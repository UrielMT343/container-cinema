package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (q *RabbitMQ) PublishHoldTicket(body []byte, ttl time.Duration) (err error) {
	q.Lock()
	defer q.Unlock()

	ttlMs := strconv.FormatInt(ttl.Milliseconds(), 10)

	slog.Info("Ticket received on RabbitMQ", "size", len(body), "ttl ms", ttlMs)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = q.pubChannel.PublishWithContext(
		ctx,
		"ticket.hold.exchange",
		"ticket.hold.created",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			Expiration:   ttlMs,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("error publishing hold ticket: %w", err)
	}

	return nil
}
