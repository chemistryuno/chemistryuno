# Chemistry UNO - Linux 快速部署指南

> **适用于 Ubuntu 22.04 + 宝塔面板**

## 🚀 30秒快速开始

```bash
# 1. 上传并解压
cd /www/wwwroot/chemistryuno/dist

# 2. 配置环境
cp .env.example .env
nano .env  # 编辑数据库等配置

# 3. 赋予权限
chmod +x chemistryuno start.sh

# 4. 启动
./start.sh
```

## 📋 配置检查清单

### ✅ 必须配置

- [ ] **数据库类型** (`DB_TYPE=sqlite` 或 `mysql`)
- [ ] **JWT密钥** (首次启动自动生成)

### 🔶 推荐配置

- [ ] **Redis** (`REDIS_ADDR=localhost:6379`)
- [ ] **域名和HTTPS** (通过Nginx反向代理)

### 🔷 可选配置

- [ ] **OAuth登录** (GitHub/Microsoft/Google/Apple)
- [ ] **WebAuthn** (硬件密钥登录)
- [ ] **SMTP邮箱** (邮件通知)

## 🎯 首次启动会自动

1. ✅ 创建所有数据库表
2. ✅ 初始化默认管理员 (`admin@chemistryuno.com` / `123456`)
3. ✅ 初始化游戏数据 (化学反应、物质等)
4. ✅ 自动检测Redis (未配置会自动降级)
5. ✅ 加载OAuth配置 (未配置会隐藏按钮)

## ⚙️ MySQL 配置示例

```bash
# 1. 在宝塔创建数据库
数据库名: chemistryuno
用户名: chemistryuno
密码: your_password

# 2. 编辑 .env
DB_TYPE=mysql
MYSQL_DSN=chemistryuno:your_password@tcp(localhost:3306)/chemistryuno?charset=utf8mb4&parseTime=True&loc=Local
```

## 🔧 守护进程 (Supervisor)

**宝塔面板设置:**

- 名称: `chemistryuno`
- 运行目录: `/www/wwwroot/chemistryuno/dist`
- 启动命令: `/www/wwwroot/chemistryuno/dist/chemistryuno`

## 🌐 Nginx 反向代理

**宝塔面板反向代理:**

- 目标URL: `http://127.0.0.1:8080`
- 发送域名: `$host`

## ❓ 常见问题

### Q: OAuth 按钮不显示？

**A**: 正常！未配置的 OAuth 会自动隐藏，不影响其他登录方式。

### Q: Redis 连接失败？

**A**: 正常！Redis 是可选的，程序会自动降级，核心功能不受影响。

### Q: 数据库初始化失败？

**A**: 检查：
- MySQL: 数据库是否已创建？用户权限是否正确？
- SQLite: 目录是否有写权限？

### Q: 如何查看日志？

**A**:
- Supervisor: 宝塔面板 → Supervisor → 查看日志
- 手动: `./chemistryuno > server.log 2>&1 &` 然后 `tail -f server.log`

## 📖 完整文档

详细部署指南请查看: [DEPLOYMENT.md](./DEPLOYMENT.md)

---

**默认管理员账号:**

- 用户名: `admin@chemistryuno.com`
- 密码: `123456`

**⚠️ 首次登录后请立即修改密码！**

---

构建于 V1.2.0 | Chemistry UNO Mendeleef
