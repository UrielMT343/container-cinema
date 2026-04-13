package rabbitmq

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"start/internal/models"
	redisclient "start/internal/redis"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (q *RabbitMQ) ConsumeTicket(insertTicketsToDB func(ctx context.Context, tickets []models.Ticket) ([]models.Ticket, error), rdb *redisclient.Redis) {
	ch, err := q.NewChannel()
	if err != nil {
		return
	}

	defer func() {
		if err := ch.Close(); err != nil {
			slog.Error("Error closing the channel", "error", err)
		}
	}()

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
			slog.Info("Message received", "body_len", len(m.Body))

			var tickets []models.Ticket

			err := json.Unmarshal(m.Body, &tickets)
			if err != nil {
				slog.Error("Error while formating the JSON", "error", err)
				errNack := m.Nack(false, false)
				if errNack != nil {
					continue
				}

				continue
			}

			if len(tickets) == 0 {
				slog.Warn("Received empty ticket array")
				errAck := m.Ack(false)
				if errAck != nil {
					continue
				}
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, err = insertTicketsToDB(ctx, tickets)
			cancel()
			if err != nil {
				slog.Error("Error inserting tickets to DB:", "error", err)
				errNack := m.Nack(false, true)
				if errNack != nil {
					continue
				}

				continue
			}

			showtimeKey := rdb.BuildShowtimeSeatsKey(tickets[0].IDShowtime)

			errDelete := rdb.DeleteKey(showtimeKey)
			if errDelete != nil {
				slog.Warn("Failed to clear cache", "key", showtimeKey)
			}

			slog.Info("Bulk tickets successfully processed", "count", len(tickets))
			errAck := m.Ack(false)
			if errAck != nil {
				continue
			}
		}
	}()

	slog.Info("Waiting for messages")
	<-listening
}
