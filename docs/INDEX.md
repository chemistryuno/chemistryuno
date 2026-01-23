# 📚 化学UNO 文档中心

欢迎来到化学UNO的文档中心！这里包含了游戏的所有使用指南。

## 🎯 快速导航

### 🚀 新手入门
- [快速开始指南](GETTING_STARTED.md) - 从零开始玩游戏
- [使用说明](使用说明.md) - 简明使用指南

### 🎮 游戏相关
- [移动端访问指南](MOBILE_ACCESS_GUIDE.md) - 在手机上玩游戏
- [管理面板指南](ADMIN_PANEL_GUIDE.md) - 修改游戏规则和配置
- [快速启动指南](QUICK_START.md) - 游戏规则速查

### 🛠️ 技术相关
- [跨平台支持说明](PLATFORM_SUPPORT.md) - Windows/Linux/Mac 使用指南
- [安全配置指南](SECURITY.md) - 密码和敏感信息管理 🔒
- [安全检查清单](SECURITY_CHECKLIST.md) - 代码安全最佳实践 🛡️
- [安全清理报告](SECURITY_CLEANUP_REPORT.md) - 调试代码清理详情 📋
- [项目优化总结](CLEANUP_SUMMARY.md) - 代码清理和优化记录
- [部署测试报告](DEPLOY_TEST_REPORT.md) - 部署测试详情

### 🔧 开发相关
- [开发者指南](DEVELOPER_GUIDE.md) - 代码结构和开发说明
- [安装指南](INSTALLATION_GUIDE.md) - 详细安装步骤

## 📋 常用命令

### 开发环境
```bash
# 一键启动
node start.js

# 或使用pnpm
pnpm run dev
```

### 生产部署
```bash
# 构建并部署
pnpm run deploy
```

### 安全检查
```bash
# 检查敏感信息
node check-security.js

# 清理调试日志
node remove-debug-logs.js
```

## 🛡️ 安全最佳实践

### ✅ 开发前
- 运行 `node check-security.js` 检查敏感信息
- 确保 `.env` 文件未被提交到版本控制

### ✅ 提交前
- 清理所有调试日志 `console.log/error/warn`
- 运行安全扫描确保无警告
- 编译检查确保无错误

### ✅ 部署前
- 设置强密码（至少12位）
- 检查所有配置文件
- 运行完整测试

详见 [安全检查清单](SECURITY_CHECKLIST.md)

## 🎮 游戏规则简介

1. **目标**：最先打完手牌获胜
2. **出牌**：打出能与上家物质发生化学反应的物质
3. **特殊牌**：
   - Au（金）：跳过下家
   - He/Ne/Ar/Kr（惰性气体）：反转方向
   - +2/+4：下家摸牌

详细规则见[主页](../README.md)

## 🔗 相关链接

- [项目主页](../README.md)
- [GitHub仓库](#)

## 💡 需要帮助？

如果遇到问题：
1. 查看对应的文档指南
2. 检查[常见问题](GETTING_STARTED.md#常见问题)
3. 提交 Issue

## 📊 最近更新

- ✅ 完成安全调试代码清理（61+处）
- ✅ 建立自动化安全扫描机制
- ✅ 统一错误处理策略
- ✅ 添加安全最佳实践文档

---

祝你玩得开心！🎮⚗️
