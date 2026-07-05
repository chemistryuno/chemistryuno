package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	onlineUsersGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "chemistryuno_online_users_total",
		Help: "Current number of active WebSocket connections (across all instances)",
	})
)

const onlineCountKey = "online:count"

// IncrOnlineCount atomically increments the shared online user counter.
func IncrOnlineCount(ctx context.Context) {
	if !GetManager().IsAvailable() {
		return
	}
	if _, err := GetManager().incr(ctx, onlineCountKey); err != nil {
		log.Printf("⚠️  INCR online:count failed: %v", err)
	}
}

// DecrOnlineCount atomically decrements the shared online user counter, min 0.
func DecrOnlineCount(ctx context.Context) {
	if !GetManager().IsAvailable() {
		return
	}
	// Lua script ensures count never goes below 0
	script := `
local v = redis.call("DECR", KEYS[1])
if tonumber(v) < 0 then
  redis.call("SET", KEYS[1], 0)
  return 0
end
return v`
	if _, err := GetManager().eval(ctx, script, []string{onlineCountKey}); err != nil {
		log.Printf("⚠️  DECR online:count failed: %v", err)
	}
}

// SetOnlineCount overrides the online count (used on server restart to sync with local hub).
func SetOnlineCount(ctx context.Context, count int64) {
	if !GetManager().IsAvailable() {
		return
	}
	if err := GetManager().set(ctx, onlineCountKey, count, 0); err != nil {
		log.Printf("⚠️  SET online:count failed: %v", err)
	}
}

// GetOnlineCount reads the current online user count from Redis.
func GetOnlineCount(ctx context.Context) (int64, error) {
	if !GetManager().IsAvailable() {
		return 0, ErrRedisUnavailable
	}
	val, err := GetManager().get(ctx, onlineCountKey)
	if err != nil {
		return 0, err
	}
	var count int64
	fmt.Sscanf(val, "%d", &count)
	return count, nil
}

// UpdateOnlineUsersMetric reads online:count and updates the Prometheus gauge.
func UpdateOnlineUsersMetric(ctx context.Context) {
	count, err := GetOnlineCount(ctx)
	if err != nil {
		return
	}
	onlineUsersGauge.Set(float64(count))
}

// -- HyperLogLog UV tracking --

func dailyUVKey(date string) string {
	return fmt.Sprintf("online:uv:%s", date)
}

// AddDailyUV adds a user ID to the HyperLogLog for today's date.
func AddDailyUV(ctx context.Context, uid uint) {
	if !GetManager().IsAvailable() {
		return
	}
	today := time.Now().Format("2006-01-02")
	key := dailyUVKey(today)
	if err := GetManager().pfAdd(ctx, key, uid); err != nil {
		log.Printf("⚠️  PFADD daily UV failed: %v", err)
		return
	}
	// Set 7-day TTL (idempotent expire)
	_ = GetManager().expire(ctx, key, 7*24*time.Hour)
}

// GetDailyUV returns the estimated unique visitor count for a given date (YYYY-MM-DD).
func GetDailyUV(ctx context.Context, date string) (int64, error) {
	if !GetManager().IsAvailable() {
		return 0, ErrRedisUnavailable
	}
	return GetManager().pfCount(ctx, dailyUVKey(date))
}
