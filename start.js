#!/usr/bin/env node
/**
 * Chemistry UNO - 快速启动脚本
 * 开发环境一键启动（跨平台：Windows/Linux/Mac）
 */

const { execSync, spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

const colors = {
  reset: '\x1b[0m',
  green: '\x1b[32m',
  red: '\x1b[31m',
  cyan: '\x1b[36m',
  magenta: '\x1b[35m'
};

function log(color, message) {
  console.log(`${colors[color]}${message}${colors.reset}`);
}

function exec(command, silent = false) {
  try {
    execSync(command, { stdio: silent ? 'pipe' : 'inherit', shell: true });
    return true;
  } catch (err) {
    return false;
  }
}

function header(title) {
  console.log('');
  log('magenta', '═'.repeat(60));
  log('magenta', `  ${title}`);
  log('magenta', '═'.repeat(60));
  console.log('');
}

async function start() {
  header('Chemistry UNO - 快速启动');
  
  log('cyan', `平台: ${process.platform} | Node.js: ${process.version}`);
  
  // 检查 pnpm
  if (!exec('pnpm --version', true)) {
    log('red', '[✗] 请先安装 pnpm: npm install -g pnpm');
    process.exit(1);
  }
  
  // 检查依赖是否已安装
  const hasNodeModules = fs.existsSync(path.join(__dirname, 'node_modules'));
  
  if (!hasNodeModules) {
    log('cyan', '首次运行，正在安装依赖...');
    if (!exec('pnpm install')) {
      log('red', '[✗] 依赖安装失败');
      process.exit(1);
    }
    log('green', '[✓] 依赖安装完成\n');
  }
  
  log('green', '正在启动开发服务器...\n');
  log('cyan', '  前端: http://localhost:3000');
  log('cyan', '  后端: http://localhost:4001');
  log('cyan', '  管理面板: http://localhost:3000/admin');
  console.log('');
  log('cyan', '按 Ctrl+C 停止服务');
  console.log('');
  
  // 启动开发服务器（跨平台）
  const isWindows = process.platform === 'win32';
  const devProcess = spawn(isWindows ? 'pnpm.cmd' : 'pnpm', ['run', 'dev'], {
    stdio: 'inherit',
    shell: true
  });
  
  // 处理退出
  process.on('SIGINT', () => {
    log('cyan', '\n正在停止服务...');
    devProcess.kill();
    process.exit(0);
  });
  
  process.on('SIGTERM', () => {
    devProcess.kill();
    process.exit(0);
  });
}

start().catch(err => {
  log('red', `[✗] 启动失败: ${err.message}`);
  process.exit(1);
});
