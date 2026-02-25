# 操作反馈系统使用文档

## 功能概述

Chemistry UNO 的操作反馈系统为玩家提供丰富的交互反馈，包括：
- 🔊 **提示音** - 基于 Web Audio API 的合成音效
- 📳 **震动反馈** - 使用 Vibration API 的触觉反馈
- ⚙️ **设置面板** - 可自定义音效开关、音量和震动开关

## 系统架构

反馈系统采用模块化设计，职责分离：

```text
┌─────────────────────────────────────┐
│      FeedbackManager (协调器)       │
│   统一的反馈接口 + 设置管理          │
└──────────┬──────────────┬───────────┘
           │              │
    ┌──────▼──────┐  ┌───▼────────┐
    │ AudioEngine │  │ Vibration  │
    │  音效引擎    │  │   Engine   │
    │             │  │  振动引擎   │
    └─────────────┘  └────────────┘
```

### 模块说明

1. **FeedbackManager** ([feedback.ts](../utils/feedback.ts))
   - 协调音效和振动引擎
   - 提供统一的 API 接口
   - 管理用户设置和 localStorage

2. **AudioEngine** ([audioEngine.ts](../utils/audioEngine.ts))
   - 独立的音效处理模块
   - 使用 Web Audio API 生成合成音
   - 支持音量控制和音效序列

3. **VibrationEngine** ([vibrationEngine.ts](../utils/vibrationEngine.ts))
   - 独立的振动处理模块
   - 使用 Vibration API 触发振动
   - 提供诊断和支持检测

### 设计优势

✅ **职责分离** - 每个模块只负责一项功能
✅ **易于测试** - 可以独立测试每个引擎
✅ **易于扩展** - 可以轻松替换音效实现（如使用音频文件）
✅ **易于维护** - 代码组织清晰，降低耦合度

## 文件位置

- **反馈管理器**: [frontend/src/utils/feedback.ts](../utils/feedback.ts)
- **音效引擎**: [frontend/src/utils/audioEngine.ts](../utils/audioEngine.ts)
- **振动引擎**: [frontend/src/utils/vibrationEngine.ts](../utils/vibrationEngine.ts)
- **设置组件**: [frontend/src/components/FeedbackSettings.vue](../components/FeedbackSettings.vue)
- **集成位置**: [frontend/src/pages/GameRoom.vue](../pages/GameRoom.vue)

## 音效类型

系统内置以下音效：

| 音效类型 | 用途 | 音色 | 音符序列 |
|---------|------|------|---------|
| `click` | 普通点击 | 正弦波 | C5 (800Hz) |
| `play-card` | 打牌 | 三角波 | C5 → E5 |
| `draw-card` | 摸牌 | 正弦波 | G4 → C5 |
| `reaction` | 化学反应 | 方波 | E5 → G5 → B5 |
| `turn-start` | 回合开始 | 正弦波 | C5 → F5 |
| `success` | 成功操作 | 三角波 | C5 → E5 → G5 |
| `error` | 错误操作 | 方波 | 低频震颤 |
| `win` | 胜利 | 三角波 | C5 → E5 → G5 → C6 |
| `lose` | 失败 | 正弦波 | G4 → F4 → D4 |
| `double-mode` | 双联模式 | 方波 | E5 → G5 |
| `special` | 特殊效果 | 锯齿波 | A5 → C6 |

## 震动模式

预设震动模式：

| 模式 | 时长 (ms) | 用途 |
|-----|----------|------|
| `light` | 10 | 轻触点击 |
| `medium` | 20 | 中等操作 |
| `heavy` | 30 | 重击操作 |
| `double` | [20, 50, 20] | 双击效果 |
| `success` | [10, 50, 10, 50, 20] | 成功序列 |
| `error` | [30, 100, 30] | 错误震颤 |
| `reaction` | [15, 30, 15, 30, 25] | 反应序列 |

## 使用方法

### 1. 基础使用

```typescript
import feedback from '../utils/feedback'

// 播放音效
feedback.playSound('click')

// 触发震动
feedback.vibrate('medium')

// 组合反馈
feedback.feedback({
  sound: 'success',
  vibration: 'success'
})
```

### 2. 快捷方法

```typescript
// 点击反馈
feedback.click()

// 打牌反馈
feedback.playCard()

// 摸牌反馈
feedback.drawCard()

// 化学反应反馈
feedback.reaction()

// 回合开始反馈
feedback.turnStart()

// 成功反馈
feedback.success()

// 错误反馈
feedback.error()

// 胜利反馈
feedback.win()

// 失败反馈
feedback.lose()

// 双联模式反馈
feedback.doubleMode()
```

### 3. 设置管理

```typescript
// 开关音效
feedback.setSoundEnabled(true)

// 开关震动
feedback.setVibrationEnabled(true)

// 设置音量 (0.0 - 1.0)
feedback.setVolume(0.5)

// 获取当前设置
const settings = feedback.getSettings()
// { soundEnabled: true, vibrationEnabled: true, volume: 0.3 }
```

## 在 GameRoom 中的集成

### 已添加反馈的操作

| 操作 | 反馈类型 | 触发时机 |
|-----|---------|---------|
| 打牌成功 | `playCard()` | 出牌/输入化学式成功后 |
| 摸牌 | `drawCard()` | 摸牌成功后 |
| 双联模式切换 | `doubleMode()` | 开启/关闭双联模式 |
| 回合开始 | `turnStart()` | 轮到玩家回合时 |
| 化学反应 | `reaction()` | 检测到 `current_reaction` 变化 |
| 胜利 | `win()` | 游戏结束且玩家第一名 |
| 完成游戏 | `success()` | 游戏结束但非第一名 |
| 操作错误 | `error()` | 出牌失败、摸牌失败等 |
| UI 点击 | `click()` | 提示面板、聊天面板等按钮 |

### 关键代码位置

**回合开始监听** (GameRoom.vue:729-737):
```typescript
watch(() => isMyTurn.value, (val) => {
  if (val) {
    fetchTurnSubstances()
    feedback.turnStart() // 回合开始音效
  } else {
    turnReadySubstances.value = []
  }
}, { immediate: true })
```

**化学反应监听** (GameRoom.vue:739-743):
```typescript
watch(() => gameState.value?.current_reaction, (newReaction, oldReaction) => {
  if (newReaction && newReaction !== oldReaction) {
    feedback.reaction() // 反应音效
  }
})
```

**游戏结束监听** (GameRoom.vue:745-755):
```typescript
watch(() => gameState.value?.status, (newStatus) => {
  if (newStatus === 'finished' && gameState.value?.finished_players) {
    const finishedPlayers = gameState.value.finished_players
    const myUID = user.value.uid
    if (finishedPlayers.length > 0 && finishedPlayers[0] === myUID) {
      feedback.win() // 胜利音效
    } else if (finishedPlayers.includes(myUID)) {
      feedback.success() // 完成音效
    }
  }
})
```

## 设置面板

FeedbackSettings 组件提供可视化设置界面：

### 功能
- ✅ 音效开关
- 🎚️ 音量滑块 (0-100%)
- ✅ 震动开关
- 🎮 测试按钮

### 位置
游戏房间右下角浮动齿轮图标 → 点击打开设置面板

### 数据持久化
所有设置自动保存到 `localStorage`:
- `chemistry-uno-sound-enabled`
- `chemistry-uno-vibration-enabled`
- `chemistry-uno-volume`

## 技术细节

### Web Audio API
- **音频上下文**: 懒加载，首次播放时创建
- **音色类型**: sine (正弦), triangle (三角), square (方波), sawtooth (锯齿)
- **音量控制**: 通过 GainNode 实现淡入淡出
- **浏览器兼容**: 支持 Chrome、Firefox、Safari、Edge

### Vibration API
- **浏览器支持**: 主要支持移动端浏览器
- **安全限制**: 仅在用户交互后可用
- **兼容性**: 桌面端自动跳过震动（无 `navigator.vibrate`）

### 性能优化
- 音效采用合成音，无需加载外部音频文件
- 单例模式，全局共享一个 AudioContext
- 设置缓存在内存和 localStorage 中

## 浏览器兼容性

| 功能 | Chrome | Firefox | Safari | Edge | 移动端 |
|-----|--------|---------|--------|------|-------|
| 音效 | ✅ | ✅ | ✅ | ✅ | ✅ |
| 震动 | 🔶 | 🔶 | ❌ | 🔶 | ✅ |

- ✅ 完全支持
- 🔶 部分支持（桌面端不支持震动）
- ❌ 不支持

## 扩展方法

### 添加自定义音效

在 `feedback.ts` 的 `playSound` 方法中添加：

```typescript
case 'my-custom-sound':
  this.playSequence([
    { frequency: 440, duration: 0.1, type: 'sine' },
    { frequency: 550, duration: 0.15, type: 'triangle', delay: 100 },
  ])
  break
```

### 添加自定义震动模式

在 `VIBRATION_PATTERNS` 中添加：

```typescript
const VIBRATION_PATTERNS = {
  // ... 现有模式
  myPattern: [50, 100, 50, 100, 100], // 自定义序列
}
```

### 创建快捷方法

在 `FeedbackManager` 类中添加：

```typescript
myCustomFeedback() {
  this.feedback({
    sound: 'my-custom-sound',
    vibration: 'myPattern'
  })
}
```

## 常见问题

### Q: 为什么移动端震动不工作？
A: 浏览器可能不支持 Vibration API（如 iOS Safari），或者需要用户先与页面交互。

### Q: 音效音量太大/太小？
A: 点击游戏房间右下角齿轮图标，打开设置面板调整音量。

### Q: 能否禁用反馈？
A: 可以，在设置面板中关闭音效和震动开关。

### Q: 反馈会消耗多少性能？
A: 非常少，音效是合成波形，震动是原生API，性能开销可忽略。

## 未来扩展

可考虑的改进方向：
- [ ] 支持外部音频文件（MP3/WAV）
- [ ] 更多音效类型（粒子加速、爆炸等）
- [ ] 音效主题包（科技风、经典风等）
- [ ] 高级音效：混响、延迟、滤波器
- [ ] 音效可视化器

---

**版本**: 1.0.0
**最后更新**: 2026-02-24
**维护者**: Chemistry UNO Team
