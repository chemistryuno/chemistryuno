#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

const refPath = path.join(__dirname, '../../ref.json');
const reactions = JSON.parse(fs.readFileSync(refPath, 'utf-8'));

console.log(`Loaded ${reactions.length} reactions from ref.json`);

let goCode = `// Reaction seed data generated from ref.json
// Generated at: ${new Date().toISOString()}
//
// Usage: append the following entries to backend/database/migrate.go
// inside initDefaultReactionsGORM().

// Continue importing reactions from ref.json
`;

reactions.forEach((reaction) => {
  let r1 = reaction.r1;
  let r2 = reaction.r2;
  let display = reaction.display;

  if (r1 > r2) {
    [r1, r2] = [r2, r1];
  }

  display = display.replace(/\\/g, '\\\\').replace(/"/g, '\\"');

  goCode += `{R1: "${r1}", R2: "${r2}", Display: "${display}", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},\n`;
});

const outputPath = path.join(__dirname, '../database/reactions_from_ref.go.txt');
fs.writeFileSync(outputPath, goCode, 'utf-8');

console.log(`Generated Go seed code for ${reactions.length} reactions`);
console.log(`Output written to ${outputPath}`);
console.log('\nPaste the generated entries into backend/database/migrate.go inside initDefaultReactionsGORM().');

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
	if err := database.InitDB("../chemistryuno.db"); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	data, err := os.ReadFile("../../ref.json")
	if err != nil {
		log.Fatalf("failed to read ref.json: %v", err)
	}

	var reactions []RefReaction
	if err := json.Unmarshal(data, &reactions); err != nil {
		log.Fatalf("failed to parse ref.json: %v", err)
	}

	fmt.Printf("Checking %d reactions...\\n\\n", len(reactions))

	missing := 0
	existing := 0

	for i, reaction := range reactions {
		r1, r2 := reaction.R1, reaction.R2

		if r1 > r2 {
			r1, r2 = r2, r1
		}

		var count int64
		database.DB.Model(&database.Reaction{}).
			Where("(r1 = ? AND r2 = ?) OR (r1 = ? AND r2 = ?)", r1, r2, r2, r1).
			Where("status = ?", "approved").
			Count(&count)

		if count == 0 {
			fmt.Printf("[%d] Missing: %s + %s -> %s\\n", i+1, r1, r2, reaction.Display)
			missing++
		} else {
			existing++
		}
	}

	fmt.Printf("\\nSummary:\\n")
	fmt.Printf("Existing: %d\\n", existing)
	fmt.Printf("Missing: %d\\n", missing)
	fmt.Printf("Total: %d\\n", len(reactions))
}
`;

const checkScriptPath = path.join(__dirname, 'check_ref_reactions.go');
fs.writeFileSync(checkScriptPath, checkScript, 'utf-8');

console.log(`\nGenerated helper check script at ${checkScriptPath}`);
console.log('Run it with: cd backend/scripts && go run check_ref_reactions.go');
