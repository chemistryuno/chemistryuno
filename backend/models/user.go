package models

import (
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

type User struct {
	UID                int                   `json:"uid" db:"uid"`
	Username           string                `json:"username" db:"username"`
	Email              string                `json:"email" db:"email"`
	Nickname           string                `json:"nickname" db:"nickname"`
	PasswordHash       string                `json:"-" db:"password"` // 不返回给前端
	Avatar             string                `json:"avatar" db:"avatar"`
	Role               string                `json:"role" db:"role"` // admin, co-worker, user
	TwoFactorEnabled   bool                  `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorSecret    string                `json:"-" db:"two_factor_secret"`
	Points             float64               `json:"points" db:"points"`
	MonthlyPoints      float64               `json:"monthly_points" db:"monthly_points"`
	Level              int                   `json:"level" db:"level"`
	XP                 int                   `json:"xp" db:"xp"`
	TotalXP            int                   `json:"total_xp" db:"total_xp"`
	NegativePlayCount  int                   `json:"negative_play_count" db:"negative_play_count"`
	BannedUntil        *time.Time            `json:"banned_until" db:"banned_until"`
	FrozenUntil        *time.Time            `json:"frozen_until" db:"frozen_until"`
	TotalGames         int                   `json:"total_games" db:"total_games"`
	WinCount           int                   `json:"win_count" db:"win_count"`
	TurnStartedAt      *time.Time            `json:"turn_started_at" db:"turn_started_at"`
	LastOfflineAt      *time.Time            `json:"last_offline_at" db:"last_offline_at"`
	LastWeeklyDecayAt  time.Time             `json:"last_weekly_decay_at" db:"last_weekly_decay_at"`
	LastMonthlyResetAt time.Time             `json:"last_monthly_reset_at" db:"last_monthly_reset_at"`
	WebAuthnIDRaw      string                `json:"-" db:"webauthn_id"`
	CreatedAt          time.Time             `json:"created_at" db:"created_at"`
	Bio                string                `json:"bio" db:"bio"`
	Wechat             string                `json:"wechat" db:"wechat"`
	QQ                 string                `json:"qq" db:"qq"`
	ShowEmail          bool                  `json:"show_email" db:"show_email"`
	Birthday           *time.Time            `json:"birthday" db:"birthday"`
	SoundVolume        float64               `json:"sound_volume" db:"sound_volume"`
	VibrationEnabled   bool                  `json:"vibration_enabled" db:"vibration_enabled"`
	EnableElementInput bool                  `json:"enable_element_input" db:"enable_element_input"`
	CustomContact      string                `json:"custom_contact" db:"custom_contact"`
	Credentials        []webauthn.Credential `json:"-" db:"-"`
}

func NormalizeRole(role string) string {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "admin":
		return "admin"
	case "co-worker", "coworker", "co_worker":
		return "co-worker"
	default:
		return "user"
	}
}

func RoleHasAdminAccess(role string) bool {
	return NormalizeRole(role) == "admin"
}

func RoleHasCoWorkerAccess(role string) bool {
	normalized := NormalizeRole(role)
	return normalized == "admin" || normalized == "co-worker"
}

func (u *User) NormalizedRole() string {
	return NormalizeRole(u.Role)
}

func (u *User) HasAdminAccess() bool {
	return RoleHasAdminAccess(u.Role)
}

func (u *User) HasCoWorkerAccess() bool {
	return RoleHasCoWorkerAccess(u.Role)
}

func (u *User) WebAuthnID() []byte {
	return []byte(u.WebAuthnIDRaw)
}

func (u *User) WebAuthnName() string {
	if u.Email != "" {
		return u.Email
	}
	return u.Username
}

func (u *User) WebAuthnDisplayName() string {
	if u.Email != "" {
		return u.Email
	}
	return u.Username
}

func (u *User) WebAuthnIcon() string {
	return u.Avatar
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

type UserCredential struct {
	ID              []byte    `json:"id" db:"id"`
	UserUID         int       `json:"user_uid" db:"user_uid"`
	PublicKey       []byte    `json:"public_key" db:"public_key"`
	AttestationType string    `json:"attestation_type" db:"attestation_type"`
	Transport       string    `json:"transport" db:"transport"`
	SignCount       uint32    `json:"sign_count" db:"sign_count"`
	UserPresent     bool      `json:"user_present" db:"user_present"`
	UserVerified    bool      `json:"user_verified" db:"user_verified"`
	BackupEligible  bool      `json:"backup_eligible" db:"backup_eligible"`
	BackupState     bool      `json:"backup_state" db:"backup_state"`
	CloneWarning    bool      `json:"clone_warning" db:"clone_warning"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type RegisterRequest struct {
	Username         string `json:"username" binding:"required,min=3,max=30"`
	Email            string `json:"email"` // 可选
	Nickname         string `json:"nickname" binding:"max=20"`
	Password         string `json:"password" binding:"required,min=6"`
	Code             string `json:"code"` // 邮箱验证码（仅当提供邮箱且SMTP开启时需要）
	SecurityQuestion string `json:"security_question" binding:"required,min=1,max=200"`
	SecurityAnswer   string `json:"security_answer" binding:"required,min=1,max=100"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"` // 用户名或邮箱
	Password   string `json:"password" binding:"required"`
}

type SendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Type  string `json:"type"` // "register" or "reset"
}

type ResetPasswordByEmailRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type AdminChangePasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type PromoteUserRequest struct {
	Role string `json:"role" binding:"required,oneof=user co-worker admin"`
}

type Bounty struct {
	ID        int       `json:"id" db:"id"`
	TargetUID int       `json:"target_uid" db:"target_uid"`
	Amount    int       `json:"amount" db:"amount"`
	IssuerUID int       `json:"issuer_uid" db:"issuer_uid"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Reaction struct {
	ID           int       `json:"id" db:"id"`
	R1           string    `json:"r1" db:"r1"`
	R2           string    `json:"r2" db:"r2"`
	Display      string    `json:"display" db:"display"`
	Status       string    `json:"status" db:"status"`
	GroupID      string    `json:"group_id" db:"group_id"`
	CreatedByUID int       `json:"created_by_uid" db:"created_by_uid"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type Feedback struct {
	ID             int     `json:"id" db:"id"`
	UserUID        int     `json:"user_uid" db:"user_uid"`
	Username       string  `json:"username"`
	Content        string  `json:"content" binding:"required"`
	Type           string  `json:"type" db:"type"`
	Status         string  `json:"status" db:"status"`
	ProcessedBy    *int    `json:"processed_by" db:"processed_by"`
	ProcessedAt    *string `json:"processed_at" db:"processed_at"`
	LastUrgedAt    *string `json:"last_urged_at" db:"last_urged_at"`
	UrgeCount      int     `json:"urge_count" db:"urge_count"`
	ResolutionNote *string `json:"resolution_note" db:"resolution_note"`
	RemoveAt       *string `json:"remove_at" db:"remove_at"`
	CreatedAt      string  `json:"created_at" db:"created_at"`
}

type ReactionRequest struct {
	Display string `json:"display" binding:"required"`
}

type UpdateAvatarRequest struct {
	Avatar string `json:"avatar" binding:"required"`
}

type UpdateProfileRequest struct {
	Nickname           *string    `json:"nickname"`
	Bio                string     `json:"bio"`
	Wechat             string     `json:"wechat"`
	QQ                 string     `json:"qq"`
	ShowEmail          bool       `json:"show_email"`
	Birthday           *time.Time `json:"birthday"`
	SoundVolume        *float64   `json:"sound_volume"`
	VibrationEnabled   *bool      `json:"vibration_enabled"`
	EnableElementInput *bool      `json:"enable_element_input"`
	CustomContact      string     `json:"custom_contact"`
}

type ChangeEmailRequest struct {
	OldCode  string `json:"old_code" binding:"required"`
	NewEmail string `json:"new_email" binding:"required,email"`
	NewCode  string `json:"new_code" binding:"required"`
}

// SetEmailRequest 无邮箱用户设置首个邮箱
type SetEmailRequest struct {
	NewEmail       string `json:"new_email" binding:"required,email"`
	NewCode        string `json:"new_code" binding:"required"` // 新邮箱验证码
	SecurityAnswer string `json:"security_answer"`             // 密保验证（无2FA时需要）
}

// VerifySecurityAnswerRequest 验证密保答案（用于敏感操作）
type VerifySecurityAnswerRequest struct {
	SecurityAnswer string `json:"security_answer" binding:"required"`
}

// ResetPasswordBySecurityQuestionRequest 通过密保问题重置密码（忘记密码且无邮箱）
type ResetPasswordBySecurityQuestionRequest struct {
	Username       string `json:"username" binding:"required"`
	SecurityAnswer string `json:"security_answer" binding:"required"`
	NewPassword    string `json:"new_password" binding:"required,min=6"`
}

// UpdateSecurityQuestionRequest 更新密保问题和答案
type UpdateSecurityQuestionRequest struct {
	SecurityQuestion string `json:"security_question" binding:"required,min=1,max=200"`
	SecurityAnswer   string `json:"security_answer" binding:"required,min=1,max=100"`
	CurrentPassword  string `json:"current_password"` // 用于验证身份
}

// LevelInfo 等级信息
type LevelInfo struct {
	Level           int    `json:"level"`
	XP              int    `json:"xp"`
	TotalXP         int    `json:"total_xp"`
	Tier            string `json:"tier"`
	TierName        string `json:"tier_name"`
	NextLevelXP     int    `json:"next_level_xp"`
	ProgressPercent int    `json:"progress_percent"`
}
