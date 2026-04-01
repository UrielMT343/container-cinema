package redisclient

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrCacheNotFound = errors.New("data not found")

func (client *Redis) GetCache(key string) (string, error) {
	val, err := client.Client.Get(context.Background(), key).Result()

	if err == redis.Nil {
		return "", ErrCacheNotFound
	} else if err != nil {
		return "", err
	} else {
		return val, nil
	}
}

func (client *Redis) SetCache(key string, value any, ttl time.Duration) error {
	data, errMarshal := json.Marshal(value)
	if errMarshal != nil {
		return errMarshal
	}

	err := client.Client.Set(context.Background(), key, data, ttl)
	if err != nil {
		return err.Err()
	}

	return nil
}

func (client *Redis) DeleteKey(key string) error {
	err := client.Client.Del(context.Background(), key)
	if err != nil {
		return err.Err()
	}

	return nil
}
