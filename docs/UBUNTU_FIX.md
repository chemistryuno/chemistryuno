# Linux Ubuntu 22 部署修复说明

## 问题诊断

在 Ubuntu 22 环境下出现的问题：
1. **无法跳转 /setup 页面** - 原因：单页应用(SPA)路由配置缺失
2. **无法连接后端服务器** - 原因：后端只监听 localhost，未监听所有网络接口
3. **创建房间提示 network error** - 原因：前后端连接问题

## 已修复内容

### 1. 添加 SPA 路由支持

创建了 `client/public/serve.json` 和 `client/build/serve.json`：
- 配置所有路由重写到 index.html
- 支持 /setup、/admin 等前端路由
- 禁用缓存确保始终获取最新内容

### 2. 修复后端服务器监听地址

修改 `server/index.ts`：
- 从只监听 localhost 改为监听所有网络接口 (0.0.0.0)
- 添加局域网访问地址显示
- 支持外部设备访问

### 3. 更新启动脚本

- 修正端口号显示（前端 4000，后端 4001）
- 更新部署脚本以包含 serve.json 配置

## 部署步骤

### 开发环境启动

```bash
# 1. 确保已安装依赖
pnpm install

# 2. 启动开发服务器（前后端同时启动）
./start.sh
# 或使用 pnpm
pnpm run dev
```

访问地址：
- 前端：http://localhost:3000
- 后端：http://localhost:4001

### 生产环境部署

```bash
# 1. 构建和部署
pnpm run deploy
# 或
node deploy-pnpm.js
```

生产环境访问地址：
- 前端：http://localhost:4000
- 后端：http://localhost:4001
- 管理面板：http://localhost:4000/admin

### 手动启动（如需分别启动）

```bash
# 终端 1 - 启动后端
cd server
pnpm install
pnpm run dev    # 开发模式
# 或
pnpm run build && node dist/index.js  # 生产模式

# 终端 2 - 启动前端
cd client
pnpm install

# 开发模式
pnpm start

# 生产模式 - 先构建
pnpm run build
# 然后使用 serve 启动 (serve 会自动加载 build/serve.json)
npx serve -s build -l 4000
```

## 关键配置说明

### serve.json 配置
```json
{
  "rewrites": [
    { "source": "**", "destination": "/index.html" }
  ]
}
```
这个配置确保所有路由请求都返回 index.html，让 React Router 处理路由。

### 后端监听地址
```typescript
const HOST = '0.0.0.0';  // 监听所有网络接口
server.listen(PORT, HOST, () => { ... });
```

### API 配置
前端 API 配置 (`client/src/config/api.ts`) 自动处理：
- localhost 访问：使用 http://localhost:4001
- 局域网访问：使用实际服务器 IP:4001

## 验证步骤

### 1. 检查后端是否正常运行
```bash
# 应该看到服务器启动信息
curl http://localhost:4001/

# 测试 API 端点
curl http://localhost:4001/api/setup/check
```

### 2. 检查前端是否能访问后端
打开浏览器开发者工具 (F12) -> Network 标签，访问页面查看请求状态。

### 3. 测试路由
访问以下地址应该都能正常加载：
- http://localhost:4000/
- http://localhost:4000/setup
- http://localhost:4000/admin

## 常见问题排查

### 问题1: 端口被占用
```bash
# 查找占用端口的进程
sudo lsof -i :4000
sudo lsof -i :4001

# 结束进程
sudo kill -9 <PID>
```

### 问题2: 防火墙阻止
```bash
# 开放端口
sudo ufw allow 4000
sudo ufw allow 4001
```

### 问题3: Node.js 版本问题
```bash
# 检查版本（需要 >= 18.0.0）
node --version

# 使用 nvm 切换版本
nvm install 18
nvm use 18
```

### 问题4: 权限问题
```bash
# 给启动脚本执行权限
chmod +x start.sh
chmod +x deploy.sh
```

### 问题4: Mixed Content / Network Error (HTTPS 环境)
**症状：**
- 控制台显示 "Mixed Content... was loaded over HTTPS, but requested an insecure XMLHttpRequest..."
- 无法连接后端或 WebSocket

**解决方案：**
如果您的网站使用了 HTTPS，必须配置 Nginx 反向代理将 `/api` 和 `/socket.io` 转发到后端端口 (4001)。

1. 查看详细配置指南：**[docs/HTTPS_NGINX_GUIDE.md](HTTPS_NGINX_GUIDE.md)**
2. 修改 Nginx 配置后，重新构建并部署前端：
```bash
pnpm run build
pnpm run deploy
```

### 问题5: CORS 错误
后端已配置支持所有来源，如仍有问题：
1. 确认后端正在运行
2. 检查浏览器控制台错误信息
3. 确认 API_BASE_URL 配置正确

## 局域网访问

后端现在会显示所有可用的局域网访问地址，例如：
```
✓ 服务器监听在 0.0.0.0:4001
✓ 局域网访问地址:
  - http://192.168.1.100:4001
```

其他设备可以通过这些地址访问：
- 前端：http://192.168.1.100:4000
- 后端：http://192.168.1.100:4001

## 环境变量配置（可选）

创建 `client/.env.production`：
```bash
# 管理员密码
REACT_APP_ADMIN=your_password_here

# 前端端口
PORT=4000

# 跳过预检检查
SKIP_PREFLIGHT_CHECK=true
```

## 日志查看

查看服务器日志：
```bash
# 如果使用 systemd 服务
sudo journalctl -u chemistryuno -f

# 或者直接查看控制台输出
```

## 重新构建

如果修改了代码，需要重新构建：
```bash
# 清理旧构建
rm -rf client/build server/dist

# 重新构建
pnpm run build

# 重新部署
pnpm run deploy
```

## 性能优化建议

1. **使用 PM2 管理进程**：
```bash
npm install -g pm2

# 启动后端
cd server
pm2 start dist/index.js --name "chemistry-uno-backend"

# 启动前端 serve
cd ../client
pm2 start "npx serve -s build -l 4000 --config build/serve.json" --name "chemistry-uno-frontend"

# 保存配置并设置开机自启
pm2 save
pm2 startup
```

2. **使用 Nginx 反向代理**（推荐生产环境）：
```nginx
server {
    listen 80;
    server_name your-domain.com;

    # 前端
    location / {
        proxy_pass http://localhost:4000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    # 后端 API
    location /api {
        proxy_pass http://localhost:4001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    # WebSocket
    location /socket.io {
        proxy_pass http://localhost:4001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

## 总结

所有修复已完成，主要改动：
1. ✅ 添加 serve.json 配置文件支持 SPA 路由
2. ✅ 修改后端监听地址为 0.0.0.0
3. ✅ 更新启动脚本的端口信息
4. ✅ 更新部署脚本包含正确配置

现在应该可以正常：
- 访问 /setup 页面
- 连接后端服务器
- 创建和加入房间

如有其他问题，请检查浏览器控制台和服务器日志获取详细错误信息。
