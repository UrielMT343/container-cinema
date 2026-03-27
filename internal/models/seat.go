package models

import "errors"

type Seat struct {
	Id           int    `db:"id" json:"id"`
	Number       string `db:"number" json:"number"`
	IdAuditorium int    `db:"id_auditorium" json:"idAuditorium"`
}

func (s *Seat) Validate() error {
	if s.Number == "" {
		return errors.New("The seat number cannot be empty")
	}
	if s.IdAuditorium <= 0 {
		return errors.New("The auditorium Id is required")
	}
	return nil
}

type ShowtimeSeat struct {
	ID     int    `json:"id"`
	Number string `json:"number"`
	Status string `json:"status"`
}
