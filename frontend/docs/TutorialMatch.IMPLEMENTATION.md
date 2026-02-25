# 教学关卡系统 - 实现文档

## 🎯 功能概述

新玩家完成大厅指引后，自动创建低难度AI对战教学关卡，并在游戏中提供实时智能提示。

## 📋 实现细节

### 1. 大厅指引调整

#### 步骤顺序优化 ([Lobby.vue:83-91](../src/pages/Lobby.vue#L83-L91))
将AI竞技场移到最后一步，作为教学关卡的入口：

```typescript
const lobbyTutorialSteps = computed(() => [
  // 1. 欢迎
  // 2. 创建房间
  // 3. 房间列表
  // 4. 导航菜单
  // 5. 个人资料
  // 6. AI竞技场 ← 移到最后
  // 7. 完成提示
])
```

### 2. 自动创建教学关卡

#### 完成回调 ([Lobby.vue:98-130](../src/pages/Lobby.vue#L98-L130))
新手指引完成时自动创建教学房间：

```typescript
const handleTutorialComplete = async () => {
  localStorage.setItem('chemistry-uno-lobby-tutorial-completed', 'true')
  showTutorial.value = false

  // 自动创建教学关卡
  await createTutorialMatch()
}

const createTutorialMatch = async () => {
  const response = await gameAPI.createRoom(
    'Tutorial: First AI Battle',  // 房间名称
    2,                              // 1人类 + 1AI
    deckID.value,                   // 默认牌组
    false,                          // 非积分模式
    true,                           // 私密房间
    undefined,                      // 无访问密钥
    true,                           // PvE模式
    20,                             // AI难度：20/100（低难度）
    1,                              // 1个AI
    false,                          // 无补位
    0, false, 0
  )

  // 标记教学模式
  localStorage.setItem('chemistry-uno-tutorial-mode', 'true')

  // 进入游戏房间
  router.push(`/room/${room.id}`)
}
```

### 3. 教学模式检测

#### 状态管理 ([GameRoom.vue:40-41](../src/pages/GameRoom.vue#L40-L41))
```typescript
const isTutorialMode = ref(false)
const tutorialHintText = ref('')
```

#### 初始化检测 ([GameRoom.vue:909-915](../src/pages/GameRoom.vue#L909-L915))
```typescript
onMounted(() => {
  // 检测教学模式
  const tutorialMode = localStorage.getItem('chemistry-uno-tutorial-mode')
  if (tutorialMode === 'true') {
    isTutorialMode.value = true
    console.log('[GameRoom] Tutorial mode activated')
  }
  // ...
})
```

#### 退出清理 ([GameRoom.vue:987-992](../src/pages/GameRoom.vue#L987-L992))
```typescript
onUnmounted(() => {
  // 清除教学模式标记
  if (isTutorialMode.value) {
    localStorage.removeItem('chemistry-uno-tutorial-mode')
    localStorage.removeItem('chemistry-uno-tutorial-welcome-shown')
    console.log('[GameRoom] Tutorial mode completed and cleared')
  }
  // ...
})
```

### 6. 教学模式欢迎提示 ⭐ NEW

#### 移动端化学键盘集成 ⭐ FIXED

```typescript
// 输入框焦点管理：移动端打开化学键盘
const handleInputFocus = () => {
  if (isMobile.value) {
    showChemicalKeyboard.value = true  // 移动端打开化学键盘
  } else {
    if (!isTutorialMode.value) {
      exitFullscreen()  // 桌面端退出全屏
    }
  }
}

// 化学键盘确认处理
const handleKeyboardConfirm = async (formula: string) => {
  substanceInput.value = formula
  showChemicalKeyboard.value = false
  await handleInputPlay()
}
```

**修复内容**：
- ✅ 移动端点击输入框正确打开化学键盘
- ✅ 教学模式禁用全屏切换，避免干扰输入
- ✅ 键盘确认后自动执行打牌操作
- ✅ 化学键盘根据牌组显示可用元素

#### 进入游戏提示 ([GameRoom.vue:977-990](../src/pages/GameRoom.vue#L977-L990))
```typescript
// 教学模式欢迎提示
if (isTutorialMode.value) {
  const tutorialWelcomeShown = localStorage.getItem('chemistry-uno-tutorial-welcome-shown')
  if (!tutorialWelcomeShown) {
    setTimeout(() => {
      showToast(
        '💡 欢迎来到教学关卡！这是一场低难度的AI对战，在你的回合时会出现实时提示帮助你学习游戏。祝你玩得开心！',
        '🎯 教学模式已开启',
        'success',
        8000
      )
      localStorage.setItem('chemistry-uno-tutorial-welcome-shown', 'true')
    }, 1500)
  }
}
```

#### 游戏开始提示 ([GameRoom.vue:350-361](../src/pages/GameRoom.vue#L350-L361))
```typescript
// 教学模式：监听游戏状态变化，游戏开始时提示
watch(() => gameState.value?.status, (newStatus, oldStatus) => {
  if (isTutorialMode.value && newStatus === 'playing' && oldStatus === 'waiting') {
    setTimeout(() => {
      showToast(
        '游戏已开始！注意查看底部的橙色提示卡片，它会在你的回合时告诉你该做什么。',
        '🎮 开始游戏',
        'info',
        6000
      )
    }, 1000)
  }
})
```

#### 提示时机
1. **进入教学关卡**（1.5秒后）
   - 标题：🎯 教学模式已开启
   - 内容：欢迎来到教学关卡，说明实时提示功能
   - 持续：8秒
   - 只显示一次（localStorage标记）

2. **游戏开始**（waiting → playing，1秒后）
   - 标题：🎮 开始游戏
   - 内容：提醒查看底部橙色提示卡片
   - 持续：6秒
   - 自动触发

### 4. 智能提示系统

#### 提示生成逻辑 ([GameRoom.vue:770-799](../src/pages/GameRoom.vue#L770-L799))
```typescript
const generateTutorialHint = () => {
  if (!gameState.value || !isMyTurn.value) {
    tutorialHintText.value = ''
    return
  }

  const myPlayer = gameState.value.players?.[gameState.value.current_player]
  if (!myPlayer) return

  const handSize = myPlayer.hand?.length || 0
  const topCard = gameState.value.discard_top

  // 根据游戏状态生成提示
  if (!topCard) {
    tutorialHintText.value = '💡 回合开始：你可以先在「化学库」中选择一个物质，然后点击「打出卡牌」开始游戏！'
  } else if (handSize === 0) {
    tutorialHintText.value = '💡 手牌用完了！点击「摸牌」按钮抽一张新牌（你将失去这个回合）'
  } else if (doubleMode.value) {
    tutorialHintText.value = '💡 双元素模式：选择第二个物质，两个物质将一起打出到战场中'
  } else if (selectedSubstance.value) {
    tutorialHintText.value = '💡 已选择物质！点击「打出卡牌」按钮将它放到战场上，或点击「双元素」同时打出两张'
  } else {
    tutorialHintText.value = '💡 轮到你了：在「化学库」中选择一个物质，让它与战场中央的卡牌发生化学反应！'
  }
}
```

#### 触发时机
1. **游戏状态加载** ([GameRoom.vue:870-873](../src/pages/GameRoom.vue#L870-L873))
   ```typescript
   if (data.game_state) {
     gameState.value = data.game_state
     if (isTutorialMode.value && isMyTurn.value) {
       generateTutorialHint()
     }
   }
   ```

2. **选择物质/双元素模式变化** ([GameRoom.vue:342-347](../src/pages/GameRoom.vue#L342-L347))
   ```typescript
   watch([selectedSubstance, doubleMode], () => {
     if (isTutorialMode.value && isMyTurn.value) {
       generateTutorialHint()
     }
   })
   ```

### 5. UI显示

#### 提示卡片 ([GameRoom.vue:1870-1896](../src/pages/GameRoom.vue#L1870-L1896))
```vue
<div v-if="isTutorialMode && tutorialHintText && isMyTurn"
     class="absolute bottom-full mb-32 sm:mb-24 left-0 right-0 flex justify-center px-4">
  <div class="bg-gradient-to-br from-amber-500 to-orange-500 backdrop-blur-xl
              border-2 border-amber-300 rounded-2xl shadow-2xl px-6 py-4 max-w-lg">
    <!-- 装饰 -->
    <div class="absolute inset-0 bg-[radial-gradient(circle_at_30%_50%,rgba(255,255,255,0.2),transparent_50%)]"></div>

    <!-- 图标 + 文本 -->
    <div class="relative flex items-start gap-3">
      <div class="w-8 h-8 rounded-full bg-white/20 flex items-center justify-center">
        <Sparkles class="w-5 h-5 text-white" />
      </div>
      <p class="text-white text-sm sm:text-base font-bold leading-relaxed">
        {{ tutorialHintText }}
      </p>
    </div>
  </div>
</div>
```

## 🎨 视觉设计

### 提示卡片样式
- **渐变背景**：amber-500 → orange-500
- **边框**：2px amber-300
- **装饰**：径向渐变 + 白色光晕
- **图标**：Sparkles 脉冲动画
- **文字**：白色粗体，易读性强
- **位置**：操作按钮上方32-24单位

### 响应式适配
- **移动端**：mb-32（底部距离更大）
- **桌面端**：mb-24
- **最大宽度**：max-w-lg（保持紧凑）

## 📊 提示触发逻辑

```
游戏开始
  ↓
检测教学模式
  ↓
显示欢迎Toast (1.5秒后)
  "欢迎来到教学关卡！..."
  ↓
等待游戏状态变为playing
  ↓
显示开始Toast (1秒后)
  "游戏已开始！注意查看底部橙色提示卡片..."
  ↓
是我的回合？ ——否——→ 不显示提示
  ↓ 是
生成提示文本
  ├→ 没有顶牌：回合开始提示
  ├→ 手牌为空：摸牌提示
  ├→ 双元素模式：双联提示
  ├→ 已选物质：打出卡牌提示
  └→ 默认：选择物质提示
  ↓
显示在UI中（橙色渐变卡片）
  ↓
监听状态变化
  ├→ 选择物质 → 更新提示
  ├→ 双元素模式 → 更新提示
  └→ 游戏状态更新 → 更新提示
```

## 🚀 使用流程

### 新玩家体验
1. **首次登录** → 触发大厅新手指引
2. **浏览大厅** → 了解创建房间、房间列表等功能
3. **完成指引** → 最后一步指向AI竞技场
4. **点击完成** → 自动创建教学关卡
5. **进入游戏** → 显示欢迎Toast（8秒）⭐
6. **游戏开始** → 显示开始提示Toast（6秒）⭐
7. **我的回合** → 显示实时橙色提示卡片
8. **完成对战** → 清除教学标记

### 老玩家
- 不触发指引（localStorage已标记）
- 正常游戏流程
- 无教学提示干扰

## 🔧 开发者调试

### 重置教学状态
```javascript
// 浏览器控制台
localStorage.removeItem('chemistry-uno-lobby-tutorial-completed')
localStorage.removeItem('chemistry-uno-tutorial-mode')
location.reload()
```

### 强制开启教学模式
```javascript
localStorage.setItem('chemistry-uno-tutorial-mode', 'true')
location.reload()
```

### 测试指引顺序
```javascript
resetLobbyTutorial()  // 重置大厅指引
showLobbyTutorial()   // 显示指引
```

## ✅ 构建状态

**最新构建**：✅ 成功（4.03s）
**文件大小**：前端 249.68 KB，后端 待测试
**状态**：生产就绪

## 🎓 脚本化教学系统（最新功能）⭐ NEW

在原有的智能提示基础上，新增了**完全脚本化的教学模式**，玩家和AI按照固定顺序出牌，确保新手严格学习游戏机制。

### 核心特性

- ✅ 固定的初始手牌（玩家：Na, Mg, O, H, Au, Ar, +2）
- ✅ 固定的出牌顺序（8个步骤）
- ✅ 出牌验证（只能打出脚本指定的牌）
- ✅ 实时进度显示（Step X/8 + 百分比）
- ✅ 每步详细提示（包含教学目标）
- ⏳ AI脚本行为（后端待实现）
- ⏳ 初始手牌设置（后端待实现）

**详细文档**：[TutorialScript.IMPLEMENTATION.md](./TutorialScript.IMPLEMENTATION.md)

### 快速启用

```javascript
// 浏览器控制台
localStorage.setItem('chemistry-uno-tutorial-mode', 'true')
location.reload()
```

## 📝 待填写文案

### 大厅指引（第6步）
```
LOBBY_AI_ARENA_TITLE    - AI竞技场标题
LOBBY_AI_ARENA_CONTENT  - AI竞技场说明（引导到教学关卡）
```

建议文案：
```
标题：AI竞技场
内容：现在让我们开始第一场AI对战！系统将为你匹配一个低难度的AI对手，在游戏中你会看到实时提示。点击「完成」开始你的第一场化学实验！
```

---

**更新时间**：2026-02-24
**版本**：Chemistry UNO V1.2.0 Mendeleef
**相关文档**：[TutorialSystem.SUMMARY.md](./TutorialSystem.SUMMARY.md)
