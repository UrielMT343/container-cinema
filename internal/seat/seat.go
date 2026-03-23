package seat

import "errors"

type Seat struct {
	Id           int `db:"id" json:"id"`
	Number       int `db:"number" json:"number"`
	IdAuditorium int `db:"id_auditorium" json:"idAuditorium"`
}

func (s *Seat) Validate() error {
	if s.Number <= 0 {
		return errors.New("The seat number must be greater than zero")
	}
	if s.IdAuditorium <= 0 {
		return errors.New("The auditorium Id is required")
	}
	return nil
}
