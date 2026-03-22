# Chemistry UNO - 命令速查表

## 📦 安装与初始化

```bash
# 首次使用：安装所有依赖并初始化项目
pnpm init
```

## 🚀 启动命令

```bash
# 开发环境 - 启动前后端（热重载）
pnpm start
# 或
pnpm dev

# 测试环境 - 使用测试数据库启动
pnpm test

# 测试环境（编译模式）- 更接近生产环境
pnpm test:build
```

## 🏗️ 构建命令

```bash
# 一键构建前后端（生产版本）
pnpm build

# 仅构建后端
pnpm build:backend

# 仅构建前端
pnpm build:frontend
```

构建产物位置：
- 完整构建：`dist/` 目录
- 后端二进制：根目录 `chemistryuno.exe` (Windows) 或 `chemistryuno` (Linux/macOS)
- 前端静态文件：`frontend/dist/`

## 🧪 测试命令

```bash
# 启动测试环境（开发模式，快速启动）
pnpm test

# 启动测试环境（编译后运行，更接近生产）
pnpm test:build

# 仅运行后端 Go 测试
pnpm go:test
```

测试环境特性：
- ✅ 自动初始化测试数据库
- ✅ 包含示例用户和数据
- ✅ 独立于开发数据库
- ✅ 前后端同时启动

## 🗄️ 数据库命令

```bash
# 初始化生产数据库
pnpm db:init

# 初始化测试数据库
pnpm db:init-test
```

## 🎨 前端命令

```bash
# 安装前端依赖
pnpm install:frontend

# 仅启动前端开发服务器
pnpm frontend:dev

# 构建并部署前端（嵌入后端）
pnpm frontend
```

## 🔧 后端命令

```bash
# 仅启动后端服务器（开发模式）
pnpm backend
```

## 🧹 清理命令

```bash
# 清理构建产物和数据库
pnpm clean

# 完全清理（包括 node_modules）
pnpm clean:all
```

## 📋 常用工作流

### 开发流程

```bash
# 1. 首次克隆项目后
pnpm init

# 2. 日常开发
pnpm dev

# 3. 构建测试
pnpm build
```

### 测试流程

```bash
# 1. 启动测试环境
pnpm test

# 2. 在浏览器测试功能
# http://localhost:5000

# 3. 运行 Go 单元测试
pnpm go:test
```

### 部署流程

```bash
# 1. 清理旧构建
pnpm clean

# 2. 构建生产版本
pnpm build

# 3. 部署 dist/ 目录内容到服务器
cd dist
./start.sh  # Linux/macOS
# 或
start.bat   # Windows
```

## 🌐 访问地址

**开发环境 (pnpm start)**
- 前端：http://localhost:5000
- 后端：http://localhost:8080

**测试环境 (pnpm test)**
- 前端：http://localhost:5000
- 后端：http://localhost:8080
- 测试账户：
  - 管理员：`admin@chemistryuno.com` / `admin123`
  - 普通用户：`test@example.com` / `test123`

**生产环境 (构建后)**
- 访问地址：http://localhost:8080（前端嵌入后端）

## 💡 提示

- 所有命令都支持在项目根目录运行
- `pnpm test` 使用独立的测试数据库，不会影响开发数据
- `pnpm build` 会生成完整的部署包到 `dist/` 目录
- Windows 用户可直接运行所有命令，路径会自动适配
- 按 `Ctrl+C` 停止任何运行中的服务

## 🔗 相关文件

- `start.js` - 开发环境启动脚本
- `test.js` - 测试环境启动脚本
- `build.js` - 构建脚本
- `init.js` - 初始化脚本
- `.env` - 环境变量配置文件（统一存放在根目录）
- `.env.example` - 环境变量配置示例

**注意**：所有配置统一从根目录 `.env` 文件读取，无需在 `backend/` 目录创建单独的配置文件。

