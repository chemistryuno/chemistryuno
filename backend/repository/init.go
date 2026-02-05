package repository

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
}
