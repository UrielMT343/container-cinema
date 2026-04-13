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

	err := client.Client.Set(context.Background(), key, data, ttl).Err()
	if err != nil {
		return err
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

func (client *Redis) SetCardinality(key string) (int64, error) {
	count, err := client.Client.SCard(context.Background(), key).Result()
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (client *Redis) SetAdd(key string, members ...any) error {
	err := client.Client.SAdd(context.Background(), key, members...).Err()
	if err != nil {
		return err
	}

	return nil
}

func (client *Redis) SetCacheNX(key string, value any, ttl time.Duration) (bool, error) {
	data, errMarshal := json.Marshal(value)
	if errMarshal != nil {
		return false, errMarshal
	}

	err := client.Client.SetArgs(context.Background(), key, data, redis.SetArgs{
		Mode: "NX",
		TTL: ttl,
	}).Err()

	if err == redis.Nil {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (client *Redis) Expire(key string, ttl time.Duration) error {
	err := client.Client.Expire(context.Background(), key, ttl).Err()
	if err != nil {
		return err
	}

	return nil
}

func (client *Redis) SetMembers(key string) ([]string, error) {
	members, err := client.Client.SMembers(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}

	return members, nil
}
