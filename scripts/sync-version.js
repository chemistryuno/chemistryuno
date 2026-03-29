#!/usr/bin/env node

/**
 * Keep workspace package versions aligned.
 * Source of truth: root package.json version.
 * Targets: frontend/package.json.
 */

const fs = require('fs');
const path = require('path');

const rootDir = path.resolve(__dirname, '..');
const rootPackagePath = path.join(rootDir, 'package.json');
const frontendPackagePath = path.join(rootDir, 'frontend', 'package.json');

function readJson(filePath) {
  return JSON.parse(fs.readFileSync(filePath, 'utf8'));
}

function writeJson(filePath, value) {
  fs.writeFileSync(filePath, `${JSON.stringify(value, null, 2)}\n`, 'utf8');
}

function syncFrontendVersion(rootVersion) {
  const frontendPackage = readJson(frontendPackagePath);
  if (frontendPackage.version === rootVersion) {
    console.log(`frontend/package.json already matches version ${rootVersion}`);
    return false;
  }

  const oldVersion = frontendPackage.version;
  frontendPackage.version = rootVersion;
  writeJson(frontendPackagePath, frontendPackage);
  console.log(`frontend/package.json: ${oldVersion} -> ${rootVersion}`);
  return true;
}

function main() {
  if (!fs.existsSync(rootPackagePath)) {
    console.error('Missing root package.json');
    process.exit(1);
  }
  if (!fs.existsSync(frontendPackagePath)) {
    console.error('Missing frontend/package.json');
    process.exit(1);
  }

  const rootPackage = readJson(rootPackagePath);
  const rootVersion = rootPackage.version;

  if (!rootVersion || typeof rootVersion !== 'string') {
    console.error('Root package.json has invalid version field');
    process.exit(1);
  }

  const changed = syncFrontendVersion(rootVersion);
  if (changed) {
    console.log('Version sync complete');
  } else {
    console.log('No changes needed');
  }
}

main();
