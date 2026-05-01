package repository

import (
	"chemistryuno/backend/cache"
	"chemistryuno/backend/database"
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ReactionRepository struct {
	db *gorm.DB
}

func NewReactionRepository() *ReactionRepository {
	return &ReactionRepository{db: database.DB}
}

// FindApprovedReactions 查找已批准的反应
func (r *ReactionRepository) FindApprovedReactions() ([]database.Reaction, error) {
	var reactions []database.Reaction
	err := r.db.Where("status = ?", "approved").Find(&reactions).Error
	return reactions, err
}

// FindDistinctSubstances 从已批准反应中返回所有不重复的物质列表
func (r *ReactionRepository) FindDistinctSubstances() ([]string, error) {
	var substances []string
	err := r.db.Raw(`
		SELECT r1 FROM reactions WHERE status = 'approved'
		UNION
		SELECT r2 FROM reactions WHERE status = 'approved'
	`).Scan(&substances).Error
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(substances))
	for _, s := range substances {
		if s != "" {
			result = append(result, s)
		}
	}
	return result, nil
}

// CheckReactionExists 检查反应是否存在 (带 Redis 缓存)
func (r *ReactionRepository) CheckReactionExists(r1, r2 string) (bool, error) {
	// 1. 尝试从 Redis 缓存获取
	ctx := context.Background()
	if val, err := cache.GetReactionCache(ctx, r1, r2); err == nil && val != "" {
		return val == "1", nil
	}

	// 2. 缓存未命中，查询数据库
	if r1 > r2 {
		r1, r2 = r2, r1
	}
	var count int64
	err := r.db.Model(&database.Reaction{}).
		Where("r1 = ? AND r2 = ? AND status = ?", r1, r2, "approved").
		Count(&count).Error
	if err != nil {
		return false, err
	}

	exists := count > 0

	// 3. 将结果写回 Redis 缓存
	_ = cache.SetReactionCache(ctx, r1, r2, exists)

	return exists, nil
}

// GetReaction 获取反应详情
func (r *ReactionRepository) GetReaction(r1, r2 string) (*database.Reaction, error) {
	if r1 > r2 {
		r1, r2 = r2, r1
	}
	var reaction database.Reaction
	err := r.db.Where("r1 = ? AND r2 = ? AND status = ?", r1, r2, "approved").
		First(&reaction).Error
	if err != nil {
		return nil, err
	}
	return &reaction, nil
}

// FindReactionsBySubstance 查找包含指定物质的反应
func (r *ReactionRepository) FindReactionsBySubstance(substance string) ([]database.Reaction, error) {
	var reactions []database.Reaction
	err := r.db.Where("(r1 = ? OR r2 = ?) AND status = ?",
		substance, substance, "approved").Find(&reactions).Error
	return reactions, err
}

// FindByGroupID 根据组ID查找反应
func (r *ReactionRepository) FindByGroupID(groupID uint) ([]database.Reaction, error) {
	var reactions []database.Reaction
	err := r.db.Where("group_id = ?", groupID).Find(&reactions).Error
	return reactions, err
}

// FindByID 根据ID查找反应
func (r *ReactionRepository) FindByID(id uint) (*database.Reaction, error) {
	var reaction database.Reaction
	err := r.db.First(&reaction, id).Error
	if err != nil {
		return nil, err
	}
	return &reaction, nil
}

// GetStatusByGroupID 获取组的状态
func (r *ReactionRepository) GetStatusByGroupID(groupID uint) (string, error) {
	var status string
	err := r.db.Model(&database.Reaction{}).
		Select("status").
		Where("group_id = ?", groupID).
		Limit(1).
		Scan(&status).Error
	return status, err
}

// UpdateStatusByGroupID 更新组状态
func (r *ReactionRepository) UpdateStatusByGroupID(groupID uint, status string) error {
	return r.db.Model(&database.Reaction{}).
		Where("group_id = ?", groupID).
		Update("status", status).Error
}

// DeleteByGroupID 删除组内所有反应
func (r *ReactionRepository) DeleteByGroupID(groupID uint) error {
	return r.db.Where("group_id = ?", groupID).Delete(&database.Reaction{}).Error
}

// Create 创建反应
func (r *ReactionRepository) Create(reaction *database.Reaction) error {
	// Canonical ordering: 确保 r1 <= r2 (字母序)
	if reaction.R1 > reaction.R2 {
		reaction.R1, reaction.R2 = reaction.R2, reaction.R1
	}
	return r.db.Create(reaction).Error
}

// FindAll 查找所有反应
func (r *ReactionRepository) FindAll() ([]database.Reaction, error) {
	var reactions []database.Reaction
	err := r.db.Order("created_at DESC").Find(&reactions).Error
	return reactions, err
}

// FindPendingGrouped 查找待审核的反应（分组）
func (r *ReactionRepository) FindPendingGrouped() ([]database.Reaction, error) {
	var reactions []database.Reaction
	err := r.db.Where("status = ?", "pending").
		Order("group_id, id").
		Find(&reactions).Error
	return reactions, err
}

// CheckDuplicateByR1R2 检查r1和r2组合是否已存在（排除指定groupID）
func (r *ReactionRepository) CheckDuplicateByR1R2(r1, r2, excludeGroupID string) (bool, string, error) {
	// 已弃用：该功能依赖旧表结构
	return false, "", nil
}

// GetGroupIDAndCreatorByID 根据ID获取group_id和创建者
func (r *ReactionRepository) GetGroupIDAndCreatorByID(id uint) (*uint, uint, error) {
	var result struct {
		GroupID      *uint
		CreatedByUID uint
	}
	err := r.db.Model(&database.Reaction{}).
		Select("group_id, created_by_uid").
		Where("id = ?", id).
		Scan(&result).Error
	return result.GroupID, result.CreatedByUID, err
}

// ReactionWithCreator 带创建者信息的反应
type ReactionWithCreator struct {
	ID                 uint      `json:"id"`
	Display            string    `json:"display"`
	R1                 string    `json:"r1"` // 反应物1
	R2                 string    `json:"r2"` // 反应物2
	Status             string    `json:"status"`
	GroupID            *uint     `json:"group_id"`
	HasInvalidElements bool      `json:"has_invalid_elements"`
	CreatedByUID       uint      `json:"created_by_uid"`
	CreatorName        string    `json:"creator_name"`
	CreatedAt          time.Time `json:"created_at"`
}

type ReactionListFilter struct {
	ViewerUID       *uint
	IncludeApproved bool
	Search          string
	Status          string
	HasInvalid      *bool
	Page            int
	PageSize        int
}

func (r *ReactionRepository) applyReactionListFilter(query *gorm.DB, filter ReactionListFilter) *gorm.DB {
	if filter.IncludeApproved && filter.ViewerUID != nil {
		query = query.Where("(reactions.status = ? OR reactions.created_by_uid = ?)", "approved", *filter.ViewerUID)
	}

	if filter.ViewerUID != nil && !filter.IncludeApproved {
		query = query.Where("reactions.created_by_uid = ?", *filter.ViewerUID)
	}

	if filter.Status != "" && filter.Status != "all" {
		statuses := make([]string, 0)
		for _, status := range strings.Split(filter.Status, ",") {
			status = strings.TrimSpace(status)
			if status != "" {
				statuses = append(statuses, status)
			}
		}
		if len(statuses) == 1 {
			query = query.Where("reactions.status = ?", statuses[0])
		} else if len(statuses) > 1 {
			query = query.Where("reactions.status IN ?", statuses)
		}
	}

	if filter.HasInvalid != nil {
		query = query.Where("reactions.has_invalid_elements = ?", *filter.HasInvalid)
	}

	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where(
			"(LOWER(reactions.display) LIKE LOWER(?) OR LOWER(reactions.r1) LIKE LOWER(?) OR LOWER(reactions.r2) LIKE LOWER(?))",
			like, like, like,
		)
	}

	return query
}

func (r *ReactionRepository) FindGroupedWithCreatorPage(filter ReactionListFilter) ([]ReactionWithCreator, int64, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}

	baseQuery := r.applyReactionListFilter(r.db.Table("reactions"), filter)
	groupExpr := "COALESCE(reactions.group_id, reactions.id)"

	groupedSubQuery := baseQuery.
		Select("MIN(reactions.id) AS id, MAX(reactions.created_at) AS latest_created_at").
		Group(groupExpr)

	var total int64
	if err := r.db.Table("(?) AS grouped_reactions", groupedSubQuery).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var results []ReactionWithCreator
	err := r.db.Table("reactions").
		Select("reactions.id, reactions.display, reactions.r1, reactions.r2, reactions.status, reactions.group_id, reactions.has_invalid_elements, reactions.created_by_uid, users.username as creator_name, reactions.created_at").
		Joins("JOIN (?) AS grouped_reactions ON reactions.id = grouped_reactions.id", groupedSubQuery).
		Joins("LEFT JOIN users ON reactions.created_by_uid = users.uid").
		Order("grouped_reactions.latest_created_at DESC, reactions.id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&results).Error
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// FindAllGroupedWithCreator 获取所有反应（按组分组，带创建者信息）
func (r *ReactionRepository) FindAllGroupedWithCreator() ([]ReactionWithCreator, error) {
	var results []ReactionWithCreator

	// 子查询找到每个group_id的第一条记录
	subQuery := r.db.Table("reactions").Select("MIN(id)").Group("COALESCE(group_id, id)")

	err := r.db.Table("reactions").
		Select("reactions.id, reactions.display, reactions.r1, reactions.r2, reactions.status, reactions.group_id, reactions.has_invalid_elements, reactions.created_by_uid, users.username as creator_name, reactions.created_at").
		Joins("LEFT JOIN users ON reactions.created_by_uid = users.uid").
		Where("reactions.id IN (?)", subQuery).
		Order("reactions.created_at DESC").
		Scan(&results).Error

	return results, err
}

// FindApprovedGrouped 获取已批准的反应（按组分组，带创建者信息）
func (r *ReactionRepository) FindApprovedGrouped() ([]ReactionWithCreator, error) {
	var results []ReactionWithCreator

	// 子查询找到每个已批准组的第一条记录
	subQuery := r.db.Table("reactions").
		Select("MIN(id)").
		Where("status = ?", "approved").
		Group("COALESCE(group_id, id)")

	err := r.db.Table("reactions").
		Select("reactions.id, reactions.display, reactions.r1, reactions.r2, reactions.status, reactions.group_id, reactions.has_invalid_elements, reactions.created_by_uid, users.username as creator_name, reactions.created_at").
		Joins("LEFT JOIN users ON reactions.created_by_uid = users.uid").
		Where("reactions.status = ? AND reactions.id IN (?)", "approved", subQuery).
		Order("reactions.created_at DESC").
		Scan(&results).Error

	return results, err
}

// FindMyReactions 获取用户提交的反应
func (r *ReactionRepository) FindMyReactions(uid uint) ([]ReactionWithCreator, error) {
	var results []ReactionWithCreator

	// 子查询找到该用户每个组的第一条记录
	subQuery := r.db.Table("reactions").
		Select("MIN(id)").
		Where("created_by_uid = ?", uid).
		Group("COALESCE(group_id, id)")

	err := r.db.Table("reactions").
		Select("reactions.id, reactions.display, reactions.r1, reactions.r2, reactions.status, reactions.group_id, reactions.has_invalid_elements, reactions.created_by_uid, users.username as creator_name, reactions.created_at").
		Joins("LEFT JOIN users ON reactions.created_by_uid = users.uid").
		Where("reactions.created_by_uid = ? AND reactions.id IN (?)", uid, subQuery).
		Order("reactions.created_at DESC").
		Scan(&results).Error

	return results, err
}

// CreateBatch 批量创建反应（用于事务）
func (r *ReactionRepository) CreateBatch(reactions []database.Reaction) error {
	for i := range reactions {
		if reactions[i].R1 > reactions[i].R2 {
			reactions[i].R1, reactions[i].R2 = reactions[i].R2, reactions[i].R1
		}
	}
	return r.db.Create(&reactions).Error
}

// BatchUpdateStatusByGroupIDs 批量更新反应组状态
func (r *ReactionRepository) BatchUpdateStatusByGroupIDs(groupIDs []uint, status string) (int64, error) {
	now := time.Now()
	updates := map[string]interface{}{
		"status": status,
	}
	if status == "approved" {
		updates["approved_at"] = &now
	}

	result := r.db.Model(&database.Reaction{}).
		Where("group_id IN ?", groupIDs).
		Updates(updates)

	return result.RowsAffected, result.Error
}

// BatchDeleteByGroupIDs 批量删除反应组
func (r *ReactionRepository) BatchDeleteByGroupIDs(groupIDs []uint) (int64, error) {
	result := r.db.Where("group_id IN ?", groupIDs).Delete(&database.Reaction{})
	return result.RowsAffected, result.Error
}
