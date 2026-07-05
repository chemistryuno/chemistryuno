package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	anticheatStream   = "anticheat:queue"
	anticheatDLStream = "anticheat:dead-letter"
	anticheatGroup    = "anticheat-consumers"
	anticheatMaxLen   = 10000

	// AnticheatMaxRetry is the maximum delivery attempts before moving to dead-letter.
	AnticheatMaxRetry = 3
)

// AnticheatQueueEvent holds the data pushed into the Streams queue.
type AnticheatQueueEvent struct {
	RoomID      string    `json:"room_id"`
	PlayerUID   uint      `json:"player_uid"`
	ContextJSON string    `json:"context_json"`
	Timestamp   time.Time `json:"timestamp"`
}

// XADDAnticheatEvent pushes an anticheat detection event onto the Streams queue.
func XADDAnticheatEvent(ctx context.Context, event AnticheatQueueEvent) error {
	if !GetManager().IsAvailable() {
		return ErrRedisUnavailable
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal anticheat event: %w", err)
	}

	t := time.Now()
	_, err = GetManager().client.XAdd(ctx, &redis.XAddArgs{
		Stream: anticheatStream,
		MaxLen: anticheatMaxLen,
		Approx: true,
		Values: map[string]interface{}{
			"payload": string(payload),
		},
	}).Result()
	redisCommandDuration.WithLabelValues("XADD").Observe(time.Since(t).Seconds())

	return err
}

// EnsureAnticheatConsumerGroup creates the consumer group if it doesn't exist.
func EnsureAnticheatConsumerGroup(ctx context.Context) error {
	if !GetManager().IsAvailable() {
		return ErrRedisUnavailable
	}
	err := GetManager().client.XGroupCreateMkStream(ctx, anticheatStream, anticheatGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return err
	}
	return nil
}

// ReadAnticheatEvents reads pending events for a consumer. Blocks up to blockMs.
func ReadAnticheatEvents(ctx context.Context, consumerName string, count int64, blockMs time.Duration) ([]redis.XStream, error) {
	if !GetManager().IsAvailable() {
		return nil, ErrRedisUnavailable
	}

	t := time.Now()
	result, err := GetManager().client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    anticheatGroup,
		Consumer: consumerName,
		Streams:  []string{anticheatStream, ">"},
		Count:    count,
		Block:    blockMs,
	}).Result()
	redisCommandDuration.WithLabelValues("XREADGROUP").Observe(time.Since(t).Seconds())

	if err == redis.Nil {
		return nil, nil // Timeout, no messages
	}
	return result, err
}

// ACKAnticheatEvent acknowledges a successfully processed event.
func ACKAnticheatEvent(ctx context.Context, messageID string) error {
	if !GetManager().IsAvailable() {
		return ErrRedisUnavailable
	}
	return GetManager().client.XAck(ctx, anticheatStream, anticheatGroup, messageID).Err()
}

// MoveToDeadLetter moves a message to the dead-letter stream and ACKs it.
func MoveToDeadLetter(ctx context.Context, messageID string, payload string, reason string) error {
	if !GetManager().IsAvailable() {
		return ErrRedisUnavailable
	}

	pipe, _ := GetManager().Pipeline()
	pipe.XAdd(ctx, &redis.XAddArgs{
		Stream: anticheatDLStream,
		MaxLen: 1000,
		Approx: true,
		Values: map[string]interface{}{
			"original_id": messageID,
			"payload":     payload,
			"reason":      reason,
			"moved_at":    time.Now().Format(time.RFC3339),
		},
	})
	pipe.XAck(ctx, anticheatStream, anticheatGroup, messageID)
	_, err := pipe.Exec(ctx)
	return err
}

// GetAnticheatPendingCount returns the number of pending (unprocessed) messages.
func GetAnticheatPendingCount(ctx context.Context) (int64, error) {
	if !GetManager().IsAvailable() {
		return 0, ErrRedisUnavailable
	}
	info, err := GetManager().client.XPending(ctx, anticheatStream, anticheatGroup).Result()
	if err != nil {
		return 0, err
	}
	return info.Count, nil
}

// ParseAnticheatEvent decodes an event from a stream message.
func ParseAnticheatEvent(msg redis.XMessage) (*AnticheatQueueEvent, error) {
	payloadStr, ok := msg.Values["payload"].(string)
	if !ok {
		return nil, fmt.Errorf("missing payload field in message %s", msg.ID)
	}
	var event AnticheatQueueEvent
	if err := json.Unmarshal([]byte(payloadStr), &event); err != nil {
		return nil, fmt.Errorf("unmarshal anticheat event: %w", err)
	}
	return &event, nil
}

// RetryCountForMessage returns the delivery count for a pending message (for dead-letter routing).
func RetryCountForMessage(ctx context.Context, messageID string) (int64, error) {
	if !GetManager().IsAvailable() {
		return 0, nil
	}
	pendingMsgs, err := GetManager().client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: anticheatStream,
		Group:  anticheatGroup,
		Start:  messageID,
		End:    "+",
		Count:  1,
	}).Result()
	if err != nil || len(pendingMsgs) == 0 {
		return 0, err
	}
	return pendingMsgs[0].RetryCount, nil
}
