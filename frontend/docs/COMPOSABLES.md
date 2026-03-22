# 组合式函数（Composables）使用指南

## 概述

为了解决 GameRoom.vue 2810 行代码的膨胀问题，我们将其拆分为多个可复用的组合式函数（Composables）。

## 架构设计

```text
┌─────────────────────────────────┐
│       GameRoom.vue (组件)        │
│        协调各个 Composables      │
└───────────┬─────────────────────┘
            │
    ┌───────┴────────┐
    │  Composables   │
    └───────┬────────┘
            │
    ┌───────┴──────────────────────────────┐
    │                                      │
┌───▼────────────┐  ┌───────────────────┐ │
│ useGameState   │  │ useCardActions    │ │
│ 游戏状态管理    │  │ 卡牌操作管理       │ │
└────────────────┘  └───────────────────┘ │
                                          │
┌───────────────┐  ┌────────────────────┐│
│ useGameSocket │  │ useGameUI          ││
│ WebSocket通信  │  │ UI状态管理          ││
└───────────────┘  └────────────────────┘│
                                          │
┌──────────────┐  ┌─────────────────────┐│
│ useTutorial  │  │ useAdminActions     ││
│ 教程管理      │  │ 管理员操作          ││
└──────────────┘  └─────────────────────┘│
                                          │
└──────────────────────────────────────────┘
```

## 模块说明

### 1. useGameState - 游戏状态管理

**职责**：管理游戏状态、房间信息和玩家数据

**文件**：[useGameState.ts](../composables/useGameState.ts)

**导出内容**：
- 状态：`gameState`, `roomInfo`, `playersInfo`, `loading`, `loadError`
- 计算属性：`allPlayers`, `isMyTurn`, `currentPlayerObj`
- 方法：`loadGameState()`, `updateGameState()`, `updateRoomInfo()`

**使用示例**：
```typescript
import { useGameState } from '@/composables'

const {
  gameState,
  roomInfo,
  allPlayers,
  isMyTurn,
  loadGameState
} = useGameState(roomId)

// 加载游戏状态
await loadGameState()

// 访问当前玩家
console.log('我的回合:', isMyTurn.value)
```

### 2. useCardActions - 卡牌操作管理

**职责**：处理卡牌点击、打牌、摸牌等操作

**文件**：[useCardActions.ts](../composables/useCardActions.ts)

**导出内容**：
- 状态：`selectedCard`, `selectedSubstance`, `doubleMode`, etc.
- 方法：`handleCardClick()`, `handlePlayCard()`, `handleDrawCard()`, etc.

**使用示例**：
```typescript
import { useCardActions } from '@/composables'

const {
  selectedCard,
  doubleMode,
  handleCardClick,
  handleDrawCard
} = useCardActions(roomId, gameState)

// 点击卡牌
await handleCardClick(card)

// 摸牌
await handleDrawCard()
```

### 3. useGameSocket - WebSocket 通信

**职责**：管理 WebSocket 连接和消息处理

**文件**：[useGameSocket.ts](../composables/useGameSocket.ts)

**导出内容**：
- 方法：`sendMessage()`, `connect()`, `disconnect()`

**使用示例**：
```typescript
import { useGameSocket } from '@/composables'

const { sendMessage } = useGameSocket(roomId, {
  onGameUpdate: (data) => {
    console.log('游戏更新:', data)
  },
  onPlayerJoined: (data) => {
    console.log('玩家加入:', data)
  }
})

// 发送消息
sendMessage('chat', { message: 'Hello!' })
```

### 4. useGameUI - UI 状态管理

**职责**：管理界面的各种 UI 状态（模态框、面板等）

**文件**：[useGameUI.ts](../composables/useGameUI.ts)

**导出内容**：
- 状态：`showHints`, `showPlayers`, `showChat`, etc.
- 方法：`closeAllModals()`, `togglePanel()`, `openModal()`

**使用示例**：
```typescript
import { useGameUI } from '@/composables'

const {
  showHints,
  showAdminModal,
  togglePanel,
  closeAllModals
} = useGameUI()

// 切换面板
togglePanel('hints')

// 关闭所有模态框
closeAllModals()
```

### 5. useTutorial - 教程管理

**职责**：处理游戏教程逻辑

**文件**：[useTutorial.ts](../composables/useTutorial.ts)

**导出内容**：
- 状态：`isTutorialMode`, `tutorialHintText`, etc.
- 方法：`validateTutorialAction()`, `advanceTutorialStep()`

**使用示例**：
```typescript
import { useTutorial } from '@/composables'

const {
  isTutorialMode,
  tutorialHintText,
  validateTutorialAction
} = useTutorial(roomInfo, gameState, isMyTurn)

// 验证教程操作
if (!validateTutorialAction(substance)) {
  console.log('教程模式：请按提示操作')
}
```

### 6. useAdminActions - 管理员操作

**职责**：处理管理员的踢人、封禁等操作

**文件**：[useAdminActions.ts](../composables/useAdminActions.ts)

**导出内容**：
- 状态：`adminTargetUser`, `banReason`, etc.
- 方法：`openAdminModal()`, `executeAdminAction()`

**使用示例**：
```typescript
import { useAdminActions } from '@/composables'

const {
  openAdminModal,
  executeAdminAction
} = useAdminActions()

// 打开管理员操作
openAdminModal(player, 'kick')

// 执行操作
await executeAdminAction()
```

## 在 GameRoom.vue 中使用

```vue
<script setup lang="ts">
import {
  useGameState,
  useCardActions,
  useGameSocket,
  useGameUI,
  useTutorial,
  useAdminActions
} from '@/composables'

const route = useRoute()
const roomId = route.params.id as string

// 1. 游戏状态
const {
  gameState,
  roomInfo,
  allPlayers,
  isMyTurn,
  loadGameState
} = useGameState(roomId)

// 2. 卡牌操作
const {
  selectedCard,
  handleCardClick,
  handleDrawCard
} = useCardActions(roomId, gameState)

// 3. WebSocket
const { sendMessage } = useGameSocket(roomId, {
  onGameUpdate: (data) => {
    updateGameState(data.game_state)
  }
})

// 4. UI 状态
const {
  showHints,
  showPlayers,
  togglePanel
} = useGameUI()

// 5. 教程
const {
  isTutorialMode,
  tutorialHintText
} = useTutorial(roomInfo, gameState, isMyTurn)

// 6. 管理员
const {
  openAdminModal,
  executeAdminAction
} = useAdminActions()

// 初始化
onMounted(async () => {
  await loadGameState()
})
</script>
```

## 优势

### 代码组织
- ✅ **清晰的职责分离** - 每个模块只负责一项功能
- ✅ **易于理解** - 每个文件只有 100-200 行代码
- ✅ **降低耦合** - 模块间通过参数传递通信

### 可维护性
- ✅ **易于定位问题** - 知道去哪个文件找
- ✅ **易于修改** - 修改不会影响其他模块
- ✅ **易于测试** - 可以独立测试每个 composable

### 可复用性
- ✅ **跨组件复用** - 其他页面也可以使用
- ✅ **按需导入** - 只导入需要的 composable
- ✅ **独立发布** - 可以作为独立包发布

## 迁移指南

从旧的 GameRoom.vue 迁移到新架构：

1. **识别功能区块** - 找出相关的状态和方法
2. **导入对应的 composable** - 使用解构获取需要的部分
3. **删除旧代码** - 删除已迁移到 composable 的代码
4. **测试功能** - 确保功能正常工作

## 最佳实践

1. **单一职责** - 每个 composable 只做一件事
2. **类型安全** - 使用 TypeScript 提供类型提示
3. **错误处理** - 在 composable 内部处理错误
4. **文档注释** - 为每个导出添加注释
5. **测试覆盖** - 为每个 composable 编写单元测试

## 未来扩展

可以继续添加的 composables：
- `useGameTimer` - 倒计时管理
- `useGameAchievements` - 成就系统
- `useGameStatistics` - 统计数据
- `useGameReplay` - 回放功能

---

**版本**: 1.0.0
**最后更新**: 2026-02-24
**维护者**: Chemistry UNO Team
