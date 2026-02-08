#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// 读取 ref.json
const refPath = path.join(__dirname, '../../ref.json');
const reactions = JSON.parse(fs.readFileSync(refPath, 'utf-8'));

console.log(`读取到 ${reactions.length} 条反应数据\n`);

// 生成 Go 代码
let goCode = `// 从 ref.json 自动生成的反应数据
// 生成时间: ${new Date().toISOString()}
//
// 使用方法：将以下代码添加到 backend/database/migrate.go 的 initDefaultReactionsGORM() 函数中

// 继续添加反应（从 ref.json 导入）
`;

reactions.forEach((reaction, index) => {
    let r1 = reaction.r1;
    let r2 = reaction.r2;
    let display = reaction.display;

    // 确保 R1 <= R2（字典序）
    if (r1 > r2) {
        [r1, r2] = [r2, r1];
    }

    // 转义显示文本中的特殊字符
    display = display.replace(/\\/g, '\\\\').replace(/"/g, '\\"');

    goCode += `{R1: "${r1}", R2: "${r2}", Display: "${display}", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},\n`;
});

// 输出到文件
const outputPath = path.join(__dirname, '../database/reactions_from_ref.go.txt');
fs.writeFileSync(outputPath, goCode, 'utf-8');

console.log(`✅ 已生成 ${reactions.length} 条反应的 Go 代码`);
console.log(`📝 输出文件：${outputPath}`);
console.log(`\n请手动将生成的代码复制到 backend/database/migrate.go 的 initDefaultReactionsGORM() 函数的 reactions 数组中`);

// 同时生成一个检查脚本，用于验证 ref.json 中的反应是否已经在数据库中
const checkScript = `
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"chemistryuno/database"
)

type RefReaction struct {
	R1      string \`json:"r1"\`
	R2      string \`json:"r2"\`
	Display string \`json:"display"\`
}

func main() {
	// 初始化数据库
	if err := database.InitDB("../chemistryuno.db"); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 读取 ref.json
	data, err := os.ReadFile("../../ref.json")
	if err != nil {
		log.Fatalf("读取 ref.json 失败: %v", err)
	}

	var reactions []RefReaction
	if err := json.Unmarshal(data, &reactions); err != nil {
		log.Fatalf("解析 ref.json 失败: %v", err)
	}

	fmt.Printf("检查 %d 条反应...\\n\\n", len(reactions))

	missing := 0
	existing := 0

	for i, reaction := range reactions {
		r1, r2 := reaction.R1, reaction.R2

		// 确保 R1 <= R2
		if r1 > r2 {
			r1, r2 = r2, r1
		}

		// 查询数据库
		var count int64
		database.DB.Model(&database.Reaction{}).
			Where("(r1 = ? AND r2 = ?) OR (r1 = ? AND r2 = ?)", r1, r2, r2, r1).
			Where("status = ?", "approved").
			Count(&count)

		if count == 0 {
			fmt.Printf("[%d] ❌ 缺失: %s + %s -> %s\\n", i+1, r1, r2, reaction.Display)
			missing++
		} else {
			existing++
		}
	}

	fmt.Printf("\\n统计结果:\\n")
	fmt.Printf("✅ 已存在: %d 条\\n", existing)
	fmt.Printf("❌ 缺失: %d 条\\n", missing)
	fmt.Printf("📊 总计: %d 条\\n", len(reactions))
}
`;

const checkScriptPath = path.join(__dirname, 'check_ref_reactions.go');
fs.writeFileSync(checkScriptPath, checkScript, 'utf-8');

console.log(`\n🔍 同时生成了检查脚本：${checkScriptPath}`);
console.log(`   使用方法：cd backend/scripts && go run check_ref_reactions.go`);
