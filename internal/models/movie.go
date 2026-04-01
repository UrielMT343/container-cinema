package models

import (
	"errors"
	"strings"
)

type Movie struct {
	ID          int    `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	DurationMin int    `db:"duration_min" json:"durationMin"`
	Synopsis    string `db:"synopsis" json:"synopsis"`
}

func (m *Movie) Validate() error {
	var errs []string

	if m.Name == "" {
		errs = append(errs, "the name must not be empty")
	}
	if m.DurationMin <= 0 {
		errs = append(errs, "the duration must be more than 0 minutes")
	}
	if m.Synopsis == "" {
		errs = append(errs, "the synopsis must not be empty")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}
