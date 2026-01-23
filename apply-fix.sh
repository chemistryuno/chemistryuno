#!/bin/bash
# 化学UNO - 快速应用 Ubuntu 修复

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo ""
echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}  化学UNO - 应用 Ubuntu 22 修复${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo ""

# 1. 创建 serve.json 配置
echo -e "${YELLOW}[1/3] 创建 serve.json 配置文件...${NC}"

# 确保目录存在
mkdir -p client/public
mkdir -p client/build

# 创建 serve.json 内容
SERVE_JSON='{
  "rewrites": [
    { "source": "**", "destination": "/index.html" }
  ],
  "headers": [
    {
      "source": "**",
      "headers": [
        {
          "key": "Cache-Control",
          "value": "no-cache"
        }
      ]
    }
  ]
}'

# 写入文件
echo "$SERVE_JSON" > client/public/serve.json
echo "$SERVE_JSON" > client/build/serve.json

echo -e "${GREEN}[✓] serve.json 配置文件已创建${NC}"
echo ""

# 2. 更新 server/index.ts
echo -e "${YELLOW}[2/3] 检查后端配置...${NC}"

if grep -q "0.0.0.0" server/index.ts; then
    echo -e "${GREEN}[✓] 后端监听地址已正确配置${NC}"
else
    echo -e "${RED}[!] 警告: 后端监听地址可能需要手动修改${NC}"
    echo -e "${YELLOW}    请在 server/index.ts 中查找 server.listen()${NC}"
    echo -e "${YELLOW}    将 HOST 设置为 '0.0.0.0'${NC}"
fi
echo ""

# 3. 设置执行权限
echo -e "${YELLOW}[3/3] 设置脚本执行权限...${NC}"
chmod +x start.sh 2>/dev/null
chmod +x deploy.sh 2>/dev/null
chmod +x verify-fix.sh 2>/dev/null
echo -e "${GREEN}[✓] 执行权限已设置${NC}"
echo ""

# 完成
echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}  修复应用完成！${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${CYAN}修复内容：${NC}"
echo -e "  ${GREEN}✓${NC} 添加了 SPA 路由支持 (serve.json)"
echo -e "  ${GREEN}✓${NC} 设置了脚本执行权限"
echo ""
echo -e "${CYAN}下一步：${NC}"
echo -e "  1. 运行验证脚本: ${YELLOW}./verify-fix.sh${NC}"
echo -e "  2. 重新构建项目: ${YELLOW}pnpm run build${NC}"
echo -e "  3. 启动服务: ${YELLOW}pnpm run deploy${NC}"
echo ""
echo -e "${CYAN}详细说明请查看: ${YELLOW}docs/UBUNTU_FIX.md${NC}"
echo ""
