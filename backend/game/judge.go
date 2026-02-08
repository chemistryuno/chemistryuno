package game

import (
	"strings"
)

// SubstanceType 化学物质类型
type SubstanceType int

const (
	TypeUnknown SubstanceType = iota
	TypeMetal
	TypeNonMetal
	TypeAcid
	TypeBase
	TypeAcidicOxide
	TypeBasicOxide
	TypeSalt
	TypeInertGas
	TypeWater
)

// Substance 物质信息
type SubstanceInfo struct {
	Name string
	Type SubstanceType
	Tags []string // 例如 "soluble", "insoluble", "oxidizing", "active_metal"
}

// 获取物质分类信息
func getSubstanceInfo(name string) SubstanceInfo {
	// 简单的硬编码数据库
	db := map[string]SubstanceInfo{
		// 单质 - 金属
		"K":  {Name: "钾", Type: TypeMetal, Tags: []string{"active_metal", "metal_before_h"}},
		"Na": {Name: "钠", Type: TypeMetal, Tags: []string{"active_metal", "metal_before_h"}},
		"Ca": {Name: "钙", Type: TypeMetal, Tags: []string{"active_metal", "metal_before_h"}},
		"Mg": {Name: "镁", Type: TypeMetal, Tags: []string{"metal_before_h"}},
		"Al": {Name: "铝", Type: TypeMetal, Tags: []string{"metal_before_h"}},
		"Zn": {Name: "锌", Type: TypeMetal, Tags: []string{"metal_before_h"}},
		"Fe": {Name: "铁", Type: TypeMetal, Tags: []string{"metal_before_h"}},
		"Sn": {Name: "锡", Type: TypeMetal, Tags: []string{"metal_before_h"}},
		"Pb": {Name: "铅", Type: TypeMetal, Tags: []string{"metal_before_h"}},
		"Cu": {Name: "铜", Type: TypeMetal, Tags: []string{"metal_after_h"}},
		"Hg": {Name: "汞", Type: TypeMetal, Tags: []string{"metal_after_h"}},
		"Ag": {Name: "银", Type: TypeMetal, Tags: []string{"metal_after_h"}},
		"Pt": {Name: "铂", Type: TypeMetal, Tags: []string{"noble_metal"}},
		"Au": {Name: "金", Type: TypeMetal, Tags: []string{"noble_metal"}},

		// 单质 - 非金属
		"H2":  {Name: "氢气", Type: TypeNonMetal, Tags: []string{"reducing"}},
		"O2":  {Name: "氧气", Type: TypeNonMetal, Tags: []string{"oxidizing"}},
		"Cl2": {Name: "氯气", Type: TypeNonMetal, Tags: []string{"oxidizing"}},
		"Br2": {Name: "溴", Type: TypeNonMetal, Tags: []string{"oxidizing"}},
		"I2":  {Name: "碘", Type: TypeNonMetal, Tags: []string{"oxidizing"}},
		"C":   {Name: "碳", Type: TypeNonMetal, Tags: []string{"reducing"}},
		"S":   {Name: "硫", Type: TypeNonMetal, Tags: []string{"reducing"}},
		"P":   {Name: "磷", Type: TypeNonMetal, Tags: []string{"reducing"}},
		"N2":  {Name: "氮气", Type: TypeNonMetal, Tags: []string{"stable"}},

		// 水
		"H2O": {Name: "水", Type: TypeWater, Tags: []string{"solvent"}},

		// 惰性气体
		"He": {Name: "氦气", Type: TypeInertGas},
		"Ne": {Name: "氖气", Type: TypeInertGas},
		"Ar": {Name: "氩气", Type: TypeInertGas},
		"Kr": {Name: "氪气", Type: TypeInertGas},

		// 酸
		"HCl":   {Name: "盐酸", Type: TypeAcid, Tags: []string{"strong_acid"}},
		"H2SO4": {Name: "硫酸", Type: TypeAcid, Tags: []string{"strong_acid"}},
		"HNO3":  {Name: "硝酸", Type: TypeAcid, Tags: []string{"strong_acid", "oxidizing"}},
		"H2CO3": {Name: "碳酸", Type: TypeAcid, Tags: []string{"weak_acid", "instable"}},

		// 碱
		"NaOH":    {Name: "氢氧化钠", Type: TypeBase, Tags: []string{"strong_base", "soluble"}},
		"KOH":     {Name: "氢氧化钾", Type: TypeBase, Tags: []string{"strong_base", "soluble"}},
		"Ca(OH)2": {Name: "氢氧化钙", Type: TypeBase, Tags: []string{"strong_base", "slightly_soluble"}},
		"Ba(OH)2": {Name: "氢氧化钡", Type: TypeBase, Tags: []string{"strong_base", "soluble"}},
		"Mg(OH)2": {Name: "氢氧化镁", Type: TypeBase, Tags: []string{"weak_base", "insoluble"}},
		"Al(OH)3": {Name: "氢氧化铝", Type: TypeBase, Tags: []string{"amphoteric", "insoluble"}},
		"Cu(OH)2": {Name: "氢氧化铜", Type: TypeBase, Tags: []string{"weak_base", "insoluble"}},
		"Fe(OH)3": {Name: "氢氧化铁", Type: TypeBase, Tags: []string{"weak_base", "insoluble"}},

		// 氧化物
		"CO2":   {Name: "二氧化碳", Type: TypeAcidicOxide},
		"SO2":   {Name: "二氧化硫", Type: TypeAcidicOxide},
		"SO3":   {Name: "三氧化硫", Type: TypeAcidicOxide},
		"P2O5":  {Name: "五氧化二磷", Type: TypeAcidicOxide},
		"Na2O":  {Name: "氧化钠", Type: TypeBasicOxide},
		"K2O":   {Name: "氧化钾", Type: TypeBasicOxide},
		"CaO":   {Name: "氧化钙", Type: TypeBasicOxide},
		"MgO":   {Name: "氧化镁", Type: TypeBasicOxide},
		"CuO":   {Name: "氧化铜", Type: TypeBasicOxide},
		"Fe2O3": {Name: "氧化铁", Type: TypeBasicOxide},

		// 盐
		"NaCl":   {Name: "氯化钠", Type: TypeSalt, Tags: []string{"soluble"}},
		"Na2CO3": {Name: "碳酸钠", Type: TypeSalt, Tags: []string{"soluble"}},
		"Na2SO4": {Name: "硫酸钠", Type: TypeSalt, Tags: []string{"soluble"}},
		"NaNO3":  {Name: "硝酸钠", Type: TypeSalt, Tags: []string{"soluble"}},
		"KCl":    {Name: "氯化钾", Type: TypeSalt, Tags: []string{"soluble"}},
		"K2CO3":  {Name: "碳酸钾", Type: TypeSalt, Tags: []string{"soluble"}},
		"K2SO4":  {Name: "硫酸钾", Type: TypeSalt, Tags: []string{"soluble"}},
		"KNO3":   {Name: "硝酸钾", Type: TypeSalt, Tags: []string{"soluble"}},
		"CaCl2":  {Name: "氯化钙", Type: TypeSalt, Tags: []string{"soluble"}},
		"CaCO3":  {Name: "碳酸钙", Type: TypeSalt, Tags: []string{"insoluble"}},
		"BaCl2":  {Name: "氯化钡", Type: TypeSalt, Tags: []string{"soluble"}},
		"BaSO4":  {Name: "硫酸钡", Type: TypeSalt, Tags: []string{"insoluble"}},
		"CuSO4":  {Name: "硫酸铜", Type: TypeSalt, Tags: []string{"soluble"}},
		"AgNO3":  {Name: "硝酸银", Type: TypeSalt, Tags: []string{"soluble"}},
		"AgCl":   {Name: "氯化银", Type: TypeSalt, Tags: []string{"insoluble"}},
		"FeCl3":  {Name: "氯化铁", Type: TypeSalt, Tags: []string{"soluble"}},
		"FeCl2":  {Name: "氯化亚铁", Type: TypeSalt, Tags: []string{"soluble"}},
		"FeSO4":  {Name: "硫酸亚铁", Type: TypeSalt, Tags: []string{"soluble"}},
	}

	if info, ok := db[name]; ok {
		return info
	}

	// 动态类型检测助力
	_, anion := getIons(name)
	if anion != "" {
		// 判断是否是酸 (H开头的非括号结构通常是酸)
		if strings.HasPrefix(name, "H") {
			return SubstanceInfo{Name: "未知酸", Type: TypeAcid}
		}
		// 判断是否是碱 (含OH)
		if anion == "OH" {
			return SubstanceInfo{Name: "未知碱", Type: TypeBase, Tags: []string{"soluble"}}
		}
		// 否则视为盐
		return SubstanceInfo{Name: "未知盐", Type: TypeSalt, Tags: []string{"soluble"}}
	}

	// 处理简单的单质
	if len(name) <= 2 {
		// 识别常见的非金属单质原子形式
		nonMetals := map[string]bool{"H": true, "O": true, "N": true, "Cl": true, "Br": true, "I": true, "F": true, "S": true, "P": true, "C": true}
		if nonMetals[name] {
			return SubstanceInfo{Name: name, Type: TypeNonMetal}
		}
		return SubstanceInfo{Name: name, Type: TypeMetal}
	}

	return SubstanceInfo{Name: name, Type: TypeUnknown}
}

// JudgeReaction 判断两个物质是否能反应
func JudgeReaction(s1, s2 string) bool {
	info1 := getSubstanceInfo(s1)
	info2 := getSubstanceInfo(s2)

	// 惰性气体不反应
	if info1.Type == TypeInertGas || info2.Type == TypeInertGas {
		return false
	}

	// 1. 酸碱中和
	if (info1.Type == TypeAcid && info2.Type == TypeBase) || (info1.Type == TypeBase && info2.Type == TypeAcid) {
		return true
	}

	// 2. 酸 + 碱性氧化物
	if (info1.Type == TypeAcid && info2.Type == TypeBasicOxide) || (info1.Type == TypeBasicOxide && info2.Type == TypeAcid) {
		return true
	}

	// 3. 碱 + 酸性氧化物
	if (info1.Type == TypeBase && info2.Type == TypeAcidicOxide) || (info1.Type == TypeAcidicOxide && info2.Type == TypeBase) {
		return true
	}

	// 4. 酸 + 盐
	if (info1.Type == TypeAcid && info2.Type == TypeSalt) || (info1.Type == TypeSalt && info2.Type == TypeAcid) {
		acidFormula := s1
		saltFormula := s2
		if info2.Type == TypeAcid {
			acidFormula = s2
			saltFormula = s1
		}
		// 强酸制弱酸 (如 HCl + Na2CO3 -> NaCl + H2O + CO2)
		if strings.Contains(saltFormula, "CO3") {
			return true
		}
		// 生成沉淀 (如 H2SO4 + BaCl2)
		_, aAcid := getIons(acidFormula)
		cCation, _ := getIons(saltFormula)
		if !isSoluble(cCation, aAcid) {
			return true
		}
		// 特例：银盐与盐酸
		if strings.Contains(acidFormula, "HCl") && strings.Contains(saltFormula, "Ag") {
			return true
		}
	}

	// 5. 碱 + 盐
	if (info1.Type == TypeBase && info2.Type == TypeSalt) || (info1.Type == TypeSalt && info2.Type == TypeBase) {
		baseFormula := s1
		saltFormula := s2
		if info2.Type == TypeBase {
			baseFormula = s2
			saltFormula = s1
		}
		bInfo := getSubstanceInfo(baseFormula)
		sInfo := getSubstanceInfo(saltFormula)
		// 必须都是可溶的 (Ca(OH)2 微溶也算)
		if (hasTag(bInfo, "soluble") || bInfo.Name == "氢氧化钙") && hasTag(sInfo, "soluble") {
			cBase, aBase := getIons(baseFormula)
			cSalt, aSalt := getIons(saltFormula)
			// 交换离子看是否有沉淀生成 (cBase+aSalt 或 cSalt+aBase)
			if !isSoluble(cBase, aSalt) || !isSoluble(cSalt, aBase) {
				return true
			}
		}
	}

	// 6. 盐 + 盐
	if info1.Type == TypeSalt && info2.Type == TypeSalt {
		// 必须都是可溶的，且生成沉淀
		if hasTag(info1, "soluble") && hasTag(info2, "soluble") {
			c1, a1 := getIons(s1)
			c2, a2 := getIons(s2)
			if !isSoluble(c1, a2) || !isSoluble(c2, a1) {
				return true
			}
		}
	}

	// 7. 金属 + 酸
	if (info1.Type == TypeMetal && info2.Type == TypeAcid) || (info2.Type == TypeMetal && info1.Type == TypeAcid) {
		metal := info1
		acid := info2
		if info2.Type == TypeMetal {
			metal = info2
			acid = info1
		}
		// 氢前金属可以与非氧化性酸反应
		if hasTag(metal, "metal_before_h") || hasTag(metal, "active_metal") {
			return true
		}
		// 铜银等可以与氧化性酸反应
		if hasTag(acid, "oxidizing") && !hasTag(metal, "noble_metal") {
			return true
		}
	}

	// 8. 金属 + 盐
	if (info1.Type == TypeMetal && info2.Type == TypeSalt) || (info2.Type == TypeMetal && info1.Type == TypeSalt) {
		metal := info1
		salt := info2
		if info2.Type == TypeMetal {
			metal = info2
			salt = info1
		}
		// 排在前面的金属可以置换出排在后面的金属（简化逻辑：如果是活泼金属置换不活泼金属盐）
		if (hasTag(metal, "metal_before_h") && (strings.Contains(salt.Name, "铜") || strings.Contains(salt.Name, "银"))) ||
			(strings.Contains(s1+s2, "Fe") && (strings.Contains(s1+s2, "CuSO4") || strings.Contains(s1+s2, "AgNO3"))) {
			return true
		}
	}

	// 9. 氧化还原 (燃烧/还原)
	if s1 == "O2" || s2 == "O2" {
		other := info1
		if s1 == "O2" {
			other = info2
		}
		// 金属氧化 (除金铂外)
		if other.Type == TypeMetal && !hasTag(other, "noble_metal") {
			return true
		}
		// 非金属燃烧
		if other.Type == TypeNonMetal && !hasTag(other, "stable") {
			return true
		}
	}

	if info1.Type == TypeNonMetal && info2.Type == TypeNonMetal {
		if (hasTag(info1, "oxidizing") && hasTag(info2, "reducing")) || (hasTag(info1, "reducing") && hasTag(info2, "oxidizing")) {
			return true
		}
	}
	if (info1.Type == TypeNonMetal && info2.Type == TypeBasicOxide) || (info2.Type == TypeNonMetal && info1.Type == TypeBasicOxide) {
		nonMetal := info1
		if info2.Type == TypeNonMetal {
			nonMetal = info2
		}
		if hasTag(nonMetal, "reducing") { // 如 C/H2 还原 CuO
			return true
		}
	}

	// 10. 与水反应
	if info1.Type == TypeWater || info2.Type == TypeWater {
		other := info1
		if info1.Type == TypeWater {
			other = info2
		}
		// 活泼金属
		if hasTag(other, "active_metal") {
			return true
		}
		// 部分氧化物
		if other.Type == TypeBasicOxide || other.Type == TypeAcidicOxide {
			// 常见的氧化钙、二氧化碳等
			formula := s1 + s2
			if strings.Contains(formula, "CaO") || strings.Contains(formula, "CO2") || strings.Contains(formula, "SO2") || strings.Contains(formula, "SO3") || strings.Contains(formula, "Na2O") {
				return true
			}
		}
	}

	return false
}

func hasTag(info SubstanceInfo, tag string) bool {
	for _, t := range info.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// 简单的溶解性判断
func isSoluble(cation, anion string) bool {
	// 钾钠铵硝全都溶
	if cation == "K" || cation == "Na" || cation == "NH4" || anion == "NO3" {
		return true
	}
	// 盐酸盐：除银、亚汞外都溶 (这里简化处理银)
	if anion == "Cl" {
		return cation != "Ag"
	}
	// 硫酸盐：除钡、铅不溶，钙、银微溶
	if anion == "SO4" {
		return cation != "Ba" && cation != "Pb" && cation != "Ca"
	}
	// 碳酸盐、磷酸盐：只溶钾钠铵
	if anion == "CO3" {
		return cation == "K" || cation == "Na" || cation == "NH4"
	}
	// 碱：只溶钾钠钡，钙微溶
	if anion == "OH" {
		return cation == "K" || cation == "Na" || cation == "Ba" || cation == "Ca"
	}
	return true
}

// 提取化学式中的离子
func getIons(formula string) (cation, anion string) {
	// 常见酸根
	anions := []string{"SO4", "CO3", "NO3", "OH", "Cl"}
	for _, a := range anions {
		if strings.Contains(formula, a) {
			anion = a
			// 提取阳离子部分
			c := strings.Replace(formula, a, "", 1)
			// 去除数字和括号
			c = strings.Map(func(r rune) rune {
				if (r >= '0' && r <= '9') || r == '(' || r == ')' {
					return -1
				}
				return r
			}, c)
			cation = c
			return
		}
	}
	// 默认处理
	return formula, ""
}
