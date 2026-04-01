package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (q *RabbitMQ) PublishTicket(body []byte) (err error) {
	ch, err := q.queue.Channel()
	if err != nil {
		return fmt.Errorf("error opening channel: %v", err)
	}

	defer func() {
		closeErr := ch.Close()
		if closeErr != nil {
			if err == nil {
				err = closeErr
			} else {
				err = fmt.Errorf("%v; close error: %v", err, closeErr)
			}
		}
	}()

	qd, err := ch.QueueDeclare("ticket", true, false, false, false, amqp.Table{amqp.QueueTypeArg: amqp.QueueTypeQuorum})
	if err != nil {
		return fmt.Errorf("error while declaring the queue: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx, "", qd.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		return fmt.Errorf("error publishing the ticket: %v", err)
	}

	slog.Info("Message published", "message", body)
	return nil
}
