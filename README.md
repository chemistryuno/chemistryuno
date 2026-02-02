# Chemistry UNO V1.0.0 Mendeleef 🧪

一个基于化学元素的UNO卡牌游戏，支持多玩家在线对战，包含完整的用户权限系统和化学反应数据库管理。

## ✨ 主要功能

- 🎮 **多玩家在线游戏**: 支持2-4人实时对战
- ⏱️ **限时操作系统**: 30秒操作限制，超时自动摸牌
- 👥 **三级权限系统**: Admin/Co-worker/User角色管理
- 🧪 **化学反应库**: 可自定义管理化学反应数据
- 🎨 **现代化界面**: Vue.js + Tailwind CSS
- ⚡ **实时通信**: WebSocket支持
- 📊 **数据统计**: 游戏历史记录和用户管理

## 技术栈

- **前端**: Vue 3.4 (Composition API) + Vite + Tailwind CSS 4
- **后端**: Go (Gin) + SQLite + WebSocket
- **状态管理**: Vue Router
- **图标**: Lucide Vue Next
- **包管理**: pnpm

## 核心机制

本项目包含一个实时化学反应判定引擎，不依赖硬编码的反应列表，而是基于化学原理动态判断：

- **酸碱理论**: 自动识别中和反应及酸性/碱性氧化物的反应。
- **复分解反应**: 结合溶解性表（K/Na/NH4溶，Cl-除Ag+等）判断沉淀生成。
- **金属活动性**: 根据 K > Na > Ca > ... 判断置换反应。
- **动态解析**: 支持 `Ca(OH)2`、`Fe2(SO4)3` 等复杂化学式的动态原子统计。

## 🚀 快速开始

### 方法一：一键初始化（推荐）

```bash
# 克隆项目
git clone <repository-url>
cd chemistryuno-alpha

# 运行初始化脚本 - 自动安装依赖和配置
npm run init

# 启动项目
npm start
```

### 方法二：手动安装

#### 前置要求

- Node.js >= 18
- Go >= 1.20
- pnpm (如未安装，运行: `npm install -g pnpm`)
- **MinGW-w64 GCC >= 8.0** (旧版会导致 CGO 编译失败)

> ⚠️ **重要**: 如果遇到 `unrecognized option '--high-entropy-va'` 错误，请运行 `fix-and-start.ps1` 自动修复。

#### 🚨 GCC版本问题修复

如果启动时遇到编译错误：

##### 方式1: 自动修复 (推荐)

```powershell
.\fix-and-start.ps1
```

**组件化启动**：

```bash
# 一键启动前端 + 后端
node start.js
```

🎨 前端: <http://localhost:3000>
📦 后端: <http://localhost:8080>

## 项目结构

```txt

chemistryuno-alpha/
├── frontend/               # 前端项目 (Vue 3 + Vite)
│   ├── src/
│   │   ├── pages/         # 页面 (Lobby, GameRoom, Profile, Admin)
│   │   ├── utils/         # WebSocket & API 工具
│   │   ├── App.vue        # 主应用组件
│   │   └── main.ts        # 入口文件
├── backend/               # 后端项目 (Go)
│   ├── game/              # 核心逻辑 (judge.go, chemistry.go)
│   ├── handlers/          # API 路由处理
│   ├── database/          # SQLite 数据库操作
│   ├── models/            # 数据模型定义
│   └── main.go            # 后端入口
├── start.js               # 一键启动脚本
└── QUICK_FIX.md           # GCC 编译问题专门说明
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
