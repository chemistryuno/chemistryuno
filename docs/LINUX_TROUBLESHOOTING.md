# Linux/Unix 部署常见问题

## 问题1: npm警告 "Unknown env config _-init-module"

### 症状
```
npm warn Unknown env config "_-init-module". This will stop working in the next major version of npm.
npm warn Unknown global config "--init.module". This will stop working in the next major version of npm.
```

### 原因
这是npm的遗留配置问题，不影响功能但会显示警告。

### 解决方案

#### 方案1: 使用清理脚本（推荐）
```bash
chmod +x clean-npm-config.sh
./clean-npm-config.sh
```

#### 方案2: 手动清理
```bash
npm config delete _-init-module
npm config delete init.module
npm config delete --global _-init-module
npm config delete --global init.module
```

#### 方案3: 自动处理
`deploy-pnpm.js` 脚本已包含自动清理逻辑，直接运行即可：
```bash
node deploy-pnpm.js
```

---

## 问题2: /setup页面无法保存密码

### 症状
- 访问首页没有自动跳转到/setup
- 在/setup页面输入密码后无法保存
- 提示"保存失败，请重试"

### 原因
1. 环境变量未正确设置
2. .env文件权限问题
3. 路径解析问题

### 解决方案

#### 1. 检查.env文件
确保根目录和client目录下都有.env文件：

**根目录 .env**:
```bash
NODE_ENV=production
PORT=4001
REACT_APP_ADMIN=your_admin_password_here
ALLOWED_ORIGINS=http://localhost:4000,http://127.0.0.1:4000
```

**client/.env.production**:
```bash
REACT_APP_ADMIN=your_admin_password_here
PORT=4000
BROWSER=none
SKIP_PREFLIGHT_CHECK=true
DISABLE_ESLINT_PLUGIN=true
```

#### 2. 设置文件权限
```bash
chmod 644 .env
chmod 644 client/.env.production
```

#### 3. 初始化设置流程

##### 步骤1: 访问setup页面
```bash
# 启动服务后，浏览器访问
http://localhost:4000/setup
```

##### 步骤2: 设置密码
- 输入管理员密码（至少6位）
- 确认密码
- 点击"完成设置"

##### 步骤3: 重启服务 ⚠️ **重要**
```bash
# 停止当前服务
Ctrl+C

# 重新启动
node start.js
# 或
pnpm run deploy
```

**说明**: 环境变量在Node.js进程启动时加载，保存密码后必须重启服务才能生效。

#### 4. 验证设置
```bash
# 检查.env文件
cat .env | grep REACT_APP_ADMIN
cat client/.env.production | grep REACT_APP_ADMIN

# 应该显示你设置的密码，而不是 your_admin_password_here
```

---

## 问题3: 首页不自动跳转到/setup

### 症状
首次访问首页没有自动跳转到/setup页面

### 原因
1. .env文件中已有密码配置（即使是占位符）
2. API检查逻辑未正确识别

### 解决方案

#### 方案1: 手动访问
直接访问setup页面：
```
http://localhost:4000/setup
```

#### 方案2: 清空密码配置
```bash
# 编辑.env文件
sed -i 's/REACT_APP_ADMIN=.*/REACT_APP_ADMIN=/' .env
sed -i 's/REACT_APP_ADMIN=.*/REACT_APP_ADMIN=/' client/.env.production

# 重启服务
node start.js
```

#### 方案3: 删除.env文件重新初始化
```bash
rm .env client/.env.production
node start.js
# 然后访问 http://localhost:4000/setup
```

---

## 完整部署流程（Linux）

### 1. 清理环境
```bash
# 清理npm配置
./clean-npm-config.sh

# 或手动清理
npm config delete _-init-module
npm config delete init.module
```

### 2. 安装依赖
```bash
pnpm install
```

### 3. 构建应用
```bash
pnpm run build
```

### 4. 初始化设置
```bash
# 启动服务
node start.js

# 浏览器访问 http://localhost:4000/setup
# 设置管理员密码

# 停止服务（Ctrl+C）
```

### 5. 重启服务
```bash
# 重新启动以加载新密码
node start.js
```

### 6. 验证
```bash
# 访问管理面板
http://localhost:4000/admin

# 使用刚才设置的密码登录
```

---

## 常用命令速查

```bash
# 一键部署（包含自动清理npm配置）
node deploy-pnpm.js

# 开发模式启动
node start.js

# 生产模式启动
cd server && node dist/index.js &
npx serve -s client/build -l 4000

# 查看日志
tail -f server/logs/*.log

# 停止服务
pkill -f "node dist/index.js"
pkill -f "serve"

# 检查端口占用
lsof -i :4000
lsof -i :4001
```

---

## 权限问题

### 文件权限
```bash
# 设置正确的文件权限
chmod 644 .env
chmod 644 client/.env.production
chmod 644 config.json
chmod +x start.js
chmod +x deploy-pnpm.js
chmod +x clean-npm-config.sh
```

### 目录权限
```bash
# 确保server和client目录可写
chmod 755 server
chmod 755 client
chmod 755 server/dist
chmod 755 client/build
```

---

## 环境变量加载顺序

系统按以下顺序查找配置：

1. `项目根目录/.env`
2. `client/.env.production`
3. `process.env` (系统环境变量)

**注意**: Node.js只在启动时加载环境变量，运行时修改不生效。

---

## 故障排查

### 1. 检查服务状态
```bash
ps aux | grep node
netstat -tulpn | grep :4001
netstat -tulpn | grep :4000
```

### 2. 查看日志
```bash
# 后端日志
journalctl -u chemistryuno-server -f

# 或直接查看进程输出
node server/dist/index.js
```

### 3. 测试API
```bash
# 检查setup状态
curl http://localhost:4001/api/setup/check

# 应返回: {"isSetup":false} 或 {"isSetup":true}
```

### 4. 验证环境变量
```bash
# 在服务器进程中打印
node -e "require('dotenv').config(); console.log(process.env.REACT_APP_ADMIN)"
```

---

## 安全建议

1. **使用强密码**: 至少12位，包含大小写字母、数字、特殊字符
2. **限制访问**: 使用防火墙限制4000/4001端口的访问
3. **HTTPS**: 生产环境使用nginx反向代理并启用HTTPS
4. **定期更新**: 定期运行 `pnpm audit` 检查安全漏洞

---

**最后更新**: 2026年1月4日
