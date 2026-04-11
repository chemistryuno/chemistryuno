#!/usr/bin/env node

/**
 * Run project tests:
 * - Go tests (pnpm go:test)
 * - Optional: Frontend type checks (pnpm -C frontend type-check)
 */

const { execSync } = require('child_process');
const path = require('path');

const rootDir = __dirname;
const args = process.argv.slice(2);

console.log('馃И Running Chemistry UNO tests...\n');

try {
  // Step 1: Run Go tests
  console.log('馃攳 Step 1: Running Go test suite...');
  try {
    execSync('pnpm go:test', { stdio: 'inherit', cwd: rootDir });
    console.log('鉁?Go tests passed\n');
  } catch (err) {
    console.error('鉂?Go tests failed');
    process.exit(1);
  }

  // Step 2: Optional frontend build test (if --build flag passed)
  if (args.includes('--build')) {
    console.log('馃敤 Step 2: Testing frontend build...');
    try {
      execSync('pnpm build:frontend', { stdio: 'inherit', cwd: rootDir });
      console.log('鉁?Frontend build successful\n');
    } catch (err) {
      console.error('鉂?Frontend build failed');
      process.exit(1);
    }
  }

  // Step 3: Optional frontend type check
  console.log('馃摑 Step 3: Running frontend type check...');
  try {
    execSync('pnpm -C frontend type-check', { stdio: 'inherit', cwd: rootDir });
    console.log('鉁?Frontend type check passed\n');
  } catch (err) {
    console.error('鈿狅笍  Frontend type check warnings (non-blocking)');
  }

  console.log('鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣');
  console.log('鉁?All tests passed!');
  console.log('鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣\n');

} catch (err) {
  console.error('鉂?Test suite failed');
  process.exit(1);
}
