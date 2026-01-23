@echo off
echo ========================================
echo Chemistry UNO - 后端启动 (兼容模式)
echo ========================================
echo.

cd /d %~dp0backend

echo 正在启动后端服务器 (使用兼容编译模式)...
echo.

REM 设置环境变量以使用兼容的链接选项
set CGO_LDFLAGS=-Wl,--no-insert-timestamp
set CGO_ENABLED=1

REM 尝试使用go build先编译
echo 步骤 1/2: 编译后端...
go build -ldflags="-s -w" -o chemistryuno-backend.exe main.go

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [错误] 编译失败！
    echo.
    echo 您的GCC版本过旧，请运行以下脚本之一来修复:
    echo   - fix-and-start.ps1  (PowerShell - 可自动下载)
    echo   - fix-and-start.bat  (批处理 - 需手动下载)
    echo.
    echo 或者手动下载新版MinGW:
    echo   https://winlibs.com/
    echo   https://github.com/niXman/mingw-builds-binaries/releases
    echo.
    pause
    exit /b 1
)

echo.
echo 步骤 2/2: 启动服务器...
echo.
chemistryuno-backend.exe

pause
