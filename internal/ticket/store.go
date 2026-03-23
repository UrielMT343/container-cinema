package ticket

import (
	"context"
	"errors"
	"fmt"
	"start/internal/database"

	"github.com/google/uuid"
)

type Store struct {
	db *database.Service
}

func New(s *database.Service) *Store {
	return &Store{db: s}
}

var ErrNotFound = errors.New("ticket not found")

func (s *Store) CreateTicket(ticket Ticket) (uuid.UUID, error) {
	pool := s.db.GetDB()

	query := `
			INSERT INTO tickets (id, id_user, id_showtime, status, id_seat)
			SELECT $1, $2, $3, $4, $5
			WHERE NOT EXISTS (
				SELECT 1 FROM tickets
				WHERE id_showtime = $3
				AND id_seat = $5
				AND status IN ('SOLD', 'HELD')
			)
			RETURNING id;
		`

	err := pool.QueryRow(context.Background(), query,
		ticket.Id, ticket.IdUser, ticket.IdShowtime, ticket.Status, ticket.IdSeat,
	).Scan(&ticket.Id)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return uuid.Nil, fmt.Errorf("Conflict: seat %v is already taken or held", ticket.IdSeat)
		}

		return uuid.Nil, fmt.Errorf("Error creating the ticket: %v", err)
	}

	return ticket.Id, nil
}

func (s *Store) UpdateTicketStatus(status string, id uuid.UUID) (Ticket, error) {
	pool := s.db.GetDB()

	var updatedTicket Ticket

	query := `
		UPDATE tickets SET status = $1
		WHERE id = $2
		RETURNING id, id_user, id_showtime, status, id_seat
	`
	err := pool.QueryRow(context.Background(), query, status, id).Scan(&updatedTicket.Id, &updatedTicket.IdUser, &updatedTicket.IdShowtime, &updatedTicket.Status, &updatedTicket.IdSeat)

	if err != nil {
		return Ticket{}, fmt.Errorf("Error updating the ticket status: %v", err)
	}

	return updatedTicket, nil
}

func (s *Store) DeleteTicket(id uuid.UUID) error {
	pool := s.db.GetDB()

	query := `
		DELETE FROM tickets
		WHERE id = $1
	`

	tag, err := pool.Exec(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("Error deleting the ticket: %v", err)
	}

	rows := tag.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
