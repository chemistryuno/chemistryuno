#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');
const readline = require('readline');

const rootDir = __dirname;

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
  red: "\x1b[31b",
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
  log("SYSTEM", "Launching Go backend...", colors.yellow);
  backendProcess = spawn('go', ['run', 'main.go'], {
    cwd: rootDir,
    shell: true,
    env: { ...process.env, FORCE_COLOR: 'true' }
  });

  backendProcess.stdout.on('data', createModuleLogger("BACKEND", colors.blue));
  backendProcess.stderr.on('data', createModuleLogger("BACKEND", colors.red));

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
