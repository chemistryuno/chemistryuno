package models

import "time"

// 卡牌类型
type Card struct {
	Type   string `json:"type"`   // 元素符号或特殊牌
	Count  int    `json:"count"`  // 剩余数量
	Effect string `json:"effect"` // 特殊效果: "reverse", "skip", "+2", "+4"
}

// 牌组配置
type DeckConfig struct {
	ID        int            `json:"id" db:"id"`
	Name      string         `json:"name" db:"name"`
	IsGlobal  bool           `json:"is_global" db:"is_global"`
	Cards     map[string]int `json:"cards"` // 元素->数量
	CreatedBy int            `json:"created_by" db:"created_by"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
}

// 游戏房间
type Room struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	HostUID    int         `json:"host_uid"`
	Players    []int       `json:"players"`
	MaxPlayers int         `json:"max_players"`
	DeckConfig *DeckConfig `json:"deck_config"`
	Status     string      `json:"status"` // "waiting", "playing", "finished"
	CreatedAt  time.Time   `json:"created_at"`
}

// 游戏状态
type GameState struct {
	RoomID        string         `json:"room_id"`
	Players       []*PlayerState `json:"players"`
	CurrentPlayer int            `json:"current_player"`
	Direction     int            `json:"direction"` // 1: 顺时针, -1: 逆时针
	LastCard      *PlayedCard    `json:"last_card"`
	DrawPile      []Card         `json:"-"` // 摸牌堆（不发送给客户端）
	DiscardPile   []PlayedCard   `json:"discard_pile"`
	Status        string         `json:"status"`        // "playing", "finished"
	TurnEndTime   int64          `json:"turn_end_time"` // 回合结束时间戳（毫秒）
}

// 玩家状态
type PlayerState struct {
	UID       int    `json:"uid"`
	Username  string `json:"username"`
	Avatar    string `json:"avatar"`
	HandCards []Card `json:"hand_cards"` // 手牌
	CardCount int    `json:"card_count"` // 手牌数量（其他玩家只能看到数量）
	IsReady   bool   `json:"is_ready"`
}

// 已出牌
type PlayedCard struct {
	Card      Card   `json:"card"`
	Substance string `json:"substance"` // 组成的物质
	PlayerUID int    `json:"player_uid"`
}
