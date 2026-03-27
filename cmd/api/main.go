package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"start/internal/database"
	"start/internal/movie"
	"start/internal/rabbitmq"
	redisClient "start/internal/redis"

	"start/internal/seat"
	"start/internal/showtime"
	"start/internal/ticket"

	"github.com/google/uuid"
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

	redisUser := os.Getenv("REDIS_USER")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisDb := os.Getenv("REDIS_DB")

	redisUrl := fmt.Sprintf("redis://%s:%s@%s:%s/%s",
		redisUser,
		redisPassword,
		redisHost,
		redisPort,
		redisDb,
	)

	rdb, err := redisClient.Connect(redisUrl)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	movieStore := movie.New(service)
	movieHandler := movie.NewHandler(movieStore)

	seatStore := seat.New(service)
	seatHandler := seat.NewHandler(seatStore, rdb)

	showtimeStore := showtime.New(service)
	showtimeHandler := showtime.NewHandler(showtimeStore, rdb)

	ticketStore := ticket.New(service)
	ticketHandler := ticket.NewHandler(ticketStore, queue, rdb)

	cfg := &Config{
		movieHanlder:    movieHandler,
		seatHandler:     seatHandler,
		showtimeHanlder: showtimeHandler,
		ticketHandler:   ticketHandler,
	}

	hanlder := routes(cfg)

	go queue.ConsumeTicket(ticketStore.CreateTicket, rdb)

	go rdb.ListenForTicketExpirations(func(expiredKey string) {
		parts := strings.Split(expiredKey, ":")
		if len(parts) != 4 {
			fmt.Println("Warning: Unknown key format expired:", expiredKey)
			return
		}

		showtimeIdStr := parts[2]
		ticketIdStr := parts[3]

		ticketUUID, err := uuid.Parse(ticketIdStr)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		errDeleteTicket := ticketStore.DeleteTicket(ticketUUID)
		if errDeleteTicket != nil {
			fmt.Println("Error:", errDeleteTicket)
			return
		}
		fmt.Println("Ticket deleted for timeout:", ticketIdStr)

		idShowtime, err := strconv.Atoi(showtimeIdStr)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		cacheShowtimeKey := rdb.BuildShowtimeSeatsKey(idShowtime)

		errDeleteKey := rdb.DeleteKey(cacheShowtimeKey)
		if errDeleteKey != nil {
			fmt.Printf("WARNING: Failed to invalidate cache for %s: %v\n", cacheShowtimeKey, errDeleteKey)
		}
	})

	fmt.Println("Server is up and running!!!")
	http.ListenAndServe(":8080", hanlder)
}
