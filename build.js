#!/usr/bin/env node

/**
 * Build frontend (Vite) and backend (Go) for production
 * - Builds frontend Vue/TS to frontend/dist/
 * - Compiles Go binary to chemistryuno (or chemistryuno.exe on Windows)
 */

const { execSync } = require('child_process');
const path = require('path');
const fs = require('fs');

const rootDir = __dirname;

console.log('馃彈锔? Building Chemistry UNO project...\n');

try {
  // Step 1: Build frontend
  console.log('馃摝 Step 1: Building frontend (Vite)...');
  try {
    execSync('pnpm build:frontend', { stdio: 'inherit', cwd: rootDir });
    console.log('鉁?Frontend build complete\n');
  } catch (err) {
    console.error('鉂?Frontend build failed:', err.message);
    process.exit(1);
  }

  // Step 2: Build backend
  console.log('馃敤 Step 2: Building backend (Go)...');
  try {
    execSync('pnpm build:backend', { stdio: 'inherit', cwd: rootDir });
    console.log('鉁?Backend build complete\n');
  } catch (err) {
    console.error('鉂?Backend build failed:', err.message);
    process.exit(1);
  }

  // Verify outputs
  console.log('馃搵 Verifying build outputs...');
  const frontendDist = path.join(rootDir, 'frontend', 'dist');
  const binaryName = process.platform === 'win32' ? 'chemistryuno.exe' : 'chemistryuno';
  const binaryPath = path.join(rootDir, binaryName);

  let distExists = false;
  let binaryExists = false;

  if (fs.existsSync(frontendDist)) {
    console.log(`鉁?Frontend dist: ${frontendDist}`);
    distExists = true;
  } else {
    console.warn(`鈿狅笍  Frontend dist not found at: ${frontendDist}`);
  }

  if (fs.existsSync(binaryPath)) {
    console.log(`鉁?Binary: ${binaryPath}`);
    binaryExists = true;
  } else {
    console.warn(`鈿狅笍  Binary not found at: ${binaryPath}`);
  }

  console.log('\n鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣');
  if (distExists && binaryExists) {
    console.log('鉁?Build successful!');
    console.log('\nNext steps:');
    console.log(`  Run: ./${binaryName} (or chemistryuno.exe on Windows)`);
    console.log('  Then visit: http://localhost:8080');
  } else {
    console.log('鈿狅笍  Build completed with warnings');
    console.log('Check output above for details');
  }
  console.log('鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣\n');

} catch (err) {
  console.error('鉂?Build failed:', err.message);
  process.exit(1);
}
