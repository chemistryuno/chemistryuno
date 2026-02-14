package game

import (
	"chemistryuno/backend/database"
	"encoding/json" // Added import
	"log"
	"strings"
	"sync"

	"gorm.io/gorm"
)

var (
	validSubstances      map[string]bool
	validSubstancesMutex sync.RWMutex
)

// RebuildSubstanceCache 从数据库重建合法物质缓存
func RebuildSubstanceCache() {
	if database.DB == nil {
		return
	}

	newCache := make(map[string]bool)

	// 只从物质百科中获取所有已批准的物质作为合法物质的唯一来源
	var substances []database.Substance
	err := database.DB.Where("status = ?", "approved").Find(&substances).Error
	if err == nil {
		for _, s := range substances {
			if s.Formula != "" {
				newCache[s.Formula] = true
			}
		}
	} else {
		log.Printf("[物质缓存] 查询已批准物质失败: %v", err)
	}

	validSubstancesMutex.Lock()
	validSubstances = newCache
	validSubstancesMutex.Unlock()

	log.Printf("[物质缓存] 已从百科加载 %d 种合法物质", len(newCache))
}

// 辅助函数：从反应方程式字符串中解析出所有物质
func parseSubstancesFromDisplay(display string) []string {
	// 常见的化学方程式符号
	seps := []string{"=", "＝", "->", "→"}
	var rhs string
	for _, sep := range seps {
		if strings.Contains(display, sep) {
			parts := strings.Split(display, sep)
			if len(parts) > 1 {
				rhs = parts[1]
				break
			}
		}
	}

	if rhs == "" {
		return nil
	}

	var result []string
	// 处理右侧生成物，按 + 分割
	products := strings.Split(rhs, "+")
	for _, p := range products {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		// 移除开头的数字系数
		i := 0
		for i < len(trimmed) && trimmed[i] >= '0' && trimmed[i] <= '9' {
			i++
		}
		formula := strings.TrimSpace(trimmed[i:])
		// 移除可能存在的状态标识如 (s), (g), (aq), (l) 或者沉淀/气体符号
		formula = strings.TrimSuffix(formula, "↓")
		formula = strings.TrimSuffix(formula, "↑")
		if formula != "" {
			result = append(result, formula)
		}
	}
	return result
}

// NormalizeSubscripts 归一化下标数字（将 Unicode 下标字符转换为普通数字）
func NormalizeSubscripts(s string) string {
	subs := map[rune]rune{
		'₀': '0', '₁': '1', '₂': '2', '₃': '3', '₄': '4',
		'₅': '5', '₆': '6', '₇': '7', '₈': '8', '₉': '9',
	}
	var res strings.Builder
	for _, r := range s {
		if v, ok := subs[r]; ok {
			res.WriteRune(v)
		} else {
			res.WriteRune(r)
		}
	}
	return res.String()
}

// IsValidSubstance 检查物质是否已录入（在合法物质缓存中）
func IsValidSubstance(substance string) bool {
	// 归一化下标数字（如 H₂O -> H2O）
	substance = NormalizeSubscripts(substance)

	// 仅对于纯游戏机制牌（非化学物质）直接放行，其他均需通过 substances 表校验
	gameMechanics := map[string]bool{
		"+2": true, "+4": true, "reverse": true,
	}
	if gameMechanics[substance] {
		return true
	}

	validSubstancesMutex.RLock()
	defer validSubstancesMutex.RUnlock()

	// 如果缓存为 nil，说明未初始化，拒绝所有物质
	if validSubstances == nil {
		log.Printf("[物质校验] ⚠️  缓存未初始化，拒绝物质: %s", substance)
		return false
	}

	// 严格从合法物质缓存（即 substances 表）中查询
	isValid := validSubstances[substance]
	if !isValid {
		log.Printf("[物质校验] ❌ 物质未录入或非合法物质: %s", substance)
	}
	return isValid
}

// SyncSubstancesFromReactions 自动遍历 reactions 表并录入 substances 表
func SyncSubstancesFromReactions() {
	if database.DB == nil {
		return
	}
	log.Println("[自动同步] 正在从化学反应同步物质百科...")
	var reactions []database.Reaction
	if err := database.DB.Select("r1, r2").Find(&reactions).Error; err != nil {
		log.Printf("[自动同步] 查询反应失败: %v", err)
		return
	}

	formulaMap := make(map[string]bool)
	for _, r := range reactions {
		if r.R1 != "" {
			formulaMap[NormalizeSubscripts(r.R1)] = true
		}
		if r.R2 != "" {
			formulaMap[NormalizeSubscripts(r.R2)] = true
		}
	}

	var formulas []string
	for f := range formulaMap {
		formulas = append(formulas, f)
	}

	if len(formulas) > 0 {
		EnsureSubstancesExist(database.DB, formulas, 100000000) // 使用系统管理员UID
	}

	// 除了从反应中同步，还确保牌组配置中出现的所有元素都被视为合法物质（作为单质）
	var deckConfigs []database.DeckConfig
	if err := database.DB.Find(&deckConfigs).Error; err == nil {
		for _, config := range deckConfigs {
			var cardMap map[string]int
			if err := json.Unmarshal(config.Cards, &cardMap); err == nil {
				var elements []string
				for cardType := range cardMap {
					// 过滤掉非化学物质的特殊卡
					isSpecial := cardType == "+2" || cardType == "+4" ||
						cardType == "He" || cardType == "Ne" ||
						cardType == "Ar" || cardType == "Kr" || cardType == "Au"
					if !isSpecial {
						elements = append(elements, cardType)
					}
				}
				EnsureSubstancesExist(database.DB, elements, 100000000)
			}
		}
	}

	log.Println("[自动同步] 物质百科同步完成")
}

// EnsureSubstancesExist 确保物质存在于百科中，若不存在则自动录入
func EnsureSubstancesExist(tx *gorm.DB, formulas []string, creatorUID uint) {
	for _, f := range formulas {
		if f == "" {
			continue
		}
		f = NormalizeSubscripts(f)
		var count int64
		tx.Model(&database.Substance{}).Where("formula = ?", f).Count(&count)
		if count == 0 {
			// 自动分析元素
			elementsMap := ParseSubstanceForElements(f)
			var elementsArr []string
			for e := range elementsMap {
				elementsArr = append(elementsArr, e)
			}
			elementsStr := strings.Join(elementsArr, ",")

			substance := &database.Substance{
				Name:         f, // 默认名称占位符为物质本身
				Formula:      f,
				Elements:     elementsStr,
				Status:       "approved", // 自动录入的物质默认设为已批准
				CreatedByUID: creatorUID,
			}
			if err := tx.Create(substance).Error; err != nil {
				log.Printf("[自动录入] 为 %s 创建物质失败: %v", f, err)
			} else {
				// 为新物质设置 group_id
				tx.Model(&database.Substance{}).Where("id = ?", substance.ID).Update("group_id", substance.ID)
				log.Printf("[自动录入] 成功将 %s 录入物质百科", f)
			}
		}
	}
}

// ParseSubstanceForElements 解析化学式获取涉及的元素
func ParseSubstanceForElements(substance string) map[string]bool {
	result := make(map[string]bool)
	i := 0
	for i < len(substance) {
		c := substance[i]
		if c >= 'A' && c <= 'Z' {
			start := i
			i++
			for i < len(substance) && substance[i] >= 'a' && substance[i] <= 'z' {
				i++
			}
			element := substance[start:i]
			result[element] = true
		} else {
			i++
		}
	}
	return result
}
