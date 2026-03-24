package models

import "time"

type Showtime struct {
	Id           int       `db:"id" json:"id"`
	IdMovie      int       `db:"id_movie" json:"idMovie"`
	IdAuditorium int       `db:"id_auditorium" json:"idAuditorium"`
	StartTime    time.Time `db:"start_time" json:"startTime"`
}
