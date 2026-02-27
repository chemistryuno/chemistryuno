package models

import (
	"encoding/json"
	"time"
)

// 卡牌类型
type Card struct {
	Type   string `json:"type"`   // 元素符号或特殊牌
	Count  int    `json:"count"`  // 剩余数量
	Effect string `json:"effect"` // 特殊效果: "reverse", "Au", "+2", "+4"
}

// 牌组配置
type DeckConfig struct {
	ID           int            `json:"id" db:"id"`
	Name         string         `json:"name" db:"name"`
	IsGlobal     bool           `json:"is_global" db:"is_global"`
	Cards        map[string]int `json:"cards"` // 元素->数量
	InitialCards int            `json:"initial_cards" db:"initial_cards"`
	CreatedBy    int            `json:"created_by" db:"created_by"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
}

// 游戏房间
type Room struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Players       []int       `json:"players"`
	ReadyUIDs     []int       `json:"ready_uids"` // 准备好的玩家UID
	Countdown     int         `json:"countdown"`  // 0 为无倒计时
	Spectators    []int       `json:"spectators"`
	MaxPlayers    int         `json:"max_players"`
	DeckConfig    *DeckConfig `json:"deck_config"`
	Status        string      `json:"status"` // "waiting", "playing", "finished"
	IsPointsMode  bool        `json:"is_points_mode"`
	IsPrivate     bool        `json:"is_private"`     // 是否为私密房间，不显示在大厅
	AccessKey     string      `json:"access_key"`     // 私密房间访问密钥
	IsDuel        bool        `json:"is_duel"`        // 是否是单挑模式
	ChallengerUID int         `json:"challenger_uid"` // 发起者 UID
	TargetUID     int         `json:"target_uid"`     // 被挑战者 UID
	CreatedAt     time.Time   `json:"created_at"`
	IsPvE         bool        `json:"is_pve"`         // 是否为人机对战模式
	PvEDifficulty int         `json:"pve_difficulty"` // 人机难度 1-100
	AICount       int         `json:"ai_count"`       // AI 数量

	// AI补位功能配置
	EnableAIBackfill     bool  `json:"enable_ai_backfill"`     // 是否启用AI补位
	AIBackfillDifficulty int   `json:"ai_backfill_difficulty"` // 补位AI难度（1-100）
	BackfilledAIUIDs     []int `json:"backfilled_ai_uids"`     // 记录自动补充的AI UID列表

	// 等级匹配系统
	IsRanked   bool `json:"is_ranked"`    // 是否为排位模式（启用严格等级匹配）
	LevelRange int  `json:"level_range"`  // 允许的等级范围（±N级）
	MinLevel   int  `json:"min_level"`    // 允许的最低等级
	MaxLevel   int  `json:"max_level"`    // 允许的最高等级
	CreatedByUID int `json:"created_by_uid"` // 房主UID（用于计算等级范围）

	// 教学脚本系统
	TutorialScript bool `json:"tutorial_script"` // 是否启用脚本化教学
}

// MarshalJSON 自定义 JSON 序列化，确保 ready_uids 和其他切片字段永远不为 null
func (r *Room) MarshalJSON() ([]byte, error) {
	type Alias Room
	// 确保切片字段不为 nil，避免 JSON 序列化为 null
	players := r.Players
	if players == nil {
		players = []int{}
	}
	readyUIDs := r.ReadyUIDs
	if readyUIDs == nil {
		readyUIDs = []int{}
	}
	spectators := r.Spectators
	if spectators == nil {
		spectators = []int{}
	}
	backfilledAIUIDs := r.BackfilledAIUIDs
	if backfilledAIUIDs == nil {
		backfilledAIUIDs = []int{}
	}

	return json.Marshal(&struct {
		Players          []int `json:"players"`
		ReadyUIDs        []int `json:"ready_uids"`
		Spectators       []int `json:"spectators"`
		BackfilledAIUIDs []int `json:"backfilled_ai_uids"`
		*Alias
	}{
		Players:          players,
		ReadyUIDs:        readyUIDs,
		Spectators:       spectators,
		BackfilledAIUIDs: backfilledAIUIDs,
		Alias:            (*Alias)(r),
	})
}

// 游戏状态
type GameState struct {
	RoomID              string         `json:"room_id"`
	Players             []*PlayerState `json:"players"`
	Spectators          []int          `json:"spectators"`
	FinishedPlayers     []int          `json:"finished_players"` // 已完成比赛的玩家UID列表
	CurrentPlayer       int            `json:"current_player"`
	Direction           int            `json:"direction"` // 1: 顺时针, -1: 逆时针
	LastCard            *PlayedCard    `json:"last_card"`
	DrawPile            []Card         `json:"-"` // 摸牌堆（不发送给客户端）
	DiscardPile         []PlayedCard   `json:"discard_pile"`
	AllUsedCards        []Card         `json:"-"`                     // 累计已排出的所有卡牌池（用于洗牌）
	OriginalPlayerCount int            `json:"original_player_count"` // 初始玩家总数
	QuittedCount        int            `json:"quitted_count"`         // 中途离开/被踢出的玩家数
	Status              string         `json:"status"`                // "playing", "finished"
	IsPointsMode        bool           `json:"is_points_mode"`        // 同步房间配置
	TurnEndTime         int64          `json:"turn_end_time"`         // 回合结束时间戳（毫秒）
	PendingDrawCount    int            `json:"pending_draw_count"`    // 当前累计需加牌数
	PendingDrawTypes    []string       `json:"pending_draw_types"`    // 当前累计加牌类型（如["+2","+4"]）
	AllowedAnyPlayer    int            `json:"allowed_any_player"`    // 允许无视反应条件直接出牌的玩家索引，-1 表示无
	PointsChanges       map[int]int    `json:"points_changes"`        // 回合结束时的积分变动 (UID -> points)
	XPChanges           map[int]int    `json:"xp_changes"`            // 回合结束时的经验变动 (UID -> xp)
	CurrentReaction     string         `json:"current_reaction"`      // 当前/最近发生的反应方程式

	// 教学脚本系统
	TutorialScriptMode  bool `json:"tutorial_script_mode"`  // 是否为脚本化教学模式
	TutorialCurrentStep int  `json:"tutorial_current_step"` // 当前教学脚本步骤（1-8）

	// 插件系统
	PendingForcedPlays int `json:"pending_forced_plays"` // 当前玩家仍需强制打出的牌数（force_play 效果）
}

// 玩家状态
type PlayerState struct {
	UID                   int    `json:"uid"`
	Username              string `json:"username"`
	Nickname              string `json:"nickname"`
	Avatar                string `json:"avatar"`
	HandCards             []Card `json:"hand_cards"` // 手牌
	CardCount             int    `json:"card_count"` // 手牌数量（其他玩家只能看到数量）
	IsReady               bool   `json:"is_ready"`
	DoubleActionAvailable bool   `json:"double_action_available"` // 是否可以使用双联反应（每2次普通行动可用1次）
	ActionProgress        int    `json:"action_progress"`         // 行动进度（0->1->2(可用)）
	IsAI                  bool   `json:"is_ai"`                   // 是否为 AI 玩家
	IsHosted              bool   `json:"is_hosted"`               // 是否处于托管状态（真人玩家超时进入）
}

// 已出牌
type PlayedCard struct {
	Card      Card     `json:"card"`
	Substance string   `json:"substance"` // 组成的物质
	PlayerUID int      `json:"player_uid"`
	Reactants []string `json:"reactants"` // 如果是双联反应，记录参与反应的两种物质
}
