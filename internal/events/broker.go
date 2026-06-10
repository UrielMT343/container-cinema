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

	clientMap, exists := b.clients[showtimeID]
	if !exists {
		return
	}

	for clientChan := range clientMap {
		select {
		case clientChan <- message:

		default:
			slog.Warn("Dropped message: client channel buffer is full")
		}
	}
}
