# 📚 Documentation Index

Chemistry UNO 项目的完整文档导航。

---

## 🚀 快速开始

**新用户必读** → [快速开始指南](./guides/QUICK_START.md)

新手开发者的 5 分钟快速上手指南，包括：
- 环境设置
- 首次运行
- 主要命令

---

## 📖 指南类文档 (docs/guides/)

| 文件 | 说明 |
|------|------|
| [QUICK_START.md](./guides/QUICK_START.md) | ⭐ 开发者快速开始（5分钟） |
| [COMMANDS.md](./guides/COMMANDS.md) | 所有可用命令汇总 |
| [DEPLOYMENT.md](./guides/DEPLOYMENT.md) | 生产环境部署指南 |

---

## 🎮 功能文档 (docs/features/)

| 文件 | 说明 |
|------|------|
| [LEVEL_SYSTEM.md](./features/LEVEL_SYSTEM.md) | 等级系统实现细节 |
| [PLUGIN_SYSTEM.md](./features/PLUGIN_SYSTEM.md) | 插件系统开发指南 |

---

## 🧪 测试文档 (docs/testing/)

| 文件 | 说明 |
|------|------|
| [TEST_FRAMEWORK_GUIDE.md](./testing/TEST_FRAMEWORK_GUIDE.md) | 测试框架使用指南 |
| [TEST_COMPLETION_REPORT.md](./testing/TEST_COMPLETION_REPORT.md) | 测试覆盖率报告 |
| [TEST_SCRIPTS_INDEX.md](./testing/TEST_SCRIPTS_INDEX.md) | 所有测试脚本索引 |
| [TEST_INTEGRATION_SUMMARY.md](./testing/TEST_INTEGRATION_SUMMARY.md) | 测试集成完成报告 |
| [PYTHON_MERGE_REPORT.md](./testing/PYTHON_MERGE_REPORT.md) | Python文件合并报告 |

---

## ✨ 性能优化 (docs/)

| 文件 | 说明 |
|------|------|
| [PERFORMANCE_OPTIMIZATION_REPORT.md](./PERFORMANCE_OPTIMIZATION_REPORT.md) | 前端性能优化完成报告（↓69% JS体积，↓72% 首屏时间） |

---

## 📝 文档规范 (docs/)

| 文件 | 说明 |
|------|------|
| [MARKDOWN_STANDARDIZATION_GUIDE.md](./MARKDOWN_STANDARDIZATION_GUIDE.md) | Markdown 书写规范指南（全局标准） |
| [MARKDOWN_STANDARDIZATION_REPORT.md](./MARKDOWN_STANDARDIZATION_REPORT.md) | Markdown 规范化执行报告（进度追踪） |

---

## ⚙️ 后端文档 (backend/docs/)

| 文件 | 说明 |
|------|------|
| [API_DOCUMENTATION.md](../backend/docs/API_DOCUMENTATION.md) | 完整的 API 文档和架构说明 |

---

## 🎨 前端文档 (frontend/docs/)

| 文件 | 说明 |
|------|------|
| [COMPOSABLES.md](../frontend/docs/COMPOSABLES.md) | Vue 3 组合式 API 框架 |
| [CSS_ARCHITECTURE.md](../frontend/docs/CSS_ARCHITECTURE.md) | CSS架构和主题系统 |
| [FEEDBACK_SYSTEM.md](../frontend/docs/FEEDBACK_SYSTEM.md) | 用户反馈系统 |
| [PING_SYSTEM.md](../frontend/docs/PING_SYSTEM.md) | 实时Ping系统 |
| [TUTORIAL_GUIDE.md](../frontend/docs/TUTORIAL_GUIDE.md) | 游戏内教程指南 |
| [TUTORIAL_IMPLEMENTATION.md](../frontend/docs/TUTORIAL_IMPLEMENTATION.md) | 教程系统实现 |

---

## 📜 法律和政策 (docs/)

| 文件 | 说明 |
|------|------|
| [PRIVACY_POLICY.md](./PRIVACY_POLICY.md) | 隐私政策 |
| [USER_AGREEMENT.md](./USER_AGREEMENT.md) | 用户协议 |

---

## 🏠 项目总体说明

- [README.md](../README.md) - 项目主页和概览

---

## 🗂️ 完整目录结构

```text
chemistryuno/
├── README.md                           # 项目总说明 ⭐
│
├── docs/                               # 📚 主文档目录
│   ├── INDEX.md                        # 文档导航（本文件）
│   ├── PRIVACY_POLICY.md              # 隐私政策
│   ├── USER_AGREEMENT.md              # 用户协议
│   │
│   ├── guides/                         # 📖 快速指南
│   │   ├── QUICK_START.md             # ⭐ 新手必读
│   │   ├── COMMANDS.md                # 命令速查表
│   │   └── DEPLOYMENT.md              # 部署指南
│   │
│   ├── features/                       # 🎮 功能文档
│   │   ├── LEVEL_SYSTEM.md            # 等级系统
│   │   └── PLUGIN_SYSTEM.md           # 插件开发
│   │
│   └── testing/                        # 🧪 测试文档
│       ├── TEST_FRAMEWORK_GUIDE.md    # 测试框架
│       ├── TEST_COMPLETION_REPORT.md  # 完成报告
│       ├── TEST_INTEGRATION_SUMMARY.md# 集成总结
│       ├── TEST_SCRIPTS_INDEX.md      # 脚本索引
│       └── PYTHON_MERGE_REPORT.md     # 合并报告
│
├── backend/
│   └── docs/
│       └── API_DOCUMENTATION.md        # 📡 API文档
│
├── frontend/
│   └── docs/
│       ├── COMPOSABLES.md             # 组合式API
│       ├── CSS_ARCHITECTURE.md        # 样式系统
│       ├── FEEDBACK_SYSTEM.md         # 反馈系统
│       ├── PING_SYSTEM.md            # Ping系统
│       ├── TUTORIAL_GUIDE.md          # 教程指南
│       └── TUTORIAL_IMPLEMENTATION.md # 实现细节
│
└── test_main.py                        # 🧪 统一测试入口
```

---

## 📚 按用途快速查找

### "我想开始开发"
1. 首先阅读 [QUICK_START.md](./guides/QUICK_START.md)
2. 查看 [COMMANDS.md](./guides/COMMANDS.md) 了解可用命令
3. 根据任务选择相应功能文档

### "我想部署到生产"
→ [DEPLOYMENT.md](./guides/DEPLOYMENT.md)

### "我想了解API"
→ [backend/docs/API_DOCUMENTATION.md](../backend/docs/API_DOCUMENTATION.md)

### "我想开发前端功能"
→ [frontend/docs/](../frontend/docs/) 查阅相关组件文档

### "我想运行测试"
→ [TEST_FRAMEWORK_GUIDE.md](./testing/TEST_FRAMEWORK_GUIDE.md)

### "我想开发插件/扩展"
→ [PLUGIN_SYSTEM.md](./features/PLUGIN_SYSTEM.md)

### "我想了解等级系统"
→ [LEVEL_SYSTEM.md](./features/LEVEL_SYSTEM.md)

---

## 📊 文档统计

- **总文档数**: 27 个
- **主要目录**: 6 个
- **快速指南**: 3 个
- **功能文档**: 2 个
- **测试文档**: 5 个
- **后端文档**: 1 个
- **前端文档**: 6 个
- **法律文挡**: 2 个

---

## ✅ 最后更新

- **日期**: 2025年3月22日
- **版本**: 2.0（重组版）
- **状态**: 文档结构优化完成

---

## 💡 寻求帮助

- 有问题？查看相应的 README.md
- 找不到答案？查看 [QUICK_START.md](./guides/QUICK_START.md)
- 仍未解决？提交 GitHub Issue

---

**提示**: 这份索引会随着项目发展持续更新。
