#!/usr/bin/env node

/**
 * Initialize project: setup .env, install dependencies, initialize database
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const rootDir = __dirname;
const envPath = path.join(rootDir, '.env');
const envExamplePath = path.join(rootDir, '.env.example');
const frontendDir = path.join(rootDir, 'frontend');
const dbPath = path.join(rootDir, 'chemistryuno.db');

console.log('馃殌 Initializing Chemistry UNO project...\n');

// Step 1: Setup .env file
console.log('馃摑 Step 1: Setting up .env file');
if (!fs.existsSync(envPath)) {
  if (fs.existsSync(envExamplePath)) {
    try {
      fs.copyFileSync(envExamplePath, envPath);
      console.log('鉁?Created .env from .env.example');
    } catch (err) {
      console.error('鉂?Failed to copy .env.example:', err.message);
      process.exit(1);
    }
  } else {
    console.error('鉂?.env.example not found');
    process.exit(1);
  }
} else {
  console.log('鉁?.env already exists, skipping');
}

// Step 2: Install frontend dependencies
console.log('\n馃摝 Step 2: Installing frontend dependencies');
if (!fs.existsSync(path.join(frontendDir, 'node_modules'))) {
  try {
    console.log('   Running: pnpm -C frontend install');
    execSync('pnpm -C frontend install', { stdio: 'inherit', cwd: rootDir });
    console.log('鉁?Frontend dependencies installed');
  } catch (err) {
    console.error('鉂?Failed to install frontend dependencies:', err.message);
    process.exit(1);
  }
} else {
  console.log('鉁?Frontend node_modules already exists, skipping');
}

// Step 3: Initialize database
console.log('\n馃梽锔? Step 3: Initializing database');
if (!fs.existsSync(dbPath)) {
  try {
    console.log('   Running: go run backend/scripts/init_db.go');
    execSync('go run backend/scripts/init_db.go', { stdio: 'inherit', cwd: rootDir });
    console.log('鉁?Database initialized');
  } catch (err) {
    console.error('鉂?Failed to initialize database:', err.message);
    process.exit(1);
  }
} else {
  console.log('鉁?Database already exists at chemistryuno.db, skipping initialization');
}

console.log('\n鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣');
console.log('鉁?Initialization complete!');
console.log('');
console.log('Next steps:');
console.log('  1. Review .env and update configuration if needed');
console.log('  2. Run: pnpm run air:install  (installs backend hot reload tool)');
console.log('  3. Run: pnpm start  (to start development servers)');
console.log('鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣鈹佲攣\n');
