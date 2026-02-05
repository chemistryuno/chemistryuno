package models

import (
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
	IsAdmin            bool                  `json:"is_admin" db:"is_admin"`
	Role               string                `json:"role" db:"role"` // admin, co-worker, user
	TwoFactorEnabled   bool                  `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorSecret    string                `json:"-" db:"two_factor_secret"`
	Points             int                   `json:"points" db:"points"`
	MonthlyPoints      int                   `json:"monthly_points" db:"monthly_points"`
	NegativePlayCount  int                   `json:"negative_play_count" db:"negative_play_count"`
	BannedUntil        *time.Time            `json:"banned_until" db:"banned_until"`
	FrozenUntil        *time.Time            `json:"frozen_until" db:"frozen_until"`
	TotalGames         int                   `json:"total_games" db:"total_games"`
	WinCount           int                   `json:"win_count" db:"win_count"`
	LastWeeklyDecayAt  time.Time             `json:"last_weekly_decay_at" db:"last_weekly_decay_at"`
	LastMonthlyResetAt time.Time             `json:"last_monthly_reset_at" db:"last_monthly_reset_at"`
	WebAuthnIDRaw      string                `json:"-" db:"webauthn_id"`
	CreatedAt          time.Time             `json:"created_at" db:"created_at"`
	Credentials        []webauthn.Credential `json:"-" db:"-"`
}

func (u *User) WebAuthnID() []byte {
	return []byte(u.WebAuthnIDRaw)
}

func (u *User) WebAuthnName() string {
	return u.Username
}

func (u *User) WebAuthnDisplayName() string {
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
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname" binding:"required,min=1,max=20"`
	Password string `json:"password" binding:"required,min=6"`
	Code     string `json:"code"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password" binding:"required"`
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
	CreatedBy int       `json:"created_by" db:"created_by"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Reaction struct {
	ID        int       `json:"id" db:"id"`
	R1        string    `json:"r1" db:"r1"`
	R2        string    `json:"r2" db:"r2"`
	Display   string    `json:"display" db:"display"`
	Status    string    `json:"status" db:"status"`
	GroupID   string    `json:"group_id" db:"group_id"`
	CreatedBy int       `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Feedback struct {
	ID             int     `json:"id" db:"id"`
	UserID         int     `json:"user_id" db:"user_id"`
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
	Nickname string `json:"nickname" binding:"required,min=1,max=20"`
}

type ChangeEmailRequest struct {
	OldCode  string `json:"old_code" binding:"required"`
	NewEmail string `json:"new_email" binding:"required,email"`
	NewCode  string `json:"new_code" binding:"required"`
}
