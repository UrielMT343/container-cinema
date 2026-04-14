package movie

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"start/internal/models"
	"start/internal/response"
)

type Handler struct {
	store *Store
}

func NewHandler(st *Store) *Handler {
	return &Handler{store: st}
}

// GetMovies retrieves all movies
// @Summary Get all movies
// @Description Retrieve a paginated list of all movies
// @Tags movies
// @Accept json
// @Produce json
// @Param limit query int false "Number of items to return"
// @Param page query int false "Page number"
// @Success 200 {object} []models.Movie "List of movies"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/movies [get]
func (h *Handler) GetMovies(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	pageStr := r.URL.Query().Get("page")

	limitDefault := 20
	pageDefault := 1

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
		slog.Error("Failed to get all movies", "error", err)
		response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	response.Respond(w, http.StatusOK, movies)
}

// InsertMovie creates a new movie
// @Summary Create a new movie
// @Description Add a new movie to the cinema's catalog
// @Tags movies
// @Accept json
// @Produce json
// @Param request body models.Movie true "Movie data"
// @Success 201 {object} models.Movie "Created movie"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload or validation error"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/movies [post]
func (h *Handler) InsertMovie(w http.ResponseWriter, r *http.Request) {
	var movie models.Movie
	err := json.NewDecoder(r.Body).Decode(&movie)
	if err != nil {
		slog.Error("Bad request on payload", "error", err, "path", r.URL.Path)
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	errValidate := movie.Validate()
	if errValidate != nil {
		slog.Error("Bad request on payload", "error", errValidate, "path", r.URL.Path)
		response.Error(w, http.StatusBadRequest, errValidate.Error())
		return
	}

	id, err := h.store.CreateMovie(movie)
	if err != nil {
		slog.Error("Failed to create new movie", "error", err)
		response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	movie.ID = id

	response.Respond(w, http.StatusCreated, movie)
}
