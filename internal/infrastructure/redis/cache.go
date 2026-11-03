package rediscache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func Connect(ctx context.Context, addr string) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	var last error
	for attempt := 0; attempt < 30; attempt++ {
		last = rdb.Ping(ctx).Err()
		if last == nil {
			return &Client{rdb: rdb}, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return nil, last
}

func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

func (c *Client) Get(ctx context.Context, key string) ([]byte, bool, error) {
	value, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return value, true, nil
}

func (c *Client) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *Client) DeletePrefix(ctx context.Context, prefix string) error {
	var cursor uint64
	for {
		keys, next, err := c.rdb.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			return nil
		}
	}
}

func (c *Client) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	count, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		return true, err
	}
	if count == 1 {
		_ = c.rdb.Expire(ctx, key, window).Err()
	}
	return count <= limit, nil
}
