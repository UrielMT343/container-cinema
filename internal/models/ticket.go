package models

import "github.com/google/uuid"

type Ticket struct {
	Id         uuid.UUID `db:"id" json:"id"`
	IdUser     int       `db:"id_user" json:"idUser"`
	IdShowtime int       `db:"id_showtime" json:"idShowtime"`
	Status     string    `db:"status" json:"status"`
	IdSeat     int       `db:"is_seat" json:"idSeat"`
}
