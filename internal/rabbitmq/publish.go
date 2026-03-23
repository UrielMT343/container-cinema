package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (q *RabbitMQ) PublishTicket(ch *amqp.Channel, ticketData TicketData) error {
	qd, err := ch.QueueDeclare("ticket", true, false, false, false, amqp.Table{amqp.QueueTypeArg: amqp.QueueTypeQuorum})
	if err != nil {
		return fmt.Errorf("Error while declaring the queue: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, err := json.Marshal(ticketData)
	if err != nil {
		return fmt.Errorf("Error while encoding data: %v", err)
	}

	err = ch.PublishWithContext(ctx, "", qd.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	fmt.Println("Message published")
	return nil
}
