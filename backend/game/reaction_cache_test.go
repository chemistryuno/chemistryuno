package game

import (
	"sync"
	"testing"
)

func TestReactionCacheBasic(t *testing.T) {
	cache := NewReactionCache(3)
	cache.enabled = true

	// Test cache miss
	if _, found := cache.Get("H2O", "NaCl"); found {
		t.Error("Expected cache miss for new entry")
	}

	// Test cache put and hit
	cache.Put("H2O", "NaCl", false)
	if result, found := cache.Get("H2O", "NaCl"); !found || result != false {
		t.Error("Expected cache hit with result=false")
	}

	cache.Put("HCl", "NaOH", true)
	if result, found := cache.Get("HCl", "NaOH"); !found || result != true {
		t.Error("Expected cache hit with result=true")
	}
}

func TestReactionCacheCommutative(t *testing.T) {
	cache := NewReactionCache(10)
	cache.enabled = true

	// Store A + B
	cache.Put("HCl", "NaOH", true)

	// Retrieve B + A (should hit same cache entry)
	if result, found := cache.Get("NaOH", "HCl"); !found || result != true {
		t.Error("Cache should be commutative: A+B = B+A")
	}
}

func TestReactionCacheLRUEviction(t *testing.T) {
	cache := NewReactionCache(3)
	cache.enabled = true

	// Fill cache to capacity
	cache.Put("A", "B", true)
	cache.Put("C", "D", true)
	cache.Put("E", "F", true)

	// Verify all three are in cache
	if _, found := cache.Get("A", "B"); !found {
		t.Error("Entry A+B should be in cache")
	}
	if _, found := cache.Get("C", "D"); !found {
		t.Error("Entry C+D should be in cache")
	}
	if _, found := cache.Get("E", "F"); !found {
		t.Error("Entry E+F should be in cache")
	}

	// Access A+B to make it recently used
	cache.Get("A", "B")

	// Add new entry (should evict C+D, the LRU)
	cache.Put("G", "H", false)

	// Verify C+D was evicted
	if _, found := cache.Get("C", "D"); found {
		t.Error("Entry C+D should have been evicted")
	}

	// Verify A+B is still in cache (it was recently accessed)
	if _, found := cache.Get("A", "B"); !found {
		t.Error("Entry A+B should still be in cache")
	}

	// Verify new entry is in cache
	if result, found := cache.Get("G", "H"); !found || result != false {
		t.Error("New entry G+H should be in cache")
	}
}

func TestReactionCacheConcurrency(t *testing.T) {
	cache := NewReactionCache(100)
	cache.enabled = true

	var wg sync.WaitGroup
	numGoroutines := 50
	operationsPerGoroutine := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key1 := string(rune('A' + (id+j)%26))
				key2 := string(rune('B' + (id+j)%26))
				cache.Put(key1, key2, (id+j)%2 == 0)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key1 := string(rune('A' + (id+j)%26))
				key2 := string(rune('B' + (id+j)%26))
				cache.Get(key1, key2)
			}
		}(i)
	}

	wg.Wait()

	// If we got here without panic, concurrency test passed
	stats := cache.Stats()
	if size := stats["size"].(int); size > 100 {
		t.Errorf("Cache size exceeded max: %d > 100", size)
	}
}

func TestReactionCacheDisabled(t *testing.T) {
	cache := NewReactionCache(10)
	cache.enabled = false

	// Operations should be no-ops when disabled
	cache.Put("H2O", "NaCl", true)
	if _, found := cache.Get("H2O", "NaCl"); found {
		t.Error("Cache should not store when disabled")
	}
}

func TestReactionCacheClear(t *testing.T) {
	cache := NewReactionCache(10)
	cache.enabled = true

	cache.Put("A", "B", true)
	cache.Put("C", "D", false)

	cache.Clear()

	// Verify cache is empty
	if _, found := cache.Get("A", "B"); found {
		t.Error("Cache should be empty after Clear()")
	}
	if _, found := cache.Get("C", "D"); found {
		t.Error("Cache should be empty after Clear()")
	}

	stats := cache.Stats()
	if size := stats["size"].(int); size != 0 {
		t.Errorf("Cache size should be 0 after clear, got %d", size)
	}
}
