# 数据库配置切换示例

## ✅ 已实现功能

本项目支持 **SQLite** 和 **MySQL** 两种数据库，可通过修改 `.env` 文件随时切换。

---

## 配置示例

### 示例1: SQLite（本地开发）

**`.env` 文件内容:**

```bash
# 数据库类型
DB_TYPE=sqlite

# SQLite 数据库文件路径
SQLITE_PATH=./chemistryuno.db

# JWT密钥（自动生成）
JWT_SECRET=

# Redis（可选）
# REDIS_ADDR=localhost:6379
# REDIS_PASSWORD=
```

**特点:**

- ✅ 零配置
- ✅ 不需要安装MySQL
- ✅ 适合快速开发和测试
- ✅ 纯Go实现，无需CGO

---

### 示例2: MySQL（生产环境）

**`.env` 文件内容:**

```bash
# 数据库类型
DB_TYPE=mysql

# MySQL 连接字符串
MYSQL_DSN=root:password@tcp(localhost:3306)/chemistryuno?charset=utf8mb4&parseTime=True&loc=Local

# JWT密钥（自动生成）
JWT_SECRET=

# Redis（生产环境推荐）
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
```

**特点:**

- ✅ 高性能
- ✅ 支持高并发
- ✅ 连接池优化（50空闲/200最大）
- ✅ 适合生产部署

---

## 切换步骤

### 从 SQLite 切换到 MySQL

1. **启动MySQL服务**

   ```bash
   # Windows: 启动MySQL服务
   # Linux: sudo systemctl start mysql
   ```

2. **创建数据库**

   ```sql
   CREATE DATABASE chemistryuno CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

3. **修改 `.env`**

   ```bash
   DB_TYPE=mysql
   MYSQL_DSN=root:your_password@tcp(localhost:3306)/chemistryuno?charset=utf8mb4&parseTime=True&loc=Local
   ```

4. **重启应用**

   ```bash
   go run .
   ```

### 从 MySQL 切换到 SQLite

1. **修改 `.env`**

   ```bash
   DB_TYPE=sqlite
   SQLITE_PATH=./chemistryuno.db
   ```

2. **重启应用**

   ```bash
   go run .
   ```

就这么简单！

---

## 环境变量说明

| 变量 | 说明 | 默认值 | 必需 |
| --- | --- | --- | --- |
| `DB_TYPE` | 数据库类型 (sqlite/mysql) | `sqlite` | ✅ |
| `SQLITE_PATH` | SQLite数据库文件路径 | `./chemistryuno.db` | 当DB_TYPE=sqlite时 |
| `MYSQL_DSN` | MySQL连接字符串 | - | 当DB_TYPE=mysql时 |
| `JWT_SECRET` | JWT密钥 | 自动生成50位 | ❌ |
| `REDIS_ADDR` | Redis地址 | - | ❌ |
| `REDIS_PASSWORD` | Redis密码 | - | ❌ |

---

## 启动日志示例

### SQLite模式

```text
2026/02/03 21:31:24 📊 使用 SQLite 数据库 (纯Go): ./chemistryuno.db
2026/02/03 21:31:24 ⚠️ 警告: Redis连接失败 (将继续运行，但缓存功能不可用)
2026/02/03 21:31:24 开始数据库迁移...
2026/02/03 21:31:24 数据库迁移完成
2026/02/03 21:31:24 管理员账户创建成功
2026/02/03 21:31:24 默认牌组配置创建成功
2026/02/03 21:31:24 数据库初始化成功
2026/02/03 21:31:24 服务器启动在 :8080
```

### MySQL模式

```text
2026/02/03 21:31:24 📊 使用 MySQL 数据库
2026/02/03 21:31:24 ✅ Redis连接成功
2026/02/03 21:31:24 开始数据库迁移...
2026/02/03 21:31:24 数据库迁移完成
2026/02/03 21:31:24 管理员账户创建成功
2026/02/03 21:31:24 默认牌组配置创建成功
2026/02/03 21:31:24 数据库初始化成功
2026/02/03 21:31:24 服务器启动在 :8080
```

---

## 性能建议

### 开发环境

```bash
DB_TYPE=sqlite
# 不需要Redis，快速启动
```

### 测试环境

```bash
DB_TYPE=mysql
REDIS_ADDR=localhost:6379
# 模拟生产环境
```

### 生产环境

```bash
DB_TYPE=mysql
MYSQL_DSN=user:pass@tcp(db.example.com:3306)/chemistryuno?charset=utf8mb4&parseTime=True&loc=Local
REDIS_ADDR=redis.example.com:6379
REDIS_PASSWORD=your_redis_password
# 完整的高性能配置
```

---

## 故障排查

### 问题1: SQLite文件权限错误

**解决**: 确保应用有写入权限，或修改 `SQLITE_PATH` 到有权限的目录

### 问题2: MySQL连接失败

**检查清单**:

- [ ] MySQL服务是否启动？
- [ ] 数据库是否已创建？
- [ ] 用户名密码是否正确？
- [ ] 端口是否正确（默认3306）？

### 问题3: 切换数据库后数据丢失

这是正常的！SQLite和MySQL是独立的数据库。如需迁移数据，请使用数据导出/导入工具。

---

## 技术实现

使用GORM的方言（Dialector）系统实现多数据库支持：

```go
// database/gorm.go
switch dbType {
case "mysql":
    dialector = mysql.Open(dsn)
case "sqlite":
    dialector = sqlite.Dialector{
        DriverName: "sqlite",  // 使用modernc.org/sqlite
        DSN:        sqlitePath,
    }
}

DB, err = gorm.Open(dialector, &gorm.Config{...})
```

所有SQL查询通过GORM抽象层，自动适配不同数据库的语法差异。

---

## 更多信息

- 📖 [QUICKSTART.md](QUICKSTART.md) - 快速启动指南
- 📚 [UPGRADE_GUIDE.md](UPGRADE_GUIDE.md) - 完整升级文档
- 🔧 [.env.example](.env.example) - 配置文件模板
