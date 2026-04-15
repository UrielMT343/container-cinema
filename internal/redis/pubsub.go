package redisclient

import (
	"context"
	"strings"
)

func (client *Redis) ListenForTicketExpirations(ctx context.Context, onExpire func(key string)) {
	pubsub := client.Client.Subscribe(context.Background(), "__keyevent@0__:expired")

	ch := pubsub.Channel()

	for msg := range ch {
		keyName := msg.Payload

		if !strings.HasPrefix(keyName, "hold:ticket:") {
			continue
		}

		onExpire(msg.Payload)
	}
}
