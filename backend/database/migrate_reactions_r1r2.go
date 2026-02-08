package database

import (
	"log"
	"strings"
)

// MigrateReactionsToR1R2 将reactions表从Reactants/Products字段迁移到R1/R2字段
// 该函数在数据库初始化后调用，进行schema升级和数据迁移
func MigrateReactionsToR1R2() error {
	log.Println("开始迁移reactions表到R1/R2结构...")

	// 步骤1: 添加新列（如果不存在），使用原生SQL添加带默认值的列
	if !DB.Migrator().HasColumn(&Reaction{}, "r1") {
		log.Println("添加r1列...")
		// SQLite不允许添加NOT NULL列到已有数据的表，所以先添加可为NULL的列
		if err := DB.Exec("ALTER TABLE reactions ADD COLUMN r1 TEXT").Error; err != nil {
			log.Printf("添加r1列失败: %v", err)
			return err
		}
	}

	if !DB.Migrator().HasColumn(&Reaction{}, "r2") {
		log.Println("添加r2列...")
		if err := DB.Exec("ALTER TABLE reactions ADD COLUMN r2 TEXT").Error; err != nil {
			log.Printf("添加r2列失败: %v", err)
			return err
		}
	}

	// 步骤2: 检查是否还有旧字段
	hasReactants := DB.Migrator().HasColumn(&Reaction{}, "reactants")
	hasProducts := DB.Migrator().HasColumn(&Reaction{}, "products")

	if !hasReactants && !hasProducts {
		log.Println("旧字段已删除，跳过数据迁移")
		return nil
	}

	// 步骤3: 从"A+B"格式填充R1/R2
	log.Println("从reactants字段迁移数据...")

	// 使用原生SQL查询（因为Reaction结构体已经没有Reactants/Products字段）
	var oldReactions []struct {
		ID        uint
		Reactants string
		Products  string
		GroupID   *uint
	}

	// 查询所有R1为空的记录（需要迁移）
	err := DB.Raw("SELECT id, reactants, products, group_id FROM reactions WHERE r1 IS NULL OR r1 = ''").Scan(&oldReactions).Error
	if err != nil {
		log.Printf("查询旧数据失败: %v", err)
		return err
	}

	log.Printf("找到 %d 条需要迁移的记录", len(oldReactions))

	// 按组ID分组处理
	groupMap := make(map[uint][]struct {
		ID        uint
		Reactants string
		Products  string
		GroupID   *uint
	})

	var noGroupReactions []struct {
		ID        uint
		Reactants string
		Products  string
		GroupID   *uint
	}

	for _, r := range oldReactions {
		if r.GroupID != nil {
			groupMap[*r.GroupID] = append(groupMap[*r.GroupID], r)
		} else {
			noGroupReactions = append(noGroupReactions, r)
		}
	}

	// 处理有GroupID的记录（可能是双向记录）
	migratedCount := 0
	deletedCount := 0

	for groupID, group := range groupMap {
		if len(group) == 1 {
			// 单条记录，检查是否为"A+B"格式
			r := group[0]
			if strings.Contains(r.Reactants, "+") {
				// "A+B"格式
				parts := strings.Split(r.Reactants, "+")
				if len(parts) == 2 {
					r1 := strings.TrimSpace(parts[0])
					r2 := strings.TrimSpace(parts[1])

					// Canonical ordering
					if r1 > r2 {
						r1, r2 = r2, r1
					}

					err := DB.Exec("UPDATE reactions SET r1 = ?, r2 = ? WHERE id = ?", r1, r2, r.ID).Error
					if err != nil {
						log.Printf("更新记录 %d 失败: %v", r.ID, err)
					} else {
						migratedCount++
					}
				}
			} else {
				// 可能是单个反应物的格式（Reactants=A, Products=B）
				r1 := strings.TrimSpace(r.Reactants)
				r2 := strings.TrimSpace(r.Products)

				if r1 != "" && r2 != "" {
					// Canonical ordering
					if r1 > r2 {
						r1, r2 = r2, r1
					}

					err := DB.Exec("UPDATE reactions SET r1 = ?, r2 = ? WHERE id = ?", r1, r2, r.ID).Error
					if err != nil {
						log.Printf("更新记录 %d 失败: %v", r.ID, err)
					} else {
						migratedCount++
					}
				}
			}
		} else if len(group) == 2 {
			// 双向记录，保留一条canonical记录
			r1 := strings.TrimSpace(group[0].Reactants)
			r2 := strings.TrimSpace(group[0].Products)

			// Canonical ordering
			if r1 > r2 {
				r1, r2 = r2, r1
			}

			// 更新第一条
			err := DB.Exec("UPDATE reactions SET r1 = ?, r2 = ? WHERE id = ?", r1, r2, group[0].ID).Error
			if err != nil {
				log.Printf("更新记录 %d 失败: %v", group[0].ID, err)
			} else {
				migratedCount++
			}

			// 删除第二条
			err = DB.Exec("DELETE FROM reactions WHERE id = ?", group[1].ID).Error
			if err != nil {
				log.Printf("删除重复记录 %d 失败: %v", group[1].ID, err)
			} else {
				deletedCount++
			}
		} else {
			log.Printf("警告: GroupID %d 有 %d 条记录（异常）", groupID, len(group))
		}
	}

	// 处理没有GroupID的记录
	for _, r := range noGroupReactions {
		if strings.Contains(r.Reactants, "+") {
			// "A+B"格式
			parts := strings.Split(r.Reactants, "+")
			if len(parts) == 2 {
				r1 := strings.TrimSpace(parts[0])
				r2 := strings.TrimSpace(parts[1])

				// Canonical ordering
				if r1 > r2 {
					r1, r2 = r2, r1
				}

				err := DB.Exec("UPDATE reactions SET r1 = ?, r2 = ? WHERE id = ?", r1, r2, r.ID).Error
				if err != nil {
					log.Printf("更新记录 %d 失败: %v", r.ID, err)
				} else {
					migratedCount++
				}
			}
		} else {
			// Reactants=A, Products=B格式
			r1 := strings.TrimSpace(r.Reactants)
			r2 := strings.TrimSpace(r.Products)

			if r1 != "" && r2 != "" {
				// Canonical ordering
				if r1 > r2 {
					r1, r2 = r2, r1
				}

				err := DB.Exec("UPDATE reactions SET r1 = ?, r2 = ? WHERE id = ?", r1, r2, r.ID).Error
				if err != nil {
					log.Printf("更新记录 %d 失败: %v", r.ID, err)
				} else {
					migratedCount++
				}
			}
		}
	}

	log.Printf("迁移完成: 更新了 %d 条记录，删除了 %d 条重复记录", migratedCount, deletedCount)

	// 步骤4: 创建索引
	log.Println("创建索引...")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_reactions_r1r2_status ON reactions(r1, r2, status)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_reactions_r2r1_status ON reactions(r2, r1, status)")

	// 步骤5: 验证数据完整性
	log.Println("验证数据完整性...")

	var nullCount int64
	DB.Raw("SELECT COUNT(*) FROM reactions WHERE (r1 IS NULL OR r1 = '') OR (r2 IS NULL OR r2 = '')").Scan(&nullCount)
	if nullCount > 0 {
		log.Printf("警告: 仍有 %d 条记录的R1或R2为空", nullCount)
	}

	var wrongOrderCount int64
	DB.Raw("SELECT COUNT(*) FROM reactions WHERE r1 > r2").Scan(&wrongOrderCount)
	if wrongOrderCount > 0 {
		log.Printf("警告: 有 %d 条记录不符合canonical ordering (r1 > r2)", wrongOrderCount)
	}

	// 步骤6: 删除旧字段（如果用户选择了完全删除）
	if hasReactants {
		log.Println("删除旧的reactants列...")
		if err := DB.Migrator().DropColumn(&Reaction{}, "reactants"); err != nil {
			log.Printf("删除reactants列失败: %v", err)
			// 不返回错误，允许继续
		}
	}

	if hasProducts {
		log.Println("删除旧的products列...")
		if err := DB.Migrator().DropColumn(&Reaction{}, "products"); err != nil {
			log.Printf("删除products列失败: %v", err)
			// 不返回错误，允许继续
		}
	}

	// 删除不再使用的bidirection字段
	if DB.Migrator().HasColumn(&Reaction{}, "bidirection") {
		log.Println("删除bidirection列...")
		if err := DB.Migrator().DropColumn(&Reaction{}, "bidirection"); err != nil {
			log.Printf("删除bidirection列失败: %v", err)
		}
	}

	log.Println("reactions表迁移完成！")
	return nil
}
