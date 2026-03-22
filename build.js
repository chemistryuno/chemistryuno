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

console.log('🏗️  Building Chemistry UNO project...\n');

try {
  // Step 1: Build frontend
  console.log('📦 Step 1: Building frontend (Vite)...');
  try {
    execSync('pnpm build:frontend', { stdio: 'inherit', cwd: rootDir });
    console.log('✅ Frontend build complete\n');
  } catch (err) {
    console.error('❌ Frontend build failed:', err.message);
    process.exit(1);
  }

  // Step 2: Build backend
  console.log('🔨 Step 2: Building backend (Go)...');
  try {
    execSync('pnpm build:backend', { stdio: 'inherit', cwd: rootDir });
    console.log('✅ Backend build complete\n');
  } catch (err) {
    console.error('❌ Backend build failed:', err.message);
    process.exit(1);
  }

  // Verify outputs
  console.log('📋 Verifying build outputs...');
  const frontendDist = path.join(rootDir, 'frontend', 'dist');
  const binaryName = process.platform === 'win32' ? 'chemistryuno.exe' : 'chemistryuno';
  const binaryPath = path.join(rootDir, binaryName);

  let distExists = false;
  let binaryExists = false;

  if (fs.existsSync(frontendDist)) {
    console.log(`✅ Frontend dist: ${frontendDist}`);
    distExists = true;
  } else {
    console.warn(`⚠️  Frontend dist not found at: ${frontendDist}`);
  }

  if (fs.existsSync(binaryPath)) {
    console.log(`✅ Binary: ${binaryPath}`);
    binaryExists = true;
  } else {
    console.warn(`⚠️  Binary not found at: ${binaryPath}`);
  }

  console.log('\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  if (distExists && binaryExists) {
    console.log('✨ Build successful!');
    console.log('\nNext steps:');
    console.log(`  Run: ./${binaryName} (or chemistryuno.exe on Windows)`);
    console.log('  Then visit: http://localhost:8080');
  } else {
    console.log('⚠️  Build completed with warnings');
    console.log('Check output above for details');
  }
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n');

} catch (err) {
  console.error('❌ Build failed:', err.message);
  process.exit(1);
}
