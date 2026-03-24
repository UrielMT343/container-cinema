package rabbitmq

import (
	"encoding/json"
	"fmt"
	"start/internal/models"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (q *RabbitMQ) ConsumeTicket(insertTicketToDB func(ticket models.Ticket) (models.Ticket, error)) {
	ch, err := q.NewChannel()
	if err != nil {
		return
	}

	qd, err := ch.QueueDeclare("ticket", true, false, false, false, amqp.Table{amqp.QueueTypeArg: amqp.QueueTypeQuorum})
	if err != nil {
		return
	}

	msgs, err := ch.Consume(qd.Name, "", false, false, false, false, nil)

	listening := make(chan struct{})

	go func() {
		for m := range msgs {
			fmt.Println("Message received: ", string(m.Body))

			var ticket models.Ticket

			err := json.Unmarshal(m.Body, &ticket)
			if err != nil {
				fmt.Printf("Error while formating the JSON: %s\n", err.Error())
				m.Nack(false, false)
				continue
			}

			createdTicket, err := insertTicketToDB(ticket)

			if err != nil {
				fmt.Printf("Error inserting ticket to DB: %s\n", err.Error())
				m.Nack(false, true)
				continue
			}

			fmt.Println("Ticket successfully processed!", createdTicket.Id)
			m.Ack(false)
		}
	}()

	fmt.Println("Waiting for messages. Exit by pressing CTRL + C")
	<-listening
}
