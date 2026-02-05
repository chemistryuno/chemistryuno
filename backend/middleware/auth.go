package middleware

import (
	"chemistryuno/repository"
	"chemistryuno/utils"
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

		// 优先从查询参数获取token（用于WebSocket）
		token = c.Query("token")

		// 如果查询参数中没有，则从Authorization头获取
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

		// 强制要求 SID 存在（防止旧版 Token 或非法 Token 绕过会话检查）
		if claims.SID == "" {
			log.Printf("[强制踢出] Token中缺少SID: UID=%d, Username=%s, IP=%s", claims.UID, claims.Username, c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证信息不完整，请重新登录"})
			c.Abort()
			return
		}

		// 验证会话是否依然有效
		if !utils.IsSessionValid(claims.SID) {
			log.Printf("[会话失效] UID=%d, SID=%s, IP=%s", claims.UID, claims.SID, c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "会话已过期或在其他设备登出"})
			c.Abort()
			return
		}

		// 验证会话是否属于该用户（防止会话劫持）
		if !utils.ValidateSessionForUser(claims.SID, claims.UID) {
			log.Printf("[会话验证失败] UID=%d, SID=%s, IP=%s", claims.UID, claims.SID, c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "会话验证失败"})
			c.Abort()
			return
		}

		// 更新活动时间及当前访问 IP
		utils.UpdateSessionActivity(claims.SID, c.ClientIP())

		// 将用户信息存入上下文
		c.Set("uid", claims.UID)
		c.Set("username", claims.Username)
		c.Set("is_admin", claims.IsAdmin)
		c.Set("role", claims.Role)
		c.Set("sid", claims.SID)

		// 检查账号冻结/封禁状态（使用Repository）
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
