package ticket

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

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
	IDUser     int `json:"idUser"`
	IDShowtime int `json:"idShowtime"`
	IDSeat     int `json:"idSeat"`
}

func NewHandler(st *Store, q *rabbitmq.RabbitMQ, r *redisclient.Redis) *Handler {
	return &Handler{store: st, queue: q, redis: r}
}

func (h *Handler) ConfirmTicket(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		slog.Error("Failed to cast value", "error", err, "value", idStr)
		response.Error(w, http.StatusBadRequest, "An unexpected error occurred")
		return
	}

	ticket, err := h.store.UpdateTicketStatus("SOLD", id)
	if err != nil {
		slog.Error("Failed to update ticket status", "error", err)
		response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	cacheShowtimeKey := h.redis.BuildShowtimeSeatsKey(ticket.IDShowtime)

	errDelete := h.redis.DeleteKey(cacheShowtimeKey)
	if errDelete != nil {
		slog.Warn("Failed to invalidate cache", "error", errDelete, "key", cacheShowtimeKey)
	}

	cacheHoldKey := h.redis.BuildHoldTicketKey(ticket.ID, ticket.IDShowtime)

	errDeleteHold := h.redis.DeleteKey(cacheHoldKey)
	if errDeleteHold != nil {
		slog.Warn("Failed to invalidate cache", "error", errDeleteHold, "key", cacheHoldKey)
	}

	response.Respond(w, http.StatusCreated, ticket)
}

func (h *Handler) HoldTicket(w http.ResponseWriter, r *http.Request) {
	var reqTicket ReqTicket
	err := json.NewDecoder(r.Body).Decode(&reqTicket)
	if err != nil {
		slog.Error("Bad request payload", "error", err, "path", r.URL.Path)
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	newUUID := uuid.New()
	status := "HOLD"
	ticket := models.Ticket{ID: newUUID, IDUser: reqTicket.IDUser, IDShowtime: reqTicket.IDShowtime, Status: status, IDSeat: reqTicket.IDSeat}

	body, err := json.Marshal(ticket)
	if err != nil {
		slog.Error("Failed to format the data", "error", err)
		response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
	}

	errPublish := h.queue.PublishTicket(body)
	if errPublish != nil {
		slog.Error("Failed to publish to queue", "error", errPublish)
		response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	keyHoldTicket := h.redis.BuildHoldTicketKey(ticket.ID, ticket.IDShowtime)

	ttl := config.HoldTicketTTLMinutes

	errSetHold := h.redis.SetCache(keyHoldTicket, ticket.ID.String(), ttl)
	if errSetHold != nil {
		slog.Error("Failed to set cache", "error", errSetHold, "key", keyHoldTicket)
		response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	cacheShowtimeKey := fmt.Sprintf("seats:showtime:%s", strconv.Itoa(ticket.IDShowtime))

	errDelete := h.redis.DeleteKey(cacheShowtimeKey)
	if errDelete != nil {
		slog.Warn("Failed to invalidate cache", "error", errDelete, "key", cacheShowtimeKey)
	}

	response.Respond(w, http.StatusAccepted, ticket)
}

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
