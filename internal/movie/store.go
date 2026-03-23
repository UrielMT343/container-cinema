package movie

import (
	"context"
	"fmt"
	"start/internal/database"
)

type Store struct {
	db *database.Service
}

type MoviePage struct {
	Data       []Movie `json:"movies"`
	TotalRows  int     `json:"total_rows"`
	TotalPages int     `json:"total_pages"`
}

func New(s *database.Service) *Store {
	return &Store{db: s}
}

func (s *Store) GetAllMovies(limit int, offset int) (MoviePage, error) {
	query := `
		SELECT * FROM movies
		LIMIT $1
		OFFSET $2
	`

	movies, err := database.QueryRows[Movie](s.db, context.Background(), query, limit, offset)
	if err != nil {
		return MoviePage{}, fmt.Errorf("Error while running the query: %v", err)
	}

	pool := s.db.GetDB()

	countQuery := `
		SELECT COUNT(*) FROM movies
	`

	var count int
	errScan := pool.QueryRow(context.Background(), countQuery).Scan(&count)
	if errScan != nil {
		return MoviePage{}, fmt.Errorf("Error while counting the rows: %v", errScan)
	}

	totalPages := count / limit

	if count%limit != 0 {
		totalPages += 1
	}

	var moviePage MoviePage = MoviePage{Data: movies, TotalRows: count, TotalPages: totalPages}

	return moviePage, nil
}

func (s *Store) CreateMovie(m Movie) (int, error) {
	pool := s.db.GetDB()

	query := `
		INSERT INTO movies (name, duration_min, synopsis)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	err := pool.QueryRow(context.Background(), query, m.Name, m.DurationMin, m.Synopsis).Scan(&m.Id)

	if err != nil {
		return 0, fmt.Errorf("Error while running the query: %v", err)
	}

	return m.Id, nil
}
