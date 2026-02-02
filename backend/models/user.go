package models

import (
	"time"
)

type User struct {
	UID               int        `json:"uid" db:"uid"`
	Username          string     `json:"username" db:"username"`
	PasswordHash      string     `json:"-" db:"password"` // 不返回给前端
	Avatar            string     `json:"avatar" db:"avatar"`
	IsAdmin           bool       `json:"is_admin" db:"is_admin"`
	Role              string     `json:"role" db:"role"` // admin, co-worker, user
	TwoFactorEnabled  bool       `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorSecret   string     `json:"-" db:"two_factor_secret"`
	Points            int        `json:"points" db:"points"`
	NegativePlayCount int        `json:"negative_play_count" db:"negative_play_count"`
	BannedUntil       *time.Time `json:"banned_until" db:"banned_until"`
	LastDecayAt       time.Time  `json:"last_decay_at" db:"last_decay_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
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
