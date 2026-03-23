package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (q *RabbitMQ) ConsumeTicket() {
	ch, err := q.NewChannel()
	if err != nil {
		return
	}

	qd, err := ch.QueueDeclare("ticket", true, false, false, false, amqp.Table{amqp.QueueTypeArg: amqp.QueueTypeQuorum})
	if err != nil {
		return
	}

	msgs, err := ch.Consume(qd.Name, "", true, false, false, false, nil)

	listening := make(chan struct{})

	go func() {
		for m := range msgs {
			fmt.Println("Message received: ", string(m.Body))
		}
	}()

	fmt.Println("Waiting for messages. Exit by pressing CTRL + C")
	<-listening
}
