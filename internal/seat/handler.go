package seat

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"start/internal/config"
	"start/internal/models"
	redisclient "start/internal/redis"
	"start/internal/response"
)

type Hander struct {
	store *Store
	redis *redisclient.Redis
}

func NewHandler(s *Store, r *redisclient.Redis) *Hander {
	return &Hander{store: s, redis: r}
}

func (h *Hander) GetSeats(w http.ResponseWriter, r *http.Request) {
	seats, err := h.store.GetAllSeats()
	if err != nil {
		slog.Error("Failed to get all seats", "error", err)
		response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	response.Respond(w, http.StatusOK, seats)
}

func (h *Hander) InsertSeat(w http.ResponseWriter, r *http.Request) {
	var seat models.Seat
	err := json.NewDecoder(r.Body).Decode(&seat)
	if err != nil {
		slog.Error("Bad request on payload", "error", err, "path", r.URL.Path)
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	errValidate := seat.Validate()
	if errValidate != nil {
		slog.Error("Bad request on payload", "error", errValidate, "path", r.URL.Path)
		response.Error(w, http.StatusBadRequest, errValidate.Error())
		return
	}

	id, err := h.store.CreateSeat(seat)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	seat.ID = id

	response.Respond(w, http.StatusCreated, seat)
}

func (h *Hander) GetSeatsByAuditorium(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error("Failed to cast value", "error", err, "value", idStr)
		response.Error(w, http.StatusBadRequest, "Invalid auditorium")
		return
	}

	seats, err := h.store.GetSeatsByAuditorium(id)
	if err != nil {
		slog.Error("Failed to get seats", "error", err, "path", r.URL.Path)
		response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	response.Respond(w, http.StatusOK, seats)
}

func (h *Hander) GetSeatsByShowtime(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error("Failed to cast value", "error", err, "value", idStr)
		response.Error(w, http.StatusBadRequest, "Invalid showtime")
		return
	}

	showtimeKey := h.redis.BuildShowtimeSeatsKey(id)

	val, err := h.redis.GetCache(showtimeKey)

	if errors.Is(err, redisclient.ErrCacheNotFound) {
		seats, err := h.store.GetSeatsByShowtime(id)
		if err != nil {
			slog.Error("Failed to get seats by showtime", "error", err, "showtime", id)
			response.Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		ttl := config.CacheTTLMinutes

		errSetCache := h.redis.SetCache(showtimeKey, seats, ttl)
		if errSetCache != nil {
			slog.Error("Failed to set cache", "error", errSetCache, "key", showtimeKey)
			response.Error(w, http.StatusInternalServerError, errSetCache.Error())
			return
		}

		response.Respond(w, http.StatusOK, seats)
	} else if err != nil {
		slog.Error("Failed to get cache", "error", err, "key", showtimeKey)
		response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
		return
	} else {
		response.Respond(w, http.StatusOK, json.RawMessage([]byte(val)))
	}
}
