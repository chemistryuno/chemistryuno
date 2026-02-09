package utils

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"strings"
	"time"
)

type Session struct {
	ID         string `json:"id"`
	UserUID    int    `json:"user_uid"`
	UserAgent  string `json:"user_agent"`
	IPAddress  string `json:"ip_address"`
	LastActive string `json:"last_active"`
	CreatedAt  string `json:"created_at"`
}

var sessionRepo *repository.SessionRepository

func getSessionRepo() *repository.SessionRepository {
	if sessionRepo == nil {
		sessionRepo = repository.SessionRepo
	}
	return sessionRepo
}

func GenerateSessionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Printf("生成Session ID失败: %v", err)
		return ""
	}
	return hex.EncodeToString(b)
}

func CreateSession(uid int, ua string, ip string) (string, error) {
	// 允许用户在同一 UA 下拥有多个会话，不再自动清理
	// 这样可以避免多标签页或相同 UA 的不同设备相互挤出的问题

	// 重试机制：防止极端情况下的Session ID冲突
	maxRetries := 3
	var sid string
	var err error

	for i := 0; i < maxRetries; i++ {
		sid = GenerateSessionID()
		if sid == "" {
			log.Printf("Session ID生成失败，重试 %d/%d", i+1, maxRetries)
			continue
		}

		session := &database.UserSession{
			ID:         sid,
			UserUID:    uint(uid),
			UserAgent:  ua,
			IPAddress:  ip,
			LastActive: time.Now(),
			CreatedAt:  time.Now(),
		}

		err = getSessionRepo().Create(session)
		if err == nil {
			// 创建成功
			log.Printf("[会话创建] UID=%d, SID=%s, IP=%s, UA=%s", uid, sid, ip, ua)
			return sid, nil
		}

		// 如果是主键冲突错误（极其罕见），重试生成新的ID
		// 其他错误直接返回
		if !isDuplicateKeyError(err) {
			log.Printf("创建会话失败: %v", err)
			return "", err
		}
		log.Printf("Session ID冲突，重试 %d/%d", i+1, maxRetries)
	}

	// 所有重试都失败
	if err != nil {
		log.Printf("创建会话失败，已重试%d次: %v", maxRetries, err)
		return "", err
	}
	// Session ID生成失败
	log.Printf("Session ID生成失败，已重试%d次", maxRetries)
	return "", errors.New("无法生成有效的Session ID")
}

// isDuplicateKeyError 检查是否是主键冲突错误
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	// SQLite 和 MySQL 的主键冲突错误信息
	return strings.Contains(errMsg, "UNIQUE constraint failed") ||
		strings.Contains(errMsg, "Duplicate entry") ||
		strings.Contains(errMsg, "duplicate key")
}

func DeleteSession(sid string) error {
	return getSessionRepo().Delete(sid)
}

func UpdateSessionActivity(sid string, ip string) {
	// 同时更新最后活跃时间和 IP，防止用户因为看到旧 IP 而误以为被盗号
	_ = getSessionRepo().UpdateActivity(sid, ip)
}

func IsSessionValid(sid string) bool {
	exists, err := getSessionRepo().Exists(sid)
	if err != nil {
		// 数据库错误时，打印警告但暂时允许通过（避免因数据库暂时不可用而大量踢出用户）
		log.Printf("[警告] Session验证时数据库错误，SID=%s: %v", sid, err)
		// 返回true以避免误踢，但下次验证会重试
		return true
	}
	return exists
}

// ValidateSessionForUser 验证会话是否属于指定用户
func ValidateSessionForUser(sid string, uid int) bool {
	valid, err := getSessionRepo().ValidateSessionForUser(sid, uint(uid))
	if err != nil {
		log.Printf("[警告] 验证用户会话时数据库错误，SID=%s, UID=%d: %v", sid, uid, err)
		// 返回true以避免误踢，但下次验证会重试
		return true
	}
	return valid
}
