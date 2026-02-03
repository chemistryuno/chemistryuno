package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

var cacheCtx = context.Background()

// CacheKey 生成缓存键
func CacheKey(prefix string, id interface{}) string {
	return fmt.Sprintf("%s:%v", prefix, id)
}

// SetCache 设置缓存（带过期时间）
func SetCache(key string, value interface{}, expiration time.Duration) error {
	if RedisClient == nil {
		return nil // Redis不可用，跳过缓存
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return RedisClient.Set(cacheCtx, key, data, expiration).Err()
}

// GetCache 获取缓存
func GetCache(key string, dest interface{}) error {
	if RedisClient == nil {
		return fmt.Errorf("redis not available")
	}

	data, err := RedisClient.Get(cacheCtx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// DeleteCache 删除缓存
func DeleteCache(keys ...string) error {
	if RedisClient == nil {
		return nil
	}

	return RedisClient.Del(cacheCtx, keys...).Err()
}

// DeleteCacheByPattern 根据模式删除缓存
func DeleteCacheByPattern(pattern string) error {
	if RedisClient == nil {
		return nil
	}

	iter := RedisClient.Scan(cacheCtx, 0, pattern, 0).Iterator()
	var keys []string

	for iter.Next(cacheCtx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return RedisClient.Del(cacheCtx, keys...).Err()
	}

	return nil
}

// GetUserByIDWithCache 通过ID获取用户（带缓存）
func GetUserByIDWithCache(uid uint) (*User, error) {
	cacheKey := CacheKey("user", uid)

	// 尝试从缓存获取
	var user User
	if err := GetCache(cacheKey, &user); err == nil {
		return &user, nil
	}

	// 缓存未命中，从数据库查询
	if err := DB.First(&user, uid).Error; err != nil {
		return nil, err
	}

	// 写入缓存（10分钟过期）
	SetCache(cacheKey, &user, 10*time.Minute)

	return &user, nil
}

// GetUserByUsernameWithCache 通过用户名获取用户（带缓存）
func GetUserByUsernameWithCache(username string) (*User, error) {
	cacheKey := CacheKey("user:username", username)

	// 尝试从缓存获取
	var user User
	if err := GetCache(cacheKey, &user); err == nil {
		return &user, nil
	}

	// 缓存未命中，从数据库查询
	if err := DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}

	// 写入缓存（10分钟过期）
	SetCache(cacheKey, &user, 10*time.Minute)

	return &user, nil
}

// InvalidateUserCache 使用户缓存失效
func InvalidateUserCache(uid uint, username string) {
	DeleteCache(
		CacheKey("user", uid),
		CacheKey("user:username", username),
	)
}

// GetReactionsWithCache 获取所有反应（带缓存）
func GetReactionsWithCache(status string) ([]Reaction, error) {
	cacheKey := CacheKey("reactions", status)

	// 尝试从缓存获取
	var reactions []Reaction
	if err := GetCache(cacheKey, &reactions); err == nil {
		return reactions, nil
	}

	// 缓存未命中，从数据库查询
	query := DB.Model(&Reaction{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&reactions).Error; err != nil {
		return nil, err
	}

	// 写入缓存（5分钟过期）
	SetCache(cacheKey, reactions, 5*time.Minute)

	return reactions, nil
}

// InvalidateReactionsCache 使反应缓存失效
func InvalidateReactionsCache() {
	DeleteCacheByPattern("reactions:*")
}

// GetSubstancesWithCache 获取所有物质（带缓存）
func GetSubstancesWithCache(status string) ([]Substance, error) {
	cacheKey := CacheKey("substances", status)

	// 尝试从缓存获取
	var substances []Substance
	if err := GetCache(cacheKey, &substances); err == nil {
		return substances, nil
	}

	// 缓存未命中，从数据库查询
	query := DB.Model(&Substance{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&substances).Error; err != nil {
		return nil, err
	}

	// 写入缓存（5分钟过期）
	SetCache(cacheKey, substances, 5*time.Minute)

	return substances, nil
}

// InvalidateSubstancesCache 使物质缓存失效
func InvalidateSubstancesCache() {
	DeleteCacheByPattern("substances:*")
}

// IncrementCounter 原子递增计数器（用于限流等场景）
func IncrementCounter(key string, expiration time.Duration) (int64, error) {
	if RedisClient == nil {
		return 0, fmt.Errorf("redis not available")
	}

	pipe := RedisClient.TxPipeline()
	incr := pipe.Incr(cacheCtx, key)
	pipe.Expire(cacheCtx, key, expiration)

	if _, err := pipe.Exec(cacheCtx); err != nil {
		return 0, err
	}

	return incr.Val(), nil
}

// CheckRateLimit 检查速率限制
func CheckRateLimit(key string, limit int64, window time.Duration) (bool, error) {
	if RedisClient == nil {
		return true, nil // Redis不可用时不限流
	}

	count, err := IncrementCounter(key, window)
	if err != nil {
		return true, err
	}

	return count <= limit, nil
}
