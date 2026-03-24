package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"start/internal/database"
	"start/internal/movie"
	"start/internal/rabbitmq"
	"start/internal/seat"
	"start/internal/showtime"
	"start/internal/ticket"
)

func main() {
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("DATABASE_HOST")
	port := os.Getenv("POSTGRES_PORT")
	db := os.Getenv("POSTGRES_DB")

	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		user, pass, host, port, db,
	)
	service, err := database.NewConnection(context.Background(), url)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	rabbitUser := os.Getenv("RABBITMQ_USER")
	rabbitPass := os.Getenv("RABBITMQ_PASSWORD")
	rabbitHost := os.Getenv("RABBITMQ_HOST")
	rabbitPort := os.Getenv("AMQP_PORT")

	rabbitUrl := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		rabbitUser,
		rabbitPass,
		rabbitHost,
		rabbitPort,
	)

	queue, err := rabbitmq.Connect(rabbitUrl)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	movieStore := movie.New(service)
	movieHandler := movie.NewHandler(movieStore)

	seatStore := seat.New(service)
	seatHandler := seat.NewHandler(seatStore)

	showtimeStore := showtime.New(service)
	showtimeHandler := showtime.NewHandler(showtimeStore)

	ticketStore := ticket.New(service)
	ticketHandler := ticket.NewHandler(ticketStore, queue)

	cfg := &Config{
		movieHanlder:    movieHandler,
		seatHandler:     seatHandler,
		showtimeHanlder: showtimeHandler,
		ticketHandler:   ticketHandler,
	}

	hanlder := routes(cfg)

	go queue.ConsumeTicket(ticketStore.CreateTicket)

	fmt.Println("Server is up and running!!!")
	http.ListenAndServe(":8080", hanlder)
}
