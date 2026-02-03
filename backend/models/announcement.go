package models

import "time"

type Announcement struct {
	ID        int        `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Type      string     `json:"type"` // info, maintenance, emergency
	Active    bool       `json:"active"`
	IsTicker  bool       `json:"is_ticker"` // 是否是滚动条公告
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}
