#!/usr/bin/env node

const assert = require('assert');
const path = require('path');
const { generateAuditReport } = require('../scripts/feature-coverage-audit');

const rootDir = path.resolve(__dirname, '..');
const report = generateAuditReport(rootDir);
const byFamily = new Map(report.coverage.map((item) => [item.family, item]));

function expectFamily(family, classification) {
  const finding = byFamily.get(family);
  assert(finding, `expected coverage entry for family: ${family}`);
  assert.strictEqual(finding.classification, classification, `expected ${family} to be ${classification}`);
  return finding;
}

try {
  const anticheat = expectFamily('anticheat', 'matched');
  assert(anticheat.frontend.direct, 'anticheat should include a frontend route/page');
  assert(anticheat.backend.support, 'anticheat should include backend support evidence');

  const health = expectFamily('health', 'backend-only');
  assert(health.backend.direct, 'health should include a backend API route');
  assert.strictEqual(health.frontend.direct, false, 'health should not have frontend coverage');

  const version = expectFamily('version', 'backend-only');
  assert(version.backend.direct, 'version should include a backend API route');
  assert.strictEqual(version.frontend.direct, false, 'version should not have frontend coverage');

  assert(report.summary.matched > 0, 'expected at least one matched family');
  assert(report.summary['backend-only'] > 0, 'expected at least one backend-only family');
  assert(report.findings.some((finding) => finding.classification === 'backend-only'), 'expected gap findings to be present');

  process.stdout.write('Feature coverage audit verification passed.\n');
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exit(1);
}
