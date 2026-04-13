package ticket

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"start/internal/database"
	"start/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Store struct {
	db *database.Service
}

func New(s *database.Service) *Store {
	return &Store{db: s}
}

var ErrNotFound = errors.New("ticket not found")

func (s *Store) CreateTicket(ctx context.Context, tickets []models.Ticket) (insertedTickets []models.Ticket, err error) {
	if len(tickets) == 0 {
		return nil, nil
	}

	pool := s.db.GetDB()

	var ids []uuid.UUID
	var seatIDs []int

	idUser := tickets[0].IDUser
	idShowtime := tickets[0].IDShowtime
	status := tickets[0].Status

	for _, t := range tickets {
		ids = append(ids, t.ID)
		seatIDs = append(seatIDs, t.IDSeat)
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("error while beginning transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	query := `
		WITH input AS (
			SELECT unnest($1::uuid[]) AS id, unnest($2::int[]) AS id_seat
		)
		INSERT INTO tickets (id, id_user, id_showtime, status, id_seat)
		SELECT i.id, $3, $4, $5, i.id_seat
		FROM input i
		WHERE NOT EXISTS (
			SELECT 1 FROM tickets t
			WHERE t.id_showtime = $4
			AND t.id_seat = ANY($2::int[])
			AND t.status IN ('SOLD', 'HELD')
		)
		RETURNING id;
	`

	rows, err := tx.Query(ctx, query, ids, seatIDs, idUser, idShowtime, status)
	if err != nil {
		return nil, fmt.Errorf("error executing bulk insert: %w", err)
	}
	defer rows.Close()

	var insertedCount int
	for rows.Next() {
		insertedCount++
	}

	if insertedCount == 0 {
		return nil, fmt.Errorf("conflict: one or more seats in %v are already taken or held", seatIDs)
	}

	errCommit := tx.Commit(ctx)
	if errCommit != nil {
		return nil, fmt.Errorf("could not commit transaction, error: %v", errCommit)
	}

	return tickets, nil
}

func (s *Store) UpdateTicketStatuses(ctx context.Context, status string, ids []uuid.UUID) ([]models.Ticket, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	pool := s.db.GetDB()

	query := `
		UPDATE tickets
		SET status = $1
		WHERE id = ANY($2::uuid[])
		RETURNING id, id_user, id_showtime, status, id_seat;
	`

	rows, err := pool.Query(ctx, query, status, ids)
	if err != nil {
		return nil, fmt.Errorf("error executing bulk update: %w", err)
	}
	defer rows.Close()

	var updatedTickets []models.Ticket

	for rows.Next() {
		var t models.Ticket
		err := rows.Scan(&t.ID, &t.IDUser, &t.IDShowtime, &t.Status, &t.IDSeat)
		if err != nil {
			return nil, fmt.Errorf("error scanning updated ticket row: %w", err)
		}
		updatedTickets = append(updatedTickets, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over updated rows: %w", err)
	}

	if len(updatedTickets) != len(ids) {
		slog.Warn("Mismatch in updated tickets", "requested", len(ids), "updated", len(updatedTickets))
	}

	return updatedTickets, nil
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
