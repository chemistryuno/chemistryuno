package repository

import (
	"chemistryuno/backend/database"

	"gorm.io/gorm"
)

// LeaderboardEntry 排行榜条目 - 只包含必要的字段
type LeaderboardEntry struct {
	UID           uint    `json:"uid" gorm:"column:uid"`
	Username      string  `json:"username" gorm:"column:username"`
	Nickname      string  `json:"nickname" gorm:"column:nickname"`
	Avatar        string  `json:"avatar" gorm:"column:avatar"`
	Points        float64 `json:"points" gorm:"column:points"`
	MonthlyPoints float64 `json:"monthly_points" gorm:"column:monthly_points"`
	Level         int     `json:"level" gorm:"column:level"`
	TotalXP       int     `json:"total_xp" gorm:"column:total_xp"`
	WinCount      int     `json:"win_count" gorm:"column:win_count"`
	TotalGames    int     `json:"total_games" gorm:"column:total_games"`
}

// TableName 指定表名
func (LeaderboardEntry) TableName() string {
	return "users"
}

// GetLeaderboardOptimized 获取排行榜 - 只查询必要的字段
func (r *UserRepository) GetLeaderboardOptimized(orderBy string, limit int) ([]LeaderboardEntry, error) {
	var entries []LeaderboardEntry

	safeOrderBy := "points"
	switch orderBy {
	case "points", "monthly_points", "total_xp", "win_count", "total_games", "created_at", "uid":
		safeOrderBy = orderBy
	}

	// 只查询必要的字段
	err := r.db.
		Select("uid, username, nickname, avatar, points, monthly_points, level, total_xp, win_count, total_games").
		Order(safeOrderBy + " DESC, uid ASC").
		Limit(limit).
		Find(&entries).Error

	return entries, err
}

// SearchUsersOptimized 搜索用户 - 只查询必要字段
type SearchUserEntry struct {
	UID           uint    `json:"uid" gorm:"column:uid"`
	Username      string  `json:"username" gorm:"column:username"`
	Nickname      string  `json:"nickname" gorm:"column:nickname"`
	Avatar        string  `json:"avatar" gorm:"column:avatar"`
	Level         int     `json:"level" gorm:"column:level"`
	Points        float64 `json:"points" gorm:"column:points"`
	MonthlyPoints float64 `json:"monthly_points" gorm:"column:monthly_points"`
	TotalXP       int     `json:"total_xp" gorm:"column:total_xp"`
	BannedUntil   *string `json:"banned_until" gorm:"column:banned_until"` // 简化为字符串
}

func (SearchUserEntry) TableName() string {
	return "users"
}

func (r *UserRepository) SearchUsersOptimized(query string) ([]SearchUserEntry, error) {
	var users []SearchUserEntry

	// 只查询必要的字段
	err := r.db.
		Select("uid, username, nickname, avatar, level, points, monthly_points, total_xp, banned_until").
		Where("username LIKE ? OR nickname LIKE ?", "%"+query+"%", "%"+query+"%").
		Limit(50).
		Find(&users).Error

	return users, err
}

// GetAllUsersOptimized 获取所有用户 - 只查询必要字段
type UserListEntry struct {
	UID         uint    `json:"uid" gorm:"column:uid"`
	Username    string  `json:"username" gorm:"column:username"`
	Nickname    string  `json:"nickname" gorm:"column:nickname"`
	Email       string  `json:"email" gorm:"column:email"`
	Avatar      string  `json:"avatar" gorm:"column:avatar"`
	IsAdmin     bool    `json:"is_admin" gorm:"column:is_admin"`
	Role        string  `json:"role" gorm:"column:role"`
	Level       int     `json:"level" gorm:"column:level"`
	Points      float64 `json:"points" gorm:"column:points"`
	TotalGames  int     `json:"total_games" gorm:"column:total_games"`
	BannedUntil *string `json:"banned_until" gorm:"column:banned_until"`
	CreatedAt   string  `json:"created_at" gorm:"column:created_at"`
}

func (UserListEntry) TableName() string {
	return "users"
}

func (r *UserRepository) GetAllUsersOptimized() ([]UserListEntry, error) {
	var users []UserListEntry

	err := r.db.
		Select("uid, username, nickname, email, avatar, is_admin, role, level, points, total_games, banned_until, created_at").
		Order("created_at DESC").
		Find(&users).Error

	return users, err
}

// GetUserBasicInfo 获取用户基本信息 - 用于列表展示
type UserBasicInfo struct {
	UID      uint    `json:"uid" gorm:"column:uid"`
	Username string  `json:"username" gorm:"column:username"`
	Nickname string  `json:"nickname" gorm:"column:nickname"`
	Avatar   string  `json:"avatar" gorm:"column:avatar"`
	Level    int     `json:"level" gorm:"column:level"`
	Points   float64 `json:"points" gorm:"column:points"`
}

func (UserBasicInfo) TableName() string {
	return "users"
}

func (r *UserRepository) GetUserBasicInfo(uid uint) (*UserBasicInfo, error) {
	var user UserBasicInfo

	err := r.db.
		Select("uid, username, nickname, avatar, level, points").
		Where("uid = ?", uid).
		First(&user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// GetTopUsers 获取积分最高的N个用户
func (r *UserRepository) GetTopUsers(limit int, includeFields string) ([]database.User, error) {
	var users []database.User

	// 如果指定了字段，只查询这些字段
	if includeFields != "" {
		err := r.db.
			Select(includeFields).
			Order("points DESC").
			Limit(limit).
			Find(&users).Error
		return users, err
	}

	// 否则查询常用字段
	err := r.db.
		Select("uid, username, nickname, avatar, level, points, monthly_points, total_games, win_count").
		Order("points DESC").
		Limit(limit).
		Find(&users).Error

	return users, err
}
