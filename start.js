#!/usr/bin/env node

const fs = require('fs');
const { spawn } = require('child_process');
const path = require('path');

const rootDir = __dirname;
const airConfigPath = path.join(rootDir, '.air.toml');

// Simple color function (ANSI escape codes)
const colors = {
  reset: "\x1b[0m",
  bright: "\x1b[1m",
  dim: "\x1b[2m",
  green: "\x1b[32m",
  blue: "\x1b[34m",
  cyan: "\x1b[36m",
  yellow: "\x1b[33m",
  magenta: "\x1b[35m",
  red: "\x1b[31m",
};

function log(module, message, color = colors.reset) {
  const timestamp = new Date().toLocaleTimeString('zh-CN', { hour12: false });
  const prefix = `${colors.dim}[${timestamp}]${colors.reset} ${color}${colors.bright}[${module}]${colors.reset}`;
  console.log(`${prefix} ${message}`);
}

console.clear();
console.log(`${colors.cyan}${colors.bright}==================================================${colors.reset}`);
console.log(`${colors.cyan}${colors.bright}   Chemistry UNO - Mendeleev Development Shell    ${colors.reset}`);
console.log(`${colors.cyan}${colors.bright}==================================================${colors.reset}\n`);

let frontendProcess = null;
let backendProcess = null;

function resolveAirCommand() {
  const exeName = process.platform === 'win32' ? 'air.exe' : 'air';
  const candidates = [];

  if (process.env.AIR_BIN) {
    candidates.push(process.env.AIR_BIN);
  }

  candidates.push(exeName);

  if (process.env.GOPATH) {
    candidates.push(path.join(process.env.GOPATH, 'bin', exeName));
  }

  if (process.env.USERPROFILE) {
    candidates.push(path.join(process.env.USERPROFILE, 'go', 'bin', exeName));
  }

  if (process.env.HOME) {
    candidates.push(path.join(process.env.HOME, 'go', 'bin', exeName));
  }

  for (const candidate of candidates) {
    if (!candidate) {
      continue;
    }

    if (!path.isAbsolute(candidate)) {
      return candidate;
    }

    if (fs.existsSync(candidate)) {
      return candidate;
    }
  }

  return exeName;
}

// Helper to pipe and prefix child process output
function createModuleLogger(moduleName, color) {
  return (data) => {
    const lines = data.toString().split('\n');
    lines.forEach(line => {
      if (line.trim()) {
        log(moduleName, line, color);
      }
    });
  };
}

log("SYSTEM", "Starting services...", colors.yellow);
log("SYSTEM", "Frontend dev server: http://127.0.0.1:5000", colors.green);
log("SYSTEM", "Backend API server:   http://127.0.0.1:8080", colors.blue);
log("SYSTEM", "Open the frontend URL above for Vite hot updates.", colors.yellow);

// Start frontend
frontendProcess = spawn('pnpm', ['-C', 'frontend', 'dev'], {
  cwd: rootDir,
  shell: true,
  env: { ...process.env, FORCE_COLOR: 'true' }
});

frontendProcess.stdout.on('data', createModuleLogger("FRONTEND", colors.green));
frontendProcess.stderr.on('data', createModuleLogger("FRONTEND", colors.red));

// Start backend
setTimeout(() => {
  const airCommand = resolveAirCommand();
  log("SYSTEM", "Launching Go backend with air...", colors.yellow);
  backendProcess = spawn(airCommand, ['-c', airConfigPath], {
    cwd: rootDir,
    shell: false,
    env: { ...process.env, FORCE_COLOR: 'true' }
  });

  backendProcess.stdout.on('data', createModuleLogger("BACKEND", colors.blue));
  backendProcess.stderr.on('data', createModuleLogger("BACKEND", colors.red));

  backendProcess.on('error', (err) => {
    if (err.code === 'ENOENT') {
      log("BACKEND", "air was not found. Run `pnpm run air:install` (or `go install github.com/air-verse/air@latest`) and try again.", colors.red);
    } else {
      log("BACKEND", `Failed to launch air: ${err.message}`, colors.red);
    }

    if (frontendProcess) frontendProcess.kill();
    process.exit(1);
  });

  backendProcess.on('exit', (code) => {
    if (code !== 0) log("BACKEND", `Service stopped with code ${code}`, colors.red);
    process.exit(code || 0);
  });
}, 1500);

frontendProcess.on('exit', (code) => {
  if (code !== 0) log("FRONTEND", `Service stopped with code ${code}`, colors.red);
  if (backendProcess) backendProcess.kill();
  process.exit(code || 0);
});

process.on('SIGINT', () => {
  console.log("\n");
  log("SYSTEM", "Shutting down services...", colors.magenta);
  if (frontendProcess) frontendProcess.kill();
  if (backendProcess) backendProcess.kill();
  process.exit(0);
});
