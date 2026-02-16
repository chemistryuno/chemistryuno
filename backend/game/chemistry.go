package game

import (
	"chemistryuno/backend/database"
	"chemistryuno/backend/models"
	"chemistryuno/backend/repository"
	"strings"
)

// 根据手牌元素获取可以组成的物质
func GetSubstancesFromElements(cards []models.Card) []string {
	elementMap := make(map[string]int)

	// 统计每种元素的数量
	for _, card := range cards {
		// 无论是普通元素还是功能牌（如 Au），都将其名称作为元素存入 map
		// 这样 AI 就能在计算可出物质时考虑功能牌
		elementMap[card.Type]++
	}

	substanceSet := make(map[string]bool)

	// 从数据库中获取所有可能的物质并进行手牌校验
	if database.DB != nil {
		substances, err := repository.SubstanceRepo.FindApproved()
		if err == nil {
			for _, sub := range substances {
				if canFormSubstance(sub.Name, elementMap) {
					substanceSet[sub.Name] = true
				}
			}
		}
	}

	// 允许打出由单个原子组成的单质（直接使用手牌中的元素符号）
	for elem, count := range elementMap {
		if count > 0 {
			substanceSet[elem] = true
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
	for elem := range required {
		// 普通反应时，仅考虑元素种类，不考虑元素系数
		if elements[elem] < 1 {
			return false
		}
	}
	return true
}

// 解析物质化学式，返回所需元素及数量
func parseSubstance(substance string) map[string]int {
	substance = NormalizeSubscripts(substance)
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
	// 去除可能存在的空格
	s1 := strings.ReplaceAll(substance1, " ", "")
	s2 := strings.ReplaceAll(substance2, " ", "")

	// 特殊卡牌逻辑：稀有气体和功能牌可以与任何物质反应（即可以接在任何牌后面）
	specialSubstances := map[string]bool{
		"He": true, "Ne": true, "Ar": true, "Kr": true, "Xe": true, "Rn": true,
		"Au": true, "+2": true, "+4": true,
	}
	if specialSubstances[s1] || specialSubstances[s2] {
		return true
	}

	// 完全依赖数据库查询验证反应是否存在
	// 优势：规则统一存储，易于管理和扩展；支持用户自定义反应
	if database.DB != nil {
		exists, err := repository.ReactionRepo.CheckReactionExists(s1, s2)
		if err == nil && exists {
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
		reactions, err := repository.ReactionRepo.FindReactionsBySubstance(substance)
		if err == nil {
			for _, reaction := range reactions {
				// 找到另一个反应物
				if reaction.R1 == substance {
					if !contains(results, reaction.R2) {
						results = append(results, reaction.R2)
					}
				} else if reaction.R2 == substance {
					if !contains(results, reaction.R1) {
						results = append(results, reaction.R1)
					}
				}
			}
		}
	}
	return results
}

// 辅助函数：检查字符串数组是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
