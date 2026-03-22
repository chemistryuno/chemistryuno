# 后端 API 文档与架构说明

## 📁 项目结构

```text
backend/
├── main.go                 # 主入口，路由定义
├── go.mod                  # Go模块依赖
├── database/              # 数据库层
│   ├── gorm.go           # GORM配置和连接
│   ├── migrate.go        # 数据库迁移
│   └── models.go         # GORM数据模型
├── repository/            # 数据访问层
│   ├── init.go           # Repository初始化
│   ├── user_repository.go
│   ├── session_repository.go
│   ├── game_repository.go
│   ├── deck_repository.go
│   ├── feedback_repository.go
│   ├── announcement_repository.go
│   ├── webauthn_repository.go
│   ├── reaction_repository.go
│   ├── substance_repository.go
│   ├── bounty_repository.go
│   ├── legacy_game_repository.go      # 兼容旧表结构
│   ├── legacy_reaction_repository.go  # 兼容旧表结构
│   └── legacy_substance_repository.go # 兼容旧表结构
├── handlers/              # HTTP处理层
│   ├── auth.go           # 认证相关：注册、登录
│   ├── 2fa.go            # 双因素认证
│   ├── webauthn.go       # WebAuthn无密码认证
│   ├── admin.go          # 管理员功能
│   ├── game.go           # 游戏房间管理
│   ├── deck.go           # 卡组配置
│   ├── feedback.go       # 用户反馈
│   ├── announcement.go   # 公告管理
│   └── points.go         # 积分和悬赏
├── middleware/            # 中间件
│   ├── auth.go           # JWT认证中间件
│   └── cors.go           # CORS跨域中间件
├── game/                  # 游戏逻辑层
│   ├── manager.go        # 房间和游戏状态管理
│   ├── chemistry.go      # 化学反应验证
│   ├── judge.go          # 游戏裁判逻辑
│   └── cron.go           # 定时任务
├── websocket/             # WebSocket实时通信
│   ├── hub.go            # WebSocket中心
│   └── client.go         # 客户端连接管理
├── models/                # 业务模型
│   ├── user.go           # 用户模型
│   ├── game.go           # 游戏模型
│   └── announcement.go   # 公告模型
└── utils/                 # 工具函数
    ├── jwt.go            # JWT令牌生成和解析
    ├── password.go       # 密码加密
    ├── session_helper.go # Session管理
    └── secret_generator.go # 密钥生成
```

## 🔗 API 端点总览

### 公开接口（无需认证）

#### 认证相关

- `POST /auth/register` - 用户注册
- `POST /auth/login` - 用户登录
- `POST /auth/2fa/verify` - 2FA验证登录
- `POST /auth/2fa/reset-password` - 2FA重置密码

#### WebAuthn

- `GET /auth/webauthn/login/begin` - 开始WebAuthn登录
- `POST /auth/webauthn/login/finish` - 完成WebAuthn登录

#### 其他

- `GET /announcements` - 获取活跃公告
- `GET /ping` - 健康检查
- `GET /health` - 详细健康状态

---

### 认证接口（需要JWT Token）

#### 用户管理

- `GET /user/info` - 获取当前用户信息
- `GET /user/game-history` - 获取游戏历史
- `PUT /user/password` - 修改密码
- `PUT /user/avatar` - 更新头像
- `DELETE /user/account` - 删除账号

#### 会话管理

- `GET /user/sessions` - 获取所有会话
- `POST /user/sessions/logout` - 撤销指定会话
- `POST /user/account/freeze` - 冻结账号

#### 2FA 管理

- `POST /user/2fa/setup` - 设置 2FA
- `POST /user/2fa/enable` - 启用 2FA
- `POST /user/2fa/disable` - 禁用 2FA

#### WebAuthn管理

- `GET /user/webauthn/register/begin` - 开始注册硬件密钥
- `POST /user/webauthn/register/finish` - 完成硬件密钥注册
- `GET /user/webauthn/credentials` - 列出所有硬件密钥
- `DELETE /user/webauthn/credentials/:id` - 删除硬件密钥

#### 反馈系统

- `POST /feedback` - 提交反馈
- `GET /feedbacks/my` - 获取我的反馈
- `POST /feedbacks/:id/urge` - 催促反馈处理
- `POST /feedback/withdraw` - 撤回反馈

#### 卡组管理

- `GET /my-decks` - 获取我的卡组
- `POST /my-decks` - 创建自定义卡组
- `PUT /my-decks/:id` - 更新卡组
- `DELETE /my-decks/:id` - 删除卡组

#### 化学反应

- `GET /reactions/my` - 获取我提交的反应
- `GET /reactions/all` - 获取所有已审核反应
- `POST /reactions` - 提交新反应
- `GET /reactions` - 获取反应列表（co-worker）
- `POST /reactions/batch` - 批量添加反应（co-worker）
- `PUT /reactions/:id` - 更新反应（co-worker）
- `PUT /reactions/approve/:group_id` - 批准反应（co-worker）
- `DELETE /reactions/:id` - 删除反应（admin）

#### 化学物质

- `GET /substances` - 获取物质列表
- `POST /substances` - 提交新物质
- `PUT /substances/:id` - 更新物质（co-worker）
- `PUT /substances/approve/:id` - 批准物质（co-worker）
- `DELETE /substances/:id` - 删除物质（admin）

#### 游戏相关

- `GET /rooms` - 获取所有房间
- `POST /rooms` - 创建房间
- `POST /game/duel` - 发起单挑
- `POST /game/duel/respond` - 响应单挑
- `GET /rooms/:id` - 获取房间状态
- `POST /rooms/:id/join` - 加入房间
- `POST /rooms/:id/leave` - 离开房间
- `POST /rooms/:id/start` - 开始游戏
- `POST /rooms/:id/play` - 出牌
- `POST /rooms/:id/play-double` - 双卡出牌
- `POST /rooms/:id/draw` - 摸牌
- `GET /rooms/:id/substances` - 获取可用物质
- `POST /game/check-reaction` - 验证化学反应

#### 积分系统

- `GET /points/leaderboard` - 获取排行榜
- `POST /points/bounty` - 设置悬赏

#### WebSocket

- `GET /ws` - WebSocket连接（实时游戏通信）

---

### 管理员接口（需要admin权限）

#### 用户管理（管理员）

- `GET /admin/users` - 获取所有用户
- `POST /admin/users` - 创建用户
- `DELETE /admin/users/:id` - 删除用户
- `PUT /admin/users/:id/password` - 重置用户密码
- `PUT /admin/users/:id/role` - 修改用户角色

#### 系统配置

- `GET /admin/deck-config` - 获取全局卡组配置
- `PUT /admin/deck-config` - 更新全局卡组配置
- `GET /admin/configs` - 获取系统配置
- `PUT /admin/configs` - 更新系统配置

#### 游戏历史

- `GET /admin/game-history` - 获取游戏历史记录

#### 反馈管理

- `GET /admin/feedbacks` - 获取所有反馈
- `PUT /admin/feedbacks/:id/status` - 更新反馈状态

#### 公告管理

- `GET /admin/announcements` - 获取所有公告
- `POST /admin/announcements` - 创建公告
- `PUT /admin/announcements/:id/status` - 更新公告状态
- `DELETE /admin/announcements/:id` - 删除公告

---

## 🔐 认证流程

### JWT Token

1. 登录成功后返回 JWT Token
2. 后续请求在 Header 中携带：`Authorization: Bearer <token>`
3. Token 包含：UID, Username, IsAdmin, Role, SID（Session ID）

### Session管理

- 每次登录创建新的Session
- Session包含：设备信息、IP地址、最后活跃时间
- Token中的SID与数据库Session关联
- 支持多设备同时登录

### 权限级别

- **user**: 普通用户
- **co-worker**: 协作者（可审核化学反应和物质）
- **admin**: 管理员（完全权限）

---

## 🎮 游戏架构

### 房间管理

- 房间状态：waiting（等待）、playing（游戏中）、finished（已结束）
- 支持普通模式和单挑模式
- 自定义卡组配置
- 观战者系统

### 实时通信

- WebSocket Hub管理所有连接
- 消息类型：
  - `game_state`: 游戏状态更新
  - `player_joined`: 玩家加入
  - `player_left`: 玩家离开
  - `game_started`: 游戏开始
  - `card_played`: 出牌
  - `turn_changed`: 回合变更
  - `game_finished`: 游戏结束

### 化学反应验证

- 自动验证元素组合是否能生成物质
- 支持正向和逆向反应
- 群组ID管理双向反应

---

## 📊 数据库设计

### 主要表结构

- `users` - 用户表
- `user_sessions` - 会话表
- `game_history` - 游戏历史
- `deck_configs` - 卡组配置
- `feedbacks` - 用户反馈
- `announcements` - 系统公告
- `reactions` - 化学反应
- `substances` - 化学物质
- `webauthn_credentials` - WebAuthn凭证
- `bounties` - 悬赏记录

### Repository模式

所有数据访问通过Repository层，提供：

- 类型安全的数据操作
- 统一的错误处理
- 易于测试和维护

---

## 🔧 中间件

### AuthMiddleware

- 验证JWT Token
- 检查Session有效性
- 防止Session劫持
- 更新活跃时间

### AdminMiddleware

- 验证管理员权限

### CoWorkerMiddleware

- 验证协作者或管理员权限

### CORSMiddleware

- 跨域资源共享配置

---

## ⚙️ 配置说明

### 环境变量

- `JWT_SECRET` - JWT密钥（自动生成）
- `DATABASE_URL` - 数据库连接（SQLite默认）
- `REDIS_URL` - Redis连接（可选）

### 数据库

- 默认：SQLite（纯Go实现，无需CGO）
- 支持：MySQL、PostgreSQL

---

## 🚀 启动流程

1. 初始化JWT密钥
2. 连接数据库
3. 运行数据库迁移
4. 初始化Repository层
5. 初始化Admin Handlers
6. 启动WebSocket Hub
7. 启动定时任务
8. 初始化WebAuthn
9. 配置路由
10. 启动HTTP服务器
11. 等待优雅关闭信号

---

## 📝 注意事项

### Legacy Repository

- `legacy_*_repository.go` 文件用于兼容旧表结构
- 主要在admin.go中使用
- 未来可能迁移到新的GORM模型

### 安全性

- 所有密码使用bcrypt加密
- Session ID使用crypto/rand生成
- 支持2FA和WebAuthn
- 会话验证防劫持

### 性能优化

- WebSocket长连接
- Redis缓存（可选）
- 数据库连接池
- 优雅关闭处理

---

生成时间: 2026-03-22
