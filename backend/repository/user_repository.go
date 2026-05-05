package repository

import (
	"chemistryuno/backend/cache"
	"chemistryuno/backend/database"
	"chemistryuno/backend/models"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

var ErrFuelCompensationAlreadyIssued = errors.New("fuel compensation already issued")

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(dbs ...*gorm.DB) *UserRepository {
	db := database.DB
	if len(dbs) > 0 && dbs[0] != nil {
		db = dbs[0]
	}
	return &UserRepository{db: db}
}

// FindByUID 根据UID查找用户
func (r *UserRepository) FindByUID(uid uint) (*database.User, error) {
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

// FindByUIDs 根据UID列表批量查找用户，返回 uid->User 映射
func (r *UserRepository) FindByUIDs(uids []uint) (map[uint]*database.User, error) {
	if len(uids) == 0 {
		return map[uint]*database.User{}, nil
	}
	var users []database.User
	err := r.db.Where("uid IN ?", uids).Find(&users).Error
	if err != nil {
		return nil, err
	}
	userMap := make(map[uint]*database.User, len(users))
	for i := range users {
		userMap[users[i].UID] = &users[i]
	}
	return userMap, nil
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

// FindByEmail 根据邮箱查找用户
func (r *UserRepository) FindByEmail(email string) (*database.User, error) {
	var user database.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// ExistsByUsername 检查用户名是否存在
func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
	if username == "" {
		return false, nil
	}
	var count int64
	err := r.db.Model(&database.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// ExistsByEmail 检查邮箱是否存在
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	if email == "" {
		return false, nil
	}
	var count int64
	err := r.db.Model(&database.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// ExistsByNickname 检查昵称是否存在
func (r *UserRepository) ExistsByNickname(nickname string) (bool, error) {
	if nickname == "" {
		return false, nil
	}
	var count int64
	err := r.db.Model(&database.User{}).Where("nickname = ?", nickname).Count(&count).Error
	return count > 0, err
}

// FindUsersWithBlankNickname returns users whose nickname is empty or whitespace-only.
func (r *UserRepository) FindUsersWithBlankNickname() ([]database.User, error) {
	var users []database.User
	err := r.db.Select("uid, username, nickname").
		Where("TRIM(COALESCE(nickname, '')) = ''").
		Order("uid ASC").
		Find(&users).Error
	return users, err
}

// FindByEmail 按邮箱或用户名查找用户
func (r *UserRepository) FindByEmailOrUsername(identifier string) (*database.User, error) {
	var user database.User
	err := r.db.Where("email = ? OR username = ?", identifier, identifier).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// FindByGithubID 根据 GitHub ID 查找用户
func (r *UserRepository) FindByGithubID(githubID string) (*database.User, error) {
	var user database.User
	err := r.db.Where("github_id = ?", githubID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByMicrosoftID 根据 Microsoft ID 查找用户
func (r *UserRepository) FindByMicrosoftID(microsoftID string) (*database.User, error) {
	var user database.User
	err := r.db.Where("microsoft_id = ?", microsoftID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByGoogleID 根据 Google ID 查找用户
func (r *UserRepository) FindByGoogleID(googleID string) (*database.User, error) {
	var user database.User
	err := r.db.Where("google_id = ?", googleID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByAppleID 根据 Apple ID 查找用户
func (r *UserRepository) FindByAppleID(appleID string) (*database.User, error) {
	var user database.User
	err := r.db.Where("apple_id = ?", appleID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
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

// UpdateNickname 更新昵称
func (r *UserRepository) UpdateNickname(uid uint, nickname string) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).Update("nickname", nickname).Error
}

// UpdateProfile 更新个人资料
func (r *UserRepository) UpdateProfile(uid uint, req *models.UpdateProfileRequest) error {
	updates := map[string]interface{}{
		"nickname":       req.Nickname,
		"bio":            req.Bio,
		"wechat":         req.Wechat,
		"qq":             req.QQ,
		"show_email":     req.ShowEmail,
		"custom_contact": req.CustomContact,
		"birthday":       req.Birthday,
	}

	if req.SoundVolume != nil {
		updates["sound_volume"] = *req.SoundVolume
	}
	if req.VibrationEnabled != nil {
		updates["vibration_enabled"] = *req.VibrationEnabled
	}
	if req.EnableElementInput != nil {
		updates["enable_element_input"] = *req.EnableElementInput
	}

	return r.db.Model(&database.User{}).Where("uid = ?", uid).Updates(updates).Error
}

// Delete 删除用户（软删除）
func (r *UserRepository) Delete(uid uint) error {
	return r.db.Unscoped().Delete(&database.User{}, uid).Error
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
	err := r.db.Model(&database.User{}).
		Where("uid = ?", uid).
		Update("points", gorm.Expr("points + ?", points)).Error

	if err == nil {
		// 失效排行榜缓存
		_ = cache.InvalidateAllLeaderboardCache(context.Background())
		// 失效用户统计缓存
		_ = cache.InvalidateUserStatsCache(context.Background(), uid)
	}

	return err
}

// UpdateFreezeStatus 更新冻结状态
func (r *UserRepository) UpdateFreezeStatus(uid uint, frozenUntil *time.Time) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).Update("frozen_until", frozenUntil).Error
}

// UpdateRoomReadyStatus 更新房间准备状态
func (r *UserRepository) UpdateRoomReadyStatus(uid uint, ready bool) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).Update("room_ready", ready).Error
}

// IncrementPoints 增加积分
func (r *UserRepository) IncrementPoints(uid uint, points int) error {
	err := r.db.Model(&database.User{}).Where("uid = ?", uid).
		Updates(map[string]interface{}{
			"points":         gorm.Expr("points + ?", points),
			"monthly_points": gorm.Expr("monthly_points + ?", points),
		}).Error

	if err == nil {
		// 失效排行榜缓存
		_ = cache.InvalidateAllLeaderboardCache(context.Background())
		// 失效用户统计缓存
		_ = cache.InvalidateUserStatsCache(context.Background(), uid)
	}

	return err
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
				"points":                1000,
				"monthly_points":        1000,
				"last_monthly_reset_at": now,
			}).Error
	}
	return nil
}

// DecayTopPlayersPoints 对排名前列的用户进行积分衰减
func (r *UserRepository) DecayTopPlayersPoints(percentage int) error {
	// 1. 获取用户总数来计算百分比对应的实际人数
	totalCount, err := r.GetUserCount()
	if err != nil {
		return err
	}
	if totalCount == 0 {
		return nil
	}

	// 计算前 N% 的实际人数 (至少1人)
	topCount := int(totalCount * int64(percentage) / 100)
	if topCount == 0 {
		topCount = 1
	}

	// 2. 获取排名阈值
	threshold, err := r.GetTopPointsThreshold(topCount)
	if err != nil {
		return err
	}

	// 3. 对排名前列且超过一周未衰减的用户积分衰减 2%
	now := time.Now()
	oneWeekAgo := now.AddDate(0, 0, -7)

	return r.db.Model(&database.User{}).
		Where("points >= ? AND points > 0 AND last_weekly_decay_at < ?", threshold, oneWeekAgo).
		Updates(map[string]interface{}{
			"points":               gorm.Expr("points * 0.98"),
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
	updateResult := r.db.Model(&database.User{}).
		Where("uid = ?", uid).
		Update("negative_play_count", gorm.Expr("negative_play_count + 1"))
	if updateResult.Error != nil {
		return 0, updateResult.Error
	}
	if updateResult.RowsAffected == 0 {
		return 0, gorm.ErrRecordNotFound
	}

	var count int
	err := r.db.Model(&database.User{}).
		Select("negative_play_count").
		Where("uid = ?", uid).
		Scan(&count).Error
	return count, err
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

// UpdateLastOfflineAt 更新最后离线时间
func (r *UserRepository) UpdateLastOfflineAt(uid uint, t time.Time) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).Update("last_offline_at", t).Error
}

// UpdateTurnStartedAt 更新回合开始时间
func (r *UserRepository) UpdateTurnStartedAt(uid uint, t time.Time) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).Update("turn_started_at", t).Error
}

// GetUserReconnectionData 获取用于重连检查的数据
func (r *UserRepository) GetUserReconnectionData(uid uint) (*time.Time, *time.Time, error) {
	var user database.User
	err := r.db.Select("turn_started_at, last_offline_at").First(&user, uid).Error
	if err != nil {
		return nil, nil, err
	}
	return user.TurnStartedAt, user.LastOfflineAt, nil
}

// DeductPointsPercentage 扣除积分百分比
func (r *UserRepository) DeductPointsPercentage(uid uint, percentage int) error {
	multiplier := float64(100-percentage) / 100.0
	return r.db.Model(&database.User{}).Where("uid = ?", uid).
		Updates(map[string]interface{}{
			"points":         gorm.Expr("points * ?", multiplier),
			"monthly_points": gorm.Expr("monthly_points * ?", multiplier),
		}).Error
}

// MultiplyPoints 积分乘以系数
func (r *UserRepository) MultiplyPoints(uid uint, multiplier float64) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).
		Update("points", gorm.Expr("points * ?", multiplier)).Error
}

// ResetMonthlyPoints 重置月度积分 (设为初始1000)
func (r *UserRepository) ResetMonthlyPoints() error {
	now := time.Now()
	// 获取本月第一天
	beginningOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return r.db.Model(&database.User{}).
		Where("last_monthly_reset_at < ?", beginningOfMonth).
		Updates(map[string]interface{}{
			"points":                1000,
			"monthly_points":        1000,
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
			"points":               gorm.Expr("points * 0.9"),
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
	err := r.db.Select("uid, username, nickname, avatar, is_admin, role, created_at, banned_until, ban_reason").
		Order("created_at DESC").
		Find(&users).Error
	return users, err
}

// DeleteNonAdmin 删除非管理员用户
func (r *UserRepository) DeleteNonAdmin(uid uint) error {
	return r.db.Unscoped().Where("uid = ?", uid).Delete(&database.User{}).Error
}

// UpdateRole 更新用户角色，并同步兼容 is_admin 字段
func (r *UserRepository) UpdateRole(uid uint, role string) error {
	normalizedRole := models.NormalizeRole(role)
	return r.db.Model(&database.User{}).
		Where("uid = ?", uid).
		Updates(map[string]interface{}{
			"role":     normalizedRole,
			"is_admin": models.RoleHasAdminAccess(normalizedRole),
		}).Error
}

// SearchUsers 搜索用户 (通过UID或用户名)
func (r *UserRepository) SearchUsers(query string) ([]database.User, error) {
	users := []database.User{}

	// 使用模型进行查询，确保 GORM 正确处理字段映射和软删除
	dbQuery := r.db.Model(&database.User{})

	// 尝试将 query 解析为数字 (UID)
	var uid uint
	isNumeric := false
	uidValue, err := strconv.ParseUint(query, 10, 32)
	if err == nil {
		uid = uint(uidValue)
		isNumeric = true
	}

	q := strings.ToLower(query)
	if isNumeric && uid > 0 {
		// 如果是数字且大于0，则优先精准匹配 UID，同时模糊匹配用户名和昵称
		err = dbQuery.Where("uid = ? OR LOWER(username) LIKE ? OR LOWER(nickname) LIKE ?", uid, "%"+q+"%", "%"+q+"%").
			Order(fmt.Sprintf("CASE WHEN uid = %d THEN 0 ELSE 1 END", uid)).
			Limit(20).Find(&users).Error
	} else {
		// 否则模糊搜索用户名和昵称（不区分大小写）
		err = dbQuery.Where("LOWER(username) LIKE ? OR LOWER(nickname) LIKE ?", "%"+q+"%", "%"+q+"%").Limit(20).Find(&users).Error
	}

	return users, err
}

// GetLeaderboard 获取排行榜
func (r *UserRepository) GetLeaderboard(orderBy string, limit int) ([]database.User, error) {
	var users []database.User
	safeOrderBy := "points"
	switch orderBy {
	case "points", "monthly_points", "total_xp", "win_count", "total_games", "created_at", "uid":
		safeOrderBy = orderBy
	}
	err := r.db.Order(safeOrderBy + " DESC").Limit(limit).Find(&users).Error
	return users, err
}

// UpdateXP 更新经验值和等级
func (r *UserRepository) UpdateXP(uid uint, xp int, totalXP int, level int) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).Updates(map[string]interface{}{
		"xp":       xp,
		"total_xp": totalXP,
		"level":    level,
	}).Error
}

// AddXP 增加经验值
func (r *UserRepository) AddXP(uid uint, xp int) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).
		Updates(map[string]interface{}{
			"xp":       gorm.Expr("xp + ?", xp),
			"total_xp": gorm.Expr("total_xp + ?", xp),
		}).Error
}

// GetMaxUID 获取当前最大 UID
func (r *UserRepository) GetMaxUID() (uint, error) {
	var maxUID uint
	err := r.db.Model(&database.User{}).Select("MAX(uid)").Scan(&maxUID).Error
	return maxUID, err
}

// UpdateSecurityQuestion 更新密保问题和哈希答案
func (r *UserRepository) UpdateSecurityQuestion(uid uint, question, hashedAnswer string) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).Updates(map[string]interface{}{
		"security_question": question,
		"security_answer":   hashedAnswer,
	}).Error
}

// UpdateEmail 设置或更新邮箱地址
func (r *UserRepository) UpdateEmail(uid uint, email string) error {
	return r.db.Model(&database.User{}).Where("uid = ?", uid).Update("email", email).Error
}

// AddFuel 向用户账户添加燃素（幂等操作）
// compensationID 用于幂等性检查，防止重复发放
func (r *UserRepository) AddFuel(uid uint, amount int, compensationID string) (int, error) {
	if amount <= 0 {
		return 0, errors.New("燃素金额必须大于0")
	}

	// 检查是否已经为该补偿ID发放过燃素
	if compensationID != "" {
		var count int64
		err := r.db.Model(&database.FuelCompensationRecord{}).
			Where("user_uid = ? AND compensation_id = ?", uid, compensationID).
			Count(&count).Error
		if err != nil {
			return 0, fmt.Errorf("检查燃素补偿记录失败: %w", err)
		}
		if count > 0 {
			return 0, ErrFuelCompensationAlreadyIssued
		}
	}

	// 在事务中执行燃素添加和记录
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Keep both legacy fuel and visible points in sync.
		if err := tx.Model(&database.User{}).Where("uid = ?", uid).Updates(map[string]interface{}{
			"fuel":   gorm.Expr("fuel + ?", amount),
			"points": gorm.Expr("points + ?", amount),
		}).Error; err != nil {
			return fmt.Errorf("更新燃素余额失败: %w", err)
		}

		// 记录燃素补偿历史
		record := &database.FuelCompensationRecord{
			UserUID:        uid,
			Amount:         amount,
			CompensationID: compensationID,
			CreatedAt:      time.Now(),
		}
		if err := tx.Create(record).Error; err != nil {
			return fmt.Errorf("记录燃素补偿失败: %w", err)
		}

		return nil
	})
	if err == nil {
		_ = cache.InvalidateAllLeaderboardCache(context.Background())
		_ = cache.InvalidateUserStatsCache(context.Background(), uid)
	}

	return amount, err
}
