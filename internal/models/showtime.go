package models

import (
	"errors"
	"strings"
	"time"
)

type Showtime struct {
	ID           int       `db:"id" json:"id"`
	IDMovie      int       `db:"id_movie" json:"idMovie"`
	IDAuditorium int       `db:"id_auditorium" json:"idAuditorium"`
	StartTime    time.Time `db:"start_time" json:"startTime"`
}

func (sh *Showtime) Validate() error {
	var errs []string

	if sh.IDMovie <= 0 {
		errs = append(errs, "the movie is required")
	}
	if sh.IDAuditorium <= 0 {
		errs = append(errs, "the auditorium Id is required")
	}

	if sh.StartTime.IsZero() {
		errs = append(errs, "the time must not be empty")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}
