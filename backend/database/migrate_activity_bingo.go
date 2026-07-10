package database

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// GameVersion 游戏版本（赛季）
type GameVersion struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	StartDate time.Time `gorm:"not null" json:"start_date"`
	EndDate   time.Time `gorm:"not null" json:"end_date"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (GameVersion) TableName() string { return "game_versions" }

// Activity 活动记录
type Activity struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:200;not null" json:"name"`
	Type      string    `gorm:"size:50;not null;index" json:"type"` // double_points, bingo, etc.
	StartTime time.Time `gorm:"not null;index" json:"start_time"`
	EndTime   time.Time `gorm:"not null;index" json:"end_time"`
	VersionID *uint     `gorm:"index" json:"version_id"`
	Params    JSON      `gorm:"type:text" json:"params"` // JSON配置，如 {"daily_limit": 3}
	IsActive  bool      `gorm:"default:true;index" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Activity) TableName() string { return "activities" }

// DailyActivityToken 每日活动使用令牌（双倍积分等）
type DailyActivityToken struct {
	ID         uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	UID        uint   `gorm:"not null;index:idx_token_uid_activity_date" json:"uid"`
	ActivityID uint   `gorm:"not null;index:idx_token_uid_activity_date" json:"activity_id"`
	Date       string `gorm:"size:10;not null;index:idx_token_uid_activity_date" json:"date"` // UTC date "2006-01-02"
	UsedCount  int    `gorm:"default:0" json:"used_count"`
}

func (DailyActivityToken) TableName() string { return "daily_activity_tokens" }

// Team 队伍
type Team struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name       string    `gorm:"size:100;not null" json:"name"`
	InviteCode string    `gorm:"size:20;uniqueIndex;not null" json:"invite_code"`
	LeaderUID  uint      `gorm:"not null;index" json:"leader_uid"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Team) TableName() string { return "teams" }

// TeamMember 队伍成员
type TeamMember struct {
	TeamID   uint      `gorm:"primaryKey;index:idx_team_member" json:"team_id"`
	UID      uint      `gorm:"primaryKey;index:idx_team_member;index:idx_member_uid" json:"uid"`
	JoinedAt time.Time `gorm:"autoCreateTime" json:"joined_at"`
}

func (TeamMember) TableName() string { return "team_members" }

// TeamChatMessage 队伍聊天消息
type TeamChatMessage struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TeamID    uint      `gorm:"not null;index" json:"team_id"`
	UID       uint      `gorm:"not null" json:"uid"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

func (TeamChatMessage) TableName() string { return "team_chat_messages" }

// BingoRoom BINGO 游戏房间
type BingoRoom struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TeamAMembers   JSON      `gorm:"type:text" json:"team_a_members"` // JSON array of UIDs
	TeamBMembers   JSON      `gorm:"type:text" json:"team_b_members"` // JSON array of UIDs
	AIMembers      JSON      `gorm:"type:text" json:"ai_members"`     // JSON array of AI UIDs (subset of TeamA + TeamB)
	AIDifficulty   int       `gorm:"default:50" json:"ai_difficulty"` // AI difficulty (10-90)
	Board          JSON      `gorm:"type:text" json:"board"`
	Status         string    `gorm:"size:30;default:waiting;index" json:"status"` // waiting, playing, finished
	TimeoutMinutes int       `gorm:"default:10" json:"timeout_minutes"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (BingoRoom) TableName() string { return "bingo_rooms" }

// BingoCell BINGO 棋盘格子（持久化快照，实际状态在内存）
type BingoCell struct {
	RoomID      uint   `gorm:"primaryKey;index:idx_bingo_cell" json:"room_id"`
	Row         int    `gorm:"primaryKey" json:"row"`
	Col         int    `gorm:"primaryKey" json:"col"`
	SubstanceID uint   `gorm:"not null" json:"substance_id"`
	OwnerTeamID *uint  `gorm:"index" json:"owner_team_id"` // nil = unoccupied
}

func (BingoCell) TableName() string { return "bingo_cells" }

// MigrateActivityBingoTables 执行活动与BINGO相关表的数据库迁移
func MigrateActivityBingoTables(db *gorm.DB) error {
	log.Println("Running activity/bingo migrations...")

	tables := []interface{}{
		&GameVersion{},
		&Activity{},
		&DailyActivityToken{},
		&Team{},
		&TeamMember{},
		&TeamChatMessage{},
		&BingoRoom{},
		&BingoCell{},
	}

	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			if err := db.Migrator().CreateTable(table); err != nil {
				log.Printf("Failed to create table %T: %v", table, err)
				return err
			}
		}
	}

	if err := db.AutoMigrate(tables...); err != nil {
		log.Printf("Warning: AutoMigrate for activity/bingo tables: %v", err)
	}

	// Drop deprecated NOT NULL columns left over from an earlier BingoRoom schema.
	// team_a_id/team_b_id predate the JSON team_*_members columns; AutoMigrate adds
	// the new columns but never removes the old NOT NULL ones, so every insert fails
	// the constraint until they are gone. DropColumn rebuilds the table on SQLite.
	for _, col := range []string{"team_a_id", "team_b_id"} {
		if db.Migrator().HasColumn(&BingoRoom{}, col) {
			if err := db.Migrator().DropColumn(&BingoRoom{}, col); err != nil {
				log.Printf("Warning: failed to drop deprecated bingo_rooms.%s: %v", col, err)
			} else {
				log.Printf("Dropped deprecated bingo_rooms.%s column", col)
			}
		}
	}

	log.Println("Activity/bingo migrations completed")
	return nil
}
