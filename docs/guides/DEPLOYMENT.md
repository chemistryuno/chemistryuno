# Chemistry UNO Mendeleef V1.2.0 - 部署指南

> **Linux Ubuntu 22.04 + 宝塔面板完整部署方案**

## 📋 目录

- [系统要求](#系统要求)
- [快速开始](#快速开始)
- [详细步骤](#详细步骤)
- [配置说明](#配置说明)
- [常见问题](#常见问题)

---

## 🔧 系统要求

### 服务器环境
- **操作系统**: Linux Ubuntu 22.04 LTS (推荐)
- **面板**: 宝塔 Linux 面板 7.x/8.x
- **架构**: x86_64 (AMD64)

### 必需服务
- ✅ **数据库** (二选一):
  - MySQL 5.7+ / 8.0+ (生产环境推荐)
  - SQLite 3.x (内置，开发/小型部署)

### 可选服务
- 🔶 **Redis 6.x+** (推荐，用于缓存和高并发优化)
- 🔶 **SMTP 邮箱服务** (用于邮箱验证码和通知)

---

## 🚀 快速开始

### 方式一：自动化部署（推荐）

```bash
# 1. 构建发布包（在开发机上）
cd /path/to/chemistryuno
pnpm build

# 2. 上传 dist 目录到服务器
scp -r dist/ user@your-server:/www/wwwroot/chemistryuno/

# 3. SSH 登录服务器
ssh user@your-server

# 4. 进入项目目录
cd /www/wwwroot/chemistryuno/dist

# 5. 配置环境变量
cp .env.example .env
nano .env  # 编辑配置

# 6. 赋予执行权限
chmod +x chemistryuno start.sh

# 7. 启动服务
./start.sh
```

### 方式二：宝塔面板部署

1. **上传文件**
   - 进入宝塔文件管理器
   - 上传 `dist` 目录到 `/www/wwwroot/chemistryuno/`

2. **配置 .env**
   - 复制 `.env.example` 为 `.env`
   - 编辑配置文件（数据库、OAuth 等）

3. **赋予权限**
   ```bash
   chmod +x /www/wwwroot/chemistryuno/dist/chemistryuno
   chmod +x /www/wwwroot/chemistryuno/dist/start.sh
   ```

4. **添加守护进程**（宝塔 > 软件商店 > Supervisor）
   - 名称: `chemistryuno`
   - 运行目录: `/www/wwwroot/chemistryuno/dist`
   - 启动命令: `/www/wwwroot/chemistryuno/dist/chemistryuno`
   - 进程数量: `1`

---

## 📝 详细步骤

### 第一步：构建项目

在**开发机**上构建项目：

```bash
# 克隆项目（如果还没有）
git clone https://github.com/yourusername/chemistryuno.git
cd chemistryuno

# 安装依赖
pnpm install
pnpm -C frontend install

# 构建发布包
pnpm build
```

构建完成后，`dist/` 目录结构如下：

```
dist/
├── chemistryuno          # Go 二进制可执行文件
├── .env.example          # 配置文件模板
├── start.sh              # Linux 启动脚本
├── README.md             # 部署说明
└── frontend/             # 前端静态文件（已嵌入二进制）
```

### 第二步：上传到服务器

#### 方式 A: SCP 上传

```bash
# 压缩 dist 目录
cd dist
tar -czf chemistryuno-dist.tar.gz *

# 上传到服务器
scp chemistryuno-dist.tar.gz user@your-server:/tmp/

# SSH 登录服务器并解压
ssh user@your-server
cd /www/wwwroot/chemistryuno
mkdir -p dist
cd dist
tar -xzf /tmp/chemistryuno-dist.tar.gz
```

#### 方式 B: 宝塔面板上传

1. 登录宝塔面板
2. 文件管理 → 进入 `/www/wwwroot/`
3. 创建目录 `chemistryuno`
4. 上传 `dist` 目录内的所有文件

### 第三步：配置环境变量

#### 1. 创建 .env 文件

```bash
cd /www/wwwroot/chemistryuno/dist
cp .env.example .env
nano .env
```

#### 2. 配置必要项

##### 数据库配置（必须）

**选项 A: 使用 SQLite（简单，适合小型部署）**

```env
DB_TYPE=sqlite
SQLITE_PATH=./chemistryuno.db
```

**选项 B: 使用 MySQL（推荐生产环境）**

1. 在宝塔面板创建数据库：
   - 数据库名: `chemistryuno`
   - 用户名: `chemistryuno`
   - 密码: `your_secure_password`

2. 配置 .env：
```env
DB_TYPE=mysql
MYSQL_DSN=chemistryuno:your_secure_password@tcp(localhost:3306)/chemistryuno?charset=utf8mb4&parseTime=True&loc=Local
```

##### JWT 密钥（必须）

```env
# 首次启动会自动生成，也可手动设置（至少 32 字符）
JWT_SECRET=your-random-secret-key-at-least-32-characters-long
```

##### Redis 配置（可选，推荐）

如果安装了 Redis：

```env
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=your_redis_password  # 如果设置了密码
# REDIS_DB=0
```

如果没有 Redis，留空即可（功能会自动降级）。

##### OAuth 配置（可选）

**GitHub OAuth**:
```env
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
GITHUB_REDIRECT_URI=https://yourdomain.com/api/auth/github/callback
```

**Microsoft OAuth**:
```env
MS_CLIENT_ID=your_ms_client_id
MS_CLIENT_SECRET=your_ms_client_secret
MS_TENANT_ID=common
MS_REDIRECT_URI=https://yourdomain.com/api/auth/ms/callback
```

**Google OAuth**:
```env
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URI=https://yourdomain.com/api/auth/google/callback
```

> **注意**: 未配置的 OAuth 提供商，其登录按钮会自动隐藏，不影响其他功能。

##### WebAuthn 配置（可选）

如果需要支持硬件密钥登录（如 YubiKey）：

```env
WEBAUTHN_RPID=yourdomain.com
WEBAUTHN_ORIGIN=https://yourdomain.com
```

##### SMTP 邮箱配置（可选）

用于邮箱验证码和通知：

```env
SMTP_HOST=smtp.qq.com          # QQ邮箱
SMTP_PORT=587
SMTP_USER=your_email@qq.com
SMTP_PASS=your_smtp_password   # QQ邮箱授权码
SMTP_FROM=your_email@qq.com
```

### 第四步：赋予执行权限

```bash
cd /www/wwwroot/chemistryuno/dist
chmod +x chemistryuno
chmod +x start.sh
```

### 第五步：首次启动和初始化

#### 方式 A: 直接启动（测试）

```bash
./chemistryuno
```

首次启动时，程序会自动：
1. ✅ 检查并加载 .env 配置
2. ✅ 连接数据库（MySQL 或 SQLite）
3. ✅ 创建所有必需的表结构
4. ✅ 初始化默认管理员账户（admin@chemistryuno.com / 123456）
5. ✅ 初始化默认游戏数据（化学反应、物质等）
6. ✅ 尝试连接 Redis（如果配置）
7. ✅ 加载 OAuth 配置

启动日志示例：

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚠️  警告: 未找到 .env 配置文件

📝 首次部署步骤:
   1. 复制配置文件模板:
      Linux/macOS: cp .env.example .env
      Windows:     copy .env.example .env

   2. 编辑 .env 文件，配置以下必要项:
      - DB_TYPE (数据库类型: sqlite 或 mysql)
      - JWT_SECRET (将在首次启动时自动生成)
      - OAuth 配置 (可选，用于第三方登录)

   3. 重新启动程序

💡 程序将使用默认配置继续启动 (SQLite 数据库)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🚀 Chemistry UNO 数据库初始化
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 使用 SQLite 数据库 (纯Go驱动, WAL模式)
   数据库文件: ./chemistryuno.db
⚙️  SQLite连接池配置: MaxIdle=2, MaxOpen=10 (WAL模式优化)
ℹ️  Redis未配置（缓存功能已禁用，不影响核心功能）
💡 如需启用Redis，请设置环境变量: REDIS_ADDR=localhost:6379
🔄 开始数据库迁移和初始化...
📊 检测到首次启动，正在初始化数据库表结构...
✅ 数据库表结构初始化成功
✅ 数据库迁移完成
🔧 初始化默认数据...
👤 创建默认管理员账户...
✅ 默认管理员账户创建成功 (admin@chemistryuno.com / 123456)
🃏 创建默认全局牌组配置...
✅ 默认牌组配置创建成功
⚗️  初始化默认物质数据...
✅ 默认物质数据初始化成功
🧪 初始化默认化学反应数据...
✅ 默认反应数据初始化成功
✅ 默认数据初始化完成
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ 数据库初始化完成
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔐 初始化 OAuth 配置...
💡 未配置的 OAuth 提供商: GitHub, Microsoft, Google, Apple (登录按钮将自动隐藏)
   如需启用，请在 .env 文件中配置相应的 CLIENT_ID 和 CLIENT_SECRET
✓ 前端静态文件服务已启用（embed 模式）
✅ 服务器准备启动在 :8080
✅ 服务器启动在 :8080
```

#### 方式 B: 使用启动脚本

```bash
./start.sh
```

#### 方式 C: 后台运行

```bash
nohup ./chemistryuno > server.log 2>&1 &

# 查看日志
tail -f server.log

# 停止服务
pkill chemistryuno
```

### 第六步：配置守护进程（推荐）

#### 使用宝塔 Supervisor（推荐）

1. **安装 Supervisor**
   - 宝塔面板 → 软件商店 → 搜索 "Supervisor"
   - 点击安装

2. **添加守护进程**
   - 进入 Supervisor 管理界面
   - 点击 "添加守护进程"
   - 填写配置：

   ```
   名称: chemistryuno
   运行目录: /www/wwwroot/chemistryuno/dist
   启动命令: /www/wwwroot/chemistryuno/dist/chemistryuno
   进程数量: 1
   ```

3. **启动进程**
   - 点击 "启动"
   - 查看状态确认运行正常

#### 使用 systemd

创建服务文件 `/etc/systemd/system/chemistryuno.service`：

```ini
[Unit]
Description=Chemistry UNO Server
After=network.target mysql.service redis.service

[Service]
Type=simple
User=www
Group=www
WorkingDirectory=/www/wwwroot/chemistryuno/dist
ExecStart=/www/wwwroot/chemistryuno/dist/chemistryuno
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable chemistryuno
sudo systemctl start chemistryuno
sudo systemctl status chemistryuno

# 查看日志
sudo journalctl -u chemistryuno -f
```

### 第七步：配置反向代理（Nginx）

#### 宝塔面板配置

1. **添加站点**
   - 网站 → 添加站点
   - 域名: `yourdomain.com`
   - 根目录: `/www/wwwroot/chemistryuno/dist` (随意，不会用到)

2. **配置反向代理**
   - 进入站点设置 → 反向代理
   - 添加反向代理：

   ```
   代理名称: chemistryuno
   目标URL: http://127.0.0.1:8080
   发送域名: $host
   ```

3. **SSL 证书**
   - 站点设置 → SSL
   - Let's Encrypt → 申请证书

#### 手动 Nginx 配置

编辑站点配置文件 `/www/server/panel/vhost/nginx/yourdomain.com.conf`：

```nginx
upstream chemistryuno {
    server 127.0.0.1:8080;
    keepalive 64;
}

server {
    listen 80;
    listen 443 ssl http2;
    server_name yourdomain.com www.yourdomain.com;

    # SSL 证书配置
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    # 强制 HTTPS
    if ($server_port !~ 443) {
        rewrite ^(/.*)$ https://$host$1 permanent;
    }

    # 日志
    access_log /www/wwwlogs/chemistryuno_access.log;
    error_log /www/wwwlogs/chemistryuno_error.log;

    # 反向代理配置
    location / {
        proxy_pass http://chemistryuno;
        proxy_http_version 1.1;

        # WebSocket 支持
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # 传递真实 IP
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 超时配置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # 静态资源缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        proxy_pass http://chemistryuno;
        proxy_cache_valid 200 7d;
        expires 7d;
        add_header Cache-Control "public, immutable";
    }
}
```

重载 Nginx：

```bash
sudo nginx -t
sudo nginx -s reload
```

### 第八步：防火墙配置

```bash
# 允许 HTTP/HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# 如果直接访问 8080 端口（不推荐生产环境）
sudo ufw allow 8080/tcp

# 查看状态
sudo ufw status
```

---

## 📚 配置说明

### 环境变量完整列表

| 变量名 | 必需 | 默认值 | 说明 |
|--------|------|--------|------|
| `DB_TYPE` | 否 | `sqlite` | 数据库类型: `sqlite` 或 `mysql` |
| `SQLITE_PATH` | 否 | `./chemistryuno.db` | SQLite 数据库文件路径 |
| `MYSQL_DSN` | MySQL时必需 | - | MySQL 连接字符串 |
| `JWT_SECRET` | 是 | 自动生成 | JWT 签名密钥（至少 32 字符） |
| `REDIS_ADDR` | 否 | - | Redis 地址（如 `localhost:6379`） |
| `REDIS_PASSWORD` | 否 | - | Redis 密码 |
| `REDIS_DB` | 否 | `0` | Redis 数据库编号 |
| `GITHUB_CLIENT_ID` | 否 | - | GitHub OAuth Client ID |
| `GITHUB_CLIENT_SECRET` | 否 | - | GitHub OAuth Client Secret |
| `GITHUB_REDIRECT_URI` | 否 | - | GitHub OAuth 回调地址 |
| `MS_CLIENT_ID` | 否 | - | Microsoft OAuth Client ID |
| `MS_CLIENT_SECRET` | 否 | - | Microsoft OAuth Client Secret |
| `MS_TENANT_ID` | 否 | `common` | Microsoft Azure AD Tenant ID |
| `MS_REDIRECT_URI` | 否 | - | Microsoft OAuth 回调地址 |
| `GOOGLE_CLIENT_ID` | 否 | - | Google OAuth Client ID |
| `GOOGLE_CLIENT_SECRET` | 否 | - | Google OAuth Client Secret |
| `GOOGLE_REDIRECT_URI` | 否 | - | Google OAuth 回调地址 |
| `APPLE_CLIENT_ID` | 否 | - | Apple OAuth Client ID |
| `APPLE_REDIRECT_URI` | 否 | - | Apple OAuth 回调地址 |
| `WEBAUTHN_RPID` | 否 | `localhost` | WebAuthn 域名 |
| `WEBAUTHN_ORIGIN` | 否 | - | WebAuthn 完整 URL |
| `SMTP_HOST` | 否 | - | SMTP 服务器地址 |
| `SMTP_PORT` | 否 | `587` | SMTP 端口 |
| `SMTP_USER` | 否 | - | SMTP 用户名 |
| `SMTP_PASS` | 否 | - | SMTP 密码 |
| `SMTP_FROM` | 否 | - | 发件人邮箱 |

### OAuth 申请指南

#### GitHub OAuth

1. 访问 https://github.com/settings/developers
2. 点击 "New OAuth App"
3. 填写信息：
   - Application name: `Chemistry UNO`
   - Homepage URL: `https://yourdomain.com`
   - Authorization callback URL: `https://yourdomain.com/api/auth/github/callback`
4. 获取 Client ID 和 Client Secret

#### Microsoft OAuth

1. 访问 https://portal.azure.com
2. Azure Active Directory → 应用注册 → 新注册
3. 重定向 URI: `https://yourdomain.com/api/auth/ms/callback`
4. 证书和密码 → 新建客户端密码

#### Google OAuth

1. 访问 https://console.cloud.google.com/
2. 创建项目 → API和服务 → OAuth 同意屏幕
3. 创建凭据 → OAuth 客户端 ID → Web 应用
4. 已获授权的重定向 URI: `https://yourdomain.com/api/auth/google/callback`

---

## ❓ 常见问题

### Q1: 首次启动时 OAuth 按钮不显示？

**A**: 这是正常的！OAuth 登录功能是可选的。

- 如果 `.env` 文件中 **未配置** OAuth CLIENT_ID，对应的登录按钮会自动隐藏
- 不影响用户名密码登录和 WebAuthn 登录
- 如需启用 OAuth，请在 `.env` 中配置相应的 CLIENT_ID 和 CLIENT_SECRET

验证配置是否生效：

```bash
# 访问配置接口
curl http://localhost:8080/api/auth/config

# 应返回类似：
{
  "github_enabled": false,
  "ms_enabled": false,
  "google_enabled": false,
  "apple_enabled": false
}
```

### Q2: 数据库初始化失败？

**A**: 检查以下几点：

1. **MySQL 情况**：
   ```bash
   # 1. 确认数据库已创建
   mysql -u root -p
   SHOW DATABASES LIKE 'chemistryuno';

   # 2. 确认用户权限
   SHOW GRANTS FOR 'chemistryuno'@'localhost';

   # 3. 测试连接
   mysql -u chemistryuno -p chemistryuno
   ```

2. **SQLite 情况**：
   ```bash
   # 确认目录有写权限
   ls -la chemistryuno.db
   chmod 666 chemistryuno.db
   chmod 777 .
   ```

### Q3: Redis 连接失败但程序仍在运行？

**A**: 这是正常的！Redis 是可选组件。

- 程序会自动降级，禁用缓存功能
- 核心功能不受影响
- 建议生产环境安装 Redis 以提升性能

安装 Redis (宝塔面板)：
- 软件商店 → 搜索 "Redis" → 安装
- 默认端口: 6379
- 在 .env 中配置: `REDIS_ADDR=localhost:6379`

### Q4: 如何修改默认管理员密码？

**A**: 首次启动后：

1. 访问 `https://yourdomain.com`
2. 使用默认账号登录:
   - 用户名: `admin@chemistryuno.com`
   - 密码: `123456`
3. 进入设置 → 修改密码

或通过数据库直接修改（需要重新哈希密码）。

### Q5: WebSocket 连接失败？

**A**: 检查 Nginx 配置：

```nginx
location / {
    proxy_pass http://127.0.0.1:8080;
    proxy_http_version 1.1;

    # 必须包含这两行
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
}
```

### Q6: 如何查看服务器日志？

**A**:

- **Supervisor 方式**: 宝塔面板 → Supervisor → 查看日志
- **systemd 方式**: `sudo journalctl -u chemistryuno -f`
- **手动运行**: `./chemistryuno > server.log 2>&1 &` 然后 `tail -f server.log`

### Q7: 如何升级到新版本？

**A**:

1. 备份数据库

   ```bash
   # SQLite
   cp chemistryuno.db chemistryuno.db.backup

   # MySQL
   mysqldump -u chemistryuno -p chemistryuno > backup.sql
   ```

2. 停止服务

   ```bash
   # Supervisor
   宝塔面板 → Supervisor → 停止

   # systemd
   sudo systemctl stop chemistryuno

   # 手动
   pkill chemistryuno
   ```

3. 替换二进制文件

   ```bash
   # 备份旧版本
   mv chemistryuno chemistryuno.old

   # 上传新版本
   # (通过 SCP 或宝塔面板上传)

   # 赋予执行权限
   chmod +x chemistryuno
   ```

4. 重启服务

   ```bash
   # Supervisor
   宝塔面板 → Supervisor → 启动

   # systemd
   sudo systemctl start chemistryuno

   # 手动
   ./start.sh
   ```

5. 验证新版本

   ```bash
   curl http://localhost:8080/api/health
   ```

### Q8: 如何备份和恢复？

**A**:

**备份**:

```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/www/backup/chemistryuno"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# 备份数据库
if [ "$DB_TYPE" = "mysql" ]; then
    mysqldump -u chemistryuno -p chemistryuno > $BACKUP_DIR/db_$DATE.sql
else
    cp chemistryuno.db $BACKUP_DIR/db_$DATE.db
fi

# 备份配置
cp .env $BACKUP_DIR/env_$DATE

echo "备份完成: $BACKUP_DIR"
```

**恢复**:

```bash
# MySQL
mysql -u chemistryuno -p chemistryuno < backup.sql

# SQLite
cp backup.db chemistryuno.db
```

### Q9: 性能优化建议？

**A**:

1. **使用 MySQL 替代 SQLite** (生产环境)
2. **启用 Redis 缓存**
3. **配置 Nginx 缓存**
4. **启用 Gzip 压缩**
5. **使用 CDN 加速静态资源**
6. **优化数据库连接池** (在 .env 中配置)

   ```env
   DB_MAX_IDLE_CONNS=10
   DB_MAX_OPEN_CONNS=100
   DB_CONN_MAX_LIFETIME=3600
   REDIS_POOL_SIZE=100
   REDIS_MIN_IDLE_CONNS=10
   ```

### Q10: 如何启用 HTTPS？

**A**: 使用宝塔面板申请免费 SSL 证书：

1. 网站设置 → SSL
2. Let's Encrypt → 申请
3. 强制 HTTPS → 开启

或手动配置：

```nginx
server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    # ... 其他配置
}
```

---

## 🔗 相关链接

- **项目地址**: https://github.com/yourusername/chemistryuno
- **问题反馈**: https://github.com/yourusername/chemistryuno/issues
- **在线文档**: https://docs.chemistryuno.com

---

## 📄 许可证

本项目采用 [MIT License](LICENSE)

---

**祝部署顺利！** 🎉

如有问题，请提交 Issue 或联系技术支持。
