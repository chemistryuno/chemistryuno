package middleware

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitStore 存储请求频率限制的数据
type RateLimitStore struct {
	mu       sync.RWMutex
	requests map[string][]time.Time // key: 标识符 (IP+endpoint), value: 请求时间戳列表
}

var rateLimitStore = &RateLimitStore{
	requests: make(map[string][]time.Time),
}

// 全局异常登录检测器实例
var globalAnomalousLoginDetector *AnomalousLoginDetector

// NewAnomalousLoginDetector 创建新的异常登录检测器
func NewAnomalousLoginDetector() *AnomalousLoginDetector {
	return &AnomalousLoginDetector{
		lastLogins:        make(map[int]LoginInfo),
		multiFailedLogins: make(map[string]int),
	}
}

// GetGlobalAnomalousLoginDetector 获取全局异常登录检测器实例
func GetGlobalAnomalousLoginDetector() *AnomalousLoginDetector {
	if globalAnomalousLoginDetector == nil {
		globalAnomalousLoginDetector = NewAnomalousLoginDetector()
	}
	return globalAnomalousLoginDetector
}

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	MaxRequests int           // 最大请求数
	TimeWindow  time.Duration // 时间窗口
	Identifier  string        // 限制标识（IP 或 UID）
}

// RateLimitMiddleware 通用速率限制中间件
func RateLimitMiddleware(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("%s:%s", c.ClientIP(), config.Identifier)

		rateLimitStore.mu.Lock()
		defer rateLimitStore.mu.Unlock()

		now := time.Now()
		// 获取该标识符的请求记录
		requests := rateLimitStore.requests[key]

		// 移除超过时间窗口的请求记录
		validRequests := []time.Time{}
		for _, reqTime := range requests {
			if now.Sub(reqTime) <= config.TimeWindow {
				validRequests = append(validRequests, reqTime)
			}
		}

		// 检查是否超过限制
		if len(validRequests) >= config.MaxRequests {
			log.Printf("⚠️  速率限制触发 - %s: %d 次请求在 %v 内", key, len(validRequests), config.TimeWindow)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": fmt.Sprintf("too many requests, please try again later (limit: %d requests per %v)",
					config.MaxRequests, config.TimeWindow),
			})
			c.Abort()
			return
		}

		// 记录当前请求
		validRequests = append(validRequests, now)
		rateLimitStore.requests[key] = validRequests

		c.Next()
	}
}

// LoginRateLimiter 为登录接口专用的速率限制中间件（5次/小时）
func LoginRateLimiter() gin.HandlerFunc {
	return RateLimitMiddleware(RateLimitConfig{
		MaxRequests: 5,
		TimeWindow:  time.Hour,
		Identifier:  "auth:login",
	})
}

// RegisterRateLimiter 为注册接口专用的速率限制中间件（3次/小时）
func RegisterRateLimiter() gin.HandlerFunc {
	return RateLimitMiddleware(RateLimitConfig{
		MaxRequests: 3,
		TimeWindow:  time.Hour,
		Identifier:  "auth:register",
	})
}

// SendCodeRateLimiter 为发送验证码接口的速率限制中间件（10次/小时）
func SendCodeRateLimiter() gin.HandlerFunc {
	return RateLimitMiddleware(RateLimitConfig{
		MaxRequests: 10,
		TimeWindow:  time.Hour,
		Identifier:  "auth:send-code",
	})
}

// AnomalousLoginDetector 异常登录检测
type AnomalousLoginDetector struct {
	mu                sync.RWMutex
	lastLogins        map[int]LoginInfo // UID -> 最后的登录信息
	multiFailedLogins map[string]int    // IP -> 失败次数
}

type LoginInfo struct {
	IP          string    // 最后登录 IP
	LastLoginAt time.Time // 最后登录时间
	UserAgent   string    // 最后的用户代理
}

var detector = &AnomalousLoginDetector{
	lastLogins:        make(map[int]LoginInfo),
	multiFailedLogins: make(map[string]int),
}

// CheckAnomalousLogin 检测异常登录并返回异常类型
func (d *AnomalousLoginDetector) CheckAnomalousLogin(uid int, ip string, userAgent string) []string {
	d.mu.Lock()
	defer d.mu.Unlock()

	anomalies := []string{}

	lastLogin, exists := d.lastLogins[uid]
	if !exists {
		// 第一次登录，记录信息
		d.lastLogins[uid] = LoginInfo{
			IP:          ip,
			LastLoginAt: time.Now(),
			UserAgent:   userAgent,
		}
		return anomalies
	}

	// 检测是否来自不同的 IP
	if lastLogin.IP != ip {
		anomalies = append(anomalies, fmt.Sprintf("login from different IP: %s -> %s", lastLogin.IP, ip))
	}

	// 检测登录时间间隔是否太短（< 5 分钟）
	timeSinceLastLogin := time.Since(lastLogin.LastLoginAt)
	if timeSinceLastLogin < 5*time.Minute {
		anomalies = append(anomalies, fmt.Sprintf("login too frequent: last login %v ago", timeSinceLastLogin))
	}

	// 检测user agent是否变化
	if lastLogin.UserAgent != userAgent {
		anomalies = append(anomalies, "user agent changed")
	}

	// 更新登录信息
	d.lastLogins[uid] = LoginInfo{
		IP:          ip,
		LastLoginAt: time.Now(),
		UserAgent:   userAgent,
	}

	return anomalies
}

// RecordLoginFailure 记录登录失败
func RecordLoginFailure(ip string) {
	detector.mu.Lock()
	defer detector.mu.Unlock()

	detector.multiFailedLogins[ip]++
	failCount := detector.multiFailedLogins[ip]

	if failCount > 3 {
		log.Printf("⚠️  登录失败超过3次 - IP: %s, 失败次数: %d", ip, failCount)
	}
}

// ClearLoginFailure 清除登录失败记录
func ClearLoginFailure(ip string) {
	detector.mu.Lock()
	defer detector.mu.Unlock()

	delete(detector.multiFailedLogins, ip)
}

// GetLoginFailureCount 获取登录失败次数
func GetLoginFailureCount(ip string) int {
	detector.mu.RLock()
	defer detector.mu.RUnlock()

	return detector.multiFailedLogins[ip]
}

// CleanupExpiredLoginRecords 定期清理过期的登录记录
func CleanupExpiredLoginRecords() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		detector.mu.Lock()

		// 清理 7 天没有登录的用户记录
		expiryTime := time.Now().Add(-7 * 24 * time.Hour)
		for uid, info := range detector.lastLogins {
			if info.LastLoginAt.Before(expiryTime) {
				delete(detector.lastLogins, uid)
			}
		}

		// 清理失败登录记录（每次清理时全部清零，重新计数）
		detector.multiFailedLogins = make(map[string]int)

		detector.mu.Unlock()
		log.Println("✅ 登录记录清理完成")
	}
}
