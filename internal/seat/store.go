package seat

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

func (s *Store) GetAllSeats() ([]models.Seat, error) {
	query := `
		SELECT * FROM seats
	`

	seats, err := database.QueryRows[models.Seat](s.db, context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("Error while getting all seats: %v", err)
	}

	return seats, nil
}

func (s *Store) CreateSeat(seat models.Seat) (int, error) {
	pool := s.db.GetDB()

	if err := seat.Validate(); err != nil {
		return 0, err
	}

	query := `
		INSERT INTO seats (number, id_auditorium)
		VALUES ($1, $2)
		RETURNING id
	`

	err := pool.QueryRow(context.Background(), query, seat.Number, seat.IdAuditorium).Scan(&seat.Id)

	if err != nil {
		return 0, fmt.Errorf("Error while creating the seat: %v", err)
	}

	return seat.Id, nil
}

func (s *Store) GetSeatsByAuditorium(idAuditorium int) ([]models.Seat, error) {
	query := `
		SELECT * FROM seats
		WHERE id_auditorium = $1
	`

	seats, err := database.QueryRows[models.Seat](s.db, context.Background(), query, idAuditorium)
	if err != nil {
		return nil, fmt.Errorf("Error while obtaining the seats: %v", err)
	}
	return seats, nil
}

func (s *Store) GetSeatsByShowtime(idShowtime int) ([]models.ShowtimeSeat, error) {
	query := `
		SELECT
		    s.id AS id,
		    s.number AS number,
		    COALESCE(t.status, 'AVAILABLE') AS status
		FROM showtimes st
		JOIN seats s ON st.id_auditorium = s.id_auditorium
		LEFT JOIN tickets t ON s.id = t.id_seat AND t.id_showtime = st.id
		WHERE st.id = $1;
	`

	seats, err := database.QueryRows[models.ShowtimeSeat](s.db, context.Background(), query, idShowtime)
	if err != nil {
		return nil, fmt.Errorf("Error while obtaining the seats: %v", err)
	}
	return seats, nil
}
