#!/bin/bash
# 化学UNO - Ubuntu 22 修复验证脚本

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo ""
echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}  化学UNO - Ubuntu 22 修复验证脚本${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo ""

# 检查函数
check_step() {
    local description=$1
    local command=$2
    
    echo -n -e "${CYAN}[→] ${description}... ${NC}"
    
    if eval "$command" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
        return 0
    else
        echo -e "${RED}✗${NC}"
        return 1
    fi
}

# 1. 检查 Node.js
echo -e "${YELLOW}1. 环境检查${NC}"
check_step "检查 Node.js" "command -v node"
if [ $? -eq 0 ]; then
    NODE_VERSION=$(node --version)
    echo -e "   版本: ${GREEN}${NODE_VERSION}${NC}"
fi

# 2. 检查 pnpm
check_step "检查 pnpm" "command -v pnpm"
if [ $? -eq 0 ]; then
    PNPM_VERSION=$(pnpm --version)
    echo -e "   版本: ${GREEN}${PNPM_VERSION}${NC}"
fi

echo ""
echo -e "${YELLOW}2. 文件检查${NC}"

# 3. 检查 serve.json
check_step "检查 client/public/serve.json" "test -f client/public/serve.json"
check_step "检查 client/build/serve.json" "test -f client/build/serve.json"

echo ""
echo -e "${YELLOW}3. 配置检查${NC}"

# 4. 检查 server/index.ts 是否包含 0.0.0.0
if grep -q "0.0.0.0" server/index.ts; then
    echo -e "${GREEN}[✓]${NC} 后端监听地址配置正确 (0.0.0.0)"
else
    echo -e "${RED}[✗]${NC} 后端监听地址未配置为 0.0.0.0"
fi

# 5. 检查端口配置
if grep -q "4001" server/index.ts; then
    echo -e "${GREEN}[✓]${NC} 后端端口配置正确 (4001)"
else
    echo -e "${RED}[✗]${NC} 后端端口配置可能有问题"
fi

echo ""
echo -e "${YELLOW}4. 端口占用检查${NC}"

# 6. 检查端口占用
PORT_4000=$(lsof -i :4000 2>/dev/null | grep LISTEN)
PORT_4001=$(lsof -i :4001 2>/dev/null | grep LISTEN)

if [ -z "$PORT_4000" ]; then
    echo -e "${GREEN}[✓]${NC} 端口 4000 可用"
else
    echo -e "${YELLOW}[!]${NC} 端口 4000 已被占用："
    echo "   $PORT_4000"
fi

if [ -z "$PORT_4001" ]; then
    echo -e "${GREEN}[✓]${NC} 端口 4001 可用"
else
    echo -e "${YELLOW}[!]${NC} 端口 4001 已被占用："
    echo "   $PORT_4001"
fi

echo ""
echo -e "${YELLOW}5. 服务测试${NC}"

# 7. 尝试连接后端
echo -n -e "${CYAN}[→] 测试后端连接 (localhost:4001)... ${NC}"
if curl -s http://localhost:4001/ > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC}"
    echo -e "   ${GREEN}后端服务正在运行${NC}"
else
    echo -e "${YELLOW}!${NC}"
    echo -e "   ${YELLOW}后端服务未运行（如已启动请忽略）${NC}"
fi

# 8. 测试 API 端点
echo -n -e "${CYAN}[→] 测试 API 端点 (/api/setup/check)... ${NC}"
if curl -s http://localhost:4001/api/setup/check > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${YELLOW}!${NC}"
    echo -e "   ${YELLOW}API 端点无响应（如已启动请忽略）${NC}"
fi

echo ""
echo -e "${YELLOW}6. 网络接口检查${NC}"

# 9. 显示可用的网络接口
echo -e "${CYAN}本机 IP 地址：${NC}"
ip -4 addr show | grep inet | grep -v "127.0.0.1" | awk '{print "   " $2}' | sed 's/\/.*$//'

echo ""
echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}  验证完成${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo ""

# 10. 提供下一步操作建议
echo -e "${GREEN}下一步操作：${NC}"
echo ""
echo -e "  ${CYAN}1. 如果服务未启动，运行：${NC}"
echo -e "     ${YELLOW}pnpm run deploy${NC}  或  ${YELLOW}./start.sh${NC}"
echo ""
echo -e "  ${CYAN}2. 启动后访问：${NC}"
echo -e "     前端: ${YELLOW}http://localhost:4000${NC}"
echo -e "     后端: ${YELLOW}http://localhost:4001${NC}"
echo -e "     管理: ${YELLOW}http://localhost:4000/admin${NC}"
echo ""
echo -e "  ${CYAN}3. 如需局域网访问，使用上面显示的 IP 地址${NC}"
echo ""
echo -e "  ${CYAN}4. 如遇到问题，查看：${NC}"
echo -e "     ${YELLOW}docs/UBUNTU_FIX.md${NC}"
echo ""
