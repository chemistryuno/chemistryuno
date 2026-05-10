package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func withMiniRedis(t *testing.T) context.Context {
	t.Helper()

	server := miniredis.RunT(t)
	previous := redisClient
	redisClient = redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
		redisClient = previous
	})

	return context.Background()
}

func TestLeaderboardCacheIncludesLastOfflineAt(t *testing.T) {
	if LeaderboardCacheIncludesLastOfflineAt(`[{"uid":1001,"last_offline_at":null}]`) != true {
		t.Fatalf("expected cache with explicit last_offline_at to be valid")
	}
	if LeaderboardCacheIncludesLastOfflineAt(`[{"uid":1001}]`) {
		t.Fatalf("expected legacy cache without last_offline_at to be invalid")
	}

	offlineAt := time.Date(2026, 5, 6, 8, 30, 0, 0, time.UTC)
	data := `[{"uid":1001,"last_offline_at":"` + offlineAt.Format(time.RFC3339) + `"}]`
	if !LeaderboardCacheIncludesLastOfflineAt(data) {
		t.Fatalf("expected cache with timestamp last_offline_at to be valid")
	}
}

func TestUnbanCompensationIdempotencyRedisKeyLifecycle(t *testing.T) {
	ctx := withMiniRedis(t)
	const userUID uint = 4101
	const eventID = "appeal_99"

	duplicate, err := CheckUnbanCompensationIdempotency(ctx, userUID, eventID)
	if err != nil {
		t.Fatalf("check empty idempotency key: %v", err)
	}
	if duplicate {
		t.Fatal("expected empty idempotency key to allow compensation")
	}

	if err := SetUnbanCompensationIdempotency(ctx, userUID, eventID, 60); err != nil {
		t.Fatalf("set idempotency key: %v", err)
	}

	duplicate, err = CheckUnbanCompensationIdempotency(ctx, userUID, eventID)
	if err != nil {
		t.Fatalf("check populated idempotency key: %v", err)
	}
	if !duplicate {
		t.Fatal("expected populated idempotency key to report duplicate compensation")
	}

	if err := DeleteUnbanCompensationIdempotency(ctx, userUID, eventID); err != nil {
		t.Fatalf("delete idempotency key: %v", err)
	}

	duplicate, err = CheckUnbanCompensationIdempotency(ctx, userUID, eventID)
	if err != nil {
		t.Fatalf("check deleted idempotency key: %v", err)
	}
	if duplicate {
		t.Fatal("expected deleted idempotency key to allow retry")
	}
}
