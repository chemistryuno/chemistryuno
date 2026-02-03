package utils

import (
	"chemistryuno/database"
	"crypto/rand"
	"encoding/hex"
	"log"
)

type Session struct {
	ID         string `json:"id"`
	UserUID    int    `json:"user_uid"`
	UserAgent  string `json:"user_agent"`
	IPAddress  string `json:"ip_address"`
	LastActive string `json:"last_active"`
	CreatedAt  string `json:"created_at"`
}

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
	_, err := database.Exec("INSERT INTO user_sessions (id, user_uid, user_agent, ip_address) VALUES (?, ?, ?, ?)", sid, uid, ua, ip)
	if err != nil {
		log.Printf("创建会话失败: %v", err)
		return "", err
	}
	return sid, nil
}

func DeleteSession(sid string) error {
	_, err := database.Exec("DELETE FROM user_sessions WHERE id = ?", sid)
	return err
}

func UpdateSessionActivity(sid string, ip string) {
	// 同时更新最后活跃时间和 IP，防止用户因为看到旧 IP 而误以为被盗号
	_, _ = database.Exec("UPDATE user_sessions SET last_active = NOW(), ip_address = ? WHERE id = ?", ip, sid)
}

func IsSessionValid(sid string) bool {
	var exists bool
	err := database.QueryRow("SELECT EXISTS(SELECT 1 FROM user_sessions WHERE id = ?)", sid).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

// ValidateSessionForUser 验证会话是否属于指定用户
func ValidateSessionForUser(sid string, uid int) bool {
	var sessionUID int
	err := database.QueryRow("SELECT user_uid FROM user_sessions WHERE id = ?", sid).Scan(&sessionUID)
	if err != nil {
		return false
	}
	return sessionUID == uid
}
