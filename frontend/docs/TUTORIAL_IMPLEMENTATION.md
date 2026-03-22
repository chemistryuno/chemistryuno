# 脚本化教学系统 - 实现文档

## 🎯 功能概述

创建完全脚本化的教学关卡，玩家和AI按照固定顺序出牌，玩家必须严格按照提示操作，不能自由发挥。

## 📋 教学脚本设计

### 初始状态
- **玩家手牌**：Na, Mg, O, H, Au, Ar, +2
- **AI手牌**：H, Cl, Br, Al, Fe, Zn, K
- **场上初始牌**：Cl₂
- **先手**：玩家

### 出牌顺序

| 步骤 | 玩家 | 操作 | 物质 | 教学目标 |
|------|------|------|------|----------|
| 1 | 玩家 | 打出 | Mg | 学习金属与非金属的反应 (Mg + Cl₂ → MgCl₂) |
| 2 | AI | 打出 | HCl | AI示范酸的使用 |
| 3 | 玩家 | 打出 | NaOH | 学习碱的合成和中和反应 (NaOH + HCl → NaCl + H₂O) |
| 4 | AI | 打出 | Br₂ | AI示范卤素单质 |
| 5 | 玩家 | 打出 | Ar | 学习稀有气体（不反应） |
| 6 | AI | 摸牌 | - | 演示手牌为空时的操作 |
| 7 | 玩家 | 打出 | Au | 学习惰性金属（不易反应） |
| 8 | 玩家 | 打出 | +2 | 学习特殊卡牌的使用 |

## 🔧 前端实现（✅ 已完成）

### 1. 脚本配置文件

**文件位置**：[frontend/src/utils/tutorialScript.ts](../src/utils/tutorialScript.ts)

```typescript
export interface TutorialStep {
  stepNumber: number
  player: 'human' | 'ai'
  action: 'play' | 'draw' | 'double'
  substance?: string
  hint: string
  aiMessage?: string
}

export const TUTORIAL_SCRIPT: TutorialStep[] = [
  {
    stepNumber: 1,
    player: 'human',
    action: 'play',
    substance: 'Mg',
    hint: '💡 第一步：从手牌中选择 <strong>Mg</strong>（镁）打出...'
  },
  // ... 8个步骤
]
```

### 2. 状态管理 ([GameRoom.vue:41-45](../src/pages/GameRoom.vue#L41-L45))

```typescript
const isTutorialMode = ref(false)           // 教学模式开关
const tutorialHintText = ref('')            // 提示文本
const tutorialCurrentStep = ref(1)          // 当前步骤 (1-8)
const tutorialScriptMode = ref(false)       // 脚本化模式开关
```

### 3. 脚本化提示生成 ([GameRoom.vue:812-852](../src/pages/GameRoom.vue#L812-L852))

```typescript
const generateTutorialHint = () => {
  if (tutorialScriptMode.value) {
    const currentStep = getTutorialStep(tutorialCurrentStep.value)
    if (currentStep && currentStep.player === 'human') {
      tutorialHintText.value = currentStep.hint  // 显示脚本提示
    }
    return
  }
  // 原有的通用教学提示...
}
```

### 4. 出牌验证 ([GameRoom.vue:1102-1120](../src/pages/GameRoom.vue#L1102-L1120))

```typescript
const handlePlayCard = async () => {
  // 脚本化教学模式：验证是否是正确的牌
  if (tutorialScriptMode.value) {
    const currentStep = getTutorialStep(tutorialCurrentStep.value)
    if (currentStep && currentStep.player === 'human') {
      if (selectedSubstance.value !== currentStep.substance) {
        showToast(
          `请按照教学提示打出 <strong>${currentStep.substance}</strong>`,
          '⚠️ 教学模式',
          'warning'
        )
        return  // 阻止非脚本内的出牌
      }
    }
  }

  // 执行出牌...
  await gameAPI.playCard(...)

  // 脚本化教学模式：进入下一步
  if (tutorialScriptMode.value) {
    tutorialCurrentStep.value++
    generateTutorialHint()
  }
}
```

### 5. UI进度显示 ([GameRoom.vue:1987-1998](../src/pages/GameRoom.vue#L1987-L1998))

```vue
<!-- 脚本进度指示器 -->
<div v-if="tutorialScriptMode" class="relative mb-3 flex items-center justify-between">
  <div class="text-[10px] font-black text-white/80 uppercase tracking-widest">
    Step {{ tutorialCurrentStep }}/8
  </div>
  <div class="flex-1 mx-3 h-1.5 bg-white/20 rounded-full overflow-hidden">
    <div
      class="h-full bg-white/60 rounded-full transition-all duration-500"
      :style="{ width: `${(tutorialCurrentStep / 8) * 100}%` }"
    ></div>
  </div>
  <div class="text-[10px] font-bold text-white/60">
    {{ Math.round((tutorialCurrentStep / 8) * 100) }}%
  </div>
</div>
```

### 6. 欢迎提示 ([GameRoom.vue:1034-1049](../src/pages/GameRoom.vue#L1034-L1049))

```typescript
if (tutorialScriptMode.value) {
  showToast(
    '🎓 欢迎来到脚本化教学关卡！你将跟随系统指引，按照固定步骤学习游戏机制。请严格按照提示的顺序出牌。',
    '📖 教学脚本已加载',
    'success',
    9000
  )
}
```

## 🎨 视觉设计

### 提示卡片
- **渐变背景**：amber-500 → orange-500
- **进度条**：白色半透明，动态宽度
- **步骤显示**：Step X/8 + 百分比
- **提示文本**：支持HTML格式（加粗关键物质）

### 进度指示器
- **样式**：极简线性进度条
- **位置**：提示卡片顶部
- **动画**：500ms平滑过渡

## 🚧 后端实现（待开发）

### 需要的后端功能

#### 1. 创建教学房间API扩展

**位置**：`backend/handlers/game.go`

```go
type CreateRoomRequest struct {
  // ... 现有字段 ...
  TutorialScript bool `json:"tutorial_script"` // 是否启用教学脚本
}

// 创建房间时
if req.TutorialScript {
  // 设置固定手牌
  game.Players[0].Hand = []string{"Na", "Mg", "O", "H", "Au", "Ar", "+2"}
  game.Players[1].Hand = []string{"H", "Cl", "Br", "Al", "Fe", "Zn", "K"}
  game.DiscardTop = "Cl2"
  game.TutorialScriptMode = true
}
```

#### 2. AI脚本行为

**位置**：`backend/game/ai.go`

```go
func (ai *AI) makeDecision(game *Game) {
  if game.TutorialScriptMode {
    // 读取教学脚本
    step := getTutorialStep(game.TutorialCurrentStep)
    if step.Player == "ai" {
      if step.Action == "play" {
        ai.playSubstance(step.Substance)
      } else if step.Action == "draw" {
        ai.drawCard()
      }
      game.TutorialCurrentStep++
      return
    }
  }
  // 原有AI逻辑...
}
```

#### 3. 游戏状态扩展

**位置**：`backend/models/game.go`

```go
type Game struct {
  // ... 现有字段 ...
  TutorialScriptMode bool `json:"tutorial_script_mode"`
  TutorialCurrentStep int `json:"tutorial_current_step"`
}
```

#### 4. 出牌验证

**位置**：`backend/game/manager.go`

```go
func (gm *GameManager) PlayCard(...) error {
  if game.TutorialScriptMode {
    step := getTutorialStep(game.TutorialCurrentStep)
    if step.Player == "human" && substance != step.Substance {
      return fmt.Errorf("请按照教学脚本打出 %s", step.Substance)
    }
  }
  // 原有逻辑...
}
```

## 📊 数据流

```
用户进入教学关卡
  ↓
localStorage.setItem('chemistry-uno-tutorial-mode', 'true')
  ↓
GameRoom检测教学模式
  ↓
tutorialScriptMode.value = true
tutorialCurrentStep.value = 1
  ↓
显示欢迎Toast（脚本化教学）
  ↓
[循环] 玩家回合
  ├→ 显示脚本提示（Step X/8 + 进度条）
  ├→ 玩家选择物质
  ├→ 验证是否是脚本中的物质
  │   ├→ 正确 → 执行出牌 → tutorialCurrentStep++
  │   └→ 错误 → 显示警告Toast，阻止出牌
  ↓
[循环] AI回合
  ├→ 后端按照脚本执行AI行动
  ├→ tutorialCurrentStep++
  ↓
重复直到第8步完成
  ↓
教学关卡完成
```

## ✅ 当前状态

### 前端（已完成）
- ✅ 教学脚本配置文件
- ✅ 脚本化模式检测
- ✅ 步骤管理和进度显示
- ✅ 出牌验证（阻止非脚本内的操作）
- ✅ UI进度条和步骤指示器
- ✅ 脚本化提示显示
- ✅ 自动步骤递进

### 后端（待开发）
- ❌ 初始手牌设置
- ❌ AI脚本行为
- ❌ 游戏状态扩展
- ❌ 服务端出牌验证

## 🚀 使用方法

### 开发者测试

```javascript
// 浏览器控制台
localStorage.setItem('chemistry-uno-tutorial-mode', 'true')
location.reload()
```

### 重置教学状态

```javascript
localStorage.removeItem('chemistry-uno-tutorial-mode')
localStorage.removeItem('chemistry-uno-tutorial-welcome-shown')
location.reload()
```

## 📝 下一步开发

### Phase 1: 后端基础 🔴 优先
1. 扩展CreateRoom API，支持`tutorial_script`参数
2. 在游戏创建时设置固定手牌
3. 扩展Game模型，添加脚本状态字段

### Phase 2: AI脚本行为 🟠 中等
1. 在AI决策中添加脚本模式分支
2. 实现AI按脚本出牌的逻辑
3. 同步脚本步骤状态

### Phase 3: 服务端验证 🟡 可选
1. 在PlayCard中添加脚本验证
2. 返回明确的错误提示
3. 防止客户端绕过验证

### Phase 4: 前后端联调 🟢 最终
1. 前端调用新的创建房间API
2. 测试完整的教学流程
3. 优化提示文案和用户体验

---

**更新时间**：2026-02-24
**版本**：Chemistry UNO V1.2.0 Mendeleef
**状态**：前端已完成，等待后端实现
**相关文档**：[TutorialMatch.IMPLEMENTATION.md](./TutorialMatch.IMPLEMENTATION.md)
