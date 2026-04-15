package redisclient

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrCacheNotFound = errors.New("data not found")

func (client *Redis) GetCache(key string, ctx context.Context) (string, error) {
	val, err := client.Client.Get(ctx, key).Result()

	if err == redis.Nil {
		return "", ErrCacheNotFound
	} else if err != nil {
		return "", err
	} else {
		return val, nil
	}
}

func (client *Redis) SetCache(key string, value any, ttl time.Duration, ctx context.Context) error {
	data, errMarshal := json.Marshal(value)
	if errMarshal != nil {
		return errMarshal
	}

	err := client.Client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return err
	}

	return nil
}

func (client *Redis) DeleteKey(key string, ctx context.Context) error {
	err := client.Client.Del(ctx, key)
	if err != nil {
		return err.Err()
	}

	return nil
}

func (client *Redis) SetCardinality(key string, ctx context.Context) (int64, error) {
	count, err := client.Client.SCard(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (client *Redis) SetAdd(key string, ctx context.Context, members ...any) error {
	err := client.Client.SAdd(ctx, key, members...).Err()
	if err != nil {
		return err
	}

	return nil
}

func (client *Redis) SetCacheNX(key string, value any, ttl time.Duration, ctx context.Context) (bool, error) {
	data, errMarshal := json.Marshal(value)
	if errMarshal != nil {
		return false, errMarshal
	}

	err := client.Client.SetArgs(ctx, key, data, redis.SetArgs{
		Mode: "NX",
		TTL:  ttl,
	}).Err()

	if err == redis.Nil {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (client *Redis) Expire(key string, ttl time.Duration, ctx context.Context) error {
	err := client.Client.Expire(ctx, key, ttl).Err()
	if err != nil {
		return err
	}

	return nil
}

func (client *Redis) SetMembers(key string, ctx context.Context) ([]string, error) {
	members, err := client.Client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	return members, nil
}
