# 快速启动指南 🚀

## 本地开发（SQLite - 推荐）

### 1. 最简单的方式

```bash
cd backend

# 复制配置文件
copy .env.example .env

# 确保 DB_TYPE=sqlite（默认配置）
# 不需要安装MySQL！

# 启动服务
go run .
```

就这么简单！数据库文件会自动创建在 `./chemistryuno.db`

### 2. 默认账户

首次启动后会自动创建管理员账户：

- **用户名**: `admin`
- **密码**: `admin123`

### 3. 访问应用

- **后端API**: <http://localhost:8080>
- **前端界面**: <http://localhost:5173（需要另开终端启动前端）>

---

## 切换到MySQL（生产环境）

### 1. 安装MySQL

确保MySQL服务已启动，并创建数据库：

```sql
CREATE DATABASE chemistryuno CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 2. 修改配置

编辑 `.env` 文件：

```bash
# 修改数据库类型
DB_TYPE=mysql

# 配置MySQL连接
MYSQL_DSN=root:your_password@tcp(localhost:3306)/chemistryuno?charset=utf8mb4&parseTime=True&loc=Local
```

### 3. 重启服务

```bash
go run .
```

GORM会自动迁移表结构和初始化数据。

---

## Redis缓存（可选）

### 为什么需要Redis？

- ⚡ 大幅提升高并发性能
- 📊 减少数据库查询压力
- 🎯 支持限流和会话管理

### 安装和启动

**Windows:**

```bash
# 下载 Redis for Windows
# 或使用 WSL: wsl --install
# 然后在WSL中: sudo apt install redis-server

redis-server
```

**Linux/Mac:**

```bash
# Ubuntu/Debian
sudo apt install redis-server
sudo systemctl start redis

# Mac
brew install redis
brew services start redis
```

### 配置

在 `.env` 中添加：

```bash
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
```

**注意**: Redis连接失败不会影响应用运行，只是缓存功能不可用。

---

## 构建生产版本

```bash
# 编译
go build -o chemistryuno.exe .

# 运行
./chemistryuno.exe
```

---

## 常见问题

### Q1: 端口已被占用？

修改 `main.go` 中的端口号，或在环境变量中设置 `PORT`。

### Q2: SQLite数据库文件在哪？

默认在 `backend/chemistryuno.db`，可以通过 `SQLITE_PATH` 环境变量修改。

### Q3: JWT密钥警告？

首次启动会自动生成，保存在 `.env` 中。生产环境请妥善保管此文件。

### Q4: Redis连接失败？

这是正常的，如果不需要缓存功能，可以忽略。应用会自动降级使用纯数据库模式。

### Q5: 如何重置数据库？

**SQLite**: 删除 `chemistryuno.db` 文件，重新启动
**MySQL**:

```sql
DROP DATABASE chemistryuno;
CREATE DATABASE chemistryuno CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

---

## 性能对比

| 场景 | SQLite | MySQL | MySQL+Redis |
|------|--------|-------|-------------|
| 本地开发 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ |
| 小团队 | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| 生产环境 | ⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| 高并发 | ⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| 配置复杂度 | 简单 | 中等 | 中等 |

---

## 技术栈

- **Go** 1.24+
- **GORM** - ORM框架
- **Gin** - Web框架
- **SQLite** / **MySQL** - 数据库
- **Redis** - 缓存（可选）
- **JWT** - 身份认证
- **WebAuthn** - 无密码登录

---

## 下一步

- 📖 阅读 [UPGRADE_GUIDE.md](UPGRADE_GUIDE.md) 了解详细架构
- 🔐 查看 [安全特性](../README.md#安全特性)
- 🎮 开始游戏开发！
