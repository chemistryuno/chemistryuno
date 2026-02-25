# 大厅新手指引 - 快速集成说明

由于原始Lobby.vue文件结构复杂，为了保证稳定性，建议采用以下方式集成新手指引：

## 方案1：精简版（推荐）

创建独立的大厅新手指引组件，在App.vue层级控制显示，避免修改Lobby.vue复杂结构。

## 方案2：完整集成

需要在Lobby.vue中添加：

### 1. 导入组件（第9行后）
```typescript
import TutorialGuide from '../components/TutorialGuide.vue'
```

### 2. 添加状态（第73行后）
```typescript
// Tutorial State
const showTutorial = ref(false)

const lobbyTutorialSteps = [
  // ... 步骤配置
]

const checkFirstTimeLobby = () => {
  const hasSeenLobbyTutorial = localStorage.getItem('chemistry-uno-lobby-tutorial-completed')
  if (!hasSeenLobbyTutorial) {
    setTimeout(() => showTutorial.value = true, 1500)
  }
}

const handleTutorialComplete = () => {
  localStorage.setItem('chemistry-uno-lobby-tutorial-completed', 'true')
  showTutorial.value = false
}

const handleTutorialClose = () => {
  showTutorial.value = false
}
```

### 3. onMounted初始化（第240行后）
```typescript
  // 检查大厅新手指引
  checkFirstTimeLobby()

  // 控制台指令
  if (typeof window !== 'undefined') {
    (window as any).showLobbyTutorial = () => {
      showTutorial.value = true
      console.log('✨ 大厅新手指引已启动')
    }
    (window as any).resetLobbyTutorial = () => {
      localStorage.removeItem('chemistry-uno-lobby-tutorial-completed')
      console.log('🔄 已重置')
    }
  }
```

### 4. 模板末尾（第1390行，</div>之后，</template>之前）
```vue
    <TutorialGuide
      :show="showTutorial"
      :steps="lobbyTutorialSteps"
      @close="handleTutorialClose"
      @complete="handleTutorialComplete"
    />
  </div>
</template>
```

## 文案占位符

```
LOBBY_WELCOME_TITLE / CONTENT - 欢迎页
LOBBY_CREATE_ROOM_TITLE / CONTENT - 创建房间按钮
LOBBY_AI_ARENA_TITLE / CONTENT - AI竞技场
LOBBY_ROOM_LIST_TITLE / CONTENT - 房间列表
LOBBY_NAVIGATION_TITLE / CONTENT - 导航菜单
LOBBY_PROFILE_TITLE / CONTENT - 个人资料
LOBBY_COMPLETE_TITLE / CONTENT - 完成提示
```

## 控制台指令

```javascript
showLobbyTutorial()   // 显示大厅指引
resetLobbyTutorial()  // 重置状态
```

## 已创建的文件

- [TutorialGuide.vue](../src/components/TutorialGuide.vue) - 通用新手指引组件
- [TutorialGuide.README.md](../src/components/TutorialGuide.README.md) - 详细文档
- [TutorialGuide.CONSOLE.md](../src/components/TutorialGuide.CONSOLE.md) - 控制台指令

## 注意事项

Lobby.vue的HTML结构非常复杂，直接修改容易破坏div平衡。建议先在GameRoom.vue中测试新手指引功能（已完成），确认无误后再谨慎集成到Lobby。

或者考虑在用户首次进入游戏房间时统一显示大厅+房间的综合指引。
