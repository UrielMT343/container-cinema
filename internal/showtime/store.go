package showtime

import (
	"context"
	"fmt"
	"start/internal/database"
	"start/internal/models"
)

type Store struct {
	db *database.Service
}

func New(s *database.Service) *Store {
	return &Store{db: s}
}

func (s *Store) GetAllShowtimes() ([]models.Showtime, error) {
	query := `
		SELECT * FROM showtimes
	`

	showtimes, err := database.QueryRows[models.Showtime](s.db, context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("Error while getting the showtimes: %v", err)
	}

	return showtimes, nil
}

func (s *Store) GetShowtimeById(id int) (models.Showtime, error) {
	pool := s.db.GetDB()

	query := `
		SELECT id, id_movie, id_auditorium, start_time FROM showtimes
		WHERE id = $1
	`

	var showtime models.Showtime
	err := pool.QueryRow(context.Background(), query, id).Scan(&showtime.Id, &showtime.IdMovie, &showtime.IdAuditorium, &showtime.StartTime)
	if err != nil {
		return models.Showtime{}, fmt.Errorf("No showtime founded: %v", err)
	}

	return showtime, nil
}
