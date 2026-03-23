package ticket

import (
	"encoding/json"
	"errors"
	"net/http"
	"start/internal/rabbitmq"
	"start/internal/response"

	"github.com/google/uuid"
)

type Handler struct {
	store *Store
	queue *rabbitmq.RabbitMQ
}

type ReqTicket struct {
	IdUser     int `json:"idUser"`
	IdShowtime int `json:"idShowtime"`
	IdSeat     int `json:"idSeat"`
}

func NewHandler(st *Store, q *rabbitmq.RabbitMQ) *Handler {
	return &Handler{store: st, queue: q}
}

func (h *Handler) ConfirmTicket(w http.ResponseWriter, r *http.Request) {
	var idStr = r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
	}

	ticket, err := h.store.UpdateTicketStatus("SOLD", id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
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
	var ticket Ticket = Ticket{Id: newUuid, IdUser: reqTicket.IdUser, IdShowtime: reqTicket.IdShowtime, Status: status, IdSeat: reqTicket.IdSeat}

	id, err := h.store.CreateTicket(ticket)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	ticket.Id = id

	response.Respond(w, http.StatusCreated, ticket)
}

func (h *Handler) CancelTicket(w http.ResponseWriter, r *http.Request) {
	var idStr = r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
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
