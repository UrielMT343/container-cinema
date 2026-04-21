package rabbitmq

import (
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	queue      *amqp.Connection
	pubChannel *amqp.Channel
	sync.Mutex
}

func Connect(connString string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ")
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("error while creating the new pub channel: %v", err)
	}

	return &RabbitMQ{queue: conn, pubChannel: ch}, nil
}

func (q *RabbitMQ) NewChannel() (*amqp.Channel, error) {
	ch, err := q.queue.Channel()
	if err != nil {
		return nil, fmt.Errorf("error while creating the new pub channel: %v", err)
	}

	return ch, nil
}

func (q *RabbitMQ) Close() error {
	err := q.queue.Close()
	if err != nil {
		return err
	}
	return nil
}

func (q *RabbitMQ) SetupHoldTopology() (err error) {
	ch, err := q.queue.Channel()
	if err != nil {
		return fmt.Errorf("error opening channel: %w", err)
	}
	defer func() {
		closeErr := ch.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	err = ch.ExchangeDeclare(
		"ticket.hold.exchange",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error declaring hold_dlx: %w", err)
	}

	err = ch.ExchangeDeclare(
		"ticket.hold.dlx.exchange",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error declaring hold_dlx: %w", err)
	}

	_, err = ch.QueueDeclare(
		"ticket.hold.queue",
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    "ticket.hold.dlx.exchange",
			"x-dead-letter-routing-key": "ticket.hold.expired",
		},
	)
	if err != nil {
		return fmt.Errorf("error declaring hold_delay_queue: %w", err)
	}

	_, err = ch.QueueDeclare(
		"ticket.hold.cleanup.queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error declaring hold_cleanup_queue: %w", err)
	}

	err = ch.QueueBind(
		"ticket.hold.queue",
		"ticket.hold.created",
		"ticket.hold.exchange",
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error binding hold_delay_queue: %w", err)
	}

	err = ch.QueueBind(
		"ticket.hold.cleanup.queue",
		"ticket.hold.expired",
		"ticket.hold.dlx.exchange",
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error binding ticket.hold.cleanup.queue: %w", err)
	}

	return nil
}
