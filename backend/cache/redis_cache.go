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

// Close 关闭 Redis 连接
func Close() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}
