#!/usr/bin/env node

const { execSync, spawn } = require('child_process');
const path = require('path');
const fs = require('fs');
const os = require('os');

console.log('🏗️  构建 Chemistry UNO V1.0.0 Mendeleef (PRODUCTION)...\n');

const isWindows = process.platform === 'win32';

// 检查并创建dist目录
const distDir = path.join(__dirname, 'dist');
if (!fs.existsSync(distDir)) {
  fs.mkdirSync(distDir, { recursive: true });
}

// 1. 构建后端
console.log('📦 构建后端 (Go)...');
const backendPath = path.join(__dirname, 'backend');
const backendBinary = isWindows ? 'chemistryuno.exe' : 'chemistryuno';
const backendOutput = path.join(backendPath, backendBinary);

try {
  const buildCmd = isWindows
    ? `cd ${backendPath} && set CGO_ENABLED=0 && go build -ldflags="-s -w" -o ${backendBinary} main.go`
    : `cd ${backendPath} && CGO_ENABLED=0 go build -ldflags="-s -w" -o ${backendBinary} main.go`;

  execSync(buildCmd, { stdio: 'inherit', shell: true });

  const stats = fs.statSync(backendOutput);
  const sizeMB = (stats.size / (1024 * 1024)).toFixed(2);
  console.log(`✅ 后端构建成功: ${backendBinary} (${sizeMB} MB)\n`);
} catch (err) {
  console.error('❌ 后端构建失败:', err.message);
  process.exit(1);
}

// 2. 构建前端
console.log('🎨 构建前端 (Vue + Vite)...');
const frontendPath = path.join(__dirname, 'frontend');

try {
  const buildCmd = isWindows
    ? `cd ${frontendPath} && pnpm build`
    : `cd ${frontendPath} && pnpm build`;

  execSync(buildCmd, { stdio: 'inherit', shell: true });
  console.log('✅ 前端构建成功: frontend/dist/\n');
} catch (err) {
  console.error('❌ 前端构建失败:', err.message);
  process.exit(1);
}

// 3. 复制构建产物到dist目录
console.log('📂 整理构建产物...');
try {
  // 复制后端二进制
  const distBackend = path.join(distDir, backendBinary);
  fs.copyFileSync(backendOutput, distBackend);

  // 复制前端构建产物
  const frontendDist = path.join(frontendPath, 'dist');
  const distFrontend = path.join(distDir, 'frontend');

  // 递归复制目录
  function copyDir(src, dest) {
    if (!fs.existsSync(dest)) {
      fs.mkdirSync(dest, { recursive: true });
    }
    const entries = fs.readdirSync(src, { withFileTypes: true });
    for (let entry of entries) {
      const srcPath = path.join(src, entry.name);
      const destPath = path.join(dest, entry.name);
      if (entry.isDirectory()) {
        copyDir(srcPath, destPath);
      } else {
        fs.copyFileSync(srcPath, destPath);
      }
    }
  }

  copyDir(frontendDist, distFrontend);

  // 创建启动脚本
  const startScript = isWindows
    ? `@echo off
echo Starting Chemistry UNO V1.0.0 Mendeleef...
start ${backendBinary}
echo Server started at http://localhost:8080
echo.
echo Press any key to stop the server...
pause > nul
taskkill /F /IM ${backendBinary} > nul 2>&1
`
    : `#!/bin/bash
echo "Starting Chemistry UNO V1.0.0 Mendeleef..."
./${backendBinary} &
PID=$!
echo "Server started at http://localhost:8080"
echo "Press Ctrl+C to stop..."
trap "kill $PID" EXIT
wait $PID
`;

  const startScriptPath = path.join(distDir, isWindows ? 'start.bat' : 'start.sh');
  fs.writeFileSync(startScriptPath, startScript);
  if (!isWindows) {
    fs.chmodSync(startScriptPath, '755');
  }

  // 创建README
  const readme = `# Chemistry UNO V1.0.0 Mendeleef

## 运行说明

### Windows
双击 \`start.bat\` 或在命令行运行：
\`\`\`
start.bat
\`\`\`

### Linux/macOS
\`\`\`bash
./start.sh
\`\`\`

### 手动运行
\`\`\`bash
./${backendBinary}
\`\`\`

## 访问地址

- **主页**: http://localhost:8080
- **API文档**: http://localhost:8080/api

## 默认账户

- 用户名: admin
- 密码: admin123

## 配置

### 环境变量（可选）

- \`JWT_SECRET\`: JWT密钥（如未设置会自动生成）
- \`REDIS_ADDR\`: Redis地址（如 localhost:6379）
- \`REDIS_PASSWORD\`: Redis密码
- \`DB_TYPE\`: 数据库类型（sqlite 或 mysql，默认 sqlite）
- \`SQLITE_PATH\`: SQLite数据库路径（默认 ./chemistryuno.db）
- \`MYSQL_DSN\`: MySQL连接字符串

### Redis配置（可选）

Redis用于缓存和会话管理，不配置也可以正常运行。

如需启用Redis：
1. 安装Redis
2. 设置环境变量 \`REDIS_ADDR=localhost:6379\`
3. 重启服务

## 技术栈

- 后端: Go 1.20+ (Gin框架)
- 前端: Vue 3 + TypeScript + Vite
- 数据库: SQLite (纯Go实现，无需CGO)
- 实时通信: WebSocket

## 支持

项目地址: https://github.com/yourusername/chemistryuno

---
构建日期: ${new Date().toLocaleString('zh-CN')}
`;

  fs.writeFileSync(path.join(distDir, 'README.md'), readme);

  console.log('✅ 构建产物已整理到 dist/ 目录\n');
} catch (err) {
  console.error('❌ 整理构建产物失败:', err.message);
  process.exit(1);
}

// 4. 显示构建信息
console.log('✨ 构建完成！\n');
console.log('📁 构建产物位置:');
console.log(`   - 后端二进制: dist/${backendBinary}`);
console.log(`   - 前端静态文件: dist/frontend/`);
console.log(`   - 启动脚本: dist/${isWindows ? 'start.bat' : 'start.sh'}`);
console.log(`   - 说明文档: dist/README.md\n`);

const distSize = execSync(`du -sh ${distDir} 2>/dev/null || echo "N/A"`, {
  encoding: 'utf8',
  shell: true
}).trim();
console.log(`📊 总大小: ${distSize.split('\t')[0] || 'N/A'}\n`);

console.log('🚀 运行方式:');
if (isWindows) {
  console.log('   cd dist && start.bat');
} else {
  console.log('   cd dist && ./start.sh');
}
