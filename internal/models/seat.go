package models

import (
	"errors"
	"strings"
)

type Seat struct {
	ID           int    `db:"id" json:"id"`
	Number       string `db:"number" json:"number"`
	IDAuditorium int    `db:"id_auditorium" json:"idAuditorium"`
}

func (s *Seat) Validate() error {
	var errs []string
	if s.Number == "" {
		errs = append(errs, "the seat number cannot be empty")
	}
	if s.IDAuditorium <= 0 {
		errs = append(errs, "the auditorium Id is required")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

type ShowtimeSeat struct {
	ID     int    `json:"id"`
	Number string `json:"number"`
	Status string `json:"status"`
}
