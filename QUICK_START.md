# 🚀 Chemistry UNO 测试框架 - 快速起步

## 5分钟快速开始

### 1️⃣ 准备环境

```bash
# 确保后端和前端都在运行
# 后端: localhost:8080
# 前端: localhost:5000

# 安装依赖（如果未安装）
pip install requests
```

### 2️⃣ 获取测试凭证

```bash
# 首次运行 - 生成测试凭证
python test_setup.py
```

或者直接创建 `test_credentials.json`：

```json
{
  "player1": {"username": "p95484", "uid": 100000011, "token": "..."},
  "player2": {"username": "q95484", "uid": 100000012, "token": "..."},
  "backend": "http://localhost:8080/api",
  "frontend": "http://localhost:5000"
}
```

### 3️⃣ 运行测试

#### 最简单的方式（运行全部）

```bash
python run_all_tests.py
```

#### 按功能选择

```bash
# 仅核心功能（最快）
python run_all_tests.py --stage 1

# 核心+设置
python run_all_tests.py --stage 1,2

# 所有测试
python run_all_tests.py

# 或运行特定套件
python run_all_tests.py --suite spectator
```

---

## 📊 输出示例

```
=======================================================================
Chemistry UNO 统一测试框架
=======================================================================

✓ 后端服务运行中
✓ 前端服务运行中  
✓ 加载凭证: p95484, q95484

▶ 运行 阶段1 - 核心游戏功能 (test_stage1_v3.py)
  ✓ 用户信息
  ✓ 会话列表
  ✓ 登录
  ... (28项测试)
✓ stage1: 28 通过

... (其他阶段) ...

───────────────────────────────────────────────────────────────────
📊 测试汇总报告
───────────────────────────────────────────────────────────────────

总API数:        101
通过:           101
失败:           0
通过率:         100.0%
执行时间:       45.3秒

🎉 所有测试通过！(101/101 APIs)
```

---

## 🎯 常用命令速查

| 命令 | 说明 | 时间 |
|------|------|------|
| `python run_all_tests.py` | 运行所有测试 | ~1分钟 |
| `python run_all_tests.py --stage 1` | 快速验证 | ~10秒 |
| `python run_all_tests.py --list` | 列出所有测试 | 立即 |
| `python run_all_tests.py --verbose` | 详细输出 | ~1分钟 |
| `python run_all_tests.py --output report.json` | 保存报告 | ~1分钟 |

---

## ⚙️ 配置选项

```bash
# 自定义API地址
python run_all_tests.py --url http://192.168.1.100:8080/api

# 详细输出
python run_all_tests.py --verbose

# 保存JSON报告
python run_all_tests.py --output report.json

# 组合使用
python run_all_tests.py --stage 1,2 --verbose --output report.json
```

---

## 📋 测试套件速览

```
┌─────────────────────┬─────┬─────────────────┐
│ 名称                │ APIs │ 典型时间         │
├─────────────────────┼─────┼─────────────────┤
│ Stage 1: 核心功能   │ 28  │ 10-15秒         │
│ Stage 2: 个人设置   │ 47  │ 20-30秒         │
│ Stage 3: 并发/边界  │ 26  │ 15-20秒         │
│ 旁观功能验证        │ 6   │ 5-10秒          │
│ 完整集成测试        │ 95  │ 45-60秒 (可选)  │
├─────────────────────┼─────┼─────────────────┤
│ 总计                │ 202 │ ~1分钟          │
└─────────────────────┴─────┴─────────────────┘
```

---

## 🔍 快速排查

### 问题：凭证文件不存在

```bash
# ❌ 错误信息
Error: 凭证文件不存在

# ✅ 解决方案
python test_setup.py
```

### 问题：后端未运行

```bash
# ❌ 错误信息
Error: 后端服务未运行（localhost:8080）

# ✅ 解决方案
# 在另一个终端启动后端
cd backend
go run main.go
```

### 问题：测试失败

```bash
# ❌ 某个测试失败
stage1: 22 通过, 6 失败

# ✅ 查看详细信息
python run_all_tests.py --stage 1 --verbose
```

### 问题：连接超时

```bash
# ❌ 错误信息
Timeout error

# ✅ 检查网络和服务
curl http://localhost:8080/api/health
```

---

## 📈 生成报告

### 保存为JSON

```bash
python run_all_tests.py --output report.json
```

### 查看报告

```bash
# Windows
type report.json

# Linux/Mac
cat report.json
```

### 报告内容示例

```json
{
  "timestamp": "2026-03-21T10:30:45",
  "total_apis": 101,
  "total_passed": 101,
  "total_failed": 0,
  "pass_rate": 100.0,
  "duration": 45.3,
  "results": [...]
}
```

---

## 🏃 高效工作流

### 日常开发

```bash
# 1. 早晨快速检查（30秒）
python run_all_tests.py --stage 1

# 2. 修改后完整验证（1分钟）
python run_all_tests.py

# 3. 提交前最终检查
python run_all_tests.py --stage 3  # 边界和并发
```

### 特定功能测试

```bash
# 只验证旁观功能
python run_all_tests.py --suite spectator

# 验证个人设置
python run_all_tests.py --stage 2
```

### CI/CD集成

```bash
# 保存报告用于分析
python run_all_tests.py --output ci_report.json

# 检查是否全部通过
if python run_all_tests.py > /dev/null 2>&1; then
    echo "Tests passed!"
    exit 0
else
    echo "Tests failed!"
    exit 1
fi
```

---

## 🎓 学习路径

### 初级用户

1. 运行 `python run_all_tests.py --list` 了解测试
2. 运行 `python run_all_tests.py --stage 1` 看基本功能
3. 查看 [TEST_FRAMEWORK_GUIDE.md](TEST_FRAMEWORK_GUIDE.md)

### 中级用户

1. 了解各个Stage的测试内容
2. 使用 `--verbose` 查看详细信息
3. 学习如何修改和扩展测试

### 高级用户

1. 创建自定义测试套件
2. 集成到CI/CD流程
3. 分析测试报告数据

---

## 💡 技巧

### 1. 快速循环测试

```bash
# 修改代码后快速验证关键功能
python run_all_tests.py --stage 1

# 确认没有破坏其他功能
python run_all_tests.py
```

### 2. 并行运行多个测试

```bash
# 终端1 - 运行Stage 1
python run_all_tests.py --stage 1

# 终端2 - 同时运行Stage 2
python run_all_tests.py --stage 2
```

### 3. 保存和比较报告

```bash
# 修改前
python run_all_tests.py --output before.json

# 修改后
python run_all_tests.py --output after.json

# 比较
diff before.json after.json
```

### 4. 自定义输出

```bash
# 只看摘要
python run_all_tests.py 2>/dev/null | tail -20

# 看实时进度
python run_all_tests.py --verbose | grep -E "✓|✗"
```

---

## 📚 更多资源

| 资源 | 说明 |
|------|------|
| [TEST_FRAMEWORK_GUIDE.md](TEST_FRAMEWORK_GUIDE.md) | 完整框架指南 |
| [TEST_SCRIPTS_INDEX.md](TEST_SCRIPTS_INDEX.md) | 各脚本详细说明 |
| [TEST_EXECUTION_PLAN.md](TEST_EXECUTION_PLAN.md) | 测试计划 |
| [test_utils.py](test_utils.py) | 工具代码 |
| [run_all_tests.py](run_all_tests.py) | 运行器代码 |

---

## ✨ 下一步

- ✅ 运行基础测试验证环境
- 📖 阅读 [TEST_FRAMEWORK_GUIDE.md](TEST_FRAMEWORK_GUIDE.md) 了解更多
- 🔧 支持新功能时创建新的测试
- 📊 集成到CI/CD流程

---

**需要帮助？** 查看完整文档或直接运行 `python run_all_tests.py --help`

**版本**: 1.0 | **状态**: ✅ 生产就绪
