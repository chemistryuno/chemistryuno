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
