#!/usr/bin/env node

const fs = require('fs');
const http = require('http');
const path = require('path');
const { spawn, spawnSync } = require('child_process');

const rootDir = path.resolve(__dirname, '..');
const e2eDir = path.join(rootDir, 'tmp', 'e2e');
const logsDir = path.join(e2eDir, 'logs');
const dbPath = path.join(e2eDir, 'chemistryuno-e2e.db');
const backendBinaryPath = path.join(e2eDir, process.platform === 'win32' ? 'chemistryuno-e2e.exe' : 'chemistryuno-e2e');
const backendURL = process.env.E2E_BACKEND_URL || 'http://127.0.0.1:8080';
const frontendURL = process.env.E2E_BASE_URL || 'http://127.0.0.1:5000';

const testEnv = {
  ...process.env,
  DB_TYPE: 'sqlite',
  SQLITE_PATH: process.env.SQLITE_PATH || dbPath,
  JWT_SECRET: process.env.JWT_SECRET || 'chemistryuno-e2e-test-secret-000000000000',
  REDIS_ENABLED: process.env.REDIS_ENABLED || 'false',
  GIN_MODE: process.env.GIN_MODE || 'release',
  CHEM_SERVER_ORIGIN: process.env.CHEM_SERVER_ORIGIN || backendURL,
  E2E_BACKEND_URL: backendURL,
  E2E_BASE_URL: frontendURL,
};

let backendProcess = null;
let stoppingBackend = false;

function ensureDirs() {
  fs.mkdirSync(e2eDir, { recursive: true });
  fs.mkdirSync(logsDir, { recursive: true });
}

function removeIfExists(filePath) {
  fs.rmSync(filePath, { force: true });
}

function runChecked(name, command, args, options = {}) {
  console.log(`\n==> ${name}`);
  console.log(`$ ${[command, ...args].join(' ')}`);
  const result = spawnSync(command, args, {
    cwd: rootDir,
    stdio: 'inherit',
    shell: process.platform === 'win32',
    env: testEnv,
    ...options,
  });

  if (result.status !== 0) {
    throw new Error(`${name} failed with exit code ${result.status}`);
  }
}

function waitForURL(url, timeoutMs) {
  const startedAt = Date.now();

  return new Promise((resolve, reject) => {
    const tick = () => {
      const req = http.get(url, (res) => {
        res.resume();
        if (res.statusCode && res.statusCode >= 200 && res.statusCode < 500) {
          resolve();
          return;
        }
        retry();
      });

      req.on('error', retry);
      req.setTimeout(2_000, () => {
        req.destroy();
        retry();
      });
    };

    const retry = () => {
      if (Date.now() - startedAt > timeoutMs) {
        reject(new Error(`Timed out waiting for ${url}`));
        return;
      }
      setTimeout(tick, 1_000);
    };

    tick();
  });
}

async function startBackend() {
  const stdoutPath = path.join(logsDir, 'backend.stdout.log');
  const stderrPath = path.join(logsDir, 'backend.stderr.log');
  const stdout = fs.openSync(stdoutPath, 'w');
  const stderr = fs.openSync(stderrPath, 'w');

  console.log('\n==> Start backend test server');
  backendProcess = spawn(backendBinaryPath, [], {
    cwd: rootDir,
    stdio: ['ignore', stdout, stderr],
    shell: false,
    env: testEnv,
  });

  backendProcess.on('exit', (code) => {
    if (!stoppingBackend && code !== null && code !== 0) {
      console.error(`Backend test server exited with code ${code}. See ${stdoutPath} and ${stderrPath}.`);
    }
  });

  await waitForURL(`${backendURL}/api/health`, 90_000);
  console.log(`Backend test server is ready at ${backendURL}`);
}

function stopBackend() {
  if (backendProcess && !backendProcess.killed) {
    stoppingBackend = true;
    if (process.platform === 'win32') {
      spawnSync('taskkill', ['/pid', String(backendProcess.pid), '/T', '/F'], {
        stdio: 'ignore',
      });
    } else {
      backendProcess.kill('SIGTERM');
    }
  }
}

async function main() {
  ensureDirs();
  removeIfExists(testEnv.SQLITE_PATH);
  removeIfExists(`${testEnv.SQLITE_PATH}-wal`);
  removeIfExists(`${testEnv.SQLITE_PATH}-shm`);

  try {
    runChecked('Reset and seed e2e database', ['go'][0], ['run', '-tags', 'scripts', 'backend/scripts/reset_test_db.go']);
    runChecked('Build backend e2e binary', 'go', ['build', '-o', backendBinaryPath, 'main.go']);
    await startBackend();
    runChecked('Run Playwright full-stack tests', 'pnpm', ['-C', 'frontend', 'test:e2e']);
    console.log('\nFull-stack e2e workflow completed successfully.');
  } catch (error) {
    console.error(`\nFull-stack e2e workflow failed: ${error.message}`);
    console.error(`Service logs: ${logsDir}`);
    process.exitCode = 1;
  } finally {
    stopBackend();
  }
}

main();
