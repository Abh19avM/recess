package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Abh19avM/recess/internal/metrics"
	"github.com/redis/go-redis/v9"
)

// PubSubBroker manages Redis Pub/Sub channels for cross-process room broadcasting.
type PubSubBroker struct {
	rdb *redis.Client
}

// NewPubSubBroker creates a new PubSubBroker instance.
func NewPubSubBroker(rdb *redis.Client) *PubSubBroker {
	return &PubSubBroker{rdb: rdb}
}

// Publish serializes and publishes an event envelope to a Redis channel.
func (b *PubSubBroker) Publish(ctx context.Context, channel string, message any) error {
	if b.rdb == nil {
		return nil
	}
	defer metrics.ObserveRedis("pubsub_publish", time.Now())

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal pubsub message: %w", err)
	}

	return b.rdb.Publish(ctx, channel, data).Err()
}

// Subscribe opens a subscription on the specified Redis channels.
func (b *PubSubBroker) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if b.rdb == nil {
		return nil
	}
	return b.rdb.Subscribe(ctx, channels...)
}

// ListenChannel starts a goroutine to read messages from a PubSub subscription.
func (b *PubSubBroker) ListenChannel(ctx context.Context, pubsub *redis.PubSub, handler func(channel string, payload []byte)) {
	if pubsub == nil {
		return
	}

	go func() {
		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				_ = pubsub.Close()
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				if msg != nil && handler != nil {
					handler(msg.Channel, []byte(msg.Payload))
				}
			}
		}
	}()
}
