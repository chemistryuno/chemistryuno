package game

import (
	"chemistryuno/backend/database"
	"log"
	"sync"
)

var (
	validSubstances      map[string]bool
	validSubstancesMutex sync.RWMutex
)

// RebuildSubstanceCache 从已批准的反应表中重建合法物质缓存
func RebuildSubstanceCache() {
	if database.DB == nil {
		return
	}

	var reactions []database.Reaction
	err := database.DB.Where("status = ?", "approved").Find(&reactions).Error
	if err != nil {
		log.Printf("[物质缓存] 查询已批准反应失败: %v", err)
		return
	}

	newCache := make(map[string]bool)
	for _, r := range reactions {
		if r.R1 != "" {
			newCache[r.R1] = true
		}
		if r.R2 != "" {
			newCache[r.R2] = true
		}
	}

	validSubstancesMutex.Lock()
	validSubstances = newCache
	validSubstancesMutex.Unlock()

	log.Printf("[物质缓存] 已加载 %d 种合法物质", len(newCache))
}

// IsValidSubstance 检查物质是否已录入（在合法物质缓存中）
func IsValidSubstance(substance string) bool {
	// 特殊卡牌直接放行
	specialTypes := map[string]bool{
		"+2": true, "+4": true, "Au": true,
		"He": true, "Ne": true, "Ar": true, "Kr": true, "Xe": true, "Rn": true,
	}
	if specialTypes[substance] {
		return true
	}

	// 单元素符号（长度1-2的大写字母开头）视为单质，直接放行
	if len(substance) <= 2 && len(substance) > 0 && substance[0] >= 'A' && substance[0] <= 'Z' {
		return true
	}

	validSubstancesMutex.RLock()
	defer validSubstancesMutex.RUnlock()

	// 如果缓存为 nil，说明未初始化，拒绝所有复杂物质
	if validSubstances == nil {
		log.Printf("[物质校验] ⚠️  缓存未初始化，拒绝物质: %s", substance)
		return false
	}

	// 查询缓存
	isValid := validSubstances[substance]
	if !isValid {
		log.Printf("[物质校验] ❌ 物质未录入: %s", substance)
	}
	return isValid
}
