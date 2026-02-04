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
func (r *ReactionRepository) GetGroupIDAndCreatorByID(id uint) (string, uint, error) {
	var result struct {
		GroupID   string
		CreatedBy uint
	}
	err := r.db.Model(&database.Reaction{}).
		Select("group_id, created_by").
		Where("id = ?", id).
		Scan(&result).Error
	return result.GroupID, result.CreatedBy, err
}

// FindPendingByStatus 根据状态查找待审核反应（支持多种状态）
type ReactionWithCreator struct {
	ID          uint   `json:"id"`
	Display     string `json:"display"`
	Status      string `json:"status"`
	GroupID     string `json:"group_id"`
	CreatedBy   uint   `json:"created_by"`
	CreatorName string `json:"creator_name"`
	CreatedAt   string `json:"created_at"`
}

func (r *ReactionRepository) FindAllGroupedWithCreator() ([]ReactionWithCreator, error) {
	var results []ReactionWithCreator
	// 已弃用：该功能依赖旧表结构
	return results, nil
}
