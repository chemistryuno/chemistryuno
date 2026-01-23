#!/bin/bash
# Chemistry UNO - 生产环境部署脚本（Linux/Mac）

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

echo ""
echo -e "${MAGENTA}═══════════════════════════════════════════════════════════${NC}"
echo -e "${MAGENTA}  Chemistry UNO - 生产环境部署${NC}"
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

echo -e "${CYAN}[→] 开始部署...${NC}"
echo ""

node deploy-pnpm.js

if [ $? -ne 0 ]; then
    echo ""
    echo -e "${RED}[✗] 部署失败${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}[✓] 部署成功！${NC}"
