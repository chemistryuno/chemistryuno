#!/bin/bash
# 化学UNO - 完整测试脚本（用于验证所有功能）

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo ""
echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}  化学UNO - 完整功能测试${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo ""

# 测试计数器
PASSED=0
FAILED=0

# 测试函数
test_endpoint() {
    local name=$1
    local url=$2
    local expected_code=${3:-200}
    
    echo -n -e "${CYAN}[测试] ${name}... ${NC}"
    
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null)
    
    if [ "$response" = "$expected_code" ]; then
        echo -e "${GREEN}✓ (HTTP $response)${NC}"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}✗ (HTTP $response, 期望 $expected_code)${NC}"
        ((FAILED++))
        return 1
    fi
}

# 检查服务是否运行
echo -e "${YELLOW}检查服务状态...${NC}"
echo ""

# 检查后端
if curl -s http://localhost:4001/ > /dev/null 2>&1; then
    echo -e "${GREEN}[✓] 后端服务运行中 (port 4001)${NC}"
else
    echo -e "${RED}[✗] 后端服务未运行${NC}"
    echo -e "${YELLOW}请先启动后端服务：cd server && pnpm run dev${NC}"
    exit 1
fi

# 检查前端
if curl -s http://localhost:4000/ > /dev/null 2>&1; then
    echo -e "${GREEN}[✓] 前端服务运行中 (port 4000)${NC}"
else
    echo -e "${RED}[✗] 前端服务未运行${NC}"
    echo -e "${YELLOW}请先启动前端服务：cd client && npx serve -s build -l 4000${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}开始API端点测试...${NC}"
echo ""

# 测试后端 API 端点
test_endpoint "根路径" "http://localhost:4001/"
test_endpoint "设置检查" "http://localhost:4001/api/setup/check"
test_endpoint "化合物列表" "http://localhost:4001/api/compounds"
test_endpoint "配置" "http://localhost:4001/api/config"

echo ""
echo -e "${YELLOW}测试前端路由...${NC}"
echo ""

# 测试前端路由
test_endpoint "主页" "http://localhost:4000/"
test_endpoint "Setup页面" "http://localhost:4000/setup"
test_endpoint "Admin页面" "http://localhost:4000/admin"

echo ""
echo -e "${YELLOW}测试创建游戏功能...${NC}"
echo ""

# 测试创建游戏
create_response=$(curl -s -X POST http://localhost:4001/api/game/create \
  -H "Content-Type: application/json" \
  -d '{
    "hostName": "TestHost",
    "maxPlayers": 4,
    "initialHandSize": 10
  }' 2>/dev/null)

if echo "$create_response" | grep -q "roomCode"; then
    room_code=$(echo "$create_response" | grep -o '"roomCode":"[^"]*"' | cut -d'"' -f4)
    echo -e "${GREEN}[✓] 创建房间成功 - 房间号: $room_code${NC}"
    ((PASSED++))
    
    # 测试房间信息
    echo -n -e "${CYAN}[测试] 获取房间信息... ${NC}"
    room_info=$(curl -s "http://localhost:4001/api/game/${room_code}/info" 2>/dev/null)
    if echo "$room_info" | grep -q "$room_code"; then
        echo -e "${GREEN}✓${NC}"
        ((PASSED++))
    else
        echo -e "${RED}✗${NC}"
        ((FAILED++))
    fi
else
    echo -e "${RED}[✗] 创建房间失败${NC}"
    echo "响应: $create_response"
    ((FAILED++))
fi

echo ""
echo -e "${YELLOW}测试加入游戏功能...${NC}"
echo ""

if [ ! -z "$room_code" ]; then
    join_response=$(curl -s -X POST http://localhost:4001/api/game/join \
      -H "Content-Type: application/json" \
      -d "{
        \"roomCode\": \"$room_code\",
        \"playerName\": \"TestPlayer\"
      }" 2>/dev/null)
    
    if echo "$join_response" | grep -q "playerId"; then
        echo -e "${GREEN}[✓] 加入房间成功${NC}"
        ((PASSED++))
    else
        echo -e "${RED}[✗] 加入房间失败${NC}"
        echo "响应: $join_response"
        ((FAILED++))
    fi
fi

echo ""
echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}  测试结果${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "  ${GREEN}通过: $PASSED${NC}"
echo -e "  ${RED}失败: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ 所有测试通过！系统运行正常。${NC}"
    echo ""
    echo -e "${CYAN}访问地址：${NC}"
    echo -e "  游戏: ${YELLOW}http://localhost:4000${NC}"
    echo -e "  管理: ${YELLOW}http://localhost:4000/admin${NC}"
    echo ""
    exit 0
else
    echo -e "${RED}✗ 部分测试失败，请检查错误信息。${NC}"
    echo ""
    echo -e "${CYAN}调试建议：${NC}"
    echo -e "  1. 查看后端日志"
    echo -e "  2. 检查浏览器控制台 (F12)"
    echo -e "  3. 运行诊断脚本: ${YELLOW}./verify-fix.sh${NC}"
    echo -e "  4. 查看故障排除文档: ${YELLOW}docs/TROUBLESHOOTING_UBUNTU.md${NC}"
    echo ""
    exit 1
fi
