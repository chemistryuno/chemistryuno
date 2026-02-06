package repository

import (
	"chemistryuno/database"
	"time"

	"gorm.io/gorm"
)

type FeedbackRepository struct {
	db *gorm.DB
}

func NewFeedbackRepository() *FeedbackRepository {
	return &FeedbackRepository{db: database.DB}
}

// Create 创建反馈
func (r *FeedbackRepository) Create(feedback *database.Feedback) error {
	return r.db.Create(feedback).Error
}

// FindByID 根据ID查找反馈
func (r *FeedbackRepository) FindByID(id uint) (*database.Feedback, error) {
	var feedback database.Feedback
	err := r.db.First(&feedback, id).Error
	if err != nil {
		return nil, err
	}
	return &feedback, nil
}

// FindByUserUID 查找用户的所有反馈
func (r *FeedbackRepository) FindByUserUID(uid uint) ([]database.Feedback, error) {
	var feedbacks []database.Feedback
	err := r.db.Where("user_uid = ?", uid).
		Order("created_at DESC").
		Find(&feedbacks).Error
	return feedbacks, err
}

// FindAll 查找所有反馈（管理员）
func (r *FeedbackRepository) FindAll() ([]database.Feedback, error) {
	var feedbacks []database.Feedback
	err := r.db.Order("created_at DESC").Find(&feedbacks).Error
	return feedbacks, err
}

// UpdateStatus 更新反馈状态
func (r *FeedbackRepository) UpdateStatus(id uint, status string, processedBy uint, note string) error {
	now := time.Now()
	return r.db.Model(&database.Feedback{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":           status,
			"processed_by_uid": processedBy,
			"processed_at":     now,
			"resolution_note":  note,
		}).Error
}

// UpdateUrge 更新催促信息
func (r *FeedbackRepository) UpdateUrge(id uint) error {
	now := time.Now()
	return r.db.Model(&database.Feedback{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_urged_at": now,
			"urge_count":    gorm.Expr("urge_count + 1"),
		}).Error
}

// SetRemovalTime 设置删除时间
func (r *FeedbackRepository) SetRemovalTime(id uint, removeAt time.Time) error {
	return r.db.Model(&database.Feedback{}).Where("id = ?", id).
		Update("remove_at", removeAt).Error
}

// Delete 删除反馈
func (r *FeedbackRepository) Delete(id uint) error {
	return r.db.Delete(&database.Feedback{}, id).Error
}

// DeleteExpired 删除已过期的反馈
func (r *FeedbackRepository) DeleteExpired() (int64, error) {
	now := time.Now()
	result := r.db.Where("remove_at IS NOT NULL AND remove_at <= ?", now).
		Delete(&database.Feedback{})
	return result.RowsAffected, result.Error
}
