#!/bin/bash
set -e

echo "============================================"
echo "  ChemistryUNO 一键打包工具"
echo "  Windows / Android / iOS"
echo "============================================"
echo ""

BUILD_WIN=0
BUILD_ANDROID=0
BUILD_IOS=0

if [ $# -eq 0 ]; then
    BUILD_WIN=1
    BUILD_ANDROID=1
    BUILD_IOS=1
else
    for arg in "$@"; do
        case "${arg,,}" in
            windows|win) BUILD_WIN=1 ;;
            android) BUILD_ANDROID=1 ;;
            ios) BUILD_IOS=1 ;;
            all) BUILD_WIN=1; BUILD_ANDROID=1; BUILD_IOS=1 ;;
        esac
    done
fi

echo "打包目标:"
[ $BUILD_WIN -eq 1 ] && echo "  [x] Windows/Linux"
[ $BUILD_ANDROID -eq 1 ] && echo "  [x] Android"
[ $BUILD_IOS -eq 1 ] && echo "  [x] iOS"
echo ""

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"

# ============================
#  Step 1: 构建前端
# ============================
echo "[1/5] 构建前端..."
cd "$PROJECT_DIR/frontend"
pnpm install --frozen-lockfile 2>/dev/null || npm install
pnpm run build 2>/dev/null || npm run build
echo "[OK] 前端构建完成"
echo ""

# ============================
#  Step 2: Windows/Linux 打包
# ============================
if [ $BUILD_WIN -eq 1 ]; then
    echo "[2/5] 打包服务端..."

    # 复制前端到后端 static
    rm -rf "$PROJECT_DIR/backend/static/dist"
    mkdir -p "$PROJECT_DIR/backend/static"
    cp -r dist "$PROJECT_DIR/backend/static/dist"

    cd "$PROJECT_DIR"
    go mod tidy
    mkdir -p bin

    # Windows amd64
    echo "  构建 Windows amd64..."
    GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/chemistryuno-windows-amd64.exe .
    echo "  [OK] bin/chemistryuno-windows-amd64.exe"

    # Linux amd64
    echo "  构建 Linux amd64..."
    GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/chemistryuno-linux-amd64 .
    echo "  [OK] bin/chemistryuno-linux-amd64"

    # macOS arm64
    echo "  构建 macOS arm64..."
    GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/chemistryuno-darwin-arm64 . 2>/dev/null && \
        echo "  [OK] bin/chemistryuno-darwin-arm64" || \
        echo "  [WARN] macOS arm64 构建跳过"

    cd "$PROJECT_DIR/frontend"
    echo "[OK] 服务端打包完成"
    echo ""
else
    echo "[2/5] 跳过服务端打包"
    echo ""
fi

# ============================
#  Step 3: 初始化 Capacitor
# ============================
if [ $BUILD_ANDROID -eq 1 ] && [ ! -d "android" ]; then
    echo "[3/5] 初始化 Capacitor Android..."
    npx cap add android
    echo "[OK] Capacitor Android 初始化完成"
    echo ""
fi

if [ $BUILD_IOS -eq 1 ] && [ ! -d "ios" ]; then
    echo "[3/5] 初始化 Capacitor iOS..."
    npx cap add ios
    echo "[OK] Capacitor iOS 初始化完成"
    echo ""
fi

# ============================
#  Step 4: Android 打包
# ============================
if [ $BUILD_ANDROID -eq 1 ]; then
    echo "[4/5] 打包 Android 客户端..."
    npx cap sync android

    cd android
    if ./gradlew assembleRelease 2>/dev/null; then
        mkdir -p "$PROJECT_DIR/bin"
        cp -f app/build/outputs/apk/release/app-release*.apk "$PROJECT_DIR/bin/chemistryuno-release.apk" 2>/dev/null || true
        echo "  [OK] bin/chemistryuno-release.apk"
    else
        echo "  [WARN] Release 构建失败, 尝试 Debug..."
        if ./gradlew assembleDebug; then
            mkdir -p "$PROJECT_DIR/bin"
            cp -f app/build/outputs/apk/debug/app-debug.apk "$PROJECT_DIR/bin/chemistryuno-debug.apk"
            echo "  [OK] bin/chemistryuno-debug.apk"
        else
            echo "  [ERROR] Android 构建失败"
        fi
    fi
    cd "$PROJECT_DIR/frontend"
    echo "[OK] Android 打包完成"
    echo ""
else
    echo "[4/5] 跳过 Android 打包"
    echo ""
fi

# ============================
#  Step 5: iOS 打包
# ============================
if [ $BUILD_IOS -eq 1 ]; then
    echo "[5/5] 打包 iOS 客户端..."
    npx cap sync ios

    if command -v xcodebuild &>/dev/null; then
        echo "  Xcode 可用, 请运行以下命令打开项目:"
        echo "    cd frontend && npx cap open ios"
        echo "  然后在 Xcode 中 Product > Archive 导出 IPA"
    else
        echo "  [WARN] 未检测到 Xcode, iOS 需要在 macOS 上构建"
        echo "    在 Mac 上执行:"
        echo "      cd frontend && npx cap sync ios && npx cap open ios"
    fi
    echo "[OK] iOS 准备完成"
    echo ""
else
    echo "[5/5] 跳过 iOS 打包"
    echo ""
fi

cd "$PROJECT_DIR"

# ============================
#  构建结果汇总
# ============================
echo "============================================"
echo "  构建结果汇总"
echo "============================================"
echo ""
echo "  输出目录: bin/"
echo ""
[ -f "bin/chemistryuno-windows-amd64.exe" ] && echo "  [Windows]  $(ls -lh bin/chemistryuno-windows-amd64.exe | awk '{print $5}')"
[ -f "bin/chemistryuno-linux-amd64" ] && echo "  [Linux]    $(ls -lh bin/chemistryuno-linux-amd64 | awk '{print $5}')"
[ -f "bin/chemistryuno-darwin-arm64" ] && echo "  [macOS]    $(ls -lh bin/chemistryuno-darwin-arm64 | awk '{print $5}')"
[ -f "bin/chemistryuno-release.apk" ] && echo "  [Android]  $(ls -lh bin/chemistryuno-release.apk | awk '{print $5}')"
[ -f "bin/chemistryuno-debug.apk" ] && echo "  [Android]  $(ls -lh bin/chemistryuno-debug.apk | awk '{print $5}')"
echo ""
echo "  iOS: 请在 macOS + Xcode 中完成最终构建"
echo ""
echo "============================================"
echo "  部署提示"
echo "============================================"
echo ""
echo "  Windows/Linux: 上传二进制文件 + .env 即可运行"
echo "  Android: 安装 APK 前需在 capacitor.config.ts 中"
echo "           设置 server.url 为后端服务器地址"
echo "  iOS:     在 Xcode 中设置签名后 Archive 导出"
echo ""
echo "构建完成!"
