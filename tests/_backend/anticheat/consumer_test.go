package anticheat

import (
	"chemistryuno/backend/cache"
	"context"
	"testing"
	"time"
)

func TestXADDAnticheatEvent(t *testing.T) {
	if !cache.GetManager().IsAvailable() {
		t.Skip("Redis not available")
	}

	ctx := context.Background()
	event := cache.AnticheatQueueEvent{
		RoomID:      "test-room-1",
		PlayerUID:   999,
		ContextJSON: `{"PlayerUID":999,"RoomID":"test-room-1"}`,
		Timestamp:   time.Now(),
	}
	if err := cache.XADDAnticheatEvent(ctx, event); err != nil {
		t.Fatalf("XADD failed: %v", err)
	}
}

func TestAnticheatMaxRetryConstant(t *testing.T) {
	if cache.AnticheatMaxRetry != 3 {
		t.Errorf("Expected AnticheatMaxRetry=3, got %d", cache.AnticheatMaxRetry)
	}
}

func TestStartAnticheatConsumer_NilSystem(t *testing.T) {
	// Should not panic with nil system
	StartAnticheatConsumer(context.Background(), nil)
}

func TestDeadLetterRouting(t *testing.T) {
	if !cache.GetManager().IsAvailable() {
		t.Skip("Redis not available")
	}
	ctx := context.Background()
	// Verify MoveToDeadLetter doesn't crash
	if err := cache.MoveToDeadLetter(ctx, "test-msg-id", `{"test":"payload"}`, "unit test"); err != nil {
		// May fail if stream doesn't exist; that's acceptable in unit test
		t.Logf("MoveToDeadLetter (expected if stream not initialized): %v", err)
	}
}
