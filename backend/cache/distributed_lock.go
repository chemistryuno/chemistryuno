package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	// ErrLockNotAcquired is returned when another process holds the lock.
	ErrLockNotAcquired = errors.New("lock not acquired: already held by another process")
	// ErrLockTokenMismatch is returned when releasing a lock with the wrong token.
	ErrLockTokenMismatch = errors.New("lock token mismatch: cannot release lock held by another process")
	// ErrRedisUnavailable is returned when Redis is not connected.
	ErrRedisUnavailable = errors.New("redis not available")
)

// releaseScript atomically releases a lock only if the token matches.
var releaseScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
else
  return 0
end
`)

func lockKey(resource string) string {
	return fmt.Sprintf("lock:%s", resource)
}

// AcquireLock tries to acquire a distributed lock for the given resource.
// Returns a unique token on success; ErrLockNotAcquired if already held.
// ttl is the maximum lock duration; the caller should release it before ttl expires.
func AcquireLock(ctx context.Context, resource string, ttl time.Duration) (string, error) {
	if !GetManager().IsAvailable() {
		return "", ErrRedisUnavailable
	}

	token := uuid.New().String()
	key := lockKey(resource)

	t := time.Now()
	ok, err := GetManager().client.SetNX(ctx, key, token, ttl).Result()
	redisCommandDuration.WithLabelValues("SETNX").Observe(time.Since(t).Seconds())

	if err != nil {
		return "", fmt.Errorf("acquire lock %s: %w", resource, err)
	}
	if !ok {
		return "", ErrLockNotAcquired
	}
	return token, nil
}

// ReleaseLock releases the lock if the token matches.
// Returns ErrLockTokenMismatch if the token does not match.
func ReleaseLock(ctx context.Context, resource, token string) error {
	if !GetManager().IsAvailable() {
		return ErrRedisUnavailable
	}

	t := time.Now()
	result, err := releaseScript.Run(ctx, GetManager().client, []string{lockKey(resource)}, token).Int()
	redisCommandDuration.WithLabelValues("EVAL").Observe(time.Since(t).Seconds())

	if err != nil {
		return fmt.Errorf("release lock %s: %w", resource, err)
	}
	if result == 0 {
		return ErrLockTokenMismatch
	}
	return nil
}

// WithLock acquires the lock, runs fn, then releases it.
// If Redis is unavailable the lock is skipped and fn runs without protection.
func WithLock(ctx context.Context, resource string, ttl time.Duration, fn func() error) error {
	token, err := AcquireLock(ctx, resource, ttl)
	if err == ErrRedisUnavailable {
		// Degrade gracefully: run without lock
		return fn()
	}
	if err != nil {
		return err
	}
	defer func() {
		_ = ReleaseLock(ctx, resource, token)
	}()
	return fn()
}
