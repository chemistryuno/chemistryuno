@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo ============================================
echo   ChemistryUNO 一键打包工具
echo   Windows / Android / iOS
echo ============================================
echo.

REM 解析参数
set BUILD_WIN=0
set BUILD_ANDROID=0
set BUILD_IOS=0
set SERVER_URL=

if "%~1"=="" (
    set BUILD_WIN=1
    set BUILD_ANDROID=1
    set BUILD_IOS=1
) else (
    for %%a in (%*) do (
        if /I "%%a"=="windows" set BUILD_WIN=1
        if /I "%%a"=="win" set BUILD_WIN=1
        if /I "%%a"=="android" set BUILD_ANDROID=1
        if /I "%%a"=="ios" set BUILD_IOS=1
        if /I "%%a"=="all" (
            set BUILD_WIN=1
            set BUILD_ANDROID=1
            set BUILD_IOS=1
        )
    )
)

echo 打包目标:
if %BUILD_WIN%==1 echo   [x] Windows
if %BUILD_ANDROID%==1 echo   [x] Android
if %BUILD_IOS%==1 echo   [x] iOS
echo.

REM ============================
REM  Step 1: 构建前端
REM ============================
echo [1/5] 构建前端...
cd frontend
call pnpm install --frozen-lockfile 2>nul || call npm install
if errorlevel 1 (
    echo [ERROR] 依赖安装失败
    goto :error
)
call pnpm run build 2>nul || call npm run build
if errorlevel 1 (
    echo [ERROR] 前端构建失败
    goto :error
)
echo [OK] 前端构建完成
echo.

REM ============================
REM  Step 2: Windows 打包
REM ============================
if %BUILD_WIN%==1 (
    echo [2/5] 打包 Windows 客户端...

    REM 复制前端到后端 static
    if exist "..\backend\static\dist" rmdir /s /q "..\backend\static\dist"
    if not exist "..\backend\static" mkdir "..\backend\static"
    xcopy /E /I /Y /Q dist "..\backend\static\dist" >nul
    if errorlevel 1 (
        echo [ERROR] 前端文件复制失败
        goto :error
    )

    REM 构建 Go 后端
    cd ..
    call go mod tidy
    if errorlevel 1 (
        echo [ERROR] go mod tidy 失败
        goto :error
    )

    if not exist "bin" mkdir bin

    REM Windows amd64
    set GOOS=windows
    set GOARCH=amd64
    call go build -ldflags="-s -w" -o bin\chemistryuno-windows-amd64.exe .
    if errorlevel 1 (
        echo [ERROR] Windows amd64 构建失败
        goto :error
    )
    echo   [OK] bin\chemistryuno-windows-amd64.exe

    REM Linux amd64 (bonus - 服务器部署用)
    set GOOS=linux
    set GOARCH=amd64
    call go build -ldflags="-s -w" -o bin\chemistryuno-linux-amd64 .
    if errorlevel 1 (
        echo   [WARN] Linux amd64 构建失败 (跳过)
    ) else (
        echo   [OK] bin\chemistryuno-linux-amd64
    )

    set GOOS=
    set GOARCH=
    cd frontend
    echo [OK] Windows 打包完成
    echo.
) else (
    echo [2/5] 跳过 Windows 打包
    echo.
)

REM ============================
REM  Step 3: 初始化 Capacitor (如需)
REM ============================
if %BUILD_ANDROID%==1 if not exist "android" (
    echo [3/5] 初始化 Capacitor...
    call npx cap add android
    if errorlevel 1 (
        echo [ERROR] Capacitor Android 初始化失败
        echo   请先运行: npm install @capacitor/core @capacitor/cli @capacitor/android
        goto :error
    )
    echo [OK] Capacitor Android 初始化完成
    echo.
)

if %BUILD_IOS%==1 if not exist "ios" (
    echo [3/5] 初始化 Capacitor iOS...
    call npx cap add ios
    if errorlevel 1 (
        echo [ERROR] Capacitor iOS 初始化失败
        echo   请先运行: npm install @capacitor/core @capacitor/cli @capacitor/ios
        echo   注意: iOS 构建需要在 macOS 上运行
        goto :error
    )
    echo [OK] Capacitor iOS 初始化完成
    echo.
)

REM ============================
REM  Step 4: Android 打包
REM ============================
if %BUILD_ANDROID%==1 (
    echo [4/5] 打包 Android 客户端...

    REM 同步 Web 资源到 Android 项目
    call npx cap sync android
    if errorlevel 1 (
        echo [ERROR] Capacitor Android 同步失败
        goto :error
    )

    REM 构建 APK
    cd android
    call gradlew.bat assembleRelease 2>nul || call .\gradlew.bat assembleRelease
    if errorlevel 1 (
        echo [WARN] Release APK 构建失败, 尝试 Debug 版本...
        call gradlew.bat assembleDebug 2>nul || call .\gradlew.bat assembleDebug
        if errorlevel 1 (
            echo [ERROR] Android APK 构建失败
            cd ..
            goto :error
        )
        REM 复制 Debug APK
        if not exist "..\bin" mkdir "..\bin"
        copy /Y app\build\outputs\apk\debug\app-debug.apk ..\bin\chemistryuno-debug.apk >nul 2>nul
        echo   [OK] bin\chemistryuno-debug.apk
    ) else (
        REM 复制 Release APK
        if not exist "..\bin" mkdir "..\bin"
        copy /Y app\build\outputs\apk\release\app-release.apk ..\bin\chemistryuno-release.apk >nul 2>nul
        copy /Y app\build\outputs\apk\release\app-release-unsigned.apk ..\bin\chemistryuno-release.apk >nul 2>nul
        echo   [OK] bin\chemistryuno-release.apk
    )
    cd ..
    echo [OK] Android 打包完成
    echo.
) else (
    echo [4/5] 跳过 Android 打包
    echo.
)

REM ============================
REM  Step 5: iOS 打包
REM ============================
if %BUILD_IOS%==1 (
    echo [5/5] 打包 iOS 客户端...

    REM 检测是否在 macOS 环境
    where xcodebuild >nul 2>nul
    if errorlevel 1 (
        echo [WARN] iOS 构建需要 macOS + Xcode 环境
        echo   已同步 Web 资源，请在 Mac 上执行:
        echo     cd frontend
        echo     npx cap sync ios
        echo     npx cap open ios
        echo   然后在 Xcode 中 Archive 导出 IPA
        call npx cap sync ios 2>nul
        echo [OK] iOS Web 资源已同步 (需要在 Mac 上完成最终构建)
    ) else (
        call npx cap sync ios
        if errorlevel 1 (
            echo [ERROR] Capacitor iOS 同步失败
            goto :error
        )
        echo   iOS 项目已同步, 请在 Xcode 中打开并构建:
        echo     npx cap open ios
        echo [OK] iOS 准备完成
    )
    echo.
) else (
    echo [5/5] 跳过 iOS 打包
    echo.
)

cd ..

REM ============================
REM  构建结果汇总
REM ============================
echo ============================================
echo   构建结果汇总
echo ============================================
echo.
echo   输出目录: bin\
echo.
if exist "bin\chemistryuno-windows-amd64.exe" (
    for %%A in (bin\chemistryuno-windows-amd64.exe) do echo   [Windows]  chemistryuno-windows-amd64.exe  (%%~zA bytes^)
)
if exist "bin\chemistryuno-linux-amd64" (
    for %%A in (bin\chemistryuno-linux-amd64) do echo   [Linux]    chemistryuno-linux-amd64  (%%~zA bytes^)
)
if exist "bin\chemistryuno-release.apk" (
    for %%A in (bin\chemistryuno-release.apk) do echo   [Android]  chemistryuno-release.apk  (%%~zA bytes^)
)
if exist "bin\chemistryuno-debug.apk" (
    for %%A in (bin\chemistryuno-debug.apk) do echo   [Android]  chemistryuno-debug.apk  (%%~zA bytes^)
)
echo.
echo   iOS: 请在 macOS + Xcode 中完成最终构建
echo.
echo ============================================
echo   部署提示
echo ============================================
echo.
echo   Windows/Linux: 上传二进制文件 + .env 即可运行
echo   Android: 安装 APK 前需在 capacitor.config.ts 中
echo            设置 server.url 为后端服务器地址
echo   iOS:     在 Xcode 中设置签名后 Archive 导出
echo.
goto :end

:error
echo.
echo [FAILED] 构建过程中发生错误，请检查上方日志
cd "%~dp0"
exit /b 1

:end
cd "%~dp0"
echo 构建完成!
