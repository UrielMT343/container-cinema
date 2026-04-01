package redisclient

import (
	"fmt"

	"github.com/google/uuid"
)

func (client *Redis) BuildHoldTicketKey(idTicket uuid.UUID, idShowtime int) string {
	return fmt.Sprintf("hold:ticket:%d:%s", idShowtime, idTicket.String())
}

func (client *Redis) BuildShowtimeSeatsKey(idShowtime int) string {
	return fmt.Sprintf("seats:showtime:%d", idShowtime)
}
