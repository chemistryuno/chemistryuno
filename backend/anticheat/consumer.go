package anticheat

import (
	"chemistryuno/backend/cache"
	"context"
	"encoding/json"
	"log"
	"runtime/debug"
	"time"

	"github.com/redis/go-redis/v9"
)

// StartAnticheatConsumer launches the Redis Streams consumer goroutine.
// Enabled when ENABLE_ANTICHEAT_STREAMS=true.
func StartAnticheatConsumer(ctx context.Context, system *System) {
	if system == nil {
		return
	}
	if err := cache.EnsureAnticheatConsumerGroup(ctx); err != nil {
		log.Printf("⚠️  Failed to create anticheat consumer group: %v", err)
	}

	go consumerLoop(ctx, system)
}

func consumerLoop(ctx context.Context, system *System) {
	consumerName := "consumer-1"
	log.Printf("✅ Anticheat Streams consumer started (consumer: %s)", consumerName)

	for {
		select {
		case <-ctx.Done():
			log.Println("ℹ️  Anticheat consumer shutting down")
			return
		default:
		}

		streams, err := cache.ReadAnticheatEvents(ctx, consumerName, 10, 5*time.Second)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("⚠️  XREADGROUP error: %v, retrying in 2s", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				consumeMessage(ctx, system, msg)
			}
		}
	}
}

func consumeMessage(ctx context.Context, system *System, msg redis.XMessage) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ Anticheat consumer panic for msg %s: %v\n%s", msg.ID, r, debug.Stack())
			// On panic, move to dead-letter so we don't loop forever
			payload, _ := msg.Values["payload"].(string)
			_ = cache.MoveToDeadLetter(ctx, msg.ID, payload, "consumer panic")
		}
	}()

	event, err := cache.ParseAnticheatEvent(msg)
	if err != nil {
		log.Printf("⚠️  Failed to parse anticheat event %s: %v", msg.ID, err)
		_ = cache.MoveToDeadLetter(ctx, msg.ID, "", err.Error())
		return
	}

	// Check retry count — move to dead-letter after max retries
	retries, _ := cache.RetryCountForMessage(ctx, msg.ID)
	if retries >= cache.AnticheatMaxRetry {
		log.Printf("⚠️  Message %s exceeded max retries (%d), moving to dead-letter", msg.ID, cache.AnticheatMaxRetry)
		payload, _ := msg.Values["payload"].(string)
		_ = cache.MoveToDeadLetter(ctx, msg.ID, payload, "max retries exceeded")
		return
	}

	// Deserialize detection context
	var detCtx DetectionContext
	if err := json.Unmarshal([]byte(event.ContextJSON), &detCtx); err != nil {
		log.Printf("⚠️  Failed to deserialize detection context for msg %s: %v", msg.ID, err)
		_ = cache.MoveToDeadLetter(ctx, msg.ID, event.ContextJSON, err.Error())
		return
	}

	// Process the anticheat detection
	_, _, err = system.ProcessGameEnd(event.RoomID, event.PlayerUID, &detCtx)
	if err != nil {
		log.Printf("⚠️  Anticheat processing failed for msg %s (room %s, player %d): %v",
			msg.ID, event.RoomID, event.PlayerUID, err)
		// Do NOT ACK — message stays in pending, will be retried
		return
	}

	// Success: ACK the message
	if err := cache.ACKAnticheatEvent(ctx, msg.ID); err != nil {
		log.Printf("⚠️  XACK failed for msg %s: %v", msg.ID, err)
	}
}
