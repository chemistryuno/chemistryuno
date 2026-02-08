#!/bin/bash
set -e

echo "🚀 开始构建 ChemistryUNO..."
echo ""

# 1. 构建前端
echo "📦 [1/4] 构建前端..."
cd frontend
npm install
npm run build
echo "✓ 前端构建完成"
echo ""

# 2. 复制前端文件到后端 static 目录
echo "📂 [2/4] 复制前端文件到后端..."
rm -rf ../backend/static/dist
mkdir -p ../backend/static
cp -r dist ../backend/static/
echo "✓ 文件复制完成"
echo ""

# 3. 构建后端
echo "🔨 [3/4] 构建后端..."
cd ../backend
go mod tidy
go build -ldflags="-s -w" -o ../bin/chemistryuno main.go
echo "✓ 后端构建完成"
echo ""

# 4. 显示构建结果
echo "✅ [4/4] 构建完成！"
echo ""
echo "📁 可执行文件位置: bin/chemistryuno"
echo "📦 文件大小: $(du -h ../bin/chemistryuno | cut -f1)"
echo ""
echo "🎯 运行方式:"
echo "   ./bin/chemistryuno"
echo ""
echo "💡 提示: 前端文件已嵌入到二进制文件中，部署时只需上传 bin/chemistryuno 和 .env 文件"
