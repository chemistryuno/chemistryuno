package utils

import (
	"chemistryuno/database"
	"chemistryuno/repository"
	"crypto/rand"
	"encoding/hex"
	"log"
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

var sessionRepo = repository.NewSessionRepository()

func GenerateSessionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func CreateSession(uid int, ua string, ip string) (string, error) {
	// 允许用户在同一 UA 下拥有多个会话，不再自动清理
	// 这样可以避免多标签页或相同 UA 的不同设备相互挤出的问题

	sid := GenerateSessionID()
	session := &database.UserSession{
		ID:         sid,
		UserUID:    uint(uid),
		UserAgent:  ua,
		IPAddress:  ip,
		LastActive: time.Now(),
		CreatedAt:  time.Now(),
	}

	err := sessionRepo.Create(session)
	if err != nil {
		log.Printf("创建会话失败: %v", err)
		return "", err
	}
	return sid, nil
}

func DeleteSession(sid string) error {
	return sessionRepo.Delete(sid)
}

func UpdateSessionActivity(sid string, ip string) {
	// 同时更新最后活跃时间和 IP，防止用户因为看到旧 IP 而误以为被盗号
	_ = sessionRepo.UpdateActivity(sid, ip)
}

func IsSessionValid(sid string) bool {
	exists, err := sessionRepo.Exists(sid)
	if err != nil {
		return false
	}
	return exists
}

// ValidateSessionForUser 验证会话是否属于指定用户
func ValidateSessionForUser(sid string, uid int) bool {
	valid, err := sessionRepo.ValidateSessionForUser(sid, uint(uid))
	if err != nil {
		return false
	}
	return valid
}
