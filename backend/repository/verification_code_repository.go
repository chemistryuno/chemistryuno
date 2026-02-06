package repository

import (
	"chemistryuno/database"
	"time"

	"gorm.io/gorm"
)

type VerificationCodeRepository struct {
	db *gorm.DB
}

func NewVerificationCodeRepository() *VerificationCodeRepository {
	return &VerificationCodeRepository{db: database.DB}
}

// FindFirst 查询有效的验证码
func (r *VerificationCodeRepository) FindFirst(email, code, codeType string) (*database.VerificationCode, error) {
	var vCode database.VerificationCode
	err := r.db.Where("email = ? AND code = ? AND type = ? AND expires_at > ?",
		email, code, codeType, time.Now()).First(&vCode).Error
	if err != nil {
		return nil, err
	}
	return &vCode, nil
}

// Delete 删除验证码
func (r *VerificationCodeRepository) Delete(vCode *database.VerificationCode) error {
	return r.db.Delete(vCode).Error
}

// Create 创建验证码
func (r *VerificationCodeRepository) Create(vCode *database.VerificationCode) error {
	return r.db.Create(vCode).Error
}

// HasRecentCode 检查最近一分钟内是否发送过验证码
func (r *VerificationCodeRepository) HasRecentCode(email string) (bool, error) {
	var count int64
	err := r.db.Model(&database.VerificationCode{}).
		Where("email = ? AND created_at > ?", email, time.Now().Add(-1*time.Minute)).
		Count(&count).Error
	return count > 0, err
}
