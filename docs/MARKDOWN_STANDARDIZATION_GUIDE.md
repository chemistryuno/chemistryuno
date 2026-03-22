# Markdown 书写规范指南

**最后更新**: 2026 年 3 月 22 日  
**规范版本**: 1.0

---

## 📋 概述

本文档定义 Chemistry UNO 项目的统一 Markdown 规范标准，确保所有文档书写风格一致、易于阅读和维护。

---

## 1️⃣ 标题规范

### 标题级别

- 仅使用 **一级到三级** 标题（`#`、`##`、`###`）
- 避免超过三级标题
- 每个标题前后各空一行

### 示例

```markdown
# 一级标题 (主标题)
## 二级标题 (章节)
### 三级标题 (小节)
```

---

## 2️⃣ 代码块规范

### 必须指定语言

所有代码块 **必须** 标注语言类型。格式：` ```语言名 `

### 支持的语言标记

| 语言 | 标记 | 说明 |
|------|------|------|
| Bash/Shell | `bash` | Shell 脚本命令 |
| Python | `python` | Python 代码 |
| JavaScript | `javascript` | JavaScript 代码 |
| TypeScript | `typescript` | TypeScript 代码 |
| Go | `go` | Go 语言代码 |
| JSON | `json` | JSON 数据 |
| YAML | `yaml` | YAML 配置文件 |
| TOML | `toml` | TOML 配置文件 |
| SQL | `sql` | SQL 查询语句 |
| HTML | `html` | HTML 标签 |
| CSS | `css` | CSS 样式 |
| Vue | `vue` | Vue 单文件组件 |
| XML | `xml` | XML 文档 |
| Markdown | `markdown` | Markdown 文本 |
| Text | `text` | 纯文本/目录结构 |

### 代码块示例

✅ **正确**：
```python
def hello_world():
    print("Hello, World!")
```

✅ **正确**：
```bash
npm run build
npm run dev
```

❌ **错误**：
```
def hello_world():
    print("Hello, World!")
```

---

## 3️⃣ 列表规范

### 无序列表

- 使用 **统一的 `-` 符号**（不用 `*` 或 `+`）
- 列表项的缩进保持一致（子列表缩进 2 个空格）
- 列表前后各空一行

### 示例

```markdown
主要功能：
- 实时聊天
- 用户认证
- 数据管理

高级功能：
- 某个功能
  - 子功能 1
  - 子功能 2
- 另一个功能
```

---

## 4️⃣ 表格规范

### 表格格式

- 使用标准 Markdown 表格语法
- 表格前后各空一行
- 列分隔符为 `|` 符号
- 每列宽度应相对均衡

### 示例

```markdown
| 功能 | 描述 | 状态 |
|------|------|------|
| API | 后端接口 | ✅ 完成 |
| 认证 | 身份验证 | ✅ 完成 |
```

---

## 5️⃣ 中文排版规范

### 中英文间距

中文与英文字母/单词之间应有 **1 个空格**。

✅ **正确**：
- Chemistry UNO 是一款游戏
- 使用 API 接口
- 支持 WebSocket 连接
- JWT 令牌

❌ **错误**：
- ChemistryUNO 是一款游戏
- 使用API接口
- 支持WebSocket连接
- JWT令牌

### 数字与中文间距

数字与中文字符之间应有 **1 个空格**。

✅ **正确**：
- 每 3 秒发送 ping
- 最近 10 次延迟
- 至少 32 字符

❌ **错误**：
- 每3秒发送ping
- 最近10次延迟
- 至少32字符

### 标点符号

- 使用中文标点符号（，。；：）
- 「」用于引用重要词汇或标题
- 避免使用英文标点（except in code blocks）

---

## 6️⃣ 链接规范

### 链接格式

使用标准 Markdown 链接格式：` [文本](URL) `

### 相对路径

- 文件链接使用相对路径：`./file.md`、`../folder/file.md`
- 不使用绝对路径或文件:// 协议

### 示例

```markdown
- [快速开始](./guides/QUICK_START.md)
- [API 文档](../backend/docs/API_DOCUMENTATION.md)
- [GitHub 仓库](https://github.com/chemistryuno/chemistryuno)
```

---

## 7️⃣ 空行规范

### 间距规则

| 位置 | 空行数 | 说明 |
|------|-------|------|
| 标题后 | 1 行 | 标题与内容间隔 |
| 段落间 | 1 行 | 段落分隔 |
| 列表前后 | 1 行 | 列表与其他内容分隔 |
| 代码块前后 | 1 行 | 代码块与其他内容分隔 |
| 表格前后 | 1 行 | 表格与其他内容分隔 |

### 示例

```markdown
# 主标题

## 小标题

这是一个段落。

- 列表项 1
- 列表项 2

这是另一个段落。

```python
code example
```

最后一个段落。
```

---

## 8️⃣ 特殊符号规范

### 强调

- 粗体：使用 `**文本**` 或 `__文本__`
- 斜体：使用 `*文本*` 或 `_文本_`
- 不使用 HTML 标签

### 代码引用

- 行内代码：使用单反引号 `` `代码` ``
- 避免在段落中使用​多行代码块

### 列表

- 编号列表：使用 `1. 2. 3.`（自动编号）
- 无序列表：使用 `-`（统一规范）

---

## 9️⃣ 文件组织规范

### 文件名

- 使用大写 + 下划线：`TEST_COMPLETION_REPORT.md`
- 不使用小写或连字符：`test-completion-report.md` ❌

### 目录结构

```
docs/
├── INDEX.md                  # 文档导航
├── guides/                   # 快速指南
│   ├── QUICK_START.md
│   ├── COMMANDS.md
│   └── DEPLOYMENT.md
├── features/                 # 功能文档
│   ├── LEVEL_SYSTEM.md
│   └── PLUGIN_SYSTEM.md
└── testing/                  # 测试文档
    ├── TEST_FRAMEWORK_GUIDE.md
    └── TEST_COMPLETION_REPORT.md
```

---

## 🔟 检查清单

在提交 Markdown 文件前，请检查以下项目：

- [ ] 所有标题使用 `#` 格式（1-3 级）
- [ ] 所有代码块都有语言标记
- [ ] 列表使用统一的 `-` 符号
- [ ] 中英文之间有空格
- [ ] 数字与中文之间有空格
- [ ] 表格格式正确
- [ ] 链接使用相对路径
- [ ] 文件名使用大写 + 下划线
- [ ] 段落和内容块之间有空行
- [ ] 无单词拼写错误

---

## 📚 参考链接

- [CommonMark 规范](https://spec.commonmark.org/)
- [GitHub Flavored Markdown](https://github.github.com/gfm/)
- [中文文案排版指北](https://github.com/sparanoid/chinese-copywriting-guidelines)

---

**版本历史**：
- v1.0 (2026-03-22) - 初版发布
