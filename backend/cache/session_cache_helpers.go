package cache

import (
	"chemistryuno/backend/repository"
	"context"
	"log"
	"time"
)

// ValidateSessionWithCache 使用缓存验证会话 - 先查Redis，miss后查DB并回填缓存
func ValidateSessionWithCache(ctx context.Context, sid string) (bool, error) {
	// 先查缓存
	cached, err := GetSessionCache(ctx, sid)
	if err != nil && err.Error() != "redis client not initialized" {
		log.Printf("⚠️  Redis session 查询异常: %v", err)
		// 继续到数据库查询
	} else if cached != nil {
		// 缓存命中，检查是否过期
		if cached.ExpiresAt > 0 && time.Now().Unix() > cached.ExpiresAt {
			// 已过期，删除缓存并返回 false
			_ = InvalidateSessionCache(ctx, sid)
			return false, nil
		}
		return true, nil
	}

	// 缓存 miss，查询数据库
	sessionRepo := repository.NewSessionRepository()
	exists, err := sessionRepo.Exists(sid)
	if err != nil {
		log.Printf("❌ 数据库 session 查询失败 SID=%s: %v", sid, err)
		return false, err
	}

	if exists {
		// 从数据库加载完整信息，缓存到Redis
		session, err := sessionRepo.FindByID(sid)
		if err == nil && session != nil {
			cacheSess := &SessionCache{
				UserUID:   session.UserUID,
				ExpiresAt: 0, // 无固定过期时间，使用 Redis TTL
				UserAgent: session.UserAgent,
				IPAddress: session.IPAddress,
			}
			_ = SetSessionCache(ctx, sid, cacheSess)
		}
	}

	return exists, nil
}

// ValidateSessionForUserWithCache 使用缓存验证会话属于指定用户
func ValidateSessionForUserWithCache(ctx context.Context, sid string, uid uint) (bool, error) {
	// 先查缓存，看会话用户是否匹配
	cached, err := GetSessionCache(ctx, sid)
	if err != nil && err.Error() != "redis client not initialized" {
		log.Printf("⚠️  Redis session 查询异常: %v", err)
	} else if cached != nil {
		// 缓存命中，直接验证
		if cached.ExpiresAt > 0 && time.Now().Unix() > cached.ExpiresAt {
			_ = InvalidateSessionCache(ctx, sid)
			return false, nil
		}
		return cached.UserUID == uid, nil
	}

	// 缓存 miss，查询数据库
	sessionRepo := repository.NewSessionRepository()
	valid, err := sessionRepo.ValidateSessionForUser(sid, uid)
	if err != nil {
		log.Printf("❌ 数据库会话验证失败 SID=%s, UID=%d: %v", sid, uid, err)
		return false, err
	}

	if valid {
		// 从数据库加载信息，缓存到Redis
		session, err := sessionRepo.FindByID(sid)
		if err == nil && session != nil {
			cacheSess := &SessionCache{
				UserUID:   session.UserUID,
				ExpiresAt: 0,
				UserAgent: session.UserAgent,
				IPAddress: session.IPAddress,
			}
			_ = SetSessionCache(ctx, sid, cacheSess)
		}
	}

	return valid, nil
}

// GetUserWithCache 使用缓存获取用户信息 - 先查Redis，miss后查DB
func GetUserWithCache(ctx context.Context, uid uint) (*UserCache, error) {
	// 先查缓存
	cached, err := GetUserCache(ctx, uid)
	if err != nil && err.Error() != "redis client not initialized" {
		log.Printf("⚠️  Redis user 查询异常: %v", err)
	} else if cached != nil {
		return cached, nil
	}

	// 缓存 miss，查询数据库
	userRepo := repository.NewUserRepository()
	dbUser, err := userRepo.FindByUID(uid)
	if err != nil {
		log.Printf("❌ 数据库用户查询失败 UID=%d: %v", uid, err)
		return nil, err
	}

	// 缓存到Redis
	cacheUser := &UserCache{
		UID:         dbUser.UID,
		Username:    dbUser.Username,
		Email:       dbUser.Email,
		IsAdmin:     dbUser.IsAdmin,
		Role:        dbUser.Role,
		BannedUntil: dbUser.BannedUntil,
		FrozenUntil: dbUser.FrozenUntil,
		BanReason:   dbUser.BanReason,
	}
	_ = SetUserCache(ctx, cacheUser)

	return cacheUser, nil
}
