package cache

import (
	"context"
	"log"
	"time"
)

// PubSubMessageHandler is called for each message received on a subscribed channel.
type PubSubMessageHandler func(channel, payload string)

// StartPubSubListener subscribes to patterns and calls handler for each message.
// It reconnects automatically with exponential backoff (max 30s).
// The returned cancel function stops the listener.
func StartPubSubListener(ctx context.Context, patterns []string, handler PubSubMessageHandler) context.CancelFunc {
	listenerCtx, cancel := context.WithCancel(ctx)

	go func() {
		backoff := time.Second
		maxBackoff := 30 * time.Second

		for {
			if !GetManager().IsAvailable() {
				select {
				case <-listenerCtx.Done():
					return
				case <-time.After(backoff):
					backoff = min(backoff*2, maxBackoff)
					continue
				}
			}

			err := runPubSubLoop(listenerCtx, patterns, handler)

			select {
			case <-listenerCtx.Done():
				return
			default:
			}

			if err != nil {
				log.Printf("⚠️  Pub/Sub loop exited (%v), reconnecting in %v", err, backoff)
			}

			select {
			case <-listenerCtx.Done():
				return
			case <-time.After(backoff):
				backoff = minDuration(backoff*2, maxBackoff)
			}
		}
	}()

	return cancel
}

func runPubSubLoop(ctx context.Context, patterns []string, handler PubSubMessageHandler) error {
	ps := GetManager().pSubscribe(ctx, patterns...)
	defer ps.Close()

	// Reset backoff after successful connection
	log.Printf("✅ Pub/Sub listener connected, patterns: %v", patterns)

	ch := ps.Channel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-ch:
			if !ok {
				return nil // channel closed, reconnect
			}
			handler(msg.Channel, msg.Payload)
		}
	}
}

// Publish sends a message to a Redis Pub/Sub channel.
// Returns nil if Redis is unavailable (non-blocking, best-effort).
func Publish(ctx context.Context, channel string, payload interface{}) error {
	if !GetManager().IsAvailable() {
		return nil // Silently skip; local broadcast already happened
	}
	return GetManager().publish(ctx, channel, payload)
}

// PublishAsync publishes without blocking the caller.
func PublishAsync(channel string, payload interface{}) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := Publish(ctx, channel, payload); err != nil {
			log.Printf("⚠️  PUBLISH %s failed: %v", channel, err)
		}
	}()
}

// SubscribeChannel subscribes to exact channels (not patterns).
func SubscribeChannel(ctx context.Context, channels []string, handler PubSubMessageHandler) context.CancelFunc {
	listenerCtx, cancel := context.WithCancel(ctx)

	go func() {
		backoff := time.Second
		maxBackoff := 30 * time.Second

		for {
			if !GetManager().IsAvailable() {
				select {
				case <-listenerCtx.Done():
					return
				case <-time.After(backoff):
					backoff = min(backoff*2, maxBackoff)
					continue
				}
			}

			err := runSubscribeLoop(listenerCtx, channels, handler)

			select {
			case <-listenerCtx.Done():
				return
			default:
			}

			if err != nil {
				log.Printf("⚠️  Subscribe loop exited (%v), reconnecting in %v", err, backoff)
			}

			select {
			case <-listenerCtx.Done():
				return
			case <-time.After(backoff):
				backoff = minDuration(backoff*2, maxBackoff)
			}
		}
	}()

	return cancel
}

func runSubscribeLoop(ctx context.Context, channels []string, handler PubSubMessageHandler) error {
	ps := GetManager().subscribe(ctx, channels...)
	defer ps.Close()

	log.Printf("✅ Subscribe listener connected, channels: %v", channels)

	ch := ps.Channel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			handler(msg.Channel, msg.Payload)
		}
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
