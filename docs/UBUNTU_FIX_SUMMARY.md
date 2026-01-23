# Ubuntu 22 部署问题修复总结

## 问题概述

在 Ubuntu 22 环境下部署化学UNO项目时遇到以下三个主要问题：

### 问题表现
1. ❌ **路由问题**: 访问 `localhost:4000` 无法跳转到 `/setup` 页面，显示 404 错误
2. ❌ **连接问题**: 无法连接后端服务器
3. ❌ **网络错误**: 创建房间时提示 "network error"

## 根本原因分析

### 1. SPA 路由配置缺失
**问题**: 生产环境使用 `serve` 命令部署时，缺少单页应用(SPA)路由配置，导致所有非根路径请求返回 404。

**技术原因**: 
- React Router 使用客户端路由
- 服务器需要配置所有路由重定向到 `index.html`
- `serve` 默认不提供这种配置

**影响**: 
- 直接访问 `/setup`、`/admin` 等路由失败
- 刷新页面后丢失当前路由

### 2. 后端监听地址配置错误
**问题**: 后端服务器只监听 `localhost` (127.0.0.1)，不监听所有网络接口。

**技术原因**:
```typescript
// 错误配置
server.listen(PORT, () => { ... });
// 默认只监听 localhost

// 正确配置
server.listen(PORT, '0.0.0.0', () => { ... });
// 监听所有网络接口
```

**影响**:
- 外部设备无法访问后端
- 某些 Linux 环境下前端也无法连接

### 3. 部署脚本配置不完整
**问题**: 部署脚本未包含 `serve.json` 配置文件参数。

**影响**: 即使创建了配置文件，`serve` 命令也不会使用

## 修复方案

### 修复 1: 添加 serve.json 配置

**文件位置**: 
- `client/public/serve.json`
- `client/build/serve.json`

**配置内容**:
```json
{
  "rewrites": [
    { "source": "**", "destination": "/index.html" }
  ],
  "headers": [
    {
      "source": "**",
      "headers": [
        {
          "key": "Cache-Control",
          "value": "no-cache"
        }
      ]
    }
  ]
}
```

**作用**:
- 所有路由请求重定向到 `index.html`
- 禁用缓存确保获取最新内容
- 支持前端路由（/setup、/admin 等）

### 修复 2: 更新后端监听配置

**文件**: `server/index.ts`

**修改内容**:
```typescript
const PORT = process.env.PORT || 4001;
const HOST = process.env.HOST || '0.0.0.0';

server.listen(PORT, HOST, () => {
  console.log(`✓ 服务器运行在 http://localhost:${PORT}`);
  console.log(`✓ 服务器监听在 ${HOST}:${PORT}`);
  console.log(`✓ WebSocket 服务已启动，等待连接...`);
  
  // 显示局域网访问地址
  const os = require('os');
  const interfaces = os.networkInterfaces();
  const addresses: string[] = [];
  
  for (const name of Object.keys(interfaces)) {
    for (const iface of interfaces[name]!) {
      if (iface.family === 'IPv4' && !iface.internal) {
        addresses.push(iface.address);
      }
    }
  }
  
  if (addresses.length > 0) {
    console.log(`✓ 局域网访问地址:`);
    addresses.forEach(addr => {
      console.log(`  - http://${addr}:${PORT}`);
    });
  }
});
```

**作用**:
- 监听所有网络接口 (0.0.0.0)
- 支持局域网访问
- 显示所有可用访问地址

### 修复 3: 更新部署脚本

**文件**: `deploy-pnpm.js`

**修改内容**:
```javascript
const serveArgs = ['serve', '-s', 'build', '-l', '4000', '--config', 'build/serve.json'];
const frontendProcess = spawn(isWindows ? 'npx.cmd' : 'npx', serveArgs, {
  cwd: path.join(process.cwd(), 'client'),
  stdio: 'inherit',
  shell: true
});
```

**作用**: 
- 使用 `--config` 参数指定配置文件
- 确保 `serve` 使用正确的路由配置

### 修复 4: 更新启动脚本端口说明

**文件**: `start.sh`

**修改**: 端口号从 3000/5000 改为 4000/4001，与实际配置一致。

## 修复验证

### 自动化脚本

创建了三个脚本帮助用户快速修复和验证：

1. **apply-fix.sh** - 快速应用修复
   ```bash
   chmod +x apply-fix.sh
   ./apply-fix.sh
   ```

2. **verify-fix.sh** - 验证修复是否成功
   ```bash
   chmod +x verify-fix.sh
   ./verify-fix.sh
   ```

3. **test-all.sh** - 完整功能测试
   ```bash
   chmod +x test-all.sh
   ./test-all.sh
   ```

### 手动验证步骤

1. **验证 serve.json 存在**:
   ```bash
   ls -la client/public/serve.json
   ls -la client/build/serve.json
   ```

2. **验证后端监听地址**:
   ```bash
   # 启动后端后检查
   netstat -tlnp | grep 4001
   # 应该显示: 0.0.0.0:4001
   ```

3. **测试前端路由**:
   ```bash
   curl http://localhost:4000/
   curl http://localhost:4000/setup
   curl http://localhost:4000/admin
   # 所有请求都应该返回 HTML 内容
   ```

4. **测试后端 API**:
   ```bash
   curl http://localhost:4001/api/setup/check
   # 应该返回 JSON 响应
   ```

5. **测试创建房间**:
   ```bash
   curl -X POST http://localhost:4001/api/game/create \
     -H "Content-Type: application/json" \
     -d '{"hostName":"Test","maxPlayers":4,"initialHandSize":10}'
   # 应该返回房间号
   ```

## 部署流程

### 开发环境

```bash
# 1. 安装依赖
pnpm install

# 2. 启动开发服务器
pnpm run dev
# 或
./start.sh

# 访问地址
# 前端: http://localhost:3000
# 后端: http://localhost:4001
```

### 生产环境

```bash
# 1. 应用修复（如需要）
chmod +x apply-fix.sh
./apply-fix.sh

# 2. 构建项目
pnpm run build

# 3. 部署
pnpm run deploy
# 或
node deploy-pnpm.js

# 4. 验证
chmod +x verify-fix.sh
./verify-fix.sh

# 5. 完整测试
chmod +x test-all.sh
./test-all.sh

# 访问地址
# 前端: http://localhost:4000
# 后端: http://localhost:4001
```

## 技术要点

### SPA 路由原理

单页应用(SPA)的路由是在客户端处理的：
1. 用户访问 `/setup`
2. 服务器需要返回 `index.html`
3. React Router 在客户端解析 `/setup` 并渲染对应组件

如果服务器直接返回 404，React Router 无法接管路由。

### 0.0.0.0 vs 127.0.0.1

- **127.0.0.1** (localhost): 只能本机访问
- **0.0.0.0**: 监听所有网络接口
  - 包括 127.0.0.1 (本机)
  - 包括局域网 IP (如 192.168.1.100)
  - 支持外部设备访问

### CORS 配置

后端已配置允许所有来源（开发环境）：
```typescript
app.use(cors({
  origin: (origin, callback) => {
    callback(null, true);
  },
  credentials: true
}));
```

生产环境建议限制具体来源。

## 文档结构

创建了完整的文档体系：

1. **[UBUNTU_FIX.md](UBUNTU_FIX.md)** - 详细修复指南
   - 问题诊断
   - 修复步骤
   - 配置说明
   - 优化建议

2. **[TROUBLESHOOTING_UBUNTU.md](TROUBLESHOOTING_UBUNTU.md)** - 问题排查清单
   - 8大常见问题
   - 详细解决方案
   - 快速修复命令

3. **修复脚本**:
   - `apply-fix.sh` - 应用修复
   - `verify-fix.sh` - 验证修复
   - `test-all.sh` - 功能测试

4. **更新主文档**:
   - README.md - 添加 Ubuntu 注意事项
   - 文档目录.md - 添加新文档索引

## 预防措施

### 1. 开发环境配置
确保 `client/.env` 正确：
```bash
PORT=4000
BROWSER=none
SKIP_PREFLIGHT_CHECK=true
```

### 2. 生产环境配置
创建 `client/.env.production`：
```bash
REACT_APP_ADMIN=your_password_here
PORT=4000
SKIP_PREFLIGHT_CHECK=true
```

### 3. 构建后检查
每次构建后验证：
```bash
# 检查 serve.json 是否被复制
ls -la client/build/serve.json

# 检查构建文件
ls -la client/build/
```

### 4. 使用 PM2 管理进程
生产环境推荐使用 PM2：
```bash
npm install -g pm2

# 后端
cd server
pm2 start dist/index.js --name "chemistry-uno-backend"

# 前端
cd ../client
pm2 start "npx serve -s build -l 4000 --config build/serve.json" \
  --name "chemistry-uno-frontend"

# 保存配置
pm2 save
pm2 startup
```

### 5. 使用 Nginx 反向代理
生产环境推荐配置 Nginx，统一管理前后端：
```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:4000;
        # ... proxy 配置
    }

    location /api {
        proxy_pass http://localhost:4001;
        # ... proxy 配置
    }

    location /socket.io {
        proxy_pass http://localhost:4001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

## 测试结果

修复后所有功能正常：
- ✅ 前端路由正常工作（/, /setup, /admin）
- ✅ 后端 API 正常响应
- ✅ WebSocket 连接成功
- ✅ 创建房间功能正常
- ✅ 加入房间功能正常
- ✅ 局域网访问正常

## 总结

本次修复解决了 Ubuntu 22 部署时的三大核心问题：
1. ✅ SPA 路由配置
2. ✅ 后端监听地址
3. ✅ 部署脚本配置

同时提供了：
- 📝 完整文档（修复指南、问题排查）
- 🔧 自动化脚本（修复、验证、测试）
- 📊 最佳实践（PM2、Nginx）

用户现在可以：
- 快速修复问题（1个命令）
- 验证修复效果（自动化测试）
- 正常使用所有功能
- 支持局域网访问

## 相关资源

- [Ubuntu 修复指南](UBUNTU_FIX.md)
- [问题排查清单](TROUBLESHOOTING_UBUNTU.md)
- [跨平台支持](PLATFORM_SUPPORT.md)
- [主文档](../README.md)

---

修复完成时间: 2026-01-09
修复版本: v2.0.1
