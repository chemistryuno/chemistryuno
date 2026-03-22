#!/usr/bin/env node

/**
 * Start both frontend and backend development servers concurrently
 * Frontend: pnpm -C frontend dev (Vite on port 5000)
 * Backend: go run main.go (server on port 8080)
 */

const { spawn } = require('child_process');
const path = require('path');

const rootDir = __dirname;

console.log('🚀 Starting Chemistry UNO development environment...\n');
console.log('Frontend: http://localhost:5000');
console.log('Backend:  http://localhost:8080');
console.log('\nPress Ctrl+C to stop both servers\n');
console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n');

let frontendProcess = null;
let backendProcess = null;
let hasError = false;

// Start frontend (Vite dev server)
console.log('[FRONTEND] Starting Vite dev server...');
frontendProcess = spawn('pnpm', ['-C', 'frontend', 'dev'], {
  cwd: rootDir,
  stdio: ['inherit', 'inherit', 'inherit'],
  shell: process.platform === 'win32',
});

frontendProcess.on('error', (err) => {
  console.error('[FRONTEND] ❌ Failed to start:', err.message);
  hasError = true;
  process.exit(1);
});

frontendProcess.on('exit', (code) => {
  if (code !== 0 && !hasError) {
    console.error(`[FRONTEND] ❌ Exited with code ${code}`);
    hasError = true;
  }
  // Cleanup: kill backend when frontend exits
  if (backendProcess) {
    backendProcess.kill();
  }
  process.exit(code || 0);
});

// Give frontend time to start, then start backend
setTimeout(() => {
  console.log('[BACKEND]  Starting Go server...');
  backendProcess = spawn('go', ['run', 'main.go'], {
    cwd: rootDir,
    stdio: ['inherit', 'inherit', 'inherit'],
  });

  backendProcess.on('error', (err) => {
    console.error('[BACKEND] ❌ Failed to start:', err.message);
    if (frontendProcess) {
      frontendProcess.kill();
    }
    process.exit(1);
  });

  backendProcess.on('exit', (code) => {
    if (code !== 0 && !hasError) {
      console.error(`[BACKEND] ❌ Exited with code ${code}`);
      hasError = true;
    }
    // Cleanup: kill frontend when backend exits
    if (frontendProcess) {
      frontendProcess.kill();
    }
    process.exit(code || 0);
  });
}, 500);

// Handle Ctrl+C gracefully
process.on('SIGINT', () => {
  console.log('\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log('⏸️  Shutting down servers...');
  
  if (frontendProcess) {
    frontendProcess.kill('SIGINT');
  }
  if (backendProcess) {
    backendProcess.kill('SIGINT');
  }

  // Force exit after timeout if processes don't terminate
  setTimeout(() => {
    console.log('⚠️  Force closing servers');
    process.exit(0);
  }, 3000);
});

process.on('SIGTERM', () => {
  if (frontendProcess) {
    frontendProcess.kill('SIGTERM');
  }
  if (backendProcess) {
    backendProcess.kill('SIGTERM');
  }
  process.exit(0);
});
