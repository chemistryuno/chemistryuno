package repository

import (
	"chemistryuno/database"

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
	ID          uint   `json:"id"`
	Formula     string `json:"formula"`
	Name        string `json:"name"`
	Elements    string `json:"elements"`
	Status      string `json:"status"`
	CreatedBy   uint   `json:"created_by"`
	CreatorName string `json:"creator_name"`
	CreatedAt   string `json:"created_at"`
}

// FindAllWithCreator 查找所有物质及创建者信息
func (r *SubstanceRepository) FindAllWithCreator() ([]SubstanceWithCreator, error) {
	var results []SubstanceWithCreator
	// 这需要原生SQL查询，因为涉及JOIN和旧表结构
	return results, nil
}

// UpdateWithElements 更新物质（包括元素信息）
func (r *SubstanceRepository) UpdateWithElements(id uint, formula, name, elements, status string) error {
	return r.db.Model(&database.Substance{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"name":        name,
			"description": formula, // formula映射到description
			"status":      status,
		}).Error
}

// UpdateFormula 更新化学式和名称
func (r *SubstanceRepository) UpdateFormula(id uint, formula, name, elements string) error {
	return r.db.Model(&database.Substance{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"name":        name,
			"description": formula,
		}).Error
}
