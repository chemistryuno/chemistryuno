
import re

def analyze():
    with open('index.html', 'r', encoding='utf-8') as f:
        content = f.read()

    # Extract REACTIONS_DB content
    start_marker = 'const REACTIONS_DB = ['
    end_marker = '];'
    start_index = content.find(start_marker)
    if start_index == -1:
        print("Could not find REACTIONS_DB")
        return
    
    end_index = content.find(end_marker, start_index)
    reactions_content = content[start_index + len(start_marker) : end_index]
    
    # Extract EQUATIONS_LIB content
    start_marker_lib = 'const EQUATIONS_LIB = {'
    end_marker_lib = '};'
    start_index_lib = content.find(start_marker_lib)
    if start_index_lib == -1:
        print("Could not find EQUATIONS_LIB")
        return
    
    end_index_lib = content.find(end_marker_lib, start_index_lib)
    equations_content = content[start_index_lib + len(start_marker_lib) : end_index_lib]

    # Parse REACTIONS_DB using regex to find { "a": "...", "b": "..." }
    reactions = []
    reaction_pattern = re.compile(r'\{\s*"a":\s*"([^"]+)",\s*"b":\s*"([^"]+)"\s*\}')
    for match in reaction_pattern.finditer(reactions_content):
        reactions.append({"a": match.group(1), "b": match.group(2)})

    # Parse EQUATIONS_LIB using regex to find 'key': 'value'
    equations_keys = set()
    lib_pattern = re.compile(r"['\"]([^'\"]+)['\"]\s*:\s*['\"]([^'\"]+)['\"]")
    for match in lib_pattern.finditer(equations_content):
        equations_keys.add(match.group(1))

    # Task 1: Check REACTIONS_DB for "O2" entries and match with EQUATIONS_LIB
    missing_equations = []
    found_o2_reactions = 0
    for entry in reactions:
        a = entry['a']
        b = entry['b']
        if a == 'O2' or b == 'O2':
            found_o2_reactions += 1
            key1 = f"{a}+{b}"
            key2 = f"{b}+{a}"
            if key1 not in equations_keys and key2 not in equations_keys:
                missing_equations.append(f"{a} + {b}")

    print(f"Total O2 reactions found in REACTIONS_DB: {found_o2_reactions}")
    missing_equations = sorted(list(set(missing_equations)))
    print("--- Missing in EQUATIONS_LIB (but in REACTIONS_DB) ---")
    if not missing_equations:
        print("None")
    for m in missing_equations:
        print(m)

    # Task 2: Check for missing common metals + O2 in REACTIONS_DB
    common_metals = ["Na", "K", "Ca", "Mg", "Al", "Zn", "Fe", "Cu", "Hg", "Ag", "Ba"]
    missing_metals_o2 = []
    
    normalized_reactions = set()
    for entry in reactions:
        pair = tuple(sorted([entry['a'], entry['b']]))
        normalized_reactions.add(pair)

    for metal in common_metals:
        pair = tuple(sorted([metal, "O2"]))
        if pair not in normalized_reactions:
            missing_metals_o2.append(metal)

    print("\n--- Common Metals missing from REACTIONS_DB (with O2) ---")
    if not missing_metals_o2:
        print("None")
    for m in missing_metals_o2:
        print(f"{m} + O2")

if __name__ == "__main__":
    analyze()
