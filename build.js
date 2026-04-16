#!/usr/bin/env node

/**
 * Build frontend and backend for production.
 * - Builds frontend to frontend/dist/
 * - Syncs frontend assets to backend/static/dist/ for Go embed
 * - Collects frontend + backend artifacts into dist/
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const rootDir = __dirname;
const frontendDir = path.join(rootDir, 'frontend');
const frontendDistDir = path.join(frontendDir, 'dist');
const backendStaticDistDir = path.join(rootDir, 'backend', 'static', 'dist');
const distDir = path.join(rootDir, 'dist');
const binaryName = process.platform === 'win32' ? 'chemistryuno.exe' : 'chemistryuno';
const rootBinaryPath = path.join(rootDir, binaryName);
const distBinaryPath = path.join(distDir, binaryName);

function ensureDir(dirPath) {
  fs.mkdirSync(dirPath, { recursive: true });
}

function resetDir(dirPath) {
  fs.rmSync(dirPath, { recursive: true, force: true });
  ensureDir(dirPath);
}

function copyDir(srcDir, destDir) {
  ensureDir(destDir);

  for (const entry of fs.readdirSync(srcDir, { withFileTypes: true })) {
    const srcPath = path.join(srcDir, entry.name);
    const destPath = path.join(destDir, entry.name);

    if (entry.isDirectory()) {
      copyDir(srcPath, destPath);
    } else {
      fs.copyFileSync(srcPath, destPath);
    }
  }
}

function writeStartScript() {
  const startScriptPath = path.join(distDir, process.platform === 'win32' ? 'start.bat' : 'start.sh');
  const content = process.platform === 'win32'
    ? `@echo off
setlocal
cd /d "%~dp0"
if not exist ".env" if exist ".env.example" copy /Y ".env.example" ".env" >nul
"%~dp0${binaryName}"
endlocal
`
    : `#!/bin/sh
set -e
cd "$(dirname "$0")"
if [ ! -f .env ] && [ -f .env.example ]; then
  cp .env.example .env
fi
./${binaryName}
`;

  fs.writeFileSync(startScriptPath, content);
  if (process.platform !== 'win32') {
    fs.chmodSync(startScriptPath, 0o755);
  }
}

console.log('Building Chemistry UNO project...\n');

try {
  console.log('Step 1: Building frontend (Vite)...');
  execSync('pnpm build:frontend', { stdio: 'inherit', cwd: rootDir });
  if (!fs.existsSync(frontendDistDir)) {
    throw new Error(`Frontend dist not found at ${frontendDistDir}`);
  }
  console.log('Frontend build complete\n');

  console.log('Step 2: Syncing frontend assets to backend/static/dist...');
  resetDir(backendStaticDistDir);
  copyDir(frontendDistDir, backendStaticDistDir);
  console.log('Frontend assets synced for embed\n');

  console.log('Step 3: Building backend (Go)...');
  fs.rmSync(rootBinaryPath, { force: true });
  execSync('pnpm build:backend', { stdio: 'inherit', cwd: rootDir });
  if (!fs.existsSync(rootBinaryPath)) {
    throw new Error(`Backend binary not found at ${rootBinaryPath}`);
  }
  console.log('Backend build complete\n');

  console.log('Step 4: Collecting build artifacts into dist/...');
  resetDir(distDir);
  fs.copyFileSync(rootBinaryPath, distBinaryPath);
  copyDir(frontendDistDir, path.join(distDir, 'frontend'));

  const envExamplePath = path.join(rootDir, '.env.example');
  if (fs.existsSync(envExamplePath)) {
    fs.copyFileSync(envExamplePath, path.join(distDir, '.env.example'));
  }

  writeStartScript();
  fs.rmSync(rootBinaryPath, { force: true });

  console.log('Artifacts collected successfully\n');
  console.log('------------------------------------------------------------');
  console.log('Build successful!');
  console.log(`Backend binary: ${distBinaryPath}`);
  console.log(`Frontend files: ${path.join(distDir, 'frontend')}`);
  console.log(`Startup script: ${path.join(distDir, process.platform === 'win32' ? 'start.bat' : 'start.sh')}`);
  console.log('------------------------------------------------------------\n');
} catch (err) {
  console.error('Build failed:', err.message);
  process.exit(1);
}
