#!/usr/bin/env python3
"""
批量迁移SQL查询到GORM
"""
import os
import re

# 需要完全迁移的文件列表
files_to_migrate = [
    "handlers/admin.go",
    "handlers/game.go",
    "handlers/points.go",
    "handlers/feedback.go",
    "handlers/deck.go",
    "handlers/webauthn.go",
    "handlers/announcement.go",
    "game/manager.go",
    "game/cron.go",
    "game/chemistry.go"
]

# 简单的替换规则
patterns = [
    # 基本查询
    (r'database\.LegacyDB\.QueryRow\((.*?)\)\.Scan', r'/* TODO: Use Repository */ database.LegacyDB.QueryRow(\1).Scan'),
    (r'database\.LegacyDB\.Exec\((.*?)\)', r'/* TODO: Use Repository */ database.LegacyDB.Exec(\1)'),
    (r'database\.LegacyDB\.Query\((.*?)\)', r'/* TODO: Use Repository */ database.LegacyDB.Query(\1)'),
]

def migrate_file(filepath):
    """迁移单个文件"""
    if not os.path.exists(filepath):
        print(f"文件不存在: {filepath}")
        return False
    
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # 应用替换规则
    modified = content
    for pattern, replacement in patterns:
        modified = re.sub(pattern, replacement, modified, flags=re.DOTALL)
    
    if modified != content:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(modified)
        print(f"✓ 已标记: {filepath}")
        return True
    else:
        print(f"- 跳过: {filepath}")
        return False

if __name__ == "__main__":
    backend_dir = os.path.dirname(os.path.abspath(__file__))
    os.chdir(backend_dir)
    
    for file in files_to_migrate:
        migrate_file(file)
    
    print("\n标记完成！请手动迁移每个TODO标记的地方。")
