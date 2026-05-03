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
	Username           string         `gorm:"size:50;uniqueIndex:idx_users_username;not null;default:''" json:"username"` // 主要登录标识符，唯一
	Email              string         `gorm:"size:100;index:idx_users_email;default:''" json:"email"`                     // 可选，允许空值（多用户无邮箱）
	Nickname           string         `gorm:"not null;size:50;default:''" json:"nickname"`
	Password           string         `gorm:"not null;default:''" json:"-"`
	Avatar             string         `gorm:"type:longtext" json:"avatar"`
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
	BannedUntil        *time.Time     `gorm:"index:idx_users_banned_until" json:"banned_until"`
	BanReason          string         `gorm:"size:255" json:"ban_reason"`
	FrozenUntil        *time.Time     `gorm:"index:idx_users_frozen_until" json:"frozen_until"`
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
	SecurityQuestion   string         `gorm:"size:500;default:''" json:"security_question"`
	SecurityAnswer     string         `gorm:"size:255;default:''" json:"-"` // 存储bcrypt哈希
	Fuel               int            `gorm:"default:0" json:"fuel"`
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
	UserUID    uint      `gorm:"not null;index:idx_user_sessions_user_uid" json:"user_uid"`
	UserAgent  string    `gorm:"type:text" json:"user_agent"`
	IPAddress  string    `gorm:"size:45" json:"ip_address"`
	LastActive time.Time `gorm:"autoCreateTime;index:idx_user_sessions_last_active" json:"last_active"`
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

func (r *Reaction) BeforeCreate(tx *gorm.DB) error {
	if r.R1 > r.R2 {
		r.R1, r.R2 = r.R2, r.R1
	}
	return nil
}

func (r *Reaction) BeforeSave(tx *gorm.DB) error {
	if r.R1 > r.R2 {
		r.R1, r.R2 = r.R2, r.R1
	}
	return nil
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
	ID                  uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	RoomID              string     `gorm:"not null;size:50;index" json:"room_id"`
	WinnerUID           *uint      `json:"winner_uid"`
	IsInvalid           bool       `gorm:"default:false;index" json:"is_invalid"`
	InvalidReason       string     `gorm:"size:255" json:"invalid_reason"`
	ReplayLog           string     `gorm:"type:longtext" json:"replay_log,omitempty"`
	ReplayPermanent     bool       `gorm:"default:false;index" json:"replay_permanent"`
	ReplayExpiresAt     *time.Time `gorm:"index" json:"replay_expires_at,omitempty"`
	ReplayClearedAt     *time.Time `json:"replay_cleared_at,omitempty"`
	CheatDetected       bool       `gorm:"default:false;index" json:"cheat_detected"`
	CheatUIDs           JSON       `gorm:"type:json" json:"cheat_uids,omitempty"`
	Players             JSON       `gorm:"not null;type:json" json:"players"`
	OriginalPlayerCount int        `json:"original_player_count"`
	QuittedCount        int        `json:"quitted_count"`
	StartedAt           time.Time  `json:"started_at"`
	FinishedAt          time.Time  `json:"finished_at"`
	CreatedAt           time.Time  `gorm:"autoCreateTime" json:"created_at"`
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
	Value     string    `gorm:"not null;type:longtext" json:"value"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}

// CheatRiskScore GORM模型 - 作弊风险评分表
type CheatRiskScore struct {
	ID                 uint          `gorm:"primaryKey;autoIncrement" json:"id"`
	RoomID             string        `gorm:"not null;size:50;index" json:"room_id"`
	PlayerUID          uint          `gorm:"not null;index" json:"player_uid"`
	RiskScore          float64       `gorm:"not null;default:0;index" json:"risk_score"`
	ResponseTimeDim    float64       `gorm:"not null;default:0" json:"response_time_dim"`
	FrequencyDim       float64       `gorm:"not null;default:0" json:"frequency_dim"`
	WinRateDim         float64       `gorm:"not null;default:0" json:"win_rate_dim"`
	PatternDim         float64       `gorm:"not null;default:0" json:"pattern_dim"`
	AccountAgeDim      float64       `gorm:"not null;default:0" json:"account_age_dim"`
	DetectionTime      time.Time     `gorm:"not null;autoCreateTime" json:"detection_time"`
	CreatedAt          time.Time     `gorm:"autoCreateTime" json:"created_at"`
}

func (CheatRiskScore) TableName() string {
	return "cheat_risk_scores"
}

// CheatSanction GORM模型 - 作弊处罚表
type CheatSanction struct {
	ID             uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	RoomID         string     `gorm:"not null;size:50;index" json:"room_id"`
	PlayerUID      uint       `gorm:"not null;index" json:"player_uid"`
	RiskScoreID    uint       `gorm:"index" json:"risk_score_id"`
	SanctionType   string     `gorm:"not null;size:20;index" json:"sanction_type"` // "observe", "warning", "mute", "ban"
	RiskScore      float64    `gorm:"not null" json:"risk_score"`
	Reason         string     `gorm:"type:text" json:"reason"`
	Duration       *int       `gorm:"comment:禁言/封号时长(分钟)" json:"duration"`
	AppliedAt      time.Time  `gorm:"not null;autoCreateTime" json:"applied_at"`
	EffectiveUntil *time.Time `json:"effective_until"`
	Status         string     `gorm:"default:active;size:20" json:"status"` // "active", "revoked", "expired"
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (CheatSanction) TableName() string {
	return "cheat_sanctions"
}

// CheatAppeal GORM模型 - 作弊申诉表
type CheatAppeal struct {
	ID                    uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	RoomID                string     `gorm:"not null;size:50;index" json:"room_id"`
	PlayerUID             uint       `gorm:"not null;index" json:"player_uid"`
	RiskScoreID           uint       `gorm:"index" json:"risk_score_id"`
	SanctionID            uint       `gorm:"index" json:"sanction_id"`
	Reason                string     `gorm:"not null;type:text" json:"reason"`
	Evidence              string     `gorm:"type:text" json:"evidence"`
	Status                string     `gorm:"not null;default:pending;size:20;index" json:"status"` // "pending", "under_review", "approved", "rejected"
	ReviewerUID           *uint      `gorm:"index" json:"reviewer_uid"`
	ReviewedAt            *time.Time `json:"reviewed_at"`
	ReviewRemark          string     `gorm:"type:text" json:"review_remark"`
	CompensationAmount    int        `gorm:"default:0" json:"compensation_amount"`
	CompensationStatus    string     `gorm:"size:20;default:'pending'" json:"compensation_status"` // "pending", "completed", "failed"
	CompensationNote      string     `gorm:"type:text" json:"compensation_note"`
	SubmittedAt           time.Time  `gorm:"not null;autoCreateTime" json:"submitted_at"`
	CreatedAt             time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (CheatAppeal) TableName() string {
	return "cheat_appeals"
}

// CheatAuditLog GORM模型 - 作弊审计日志表
type CheatAuditLog struct {
	ID                      uint          `gorm:"primaryKey;autoIncrement" json:"id"`
	EventType               string        `gorm:"not null;size:50;index" json:"event_type"` // "detection", "sanction", "appeal", "review", "revoke"
	RoomID                  string        `gorm:"size:50;index" json:"room_id"`
	PlayerUID               uint          `gorm:"index" json:"player_uid"`
	OperatorUID             *uint         `gorm:"index" json:"operator_uid"` // 操作人(审核人或系统)
	RiskScoreID             *uint         `json:"risk_score_id"`
	SanctionID              *uint         `json:"sanction_id"`
	AppealID                *uint         `json:"appeal_id"`
	RiskScore               *float64      `json:"risk_score"`
	SanctionType            string        `gorm:"size:20" json:"sanction_type"`
	OldStatus               string        `gorm:"size:50" json:"old_status"`
	NewStatus               string        `gorm:"size:50" json:"new_status"`
	Details                 JSON          `gorm:"type:json" json:"details"`
	Remark                  string        `gorm:"type:text" json:"remark"`
	ApprovalNote            *string       `gorm:"type:text" json:"approval_note"`           // Admin notes during appeal approval
	CompensationAmount      *int          `gorm:"index" json:"compensation_amount"`         // Fuel amount for unban compensation
	CompensationStatus      *string       `gorm:"size:20" json:"compensation_status"`       // pending/ok/failed
	CompensationMessage     *string       `gorm:"type:text" json:"compensation_message"`    // Message sent to player
	CompensationNote        *string       `gorm:"type:text" json:"compensation_note"`       // Failure reason or notes
	CompensationDate        *time.Time    `gorm:"index" json:"compensation_date"`           // When compensation was issued/attempted
	CreatedAt               time.Time     `gorm:"not null;autoCreateTime;index" json:"created_at"`
}

func (CheatAuditLog) TableName() string {
	return "cheat_audit_logs"
}

// FuelCompensationRecord 燃素补偿记录表
type FuelCompensationRecord struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserUID        uint      `gorm:"not null;index" json:"user_uid"`
	Amount         int       `gorm:"not null" json:"amount"`
	CompensationID string    `gorm:"size:100;index" json:"compensation_id"`
	Reason         string    `gorm:"size:255;default:''" json:"reason"`
	CreatedAt      time.Time `gorm:"not null;autoCreateTime;index" json:"created_at"`
}

func (FuelCompensationRecord) TableName() string {
	return "fuel_compensation_records"
}

