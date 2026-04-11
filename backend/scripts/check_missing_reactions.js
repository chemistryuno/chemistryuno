#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

const refPath = path.join(__dirname, '../../ref.json');
const reactions = JSON.parse(fs.readFileSync(refPath, 'utf-8'));

const migratePath = path.join(__dirname, '../database/migrate.go');
const migrateContent = fs.readFileSync(migratePath, 'utf-8');

console.log('Checking whether reactions from ref.json already exist in migrate.go...\n');
console.log(`Total reactions to check: ${reactions.length}\n`);

const missing = [];
const existing = [];

reactions.forEach((reaction, index) => {
  let r1 = reaction.r1;
  let r2 = reaction.r2;
  const display = reaction.display;

  if (r1 > r2) {
    [r1, r2] = [r2, r1];
  }

  const pattern1 = `R1: "${r1}", R2: "${r2}"`;
  const pattern2 = `R1: "${r2}", R2: "${r1}"`;

  if (migrateContent.includes(pattern1) || migrateContent.includes(pattern2)) {
    existing.push({ r1, r2, display });
  } else {
    missing.push({ r1, r2, display, index: index + 1 });
  }
});

console.log(`Existing reactions: ${existing.length}`);

if (missing.length > 0) {
  console.log(`Missing reactions: ${missing.length}`);
  console.log('Missing list:');
  missing.forEach((item) => {
    console.log(`  [${item.index}] ${item.r1} + ${item.r2} -> ${item.display}`);
  });
} else {
  console.log('All reactions from ref.json are already present in the database seed.');
}

console.log('\nSummary');
console.log('-'.repeat(60));
console.log(`Existing: ${existing.length} (${(existing.length / reactions.length * 100).toFixed(1)}%)`);
console.log(`Missing: ${missing.length} (${(missing.length / reactions.length * 100).toFixed(1)}%)`);
console.log(`Total: ${reactions.length}`);
console.log('-'.repeat(60));

if (missing.length > 0) {
  console.log('\nTip: only the missing reactions will be written to backend/database/missing_reactions.go.txt\n');

  let missingCode = `// Missing reaction seed data extracted from ref.json
// Generated at: ${new Date().toISOString()}
// Missing count: ${missing.length}
//
// Usage: append the following entries to backend/database/migrate.go
// inside initDefaultReactionsGORM().

`;

  missing.forEach((item) => {
    const escapedDisplay = item.display.replace(/\\/g, '\\\\').replace(/"/g, '\\"');
    missingCode += `{R1: "${item.r1}", R2: "${item.r2}", Display: "${escapedDisplay}", GroupID: func() *uint { id := groupIDBase + uint(reactionCount); reactionCount++; return &id }(), CreatedByUID: 100000000, Status: "approved"},\n`;
  });

  const outputPath = path.join(__dirname, '../database/missing_reactions.go.txt');
  fs.writeFileSync(outputPath, missingCode, 'utf-8');
}
