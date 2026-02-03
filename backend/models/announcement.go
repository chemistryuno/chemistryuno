package models

import "time"

type Announcement struct {
	ID              int        `json:"id"`
	Title           string     `json:"title"`
	Content         string     `json:"content"`
	Type            string     `json:"type"` // info, maintenance, emergency
	Active          bool       `json:"active"`
	IsTicker        bool       `json:"is_ticker"`     // 是否是滚动条公告
	IsPersistent    bool       `json:"is_persistent"` // 是否是常驻公告（在特定区域始终显示）
	OnJoin          bool       `json:"on_join"`       // 玩家加入时是否触发
	CronInterval    int        `json:"cron_interval"` // 定时广播间隔（分钟），0表示不定时
	CloseDelay      int        `json:"close_delay"`   // 强制延迟关闭时间（秒）
	CreatedAt       time.Time  `json:"created_at"`
	ExpiresAt       *time.Time `json:"expires_at"`
	LastBroadcastAt *time.Time `json:"last_broadcast_at"`
}
