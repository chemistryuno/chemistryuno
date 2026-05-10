// --- 数据定义 ---
const CHEMIST_NAMES = ["门捷列夫", "拉瓦锡", "波义耳", "道尔顿", "居里夫人", "诺贝尔", "范霍夫", "阿伦尼乌斯", "勒沙特列", "鲍林", "舍勒", "海洛夫斯基", "维尔纳", "哈伯", "费舍尔", "普利斯特里", "卡文迪许", "侯德榜", "徐寿", "庄长恭"];
const formatFormula = (f) => f.replace(/(\d+)/g, '<sub>$1</sub>');

function showMessageModal(title, content) {
    document.getElementById('message-modal-title').textContent = title;
    document.getElementById('message-modal-content').innerHTML = content;
    document.getElementById('message-modal').style.display = 'flex';
}

const EQUATIONS_LIB = {
    'H2+O2': '2H₂ + O₂ = 2H₂O', 'C+O2': 'C + O₂ = CO₂', 'S+O2': 'S + O₂ = SO₂', 'P+O2': '4P + 5O₂ = 2P₂O₅',
    'Fe+O2': '3Fe + 2O₂ = Fe₃O₄', 'Mg+O2': '2Mg + O₂ = 2MgO', 'CO+O2': '2CO + O₂ = 2CO₂',
    'HCl+NaOH': 'HCl + NaOH = NaCl + H₂O', 'H2SO4+NaOH': 'H₂SO₄ + 2NaOH = Na₂SO₄ + 2H₂O', 'HNO3+NaOH': 'HNO₃ + NaOH = NaNO₃ + H₂O',
    'HCl+Ca(OH)2': '2HCl + Ca(OH)₂ = CaCl₂ + 2H₂O', 'H2SO4+BaCl2': 'H₂SO₄ + BaCl₂ = BaSO₄↓ + 2HCl', 'HCl+AgNO3': 'HCl + AgNO₃ = AgCl↓ + HNO₃',
    'NaCl+AgNO3': 'NaCl + AgNO₃ = AgCl↓ + NaNO₃', 'Fe+HCl': 'Fe + 2HCl = FeCl₂ + H₂↑', 'Zn+HCl': 'Zn + 2HCl = ZnCl₂ + H₂↑',
    'Mg+HCl': 'Mg + 2HCl = MgCl₂ + H₂↑', 'Al+HCl': '2Al + 6HCl = 2AlCl₃ + 3H₂↑', 'Fe+CuSO4': 'Fe + CuSO₄ = FeSO₄ + Cu',
    'Zn+CuSO4': 'Zn + CuSO₄ = ZnSO₄ + Cu', 'CO2+Ca(OH)2': 'CO₂ + Ca(OH)₂ = CaCO₃↓ + H₂O', 'CO2+NaOH': 'CO₂ + 2NaOH = Na₂CO₃ + H₂O',
    'CO2+H2O': 'CO₂ + H₂O = H₂CO₃', 'CaO+H2O': 'CaO + H₂O = Ca(OH)₂', 'MgO+H2O': 'MgO + H₂O = Mg(OH)₂',
    'CaCO3+HCl': 'CaCO₃ + 2HCl = CaCl₂ + H₂O + CO₂↑', 'Na2CO3+HCl': 'Na₂CO₃ + 2HCl = 2NaCl + H₂O + CO₂↑', 'NaHCO3+HCl': 'NaHCO₃ + HCl = NaCl + H₂O + CO₂↑',
    'NH3+HCl': 'NH₃ + HCl = NH₄Cl', 'Cu+AgNO3': 'Cu + 2AgNO₃ = Cu(NO₃)₂ + 2Ag', 'CuO+H2': 'CuO + H₂ = Cu + H₂O',
    'Fe2O3+CO': 'Fe₂O₃ + 3CO = 2Fe + 3CO₂', 'CuO+C': '2CuO + C = 2Cu + CO₂↑', 'BaCl2+Na2SO4': 'BaCl₂ + Na₂SO₄ = BaSO₄↓ + 2NaCl',
    'MgCl2+NaOH': 'MgCl₂ + 2NaOH = Mg(OH)₂↓ + 2NaCl', 'N2+H2': 'N₂ + 3H₂ = 2NH₃', 'NO+O2': '2NO + O₂ = 2NO₂',
    'NO2+H2O': '3NO₂ + H₂O = 2HNO₃ + NO', 'Na2O+H2O': 'Na₂O + H₂O = 2NaOH', 'K2O+H2O': 'K₂O + H₂O = 2KOH',
    'SO3+H2O': 'SO₃ + H₂O = H₂SO₄', 'BaO+H2O': 'BaO + H₂O = Ba(OH)₂', 'CuCl2+NaOH': 'CuCl₂ + 2NaOH = Cu(OH)₂↓ + 2NaCl',
    'FeCl3+NaOH': 'FeCl₃ + 3NaOH = Fe(OH)₃↓ + 3NaCl', 'AlCl3+NaOH': 'AlCl₃ + 3NaOH = Al(OH)₃↓ + 3NaCl',
    'Cu(OH)2+HCl': 'Cu(OH)₂ + 2HCl = CuCl₂ + 2H₂O', 'Fe(OH)3+H2SO4': '2Fe(OH)₃ + 3H₂SO₄ = Fe₂(SO₄)₃ + 6H₂O',
    'AgNO3+BaCl2': '2AgNO₃ + BaCl₂ = 2AgCl↓ + Ba(NO₃)₂', 'AgNO3+Cu': '2AgNO₃ + Cu = Cu(NO₃)₂ + 2Ag',
    'Zn+AgNO3': 'Zn + 2AgNO₃ = Zn(NO₃)₂ + 2Ag', 'Fe+AgNO3': 'Fe + 2AgNO₃ = Fe(NO₃)₂ + 2Ag',
    'NH4Cl+NaOH': 'NH₄Cl + NaOH = NaCl + NH₃↑ + H₂O', 'Ca(OH)2+Na2CO3': 'Ca(OH)₂ + Na₂CO₃ = CaCO₃↓ + 2NaOH',
    'Ba(OH)2+Na2SO4': 'Ba(OH)₂ + Na₂SO₄ = BaSO₄↓ + 2NaOH', 'Na+H2O': '2Na + 2H₂O = 2NaOH + H₂↑',
    'K+H2O': '2K + 2H₂O = 2KOH + H₂↑', 'Ca+H2O': 'Ca + 2H₂O = Ca(OH)₂ + H₂↑', 'Na+Cl2': '2Na + Cl₂ = 2NaCl',
    'Mg+Cl2': 'Mg + Cl₂ = MgCl₂', 'Fe+Cl2': '2Fe + 3Cl₂ = 2FeCl₃', 'Al+NaOH': '2Al + 2NaOH + 2H₂O = 2NaAlO₂ + 3H₂↑',
    'HgO+C': '2HgO + C = 2Hg + CO₂↑', 'Hg+O2': '2Hg + O₂ = 2HgO',
    'NaH+H2O': 'NaH + H₂O = NaOH + H₂↑', 'CaH2+H2O': 'CaH₂ + 2H₂O = Ca(OH)₂ + 2H₂↑',
    'Na2O2+H2O': '2Na₂O₂ + 2H₂O = 4NaOH + O₂↑', 'Na2O2+CO2': '2Na₂O₂ + 2CO₂ = 2Na₂CO₃ + O₂',
    'Al+Fe2O3': '2Al + Fe₂O₃ = Al₂O₃ + 2Fe', 'F2+H2O': '2F₂ + 2H₂O = 4HF + O₂',
    'H2S+O2': '2H₂S + 3O₂ = 2SO₂ + 2H₂O', 'H2S+CuSO4': 'H₂S + CuSO₄ = CuS↓ + H₂SO₄',
    'H2S+AgNO3': 'H₂S + 2AgNO₃ = Ag₂S↓ + 2HNO₃', 'SO2+NaOH': 'SO₂ + 2NaOH = Na₂SO₃ + H₂O',
    'SO2+Ca(OH)2': 'SO₂ + Ca(OH)₂ = CaSO₃↓ + H₂O', 'SO3+NaOH': 'SO₃ + 2NaOH = Na₂SO₄ + H₂O',
    'SO3+Ca(OH)2': 'SO₃ + Ca(OH)₂ = CaSO₄ + H₂O', 'P2O5+NaOH': 'P₂O₅ + 6NaOH = 2Na₃PO₄ + 3H₂O',
    'P2O5+Ca(OH)2': 'P₂O₅ + 3Ca(OH)₂ = Ca₃(PO₄)₂↓ + 3H₂O', 'CuO+H2SO4': 'CuO + H₂SO₄ = CuSO₄ + H₂O',
    'CuO+HCl': 'CuO + 2HCl = CuCl₂ + H₂O', 'Fe2O3+H2SO4': 'Fe₂O₃ + 3H₂SO₄ = Fe₂(SO₄)₃ + 3H₂O',
    'CaO+CO2': 'CaO + CO₂ = CaCO₃', 'CaO+SO2': 'CaO + SO₂ = CaSO₃', 'MgO+CO2': 'MgO + CO₂ = MgCO₃',
    'Cl2+NaBr': 'Cl₂ + 2NaBr = 2NaCl + Br₂', 'Cl2+KI': 'Cl₂ + 2KI = 2KCl + I₂', 'Br2+NaI': 'Br₂ + 2NaI = 2NaBr + I₂',
    'F2+NaCl': 'F₂ + 2NaCl = 2NaF + Cl₂', 'HF+SiO2': '4HF + SiO₂ = SiF₄↑ + 2H₂O', 'HF+NaOH': 'HF + NaOH = NaF + H₂O',
    'Fe+S': 'Fe + S = FeS', 'Cu+S': '2Cu + S = Cu₂S', 'H2+S': 'H₂ + S = H₂S',
    'Na2O2+H2S': 'Na₂O₂ + H₂S = 2NaOH + S↓', 'Na2O2+SO2': 'Na₂O₂ + SO₂ = Na₂SO₄',
    'Mg+CO2': '2Mg + CO₂ = 2MgO + C',
    'NH3+H2SO4': '2NH₃ + H₂SO₄ = (NH₄)₂SO₄', 'H3PO4+NaOH': 'H₃PO₄ + 3NaOH = Na₃PO₄ + 3H₂O', 'Na3PO4+BaCl2': '2Na₃PO₄ + 3BaCl₂ = Ba₃(PO₄)₂↓ + 6NaCl',
    'HI+NaOH': 'HI + NaOH = NaI + H₂O', 'HBr+KOH': 'HBr + KOH = KBr + H₂O', 'KClO3+S': '2KClO₃ + 3S = 2KCl + 3SO₂↑',
    'NH4NO3+NaOH': 'NH₄NO₃ + NaOH = NaNO₃ + NH₃↑ + H₂O', 'ZnSO4+BaCl2': 'ZnSO₄ + BaCl₂ = BaSO₄↓ + ZnCl₂',
    'MgSO4+Ba(OH)2': 'MgSO₄ + Ba(OH)₂ = BaSO₄↓ + Mg(OH)₂↓', 'Al2(SO4)3+KOH': 'Al₂(SO₄)₃ + 6KOH = 2Al(OH)₃↓ + 3K₂SO₄',
    'Al+Cl2': '2Al + 3Cl₂ = 2AlCl₃', 'Al2O3+NaOH': 'Al₂O₃ + 2NaOH = 2NaAlO₂ + H₂O',
    'Na+H2': '2Na + H₂ = 2NaH', 'K+H2': '2K + H₂ = 2KH', 'Mg+H2': 'Mg + H₂ = MgH₂', 'Ca+H2': 'Ca + H₂ = CaH₂', 'Ba+H2': 'Ba + H₂ = BaH₂',
    'KH+H2O': 'KH + H₂O = KOH + H₂↑', 'MgH2+H2O': 'MgH₂ + 2H₂O = Mg(OH)₂ + 2H₂↑', 'BaH2+H2O': 'BaH₂ + 2H₂O = Ba(OH)₂ + 2H₂↑',
    'NaH+HCl': 'NaH + HCl = NaCl + H₂↑', 'NaH+H2SO4': '2NaH + H₂SO₄ = Na₂SO₄ + 2H₂↑', 'CaH2+HCl': 'CaH₂ + 2HCl = CaCl₂ + 2H₂↑',
    'CaH2+H2SO4': 'CaH₂ + H₂SO₄ = CaSO₄ + 2H₂↑', 'KH+HCl': 'KH + HCl = KCl + H₂↑', 'MgH2+HCl': 'MgH₂ + 2HCl = MgCl₂ + 2H₂↑',
    'K+O2': '4K + O₂ = 2K₂O', 'Na+O2': '4Na + O₂ = 2Na₂O', 'Ca+O2': '2Ca + O₂ = 2CaO',
    'Na+O2->Na2O2': '2Na + O₂ = Na₂O₂', 'K+O2->KO2': 'K + O₂ = KO₂',
    'Fe+O2->Fe2O3': '4Fe + 3O₂ = 2Fe₂O₃', 'Fe+O2->FeO': '2Fe + O₂ = 2FeO',
    'Cu+O2->Cu2O': '4Cu + O₂ = 2Cu₂O',
    'SO2+KOH': 'SO₂ + 2KOH = K₂SO₃ + H₂O', 'SO3+KOH': 'SO₃ + 2KOH = K₂SO₄ + H₂O', 'P2O5+KOH': 'P₂O₅ + 6KOH = 2K₃PO₄ + 3H₂O',
    'CuO+HNO3': 'CuO + 2HNO₃ = Cu(NO₃)₂ + H₂O', 'Fe2O3+HCl': 'Fe₂O₃ + 6HCl = 2FeCl₃ + 3H₂O', 'Fe2O3+HNO3': 'Fe₂O₃ + 6HNO₃ = 2Fe(NO₃)₃ + 3H₂O',
    'MgO+HCl': 'MgO + 2HCl = MgCl₂ + H₂O', 'MgO+H2SO4': 'MgO + H₂SO₄ = MgSO₄ + H₂O', 'MgO+HNO3': 'MgO + 2HNO₃ = Mg(NO₃)₂ + H₂O',
    'CaO+HCl': 'CaO + 2HCl = CaCl₂ + H₂O', 'CaO+HNO3': 'CaO + 2HNO₃ = Ca(NO₃)₂ + H₂O', 'Na2O+HCl': 'Na₂O + 2HCl = 2NaCl + H₂O',
    'Na2O+H2SO4': 'Na₂O + H₂SO₄ = Na₂SO₄ + H₂O', 'CaO+SO3': 'CaO + SO₃ = CaSO₄', 'MgO+SO2': 'MgO + SO₂ = MgSO₃',
    'BaO+CO2': 'BaO + CO₂ = BaCO₃', 'BaO+SO2': 'BaO + SO₂ = BaSO₃', 'BaO+SO3': 'BaO + SO₃ = BaSO₄',
    'Na2O+CO2': 'Na₂O + CO₂ = Na₂CO₃', 'Na2O+SO2': 'Na₂O + SO₂ = Na₂SO₃', 'H3PO4+KOH': 'H₃PO₄ + 3KOH = K₃PO₄ + 3H₂O',
    'H3PO4+Ca(OH)2': '2H₃PO₄ + 3Ca(OH)₂ = Ca₃(PO₄)₂↓ + 6H₂O', 'Cl2+NaI': 'Cl₂ + 2NaI = 2NaCl + I₂',
    'Cl2+KBr': 'Cl₂ + 2KBr = 2KCl + Br₂', 'Br2+KI': 'Br₂ + 2KI = 2KBr + I₂', 'Cl2+H2O': 'Cl₂ + H₂O = HCl + HClO',
    'Br2+H2O': 'Br₂ + H₂O = HBr + HBrO', 'Cl2+NaOH': 'Cl₂ + 2NaOH = NaCl + NaClO + H₂O', 'Cl2+Ca(OH)2': '2Cl₂ + 2Ca(OH)₂ = CaCl₂ + Ca(ClO)₂ + 2H₂O',
    'F2+H2': 'F₂ + H₂ = 2HF', 'Cl2+H2': 'Cl₂ + H₂ = 2HCl', 'Br2+H2': 'Br₂ + H₂ = 2HBr', 'I2+H2': 'I₂ + H₂ = 2HI',
    'Na+Br2': '2Na + Br₂ = 2NaBr', 'Na+I2': '2Na + I₂ = 2NaI', 'Fe+Br2': '2Fe + 3Br₂ = 2FeBr₃', 'Cu+Cl2': 'Cu + Cl₂ = CuCl₂',
    'AgNO3+NaBr': 'AgNO₃ + NaBr = AgBr↓ + NaNO₃', 'AgNO3+NaI': 'AgNO₃ + NaI = AgI↓ + NaNO₃', 'AgNO3+KI': 'AgNO₃ + KI = AgI↓ + KNO₃',
    'NaF+CaCl2': '2NaF + CaCl₂ = CaF₂↓ + 2NaCl', 'HF+CaO': '2HF + CaO = CaF₂ + H₂O',
    'SO2+O2': '2SO₂ + O₂ = 2SO₃', 'SO2+H2O': 'SO₂ + H₂O = H₂SO₃', 'Na2S+HCl': 'Na₂S + 2HCl = 2NaCl + H₂S↑',
    'H2SO3+O2': '2H₂SO₃ + O₂ = 2H₂SO₄', 'H2SO3+NaOH': 'H₂SO₃ + 2NaOH = Na₂SO₃ + 2H₂O', 'Na2SO3+S': 'Na₂SO₃ + S = Na₂S₂O₃',
    'H2S+SO2': '2H₂S + SO₂ = 3S↓ + 2H₂O', 'Mg+Br2': 'Mg + Br₂ = MgBr₂', 'Mg+I2': 'Mg + I₂ = MgI₂',
    'Al+Br2': '2Al + 3Br₂ = 2AlBr₃', 'Al+I2': '2Al + 3I₂ = 2AlI₃', 'Fe+I2': 'Fe + I₂ = FeI₂', 'Zn+Br2': 'Zn + Br₂ = ZnBr₂',
    'Zn+I2': 'Zn + I₂ = ZnI₂', 'Cu+Br2': 'Cu + Br₂ = CuBr₂', 'Ca+Br2': 'Ca + Br₂ = CaBr₂', 'Ca+I2': 'Ca + I₂ = CaI₂',
    'Fe+H2SO4': 'Fe + H₂SO₄ = FeSO₄ + H₂↑', 'Zn+H2SO4': 'Zn + H₂SO₄ = ZnSO₄ + H₂↑', 'Mg+H2SO4': 'Mg + H₂SO₄ = MgSO₄ + H₂↑',
    'Al+H2SO4': '2Al + 3H₂SO₄ = Al₂(SO₄)₃ + 3H₂↑', 'Na+HCl': '2Na + 2HCl = 2NaCl + H₂↑', 'Na+H2SO4': '2Na + H₂SO₄ = Na₂SO₄ + 2H₂↑',
    'Na+HNO3': '8Na + 10HNO₃ = 8NaNO₃ + NH₄NO₃ + 3H₂O', 'K+HCl': '2K + 2HCl = 2KCl + H₂↑', 'K+H2SO4': '2K + H₂SO₄ = K₂SO₄ + H₂↑',
    'K+HNO3': '8K + 10HNO₃ = 8KNO₃ + NH₄NO₃ + 3H₂O',
    'Ca+HCl': 'Ca + 2HCl = CaCl₂ + H₂↑', 'Ca+H2SO4': 'Ca + H₂SO₄ = CaSO₄ + H₂↑',
    'Ca+HNO3': '4Ca + 10HNO₃ = 4Ca(NO₃)₂ + NH₄NO₃ + 3H₂O', 'Ba+HCl': 'Ba + 2HCl = BaCl₂ + H₂↑', 'Ba+H2SO4': 'Ba + H₂SO₄ = BaSO₄ + H₂↑',
    'Ba+HNO3': '4Ba + 10HNO₃ = 4Ba(NO₃)₂ + NH₄NO₃ + 3H₂O', 'Cu+HNO3': 'Cu + 4HNO₃(浓) = Cu(NO₃)₂ + 2NO₂↑ + 2H₂O',
    'Ag+HNO3': 'Ag + 2HNO₃(浓) = AgNO₃ + NO₂↑ + H₂O', 'Hg+HNO3': 'Hg + 4HNO₃(浓) = Hg(NO₃)₂ + 2NO₂↑ + 2H₂O',
    'Fe+HNO3': 'Fe + 4HNO₃(稀) = Fe(NO₃)₃ + NO↑ + 2H₂O', 'Zn+HNO3': '4Zn + 10HNO₃(稀) = 4Zn(NO₃)₂ + NH₄NO₃ + 3H₂O',
    'Al+HNO3': '8Al + 30HNO₃(稀) = 8Al(NO₃)₃ + 3NH₄NO₃ + 9H₂O',
    'Mg+H3PO4': '3Mg + 2H₃PO₄ = Mg₃(PO₄)₂ + 3H₂↑', 'Zn+H3PO4': '3Zn + 2H₃PO₄ = Zn₃(PO₄)₂ + 3H₂↑', 'Fe+H3PO4': '3Fe + 2H₃PO₄ = Fe₃(PO₄)₂ + 3H₂↑',
    'Na+H3PO4': '6Na + 2H₃PO₄ = 2Na₃PO₄ + 3H₂↑', 'K+H3PO4': '6K + 2H₃PO₄ = 2K₃PO₄ + 3H₂↑',
    'Mg+HI': 'Mg + 2HI = MgI₂ + H₂↑', 'Zn+HI': 'Zn + 2HI = ZnI₂ + H₂↑', 'Fe+HI': 'Fe + 2HI = FeI₂ + H₂↑',
    'Mg+HBr': 'Mg + 2HBr = MgBr₂ + H₂↑', 'Zn+HBr': 'Zn + 2HBr = ZnBr₂ + H₂↑', 'Fe+HBr': 'Fe + 2HBr = FeBr₂ + H₂↑',
    'Mg+HF': 'Mg + 2HF = MgF₂ + H₂↑', 'Zn+HF': 'Zn + 2HF = ZnF₂ + H₂↑', 'Al+HF': '2Al + 6HF = 2AlF₃ + 3H₂↑',
    'K+HI': '2K + 2HI = 2KI + H₂↑', 'K+HBr': '2K + 2HBr = 2KBr + H₂↑', 'K+HF': '2K + 2HF = 2KF + H₂↑',
    'Na+HI': '2Na + 2HI = 2NaI + H₂↑', 'Na+HBr': '2Na + 2HBr = 2NaBr + H₂↑', 'Na+HF': '2Na + 2HF = 2NaF + H₂↑',
    'Ca+HI': 'Ca + 2HI = CaI₂ + H₂↑', 'Ca+HBr': 'Ca + 2HBr = CaBr₂ + H₂↑', 'Ca+HF': 'Ca + 2HF = CaF₂ + H₂↑',
    'Ba+HI': 'Ba + 2HI = BaI₂ + H₂↑', 'Ba+HBr': 'Ba + 2HBr = BaBr₂ + H₂↑', 'Ba+HF': 'Ba + 2HF = BaF₂ + H₂↑',
    'K+H2S': '2K + H₂S = K₂S + H₂↑', 'Na+H2S': '2Na + H₂S = Na₂S + H₂↑', 'Mg+H2S': 'Mg + H₂S = MgS + H₂↑',
    'Ca+H2S': 'Ca + H₂S = CaS + H₂↑', 'Ba+H2S': 'Ba + H₂S = BaS + H₂↑', 'Fe+H2S': 'Fe + H₂S = FeS + H₂↑',
    'K+H2SO3': '2K + H₂SO₃ = K₂SO₃ + H₂↑', 'Na+H2SO3': '2Na + H₂SO₃ = Na₂SO₃ + H₂↑', 'Mg+H2SO3': 'Mg + H₂SO₃ = MgSO₃ + H₂↑',
    'Ca+H2SO3': 'Ca + H₂SO₃ = CaSO₃ + H₂↑', 'Ba+H2SO3': 'Ba + H₂SO₃ = BaSO₃ + H₂↑',
    'H2O2+FeSO4': 'H₂O₂ + 2FeSO₄ + H₂SO₄ = Fe₂(SO₄)₃ + 2H₂O', 'H2O2+Na2SO3': 'H₂O₂ + Na₂SO₃ = Na₂SO₄ + H₂O',
    'H2O2+KI': 'H₂O₂ + 2KI = 2KOH + I₂', 'H2O2+H2S': 'H₂O₂ + H₂S = S↓ + 2H₂O', 'H2O2+SO2': 'H₂O₂ + SO₂ = H₂SO₄',
    'FeO+O2': '4FeO + O₂ = 2Fe₂O₃', 'Cu2O+O2': '2Cu₂O + O₂ = 4CuO',
    'Cl2+FeSO4': '3Cl₂ + 6FeSO₄ = 2Fe₂(SO₄)₃ + 2FeCl₃', 'Cl2+Na2SO3': 'Cl₂ + Na₂SO₃ + H₂O = Na₂SO₄ + 2HCl',
    'Cl2+H2S': 'Cl₂ + H₂S = S↓ + 2HCl', 'Cl2+FeCl2': 'Cl₂ + 2FeCl₂ = 2FeCl₃', 'Br2+FeSO4': '3Br₂ + 6FeSO₄ = 2Fe₂(SO₄)₃ + 2FeBr₃',
    'Br2+Na2SO3': 'Br₂ + Na₂SO₃ + H₂O = Na₂SO₄ + 2HBr', 'Br2+H2S': 'Br₂ + H₂S = S↓ + 2HBr',
    'FeCl3+Cu': '2FeCl₃ + Cu = 2FeCl₂ + CuCl₂', 'FeCl3+Fe': '2FeCl₃ + Fe = 3FeCl₂', 'FeCl3+KI': '2FeCl₃ + 2KI = 2FeCl₂ + 2KCl + I₂',
    'FeCl3+H2S': '2FeCl₃ + H₂S = 2FeCl₂ + S↓ + 2HCl', 'FeCl3+Na2SO3': '2FeCl₃ + Na₂SO₃ + H₂O = 2FeCl₂ + Na₂SO₄ + 2HCl',
    'HNO3+C': '4HNO₃(浓) + C = CO₂↑ + 4NO₂↑ + 2H₂O', 'HNO3+S': '6HNO₃(浓) + S = H₂SO₄ + 6NO₂↑ + 2H₂O',
    'HNO3+P': '5HNO₃(浓) + P = H₃PO₄ + 5NO₂↑ + H₂O', 'HNO3+FeO': 'FeO + 4HNO₃(浓) = Fe(NO₃)₃ + NO₂↑ + 2H₂O',
    'H2SO4+Cu': 'Cu + 2H₂SO₄(浓) = CuSO₄ + SO₂↑ + 2H₂O', 'H2SO4+C': 'C + 2H₂SO₄(浓) = CO₂↑ + 2SO₂↑ + 2H₂O',
    'H2SO4+S': 'S + 2H₂SO₄(浓) = 3SO₂↑ + 2H₂O', 'Na2O2+FeSO4': '3Na₂O₂ + 6FeSO₄ + 6H₂O = 4Fe(OH)₃↓ + 2Fe₂(SO₄)₃ + 6Na⁺',
    'O2+Fe(OH)2': '4Fe(OH)₂ + O₂ + 2H₂O = 4Fe(OH)₃', 'S+NaOH': '3S + 6NaOH = 2Na₂S + Na₂SO₃ + 3H₂O',
    'NO2+NaOH': '2NO₂ + 2NaOH = NaNO₂ + NaNO₃ + H₂O', 'Al+Fe3O4': '8Al + 3Fe₃O₄ = 4Al₂O₃ + 9Fe',
    'Al+CuO': '2Al + 3CuO = Al₂O₃ + 3Cu', 'C+CuO': 'C + 2CuO = 2Cu + CO₂↑', 'CO+CuO': 'CO + CuO = Cu + CO₂',
    'F2+NaBr': 'F₂ + 2NaBr = 2NaF + Br₂', 'F2+NaI': 'F₂ + 2NaI = 2NaF + I₂', 'F2+KCl': 'F₂ + 2KCl = 2KF + Cl₂',
    'F2+KBr': 'F₂ + 2KBr = 2KF + Br₂', 'F2+KI': 'F₂ + 2KI = 2KF + I₂', 'F2+Na': 'F₂ + 2Na = 2NaF',
    'F2+K': 'F₂ + 2K = 2KF', 'F2+Mg': 'F₂ + Mg = MgF₂', 'F2+Ca': 'F₂ + Ca = CaF₂', 'F2+Ba': 'F₂ + Ba = BaF₂',
    'F2+Al': '3F₂ + 2Al = 2AlF₃', 'F2+Fe': '3F₂ + 2Fe = 2FeF₃', 'F2+Cu': 'F₂ + Cu = CuF₂', 'F2+Ag': 'F₂ + 2Ag = 2AgF',
    'F2+Hg': 'F₂ + Hg = HgF₂', 'F2+Zn': 'F₂ + Zn = ZnF₂', 'F2+NH3': '3F₂ + 2NH₃ = 6HF + N₂',
    'F2+S': '3F₂ + S = SF₆', 'F2+P': '5F₂ + 2P = 2PF₅', 'F2+C': '2F₂ + C = CF₄', 'HF+KOH': 'HF + KOH = KF + H₂O',
    'HF+Ba(OH)2': '2HF + Ba(OH)₂ = BaF₂ + 2H₂O', 'HF+MgO': '2HF + MgO = MgF₂ + H₂O', 'HF+Al2O3': '6HF + Al₂O₃ = 2AlF₃ + 3H₂O',
    'NaF+BaCl2': '2NaF + BaCl₂ = BaF₂↓ + 2NaCl', 'KF+CaCl2': '2KF + CaCl₂ = CaF₂↓ + 2KCl', 'Zn+O2': '2Zn + O₂ = 2ZnO',
    'Ag+O2': '4Ag + O₂ = 2Ag₂O', 'Si+O2': 'Si + O₂ = SiO₂', 'Cu+O2': '2Cu + O₂ = 2CuO', 'Al+O2': '4Al + 3O₂ = 2Al₂O₃',
    'Ba+O2': '2Ba + O₂ = 2BaO', 'N2+O2': 'N₂ + O₂ = 2NO', 'Cl2+O2': '2Cl₂ + 7O₂ = 2Cl₂O₇', 'K+Cl2': '2K + Cl₂ = 2KCl',
    'K+Br2': '2K + Br₂ = 2KBr', 'K+I2': '2K + I₂ = 2KI', 'K+S': '2K + S = K₂S', 'Na+S': '2Na + S = Na₂S',
    'Mg+S': 'Mg + S = MgS', 'Ca+S': 'Ca + S = CaS', 'Ba+S': 'Ba + S = BaS', 'Al+S': '2Al + 3S = Al₂S₃',
    'Zn+S': 'Zn + S = ZnS', 'Ca+Cl2': 'Ca + Cl₂ = CaCl₂', 'Ba+Cl2': 'Ba + Cl₂ = BaCl₂', 'Zn+Cl2': 'Zn + Cl₂ = ZnCl₂',
    'Ag+Cl2': '2Ag + Cl₂ = 2AgCl', 'Hg+Cl2': 'Hg + Cl₂ = HgCl₂', 'Ba+Br2': 'Ba + Br₂ = BaBr₂', 'Ba+I2': 'Ba + I₂ = BaI₂',
    'Hg+Br2': 'Hg + Br₂ = HgBr₂', 'Hg+I2': 'Hg + I₂ = HgI₂', 'Ag+Br2': '2Ag + Br₂ = 2AgBr', 'Ag+I2': '2Ag + I₂ = 2AgI',
    'Cu+I2': '2Cu + I₂ = 2CuI', 'P+Cl2': '2P + 3Cl₂ = 2PCl₃', 'P+S': '2P + 5S = P₂S₅', 'Mg+ZnCl2': 'Mg + ZnCl₂ = MgCl₂ + Zn',
    'Mg+ZnSO4': 'Mg + ZnSO₄ = MgSO₄ + Zn', 'Mg+FeCl2': 'Mg + FeCl₂ = MgCl₂ + Fe', 'Mg+FeSO4': 'Mg + FeSO₄ = MgSO₄ + Fe',
    'Mg+CuCl2': 'Mg + CuCl₂ = MgCl₂ + Cu', 'Mg+AlCl3': '3Mg + 2AlCl₃ = 3MgCl₂ + 2Al', 'Mg+Al2(SO4)3': '3Mg + Al₂(SO₄)₃ = 3MgSO₄ + 2Al',
    'Zn+FeCl2': 'Zn + FeCl₂ = ZnCl₂ + Fe', 'Zn+FeSO4': 'Zn + FeSO₄ = ZnSO₄ + Fe', 'Zn+CuCl2': 'Zn + CuCl₂ = ZnCl₂ + Cu',
    'Fe+CuCl2': 'Fe + CuCl₂ = FeCl₂ + Cu', 'Al+ZnCl2': '2Al + 3ZnCl₂ = 2AlCl₃ + 3Zn', 'Al+FeCl2': '2Al + 3FeCl₂ = 2AlCl₃ + 3Fe',
    'Al+CuCl2': '2Al + 3CuCl₂ = 2AlCl₃ + 3Cu',
    'Mg+HNO3': 'Mg + 2HNO₃ = Mg(NO₃)₂ + H₂↑', 'H2+Fe2O3': '3H₂ + Fe₂O₃ = 2Fe + 3H₂O', 'Ca+H3PO4': '3Ca + 2H₃PO₄ = Ca₃(PO₄)₂↓ + 3H₂↑',
    'Al+H3PO4': '2Al + 2H₃PO₄ = 2AlPO₄↓ + 3H₂↑', 'Al+HI': '2Al + 6HI = 2AlI₃ + 3H₂↑', 'Al+HBr': '2Al + 6HBr = 2AlBr₃ + 3H₂↑',
    'Al+H2S': '2Al + 3H₂S = Al₂S₃ + 3H₂↑', 'Al+H2SO3': '2Al + 3H₂SO₃ = Al₂(SO₃)₃ + 3H₂↑', 'Zn+H2S': 'Zn + H₂S = ZnS↓ + H₂↑',
    'Fe+H2SO3': 'Fe + H₂SO₃ = FeSO₃ + H₂↑', 'Mg+CuSO4': 'Mg + CuSO₄ = MgSO₄ + Cu'
};

function showReactionEquation(a, b) {
    if (!a || !b) {
        console.log("Reaction matching skipped: current or last card is null", {a, b});
        return;
    }
    
    const fA = a.formula, fB = b.formula;
    console.log(`Attempting to match equation for: ${fA} + ${fB}`);
    let eq = null;

    // 1. 优先尝试直接反应物匹配 (UNO 模式：反应物打在反应物上)
    eq = EQUATIONS_LIB[`${fA}+${fB}`] || EQUATIONS_LIB[`${fB}+${fA}`];
    if (eq) console.log(`Direct match found: ${eq}`);

    // 2. 尝试路径匹配 (合成模式：产物打在反应物上)
    if (!eq) {
        const getOtherReactantKey = (el) => {
            const map = { 'O': 'O2', 'H': 'H2', 'Cl': 'Cl2', 'N': 'N2', 'F': 'F2', 'Br': 'Br2', 'I': 'I2' };
            return map[el] || el;
        };

        const aComp = Object.keys(a.composition || {});
        const bComp = Object.keys(b.composition || {});
        
        // 场景 A: b 包含 a 之外的元素 (如 a=O2, b=H2O，玩家合成了水)
        const extraInB = bComp.filter(el => !aComp.includes(el));
        if (extraInB.length > 0) {
            const other = getOtherReactantKey(extraInB[0]);
            eq = EQUATIONS_LIB[`${fA}+${other}->${fB}`] || 
                 EQUATIONS_LIB[`${other}+${fA}->${fB}`] || 
                 EQUATIONS_LIB[`${fA}+${other}`] || 
                 EQUATIONS_LIB[`${other}+${fA}`];
            if (eq) console.log(`Path A match found (${fA} + ${other} -> ${fB}): ${eq}`);
        }
        
        // 场景 B: a 包含 b 之外的元素 (如 a=H2O, b=O2，玩家分解了水 - 概率较低)
        if (!eq) {
            const extraInA = aComp.filter(el => !bComp.includes(el));
            if (extraInA.length > 0) {
                const other = getOtherReactantKey(extraInA[0]);
                eq = EQUATIONS_LIB[`${fB}+${other}->${fA}`] || 
                     EQUATIONS_LIB[`${other}+${fB}->${fA}`] ||
                     EQUATIONS_LIB[`${fB}+${other}`] ||
                     EQUATIONS_LIB[`${other}+${fB}`];
                if (eq) console.log(`Path B match found (${fB} + ${other} -> ${fA}): ${eq}`);
            }
        }
    }
    
    if (!eq) {
        console.log(`No equation found in LIB for ${fA} and ${fB}`);
        return;
    }

    // 同步记录到实验日志
    const logFormatted = eq.replace(/([a-zA-Z\)])(\d+)/g, '$1<sub>$2</sub>')
                           .replace(/->/g, '→')
                           .replace(/\+/g, ' + ')
                           .replace(/=/g, ' = ');
    log(`<b>化学反应:</b> ${logFormatted}`);

    const overlay = document.getElementById('equation-overlay');
    const text = document.getElementById('equation-text');
    
    // 改进后的方程式美化器：
    // 1. 只在该位数字前有字母或括号时才作为下标 (如 H2O -> H₂O, 但 2H2 -> 2H₂)
    let formatted = eq.replace(/([a-zA-Z\)])(\d+)/g, '$1<sub>$2</sub>');
    // 2. 替换各种符号
    formatted = formatted.replace(/->/g, ' → ')
                         .replace(/=/g, ' = ')
                         .replace(/\+/g, ' + ');
    
    text.innerHTML = formatted;
    
    overlay.style.opacity = '1';
    overlay.style.transform = 'translateY(0)';
    
    // 5秒后自动隐藏
    if (window.eqTimeout) clearTimeout(window.eqTimeout);
    window.eqTimeout = setTimeout(() => {
        overlay.style.opacity = '0';
        overlay.style.transform = 'translateY(-20px)';
    }, 5000);
}
const ELEMENTS_DATA = {
    'H': {name:'氢',bg:'bg-H'}, 'O': {name:'氧',bg:'bg-O'}, 'C': {name:'碳',bg:'bg-C'}, 'N': {name:'氮'}, 'F': {name:'氟'}, 'Na': {name:'钠'}, 'Mg': {name:'镁'}, 'Al': {name:'铝'}, 'Si': {name:'硅'}, 'P': {name:'磷'}, 'S': {name:'硫'}, 'Cl': {name:'氯'}, 'K': {name:'钾'}, 'Ca': {name:'钙'}, 'Fe': {name:'铁'}, 'Cu': {name:'铜'}, 'Zn': {name:'锌'}, 'Br': {name:'溴'}, 'I': {name:'碘'}, 'Ag': {name:'银'}, 'Ba': {name:'钡'}, 'Hg': {name:'汞'}
};
const SPECIAL_CARDS = {
    'He': {name:'反转',bg:'bg-skip',eff:'reverse'}, 'Ne': {name:'反转',bg:'bg-skip',eff:'reverse'}, 'Ar': {name:'反转',bg:'bg-skip',eff:'reverse'}, 'Kr': {name:'反转',bg:'bg-skip',eff:'reverse'}, 'Au': {name:'跳过',bg:'bg-skip',eff:'skip'}
};

// 占位符，将被爬虫替换
const COMPOUNDS_DB = [
    { "formula": "H2O", "name": "水", "composition": { "H": 2, "O": 1 } },
    { "formula": "CO2", "name": "二氧化碳", "composition": { "C": 1, "O": 2 } },
    { "formula": "HCl", "name": "盐酸", "composition": { "H": 1, "Cl": 1 } },
    { "formula": "NaOH", "name": "氢氧化钠", "composition": { "Na": 1, "O": 1, "H": 1 } },
    { "formula": "NaCl", "name": "氯化钠", "composition": { "Na": 1, "Cl": 1 } },
    { "formula": "H2SO4", "name": "硫酸", "composition": { "H": 2, "S": 1, "O": 4 } },
    { "formula": "O2", "name": "氧气", "composition": { "O": 2 } },
    { "formula": "H2", "name": "氢气", "composition": { "H": 2 } },
    { "formula": "Fe", "name": "铁", "composition": { "Fe": 1 } },
    { "formula": "CuSO4", "name": "硫酸铜", "composition": { "Cu": 1, "S": 1, "O": 4 } },
    { "formula": "Cu", "name": "铜", "composition": { "Cu": 1 } },
    { "formula": "Zn", "name": "锌", "composition": { "Zn": 1 } },
    { "formula": "CO", "name": "一氧化碳", "composition": { "C": 1, "O": 1 } },
    { "formula": "CaCO3", "name": "碳酸钙", "composition": { "Ca": 1, "C": 1, "O": 3 } },
    { "formula": "CaO", "name": "氧化钙", "composition": { "Ca": 1, "O": 1 } },
    { "formula": "Ca(OH)2", "name": "氢氧化钙", "composition": { "Ca": 1, "O": 2, "H": 2 } },
    { "formula": "NH3", "name": "氨气", "composition": { "N": 1, "H": 3 } },
    { "formula": "HNO3", "name": "硝酸", "composition": { "H": 1, "N": 1, "O": 3 } },
    { "formula": "AgNO3", "name": "硝酸银", "composition": { "Ag": 1, "N": 1, "O": 3 } },
    { "formula": "AgCl", "name": "氯化银", "composition": { "Ag": 1, "Cl": 1 } },
    { "formula": "MgO", "name": "氧化镁", "composition": { "Mg": 1, "O": 1 } },
    { "formula": "Mg", "name": "镁", "composition": { "Mg": 1 } },
    { "formula": "Al", "name": "铝", "composition": { "Al": 1 } },
    { "formula": "Al2O3", "name": "氧化铝", "composition": { "Al": 2, "O": 3 } },
    { "formula": "Fe2O3", "name": "氧化铁", "composition": { "Fe": 2, "O": 3 } },
    { "formula": "Fe3O4", "name": "四氧化三铁", "composition": { "Fe": 3, "O": 4 } },
    { "formula": "Na2CO3", "name": "碳酸钠", "composition": { "Na": 2, "C": 1, "O": 3 } },
    { "formula": "NaHCO3", "name": "碳酸氢钠", "composition": { "Na": 1, "H": 1, "C": 1, "O": 3 } },
    { "formula": "KClO3", "name": "氯酸钾", "composition": { "K": 1, "Cl": 1, "O": 3 } },
    { "formula": "KCl", "name": "氯化钾", "composition": { "K": 1, "Cl": 1 } },
    { "formula": "H2O2", "name": "过氧化氢", "composition": { "H": 2, "O": 2 } },
    { "formula": "SO2", "name": "二氧化硫", "composition": { "S": 1, "O": 2 } },
    { "formula": "BaCl2", "name": "氯化钡", "composition": { "Ba": 1, "Cl": 2 } },
    { "formula": "BaSO4", "name": "硫酸钡", "composition": { "Ba": 1, "S": 1, "O": 4 } },
    { "formula": "C", "name": "碳", "composition": { "C": 1 } },
    { "formula": "S", "name": "硫", "composition": { "S": 1 } },
    { "formula": "P", "name": "磷", "composition": { "P": 1 } },
    { "formula": "P2O5", "name": "五氧化二磷", "composition": { "P": 2, "O": 5 } },
    { "formula": "CuO", "name": "氧化铜", "composition": { "Cu": 1, "O": 1 } },
    { "formula": "FeCl3", "name": "氯化铁", "composition": { "Fe": 1, "Cl": 3 } },
    { "formula": "FeSO4", "name": "硫酸亚铁", "composition": { "Fe": 1, "S": 1, "O": 4 } },
    { "formula": "KOH", "name": "氢氧化钾", "composition": { "K": 1, "O": 1, "H": 1 } },
    { "formula": "MgCl2", "name": "氯化镁", "composition": { "Mg": 1, "Cl": 2 } },
    { "formula": "CaCl2", "name": "氯化钙", "composition": { "Ca": 1, "Cl": 2 } },
    { "formula": "NH4Cl", "name": "氯化铵", "composition": { "N": 1, "H": 4, "Cl": 1 } },
    { "formula": "Na2SO4", "name": "硫酸钠", "composition": { "Na": 2, "S": 1, "O": 4 } },
    { "formula": "NO", "name": "一氧化氮", "composition": { "N": 1, "O": 1 } },
    { "formula": "NO2", "name": "二氧化氮", "composition": { "N": 1, "O": 2 } },
    { "formula": "N2", "name": "氮气", "composition": { "N": 2 } },
    { "formula": "Hg", "name": "汞", "composition": { "Hg": 1 } },
    { "formula": "HgO", "name": "氧化汞", "composition": { "Hg": 1, "O": 1 } },
    { "formula": "Na2O", "name": "氧化钠", "composition": { "Na": 2, "O": 1 } },
    { "formula": "K2O", "name": "氧化钾", "composition": { "K": 2, "O": 1 } },
    { "formula": "SO3", "name": "三氧化硫", "composition": { "S": 1, "O": 3 } },
    { "formula": "BaO", "name": "氧化钡", "composition": { "Ba": 1, "O": 1 } },
    { "formula": "CuCl2", "name": "氯化铜", "composition": { "Cu": 1, "Cl": 2 } },
    { "formula": "Cu(OH)2", "name": "氢氧化铜", "composition": { "Cu": 1, "O": 2, "H": 2 } },
    { "formula": "Fe(OH)3", "name": "氢氧化铁", "composition": { "Fe": 1, "O": 3, "H": 3 } },
    { "formula": "NH4NO3", "name": "硝酸铵", "composition": { "N": 2, "H": 4, "O": 3 } },
    { "formula": "ZnCl2", "name": "氯化锌", "composition": { "Zn": 1, "Cl": 2 } },
    { "formula": "ZnSO4", "name": "硫酸锌", "composition": { "Zn": 1, "S": 1, "O": 4 } },
    { "formula": "AlCl3", "name": "氯化铝", "composition": { "Al": 1, "Cl": 3 } },
    { "formula": "Ag", "name": "银", "composition": { "Ag": 1 } },
    { "formula": "H3PO4", "name": "磷酸", "composition": { "H": 3, "P": 1, "O": 4 } },
    { "formula": "Na3PO4", "name": "磷酸钠", "composition": { "Na": 3, "P": 1, "O": 4 } },
    { "formula": "Cl2", "name": "氯气", "composition": { "Cl": 2 } },
    { "formula": "Br2", "name": "溴单质", "composition": { "Br": 2 } },
    { "formula": "I2", "name": "碘单质", "composition": { "I": 2 } },
    { "formula": "HI", "name": "碘化氢", "composition": { "H": 1, "I": 1 } },
    { "formula": "HBr", "name": "溴化氢", "composition": { "H": 1, "Br": 1 } },
    { "formula": "K2SO4", "name": "硫酸钾", "composition": { "K": 2, "S": 1, "O": 4 } },
    { "formula": "MgSO4", "name": "硫酸镁", "composition": { "Mg": 1, "S": 1, "O": 4 } },
    { "formula": "CaSO4", "name": "硫酸钙", "composition": { "Ca": 1, "S": 1, "O": 4 } },
    { "formula": "Na2O2", "name": "过氧化钠", "composition": { "Na": 2, "O": 2 } },
    { "formula": "Na", "name": "钠", "composition": { "Na": 1 } },
    { "formula": "K", "name": "钾", "composition": { "K": 1 } },
    { "formula": "Ca", "name": "钙", "composition": { "Ca": 1 } },
    { "formula": "NaAlO2", "name": "偏铝酸钠", "composition": { "Na": 1, "Al": 1, "O": 2 } },
    { "formula": "K2CO3", "name": "碳酸钾", "composition": { "K": 2, "C": 1, "O": 3 } },
    { "formula": "NaH", "name": "氢化钠", "composition": { "Na": 1, "H": 1 } },
    { "formula": "KH", "name": "氢化钾", "composition": { "K": 1, "H": 1 } },
    { "formula": "MgH2", "name": "氢化镁", "composition": { "Mg": 1, "H": 2 } },
    { "formula": "BaH2", "name": "氢化钡", "composition": { "Ba": 1, "H": 2 } },
    { "formula": "KO2", "name": "超氧化钾", "composition": { "K": 1, "O": 2 } },
    { "formula": "CaH2", "name": "氢化钙", "composition": { "Ca": 1, "H": 2 } },
    { "formula": "NaOH", "name": "氢氧化钠", "composition": { "Na": 1, "O": 1, "H": 1 } },
    { "formula": "Na2SO3", "name": "亚硫酸钠", "composition": { "Na": 2, "S": 1, "O": 3 } },
    { "formula": "K2SO3", "name": "亚硫酸钾", "composition": { "K": 2, "S": 1, "O": 3 } },
    { "formula": "NaHSO4", "name": "硫酸氢钠", "composition": { "Na": 1, "H": 1, "S": 1, "O": 4 } },
    { "formula": "NaHSO3", "name": "亚硫酸氢钠", "composition": { "Na": 1, "H": 1, "S": 1, "O": 3 } },
    { "formula": "CaSO3", "name": "亚硫酸钙", "composition": { "Ca": 1, "S": 1, "O": 3 } },
    { "formula": "BaCO3", "name": "碳酸钡", "composition": { "Ba": 1, "C": 1, "O": 3 } },
    { "formula": "NaBr", "name": "溴化钠", "composition": { "Na": 1, "Br": 1 } },
    { "formula": "NaI", "name": "碘化钠", "composition": { "Na": 1, "I": 1 } },
    { "formula": "KBr", "name": "溴化钾", "composition": { "K": 1, "Br": 1 } },
    { "formula": "AgBr", "name": "溴化银", "composition": { "Ag": 1, "Br": 1 } },
    { "formula": "AgI", "name": "碘化银", "composition": { "Ag": 1, "I": 1 } },
    { "formula": "F2", "name": "氟气", "composition": { "F": 2 } },
    { "formula": "HF", "name": "氢氟酸", "composition": { "H": 1, "F": 1 } },
    { "formula": "NaF", "name": "氟化钠", "composition": { "Na": 1, "F": 1 } },
    { "formula": "CaF2", "name": "氟化钙", "composition": { "Ca": 1, "F": 2 } },
    { "formula": "HClO", "name": "次氯酸", "composition": { "H": 1, "Cl": 1, "O": 1 } },
    { "formula": "NaClO", "name": "次氯酸钠", "composition": { "Na": 1, "Cl": 1, "O": 1 } },
    { "formula": "H2S", "name": "硫化氢", "composition": { "H": 2, "S": 1 } },
    { "formula": "Na2S", "name": "硫化钠", "composition": { "Na": 2, "S": 1 } },
    { "formula": "FeS", "name": "硫化亚铁", "composition": { "Fe": 1, "S": 1 } },
    { "formula": "CuS", "name": "硫化铜", "composition": { "Cu": 1, "S": 1 } },
    { "formula": "ZnS", "name": "硫化锌", "composition": { "Zn": 1, "S": 1 } },
    { "formula": "Ag2S", "name": "硫化银", "composition": { "Ag": 2, "S": 1 } },
    { "formula": "H2SO3", "name": "亚硫酸", "composition": { "H": 2, "S": 1, "O": 3 } },
    { "formula": "BaSO3", "name": "亚硫酸钡", "composition": { "Ba": 1, "S": 1, "O": 3 } },
    { "formula": "MgBr2", "name": "溴化镁", "composition": { "Mg": 1, "Br": 2 } },
    { "formula": "MgI2", "name": "碘化镁", "composition": { "Mg": 1, "I": 2 } },
    { "formula": "AlBr3", "name": "溴化铝", "composition": { "Al": 1, "Br": 3 } },
    { "formula": "AlI3", "name": "碘化铝", "composition": { "Al": 1, "I": 3 } },
    { "formula": "ZnBr2", "name": "溴化锌", "composition": { "Zn": 1, "Br": 2 } },
    { "formula": "ZnI2", "name": "碘化锌", "composition": { "Zn": 1, "I": 2 } },
    { "formula": "CuBr2", "name": "溴化铜", "composition": { "Cu": 1, "Br": 2 } },
    { "formula": "FeBr3", "name": "溴化铁", "composition": { "Fe": 1, "Br": 3 } },
    { "formula": "FeI2", "name": "碘化亚铁", "composition": { "Fe": 1, "I": 2 } },
    { "formula": "CaBr2", "name": "溴化钙", "composition": { "Ca": 1, "Br": 2 } },
    { "formula": "CaI2", "name": "碘化钙", "composition": { "Ca": 1, "I": 2 } },
    { "formula": "KF", "name": "氟化钾", "composition": { "K": 1, "F": 1 } },
    { "formula": "BaF2", "name": "氟化钡", "composition": { "Ba": 1, "F": 2 } },
    { "formula": "MgF2", "name": "氟化镁", "composition": { "Mg": 1, "F": 2 } },
    { "formula": "AlF3", "name": "氟化铝", "composition": { "Al": 1, "F": 3 } },
    { "formula": "FeF3", "name": "氟化铁", "composition": { "Fe": 1, "F": 3 } },
    { "formula": "CuF2", "name": "氟化铜", "composition": { "Cu": 1, "F": 2 } },
    { "formula": "AgF", "name": "氟化银", "composition": { "Ag": 1, "F": 1 } },
    { "formula": "HgF2", "name": "氟化汞", "composition": { "Hg": 1, "F": 2 } },
    { "formula": "ZnF2", "name": "氟化锌", "composition": { "Zn": 1, "F": 2 } },
    { "formula": "SiO2", "name": "二氧化硅", "composition": { "Si": 1, "O": 2 } },
    { "formula": "SiF4", "name": "四氟化硅", "composition": { "Si": 1, "F": 4 } },
    { "formula": "ZnO", "name": "氧化锌", "composition": { "Zn": 1, "O": 1 } },
    { "formula": "Cu2O", "name": "氧化亚铜", "composition": { "Cu": 2, "O": 1 } },
    { "formula": "FeO", "name": "氧化亚铁", "composition": { "Fe": 1, "O": 1 } },
    { "formula": "Ag2O", "name": "氧化银", "composition": { "Ag": 2, "O": 1 } },
    { "formula": "CO2", "name": "二氧化碳", "composition": { "C": 1, "O": 2 } },
    { "formula": "SO2", "name": "二氧化硫", "composition": { "S": 1, "O": 2 } },
    { "formula": "SO3", "name": "三氧化硫", "composition": { "S": 1, "O": 3 } },
    { "formula": "P2O5", "name": "五氧化二磷", "composition": { "P": 2, "O": 5 } },
    { "formula": "NO", "name": "一氧化氮", "composition": { "N": 1, "O": 1 } },
    { "formula": "NO2", "name": "二氧化氮", "composition": { "N": 1, "O": 2 } },
    { "formula": "N2O", "name": "一氧化二氮", "composition": { "N": 2, "O": 1 } },
    { "formula": "Cl2O7", "name": "七氧化二氯", "composition": { "Cl": 2, "O": 7 } },
    { "formula": "K2S", "name": "硫化钾", "composition": { "K": 2, "S": 1 } },
    { "formula": "MgS", "name": "硫化镁", "composition": { "Mg": 1, "S": 1 } },
    { "formula": "CaS", "name": "硫化钙", "composition": { "Ca": 1, "S": 1 } },
    { "formula": "BaS", "name": "硫化钡", "composition": { "Ba": 1, "S": 1 } },
    { "formula": "MgSO3", "name": "亚硫酸镁", "composition": { "Mg": 1, "S": 1, "O": 3 } },
    { "formula": "Al2S3", "name": "硫化铝", "composition": { "Al": 2, "S": 3 } },
    { "formula": "Al2(SO3)3", "name": "亚硫酸铝", "composition": { "Al": 2, "S": 3, "O": 9 } },
    { "formula": "FeSO3", "name": "亚硫酸亚铁", "composition": { "Fe": 1, "S": 1, "O": 3 } },
    { "formula": "KI", "name": "碘化钾", "composition": { "K": 1, "I": 1 } },
    { "formula": "BaBr2", "name": "溴化钡", "composition": { "Ba": 1, "Br": 2 } },
    { "formula": "BaI2", "name": "碘化钡", "composition": { "Ba": 1, "I": 2 } },
    { "formula": "CaS", "name": "硫化钙", "composition": { "Ca": 1, "S": 1 } },
    { "formula": "BaS", "name": "硫化钡", "composition": { "Ba": 1, "S": 1 } },
    { "formula": "Al2S3", "name": "硫化铝", "composition": { "Al": 2, "S": 3 } },
    { "formula": "BaBr2", "name": "溴化钡", "composition": { "Ba": 1, "Br": 2 } },
    { "formula": "BaI2", "name": "碘化钡", "composition": { "Ba": 1, "I": 2 } },
    { "formula": "FeCl2", "name": "氯化亚铁", "composition": { "Fe": 1, "Cl": 2 } },
    { "formula": "Al2(SO4)3", "name": "硫酸铝", "composition": { "Al": 2, "S": 3, "O": 12 } }
];
const REACTIONS_DB = [
    { "a": "H2", "b": "O2" }, { "a": "C", "b": "O2" }, { "a": "S", "b": "O2" }, { "a": "P", "b": "O2" },
    { "a": "Fe", "b": "O2" }, { "a": "Mg", "b": "O2" }, { "a": "CO", "b": "O2" },
    { "a": "HCl", "b": "NaOH" }, { "a": "H2SO4", "b": "NaOH" }, { "a": "HNO3", "b": "NaOH" }, { "a": "HCl", "b": "Ca(OH)2" },
    { "a": "H2SO4", "b": "BaCl2" }, { "a": "HCl", "b": "AgNO3" }, { "a": "NaCl", "b": "AgNO3" }, { "a": "Fe", "b": "HCl" },
    { "a": "Zn", "b": "HCl" }, { "a": "Mg", "b": "HCl" }, { "a": "Al", "b": "HCl" }, { "a": "Fe", "b": "CuSO4" },
    { "a": "Zn", "b": "CuSO4" }, { "a": "CO2", "b": "Ca(OH)2" }, { "a": "CO2", "b": "NaOH" }, { "a": "CO2", "b": "H2O" },
    { "a": "CaO", "b": "H2O" }, { "a": "MgO", "b": "H2O" }, { "a": "CaCO3", "b": "HCl" }, { "a": "Na2CO3", "b": "HCl" },
    { "a": "NaHCO3", "b": "HCl" }, { "a": "NH3", "b": "HCl" }, { "a": "Cu", "b": "AgNO3" }, { "a": "CuO", "b": "H2" },
    { "a": "Fe2O3", "b": "CO" }, { "a": "CuO", "b": "C" }, { "a": "BaCl2", "b": "Na2SO4" }, { "a": "MgCl2", "b": "NaOH" },
    { "a": "N2", "b": "H2" },
    { "a": "NO", "b": "O2" }, { "a": "NO2", "b": "H2O" }, { "a": "NH3", "b": "H2SO4" },
    { "a": "HgO", "b": "C" }, { "a": "Hg", "b": "O2" },
    { "a": "Na2O", "b": "H2O" }, { "a": "K2O", "b": "H2O" }, { "a": "SO3", "b": "H2O" },
    { "a": "BaO", "b": "H2O" }, { "a": "CuCl2", "b": "NaOH" }, { "a": "FeCl3", "b": "NaOH" }, { "a": "AlCl3", "b": "NaOH" },
    { "a": "Cu(OH)2", "b": "HCl" }, { "a": "Fe(OH)3", "b": "H2SO4" },
    { "a": "AgNO3", "b": "BaCl2" }, { "a": "AgNO3", "b": "Cu" }, { "a": "Zn", "b": "AgNO3" }, { "a": "Fe", "b": "AgNO3" },
    { "a": "H3PO4", "b": "NaOH" }, { "a": "Na3PO4", "b": "BaCl2" }, { "a": "HI", "b": "NaOH" }, { "a": "HBr", "b": "KOH" },
    { "a": "KClO3", "b": "S" },
    { "a": "NH4Cl", "b": "NaOH" }, { "a": "NH4NO3", "b": "NaOH" }, { "a": "Ca(OH)2", "b": "Na2CO3" }, { "a": "Ba(OH)2", "b": "Na2SO4" },
    { "a": "ZnSO4", "b": "BaCl2" }, { "a": "MgSO4", "b": "Ba(OH)2" }, { "a": "Al2(SO4)3", "b": "KOH" }, { "a": "I2", "b": "H2" },
    { "a": "Na", "b": "H2O" }, { "a": "K", "b": "H2O" }, { "a": "Ca", "b": "H2O" },
    { "a": "Na", "b": "Cl2" }, { "a": "Mg", "b": "Cl2" }, { "a": "Al", "b": "Cl2" }, { "a": "Fe", "b": "Cl2" },
    { "a": "Na2O2", "b": "H2O" }, { "a": "Na2O2", "b": "CO2" },
    { "a": "Mg", "b": "CO2" },
    { "a": "Al", "b": "NaOH" }, { "a": "Al2O3", "b": "NaOH" },
    // 氢化物合成
    { "a": "Na", "b": "H2" }, { "a": "K", "b": "H2" },
    { "a": "Mg", "b": "H2" }, { "a": "Ca", "b": "H2" }, { "a": "Ba", "b": "H2" },
    // 氢化物反应
    { "a": "NaH", "b": "H2O" }, { "a": "CaH2", "b": "H2O" },
    { "a": "KH", "b": "H2O" }, { "a": "MgH2", "b": "H2O" }, { "a": "BaH2", "b": "H2O" },
    { "a": "NaH", "b": "HCl" }, { "a": "NaH", "b": "H2SO4" }, { "a": "CaH2", "b": "HCl" }, { "a": "CaH2", "b": "H2SO4" },
    { "a": "KH", "b": "HCl" }, { "a": "MgH2", "b": "HCl" },
    { "a": "K", "b": "O2" }, { "a": "Na", "b": "O2" }, { "a": "Ca", "b": "O2" },
    { "a": "SO2", "b": "NaOH" }, { "a": "SO2", "b": "KOH" }, { "a": "SO2", "b": "Ca(OH)2" },
    { "a": "SO3", "b": "NaOH" }, { "a": "SO3", "b": "KOH" }, { "a": "SO3", "b": "Ca(OH)2" },
    { "a": "P2O5", "b": "NaOH" }, { "a": "P2O5", "b": "KOH" }, { "a": "P2O5", "b": "Ca(OH)2" },
    { "a": "CuO", "b": "H2SO4" }, { "a": "CuO", "b": "HCl" }, { "a": "CuO", "b": "HNO3" },
    { "a": "Fe2O3", "b": "H2SO4" }, { "a": "Fe2O3", "b": "HCl" }, { "a": "Fe2O3", "b": "HNO3" },
    { "a": "MgO", "b": "HCl" }, { "a": "MgO", "b": "H2SO4" }, { "a": "MgO", "b": "HNO3" },
    { "a": "CaO", "b": "HCl" }, { "a": "CaO", "b": "HNO3" },
    { "a": "Na2O", "b": "HCl" }, { "a": "Na2O", "b": "H2SO4" },
    { "a": "CaO", "b": "CO2" }, { "a": "CaO", "b": "SO2" }, { "a": "CaO", "b": "SO3" },
    { "a": "MgO", "b": "CO2" }, { "a": "MgO", "b": "SO2" },
    { "a": "BaO", "b": "CO2" }, { "a": "BaO", "b": "SO2" }, { "a": "BaO", "b": "SO3" },
    { "a": "Na2O", "b": "CO2" }, { "a": "Na2O", "b": "SO2" },
    { "a": "H3PO4", "b": "KOH" }, { "a": "H3PO4", "b": "Ca(OH)2" },
    { "a": "Cl2", "b": "NaBr" }, { "a": "Cl2", "b": "NaI" }, { "a": "Cl2", "b": "KBr" }, { "a": "Cl2", "b": "KI" },
    { "a": "Br2", "b": "NaI" }, { "a": "Br2", "b": "KI" },
    { "a": "Cl2", "b": "H2O" }, { "a": "Br2", "b": "H2O" }, { "a": "F2", "b": "H2O" },
    { "a": "Cl2", "b": "NaOH" }, { "a": "Cl2", "b": "Ca(OH)2" },
    { "a": "F2", "b": "H2" }, { "a": "Cl2", "b": "H2" }, { "a": "Br2", "b": "H2" }, { "a": "I2", "b": "H2" },
    { "a": "Na", "b": "Br2" }, { "a": "Na", "b": "I2" },
    { "a": "Fe", "b": "Cl2" }, { "a": "Fe", "b": "Br2" }, { "a": "Cu", "b": "Cl2" },
    { "a": "AgNO3", "b": "NaBr" }, { "a": "AgNO3", "b": "NaI" }, { "a": "AgNO3", "b": "KI" },
    { "a": "NaF", "b": "CaCl2" }, { "a": "HF", "b": "CaO" }, { "a": "HF", "b": "NaOH" },
    { "a": "SO2", "b": "O2" }, { "a": "SO2", "b": "H2O" }, { "a": "H2S", "b": "O2" },
    { "a": "H2S", "b": "CuSO4" }, { "a": "H2S", "b": "AgNO3" }, { "a": "Na2S", "b": "HCl" },
    { "a": "Fe", "b": "S" }, { "a": "Cu", "b": "S" }, { "a": "H2", "b": "S" },
    { "a": "H2SO3", "b": "O2" }, { "a": "H2SO3", "b": "NaOH" }, { "a": "Na2SO3", "b": "S" },
    { "a": "H2S", "b": "SO2" },
    { "a": "Mg", "b": "Br2" }, { "a": "Mg", "b": "I2" },
    { "a": "Al", "b": "Br2" }, { "a": "Al", "b": "I2" },
    { "a": "Fe", "b": "Br2" }, { "a": "Fe", "b": "I2" },
    { "a": "Zn", "b": "Br2" }, { "a": "Zn", "b": "I2" },
    { "a": "Cu", "b": "Br2" },
    { "a": "Ca", "b": "Br2" }, { "a": "Ca", "b": "I2" },
    // 金属与酸的反应
    { "a": "Fe", "b": "H2SO4" }, { "a": "Zn", "b": "H2SO4" }, { "a": "Mg", "b": "H2SO4" }, { "a": "Al", "b": "H2SO4" },
    { "a": "Na", "b": "HCl" }, { "a": "Na", "b": "H2SO4" }, { "a": "Na", "b": "HNO3" },
    { "a": "K", "b": "HCl" }, { "a": "K", "b": "H2SO4" }, { "a": "K", "b": "HNO3" },
    { "a": "Ca", "b": "HCl" }, { "a": "Ca", "b": "H2SO4" }, { "a": "Ca", "b": "HNO3" },
    { "a": "Ba", "b": "HCl" }, { "a": "Ba", "b": "H2SO4" }, { "a": "Ba", "b": "HNO3" },
    { "a": "Cu", "b": "HNO3" }, { "a": "Ag", "b": "HNO3" }, { "a": "Hg", "b": "HNO3" },
    { "a": "Fe", "b": "HNO3" }, { "a": "Zn", "b": "HNO3" }, { "a": "Mg", "b": "HNO3" }, { "a": "Al", "b": "HNO3" },
    { "a": "Mg", "b": "H3PO4" }, { "a": "Zn", "b": "H3PO4" }, { "a": "Fe", "b": "H3PO4" },
    { "a": "Na", "b": "H3PO4" }, { "a": "K", "b": "H3PO4" },
    { "a": "Mg", "b": "HI" }, { "a": "Zn", "b": "HI" }, { "a": "Fe", "b": "HI" },
    { "a": "Mg", "b": "HBr" }, { "a": "Zn", "b": "HBr" }, { "a": "Fe", "b": "HBr" },
    { "a": "Mg", "b": "HF" }, { "a": "Zn", "b": "HF" }, { "a": "Al", "b": "HF" },
    // 更多系统性的活泼金属与酸反应
    { "a": "K", "b": "HI" }, { "a": "K", "b": "HBr" }, { "a": "K", "b": "HF" }, { "a": "K", "b": "H2S" }, { "a": "K", "b": "H2SO3" },
    { "a": "Na", "b": "HI" }, { "a": "Na", "b": "HBr" }, { "a": "Na", "b": "HF" }, { "a": "Na", "b": "H2S" }, { "a": "Na", "b": "H2SO3" },
    { "a": "Mg", "b": "HI" }, { "a": "Mg", "b": "HBr" }, { "a": "Mg", "b": "HF" }, { "a": "Mg", "b": "H2S" }, { "a": "Mg", "b": "H2SO3" },
    { "a": "Al", "b": "HI" }, { "a": "Al", "b": "HBr" }, { "a": "Al", "b": "HF" }, { "a": "Al", "b": "H2S" }, { "a": "Al", "b": "H2SO3" },
    { "a": "Zn", "b": "HI" }, { "a": "Zn", "b": "HBr" }, { "a": "Zn", "b": "HF" }, { "a": "Zn", "b": "H2S" },
    { "a": "Fe", "b": "HI" }, { "a": "Fe", "b": "HBr" }, { "a": "Fe", "b": "H2S" }, { "a": "Fe", "b": "H2SO3" },
    // 大量氧化还原反应
    { "a": "H2O2", "b": "FeSO4" }, { "a": "H2O2", "b": "Na2SO3" }, { "a": "H2O2", "b": "KI" },
    { "a": "H2O2", "b": "H2S" }, { "a": "H2O2", "b": "SO2" },
    { "a": "Cl2", "b": "FeSO4" }, { "a": "Cl2", "b": "Na2SO3" }, { "a": "Cl2", "b": "H2S" },
    { "a": "Cl2", "b": "KI" }, { "a": "Cl2", "b": "NaBr" }, { "a": "Cl2", "b": "FeCl2" },
    { "a": "Br2", "b": "FeSO4" }, { "a": "Br2", "b": "Na2SO3" }, { "a": "Br2", "b": "H2S" },
    { "a": "Br2", "b": "KI" },
    { "a": "FeCl3", "b": "Cu" }, { "a": "FeCl3", "b": "Fe" }, { "a": "FeCl3", "b": "KI" },
    { "a": "FeCl3", "b": "H2S" }, { "a": "FeCl3", "b": "Na2SO3" },
    { "a": "HNO3", "b": "Cu" }, { "a": "HNO3", "b": "Fe" }, { "a": "HNO3", "b": "C" },
    { "a": "HNO3", "b": "S" }, { "a": "HNO3", "b": "P" }, { "a": "HNO3", "b": "FeO" },
    { "a": "H2SO4", "b": "Cu" }, { "a": "H2SO4", "b": "C" }, { "a": "H2SO4", "b": "S" },
    { "a": "Na2O2", "b": "H2S" }, { "a": "Na2O2", "b": "SO2" }, { "a": "Na2O2", "b": "FeSO4" },
    { "a": "O2", "b": "SO2" }, { "a": "O2", "b": "NO" }, { "a": "O2", "b": "Fe(OH)2" },
    { "a": "O2", "b": "FeO" }, { "a": "O2", "b": "Cu2O" },
    { "a": "Cl2", "b": "NaOH" }, { "a": "Cl2", "b": "Ca(OH)2" }, { "a": "S", "b": "NaOH" },
    { "a": "NO2", "b": "H2O" }, { "a": "NO2", "b": "NaOH" },
    { "a": "Al", "b": "Fe2O3" }, { "a": "Al", "b": "Fe3O4" }, { "a": "Al", "b": "CuO" },
    { "a": "H2", "b": "CuO" }, { "a": "H2", "b": "Fe2O3" }, { "a": "C", "b": "CuO" },
    { "a": "CO", "b": "CuO" }, { "a": "CO", "b": "Fe2O3" },
    // F 相关反应
    { "a": "F2", "b": "NaCl" }, { "a": "F2", "b": "NaBr" }, { "a": "F2", "b": "NaI" },
    { "a": "F2", "b": "KCl" }, { "a": "F2", "b": "KBr" }, { "a": "F2", "b": "KI" },
    { "a": "F2", "b": "Na" }, { "a": "F2", "b": "K" },
    { "a": "F2", "b": "Mg" }, { "a": "F2", "b": "Ca" }, { "a": "F2", "b": "Ba" },
    { "a": "F2", "b": "Al" }, { "a": "F2", "b": "Fe" }, { "a": "F2", "b": "Cu" },
    { "a": "F2", "b": "Ag" }, { "a": "F2", "b": "Hg" }, { "a": "F2", "b": "Zn" },
    { "a": "F2", "b": "NH3" }, { "a": "F2", "b": "S" }, { "a": "F2", "b": "P" },
    { "a": "F2", "b": "C" }, { "a": "HF", "b": "SiO2" }, { "a": "HF", "b": "KOH" },
    { "a": "HF", "b": "Ba(OH)2" }, { "a": "HF", "b": "MgO" }, { "a": "HF", "b": "Al2O3" },
    { "a": "NaF", "b": "BaCl2" }, { "a": "KF", "b": "CaCl2" },
    // 所有支持元素的氧化反应
    { "a": "Zn", "b": "O2" }, { "a": "Ag", "b": "O2" },
    { "a": "Si", "b": "O2" },
    { "a": "Cu", "b": "O2" }, { "a": "Al", "b": "O2" }, { "a": "Ba", "b": "O2" },
    { "a": "N2", "b": "O2" },
    // 非金属氧化 (新增)
    { "a": "C", "b": "O2" }, { "a": "S", "b": "O2" }, { "a": "P", "b": "O2" },
    { "a": "H2", "b": "O2" }, { "a": "Si", "b": "O2" }, { "a": "N2", "b": "O2" },
    { "a": "SO2", "b": "O2" }, { "a": "NO", "b": "O2" }, { "a": "CO", "b": "O2" },
    { "a": "Cl2", "b": "O2" },
    // 活泼金属与活泼非金属反应 (卤素、硫)
    { "a": "K", "b": "Cl2" }, { "a": "K", "b": "Br2" }, { "a": "K", "b": "I2" }, { "a": "K", "b": "S" },
    { "a": "Na", "b": "S" }, { "a": "Mg", "b": "S" },
    { "a": "Ca", "b": "S" }, { "a": "Ba", "b": "S" }, { "a": "Al", "b": "S" },
    { "a": "Zn", "b": "S" }, { "a": "Ca", "b": "Cl2" }, { "a": "Ba", "b": "Cl2" },
    { "a": "Zn", "b": "Cl2" }, { "a": "Ag", "b": "Cl2" }, { "a": "Hg", "b": "Cl2" }, { "a": "Cu", "b": "Cl2" },
    { "a": "Ba", "b": "Br2" }, { "a": "Ba", "b": "I2" }, { "a": "Hg", "b": "Br2" }, { "a": "Hg", "b": "I2" },
    { "a": "Ag", "b": "Br2" }, { "a": "Ag", "b": "I2" }, { "a": "Cu", "b": "I2" },
    { "a": "P", "b": "Cl2" }, { "a": "P", "b": "S" },
    // 活泼金属与酸反应 (扩展)
    { "a": "K", "b": "HCl" }, { "a": "K", "b": "H2SO4" }, { "a": "K", "b": "HNO3" }, { "a": "K", "b": "H3PO4" },
    { "a": "Na", "b": "HCl" }, { "a": "Na", "b": "H2SO4" }, { "a": "Na", "b": "HNO3" }, { "a": "Na", "b": "H3PO4" },
    { "a": "Ca", "b": "HCl" }, { "a": "Ca", "b": "H2SO4" }, { "a": "Ca", "b": "HNO3" }, { "a": "Ca", "b": "H3PO4" },
    { "a": "Mg", "b": "HCl" }, { "a": "Mg", "b": "H2SO4" }, { "a": "Mg", "b": "HNO3" }, { "a": "Mg", "b": "H3PO4" },
    { "a": "Al", "b": "HCl" }, { "a": "Al", "b": "H2SO4" }, { "a": "Al", "b": "HNO3" }, { "a": "Al", "b": "H3PO4" },
    { "a": "Zn", "b": "HCl" }, { "a": "Zn", "b": "H2SO4" }, { "a": "Zn", "b": "HNO3" },
    { "a": "Fe", "b": "HCl" }, { "a": "Fe", "b": "H2SO4" }, { "a": "Fe", "b": "HNO3" },
    { "a": "Ba", "b": "HCl" }, { "a": "Ba", "b": "H2SO4" }, { "a": "Ba", "b": "HNO3" },
    { "a": "K", "b": "HI" }, { "a": "K", "b": "HBr" }, { "a": "K", "b": "HF" }, { "a": "K", "b": "H2S" }, { "a": "K", "b": "H2SO3" },
    { "a": "Na", "b": "HI" }, { "a": "Na", "b": "HBr" }, { "a": "Na", "b": "HF" }, { "a": "Na", "b": "H2S" }, { "a": "Na", "b": "H2SO3" },
    { "a": "Mg", "b": "HI" }, { "a": "Mg", "b": "HBr" }, { "a": "Mg", "b": "HF" }, { "a": "Mg", "b": "H2S" }, { "a": "Mg", "b": "H2SO3" },
    { "a": "Al", "b": "HI" }, { "a": "Al", "b": "HBr" }, { "a": "Al", "b": "HF" }, { "a": "Al", "b": "H2S" }, { "a": "Al", "b": "H2SO3" },
    { "a": "Zn", "b": "HI" }, { "a": "Zn", "b": "HBr" }, { "a": "Zn", "b": "HF" }, { "a": "Zn", "b": "H2S" },
    { "a": "Fe", "b": "HI" }, { "a": "Fe", "b": "HBr" }, { "a": "Fe", "b": "H2S" }, { "a": "Fe", "b": "H2SO3" },
    // 置换反应 (金属置换金属)
    { "a": "Mg", "b": "ZnCl2" }, { "a": "Mg", "b": "ZnSO4" }, { "a": "Mg", "b": "FeCl2" }, { "a": "Mg", "b": "FeSO4" },
    { "a": "Mg", "b": "CuCl2" }, { "a": "Mg", "b": "CuSO4" }, { "a": "Mg", "b": "AlCl3" }, { "a": "Mg", "b": "Al2(SO4)3" },
    { "a": "Zn", "b": "FeCl2" }, { "a": "Zn", "b": "FeSO4" }, { "a": "Zn", "b": "CuCl2" }, { "a": "Zn", "b": "CuSO4" },
    { "a": "Fe", "b": "CuCl2" }, { "a": "Fe", "b": "CuSO4" }, { "a": "Fe", "b": "AgNO3" },
    { "a": "Al", "b": "ZnCl2" }, { "a": "Al", "b": "FeCl2" }, { "a": "Al", "b": "CuCl2" },
    { "a": "Cu", "b": "AgNO3" }
];

let deck = [], players = [{id:0,name:'玩家',hand:[]},{id:1,name:'电脑 1',hand:[]},{id:2,name:'电脑 2',hand:[]}], cur = 0, dir = 1, last = null, anyCardMode = false, isAuWaiting = false, waitingForName = '', stagnantTurns = 0, aiFastMode = false;
let timerInterval, timeLeft = 30;
let totalTurns = 1, gameStartTime = 0, durationInterval;
let aiThinkingTimeout;

// 成就追踪数据
let matchStats = {
    synthesisCount: 0,
    specialCardsUsed: 0,
    elementsUsed: new Set(),
    consecutiveNoDraw: 0,
    maxConsecutiveNoDraw: 0,
    nobleGases: new Set(),
    eggsTriggered: new Set()
};

const ACHIEVEMENTS_DB = [
    { id: 'win', name: '首战告捷', desc: '赢得一场化学竞赛', icon: '🧪' },
    { id: 'fast', name: '科研神速', desc: '在 20 回合内获胜', icon: '⚡' },
    { id: 'synth', name: '合成大师', desc: '单局累计合成 5 次', icon: '🔬' },
    { id: 'diverse', name: '元素周期领主', desc: '合成涉及 5 种不同元素', icon: '🌌' },
    { id: 'noble', name: '稀有气体之光', desc: '使用 3 种及以上稀有气体牌', icon: '✨' },
    { id: 'streak', name: '灵感如潮', desc: '连续 5 回合未摸牌获胜', icon: '🔥' },
    // 隐藏彩蛋成就
    { id: 'egg_heisenberg', name: '绝命毒师', desc: '以 Heisenberg 之名开启实验', icon: '🎩', hidden: true },
    { id: 'egg_lavoisier', name: '拉瓦锡的信徒', desc: '向近代化学之父致敬', icon: '⚖️', hidden: true },
    { id: 'egg_alchemist', name: '禁忌炼金术', desc: '秘密掌握了物质转化的捷径', icon: '🔱', hidden: true },
    { id: 'egg_kinetics', name: '反应动力学专家', desc: '成功干扰了实验室的时间流速', icon: '🕒', hidden: true }
];

const THINKING_MESSAGES = {
    short: ["这一步显而易见...", "让我想想...", "这种简单的反应，瞬间就能计算出来。", "不需要查资料，我记得它的性质。"],
    medium: ["这个电子得失守恒吗？", "我正在脑海里配平这个方程式...", "这种组合并不常见，我得确认一下...", "他在期待我出错吗？天真。"],
    long: ["我翻阅了整本《化学手册》，似乎找到了关键...", "这是一个深奥的过程，我在思考量子力学的解释...", "让我想想拉瓦锡对此会怎么说...", "这个化学平衡移动的方向... 我需要一些时间来确认。"]
};

const NICKNAME_EGGS = {
    'HEISENBERG': { title: '传奇导师 (Breaking Bad)', msg: 'Say my name. 你在化学界拥有不可动摇的地位。' },
    'WALTER WHITE': { title: '传奇导师 (Breaking Bad)', msg: 'You are the one who knocks! 这批货（反应）完美无瑕。' },
    'LAVOISIER': { title: '近代化学之父', msg: '物质守恒定律在你的策略中体现得淋漓尽致。' },
    '拉瓦锡': { title: '近代化学之父', msg: '物质守恒定律在你的策略中体现得淋漓尽致。' },
    'CURIE': { title: '放射性拓荒者', msg: '你的智慧比镭还要耀眼。' },
    '居里夫人': { title: '放射性拓荒者', msg: '你的智慧比镭还要耀眼。' },
    'NOBEL': { title: '诺贝尔奖得主', msg: '这项实验成果足以让你载入教科书。' },
    'DALTON': { title: '原子论奠基人', msg: '你看到了物质最本质的结构，并赢得了比赛。' }
};

function showAiBubble(name, delay) {
    const bubble = document.getElementById('ai-bubble');
    const nameEl = document.getElementById('bubble-name');
    const textEl = document.getElementById('bubble-text');
    const badge = document.getElementById(`ai-badge-${cur}`);
    
    nameEl.textContent = `${name} 正在思考...`;
    let pool = THINKING_MESSAGES.medium;
    if (delay < 8000) pool = THINKING_MESSAGES.short;
    else if (delay > 14000) pool = THINKING_MESSAGES.long;
    
    textEl.textContent = pool[Math.floor(Math.random() * pool.length)];
    
    if (badge) {
        const badgeRect = badge.getBoundingClientRect();
        const containerRect = document.getElementById('game-container').getBoundingClientRect();
        const bubbleWidth = 220; // 与 CSS 中的 width 一致
        
        // 计算泡泡左边缘位置 (初始尝试居中于 badge)
        let targetLeft = (badgeRect.left - containerRect.left + badgeRect.width / 2) - bubbleWidth / 2;
        
        // 边界检测：防止超出右侧
        if (targetLeft + bubbleWidth > containerRect.width - 20) {
            targetLeft = containerRect.width - bubbleWidth - 20;
        }
        // 边界检测：防止超出左侧
        if (targetLeft < 20) {
            targetLeft = 20;
        }

        bubble.style.left = `${targetLeft}px`;
        bubble.style.top = `${badgeRect.bottom - containerRect.top + 10}px`;
        
        // 动态调整小箭头偏移，使其依然指向 Badge 中心
        // arrowX 是 badge 中心点相对于泡泡左边缘的偏移
        const badgeCenterInContainer = badgeRect.left - containerRect.left + badgeRect.width / 2;
        const arrowX = badgeCenterInContainer - targetLeft;
        bubble.style.setProperty('--arrow-x', `${arrowX}px`);
    }
    
    bubble.style.display = 'block';
}

function showUserAgreement() {
    document.getElementById('agreement-modal').style.display = 'flex';
}

function hideUserAgreement() {
    document.getElementById('agreement-modal').style.display = 'none';
}

function updateDurationUI() {
    const now = Date.now();
    const diff = Math.floor((now - gameStartTime) / 1000);
    const m = Math.floor(diff / 60).toString().padStart(2, '0');
    const s = (diff % 60).toString().padStart(2, '0');
    document.getElementById('stat-duration').textContent = `${m}:${s}`;
}

function startGame() {
    const nickname = document.getElementById('nickname-input').value.trim() || '玩家';
    players[0].name = nickname;
    
    // 触发昵称彩蛋
    const upperNick = nickname.toUpperCase();
    if (['HEISENBERG', 'WALTER WHITE'].includes(upperNick)) {
        triggerEggAchievement('egg_heisenberg');
    } else if (['LAVOISIER', '拉瓦锡'].includes(upperNick)) {
        triggerEggAchievement('egg_lavoisier');
    }

    // 随机分配 AI 姓名（不重复）
    const aiNames = [...CHEMIST_NAMES].sort(() => Math.random() - 0.5);
    players[1].name = aiNames[0];
    players[2].name = aiNames[1];

    document.getElementById('lobby').style.display = 'none';
    
    // 重置游戏状态
    cur = 0;
    dir = 1;
    anyCardMode = false;
    isAuWaiting = false;
    stagnantTurns = 0;
    totalTurns = 1;
    gameStartTime = Date.now();
    if (durationInterval) clearInterval(durationInterval);
    durationInterval = setInterval(updateDurationUI, 1000);
    updateDurationUI();

    // 重置成就数据
    matchStats = {
        synthesisCount: 0,
        specialCardsUsed: 0,
        elementsUsed: new Set(),
        consecutiveNoDraw: 0,
        maxConsecutiveNoDraw: 0,
        nobleGases: new Set(),
        eggsTriggered: new Set()
    };
    
    deck = [];
    for(let i=0; i<12; i++) { deck.push({id:'H'+i,t:'H'}); deck.push({id:'O'+i,t:'O'}); }
    Object.keys(ELEMENTS_DATA).forEach(e => { if(e!=='H'&&e!=='O') for(let i=0; i<4; i++) deck.push({id:e+i,t:e}); });
    Object.keys(SPECIAL_CARDS).forEach(s => { for(let i=0; i<4; i++) deck.push({id:s+i,t:s}); });
    deck.sort(() => Math.random() - 0.5);
    
    players.forEach(p => p.hand = deck.splice(0, 10));
    
    // 弹出初始底牌选择框
    showStartCardModal();
}

function showStartCardModal() {
    const modal = document.getElementById('start-card-modal');
    const grid = document.getElementById('start-card-options');
    grid.innerHTML = '';
    modal.style.display = 'flex';

    // 随机选出 8 种物质作为备选底牌
    const candidates = [];
    const indices = new Set();
    while(indices.size < Math.min(8, COMPOUNDS_DB.length)) {
        indices.add(Math.floor(Math.random() * COMPOUNDS_DB.length));
    }
    indices.forEach(idx => candidates.push(COMPOUNDS_DB[idx]));

    candidates.forEach(c => {
        const btn = document.createElement('button');
        btn.className = 'compound-btn';
        btn.innerHTML = `<span class="formula">${formatFormula(c.formula)}</span><span class="name">${c.name}</span>`;
        btn.onclick = () => {
            last = c;
            modal.style.display = 'none';
            initGameUI();
        };
        grid.appendChild(btn);
    });
}

function initGameUI() {
    document.getElementById('game-container').style.display = 'flex';
    resetTimer();
    updateUI();
}

function resetTimer() {
    clearInterval(timerInterval);
    timeLeft = 30;
    updateTimerUI();
    timerInterval = setInterval(() => {
        timeLeft--;
        updateTimerUI();
        if (timeLeft <= 0) {
            clearInterval(timerInterval);
            handleTimeout();
        }
    }, 1000);
}

function updateTimerUI() {
    const el = document.getElementById('countdown');
    if (el) {
        el.textContent = `⏲️ ${timeLeft}s`;
        el.style.color = timeLeft <= 10 ? '#ef4444' : '#475569';
    }
}

function handleTimeout() {
    if (cur === 0) {
        log("由于长时间未操作，自动摸牌罚并跳过回合");
        playerDraw();
    } else {
        nextTurn();
    }
}


function updateUI() {
    // 更新中心区域状态
    const desktopLabel = document.getElementById('desktop-label');
    const substanceBox = document.getElementById('current-substance-box');
    const waitingBox = document.getElementById('waiting-box');
    
    if (isAuWaiting) {
        desktopLabel.style.display = 'none';
        substanceBox.style.display = 'none';
        waitingBox.style.display = 'block';
        document.getElementById('waiting-player-name').textContent = waitingForName;
    } else {
        desktopLabel.style.display = 'block';
        substanceBox.style.display = 'block';
        waitingBox.style.display = 'none';
    }

    // 更新侧边统计面板
    const roundCount = Math.ceil(totalTurns / 3);
    document.getElementById('stat-turns').textContent = roundCount;
    const nextIdx = (cur + dir + 3) % 3;
    document.getElementById('stat-next-player').textContent = players[nextIdx].name;

    document.getElementById('formula-display').innerHTML = formatFormula(last.formula);
    document.getElementById('name-display').textContent = last.name;
    document.getElementById('hand-count').textContent = `我的手牌 (${players[0].hand.length})`;
    document.getElementById('status-dir').textContent = `方向: ${dir===1?'➡️':'⬅️'}`;
    
    // 下方控制台状态
    const consoleEl = document.getElementById('control-console');
    if (cur === 0) {
        consoleEl.style.opacity = '1';
        consoleEl.style.pointerEvents = 'auto';
    } else {
        consoleEl.style.opacity = '0.5';
        consoleEl.style.pointerEvents = 'none';
    }

    // AI 状态
    document.getElementById('ai-status').innerHTML = players.slice(1).map(p => `
        <div id="ai-badge-${p.id}" class="status-badge" style="${cur===p.id?'border: 1px solid #3b82f6; background: #eff6ff;':''}">
            ${p.name}: ${p.hand.length}张
        </div>
    `).join('');

    // 手牌渲染
    const cont = document.getElementById('hand-container');
    cont.innerHTML = '';
    players[0].hand.forEach(c => {
        const d = document.createElement('div');
        const info = ELEMENTS_DATA[c.t] || SPECIAL_CARDS[c.t];
        d.className = `card ${info.bg || ''}`;
        if(cur !== 0) d.style.opacity = '0.5';
        d.innerHTML = `<strong>${c.t}</strong><div style="font-size:0.6rem;margin-top:4px;">${info.name}</div>`;
        d.onclick = () => { 
            if(cur !== 0) return;
            if(SPECIAL_CARDS[c.t]) playSpecial(0, c);
        };
        cont.appendChild(d);
    });

    // 提示区：支持点击合成，且只显示前 3 个可能的选项
    const hints = document.getElementById('hints');
    hints.innerHTML = '';
    if(cur === 0) {
        const possible = getPossible(players[0].hand);
        const partial = possible.slice(0, 3); // 只显示前 3 个提示
        partial.forEach(p => {
            const b = document.createElement('div');
            b.className = 'compound-btn';
            b.innerHTML = `<span>${p.name}</span><small style="opacity:0.6;">${formatFormula(p.formula)}</small>`;
            b.onclick = () => {
                const consumed = [];
                const tempHand = [...players[0].hand];
                Object.keys(p.composition).forEach(el => {
                    const j = tempHand.findIndex(x => x.t === el);
                    if (j > -1) {
                        consumed.push(tempHand[j].id);
                        tempHand.splice(j, 1);
                    }
                });
                playSub(0, p, consumed);
            };
            hints.appendChild(b);
        });
        if(possible.length === 0) hints.innerHTML = '<span style="color:#aaa;font-style:italic;">暂无可进行的合成，需摸牌或出特殊牌</span>';
        else if(possible.length > 3) {
            const more = document.createElement('small');
            more.style.color = '#94a3b8';
            more.style.marginLeft = '10px';
            more.textContent = `...及其他 ${possible.length - 3} 种可能`;
            hints.appendChild(more);
        }
    }

    // 更新摸牌按钮文本
    const drawBtn = document.querySelector('.btn-warning');
    drawBtn.textContent = '没有牌？摸 2 张';
    drawBtn.style.background = '#f59e0b';
}

function getPossible(hand) {
    // 无需考虑系数：只需检查手牌中是否包含组成物质的所有种类元素
    const availableElements = new Set(hand.map(c => c.t));
    return COMPOUNDS_DB.filter(c => 
        Object.keys(c.composition).every(el => availableElements.has(el))
    ).filter(c => canReact(c, last));
}

// --- 离子反应逻辑 ---
function getIons(comp) {
    const f = comp.formula;
    if (!f) return null;

    // 1. 特殊处理酸
    if (f === 'HCl') return { c: 'H+', a: 'Cl' };
    if (f === 'H2SO4') return { c: 'H+', a: 'SO4' };
    if (f === 'HNO3') return { c: 'H+', a: 'NO3' };
    if (f === 'H3PO4') return { c: 'H+', a: 'PO4' };
    if (f === 'HI') return { c: 'H+', a: 'I' };
    if (f === 'HBr') return { c: 'H+', a: 'Br' };
    if (f === 'HF') return { c: 'H+', a: 'F' };
    if (f === 'HClO') return { c: 'H+', a: 'ClO' };
    if (f === 'H2S') return { c: 'H+', a: 'S' };
    if (f === 'H2SO3') return { c: 'H+', a: 'SO3' };

    // 2. 处理常见的阳离子和阴离子原子团
    const cations = ['Na', 'K', 'NH4', 'Ag', 'Ca', 'Ba', 'Mg', 'Al', 'Zn', 'Fe', 'Cu', 'Hg'];
    const anions = ['OH', 'NO3', 'SO4', 'CO3', 'HCO3', 'Cl', 'PO4', 'I', 'Br', 'SO3', 'HSO4', 'HSO3', 'NO2', 'F', 'ClO', 'S'];

    let cation = '', anion = '';
    for (let c of cations) { if (f.startsWith(c)) { cation = c; break; } }
    for (let a of anions) { if (f.endsWith(a) || f.includes(a + ')')) { anion = a; break; } }
    
    // 特殊结尾处理
    if (!anion && f.endsWith('Cl')) anion = 'Cl';
    if (!anion && f.endsWith('I')) anion = 'I';
    if (!anion && f.endsWith('Br')) anion = 'Br';
    if (!anion && f.endsWith('F')) anion = 'F';
    if (!anion && f.endsWith('S')) anion = 'S';

    if (cation && anion) return { c: cation, a: anion };
    return null;
}

const PRECIPITATES = new Set(['AgCl', 'AgBr', 'AgI', 'Ag2CO3', 'Ag3PO4', 'BaSO4', 'BaCO3', 'Ba3(PO4)2', 'CaCO3', 'Ca3(PO4)2', 'Mg(OH)2', 'MgCO3', 'Cu(OH)2', 'CuS', 'Fe(OH)3', 'Fe(OH)2', 'FeS', 'Al(OH)3', 'Zn(OH)2', 'ZnS', 'HgS', 'BaSO3', 'CaSO3', 'Ag2SO4', 'CaF2', 'Ag2S']);

function canReact(a, b) {
    console.log(`Checking canReact for: ${a?.formula} vs ${b?.formula} (AnyCardMode: ${anyCardMode})`);
    if(!a || !b) return true;
    if(anyCardMode) return true;
    
    // 1. 基础数据库匹配 (氧化还原、置换等)
    if (REACTIONS_DB.some(r => (r.a===a.formula && r.b===b.formula) || (r.a===b.formula && r.b===a.formula))) {
        console.log("canReact: Matched in REACTIONS_DB");
        return true;
    }

    // 2. 氧气通配：大部分单质和化合物都能与氧气反应（燃烧或缓慢氧化）
    if (a.formula === 'O2' || b.formula === 'O2') {
        const other = a.formula === 'O2' ? b : a;
        // 除了极个别物质（如金、铂、氦氖氩氪等惰性气体，已经在 SPECIAL_CARDS 处理了，以及部分高价氧化物）
        const NO_O2_REACTION = ['CO2', 'SO3', 'Al2O3', 'SiO2', 'H2O', 'BaSO4', 'H2SO4'];
        if (!NO_O2_REACTION.includes(other.formula)) {
            console.log("canReact: Matched via Oxygen wildcard");
            return true;
        }
    }

    // 3. 离子反应判断
    const iA = getIons(a), iB = getIons(b);
    if (iA && iB) {
        const cats = [iA.c, iB.c], anis = [iA.a, iB.a];

        // 生成水 (H+ + OH-)
        if (cats.includes('H+') && anis.includes('OH')) {
            console.log("canReact: Matched Ion reaction (Water formation)");
            return true;
        }
        // 生成气体 (H+ + CO3/HCO3/SO3/HSO3/S 或 OH- + NH4)
        if (cats.includes('H+') && (anis.includes('CO3') || anis.includes('HCO3') || anis.includes('SO3') || anis.includes('HSO3') || anis.includes('S'))) {
            console.log("canReact: Matched Ion reaction (Gas formation - Acid)");
            return true;
        }
        if (anis.includes('OH') && cats.includes('NH4')) {
            console.log("canReact: Matched Ion reaction (Gas formation - Ammonia)");
            return true;
        }

        // 生成沉淀 (检查交叉组合)
        const check = (c, n) => {
            const combinations = [c+n, c+'2'+n, c+'3'+n, c+'2('+n+')3', c+'('+n+')2', c+'('+n+')3'];
            return combinations.some(combined => PRECIPITATES.has(combined));
        };
        if (check(iA.c, iB.a) || check(iB.c, iA.a)) {
            console.log("canReact: Matched Ion reaction (Precipitate formation)");
            return true;
        }
    }
    console.log("canReact: NO MATCH found");
    return false;
}
function playSub(pIdx, comp, ids) {
    // 在更新 last 之前，这其实就是一次反应过程
    console.log(`playSub: Calling showReactionEquation with last=${last?.formula} and comp=${comp?.formula}`);
    showReactionEquation(last, comp);
    
    // 成就统计
    if (pIdx === 0) {
        matchStats.synthesisCount++;
        if (comp.composition) {
            Object.keys(comp.composition).forEach(el => matchStats.elementsUsed.add(el));
        }
        matchStats.consecutiveNoDraw++;
        matchStats.maxConsecutiveNoDraw = Math.max(matchStats.maxConsecutiveNoDraw, matchStats.consecutiveNoDraw);
    }

    players[pIdx].hand = players[pIdx].hand.filter(c => !ids.includes(c.id));
    last = comp;
    anyCardMode = false;
    isAuWaiting = false;
    stagnantTurns = 0;
    log(`${players[pIdx].name} 合成了 ${comp.name}`);
    if(checkWin(pIdx)) return;
    nextTurn();
}

function playSpecial(pIdx, card) {
    const eff = SPECIAL_CARDS[card.t].eff;
    
    // 成就统计
    if (pIdx === 0) {
        matchStats.specialCardsUsed++;
        if (['He', 'Ne', 'Ar', 'Kr'].includes(card.t)) {
            matchStats.nobleGases.add(card.t);
        }
        matchStats.consecutiveNoDraw++;
        matchStats.maxConsecutiveNoDraw = Math.max(matchStats.maxConsecutiveNoDraw, matchStats.consecutiveNoDraw);
    }

    // 记录当前的 last，因为 playSub 或其他逻辑可能需要它来展示反应
    const prevLast = last;

    players[pIdx].hand = players[pIdx].hand.filter(c => c.id !== card.id);
    let nxt = (cur + dir + 3) % 3;
    stagnantTurns = 0;
    
    if(eff==='reverse') {
        dir *= -1;
        anyCardMode = false;
        log(`游戏方向已反转！`);
    } else if(eff==='skip') {
        anyCardMode = true;
        const skipped = nxt;
        nxt = (skipped + dir + 3) % 3; // 跳过一位
        isAuWaiting = true;
        waitingForName = players[nxt].name;
        log(`金牌 (Au) 激活！${players[skipped].name} 被跳过，${players[nxt].name} 进入任意反应模式`);
    }
    
    if(eff !== 'skip') anyCardMode = false;
    
    // 如果是特殊牌，通常它本身不参与化学反应显示，
    // 但如果我们需要在打出后依然显示方程式（例如某些特殊交互），可以在此处理。
    // 目前 last 保持不变，除非打出了具体的物质卡。

    if(checkWin(pIdx)) return;
    cur = nxt;
    resetTimer();
    updateUI();
    if(players[cur].id !== 0) setTimeout(aiTurn, 1500);
}

function playerDraw() {
    if (cur !== 0) return;
    const num = 2;
    players[0].hand.push(...deck.splice(0, num));
    log(`您摸了 ${num} 张牌`);
    anyCardMode = false;
    isAuWaiting = false;
    stagnantTurns++;
    matchStats.consecutiveNoDraw = 0; // 重置连续不出牌统计
    nextTurn();
}

function setInputMsg(m, isErr = false) {
    const el = document.getElementById('input-msg');
    el.textContent = m;
    el.style.color = isErr ? '#ef4444' : '#64748b';
    if(isErr) setTimeout(() => setInputMsg('输入名称后点击合成'), 3000);
}

function handleManualSubmit() {
    if (cur !== 0) {
        setInputMsg("现在不是您的回合", true);
        return;
    }
    const inputEl = document.getElementById('substance-input');
    const rawName = inputEl.value.trim().toUpperCase();
    if (!rawName) return;

    // --- 实验室彩蛋系统 ---
    if (rawName === 'HEISENBERG') {
        triggerEggAchievement('egg_heisenberg');
        log("海森堡：不确定性原理！你与一位随机化学家交换了手牌。");
        const target = Math.floor(Math.random() * 2) + 1;
        [players[0].hand, players[target].hand] = [players[target].hand, players[0].hand];
        inputEl.value = ''; updateUI(); return;
    }
    if (rawName === 'NOBEL') {
        log("诺贝尔：炸药奖发放！除你以外所有人强制摸 3 张牌。");
        players.slice(1).forEach(p => p.hand.push(...deck.splice(0, 3)));
        inputEl.value = ''; updateUI(); return;
    }
    if (rawName === 'ALCHEMY') {
        triggerEggAchievement('egg_alchemist');
        log("炼金术：你的前三张牌已转化为纯金(Au)！");
        for(let i=0; i<Math.min(3, players[0].hand.length); i++) {
            players[0].hand[i] = {id:'Au_egg_'+Date.now()+i, t:'Au'};
        }
        inputEl.value = ''; updateUI(); return;
    }
    if (rawName === 'KINETICS') {
        triggerEggAchievement('egg_kinetics');
        aiFastMode = !aiFastMode;
        log(aiFastMode ? "动力学系统：实验室催化剂已就位，AI 行动将大幅加速！" : "动力学系统：移除催化剂，AI 行动恢复正常速度。");
        inputEl.value = ''; return;
    }

    // 查找物质
    const comp = COMPOUNDS_DB.find(c => c.formula.toUpperCase() === rawName || c.name === rawName);
    if (!comp) {
        setInputMsg(`找不到物质: "${rawName}"`, true);
        return;
    }

    // 1. 检查手牌是否拥有所需元素
    const availableElements = new Set(players[0].hand.map(c => c.t));
    const hasElements = Object.keys(comp.composition).every(el => availableElements.has(el));
    
    if (!hasElements) {
        setInputMsg(`缺少必要元素牌`, true);
        return;
    }

    // 2. 检查是否能与当前物质反应
    if (!canReact(comp, last)) {
        setInputMsg(`${comp.name} 无法与当前物质反应`, true);
        return;
    }

    // 3. 执行合成
    const consumed = [];
    const tempHand = [...players[0].hand];
    Object.keys(comp.composition).forEach(el => {
        const j = tempHand.findIndex(x => x.t === el);
        if (j > -1) {
            consumed.push(tempHand[j].id);
            tempHand.splice(j, 1);
        }
    });
    
    inputEl.value = ''; // 清空输入
    playSub(0, comp, consumed);
}

function nextTurn() {
    document.getElementById('ai-bubble').style.display = 'none';
    if (aiThinkingTimeout) clearTimeout(aiThinkingTimeout);

    if (stagnantTurns >= 3) {
        anyCardMode = true;
        stagnantTurns = 0;
        log("陷入实验室僵局！化学反应壁垒暂时消失，下一玩家可随意出牌。");
    }
    cur = (cur + dir + 3) % 3;
    totalTurns++;
    resetTimer();
    updateUI();
    if(players[cur].id !== 0) {
        let aiDelay = Math.floor(Math.random() * (20000 - 3000 + 1)) + 3000;
        if (aiFastMode) aiDelay = 800; // 加速模式下固定为 0.8s
        
        aiThinkingTimeout = setTimeout(() => {
            showAiBubble(players[cur].name, aiDelay);
        }, aiDelay / 3);
        setTimeout(aiTurn, aiDelay);
    }
}

function aiTurn() {
    const p = players[cur];
    
    // 优先功能牌
    const sp = p.hand.find(c => SPECIAL_CARDS[c.t]);
    if(sp) { playSpecial(cur, sp); return; }
    // 尝试合成
    const poss = getPossible(p.hand);
    if(poss.length > 0) {
        const c = poss[0], ids = [];
        const tmp = [...p.hand];
        // 无需考虑系数：AI 也是每种元素只消耗一张
        Object.keys(c.composition).forEach(el => {
            const j = tmp.findIndex(x => x.t === el);
            if (j > -1) {
                ids.push(tmp[j].id);
                tmp.splice(j, 1);
            }
        });
        playSub(cur, c, ids);
    } else {
        p.hand.push(...deck.splice(0,2));
        log(`${p.name} 无法出牌，摸了 2 张`);
        stagnantTurns++;
        nextTurn();
    }
}

function checkWin(i) {
    if(players[i].hand.length === 0) {
        showVictory(players[i].name);
        return true;
    }
    return false;
}

// --- 对局保存与复现 ---
function exportGame() {
    const state = {
        v: 1, // 版本号
        p: players,
        c: cur,
        d: dir,
        lf: last?.formula,
        am: anyCardMode,
        aup: isAuWaiting,
        aun: waitingForName,
        st: stagnantTurns,
        tt: totalTurns,
        af: aiFastMode,
        gs: gameStartTime,
        dk: deck
    };
    
    // 使用 encodeURIComponent 处理中文字符后再 base64
    const jsonStr = JSON.stringify(state);
    const base64 = btoa(encodeURIComponent(jsonStr));
    
    // 校验位计算 (取所有字符编码总和的后3位)
    let sum = 0;
    for (let i = 0; i < base64.length; i++) sum += base64.charCodeAt(i);
    const checksum = (sum % 1000).toString().padStart(3, '0');
    
    const finalStr = base64 + checksum;
    
    // 使用自定义弹窗 (修正 ID 匹配)
    const textarea = document.getElementById('experiment-data-text');
    if (textarea) {
        textarea.value = finalStr;
    }
    document.getElementById('save-experiment-modal').style.display = 'flex';
}

function copyExperimentCode(e) {
    const area = document.getElementById('save-code-area');
    area.select();
    document.execCommand('copy');
    const btn = e.target;
    const originText = btn.innerText;
    btn.innerText = "已复制！";
    btn.style.background = "#22c55e"; 
    setTimeout(() => {
        btn.innerText = originText;
        btn.style.background = "";
    }, 2000);
}

function importGame() {
    document.getElementById('load-code-area').value = "";
    document.getElementById('load-experiment-modal').style.display = 'flex';
}

function processImportGame() {
    const code = document.getElementById('load-code-area').value.trim();
    if (!code) {
        document.getElementById('load-experiment-modal').style.display = 'none';
        return;
    }
    if (code.length < 4) {
        showMessageModal("实验复现失败", "实验代码长度不正确，请检查是否完整复制。");
        return;
    }
    
    const checksumStr = code.slice(-3);
    const base64 = code.slice(0, -3);
    
    // 校验
    let sum = 0;
    for (let i = 0; i < base64.length; i++) sum += base64.charCodeAt(i);
    const calculatedChecksum = (sum % 1000).toString().padStart(3, '0');
    
    if (checksumStr !== calculatedChecksum) {
        showMessageModal("实验复现失败", "实验代码校验失败，请确保代码输入完整且正确！");
        return;
    }
    
    try {
        const jsonStr = decodeURIComponent(atob(base64));
        const state = JSON.parse(jsonStr);
        
        // 应用状态
        players = state.p;
        cur = state.c;
        dir = state.d;
        anyCardMode = state.am;
        isAuWaiting = state.aup;
        waitingForName = state.aun;
        stagnantTurns = state.st;
        totalTurns = state.tt;
        aiFastMode = state.af;
        gameStartTime = state.gs;
        deck = state.dk;
        
        // 恢复 last 卡片对象
        if (state.lf) {
            last = COMPOUNDS_DB.find(f => f.formula === state.lf);
        } else {
            last = null;
        }

        // 刷新 UI
        document.getElementById('load-experiment-modal').style.display = 'none';
        document.getElementById('lobby').style.display = 'none';
        
        initGameUI();
        
        // 恢复实验时长计时
        if (durationInterval) clearInterval(durationInterval);
        durationInterval = setInterval(updateDurationUI, 1000);
        updateDurationUI();

        log(`<span style="color:#8b5cf6">系统: 实验记录已复现，对局开始。</span>`);
        
        // 处理金(Au)产生的等待状态：如果加载时处于等待期，直接跳过等待
        if (isAuWaiting) {
            isAuWaiting = false;
            const nextTargetIdx = players.findIndex(p => p.name === waitingForName);
            if (nextTargetIdx !== -1) cur = nextTargetIdx;
            resetTimer();
            updateUI();
        }

        // 如果当前是 AI 回合，则启动 AI 行动逻辑
        if (cur !== 0) {
            let aiDelay = Math.floor(Math.random() * (6000 - 2000 + 1)) + 2000;
            if (aiFastMode) aiDelay = 800;
            
            if (aiThinkingTimeout) clearTimeout(aiThinkingTimeout);
            aiThinkingTimeout = setTimeout(() => {
                showAiBubble(players[cur].name, aiDelay);
            }, aiDelay / 3);
            setTimeout(aiTurn, aiDelay);
        }
        
    } catch(e) {
        console.error(e);
        showMessageModal("实验复现失败", "读取失败：代码内容损坏或版本不兼容！");
    }
}

function showVictory(name) {
    clearInterval(timerInterval);
    if (durationInterval) clearInterval(durationInterval);
    const screen = document.getElementById('victory-screen');
    const nameEl = document.getElementById('winner-name');
    const msgEl = document.getElementById('victory-msg');
    const achContainer = document.getElementById('achievements-container');
    
    achContainer.innerHTML = '';
    const isPlayer = (name === players[0].name);

    const egg = NICKNAME_EGGS[name.toUpperCase()];
    if (isPlayer && egg) {
        nameEl.textContent = `${name} (${egg.title}) 胜利！`;
        msgEl.textContent = egg.msg;
        nameEl.style.color = '#2563eb';
    } else {
        nameEl.textContent = `${name} 胜利！`;
        nameEl.style.color = '#1e3a8a';
        msgEl.textContent = isPlayer ? '杰出的科学家！你完美地控制了实验室的所有反应。' : '实验结论：电脑在本次推演中获得了先机。';
    }

    // 计算并保存本次比赛成就
    const history = JSON.parse(localStorage.getItem('chemistry_achievements') || '[]');
    const historySet = new Set(history);

    if (isPlayer) {
        const sessionAchievements = [];
        sessionAchievements.push('win');
        if (totalTurns <= 20) sessionAchievements.push('fast');
        if (matchStats.synthesisCount >= 5) sessionAchievements.push('synth');
        if (matchStats.elementsUsed.size >= 5) sessionAchievements.push('diverse');
        if (matchStats.nobleGases.size >= 3) sessionAchievements.push('noble');
        if (matchStats.maxConsecutiveNoDraw >= 5) sessionAchievements.push('streak');
        
        // 将本次比赛获得的成就合并到历史记录
        sessionAchievements.forEach(id => {
            if (!historySet.has(id)) {
                historySet.add(id);
                history.push(id);
            }
        });
        matchStats.eggsTriggered.forEach(id => {
            if (!historySet.has(id)) {
                historySet.add(id);
                history.push(id);
            }
        });
        localStorage.setItem('chemistry_achievements', JSON.stringify(history));
    }

    // 渲染列表
    ACHIEVEMENTS_DB.forEach((ach) => {
        // 如果是隐藏成就且未获得，则不显示
        if (ach.hidden && !historySet.has(ach.id)) return;
        
        const hasEarned = historySet.has(ach.id);
        const item = document.createElement('div');
        item.className = `achievement-badge ${hasEarned ? 'earned' : ''}`;
        item.innerHTML = `
            <span class="achievement-icon">${ach.icon}</span>
            <span class="achievement-name">${ach.name}</span>
            <span class="achievement-desc">${ach.desc}</span>
        `;
        achContainer.appendChild(item);
    });
    
    screen.style.display = 'flex';
}

function triggerEggAchievement(id) {
    if (matchStats.eggsTriggered.has(id)) return;
    const meta = ACHIEVEMENTS_DB.find(a => a.id === id);
    if (!meta) return;

    matchStats.eggsTriggered.add(id);
    
    // 持久化存储
    const history = JSON.parse(localStorage.getItem('chemistry_achievements') || '[]');
    if (!history.includes(id)) {
        history.push(id);
        localStorage.setItem('chemistry_achievements', JSON.stringify(history));
    }
    
    const popup = document.getElementById('egg-popup');
    const nameDisplay = document.getElementById('egg-name-display');
    const eggIcon = popup.querySelector('.egg-icon');
    if (eggIcon) eggIcon.textContent = meta.icon;
    nameDisplay.textContent = meta.name;
    
    popup.classList.add('show');
    setTimeout(() => {
        popup.classList.remove('show');
    }, 3000);
}

function log(m) {
    // 顶部提示栏支持 HTML 渲染（如下标、颜色等）
    document.getElementById('msg').innerHTML = m;
    const logList = document.getElementById('log-list');
    const item = document.createElement('div');
    item.style.cssText = 'font-size: 0.8rem; color: #475569; margin-bottom: 6px; padding-bottom: 4px; border-bottom: 1px dotted #e2e8f0; animation: fadeInLog 0.3s ease;';
    const time = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    item.innerHTML = `<span style="color: #94a3b8; font-size: 0.7rem;">[${time}]</span> ${m}`;
    logList.appendChild(item);
    
    // 自动滚动到底部以确保实时看到最新内容
    logList.scrollTop = logList.scrollHeight;
    
    // Keep only latest 30 logs
    if (logList.children.length > 30) {
        logList.removeChild(logList.firstChild);
    }
}

// ===== 移动端优化代码 =====
// 1. 防止双击缩放
document.addEventListener('touchstart', function(e) {
    if (e.touches.length > 1) {
        e.preventDefault();
    }
}, { passive: false });

// 2. 禁止长按菜单
document.addEventListener('contextmenu', function(e) {
    if (e.target.tagName !== 'INPUT' && e.target.tagName !== 'TEXTAREA') {
        e.preventDefault();
    }
});

// 3. 优化手机键盘
document.getElementById('substance-input').addEventListener('focus', function() {
    setTimeout(() => {
        this.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }, 500);
});

// 4. 处理软键盘显示/隐藏
let lastHeight = window.innerHeight;
window.addEventListener('resize', function() {
    const currentHeight = window.innerHeight;
    const diff = lastHeight - currentHeight;
    
    if (diff > 100) {
        // 软键盘显示，可能需要调整布局
        document.body.style.overflow = 'hidden';
    } else if (diff < -100) {
        // 软键盘隐藏
        document.body.style.overflow = 'auto';
    }
    lastHeight = currentHeight;
});

// 5. 提升移动端卡片点击精度
const isMobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent);
if (isMobile) {
    // 增加卡片的可点击区域反馈
    document.addEventListener('touchend', function(e) {
        const target = e.target.closest('.card');
        if (target) {
            target.style.transform = '';
        }
    });
}

// 6. 防止iOS下input聚焦时页面缩放
if (/iPhone|iPad|iPod/.test(navigator.userAgent)) {
    document.addEventListener('touchstart', function(e) {
        if (e.target.tagName === 'INPUT') {
            e.target.style.fontSize = '16px';
        }
    });
}

// 7. 隐藏地址栏（仅当在全屏模式下）
if (window.scrollY === 0) {
    setTimeout(() => {
        window.scrollTo(0, 1);
    }, 100);
}
