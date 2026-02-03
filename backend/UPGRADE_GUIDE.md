# 数据库升级指南

## 更新内容

### 1. GORM + Redis 高并发优化

- ✅ 将原有的 `database/sql` 原生SQL查询迁移到GORM ORM框架
- ✅ 添加Redis缓存层，提高高并发场景下的性能
- ✅ 优化数据库连接池配置（50空闲/200最大连接）
- ✅ Redis连接池配置（100连接/10空闲连接）
- ✅ **支持SQLite和MySQL双数据库切换**

### 2. 灵活的数据库支持

- **SQLite**: 本地开发调试首选，零配置快速启动
- **MySQL**: 生产环境推荐，支持高并发和大数据量

切换方式：修改 `.env` 中的 `DB_TYPE` 即可！

### 3. 代码兼容性

所有旧代码无需修改即可运行！通过 `database.LegacyDB` 兼容层：

```go
// 旧代码（仍然有效）
database.LegacyDB.Query("SELECT * FROM users WHERE id = ?", id)
database.LegacyDB.QueryRow("SELECT name FROM users WHERE id = ?", id).Scan(&name)
database.LegacyDB.Exec("UPDATE users SET name = ? WHERE id = ?", name, id)

// 新代码（推荐使用GORM）
var user models.User
database.DB.Where("id = ?", id).First(&user)

// 使用Redis缓存
user, err := database.GetUserByIDWithCache(userID)
```

## 环境配置

### 方案1: SQLite（本地开发推荐）

**优点:**

- ✅ 零配置，开箱即用
- ✅ 无需安装MySQL服务
- ✅ 轻量级，适合快速开发调试
- ✅ 纯Go实现，无需CGO

**配置 `.env`:**

```bash
# 数据库类型选择
DB_TYPE=sqlite

# SQLite 数据库文件路径
SQLITE_PATH=./chemistryuno.db

# JWT密钥（首次启动会自动生成）
JWT_SECRET=

# Redis配置（可选，不配置则不使用缓存）
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
```

### 方案2: MySQL（生产环境推荐）

**优点:**

- ✅ 高性能，支持高并发
- ✅ 完善的备份和恢复方案
- ✅ 支持主从复制和集群
- ✅ 适合大数据量场景

**配置 `.env`:**

```bash
# 数据库类型选择
DB_TYPE=mysql

# MySQL 连接配置
MYSQL_DSN=root:password@tcp(localhost:3306)/chemistryuno?charset=utf8mb4&parseTime=True&loc=Local

# JWT密钥（首次启动会自动生成）
JWT_SECRET=

# Redis配置（生产环境强烈推荐）
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
```

### 安装依赖

```bash
cd backend
go mod tidy
```

### 启动应用

```bash
# 开发环境
go run .

# 生产环境
go build -o main.exe .
./main.exe
```

## 性能优化特性

### 1. 数据库连接池

- 最大空闲连接: 50
- 最大打开连接: 200
- 连接生命周期: 1小时
- 预编译语句缓存: 已启用
- 跳过默认事务: 已启用（提升写入性能）

### 2. Redis缓存策略

- 用户数据缓存: 10分钟 TTL
- 反应/物质数据: 5分钟 TTL
- 自动缓存失效机制
- 优雅降级（Redis不可用时自动降级到数据库）

### 3. 缓存辅助函数

```go
// 通用缓存操作
database.SetCache(ctx, key, value, ttl)
database.GetCache(ctx, key, &target)
database.DeleteCache(ctx, key)

// 业务缓存
user, err := database.GetUserByIDWithCache(userID)
reactions, err := database.GetReactionsWithCache()
substances, err := database.GetSubstancesWithCache()

// 缓存失效
database.InvalidateUserCache(userID)
database.InvalidateReactionCache()

// 限流
allowed, err := database.CheckRateLimit(ctx, userID, limit, window)
```

## 数据库模型

### 新增GORM模型（database/models.go）

所有模型都添加了：

- 自动时间戳（CreatedAt, UpdatedAt）
- 软删除支持（DeletedAt）
- 索引优化
- 外键约束

主要模型：

- `User` - 用户表
- `UserSession` - 会话表
- `WebAuthnCredential` - WebAuthn凭证
- `Reaction` - 化学反应
- `Substance` - 物质
- `Feedback` - 反馈
- `DeckConfig` - 牌组配置
- `GameHistory` - 游戏历史
- `Bounty` - 悬赏
- `Announcement` - 公告
- `SystemConfig` - 系统配置

### 自动数据库迁移

首次启动时会自动：

1. 创建所有表结构
2. 创建索引和外键
3. 初始化默认管理员账户（用户名: admin, 密码: admin123）
4. 初始化默认牌组配置

## 逐步迁移到GORM

虽然旧代码可以继续使用，但推荐逐步迁移到GORM以获得更好的性能和类型安全：

### 查询示例对比

```go
// 旧方式
var name string
err := database.LegacyDB.QueryRow("SELECT name FROM users WHERE id = ?", id).Scan(&name)

// 新方式（GORM）
var user models.User
err := database.DB.Where("id = ?", id).First(&user).Error
name := user.Name

// 新方式（带缓存）
user, err := database.GetUserByIDWithCache(id)
name := user.Name
```

### 插入示例对比

```go
// 旧方式
_, err := database.LegacyDB.Exec("INSERT INTO users (name, email) VALUES (?, ?)", name, email)

// 新方式
user := &models.User{Name: name, Email: email}
err := database.DB.Create(user).Error
```

### 更新示例对比

```go
// 旧方式
_, err := database.LegacyDB.Exec("UPDATE users SET name = ? WHERE id = ?", newName, id)

// 新方式
err := database.DB.Model(&models.User{}).Where("id = ?", id).Update("name", newName).Error
```

## 故障排查

### 1. 数据库连接失败

检查 `MYSQL_DSN` 环境变量是否正确配置，MySQL服务是否启动。

### 2. Redis连接失败

Redis是可选的。如果Redis不可用，系统会自动降级到纯数据库模式，不会影响功能。

### 3. JWT密钥警告

首次启动时如果 `JWT_SECRET` 为空，系统会自动生成一个50位的安全密钥并保存到 `.env` 文件。

### 4. 表结构不匹配

删除旧数据库，重新启动应用让GORM自动创建新表结构：

```sql
DROP DATABASE chemistryuno;
CREATE DATABASE chemistryuno CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## 性能监控建议

1. 监控数据库连接池使用情况
2. 监控Redis缓存命中率
3. 定期清理过期会话（已有定时任务）
4. 根据实际负载调整连接池配置

## 回滚方案

如果遇到问题需要回滚到旧版本：

1. 使用git恢复旧代码：

```bash
git checkout <旧版本commit>
```

1. 或者使用备份的 `database/db_old.go.bak` 文件：

```bash
cp database/db_old.go.bak database/db.go
```

## 支持

如有问题，请检查：

1. 日志文件中的错误信息
2. MySQL连接是否正常
3. Redis服务是否启动（可选）
4. 环境变量配置是否正确
