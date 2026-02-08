@echo off
chcp 65001 >nul
echo 🚀 开始构建 ChemistryUNO...
echo.

REM 1. 构建前端
echo 📦 [1/4] 构建前端...
cd frontend
call npm install
if errorlevel 1 (
    echo ❌ npm install 失败
    exit /b 1
)
call npm run build
if errorlevel 1 (
    echo ❌ 前端构建失败
    exit /b 1
)
echo ✓ 前端构建完成
echo.

REM 2. 复制前端文件到后端 static 目录
echo 📂 [2/4] 复制前端文件到后端...
if exist "..\backend\static\dist" (
    rmdir /s /q "..\backend\static\dist"
)
if not exist "..\backend\static" (
    mkdir "..\backend\static"
)
xcopy /E /I /Y /Q dist "..\backend\static\dist"
if errorlevel 1 (
    echo ❌ 文件复制失败
    exit /b 1
)
echo ✓ 文件复制完成
echo.

REM 3. 构建后端
echo 🔨 [3/4] 构建后端...
cd ..\backend
call go mod tidy
if errorlevel 1 (
    echo ❌ go mod tidy 失败
    exit /b 1
)
call go build -ldflags="-s -w" -o ..\bin\chemistryuno.exe main.go
if errorlevel 1 (
    echo ❌ 后端构建失败
    exit /b 1
)
echo ✓ 后端构建完成
echo.

REM 4. 显示构建结果
echo ✅ [4/4] 构建完成！
echo.
echo 📁 可执行文件位置: bin\chemistryuno.exe
for %%A in (..\bin\chemistryuno.exe) do echo 📦 文件大小: %%~zA 字节
echo.
echo 🎯 运行方式:
echo    bin\chemistryuno.exe
echo.
echo 💡 提示: 前端文件已嵌入到二进制文件中，部署时只需上传 bin\chemistryuno.exe 和 .env 文件
echo.

cd ..
