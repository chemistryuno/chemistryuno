package repository

import (
	"chemistryuno/backend/cache"
	"chemistryuno/backend/database"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// GetLeaderboardWithCache 使用缓存获取排行榜 - 先查Redis，miss后查DB
func GetLeaderboardWithCache(ctx context.Context, orderBy string, limit int) ([]cache.LeaderboardCache, error) {
	// 先查缓存
	cached, err := cache.GetLeaderboardCache(ctx, orderBy, limit)
	if err != nil && err.Error() != "redis client not initialized" {
		log.Printf("⚠️  Redis leaderboard 查询异常: %v", err)
	} else if cached != "" {
		var cacheEntries []cache.LeaderboardCache
		if err := json.Unmarshal([]byte(cached), &cacheEntries); err == nil && len(cacheEntries) > 0 && cache.LeaderboardCacheIncludesLastOfflineAt(cached) {
			return cacheEntries, nil
		}
	}

	// 缓存 miss，查询数据库
	userRepo := NewUserRepository()
	entries, err := userRepo.GetLeaderboardOptimized(orderBy, limit)
	if err != nil {
		log.Printf("❌ 数据库 leaderboard 查询失败: %v", err)
		return nil, err
	}

	// 转换为缓存格式
	cacheEntries := make([]cache.LeaderboardCache, len(entries))
	for i, entry := range entries {
		cacheEntries[i] = cache.LeaderboardCache{
			UID:           entry.UID,
			Username:      entry.Username,
			Nickname:      entry.Nickname,
			Avatar:        entry.Avatar,
			Points:        entry.Points,
			MonthlyPoints: entry.MonthlyPoints,
			Level:         entry.Level,
			TotalXP:       entry.TotalXP,
			WinCount:      entry.WinCount,
			TotalGames:    entry.TotalGames,
			LastOfflineAt: entry.LastOfflineAt,
		}
	}

	// 缓存到 Redis K/V
	if data, err := json.Marshal(cacheEntries); err == nil {
		_ = cache.SetLeaderboardCache(ctx, orderBy, limit, string(data))
	}

	// 如果 ZSET 还不存在，异步重建（冷启动或 Redis 清空后）
	go func() {
		zCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		exists, _ := cache.LeaderboardZSetExists(zCtx, orderBy)
		if !exists {
			RebuildLeaderboardZSet(zCtx, orderBy)
		}
	}()

	return cacheEntries, nil
}

// GetBountyTotalWithCache 使用缓存获取赏金总额 - 先查Redis，miss后查DB
func GetBountyTotalWithCache(ctx context.Context, uid uint) (int, error) {
	// 先查缓存
	cached, err := cache.GetBountyTotalCache(ctx, uid)
	if err != nil && err.Error() != "redis client not initialized" {
		log.Printf("⚠️  Redis bounty 查询异常: %v", err)
	} else if cached != "" {
		amount, err := strconv.ParseFloat(cached, 64)
		if err == nil && amount > 0 {
			return int(amount), nil
		}
	}

	// 缓存 miss，查询数据库
	bountyRepo := NewBountyRepository()
	total, err := bountyRepo.GetTotalBounty(uid)
	if err != nil {
		log.Printf("❌ 数据库 bounty 查询失败 UID=%d: %v", uid, err)
		return 0, err
	}

	// 缓存到 Redis
	_ = cache.SetBountyTotalCache(ctx, uid, float64(total))

	return total, nil
}

// RebuildLeaderboardZSet rebuilds the ZSET for a given leaderboard field from SQL.
// It runs in batches of 500 to avoid oversized pipelines.
func RebuildLeaderboardZSet(ctx context.Context, field string) {
	type scoreRow struct {
		UID   uint
		Score float64
	}

	colMap := map[string]string{
		"points":         "points",
		"monthly_points": "monthly_points",
		"total_xp":       "total_xp",
		"win_count":      "win_count",
		"total_games":    "total_games",
	}
	col, ok := colMap[field]
	if !ok {
		return
	}

	var rows []scoreRow
	if err := database.DB.Model(&database.User{}).
		Select(fmt.Sprintf("uid, CAST(%s AS DECIMAL(20,4)) as score", col)).
		Scan(&rows).Error; err != nil {
		log.Printf("⚠️  RebuildLeaderboardZSet query failed for %s: %v", field, err)
		return
	}

	if len(rows) == 0 {
		return
	}

	batchSize := 500
	for i := 0; i < len(rows); i += batchSize {
		end := i + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := rows[i:end]

		pipe, err := cache.GetManager().Pipeline()
		if err != nil {
			return
		}
		key := fmt.Sprintf("lb:zset:%s", field)
		for _, row := range batch {
			pipe.ZAdd(ctx, key, redis.Z{Score: row.Score, Member: fmt.Sprintf("%d", row.UID)})
		}
		if _, err := pipe.Exec(ctx); err != nil {
			log.Printf("⚠️  RebuildLeaderboardZSet pipeline exec failed: %v", err)
			return
		}
	}

	log.Printf("✅ Rebuilt leaderboard ZSET for '%s' with %d entries", field, len(rows))
}

// GetLevelConfigsWithCache 使用缓存获取等级配置 - 先查Redis，miss后查DB
func GetLevelConfigsWithCache(ctx context.Context) ([]database.LevelConfig, error) {
	// 先查缓存
	cached, err := cache.GetLevelConfigsCache(ctx)
	if err != nil && err.Error() != "redis client not initialized" {
		log.Printf("⚠️  Redis level configs 查询异常: %v", err)
	} else if cached != "" {
		var configs []database.LevelConfig
		if err := json.Unmarshal([]byte(cached), &configs); err == nil {
			return configs, nil
		}
	}

	// 缓存 miss，查询数据库
	var configs []database.LevelConfig
	if err := database.DB.Find(&configs).Error; err != nil {
		log.Printf("❌ 数据库 level configs 查询失败: %v", err)
		return nil, err
	}

	// 缓存到 Redis
	if data, err := json.Marshal(configs); err == nil {
		_ = cache.SetLevelConfigsCache(ctx, string(data))
	}

	return configs, nil
}

// GetApprovedReactionsWithCache 使用缓存获取已批准反应 - 先查Redis，miss后查DB
func GetApprovedReactionsWithCache(ctx context.Context) ([]ReactionWithCreator, error) {
	// 先查缓存
	cached, err := cache.GetApprovedReactionsCache(ctx)
	if err != nil && err.Error() != "redis client not initialized" {
		log.Printf("⚠️  Redis approved reactions 查询异常: %v", err)
	} else if cached != "" {
		var reactions []ReactionWithCreator
		if err := json.Unmarshal([]byte(cached), &reactions); err == nil {
			return reactions, nil
		}
	}

	// 缓存 miss，查询数据库
	reactionRepo := NewReactionRepository()
	reactions, err := reactionRepo.FindApprovedGrouped()
	if err != nil {
		log.Printf("❌ 数据库 approved reactions 查询失败: %v", err)
		return nil, err
	}

	// 缓存到 Redis
	if data, err := json.Marshal(reactions); err == nil {
		_ = cache.SetApprovedReactionsCache(ctx, string(data))
	}

	return reactions, nil
}

// GetUserStatsWithCache 使用缓存获取用户统计 - 先查Redis，miss后查DB
func GetUserStatsWithCache(ctx context.Context, uid uint) (map[string]interface{}, error) {
	// 先查缓存
	cached, err := cache.GetUserStatsCache(ctx, uid)
	if err != nil && err.Error() != "redis client not initialized" {
		log.Printf("⚠️  Redis user stats 查询异常: %v", err)
	} else if cached != "" {
		var stats map[string]interface{}
		if err := json.Unmarshal([]byte(cached), &stats); err == nil {
			return stats, nil
		}
	}

	// 缓存 miss，查询数据库
	userRepo := NewUserRepository()
	user, err := userRepo.FindByUID(uid)
	if err != nil || user == nil {
		log.Printf("❌ 数据库 user stats 查询失败 UID=%d: %v", uid, err)
		return nil, err
	}

	// 构建统计数据
	stats := map[string]interface{}{
		"uid":            user.UID,
		"username":       user.Username,
		"nickname":       user.Nickname,
		"avatar":         user.Avatar,
		"level":          user.Level,
		"points":         user.Points,
		"monthly_points": user.MonthlyPoints,
		"total_xp":       user.TotalXP,
		"win_count":      user.WinCount,
		"total_games":    user.TotalGames,
	}

	// 缓存到 Redis
	if data, err := json.Marshal(stats); err == nil {
		_ = cache.SetUserStatsCache(ctx, uid, string(data))
	}

	return stats, nil
}

// InvalidateLeaderboardCacheOnPointsChange 在用户积分变化时失效排行榜缓存
func InvalidateLeaderboardCacheOnPointsChange(ctx context.Context) error {
	// 失效所有排行榜缓存
	return cache.InvalidateAllLeaderboardCache(ctx)
}

// InvalidateBountyTotalCacheOnChange 在赏金变化时失效赏金缓存
func InvalidateBountyTotalCacheOnChange(ctx context.Context, uid uint) error {
	return cache.InvalidateBountyTotalCache(ctx, uid)
}

// InvalidateLevelConfigsCacheOnChange 在等级配置变化时失效缓存
func InvalidateLevelConfigsCacheOnChange(ctx context.Context) error {
	return cache.InvalidateLevelConfigsCache(ctx)
}

// InvalidateApprovedReactionsCacheOnChange 在反应批准状态变化时失效缓存
func InvalidateApprovedReactionsCacheOnChange(ctx context.Context) error {
	return cache.InvalidateApprovedReactionsCache(ctx)
}

// InvalidateUserStatsCacheOnChange 在用户统计变化时失效缓存
func InvalidateUserStatsCacheOnChange(ctx context.Context, uid uint) error {
	return cache.InvalidateUserStatsCache(ctx, uid)
}

// ParseBountyTotal 解析赏金总额字符串为float64
func ParseBountyTotal(bountyStr string) float64 {
	if bountyStr == "" {
		return 0
	}
	amount, err := strconv.ParseFloat(bountyStr, 64)
	if err != nil {
		return 0
	}
	return amount
}
