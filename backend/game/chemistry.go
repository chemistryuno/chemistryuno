package game

import (
	"chemistryuno/database"
	"chemistryuno/models"
)

// 化学反应数据库（简化版，实际应该更完整）
var reactionDB = map[string][]string{
	// 水相关
	"H2O": {"H2", "O2", "H2O2", "NaOH", "HCl", "H2SO4", "Na2O", "CaO", "CO2"},

	// 氢气
	"H2": {"O2", "Cl2", "N2", "C", "CuO", "Fe2O3"},

	// 氧气
	"O2": {"H2", "C", "S", "P", "Fe", "Cu", "Mg", "Al", "Na", "K", "Ca"},

	// 酸
	"HCl":   {"NaOH", "Na2CO3", "Fe", "Zn", "Mg", "Al", "CuO", "FeO"},
	"H2SO4": {"NaOH", "BaCl2", "Fe", "Cu", "Zn", "Mg", "Al"},

	// 碱
	"NaOH": {"HCl", "H2SO4", "CO2", "CuSO4", "FeCl3", "Al"},

	// 盐
	"NaCl":   {"AgNO3", "H2SO4"},
	"Na2CO3": {"HCl", "CaCl2", "BaCl2"},
	"CuSO4":  {"NaOH", "Fe", "Zn", "BaCl2"},

	// 氧化物
	"CO2":   {"NaOH", "Ca(OH)2", "C", "H2O"},
	"CaO":   {"H2O", "HCl", "CO2"},
	"CuO":   {"H2", "C", "HCl", "H2SO4"},
	"Fe2O3": {"H2", "C", "CO", "HCl"},

	// 单质
	"Fe":  {"O2", "S", "HCl", "H2SO4", "CuSO4", "AgNO3"},
	"Cu":  {"O2", "S", "Cl2", "AgNO3", "H2SO4"},
	"Zn":  {"O2", "HCl", "H2SO4", "CuSO4"},
	"Mg":  {"O2", "HCl", "H2SO4"},
	"Al":  {"O2", "HCl", "H2SO4", "NaOH"},
	"C":   {"O2", "CuO", "Fe2O3", "CO2"},
	"S":   {"O2", "Fe", "Cu", "Hg"},
	"Cl2": {"H2", "Fe", "Cu", "Na", "NaBr", "KI"},
	"Br2": {"H2", "KI", "Fe"},
	"I2":  {"H2", "Zn"},

	// 稀有气体与特殊卡牌（作为特殊物质，通常允许反应）
	"He": {"*"},
	"Ne": {"*"},
	"Ar": {"*"},
	"Kr": {"*"},
	"Au": {"*"},
	"+2": {"*"},
	"+4": {"*"},
	// "Choice": {"*"},
}

// 元素组成物质的映射
var elementSubstances = map[string][]string{
	"H":  {"H2", "H2O", "HCl", "H2SO4", "NH3", "CH4", "H2O2", "NaOH", "Ca(OH)2", "HI", "HBr", "HF", "H2S", "H2SO3", "H3PO4", "KH", "NaH", "MgH2", "CaH2", "BaH2"},
	"O":  {"O2", "H2O", "CO2", "CaO", "CuO", "Fe2O3", "Al2O3", "Na2O", "MgO", "SO2", "SO3", "H2SO4", "NaOH", "Ca(OH)2", "MnO2", "Fe3O4", "Ag2O", "K2O", "BaO", "P2O5", "H2O2", "Na2O2", "NaClO", "Ca(ClO)2", "HBrO", "HClO"},
	"C":  {"C", "CO2", "CO", "CH4", "C2H5OH", "CH3COOH", "CaCO3", "Na2CO3"},
	"N":  {"N2", "NH3", "NO", "NO2", "HNO3", "NH4Cl", "KNO3", "NaNO3", "NH4NO3", "AgNO3", "Cu(NO3)2", "Mg(NO3)2", "Ca(NO3)2", "Ba(NO3)2", "Fe(NO3)3"},
	"F":  {"F2", "HF", "CaF2", "MgF2", "AlF3", "ZnF2", "NaF", "KF", "BaF2", "SiF4", "PF5", "CF4"},
	"Na": {"Na", "NaCl", "NaOH", "Na2O", "Na2SO4", "Na2CO3", "NaHCO3", "Na2O2", "NaI", "NaBr", "Na2S", "Na2SO3", "Na2S2O3", "Na3PO4", "NaAlO2", "NaClO", "NaH"},
	"Mg": {"Mg", "MgO", "MgCl2", "MgSO4", "Mg(OH)2", "MgS", "MgBr2", "MgI2", "Mg(NO3)2", "MgH2", "Mg3(PO4)2"},
	"Al": {"Al", "Al2O3", "AlCl3", "Al(OH)3", "Al2(SO4)3", "Al2S3", "AlBr3", "AlI3", "AlF3", "NaAlO2", "AlPO4"},
	"Si": {"Si", "SiO2", "H2SiO3", "SiF4"},
	"P":  {"P", "P2O5", "H3PO4", "Na3PO4", "K3PO4", "Ca3(PO4)2", "PCl3", "P2S5", "PF5"},
	"S":  {"S", "SO2", "SO3", "H2SO4", "H2S", "CuSO4", "FeSO4", "ZnSO4", "FeS", "Cu2S", "Na2S", "Na2SO3", "H2SO3", "BaSO4", "Na2S2O3", "Al2S3", "ZnS", "P2S5", "K2S", "MgS", "CaS", "BaS", "Na2SO4", "K2SO4", "MgSO4", "CaSO4", "Fe2(SO4)3"},
	"Cl": {"Cl2", "HCl", "NaCl", "KCl", "CaCl2", "MgCl2", "AlCl3", "FeCl3", "CuCl2", "FeCl2", "ZnCl2", "AgCl", "HgCl2", "NH4Cl", "BaCl2", "PCl3", "KClO3", "NaClO", "Ca(ClO)2", "HClO"},
	"K":  {"K", "KCl", "KOH", "K2O", "KNO3", "K2SO4", "K2CO3", "K3PO4", "KI", "KBr", "K2S", "KClO3", "KH", "KO2"},
	"Ca": {"Ca", "CaO", "Ca(OH)2", "CaCl2", "CaCO3", "CaSO4", "Ca3(PO4)2", "CaF2", "CaS", "Ca(NO3)2", "Ca(ClO)2", "CaH2"},
	"Mn": {"Mn", "MnO2", "KMnO4"},
	"Fe": {"Fe", "FeO", "Fe2O3", "Fe3O4", "FeCl2", "FeCl3", "FeSO4", "Fe(OH)2", "Fe(OH)3", "FeS", "FeBr3", "FeI2", "Fe(NO3)2", "Fe(NO3)3", "Fe2(SO4)3", "Fe3(PO4)2", "FeF3"},
	"Cu": {"Cu", "CuO", "Cu2O", "CuCl2", "CuSO4", "Cu(OH)2", "Cu2S", "CuBr2", "CuI", "Cu(NO3)2", "CuS"},
	"Zn": {"Zn", "ZnO", "ZnCl2", "ZnSO4", "ZnS", "ZnBr2", "ZnI2", "Zn3(PO4)2", "ZnF2"},
	"Br": {"Br2", "HBr", "NaBr", "KBr", "AgBr", "FeBr3", "MgBr2", "AlBr3", "ZnBr2", "CuBr2", "CaBr2", "HgBr2", "BaBr2", "HBrO"},
	"I":  {"I2", "HI", "KI", "NaI", "AgI", "FeI2", "MgI2", "AlI3", "ZnI2", "CuI", "CaI2", "HgI2", "BaI2"},
	"Ag": {"Ag", "AgNO3", "AgCl", "Ag2O", "AgBr", "AgI", "Ag2S", "AgF"},
	"Hg": {"Hg", "HgO", "HgCl2", "HgBr2", "HgI2"},
	"Ba": {"Ba", "BaO", "Ba(OH)2", "BaCl2", "BaSO4", "Ba(NO3)2", "BaF2", "BaS", "BaH2", "Ba3(PO4)2"},
	// 特殊卡牌元素
	"He": {"He"},
	"Ne": {"Ne"},
	"Ar": {"Ar"},
	"Kr": {"Kr"},
	"Au": {"Au"},
	"+2": {"+2"},
	"+4": {"+4"},
	// "Choice": {"Choice"},
}

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

	// 查找所有可能的物质
	for element := range elementMap {
		if substances, ok := elementSubstances[element]; ok {
			for _, sub := range substances {
				if canFormSubstance(sub, elementMap) {
					substanceSet[sub] = true
				}
			}
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
		// 不考虑系数，只考虑元素种类是否存在
		if elements[elem] < 1 {
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
		"He": true, "Ne": true, "Ar": true, "Kr": true, "Au": true, "+2": true, "+4": true,
	}
	if specialSubstances[substance1] || specialSubstances[substance2] {
		return true
	}

	// 首先检查硬编码的反应数据库
	if products, ok := reactionDB[substance1]; ok {
		for _, product := range products {
			if product == substance2 || product == "*" {
				return true
			}
		}
	}

	if products, ok := reactionDB[substance2]; ok {
		for _, product := range products {
			if product == substance1 || product == "*" {
				return true
			}
		}
	}

	// 检查数据库中的自定义反应
	if database.DB != nil {
		var count int
		err := database.DB.QueryRow("SELECT COUNT(*) FROM reactions WHERE r1 = ? AND r2 = ? AND status = 'approved'", substance1, substance2).Scan(&count)
		if err == nil && count > 0 {
			return true
		}
	}

	// 如果数据库中也没有，则使用通用的化学逻辑判定
	return JudgeReaction(substance1, substance2)
}

// 获取能与指定物质反应的所有物质
func GetReactableSubstances(substance string) []string {
	var results []string
	if products, ok := reactionDB[substance]; ok {
		results = append(results, products...)
	}

	// 从数据库获取
	if database.DB != nil {
		rows, err := database.DB.Query("SELECT r2 FROM reactions WHERE r1 = ? AND status = 'approved'", substance)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var r2 string
				if err := rows.Scan(&r2); err == nil {
					// 避免重复
					found := false
					for _, r_exist := range results {
						if r_exist == r2 {
							found = true
							break
						}
					}
					if !found {
						results = append(results, r2)
					}
				}
			}
		}
	}
	return results
}
