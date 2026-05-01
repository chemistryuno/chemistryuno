# 🧪 Chemistry UNO（化学版 UNO）

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://golang.org)
[![Vue Version](https://img.shields.io/badge/Vue-3.4-4FC08D?logo=vue.js)](https://vuejs.org)
[![Vite](https://img.shields.io/badge/Vite-5.x-646CFF?logo=vite)](https://vitejs.dev)

**Chemistry UNO** 是一个将 UNO 回合制卡牌机制与化学反应判定结合的多人在线游戏系统。  
它包含完整的前后端、实时通信、账户安全体系、插件扩展系统、管理后台与内容审核流程。

---

## 🗂️ 目录

- [当前版本](#-当前版本)
- [运行模式速览](#-运行模式速览)
- [功能总览（当前版本）](#-功能总览当前版本)
- [化学规则与出牌逻辑（摘要）](#-化学规则与出牌逻辑摘要)
- [安全设计](#️-安全设计)
- [技术栈](#-技术栈)
- [快速开始（开发）](#-快速开始开发)
- [常用脚本](#-常用脚本)
- [环境变量（核心项）](#️-环境变量核心项)
- [API 与路由分组（摘要）](#-api-与路由分组摘要)
- [项目结构](#-项目结构)
- [文件职责与分层约定](#-文件职责与分层约定)
- [测试建议](#-测试建议)
- [文档索引](#-文档索引)
- [FAQ / 排障](#-faq--排障)
- [贡献](#-贡献)

---

## 📌 当前版本

- 版本：`1.2.1`
- 代号：`Mendeleef`
- 默认端口：前端 `5000` / 后端 `8080`
- 默认数据库：`SQLite`（可切换 `MySQL`）

---

## 🚦 运行模式速览

| 模式 | 命令 | 前端入口 | 后端入口 | 说明 |
| --- | --- | --- | --- | --- |
| 开发 | `pnpm start` | `http://localhost:5000` | `http://localhost:8080` | 前后端分离开发、热更新 |
| 测试 | `pnpm test` | `http://localhost:5000` | `http://localhost:8080` | 独立测试数据环境 |
| 生产构建 | `pnpm build` | `http://localhost:8080` | `http://localhost:8080` | 前端静态资源嵌入后端单体运行 |
| Electron 客户端（开发） | `pnpm electron:dev` | Electron 窗口 | `http://localhost:8080` | 使用 Electron 桌面壳加载 Vite 页面 |
| Windows 客户端（安装包） | `pnpm electron:pack:win` | Electron 安装包 | `http://localhost:8080` | 生成 Windows 安装包（NSIS） |
| Android 客户端（调试包） | `pnpm android:build:debug` | Android WebView | 通过 `CHEM_SERVER_ORIGIN` / `CHEM_ANDROID_API_ORIGIN` 指定 | 生成 Android Debug APK |

---

## ✨ 功能总览（当前版本）

### 1) 账号与身份系统

- 用户注册/登录（支持用户名模式与邮箱模式）
- 邮箱验证码（注册/重置/改密）
- 2FA（TOTP）启用、校验、关闭
- WebAuthn（硬件密钥）：
  - 无密码登录
  - WebAuthn 辅助密码重置
  - WebAuthn 辅助改密
  - 凭证管理（增删查）
- OAuth 第三方登录：
  - GitHub / Microsoft / Google / Apple
  - OAuth 账号绑定/解绑
- 会话管理（设备会话列表、单会话下线）
- 账号冻结（定时冻结）
- RBAC 角色权限（`admin` / `co-worker` / `user`）

### 2) 核心对战系统

- 房间系统：
  - 创建/加入/离开房间
  - 公开房间与私密房间（访问密钥）
  - 准备/开始流程
  - 实时房间状态与在线广播
- 模式支持：
  - PvP 多人对战
  - PvE 人机对战（AI 数量、AI 难度）
  - AI 自动补位（开局补空位，可配置难度）
  - 排位开关与等级范围匹配参数
- 回合机制：
  - 限时回合（超时处理）
  - 托管/自动接管逻辑
  - 重连与状态同步
  - 完赛玩家状态与观战态处理
- 出牌机制：
  - 普通出牌（化学物质）
  - 双联反应（双物质组合）
  - 摸牌与惩罚结算
  - 特殊牌：`+2`、`+4`、`reverse`、`Au`、稀有气体（跳过/转向规则）
- AI 系统：
  - 难度驱动的策略决策
  - 威胁检测、协作策略、卡位策略
  - 随机策略与最优策略混合
  - 教学脚本 AI（固定步骤）

### 3) 化学引擎与数据内容

- 化学式解析与元素需求计算（支持复杂化学式）
- 反应判定与提示查询
- 物质与反应数据分离管理
- 物质/反应提交、审核、拒绝、批量审批
- 重复/待完善内容自动标记
- 牌组系统：
  - 全局牌组（管理员）
  - 个人牌组（玩家）
  - 初始手牌数配置

### 4) 教学与新手引导

- 大厅新手引导（分步骤聚焦 UI）
- 引导完成后可自动进入教学关卡
- 教学关卡脚本模式（固定步骤、固定手牌、步骤提示）
- 教学提示（回合内动态提示）
- 跳过教程后可记录状态，避免反复弹出

### 5) 社交与社区功能

- 全局聊天（大厅）
- 私聊（好友间）
- 好友系统：
  - 发送请求
  - 同意/拒绝
  - 好友备注
  - 删除好友
- 对战邀请（私聊内游戏邀请信息）
- 用户反馈系统（提交、催办、撤回）
- 公告系统：
  - 跑马灯公告
  - 持久公告
  - 入场公告
  - 定时公告

### 6) 积分、等级与竞技

- 总积分排行榜、月积分排行榜
- 悬赏系统（Bounty）
- 等级与经验系统（XP、Level）
- 实时积分结算（含 PvE 难度修正逻辑）

### 7) 管理后台能力（Admin）

- 用户管理（查、建、删、改密、改角色）
- 封禁与踢出
- 全局牌组配置管理
- 游戏历史与反馈处理
- 系统配置、游戏时间配置
- 公告全生命周期管理
- Excel 导出（物质/反应/全量）
- 批量审批（物质/反应）
- 管理广播（全局 / 房间 / 指定用户）
- 活跃房间查询
- 插件管理（见下）
- 服务器计划重启/取消重启

### 8) 插件系统（Plugin）

- 插件卡牌注册与运行时加载
- 插件脚本读取（前端只读查看）
- 管理端插件功能：
  - 创建/更新/删除插件
  - 上传安装 `.cumod`
  - 插件卡牌增删改查
  - 热重载插件
- 插件事件机制（房间创建、回合切换、出牌等事件）

### 9) 实时通信与可运维性

- WebSocket Hub（房间维度广播、用户定向推送）
- 健康检查（`/api/health`）与 `ping`
- 优雅停机（SIGINT/SIGTERM）
- 前端静态资源 embed（后端可单体运行）
- 后台定时清理任务（如过期反馈）

---

## 🧠 化学规则与出牌逻辑（摘要）

- 手牌中的元素可组合成可用物质后出牌
- 若场上有物质限制，下一手需与场上物质满足反应条件（除特殊牌）
- `+2/+4` 支持叠加，未防御时需一次性结算摸牌
- `Au` 可重置场面并跳过目标
- 稀有气体触发特殊稳定性逻辑（转向/跳过效果）
- 支持双联反应（`play-double`）

> 更完整数据模型与接口可参考 `backend/API_DOCUMENTATION.md`。

---

## 🛡️ 安全设计

- 密码哈希：Argon2
- 鉴权：JWT + SID 会话双层校验
- 鉴权中间件：统一验证 UID / SID / 封禁状态 / 冻结状态
- 2FA + WebAuthn + OAuth 组合式登录恢复能力
- 管理接口全部走 `AuthMiddleware + AdminMiddleware`

---

## 🧰 技术栈

### 后端

- Go 1.24+
- Gin
- GORM
- SQLite（modernc）/ MySQL
- WebSocket（gorilla/websocket）
- WebAuthn（go-webauthn）
- TOTP（pquerna/otp）

### 前端

- Vue 3 + Composition API
- TypeScript
- Vite
- Tailwind CSS 4
- Axios
- Vue Router

---

## 🚀 快速开始（开发）

### 依赖要求

- Node.js >= 18
- pnpm
- Go >= 1.24

### 初始化并启动

```bash
git clone <your-repo-url>
cd chemistryuno
pnpm run init
pnpm start
```

启动后：

- 前端：`http://localhost:5000`
- 后端：`http://localhost:8080`

---

## 📜 常用脚本

### 根目录

- `pnpm run init`：初始化依赖与项目环境
- `pnpm start`：启动前后端开发环境
- `pnpm build`：一体化构建
- `pnpm build:frontend`：仅构建前端
- `pnpm build:backend`：仅构建后端
- `pnpm electron:dev`：启动 Electron 客户端（开发模式）
- `pnpm electron:run`：构建前端后启动 Electron 客户端（生产渲染资源）
- `pnpm electron:pack:win`：构建并打包 Windows 客户端安装包（输出到 `frontend/release`）
- `pnpm android:add`：初始化 Android 工程（首次执行）
- `pnpm android:sync`：用指定 API 地址构建前端并同步到 Android 工程
- `pnpm android:build:debug`：生成 Android 调试 APK
- `pnpm android:build:release`：生成 Android Release APK（未提供签名变量时输出 unsigned 包）
- `pnpm go:test`：执行 Go 测试
- `pnpm test`：项目测试脚本入口

### 前端目录

- `pnpm -C frontend dev`
- `pnpm -C frontend build`
- `pnpm -C frontend type-check`

---

## ⚙️ 环境变量（核心项）

复制 `.env.example` 为 `.env` 后按需修改。

### 基础配置

- `APP_VERSION` / `APP_VERSION_NAME`
- `DB_TYPE=sqlite|mysql`
- `SQLITE_PATH`（SQLite）
- `MYSQL_DSN`（MySQL）
- `JWT_SECRET`
- `REDIS_ADDR`（可选）
- `VITE_SERVER_ORIGIN`：前端运行时和 Vite 开发代理使用的服务器地址，例如 `http://127.0.0.1:8080`
- `CHEM_SERVER_ORIGIN`：Electron / Android 打包时共享使用的服务器地址；未单独指定平台变量时会回退到它
- `CHEM_ANDROID_API_ORIGIN`：Android 专用服务器地址，优先级高于 `CHEM_SERVER_ORIGIN`

### 安全配置

- `WEBAUTHN_RPID`
- `WEBAUTHN_RP_ORIGIN`
- `WEBAUTHN_ORIGIN`

### SMTP（启用邮箱模式）

- `SMTP_HOST` / `SMTP_PORT` / `SMTP_USER` / `SMTP_PASS` / `SMTP_FROM`

### OAuth（第三方登录）

- GitHub：`GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET` / `GITHUB_REDIRECT_URI`
- Microsoft：`MS_CLIENT_ID` / `MS_CLIENT_SECRET` / `MS_TENANT_ID` / `MS_REDIRECT_URI`
- Google：`GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` / `GOOGLE_REDIRECT_URI`
- Apple：`APPLE_CLIENT_ID` / `APPLE_CLIENT_SECRET` / `APPLE_REDIRECT_URI`

> Apple 当前实现使用静态 `APPLE_CLIENT_SECRET`。  
> 若仅配置 `CLIENT_ID` 而未配置必要字段，前端会自动隐藏该 OAuth 入口。

---

## 🔌 API 与路由分组（摘要）

- 公开：`/api/auth/*`、`/api/version`、`/api/health`、`/api/announcements`、`/api/hints`
- 鉴权：`/api/user/*`、`/api/rooms/*`、`/api/friends/*`、`/api/chat/*`、`/api/points/*`
- 管理：`/api/admin/*`
- WebSocket：`/api/ws`

---

## 📁 项目结构

```text
.
├── backend/
│   ├── anticheat/            # 反作弊系统（检测、处罚、申诉、审计）
│   ├── cache/                # 缓存层（Redis）
│   ├── config/               # 配置文件（包含反作弊配置）
│   ├── database/             # 数据库模型与迁移
│   ├── deckcfg/              # 卡组配置管理
│   ├── game/                 # 游戏核心逻辑（规则、AI、回合、房间）
│   ├── handlers/             # HTTP 处理器（auth/game/admin/plugin...）
│   ├── middleware/           # 中间件（认证、CORS、速率限制）
│   ├── models/               # 数据结构定义
│   ├── plugins/              # 插件系统（脚本运行时）
│   ├── repository/           # 数据访问层（DAO）
│   ├── router/               # 路由注册
│   ├── scripts/              # 工具脚本（初始化、迁移等）
│   ├── static/               # 静态资源、法律文档
│   ├── utils/                # 工具函数库
│   └── websocket/            # WebSocket 连接管理
├── docs/                     # 📚 集中式文档库
│   ├── anticheat/            # 反作弊系统文档
│   ├── api/                  # API 文档
│   ├── guides/               # 用户与开发指南
│   ├── architecture/         # 架构设计文档
│   └── legal/                # 法律文档（隐私政策等）
├── frontend/                 # Vue 3 前端应用
│   ├── src/
│   │   ├── components/       # Vue 组件
│   │   ├── views/            # 页面组件
│   │   ├── stores/           # Pinia 状态管理
│   │   ├── api/              # API 请求层
│   │   └── utils/            # 工具函数
│   ├── electron/             # Electron 桌面壳
│   └── public/               # 公开资源
├── scripts/                  # 项目级工具脚本
├── tools/                    # Go 工具集（数据库初始化等）
├── main.go                   # 后端入口
├── package.json              # 项目元数据与脚本
└── README.md                 # 本文档
```
│   ├── router/               # 路由装配层（分组、鉴权绑定、路径注册）
│   ├── middleware/           # 鉴权、权限、CORS
│   ├── repository/           # 数据访问层
│   ├── websocket/            # WS Hub & Client
│   ├── scripts/              # 后端脚本/脚本测试
│   └── static/               # 前端构建产物嵌入目录
├── frontend/
│   └── src/
│       ├── pages/            # 页面（Lobby/GameRoom/Admin/Profile/Login...）
│       ├── components/       # 组件
│       ├── composables/      # 复用逻辑
│       └── utils/            # API/WS/工具函数
├── tools/                    # 工具模块（入口在 tools/cmd/*）
├── COMMANDS.md
├── QUICKSTART.md
├── DEPLOYMENT.md
└── backend/API_DOCUMENTATION.md
```

---

## 🧭 文件职责与分层约定

- 工程入口脚本（`init.js` / `start.js` / `build.js` / `test.js`）仅负责流程编排，不承载业务规则。
- 后端业务逻辑放在 `backend/` 分层内：`router` 处理路由装配，`handlers` 处理协议边界，`repository` 处理数据访问，`game` 处理领域规则。
- 数据修复与迁移脚本统一放在 `backend/scripts/`，避免与通用工具目录重复。
- 仓库级自动化脚本放在 `scripts/`。
- 详细职责说明见：`docs/FILE_RESPONSIBILITIES.md`。

---

## 🧪 测试建议

- 后端：`go test ./backend/...`
- 前端：`pnpm -C frontend build` + `pnpm -C frontend type-check`
- OAuth 自动化脚本测试：`go test -tags scripts backend/scripts/oauth_third_party_test.go -v`
- 端到端手测建议：登录/建房/出牌/重连/结算/退出

---

## 📚 文档索引

- 部署：`DEPLOYMENT.md`
- 快速上手：`QUICKSTART.md`
- 命令速查：`COMMANDS.md`
- API 文档：`backend/API_DOCUMENTATION.md`
- 等级系统：`LEVEL_SYSTEM_DOCS.md`
- 文件职责：`docs/FILE_RESPONSIBILITIES.md`

---

## ❓ FAQ / 排障

### 1) 前端能开，后端接口全 404？

- 确认后端是否已启动在 `:8080`。
- 本地开发应通过 `pnpm start` 一起启动；仅开前端会导致 API 无法访问。

### 2) OAuth 登录按钮不显示？

- 按钮受后端配置开关控制。
- 例如 GitHub 至少需要：
  - `GITHUB_CLIENT_ID`
  - `GITHUB_CLIENT_SECRET`
- Apple 目前还需要：
  - `APPLE_CLIENT_ID`
  - `APPLE_CLIENT_SECRET`
  - `APPLE_REDIRECT_URI`

### 3) OAuth 授权后弹窗关闭但没登录？

- 新版本已处理弹窗关闭与消息回传竞态。
- 若仍复现，优先检查浏览器是否拦截了弹窗/跨窗口消息。
- 可先跑脚本测试定位：`go test -tags scripts backend/scripts/oauth_third_party_test.go -v`

### 4) 首次启动提示找不到 `.env`？

- 复制模板：`cp .env.example .env`（Windows 用 `copy`）。
- 至少确保 `JWT_SECRET` 可用（系统也支持首次自动生成）。

### 5) 生产环境推荐怎么跑？

- 使用 `pnpm build` 生成产物后，前端静态文件会输出到 `dist/frontend/`，后端可执行文件会输出到 `dist/`，直接运行 `dist` 内启动脚本即可。
- 生产建议前置 Nginx/Caddy，并启用 HTTPS（尤其 WebAuthn/OAuth）。

---

## 🤝 贡献

欢迎提交 Issue / PR。  
建议提交前完成以下检查：

1. `go test ./backend/...`
2. `pnpm -C frontend build`
3. 关键页面功能手测（登录、建房、对局、退出）

---

## 📄 许可证

MIT License

---

## 🏗️ 文件职责与分层约定

详见 [docs/architecture/FILE_RESPONSIBILITIES.md](docs/architecture/FILE_RESPONSIBILITIES.md)

**核心原则**：

- `handlers`：HTTP 请求解析、应答塑形、认证检验，禁止直接操作 SQL
- `repository`：数据访问与查询组合，禁止 HTTP 相关逻辑
- `game`：游戏引擎、规则、AI 决策、房间生命周期
- `middleware`：跨层请求过滤（认证、速率限制、CORS）
- `websocket`：Socket 生命周期与消息发布
- `scripts`：后端数据库/数据迁移脚本
- `tools/cmd/*`：独立 Go 工具模块与一次性维护工具

---

## 🧪 测试建议

### 后端测试

```bash
# 运行所有 Go 测试
pnpm run go:test

# 特定包测试
go test ./backend/game -v
go test ./backend/anticheat -v

# 覆盖率统计
go test -cover ./...
```

### 前端测试

```bash
# 运行前端单元测试（如配置）
pnpm -C frontend test

# 类型检查
pnpm -C frontend type-check
```

### E2E 测试

```bash
# 启动测试环境
pnpm test

# 后续可集成 Playwright / Cypress
```

### 反作弊系统测试

关键测试覆盖：
- 风险评分计算精度
- 多维度检测准确性
- 处罚等级映射
- 申诉工作流
- 审计日志完整性

详见 `backend/anticheat/anticheat_test.go` 和集成测试。

---

## 📚 文档索引

### 核心文档

| 文档 | 位置 | 用途 |
|-----|-----|-----|
| **反作弊系统指南** | [docs/anticheat/ANTICHEAT_GUIDE.md](docs/anticheat/ANTICHEAT_GUIDE.md) | 完整的反作弊系统参考（架构、检测、配置、运维） |
| **文件职责** | [docs/architecture/FILE_RESPONSIBILITIES.md](docs/architecture/FILE_RESPONSIBILITIES.md) | 代码分层约定与文件所有权规则 |
| **隐私政策** | [docs/legal/PRIVACY_POLICY.md](docs/legal/PRIVACY_POLICY.md) | 用户隐私相关 |
| **用户协议** | [docs/legal/USER_AGREEMENT.md](docs/legal/USER_AGREEMENT.md) | 用户服务条款 |

### 在线资源与参考

- **API 文档**：启动后端后访问 `http://localhost:8080/swagger`（如配置 Swagger）
- **数据库模型**：见 `backend/database/models.go`
- **游戏规则**：见 `backend/game/judge.go` 和 `backend/game/chemistry.go`
- **插件开发**：见 `backend/plugins/runtime.go` 和管理员文档

---

## ❓ FAQ / 排障

### Q1: 启动时提示"数据库连接失败"

**原因**：未初始化数据库或配置错误

**解决**：
```bash
pnpm run init
# 或手动初始化
pnpm run db:init
```

### Q2: 前端无法连接后端（CORS 错误）

**原因**：VITE_SERVER_ORIGIN 配置错误或后端未启动

**解决**：
```bash
# 确认后端在运行
curl http://localhost:8080/api/health

# 检查 .env 中的 VITE_SERVER_ORIGIN 设置
# 应为 http://127.0.0.1:8080（开发环境）
```

### Q3: Electron 打包后白屏

**原因**：CHEM_SERVER_ORIGIN 配置指向本地开发服务器

**解决**：
```bash
# 设置正确的生产服务器地址
export CHEM_SERVER_ORIGIN=https://your-production-server.com
pnpm electron:pack:win
```

### Q4: 反作弊系统检测量偏低

**原因**：配置阈值过高或检测维度权重不合理

**解决**：
1. 查看 [docs/anticheat/ANTICHEAT_GUIDE.md](docs/anticheat/ANTICHEAT_GUIDE.md) 的配置示例
2. 通过管理后台逐步调整权重和阈值
3. 持续监控申诉率，确保误判率可接受

### Q5: 如何在 Android 上运行游戏

**解决**：
```bash
pnpm android:add              # 首次初始化 Android 工程
pnpm android:sync             # 构建前端并同步到 Android
pnpm android:build:debug      # 生成调试 APK
# 然后用 ADB 安装：adb install release/app-debug.apk
```

---

## 🤝 贡献

欢迎贡献代码、报告问题、改进文档！

### 贡献流程

1. **创建 Issue** - 讨论计划的功能或问题修复
2. **Fork & Branch** - 从 `main` 创建特性分支（如 `feature/xxx` 或 `fix/xxx`）
3. **开发与测试** - 确保代码通过测试并遵循分层约定
4. **创建 Pull Request** - 详细描述改动内容
5. **代码审查** - 参与讨论并根据反馈调整
6. **合并** - 由项目维护者合并至主分支

### 编码规范

- **Go**：遵循 [Effective Go](https://golang.org/doc/effective_go) 和 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- **Vue/TypeScript**：
  - 使用 Composition API
  - 启用 TypeScript 严格模式
  - 遵循 ESLint 规则
  - 每行不超过 120 字符

### 分支命名

- `feature/xxx` - 新功能
- `fix/xxx` - Bug 修复
- `refactor/xxx` - 代码重构
- `docs/xxx` - 文档改进
- `chore/xxx` - 构建、依赖更新等

### 提交消息规范

```
<type>(<scope>): <subject>

<body>

<footer>
```

示例：
```
feat(anticheat): add pattern detection dimension

Implements a new detection strategy based on operation pattern analysis.
Adds periodic pattern checks during game end processing.

Closes #123
```

### 测试要求

- 新功能必须有对应的单元测试
- 修复 Bug 时应添加回归测试
- 确保所有测试通过：`pnpm test` 和 `pnpm go:test`

---

## 📞 支持与反馈

- **Issue 追踪**：使用 GitHub Issues 报告 Bug 和功能请求
- **讨论**：在 Discussions 中参与社区讨论
- **安全问题**：请勿在 Issue 中公开，改为私密报告给维护者

---

## 📄 许可证

本项目采用 MIT License 开源。详见 [LICENSE](LICENSE) 文件。

---

**Chemistry UNO V1.2.1 "Mendeleef"**  
让化学学习与卡牌策略真正结合。
