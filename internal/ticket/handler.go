package ticket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"start/internal/auth"
	"start/internal/config"
	"start/internal/models"
	"start/internal/rabbitmq"
	redisclient "start/internal/redis"
	"start/internal/response"

	"github.com/google/uuid"
)

type Handler struct {
	store *Store
	queue *rabbitmq.RabbitMQ
	redis *redisclient.Redis
}

type ReqTicket struct {
	IDUser     int   `json:"idUser"`
	IDShowtime int   `json:"idShowtime"`
	IDSeats    []int `json:"idSeats"`
}

func NewHandler(st *Store, q *rabbitmq.RabbitMQ, r *redisclient.Redis) *Handler {
	return &Handler{store: st, queue: q, redis: r}
}

// ConfirmTicket confirms held tickets and marks them as sold
// @Summary Confirm held tickets
// @Description Confirm held tickets by cart ID and mark them as sold
// @Tags tickets
// @Accept json
// @Produce json
// @Success 200 {object} []models.Ticket "Confirmed tickets"
// @Failure 400 {object} response.ErrorResponse "Invalid ticket ID"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /user/ticket/pay [patch]
func (h *Handler) ConfirmTicket(w http.ResponseWriter, r *http.Request) {
	cartID, ok := r.Context().Value(auth.CartContextKey).(string)
	if !ok {
		response.Error(w, http.StatusInternalServerError, "Cart ID lost in transit")
		return
	}

	ticketIDsStr, err := h.redis.SetMembers(cartID)
	if err != nil {
		slog.Error("Failed to retrieve cart items", "error", err)
		response.Error(w, http.StatusInternalServerError, "Cart items lost")
		return
	}

	var ticketUUIDSlice []uuid.UUID

	for _, s := range ticketIDsStr {
		id, err := uuid.Parse(s)
		if err != nil {
			slog.Error("Failed to parse the ticket ID", "error", err, "ticketID", s)
			response.Error(w, http.StatusBadRequest, "Invalid ticket ID")
			return
		}

		ticketUUIDSlice = append(ticketUUIDSlice, id)
	}

	ctx := context.Background()
	updatedTickets, errUpdate := h.store.UpdateTicketStatuses(ctx, "SOLD", ticketUUIDSlice)
	if errUpdate != nil {
		slog.Error("Failed to updated tickets", "error", errUpdate)
		response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	IDShowtime := updatedTickets[0].IDShowtime

	cacheShowtimeKey := h.redis.BuildShowtimeSeatsKey(IDShowtime)

	errDelete := h.redis.DeleteKey(cacheShowtimeKey)
	if errDelete != nil {
		slog.Warn("Failed to invalidate cache", "error", errDelete, "key", cacheShowtimeKey)
	}

	for _, ticket := range updatedTickets {
		cacheHoldKey := h.redis.BuildHoldTicketKey(ticket.ID, ticket.IDShowtime)
		errDeleteHold := h.redis.DeleteKey(cacheHoldKey)
		if errDeleteHold != nil {
			slog.Warn("Failed to invalidate cache", "error", errDeleteHold, "key", cacheHoldKey)
		}
	}

	response.Respond(w, http.StatusOK, updatedTickets)
}

// HoldTicket holds seats for a user
// @Summary Hold tickets for seats
// @Description Hold seats for a user by creating temporary ticket reservations
// @Tags tickets
// @Accept json
// @Produce json
// @Param request body ReqTicket true "Request payload"
// @Success 202 {object} []models.Ticket "Held tickets"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload or seat selection"
// @Failure 409 {object} response.ErrorResponse "Seat already taken"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /user/ticket/hold [post]
func (h *Handler) HoldTicket(w http.ResponseWriter, r *http.Request) {
	cartID, ok := r.Context().Value(auth.CartContextKey).(string)
	if !ok {
		response.Error(w, http.StatusInternalServerError, "Cart ID lost in transit")
		return
	}
	slog.Info("Holding tickets for cart", "cart_id", cartID)

	var reqTicket ReqTicket
	err := json.NewDecoder(r.Body).Decode(&reqTicket)
	if err != nil {
		slog.Error("Bad request payload", "error", err, "path", r.URL.Path)
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	requestedSeats := int64(len(reqTicket.IDSeats))
	if requestedSeats == 0 {
		response.Error(w, http.StatusBadRequest, "At least one seat must be selected")
		return
	}

	seatLimit := int64(config.SeatsLimit)
	currentSeats, _ := h.redis.SetCardinality(cartID)

	if currentSeats+requestedSeats > seatLimit {
		slog.Error("Too many seats taken", "current", currentSeats, "requested", requestedSeats)
		message := fmt.Sprintf("You can only hold %d seats at a time.", seatLimit)
		response.Error(w, http.StatusBadRequest, message)
		return
	}

	ttl := time.Duration(config.HoldTicketTTLMinutes)

	for _, seatID := range reqTicket.IDSeats {
		snipeKey := fmt.Sprintf("showtime:%d:seat:%d", reqTicket.IDShowtime, seatID)

		acquired, errSetNX := h.redis.SetCacheNX(snipeKey, cartID, ttl)
		if errSetNX != nil {
			slog.Error("Redis server error during snipe guard", "error", errSetNX)
			response.Error(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}
		if !acquired {
			slog.Warn("Seat snipe attempt blocked", "seat", seatID, "cart", cartID)
			response.Error(w, http.StatusConflict, fmt.Sprintf("Seat %d was just taken!", seatID))
			return
		}
	}

	var tickets []models.Ticket
	for _, seatID := range reqTicket.IDSeats {
		ticket := models.Ticket{
			ID:         uuid.New(),
			IDUser:     reqTicket.IDUser,
			IDShowtime: reqTicket.IDShowtime,
			Status:     "HOLD",
			IDSeat:     seatID,
		}
		tickets = append(tickets, ticket)

		keyHoldTicket := h.redis.BuildHoldTicketKey(ticket.ID, ticket.IDShowtime)
		if errSetHold := h.redis.SetCache(keyHoldTicket, ticket.ID.String(), ttl); errSetHold != nil {
			slog.Error("Failed to set cache", "error", errSetHold, "key", keyHoldTicket)
		}

		if errAdd := h.redis.SetAdd(cartID, ticket.ID.String()); errAdd != nil {
			slog.Error("Failed to add to cart key", "error", errAdd, "cartID", cartID)
		}
	}

	if errExpire := h.redis.Expire(cartID, ttl); errExpire != nil {
		slog.Error("Failed to set TTL on cart", "error", errExpire, "cartID", cartID)
	}

	body, err := json.Marshal(tickets)
	if err == nil {
		errPublish := h.queue.PublishTicket(body)
		if errPublish != nil {
			slog.Error("Failed to publish to queue", "error", errPublish)
		}
	}

	cacheShowtimeKey := fmt.Sprintf("seats:showtime:%d", reqTicket.IDShowtime)
	if errDelete := h.redis.DeleteKey(cacheShowtimeKey); errDelete != nil {
		slog.Warn("Failed to invalidate cache", "error", errDelete, "key", cacheShowtimeKey)
	}

	response.Respond(w, http.StatusAccepted, tickets)
}

// CancelTicket cancels a ticket by ID
// @Summary Cancel a ticket
// @Description Cancel a ticket by its ID
// @Tags tickets
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Success 204 "Ticket cancelled successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid ticket ID"
// @Failure 404 {object} response.ErrorResponse "Ticket not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /user/ticket/{id} [delete]
func (h *Handler) CancelTicket(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		slog.Error("Failed to cast value", "error", err, "value", idStr)
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	errDelete := h.store.DeleteTicket(id)

	if errDelete != nil {
		if errors.Is(errDelete, ErrNotFound) {
			slog.Error("Failed to get the data", "error", ErrNotFound)
			response.Error(w, http.StatusNotFound, ErrNotFound.Error())
			return
		}
		slog.Error("Failed to delete the ticket", "error", errDelete)
		response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	response.Respond(w, http.StatusNoContent, nil)
}

// BeginCheckout begins the checkout process by creating a cart
// @Summary Begin checkout process
// @Description Create a new cart for the checkout process
// @Tags tickets
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "Cart created"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /public/checkout/begin [post]
func (h *Handler) BeginCheckout(w http.ResponseWriter, r *http.Request) {
	cartID := uuid.New()
	cartIDstr := "cart_" + cartID.String()

	tokenExp := config.CartIDCookieTTLMinutes

	httpCookie := http.Cookie{
		Name:     "cart_id",
		Value:    cartIDstr,
		Expires:  time.Now().Add(tokenExp),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}

	http.SetCookie(w, &httpCookie)

	response.Respond(w, http.StatusOK, "Cart created")
}
