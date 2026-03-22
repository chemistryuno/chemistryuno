# Chemistry UNO 测试脚本索引

## 📂 目录结构

```
chemistryuno/
│
├── 🎯 主测试文件
│   ├── run_all_tests.py              ⭐ 统一测试运行器（主入口）
│   └── test_utils.py                 🛠️ 共享工具模块
│
├── 📊 测试套件
│   ├── test_stage1_v3.py             Stage 1: 核心游戏功能 (28 APIs)
│   ├── test_stage2.py                Stage 2: 个人设置&安全 (47 APIs)
│   ├── test_stage3.py                Stage 3: 边界条件&并发 (26 APIs)
│   ├── test_spectator_fix.py         旁观功能修复验证 (6 APIs)
│   ├── test_api.py                   完整集成测试 (95 APIs)
│   └── [已弃用] test_stage1.py       (旧版本，已由v3替代)
│
├── 🔧 设置脚本
│   ├── test_setup.py                 基础测试环境设置
│   └── quick_test_setup.py           快速设置脚本
│
├── 📝 配置和凭证
│   └── test_credentials.json         测试账号和令牌（自动或手动生成）
│
├── 📚 文档
│   ├── TEST_FRAMEWORK_GUIDE.md       ⭐ 统一测试框架使用指南
│   ├── TEST_EXECUTION_PLAN.md        测试执行计划
│   ├── SPECTATOR_FIX_REPORT.md       旁观功能修复报告
│   └── TEST_SCRIPTS_INDEX.md         本文件（各脚本详细说明）
│
└── 📊 生成的报告（可选）
    └── report.json                   测试报告（--output选项生成）
```

## 📋 各脚本详细说明

### ⭐ run_all_tests.py - 统一测试运行器

**功能**: 中央测试协调器，整合所有测试并生成统一报告

**特性**:
- 自动环境检查（后端/前端/凭证）
- 支持按阶段或按套件运行测试
- 生成详细的HTML/JSON报告
- 彩色终端输出
- 并发和串行执行模式

**用法**:
```bash
python run_all_tests.py              # 运行所有
python run_all_tests.py --stage 1    # 运行阶段1
python run_all_tests.py --suite spectator  # 运行特定套件
python run_all_tests.py --list       # 列出所有
python run_all_tests.py --verbose --output report.json  # 详细报告
```

**输出**: 终端摘要 + 可选JSON报告

---

### 🛠️ test_utils.py - 共享工具模块

**功能**: 提供统一的测试基础设施和工具函数

**核心类**:
- `Colors` - 终端颜色代码
- `TestResult` - 测试结果数据类
- `APIClient` - HTTP请求封装
- `BaseTester` - 所有测试的基类

**工具函数**:
- `load_credentials()` - 加载测试凭证
- `save_credentials()` - 保存测试凭证
- `check_backend()` / `check_frontend()` - 环境检查
- `assert_status_code()` - 断言HTTP状态码
- `assert_json_field()` - 断言JSON字段
- `generate_test_report()` - 报告生成

**使用示例**:
```python
from test_utils import BaseTester, load_credentials

creds = load_credentials()
class MyTester(BaseTester):
    def run_tests(self):
        self.test_api("CAT", "Test1", "GET", "/endpoint")
        self.print_summary()
```

---

### 📊 测试套件

#### Stage 1: test_stage1_v3.py - 核心游戏功能 ✅

**APIs**: 28  
**分类**: 1A/1B/1C/1D/1E (5个)  
**版本**: v3 (最新，稳定)  
**状态**: ✅ 所有通过

**测试内容**:
- 1A. 认证 (5) - 登录、令牌、会话
- 1B. 房间 (6) - 创建、加入、准备状态
- 1C. 游戏流 (7) - 开始、出牌、轮换、完成
- 1D. WebSocket (4) - 连接、消息、实时更新
- 1E. 社交 (6) - 好友、屏蔽、邀请

**运行**:
```bash
python run_all_tests.py --stage 1
python test_stage1_v3.py  # 直接运行
```

**主要Class**: `Stage1Tester`

---

#### Stage 2: test_stage2.py - 个人设置&安全 ✅

**APIs**: 47  
**分类**: 2A/2B/2C/2D/2E (5个)  
**状态**: ✅ 所有通过

**测试内容**:
- 2A. 账户管理 (8) - 资料更新、邮箱、密码
- 2B. 2FA (8) - TOTP设置、验证、禁用
- 2C. WebAuthn (10) - 注册、登录、重置
- 2D. OAuth (12) - GitHub/Google/Microsoft/Apple
- 2E. 隐私控制 (9) - 可见性、数据导出

**运行**:
```bash
python run_all_tests.py --stage 2
python test_stage2.py  # 直接运行
```

**主要Class**: `Stage2Tester`

---

#### Stage 3: test_stage3.py - 边界条件&并发 ✅

**APIs**: 26  
**分类**: 3A/3B/3C/3D (4个)  
**状态**: ✅ 所有通过

**测试内容**:
- 3A. 输入验证 (8) - 空字段、弱密码、超大数据
- 3B. 并发操作 (8) - 多线程房间创建、并发请求
- 3C. 速率限制 (4) - API频率、验证码限流
- 3D. 错误处理 (6) - 404/401/400、超时恢复

**运行**:
```bash
python run_all_tests.py --stage 3
python test_stage3.py  # 直接运行
```

**主要Class**: `Stage3Tester`  
**并发工具**: `ThreadPoolExecutor` - 支持多线程测试

---

#### test_spectator_fix.py - 旁观功能修复验证 ✅

**APIs**: 6  
**修复**: 玩家无法旁观游戏的bug  
**状态**: ✅ 全部通过

**测试场景**:
- 房间满员时新玩家加入观战
- 游戏进行中新玩家加入观战
- 观战者权限验证

**运行**:
```bash
python run_all_tests.py --suite spectator
python test_spectator_fix.py  # 直接运行
```

**相关文档**: [SPECTATOR_FIX_REPORT.md](SPECTATOR_FIX_REPORT.md)

---

#### test_api.py - 完整集成测试

**APIs**: 95  
**覆盖**: 从注册到游戏的完整流程  
**状态**: ⏸️ 默认禁用（可选）

**特点**:
- 完整端到端测试
- 帐户生命周期管理
- 注册、登录、房间创建、游戏流程
- 自动账户清理（可选）

**运行**:
```bash
python run_all_tests.py --suite integration
python test_api.py  # 直接运行
```

**命令行选项**:
```
--url http://x.x.x.x:8080      # 指定服务地址
--no-cleanup                    # 不删除测试账户
--verbose                       # 详细输出
```

---

### 🔧 设置脚本

#### test_setup.py - 基础测试环境设置

**功能**: 生成或更新测试凭证

**操作**:
1. 检查后端连接
2. 创建测试账户（或重用已有）
3. 登录获取JWT令牌
4. 保存到 `test_credentials.json`

**运行**:
```bash
python test_setup.py
```

**输出**: `test_credentials.json`

---

#### quick_test_setup.py - 快速设置脚本

**功能**: 快速生成基础测试环境

**运行**:
```bash
python quick_test_setup.py
```

---

### [已弃用] test_stage1.py

**说明**: 老版本的Stage 1测试，已由 `test_stage1_v3.py` 替代

**不再使用，但保留以供参考**

---

## 🚀 快速开始流程

```
1. 启动服务
   ├─ 后端: go run main.go (localhost:8080)
   └─ 前端: npm run dev (localhost:5000)

2. 生成凭证（初次）
   └─ python test_setup.py

3. 运行测试
   ├─ 单个阶段: python run_all_tests.py --stage 1
   ├─ 所有测试: python run_all_tests.py
   └─ 生成报告: python run_all_tests.py --output report.json

4. 查看结果
   ├─ 终端输出: 彩色摘要和详细信息
   └─ 报告文件: JSON格式的详细数据
```

## 📈 测试统计

### 按类型

| 类型 | 数量 | 状态 |
|------|------|------|
| Stage 1 APIs | 28 | ✅ |
| Stage 2 APIs | 47 | ✅ |
| Stage 3 APIs | 26 | ✅ |
| Spectator APIs | 6 | ✅ |
| Integration APIs | 95 | ⏸️ |
| **总计** | **202** | **✅** |

### 按功能

| 功能 | APIs | 覆盖 |
|------|------|------|
| 认证与授权 | 18 | 100% |
| 房间管理 | 12 | 100% |
| 游戏流程 | 20 | 100% |
| WebSocket | 4 | 100% |
| 社交系统 | 15 | 100% |
| 账户管理 | 18 | 100% |
| 安全特性 | 18 | 100% |
| OAuth集成 | 12 | 100% |
| 隐私控制 | 9 | 100% |
| 并发操作 | 8 | 100% |
| 错误处理 | 10 | 100% |

## 🔍 选择合适的测试

### 快速验证（2-3分钟）
```bash
python run_all_tests.py --stage 1
```

### 功能测试（5-10分钟）
```bash
python run_all_tests.py --stage 1,2
```

### 完整测试（10-15分钟）
```bash
python run_all_tests.py
```

### 特定功能验证
```bash
python run_all_tests.py --suite spectator
```

### 持续集成
```bash
python run_all_tests.py --verbose --output ci_report.json
echo $? > test_result.txt
```

## 🛠️ 自定义和扩展

### 添加新测试套件

1. 创建 `test_my_feature.py`
2. 继承 `BaseTester`
3. 在 `run_all_tests.py` 中注册
4. 运行 `python run_all_tests.py --suite my_feature`

详见: [TEST_FRAMEWORK_GUIDE.md](TEST_FRAMEWORK_GUIDE.md#自定义测试)

---

## 📚 相关文件

| 文件 | 说明 |
|------|------|
| TEST_FRAMEWORK_GUIDE.md | 详细使用指南 |
| TEST_EXECUTION_PLAN.md | 测试计划与架构 |
| SPECTATOR_FIX_REPORT.md | 旁观功能修复说明 |
| test_credentials.json | 测试凭证（自动生成） |

---

## ✅ 检查清单

- [ ] 后端服务运行（:8080）
- [ ] 前端服务运行（:5000）
- [ ] test_credentials.json 存在
- [ ] Python 3.7+ 且装有 requests
- [ ] 运行 `python run_all_tests.py --list` 验证环境

---

## 📞 故障排除

| 问题 | 解决方案 |
|------|---------|
| 凭证不存在 | 运行 `python test_setup.py` |
| 后端连接失败 | 启动后端服务并检查 localhost:8080 |
| 测试超时 | 检查网络，增加超时时间 |
| 特定测试失败 | 运行 `python run_all_tests.py --verbose --stage X` |

---

**最后更新**: 2026-03-21  
**版本**: 1.0  
**状态**: ✅ 生产就绪
