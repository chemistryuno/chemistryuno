package repository

import (
	"database/sql"
	"errors"
)

// LegacyReactionRepository 处理旧表结构的反应Repository
// 旧表结构使用 r1, r2, display, group_id 等字段
type LegacyReactionRepository struct {
	db *sql.DB
}

func NewLegacyReactionRepository(db *sql.DB) *LegacyReactionRepository {
	return &LegacyReactionRepository{db: db}
}

// ReactionGroupInfo 反应组信息
type ReactionGroupInfo struct {
	ID          int
	Display     string
	Status      string
	GroupID     string
	CreatedBy   int
	CreatorName string
	CreatedAt   string
}

// ReactionInfo 反应信息
type ReactionInfo struct {
	ID        int
	Display   string
	R1        string
	R2        string
	CreatedAt string
}

// GetAllReactionsGrouped 获取所有反应（按组分组）
func (r *LegacyReactionRepository) GetAllReactionsGrouped() ([]ReactionGroupInfo, error) {
	rows, err := r.db.Query(`
		SELECT MIN(r.id), r.display, r.status, r.group_id, r.created_by, u.username, MIN(r.created_at)
		FROM reactions r
		LEFT JOIN users u ON r.created_by = u.UID
		GROUP BY r.display, r.status, r.group_id, r.created_by, u.username
		ORDER BY 
			CASE 
				WHEN r.status = 'pending_admin' THEN 1 
				WHEN r.status = 'pending_coworker' THEN 2 
				ELSE 3 
			END, 
			MIN(r.created_at) DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ReactionGroupInfo
	for rows.Next() {
		var info ReactionGroupInfo
		err := rows.Scan(&info.ID, &info.Display, &info.Status, &info.GroupID,
			&info.CreatedBy, &info.CreatorName, &info.CreatedAt)
		if err != nil {
			continue
		}
		results = append(results, info)
	}
	return results, nil
}

// GetApprovedReactionsGrouped 获取已批准的反应（按组分组）
func (r *LegacyReactionRepository) GetApprovedReactionsGrouped() ([]ReactionInfo, error) {
	rows, err := r.db.Query(`
		SELECT MIN(id), display, r1, r2, MIN(created_at)
		FROM reactions
		WHERE status = 'approved'
		GROUP BY group_id, display
		ORDER BY MIN(created_at) DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ReactionInfo
	for rows.Next() {
		var info ReactionInfo
		err := rows.Scan(&info.ID, &info.Display, &info.R1, &info.R2, &info.CreatedAt)
		if err != nil {
			continue
		}
		results = append(results, info)
	}
	return results, nil
}

// GetMyReactions 获取我提交的反应
func (r *LegacyReactionRepository) GetMyReactions(uid int) ([]ReactionGroupInfo, error) {
	rows, err := r.db.Query(`
		SELECT MIN(id), display, status, MIN(created_at)
		FROM reactions
		WHERE created_by = ?
		GROUP BY group_id, display, status
		ORDER BY MIN(created_at) DESC
	`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ReactionGroupInfo
	for rows.Next() {
		var info ReactionGroupInfo
		err := rows.Scan(&info.ID, &info.Display, &info.Status, &info.CreatedAt)
		if err != nil {
			continue
		}
		results = append(results, info)
	}
	return results, nil
}

// GetStatusByGroupID 根据group_id获取状态
func (r *LegacyReactionRepository) GetStatusByGroupID(groupID string) (string, error) {
	var status string
	err := r.db.QueryRow("SELECT status FROM reactions WHERE group_id = ? LIMIT 1", groupID).Scan(&status)
	if err != nil {
		return "", err
	}
	return status, nil
}

// UpdateStatusByGroupID 更新组状态
func (r *LegacyReactionRepository) UpdateStatusByGroupID(groupID, status string) error {
	_, err := r.db.Exec("UPDATE reactions SET status = ? WHERE group_id = ?", status, groupID)
	return err
}

// GetGroupIDAndCreatorByID 根据ID获取group_id和创建者
func (r *LegacyReactionRepository) GetGroupIDAndCreatorByID(id int) (string, int, error) {
	var groupID string
	var createdBy int
	err := r.db.QueryRow("SELECT group_id, created_by FROM reactions WHERE id = ?", id).Scan(&groupID, &createdBy)
	if err != nil {
		return "", 0, err
	}
	return groupID, createdBy, nil
}

// DeleteByGroupID 删除组内所有反应
func (r *LegacyReactionRepository) DeleteByGroupID(groupID string) error {
	_, err := r.db.Exec("DELETE FROM reactions WHERE group_id = ?", groupID)
	return err
}

// GetCreatorByGroupID 根据group_id获取创建者
func (r *LegacyReactionRepository) GetCreatorByGroupID(groupID string) (int, error) {
	var creatorID int
	err := r.db.QueryRow("SELECT created_by FROM reactions WHERE group_id = ? LIMIT 1", groupID).Scan(&creatorID)
	if err != nil {
		return 0, err
	}
	return creatorID, nil
}

// CheckDuplicateReactants 检查反应物组合是否已存在
func (r *LegacyReactionRepository) CheckDuplicateReactants(r1, r2, excludeGroupID string) (bool, string, error) {
	var existingDisplay string
	query := `
		SELECT display FROM reactions 
		WHERE status != 'rejected' 
		AND ((r1 = ? AND r2 = ?) OR (r1 = ? AND r2 = ?))`
	args := []interface{}{r1, r2, r2, r1}

	if excludeGroupID != "" {
		query += " AND group_id != ?"
		args = append(args, excludeGroupID)
	}
	query += " LIMIT 1"

	err := r.db.QueryRow(query, args...).Scan(&existingDisplay)
	if err == sql.ErrNoRows {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	return true, existingDisplay, nil
}

// BeginTx 开始事务
func (r *LegacyReactionRepository) BeginTx() (*sql.Tx, error) {
	return r.db.Begin()
}

// SaveReaction 保存反应（在事务中）
func (r *LegacyReactionRepository) SaveReaction(tx *sql.Tx, r1, r2, display, status, groupID string, createdBy int) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}
	_, err := tx.Exec(`
		INSERT INTO reactions (r1, r2, display, status, group_id, created_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`, r1, r2, display, status, groupID, createdBy)
	return err
}
