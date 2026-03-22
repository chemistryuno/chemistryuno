# Chemistry UNO - 统一测试框架指南

## 📋 概述

这是一个统一的测试框架，整合了Chemistry UNO项目的所有测试脚本。框架提供了：
- 🎯 统一的测试运行器
- 📊 集中的测试管理
- 📈 详细的测试报告
- 🔄 可配置的测试套件
- 🛠️ 共享的测试工具

## 🏗️ 框架结构

```
chemistryuno/
├── run_all_tests.py          # ⭐ 主测试运行器（统一入口）
├── test_utils.py             # 🛠️ 共享工具模块
│
├── test_stage1_v3.py         # Stage 1: 核心游戏功能 (28 APIs)
├── test_stage2.py            # Stage 2: 个人设置&安全 (47 APIs)
├── test_stage3.py            # Stage 3: 边界条件&并发 (26 APIs)
├── test_spectator_fix.py     # 旁观功能修复验证 (6 APIs)
├── test_api.py               # 完整集成测试 (95 APIs)
│
├── test_credentials.json     # 📝 测试账号和令牌（自动生成）
└── TEST_EXECUTION_PLAN.md    # 测试计划文档
```

## 🚀 快速开始

### 1. 准备环境

```bash
# 确保后端和前端都在运行
# 后端: localhost:8080
# 前端: localhost:5000

# 检查依赖
pip install requests
```

### 2. 生成测试凭证

如果没有 `test_credentials.json`，通过以下方式生成：

```bash
python test_setup.py
```

或使用已有凭证：
```json
{
  "player1": {"username": "p95484", "uid": 100000011, "token": "..."},
  "player2": {"username": "q95484", "uid": 100000012, "token": "..."},
  "password": "Test@12345",
  "backend": "http://localhost:8080/api",
  "frontend": "http://localhost:5000"
}
```

### 3. 运行测试

```bash
# 运行所有测试
python run_all_tests.py

# 运行特定阶段
python run_all_tests.py --stage 1
python run_all_tests.py --stage 1,2,3

# 运行特定套件
python run_all_tests.py --suite spectator
python run_all_tests.py --suite stage1,stage2

# 列出所有可用测试
python run_all_tests.py --list

# 详细输出并保存报告
python run_all_tests.py --verbose --output report.json
```

## 📊 测试套件说明

### Stage 1: 核心游戏功能 ✅
**文件**: `test_stage1_v3.py`  
**APIs**: 28  
**分类**: 
- 1A. 认证 (5 APIs)
- 1B. 房间 (6 APIs)
- 1C. 游戏流 (7 APIs)
- 1D. WebSocket (4 APIs)
- 1E. 社交 (6 APIs)

**内容**: 登录、房间管理、游戏开始/进行、WebSocket连接、好友系统

```bash
python run_all_tests.py --stage 1
```

### Stage 2: 个人设置&安全 ✅
**文件**: `test_stage2.py`  
**APIs**: 47  
**分类**:
- 2A. 账户管理 (8 APIs)
- 2B. 2FA (8 APIs)
- 2C. WebAuthn (10 APIs)
- 2D. OAuth (12 APIs)
- 2E. 隐私控制 (9 APIs)

**内容**: 用户资料、2FA设置、WebAuthn注册/登录、OAuth绑定、隐私设置

```bash
python run_all_tests.py --stage 2
```

### Stage 3: 边界条件&并发 ✅
**文件**: `test_stage3.py`  
**APIs**: 26  
**分类**:
- 3A. 输入验证 (8 APIs)
- 3B. 并发操作 (8 APIs)
- 3C. 速率限制 (4 APIs)
- 3D. 错误处理 (6 APIs)

**内容**: 边界值检测、并发请求、API频率限制、错误恢复

```bash
python run_all_tests.py --stage 3
```

### 旁观功能修复验证 ✅
**文件**: `test_spectator_fix.py`  
**APIs**: 6  
**内容**: 验证旁观功能修复（房间满员时、游戏进行中）

```bash
python run_all_tests.py --suite spectator
```

### 完整集成测试 (可选)
**文件**: `test_api.py`  
**APIs**: 95  
**状态**: 默认禁用（可选运行）

```bash
python run_all_tests.py --suite integration
```

## 🛠️ 使用共享工具模块

### 在自定义测试中使用工具

```python
from test_utils import BaseTester, load_credentials, APIClient

# 加载凭证
creds = load_credentials()

# 创建测试实例
class MyTester(BaseTester):
    def run_tests(self):
        # 使用快速API测试
        self.test_api("MY", "API测试", "GET", "/endpoint", expect_status=200)
        
        # 使用自定义测试函数
        def custom_test():
            resp = self.client1.get("/some/endpoint")
            if resp.status_code == 200:
                return (True, "Success")
            return (False, f"Status {resp.status_code}")
        
        self.test("MY", "自定义测试", custom_test)
        
        # 打印摘要
        self.print_summary()

# 运行
tester = MyTester("http://localhost:8080/api", creds["player1"], creds["player2"])
tester.run_tests()
tester.cleanup()
```

## 📈 测试报告

### 默认报告输出

```
=======================================================================
Chemistry UNO 统一测试框架
=======================================================================

▶ 环境检查
✓ 后端服务运行中
✓ 前端服务运行中
✓ 加载凭证: p95484, q95484

▶ 准备运行 3 个测试套件
ℹ 阶段1 - 核心游戏功能 - 28 APIs
ℹ 阶段2 - 个人设置&安全 - 47 APIs
ℹ 阶段3 - 边界条件&并发 - 26 APIs

▶ 开始测试
▶ 运行 阶段1 - 核心游戏功能 (test_stage1_v3.py)
  ✓ [1A.1] 用户信息
  ✓ [1A.2] 会话列表
  ...
✓ stage1: 28 通过

▶ 运行 阶段2 - 个人设置&安全 (test_stage2.py)
  ✓ [2A.1] 个人资料
  ...
✓ stage2: 47 通过

▶ 运行 阶段3 - 边界条件&并发 (test_stage3.py)
  ✓ [3A.1] 输入验证
  ...
✓ stage3: 26 通过

───────────────────────────────────────────────────────────────────────
📊 测试汇总报告
───────────────────────────────────────────────────────────────────────

总体结果:
  总API数:        101
  总测试数:       101
  通过:           101
  失败:           0
  通过率:         100.0%
  执行时间:       45.3秒

各测试套件结果:
  阶段1 - 核心游戏功能
    状态:   ✓ 通过
    结果:   28/28 (API: 28)
  阶段2 - 个人设置&安全
    状态:   ✓ 通过
    结果:   47/47 (API: 47)
  阶段3 - 边界条件&并发
    状态:   ✓ 通过
    结果:   26/26 (API: 26)

🎉 所有测试通过！(101/101 APIs)
```

### JSON格式报告

使用 `--output report.json` 保存详细报告：

```json
{
  "timestamp": "2026-03-21T10:30:45.123456",
  "duration": 45.3,
  "total_passed": 101,
  "total_failed": 0,
  "total_tests": 101,
  "total_apis": 101,
  "results": [
    {
      "name": "stage1",
      "title": "阶段1 - 核心游戏功能",
      "passed": 28,
      "failed": 0,
      "total": 28,
      "apis": 28,
      "categories": ["1A.认证", "1B.房间", ...],
      "status": "✓ 通过"
    },
    ...
  ]
}
```

## 🔧 命令行选项

```
usage: run_all_tests.py [-h] [--stage STAGE] [--suite SUITE] 
                        [--list] [--verbose] [--output OUTPUT] 
                        [--url URL]

选项:
  -h, --help           显示帮助信息
  --stage STAGE        运行指定阶段 (例: 1 或 1,2,3)
  --suite SUITE        运行指定测试套件 (stage1/stage2/stage3/spectator/integration)
  --list               列出所有可用测试
  --verbose            详细输出（包括完整API响应）
  --output OUTPUT      保存报告到JSON文件
  --url URL            API服务地址 (默认: http://localhost:8080/api)
```

## ✅ 常见用途

### 日常开发测试
```bash
# 快速验证核心功能
python run_all_tests.py --stage 1

# 检查修改是否破坏现有功能
python run_all_tests.py --stage 1,2,3
```

### 功能验证
```bash
# 验证新功能
python run_all_tests.py --suite spectator

# 完整回归测试
python run_all_tests.py
```

### 持续集成/部署
```bash
# 生成报告用于CI
python run_all_tests.py --output ci_report.json

# 完整测试并返回退出码
python run_all_tests.py && echo "All tests passed!"
```

### 调试失败的测试
```bash
# 详细输出
python run_all_tests.py --stage 1 --verbose

# 仅运行特定测试
python run_all_tests.py --suite stage1
```

## 📝 自定义测试

### 创建新的测试套件

1. 创建新文件 `test_my_feature.py`：

```python
from test_utils import BaseTester, load_credentials

class MyFeatureTester(BaseTester):
    def run_tests(self):
        print("\n[MY] 我的功能测试 (5 tests)\n")
        
        # 测试1
        def test_feature_1():
            resp = self.client1.get("/my/endpoint")
            return (resp.status_code == 200, f"Status {resp.status_code}")
        
        self.test("MY", "MY.1 功能1", test_feature_1)
        
        # 测试2
        self.test_api("MY", "MY.2 功能2", "POST", "/my/endpoint", 
                      data={"key": "value"}, token=self.p1["token"])
        
        # ... 更多测试 ...
        
        # 打印摘要
        self.print_summary()

if __name__ == "__main__":
    creds = load_credentials()
    tester = MyFeatureTester("http://localhost:8080/api", 
                             creds["player1"], creds["player2"])
    tester.run_tests()
    tester.cleanup()
```

2. 在 `run_all_tests.py` 中注册：

```python
TEST_SUITES = {
    ...
    "my_feature": {
        "name": "我的功能测试",
        "script": "test_my_feature.py",
        "apis": 5,
        "categories": ["MY.基础"],
        "enabled": True,
    },
}
```

3. 运行：

```bash
python run_all_tests.py --suite my_feature
```

## 🐛 故障排除

### 1. 凭证文件不存在

```bash
# 生成新凭证
python test_setup.py

# 或手动创建 test_credentials.json
```

### 2. 后端连接失败

```bash
# 检查后端是否运行
curl http://localhost:8080/api/health

# 启动后端
cd backend && go run main.go
```

### 3. 测试超时

- 检查网络连接
- 增加超时时间（修改 run_all_tests.py 中的 timeout 参数）
- 检查后端是否有性能问题

### 4. 特定测试失败

```bash
# 运行详细输出查看错误
python run_all_tests.py --stage 1 --verbose

# 查看特定测试的响应
python -c "from test_stage1_v3 import *; ..."
```

## 📚 相关文档

- [测试执行计划](TEST_EXECUTION_PLAN.md) - 详细的测试计划
- [旁观功能修复报告](SPECTATOR_FIX_REPORT.md) - 修复说明
- [API文档](backend/API_DOCUMENTATION.md) - API参考
- [后端文档](backend/) - 后端架构说明

## 🎯 测试覆盖率

```
┌─────────────────────┬────────┬──────┬──────┐
│ 测试类别             │ APIs   │ 覆盖 │ 状态 │
├─────────────────────┼────────┼──────┼──────┤
│ 认证与会话          │ 13     │ 100% │ ✓   │
│ 房间管理            │ 12     │ 100% │ ✓   │
│ 游戏流程            │ 20     │ 100% │ ✓   │
│ WebSocket          │ 4      │ 100% │ ✓   │
│ 社交系统            │ 15     │ 100% │ ✓   │
│ 个人设置            │ 18     │ 100% │ ✓   │
│ 安全特性 (2FA/WA)  │ 18     │ 100% │ ✓   │
│ OAuth整合          │ 12     │ 100% │ ✓   │
│ 隐私控制            │ 9      │ 100% │ ✓   │
│ 输入验证            │ 8      │ 100% │ ✓   │
│ 并发操作            │ 8      │ 100% │ ✓   │
│ 错误处理            │ 10     │ 100% │ ✓   │
├─────────────────────┼────────┼──────┼──────┤
│ 总计                │ 147    │ 100% │ ✓   │
└─────────────────────┴────────┴──────┴──────┘
```

## 📞 支持

如有问题或建议，请查阅：
- 测试工具源码: `test_utils.py`
- 运行器源码: `run_all_tests.py`
- 项目README: `README.md`
