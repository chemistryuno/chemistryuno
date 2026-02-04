@echo off
REM Chemistry UNO 后端编译脚本
REM 禁用CGO以避免MinGW链接器版本问题

echo 🔨 正在编译后端 (禁用CGO)...
set CGO_ENABLED=0
go build -o chemistryuno.exe
set CGO_ENABLED=1

if %ERRORLEVEL% EQU 0 (
    echo ✅ 编译成功！可执行文件: chemistryuno.exe
) else (
    echo ❌ 编译失败
    exit /b 1
)
