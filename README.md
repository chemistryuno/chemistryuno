# ⚗️ 化学UNO - Chemistry UNO Game

一个基于化学知识的创意卡牌游戏网络应用。玩家需要打出能与上一个物质发生化学反应的物质来继续游戏。

## � 快速开始

### 方法一：一键启动（推荐）

```bash
# Windows
node start.js

# 或使用 pnpm
pnpm run dev
```

### 方法二：手动安装

```bash
# 1. 安装依赖
pnpm install

# 2. 启动开发服务器
pnpm run dev
```

访问：
- **游戏**: http://localhost:3000
- **管理面板**: http://localhost:3000/admin

## 📦 生产部署

**Windows:**
```bash
# 双击运行
deploy.bat

# 或在命令行
pnpm run deploy
```

**Linux/Mac:**
```bash
# 添加执行权限（首次运行）
chmod +x deploy.sh

# 运行部署脚本
./deploy.sh

# 或使用 pnpm
pnpm run deploy
```

访问：
- **游戏**: http://localhost:4000
- **后端API**: http://localhost:4001
- **管理面板**: http://localhost:4000/admin

## 📚 详细文档

📋 **[查看完整文档目录](docs/文档目录.md)** - 所有文档的详细说明和快速查找

### 常用文档

- [快速入门指南](docs/GETTING_STARTED.md) - 开始游戏的详细说明
- [使用说明](docs/使用说明.md) - 简明使用指南
- [管理面板指南](docs/ADMIN_PANEL_GUIDE.md) - 如何管理游戏规则
- [移动端访问指南](docs/MOBILE_ACCESS_GUIDE.md) - 在手机上玩游戏
- [跨平台支持说明](docs/PLATFORM_SUPPORT.md) - Windows/Linux/Mac 使用指南
- **[Ubuntu 22 修复指南](docs/UBUNTU_FIX.md)** - 🔧 解决 Linux 部署问题

### 技术文档

- [跨平台支持说明](docs/PLATFORM_SUPPORT.md) - Windows/Linux/Mac 使用指南
- [安全配置指南](docs/SECURITY.md) - 密码和敏感信息管理 🔒
- [项目优化总结](docs/CLEANUP_SUMMARY.md) - 代码清理和优化记录
- [部署测试报告](docs/DEPLOY_TEST_REPORT.md) - 部署测试详情

## ⚠️ Ubuntu/Linux 用户注意

如果在 Ubuntu 22 或其他 Linux 发行版上遇到以下问题：
- ❌ 无法跳转到 /setup 页面
- ❌ 无法连接后端服务器
- ❌ 创建房间提示 network error

请运行快速修复脚本：
```bash
# 1. 应用修复
chmod +x apply-fix.sh
./apply-fix.sh

# 2. 验证修复
chmod +x verify-fix.sh
./verify-fix.sh

# 3. 重新构建和部署
pnpm run build
pnpm run deploy
```

详细说明请查看：**[Ubuntu 22 修复指南](docs/UBUNTU_FIX.md)**

## 🎮 游戏规则

### 牌组成分（可在管理面板自定义）
- **H、O**：各12张（常见元素）
- **C、N、F、Na、Mg、Al、Si、P、S、Cl、K、Ca、Mn、Fe、Cu、Zn、Br、I、Ag**：各4张
- **+2**：8张（摸2张牌）
- **+4**：4张（摸4张牌）
- **He、Ne、Ar、Kr**：各1张（反转游戏方向）
- **Au**：4张（跳过下一位玩家）

### 游戏流程
1. **初始化**：每位玩家获得10张牌
2. **出牌规则**：
   - 上家打出一种物质
   - 下家必须打出能与该物质**发生化学反应**的物质
   - 系统自动展示当前元素能组成的所有物质
   - 玩家从物质列表中选择合适的物质打出
3. **无法打出**：摸2张牌，回合结束
4. **特殊卡牌**：
   - He/Ne/Ar/Kr：反转游戏方向
   - Au：跳过下一位玩家
   - +2/+4：指定下一位玩家额外摸牌
5. **胜利条件**：首先清空手牌的玩家获胜

## 🏗️ 项目结构

```
chemistryuno/
├── client/              # React 前端
│   ├── src/
│   │   ├── components/  # 游戏组件
│   │   └── config/      # 配置文件
│   └── build/          # 构建输出
├── server/              # Node.js 后端
│   ├── index.ts        # 服务器入口
│   ├── gameLogic.ts    # 游戏逻辑
│   ├── database.ts     # 化学数据库
│   └── rules.ts        # 游戏规则
├── docs/               # 文档目录
│   ├── GETTING_STARTED.md      # 快速入门
│   ├── 使用说明.md              # 使用说明
│   ├── ADMIN_PANEL_GUIDE.md    # 管理面板
│   ├── MOBILE_ACCESS_GUIDE.md  # 移动端访问
│   ├── PLATFORM_SUPPORT.md     # 跨平台支持
│   ├── CLEANUP_SUMMARY.md      # 优化总结
│   └── DEPLOY_TEST_REPORT.md   # 部署测试
├── config.json         # 游戏配置（卡牌、反应数据库）
├── start.js           # 一键启动脚本
├── start.bat / start.sh # 平台启动脚本
└── deploy-pnpm.js     # 生产部署脚本
```

## 🛠️ 可用命令

```bash
# 开发环境
pnpm run dev          # 启动开发服务器
node start.js         # 一键启动（含依赖检查）

# 构建
pnpm run build        # 构建前端和后端
pnpm run build:client # 仅构建前端
pnpm run build:server # 仅构建后端

# 生产部署
pnpm run deploy       # 构建并启动生产服务器
```

## 🔧 配置说明

### 管理员密码

首次部署后，请立即修改管理员密码：

1. 编辑 `client/.env.production` 文件
2. 设置强密码：
```env
REACT_APP_ADMIN=your_strong_password_here
```
3. 重新构建：`pnpm run build:client`

### 游戏配置
编辑 `config.json` 可以修改：
- 卡牌数量和类型
- 化学反应数据库
- 游戏超时设置
- 特殊卡牌效果

## 📱 移动端访问

1. 确保电脑和手机在同一WiFi网络
2. 启动游戏后，查看控制台显示的二维码
3. 用手机扫描二维码即可访问

详见[移动端访问指南](docs/MOBILE_ACCESS_GUIDE.md)

## 🎯 技术栈

- **前端**: React 18, TypeScript, Socket.IO Client
- **后端**: Node.js, Express, Socket.IO, TypeScript
- **包管理**: pnpm
- **构建工具**: React Scripts, TypeScript Compiler

## 📋 系统要求

- Node.js >= 18.0.0
- pnpm >= 8.0.0

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License
├── package.json                 # 项目主配置（pnpm workspace）
├── pnpm-workspace.yaml          # pnpm工作区配置
├── tsconfig.json                # TypeScript根配置
├── config.json                  # 游戏配置文件（元素、物质、反应）
├── Dockerfile                   # Docker开发环境配置
├── Dockerfile.production        # Docker生产环境配置
├── docker-compose.yml           # Docker Compose开发配置
├── docker-compose.production.yml # Docker Compose生产配置
├── healthcheck.ts               # 健康检查脚本
├── server/                      # 后端（Node.js + TypeScript + Express + Socket.IO）
│   ├── package.json
│   ├── tsconfig.json            # 后端TypeScript配置
│   ├── index.ts                 # 主服务器文件（WebSocket + REST API）
│   ├── gameLogic.ts             # 游戏逻辑和化学反应匹配
│   ├── database.ts              # 化学数据库类
│   ├── rules.ts                 # 游戏规则引擎
│   ├── configService.ts         # 配置管理服务
│   ├── *.js                     # 编译前的旧JS文件（待清理）
│   └── dist/                    # TypeScript编译输出目录
├── client/                      # 前端（React 18 + TypeScript）
│   ├── package.json
│   ├── tsconfig.json            # 前端TypeScript配置
│   ├── public/
│   │   └── index.html
│   └── src/
│       ├── index.tsx            # 入口文件
│       ├── index.css
│       ├── App.tsx              # 主应用组件
│       ├── App.css
│       ├── config/
│       │   └── api.ts           # API配置（支持移动端）
│       ├── utils/
│       │   └── chemistryFormatter.ts  # 化学式格式化
│       └── components/
│           ├── GameLobby.tsx    # 游戏大厅（创建/加入游戏）
│           ├── GameLobby.css
│           ├── GameBoard.tsx    # 游戏主界面
│           ├── GameBoard.css
│           ├── Card.tsx         # 卡牌组件
│           ├── Card.css
│           ├── CompoundSelector.tsx # 物质选择浮窗
│           ├── CompoundSelector.css
│           ├── Setup.tsx        # 游戏设置组件
│           ├── Setup.css
│           ├── AdminPanel.tsx   # 管理面板（修改反应规则）
│           ├── AdminPanel.css
│           ├── AdminLogin.tsx   # 管理员登录
│           ├── AdminLogin.css
│           └── *.js             # 编译前的旧JS文件（待清理）
└── docs/                        # 完整文档目录
    ├── README.md                # 文档中心
    ├── ADMIN_PANEL_GUIDE.md     # 管理面板使用指南
    ├── MOBILE_ACCESS_GUIDE.md   # 移动端访问指南
    ├── DEPLOYMENT_GUIDE.md      # 部署指南
    ├── DEVELOPER_GUIDE.md       # 开发者指南
    └── ... （更多文档）
```

## 🚀 快速开始

### 前置要求
- **Node.js** >= 14.0
- **pnpm** >= 8.0 ([安装指南](https://pnpm.io/installation))

> 💡 项目已从 npm 迁移到 pnpm，具有更快的安装速度和更小的磁盘占用。

### 开发环境

```bash
# 1. 安装依赖（使用pnpm workspace自动安装所有子包）
pnpm install

# 2. 启动开发服务器（并发启动前后端）
pnpm start
# 或使用开发模式（热重载）
pnpm run dev

# 3. 访问应用
# 前端：http://localhost:3000
# 后端：http://localhost:5000
```

### 生产环境部署

#### 快速部署

```bash
# 默认 Docker 部署
pnpm run prod:deploy:docker

# PM2 部署（Linux/macOS）
pnpm run prod:deploy:pm2

# Systemd 部署（Linux）
pnpm run prod:deploy:systemd

# 旧版本支持（仅 Docker 或无 Docker）
pnpm run deploy:prod          # Docker 模式
pnpm run deploy:prod:no-docker # 无 Docker 模式
```

#### 部署模式选择

| 模式 | 推荐场景 | 依赖 |
|------|--------|------|
| **Docker** | 生产环境，多服务器 | Docker, Docker Compose |
| **PM2** | Linux/macOS 生产环境 | Node.js, PM2 |
| **Systemd** | Linux 系统集成 | Node.js, systemd |
| **Direct** | 开发/测试 | Node.js 仅 |

#### 服务管理

```bash
# 启动/停止/重启/查看日志
pnpm run prod:start docker
pnpm run prod:stop docker
pnpm run prod:restart docker
pnpm run prod:logs docker

# 健康检查和更新
pnpm run prod:health
pnpm run prod:upgrade docker
pnpm run prod:backup docker
```

详见 [完整部署指南](docs/PRODUCTION_ENVIRONMENT.md) 或 [快速部署指南](docs/QUICK_DEPLOY.md)

### 游戏入门

#### 电脑端
1. 打开浏览器访问 `http://localhost:4000`
2. 输入玩家名称
3. 选择"创建游戏"并选择玩家数量
4. 点击"创建游戏"按钮
5. 分享房间二维码或房间号给其他玩家

#### 移动端（手机/平板）
1. **确保设备与电脑在同一WiFi网络**
2. **获取电脑IP地址**：
   - Windows: 运行 `ipconfig`，查看 IPv4 地址
   - macOS/Linux: 运行 `ifconfig` 或 `ip addr`
3. **手机浏览器访问**：`http://[电脑IP]:4000`
   - 例如：`http://192.168.1.100:4000`
4. 扫描房间二维码或输入房间号加入游戏

> 💡 提示：创建房间后会自动生成二维码，手机扫码即可快速加入！

#### 管理面板（修改游戏配置）
1. 访问 `http://localhost:4000/admin`
2. 输入管理员密码（在 `.env` 文件中配置 `REACT_APP_ADMIN`）
3. 可以添加、修改、删除化学反应规则
4. 点击"保存配置"使修改生效

详细使用说明：[管理面板指南](docs/ADMIN_PANEL_GUIDE.md)

## 🔌 API 端点

### REST API

| 方法 | 端点 | 描述 |
|-----|------|------|
| POST | `/api/game/create` | 创建新游戏 |
| GET | `/api/game/:gameId/:playerId` | 获取游戏状态 |
| POST | `/api/compounds` | 获取元素能组成的物质 |
| POST | `/api/reaction/check` | 检查两个物质是否能反应 |
| GET | `/api/qrcode` | 生成房间二维码 |
| GET | `/api/config` | 获取游戏配置 |
| POST | `/api/config` | 保存游戏配置（管理员）|
| GET | `/api/mobile-info` | 获取移动端访问信息 |

### WebSocket 事件

#### 发送事件
- `joinGame` - 加入游戏（支持重新加入）
- `playCard` - 打出卡牌
- `drawCard` - 摸牌
- `startGame` - 开始游戏（房主）
- `disconnect` - 断开连接

#### 接收事件
- `gameStateUpdate` - 游戏状态更新
- `gameOver` - 游戏结束
- `playerRejoined` - 玩家重新加入
- `error` - 错误消息

## 🧪 核心算法

### 物质反应匹配

1. **元素提取**：从化学式中提取所有元素
   ```
   H2O → [H, O]
   Ca(OH)2 → [Ca, O, H]
   ```

2. **可用物质查找**：根据玩家拥有的元素，查找能组成的所有物质
   ```
   拥有元素: [H, O, Na, Cl]
   可组成: [H2O, NaCl, HCl, NaOH, ...]
   ```

3. **反应验证**：检查两个物质是否在反应数据库中有记录
   ```
   AgNO3 + NaCl → AgCl↓ + NaNO3 ✓
   ```

## 🎯 主要功能

### 核心游戏功能
- ✅ **实时多人游戏**：基于WebSocket的实时同步
- ✅ **智能物质推荐**：根据持有卡牌自动推荐可打出的物质
- ✅ **化学反应验证**：自动检验物质是否能反应
- ✅ **丰富的卡牌类型**：元素卡、特殊卡牌、功能卡
- ✅ **游戏逻辑完整**：支持方向反转、跳过、摸牌等规则
- ✅ **观战模式**：支持中途断线重连和观战
- ✅ **昵称记忆**：自动记住玩家昵称（会话级别）

### 移动端支持
- ✅ **跨设备访问**：支持手机、平板通过局域网访问
- ✅ **二维码分享**：自动生成房间二维码，扫码即可加入
- ✅ **自动API检测**：自动识别访问设备，选择正确的服务器地址
- ✅ **触摸优化**：移动端触摸交互优化
- ✅ **响应式设计**：支持不同屏幕尺寸

### 管理功能
- ✅ **管理面板**：可视化修改游戏配置（访问 `/admin`）
- ✅ **反应规则管理**：动态添加/修改/删除化学反应规则
- ✅ **配置持久化**：配置保存到 `config.json` 文件
- ✅ **实时生效**：配置修改后立即生效，无需重启服务器

### 网络功能
- ✅ **局域网支持**：支持同一WiFi下多设备联机
- ✅ **跨平台**：Windows/macOS/Linux 全平台支持
- ✅ **服务器状态页**：访问服务器根路径查看状态和快速链接

## 🛠️ 技术栈

### 后端
- **Node.js >= 14.0** - JavaScript运行时
- **TypeScript 5.3+** - 类型安全的JavaScript超集
- **Express.js** - Web框架
- **Socket.IO 4.5+** - 实时双向通信
- **ts-node** - TypeScript直接执行
- **QRCode** - 二维码生成

### 前端
- **React 18** - UI框架
- **TypeScript 5.3+** - 类型安全开发
- **Socket.IO-client 4.5+** - WebSocket客户端
- **Axios** - HTTP客户端
- **React Scripts 5.0** - 构建工具链
- **CSS3** - 样式设计

### 工具链
- **pnpm 8.15+** - 快速、节省磁盘空间的包管理器
- **pnpm workspace** - Monorepo管理
- **Docker** - 容器化部署
- **concurrently** - 并发任务运行

### 开发工具
- **nodemon** - 开发热重载
- **TypeScript Compiler** - 类型检查和编译

## 📝 扩展可能性

1. **增强物质库**：添加更多化学物质和反应
2. **用户系统**：注册、登录、排行榜
3. **游戏记录**：保存游戏历史和分析
4. **AI对手**：实现智能机器人玩家
5. **移动端优化**：完全响应式设计
6. **教育模式**：显示化学反应方程式和原理
7. **多语言支持**：国际化适配

## ⚠️ 已知限制

- 仅支持2-12人游戏（4人游戏已经过验证）
- 物质库为简化示例（可根据需要扩展）
- 暂不支持离线游戏存档

---

## 📚 完整文档

所有详细文档已整理到 **[docs/](docs/)** 目录：

| 文档 | 说明 | 适用人群 |
|------|------|--------|
| [GETTING_STARTED.md](docs/GETTING_STARTED.md) | 启动应用完整指南 | 👥 所有用户 |
| [QUICK_REFERENCE.md](docs/QUICK_REFERENCE.md) | 快速参考卡 | ⚡ 快速查阅 |
| [QUICK_START.md](docs/QUICK_START.md) | 游戏规则速查表 | 🎮 游戏玩家 |
| [INSTALLATION_GUIDE.md](docs/INSTALLATION_GUIDE.md) | 详细安装指南 | 🔧 安装问题 |
| [DEVELOPER_GUIDE.md](docs/DEVELOPER_GUIDE.md) | 技术架构和API | 👨‍💻 开发者 |
| [PROJECT_SUMMARY.md](docs/PROJECT_SUMMARY.md) | 项目技术总结 | 📊 项目概览 |
| [FILE_MANIFEST.md](docs/FILE_MANIFEST.md) | 完整文件清单 | 📋 文件说明 |
| [CLEANUP_SUMMARY.md](docs/CLEANUP_SUMMARY.md) | 项目简化说明 | 📝 了解变更 |
| [INDEX.md](docs/INDEX.md) | 项目索引 | 🔍 查找内容 |
| [COMPLETION_REPORT.md](docs/COMPLETION_REPORT.md) | 完成报告 | ✅ 项目状态 |

## 📄 许可证

MIT License

## 👥 贡献

欢迎提交问题和改进建议！

---

**享受化学UNO的乐趣！** ⚗️🃏
