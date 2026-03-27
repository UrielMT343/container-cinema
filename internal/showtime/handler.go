package showtime

import (
	"encoding/json"
	"errors"
	"net/http"
	"start/internal/config"
	redisClient "start/internal/redis"
	"start/internal/response"
	"strconv"
)

type Handler struct {
	store *Store
	redis *redisClient.Redis
}

func NewHandler(s *Store, redis *redisClient.Redis) *Handler {
	return &Handler{store: s, redis: redis}
}

func (h *Handler) GetShowtimes(w http.ResponseWriter, r *http.Request) {
	key := "showtime:all"

	val, err := h.redis.GetCache(key)

	if errors.Is(err, redisClient.ErrCacheNotFound) {
		showtimes, err := h.store.GetAllShowtimes()
		if err != nil {
			response.Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		ttl := config.CacheTTLMinutes

		errSetCache := h.redis.SetCache(key, showtimes, ttl)
		if errSetCache != nil {
			response.Error(w, http.StatusInternalServerError, errSetCache.Error())
			return
		}
	} else if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	} else {
		response.Respond(w, http.StatusOK, json.RawMessage([]byte(val)))
	}
}

func (h *Handler) GetShowtimesById(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid id")
		return
	}

	showtime, err := h.store.GetShowtimeById(id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Respond(w, http.StatusOK, showtime)
}
