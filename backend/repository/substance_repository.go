package repository

import (
	"chemistryuno/backend/database"
	"time"

	"gorm.io/gorm"
)

type SubstanceRepository struct {
	db *gorm.DB
}

func NewSubstanceRepository() *SubstanceRepository {
	return &SubstanceRepository{db: database.DB}
}

// FindApproved 查找已批准的物质
func (r *SubstanceRepository) FindApproved() ([]database.Substance, error) {
	var substances []database.Substance
	err := r.db.Where("status = ?", "approved").Find(&substances).Error
	return substances, err
}

// FindRandomApproved 随机获取一个已批准的物质
func (r *SubstanceRepository) FindRandomApproved() (*database.Substance, error) {
	var substance database.Substance
	// 使用 SQLite 的 RANDOM() 函数
	err := r.db.Where("status = ?", "approved").Order("RANDOM()").First(&substance).Error
	if err != nil {
		return nil, err
	}
	return &substance, nil
}

// FindByID 根据ID查找物质
func (r *SubstanceRepository) FindByID(id uint) (*database.Substance, error) {
	var substance database.Substance
	err := r.db.First(&substance, id).Error
	if err != nil {
		return nil, err
	}
	return &substance, nil
}

// GetStatus 获取物质状态
func (r *SubstanceRepository) GetStatus(id uint) (string, error) {
	var status string
	err := r.db.Model(&database.Substance{}).
		Select("status").
		Where("id = ?", id).
		Scan(&status).Error
	return status, err
}

// UpdateStatus 更新物质状态
func (r *SubstanceRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&database.Substance{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// Create 创建物质
func (r *SubstanceRepository) Create(substance *database.Substance) error {
	return r.db.Create(substance).Error
}

// Update 更新物质
func (r *SubstanceRepository) Update(substance *database.Substance) error {
	return r.db.Save(substance).Error
}

// Delete 删除物质
func (r *SubstanceRepository) Delete(id uint) error {
	return r.db.Delete(&database.Substance{}, id).Error
}

// FindAll 查找所有物质
func (r *SubstanceRepository) FindAll() ([]database.Substance, error) {
	var substances []database.Substance
	err := r.db.Order("created_at DESC").Find(&substances).Error
	return substances, err
}

// FindPending 查找待审核的物质
func (r *SubstanceRepository) FindPending() ([]database.Substance, error) {
	var substances []database.Substance
	err := r.db.Where("status = ?", "pending").
		Order("created_at DESC").
		Find(&substances).Error
	return substances, err
}

// SubstanceWithCreator 物质及创建者信息
type SubstanceWithCreator struct {
	ID           uint      `json:"id"`
	Formula      string    `json:"formula"`
	Name         string    `json:"name"`
	Elements     string    `json:"elements"`
	Status       string    `json:"status"`
	CreatedByUID uint      `json:"created_by_uid"`
	CreatorName  string    `json:"creator_name"`
	CreatedAt    time.Time `json:"created_at"`
}

// FindAllWithCreator 查找所有物质及创建者信息
func (r *SubstanceRepository) FindAllWithCreator() ([]SubstanceWithCreator, error) {
	var results []SubstanceWithCreator

	err := r.db.Table("substances").
		Select("substances.id, substances.formula, substances.name, substances.elements, substances.status, substances.created_by_uid, users.username as creator_name, substances.created_at").
		Joins("LEFT JOIN users ON substances.created_by_uid = users.uid").
		Order("substances.created_at DESC").
		Scan(&results).Error

	return results, err
}

// UpdateWithElements 更新物质（包括元素信息）
func (r *SubstanceRepository) UpdateWithElements(id uint, formula, name, elements, status string) error {
	return r.db.Model(&database.Substance{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"name":     name,
			"formula":  formula,
			"elements": elements,
			"status":   status,
		}).Error
}

// UpdateFormula 更新化学式和名称
func (r *SubstanceRepository) UpdateFormula(id uint, formula, name, elements string) error {
	return r.db.Model(&database.Substance{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"name":     name,
			"formula":  formula,
			"elements": elements,
		}).Error
}

// FindByGroupID 根据组ID查找物质
func (r *SubstanceRepository) FindByGroupID(groupID uint) ([]database.Substance, error) {
	var substances []database.Substance
	err := r.db.Where("group_id = ?", groupID).Find(&substances).Error
	return substances, err
}

// GetStatusByGroupID 获取组的状态
func (r *SubstanceRepository) GetStatusByGroupID(groupID uint) (string, error) {
	var status string
	err := r.db.Model(&database.Substance{}).
		Select("status").
		Where("group_id = ?", groupID).
		Limit(1).
		Scan(&status).Error
	return status, err
}

// UpdateStatusByGroupID 更新组状态
func (r *SubstanceRepository) UpdateStatusByGroupID(groupID uint, status string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status": status,
	}
	if status == "approved" {
		updates["approved_at"] = &now
	}
	return r.db.Model(&database.Substance{}).
		Where("group_id = ?", groupID).
		Updates(updates).Error
}

// DeleteByGroupID 删除组内所有物质
func (r *SubstanceRepository) DeleteByGroupID(groupID uint) error {
	return r.db.Where("group_id = ?", groupID).Delete(&database.Substance{}).Error
}

// GetGroupIDAndCreatorByID 根据ID获取group_id和创建者
func (r *SubstanceRepository) GetGroupIDAndCreatorByID(id uint) (*uint, uint, error) {
	var result struct {
		GroupID      *uint
		CreatedByUID uint
	}
	err := r.db.Model(&database.Substance{}).
		Select("group_id, created_by_uid").
		Where("id = ?", id).
		Scan(&result).Error
	return result.GroupID, result.CreatedByUID, err
}

// SubstanceWithCreatorExtended 扩展的物质及创建者信息
type SubstanceWithCreatorExtended struct {
	ID               uint      `json:"id"`
	Formula          string    `json:"formula"`
	Name             string    `json:"name"`
	Elements         string    `json:"elements"`
	Status           string    `json:"status"`
	GroupID          *uint     `json:"group_id"`
	NeedsImprovement bool      `json:"needs_improvement"`
	CreatedByUID     uint      `json:"created_by_uid"`
	CreatorName      string    `json:"creator_name"`
	CreatedAt        time.Time `json:"created_at"`
}

// FindAllGroupedWithCreator 获取所有物质（按组分组，带创建者信息）
func (r *SubstanceRepository) FindAllGroupedWithCreator() ([]SubstanceWithCreatorExtended, error) {
	var results []SubstanceWithCreatorExtended

	// 子查询找到每个group_id的第一条记录
	subQuery := r.db.Table("substances").Select("MIN(id)").Group("group_id")

	err := r.db.Table("substances").
		Select("substances.id, substances.formula, substances.name, substances.elements, substances.status, substances.group_id, substances.needs_improvement, substances.created_by_uid, users.username as creator_name, substances.created_at").
		Joins("LEFT JOIN users ON substances.created_by_uid = users.uid").
		Where("substances.id IN (?)", subQuery).
		Order("substances.needs_improvement DESC, substances.created_at DESC").
		Scan(&results).Error

	return results, err
}

// FindApprovedGrouped 获取已批准的物质（按组分组，带创建者信息）
func (r *SubstanceRepository) FindApprovedGrouped() ([]SubstanceWithCreatorExtended, error) {
	var results []SubstanceWithCreatorExtended

	// 子查询找到每个已批准组的第一条记录
	subQuery := r.db.Table("substances").
		Select("MIN(id)").
		Where("status = ?", "approved").
		Group("group_id")

	err := r.db.Table("substances").
		Select("substances.id, substances.formula, substances.name, substances.elements, substances.status, substances.group_id, substances.needs_improvement, substances.created_by_uid, users.username as creator_name, substances.created_at").
		Joins("LEFT JOIN users ON substances.created_by_uid = users.uid").
		Where("substances.status = ? AND substances.id IN (?)", "approved", subQuery).
		Order("substances.needs_improvement DESC, substances.created_at DESC").
		Scan(&results).Error

	return results, err
}

// FindMySubstances 获取用户提交的物质
func (r *SubstanceRepository) FindMySubstances(uid uint) ([]SubstanceWithCreatorExtended, error) {
	var results []SubstanceWithCreatorExtended

	// 子查询找到该用户每个组的第一条记录
	subQuery := r.db.Table("substances").
		Select("MIN(id)").
		Where("created_by_uid = ?", uid).
		Group("group_id")

	err := r.db.Table("substances").
		Select("substances.id, substances.formula, substances.name, substances.elements, substances.status, substances.group_id, substances.needs_improvement, substances.created_by_uid, users.username as creator_name, substances.created_at").
		Joins("LEFT JOIN users ON substances.created_by_uid = users.uid").
		Where("substances.created_by_uid = ? AND substances.id IN (?)", uid, subQuery).
		Order("substances.created_at DESC").
		Scan(&results).Error

	return results, err
}

// MarkNeedsImprovement 标记物质需要完善
func (r *SubstanceRepository) MarkNeedsImprovement(groupID uint, needsImprovement bool) error {
	return r.db.Model(&database.Substance{}).
		Where("group_id = ?", groupID).
		Update("needs_improvement", needsImprovement).Error
}

// FindDuplicatesByNameFormula 查找名称和化学式相同的物质（用于标记待完善）
func (r *SubstanceRepository) FindDuplicatesByNameFormula() ([]struct {
	Name    string
	Formula string
	Count   int64
}, error) {
	var results []struct {
		Name    string
		Formula string
		Count   int64
	}

	err := r.db.Table("substances").
		Select("name, formula, COUNT(*) as count").
		Where("status = ?", "approved").
		Group("name, formula").
		Having("COUNT(*) > 1").
		Scan(&results).Error

	return results, err
}
