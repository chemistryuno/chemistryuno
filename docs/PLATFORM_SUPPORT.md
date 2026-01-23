# 🖥�?跨平台支持说�?

化学UNO 完全支持 Windows、Linux �?macOS 系统�?

## 📋 可用脚本

| 脚本文件 | 系统 | 用�?|
|---------|------|------|
| `start.bat` | Windows | 一键启动开发环�?|
| `start.sh` | Linux/Mac | 一键启动开发环�?|
| `start.js` | 所有系�?| Node.js 启动脚本（跨平台�?|
| `deploy.bat` | Windows | 一键生产部�?|
| `deploy.sh` | Linux/Mac | 一键生产部�?|
| `deploy-pnpm.js` | 所有系�?| Node.js 部署脚本（跨平台�?|

## 🚀 快速启�?

### Windows

```cmd
# 方式一：双击运�?
start.bat

# 方式二：命令�?
node start.js
```

### Linux / macOS

```bash
# 首次运行需要添加执行权�?
chmod +x start.sh deploy.sh

# 启动开发环�?
./start.sh

# 或使�?Node.js 脚本
node start.js
```

### 通用方式（所有系统）

```bash
# 安装依赖（首次运行）
pnpm install

# 启动开发服务器
pnpm run dev

# 或直接使�?
node start.js
```

## 📦 生产部署

### Windows

```cmd
# 方式一：双击运�?
deploy.bat

# 方式二：命令�?
pnpm run deploy
```

### Linux / macOS

```bash
# 运行部署脚本
./deploy.sh

# 或使�?pnpm
pnpm run deploy
```

## 🔧 平台特定说明

### Windows

- 脚本使用 UTF-8 编码（`chcp 65001`）以正确显示中文
- 批处理文件（`.bat`）可直接双击运行
- 支持 PowerShell �?CMD

### Linux / macOS

- Shell 脚本（`.sh`）需要执行权限：`chmod +x *.sh`
- 使用 ANSI 颜色代码显示彩色输出
- 支持 bash 和其他兼�?shell

### 跨平�?

- 所�?`.js` 脚本都是跨平台的
- 自动检测操作系统并调整命令
- Windows 使用 `.cmd` 后缀（如 `pnpm.cmd`�?
- Linux/Mac 直接使用命令（如 `pnpm`�?

## 📱 端口配置

默认端口�?
- **开发环�?*：前�?3000，后�?4001
- **生产环境**：前�?4000，后�?4001

### 修改端口

**前端端口**�?
```bash
# 创建或编�?client/.env
PORT=3001
```

**后端端口**�?
编辑 [server/index.ts](server/index.ts)，修�?`PORT` 常量

## 🛠�?常见问题

### Linux/Mac: Permission denied

```bash
# 添加执行权限
chmod +x start.sh deploy.sh
```

### Windows: 脚本无法运行

```cmd
# 确保以管理员身份运行
# 或使�?PowerShell 执行策略
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### 命令未找�?

```bash
# 确保 Node.js �?pnpm 已安�?
node --version
pnpm --version

# 安装 pnpm
npm install -g pnpm
```

### 端口被占�?

```bash
# Linux/Mac: 查找并终止占用端口的进程
lsof -ti:3000 | xargs kill -9
lsof -ti:4001 | xargs kill -9

# Windows: 使用任务管理器或命令
netstat -ano | findstr :3000
taskkill /PID <PID> /F
```

## 🐳 Docker 支持（可选）

虽然我们移除�?Docker 相关的部署脚本以简化项目，但你仍然可以使用 Docker�?

```dockerfile
# 创建简单的 Dockerfile
FROM node:18-alpine

RUN corepack enable && corepack prepare pnpm@8.15.0 --activate

WORKDIR /app
COPY . .

RUN pnpm install
RUN pnpm run build

EXPOSE 4000 4001

CMD ["pnpm", "run", "deploy"]
```

## 📊 系统要求

| 要求 | 版本 |
|-----|------|
| Node.js | >= 18.0.0 |
| pnpm | >= 8.0.0 |
| 内存 | >= 2GB |
| 磁盘空间 | >= 500MB |

## 🔗 相关链接

- [主文档](../README.md)
- [快速开始](GETTING_STARTED.md)
- [使用说明](使用说明.md)

---

所有脚本都经过跨平台测试，确保在不同系统上都能正常工作！✅
