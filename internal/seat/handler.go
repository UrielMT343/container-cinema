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

// GetSeats retrieves all seats
// @Summary Get all seats
// @Description Retrieve a list of all seats in the cinema
// @Tags seats
// @Accept json
// @Produce json
// @Success 200 {object} []models.Seat "List of seats"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/seats [get]
func (h *Hander) GetSeats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	seats, err := h.store.GetAllSeats(ctx)
	if err != nil {
		slog.Error("Failed to get all seats", "error", err)
		response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	response.Respond(w, http.StatusOK, seats)
}

// InsertSeat creates a new seat
// @Summary Create a new seat
// @Description Add a new seat to the cinema's seat catalog
// @Tags seats
// @Accept json
// @Produce json
// @Param request body models.Seat true "Seat data"
// @Success 201 {object} models.Seat "Created seat"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload or validation error"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/seats [post]
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

	ctx := r.Context()

	id, err := h.store.CreateSeat(ctx, seat)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	seat.ID = id

	response.Respond(w, http.StatusCreated, seat)
}

// GetSeatsByAuditorium retrieves seats by auditorium ID
// @Summary Get seats by auditorium
// @Description Retrieve all seats belonging to a specific auditorium
// @Tags seats
// @Accept json
// @Produce json
// @Param id path int true "Auditorium ID"
// @Success 200 {object} []models.Seat "List of seats"
// @Failure 400 {object} response.ErrorResponse "Invalid auditorium ID"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/seats/auditorium/{id} [get]
func (h *Hander) GetSeatsByAuditorium(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error("Failed to cast value", "error", err, "value", idStr)
		response.Error(w, http.StatusBadRequest, "Invalid auditorium")
		return
	}

	ctx := r.Context()

	seats, err := h.store.GetSeatsByAuditorium(ctx, id)
	if err != nil {
		slog.Error("Failed to get seats", "error", err, "path", r.URL.Path)
		response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	response.Respond(w, http.StatusOK, seats)
}

// GetSeatsByShowtime retrieves seats by showtime ID
// @Summary Get seats by showtime
// @Description Retrieve seat availability for a specific showtime
// @Tags seats
// @Accept json
// @Produce json
// @Param id path int true "Showtime ID"
// @Success 200 {object} []models.Seat "List of seats with availability"
// @Failure 400 {object} response.ErrorResponse "Invalid showtime ID"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /public/seats/showtime/{id} [get]
func (h *Hander) GetSeatsByShowtime(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error("Failed to cast value", "error", err, "value", idStr)
		response.Error(w, http.StatusBadRequest, "Invalid showtime")
		return
	}

	ctx := r.Context()

	showtimeKey := h.redis.BuildShowtimeSeatsKey(id)

	val, err := h.redis.GetCache(showtimeKey, ctx)

	if errors.Is(err, redisclient.ErrCacheNotFound) {
		seats, err := h.store.GetSeatsByShowtime(ctx, id)
		if err != nil {
			slog.Error("Failed to get seats by showtime", "error", err, "showtime", id)
			response.Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		ttl := config.CacheTTLMinutes

		errSetCache := h.redis.SetCache(showtimeKey, seats, ttl, ctx)
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
