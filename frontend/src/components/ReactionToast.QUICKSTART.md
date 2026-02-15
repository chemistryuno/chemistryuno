# 化学反应通知系统 - 快速集成指南

## 🚀 5分钟快速上手

### 步骤1：复制组件文件

将以下文件复制到你的项目：
```
frontend/src/components/
  ├── ReactionToast.vue          ← 主组件
  ├── ReactionToast.README.md    ← 详细文档
  └── ReactionToastDemo.vue      ← 演示页面（可选）
```

### 步骤2：在游戏页面中引入

在 `GameRoom.vue` 或你的主游戏组件中：

```vue
<template>
  <div class="game-room">
    <!-- 你的游戏界面 -->
    <div class="game-content">
      <!-- ... -->
    </div>

    <!-- 添加反应通知组件 -->
    <ReactionToast ref="reactionToast" />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import ReactionToast from '@/components/ReactionToast.vue'

const reactionToast = ref(null)

// 暴露给其他地方调用
defineExpose({
  showReactionToast: (type, equation, name, energy) => {
    reactionToast.value?.showToast(type, equation, name, energy)
  }
})
</script>
```

### 步骤3：触发通知

有多种方式触发通知：

#### 方式A：直接调用
```javascript
// 在同一组件内
reactionToast.value?.showToast(
  'synthesis',           // 反应类型
  'H₂ + O₂ → H₂O',      // 化学方程式
  'Water Formation',     // 反应名称
  85                     // 能量值(0-100)
)
```

#### 方式B：通过WebSocket监听
```javascript
import { websocket } from '@/utils/websocket'

// 监听游戏事件
websocket.on('reaction_occurred', (data) => {
  reactionToast.value?.showToast(
    data.reactionType,    // 后端返回的反应类型
    data.equation,        // 后端返回的方程式
    data.name,           // 后端返回的反应名称
    data.energy || 75    // 后端返回的能量值
  )
})
```

#### 方式C：通过游戏事件触发
```javascript
// 当玩家出牌成功
const onCardPlayed = (cardData) => {
  if (cardData.hasReaction) {
    reactionToast.value?.showToast(
      cardData.reactionType,
      cardData.equation,
      cardData.name,
      cardData.energy
    )
  }
}
```

---

## 📝 反应类型映射

### 后端 → 前端类型映射

如果你的后端使用不同的命名，需要做映射：

```javascript
const reactionTypeMap = {
  'synthesis': 'synthesis',
  'combination': 'synthesis',      // 别名
  'decomposition': 'decomposition',
  'breakdown': 'decomposition',    // 别名
  'displacement': 'displacement',
  'single_replacement': 'displacement',  // 别名
  'combustion': 'combustion',
  'burning': 'combustion',         // 别名
  'neutralization': 'neutralization',
  'acid_base': 'neutralization'    // 别名
}

// 使用映射
const mappedType = reactionTypeMap[backendType] || 'synthesis'
reactionToast.value?.showToast(mappedType, equation, name, energy)
```

---

## 🎮 游戏场景示例

### 场景1：玩家出牌触发反应

```javascript
const handleCardPlay = async (card) => {
  try {
    const response = await gameAPI.playCard(roomId, card, substance)

    // 检查是否有反应发生
    if (response.data.reaction) {
      const { type, equation, name, energy } = response.data.reaction
      reactionToast.value?.showToast(type, equation, name, energy)
    }
  } catch (error) {
    console.error('出牌失败', error)
  }
}
```

### 场景2：AI出牌触发反应

```javascript
websocket.on('ai_played', (data) => {
  // AI出牌成功，检查是否有反应
  if (data.reaction) {
    setTimeout(() => {
      reactionToast.value?.showToast(
        data.reaction.type,
        data.reaction.equation,
        data.reaction.name,
        data.reaction.energy
      )
    }, 500)  // 延迟500ms，让动画更自然
  }
})
```

### 场景3：连锁反应

```javascript
// 当发生连锁反应时，延迟显示每个反应
const showChainReactions = (reactions) => {
  reactions.forEach((reaction, index) => {
    setTimeout(() => {
      reactionToast.value?.showToast(
        reaction.type,
        reaction.equation,
        reaction.name,
        reaction.energy
      )
    }, index * 800)  // 每个反应间隔800ms
  })
}
```

---

## ⚙️ 自定义配置

### 修改显示时长

在 `ReactionToast.vue` 中修改：

```javascript
// 找到这行（约第43行）
setTimeout(() => {
  // ...
}, 4000)  // 改成你想要的毫秒数，如 3000 = 3秒
```

### 修改位置

在 `ReactionToast.vue` 的 CSS 中修改：

```css
.reaction-toast-container {
  position: fixed;
  top: 80px;      /* 修改顶部距离 */
  right: 20px;    /* 修改右侧距离 */
  /* 或者改成左侧显示 */
  /* left: 20px; */
}
```

### 限制最大数量

```javascript
const showToast = (type, equation, name, energy = 75) => {
  // 限制最多显示3个
  if (toasts.value.length >= 3) {
    toasts.value.shift()  // 移除最早的
  }

  // ... 原有代码 ...
}
```

---

## 🎨 自定义颜色

### 添加新的反应类型

在 `ReactionToast.vue` 的 CSS 中添加：

```css
/* 自定义反应类型 - 示例：还原反应 */
.reaction-reduction {
  background: linear-gradient(135deg,
    rgba(99, 102, 241, 0.15) 0%,
    rgba(79, 70, 229, 0.25) 50%,
    rgba(67, 56, 202, 0.15) 100%);
  box-shadow:
    0 8px 32px rgba(99, 102, 241, 0.3),
    0 0 0 1px rgba(99, 102, 241, 0.2),
    inset 0 0 60px rgba(99, 102, 241, 0.1);
}

.reaction-reduction .glow-border {
  background: linear-gradient(45deg,
    #6366f1 0%,
    #4f46e5 25%,
    #4338ca 50%,
    #4f46e5 75%,
    #6366f1 100%);
}

.reaction-reduction .energy-bar {
  background: linear-gradient(90deg, #6366f1, #4338ca);
}
```

然后在TypeScript定义中添加类型：

```typescript
type: 'synthesis' | 'decomposition' | 'displacement' | 'combustion' | 'neutralization' | 'reduction'
```

---

## 🐛 故障排查

### 问题1：通知不显示

**检查清单**：
- [ ] 组件是否正确引入？
- [ ] ref 是否正确绑定？
- [ ] showToast 方法是否正确调用？
- [ ] 控制台是否有错误？

```javascript
// 添加调试日志
console.log('Toast Ref:', reactionToast.value)
console.log('Calling showToast...')
reactionToast.value?.showToast('synthesis', 'H₂ + O₂ → H₂O', 'Test', 75)
```

### 问题2：通知位置不对

```css
/* 检查是否有其他z-index冲突 */
.reaction-toast-container {
  z-index: 9999;  /* 确保足够高 */
}
```

### 问题3：字体未加载

确保网络连接正常，或下载字体到本地：

```html
<!-- 或使用本地字体 -->
<style>
@font-face {
  font-family: 'Orbitron';
  src: url('@/assets/fonts/Orbitron-Bold.woff2') format('woff2');
}
</style>
```

---

## 📱 移动端优化建议

### 调整移动端大小

```css
@media (max-width: 640px) {
  .reaction-toast {
    width: calc(100vw - 24px);  /* 自适应宽度 */
    padding: 16px;              /* 减小padding */
  }

  .reaction-equation {
    font-size: 14px;            /* 减小字体 */
  }
}
```

### 触摸关闭功能（可选）

```vue
<div
  @click="closeToast(toast.id)"
  class="reaction-toast"
>
  <!-- ... -->
</div>

<script>
const closeToast = (id) => {
  const idx = toasts.value.findIndex(t => t.id === id)
  if (idx !== -1) {
    toasts.value.splice(idx, 1)
  }
}
</script>
```

---

## 🎯 性能优化建议

### 1. 避免频繁触发

```javascript
let lastToastTime = 0
const MIN_INTERVAL = 500  // 最小间隔500ms

const showReactionToastThrottled = (type, equation, name, energy) => {
  const now = Date.now()
  if (now - lastToastTime < MIN_INTERVAL) {
    return  // 跳过过于频繁的通知
  }
  lastToastTime = now
  reactionToast.value?.showToast(type, equation, name, energy)
}
```

### 2. 限制粒子数量

如果性能较差，可以减少粒子：

```vue
<!-- 从12个减少到6个 -->
<div v-for="i in 6" :key="i" class="particle"></div>
```

### 3. 禁用动画（低端设备）

```javascript
const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches

if (prefersReducedMotion) {
  // 禁用复杂动画
  document.documentElement.style.setProperty('--animation-duration', '0s')
}
```

---

## ✅ 完成！

恭喜！你已经成功集成了化学反应通知系统。

### 下一步
1. 在真实游戏场景中测试
2. 根据反馈调整颜色和动效
3. 添加音效（可选）
4. 记录反应历史（可选）

### 获取帮助
- 查看完整文档：`ReactionToast.README.md`
- 运行演示页面：`ReactionToastDemo.vue`
- 查看示例代码：文档中的"游戏中集成示例"章节

**祝你的Chemistry UNO游戏更加精彩！** 🎮✨
