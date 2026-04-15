package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"start/internal/database"
	"start/internal/movie"
	"start/internal/rabbitmq"
	redisclient "start/internal/redis"
	"start/internal/seat"
	"start/internal/showtime"
	"start/internal/ticket"
	"start/internal/user"

	"github.com/google/uuid"
)

// @title           Cloud Cinema API
// @version         2.0
// @description     This is the distributed backend for the Cloud Cinema ticket booking system.
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting Cloud Cinema API", "version", "1.0.0", "env", "testing")

	rootCtx, stopCtx := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopCtx()

	pgUser := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("DATABASE_HOST")
	port := os.Getenv("POSTGRES_PORT")
	db := os.Getenv("POSTGRES_DB")

	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		pgUser, pass, host, port, db,
	)
	service, err := database.NewConnection(rootCtx, url)
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

	rdb, err := redisclient.Connect(redisURL, rootCtx)
	if err != nil {
		slog.Error("Critical startup error", "error", err)
		os.Exit(1)
	}

	tokenSecret := os.Getenv("JWT_SECRET")

	movieStore := movie.New(service)
	movieHandler := movie.NewHandler(movieStore)

	seatStore := seat.New(service)
	seatHandler := seat.NewHandler(seatStore, rdb)

	showtimeStore := showtime.New(service)
	showtimeHandler := showtime.NewHandler(showtimeStore, rdb)

	ticketStore := ticket.New(service)
	ticketHandler := ticket.NewHandler(ticketStore, queue, rdb)

	userStore := user.New(service)
	userHandler := user.NewHandler(userStore, tokenSecret)

	cfg := &Config{
		movieHanlder:    movieHandler,
		seatHandler:     seatHandler,
		showtimeHanlder: showtimeHandler,
		ticketHandler:   ticketHandler,
		userHandler:     userHandler,
	}

	apiVersion := os.Getenv("API_VERSION")
	basePrefix := fmt.Sprintf("/api/%s", apiVersion)

	handler := routes(cfg, basePrefix, tokenSecret)

	go queue.ConsumeTicket(rootCtx, ticketStore.CreateTicket, rdb)

	go rdb.ListenForTicketExpirations(rootCtx, func(expiredKey string) {
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

		errDeleteTicket := ticketStore.DeleteTicket(rootCtx, ticketUUID)
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

		errDeleteKey := rdb.DeleteKey(cacheShowtimeKey, rootCtx)
		if errDeleteKey != nil {
			slog.Warn("Failed to invalidate cache", "showtimeKey", cacheShowtimeKey, "error", errDeleteKey)
		}
	})

	apiPort := os.Getenv("API_PORT")

	apiAddr := fmt.Sprintf(":%s", apiPort)

	srv := &http.Server{
		Addr:    apiAddr,
		Handler: handler,
	}

	slog.Info("Server started", "Port", apiPort)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Critical server error", "error", err)
		}
	}()
	<-rootCtx.Done()
	slog.Info("Shutdown signal received, initiating graceful teardown...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	service.CloseConnection()
	errRabbitClose := queue.Close()
	if errRabbitClose != nil {
		slog.Error("Error closing Rabbit service", "error", errRabbitClose)
	}

	errRedisClose := rdb.Close()
	if errRedisClose != nil {
		slog.Error("Error closing Redis service", "error", errRedisClose)
	}

	slog.Info("Cloud Cinema API cleanly stopped")
}
