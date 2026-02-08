package repository

import (
	"chemistryuno/database"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{db: database.DB}
}

// Create 创建会话
func (r *SessionRepository) Create(session *database.UserSession) error {
	return r.db.Create(session).Error
}

// FindByID 根据ID查找会话
func (r *SessionRepository) FindByID(id string) (*database.UserSession, error) {
	var session database.UserSession
	err := r.db.Where("id = ?", id).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// FindByUserUID 查找用户的所有会话
func (r *SessionRepository) FindByUserUID(uid uint) ([]database.UserSession, error) {
	var sessions []database.UserSession
	err := r.db.Where("user_uid = ?", uid).
		Order("last_active DESC").
		Find(&sessions).Error
	return sessions, err
}

// UpdateActivity 更新会话活动时间和IP（带重试机制）
func (r *SessionRepository) UpdateActivity(id string, ip string) error {
	// 尝试更新，如果因为数据库锁定失败，最多重试3次
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		err := r.db.Model(&database.UserSession{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"last_active": time.Now(),
				"ip_address":  ip,
			}).Error

		if err == nil {
			return nil
		}

		lastErr = err
		// 如果是数据库锁定错误，等待后重试
		if isDatabaseBusyError(err) && i < maxRetries-1 {
			time.Sleep(time.Millisecond * time.Duration(50*(i+1))) // 递增退避
			continue
		}

		// 其他错误直接返回
		if i == maxRetries-1 {
			log.Printf("[Session更新失败] SID=%s, IP=%s: %v (已重试%d次)", id, ip, err, maxRetries)
		}
		break
	}

	return lastErr
}

// isDatabaseBusyError 检查是否是数据库忙碌/锁定错误
func isDatabaseBusyError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	// 检查SQLite常见的锁定错误
	return strings.Contains(errMsg, "database is locked") ||
		strings.Contains(errMsg, "SQLITE_BUSY") ||
		strings.Contains(errMsg, "database locked")
}

// Delete 删除会话
func (r *SessionRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&database.UserSession{}).Error
}

// DeleteByUserUID 删除用户的所有会话
func (r *SessionRepository) DeleteByUserUID(uid uint) error {
	return r.db.Where("user_uid = ?", uid).Delete(&database.UserSession{}).Error
}

// DeleteByIDAndUserUID 删除指定用户的指定会话
func (r *SessionRepository) DeleteByIDAndUserUID(id string, uid uint) error {
	return r.db.Where("id = ? AND user_uid = ?", id, uid).Delete(&database.UserSession{}).Error
}

// Exists 检查会话是否存在
func (r *SessionRepository) Exists(id string) (bool, error) {
	var count int64
	err := r.db.Model(&database.UserSession{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		// 数据库错误时，记录日志并返回错误而不是直接false
		log.Printf("[Session验证错误] 数据库查询失败 SID=%s: %v", id, err)
		return false, err
	}
	return count > 0, nil
}

// ValidateSessionForUser 验证会话是否属于指定用户
func (r *SessionRepository) ValidateSessionForUser(id string, uid uint) (bool, error) {
	var count int64
	err := r.db.Model(&database.UserSession{}).
		Where("id = ? AND user_uid = ?", id, uid).
		Count(&count).Error
	if err != nil {
		log.Printf("[Session验证错误] 验证用户会话失败 SID=%s, UID=%d: %v", id, uid, err)
		return false, err
	}
	return count > 0, nil
}

// CleanupInactive 清理24小时未活动的会话
func (r *SessionRepository) CleanupInactive() (int64, error) {
	result := r.db.Where("last_active < ?", time.Now().Add(-24*time.Hour)).
		Delete(&database.UserSession{})
	return result.RowsAffected, result.Error
}
