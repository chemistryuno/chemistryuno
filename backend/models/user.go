package models

import (
	"time"
)

type User struct {
	UID              int       `json:"uid" db:"uid"`
	Username         string    `json:"username" db:"username"`
	PasswordHash     string    `json:"-" db:"password"` // 不返回给前端
	Avatar           string    `json:"avatar" db:"avatar"`
	IsAdmin          bool      `json:"is_admin" db:"is_admin"`
	Role             string    `json:"role" db:"role"` // admin, co-worker, user
	TwoFactorEnabled bool      `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorSecret  string    `json:"-" db:"two_factor_secret"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
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

type ChemicalReaction struct {
	ID        int       `json:"id" db:"id"`
	Reactant  string    `json:"reactant" db:"reactant"`
	Product   string    `json:"product" db:"product"`
	Type      string    `json:"type" db:"type"`
	CreatedBy int       `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type ReactionRequest struct {
	Reactant string `json:"reactant" binding:"required"`
	Product  string `json:"product" binding:"required"`
	Type     string `json:"type" binding:"required"`
}

type UpdateAvatarRequest struct {
	Avatar string `json:"avatar" binding:"required"`
}
