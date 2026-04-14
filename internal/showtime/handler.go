package showtime

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"start/internal/config"
	redisclient "start/internal/redis"
	"start/internal/response"
)

type Handler struct {
	store *Store
	redis *redisclient.Redis
}

func NewHandler(s *Store, redis *redisclient.Redis) *Handler {
	return &Handler{store: s, redis: redis}
}

// GetShowtimes retrieves all showtimes
// @Summary Get all showtimes
// @Description Retrieve a list of all movie showtimes
// @Tags showtimes
// @Accept json
// @Produce json
// @Success 200 {object} []models.Showtime "List of showtimes"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /public/showtimes [get]
func (h *Handler) GetShowtimes(w http.ResponseWriter, r *http.Request) {
	key := "showtime:all"

	val, err := h.redis.GetCache(key)

	if errors.Is(err, redisclient.ErrCacheNotFound) {
		showtimes, err := h.store.GetAllShowtimes()
		if err != nil {
			slog.Error("Failed to get showtimes", "error", err)
			response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
			return
		}

		ttl := config.CacheTTLMinutes

		errSetCache := h.redis.SetCache(key, showtimes, ttl)
		if errSetCache != nil {
			slog.Error("Failed to set cache", "error", errSetCache, "key", key)
			response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
			return
		}
	} else if err != nil {
		slog.Error("Failed to get showtimes cache", "error", err, "key", key)
		response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
		return
	} else {
		response.Respond(w, http.StatusOK, json.RawMessage([]byte(val)))
	}
}

// GetShowtimesByID retrieves a showtime by ID
// @Summary Get showtime by ID
// @Description Retrieve details of a specific showtime by its ID
// @Tags showtimes
// @Accept json
// @Produce json
// @Param id path int true "Showtime ID"
// @Success 200 {object} models.Showtime "Showtime details"
// @Failure 400 {object} response.ErrorResponse "Invalid showtime ID"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /public/showtimes/{id} [get]
func (h *Handler) GetShowtimesByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error("Failed to cast value", "error", err, "path", r.URL.Path)
		response.Error(w, http.StatusBadRequest, "Invalid id")
		return
	}

	showtime, err := h.store.GetShowtimeByID(id)
	if err != nil {
		slog.Error("Failed to get showtime", "error", err, "path", r.URL.Path)
		response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	response.Respond(w, http.StatusOK, showtime)
}
