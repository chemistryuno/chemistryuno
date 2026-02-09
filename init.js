#!/usr/bin/env node

const { spawn, execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

console.log('🧪 ============================================');
console.log('🧪 Chemistry UNO V1.0.0 Mendeleef - 项目初始化脚本');
console.log('🧪 ============================================\n');

const isWindows = process.platform === 'win32';

// 颜色输出函数
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
};

function log(message, color = colors.reset) {
  console.log(`${color}${message}${colors.reset}`);
}

function success(message) {
  log(`✅ ${message}`, colors.green);
}

function error(message) {
  log(`❌ ${message}`, colors.red);
}

function info(message) {
  log(`ℹ️  ${message}`, colors.blue);
}

function warning(message) {
  log(`⚠️  ${message}`, colors.yellow);
}

// 检查命令是否存在
function commandExists(command) {
  try {
    execSync(`${isWindows ? 'where' : 'which'} ${command}`, { stdio: 'ignore' });
    return true;
  } catch (error) {
    return false;
  }
}

// 执行命令并显示输出
function runCommand(command, args, options = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      stdio: 'inherit',
      shell: isWindows,
      ...options
    });

    child.on('close', (code) => {
      if (code === 0) {
        resolve();
      } else {
        reject(new Error(`命令失败: ${command} ${args.join(' ')}`));
      }
    });
  });
}

// 检查并安装依赖
async function checkAndInstallDependencies() {
  log('\n📦 检查系统依赖...', colors.cyan);

  // 检查 Node.js
  if (!commandExists('node')) {
    error('Node.js 未安装。请从 https://nodejs.org/ 下载并安装。');
    process.exit(1);
  }
  success('Node.js 已安装');

  // 检查 Go
  if (!commandExists('go')) {
    error('Go 未安装。请从 https://golang.org/dl/ 下载并安装。');
    process.exit(1);
  }
  success('Go 已安装');

  // 检查 pnpm
  if (!commandExists('pnpm')) {
    warning('pnpm 未安装，正在安装...');
    try {
      await runCommand('npm', ['install', '-g', 'pnpm']);
      success('pnpm 安装成功');
    } catch (err) {
      error('pnpm 安装失败: ' + err.message);
      process.exit(1);
    }
  } else {
    success('pnpm 已安装');
  }
}

// 安装前端依赖
async function installFrontendDependencies() {
  log('\n🎨 安装前端依赖...', colors.cyan);
  const frontendPath = path.join(__dirname, 'frontend');
  
  if (!fs.existsSync(path.join(frontendPath, 'package.json'))) {
    error('前端 package.json 不存在');
    process.exit(1);
  }

  try {
    await runCommand('pnpm', ['install'], { cwd: frontendPath });
    success('前端依赖安装成功');
  } catch (err) {
    error('前端依赖安装失败: ' + err.message);
    process.exit(1);
  }
}

// 安装后端依赖
async function installBackendDependencies() {
  log('\n🏗️  安装后端依赖...', colors.cyan);

  if (!fs.existsSync(path.join(__dirname, 'go.mod'))) {
    error('后端 go.mod 不存在');
    process.exit(1);
  }

  try {
    await runCommand('go', ['mod', 'tidy'], { cwd: __dirname });
    success('后端依赖安装成功');
  } catch (err) {
    error('后端依赖安装失败: ' + err.message);
    process.exit(1);
  }
}

// 初始化数据库
async function initializeDatabase() {
  log('\n🗄️  初始化数据库...', colors.cyan);
  
  try {
    info('正在创建数据库表和默认数据...');
    await runCommand('go', ['run', 'tools/init_db.go'], { 
      env: { ...process.env, INIT_DB: 'true' }
    });
    success('数据库初始化成功');
  } catch (err) {
    warning('数据库将在首次启动时自动初始化');
  }
}

// 创建配置文件
async function createConfigFiles() {
  log('\n⚙️  创建配置文件...', colors.cyan);

  // 创建 .env 文件（如果不存在）
  const envPath = path.join(__dirname, 'backend', '.env');
  if (!fs.existsSync(envPath)) {
    const envContent = `# Chemistry UNO Mendeleef 配置文件
# 后端配置
PORT=8080
JWT_SECRET=chemistry-uno-secret-key-change-in-production
SQLITE_PATH=./chemistryuno.db

# 前端配置  
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080/ws

# 开发环境配置
NODE_ENV=development
GIN_MODE=debug
`;
    
    fs.writeFileSync(envPath, envContent);
    success('创建 .env 配置文件');
  } else {
    info('.env 配置文件已存在');
  }

  // 创建 .gitignore 文件（如果不完整）
  const gitignorePath = path.join(__dirname, '.gitignore');
  const additionalIgnores = `
# 数据库文件
*.db
*.sqlite
*.sqlite3

# 日志文件
*.log
logs/

# 环境变量文件
.env.local
.env.*.local

# IDE 配置文件
.vscode/
.idea/
*.swp
*.swo

# 操作系统文件
.DS_Store
Thumbs.db

# 临时文件
tmp/
temp/
`;

  if (fs.existsSync(gitignorePath)) {
    const currentContent = fs.readFileSync(gitignorePath, 'utf8');
    if (!currentContent.includes('*.db')) {
      fs.appendFileSync(gitignorePath, additionalIgnores);
      success('更新 .gitignore 文件');
    } else {
      info('.gitignore 文件已是最新');
    }
  }
}

// 验证安装
async function validateInstallation() {
  log('\n🔍 验证安装...', colors.cyan);

  // 检查前端构建
  const frontendPath = path.join(__dirname, 'frontend');
  if (!fs.existsSync(path.join(frontendPath, 'node_modules'))) {
    error('前端依赖缺失');
    return false;
  }

  // 检查后端模块
  if (!fs.existsSync(path.join(__dirname, 'go.sum'))) {
    error('后端依赖缺失');
    return false;
  }

  success('所有组件验证通过');
  return true;
}

// 显示启动信息
function showStartupInfo() {
  log('\n🎉 初始化完成！', colors.green);
  log('\n📋 启动说明:', colors.cyan);
  log('  🚀 启动完整项目:       pnpm start', colors.bright);
  log('  🎨 仅启动前端:         pnpm run frontend', colors.bright);
  log('  🏗️  仅启动后端:         pnpm run backend', colors.bright);
  log('\n🌐 访问地址:', colors.cyan);
  log('  前端: http://localhost:5000', colors.bright);
  log('  后端: http://localhost:8080', colors.bright);
  log('\n👥 默认管理员账户:', colors.cyan);
  log('  用户名: admin@chemistryuno.com', colors.bright);
  log('  密码: admin123', colors.bright);
  log('\n📁 项目结构:', colors.cyan);
  log('  📂 frontend/     - Vue.js 前端应用', colors.bright);
  log('  📂 backend/      - Go 后端API服务', colors.bright);
  log('  📄 start.js      - 项目启动脚本', colors.bright);
  log('  📄 package.json  - 项目配置文件', colors.bright);
  log('\n🛠️  开发工具:', colors.cyan);
  log('  🔧 热重载:           自动检测文件变化', colors.bright);
  log('  📊 数据库管理:       内置SQLite数据库', colors.bright);
  log('  🔐 用户权限系统:     admin/co-worker/user三级权限', colors.bright);
  log('  🧪 化学反应库:       支持自定义化学反应数据', colors.bright);
  log('\n===================================', colors.magenta);
}

// 主初始化函数
async function main() {
  try {
    await checkAndInstallDependencies();
    await installFrontendDependencies();
    await installBackendDependencies();
    await createConfigFiles();
    await initializeDatabase();
    
    const isValid = await validateInstallation();
    if (isValid) {
      showStartupInfo();
    } else {
      error('初始化过程中遇到问题，请检查错误信息。');
      process.exit(1);
    }
    
  } catch (err) {
    error('初始化失败: ' + err.message);
    process.exit(1);
  }
}

// 运行初始化
main().catch((err) => {
  error('未知错误: ' + err.message);
  process.exit(1);
});