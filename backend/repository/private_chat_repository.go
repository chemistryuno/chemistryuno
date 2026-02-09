package repository

import (
	"chemistryuno/backend/database"

	"gorm.io/gorm"
)

type PrivateChatRepository struct {
	db *gorm.DB
}

func NewPrivateChatRepository() *PrivateChatRepository {
	return &PrivateChatRepository{db: database.DB}
}

// SavePrivateMessage 保存私聊消息
func (r *PrivateChatRepository) SavePrivateMessage(senderUID, receiverUID uint, message string, isGameInvite bool, roomID string) error {
	chat := database.PrivateChat{
		SenderUID:    senderUID,
		ReceiverUID:  receiverUID,
		Message:      message,
		IsGameInvite: isGameInvite,
		RoomID:       roomID,
	}
	return r.db.Create(&chat).Error
}

// GetMessagesBetweenUsers 获取两个用户之间的历史消息（双向）
func (r *PrivateChatRepository) GetMessagesBetweenUsers(uid1, uid2 uint, limit int) ([]database.PrivateChat, error) {
	messages := []database.PrivateChat{}
	err := r.db.
		Preload("Sender").
		Preload("Receiver").
		Where("(sender_uid = ? AND receiver_uid = ?) OR (sender_uid = ? AND receiver_uid = ?)", uid1, uid2, uid2, uid1).
		Order("created_at desc").
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// 反转顺序，让前端收到的顺序是时间递增
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// DeleteGameInvitesByRoom 删除指定房间的游戏邀请消息
func (r *PrivateChatRepository) DeleteGameInvitesByRoom(roomID string) error {
	return r.db.
		Where("is_game_invite = ? AND room_id = ?", true, roomID).
		Delete(&database.PrivateChat{}).Error
}
