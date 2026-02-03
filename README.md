# 🧪 Chemistry UNO (化学版 UNO)

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?logo=go)](https://golang.org)
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
  - **全量审计**: [Admin.vue](frontend/src/pages/Admin.vue) 提供详尽的游戏历史溯源（Reactor Logs）。
  - **反馈系统**: 玩家反馈实时收集与处理。
- 🎨 **现代化体验**:
  - **Modern UI**: 使用 Tailwind CSS 4 构建的现代化响应式界面。
  - **公式编辑器**: [EquationEditor.vue](frontend/src/components/EquationEditor.vue) 内置化学公式解析与渲染。

## 🛠️ 技术栈

### 后端 (Backend)

- **语言**: Go (1.20+)
- **框架**: Gin Web Framework
- **数据库**: SQLite (通过 CGO 连接)
- **核心逻辑**: [judge.go](backend/game/judge.go) (化学反应逻辑), [chemistry.go](backend/game/chemistry.go) (物质解析)
- **安全安全**: WebAuthn (go-webauthn), TOTP (pquerna/otp), Argon2, JWT

### 前端 (Frontend)

- **框架**: Vue 3 (Composition API)
- **构建工具**: Vite / TypeScript
- **样式**: Tailwind CSS 4
- **状态管理**: 响应式 API
- **图标**: Lucide Icons

## 🧬 化学逻辑引擎

本项目包含一个独特的化学逻辑判断模块 [judge.go](backend/game/judge.go)，它模仿了真实的化学反应判定过程：

1. **动态解析**: 支持 `Fe(OH)3`、`Ca(HCO3)2` 等复杂化学式的原子统计与类型识别。
2. **反应模拟**:
    - **酸碱理论**: 自动识别酸、碱、盐及氧化物间的反应规律。
    - **溶解性判定**: 结合溶解性规则判断沉淀生成。
    - **金属活动性顺序**: 严格遵循 K-Au 序列。

## 🚀 快速开始

### 环境依赖

- **Node.js**: >= 18
- **Go**: >= 1.20
- **GCC**: MinGW-w64 (建议版本 >= 8.0，用于编译 CGO 代码)
- **pnpm**: `npm install -g pnpm`

### 一键安装与启动

项目内置了便捷的初始化脚本，支持 Windows 环境：

```bash
# 1. 克隆并进入目录
git clone https://github.com/your-repo/chemistryuno.git
cd chemistryuno

# 2. 运行初始化脚本 (安装依赖 + 环境检查)
pnpm run init

# 3. 启动全栈项目
pnpm start
```

启动后：

- 🌐 前端地址: `http://localhost:5173`
- ⚙️ 后端 API: `http://localhost:8080`

### 🔧 常见问题解决 (GCC)

如果在启动后端时遇到 `unrecognized option '--high-entropy-va'` 错误：

- 请运行根目录下的 [init.bat](init.bat) 进行修复。
- 或执行 [start.js](start.js) 脚本，它会自动处理环境变量配置。

## 📂 目录结构

```text
.
├── backend/            # Go 后端源码
│   ├── game/           # 核心逻辑 (化学判定 [judge.go](backend/game/judge.go))
│   ├── handlers/       # API 路由处理器 (WebAuthn, Game, Admin)
│   ├── models/         # 数据库模型
│   └── websocket/      # 通信层 [hub.go](backend/websocket/hub.go)
├── frontend/           # Vue 前端源码
│   ├── src/
│   │   ├── components/ # 基础组件
│   │   ├── pages/      # 业务页面 ([Lobby.vue](frontend/src/pages/Lobby.vue))
│   │   └── utils/      # [api.ts](frontend/src/utils/api.ts) & [websocket.ts](frontend/src/utils/websocket.ts)
├── init.bat            # 环境修复脚本
└── start.js            # Node.js 引导启动程序
```

## 🛡️ 安全架构

本项目采用分级安全验证逻辑，确保高价值账户操作的安全：

- **验证优先级**: 系统会自动检测用户绑定的安全凭证，按 `FIDO2 硬件密钥 > 2FA 动态码 > 传统密码` 的顺序提示验证。
- **零密码认证**: 完成 WebAuthn 注册后，用户可完全脱离密码，使用生物识别或硬件按钮完成安全挑战。
- **数据安全**: 敏感信息采用 Argon2 算法进行不可逆哈希存储，API 通信基于 JWT 无状态令牌。

## 📊 管理能力

管理员通过 [Admin.vue](frontend/src/pages/Admin.vue) 可进行全方位管控：

- **实时监控**: 追踪每一个 Reactor 实例的运行状态。
- **物质审核**: 对玩家提交的新物质合成公式进行合规性审查。
- **数据导出**: 支持游戏历史与积分排行的导出与可视化。

## 🤝 贡献与反馈

欢迎提交 Issue 或 Pull Request 来改进化学平衡判定系统或丰富 UI 设计！

---

**Chemistry UNO V1.0.0 "Mendeleef"** - 让化学学习变得更有趣。

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
