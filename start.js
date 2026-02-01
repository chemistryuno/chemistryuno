#!/usr/bin/env node

const { spawn, execSync } = require('child_process');
const path = require('path');
const os = require('os');

console.log('🚀 正在启动 Chemistry UNO Alpha...\n');

const isWindows = process.platform === 'win32';
const shell = isWindows ? 'cmd.exe' : true;

// 启动后端 (Go)
console.log('📦 启动后端服务器...');
const backendPath = path.join(__dirname, 'backend');

// 设置Go环境变量，禁用ASLR以兼容旧版GCC (可选，根据环境设置)
const backendEnv = Object.assign({}, process.env, {
  'CGO_LDFLAGS': '-g -O2'
});

const backendProcess = spawn('go', ['run', 'main.go'], {
  cwd: backendPath,
  shell: true,
  stdio: 'inherit',
  env: backendEnv
});

backendProcess.on('error', (err) => {
  console.error('❌ 后端启动失败:', err.message);
  process.exit(1);
});

// 等待1.5秒后启动前端，确保后端先建立基础
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

