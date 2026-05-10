package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func TestInitRedisWithPassword(t *testing.T) {
	server := miniredis.RunT(t)
	server.RequireAuth("secret")

	previous := redisClient
	t.Cleanup(func() {
		_ = Close()
		redisClient = previous
	})

	if err := InitRedis(server.Addr(), "", "secret", 0); err != nil {
		t.Fatalf("init redis with password: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := SetReactionCache(ctx, "H2", "O2", true); err != nil {
		t.Fatalf("set reaction cache with authenticated redis: %v", err)
	}

	got, err := GetReactionCache(ctx, "O2", "H2")
	if err != nil {
		t.Fatalf("get reaction cache with authenticated redis: %v", err)
	}
	if got != "1" {
		t.Fatalf("expected cached reaction to be 1, got %q", got)
	}
}
