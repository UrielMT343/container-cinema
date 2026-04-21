package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"start/internal/config"
	"start/internal/models"
	redisclient "start/internal/redis"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

func (q *RabbitMQ) CleanupDLXTickets(ctx context.Context, deleteTicketsFromDB func(ctx context.Context, ticketID uuid.UUID) error, rdb *redisclient.Redis) {
	ch, err := q.NewChannel()
	if err != nil {
		return
	}

	defer func() {
		if err := ch.Close(); err != nil {
			slog.Error("Error closing the channel", "error", err)
		}
	}()

	workerCount := config.WorkerCount

	errQos := ch.Qos(workerCount, 0, false)
	if errQos != nil {
		slog.Error("Failed to set channel prefetch", "error", errQos)
		return
	}

	msgs, errConsume := ch.ConsumeWithContext(ctx, "ticket.hold.cleanup.queue", "cleanup-consumer",
		false, false, false, false, nil)
	if errConsume != nil {
		return
	}

	eg, groupCtx := errgroup.WithContext(ctx)

	for i := 0; i < workerCount; i++ {
		eg.Go(func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("Worker paccked!", "panic", r)
					err = fmt.Errorf("worker panic: %v", r)
				}
			}()

			for m := range msgs {
				var tickets []models.Ticket

				err := json.Unmarshal(m.Body, &tickets)
				if err != nil {
					slog.Error("Error while formating the JSON", "error", err)
					_ = m.Nack(false, false)
					continue
				}

				if len(tickets) == 0 {
					slog.Warn("Received empty ticket array")
					_ = m.Ack(false)
					continue
				}

				var fatalError bool

				for _, ticket := range tickets {
					ctxDelete, cancel := context.WithTimeout(groupCtx, 5*time.Second)
					err = deleteTicketsFromDB(ctxDelete, ticket.ID)
					cancel()
					if err != nil {
						if errors.Is(err, models.ErrorTicketNotHeld) {
							slog.Info("Ticket paid or already freed, skipping", "id", ticket.ID)
							continue
						} else {
							slog.Error("Error deleting ticket from DB", "error", err)
							_ = m.Nack(false, true)
							fatalError = true
							break
						}
					}
				}

				if fatalError {
					return err
				}

				slog.Info("Ticket deleted by timeout", "count", len(tickets))
				errAck := m.Ack(false)
				if errAck != nil {
					slog.Error("Failed to Ack message", "error", errAck)
					return err
				}

				cacheShowtimeKey := rdb.BuildShowtimeSeatsKey(tickets[0].IDShowtime)

				errDeleteKey := rdb.DeleteKey(cacheShowtimeKey, ctx)
				if errDeleteKey != nil {
					slog.Warn("Failed to invalidate cache", "showtimeKey", cacheShowtimeKey, "error", errDeleteKey)
				}
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		slog.Error("Worker pool shut down with error", "error", err)
	}
}
