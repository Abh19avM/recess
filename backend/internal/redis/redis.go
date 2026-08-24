package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps the go-redis client.
type Client struct {
	RDB *redis.Client
}

// New initializes a new Redis client from a connection URL with connection pooling.
func New(ctx context.Context, redisURL string) (*Client, error) {
	if redisURL == "" || redisURL == "in-memory" {
		return nil, fmt.Errorf("redis URL not configured (in-memory mode)")
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	opts.DialTimeout = 5 * time.Second
	opts.ReadTimeout = 3 * time.Second
	opts.WriteTimeout = 3 * time.Second
	opts.PoolSize = 20
	opts.MinIdleConns = 5

	rdb := redis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		slog.Warn("initial redis ping failed (will retry on readiness checks)", "error", err)
	} else {
		slog.Info("connected to Redis successfully")
	}

	return &Client{RDB: rdb}, nil
}

// Ping checks if Redis is reachable and returns latency.
func (c *Client) Ping(ctx context.Context) (time.Duration, error) {
	if c == nil || c.RDB == nil {
		return 0, fmt.Errorf("redis client not initialized")
	}
	start := time.Now()
	_, err := c.RDB.Ping(ctx).Result()
	latency := time.Since(start)
	return latency, err
}

// Close gracefully closes the Redis client connection.
func (c *Client) Close() error {
	if c != nil && c.RDB != nil {
		slog.Info("closing Redis client connection")
		return c.RDB.Close()
	}
	return nil
}
