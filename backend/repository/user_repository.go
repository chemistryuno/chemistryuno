package repository

import (
	"chemistryuno/database"
	"errors"
	"time"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository() *UserRepository {
	return &UserRepository{db: database.DB}
}

// FindByID 根据UID查找用户
func (r *UserRepository) FindByID(uid uint) (*database.User, error) {
	var user database.User
	err := r.db.First(&user, uid).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名查找用户
func (r *UserRepository) FindByUsername(username string) (*database.User, error) {
	var user database.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// Exists 检查用户名是否存在
func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&database.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// Create 创建新用户
func (r *UserRepository) Create(user *database.User) error {
	return r.db.Create(user).Error
}

// Update 更新用户
func (r *UserRepository) Update(user *database.User) error {
	return r.db.Save(user).Error
}

// UpdatePassword 更新密码
func (r *UserRepository) UpdatePassword(uid uint, newPasswordHash string) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).Update("password", newPasswordHash).Error
}

// UpdateAvatar 更新头像
func (r *UserRepository) UpdateAvatar(uid uint, avatar string) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).Update("avatar", avatar).Error
}

// Delete 删除用户（软删除）
func (r *UserRepository) Delete(uid uint) error {
	return r.db.Delete(&database.User{}, uid).Error
}

// CheckBanStatus 检查封禁状态
func (r *UserRepository) CheckBanStatus(uid uint) (bannedUntil *time.Time, frozenUntil *time.Time, banReason string, err error) {
	var user database.User
	err = r.db.Select("banned_until, frozen_until, ban_reason").First(&user, uid).Error
	if err != nil {
		return nil, nil, "", err
	}
	return user.BannedUntil, user.FrozenUntil, user.BanReason, nil
}

// UpdateBanStatus 更新封禁状态
func (r *UserRepository) UpdateBanStatus(uid uint, bannedUntil *time.Time) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).Update("banned_until", bannedUntil).Error
}

// UpdateBanStatusWithReason 更新封禁状态及原因
func (r *UserRepository) UpdateBanStatusWithReason(uid uint, bannedUntil *time.Time, reason string) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).Updates(map[string]interface{}{
		"banned_until": bannedUntil,
		"ban_reason":   reason,
	}).Error
}

// AddPoints 增加用户积分
func (r *UserRepository) AddPoints(uid uint, points int) error {
	return r.db.Model(&database.User{}).
		Where("uid = ?", uid).
		Update("points", gorm.Expr("points + ?", points)).Error
}

// UpdateFreezeStatus 更新冻结状态
func (r *UserRepository) UpdateFreezeStatus(uid uint, frozenUntil *time.Time) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).Update("frozen_until", frozenUntil).Error
}

// IncrementPoints 增加积分
func (r *UserRepository) IncrementPoints(uid uint, points int) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).
		Updates(map[string]interface{}{
			"points":         gorm.Expr("points + ?", points),
			"monthly_points": gorm.Expr("monthly_points + ?", points),
		}).Error
}

// IncrementMonthlyPoints 增加月度积分
func (r *UserRepository) IncrementMonthlyPoints(uid uint, points int) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).
		Update("monthly_points", gorm.Expr("monthly_points + ?", points)).Error
}

// ResetMonthlyPointsIfNeeded 如果需要，重置月度积分
func (r *UserRepository) ResetMonthlyPointsIfNeeded() error {
	now := time.Now()
	// 获取上次重置时间，如果是新月份则重置
	var lastReset time.Time
	err := r.db.Model(&database.User{}).
		Select("MAX(last_monthly_reset_at)").
		Scan(&lastReset).Error
	if err != nil {
		return err
	}

	// 如果是新月份，重置所有用户的月度积分
	if lastReset.Month() != now.Month() || lastReset.Year() != now.Year() {
		return r.db.Model(&database.User{}).
			Updates(map[string]interface{}{
				"monthly_points":        0,
				"last_monthly_reset_at": now,
			}).Error
	}
	return nil
}

// DecayTopPlayersPoints 对排名前列的用户进行积分衰减
func (r *UserRepository) DecayTopPlayersPoints(topCount int) error {
	// 获取排名阈值
	threshold, err := r.GetTopPointsThreshold(topCount)
	if err != nil {
		return err
	}

	// 对排名前列的用户积分衰减10%
	now := time.Now()
	return r.db.Model(&database.User{}).
		Where("points >= ? AND points > 0", threshold).
		Updates(map[string]interface{}{
			"points":               gorm.Expr("CAST(points * 0.9 AS INTEGER)"),
			"last_weekly_decay_at": now,
		}).Error
}

// IncrementGameStats 增加游戏统计
func (r *UserRepository) IncrementTotalGames(uid uint) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).
		Update("total_games", gorm.Expr("total_games + 1")).Error
}

func (r *UserRepository) IncrementWinCount(uid uint) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).
		Update("win_count", gorm.Expr("win_count + 1")).Error
}

// Update2FASecret 更新2FA密钥
func (r *UserRepository) Update2FASecret(uid uint, secret string) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).Update("two_factor_secret", secret).Error
}

// Enable2FA 启用2FA
func (r *UserRepository) Enable2FA(uid uint) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).Update("two_factor_enabled", true).Error
}

// Disable2FA 禁用2FA
func (r *UserRepository) Disable2FA(uid uint) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).
		Updates(map[string]interface{}{
			"two_factor_enabled": false,
			"two_factor_secret":  "",
		}).Error
}

// UpdateWebAuthnID 更新WebAuthn ID
func (r *UserRepository) UpdateWebAuthnID(uid uint, webAuthnID string) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).Update("webauthn_id", webAuthnID).Error
}

// GetAllUsers 获取所有用户（管理员功能）
func (r *UserRepository) GetAllUsers() ([]database.User, error) {
	var users []database.User
	err := r.db.Order("uid DESC").Find(&users).Error
	return users, err
}

// IncrementNegativePlayCount 增加消极游戏计数
func (r *UserRepository) IncrementNegativePlayCount(uid uint) (int, error) {
	var user database.User
	err := r.db.First(&user, uid).Error
	if err != nil {
		return 0, err
	}

	user.NegativePlayCount++
	err = r.db.Save(&user).Error
	return user.NegativePlayCount, err
}

// ResetNegativePlayCount 重置消极游戏计数
func (r *UserRepository) ResetNegativePlayCount(uid uint) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).Update("negative_play_count", 0).Error
}

// GetNegativePlayCount 获取消极游戏计数
func (r *UserRepository) GetNegativePlayCount(uid uint) (int, error) {
	var user database.User
	err := r.db.Select("negative_play_count").First(&user, uid).Error
	return user.NegativePlayCount, err
}

// UpdateNegativePlayCount 更新消极游戏计数
func (r *UserRepository) UpdateNegativePlayCount(uid uint, count int) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).Update("negative_play_count", count).Error
}

// DeductPoints 扣除积分
func (r *UserRepository) DeductPoints(uid uint, points int) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).
		Updates(map[string]interface{}{
			"points":         gorm.Expr("points - ?", points),
			"monthly_points": gorm.Expr("monthly_points - ?", points),
		}).Error
}

// DeductPointsPercentage 扣除积分百分比
func (r *UserRepository) DeductPointsPercentage(uid uint, percentage int) error {
	multiplier := float64(100-percentage) / 100.0
	return r.db.Model(&database.User{}).Where("uid = ?", uid).
		Updates(map[string]interface{}{
			"points":         gorm.Expr("CAST(points * ? AS INTEGER)", multiplier),
			"monthly_points": gorm.Expr("CAST(monthly_points * ? AS INTEGER)", multiplier),
		}).Error
}

// MultiplyPoints 积分乘以系数
func (r *UserRepository) MultiplyPoints(uid uint, multiplier float64) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).
		Update("points", gorm.Expr("CAST(points * ? AS INTEGER)", multiplier)).Error
}

// ResetMonthlyPoints 重置月度积分
func (r *UserRepository) ResetMonthlyPoints() error {
	now := time.Now()
	return r.db.Model(&database.User{}).
		Where("strftime('%Y-%m', last_monthly_reset_at) != ?", now.Format("2006-01")).
		Updates(map[string]interface{}{
			"monthly_points":        0,
			"last_monthly_reset_at": now,
		}).Error
}

// GetTopPointsThreshold 获取前N名的积分阈值
func (r *UserRepository) GetTopPointsThreshold(topCount int) (int, error) {
	var threshold int
	err := r.db.Model(&database.User{}).
		Select("points").
		Order("points DESC").
		Limit(1).
		Offset(topCount - 1).
		Scan(&threshold).Error
	return threshold, err
}

// DecayPointsForLowRankers 对排名靠后的用户进行积分衰减
func (r *UserRepository) DecayPointsForLowRankers(threshold int) error {
	now := time.Now()
	return r.db.Model(&database.User{}).
		Where("points < ? AND points > 0", threshold).
		Updates(map[string]interface{}{
			"points":               gorm.Expr("CAST(points * 0.9 AS INTEGER)"),
			"last_weekly_decay_at": now,
		}).Error
}

// GetUserCount 获取用户总数
func (r *UserRepository) GetUserCount() (int64, error) {
	var count int64
	err := r.db.Model(&database.User{}).Count(&count).Error
	return count, err
}

// GetAllUsersOrderByCreatedAt 获取所有用户，按创建时间倒序
func (r *UserRepository) GetAllUsersOrderByCreatedAt() ([]database.User, error) {
	var users []database.User
	err := r.db.Select("uid, username, avatar, is_admin, role, created_at").
		Order("created_at DESC").
		Find(&users).Error
	return users, err
}

// DeleteNonAdmin 删除非管理员用户
func (r *UserRepository) DeleteNonAdmin(uid uint) error {
	return r.db.Where("uid = ? AND is_admin = ?", uid, false).Delete(&database.User{}).Error
}

// UpdateRole 更新用户角色和管理员状态
func (r *UserRepository) UpdateRole(uid uint, role string, isAdmin bool) error {
	return r.db.Model(&database.User{}).
		Where("uid = ?", uid).
		Updates(map[string]interface{}{
			"role":     role,
			"is_admin": isAdmin,
		}).Error
}

// FindIsAdminByID 根据UID查找是否是管理员
func (r *UserRepository) FindIsAdminByID(uid uint) (bool, error) {
	var isAdmin bool
	err := r.db.Model(&database.User{}).
		Select("is_admin").
		Where("uid = ?", uid).
		Scan(&isAdmin).Error
	return isAdmin, err
}

// SearchUsers 搜索用户 (通过UID或用户名)
func (r *UserRepository) SearchUsers(query string) ([]database.User, error) {
	var users []database.User
	
	dbQuery := r.db.Select("uid, username, avatar, points, win_count, total_games")
	
	// 尝试将 query 解析为数字 (UID)
	var uid uint
	_, err := fmt.Sscanf(query, "%d", &uid)
	
	if err == nil && uid > 0 {
		// 如果是数字且大于0，则同时搜索 UID 和用户名
		err = dbQuery.Where("uid = ? OR username LIKE ?", uid, "%"+query+"%").Limit(20).Find(&users).Error
	} else {
		// 否则仅模糊搜索用户名
		err = dbQuery.Where("username LIKE ?", "%"+query+"%").Limit(20).Find(&users).Error
	}
	
	return users, err
}

// GetLeaderboard 获取排行榜
func (r *UserRepository) GetLeaderboard(orderBy string, limit int) ([]database.User, error) {
	var users []database.User
	err := r.db.Order(orderBy + " DESC").Limit(limit).Find(&users).Error
	return users, err
}
