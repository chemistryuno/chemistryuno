package repository

import (
	"chemistryuno/database"

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

func (r *FriendshipRepository) CreateRequest(userID, friendID uint) error {
	// 如果已经存在（无论是 pending 还是 accepted），不再创建
	var existing database.Friendship
	err := r.db.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)", userID, friendID, friendID, userID).First(&existing).Error
	if err == nil {
		return nil // 已经存在记录
	}

	f := database.Friendship{
		UserID:   userID,
		FriendID: friendID,
		Status:   "pending",
	}
	return r.db.Create(&f).Error
}

func (r *FriendshipRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&database.Friendship{}).Where("id = ?", id).Update("status", status).Error
}

func (r *FriendshipRepository) GetPendingRequests(uid uint) ([]database.Friendship, error) {
	var requests []database.Friendship
	err := r.db.Preload("Friend").Preload("User").Where("friend_id = ? AND status = ?", uid, "pending").Find(&requests).Error
	return requests, err
}

func (r *FriendshipRepository) GetFriends(uid uint) ([]database.User, error) {
	var friendships []database.Friendship
	// 查找所有 status 为 accepted 且涉及到 uid 的记录
	err := r.db.Where("status = ? AND (user_id = ? OR friend_id = ?)", "accepted", uid, uid).Find(&friendships).Error
	if err != nil {
		return nil, err
	}

	friendIDs := make([]uint, 0)
	for _, f := range friendships {
		if f.UserID == uid {
			friendIDs = append(friendIDs, f.FriendID)
		} else {
			friendIDs = append(friendIDs, f.UserID)
		}
	}

	if len(friendIDs) == 0 {
		return []database.User{}, nil
	}

	var friends []database.User
	err = r.db.Where("uid IN ?", friendIDs).Find(&friends).Error
	return friends, err
}

func (r *FriendshipRepository) IsFriend(uid1, uid2 uint) (bool, error) {
	var count int64
	err := r.db.Model(&database.Friendship{}).
		Where("status = ? AND ((user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?))", "accepted", uid1, uid2, uid2, uid1).
		Count(&count).Error
	return count > 0, err
}

func (r *FriendshipRepository) DeleteFriendship(uid1, uid2 uint) error {
	return r.db.Where("((user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?))", uid1, uid2, uid2, uid1).Delete(&database.Friendship{}).Error
}

func (r *FriendshipRepository) GetFriendshipByID(id uint) (*database.Friendship, error) {
	var f database.Friendship
	err := r.db.First(&f, id).Error
	return &f, err
}
