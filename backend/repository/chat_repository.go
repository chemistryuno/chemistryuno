package repository

import (
	"chemistryuno/backend/database"
	"time"

	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository() *ChatRepository {
	return &ChatRepository{db: database.DB}
}

func (r *ChatRepository) SaveChatMessage(uid uint, username, avatar, message string) error {
	chat := database.GlobalChat{
		UserUID:  uid,
		Username: username,
		Avatar:   avatar,
		Message:  message,
	}
	return r.db.Create(&chat).Error
}

func (r *ChatRepository) GetRecentMessages(limit int) ([]database.GlobalChat, error) {
	messages := []database.GlobalChat{}
	err := r.db.Order("created_at desc").Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// 反转顺序，让前端收到的顺序是时间递增
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// ClearAllMessages 清除所有大厅聊天消息（每天0:00执行）
func (r *ChatRepository) ClearAllMessages() error {
	return r.db.Exec("DELETE FROM global_chats").Error
}

// DeleteOldMessages 删除24小时前的大厅聊天消息
func (r *ChatRepository) DeleteOldMessages() (int64, error) {
	// 计算24小时前的时间
	cutoffTime := time.Now().Add(-24 * time.Hour)
	result := r.db.Where("created_at < ?", cutoffTime).Delete(&database.GlobalChat{})
	return result.RowsAffected, result.Error
}
