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

var ErrInsertConflict error

func (s *Store) CreateTickets(ctx context.Context, tickets []models.Ticket) (insertedTickets []models.Ticket, err error) {
	if len(tickets) == 0 {
		return nil, nil
	}

	pool := s.db.GetDB()

	var ids []uuid.UUID
	var seatIDs []int

	idUser := tickets[0].IDUser
	idShowtime := tickets[0].IDShowtime
	status := tickets[0].Status
	email := tickets[0].Email

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
		INSERT INTO tickets (id, id_user, id_showtime, status, id_seat, email)
		SELECT i.id, $3, $4, $5, i.id_seat, $6
		FROM input i
		WHERE NOT EXISTS (
			SELECT 1 FROM tickets t
			WHERE t.id_showtime = $4
			AND t.id_seat = ANY($2::int[])
			AND t.status IN ('SOLD', 'HOLD')
		)
		RETURNING id;
	`

	rows, err := tx.Query(ctx, query, ids, seatIDs, idUser, idShowtime, status, email)
	if err != nil {
		return nil, fmt.Errorf("error executing bulk insert: %w", err)
	}
	defer rows.Close()

	var insertedCount int
	for rows.Next() {
		insertedCount++
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("fatal Postgres error: %w", err)
	}

	if insertedCount == 0 {
		strError := fmt.Sprintf("conflict: one or more seats in %d are already taken or held", seatIDs)
		ErrInsertConflict = errors.New(strError)
		return nil, ErrInsertConflict
	}

	errCommit := tx.Commit(ctx)
	if errCommit != nil {
		return nil, fmt.Errorf("could not commit transaction, error: %v", errCommit)
	}

	return tickets, nil
}

func (s *Store) UpdateTicketStatuses(ctx context.Context, ticketIDs []uuid.UUID, email *string) ([]models.Ticket, error) {
	if len(ticketIDs) == 0 {
		return nil, nil
	}

	pool := s.db.GetDB()

	query := `
		UPDATE tickets
		SET status = 'SOLD',
			email = COALESCE($1, email)
		WHERE id = ANY($2::uuid[])
		RETURNING *;
	`

	rows, err := pool.Query(ctx, query, email, ticketIDs)
	if err != nil {
		return nil, fmt.Errorf("error executing bulk update: %w", err)
	}
	defer rows.Close()

	var updatedTickets []models.Ticket

	for rows.Next() {
		var t models.Ticket
		err := rows.Scan(&t.ID, &t.IDUser, &t.IDShowtime, &t.Status, &t.IDSeat, &t.Email)
		if err != nil {
			return nil, fmt.Errorf("error scanning updated ticket row: %w", err)
		}
		updatedTickets = append(updatedTickets, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over updated rows: %w", err)
	}

	if len(updatedTickets) != len(ticketIDs) {
		slog.Warn("Mismatch in updated tickets", "requested", len(ticketIDs), "updated", len(updatedTickets))
	}

	return updatedTickets, nil
}

func (s *Store) DeleteTicket(ctx context.Context, id uuid.UUID) error {
	pool := s.db.GetDB()

	query := `
		DELETE FROM tickets
		WHERE id = $1
		AND status = 'HOLD'
	`

	tag, err := pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting the ticket: %v", err)
	}

	rows := tag.RowsAffected()
	if rows == 0 {
		return models.ErrorTicketNotHeld
	}

	return nil
}

func (s *Store) CheckIfSeatOccupied(ctx context.Context, seatIDs []int, showtimeID int) ([]string, error) {
	if len(seatIDs) == 0 {
		return nil, nil
	}

	pool := s.db.GetDB()

	var occupiedSeats []string

	query := `
		SELECT COALESCE(
  			array_agg(DISTINCT s.number ORDER BY s.number),
   			ARRAY[]::text[]
		) AS occupied_seat_numbers
		FROM tickets t
		INNER JOIN seats s ON s.id = t.id_seat
		WHERE t.id_showtime = $1
	  	AND t.id_seat = ANY($2::int[])
	  	AND t.status IN ('HELD', 'SOLD');
	`
	err := pool.QueryRow(ctx, query, showtimeID, seatIDs).Scan(&occupiedSeats)
	if err != nil {
		return nil, err
	}

	if len(occupiedSeats) == 0 {
		return nil, nil
	}

	return occupiedSeats, nil
}

func (s *Store) CheckShowtimeExists(ctx context.Context, showtimeID int) (bool, error) {
	pool := s.db.GetDB()

	var exists bool

	query := `
		SELECT EXISTS(SELECT 1 FROM showtimes WHERE id = $1);
	`
	err := pool.QueryRow(ctx, query, showtimeID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
