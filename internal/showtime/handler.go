package showtime

import (
	"net/http"
	"start/internal/response"
	"strconv"
)

type Handler struct {
	store *Store
}

func NewHandler(s *Store) *Handler {
	return &Handler{store: s}
}

func (h *Handler) GetShowtimes(w http.ResponseWriter, r *http.Request) {
	showtimes, err := h.store.GetAllShowtimes()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Respond(w, http.StatusOK, showtimes)
}

func (h *Handler) GetShowtimesById(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid id")
	}

	showtime, err := h.store.GetShowtimeById(id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Respond(w, http.StatusOK, showtime)
}
