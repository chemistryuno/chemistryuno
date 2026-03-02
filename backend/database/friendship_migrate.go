package database

import (
	"log"
	"sort"

	"gorm.io/gorm"
)

func friendshipStatusPriority(status string) int {
	switch status {
	case "accepted":
		return 3
	case "pending":
		return 2
	case "declined":
		return 1
	default:
		return 0
	}
}

func pickFriendshipToKeep(items []Friendship) Friendship {
	copied := append([]Friendship(nil), items...)
	sort.SliceStable(copied, func(i, j int) bool {
		left := copied[i]
		right := copied[j]
		if friendshipStatusPriority(left.Status) != friendshipStatusPriority(right.Status) {
			return friendshipStatusPriority(left.Status) > friendshipStatusPriority(right.Status)
		}
		if !left.UpdatedAt.Equal(right.UpdatedAt) {
			return left.UpdatedAt.After(right.UpdatedAt)
		}
		return left.ID > right.ID
	})
	return copied[0]
}

func mergeFriendshipRemarks(items []Friendship, userUID, friendUID uint, defaultUserRemark, defaultFriendRemark string) (string, string) {
	userRemark := defaultUserRemark
	friendRemark := defaultFriendRemark

	for _, item := range items {
		if userRemark == "" {
			if item.UserUID == userUID && item.UserRemark != "" {
				userRemark = item.UserRemark
			} else if item.FriendUID == userUID && item.FriendRemark != "" {
				userRemark = item.FriendRemark
			}
		}

		if friendRemark == "" {
			if item.UserUID == friendUID && item.UserRemark != "" {
				friendRemark = item.UserRemark
			} else if item.FriendUID == friendUID && item.FriendRemark != "" {
				friendRemark = item.FriendRemark
			}
		}

		if userRemark != "" && friendRemark != "" {
			break
		}
	}

	return userRemark, friendRemark
}

func migrateFriendshipPairKey() error {
	var friendships []Friendship
	if err := DB.Order("id ASC").Find(&friendships).Error; err != nil {
		return err
	}

	grouped := make(map[string][]Friendship)
	for _, item := range friendships {
		key := FriendshipPairKey(item.UserUID, item.FriendUID)
		grouped[key] = append(grouped[key], item)
	}

	if err := DB.Transaction(func(tx *gorm.DB) error {
		for key, items := range grouped {
			keep := pickFriendshipToKeep(items)
			userRemark, friendRemark := mergeFriendshipRemarks(items, keep.UserUID, keep.FriendUID, keep.UserRemark, keep.FriendRemark)

			updates := map[string]interface{}{}
			if keep.PairKey != key {
				updates["pair_key"] = key
			}
			if keep.UserRemark != userRemark {
				updates["user_remark"] = userRemark
			}
			if keep.FriendRemark != friendRemark {
				updates["friend_remark"] = friendRemark
			}

			if len(updates) > 0 {
				if err := tx.Model(&Friendship{}).Where("id = ?", keep.ID).Updates(updates).Error; err != nil {
					return err
				}
			}

			if len(items) <= 1 {
				continue
			}

			duplicateIDs := make([]uint, 0, len(items)-1)
			for _, item := range items {
				if item.ID != keep.ID {
					duplicateIDs = append(duplicateIDs, item.ID)
				}
			}

			if len(duplicateIDs) == 0 {
				continue
			}

			if err := tx.Where("id IN ?", duplicateIDs).Delete(&Friendship{}).Error; err != nil {
				return err
			}

			log.Printf("Deduplicated friendship pair %s, kept id=%d and removed %d duplicate rows", key, keep.ID, len(duplicateIDs))
		}
		return nil
	}); err != nil {
		return err
	}

	if DB.Migrator().HasIndex(&Friendship{}, "idx_friendship_pair_key_unique") {
		return nil
	}

	if err := DB.Exec("CREATE UNIQUE INDEX idx_friendship_pair_key_unique ON friendships(pair_key)").Error; err != nil {
		return err
	}

	log.Println("Friendship pair_key unique index ensured")
	return nil
}
