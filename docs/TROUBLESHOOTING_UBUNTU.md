# Ubuntu 22 常见问题排查清单

## 问题检查流程

### 第一步：确认问题症状

请勾选你遇到的问题：
- [ ] 访问 `localhost:4000` 显示空白页面
- [ ] 访问 `localhost:4000/setup` 显示 404 错误
- [ ] 浏览器控制台显示 CORS 错误
- [ ] 浏览器控制台显示 "Failed to fetch" 或 "Network error"
- [ ] 创建房间时提示 "network error"
- [ ] 无法连接 WebSocket

### 第二步：运行诊断脚本

```bash
# 给脚本执行权限
chmod +x verify-fix.sh

# 运行验证
./verify-fix.sh
```

### 第三步：根据诊断结果修复

## 问题 1: 404 错误 - 路由不工作

**症状：**
- 访问 `/setup` 或其他路由返回 404
- 刷新页面后丢失当前页面

**原因：**
单页应用(SPA)需要配置所有路由重定向到 index.html

**解决方案：**

1. 检查是否存在 `serve.json`：
```bash
ls -la client/build/serve.json
ls -la client/public/serve.json
```

2. 如果不存在，创建文件：
```bash
# 确保目录存在
mkdir -p client/build client/public

# 创建配置文件
cat > client/public/serve.json << 'EOF'
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
EOF

# 复制到 build 目录
cp client/public/serve.json client/build/serve.json
```

3. 重新启动前端服务：
```bash
# 停止当前服务 (Ctrl+C)
# 重新启动
cd client
npx serve -s build -l 4000 --config build/serve.json
```

## 问题 2: 无法连接后端 - CORS 或连接被拒绝

**症状：**
- 浏览器控制台显示 CORS 错误
- 显示 "Failed to connect to localhost:4001"
- Network error

**诊断步骤：**

1. **检查后端是否运行：**
```bash
# 测试后端连接
curl http://localhost:4001/

# 应该返回 HTML 页面，表示后端正在运行
```

如果返回 "Connection refused"：
```bash
# 后端未运行，启动它
cd server
pnpm run dev
# 或生产模式
node dist/index.js
```

2. **检查后端监听地址：**
```bash
# 查看后端监听在哪个地址
netstat -tlnp | grep 4001
# 或
ss -tlnp | grep 4001
```

应该显示：
```
tcp   0   0 0.0.0.0:4001   0.0.0.0:*   LISTEN
```

如果显示 `127.0.0.1:4001`，则需要修改后端配置。

3. **修复后端监听地址：**

编辑 `server/index.ts`，找到最后的 `server.listen` 部分：

```typescript
// 修改前（只监听 localhost）
const PORT = process.env.PORT || 4001;
server.listen(PORT, () => {
  console.log(`✓ 服务器运行在 http://localhost:${PORT}`);
});

// 修改后（监听所有接口）
const PORT = process.env.PORT || 4001;
const HOST = process.env.HOST || '0.0.0.0';

server.listen(PORT, HOST, () => {
  console.log(`✓ 服务器运行在 http://localhost:${PORT}`);
  console.log(`✓ 服务器监听在 ${HOST}:${PORT}`);
});
```

4. **重新编译并启动后端：**
```bash
cd server
pnpm run build
node dist/index.js
```

## 问题 3: 端口被占用

**症状：**
- 启动服务时显示 "port already in use"
- 无法启动前端或后端

**解决方案：**

1. **查找占用端口的进程：**
```bash
# 查找端口 4000
sudo lsof -i :4000

# 查找端口 4001
sudo lsof -i :4001
```

2. **结束占用进程：**
```bash
# 方法 1: 使用 PID
sudo kill -9 <PID>

# 方法 2: 使用 fuser
sudo fuser -k 4000/tcp
sudo fuser -k 4001/tcp
```

3. **更改端口（可选）：**

修改前端端口 (client/.env)：
```bash
PORT=4002
```

修改后端端口 (server/index.ts)：
```typescript
const PORT = process.env.PORT || 4003;
```

同时更新前端 API 配置 (client/src/config/api.ts)：
```typescript
return 'http://localhost:4003';
```

## 问题 4: 防火墙阻止连接

**症状：**
- 本机访问正常，其他设备无法访问
- 局域网设备连接失败

**解决方案：**

1. **检查防火墙状态：**
```bash
sudo ufw status
```

2. **开放端口：**
```bash
# 开放前端端口
sudo ufw allow 4000/tcp

# 开放后端端口
sudo ufw allow 4001/tcp

# 重新加载防火墙
sudo ufw reload
```

3. **或暂时禁用防火墙（测试用）：**
```bash
sudo ufw disable
```

记得测试完后重新启用：
```bash
sudo ufw enable
```

## 问题 5: 浏览器缓存问题

**症状：**
- 修改代码后页面没有更新
- 显示旧版本的页面

**解决方案：**

1. **清除浏览器缓存：**
   - Chrome/Edge: `Ctrl + Shift + Delete`
   - Firefox: `Ctrl + Shift + Delete`

2. **硬刷新：**
   - `Ctrl + Shift + R` (Linux)
   - `Ctrl + F5`

3. **使用无痕模式测试：**
   - Chrome/Edge: `Ctrl + Shift + N`
   - Firefox: `Ctrl + Shift + P`

4. **清除构建缓存并重新构建：**
```bash
# 清除前端构建
rm -rf client/build
rm -rf client/node_modules/.cache

# 清除后端构建
rm -rf server/dist

# 重新构建
pnpm run build
```

## 问题 6: 依赖安装失败

**症状：**
- `pnpm install` 报错
- 缺少某些包

**解决方案：**

1. **清除 pnpm 缓存：**
```bash
pnpm store prune
```

2. **删除 lock 文件重新安装：**
```bash
rm pnpm-lock.yaml
rm -rf node_modules
rm -rf client/node_modules
rm -rf server/node_modules

pnpm install
```

3. **更新 pnpm：**
```bash
npm install -g pnpm@latest
```

4. **检查 Node.js 版本：**
```bash
node --version  # 应该 >= 18.0.0
```

如果版本过低：
```bash
# 使用 nvm 安装新版本
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
source ~/.bashrc
nvm install 18
nvm use 18
```

## 问题 7: WebSocket 连接失败

**症状：**
- 浏览器控制台显示 WebSocket 错误
- 实时功能不工作（如玩家加入通知）

**诊断：**

1. **检查 WebSocket 连接：**
打开浏览器开发者工具 -> Network -> WS (WebSocket)

2. **检查 Socket.IO 配置：**
确认 `client/src/config/api.ts` 中的 `SOCKET_URL` 正确：
```typescript
export const SOCKET_URL = API_BASE_URL; // 应该是 http://localhost:4001
```

3. **检查后端 CORS 配置：**
`server/index.ts` 中应该包含：
```typescript
const io = new SocketIOServer(server, {
  cors: {
    origin: (origin, callback) => {
      callback(null, true); // 允许所有来源（开发环境）
    },
    methods: ["GET", "POST"],
    credentials: true
  }
});
```

## 问题 8: 环境变量未生效

**症状：**
- 管理员密码不工作
- 端口配置未生效

**解决方案：**

1. **检查 .env 文件：**
```bash
# 开发环境
cat client/.env

# 生产环境
cat client/.env.production
```

2. **确保文件格式正确：**
```bash
# client/.env
PORT=4000
BROWSER=none
REACT_APP_ADMIN=your_password
SKIP_PREFLIGHT_CHECK=true
```

3. **重新构建：**
环境变量在构建时被嵌入，修改后需要重新构建：
```bash
pnpm run build
```

## 快速修复命令汇总

```bash
# 1. 应用所有修复
chmod +x apply-fix.sh && ./apply-fix.sh

# 2. 清理并重新安装
rm -rf node_modules client/node_modules server/node_modules
rm -rf client/build server/dist
pnpm install

# 3. 重新构建
pnpm run build

# 4. 启动服务
pnpm run deploy

# 5. 验证修复
chmod +x verify-fix.sh && ./verify-fix.sh
```

## 仍然无法解决？

1. **收集日志信息：**
```bash
# 后端日志
cd server
pnpm run dev 2>&1 | tee backend.log

# 前端日志
cd client
npx serve -s build -l 4000 2>&1 | tee frontend.log
```

2. **检查系统日志：**
```bash
# 查看系统日志
journalctl -xe

# 查看 nginx 日志（如果使用）
sudo tail -f /var/log/nginx/error.log
```

3. **浏览器控制台：**
   - 打开开发者工具 (F12)
   - 查看 Console 标签的错误信息
   - 查看 Network 标签的请求失败详情

4. **提供详细信息：**
   - 操作系统版本：`lsb_release -a`
   - Node.js 版本：`node --version`
   - pnpm 版本：`pnpm --version`
   - 完整的错误信息
   - 浏览器类型和版本

## 相关文档

- [Ubuntu 22 修复指南](UBUNTU_FIX.md) - 详细修复说明
- [快速入门指南](GETTING_STARTED.md) - 基本使用说明
- [部署测试报告](DEPLOY_TEST_REPORT.md) - 部署测试记录
