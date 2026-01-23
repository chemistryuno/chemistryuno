#!/bin/bash
# 清理npm遗留配置

echo "🧹 清理npm遗留配置..."

# 检查并删除可能导致警告的npm配置
if command -v npm &> /dev/null; then
  # 删除可能的遗留全局配置
  npm config delete _-init-module 2>/dev/null || true
  npm config delete init.module 2>/dev/null || true
  npm config delete --global _-init-module 2>/dev/null || true
  npm config delete --global init.module 2>/dev/null || true
  
  echo "✅ npm配置已清理"
else
  echo "⚠️  npm未安装，跳过"
fi

echo "✅ 清理完成"
