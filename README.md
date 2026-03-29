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
│   ├── game/                 # 游戏核心逻辑（规则、AI、回合、房间）
│   ├── handlers/             # HTTP 处理器（auth/game/admin/plugin...）
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

- 使用 `pnpm build` 生成产物后运行 `dist` 内启动脚本。
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

**Chemistry UNO V1.2.1 "Mendeleef"**  
让化学学习与卡牌策略真正结合。
