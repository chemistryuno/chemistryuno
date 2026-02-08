@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo.
echo ╔════════════════════════════════════════════════════════╗
echo ║         Chemistry UNO - 数据库初始化工具                ║
echo ╚════════════════════════════════════════════════════════╝
echo.

:: 检查后端进程是否正在运行
tasklist /FI "IMAGENAME eq chemistryuno.exe" 2>NUL | find /I /N "chemistryuno.exe" >NUL
if "%ERRORLEVEL%"=="0" (
    echo ⚠️  警告: 检测到后端服务正在运行！
    echo.
    echo 选项：
    echo   [1] 停止后端服务并继续初始化
    echo   [2] 取消初始化
    echo.
    set /p choice="请选择 (1/2): "

    if "!choice!"=="1" (
        echo.
        echo 🛑 正在停止后端服务...
        taskkill /F /IM chemistryuno.exe >nul 2>&1
        timeout /t 2 /nobreak >nul
        echo ✅ 后端服务已停止
    ) else (
        echo.
        echo ❌ 已取消初始化
        pause
        exit /b 0
    )
)

:: 同时检查 Go 进程（可能是 go run main.go）
tasklist /FI "IMAGENAME eq go.exe" 2>NUL | find /I /N "go.exe" >NUL
if "%ERRORLEVEL%"=="0" (
    echo.
    echo ⚠️  警告: 检测到 Go 进程正在运行（可能是 go run main.go）
    echo.
    set /p choice2="是否停止？(y/n): "

    if /i "!choice2!"=="y" (
        echo 🛑 正在停止 Go 进程...
        taskkill /F /IM go.exe >nul 2>&1
        timeout /t 2 /nobreak >nul
        echo ✅ Go 进程已停止
    )
)

echo.
echo 🔄 开始初始化数据库...
echo ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo.

:: 运行数据库初始化
pnpm db:init

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    echo ❌ 数据库初始化失败！
    echo.
    echo 错误码: %ERRORLEVEL%
    echo.
    pause
    exit /b %ERRORLEVEL%
)

echo.
echo ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo ✅ 数据库初始化完成！
echo.
echo 是否启动后端服务？
echo   [1] 是，启动后端服务
echo   [2] 否，完成并退出
echo.
set /p restart="请选择 (1/2): "

if "!restart!"=="1" (
    echo.
    echo 🚀 正在启动后端服务...
    start "Chemistry UNO Backend" cmd /k "cd /d %~dp0 && pnpm backend"
    echo.
    echo ✅ 后端服务已在新窗口中启动
    echo.
    timeout /t 3 /nobreak >nul
) else (
    echo.
    echo 💡 提示: 使用 'pnpm backend' 命令启动后端服务
)

echo.
echo 按任意键退出...
pause >nul
