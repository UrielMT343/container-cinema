package movie

import (
	"context"
	"fmt"

	"start/internal/database"
	"start/internal/models"
)

type Store struct {
	db *database.Service
}

type MoviePage struct {
	Data       []models.Movie `json:"movies"`
	TotalRows  int            `json:"total_rows"`
	TotalPages int            `json:"total_pages"`
}

func New(s *database.Service) *Store {
	return &Store{db: s}
}

func (s *Store) GetAllMovies(ctx context.Context, limit int, offset int) (MoviePage, error) {
	query := `
		SELECT * FROM movies
		LIMIT $1
		OFFSET $2
	`

	movies, err := database.QueryRows[models.Movie](s.db, ctx, query, limit, offset)
	if err != nil {
		return MoviePage{}, fmt.Errorf("error while running the query: %v", err)
	}

	pool := s.db.GetDB()

	countQuery := `
		SELECT COUNT(*) FROM movies
	`

	var count int
	errScan := pool.QueryRow(context.Background(), countQuery).Scan(&count)
	if errScan != nil {
		return MoviePage{}, fmt.Errorf("error while counting the rows: %v", errScan)
	}

	totalPages := count / limit

	if count%limit != 0 {
		totalPages += 1
	}

	moviePage := MoviePage{Data: movies, TotalRows: count, TotalPages: totalPages}

	return moviePage, nil
}

func (s *Store) CreateMovie(ctx context.Context, m models.Movie) (int, error) {
	pool := s.db.GetDB()

	query := `
		INSERT INTO movies (name, duration_min, synopsis)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	err := pool.QueryRow(ctx, query, m.Name, m.DurationMin, m.Synopsis).Scan(&m.ID)
	if err != nil {
		return 0, fmt.Errorf("error while running the query: %v", err)
	}

	return m.ID, nil
}
