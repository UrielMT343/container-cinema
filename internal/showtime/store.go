package showtime

import (
	"context"
	"fmt"
	"time"

	"start/internal/database"
)

type Store struct {
	db *database.Service
}

type dbShowtimeRow struct {
	ID             int       `db:"id"`
	StartTime      time.Time `db:"start_time"`
	MovieID        int       `db:"movie_id"`
	MovieName      string    `db:"movie_name"`
	AuditoriumID   int       `db:"auditorium_id"`
	AuditoriumName string    `db:"auditorium_name"`
}

func New(s *database.Service) *Store {
	return &Store{db: s}
}

func (s *Store) GetAllShowtimes(ctx context.Context) ([]ShowtimeDetailsResponse, error) {
	query := `
		SELECT
			s.id,
			s.start_time,
			m.id AS movie_id,
			m.name AS movie_name,
			a.id AS auditorium_id,
			a.name AS auditorium_name
		FROM showtimes s
		INNER JOIN movies m ON s.id_movie = m.id
		INNER JOIN auditoriums a ON s.id_auditorium = a.id
		-- WHERE s.start_time >= NOW()
		ORDER BY s.start_time ASC;
	`

	dbRows, err := database.QueryRows[dbShowtimeRow](s.db, ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error while getting the showtimes: %v", err)
	}

	showtimes := make([]ShowtimeDetailsResponse, 0, len(dbRows))
	for _, row := range dbRows {
		showtimes = append(showtimes, ShowtimeDetailsResponse{
			ID:        row.ID,
			StartTime: row.StartTime,
			Movie: MovieSummary{
				ID:   row.MovieID,
				Name: row.MovieName,
			},
			Auditorium: AuditoriumSummary{
				ID:   row.AuditoriumID,
				Name: row.AuditoriumName,
			},
		})
	}

	return showtimes, nil
}

func (s *Store) GetShowtimeByID(ctx context.Context, id int) (ShowtimeDetailsResponse, error) {
	pool := s.db.GetDB()

	query := `
		SELECT
			s.id,
			s.start_time,
			m.id AS movie_id,
			m.name AS movie_name,
			a.id AS auditorium_id,
			a.name AS auditorium_name
		FROM showtimes s
		INNER JOIN movies m ON s.id_movie = m.id
		INNER JOIN auditoriums a ON s.id_auditorium = a.id
		WHERE s.id = $1
		-- WHERE s.start_time >= NOW()
	`

	var showtime ShowtimeDetailsResponse
	err := pool.QueryRow(ctx, query, id).Scan(
	    &showtime.ID,
		&showtime.StartTime,
		&showtime.Movie.ID,
		&showtime.Movie.Name,
		&showtime.Auditorium.ID,
		&showtime.Auditorium.Name,
	)
	if err != nil {
		return ShowtimeDetailsResponse{}, fmt.Errorf("no showtime founded: %v", err)
	}

	return showtime, nil
}
