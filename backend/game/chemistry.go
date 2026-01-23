package game

import "chemistryuno/models"

// 化学反应数据库（简化版，实际应该更完整）
var reactionDB = map[string][]string{
	// 水相关
	"H2O": {"H2", "O2", "H2O2", "NaOH", "HCl", "H2SO4", "Na2O", "CaO", "CO2"},
	
	// 氢气
	"H2": {"O2", "Cl2", "N2", "C", "CuO", "Fe2O3"},
	
	// 氧气
	"O2": {"H2", "C", "S", "P", "Fe", "Cu", "Mg", "Al", "Na", "K", "Ca"},
	
	// 酸
	"HCl": {"NaOH", "Na2CO3", "Fe", "Zn", "Mg", "Al", "CuO", "FeO"},
	"H2SO4": {"NaOH", "BaCl2", "Fe", "Cu", "Zn", "Mg", "Al"},
	
	// 碱
	"NaOH": {"HCl", "H2SO4", "CO2", "CuSO4", "FeCl3", "Al"},
	
	// 盐
	"NaCl": {"AgNO3", "H2SO4"},
	"Na2CO3": {"HCl", "CaCl2", "BaCl2"},
	"CuSO4": {"NaOH", "Fe", "Zn", "BaCl2"},
	
	// 氧化物
	"CO2": {"NaOH", "Ca(OH)2", "C", "H2O"},
	"CaO": {"H2O", "HCl", "CO2"},
	"CuO": {"H2", "C", "HCl", "H2SO4"},
	"Fe2O3": {"H2", "C", "CO", "HCl"},
	
	// 单质
	"Fe": {"O2", "S", "HCl", "H2SO4", "CuSO4", "AgNO3"},
	"Cu": {"O2", "S", "Cl2", "AgNO3", "H2SO4"},
	"Zn": {"O2", "HCl", "H2SO4", "CuSO4"},
	"Mg": {"O2", "HCl", "H2SO4"},
	"Al": {"O2", "HCl", "H2SO4", "NaOH"},
	"C": {"O2", "CuO", "Fe2O3", "CO2"},
	"S": {"O2", "Fe", "Cu", "Hg"},
	"Cl2": {"H2", "Fe", "Cu", "Na", "NaBr", "KI"},
	"Br2": {"H2", "KI", "Fe"},
	"I2": {"H2", "Zn"},
	
	// 金属盐
	"AgNO3": {"NaCl", "HCl", "Fe", "Cu", "Zn"},
	"BaCl2": {"H2SO4", "Na2SO4", "Na2CO3"},
	"FeCl3": {"NaOH", "KSCN", "Fe"},
}

// 元素组成物质的映射
var elementSubstances = map[string][]string{
	"H":  {"H2", "H2O", "HCl", "H2SO4", "NH3", "CH4", "H2O2", "NaOH", "Ca(OH)2"},
	"O":  {"O2", "H2O", "CO2", "CaO", "CuO", "Fe2O3", "Al2O3", "Na2O", "MgO", "SO2", "SO3", "H2SO4", "NaOH", "Ca(OH)2"},
	"C":  {"C", "CO2", "CO", "CH4", "C2H5OH", "CH3COOH", "CaCO3", "Na2CO3"},
	"N":  {"N2", "NH3", "NO", "NO2", "HNO3", "NH4Cl"},
	"F":  {"F2", "HF", "CaF2"},
	"Na": {"Na", "NaCl", "NaOH", "Na2O", "Na2SO4", "Na2CO3", "NaHCO3"},
	"Mg": {"Mg", "MgO", "MgCl2", "MgSO4"},
	"Al": {"Al", "Al2O3", "AlCl3", "Al(OH)3"},
	"Si": {"Si", "SiO2", "H2SiO3"},
	"P":  {"P", "P2O5", "H3PO4"},
	"S":  {"S", "SO2", "SO3", "H2SO4", "H2S", "CuSO4", "FeSO4", "ZnSO4"},
	"Cl": {"Cl2", "HCl", "NaCl", "KCl", "CaCl2", "MgCl2", "AlCl3", "FeCl3", "CuCl2"},
	"K":  {"K", "KCl", "KOH", "K2O", "KNO3", "K2SO4", "K2CO3"},
	"Ca": {"Ca", "CaO", "Ca(OH)2", "CaCl2", "CaCO3", "CaSO4"},
	"Mn": {"Mn", "MnO2", "KMnO4"},
	"Fe": {"Fe", "FeO", "Fe2O3", "Fe3O4", "FeCl2", "FeCl3", "FeSO4"},
	"Cu": {"Cu", "CuO", "Cu2O", "CuCl2", "CuSO4"},
	"Zn": {"Zn", "ZnO", "ZnCl2", "ZnSO4"},
	"Br": {"Br2", "HBr", "NaBr", "KBr"},
	"I":  {"I2", "HI", "KI"},
	"Ag": {"Ag", "AgNO3", "AgCl", "Ag2O"},
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
	for elem, count := range required {
		if elements[elem] < count {
			return false
		}
	}
	return true
}

// 解析物质化学式，返回所需元素及数量
func parseSubstance(substance string) map[string]int {
	// 简化版解析，实际应该用更复杂的化学式解析器
	result := make(map[string]int)
	
	// 这里是硬编码的常见物质组成
	substanceElements := map[string]map[string]int{
		"H2":        {"H": 2},
		"O2":        {"O": 2},
		"H2O":       {"H": 2, "O": 1},
		"CO2":       {"C": 1, "O": 2},
		"NaCl":      {"Na": 1, "Cl": 1},
		"HCl":       {"H": 1, "Cl": 1},
		"NaOH":      {"Na": 1, "O": 1, "H": 1},
		"H2SO4":     {"H": 2, "S": 1, "O": 4},
		"CaCO3":     {"Ca": 1, "C": 1, "O": 3},
		"Fe2O3":     {"Fe": 2, "O": 3},
		"CuSO4":     {"Cu": 1, "S": 1, "O": 4},
		"AgNO3":     {"Ag": 1, "N": 1, "O": 3},
		// ... 更多物质
	}
	
	if elems, ok := substanceElements[substance]; ok {
		return elems
	}
	
	// 默认返回单质
	result[substance] = 1
	return result
}

// 检查两个物质是否能反应
func CanReact(substance1, substance2 string) bool {
	if products, ok := reactionDB[substance1]; ok {
		for _, product := range products {
			if product == substance2 {
				return true
			}
		}
	}
	
	if products, ok := reactionDB[substance2]; ok {
		for _, product := range products {
			if product == substance1 {
				return true
			}
		}
	}
	
	return false
}

// 获取能与指定物质反应的所有物质
func GetReactableSubstances(substance string) []string {
	if products, ok := reactionDB[substance]; ok {
		return products
	}
	return []string{}
}
