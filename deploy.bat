@echo off
chcp 65001 >nul
echo.
echo ═══════════════════════════════════════════════════════════
echo   Chemistry UNO - 生产环境部署
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

echo [→] 开始部署...
echo.

call node deploy-pnpm.js

if %errorlevel% neq 0 (
    echo.
    echo [✗] 部署失败
    pause
    exit /b 1
)

echo.
echo [✓] 部署成功！
pause
