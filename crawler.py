import os
import requests
import json
import re

# 配置
REPO_OWNER = "open-reaction-database"
REPO_NAME = "ord-data"
BASE_URL = f"https://api.github.com/repos/{REPO_OWNER}/{REPO_NAME}/contents/data"
INDEX_HTML_PATH = "index.html"

# 允许的元素，用于简单的 UNO 规则
ALLOWED_ELEMENTS = {'H', 'O', 'C', 'N', 'F', 'Na', 'Mg', 'Al', 'Si', 'P', 'S', 'Cl', 'K', 'Ca', 'Mn', 'Fe', 'Cu', 'Zn', 'Br', 'I', 'Ag'}

def get_formula_from_smiles(smiles):
    """极简 SMILES 转化学式"""
    clean_smiles = re.sub(r'[\[\]\(\)\-\=\#\.\+\@\\\/0-9]', '', smiles)
    counts = {}
    i = 0
    while i < len(clean_smiles):
        char = clean_smiles[i]
        curr_el = char
        if i + 1 < len(clean_smiles) and clean_smiles[i+1].islower():
            curr_el += clean_smiles[i+1]
            i += 1
        if curr_el in ALLOWED_ELEMENTS:
            counts[curr_el] = counts.get(curr_el, 0) + 1
        i += 1
    if not counts: return None
    formula = "".join(f"{el}{counts[el] if counts[el]>1 else ''}" for el in sorted(counts.keys()))
    return formula, counts

def fetch_ord_data():
    print("正在通过 GitHub API 获取 ORD 数据样本...")
    compounds = {}
    reactions = set()
    headers = {"User-Agent": "ChemistryUNOCrawler"}
    
    # 基础化学反应库（作为保险，确保游戏始终有高质量反应）
    base_compounds = [
        {"formula": "H2O", "name": "水", "composition": {"H": 2, "O": 1}},
        {"formula": "CO2", "name": "二氧化碳", "composition": {"C": 1, "O": 2}},
        {"formula": "HCl", "name": "盐酸", "composition": {"H": 1, "Cl": 1}},
        {"formula": "NaOH", "name": "氢氧化钠", "composition": {"Na": 1, "O": 1, "H": 1}},
        {"formula": "NaCl", "name": "氯化钠", "composition": {"Na": 1, "Cl": 1}},
        {"formula": "H2SO4", "name": "硫酸", "composition": {"H": 2, "S": 1, "O": 4}},
        {"formula": "O2", "name": "氧气", "composition": {"O": 2}},
        {"formula": "H2", "name": "氢气", "composition": {"H": 2}},
        {"formula": "Fe", "name": "铁", "composition": {"Fe": 1}}
    ]
    base_reactions = [
        ("HCl", "NaOH"), ("H2SO4", "NaOH"), ("H2", "O2"), ("Fe", "O2"), ("HCl", "Fe")
    ]
    
    for bc in base_compounds: compounds[bc['formula']] = bc
    for br in base_reactions: reactions.add(tuple(sorted(br)))

    try:
        res = requests.get(BASE_URL, headers=headers, timeout=10)
        if res.status_code == 200:
            dirs = [d for d in res.json() if d['type'] == 'dir'][:1] 
            for d in dirs:
                files_res = requests.get(d['url'], headers=headers, timeout=10)
                files = [f for f in files_res.json() if f['name'].endswith('.json')][:2]
                for f in files:
                    print(f"解析数据集: {f['name']}...")
                    data_res = requests.get(f['download_url'], headers=headers)
                    data = data_res.json()
                    for rxn in data.get('reactions', []):
                        inputs = []
                        for input_data in rxn.get('inputs', {}).values():
                            for comp in input_data.get('components', []):
                                smiles = next((i['value'] for i in comp.get('identifiers', []) if i['type'] == 'SMILES'), None)
                                if smiles:
                                    res_formula = get_formula_from_smiles(smiles)
                                    if res_formula:
                                        formula, composition = res_formula
                                        if len(formula) < 12:
                                            name = comp.get('name', formula)
                                            compounds[formula] = {"formula": formula, "name": name, "composition": composition}
                                            inputs.append(formula)
                        if len(inputs) >= 2:
                            ui = list(set(inputs))
                            for i in range(len(ui)):
                                for j in range(i+1, len(ui)):
                                    reactions.add(tuple(sorted([ui[i], ui[j]])))
    except Exception as e:
        print(f"API 限制或网络问题 (使用内置基础库): {e}")

    return list(compounds.values()), [{"a": r[0], "b": r[1]} for r in reactions]

def update_index_html(compounds, reactions):
    if not os.path.exists(INDEX_HTML_PATH): return
    with open(INDEX_HTML_PATH, "r", encoding="utf-8") as f:
        content = f.read()
    
    # 强制 UTF-8 编码 JSON
    comp_json = json.dumps(compounds, ensure_ascii=False, indent=8)
    reac_json = json.dumps(reactions, ensure_ascii=False, indent=8)
    
    content = re.sub(r"const COMPOUNDS_DB = \[.*?\];", f"const COMPOUNDS_DB = {comp_json};", content, flags=re.DOTALL)
    content = re.sub(r"const REACTIONS_DB = \[.*?\];", f"const REACTIONS_DB = {reac_json};", content, flags=re.DOTALL)
    
    with open(INDEX_HTML_PATH, "w", encoding="utf-8") as f:
        f.write(content)
    print(f"Updated {INDEX_HTML_PATH}. Compounds: {len(compounds)}, Reactions: {len(reactions)}")

if __name__ == "__main__":
    c, r = fetch_ord_data()
    update_index_html(c, r)
