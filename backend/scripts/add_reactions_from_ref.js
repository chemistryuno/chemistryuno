#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// 璇诲彇 ref.json
const refPath = path.join(__dirname, '../../ref.json');
const reactions = JSON.parse(fs.readFileSync(refPath, 'utf-8'));

console.log(`璇诲彇鍒?${reactions.length} 鏉″弽搴旀暟鎹甛n`);

// 鐢熸垚 Go 浠ｇ爜
let goCode = `// 浠?ref.json 鑷姩鐢熸垚鐨勫弽搴旀暟鎹?
// 鐢熸垚鏃堕棿: ${new Date().toISOString()}
//
// 浣跨敤鏂规硶锛氬皢浠ヤ笅浠ｇ爜娣诲姞鍒?backend/database/migrate.go 鐨?initDefaultReactionsGORM() 鍑芥暟涓?

// 缁х画娣诲姞鍙嶅簲锛堜粠 ref.json 瀵煎叆锛?
`;

reactions.forEach((reaction, index) => {
    let r1 = reaction.r1;
    let r2 = reaction.r2;
    let display = reaction.display;

    // 纭繚 R1 <= R2锛堝瓧鍏稿簭锛?
    if (r1 > r2) {
        [r1, r2] = [r2, r1];
    }

    // 杞箟鏄剧ず鏂囨湰涓殑鐗规畩瀛楃
    display = display.replace(/\\/g, '\\\\').replace(/"/g, '\\"');

    goCode += `{R1: "${r1}", R2: "${r2}", Display: "${display}", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},\n`;
});

// 杈撳嚭鍒版枃浠?
const outputPath = path.join(__dirname, '../database/reactions_from_ref.go.txt');
fs.writeFileSync(outputPath, goCode, 'utf-8');

console.log(`鉁?宸茬敓鎴?${reactions.length} 鏉″弽搴旂殑 Go 浠ｇ爜`);
console.log(`馃摑 杈撳嚭鏂囦欢锛?{outputPath}`);
console.log(`\n璇锋墜鍔ㄥ皢鐢熸垚鐨勪唬鐮佸鍒跺埌 backend/database/migrate.go 鐨?initDefaultReactionsGORM() 鍑芥暟鐨?reactions 鏁扮粍涓璥);

// 鍚屾椂鐢熸垚涓€涓鏌ヨ剼鏈紝鐢ㄤ簬楠岃瘉 ref.json 涓殑鍙嶅簲鏄惁宸茬粡鍦ㄦ暟鎹簱涓?
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
	// 鍒濆鍖栨暟鎹簱
	if err := database.InitDB("../chemistryuno.db"); err != nil {
		log.Fatalf("鏁版嵁搴撳垵濮嬪寲澶辫触: %v", err)
	}

	// 璇诲彇 ref.json
	data, err := os.ReadFile("../../ref.json")
	if err != nil {
		log.Fatalf("璇诲彇 ref.json 澶辫触: %v", err)
	}

	var reactions []RefReaction
	if err := json.Unmarshal(data, &reactions); err != nil {
		log.Fatalf("瑙ｆ瀽 ref.json 澶辫触: %v", err)
	}

	fmt.Printf("妫€鏌?%d 鏉″弽搴?..\\n\\n", len(reactions))

	missing := 0
	existing := 0

	for i, reaction := range reactions {
		r1, r2 := reaction.R1, reaction.R2

		// 纭繚 R1 <= R2
		if r1 > r2 {
			r1, r2 = r2, r1
		}

		// 鏌ヨ鏁版嵁搴?
		var count int64
		database.DB.Model(&database.Reaction{}).
			Where("(r1 = ? AND r2 = ?) OR (r1 = ? AND r2 = ?)", r1, r2, r2, r1).
			Where("status = ?", "approved").
			Count(&count)

		if count == 0 {
			fmt.Printf("[%d] 鉂?缂哄け: %s + %s -> %s\\n", i+1, r1, r2, reaction.Display)
			missing++
		} else {
			existing++
		}
	}

	fmt.Printf("\\n缁熻缁撴灉:\\n")
	fmt.Printf("鉁?宸插瓨鍦? %d 鏉\n", existing)
	fmt.Printf("鉂?缂哄け: %d 鏉\n", missing)
	fmt.Printf("馃搳 鎬昏: %d 鏉\n", len(reactions))
}
`;

const checkScriptPath = path.join(__dirname, 'check_ref_reactions.go');
fs.writeFileSync(checkScriptPath, checkScript, 'utf-8');

console.log(`\n馃攳 鍚屾椂鐢熸垚浜嗘鏌ヨ剼鏈細${checkScriptPath}`);
console.log(`   浣跨敤鏂规硶锛歝d backend/scripts && go run check_ref_reactions.go`);
