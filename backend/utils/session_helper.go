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
	sid := GenerateSessionID()
	_, err := database.DB.Exec("INSERT INTO user_sessions (id, user_uid, user_agent, ip_address) VALUES (?, ?, ?, ?)", sid, uid, ua, ip)
	if err != nil {
		log.Printf("创建会话失败: %v", err)
		return "", err
	}
	return sid, nil
}

func DeleteSession(sid string) error {
	_, err := database.DB.Exec("DELETE FROM user_sessions WHERE id = ?", sid)
	return err
}

func UpdateSessionActivity(sid string) {
	_, _ = database.DB.Exec("UPDATE user_sessions SET last_active = CURRENT_TIMESTAMP WHERE id = ?", sid)
}

func IsSessionValid(sid string) bool {
	var exists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM user_sessions WHERE id = ?)", sid).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}
