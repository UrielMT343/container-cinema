package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"start/internal/models"
	redisclient "start/internal/redis"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (q *RabbitMQ) ConsumeTicket(insertTicketToDB func(ticket models.Ticket) (models.Ticket, error), rdb *redisclient.Redis) {
	ch, err := q.NewChannel()
	if err != nil {
		return
	}

	qd, err := ch.QueueDeclare("ticket", true, false, false, false, amqp.Table{amqp.QueueTypeArg: amqp.QueueTypeQuorum})
	if err != nil {
		return
	}

	msgs, err := ch.Consume(qd.Name, "", false, false, false, false, nil)
	if err != nil {
		return
	}

	listening := make(chan struct{})

	go func() {
		for m := range msgs {
			slog.Info("Message received", "body", m.Body)

			var ticket models.Ticket

			err := json.Unmarshal(m.Body, &ticket)
			if err != nil {
				slog.Error("Error while formating the JSON", "error", err)
				errNack := m.Nack(false, false)
				if errNack != nil {
					continue
				}

				continue
			}

			createdTicket, err := insertTicketToDB(ticket)
			if err != nil {
				slog.Error("Error inserting ticket to DB:", "error", err)
				errNack := m.Nack(false, true)
				if errNack != nil {
					continue
				}

				continue
			}

			showtimeKey := fmt.Sprintf("seats:showtime:%v", ticket.IDShowtime)

			errDelete := rdb.DeleteKey(showtimeKey)
			if errDelete != nil {
				slog.Warn("Failed to clear cache", "key", showtimeKey)
			}

			slog.Info("Ticket successfully processed", "ticket", createdTicket.ID)
			errAck := m.Ack(false)
			if errAck != nil {
				continue
			}
		}
	}()

	slog.Info("Waiting for messages")
	<-listening
}
