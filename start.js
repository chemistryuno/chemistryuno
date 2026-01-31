#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');

console.log('🚀 正在启动 Chemistry UNO Alpha...\n');

const isWindows = process.platform === 'win32';

// 启动后端 (Go)
console.log('📦 启动后端服务器...');
const backendPath = path.join(__dirname, 'backend');

// 设置Go环境变量，禁用ASLR以兼容旧版GCC
const backendEnv = Object.assign({}, process.env, {
  'CGO_LDFLAGS': '-g -O2'
});

const backendProcess = spawn('go', ['run', 'main.go'], {
  cwd: backendPath,
  shell: isWindows,
  stdio: 'inherit',
  env: backendEnv
});

backendProcess.on('error', (err) => {
  console.error('❌ 后端启动失败:', err.message);
  process.exit(1);
});

// 等待1秒后启动前端
setTimeout(() => {
  console.log('\n🎨 启动前端开发服务器...');
  const frontendPath = path.join(__dirname, 'frontend');
  const frontendProcess = spawn('pnpm', ['dev'], {
    cwd: frontendPath,
    shell: isWindows,
    stdio: 'inherit'
  });

  frontendProcess.on('error', (err) => {
    console.error('❌ 前端启动失败:', err.message);
    console.log('💡 提示: 请确保已安装 pnpm (npm install -g pnpm)');
    backendProcess.kill();
    process.exit(1);
  });

  // 监听退出信号
  process.on('SIGINT', () => {
    console.log('\n\n🛑 正在停止服务...');
    backendProcess.kill();
    frontendProcess.kill();
    process.exit(0);
  });

  process.on('SIGTERM', () => {
    backendProcess.kill();
    frontendProcess.kill();
    process.exit(0);
  });

}, 1000);

console.log('\n✨ 启动完成！');
// 尝试获取本机局域网 IPv4 地址并输出可访问的 URL
const os = require('os')
const nets = os.networkInterfaces()
let lanIp = 'localhost'
for (const name of Object.keys(nets)) {
  for (const net of nets[name]) {
    // 跳过内部和非 IPv4
    if (net.family === 'IPv4' && !net.internal) {
      lanIp = net.address
      break
    }
  }
  if (lanIp !== 'localhost') break
}

console.log('📍 前端地址: http://' + lanIp + ':3000 (或 http://localhost:3000)');
console.log('📍 后端地址: http://' + lanIp + ':8080 (或 http://localhost:8080)');
console.log('\n按 Ctrl+C 停止所有服务\n');
