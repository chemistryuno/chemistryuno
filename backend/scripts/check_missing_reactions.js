#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// 璇诲彇 ref.json
const refPath = path.join(__dirname, '../../ref.json');
const reactions = JSON.parse(fs.readFileSync(refPath, 'utf-8'));

// 璇诲彇 migrate.go
const migratePath = path.join(__dirname, '../database/migrate.go');
const migrateContent = fs.readFileSync(migratePath, 'utf-8');

console.log('馃攳 妫€鏌?ref.json 涓殑鍙嶅簲鏄惁宸插瓨鍦ㄤ簬鏁版嵁搴撳垵濮嬪寲浠ｇ爜涓璡n');
console.log(`鎬诲叡闇€瑕佹鏌ワ細${reactions.length} 鏉″弽搴擻n`);

let missing = [];
let existing = [];

reactions.forEach((reaction, index) => {
    let r1 = reaction.r1;
    let r2 = reaction.r2;
    let display = reaction.display;

    // 纭繚 R1 <= R2锛堝瓧鍏稿簭锛?
    if (r1 > r2) {
        [r1, r2] = [r2, r1];
    }

    // 妫€鏌ユ槸鍚﹀瓨鍦?(妫€鏌ヤ袱绉嶅彲鑳界殑椤哄簭)
    const pattern1 = `R1: "${r1}", R2: "${r2}"`;
    const pattern2 = `R1: "${r2}", R2: "${r1}"`;

    if (migrateContent.includes(pattern1) || migrateContent.includes(pattern2)) {
        existing.push({ r1, r2, display });
    } else {
        missing.push({ r1, r2, display, index: index + 1 });
    }
});

console.log('鉁?宸插瓨鍦ㄧ殑鍙嶅簲锛?, existing.length, '鏉n');

if (missing.length > 0) {
    console.log('鉂?缂哄け鐨勫弽搴旓細', missing.length, '鏉n');
    console.log('璇︾粏鍒楄〃锛?);
    missing.forEach((item) => {
        console.log(`  [${item.index}] ${item.r1} + ${item.r2} -> ${item.display}`);
    });
} else {
    console.log('馃帀 澶浜嗭紒ref.json 涓殑鎵€鏈夊弽搴旈兘宸插瓨鍦ㄤ簬鏁版嵁搴撲腑锛?);
}

console.log('\n缁熻缁撴灉锛?);
console.log('鈹?.repeat(60));
console.log(`鉁?宸插瓨鍦? ${existing.length} 鏉?(${(existing.length / reactions.length * 100).toFixed(1)}%)`);
console.log(`鉂?缂哄け: ${missing.length} 鏉?(${(missing.length / reactions.length * 100).toFixed(1)}%)`);
console.log(`馃搳 鎬昏: ${reactions.length} 鏉);
console.log('鈹?.repeat(60));

// 濡傛灉鏈夌己澶辩殑鍙嶅簲锛岀敓鎴愬彧鍖呭惈缂哄け鍙嶅簲鐨?Go 浠ｇ爜
if (missing.length > 0) {
    console.log('\n馃挕 鎻愮ず锛氬彲浠ュ彧娣诲姞缂哄け鐨勫弽搴旓紝鑰屼笉鏄叏閮ㄩ噸鏂版坊鍔?);
    console.log('   鐢熸垚鐨勪唬鐮佸凡淇濆瓨鍒?backend/database/missing_reactions.go.txt\n');

    let missingCode = `// 缂哄け鐨勫弽搴旀暟鎹紙浠?ref.json 涓彁鍙栵級
// 鐢熸垚鏃堕棿: ${new Date().toISOString()}
// 缂哄け鏁伴噺: ${missing.length} 鏉?
//
// 浣跨敤鏂规硶锛氬皢浠ヤ笅浠ｇ爜娣诲姞鍒?backend/database/migrate.go 鐨?initDefaultReactionsGORM() 鍑芥暟鐨?reactions 鏁扮粍涓?

`;

    missing.forEach((item) => {
        const display = item.display.replace(/\\/g, '\\\\').replace(/"/g, '\\"');
        missingCode += `{R1: "${item.r1}", R2: "${item.r2}", Display: "${display}", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},\n`;
    });

    const outputPath = path.join(__dirname, '../database/missing_reactions.go.txt');
    fs.writeFileSync(outputPath, missingCode, 'utf-8');
}
