#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// 读取 ref.json
const refPath = path.join(__dirname, '../../ref.json');
const reactions = JSON.parse(fs.readFileSync(refPath, 'utf-8'));

// 读取 migrate.go
const migratePath = path.join(__dirname, '../database/migrate.go');
const migrateContent = fs.readFileSync(migratePath, 'utf-8');

console.log('🔍 检查 ref.json 中的反应是否已存在于数据库初始化代码中\n');
console.log(`总共需要检查：${reactions.length} 条反应\n`);

let missing = [];
let existing = [];

reactions.forEach((reaction, index) => {
    let r1 = reaction.r1;
    let r2 = reaction.r2;
    let display = reaction.display;

    // 确保 R1 <= R2（字典序）
    if (r1 > r2) {
        [r1, r2] = [r2, r1];
    }

    // 检查是否存在 (检查两种可能的顺序)
    const pattern1 = `R1: "${r1}", R2: "${r2}"`;
    const pattern2 = `R1: "${r2}", R2: "${r1}"`;

    if (migrateContent.includes(pattern1) || migrateContent.includes(pattern2)) {
        existing.push({ r1, r2, display });
    } else {
        missing.push({ r1, r2, display, index: index + 1 });
    }
});

console.log('✅ 已存在的反应：', existing.length, '条\n');

if (missing.length > 0) {
    console.log('❌ 缺失的反应：', missing.length, '条\n');
    console.log('详细列表：');
    missing.forEach((item) => {
        console.log(`  [${item.index}] ${item.r1} + ${item.r2} -> ${item.display}`);
    });
} else {
    console.log('🎉 太棒了！ref.json 中的所有反应都已存在于数据库中！');
}

console.log('\n统计结果：');
console.log('━'.repeat(60));
console.log(`✅ 已存在: ${existing.length} 条 (${(existing.length / reactions.length * 100).toFixed(1)}%)`);
console.log(`❌ 缺失: ${missing.length} 条 (${(missing.length / reactions.length * 100).toFixed(1)}%)`);
console.log(`📊 总计: ${reactions.length} 条`);
console.log('━'.repeat(60));

// 如果有缺失的反应，生成只包含缺失反应的 Go 代码
if (missing.length > 0) {
    console.log('\n💡 提示：可以只添加缺失的反应，而不是全部重新添加');
    console.log('   生成的代码已保存到 backend/database/missing_reactions.go.txt\n');

    let missingCode = `// 缺失的反应数据（从 ref.json 中提取）
// 生成时间: ${new Date().toISOString()}
// 缺失数量: ${missing.length} 条
//
// 使用方法：将以下代码添加到 backend/database/migrate.go 的 initDefaultReactionsGORM() 函数的 reactions 数组中

`;

    missing.forEach((item) => {
        const display = item.display.replace(/\\/g, '\\\\').replace(/"/g, '\\"');
        missingCode += `{R1: "${item.r1}", R2: "${item.r2}", Display: "${display}", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},\n`;
    });

    const outputPath = path.join(__dirname, '../database/missing_reactions.go.txt');
    fs.writeFileSync(outputPath, missingCode, 'utf-8');
}
