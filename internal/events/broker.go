package events

import (
	"log/slog"
	"sync"
)

type ShowtimeBroker struct {
	clients map[int]map[chan []byte]struct{}
	mu      sync.RWMutex
}

func NewShowtimeBroker() *ShowtimeBroker {
	return &ShowtimeBroker{
		clients: make(map[int]map[chan []byte]struct{}),
	}
}

func (b *ShowtimeBroker) AddClient(showtimeID int) chan []byte {
	b.mu.Lock()
	defer b.mu.Unlock()

	clientChan := make(chan []byte, 10)
	if b.clients[showtimeID] == nil {
		b.clients[showtimeID] = make(map[chan []byte]struct{})
	}
	b.clients[showtimeID][clientChan] = struct{}{}

	return clientChan
}

func (b *ShowtimeBroker) RemoveClient(showtimeID int, clientChan chan []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if clientMap, exists := b.clients[showtimeID]; exists {
		delete(clientMap, clientChan)

		close(clientChan)

		if len(clientMap) == 0 {
			delete(b.clients, showtimeID)
		}
	}
}

func (b *ShowtimeBroker) Broadcast(showtimeID int, message []byte) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	slog.Info("Message received for broadcast", "size", len(message))

	clientMap, exists := b.clients[showtimeID]
	if !exists {
		slog.Warn("Broadcast aborted: No clients listening for this showtime",
					"target_showtime_id", showtimeID,
					"active_rooms_count", len(b.clients))
		return
	}

	slog.Info("Broadcasting to clients",
	    "showtime_id", showtimeID,
	    "client_count", len(clientMap),
	    "message", string(message),
	)

	for clientChan := range clientMap {
		select {
		case clientChan <- message:
			slog.Info("Successfully pushed message to a client channel")
		default:
			slog.Warn("Dropped message: client channel buffer is full")
		}
	}
}
