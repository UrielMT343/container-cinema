package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	queue *amqp.Connection
}

func Connect(connString string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(connString)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to RabbitMQ")
	}
	return &RabbitMQ{queue: conn}, nil
}

func (q *RabbitMQ) NewChannel() (*amqp.Channel, error) {
	ch, err := q.queue.Channel()
	if err != nil {
		return nil, fmt.Errorf("Error while creating the new channel: %v", err)
	}

	return ch, nil
}
