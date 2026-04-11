package middleware

import (
	"chemistryuno/backend/repository"
	"chemistryuno/backend/utils"
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// 优先从查询参数获取token（用于WebSocket的向后兼容）
		token = c.Query("token")

		// 如果查询参数中没有，则从Cookie获取
		if token == "" {
			token, _ = c.Cookie("access_token")
		}

		// 如果Cookie也没有，则从Authorization头获取
		if token == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证信息"})
				c.Abort()
				return
			}

			// Bearer token格式
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "认证格式错误"})
				c.Abort()
				return
			}
			token = parts[1]
		}

		claims, err := utils.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
			c.Abort()
			return
		}

		// 检查token类型（API调用只接受access token）
		if claims.TokenType != "access" && claims.TokenType != "" {
			// TokenType为空表示旧版token（兼容），为"refresh"表示刷新令牌不能用于API调用
			if claims.TokenType == "refresh" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "刷新令牌不能用于API调用，请刷新后重试"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token类型"})
			}
			c.Abort()
			return
		}

		// 强制要求 SID 存在（防止旧版 Token 或非法 Token 绕过会话检查）
		if claims.SID == "" {
			log.Printf("[强制踢出] Token中缺少SID: UID=%d, IP=%s", claims.UID, c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证信息不完整，请重新登录"})
			c.Abort()
			return
		}

		// AI账号（UID < 0）的会话永不过期，跳过会话验证
		if claims.UID < 0 {
			// 将用户信息存入上下文
			c.Set("uid", claims.UID)
			c.Set("is_admin", claims.IsAdmin)
			c.Set("role", claims.Role)
			c.Next()
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		// 使用缓存验证会话 - 先查Redis，miss后查DB
		sessionValid, err := repository.ValidateSessionWithCache(ctx, claims.SID)
		if err != nil {
			log.Printf("❌ 会话验证失败 SID=%s: %v", claims.SID, err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "会话验证失败"})
			c.Abort()
			return
		}
		if !sessionValid {
			log.Printf("[会话失效] UID=%d, SID=%s, IP=%s", claims.UID, claims.SID, c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "会话已过期或在其他设备登出"})
			c.Abort()
			return
		}

		// 验证会话是否属于该用户（防止会话劫持）
		userValid, err := repository.ValidateSessionForUserWithCache(ctx, claims.SID, uint(claims.UID))
		if err != nil {
			log.Printf("❌ 会话用户验证失败 SID=%s, UID=%d: %v", claims.SID, claims.UID, err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "会话验证失败"})
			c.Abort()
			return
		}
		if !userValid {
			log.Printf("[会话验证失败] UID=%d, SID=%s, IP=%s", claims.UID, claims.SID, c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "会话验证失败"})
			c.Abort()
			return
		}

		// 更新活动时间及当前访问 IP（异步，不阻塞）
		go utils.UpdateSessionActivity(claims.SID, c.ClientIP())

		// 将用户信息存入上下文
		c.Set("uid", claims.UID)
		c.Set("email", claims.Email)
		c.Set("is_admin", claims.IsAdmin)
		c.Set("role", claims.Role)
		c.Set("sid", claims.SID)

		// 检查账号冻结/封禁状态 - 使用缓存
		cachedUser, err := repository.GetUserWithCache(ctx, uint(claims.UID))
		if err != nil {
			// 降级到旧方式
			log.Printf("⚠️  缓存查询失败，降级到数据库: %v", err)
			userRepo := repository.NewUserRepository()
			bannedUntil, frozenUntil, reason, err := userRepo.CheckBanStatus(uint(claims.UID))
			if err == nil {
				now := time.Now()
				if bannedUntil != nil && bannedUntil.After(now) {
					msg := "账号已被封禁"
					if reason != "" {
						msg = reason
					}
					c.JSON(http.StatusForbidden, gin.H{"error": msg + " (截至 " + bannedUntil.Format("2006-01-02 15:04:05") + ")"})
					c.Abort()
					return
				}
				if frozenUntil != nil && frozenUntil.After(now) {
					c.JSON(http.StatusForbidden, gin.H{"error": "账号已被冻结至 " + frozenUntil.Format("2006-01-02 15:04:05")})
					c.Abort()
					return
				}
			}
		} else if cachedUser != nil {
			// 使用缓存的用户信息
			now := time.Now()
			if cachedUser.BannedUntil != nil && cachedUser.BannedUntil.After(now) {
				msg := "账号已被封禁"
				if cachedUser.BanReason != "" {
					msg = cachedUser.BanReason
				}
				c.JSON(http.StatusForbidden, gin.H{"error": msg + " (截至 " + cachedUser.BannedUntil.Format("2006-01-02 15:04:05") + ")"})
				c.Abort()
				return
			}
			if cachedUser.FrozenUntil != nil && cachedUser.FrozenUntil.After(now) {
				c.JSON(http.StatusForbidden, gin.H{"error": "账号已被冻结至 " + cachedUser.FrozenUntil.Format("2006-01-02 15:04:05")})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// 管理员权限中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Co-worker权限中间件（co-worker或admin）
func CoWorkerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
			c.Abort()
			return
		}

		roleStr := role.(string)
		if roleStr != "admin" && roleStr != "co-worker" {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要co-worker或管理员权限"})
			c.Abort()
			return
		}
		c.Next()
	}
}
