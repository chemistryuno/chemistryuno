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
	Points             float64        `gorm:"default:1000" json:"points"`
	MonthlyPoints      float64        `gorm:"default:1000" json:"monthly_points"`
	Level              int            `gorm:"default:1" json:"level"`
	XP                 int            `gorm:"default:0" json:"xp"`
	TotalXP            int            `gorm:"default:0" json:"total_xp"`
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
	Bio                string         `gorm:"size:500;default:''" json:"bio"`
	Wechat             string         `gorm:"size:100;default:''" json:"wechat"`
	QQ                 string         `gorm:"size:100;default:''" json:"qq"`
	ShowEmail          bool           `gorm:"default:false" json:"show_email"`
	Birthday           *time.Time     `json:"birthday"`
	SoundVolume        float64        `gorm:"default:0.3" json:"sound_volume"`
	VibrationEnabled   bool           `gorm:"default:true" json:"vibration_enabled"`
	EnableElementInput bool           `gorm:"default:true" json:"enable_element_input"`
	CustomContact      string         `gorm:"size:255;default:''" json:"custom_contact"`
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

// GlobalChat GORM模型 - 全服聊天记录表（大厅聊天，每天0:00清除）
type GlobalChat struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserUID   uint      `gorm:"not null;index" json:"user_uid"`
	Username  string    `gorm:"size:50" json:"username"`
	Nickname  string    `gorm:"size:50" json:"nickname"`
	Avatar    string    `gorm:"type:text" json:"avatar"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

func (GlobalChat) TableName() string {
	return "global_chats"
}

// PrivateChat GORM模型 - 私聊消息记录表
type PrivateChat struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SenderUID    uint      `gorm:"not null;index:idx_sender" json:"sender_uid"`
	ReceiverUID  uint      `gorm:"not null;index:idx_receiver" json:"receiver_uid"`
	Message      string    `gorm:"type:text;not null" json:"message"`
	IsGameInvite bool      `gorm:"default:false" json:"is_game_invite"`              // 是否为游戏邀请
	RoomID       string    `gorm:"size:100;index:idx_room" json:"room_id,omitempty"` // 关联的游戏房间ID，用于游戏邀请
	CreatedAt    time.Time `gorm:"autoCreateTime;index" json:"created_at"`

	Sender   User `gorm:"foreignKey:SenderUID" json:"sender"`
	Receiver User `gorm:"foreignKey:ReceiverUID" json:"receiver"`
}

func (PrivateChat) TableName() string {
	return "private_chats"
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
	PairKey      string    `gorm:"size:50;not null;default:''" json:"-"`
	Status       string    `gorm:"default:pending;size:20" json:"status"` // pending, accepted, declined
	HelloMessage string    `gorm:"size:255" json:"hello_message"`         // 发时附带的消息
	UserRemark   string    `gorm:"size:100" json:"user_remark"`           // UserUID 对 FriendUID 的备注
	FriendRemark string    `gorm:"size:100" json:"friend_remark"`         // FriendUID 对 UserUID 的备注
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	User   User `gorm:"foreignKey:UserUID" json:"-"`
	Friend User `gorm:"foreignKey:FriendUID" json:"friend"`
}

func FriendshipPairKey(uid1, uid2 uint) string {
	if uid1 < uid2 {
		return fmt.Sprintf("%d:%d", uid1, uid2)
	}
	return fmt.Sprintf("%d:%d", uid2, uid1)
}

func (f *Friendship) BeforeCreate(tx *gorm.DB) error {
	if f.PairKey == "" {
		f.PairKey = FriendshipPairKey(f.UserUID, f.FriendUID)
	}
	return nil
}

func (Friendship) TableName() string {
	return "friendships"
}

// Reaction GORM模型 - 化学反应表
type Reaction struct {
	ID                 uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	R1                 string     `gorm:"size:255;index:idx_r1r2_status,priority:1" json:"r1"`
	R2                 string     `gorm:"size:255;index:idx_r1r2_status,priority:2;index:idx_r2r1_status,priority:1" json:"r2"`
	Display            string     `gorm:"size:1000" json:"display"`
	HasInvalidElements bool       `gorm:"default:false" json:"has_invalid_elements"`
	CreatedByUID       uint       `gorm:"not null" json:"created_by_uid"`
	Status             string     `gorm:"default:pending;size:20;index:idx_r1r2_status,priority:3;index:idx_r2r1_status,priority:3;index" json:"status"`
	GroupID            *uint      `gorm:"index" json:"group_id"`
	SubmittedAt        time.Time  `gorm:"autoCreateTime" json:"submitted_at"`
	ApprovedAt         *time.Time `json:"approved_at"`
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (Reaction) TableName() string {
	return "reactions"
}

// Substance GORM模型 - 化学物质表
type Substance struct {
	ID                 uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name               string     `gorm:"not null;size:255;index:idx_name_formula" json:"name"`
	Formula            string     `gorm:"size:255;index:idx_name_formula" json:"formula"`
	Elements           string     `gorm:"size:500" json:"elements"`
	HasInvalidElements bool       `gorm:"default:false;index" json:"has_invalid_elements"`
	CreatedByUID       uint       `gorm:"not null" json:"created_by_uid"`
	Status             string     `gorm:"default:pending;size:20;index" json:"status"`
	GroupID            *uint      `gorm:"index" json:"group_id"`
	NeedsImprovement   bool       `gorm:"default:false;index" json:"needs_improvement"`
	SubmittedAt        time.Time  `gorm:"autoCreateTime" json:"submitted_at"`
	ApprovedAt         *time.Time `json:"approved_at"`
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"created_at"`
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

// Survey GORM模型 - 问卷调查表
type Survey struct {
	ID           uint             `gorm:"primaryKey;autoIncrement" json:"id"`
	Title        string           `gorm:"not null;size:255" json:"title"`
	Description  string           `gorm:"type:text" json:"description"`
	IsActive     bool             `gorm:"not null;index" json:"is_active"`
	RewardPoints int              `gorm:"default:0" json:"reward_points"`
	RewardExp    int              `gorm:"default:0" json:"reward_exp"`
	Questions    []SurveyQuestion `gorm:"foreignKey:SurveyID;constraint:OnDelete:CASCADE" json:"questions,omitempty"`
	CreatedBy    uint             `gorm:"not null" json:"created_by"`
	CreatedAt    time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Survey) TableName() string {
	return "surveys"
}

// SurveyQuestion 问卷题目模型
type SurveyQuestion struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SurveyID    uint      `gorm:"not null;index" json:"survey_id"`
	Title       string    `gorm:"not null;size:500" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Type        string    `gorm:"not null;size:20" json:"type"` // radio, checkbox, text, textarea
	Options     JSON      `gorm:"type:text" json:"options"`     // 选项数组 ["Option 1", "Option 2"]
	IsRequired  bool      `gorm:"not null" json:"is_required"`
	Order       int       `gorm:"default:0" json:"order"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (SurveyQuestion) TableName() string {
	return "survey_questions"
}

// SurveyResponse 问卷结果容器 (每个用户每个问卷一份)
type SurveyResponse struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	SurveyID  uint           `gorm:"not null;index" json:"survey_id"`
	UserUID   uint           `gorm:"not null;index" json:"user_uid"`
	Answers   []SurveyAnswer `gorm:"foreignKey:ResponseID;constraint:OnDelete:CASCADE" json:"answers,omitempty"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

func (SurveyResponse) TableName() string {
	return "survey_responses"
}

// SurveyAnswer 单个题目的回答
type SurveyAnswer struct {
	ID         uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	ResponseID uint   `gorm:"not null;index" json:"response_id"`
	QuestionID uint   `gorm:"not null;index" json:"question_id"`
	Answer     string `gorm:"type:text" json:"answer"` // JSON格式存储
}

func (SurveyAnswer) TableName() string {
	return "survey_answers"
}

// SurveyDismissal 记录玩家选择不再提醒某个问卷
type SurveyDismissal struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserUID   uint      `gorm:"not null;uniqueIndex:idx_user_survey" json:"user_uid"`
	SurveyID  uint      `gorm:"not null;uniqueIndex:idx_user_survey" json:"survey_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (SurveyDismissal) TableName() string {
	return "survey_dismissals"
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
	Key       string    `gorm:"primaryKey;column:key;size:100" json:"key"`
	Value     string    `gorm:"not null;type:text" json:"value"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}
