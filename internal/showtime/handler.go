package showtime

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"start/internal/config"
	redisclient "start/internal/redis"
	"start/internal/response"
)

type ShowtimeDetailsResponse struct {
	ID         int               `json:"id"`
	StartTime  time.Time         `json:"startTime"`
	Movie      MovieSummary      `json:"movie"`
	Auditorium AuditoriumSummary `json:"auditorium"`
}

type MovieSummary struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type AuditoriumSummary struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type showtimeStore interface {
	GetAllShowtimes(ctx context.Context) ([]ShowtimeDetailsResponse, error)
	GetShowtimeByID(ctx context.Context, id int) (ShowtimeDetailsResponse, error)
}

type cacheService interface {
	GetCache(key string, ctx context.Context) (string, error)
	SetCache(key string, value any, ttl time.Duration, ctx context.Context) error
	BuildAllShowtimesKey() string
}

type Handler struct {
	store showtimeStore
	cache cacheService
}

func NewHandler(s showtimeStore, c cacheService) *Handler {
	return &Handler{store: s, cache: c}
}

// GetShowtimes retrieves all showtimes
// @Summary Get all showtimes
// @Description Retrieve a list of all movie showtimes
// @Tags showtimes
// @Accept json
// @Produce json
// @Success 200 {object} []ShowtimeDetailsResponse "List of showtimes"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /public/showtimes [get]
func (h *Handler) GetShowtimes(w http.ResponseWriter, r *http.Request) {
	key := h.cache.BuildAllShowtimesKey()

	ctx := r.Context()

	val, err := h.cache.GetCache(key, ctx)

	if errors.Is(err, redisclient.ErrCacheNotFound) {
		showtimes, errGet := h.store.GetAllShowtimes(ctx)
		if errGet != nil {
			slog.Error("Failed to get showtimes", "error", errGet)
			response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
			return
		}

		ttl := config.CacheTTLMinutes

		errSetCache := h.cache.SetCache(key, showtimes, ttl, ctx)
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
// @Success 200 {object} ShowtimeDetailsResponse "Showtime details"
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

	ctx := r.Context()
	showtime, err := h.store.GetShowtimeByID(ctx, id)
	if err != nil {
		slog.Error("Failed to get showtime", "error", err, "path", r.URL.Path)
		response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	response.Respond(w, http.StatusOK, showtime)
}
