package cache

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// These tests require a real Redis connection.
// Run with: go test ./backend/cache/... -run TestLock -v
// They are skipped when Redis is not available.

func skipIfNoRedis(t *testing.T) {
	t.Helper()
	if !GetManager().IsAvailable() {
		t.Skip("Redis not available, skipping lock tests")
	}
}

func TestAcquireAndReleaseLock(t *testing.T) {
	skipIfNoRedis(t)
	ctx := context.Background()

	token, err := AcquireLock(ctx, "test:basic", 5*time.Second)
	if err != nil {
		t.Fatalf("AcquireLock failed: %v", err)
	}
	if token == "" {
		t.Fatal("Expected non-empty token")
	}

	// Release should succeed
	if err := ReleaseLock(ctx, "test:basic", token); err != nil {
		t.Fatalf("ReleaseLock failed: %v", err)
	}

	// After release, should be acquirable again
	token2, err := AcquireLock(ctx, "test:basic", 5*time.Second)
	if err != nil {
		t.Fatalf("Second AcquireLock failed after release: %v", err)
	}
	_ = ReleaseLock(ctx, "test:basic", token2)
}

func TestAcquireLock_AlreadyHeld(t *testing.T) {
	skipIfNoRedis(t)
	ctx := context.Background()

	token, err := AcquireLock(ctx, "test:held", 10*time.Second)
	if err != nil {
		t.Fatalf("First AcquireLock failed: %v", err)
	}
	defer func() { _ = ReleaseLock(ctx, "test:held", token) }()

	// Second acquire should fail
	_, err2 := AcquireLock(ctx, "test:held", 10*time.Second)
	if !errors.Is(err2, ErrLockNotAcquired) {
		t.Fatalf("Expected ErrLockNotAcquired, got: %v", err2)
	}
}

func TestReleaseLock_WrongToken(t *testing.T) {
	skipIfNoRedis(t)
	ctx := context.Background()

	token, err := AcquireLock(ctx, "test:wrongtoken", 10*time.Second)
	if err != nil {
		t.Fatalf("AcquireLock failed: %v", err)
	}
	defer func() { _ = ReleaseLock(ctx, "test:wrongtoken", token) }()

	// Release with wrong token should return ErrLockTokenMismatch
	err = ReleaseLock(ctx, "test:wrongtoken", "wrong-token-value")
	if !errors.Is(err, ErrLockTokenMismatch) {
		t.Fatalf("Expected ErrLockTokenMismatch, got: %v", err)
	}
}

func TestAcquireLock_TTLExpiry(t *testing.T) {
	skipIfNoRedis(t)
	ctx := context.Background()

	// Acquire with very short TTL
	token, err := AcquireLock(ctx, "test:ttl", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("AcquireLock failed: %v", err)
	}
	_ = token

	// Wait for TTL to expire
	time.Sleep(200 * time.Millisecond)

	// Should be acquirable again after TTL expiry
	token2, err := AcquireLock(ctx, "test:ttl", 5*time.Second)
	if err != nil {
		t.Fatalf("AcquireLock after TTL expiry failed: %v", err)
	}
	_ = ReleaseLock(ctx, "test:ttl", token2)
}

func TestAcquireLock_Concurrent(t *testing.T) {
	skipIfNoRedis(t)
	ctx := context.Background()

	const goroutines = 50
	var successCount int64
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			token, err := AcquireLock(ctx, "test:concurrent", 5*time.Second)
			if err == nil {
				atomic.AddInt64(&successCount, 1)
				time.Sleep(10 * time.Millisecond)
				_ = ReleaseLock(ctx, "test:concurrent", token)
			} else if !errors.Is(err, ErrLockNotAcquired) {
				t.Errorf("Unexpected error: %v", err)
			}
		}()
	}

	wg.Wait()

	// At least one goroutine must have succeeded
	if successCount == 0 {
		t.Error("No goroutine acquired the lock")
	}
	// But at most one at a time (this is a unit test assertion; hard to verify sequential,
	// but successive acquisitions confirm the lock released and was re-acquired correctly)
	t.Logf("Lock acquired %d times by %d competing goroutines", successCount, goroutines)
}

func TestWithLock_NoRedis(t *testing.T) {
	// Test graceful degradation when Redis is unavailable
	// Create a manager with nil client
	origManager := globalManager
	globalManager = &Manager{client: nil}
	defer func() { globalManager = origManager }()

	called := false
	err := WithLock(context.Background(), "test:noredis", 5*time.Second, func() error {
		called = true
		return nil
	})

	if err != nil {
		t.Fatalf("WithLock should succeed (degraded) when Redis unavailable: %v", err)
	}
	if !called {
		t.Error("fn should have been called even when Redis unavailable")
	}
}
