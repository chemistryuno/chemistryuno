package game

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/repository"
	"log"
)

// MarkDuplicateSubstancesForImprovement 标记名称和化学式相同的物质为待完善
func MarkDuplicateSubstancesForImprovement() {
	if database.DB == nil {
		return
	}

	log.Println("[物质维护] 正在检查重复物质...")

	substanceRepo := repository.NewSubstanceRepository()

	// 查找重复的物质
	duplicates, err := substanceRepo.FindDuplicatesByNameFormula()
	if err != nil {
		log.Printf("[物质维护] 查找重复物质失败: %v", err)
		return
	}

	if len(duplicates) == 0 {
		log.Println("[物质维护] 未发现重复物质")
		return
	}

	log.Printf("[物质维护] 发现 %d 组重复物质，正在标记为待完善...", len(duplicates))

	// 对每组重复物质进行标记
	for _, dup := range duplicates {
		// 查找该名称和化学式的所有物质
		var substances []database.Substance
		err := database.DB.Where("name = ? AND formula = ? AND status = ?",
			dup.Name, dup.Formula, "approved").Find(&substances).Error
		if err != nil {
			log.Printf("[物质维护] 查询物质失败 (%s, %s): %v", dup.Name, dup.Formula, err)
			continue
		}

		// 标记所有相关物质为待完善
		for _, sub := range substances {
			if sub.GroupID != nil {
				err := substanceRepo.MarkNeedsImprovement(*sub.GroupID, true)
				if err != nil {
					log.Printf("[物质维护] 标记物质组 %d 失败: %v", *sub.GroupID, err)
				}
			}
		}

		log.Printf("[物质维护] 已标记重复物质: %s (%s), 共 %d 条记录", dup.Name, dup.Formula, dup.Count)
	}

	log.Println("[物质维护] 重复物质标记完成")
}

// MarkIncompleteNameSubstances 标记名称等于化学式的物质为待完善
func MarkIncompleteNameSubstances() {
	if database.DB == nil {
		return
	}

	log.Println("[物质维护] 检查名称等于化学式的物质...")

	substanceRepo := repository.NewSubstanceRepository()
	substances, err := substanceRepo.FindSubstancesWhereNameEqualsFormula()

	if err != nil {
		log.Printf("[物质维护] 查询失败: %v", err)
		return
	}

	if len(substances) == 0 {
		log.Println("[物质维护] 未发现名称等于化学式的物质")
		return
	}

	log.Printf("[物质维护] 发现 %d 个待完善物质，正在标记...", len(substances))

	marked := 0
	for _, sub := range substances {
		if sub.GroupID != nil {
			if err := substanceRepo.MarkNeedsImprovement(*sub.GroupID, true); err != nil {
				log.Printf("[物质维护] 标记失败 (GroupID=%d): %v", *sub.GroupID, err)
			} else {
				marked++
			}
		}
	}

	log.Printf("[物质维护] 成功标记 %d/%d 个物质", marked, len(substances))
}

// CleanInvalidData 启动时自动清理非法反应和物质数据
func CleanInvalidData() {
	if database.DB == nil {
		return
	}

	log.Println("[数据清理] 正在检查并清理非法反应和物质数据...")

	// 1. 清理非法的反应 (Reaction)
	// 非法情况：R1 或 R2 为空，或者 Display 为空
	res := database.DB.Where("r1 = ? OR r2 = ? OR r1 IS NULL OR r2 IS NULL OR display = ? OR display IS NULL", "", "", "").Delete(&database.Reaction{})
	if res.Error != nil {
		log.Printf("[数据清理] 清理非法反应失败: %v", res.Error)
	} else if res.RowsAffected > 0 {
		log.Printf("[数据清理] 已删除 %d 条非法反应记录", res.RowsAffected)
	}

	// 2. 清理非法的物质 (Substance)
	// 非法情况：Name 或 Formula 为空
	res = database.DB.Where("name = ? OR formula = ? OR name IS NULL OR formula IS NULL", "", "").Delete(&database.Substance{})
	if res.Error != nil {
		log.Printf("[数据清理] 清理非法物质失败: %v", res.Error)
	} else if res.RowsAffected > 0 {
		log.Printf("[数据清理] 已删除 %d 条非法物质记录", res.RowsAffected)
	}

	// 3. 清理 has_invalid_elements 为 true 的已通过数据（可选，根据业务需求，如果 has_invalid_elements 是标记为非法的）
	// 这里我们只清理那些真正数据缺失的硬伤数据。标记为 has_invalid_elements 的通常是软性数据问题。

	log.Println("[数据清理] 数据清理完成")
}
