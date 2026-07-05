package middleware

import (
	"testing"
	"time"
)

func TestRateLimiterCleanup(t *testing.T) {
	store := &RateLimitStore{
		requests: make(map[string][]time.Time),
	}

	now := time.Now()

	// Add some recent and old entries
	store.requests["key1"] = []time.Time{
		now.Add(-30 * time.Hour), // Old - should be removed
		now.Add(-1 * time.Hour),  // Recent - should be kept
	}
	store.requests["key2"] = []time.Time{
		now.Add(-48 * time.Hour), // Old - should be removed
		now.Add(-50 * time.Hour), // Very old - should be removed
	}
	store.requests["key3"] = []time.Time{
		now.Add(-10 * time.Minute), // Recent - should be kept
		now.Add(-5 * time.Minute),  // Recent - should be kept
	}

	// Verify initial state
	if len(store.requests) != 3 {
		t.Errorf("Expected 3 keys initially, got %d", len(store.requests))
	}

	// Run cleanup with 24-hour retention
	store.cleanup(24 * time.Hour)

	// Verify key2 is removed (all entries old)
	if _, exists := store.requests["key2"]; exists {
		t.Error("key2 should have been removed (all entries expired)")
	}

	// Verify key1 still exists but only has one entry
	if requests, exists := store.requests["key1"]; !exists {
		t.Error("key1 should still exist")
	} else if len(requests) != 1 {
		t.Errorf("key1 should have 1 entry after cleanup, got %d", len(requests))
	}

	// Verify key3 still exists with both entries
	if requests, exists := store.requests["key3"]; !exists {
		t.Error("key3 should still exist")
	} else if len(requests) != 2 {
		t.Errorf("key3 should have 2 entries after cleanup, got %d", len(requests))
	}

	// Final check: should have 2 keys remaining
	if len(store.requests) != 2 {
		t.Errorf("Expected 2 keys after cleanup, got %d", len(store.requests))
	}
}

func TestRateLimiterCleanupEmptyStore(t *testing.T) {
	store := &RateLimitStore{
		requests: make(map[string][]time.Time),
	}

	// Cleanup should not panic on empty store
	store.cleanup(24 * time.Hour)

	if len(store.requests) != 0 {
		t.Error("Empty store should remain empty after cleanup")
	}
}

func TestRateLimiterCleanupNoExpiredEntries(t *testing.T) {
	store := &RateLimitStore{
		requests: make(map[string][]time.Time),
	}

	now := time.Now()

	// Add only recent entries
	store.requests["key1"] = []time.Time{
		now.Add(-1 * time.Hour),
		now.Add(-30 * time.Minute),
	}
	store.requests["key2"] = []time.Time{
		now.Add(-10 * time.Minute),
	}

	// Run cleanup
	store.cleanup(24 * time.Hour)

	// All keys should still exist
	if len(store.requests) != 2 {
		t.Errorf("Expected 2 keys after cleanup (nothing expired), got %d", len(store.requests))
	}

	// All entries should be retained
	if len(store.requests["key1"]) != 2 {
		t.Error("key1 should have 2 entries (nothing expired)")
	}
	if len(store.requests["key2"]) != 1 {
		t.Error("key2 should have 1 entry (nothing expired)")
	}
}
