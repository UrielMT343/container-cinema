package models

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

type Ticket struct {
	ID         uuid.UUID `db:"id" json:"id"`
	IDUser     *int      `db:"id_user" json:"idUser"`
	IDShowtime int       `db:"id_showtime" json:"idShowtime"`
	Status     string    `db:"status" json:"status"`
	IDSeat     int       `db:"is_seat" json:"idSeat"`
	Email      *string   `json:"email,omitempty" db:"email"`
}

var ErrorTicketNotHeld = errors.New("ticket not held")

func (t *Ticket) Validate() error {
	var errs []string

	if t.ID == uuid.Nil {
		errs = append(errs, "the ticket ID is required")
	}

	if t.IDShowtime <= 0 {
		errs = append(errs, "the showtime is required")
	}

	if t.Status == "" {
		errs = append(errs, "the status must not be empty")
	}

	if t.IDSeat <= 0 {
		errs = append(errs, "the seat is required")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}
