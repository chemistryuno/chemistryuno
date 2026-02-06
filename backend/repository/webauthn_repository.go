package repository

import (
	"chemistryuno/database"

	"gorm.io/gorm"
)

type WebAuthnRepository struct {
	db *gorm.DB
}

func NewWebAuthnRepository() *WebAuthnRepository {
	return &WebAuthnRepository{db: database.DB}
}

// Create 创建WebAuthn凭证
func (r *WebAuthnRepository) Create(credential *database.WebAuthnCredential) error {
	return r.db.Create(credential).Error
}

// FindByUserUID 查找用户的所有凭证
func (r *WebAuthnRepository) FindByUserUID(uid uint) ([]database.WebAuthnCredential, error) {
	var credentials []database.WebAuthnCredential
	err := r.db.Where("user_uid = ?", uid).Find(&credentials).Error
	return credentials, err
}

// FindByID 根据凭证ID查找
func (r *WebAuthnRepository) FindByID(id string) (*database.WebAuthnCredential, error) {
	var credential database.WebAuthnCredential
	err := r.db.Where("id = ?", id).First(&credential).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

// UpdateSignCount 更新签名计数
func (r *WebAuthnRepository) UpdateSignCount(id string, signCount uint32) error {
	return r.db.Model(&database.WebAuthnCredential{}).
		Where("id = ?", id).
		Update("sign_count", signCount).Error
}

// Delete 删除凭证
func (r *WebAuthnRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&database.WebAuthnCredential{}).Error
}

// DeleteByUserUID 删除用户的所有凭证
func (r *WebAuthnRepository) DeleteByUserUID(uid uint) error {
	return r.db.Where("user_uid = ?", uid).Delete(&database.WebAuthnCredential{}).Error
}
