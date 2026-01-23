#!/bin/bash
# 单端口集成模式启动脚本 (推荐)
# 只启动后端，由后端托管前端页面

# 颜色定义
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}  化学UNO - 集成模式启动${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo ""

# 检查构建
if [ ! -d "server/dist" ] || [ ! -d "client/build" ]; then
    echo -e "${CYAN}[→] 正在构建项目...${NC}"
    pnpm run build
fi

echo -e "${GREEN}[✓] 正在启动服务 (Port 4001)...${NC}"
echo -e "${CYAN}  服务地址: http://localhost:4001${NC}"
echo -e "${CYAN}  (前端页面已集成)${NC}"
echo ""

# 启动后端 (集成模式)
cd server
node dist/index.js
