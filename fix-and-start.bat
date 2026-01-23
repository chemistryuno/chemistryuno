@echo off
echo ============================================
echo Chemistry UNO Alpha - 快速修复GCC问题
echo ============================================
echo.
echo 检测到您的GCC版本过旧，正在下载并配置新版本...
echo.

set MINGW_DIR=%~dp0tools\mingw64
set MINGW_ZIP=%~dp0tools\mingw64.7z

if exist "%MINGW_DIR%\bin\gcc.exe" (
    echo 检测到已有新版MinGW，跳过下载...
    goto :setup_env
)

echo 正在创建工具目录...
if not exist "%~dp0tools" mkdir "%~dp0tools"

echo.
echo 请从以下地址下载MinGW-w64:
echo https://github.com/niXman/mingw-builds-binaries/releases
echo.
echo 推荐下载: x86_64-posix-seh (任何8.0+版本)
echo 下载后解压到: %~dp0tools\mingw64
echo.
echo 或者访问: https://winlibs.com/ 下载最新版本
echo.
pause
echo.
echo 如果您已经下载并解压到正确位置，按任意键继续...
pause > nul

:setup_env
if not exist "%MINGW_DIR%\bin\gcc.exe" (
    echo.
    echo [错误] 未找到新版GCC，请确保已解压到: %MINGW_DIR%
    pause
    exit /b 1
)

echo.
echo 设置临时环境变量...
set PATH=%MINGW_DIR%\bin;%PATH%

echo.
echo 验证GCC版本...
gcc --version

echo.
echo 正在启动后端服务器...
cd "%~dp0backend"
go run main.go

pause
