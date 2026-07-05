package game

import (
	"chemistryuno/backend/metrics"
	"container/list"
	"os"
	"sort"
	"strings"
	"sync"
)

// ReactionCache provides LRU caching for chemical reaction judgments
type ReactionCache struct {
	mu       sync.Mutex
	cache    sync.Map // map[string]bool - thread-safe for reads
	lru      *list.List
	entries  map[string]*list.Element
	maxSize  int
	enabled  bool
	hitCount uint64
	missCount uint64
}

type cacheEntry struct {
	key    string
	result bool
}

var globalReactionCache *ReactionCache

func init() {
	globalReactionCache = NewReactionCache(10000)
}

// NewReactionCache creates a new reaction cache with specified maximum size
func NewReactionCache(maxSize int) *ReactionCache {
	enabled := true
	if env := os.Getenv("ENABLE_REACTION_CACHE"); env == "false" || env == "0" {
		enabled = false
	}

	return &ReactionCache{
		lru:     list.New(),
		entries: make(map[string]*list.Element),
		maxSize: maxSize,
		enabled: enabled,
	}
}

// makeKey creates a commutative cache key from two substances (order-independent)
func makeKey(s1, s2 string) string {
	// Sort substances to ensure A+B and B+A generate the same key
	substances := []string{
		NormalizeSubscripts(s1),
		NormalizeSubscripts(s2),
	}
	sort.Strings(substances)
	return strings.Join(substances, "|")
}

// Get retrieves a cached reaction result
func (rc *ReactionCache) Get(s1, s2 string) (bool, bool) {
	if !rc.enabled {
		return false, false
	}

	key := makeKey(s1, s2)

	// Try fast read from sync.Map (lock-free)
	if val, ok := rc.cache.Load(key); ok {
		metrics.CacheHitsTotal.WithLabelValues("reaction").Inc()

		// Update LRU position (requires lock)
		rc.mu.Lock()
		if element, exists := rc.entries[key]; exists {
			rc.lru.MoveToFront(element)
		}
		rc.mu.Unlock()

		return val.(bool), true
	}

	metrics.CacheMissesTotal.WithLabelValues("reaction").Inc()
	return false, false
}

// Put stores a reaction result in the cache
func (rc *ReactionCache) Put(s1, s2 string, result bool) {
	if !rc.enabled {
		return
	}

	key := makeKey(s1, s2)

	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Check if already exists, update and move to front
	if element, exists := rc.entries[key]; exists {
		rc.lru.MoveToFront(element)
		element.Value.(*cacheEntry).result = result
		rc.cache.Store(key, result)
		return
	}

	// Add new entry
	entry := &cacheEntry{
		key:    key,
		result: result,
	}
	element := rc.lru.PushFront(entry)
	rc.entries[key] = element
	rc.cache.Store(key, result)

	// Evict LRU if over capacity
	if rc.lru.Len() > rc.maxSize {
		rc.evictLRU()
	}
}

// evictLRU removes the least recently used entry (must be called with lock held)
func (rc *ReactionCache) evictLRU() {
	element := rc.lru.Back()
	if element != nil {
		rc.lru.Remove(element)
		entry := element.Value.(*cacheEntry)
		delete(rc.entries, entry.key)
		rc.cache.Delete(entry.key)
	}
}

// Stats returns cache statistics
func (rc *ReactionCache) Stats() map[string]interface{} {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	return map[string]interface{}{
		"enabled":    rc.enabled,
		"size":       rc.lru.Len(),
		"max_size":   rc.maxSize,
		"hit_count":  rc.hitCount,
		"miss_count": rc.missCount,
	}
}

// Clear removes all entries from the cache
func (rc *ReactionCache) Clear() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.lru.Init()
	rc.entries = make(map[string]*list.Element)
	rc.cache = sync.Map{}
}
