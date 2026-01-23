# Chemistry UNO Alpha 🧪

化学元素版 UNO 游戏 - 基于 React + TypeScript + Go

## 技术栈

- **前端**: React 18 + TypeScript + Vite
- **后端**: Go + WebSocket
- **包管理**: pnpm

## 快速开始

### 前置要求

- Node.js >= 18
- Go >= 1.20
- pnpm (如未安装，运行: `npm install -g pnpm`)
- **MinGW-w64 GCC >= 8.0** (旧版会导致编译失败)

> ⚠️ **重要**: 如果遇到 `unrecognized option '--high-entropy-va'` 错误，说明您的GCC版本过旧。
> 请运行 `fix-and-start.ps1` 自动修复，或参考 [BACKEND_BUILD_FIX.md](BACKEND_BUILD_FIX.md)

### 安装依赖

```bash
# 使用 pnpm 安装前端依赖
cd frontend
pnpm install
cd ..
```

### 🚨 GCC版本问题修复

如果启动时遇到编译错误，请选择以下方式之一：

**方式1: 自动修复 (推荐)**
```powershell
.\fix-and-start.ps1
```

**方式2: 手动安装新版GCC**
1. 下载 [WinLibs MinGW](https://winlibs.com/) 或 [MinGW-w64](https://github.com/niXman/mingw-builds-binaries/releases)
2. 解压到 `tools/mingw64` 目录
3. 运行 `fix-and-start.bat`

**方式3: 兼容模式启动**
```batch
start-backend-compatible.bat
```

### 一键启动

在项目根目录下运行：

```bash
# 使用 Node.js 运行启动脚本
node start.js

# 或使用 pnpm
pnpm start
```

这将同时启动：
- 🎨 前端开发服务器: http://localhost:3000
- 📦 后端 API 服务器: http://localhost:8080

按 `Ctrl+C` 停止所有服务。

### 单独启动

如需单独启动服务：

```bash
# 仅启动前端
pnpm frontend

# 仅启动后端
pnpm backend
```

## 项目结构

```
chemistryuno-alpha/
├── frontend/               # 前端项目 (React + TypeScript)
│   ├── src/
│   │   ├── pages/         # 页面组件
│   │   ├── utils/         # 工具函数 (TypeScript)
│   │   ├── App.tsx        # 主应用组件
│   │   └── main.tsx       # 入口文件
│   ├── tsconfig.json      # TypeScript 配置
│   └── vite.config.ts     # Vite 配置
├── backend/               # 后端项目 (Go)
│   ├── handlers/          # 路由处理器
│   ├── models/            # 数据模型
│   ├── websocket/         # WebSocket 处理
│   └── main.go            # 入口文件
├── start.js               # 一键启动脚本
├── package.json           # 根项目配置
└── pnpm-workspace.yaml    # pnpm 工作区配置
```

## 开发说明

### TypeScript 类型检查

```bash
cd frontend
pnpm type-check
```

### 构建生产版本

```bash
cd frontend
pnpm build
```

## 更新日志

### v1.0.0
- ✅ 迁移到 TypeScript
- ✅ 切换包管理器为 pnpm
- ✅ 添加一键启动脚本
- ✅ 配置 pnpm workspace

---

**快乐编码！** 🚀
