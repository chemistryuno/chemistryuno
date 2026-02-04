@echo off
chcp 65001 >nul
title Chemistry UNO Mendeleef - 初始化脚本

echo.
echo 🧪 ============================================
echo 🧪 Chemistry UNO Mendeleef - 项目初始化脚本
echo 🧪 ============================================
echo.

:: 检查 Node.js
where node >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Node.js 未安装。请从 https://nodejs.org/ 下载并安装。
    pause
    exit /b 1
)
echo ✅ Node.js 已安装

:: 检查 Go
where go >nul 2>&1  
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Go 未安装。请从 https://golang.org/dl/ 下载并安装。
    pause
    exit /b 1
)
echo ✅ Go 已安装

:: 检查并安装 pnpm
where pnpm >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ⚠️  pnpm 未安装，正在安装...
    npm install -g pnpm
    if %ERRORLEVEL% NEQ 0 (
        echo ❌ pnpm 安装失败
        pause
        exit /b 1
    )
    echo ✅ pnpm 安装成功
) else (
    echo ✅ pnpm 已安装
)

:: 安装前端依赖
echo.
echo 🎨 安装前端依赖...
cd frontend
if not exist package.json (
    echo ❌ 前端 package.json 不存在
    pause
    exit /b 1
)
pnpm install
if %ERRORLEVEL% NEQ 0 (
    echo ❌ 前端依赖安装失败
    pause
    exit /b 1
)
echo ✅ 前端依赖安装成功
cd ..

:: 安装后端依赖
echo.
echo 🏗️  安装后端依赖...
cd backend
if not exist go.mod (
    echo ❌ 后端 go.mod 不存在
    pause
    exit /b 1
)
go mod tidy
if %ERRORLEVEL% NEQ 0 (
    echo ❌ 后端依赖安装失败
    pause
    exit /b 1
)
echo ✅ 后端依赖安装成功
cd ..

:: 创建配置文件
echo.
echo ⚙️  创建配置文件...
if not exist backend\.env (
    echo # Chemistry UNO Mendeleef 配置文件> backend\.env
    echo # 后端配置>> backend\.env
    echo PORT=8080>> backend\.env
    echo JWT_SECRET=chemistry-uno-secret-key-change-in-production>> backend\.env
    echo SQLITE_PATH=./chemistryuno.db>> backend\.env
    echo.>> backend\.env
    echo # 前端配置>> backend\.env
    echo VITE_API_URL=http://localhost:8080>> backend\.env
    echo VITE_WS_URL=ws://localhost:8080/ws>> backend\.env
    echo.>> backend\.env
    echo # 开发环境配置>> backend\.env
    echo NODE_ENV=development>> backend\.env
    echo GIN_MODE=debug>> backend\.env
    echo ✅ 创建 .env 配置文件
) else (
    echo ℹ️  .env 配置文件已存在
)

:: 显示完成信息
echo.
echo 🎉 初始化完成！
echo.
echo 📋 启动说明:
echo   🚀 启动完整项目:       npm start
echo   🎨 仅启动前端:         npm run frontend  
echo   🏗️  仅启动后端:         npm run backend
echo.
echo 🌐 访问地址:
echo   前端: http://localhost:5000
echo   后端: http://localhost:8080
echo.
echo 👥 默认管理员账户:
echo   用户名: admin
echo   密码: admin123
echo.
echo 🛠️  开发工具:
echo   🔧 热重载:           自动检测文件变化
echo   📊 数据库管理:       内置SQLite数据库
echo   🔐 用户权限系统:     admin/co-worker/user三级权限
echo   🧪 化学反应库:       支持自定义化学反应数据
echo.
echo ===================================
echo.
pause