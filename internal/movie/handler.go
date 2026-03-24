package movie

import (
	"encoding/json"
	"net/http"
	"start/internal/models"
	"start/internal/response"
	"strconv"
)

type Handler struct {
	store *Store
}

func NewHandler(st *Store) *Handler {
	return &Handler{store: st}
}

func (h *Handler) GetMovies(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	pageStr := r.URL.Query().Get("page")

	var limitDefault int = 20
	var pageDefault int = 1

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit > 100 || limit < 1 || limitStr == "" {
		limit = limitDefault
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 || pageStr == "" {
		page = pageDefault
	}

	offset := (page - 1) * limit

	movies, err := h.store.GetAllMovies(limit, offset)

	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.Respond(w, http.StatusOK, movies)
}

func (h *Handler) InsertMovie(w http.ResponseWriter, r *http.Request) {
	var movie models.Movie
	err := json.NewDecoder(r.Body).Decode(&movie)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.store.CreateMovie(movie)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	movie.Id = id

	response.Respond(w, http.StatusCreated, movie)
}
