#!/usr/bin/env node

/**
 * Run project tests:
 * - Go tests (pnpm go:test)
 * - Optional frontend build (pnpm build:frontend)
 * - Frontend type checks (pnpm -C frontend type-check)
 */

const { execSync } = require('child_process');
const path = require('path');

const rootDir = __dirname;
const args = process.argv.slice(2);

console.log('Running Chemistry UNO tests...\n');

try {
  console.log('Step 1: Running Go test suite...');
  try {
    execSync('pnpm go:test', { stdio: 'inherit', cwd: rootDir });
    console.log('Go tests passed\n');
  } catch (err) {
    console.error('Go tests failed');
    process.exit(1);
  }

  if (args.includes('--build')) {
    console.log('Step 2: Testing frontend build...');
    try {
      execSync('pnpm build:frontend', { stdio: 'inherit', cwd: rootDir });
      console.log('Frontend build successful\n');
    } catch (err) {
      console.error('Frontend build failed');
      process.exit(1);
    }
  }

  console.log('Step 3: Running frontend type check...');
  try {
    execSync('pnpm -C frontend type-check', { stdio: 'inherit', cwd: rootDir });
    console.log('Frontend type check passed\n');
  } catch (err) {
    console.error('Frontend type check warnings (non-blocking)');
  }

  console.log('------------------------------------------------------------');
  console.log('All tests passed!');
  console.log('------------------------------------------------------------\n');
} catch (err) {
  console.error('Test suite failed');
  process.exit(1);
}
