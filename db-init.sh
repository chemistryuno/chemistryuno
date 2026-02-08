#!/bin/bash

# 设置 UTF-8 编码
export LANG=en_US.UTF-8

echo ""
echo "╔════════════════════════════════════════════════════════╗"
echo "║         Chemistry UNO - 数据库初始化工具               ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# 检查后端进程是否正在运行
if pgrep -x "chemistryuno" > /dev/null; then
    echo "⚠️  警告: 检测到后端服务正在运行！"
    echo ""
    echo "选项："
    echo "  [1] 停止后端服务并继续初始化"
    echo "  [2] 取消初始化"
    echo ""
    read -p "请选择 (1/2): " choice

    if [ "$choice" = "1" ]; then
        echo ""
        echo "🛑 正在停止后端服务..."
        pkill -9 chemistryuno
        sleep 2
        echo "✅ 后端服务已停止"
    else
        echo ""
        echo "❌ 已取消初始化"
        exit 0
    fi
fi

# 同时检查 Go 进程
if pgrep -x "go" > /dev/null; then
    echo ""
    echo "⚠️  警告: 检测到 Go 进程正在运行（可能是 go run main.go）"
    echo ""
    read -p "是否停止？(y/n): " choice2

    if [ "$choice2" = "y" ] || [ "$choice2" = "Y" ]; then
        echo "🛑 正在停止 Go 进程..."
        pkill -9 go
        sleep 2
        echo "✅ Go 进程已停止"
    fi
fi

echo ""
echo "🔄 开始初始化数据库..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 运行数据库初始化
pnpm db:init
exit_code=$?

if [ $exit_code -ne 0 ]; then
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "❌ 数据库初始化失败！"
    echo ""
    echo "错误码: $exit_code"
    echo ""
    read -p "按回车键退出..."
    exit $exit_code
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ 数据库初始化完成！"
echo ""
echo "是否启动后端服务？"
echo "  [1] 是，启动后端服务"
echo "  [2] 否，完成并退出"
echo ""
read -p "请选择 (1/2): " restart

if [ "$restart" = "1" ]; then
    echo ""
    echo "🚀 正在启动后端服务..."

    # 根据操作系统选择终端
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        osascript -e 'tell app "Terminal" to do script "cd '$(pwd)' && pnpm backend"'
    else
        # Linux
        if command -v gnome-terminal &> /dev/null; then
            gnome-terminal -- bash -c "cd $(pwd) && pnpm backend; exec bash"
        elif command -v xterm &> /dev/null; then
            xterm -e "cd $(pwd) && pnpm backend" &
        else
            nohup pnpm backend > backend.log 2>&1 &
            echo "✅ 后端服务已在后台启动，日志输出到 backend.log"
        fi
    fi

    sleep 2
else
    echo ""
    echo "💡 提示: 使用 'pnpm backend' 命令启动后端服务"
fi

echo ""
read -p "按回车键退出..."
