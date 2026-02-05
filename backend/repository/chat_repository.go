package repository

import (
	"chemistryuno/database"

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
