package events

import (
	"log/slog"
	"net/http"
	"strconv"
)

type Handler struct {
	broker *ShowtimeBroker
}

func NewHandler(b *ShowtimeBroker) *Handler {
	return &Handler{broker: b}
}

// StreamShowtimeSeats godoc
// @Summary      Stream real-time seat updates
// @Description  Establishes a Server-Sent Events (SSE) connection to broadcast seat holds and purchases.
// @Description  The stream remains open indefinitely. Each event contains a JSON payload in the `data` field representing the seat status.
// @Tags         Events
// @Produce      text/event-stream
// @Param        id   path      int  true  "Showtime ID"
// @Success      200  {object}  events.SeatEventPayload  "The JSON payload sent inside the SSE 'data' block"
// @Router       /showtimes/event/{id} [get]
func (h *Handler) StreamShowtimeSeats(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	showtimeID, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error("Failed to cast value", "error", err, "path", r.URL.Path)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	rc := http.NewResponseController(w)

	_ = rc.Flush()

	clientChan := h.broker.AddClient(showtimeID)

	defer h.broker.RemoveClient(showtimeID, clientChan)

	for {
		select {
		case <-r.Context().Done():
			slog.Info("Client disconnected from showtime", "showtime_id", showtimeID)
			return

		case msg := <-clientChan:
			slog.Info("Message received", "msg",  string(msg))
			_, err := w.Write(msg)
			if err != nil {
				slog.Error("Failed to write SSE data", "error", err)
				return
			}

			err = rc.Flush()
			if err != nil {
				slog.Error("Failed to flush SSE data", "error", err)
				return
			}
		}
	}
}
