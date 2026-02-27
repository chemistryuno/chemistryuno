package models

// PluginCardDef 是插件卡牌的业务层定义（来自 DB，加载到内存 registry）
type PluginCardDef struct {
	ID           uint   `json:"id"`
	PluginID     uint   `json:"plugin_id"`
	Symbol       string `json:"symbol"`
	DisplayName  string `json:"display_name"`
	EffectType   string `json:"effect_type"`   // "swap" | "force_play" | "convert"
	EffectConfig string `json:"effect_config"` // JSON 原始字符串
	DefaultCount int    `json:"default_count"`
	Color        string `json:"color"`
}

// SwapConfig swap 效果配置
type SwapConfig struct {
	Count int `json:"count"` // 随机交换的手牌张数
}

// ForcePlayConfig force_play 效果配置
type ForcePlayConfig struct {
	Count int `json:"count"` // 下一位玩家必须额外打出的张数
}

// ConvertConfig convert 效果配置
type ConvertConfig struct {
	SourceCount int `json:"source_count"` // 消耗自身的张数
	TargetCount int `json:"target_count"` // 从摸牌堆摸取的张数
}
