package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

// SessionCache 缓存的会话信息
type SessionCache struct {
	UserUID   uint   `json:"user_uid"`
	ExpiresAt int64  `json:"expires_at"`
	UserAgent string `json:"user_agent,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
}

// UserCache 缓存的用户信息（用于中间件）
type UserCache struct {
	UID         uint       `json:"uid"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	IsAdmin     bool       `json:"is_admin"`
	Role        string     `json:"role"`
	BannedUntil *time.Time `json:"banned_until,omitempty"`
	FrozenUntil *time.Time `json:"frozen_until,omitempty"`
	BanReason   string     `json:"ban_reason,omitempty"`
}

// InitRedis 初始化 Redis 连接
func InitRedis(addr string) error {
	redisClient = redis.NewClient(&redis.Options{
		Addr:         addr,
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   2,
		PoolTimeout:  4 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("❌ Redis 连接失败: %v", err)
		return err
	}

	log.Println("✅ Redis 连接成功")
	return nil
}

// GetRedisClient 获取 Redis 客户端
func GetRedisClient() *redis.Client {
	return redisClient
}

// sessionKey 生成会话缓存 key
func sessionKey(sid string) string {
	return fmt.Sprintf("session:%s", sid)
}

// userKey 生成用户信息缓存 key
func userKey(uid uint) string {
	return fmt.Sprintf("user:%d", uid)
}

// reactionKey 生成反应验证缓存 key，使用规范化的 r1, r2
func reactionKey(r1, r2 string) string {
	if r1 > r2 {
		r1, r2 = r2, r1
	}
	return fmt.Sprintf("reaction:%s:%s", r1, r2)
}

// GetReactionCache 从缓存获取反应验证结果 (1 为存在, 0 为不存在, 空为 miss)
func GetReactionCache(ctx context.Context, r1, r2 string) (string, error) {
	if redisClient == nil {
		return "", fmt.Errorf("redis client not initialized")
	}

	val, err := redisClient.Get(ctx, reactionKey(r1, r2)).Result()
	if err == redis.Nil {
		return "", nil // Cache miss
	}
	return val, err
}

// SetReactionCache 将反应验证结果缓存 1 小时 (status=1 存在, status=0 不存在)
func SetReactionCache(ctx context.Context, r1, r2 string, exists bool) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	val := "0"
	if exists {
		val = "1"
	}

	return redisClient.Set(ctx, reactionKey(r1, r2), val, 1*time.Hour).Err()
}

// InvalidateReactionCache 失效特定的反应缓存
func InvalidateReactionCache(ctx context.Context, r1, r2 string) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return redisClient.Del(ctx, reactionKey(r1, r2)).Err()
}

// GetSessionCache 从缓存获取会话 (先 Redis，miss 则返回 nil)
func GetSessionCache(ctx context.Context, sid string) (*SessionCache, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	val, err := redisClient.Get(ctx, sessionKey(sid)).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		log.Printf("❌ Redis session 查询失败 SID=%s: %v", sid, err)
		return nil, err
	}

	var sess SessionCache
	if err := json.Unmarshal([]byte(val), &sess); err != nil {
		log.Printf("❌ 解析 session 缓存失败: %v", err)
		return nil, err
	}

	return &sess, nil
}

// SetSessionCache 将会话缓存到 Redis，TTL 为 24 小时
func SetSessionCache(ctx context.Context, sid string, session *SessionCache) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	// TTL: 24 小时
	ttl := 24 * time.Hour
	if session.ExpiresAt > 0 {
		expireTime := time.Unix(session.ExpiresAt, 0)
		ttl = time.Until(expireTime)
		if ttl < 0 {
			ttl = 1 * time.Minute // 已过期，设置 1 分钟后自动清理
		}
	}

	if err := redisClient.Set(ctx, sessionKey(sid), data, ttl).Err(); err != nil {
		log.Printf("❌ Redis session 缓存设置失败 SID=%s: %v", sid, err)
		return err
	}

	return nil
}

// InvalidateSessionCache 删除会话缓存
func InvalidateSessionCache(ctx context.Context, sid string) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	if err := redisClient.Del(ctx, sessionKey(sid)).Err(); err != nil {
		log.Printf("⚠️  删除 session 缓存失败 SID=%s: %v", sid, err)
		return err
	}

	return nil
}

// GetUserCache 从缓存获取用户信息 (用于减少中间件查询)
func GetUserCache(ctx context.Context, uid uint) (*UserCache, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	val, err := redisClient.Get(ctx, userKey(uid)).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		log.Printf("❌ Redis user 查询失败 UID=%d: %v", uid, err)
		return nil, err
	}

	var user UserCache
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		log.Printf("❌ 解析 user 缓存失败: %v", err)
		return nil, err
	}

	return &user, nil
}

// SetUserCache 将用户信息缓存到 Redis，TTL 为 5 分钟（考虑到封禁/冻结可能变化）
func SetUserCache(ctx context.Context, user *UserCache) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	// TTL: 5 分钟（用户状态变化相对频繁）
	ttl := 5 * time.Minute

	if err := redisClient.Set(ctx, userKey(user.UID), data, ttl).Err(); err != nil {
		log.Printf("❌ Redis user 缓存设置失败 UID=%d: %v", user.UID, err)
		return err
	}

	return nil
}

// InvalidateUserCache 删除用户缓存
func InvalidateUserCache(ctx context.Context, uid uint) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	if err := redisClient.Del(ctx, userKey(uid)).Err(); err != nil {
		log.Printf("⚠️  删除 user 缓存失败 UID=%d: %v", uid, err)
		return err
	}

	return nil
}

// InvalidateUserCacheMultiple 批量删除用户缓存
func InvalidateUserCacheMultiple(ctx context.Context, uids ...uint) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	keys := make([]string, len(uids))
	for i, uid := range uids {
		keys[i] = userKey(uid)
	}

	if err := redisClient.Del(ctx, keys...).Err(); err != nil {
		log.Printf("⚠️  批量删除 user 缓存失败: %v", err)
		return err
	}

	return nil
}

// validLeaderboardOrderBy 验证排行榜排序方式白名单
func validLeaderboardOrderBy(orderBy string) bool {
	validValues := map[string]bool{
		"points":         true,
		"monthly_points": true,
		"total_xp":       true,
		"win_count":      true,
		"total_games":    true,
		"created_at":     true,
		"uid":            true,
	}
	return validValues[orderBy]
}

// leaderboardKey 生成排行榜缓存 key
func leaderboardKey(orderBy string, limit int) string {
	if !validLeaderboardOrderBy(orderBy) {
		log.Printf("⚠️  非法的排行榜排序方式: %s, 使用默认值 'points'", orderBy)
		orderBy = "points"
	}
	return fmt.Sprintf("leaderboard:%s:%d", orderBy, limit)
}

// GetLeaderboardCache 从缓存获取排行榜
func GetLeaderboardCache(ctx context.Context, orderBy string, limit int) (string, error) {
	if redisClient == nil {
		return "", fmt.Errorf("redis client not initialized")
	}

	val, err := redisClient.Get(ctx, leaderboardKey(orderBy, limit)).Result()
	if err == redis.Nil {
		return "", nil // Cache miss
	}
	return val, err
}

// SetLeaderboardCache 将排行榜缓存到 Redis，TTL 为 5 分钟
func SetLeaderboardCache(ctx context.Context, orderBy string, limit int, data string) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	ttl := 5 * time.Minute
	if err := redisClient.Set(ctx, leaderboardKey(orderBy, limit), data, ttl).Err(); err != nil {
		log.Printf("⚠️  Redis leaderboard 缓存设置失败 orderBy=%s, limit=%d: %v", orderBy, limit, err)
		return err
	}

	return nil
}

// InvalidateLeaderboardCache 删除排行榜缓存
func InvalidateLeaderboardCache(ctx context.Context, orderBy string) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	// 验证排序方式白名单
	if !validLeaderboardOrderBy(orderBy) {
		log.Printf("⚠️  尝试使用非法的排行榜排序方式进行缓存失效: %s", orderBy)
		return fmt.Errorf("invalid leaderboard orderBy value: %s", orderBy)
	}

	// 删除所有该排序方式的排行榜缓存（支持多个limit值）
	pattern := fmt.Sprintf("leaderboard:%s:*", orderBy)
	iter := redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if len(keys) > 0 {
		if err := redisClient.Del(ctx, keys...).Err(); err != nil {
			log.Printf("⚠️  删除 leaderboard 缓存失败: %v", err)
			return err
		}
	}

	return nil
}

// bountyTotalKey 生成赏金总额缓存 key
func bountyTotalKey(uid uint) string {
	return fmt.Sprintf("bounty:total:%d", uid)
}

// GetBountyTotalCache 从缓存获取用户赏金总额
func GetBountyTotalCache(ctx context.Context, uid uint) (string, error) {
	if redisClient == nil {
		return "", fmt.Errorf("redis client not initialized")
	}

	val, err := redisClient.Get(ctx, bountyTotalKey(uid)).Result()
	if err == redis.Nil {
		return "", nil // Cache miss
	}
	return val, err
}

// SetBountyTotalCache 将赏金总额缓存到 Redis，TTL 为 10 分钟
func SetBountyTotalCache(ctx context.Context, uid uint, amount float64) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	ttl := 10 * time.Minute
	if err := redisClient.Set(ctx, bountyTotalKey(uid), fmt.Sprintf("%.2f", amount), ttl).Err(); err != nil {
		log.Printf("⚠️  Redis bounty 缓存设置失败 UID=%d: %v", uid, err)
		return err
	}

	return nil
}

// InvalidateBountyTotalCache 删除用户赏金缓存
func InvalidateBountyTotalCache(ctx context.Context, uid uint) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	if err := redisClient.Del(ctx, bountyTotalKey(uid)).Err(); err != nil {
		log.Printf("⚠️  删除 bounty 缓存失败 UID=%d: %v", uid, err)
		return err
	}

	return nil
}

// InvalidateBountyTotalCacheMultiple 批量删除赏金缓存
func InvalidateBountyTotalCacheMultiple(ctx context.Context, uids ...uint) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	keys := make([]string, len(uids))
	for i, uid := range uids {
		keys[i] = bountyTotalKey(uid)
	}

	if err := redisClient.Del(ctx, keys...).Err(); err != nil {
		log.Printf("⚠️  批量删除 bounty 缓存失败: %v", err)
		return err
	}

	return nil
}

// levelConfigsKey 生成等级配置缓存 key
func levelConfigsKey() string {
	return "level:configs"
}

// GetLevelConfigsCache 从缓存获取等级配置
func GetLevelConfigsCache(ctx context.Context) (string, error) {
	if redisClient == nil {
		return "", fmt.Errorf("redis client not initialized")
	}

	val, err := redisClient.Get(ctx, levelConfigsKey()).Result()
	if err == redis.Nil {
		return "", nil // Cache miss
	}
	return val, err
}

// SetLevelConfigsCache 将等级配置缓存到 Redis，TTL 为 24 小时
func SetLevelConfigsCache(ctx context.Context, data string) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	ttl := 24 * time.Hour
	if err := redisClient.Set(ctx, levelConfigsKey(), data, ttl).Err(); err != nil {
		log.Printf("⚠️  Redis level configs 缓存设置失败: %v", err)
		return err
	}

	return nil
}

// InvalidateLevelConfigsCache 删除等级配置缓存
func InvalidateLevelConfigsCache(ctx context.Context) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	if err := redisClient.Del(ctx, levelConfigsKey()).Err(); err != nil {
		log.Printf("⚠️  删除 level configs 缓存失败: %v", err)
		return err
	}

	return nil
}

// approvedReactionsKey 生成已批准反应缓存 key
func approvedReactionsKey() string {
	return "reactions:approved:list"
}

// GetApprovedReactionsCache 从缓存获取已批准的反应
func GetApprovedReactionsCache(ctx context.Context) (string, error) {
	if redisClient == nil {
		return "", fmt.Errorf("redis client not initialized")
	}

	val, err := redisClient.Get(ctx, approvedReactionsKey()).Result()
	if err == redis.Nil {
		return "", nil // Cache miss
	}
	return val, err
}

// SetApprovedReactionsCache 将已批准反应缓存到 Redis，TTL 为 1 小时
func SetApprovedReactionsCache(ctx context.Context, data string) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	ttl := 1 * time.Hour
	if err := redisClient.Set(ctx, approvedReactionsKey(), data, ttl).Err(); err != nil {
		log.Printf("⚠️  Redis approved reactions 缓存设置失败: %v", err)
		return err
	}

	return nil
}

// InvalidateApprovedReactionsCache 删除已批准反应缓存
func InvalidateApprovedReactionsCache(ctx context.Context) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	if err := redisClient.Del(ctx, approvedReactionsKey()).Err(); err != nil {
		log.Printf("⚠️  删除 approved reactions 缓存失败: %v", err)
		return err
	}

	return nil
}

// userStatsKey 生成用户统计缓存 key
func userStatsKey(uid uint) string {
	return fmt.Sprintf("user:stats:%d", uid)
}

// GetUserStatsCache 从缓存获取用户统计
func GetUserStatsCache(ctx context.Context, uid uint) (string, error) {
	if redisClient == nil {
		return "", fmt.Errorf("redis client not initialized")
	}

	val, err := redisClient.Get(ctx, userStatsKey(uid)).Result()
	if err == redis.Nil {
		return "", nil // Cache miss
	}
	return val, err
}

// SetUserStatsCache 将用户统计缓存到 Redis，TTL 为 15 分钟
func SetUserStatsCache(ctx context.Context, uid uint, data string) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	ttl := 15 * time.Minute
	if err := redisClient.Set(ctx, userStatsKey(uid), data, ttl).Err(); err != nil {
		log.Printf("⚠️  Redis user stats 缓存设置失败 UID=%d: %v", uid, err)
		return err
	}

	return nil
}

// InvalidateUserStatsCache 删除用户统计缓存
func InvalidateUserStatsCache(ctx context.Context, uid uint) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	if err := redisClient.Del(ctx, userStatsKey(uid)).Err(); err != nil {
		log.Printf("⚠️  删除 user stats 缓存失败 UID=%d: %v", uid, err)
		return err
	}

	return nil
}

// InvalidateUserStatsCacheMultiple 批量删除用户统计缓存
func InvalidateUserStatsCacheMultiple(ctx context.Context, uids ...uint) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	keys := make([]string, len(uids))
	for i, uid := range uids {
		keys[i] = userStatsKey(uid)
	}

	if err := redisClient.Del(ctx, keys...).Err(); err != nil {
		log.Printf("⚠️  批量删除 user stats 缓存失败: %v", err)
		return err
	}

	return nil
}

// LeaderboardCache 缓存的排行榜数据
type LeaderboardCache struct {
	UID           uint    `json:"uid"`
	Username      string  `json:"username"`
	Nickname      string  `json:"nickname"`
	Avatar        string  `json:"avatar"`
	Points        float64 `json:"points"`
	MonthlyPoints float64 `json:"monthly_points"`
	Level         int     `json:"level"`
	TotalXP       int     `json:"total_xp"`
	WinCount      int     `json:"win_count"`
	TotalGames    int     `json:"total_games"`
}

// leaderboardKey 生成排行榜缓存 key
func leaderboardKey(orderBy string, limit int) string {
	return fmt.Sprintf("leaderboard:%s:%d", orderBy, limit)
}

// GetLeaderboardCache 从缓存获取排行榜
func GetLeaderboardCache(ctx context.Context, orderBy string, limit int) ([]LeaderboardCache, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	val, err := redisClient.Get(ctx, leaderboardKey(orderBy, limit)).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		log.Printf("❌ Redis leaderboard 查询失败 orderBy=%s: %v", orderBy, err)
		return nil, err
	}

	var leaderboard []LeaderboardCache
	if err := json.Unmarshal([]byte(val), &leaderboard); err != nil {
		log.Printf("❌ 解析 leaderboard 缓存失败: %v", err)
		return nil, err
	}

	return leaderboard, nil
}

// SetLeaderboardCache 将排行榜缓存到 Redis，TTL 为 5 分钟
func SetLeaderboardCache(ctx context.Context, orderBy string, limit int, leaderboard []LeaderboardCache) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	data, err := json.Marshal(leaderboard)
	if err != nil {
		return err
	}

	ttl := 5 * time.Minute
	if err := redisClient.Set(ctx, leaderboardKey(orderBy, limit), data, ttl).Err(); err != nil {
		log.Printf("❌ Redis leaderboard 缓存设置失败 orderBy=%s: %v", orderBy, err)
		return err
	}

	return nil
}

// InvalidateLeaderboardCache 删除排行榜缓存
func InvalidateLeaderboardCache(ctx context.Context, orderBy string) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	// 删除所有该排序方式的排行榜缓存（支持多个limit值）
	pattern := fmt.Sprintf("leaderboard:%s:*", orderBy)
	iter := redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if len(keys) > 0 {
		if err := redisClient.Del(ctx, keys...).Err(); err != nil {
			log.Printf("⚠️  删除 leaderboard 缓存失败: %v", err)
			return err
		}
	}

	return nil
}

// InvalidateAllLeaderboardCache 删除所有排行榜缓存
func InvalidateAllLeaderboardCache(ctx context.Context) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	pattern := "leaderboard:*"
	iter := redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if len(keys) > 0 {
		if err := redisClient.Del(ctx, keys...).Err(); err != nil {
			log.Printf("⚠️  删除所有 leaderboard 缓存失败: %v", err)
			return err
		}
	}

	return nil
}

// bountyTotalKey 生成赏金总额缓存 key
func bountyTotalKey(uid uint) string {
	return fmt.Sprintf("bounty:total:%d", uid)
}

// GetBountyTotalCache 从缓存获取赏金总额
func GetBountyTotalCache(ctx context.Context, uid uint) (float64, error) {
	if redisClient == nil {
		return 0, fmt.Errorf("redis client not initialized")
	}

	val, err := redisClient.Get(ctx, bountyTotalKey(uid)).Result()
	if err == redis.Nil {
		return 0, nil // Cache miss
	}
	if err != nil {
		log.Printf("❌ Redis bounty 查询失败 UID=%d: %v", uid, err)
		return 0, err
	}

	amount, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Printf("❌ 解析 bounty 缓存失败: %v", err)
		return 0, err
	}

	return amount, nil
}

// SetBountyTotalCache 将赏金总额缓存到 Redis，TTL 为 10 分钟
func SetBountyTotalCache(ctx context.Context, uid uint, amount float64) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	ttl := 10 * time.Minute
	if err := redisClient.Set(ctx, bountyTotalKey(uid), fmt.Sprintf("%.2f", amount), ttl).Err(); err != nil {
		log.Printf("❌ Redis bounty 缓存设置失败 UID=%d: %v", uid, err)
		return err
	}

	return nil
}

// InvalidateBountyTotalCache 删除赏金缓存
func InvalidateBountyTotalCache(ctx context.Context, uid uint) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	if err := redisClient.Del(ctx, bountyTotalKey(uid)).Err(); err != nil {
		log.Printf("⚠️  删除 bounty 缓存失败 UID=%d: %v", uid, err)
		return err
	}

	return nil
}

// InvalidateBountyTotalCacheMultiple 批量删除赏金缓存
func InvalidateBountyTotalCacheMultiple(ctx context.Context, uids ...uint) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	keys := make([]string, len(uids))
	for i, uid := range uids {
		keys[i] = bountyTotalKey(uid)
	}

	if err := redisClient.Del(ctx, keys...).Err(); err != nil {
		log.Printf("⚠️  批量删除 bounty 缓存失败: %v", err)
		return err
	}

	return nil
}

// levelConfigsKey 生成等级配置缓存 key
func levelConfigsKey() string {
	return "level:configs"
}

// GetLevelConfigsCache 从缓存获取等级配置
func GetLevelConfigsCache(ctx context.Context) (string, error) {
	if redisClient == nil {
		return "", fmt.Errorf("redis client not initialized")
	}

	val, err := redisClient.Get(ctx, levelConfigsKey()).Result()
	if err == redis.Nil {
		return "", nil // Cache miss
	}
	if err != nil {
		log.Printf("❌ Redis level configs 查询失败: %v", err)
		return "", err
	}

	return val, nil
}

// SetLevelConfigsCache 将等级配置缓存到 Redis，TTL 为 24 小时
func SetLevelConfigsCache(ctx context.Context, data string) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	ttl := 24 * time.Hour
	if err := redisClient.Set(ctx, levelConfigsKey(), data, ttl).Err(); err != nil {
		log.Printf("❌ Redis level configs 缓存设置失败: %v", err)
		return err
	}

	return nil
}

// InvalidateLevelConfigsCache 删除等级配置缓存
func InvalidateLevelConfigsCache(ctx context.Context) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	if err := redisClient.Del(ctx, levelConfigsKey()).Err(); err != nil {
		log.Printf("⚠️  删除 level configs 缓存失败: %v", err)
		return err
	}

	return nil
}

// approvedReactionsKey 生成已批准反应缓存 key
func approvedReactionsKey() string {
	return "reactions:approved:list"
}

// GetApprovedReactionsCache 从缓存获取已批准反应
func GetApprovedReactionsCache(ctx context.Context) (string, error) {
	if redisClient == nil {
		return "", fmt.Errorf("redis client not initialized")
	}

	val, err := redisClient.Get(ctx, approvedReactionsKey()).Result()
	if err == redis.Nil {
		return "", nil // Cache miss
	}
	if err != nil {
		log.Printf("❌ Redis approved reactions 查询失败: %v", err)
		return "", err
	}

	return val, nil
}

// SetApprovedReactionsCache 将已批准反应缓存到 Redis，TTL 为 1 小时
func SetApprovedReactionsCache(ctx context.Context, data string) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	ttl := 1 * time.Hour
	if err := redisClient.Set(ctx, approvedReactionsKey(), data, ttl).Err(); err != nil {
		log.Printf("❌ Redis approved reactions 缓存设置失败: %v", err)
		return err
	}

	return nil
}

// InvalidateApprovedReactionsCache 删除已批准反应缓存
func InvalidateApprovedReactionsCache(ctx context.Context) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	if err := redisClient.Del(ctx, approvedReactionsKey()).Err(); err != nil {
		log.Printf("⚠️  删除 approved reactions 缓存失败: %v", err)
		return err
	}

	return nil
}

// userStatsKey 生成用户统计缓存 key
func userStatsKey(uid uint) string {
	return fmt.Sprintf("user:stats:%d", uid)
}

// GetUserStatsCache 从缓存获取用户统计
func GetUserStatsCache(ctx context.Context, uid uint) (string, error) {
	if redisClient == nil {
		return "", fmt.Errorf("redis client not initialized")
	}

	val, err := redisClient.Get(ctx, userStatsKey(uid)).Result()
	if err == redis.Nil {
		return "", nil // Cache miss
	}
	if err != nil {
		log.Printf("❌ Redis user stats 查询失败 UID=%d: %v", uid, err)
		return "", err
	}

	return val, nil
}

// SetUserStatsCache 将用户统计缓存到 Redis，TTL 为 15 分钟
func SetUserStatsCache(ctx context.Context, uid uint, data string) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	ttl := 15 * time.Minute
	if err := redisClient.Set(ctx, userStatsKey(uid), data, ttl).Err(); err != nil {
		log.Printf("❌ Redis user stats 缓存设置失败 UID=%d: %v", uid, err)
		return err
	}

	return nil
}

// InvalidateUserStatsCache 删除用户统计缓存
func InvalidateUserStatsCache(ctx context.Context, uid uint) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	if err := redisClient.Del(ctx, userStatsKey(uid)).Err(); err != nil {
		log.Printf("⚠️  删除 user stats 缓存失败 UID=%d: %v", uid, err)
		return err
	}

	return nil
}

// InvalidateUserStatsCacheMultiple 批量删除用户统计缓存
func InvalidateUserStatsCacheMultiple(ctx context.Context, uids ...uint) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	keys := make([]string, len(uids))
	for i, uid := range uids {
		keys[i] = userStatsKey(uid)
	}

	if err := redisClient.Del(ctx, keys...).Err(); err != nil {
		log.Printf("⚠️  批量删除 user stats 缓存失败: %v", err)
		return err
	}

	return nil
}

// Close 关闭 Redis 连接
func Close() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}
