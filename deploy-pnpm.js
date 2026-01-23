#!/usr/bin/env node
/**
 * Chemistry UNO - 生产环境部署脚本
 * 构建并启动应用
 */

const { execSync, spawn } = require('child_process');
const fs = require('fs');
const path = require('path');

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

function exec(command, description) {
  if (description) {
    log('cyan', `[→] ${description}`);
  }
  try {
    execSync(command, { stdio: 'inherit', shell: true });
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

async function deploy() {
  header('Chemistry UNO - 生产环境部署');
  
  log('cyan', `平台: ${process.platform} | Node.js: ${process.version}`);
  
  // 清理npm遗留配置（Linux/Mac）
  if (process.platform !== 'win32') {
    log('cyan', '\n清理npm配置...');
    try {
      execSync('npm config delete _-init-module 2>/dev/null || true', { stdio: 'ignore', shell: true });
      execSync('npm config delete init.module 2>/dev/null || true', { stdio: 'ignore', shell: true });
      execSync('npm config delete --global _-init-module 2>/dev/null || true', { stdio: 'ignore', shell: true });
      execSync('npm config delete --global init.module 2>/dev/null || true', { stdio: 'ignore', shell: true });
    } catch (e) {
      // 忽略错误
    }
  }
  
  // 检查 pnpm
  if (!exec('pnpm --version', '\n检查 pnpm...')) {
    log('red', '[✗] 请先安装 pnpm: npm install -g pnpm');
    process.exit(1);
  }
  log('green', '[✓] pnpm 已安装');
  
  // 安装依赖
  if (!exec('pnpm install', '\n安装依赖...')) {
    log('red', '[✗] 依赖安装失败');
    process.exit(1);
  }
  log('green', '[✓] 依赖安装完成');
  
  // 检查并创建环境变量文件
  const clientEnvFile = path.join(process.cwd(), 'client', '.env.production');
  if (!fs.existsSync(clientEnvFile)) {
    const envContent = `REACT_APP_ADMIN=your_admin_password_here
PORT=4000
BROWSER=none
SKIP_PREFLIGHT_CHECK=true
DISABLE_ESLINT_PLUGIN=true
`;
    fs.writeFileSync(clientEnvFile, envContent, 'utf-8');
    log('green', '[✓] 已创建 client/.env.production');
    log('yellow', '[!] 请修改 client/.env.production 中的 REACT_APP_ADMIN 为您的管理员密码');
  }
  
  // 构建应用
  if (!exec('pnpm run build', '\n构建应用...')) {
    log('red', '[✗] 构建失败');
    process.exit(1);
  }
  log('green', '[✓] 构建完成');
  
  // 成功信息
  console.log('');
  log('green', '═'.repeat(60));
  log('green', '  部署完成！');
  log('green', '═'.repeat(60));
  console.log('');
  
  log('cyan', '启动服务：');
  log('cyan', '  后端: cd server && node dist/index.js');
  log('cyan', '  前端: npx serve -s client/build -l 4000');
  console.log('');
  log('cyan', '访问地址: http://localhost:4000');
  console.log('');
  log('green', '正在启动服务...\n');

  // 跨平台命令处理
  const isWindows = process.platform === 'win32';
  
  // 启动后端
  const backendProcess = spawn(isWindows ? 'node.exe' : 'node', ['dist/index.js'], {
    cwd: path.join(process.cwd(), 'server'),
    stdio: 'inherit',
    shell: true
  });

  // 等待后端启动后启动前端
  setTimeout(() => {
    // 自动查找 build/serve.json，不需要显式指定
    const serveArgs = ['serve', '-s', 'build', '-l', '4000'];
    const frontendProcess = spawn(isWindows ? 'npx.cmd' : 'npx', serveArgs, {
      cwd: path.join(process.cwd(), 'client'),
      stdio: 'inherit',
      shell: true
    });

    console.log('');
    log('green', '✓ 服务已启动');
    log('cyan', '  前端: http://localhost:4000');
    log('cyan', '  后端: http://localhost:4001');
    log('cyan', '  管理面板: http://localhost:4000/admin');
    console.log('');
    
    // 处理退出
    process.on('SIGINT', () => {
      log('cyan', '\n正在停止服务...');
      backendProcess.kill();
      frontendProcess.kill();
      process.exit(0);
    });
  }, 2000);
}

deploy().catch(err => {
  log('red', `[✗] 部署失败: ${err.message}`);
  process.exit(1);
});
