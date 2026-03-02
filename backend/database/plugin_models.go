package database

import "time"

// Plugin 插件表
type Plugin struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"not null;size:100" json:"name"`
	Description  string    `gorm:"type:text" json:"description"`
	Author       string    `gorm:"size:100" json:"author"`                    // 来自 manifest.json 的作者名
	Version      string    `gorm:"size:32;default:'1.0.0'" json:"version"`    // 来自 manifest.json 的版本号
	CumodHash    string    `gorm:"size:64;index" json:"cumod_hash,omitempty"` // .cumod 文件 SHA256（防重复安装）
	AuthorUID    int       `json:"author_uid"`                                // 手动创建时的管理员 UID
	Script       string    `gorm:"type:text" json:"-"`                        // 可选：客户端插件脚本（JS），不直接下发
	ServerScript string    `gorm:"type:text" json:"-"`                        // 可选：服务端插件脚本（JS）
	ConfigSchema string    `gorm:"type:text" json:"config_schema"`            // JSON: 插件配置 schema
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// PluginCard 插件卡牌表
type PluginCard struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	PluginID     uint      `gorm:"not null;index" json:"plugin_id"`
	Symbol       string    `gorm:"uniqueIndex;not null;size:32" json:"symbol"` // 卡牌唯一标识，如 "SWAP3"
	DisplayName  string    `gorm:"size:64" json:"display_name"`                // 展示名
	EffectType   string    `gorm:"not null;size:32" json:"effect_type"`        // "swap" | "force_play" | "convert"
	EffectConfig string    `gorm:"type:text;not null" json:"effect_config"`    // JSON 配置
	DefaultCount int       `gorm:"default:2" json:"default_count"`             // 默认牌组中的数量
	Color        string    `gorm:"size:32" json:"color"`                       // 前端颜色提示，如 "#06b6d4"
	CreatedAt    time.Time `json:"created_at"`
}
