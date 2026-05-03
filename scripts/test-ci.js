#!/usr/bin/env node

require('child_process').execFileSync(
  process.execPath,
  [require('path').join(__dirname, 'test-runner.js'), 'ci'],
  { stdio: 'inherit' }
);
