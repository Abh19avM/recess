package redis

import (
	"context"
	"testing"
)

func TestRedisNilPing(t *testing.T) {
	var c *Client
	_, err := c.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error when pinging nil redis client, got nil")
	}

	c = &Client{RDB: nil}
	_, err = c.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error when pinging redis with nil client, got nil")
	}

	// Calling Close on nil client should not panic
	if err := c.Close(); err != nil {
		t.Errorf("unexpected error on nil client close: %v", err)
	}
}

func TestRedisInvalidURL(t *testing.T) {
	ctx := context.Background()
	_, err := New(ctx, "invalid-redis-url-:::://")
	if err == nil {
		t.Fatal("expected error when initializing with invalid redis URL")
	}
}
