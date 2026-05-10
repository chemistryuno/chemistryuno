#!/usr/bin/env node

const path = require('path');
const { execFileSync } = require('child_process');

const args = process.argv.slice(2);
const mode = args.includes('--build') ? 'ci' : 'quick';
const runnerPath = path.join(__dirname, '..', 'scripts', 'test-runner.js');

execFileSync(process.execPath, [runnerPath, mode], {
  stdio: 'inherit',
});
