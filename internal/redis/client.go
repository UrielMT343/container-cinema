package redisclient

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func Connect(connString string) (*Redis, error) {
	opt, err := redis.ParseURL(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis")
	}

	client := redis.NewClient(opt)

	pong, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("error: %v", err)
	}

	fmt.Println(pong)

	return &Redis{Client: client}, nil
}
