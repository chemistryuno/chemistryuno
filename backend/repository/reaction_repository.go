package repository

import (
	"chemistryuno/database"

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

// CheckReactionExists 检查反应是否存在
func (r *ReactionRepository) CheckReactionExists(r1, r2 string) (bool, error) {
	var count int64
	err := r.db.Model(&database.Reaction{}).
		Where("((reactants = ? AND products = ?) OR (reactants = ? AND products = ?)) AND status = ?",
			r1, r2, r2, r1, "approved").
		Count(&count).Error
	return count > 0, err
}

// FindReactionsBySubstance 查找包含指定物质的反应
func (r *ReactionRepository) FindReactionsBySubstance(substance string) ([]database.Reaction, error) {
	var reactions []database.Reaction
	err := r.db.Where("(reactants = ? OR products = ?) AND status = ?",
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
	ID           uint   `json:"id"`
	Display      string `json:"display"`
	Status       string `json:"status"`
	GroupID      *uint  `json:"group_id"`
	CreatedByUID uint   `json:"created_by_uid"`
	CreatorName  string `json:"creator_name"`
	CreatedAt    string `json:"created_at"`
}

// FindAllGroupedWithCreator 获取所有反应（按组分组，带创建者信息）
func (r *ReactionRepository) FindAllGroupedWithCreator() ([]ReactionWithCreator, error) {
	var results []ReactionWithCreator

	// 子查询找到每个group_id的第一条记录
	subQuery := r.db.Table("reactions").Select("MIN(id)").Group("group_id")

	err := r.db.Table("reactions").
		Select("reactions.id, reactions.display, reactions.status, reactions.group_id, reactions.created_by_uid, users.username as creator_name, reactions.created_at").
		Joins("LEFT JOIN users ON reactions.created_by_uid = users.uid").
		Where("reactions.id IN (?)", subQuery).
		Order("reactions.created_at DESC").
		Scan(&results).Error

	return results, err
}

// FindApprovedGrouped 获取已批准的反应（按组分组）
func (r *ReactionRepository) FindApprovedGrouped() ([]ReactionWithCreator, error) {
	var results []ReactionWithCreator

	// 子查询找到每个已批准组的第一条记录
	subQuery := r.db.Table("reactions").
		Select("MIN(id)").
		Where("status = ?", "approved").
		Group("group_id")

	err := r.db.Table("reactions").
		Select("reactions.id, reactions.display, reactions.reactants as r1, reactions.products as r2, reactions.created_at").
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
		Group("group_id")

	err := r.db.Table("reactions").
		Select("reactions.id, reactions.display, reactions.status, reactions.created_at").
		Where("reactions.created_by_uid = ? AND reactions.id IN (?)", uid, subQuery).
		Order("reactions.created_at DESC").
		Scan(&results).Error

	return results, err
}

// CreateBatch 批量创建反应（用于事务）
func (r *ReactionRepository) CreateBatch(reactions []database.Reaction) error {
	return r.db.Create(&reactions).Error
}
