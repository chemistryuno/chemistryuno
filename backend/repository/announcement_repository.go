package repository

import (
	"chemistryuno/backend/database"
	"time"

	"gorm.io/gorm"
)

type AnnouncementRepository struct {
	db *gorm.DB
}

func NewAnnouncementRepository() *AnnouncementRepository {
	return &AnnouncementRepository{db: database.DB}
}

// Announcement GORM模型
type Announcement struct {
	ID              uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Title           string     `gorm:"size:255" json:"title"`
	Content         string     `gorm:"type:text;not null" json:"content"`
	Type            string     `gorm:"size:50;default:info" json:"type"`
	Active          bool       `gorm:"default:true" json:"active"`
	IsTicker        bool       `gorm:"default:false" json:"is_ticker"`
	IsPersistent    bool       `gorm:"default:false" json:"is_persistent"`
	OnJoin          bool       `gorm:"default:false" json:"on_join"`
	CronInterval    int        `gorm:"default:0" json:"cron_interval"`
	CloseDelay      int        `gorm:"default:0" json:"close_delay"`
	LastBroadcastAt *time.Time `json:"last_broadcast_at"`
	ExpiresAt       *time.Time `json:"expires_at"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (Announcement) TableName() string {
	return "announcements"
}

// FindActive 查找活跃的公告
func (r *AnnouncementRepository) FindActive() ([]Announcement, error) {
	var announcements []Announcement
	now := time.Now()
	err := r.db.Where("active = ? AND (expires_at IS NULL OR expires_at > ?)", true, now).
		Find(&announcements).Error
	return announcements, err
}

// FindAll 查找所有公告
func (r *AnnouncementRepository) FindAll() ([]Announcement, error) {
	var announcements []Announcement
	err := r.db.Order("created_at DESC").Find(&announcements).Error
	return announcements, err
}

// Create 创建公告
func (r *AnnouncementRepository) Create(announcement *Announcement) error {
	return r.db.Create(announcement).Error
}

// Update 更新公告完整信息
func (r *AnnouncementRepository) Update(id uint, announcement *Announcement) error {
	updates := map[string]interface{}{
		"title":         announcement.Title,
		"content":       announcement.Content,
		"type":          announcement.Type,
		"is_ticker":     announcement.IsTicker,
		"is_persistent": announcement.IsPersistent,
		"on_join":       announcement.OnJoin,
		"cron_interval": announcement.CronInterval,
		"close_delay":   announcement.CloseDelay,
		"expires_at":    announcement.ExpiresAt,
	}

	return r.db.Model(&Announcement{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// UpdateActive 更新公告激活状态
func (r *AnnouncementRepository) UpdateActive(id uint, active bool) error {
	return r.db.Model(&Announcement{}).
		Where("id = ?", id).
		Update("active", active).Error
}

// Delete 删除公告
func (r *AnnouncementRepository) Delete(id uint) error {
	return r.db.Delete(&Announcement{}, id).Error
}

// DeactivateExpired 停用过期公告
func (r *AnnouncementRepository) DeactivateExpired() (int64, error) {
	result := r.db.Model(&Announcement{}).
		Where("active = ? AND expires_at IS NOT NULL AND expires_at < ?", true, time.Now()).
		Update("active", false)
	return result.RowsAffected, result.Error
}

// FindCronAnnouncements 查找定时广播公告
func (r *AnnouncementRepository) FindCronAnnouncements() ([]Announcement, error) {
	var announcements []Announcement
	err := r.db.Where("active = ? AND cron_interval > 0", true).
		Find(&announcements).Error
	return announcements, err
}

// FindOnJoinAnnouncements 查找登录时显示的公告
func (r *AnnouncementRepository) FindOnJoinAnnouncements() ([]Announcement, error) {
	var announcements []Announcement
	now := time.Now()
	err := r.db.Where("active = ? AND on_join = ? AND (expires_at IS NULL OR expires_at > ?)",
		true, true, now).
		Find(&announcements).Error
	return announcements, err
}

// UpdateLastBroadcast 更新最后广播时间
func (r *AnnouncementRepository) UpdateLastBroadcast(id uint, time time.Time) error {
	return r.db.Model(&Announcement{}).
		Where("id = ?", id).
		Update("last_broadcast_at", time).Error
}

// FindRandomHints 获取随机3条类型为 'hint' 的公告
func (r *AnnouncementRepository) FindRandomHints() ([]Announcement, error) {
	var announcements []Announcement
	// 使用数据库特定的随机排序，这里假设使用的是 SQLite/MySQL/PostgreSQL 支持 RANDOM() 或 RAND()
	// GORM 可以通过 Order("RANDOM()") 来实现，但为了兼容性，也可以查出所有 hint 后再随机
	// 这里使用 Order("RAND() / RANDOM()") 的通用写法是不太稳妥，但如果是 SQLite 则是 RANDOM()

	// 先根据活跃状态和类型过滤
	query := r.db.Where("active = ? AND type = ?", true, "hint")

	// 在 SQLite 中是 RANDOM()
	err := query.Order(randomOrder(r.db)).Limit(3).Find(&announcements).Error
	return announcements, err
}
