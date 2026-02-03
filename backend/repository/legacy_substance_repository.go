package repository

import (
	"database/sql"
)

// LegacySubstanceRepository 处理旧表结构的物质Repository
type LegacySubstanceRepository struct {
	db *sql.DB
}

func NewLegacySubstanceRepository(db *sql.DB) *LegacySubstanceRepository {
	return &LegacySubstanceRepository{db: db}
}

// SubstanceInfo 物质信息
type SubstanceInfo struct {
	ID          int
	Formula     string
	Name        string
	Elements    string
	Status      string
	CreatedBy   int
	CreatorName string
	CreatedAt   string
}

// GetAllSubstances 获取所有物质
func (r *LegacySubstanceRepository) GetAllSubstances() ([]SubstanceInfo, error) {
	rows, err := r.db.Query(`
		SELECT s.id, s.formula, s.name, s.elements, s.status, s.created_by, u.username, s.created_at 
		FROM substances s
		LEFT JOIN users u ON s.created_by = u.UID
		ORDER BY 
			CASE 
				WHEN s.status = 'pending_admin' THEN 1 
				WHEN s.status = 'pending_coworker' THEN 2 
				ELSE 3 
			END, 
			s.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SubstanceInfo
	for rows.Next() {
		var info SubstanceInfo
		var formula, name, elements, status, createdAt sql.NullString
		var createdBy sql.NullInt64
		var creatorName sql.NullString

		err := rows.Scan(&info.ID, &formula, &name, &elements, &status, &createdBy, &creatorName, &createdAt)
		if err != nil {
			continue
		}

		info.Formula = formula.String
		info.Name = name.String
		info.Elements = elements.String
		info.Status = status.String
		if createdBy.Valid {
			info.CreatedBy = int(createdBy.Int64)
		}
		if creatorName.Valid {
			info.CreatorName = creatorName.String
		} else {
			info.CreatorName = "系统"
		}
		info.CreatedAt = createdAt.String

		results = append(results, info)
	}
	return results, nil
}

// GetStatusByID 根据ID获取状态
func (r *LegacySubstanceRepository) GetStatusByID(id int) (string, error) {
	var status string
	err := r.db.QueryRow("SELECT status FROM substances WHERE id = ?", id).Scan(&status)
	if err != nil {
		return "", err
	}
	return status, nil
}

// UpdateStatus 更新物质状态
func (r *LegacySubstanceRepository) UpdateStatus(id int, status string) error {
	_, err := r.db.Exec("UPDATE substances SET status = ? WHERE id = ?", status, id)
	return err
}

// UpdateWithElements 更新物质（包括formula、name、elements）
func (r *LegacySubstanceRepository) UpdateWithElements(id int, formula, name, elements, status string) error {
	_, err := r.db.Exec("UPDATE substances SET formula = ?, name = ?, elements = ?, status = ? WHERE id = ?",
		formula, name, elements, status, id)
	return err
}

// Update 更新物质（formula、name、elements）
func (r *LegacySubstanceRepository) Update(id int, formula, name, elements string) error {
	_, err := r.db.Exec("UPDATE substances SET formula = ?, name = ?, elements = ? WHERE id = ?",
		formula, name, elements, id)
	return err
}

// Create 创建物质
func (r *LegacySubstanceRepository) Create(formula, name, elements, status string, createdBy int) error {
	_, err := r.db.Exec("INSERT INTO substances (formula, name, elements, status, created_by) VALUES (?, ?, ?, ?, ?)",
		formula, name, elements, status, createdBy)
	return err
}

// Delete 删除物质
func (r *LegacySubstanceRepository) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM substances WHERE id = ?", id)
	return err
}
