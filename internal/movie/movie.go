package movie

import "errors"

type Movie struct {
	Id          int    `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	DurationMin int    `db:"duration_min" json:"durationMin"`
	Synopsis    string `db:"synopsis" json:"synopsis"`
}

func (m *Movie) Validate() error {
	if m.Name == "" {
		return errors.New("The name must not be empty")
	}
	if m.DurationMin <= 0 {
		return errors.New("The duration must be more than 0 minutes")
	}
	if m.Synopsis == "" {
		return errors.New("The synopsis must not be empty")
	}
	return nil
}
