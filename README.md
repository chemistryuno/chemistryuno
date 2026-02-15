# 🧪 Chemistry UNO (化学版 UNO)

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://golang.org)
[![Vue Version](https://img.shields.io/badge/Vue-3.4-4FC08D?logo=vue.js)](https://vuejs.org)
[![Vite](https://img.shields.io/badge/Vite-Latest-646CFF?logo=vite)](https://vitejs.dev)

**Chemistry UNO** 是一款将经典 UNO 玩法与化学反应逻辑相结合的多人在线卡牌游戏。玩家不仅需要考虑出牌策略，还需要利用手中的化学元素进行物质合成并触发特殊化学反应效果。

---

## ✨ 核心特性

- 🎮 **创新玩法**:
  - **元素卡牌**: 基础牌由化学元素组成。
    - **物质合成**: 动态判断玩家当前手牌是否能组成特定化学物质并出牌。
    - **判定引擎**: 基于酸碱中和、复分解、金属活动性等化学原理的动态反应判定（非简单的硬编码列表）。
- ⚡ **实时竞技**:
  - 基于 **WebSocket** 的实时多人对战。
  - 30秒限时操作系统，超时自动托管，保证游戏流畅度。
- 🔐 **极致安全与身份恢复**:
  - **WebAuthn**: 支持 FIDO2/硬件安全密钥（如 Yubikey）进行零密码登录。
  - **多维度找回**: 业界领先的“双因子找回逻辑”——若不慎丢失密码，可通过物理密钥或 2FA 动态码安全重置。
  - **2FA**: 集成 Google Authenticator 等两步验证方案。
  - **RBAC**: 完善的权限系统（Admin, Co-worker, User）。
- 📊 **管理后台**:
  - **化学库管理**: 实时管理物质数据状态（草稿/已审核）。
  - **全量审计**: 前端管理页提供详尽的游戏历史溯源（Reactor Logs）。
  - **反馈系统**: 玩家反馈实时收集与处理。
- 🎨 **现代化体验**:
  - **Modern UI**: 使用 Tailwind CSS 4 构建的现代化响应式界面。
  - **公式编辑器**: 内置化学公式解析与渲染。

## 🛠️ 技术栈

### 后端 (Backend)

- **语言**: Go (1.24+)
- **框架**: Gin Web Framework
- **数据库**: SQLite（modernc 纯 Go 驱动，默认 WAL 模式）或 MySQL（可选）
- **核心逻辑**: 化学反应裁判与物质解析模块（backend/game）
- **安全安全**: WebAuthn (go-webauthn), TOTP (pquerna/otp), Argon2, JWT

### 前端 (Frontend)

- **框架**: Vue 3 (Composition API)
- **构建工具**: Vite / TypeScript
- **样式**: Tailwind CSS 4
- **状态管理**: 响应式 API
- **图标**: Lucide Icons

## 🧬 化学逻辑引擎

本项目包含一个独特的化学逻辑判断模块，模仿了真实的化学反应判定过程：

1. **动态解析**: 支持 `Fe(OH)3`、`Ca(HCO3)2` 等复杂化学式的原子统计与类型识别。
2. **反应模拟**:
    - **酸碱理论**: 自动识别酸、碱、盐及氧化物间的反应规律。
    - **溶解性判定**: 结合溶解性规则判断沉淀生成。
    - **金属活动性顺序**: 严格遵循 K-Au 序列。

## 🚀 快速开始

### 环境依赖

- **Node.js**: >= 18
- **Go**: >= 1.24
- **pnpm**: `npm install -g pnpm`

### 一键安装与启动（开发）

项目内置了便捷脚本（Windows/Linux/macOS 通用）：

```bash
# 1. 克隆并进入目录
git clone https://github.com/your-repo/chemistryuno.git
cd chemistryuno

# 2. 运行初始化脚本（安装依赖 + 环境检查）
pnpm run init

# 3. 启动全栈项目（后端 :8080 + 前端 :5000）
pnpm start
```

启动后：

- 🌐 前端地址: `http://localhost:5000`
- ⚙️ 后端 API: `http://localhost:8080`

### 常见问题

- 默认数据库为 SQLite（单文件，免安装）。如需 MySQL，设置 `DB_TYPE=mysql` 并配置 `MYSQL_DSN` 或 `MYSQL_*` 相关环境变量。
- Redis 为可选；未配置将自动降级但核心功能不受影响。
- 如需更多命令与排错，请参见根目录的 COMMANDS 与部署文档。

> 详细命令与排错见：`COMMANDS.md`、`DEPLOYMENT.md`、`QUICKSTART.md`、`backend/API_DOCUMENTATION.md`

## 📂 目录结构

```text
.
├── backend/            # Go 后端源码
│   ├── game/           # 核心逻辑（化学判定/裁判/定时任务）
│   ├── handlers/       # API 路由处理器（Auth/WebAuthn/Game/Admin）
│   ├── models/         # 数据库模型
│   └── websocket/      # 通信层（Hub/Client）
├── frontend/           # Vue 前端源码
│   ├── src/
│   │   ├── components/ # 基础组件
│   │   ├── pages/      # 业务页面（Lobby/GameRoom/Admin 等）
│   │   └── utils/      # API/WS 工具
├── start.js            # 开发环境一键启动脚本
├── build.js            # 生产构建脚本（前后端）
├── COMMANDS.md         # 命令速查表
├── QUICKSTART.md       # Linux/面板快速部署
└── DEPLOYMENT.md       # 完整部署指南
```

## 🛡️ 安全架构

本项目采用分级安全验证逻辑，确保高价值账户操作的安全：

- **验证优先级**: 系统会自动检测用户绑定的安全凭证，按 `FIDO2 硬件密钥 > 2FA 动态码 > 传统密码` 的顺序提示验证。
- **零密码认证**: 完成 WebAuthn 注册后，用户可完全脱离密码，使用生物识别或硬件按钮完成安全挑战。
- **数据安全**: 敏感信息采用 Argon2 算法进行不可逆哈希存储，API 通信基于 JWT 无状态令牌。

## 📊 管理能力

管理员可进行全方位管控：

- **实时监控**: 追踪每一个 Reactor 实例的运行状态。
- **物质审核**: 对玩家提交的新物质合成公式进行合规性审查。
- **数据导出**: 支持游戏历史与积分排行的导出与可视化。

## ⚙️ 基本配置

- `.env`（根目录）：首次运行可复制 `.env.example`，关键变量：
  - `DB_TYPE`：`sqlite`（默认）或 `mysql`
  - `SQLITE_PATH`：SQLite 数据文件路径，默认 `./chemistryuno.db`
  - `MYSQL_DSN` 或 `MYSQL_HOST/PORT/USER/PASSWORD/DATABASE`
  - `JWT_SECRET`：JWT 签名密钥（首次启动会自动生成或覆盖）
  - `REDIS_ADDR`：Redis 地址（可选）
  - `APP_VERSION`、`APP_VERSION_NAME`：版本信息（可选）

## 🏗️ 构建与发布

- 开发构建前端：`pnpm -C frontend build`
- 一体化构建（前后端）：`pnpm build`
- 产物：
  - 后端二进制：根目录 `chemistryuno(.exe)`
  - 前端静态文件：`backend/static/dist`（构建时自动嵌入）
  - 完整包：`dist/`（包含运行脚本 start.sh/.bat）

## 🤝 贡献与反馈

欢迎提交 Issue 或 Pull Request 来改进化学平衡判定系统或丰富 UI 设计！

---

**Chemistry UNO V1.2.1 "Mendeleef"** - 让化学学习变得更有趣。

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

### v1.2.1

- ✅ 默认使用 modernc 纯 Go SQLite（免 CGO）
- ✅ 统一开发端口：前端 5000 / 后端 8080
- ✅ 新增命令速查、部署与快速启动文档
- ✅ 增强管理后台与化学引擎稳定性

### v1.0.0

- ✅ 迁移到 TypeScript
- ✅ 切换包管理器为 pnpm
- ✅ 添加一键启动脚本
- ✅ 配置 pnpm workspace

---

**快乐编码！** 🚀
