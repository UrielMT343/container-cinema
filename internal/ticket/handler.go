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
	"start/internal/response"

	"github.com/google/uuid"
)

type TicketRequest struct {
	IDUser     *int    `json:"idUser,omitempty"`
	Email      *string `json:"email,omitempty"`
	IDShowtime int     `json:"idShowtime"`
	IDSeats    []int   `json:"idSeats"`
}

type ConfirmRequest struct {
	TicketIDs []string `json:"ticketIds"`
	Email     *string  `json:"email,omitempty"`
}

type TicketResponse struct {
	ID         uuid.UUID `json:"id"`
	IDShowtime int       `json:"idShowtime"`
	Status     string    `json:"status"`
	IDSeat     int       `json:"idSeat"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

type ticketStore interface {
	UpdateTicketStatuses(ctx context.Context, ticketIDs []uuid.UUID, email *string) ([]models.Ticket, error)
	CheckShowtimeExists(ctx context.Context, idShowtime int) (bool, error)
	CreateTickets(ctx context.Context, tickets []models.Ticket) ([]models.Ticket, error)
	DeleteTicket(ctx context.Context, id uuid.UUID) error
}

type cacheService interface {
	GetCache(key string, ctx context.Context) (string, error)
    SetCache(key string, value any, ttl time.Duration, ctx context.Context) error
    BuildShowtimeSeatsKey(idShowtime int) string
    BuildCartLimitKey(cartID string) string
    BuildSeatsCheckKey(idShowtime int, idSeat int) string
    DeleteKey(key string, ctx context.Context) error
    IncrBy(key string, value int64, ctx context.Context) (int64, error)
    DecrBy(key string, value int64, ctx context.Context) (int64, error)
    SetCacheNX(key string, value any, ttl time.Duration, ctx context.Context) (bool, error)
    Expire(key string, ttl time.Duration, ctx context.Context) error
}

type queueService interface {
	PublishHoldTicket(body []byte, ttl time.Duration) (err error)
}

type Handler struct {
	store ticketStore
	queue queueService
	cache cacheService
}

func NewHandler(s ticketStore, q queueService, c cacheService) *Handler {
	return &Handler{store: s, queue: q, cache: c}
}

// ConfirmTicket confirms held tickets and marks them as sold
// @Summary Confirm held tickets
// @Description Confirm held tickets by cart ID and mark them as sold
// @Tags tickets
// @Accept json
// @Produce json
// @Param request body ConfirmRequest true "Request payload"
// @Success 200 {object} []TicketResponse "Confirmed tickets"
// @Failure 400 {object} response.ErrorResponse "Invalid ticket ID"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /user/ticket/pay [patch]
func (h *Handler) ConfirmTicket(w http.ResponseWriter, r *http.Request) {
	cartID, ok := r.Context().Value(auth.CartContextKey).(string)
	if !ok {
		response.Error(w, http.StatusInternalServerError, "Cart ID lost in transit")
		return
	}

	ctx := r.Context()

	var req ConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if len(req.TicketIDs) == 0 {
		response.Error(w, http.StatusBadRequest, "No tickets provided for checkout")
		return
	}

	var ticketUUIDSlice []uuid.UUID
	for _, idStr := range req.TicketIDs {
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			slog.Error("Frontend sent an invalid UUID string", "error", err, "bad_id", idStr)

			errorMsg := fmt.Sprintf("Invalid ticket ID format: %s", idStr)
			response.Error(w, http.StatusBadRequest, errorMsg)
			return
		}
		ticketUUIDSlice = append(ticketUUIDSlice, parsedID)
	}

	updatedTickets, errUpdate := h.store.UpdateTicketStatuses(ctx, ticketUUIDSlice, req.Email)
	if errUpdate != nil {
		slog.Error("Failed to updated tickets", "error", errUpdate)
		response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	if len(updatedTickets) == 0 {
		response.Error(w, http.StatusConflict, "Tickets expired or already sold")
		return
	}

	IDShowtime := updatedTickets[0].IDShowtime

	bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	go func(showtimeID int, cID string) {
		defer cancel()

		cacheShowtimeKey := h.cache.BuildShowtimeSeatsKey(showtimeID)
		if err := h.cache.DeleteKey(cacheShowtimeKey, bgCtx); err != nil {
			slog.Warn("Failed to invalidate showtime cache", "error", err, "key", cacheShowtimeKey)
		}

		cartLimitKey := h.cache.BuildCartLimitKey(cID)
		if err := h.cache.DeleteKey(cartLimitKey, bgCtx); err != nil {
			slog.Warn("Failed to invalidate cart cache", "error", err, "key", cartLimitKey)
		}
	}(IDShowtime, cartID)

	response.Respond(w, http.StatusOK, updatedTickets)
}

// HoldTicket holds seats for a user
// @Summary Hold tickets for seats
// @Description Hold seats for a user by creating temporary ticket reservations
// @Tags tickets
// @Accept json
// @Produce json
// @Param request body TicketRequest true "Request payload"
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

	ctx := r.Context()

	var reqTicket TicketRequest
	err := json.NewDecoder(r.Body).Decode(&reqTicket)
	if err != nil {
		slog.Error("Bad request payload", "error", err, "path", r.URL.Path)
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	showtimeExists, err := h.store.CheckShowtimeExists(ctx, reqTicket.IDShowtime)
	if err != nil {
		slog.Error("Database failed to verify showtime", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to verify showtime availability")
		return
	}

	if !showtimeExists {
		slog.Warn("User attempted to book tickets for a non-existent showtime", "showtimeID", reqTicket.IDShowtime)
		response.Error(w, http.StatusNotFound, "The selected showtime does not exist")
		return
	}

	requestedSeats := int64(len(reqTicket.IDSeats))
	if requestedSeats == 0 {
		response.Error(w, http.StatusBadRequest, "At least one seat must be selected")
		return
	}

	seatLimit := int64(config.SeatsLimit)

	cartLimitKey := h.cache.BuildCartLimitKey(cartID)

	newTotal, errIncr := h.cache.IncrBy(cartLimitKey, requestedSeats, ctx)
	if errIncr != nil {
		slog.Error("Failed to increment cart limit in Redis", "error", errIncr)
		response.Error(w, http.StatusInternalServerError, "Failed to verify cart limits")
		return
	}

	if newTotal > seatLimit {
		_, _ = h.cache.DecrBy(cartLimitKey, requestedSeats, ctx)

		slog.Warn("Seat limit exceeded", "current_total_attempted", newTotal, "limit", seatLimit)
		message := fmt.Sprintf("You can only hold %d seats at a time.", seatLimit)
		response.Error(w, http.StatusBadRequest, message)
		return
	}

	snipeTTL := 5 * time.Second

	for _, seatID := range reqTicket.IDSeats {
		snipeKey := h.cache.BuildSeatsCheckKey(reqTicket.IDShowtime, seatID)

		acquired, errSetNX := h.cache.SetCacheNX(snipeKey, cartID, snipeTTL, ctx)
		if errSetNX != nil {
			slog.Error("Redis server error during snipe guard", "error", errSetNX)
			response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
			return
		}

		if !acquired {
			slog.Warn("Seat snipe attempt blocked by Redis", "seat", seatID, "cart", cartID)
			response.Error(w, http.StatusConflict, fmt.Sprintf("Seat %d is currently being purchased by someone else!", seatID))
			return
		}
	}

	if reqTicket.IDUser != nil && *reqTicket.IDUser == 0 {
	    reqTicket.IDUser = nil
	}

	if reqTicket.Email != nil && *reqTicket.Email == "" {
	    reqTicket.Email = nil
	}

	ttl := time.Duration(config.HoldTicketTTLMinutes)

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
	}

	if errExpire := h.cache.Expire(cartID, ttl, ctx); errExpire != nil {
		slog.Error("Failed to set TTL on cart", "error", errExpire, "cartID", cartID)
	}

	insertedTickets, errInsert := h.store.CreateTickets(ctx, tickets)
	if errInsert != nil {
		if errors.Is(errInsert, ErrInsertConflict) {
			slog.Error("Database rejected ticket insert", "error", errInsert)
			response.Error(w, http.StatusConflict, ErrInsertConflict.Error())
			return
		} else {
			_, _ = h.cache.DecrBy(cartLimitKey, requestedSeats, ctx)

			slog.Error("Error while inserting the tickets", "error", errInsert)
			response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
			return
		}
	}

	body, err := json.Marshal(insertedTickets)
	if err == nil {
		errPublish := h.queue.PublishHoldTicket(body, ttl)
		if errPublish != nil {
			slog.Error("Failed to publish to queue", "error", errPublish)
		}
	}

	cacheShowtimeKey := fmt.Sprintf("seats:showtime:%d", reqTicket.IDShowtime)
	if errDelete := h.cache.DeleteKey(cacheShowtimeKey, ctx); errDelete != nil {
		slog.Warn("Failed to invalidate cache", "error", errDelete, "key", cacheShowtimeKey)
	}

	responseTickets := make([]TicketResponse, 0, len(insertedTickets))

	expiresAt := time.Now().Add(ttl)

	for _, t := range insertedTickets {
		newTicket := TicketResponse {
			ID: t.ID,
			IDShowtime: t.IDShowtime,
			IDSeat: t.IDSeat,
			Status: t.Status,
			ExpiresAt: expiresAt,
		}
		responseTickets = append(responseTickets, newTicket)
	}

	response.Respond(w, http.StatusAccepted, responseTickets)
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

	ctx := r.Context()
	errDelete := h.store.DeleteTicket(ctx, id)

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
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
	}

	http.SetCookie(w, &httpCookie)

	response.Respond(w, http.StatusOK, "Cart created")
}
