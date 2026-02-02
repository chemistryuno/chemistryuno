package game

import (
	"chemistryuno/database"
	"chemistryuno/models"
)

// 根据手牌元素获取可以组成的物质
func GetSubstancesFromElements(cards []models.Card) []string {
	elementMap := make(map[string]int)

	// 统计每种元素的数量
	for _, card := range cards {
		if card.Effect == "" { // 只处理普通元素牌
			elementMap[card.Type]++
		}
	}

	substanceSet := make(map[string]bool)

	// 从数据库中获取所有可能的物质并进行手牌校验
	if database.DB != nil {
		rows, err := database.DB.Query("SELECT formula FROM substances")
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var formula string
				if err := rows.Scan(&formula); err == nil {
					if canFormSubstance(formula, elementMap) {
						substanceSet[formula] = true
					}
				}
			}
		}
	}

	// 稀有气体单质处理
	nobleGases := []string{"He", "Ne", "Ar", "Kr", "Xe", "Rn", "Au"}
	for _, gas := range nobleGases {
		if elementMap[gas] > 0 {
			substanceSet[gas] = true
		}
	}

	// 转换为列表
	result := []string{}
	for sub := range substanceSet {
		result = append(result, sub)
	}

	return result
}

// 检查是否可以用当前元素组成某个物质
func canFormSubstance(substance string, elements map[string]int) bool {
	required := parseSubstance(substance)
	for elem, count := range required {
		// 校验手牌中该元素的数量是否满足化学式需求
		if elements[elem] < count {
			return false
		}
	}
	return true
}

// 解析物质化学式，返回所需元素及数量
func parseSubstance(substance string) map[string]int {
	result := make(map[string]int)
	stack := []map[string]int{result}

	i := 0
	for i < len(substance) {
		c := substance[i]
		if c == '(' {
			stack = append(stack, make(map[string]int))
			i++
		} else if c == ')' {
			i++
			count := 0
			for i < len(substance) && substance[i] >= '0' && substance[i] <= '9' {
				count = count*10 + int(substance[i]-'0')
				i++
			}
			if count == 0 {
				count = 1
			}

			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			parent := stack[len(stack)-1]

			for k, v := range top {
				parent[k] += v * count
			}
		} else if c >= 'A' && c <= 'Z' {
			start := i
			i++
			for i < len(substance) && substance[i] >= 'a' && substance[i] <= 'z' {
				i++
			}
			element := substance[start:i]

			count := 0
			for i < len(substance) && substance[i] >= '0' && substance[i] <= '9' {
				count = count*10 + int(substance[i]-'0')
				i++
			}
			if count == 0 {
				count = 1
			}

			stack[len(stack)-1][element] += count
		} else {
			i++
		}
	}

	// 如果解析结果为空且物质长度不为0，可能是一些特殊符号或错误
	if len(result) == 0 && len(substance) > 0 {
		result[substance] = 1
	}

	return result
}

// 检查两个物质是否能反应
func CanReact(substance1, substance2 string) bool {
	// 特殊卡牌逻辑：稀有气体和功能牌可以与任何物质反应（即可以接在任何牌后面）
	specialSubstances := map[string]bool{
		"He": true, "Ne": true, "Ar": true, "Kr": true, "Xe": true, "Rn": true,
		"Au": true, "+2": true, "+4": true,
	}
	if specialSubstances[substance1] || specialSubstances[substance2] {
		return true
	}

	// 优先且唯一通过数据库查询判定，确保反应的严谨性
	if database.DB != nil {
		var count int
		// 查询已批准的反应，检查该组合是否存在
		err := database.DB.QueryRow("SELECT COUNT(*) FROM reactions WHERE ((r1 = ? AND r2 = ?) OR (r1 = ? AND r2 = ?)) AND status = 'approved'", substance1, substance2, substance2, substance1).Scan(&count)
		if err == nil && count > 0 {
			return true
		}
	}

	return false
}

// 获取能与指定物质反应的所有物质
func GetReactableSubstances(substance string) []string {
	var results []string

	// 严格从数据库获取所有允许接续的反应物
	if database.DB != nil {
		rows, err := database.DB.Query("SELECT r1, r2 FROM reactions WHERE (r1 = ? OR r2 = ?) AND status = 'approved'", substance, substance)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var r1, r2 string
				if err := rows.Scan(&r1, &r2); err == nil {
					target := r2
					if r2 == substance {
						target = r1
					}

					// 避免重复
					found := false
					for _, r_exist := range results {
						if r_exist == target {
							found = true
							break
						}
					}
					if !found {
						results = append(results, target)
					}
				}
			}
		}
	}
	return results
}
