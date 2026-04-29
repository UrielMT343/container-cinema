package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
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
)

// @title           Cloud Cinema API
// @version         1.0
// @description     This is the distributed backend for the Cloud Cinema ticket booking system.
// @host            localhost
// @BasePath        /api/v1
// @schemes         https
func main() {
	isHealthCheck := flag.Bool("healthcheck", false, "Run internal container healthcheck")
	flag.Parse()

	if *isHealthCheck {
		apiPort := os.Getenv("API_PORT")
		if apiPort == "" {
			apiPort = "8080"
		}

		url := fmt.Sprintf("http://localhost:%s/health", apiPort)
		client := http.Client{Timeout: 2 * time.Second}

		resp, err := client.Get(url)
		if err != nil || resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}

		os.Exit(0)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting Cloud Cinema API", "version", "1.0.0", "env", "testing")

	rootCtx, stopCtx := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopCtx()

	pgUser := os.Getenv("POSTGRES_USER")
	pgPass := os.Getenv("POSTGRES_PASSWORD")
	pgHost := os.Getenv("DATABASE_HOST")
	pgPort := os.Getenv("POSTGRES_PORT")
	pgDB := os.Getenv("POSTGRES_DB")

	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		pgUser, pgPass, pgHost, pgPort, pgDB,
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

	errSetup := queue.SetupHoldTopology()
	if errSetup != nil {
		slog.Error("Critical startup error", "error", errSetup)
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

	go queue.CleanupDLXTickets(rootCtx, ticketStore.DeleteTicket, rdb)

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
