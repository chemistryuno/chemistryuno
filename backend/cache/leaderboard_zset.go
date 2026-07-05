package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// leaderboardZSetKey returns the ZSET key for a given leaderboard field.
func leaderboardZSetKey(field string) string {
	return fmt.Sprintf("lb:zset:%s", field)
}

// ZADDLeaderboard updates a player's score in the leaderboard ZSET.
func ZADDLeaderboard(ctx context.Context, field string, uid uint, score float64) error {
	if !GetManager().IsAvailable() {
		return nil // Silently skip; SQL is authoritative
	}
	key := leaderboardZSetKey(field)
	t := time.Now()
	err := GetManager().client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: fmt.Sprintf("%d", uid),
	}).Err()
	redisCommandDuration.WithLabelValues("ZADD").Observe(time.Since(t).Seconds())
	if err != nil {
		log.Printf("⚠️  ZADD leaderboard:%s uid=%d: %v", field, uid, err)
	}
	return err
}

// ZADDLeaderboardAsync updates the ZSET in a goroutine, non-blocking.
func ZADDLeaderboardAsync(field string, uid uint, score float64) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = ZADDLeaderboard(ctx, field, uid, score)
	}()
}

// ZREVRANGELeaderboard fetches the top-N entries from the ZSET (highest score first).
// Returns slice of (uid_string, score) pairs; empty slice on miss.
func ZREVRANGELeaderboard(ctx context.Context, field string, limit int) ([]redis.Z, error) {
	if !GetManager().IsAvailable() {
		return nil, fmt.Errorf("redis not available")
	}
	key := leaderboardZSetKey(field)
	t := time.Now()
	result, err := GetManager().client.ZRevRangeWithScores(ctx, key, 0, int64(limit-1)).Result()
	redisCommandDuration.WithLabelValues("ZREVRANGE").Observe(time.Since(t).Seconds())
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ZREVRANKLeaderboard returns the 1-based rank of uid in the leaderboard (1 = top).
// Returns 0 if the player is not in the ZSET.
func ZREVRANKLeaderboard(ctx context.Context, field string, uid uint) (int64, error) {
	if !GetManager().IsAvailable() {
		return 0, fmt.Errorf("redis not available")
	}
	key := leaderboardZSetKey(field)
	t := time.Now()
	rank, err := GetManager().client.ZRevRank(ctx, key, fmt.Sprintf("%d", uid)).Result()
	redisCommandDuration.WithLabelValues("ZREVRANK").Observe(time.Since(t).Seconds())
	if err == redis.Nil {
		return 0, nil // Not in leaderboard
	}
	if err != nil {
		return 0, err
	}
	return rank + 1, nil // Convert 0-based to 1-based
}

// LeaderboardZSetExists reports whether the ZSET has been initialized.
func LeaderboardZSetExists(ctx context.Context, field string) (bool, error) {
	if !GetManager().IsAvailable() {
		return false, nil
	}
	count, err := GetManager().client.ZCard(ctx, leaderboardZSetKey(field)).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ZREMLeaderboard removes a player from all leaderboard ZSETs.
func ZREMLeaderboard(ctx context.Context, uid uint) {
	if !GetManager().IsAvailable() {
		return
	}
	member := fmt.Sprintf("%d", uid)
	for _, field := range []string{"points", "monthly_points", "total_xp", "win_count", "total_games"} {
		_ = GetManager().client.ZRem(ctx, leaderboardZSetKey(field), member).Err()
	}
}
