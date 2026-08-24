package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter manages Redis-backed sliding window rate limits.
type RateLimiter struct {
	rdb *redis.Client
}

// NewRateLimiter creates a new RateLimiter.
func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{rdb: rdb}
}

// Allow checks if an action is permitted within the sliding window, and records the attempt.
func (l *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
	if l.rdb == nil {
		return true, limit, nil // Open when redis is unavailable in degraded mode
	}

	now := time.Now().UnixNano()
	windowStart := now - window.Nanoseconds()

	pipe := l.rdb.TxPipeline()
	// 1. Remove entries older than window
	pipe.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", windowStart))
	// 2. Add current timestamp
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d", now)})
	// 3. Count remaining entries in window
	countCmd := pipe.ZCard(ctx, key)
	// 4. Set TTL on the sorted set
	pipe.Expire(ctx, key, window*2)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return true, limit, err // Fail open on error
	}

	count := int(countCmd.Val())
	allowed := count <= limit
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	return allowed, remaining, nil
}
