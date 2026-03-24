package rabbitmq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (q *RabbitMQ) PublishTicket(body []byte) error {
	ch, err := q.queue.Channel()
	if err != nil {
		return fmt.Errorf("Error opening channel: %v", err)
	}

	defer ch.Close()

	qd, err := ch.QueueDeclare("ticket", true, false, false, false, amqp.Table{amqp.QueueTypeArg: amqp.QueueTypeQuorum})
	if err != nil {
		return fmt.Errorf("Error while declaring the queue: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx, "", qd.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	fmt.Println("Message published")
	return nil
}
