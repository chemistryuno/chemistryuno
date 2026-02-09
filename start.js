#!/usr/bin/env node

const { spawn, execSync } = require('child_process');
const path = require('path');
const os = require('os');
const fs = require('fs');

console.log('🚀 正在启动 Chemistry UNO V1.0.0 Mendeleef (DEVELOPMENT)...\n');

const isWindows = process.platform === 'win32';

// 检查Redis是否可用（可选功能）
function checkRedis() {
  try {
    const net = require('net');
    const redisAddr = process.env.REDIS_ADDR || 'localhost:6379';
    const [host, port] = redisAddr.split(':');

    const client = net.createConnection({ host, port: parseInt(port) }, () => {
      console.log('✅ Redis连接正常:', redisAddr);
      client.end();
    });

    client.on('error', () => {
      console.log('⚠️  Redis未连接，缓存功能已禁用（这不影响核心功能）');
      if (!process.env.REDIS_ADDR) {
        console.log('💡 提示: 如需启用Redis，请设置环境变量 REDIS_ADDR=localhost:6379');
      }
    });

    client.setTimeout(1000);
    client.on('timeout', () => {
      client.destroy();
    });
  } catch (err) {
    console.log('⚠️  Redis检测失败（可选功能）');
  }
}

// 检查并加载.env文件（如果存在）
function loadEnvFile() {
  const envPath = path.join(__dirname, 'backend', '.env');
  if (fs.existsSync(envPath)) {
    const envContent = fs.readFileSync(envPath, 'utf8');
    envContent.split('\n').forEach(line => {
      const [key, ...valueParts] = line.split('=');
      if (key && valueParts.length > 0) {
        const value = valueParts.join('=').trim();
        if (!process.env[key]) {
          process.env[key] = value;
        }
      }
    });
    console.log('✅ 已加载环境配置文件 (.env)');
  }
}

console.log('🔍 检查环境配置...');
loadEnvFile();
checkRedis();
console.log('');

// 启动后端 (Go)
console.log('📦 启动后端服务器...');

// 禁用CGO以避免MinGW链接器问题（modernc.org/sqlite是纯Go实现，不需要CGO）
const backendEnv = Object.assign({}, process.env, {
  'CGO_ENABLED': '0'
});

const backendProcess = spawn('go', ['run', 'main.go'], {
  cwd: __dirname,
  shell: true,
  stdio: 'inherit',
  env: backendEnv
});

backendProcess.on('error', (err) => {
  console.error('❌ 后端启动失败:', err.message);
  process.exit(1);
});

backendProcess.on('exit', (code) => {
  if (code !== 0 && code !== null) {
    console.error(`❌ 后端进程异常退出，退出码: ${code}`);
    process.exit(1);
  }
});

// 等待 5 秒后启动前端，确保后端有足够时间编译并启动
setTimeout(() => {
  console.log('\n🎨 启动前端开发服务器...');
  const frontendPath = path.join(__dirname, 'frontend');
  
  // 智能寻找 pnpm 命令
  const pnpmCmd = isWindows ? 'pnpm.cmd' : 'pnpm';
  
  const frontendProcess = spawn(pnpmCmd, ['dev'], {
    cwd: frontendPath,
    shell: true,
    stdio: 'inherit'
  });

  frontendProcess.on('error', (err) => {
    console.error('❌ 前端启动失败:', err.message);
    console.log('💡 提示: 请确保已安装 pnpm (npm install -g pnpm)');
    backendProcess.kill();
    process.exit(1);
  });

  frontendProcess.on('exit', (code) => {
    if (code !== 0 && code !== null) {
      console.error(`❌ 前端进程异常退出，退出码: ${code}`);
      backendProcess.kill();
      process.exit(1);
    }
  });

  // 统一退出逻辑
  const quit = () => {
    console.log('\n\n🛑 正在停止所有服务...');
    backendProcess.kill();
    frontendProcess.kill();
    process.exit(0);
  };

  process.on('SIGINT', quit);
  process.on('SIGTERM', quit);

  // 前端端口通常在 package.json 或 vite.config.ts 中定义
  // 这里同步为默认的 5000 (根据项目配置)
  const frontendPort = 5000;
  const backendPort = 8080;

  // 延时显示访问地址
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

    console.log('\n✨ 服务已就绪！');
    console.log(`📍 本地访问: http://localhost:${frontendPort}`);
    if (lanIp !== 'localhost') {
      console.log(`🌐 局域网访问: http://${lanIp}:${frontendPort}`);
    }
    console.log(`🔗 后端 API: http://localhost:${backendPort}`);
    console.log('\n按 Ctrl+C 停止所有服务');
  }, 2000);

}, 1500);

