package repository

import (
	"chemistryuno/database"
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

// FindByUserID 查找用户的所有会话
func (r *SessionRepository) FindByUserID(uid uint) ([]database.UserSession, error) {
	var sessions []database.UserSession
	err := r.db.Where("user_uid = ?", uid).
		Order("last_active DESC").
		Find(&sessions).Error
	return sessions, err
}

// UpdateActivity 更新会话活动时间和IP
func (r *SessionRepository) UpdateActivity(id string, ip string) error {
	return r.db.Model(&database.UserSession{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_active": time.Now(),
			"ip_address":  ip,
		}).Error
}

// Delete 删除会话
func (r *SessionRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&database.UserSession{}).Error
}

// DeleteByUserID 删除用户的所有会话
func (r *SessionRepository) DeleteByUserID(uid uint) error {
	return r.db.Where("user_uid = ?", uid).Delete(&database.UserSession{}).Error
}

// DeleteByIDAndUserID 删除指定用户的指定会话
func (r *SessionRepository) DeleteByIDAndUserID(id string, uid uint) error {
	return r.db.Where("id = ? AND user_uid = ?", id, uid).Delete(&database.UserSession{}).Error
}

// Exists 检查会话是否存在
func (r *SessionRepository) Exists(id string) (bool, error) {
	var count int64
	err := r.db.Model(&database.UserSession{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// ValidateSessionForUser 验证会话是否属于指定用户
func (r *SessionRepository) ValidateSessionForUser(id string, uid uint) (bool, error) {
	var count int64
	err := r.db.Model(&database.UserSession{}).
		Where("id = ? AND user_uid = ?", id, uid).
		Count(&count).Error
	return count > 0, err
}

// CleanupInactive 清理24小时未活动的会话
func (r *SessionRepository) CleanupInactive() (int64, error) {
	result := r.db.Where("last_active < ?", time.Now().Add(-24*time.Hour)).
		Delete(&database.UserSession{})
	return result.RowsAffected, result.Error
}
