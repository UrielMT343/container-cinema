package seat

import (
	"context"
	"fmt"
	"start/internal/database"
)

type Store struct {
	db *database.Service
}

func New(s *database.Service) *Store {
	return &Store{db: s}
}

func (s *Store) GetAllSeats() ([]Seat, error) {
	seats, err := database.QueryRows[Seat](s.db, context.Background(), "SELECT * FROM seats")
	if err != nil {
		return nil, fmt.Errorf("Error while getting all seats: %v", err)
	}

	return seats, nil
}

func (s *Store) CreateSeat(seat Seat) (int, error) {
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

func (s *Store) GetSeatsByAuditorium(idAuditorium int) ([]Seat, error) {
	query := `
		SELECT * FROM seats
		WHERE id_auditorium = $1
	`

	seats, err := database.QueryRows[Seat](s.db, context.Background(), query, idAuditorium)
	if err != nil {
		return nil, fmt.Errorf("Error while obtaining the seats: %v", err)
	}
	return seats, nil
}
