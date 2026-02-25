# 教学脚本系统 - 测试指南

## 🔧 修复内容

### 1. 服务端验证（✅ 已完成）

**位置**: [backend/game/manager.go](../../backend/game/manager.go)

**功能**: 在PlayCard函数中添加了严格的教学模式验证

```go
// 🎓 教学脚本模式：严格验证出牌是否符合当前步骤
if gameRoom.GameState.TutorialScriptMode {
    currentStep := gameRoom.GameState.TutorialCurrentStep

    // 查找当前步骤
    var currentScriptStep *TutorialStep
    for i := range tutorialSteps {
        if tutorialSteps[i].StepNumber == currentStep {
            currentScriptStep = &tutorialSteps[i]
            break
        }
    }

    if currentScriptStep != nil && currentScriptStep.Player == "human" {
        if substance != currentScriptStep.Substance {
            log.Printf("[教学脚本] ❌ 玩家尝试打出 %s，但当前步骤 %d 要求打出 %s",
                substance, currentStep, currentScriptStep.Substance)
            return errors.New(fmt.Sprintf("请按照教程出牌，当前步骤应打出 %s", currentScriptStep.Substance))
        }
        log.Printf("[教学脚本] ✅ 玩家正确打出 %s (步骤 %d)", substance, currentStep)
    }
}
```

**效果**:
- ✅ 阻止玩家打出错误的牌
- ✅ 返回明确的错误提示："请按照教程出牌，当前步骤应打出 XXX"
- ✅ 前端会显示Toast错误提示

### 2. 步骤自动递增（✅ 已完成）

**位置**: [backend/game/manager.go](../../backend/game/manager.go)

**功能**: PlayCard成功后自动递增教学步骤

```go
// 🎓 教学脚本模式：递增步骤
if gameRoom.GameState.TutorialScriptMode {
    gameRoom.GameState.TutorialCurrentStep++
    log.Printf("[教学脚本] 📈 步骤递增至 %d", gameRoom.GameState.TutorialCurrentStep)
}
```

**效果**:
- ✅ 玩家打牌成功后，步骤自动+1
- ✅ AI能获得正确的步骤号执行对应操作
- ✅ 前端UI步骤指示器自动更新

### 3. AI增强日志（✅ 已完成）

**位置**: [backend/game/ai_controller.go](../../backend/game/ai_controller.go)

**功能**: executeTutorialScript函数增加详细日志

```go
func (gr *GameRoom) executeTutorialScript() {
    log.Printf("[教学脚本] 🎯 AI触发教学脚本执行")

    currentStep := gr.GameState.TutorialCurrentStep
    log.Printf("[教学脚本] 📊 当前步骤: %d", currentStep)

    // ...

    currentPlayer := gr.GameState.Players[gr.GameState.CurrentPlayer]
    log.Printf("[教学脚本] 当前AI玩家: %s (UID:%d), 手牌数: %d",
        currentPlayer.Nickname, currentPlayer.UID, currentPlayer.CardCount)
}
```

**效果**:
- ✅ 可以清楚看到AI何时被触发
- ✅ 可以看到AI当前步骤和手牌信息
- ✅ 便于定位AI不出牌的问题

## 🧪 测试流程

### 准备工作

1. **启动后端**
```bash
cd d:/SystemFolders/Desktop/chemistryuno
./backend.exe
```

2. **启动前端**
```bash
cd frontend
npm run dev
```

3. **打开浏览器** http://localhost:5173

### 测试步骤

#### 测试1: 创建教学关卡

1. 登录/注册账号
2. 首次进入会触发大厅指引
3. 完成指引后，点击"AI实验室"
4. 自动创建教学关卡

**预期后端日志**:
```
[教学脚本] 启用脚本化教学模式，使用固定手牌和初始配置
[教学脚本] 开始初始化教学关卡...
[教学脚本] 👤 人类玩家: Player_XXX (UID:123), 手牌: [Na Mg O H Au Ar +2]
[教学脚本] 🤖 AI玩家: 门捷列夫 (UID:-1), 手牌: [H Cl Br Al Fe Zn K]
[教学脚本] 🎴 初始场上牌: Cl2
[教学脚本] ✅ 教学关卡初始化完成
```

**预期前端显示**:
- ✅ 手牌区显示: Na, Mg, O, H, Au, Ar, +2
- ✅ 场上显示: Cl₂
- ✅ 显示欢迎Toast: "欢迎来到脚本化教学关卡！"
- ✅ 显示进度条: Step 1/8 (12.5%)
- ✅ 显示提示: "第一步：从手牌中选择 **Mg**（镁）打出..."

#### 测试2: 玩家出错牌（服务端验证）

1. 在步骤1时，尝试点击 **Na**（错误的牌）

**预期行为**:
- ❌ 后端返回错误: "请按照教程出牌，当前步骤应打出 Mg"
- ❌ 前端显示Toast警告
- ❌ 牌不会被打出
- ✅ 手牌保持不变
- ✅ 步骤仍然是1

**预期后端日志**:
```
[教学脚本] ❌ 玩家尝试打出 Na，但当前步骤 1 要求打出 Mg
```

#### 测试3: 玩家出正确牌

1. 点击 **Mg**（正确的牌）

**预期行为**:
- ✅ 后端接受出牌
- ✅ Mg被打出到场上
- ✅ 步骤递增到2
- ✅ 前端进度条更新: Step 2/8 (25%)
- ✅ 提示更新: "⚗️ AI 的回合"

**预期后端日志**:
```
[教学脚本] ✅ 玩家正确打出 Mg (步骤 1)
[教学脚本] 📈 步骤递增至 2
```

#### 测试4: AI自动出牌

**预期行为**:
- ⏱️ 1秒后AI自动行动
- ✅ AI打出 HCl
- ✅ 步骤递增到3
- ✅ 场上显示: HCl
- ✅ 前端进度条更新: Step 3/8 (37.5%)
- ✅ 提示更新: "第二步：使用 **Na** 和 **O**、**H** 合成 **NaOH**..."

**预期后端日志**:
```
[AI] 🤖 AI -1 立即行动...
[教学脚本] 🎯 AI触发教学脚本执行
[教学脚本] 📊 当前步骤: 2
[教学脚本] 🤖 执行步骤 2: AI play HCl
[教学脚本] 当前AI玩家: 门捷列夫 (UID:-1), 手牌数: 7
[教学脚本] ✅ AI打出了 HCl
[教学脚本] 📈 步骤递增至 3
```

#### 测试5: 玩家合成化合物

1. 打开化学键盘
2. 输入 **NaOH**
3. 点击确认

**预期行为**:
- ✅ 消耗 Na + O + H 三张牌
- ✅ 打出 NaOH 到场上
- ✅ 显示反应方程式: NaOH + HCl → NaCl + H₂O
- ✅ 步骤递增到4

#### 测试6: 完整流程

按照8步流程完整测试：

| 步骤 | 玩家 | 操作 | 验证点 |
|------|------|------|--------|
| 1 | 玩家 | 打出Mg | ✅ 可以打，其他牌被阻止 |
| 2 | AI | 自动打HCl | ✅ 1秒后自动执行 |
| 3 | 玩家 | 合成NaOH | ✅ 消耗Na+O+H，显示中和反应 |
| 4 | AI | 自动打Br₂ | ✅ 1秒后自动执行 |
| 5 | 玩家 | 打出Ar | ✅ 稀有气体，不反应 |
| 6 | AI | 自动摸牌 | ✅ AI手牌+1 |
| 7 | 玩家 | 打出Au | ✅ 惰性金属 |
| 8 | 玩家 | 打出+2 | ✅ 特殊牌，教学完成 |

## 🔍 故障排查

### 问题1: AI不出牌

**检查后端日志**:
```bash
# 搜索教学脚本相关日志
grep "\[教学脚本\]" backend.log
grep "\[AI\]" backend.log
```

**可能原因**:
1. ❌ TutorialScriptMode未设置 → 检查initTutorialGame是否被调用
2. ❌ AI手牌中没有目标牌 → 检查固定手牌分配是否正确
3. ❌ 步骤号不匹配 → 检查TutorialCurrentStep是否正确递增

### 问题2: 玩家可以打错牌

**检查**:
1. ❌ 后端是否有服务端验证的日志
2. ❌ 前端是否绕过了验证

**确认**:
```bash
# 查看PlayCard验证日志
grep "玩家尝试打出" backend.log
```

### 问题3: 步骤不递增

**检查后端日志**:
```bash
grep "步骤递增至" backend.log
```

**确认**:
- ✅ PlayCard成功返回后应该有递增日志
- ✅ AI执行脚本后应该有递增日志

## 📊 成功指标

### 后端日志完整流程

```
[教学脚本] 启用脚本化教学模式
[教学脚本] ✅ 教学关卡初始化完成
[教学脚本] ✅ 玩家正确打出 Mg (步骤 1)
[教学脚本] 📈 步骤递增至 2
[教学脚本] 🎯 AI触发教学脚本执行
[教学脚本] ✅ AI打出了 HCl
[教学脚本] 📈 步骤递增至 3
[教学脚本] ✅ 玩家正确打出 NaOH (步骤 3)
[教学脚本] 📈 步骤递增至 4
... (继续直到步骤8)
```

### 前端UI完整流程

- ✅ 进度条从12.5%增长到100%
- ✅ 步骤显示从Step 1/8到Step 8/8
- ✅ 错误提示Toast正常显示
- ✅ AI操作1秒后自动执行
- ✅ 化学反应方程式正确显示

---

**更新时间**: 2026-02-24
**版本**: Chemistry UNO V1.2.0 Mendeleef
**状态**: ✅ 修复完成，准备测试
**修复内容**:
1. ✅ 添加服务端PlayCard验证
2. ✅ 添加教学步骤自动递增
3. ✅ 增强AI脚本执行日志
