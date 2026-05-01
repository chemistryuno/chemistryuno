package game

import (
	"chemistryuno/backend/database"
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
	// 1. 尝试从硬编码库获取
	if info, ok := substanceDB[name]; ok {
		return info
	}

	// 2. 尝试从数据库获取基本信息
	var sub database.Substance
	err := database.DB.Where("formula = ? OR name = ?", name, name).First(&sub).Error
	if err == nil {
		// 数据库中现在没有 Category 和 Tags 字段了，只能返回基础信息
		return SubstanceInfo{
			Name: sub.Name,
			Type: TypeUnknown, // 默认为未知，靠后续动态检测
			Tags: []string{},
		}
	}

	// 3. 动态类型检测助力
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

func mapCategoryToType(cat string) SubstanceType {
	switch cat {
	case "metal":
		return TypeMetal
	case "nonmetal":
		return TypeNonMetal
	case "acid":
		return TypeAcid
	case "base":
		return TypeBase
	case "acidic_oxide":
		return TypeAcidicOxide
	case "basic_oxide":
		return TypeBasicOxide
	default:
		return TypeUnknown
	}
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

var substanceDB = map[string]SubstanceInfo{
	// 基础单质和常见物质
	"H2O":        {Name: "水", Type: TypeWater, Tags: []string{"solvent"}},
	"CO2":        {Name: "二氧化碳", Type: TypeAcidicOxide},
	"HCl":        {Name: "盐酸", Type: TypeAcid, Tags: []string{"strong_acid"}},
	"NaOH":       {Name: "氢氧化钠", Type: TypeBase, Tags: []string{"strong_base", "soluble"}},
	"NaCl":       {Name: "氯化钠", Type: TypeSalt, Tags: []string{"soluble"}},
	"H2SO4":      {Name: "硫酸", Type: TypeAcid, Tags: []string{"strong_acid"}},
	"O2":         {Name: "氧气", Type: TypeNonMetal, Tags: []string{"oxidizing"}},
	"H2":         {Name: "氢气", Type: TypeNonMetal, Tags: []string{"reducing"}},
	"Fe":         {Name: "铁", Type: TypeMetal, Tags: []string{"metal_before_h"}},
	"CuSO4":      {Name: "硫酸铜", Type: TypeSalt, Tags: []string{"soluble"}},
	"Cu":         {Name: "铜", Type: TypeMetal, Tags: []string{"metal_after_h"}},
	"Zn":         {Name: "锌", Type: TypeMetal, Tags: []string{"metal_before_h"}},
	"CO":         {Name: "一氧化碳", Type: TypeNonMetal, Tags: []string{"reducing"}},
	"CaCO3":      {Name: "碳酸钙", Type: TypeSalt, Tags: []string{"insoluble"}},
	"CaO":        {Name: "氧化钙", Type: TypeBasicOxide},
	"Ca(OH)2":    {Name: "氢氧化钙", Type: TypeBase, Tags: []string{"strong_base", "slightly_soluble"}},
	"NH3":        {Name: "氨气", Type: TypeBase, Tags: []string{"weak_base", "soluble"}},
	"HNO3":       {Name: "硝酸", Type: TypeAcid, Tags: []string{"strong_acid", "oxidizing"}},
	"AgNO3":      {Name: "硝酸银", Type: TypeSalt, Tags: []string{"soluble"}},
	"AgCl":       {Name: "氯化银", Type: TypeSalt, Tags: []string{"insoluble"}},
	"MgO":        {Name: "氧化镁", Type: TypeBasicOxide},
	"Mg":         {Name: "镁", Type: TypeMetal, Tags: []string{"metal_before_h"}},
	"Al":         {Name: "铝", Type: TypeMetal, Tags: []string{"metal_before_h"}},
	"Al2O3":      {Name: "氧化铝", Type: TypeBasicOxide},
	"Fe2O3":      {Name: "氧化铁", Type: TypeBasicOxide},
	"Fe3O4":      {Name: "四氧化三铁", Type: TypeBasicOxide},
	"Na2CO3":     {Name: "碳酸钠", Type: TypeSalt, Tags: []string{"soluble"}},
	"NaHCO3":     {Name: "碳酸氢钠", Type: TypeSalt, Tags: []string{"soluble"}},
	"KClO3":      {Name: "氯酸钾", Type: TypeSalt, Tags: []string{"soluble"}},
	"KCl":        {Name: "氯化钾", Type: TypeSalt, Tags: []string{"soluble"}},
	"H2O2":       {Name: "过氧化氢", Type: TypeNonMetal, Tags: []string{"oxidizing"}},
	"SO2":        {Name: "二氧化硫", Type: TypeAcidicOxide},
	"BaCl2":      {Name: "氯化钡", Type: TypeSalt, Tags: []string{"soluble"}},
	"BaSO4":      {Name: "硫酸钡", Type: TypeSalt, Tags: []string{"insoluble"}},
	"C":          {Name: "碳", Type: TypeNonMetal, Tags: []string{"reducing"}},
	"S":          {Name: "硫", Type: TypeNonMetal, Tags: []string{"reducing"}},
	"P":          {Name: "磷", Type: TypeNonMetal, Tags: []string{"reducing"}},
	"P2O5":       {Name: "五氧化二磷", Type: TypeAcidicOxide},
	"CuO":        {Name: "氧化铜", Type: TypeBasicOxide},
	"FeCl3":      {Name: "氯化铁", Type: TypeSalt, Tags: []string{"soluble"}},
	"FeSO4":      {Name: "硫酸亚铁", Type: TypeSalt, Tags: []string{"soluble"}},
	"KOH":        {Name: "氢氧化钾", Type: TypeBase, Tags: []string{"strong_base", "soluble"}},
	"MgCl2":      {Name: "氯化镁", Type: TypeSalt, Tags: []string{"soluble"}},
	"CaCl2":      {Name: "氯化钙", Type: TypeSalt, Tags: []string{"soluble"}},
	"NH4Cl":      {Name: "氯化铵", Type: TypeSalt, Tags: []string{"soluble"}},
	"Na2SO4":     {Name: "硫酸钠", Type: TypeSalt, Tags: []string{"soluble"}},
	"NO":         {Name: "一氧化氮", Type: TypeNonMetal, Tags: []string{"reducing"}},
	"NO2":        {Name: "二氧化氮", Type: TypeNonMetal, Tags: []string{"oxidizing"}},
	"N2":         {Name: "氮气", Type: TypeNonMetal, Tags: []string{"stable"}},
	"Hg":         {Name: "汞", Type: TypeMetal, Tags: []string{"metal_after_h"}},
	"HgO":        {Name: "氧化汞", Type: TypeBasicOxide},
	"Na2O":       {Name: "氧化钠", Type: TypeBasicOxide},
	"K2O":        {Name: "氧化钾", Type: TypeBasicOxide},
	"SO3":        {Name: "三氧化硫", Type: TypeAcidicOxide},
	"BaO":        {Name: "氧化钡", Type: TypeBasicOxide},
	"CuCl2":      {Name: "氯化铜", Type: TypeSalt, Tags: []string{"soluble"}},
	"Cu(OH)2":    {Name: "氢氧化铜", Type: TypeBase, Tags: []string{"weak_base", "insoluble"}},
	"Fe(OH)3":    {Name: "氢氧化铁", Type: TypeBase, Tags: []string{"weak_base", "insoluble"}},
	"NH4NO3":     {Name: "硝酸铵", Type: TypeSalt, Tags: []string{"soluble"}},
	"ZnCl2":      {Name: "氯化锌", Type: TypeSalt, Tags: []string{"soluble"}},
	"ZnSO4":      {Name: "硫酸锌", Type: TypeSalt, Tags: []string{"soluble"}},
	"AlCl3":      {Name: "氯化铝", Type: TypeSalt, Tags: []string{"soluble"}},
	"Ag":         {Name: "银", Type: TypeMetal, Tags: []string{"metal_after_h"}},
	"H3PO4":      {Name: "磷酸", Type: TypeAcid, Tags: []string{"weak_acid"}},
	"Na3PO4":     {Name: "磷酸钠", Type: TypeSalt, Tags: []string{"soluble"}},
	"Cl2":        {Name: "氯气", Type: TypeNonMetal, Tags: []string{"oxidizing"}},
	"Br2":        {Name: "溴单质", Type: TypeNonMetal, Tags: []string{"oxidizing"}},
	"I2":         {Name: "碘单质", Type: TypeNonMetal, Tags: []string{"oxidizing"}},
	"HI":         {Name: "碘化氢", Type: TypeAcid, Tags: []string{"strong_acid"}},
	"HBr":        {Name: "溴化氢", Type: TypeAcid, Tags: []string{"strong_acid"}},
	"K2SO4":      {Name: "硫酸钾", Type: TypeSalt, Tags: []string{"soluble"}},
	"MgSO4":      {Name: "硫酸镁", Type: TypeSalt, Tags: []string{"soluble"}},
	"CaSO4":      {Name: "硫酸钙", Type: TypeSalt, Tags: []string{"slightly_soluble"}},
	"Na2O2":      {Name: "过氧化钠", Type: TypeBasicOxide, Tags: []string{"oxidizing"}},
	"Na":         {Name: "钠", Type: TypeMetal, Tags: []string{"active_metal", "metal_before_h"}},
	"K":          {Name: "钾", Type: TypeMetal, Tags: []string{"active_metal", "metal_before_h"}},
	"Ca":         {Name: "钙", Type: TypeMetal, Tags: []string{"active_metal", "metal_before_h"}},
	"NaAlO2":     {Name: "偏铝酸钠", Type: TypeSalt, Tags: []string{"soluble"}},
	"K2CO3":      {Name: "碳酸钾", Type: TypeSalt, Tags: []string{"soluble"}},
	"NaH":        {Name: "氢化钠", Type: TypeSalt, Tags: []string{"reducing"}},
	"KH":         {Name: "氢化钾", Type: TypeSalt, Tags: []string{"reducing"}},
	"MgH2":       {Name: "氢化镁", Type: TypeSalt, Tags: []string{"reducing"}},
	"BaH2":       {Name: "氢化钡", Type: TypeSalt, Tags: []string{"reducing"}},
	"KO2":        {Name: "超氧化钾", Type: TypeBasicOxide, Tags: []string{"oxidizing"}},
	"CaH2":       {Name: "氢化钙", Type: TypeSalt, Tags: []string{"reducing"}},
	"Na2SO3":     {Name: "亚硫酸钠", Type: TypeSalt, Tags: []string{"soluble"}},
	"K2SO3":      {Name: "亚硫酸钾", Type: TypeSalt, Tags: []string{"soluble"}},
	"NaHSO4":     {Name: "硫酸氢钠", Type: TypeSalt, Tags: []string{"soluble"}},
	"NaHSO3":     {Name: "亚硫酸氢钠", Type: TypeSalt, Tags: []string{"soluble"}},
	"CaSO3":      {Name: "亚硫酸钙", Type: TypeSalt, Tags: []string{"insoluble"}},
	"BaCO3":      {Name: "碳酸钡", Type: TypeSalt, Tags: []string{"insoluble"}},
	"NaBr":       {Name: "溴化钠", Type: TypeSalt, Tags: []string{"soluble"}},
	"NaI":        {Name: "碘化钠", Type: TypeSalt, Tags: []string{"soluble"}},
	"KBr":        {Name: "溴化钾", Type: TypeSalt, Tags: []string{"soluble"}},
	"AgBr":       {Name: "溴化银", Type: TypeSalt, Tags: []string{"insoluble"}},
	"AgI":        {Name: "碘化银", Type: TypeSalt, Tags: []string{"insoluble"}},
	"F2":         {Name: "氟气", Type: TypeNonMetal, Tags: []string{"oxidizing"}},
	"HF":         {Name: "氢氟酸", Type: TypeAcid, Tags: []string{"weak_acid"}},
	"NaF":        {Name: "氟化钠", Type: TypeSalt, Tags: []string{"soluble"}},
	"CaF2":       {Name: "氟化钙", Type: TypeSalt, Tags: []string{"insoluble"}},
	"HClO":       {Name: "次氯酸", Type: TypeAcid, Tags: []string{"weak_acid", "oxidizing"}},
	"NaClO":      {Name: "次氯酸钠", Type: TypeSalt, Tags: []string{"soluble", "oxidizing"}},
	"H2S":        {Name: "硫化氢", Type: TypeAcid, Tags: []string{"weak_acid"}},
	"Na2S":       {Name: "硫化钠", Type: TypeSalt, Tags: []string{"soluble"}},
	"FeS":        {Name: "硫化亚铁", Type: TypeSalt, Tags: []string{"insoluble"}},
	"CuS":        {Name: "硫化铜", Type: TypeSalt, Tags: []string{"insoluble"}},
	"ZnS":        {Name: "硫化锌", Type: TypeSalt, Tags: []string{"insoluble"}},
	"Ag2S":       {Name: "硫化银", Type: TypeSalt, Tags: []string{"insoluble"}},
	"H2SO3":      {Name: "亚硫酸", Type: TypeAcid, Tags: []string{"weak_acid"}},
	"BaSO3":      {Name: "亚硫酸钡", Type: TypeSalt, Tags: []string{"insoluble"}},
	"MgBr2":      {Name: "溴化镁", Type: TypeSalt, Tags: []string{"soluble"}},
	"MgI2":       {Name: "碘化镁", Type: TypeSalt, Tags: []string{"soluble"}},
	"AlBr3":      {Name: "溴化铝", Type: TypeSalt, Tags: []string{"soluble"}},
	"AlI3":       {Name: "碘化铝", Type: TypeSalt, Tags: []string{"soluble"}},
	"ZnBr2":      {Name: "溴化锌", Type: TypeSalt, Tags: []string{"soluble"}},
	"ZnI2":       {Name: "碘化锌", Type: TypeSalt, Tags: []string{"soluble"}},
	"CuBr2":      {Name: "溴化铜", Type: TypeSalt, Tags: []string{"soluble"}},
	"FeBr3":      {Name: "溴化铁", Type: TypeSalt, Tags: []string{"soluble"}},
	"FeI2":       {Name: "碘化亚铁", Type: TypeSalt, Tags: []string{"soluble"}},
	"CaBr2":      {Name: "溴化钙", Type: TypeSalt, Tags: []string{"soluble"}},
	"CaI2":       {Name: "碘化钙", Type: TypeSalt, Tags: []string{"soluble"}},
	"KF":         {Name: "氟化钾", Type: TypeSalt, Tags: []string{"soluble"}},
	"BaF2":       {Name: "氟化钡", Type: TypeSalt, Tags: []string{"insoluble"}},
	"MgF2":       {Name: "氟化镁", Type: TypeSalt, Tags: []string{"insoluble"}},
	"AlF3":       {Name: "氟化铝", Type: TypeSalt, Tags: []string{"insoluble"}},
	"FeF3":       {Name: "氟化铁", Type: TypeSalt, Tags: []string{"insoluble"}},
	"CuF2":       {Name: "氟化铜", Type: TypeSalt, Tags: []string{"insoluble"}},
	"AgF":        {Name: "氟化银", Type: TypeSalt, Tags: []string{"soluble"}},
	"HgF2":       {Name: "氟化汞", Type: TypeSalt, Tags: []string{"insoluble"}},
	"ZnF2":       {Name: "氟化锌", Type: TypeSalt, Tags: []string{"insoluble"}},
	"SiO2":       {Name: "二氧化硅", Type: TypeAcidicOxide},
	"SiF4":       {Name: "四氟化硅", Type: TypeNonMetal},
	"ZnO":        {Name: "氧化锌", Type: TypeBasicOxide},
	"Cu2O":       {Name: "氧化亚铜", Type: TypeBasicOxide},
	"FeO":        {Name: "氧化亚铁", Type: TypeBasicOxide},
	"Ag2O":       {Name: "氧化银", Type: TypeBasicOxide},
	"N2O":        {Name: "一氧化二氮", Type: TypeNonMetal},
	"Cl2O7":      {Name: "七氧化二氯", Type: TypeAcidicOxide, Tags: []string{"oxidizing"}},
	"K2S":        {Name: "硫化钾", Type: TypeSalt, Tags: []string{"soluble"}},
	"MgS":        {Name: "硫化镁", Type: TypeSalt, Tags: []string{"soluble"}},
	"CaS":        {Name: "硫化钙", Type: TypeSalt, Tags: []string{"soluble"}},
	"BaS":        {Name: "硫化钡", Type: TypeSalt, Tags: []string{"soluble"}},
	"MgSO3":      {Name: "亚硫酸镁", Type: TypeSalt, Tags: []string{"insoluble"}},
	"Al2S3":      {Name: "硫化铝", Type: TypeSalt, Tags: []string{"insoluble"}},
	"Al2(SO3)3":  {Name: "亚硫酸铝", Type: TypeSalt, Tags: []string{"insoluble"}},
	"FeSO3":      {Name: "亚硫酸亚铁", Type: TypeSalt, Tags: []string{"insoluble"}},
	"KI":         {Name: "碘化钾", Type: TypeSalt, Tags: []string{"soluble"}},
	"BaBr2":      {Name: "溴化钡", Type: TypeSalt, Tags: []string{"soluble"}},
	"BaI2":       {Name: "碘化钡", Type: TypeSalt, Tags: []string{"soluble"}},
	"FeCl2":      {Name: "氯化亚铁", Type: TypeSalt, Tags: []string{"soluble"}},
	"Al2(SO4)3":  {Name: "硫酸铝", Type: TypeSalt, Tags: []string{"soluble"}},
	"Ba(OH)2":    {Name: "氢氧化钡", Type: TypeBase, Tags: []string{"strong_base", "soluble"}},
	"Ba":         {Name: "钡", Type: TypeMetal, Tags: []string{"active_metal", "metal_before_h"}},
	"KHCO3":      {Name: "碳酸氢钾", Type: TypeSalt, Tags: []string{"soluble"}},
	"Ca(HCO3)2":  {Name: "碳酸氢钙", Type: TypeSalt, Tags: []string{"soluble"}},
	"Mg(OH)2":    {Name: "氢氧化镁", Type: TypeBase, Tags: []string{"weak_base", "insoluble"}},
	"PCl3":       {Name: "三氯化磷", Type: TypeNonMetal},
	"P2S5":       {Name: "五硫化二磷", Type: TypeNonMetal},
	"PF5":        {Name: "五氟化磷", Type: TypeNonMetal},
	"CF4":        {Name: "四氟化碳", Type: TypeNonMetal},
	"SF6":        {Name: "六氟化硫", Type: TypeNonMetal},
	"Na2S2O3":    {Name: "硫代硫酸钠", Type: TypeSalt, Tags: []string{"soluble"}},
	"K2S2O3":     {Name: "硫代硫酸钾", Type: TypeSalt, Tags: []string{"soluble"}},
	"Hg2O":       {Name: "氧化亚汞", Type: TypeBasicOxide},
	"HgCl2":      {Name: "氯化汞", Type: TypeSalt, Tags: []string{"soluble"}},
	"HgBr2":      {Name: "溴化汞", Type: TypeSalt, Tags: []string{"soluble"}},
	"HgI2":       {Name: "碘化汞", Type: TypeSalt, Tags: []string{"insoluble"}},
	"Hg(NO3)2":   {Name: "硝酸汞", Type: TypeSalt, Tags: []string{"soluble"}},
	"Al(OH)3":    {Name: "氢氧化铝", Type: TypeBase, Tags: []string{"amphoteric", "insoluble"}},
	"Zn(OH)2":    {Name: "氢氧化锌", Type: TypeBase, Tags: []string{"amphoteric", "insoluble"}},
	"KNO3":       {Name: "硝酸钾", Type: TypeSalt, Tags: []string{"soluble"}},
	"Si":         {Name: "硅", Type: TypeNonMetal},
	"Na2SiO3":    {Name: "硅酸钠", Type: TypeSalt, Tags: []string{"soluble"}},
	"CaSiO3":     {Name: "硅酸钙", Type: TypeSalt, Tags: []string{"insoluble"}},
	"H2SiO3":     {Name: "硅酸", Type: TypeAcid, Tags: []string{"weak_acid", "insoluble"}},
	"SiCl4":      {Name: "四氯化硅", Type: TypeNonMetal},
	"Mn":         {Name: "锰", Type: TypeMetal, Tags: []string{"metal_before_h"}},
	"MnO":        {Name: "一氧化锰", Type: TypeBasicOxide},
	"MnO2":       {Name: "二氧化锰", Type: TypeBasicOxide, Tags: []string{"catalyst"}},
	"MnCl2":      {Name: "氯化锰", Type: TypeSalt, Tags: []string{"soluble"}},
	"MnSO4":      {Name: "硫酸锰", Type: TypeSalt, Tags: []string{"soluble"}},
	"MnCO3":      {Name: "碳酸锰", Type: TypeSalt, Tags: []string{"insoluble"}},
	"Mn(OH)2":    {Name: "氢氧化锰", Type: TypeBase, Tags: []string{"weak_base", "insoluble"}},
	"MnS":        {Name: "硫化锰", Type: TypeSalt, Tags: []string{"insoluble"}},
	"KMnO4":      {Name: "高锰酸钾", Type: TypeSalt, Tags: []string{"soluble", "oxidizing"}},
	"Mn(NO3)2":   {Name: "硝酸锰", Type: TypeSalt, Tags: []string{"soluble"}},
	"MnBr2":      {Name: "溴化锰", Type: TypeSalt, Tags: []string{"soluble"}},
	"MnI2":       {Name: "碘化锰", Type: TypeSalt, Tags: []string{"soluble"}},
	"MnF2":       {Name: "氟化锰", Type: TypeSalt, Tags: []string{"insoluble"}},
	"H2SiF6":     {Name: "六氟硅酸", Type: TypeAcid, Tags: []string{"strong_acid"}},
	"H4SiO4":     {Name: "原硅酸", Type: TypeAcid, Tags: []string{"weak_acid"}},
	"PCl5":       {Name: "五氯化磷", Type: TypeNonMetal},
	"H3PO3":      {Name: "亚磷酸", Type: TypeAcid, Tags: []string{"weak_acid"}},
	"PH4Cl":      {Name: "氯化铵磷", Type: TypeSalt, Tags: []string{"soluble"}},
	"PH3":        {Name: "磷化氢", Type: TypeNonMetal, Tags: []string{"toxic"}},
	"Ca(H2PO4)2": {Name: "磷酸二氢钙", Type: TypeSalt, Tags: []string{"soluble"}},
	"Ca3(PO4)2":  {Name: "磷酸钙", Type: TypeSalt, Tags: []string{"insoluble"}},
	"P2O3":       {Name: "三氧化二磷", Type: TypeAcidicOxide},
	"Ag3PO4":     {Name: "磷酸银", Type: TypeSalt, Tags: []string{"insoluble"}},
	"SiH4":       {Name: "硅化氢", Type: TypeNonMetal},
	"He":         {Name: "氦气", Type: TypeInertGas},
	"Ne":         {Name: "氖气", Type: TypeInertGas},
	"Ar":         {Name: "氩气", Type: TypeInertGas},
	"Kr":         {Name: "氪气", Type: TypeInertGas},
	"Xe":         {Name: "氙气", Type: TypeInertGas},
	"Rn":         {Name: "氡气", Type: TypeInertGas},
}
