#!/usr/bin/env node

const { execSync } = require('child_process');
const path = require('path');

const rootDir = path.resolve(__dirname, '..');

const steps = [
  {
    name: 'Run backend tests',
    command: 'go test ./backend/...'
  },
  {
    name: 'Run frontend type-check',
    command: 'pnpm -C frontend type-check'
  },
  {
    name: 'Run frontend build',
    command: 'pnpm -C frontend build'
  }
];

console.log('Starting CI test script...');

for (const step of steps) {
  console.log(`\n==> ${step.name}`);

  try {
    execSync(step.command, {
      cwd: rootDir,
      stdio: 'inherit'
    });
  } catch (error) {
    const exitCode = typeof error.status === 'number' ? error.status : 1;
    console.error(`${step.name} failed with exit code ${exitCode}`);
    process.exit(exitCode);
  }
}

console.log('\nCI test script completed successfully.');
