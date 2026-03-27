package ticket

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"start/internal/config"
	"start/internal/models"
	"start/internal/rabbitmq"
	redisClient "start/internal/redis"
	"start/internal/response"

	"github.com/google/uuid"
)

type Handler struct {
	store *Store
	queue *rabbitmq.RabbitMQ
	redis *redisClient.Redis
}

type ReqTicket struct {
	IdUser     int `json:"idUser"`
	IdShowtime int `json:"idShowtime"`
	IdSeat     int `json:"idSeat"`
}

func NewHandler(st *Store, q *rabbitmq.RabbitMQ, r *redisClient.Redis) *Handler {
	return &Handler{store: st, queue: q, redis: r}
}

func (h *Handler) ConfirmTicket(w http.ResponseWriter, r *http.Request) {
	var idStr = r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	ticket, err := h.store.UpdateTicketStatus("SOLD", id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	cacheShowtimeKey := h.redis.BuildShowtimeSeatsKey(ticket.IdShowtime)

	errDelete := h.redis.DeleteKey(cacheShowtimeKey)
	if errDelete != nil {
		fmt.Printf("WARNING: Failed to invalidate cache for %s: %v\n", cacheShowtimeKey, errDelete)
	}

	cacheHoldKey := h.redis.BuildHoldTicketKey(ticket.Id, ticket.IdShowtime)

	errDeleteHold := h.redis.DeleteKey(cacheHoldKey)
	if errDeleteHold != nil {
		fmt.Printf("WARNING: Failed to invalidate cache for %s: %v\n", cacheHoldKey, errDeleteHold)
	}

	response.Respond(w, http.StatusCreated, ticket)
}

func (h *Handler) HoldTicket(w http.ResponseWriter, r *http.Request) {
	var reqTicket ReqTicket
	err := json.NewDecoder(r.Body).Decode(&reqTicket)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	var newUuid uuid.UUID = uuid.New()
	var status string = "HOLD"
	var ticket models.Ticket = models.Ticket{Id: newUuid, IdUser: reqTicket.IdUser, IdShowtime: reqTicket.IdShowtime, Status: status, IdSeat: reqTicket.IdSeat}

	body, err := json.Marshal(ticket)

	errPublish := h.queue.PublishTicket(body)
	if errPublish != nil {
		response.Error(w, http.StatusInternalServerError, errPublish.Error())
		return
	}

	keyHoldTicket := h.redis.BuildHoldTicketKey(ticket.Id, ticket.IdShowtime)

	ttl := config.HoldTicketTTLMinutes

	errSetHold := h.redis.SetCache(keyHoldTicket, ticket.Id.String(), ttl)
	if errSetHold != nil {
		response.Error(w, http.StatusInternalServerError, errSetHold.Error())
		return
	}

	cacheShowtimeKey := fmt.Sprintf("seats:showtime:%s", strconv.Itoa(ticket.IdShowtime))

	errDelete := h.redis.DeleteKey(cacheShowtimeKey)
	if errDelete != nil {
		fmt.Printf("WARNING: Failed to invalidate cache for %s: %v\n", cacheShowtimeKey, errDelete)
	}

	response.Respond(w, http.StatusAccepted, ticket)
}

func (h *Handler) CancelTicket(w http.ResponseWriter, r *http.Request) {
	var idStr = r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	errDelete := h.store.DeleteTicket(id)

	if errDelete != nil {
		if errors.Is(errDelete, ErrNotFound) {
			response.Error(w, http.StatusNotFound, errDelete.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, errDelete.Error())
		return
	}

	response.Respond(w, http.StatusNoContent, nil)
}
