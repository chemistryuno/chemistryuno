@echo off
REM 单端口集成模式启动脚本 (Windows)
REM 只启动后端，由后端托管前端页面

echo ===========================================================
echo   Chemical UNO - Integrated Mode Start
echo ===========================================================
echo.

if not exist "server\dist" (
    echo [!] Server build not found. Please run 'npm run build' first.
    pause
    exit /b
)

if not exist "client\build" (
    echo [!] Client build not found. Please run 'npm run build' first.
    pause
    exit /b
)

echo [✓] Starting Server on Port 4001...
echo     Serving Integrated Frontend...
echo.

cd server
node dist/index.js
pause