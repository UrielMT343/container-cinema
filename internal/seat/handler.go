package seat

import (
	"encoding/json"
	"net/http"
	"start/internal/response"
	"strconv"
)

type Hander struct {
	store *Store
}

func NewHandler(s *Store) *Hander {
	return &Hander{store: s}
}

func (h *Hander) GetSeats(w http.ResponseWriter, r *http.Request) {
	seats, err := h.store.GetAllSeats()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Respond(w, http.StatusOK, seats)
}

func (h *Hander) InsertSeat(w http.ResponseWriter, r *http.Request) {
	var seat Seat
	err := json.NewDecoder(r.Body).Decode(&seat)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.store.CreateSeat(seat)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	seat.Id = id

	response.Respond(w, http.StatusCreated, seat)
}

func (h *Hander) GetSeatsByAuditorium(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid auditorium")
	}

	seats, err := h.store.GetSeatsByAuditorium(id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Respond(w, http.StatusOK, seats)
}
