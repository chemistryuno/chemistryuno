package database

import (
	"database/sql/driver"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// JSON 自定义类型用于处理数据库序列化和 JSON 序列化
type JSON []byte

// Scan 实现 sql.Scanner 接口
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = append((*j)[0:0], v...)
	case string:
		*j = append((*j)[0:0], v...)
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (j JSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (j *JSON) UnmarshalJSON(b []byte) error {
	*j = append((*j)[0:0], b...)
	return nil
}

// User GORM模型 - 用户表
type User struct {
	UID                uint           `gorm:"primaryKey;autoIncrement" json:"uid"`
	Username           string         `gorm:"size:50;index;default:''" json:"username"` // 旧系统保留字段，邮箱模式下可为空
	Email              string         `gorm:"unique;size:100;index;default:null" json:"email"`
	Nickname           string         `gorm:"not null;size:50;default:''" json:"nickname"`
	Password           string         `gorm:"not null;default:''" json:"-"`
	Avatar             string         `gorm:"type:text" json:"avatar"`
	IsAdmin            bool           `gorm:"default:false" json:"is_admin"`
	Role               string         `gorm:"default:user;size:20" json:"role"`
	TwoFactorEnabled   bool           `gorm:"default:false" json:"two_factor_enabled"`
	TwoFactorSecret    string         `json:"-"`
	Points             int            `gorm:"default:1000" json:"points"`
	MonthlyPoints      int            `gorm:"default:1000" json:"monthly_points"`
	NegativePlayCount  int            `gorm:"default:0" json:"negative_play_count"`
	BannedUntil        *time.Time     `json:"banned_until"`
	BanReason          string         `gorm:"size:255" json:"ban_reason"`
	FrozenUntil        *time.Time     `json:"frozen_until"`
	TotalGames         int            `gorm:"default:0" json:"total_games"`
	WinCount           int            `gorm:"default:0" json:"win_count"`
	TurnStartedAt      *time.Time     `json:"turn_started_at"`
	RoomReady          bool           `gorm:"default:false" json:"room_ready"`
	LastOfflineAt      *time.Time     `json:"last_offline_at"`
	LastWeeklyDecayAt  time.Time      `json:"last_weekly_decay_at"`
	LastMonthlyResetAt time.Time      `json:"last_monthly_reset_at"`
	WebAuthnID         string         `gorm:"size:100" json:"-"`
	GithubID           string         `gorm:"size:100;index" json:"github_id,omitempty"`
	MicrosoftID        string         `gorm:"size:100;index" json:"microsoft_id,omitempty"`
	GoogleID           string         `gorm:"size:100;index" json:"google_id,omitempty"`
	AppleID            string         `gorm:"size:100;index" json:"apple_id,omitempty"`
	OAuthProvider      string         `gorm:"size:20" json:"oauth_provider,omitempty"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

// VerificationCode 邮箱验证码
type VerificationCode struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Email     string    `gorm:"index;size:100;not null"`
	Code      string    `gorm:"size:10;not null"`
	Type      string    `gorm:"size:20;not null"` // "register" or "reset"
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

func (VerificationCode) TableName() string {
	return "verification_codes"
}

// UserSession GORM模型 - 用户会话表
type UserSession struct {
	ID         string    `gorm:"primaryKey;size:64" json:"id"`
	UserUID    uint      `gorm:"not null;index" json:"user_uid"`
	UserAgent  string    `gorm:"type:text" json:"user_agent"`
	IPAddress  string    `gorm:"size:45" json:"ip_address"`
	LastActive time.Time `gorm:"autoCreateTime" json:"last_active"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (UserSession) TableName() string {
	return "user_sessions"
}

// GlobalChat GORM模型 - 全服聊天记录表
type GlobalChat struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserUID   uint      `gorm:"not null;index" json:"user_uid"`
	Username  string    `gorm:"size:50" json:"username"`
	Avatar    string    `gorm:"type:text" json:"avatar"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

func (GlobalChat) TableName() string {
	return "global_chats"
}

// WebAuthnCredential GORM模型 - WebAuthn凭证表
type WebAuthnCredential struct {
	ID              string    `gorm:"primaryKey;size:255" json:"id"`
	UserUID         uint      `gorm:"not null;index" json:"user_uid"`
	PublicKey       []byte    `gorm:"type:blob" json:"-"`
	AttestationType string    `gorm:"size:50" json:"attestation_type"`
	AAGUID          []byte    `gorm:"size:64" json:"-"`
	SignCount       uint32    `json:"sign_count"`
	CloneWarning    bool      `gorm:"default:false" json:"clone_warning"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (WebAuthnCredential) TableName() string {
	return "webauthn_credentials"
}

// Friendship GORM模型 - 好友关系表
type Friendship struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserUID      uint      `gorm:"not null;index:idx_friendship" json:"user_uid"`   // 发起方
	FriendUID    uint      `gorm:"not null;index:idx_friendship" json:"friend_uid"` // 接收方
	Status       string    `gorm:"default:pending;size:20" json:"status"`           // pending, accepted, declined
	HelloMessage string    `gorm:"size:255" json:"hello_message"`                   // 发时附带的消息
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	User   User `gorm:"foreignKey:UserUID" json:"-"`
	Friend User `gorm:"foreignKey:FriendUID" json:"friend"`
}

func (Friendship) TableName() string {
	return "friendships"
}

// Reaction GORM模型 - 化学反应表
type Reaction struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Reactants    string     `gorm:"not null;size:500" json:"reactants"`
	Products     string     `gorm:"not null;size:500" json:"products"`
	Display      string     `gorm:"size:1000" json:"display"`
	CreatedByUID uint       `gorm:"not null" json:"created_by_uid"`
	Status       string     `gorm:"default:pending;size:20" json:"status"`
	Bidirection  bool       `gorm:"default:false" json:"bidirection"`
	GroupID      *uint      `gorm:"index" json:"group_id"`
	SubmittedAt  time.Time  `gorm:"autoCreateTime" json:"submitted_at"`
	ApprovedAt   *time.Time `json:"approved_at"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (Reaction) TableName() string {
	return "reactions"
}

// Substance GORM模型 - 化学物质表
type Substance struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string     `gorm:"not null;unique;size:255" json:"name"`
	Description  string     `gorm:"type:text" json:"description"`
	Formula      string     `gorm:"size:255" json:"formula"`
	Elements     string     `gorm:"size:500" json:"elements"`
	CreatedByUID uint       `gorm:"not null" json:"created_by_uid"`
	Status       string     `gorm:"default:pending;size:20" json:"status"`
	SubmittedAt  time.Time  `gorm:"autoCreateTime" json:"submitted_at"`
	ApprovedAt   *time.Time `json:"approved_at"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (Substance) TableName() string {
	return "substances"
}

// Feedback GORM模型 - 用户反馈表
type Feedback struct {
	ID             uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserUID        uint       `gorm:"not null;index" json:"user_uid"`
	Type           string     `gorm:"not null;size:20" json:"type"`
	Content        string     `gorm:"not null;type:text" json:"content"`
	Status         string     `gorm:"default:pending;size:20" json:"status"`
	ProcessedByUID *uint      `json:"processed_by_uid"`
	ProcessedAt    *time.Time `json:"processed_at"`
	LastUrgedAt    *time.Time `json:"last_urged_at"`
	UrgeCount      int        `gorm:"default:0" json:"urge_count"`
	ResolutionNote string     `gorm:"type:text" json:"resolution_note"`
	RemoveAt       *time.Time `json:"remove_at"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (Feedback) TableName() string {
	return "feedbacks"
}

// DeckConfig GORM模型 - 牌组配置表
type DeckConfig struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"not null;size:100" json:"name"`
	Cards        JSON      `gorm:"not null;type:text" json:"cards"`
	InitialCards int       `gorm:"default:10" json:"initial_cards"`
	CreatedByUID uint      `gorm:"not null;index" json:"created_by_uid"`
	IsGlobal     bool      `gorm:"default:false" json:"is_global"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (DeckConfig) TableName() string {
	return "deck_configs"
}

// GameHistory GORM模型 - 游戏历史表
type GameHistory struct {
	ID                  uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RoomID              string    `gorm:"not null;size:50;index" json:"room_id"`
	WinnerUID           *uint     `json:"winner_uid"`
	Players             JSON      `gorm:"not null;type:json" json:"players"`
	OriginalPlayerCount int       `json:"original_player_count"`
	QuittedCount        int       `json:"quitted_count"`
	StartedAt           time.Time `json:"started_at"`
	FinishedAt          time.Time `json:"finished_at"`
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (GameHistory) TableName() string {
	return "game_history"
}

// Bounty GORM模型 - 悬赏表
type Bounty struct {
	ID        uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	IssuerUID uint       `gorm:"not null;index" json:"issuer_uid"`
	TargetUID uint       `gorm:"not null;index" json:"target_uid"`
	Amount    int        `gorm:"not null" json:"amount"`
	Status    string     `gorm:"default:active;size:20" json:"status"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	ClaimedAt *time.Time `json:"claimed_at"`
}

func (Bounty) TableName() string {
	return "bounties"
}

// Announcement GORM模型 - 公告表
type Announcement struct {
	ID              uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Title           string     `gorm:"size:255" json:"title"`
	Content         string     `gorm:"not null;type:text" json:"content"`
	Type            string     `gorm:"default:info;size:20" json:"type"`
	Active          bool       `gorm:"default:true" json:"active"`
	IsTicker        bool       `gorm:"default:false" json:"is_ticker"`
	IsPersistent    bool       `gorm:"default:false" json:"is_persistent"`
	OnJoin          bool       `gorm:"default:false" json:"on_join"`
	CronInterval    int        `gorm:"default:0" json:"cron_interval"`
	CloseDelay      int        `gorm:"default:0" json:"close_delay"`
	LastBroadcastAt *time.Time `json:"last_broadcast_at"`
	ExpiresAt       *time.Time `json:"expires_at"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Announcement) TableName() string {
	return "announcements"
}

// SystemConfig GORM模型 - 系统配置表
type SystemConfig struct {
	Key       string    `gorm:"primaryKey;size:100" json:"key"`
	Value     string    `gorm:"not null;type:text" json:"value"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}
