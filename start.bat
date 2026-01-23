@echo off
chcp 65001 >nul
echo.
echo ═══════════════════════════════════════════════════════════
echo   Chemistry UNO - 化学UNO 快速启动
echo ═══════════════════════════════════════════════════════════
echo.

REM 检查 Node.js
where node >nul 2>nul
if %errorlevel% neq 0 (
    echo [✗] 请先安装 Node.js: https://nodejs.org/
    pause
    exit /b 1
)

REM 检查 pnpm
where pnpm >nul 2>nul
if %errorlevel% neq 0 (
    echo [✗] 请先安装 pnpm: npm install -g pnpm
    pause
    exit /b 1
)

REM 检查依赖是否已安装
if not exist "node_modules" (
    echo [→] 首次运行，正在安装依赖...
    call pnpm install
    if %errorlevel% neq 0 (
        echo [✗] 依赖安装失败
        pause
        exit /b 1
    )
    echo [✓] 依赖安装完成
    echo.
)

echo [✓] 正在启动开发服务器...
echo.
echo   前端: http://localhost:3000
echo   后端: http://localhost:5000
echo   管理面板: http://localhost:3000/admin
echo.
echo 按 Ctrl+C 停止服务
echo.

REM 启动开发服务器
call pnpm run dev
