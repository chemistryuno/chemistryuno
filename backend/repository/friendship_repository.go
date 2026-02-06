package repository

import (
	"chemistryuno/database"
	"time"

	"gorm.io/gorm"
)

type FriendshipRepository struct {
	db *gorm.DB
}

func NewFriendshipRepository() *FriendshipRepository {
	return &FriendshipRepository{
		db: database.DB,
	}
}

func (r *FriendshipRepository) CreateRequest(userUID, friendUID uint, message string) error {
	// 如果已经存在（无论是 pending 还是 accepted），不再创建
	var existing database.Friendship
	err := r.db.Where("(user_uid = ? AND friend_uid = ?) OR (user_uid = ? AND friend_uid = ?)", userUID, friendUID, friendUID, userUID).First(&existing).Error
	if err == nil {
		// 如果是已经接受了，或者过期的 pending，可能需要不同处理，但这里先按用户逻辑：已有记录就不再创建
		// 如果是 pending 状态，更新 hello_message 和 CreatedAt 也是一种思路，但先保持简单
		return nil
	}

	f := database.Friendship{
		UserUID:      userUID,
		FriendUID:    friendUID,
		Status:       "pending",
		HelloMessage: message,
	}
	return r.db.Create(&f).Error
}

func (r *FriendshipRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&database.Friendship{}).Where("id = ?", id).Update("status", status).Error
}

func (r *FriendshipRepository) GetPendingRequests(uid uint) ([]database.Friendship, error) {
	var requests []database.Friendship
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	err := r.db.Preload("Friend").Preload("User").
		Where("friend_uid = ? AND status = ? AND created_at > ?", uid, "pending", sevenDaysAgo).
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

func (r *FriendshipRepository) GetFriends(uid uint) ([]database.User, error) {
	var friendships []database.Friendship
	// 查找所有 status 为 accepted 且涉及到 uid 的记录
	err := r.db.Where("status = ? AND (user_uid = ? OR friend_uid = ?)", "accepted", uid, uid).Find(&friendships).Error
	if err != nil {
		return nil, err
	}

	friendUIDs := make([]uint, 0)
	for _, f := range friendships {
		if f.UserUID == uid {
			friendUIDs = append(friendUIDs, f.FriendUID)
		} else {
			friendUIDs = append(friendUIDs, f.UserUID)
		}
	}

	if len(friendUIDs) == 0 {
		return []database.User{}, nil
	}

	var friends []database.User
	err = r.db.Where("uid IN ?", friendUIDs).Find(&friends).Error
	return friends, err
}

func (r *FriendshipRepository) IsFriend(uid1, uid2 uint) (bool, error) {
	var count int64
	err := r.db.Model(&database.Friendship{}).
		Where("status = ? AND ((user_uid = ? AND friend_uid = ?) OR (user_uid = ? AND friend_uid = ?))", "accepted", uid1, uid2, uid2, uid1).
		Count(&count).Error
	return count > 0, err
}

func (r *FriendshipRepository) DeleteFriendship(uid1, uid2 uint) error {
	return r.db.Where("((user_uid = ? AND friend_uid = ?) OR (user_uid = ? AND friend_uid = ?))", uid1, uid2, uid2, uid1).Delete(&database.Friendship{}).Error
}

func (r *FriendshipRepository) GetFriendshipByID(id uint) (*database.Friendship, error) {
	var f database.Friendship
	err := r.db.First(&f, id).Error
	return &f, err
}

func (r *FriendshipRepository) CleanupExpiredRequests() error {
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	// 删除状态为 pending 且创建时间超过 7 天的请求
	return r.db.Where("status = ? AND created_at < ?", "pending", sevenDaysAgo).Delete(&database.Friendship{}).Error
}
