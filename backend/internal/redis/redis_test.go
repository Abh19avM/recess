package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/Abh19avM/recess/internal/redis"
)

// 1. Test Key Formatters and Canonical Conventions
func TestKeyFormatters(t *testing.T) {
	if got := redis.PresenceKey("usr_123"); got != "online:user:usr_123" {
		t.Errorf("expected online:user:usr_123, got %s", got)
	}
	if got := redis.GameKey("game_456"); got != "game:game_456" {
		t.Errorf("expected game:game_456, got %s", got)
	}
	if got := redis.RoomKey("room_789"); got != "room:room_789" {
		t.Errorf("expected room:room_789, got %s", got)
	}
	if got := redis.QueueKey("hand_cricket"); got != "queue:hand_cricket" {
		t.Errorf("expected queue:hand_cricket, got %s", got)
	}
	if got := redis.RateUserKey("usr_123"); got != "rate:user:usr_123" {
		t.Errorf("expected rate:user:usr_123, got %s", got)
	}
	if got := redis.RateIPKey("192.168.1.1"); got != "rate:ip:192.168.1.1" {
		t.Errorf("expected rate:ip:192.168.1.1, got %s", got)
	}
	if got := redis.LeaderboardKey("xo"); got != "leaderboard:xo" {
		t.Errorf("expected leaderboard:xo, got %s", got)
	}
	if got := redis.RoomChannel("room_789"); got != "pubsub:room:room_789" {
		t.Errorf("expected pubsub:room:room_789, got %s", got)
	}
}

// 2. Test Nil-Safety / Degraded Mode for all Redis Stores
func TestNilRedisStoresDegradedMode(t *testing.T) {
	ctx := context.Background()

	// Presence Store
	ps := redis.NewPresenceStore(nil)
	if err := ps.SetOnline(ctx, redis.UserPresence{UserID: "u1"}); err != nil {
		t.Errorf("expected nil error on nil presence store, got %v", err)
	}
	if ok, err := ps.IsOnline(ctx, "u1"); ok || err != nil {
		t.Errorf("expected false, nil for IsOnline on nil store")
	}

	// Game Cache Store
	gs := redis.NewGameCacheStore(nil)
	if err := gs.SaveGameState(ctx, "g1", map[string]string{"foo": "bar"}); err != nil {
		t.Errorf("expected nil error on nil game cache store, got %v", err)
	}
	var dest map[string]string
	if found, err := gs.GetGameState(ctx, "g1", &dest); found || err != nil {
		t.Errorf("expected false, nil on nil game cache store")
	}

	// Leaderboard Store
	ls := redis.NewLeaderboardStore(nil)
	if err := ls.UpdateScore(ctx, "xo", "u1", 100); err != nil {
		t.Errorf("expected nil error on nil leaderboard store, got %v", err)
	}
	top, err := ls.GetTop(ctx, "xo", 10)
	if err != nil || len(top) != 0 {
		t.Errorf("expected empty list on nil leaderboard store")
	}

	// Rate Limiter
	rl := redis.NewRateLimiter(nil)
	allowed, remaining, err := rl.Allow(ctx, "rate:test", 10, time.Minute)
	if !allowed || remaining != 10 || err != nil {
		t.Errorf("expected fail-open behavior on nil rate limiter")
	}

	// PubSub Broker
	pb := redis.NewPubSubBroker(nil)
	if err := pb.Publish(ctx, "test_channel", "msg"); err != nil {
		t.Errorf("expected nil error on nil pubsub publish")
	}
}

// 3. Test Redis Client Connection & Health Check Ping
func TestRedisClient_NilPing(t *testing.T) {
	var c *redis.Client
	latency, err := c.Ping(context.Background())
	if err == nil {
		t.Errorf("expected error for nil client ping, got latency: %v", latency)
	}
}

func TestRedisClient_InvalidURL(t *testing.T) {
	_, err := redis.New(context.Background(), "invalid://url")
	if err == nil {
		t.Errorf("expected error for invalid redis url")
	}
}
