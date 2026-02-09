#!/usr/bin/env node

const { spawn, execSync } = require('child_process');
const path = require('path');
const os = require('os');
const fs = require('fs');

console.log('🧪 正在启动 Chemistry UNO V1.0.0 Mendeleef (TEST ENVIRONMENT)...\n');

const isWindows = process.platform === 'win32';

// 颜色输出函数
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  cyan: '\x1b[36m',
};

function log(message, color = colors.reset) {
  console.log(`${color}${message}${colors.reset}`);
}

// 1. 初始化测试数据库
console.log('🗄️  初始化测试数据库...');
try {
  execSync('go run scripts/setup_test_env.go', {
    cwd: path.join(__dirname, 'backend'),
    stdio: 'inherit',
    shell: true
  });
  log('✅ 测试数据库初始化成功\n', colors.green);
} catch (err) {
  log('❌ 测试数据库初始化失败', colors.red);
  console.error(err.message);
  process.exit(1);
}

// 2. 选择后端启动方式
const useCompiledBinary = process.argv.includes('--build');

let backendProcess;

if (useCompiledBinary) {
  // 方式A: 编译后运行
  console.log('📦 编译后端（生产模式）...');
  try {
    const backendBinary = isWindows ? 'chemistryuno.exe' : 'chemistryuno';
    const buildCmd = `go build -o ${backendBinary} main.go`;

    execSync(buildCmd, {
      cwd: __dirname,
      stdio: 'inherit',
      shell: true,
      env: { ...process.env, CGO_ENABLED: '0' }
    });

    log('✅ 后端编译成功\n', colors.green);

    console.log('🚀 启动后端服务器（编译版本）...');
    const binaryPath = isWindows ? `.\\${backendBinary}` : `./${backendBinary}`;

    backendProcess = spawn(binaryPath, [], {
      cwd: __dirname,
      shell: true,
      stdio: 'inherit'
    });
  } catch (err) {
    log('❌ 后端编译失败', colors.red);
    console.error(err.message);
    process.exit(1);
  }
} else {
  // 方式B: 直接运行（开发模式，更快）
  console.log('🚀 启动后端服务器（开发模式）...');

  const backendEnv = Object.assign({}, process.env, {
    'CGO_ENABLED': '0',
    'GIN_MODE': 'debug'
  });

  backendProcess = spawn('go', ['run', 'main.go'], {
    cwd: __dirname,
    shell: true,
    stdio: 'inherit',
    env: backendEnv
  });
}

backendProcess.on('error', (err) => {
  log('❌ 后端启动失败: ' + err.message, colors.red);
  process.exit(1);
});

backendProcess.on('exit', (code) => {
  if (code !== 0 && code !== null) {
    log(`❌ 后端进程异常退出，退出码: ${code}`, colors.red);
    process.exit(1);
  }
});

// 3. 等待后端启动完成后启动前端
setTimeout(() => {
  console.log('\n🎨 启动前端开发服务器...');
  const frontendPath = path.join(__dirname, 'frontend');

  const pnpmCmd = isWindows ? 'pnpm.cmd' : 'pnpm';

  const frontendProcess = spawn(pnpmCmd, ['dev'], {
    cwd: frontendPath,
    shell: true,
    stdio: 'inherit'
  });

  frontendProcess.on('error', (err) => {
    log('❌ 前端启动失败: ' + err.message, colors.red);
    log('💡 提示: 请确保已安装 pnpm (npm install -g pnpm)', colors.yellow);
    backendProcess.kill();
    process.exit(1);
  });

  frontendProcess.on('exit', (code) => {
    if (code !== 0 && code !== null) {
      log(`❌ 前端进程异常退出，退出码: ${code}`, colors.red);
      backendProcess.kill();
      process.exit(1);
    }
  });

  // 统一退出逻辑
  const quit = () => {
    console.log('\n\n🛑 正在停止测试环境...');
    backendProcess.kill();
    frontendProcess.kill();
    process.exit(0);
  };

  process.on('SIGINT', quit);
  process.on('SIGTERM', quit);

  // 4. 延时显示测试环境信息
  setTimeout(() => {
    const nets = os.networkInterfaces();
    let lanIp = 'localhost';
    for (const name of Object.keys(nets)) {
      for (const net of nets[name]) {
        if (net.family === 'IPv4' && !net.internal) {
          lanIp = net.address;
          break;
        }
      }
      if (lanIp !== 'localhost') break;
    }

    log('\n✨ 测试环境已就绪！', colors.green);
    log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━', colors.cyan);
    log('📍 访问地址:', colors.bright);
    log(`   本地访问:  http://localhost:5000`, colors.reset);
    if (lanIp !== 'localhost') {
      log(`   局域网访问: http://${lanIp}:5000`, colors.reset);
    }
    log(`   后端 API:  http://localhost:8080`, colors.reset);
    log('', colors.reset);
    log('👤 测试账户:', colors.bright);
    log('   管理员:    admin@chemistryuno.com / admin123', colors.reset);
    log('   普通用户:  test@example.com / test123', colors.reset);
    log('', colors.reset);
    log('🗄️  数据库:', colors.bright);
    log('   测试数据库已初始化，包含示例数据', colors.reset);
    log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━', colors.cyan);
    log('\n💡 提示:', colors.yellow);
    log('   - 测试环境使用独立的测试数据库', colors.reset);
    log('   - 按 Ctrl+C 停止测试环境', colors.reset);
    if (!useCompiledBinary) {
      log('   - 使用 --build 参数可以编译后运行（更接近生产环境）', colors.reset);
    }
    log('', colors.reset);
  }, 2000);

}, 1500);
