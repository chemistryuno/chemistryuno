package repository

import (
	"gorm.io/gorm"
)

var (
	UserRepo         *UserRepository
	SessionRepo      *SessionRepository
	WebAuthnRepo     *WebAuthnRepository
	FeedbackRepo     *FeedbackRepository
	DeckRepo         *DeckRepository
	GameRepo         *GameRepository
	AnnouncementRepo *AnnouncementRepository
	ReactionRepo     *ReactionRepository
	SubstanceRepo    *SubstanceRepository
	BountyRepo       *BountyRepository
	FriendshipRepo   *FriendshipRepository
	ChatRepo         *ChatRepository
	PluginRepo       *PluginRepository
)

func InitRepositories() {
	UserRepo = NewUserRepository()
	SessionRepo = NewSessionRepository()
	WebAuthnRepo = NewWebAuthnRepository()
	FeedbackRepo = NewFeedbackRepository()
	DeckRepo = NewDeckRepository()
	GameRepo = NewGameRepository()
	AnnouncementRepo = NewAnnouncementRepository()
	ReactionRepo = NewReactionRepository()
	SubstanceRepo = NewSubstanceRepository()
	BountyRepo = NewBountyRepository()
	FriendshipRepo = NewFriendshipRepository()
	ChatRepo = NewChatRepository()
	PluginRepo = NewPluginRepository()
}

// randomOrder 根据数据库类型返回随机排序的 SQL 片段
func randomOrder(db *gorm.DB) string {
	if db.Dialector.Name() == "mysql" {
		return "RAND()"
	}
	return "RANDOM()"
}
