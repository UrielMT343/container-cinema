package ticket

import (
	"context"
	"errors"
	"fmt"

	"start/internal/database"
	"start/internal/models"

	"github.com/google/uuid"
)

type Store struct {
	db *database.Service
}

func New(s *database.Service) *Store {
	return &Store{db: s}
}

var ErrNotFound = errors.New("ticket not found")

func (s *Store) CreateTicket(ticket models.Ticket) (models.Ticket, error) {
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
		ticket.ID, ticket.IDUser, ticket.IDShowtime, ticket.Status, ticket.IDSeat,
	).Scan(&ticket.ID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return models.Ticket{}, fmt.Errorf("conflict: seat %v is already taken or held", ticket.IDSeat)
		}

		return models.Ticket{}, fmt.Errorf("error creating the ticket: %v", err)
	}

	return ticket, nil
}

func (s *Store) UpdateTicketStatus(status string, id uuid.UUID) (models.Ticket, error) {
	pool := s.db.GetDB()

	var updatedTicket models.Ticket

	query := `
		UPDATE tickets SET status = $1
		WHERE id = $2
		RETURNING id, id_user, id_showtime, status, id_seat
	`
	err := pool.QueryRow(context.Background(), query, status, id).Scan(&updatedTicket.ID, &updatedTicket.IDUser, &updatedTicket.IDShowtime, &updatedTicket.Status, &updatedTicket.IDSeat)
	if err != nil {
		return models.Ticket{}, fmt.Errorf("error updating the ticket status: %v", err)
	}

	return updatedTicket, nil
}

func (s *Store) DeleteTicket(id uuid.UUID) error {
	pool := s.db.GetDB()

	query := `
		DELETE FROM tickets
		WHERE id = $1
		AND status = 'HOLD'
	`

	tag, err := pool.Exec(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("error deleting the ticket: %v", err)
	}

	rows := tag.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
