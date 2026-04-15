package redisclient

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func Connect(connString string, ctx context.Context) (*Redis, error) {
	opt, err := redis.ParseURL(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis")
	}

	client := redis.NewClient(opt)

	errPing := client.Ping(ctx).Err()
	if errPing != nil {
		return nil, fmt.Errorf("error: %v", errPing)
	}

	return &Redis{Client: client}, nil
}

func (client *Redis) Close() error {
	err := client.Client.Close()
	if err != nil {
		return err
	}
	return nil
}
