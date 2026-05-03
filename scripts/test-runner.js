#!/usr/bin/env node

const { execSync } = require('child_process');
const path = require('path');

const rootDir = path.resolve(__dirname, '..');

const stepSets = {
  quick: [
    { name: 'Run backend tests', command: 'go test ./backend/...' },
    { name: 'Run frontend type-check', command: 'pnpm -C frontend type-check' },
    { name: 'Run frontend unit/component tests', command: 'pnpm -C frontend test' },
  ],
  ci: [
    { name: 'Run backend tests', command: 'go test ./backend/...' },
    { name: 'Run frontend type-check', command: 'pnpm -C frontend type-check' },
    { name: 'Run frontend unit/component tests', command: 'pnpm -C frontend test' },
    { name: 'Run frontend production build', command: 'pnpm -C frontend build' },
  ],
  release: [
    { name: 'Run standard CI gate', command: 'node scripts/test-runner.js ci' },
    { name: 'Run backend script-tag OAuth tests', command: 'go test -tags scripts backend/scripts/oauth_third_party_test.go -v' },
    { name: 'Run full-stack e2e tests', command: 'node scripts/run-e2e.js' },
    { name: 'Run full production build', command: 'pnpm build:full' },
  ],
  coverage: [
    { name: 'Generate backend coverage summary', command: 'go test ./backend/... -coverprofile=tmp/coverage/backend.out' },
    { name: 'Print backend coverage functions', command: 'go tool cover -func=tmp/coverage/backend.out' },
    { name: 'Generate frontend coverage summary', command: 'pnpm -C frontend exec vitest run --coverage.enabled=true --coverage.reporter=text' },
  ],
};

function ensureCoverageDirs(mode) {
  if (mode !== 'coverage') {
    return;
  }

  const fs = require('fs');
  fs.mkdirSync(path.join(rootDir, 'tmp', 'coverage'), { recursive: true });
}

function runStep(step) {
  console.log(`\n==> ${step.name}`);
  console.log(`$ ${step.command}`);

  try {
    execSync(step.command, {
      cwd: rootDir,
      stdio: 'inherit',
      env: { ...process.env },
    });
  } catch (error) {
    const exitCode = typeof error.status === 'number' ? error.status : 1;
    console.error(`\n${step.name} failed with exit code ${exitCode}`);
    console.error(`Failed command: ${step.command}`);
    process.exit(exitCode);
  }
}

function main() {
  const mode = process.argv[2] || 'quick';
  const steps = stepSets[mode];

  if (!steps) {
    console.error(`Unknown test workflow "${mode}". Expected one of: ${Object.keys(stepSets).join(', ')}`);
    process.exit(1);
  }

  ensureCoverageDirs(mode);

  console.log(`Starting Chemistry UNO ${mode} test workflow...`);
  for (const step of steps) {
    runStep(step);
  }
  console.log(`\nChemistry UNO ${mode} test workflow completed successfully.`);
}

main();
