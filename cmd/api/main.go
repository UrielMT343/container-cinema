package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"start/internal/database"
	"start/internal/movie"
	"start/internal/rabbitmq"
	redisclient "start/internal/redis"
	"start/internal/seat"
	"start/internal/showtime"
	"start/internal/ticket"

	"github.com/google/uuid"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting Cloud Cinema API", "version", "1.0.0", "env", "testing")

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
		slog.Error("Critical startup error", "error", err)
		os.Exit(1)
	}

	rabbitUser := os.Getenv("RABBITMQ_USER")
	rabbitPass := os.Getenv("RABBITMQ_PASSWORD")
	rabbitHost := os.Getenv("RABBITMQ_HOST")
	rabbitPort := os.Getenv("AMQP_PORT")

	rabbitURL := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		rabbitUser,
		rabbitPass,
		rabbitHost,
		rabbitPort,
	)

	queue, err := rabbitmq.Connect(rabbitURL)
	if err != nil {
		slog.Error("Critical startup error", "error", err)
		os.Exit(1)
	}

	redisUser := os.Getenv("REDIS_USER")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisDB := os.Getenv("REDIS_DB")

	redisURL := fmt.Sprintf("redis://%s:%s@%s:%s/%s",
		redisUser,
		redisPassword,
		redisHost,
		redisPort,
		redisDB,
	)

	rdb, err := redisclient.Connect(redisURL)
	if err != nil {
		slog.Error("Critical startup error", "error", err)
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
			slog.Warn("Unknown key format expired:", "expiredKey", expiredKey)
			return
		}

		showtimeIDStr := parts[2]
		ticketIDStr := parts[3]

		ticketUUID, err := uuid.Parse(ticketIDStr)
		if err != nil {
			slog.Error("Failed to parse the ticket", "ticket", ticketIDStr)
			return
		}

		errDeleteTicket := ticketStore.DeleteTicket(ticketUUID)
		if errDeleteTicket != nil {
			slog.Error("Failed to delete the ticket", "ticket", ticketUUID)
			return
		}
		slog.Info("Ticket deleted by timeout", "ticket", ticketIDStr)

		idShowtime, err := strconv.Atoi(showtimeIDStr)
		if err != nil {
			slog.Error("Failed to cast showtime ID to string", "showtimeID", showtimeIDStr)
			return
		}

		cacheShowtimeKey := rdb.BuildShowtimeSeatsKey(idShowtime)

		errDeleteKey := rdb.DeleteKey(cacheShowtimeKey)
		if errDeleteKey != nil {
			slog.Warn("Failed to invalidate cache", "showtimeKey", cacheShowtimeKey, "error", errDeleteKey)
		}
	})

	slog.Info("Server started", "Port", 8080)
	errServer := http.ListenAndServe(":8080", hanlder)
	if errServer != nil {
		slog.Error("Critical startup Error", "error", errServer)
		os.Exit(1)
	}
}
