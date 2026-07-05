package cache

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
)

var (
	redisCommandDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chemistryuno_redis_command_duration_seconds",
			Help:    "Duration of Redis commands",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1},
		},
		[]string{"command"},
	)

	redisPoolHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chemistryuno_redis_pool_hits_total",
			Help: "Total Redis connection pool hits/misses",
		},
		[]string{"result"}, // "hit" or "miss"
	)
)

// Manager wraps the Redis client with nil-safety, metrics, and helper methods.
type Manager struct {
	client *redis.Client
}

// globalManager is the package-level singleton, initialized by InitRedis.
var globalManager *Manager

// GetManager returns the global Manager. The returned manager is always non-nil;
// callers should check IsAvailable() before performing operations.
func GetManager() *Manager {
	if globalManager == nil {
		globalManager = &Manager{}
	}
	return globalManager
}

// IsAvailable reports whether Redis is connected and usable.
func (m *Manager) IsAvailable() bool {
	return m != nil && m.client != nil
}

// Client returns the underlying redis.Client, may be nil.
func (m *Manager) Client() *redis.Client {
	if m == nil {
		return nil
	}
	return m.client
}

// Pipeline returns a Redis pipeline for batching commands.
func (m *Manager) Pipeline() (redis.Pipeliner, error) {
	if !m.IsAvailable() {
		return nil, errNotInitialized
	}
	return m.client.Pipeline(), nil
}

// TxPipeline returns a transactional pipeline (MULTI/EXEC).
func (m *Manager) TxPipeline() (redis.Pipeliner, error) {
	if !m.IsAvailable() {
		return nil, errNotInitialized
	}
	return m.client.TxPipeline(), nil
}

// -- Low-level instrumented wrappers --

func (m *Manager) get(ctx context.Context, key string) (string, error) {
	t := time.Now()
	val, err := m.client.Get(ctx, key).Result()
	redisCommandDuration.WithLabelValues("GET").Observe(time.Since(t).Seconds())
	if err == redis.Nil {
		redisPoolHits.WithLabelValues("miss").Inc()
	} else if err == nil {
		redisPoolHits.WithLabelValues("hit").Inc()
	}
	return val, err
}

func (m *Manager) set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	t := time.Now()
	err := m.client.Set(ctx, key, value, ttl).Err()
	redisCommandDuration.WithLabelValues("SET").Observe(time.Since(t).Seconds())
	return err
}

func (m *Manager) del(ctx context.Context, keys ...string) error {
	t := time.Now()
	err := m.client.Del(ctx, keys...).Err()
	redisCommandDuration.WithLabelValues("DEL").Observe(time.Since(t).Seconds())
	return err
}

func (m *Manager) setNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	t := time.Now()
	ok, err := m.client.SetNX(ctx, key, value, ttl).Result()
	redisCommandDuration.WithLabelValues("SETNX").Observe(time.Since(t).Seconds())
	return ok, err
}

func (m *Manager) eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	t := time.Now()
	val, err := m.client.Eval(ctx, script, keys, args...).Result()
	redisCommandDuration.WithLabelValues("EVAL").Observe(time.Since(t).Seconds())
	return val, err
}

func (m *Manager) publish(ctx context.Context, channel string, message interface{}) error {
	t := time.Now()
	err := m.client.Publish(ctx, channel, message).Err()
	redisCommandDuration.WithLabelValues("PUBLISH").Observe(time.Since(t).Seconds())
	return err
}

func (m *Manager) subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return m.client.Subscribe(ctx, channels...)
}

func (m *Manager) pSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	return m.client.PSubscribe(ctx, patterns...)
}

func (m *Manager) incr(ctx context.Context, key string) (int64, error) {
	t := time.Now()
	val, err := m.client.Incr(ctx, key).Result()
	redisCommandDuration.WithLabelValues("INCR").Observe(time.Since(t).Seconds())
	return val, err
}

func (m *Manager) decr(ctx context.Context, key string) (int64, error) {
	t := time.Now()
	val, err := m.client.Decr(ctx, key).Result()
	redisCommandDuration.WithLabelValues("DECR").Observe(time.Since(t).Seconds())
	return val, err
}

func (m *Manager) expire(ctx context.Context, key string, ttl time.Duration) error {
	return m.client.Expire(ctx, key, ttl).Err()
}

func (m *Manager) pfAdd(ctx context.Context, key string, els ...interface{}) error {
	t := time.Now()
	err := m.client.PFAdd(ctx, key, els...).Err()
	redisCommandDuration.WithLabelValues("PFADD").Observe(time.Since(t).Seconds())
	return err
}

func (m *Manager) pfCount(ctx context.Context, keys ...string) (int64, error) {
	t := time.Now()
	val, err := m.client.PFCount(ctx, keys...).Result()
	redisCommandDuration.WithLabelValues("PFCOUNT").Observe(time.Since(t).Seconds())
	return val, err
}
