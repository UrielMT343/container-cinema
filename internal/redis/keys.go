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

func (client *Redis) BuildCartKey(cartIDstr string) string {
	return fmt.Sprintf("cart_%s", cartIDstr)
}

func (client *Redis) BuildSeatsCheckKey(idShowtime int, idSeat int) string {
	return fmt.Sprintf("showtime:%d:seat:%d", idShowtime, idSeat)
}

func (client *Redis) BuildCartLimitKey(cartIDstr string) string {
	return fmt.Sprintf("cart:%s:count", cartIDstr)
}
