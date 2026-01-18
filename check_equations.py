import json
import re

with open('index.html', 'r', encoding='utf-8') as f:
    content = f.read()

# Extract EQUATIONS_LIB
eq_match = re.search(r'const EQUATIONS_LIB = (\{.*?\});', content, re.DOTALL)
if eq_match:
    eq_str = eq_match.group(1)
    # Use a single regex to catch pairs
    pairs = re.findall(r"'(.+?)':\s*'(.+?)'", eq_str)
    equations_lib = {k: v for k, v in pairs}
    print(f"Total equations found: {len(equations_lib)}")
else:
    print("EQUATIONS_LIB not found")
    equations_lib = {}

# Extract REACTIONS_DB
re_match = re.search(r'const REACTIONS_DB = (\[.*?\]);', content, re.DOTALL)
if re_match:
    re_str = re_match.group(1)
    # Convert JS array to Python list of dicts. 
    # Use simple regex for this specific structure.
    reactions_db = []
    for pair in re.findall(r'\{\s*"a":\s*"(.+?)",\s*"b":\s*"(.+?)"\s*\}', re_str):
        reactions_db.append({"a": pair[0], "b": pair[1]})
    print(f"Total reactions found: {len(reactions_db)}")
else:
    print("REACTIONS_DB not found")
    reactions_db = []

missing = []
for r in reactions_db:
    key1 = f"{r['a']}+{r['b']}"
    key2 = f"{r['b']}+{r['a']}"
    if key1 not in equations_lib and key2 not in equations_lib:
        missing.append(f"{r['a']}+{r['b']}")

print("Missing keys from REACTIONS_DB in EQUATIONS_LIB:")
for m in missing:
    print(m)

placeholders = []
for k, v in equations_lib.items():
    if "反应中" in v:
        placeholders.append(k)

print("\nKeys with '反应中' in EQUATIONS_LIB:")
for p in placeholders:
    print(p)
