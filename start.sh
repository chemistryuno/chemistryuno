#!/bin/bash
# Chemistry UNO - 快速启动脚本（Linux/Mac）

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

echo ""
echo -e "${MAGENTA}═══════════════════════════════════════════════════════════${NC}"
echo -e "${MAGENTA}  Chemistry UNO - 化学UNO 快速启动${NC}"
echo -e "${MAGENTA}═══════════════════════════════════════════════════════════${NC}"
echo ""

# 检查 Node.js
if ! command -v node &> /dev/null; then
    echo -e "${RED}[✗] 请先安装 Node.js: https://nodejs.org/${NC}"
    exit 1
fi

# 检查 pnpm
if ! command -v pnpm &> /dev/null; then
    echo -e "${RED}[✗] 请先安装 pnpm: npm install -g pnpm${NC}"
    exit 1
fi

# 检查依赖是否已安装
if [ ! -d "node_modules" ]; then
    echo -e "${CYAN}[→] 首次运行，正在安装依赖...${NC}"
    pnpm install
    if [ $? -ne 0 ]; then
        echo -e "${RED}[✗] 依赖安装失败${NC}"
        exit 1
    fi
    echo -e "${GREEN}[✓] 依赖安装完成${NC}"
    echo ""
fi

echo -e "${GREEN}[✓] 正在启动开发服务器...${NC}"
echo ""
echo -e "${CYAN}  前端: http://localhost:4000${NC}"
echo -e "${CYAN}  后端: http://localhost:4001${NC}"
echo -e "${CYAN}  管理面板: http://localhost:4000/admin${NC}"
echo ""
echo -e "${CYAN}按 Ctrl+C 停止服务${NC}"
echo ""

# 启动开发服务器
pnpm run dev
