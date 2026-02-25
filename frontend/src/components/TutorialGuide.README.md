# 新手指引系统 (Tutorial Guide)

## 功能概述

精美的新手指引系统，包含：
- ✨ **聚光灯高亮**：动态定位并高亮目标元素
- 🎯 **智能定位**：提示卡片自动跟随目标元素（上/下/左/右/居中）
- 🎨 **科技风动画**：
  - 扫描线效果
  - 发光边框脉冲
  - 角落装饰动画
  - 粒子浮动背景
  - 弹性缩放/滑入动效
- 📱 **响应式设计**：完美适配移动端和桌面端
- 🔄 **步骤管理**：支持前进/后退，显示进度条

## 文案占位符说明

所有文案都使用占位符，方便后续统一管理和国际化。

### 默认步骤配置

```typescript
const defaultSteps: TutorialStep[] = [
  {
    id: 'welcome',
    titlePlaceholder: 'TUTORIAL_WELCOME_TITLE',        // 欢迎标题
    contentPlaceholder: 'TUTORIAL_WELCOME_CONTENT',    // 欢迎内容
    position: 'center'
  },
  {
    id: 'hand-cards',
    titlePlaceholder: 'TUTORIAL_HAND_CARDS_TITLE',     // 手牌区标题
    contentPlaceholder: 'TUTORIAL_HAND_CARDS_CONTENT', // 手牌区说明
    targetSelector: '.hand-container-mobile',           // 目标元素
    position: 'top',                                    // 提示框在目标上方
    spotlightRadius: 180
  },
  {
    id: 'operation-area',
    titlePlaceholder: 'TUTORIAL_OPERATION_AREA_TITLE',
    contentPlaceholder: 'TUTORIAL_OPERATION_AREA_CONTENT',
    targetSelector: '.operation-area',
    position: 'bottom',
    spotlightRadius: 200
  },
  {
    id: 'center-play',
    titlePlaceholder: 'TUTORIAL_CENTER_PLAY_TITLE',
    contentPlaceholder: 'TUTORIAL_CENTER_PLAY_CONTENT',
    targetSelector: '.center-play-area',
    position: 'bottom',
    spotlightRadius: 220
  },
  {
    id: 'complete',
    titlePlaceholder: 'TUTORIAL_COMPLETE_TITLE',
    contentPlaceholder: 'TUTORIAL_COMPLETE_CONTENT',
    position: 'center'
  }
]
```

### 文案替换示例

在 `TutorialGuide.vue` 中，找到对应的占位符替换为实际文案：

```vue
<!-- 当前：显示占位符 -->
<h3 class="text-xl sm:text-2xl font-black text-white tracking-tight">
  {{ tutorialSteps[currentStep].titlePlaceholder }}
</h3>

<p class="text-slate-300 text-sm sm:text-base leading-relaxed mb-8">
  {{ tutorialSteps[currentStep].contentPlaceholder }}
</p>
```

**方案A：直接替换默认步骤中的占位符**
```typescript
const defaultSteps: TutorialStep[] = [
  {
    id: 'welcome',
    titlePlaceholder: '欢迎来到化学UNO',
    contentPlaceholder: '让我们用2分钟快速了解游戏规则，开启你的化学冒险之旅！',
    position: 'center'
  },
  // ...
]
```

**方案B：创建文案映射对象（推荐）**
```typescript
// 在 TutorialGuide.vue 中添加
const tutorialTexts: Record<string, { title: string, content: string }> = {
  'TUTORIAL_WELCOME_TITLE': {
    title: '欢迎来到化学UNO',
    content: '让我们用2分钟快速了解游戏规则，开启你的化学冒险之旅！'
  },
  'TUTORIAL_HAND_CARDS_TITLE': {
    title: '你的手牌区',
    content: '这里是你的手牌，点击卡片可以出牌。每张卡片代表一种化学元素或特殊效果。'
  },
  'TUTORIAL_OPERATION_AREA_TITLE': {
    title: '操作区域',
    content: '在你的回合，这里会显示可用操作：出牌、合成物质、抽牌等。注意倒计时！'
  },
  'TUTORIAL_CENTER_PLAY_TITLE': {
    title: '卡牌战场',
    content: '中心区域显示当前场上的卡牌和玩家信息。观察其他玩家的状态来制定策略。'
  },
  'TUTORIAL_COMPLETE_TITLE': {
    title: '准备开始！',
    content: '你已经掌握了基本操作。现在开始你的化学UNO之旅吧！'
  }
}

// 在模板中使用
<h3>{{ tutorialTexts[tutorialSteps[currentStep].titlePlaceholder].title }}</h3>
<p>{{ tutorialTexts[tutorialSteps[currentStep].titlePlaceholder].content }}</p>
```

## 动画效果说明

### 1. 聚光灯高亮
- 自动计算目标元素位置和大小
- 发光边框（青色 `#06b6d4`）脉冲动画（3秒周期）
- 扫描线上下滑动（3秒周期）
- 四角装饰呼吸效果（2秒周期）

### 2. 提示卡片
- **居中模式**：弹性缩放进入（0.5秒）
- **跟随模式**：滑入动画（0.4秒）
- 带箭头指示器指向目标元素
- 进度条实时显示当前步骤

### 3. 背景粒子
- 20个粒子从底部浮动到顶部（8秒周期）
- 淡入淡出效果
- 随机水平偏移

### 4. 卡片装饰
- 顶部渐变条
- 左上角青色发光球体（高斯模糊）
- 径向渐变背景
- 3个位置的 ping 动画粒子

## 触发逻辑

### 自动触发（首次进入）
```typescript
// GameRoom.vue 中已实现
const checkFirstTimeUser = () => {
  const hasSeenTutorial = localStorage.getItem('chemistry-uno-tutorial-completed')
  if (!hasSeenTutorial) {
    setTimeout(() => {
      showTutorial.value = true
    }, 1000)  // 延迟1秒，等待页面元素加载
  }
}
```

### 手动触发
可以在设置菜单或帮助页面添加按钮：
```vue
<button @click="showTutorial = true">
  重新查看新手教程
</button>
```

## Props & Events

### Props
```typescript
interface Props {
  show: boolean          // 控制显示/隐藏
  steps?: TutorialStep[] // 自定义步骤（可选）
}
```

### Events
```typescript
interface Emits {
  close: []    // 用户点击跳过或关闭按钮
  complete: [] // 用户完成所有步骤
}
```

### TutorialStep 类型
```typescript
interface TutorialStep {
  id: string                // 步骤唯一标识
  titlePlaceholder: string  // 标题占位符
  contentPlaceholder: string // 内容占位符
  targetSelector?: string   // CSS选择器（可选）
  position?: 'top' | 'bottom' | 'left' | 'right' | 'center' // 提示框位置
  spotlightRadius?: number  // 聚光灯半径（可选，默认150）
}
```

## 自定义步骤示例

```vue
<TutorialGuide
  :show="showTutorial"
  :steps="customSteps"
  @close="handleClose"
  @complete="handleComplete"
/>

<script setup>
const customSteps = [
  {
    id: 'custom-1',
    titlePlaceholder: '自定义标题1',
    contentPlaceholder: '这是自定义内容...',
    targetSelector: '.my-custom-element',
    position: 'right',
    spotlightRadius: 150
  },
  // ...更多步骤
]
</script>
```

## CSS类名说明

### 已添加的目标元素类名
- `.hand-container-mobile` - 手牌区域
- `.operation-area` - 操作区域（回合操作按钮）
- `.center-play-area` - 中心卡牌战场

### 添加更多目标
如需高亮其他元素，在对应HTML元素上添加类名，然后在步骤中使用 `targetSelector`。

## 注意事项

1. **元素必须已渲染**：确保目标元素在显示教程前已渲染到DOM
2. **响应式适配**：聚光灯和提示框会自动响应窗口大小变化
3. **z-index**：教程层级为 `z-[9999]`，确保在所有内容之上
4. **性能优化**：使用了 `nextTick` 和防抖确保计算准确
5. **LocalStorage**：使用 `chemistry-uno-tutorial-completed` 标记完成状态

## 样式调整

### 修改主题色
在 `TutorialGuide.vue` 中搜索并替换：
- `cyan-400` / `#06b6d4` - 主题青色
- `blue-500` / `#3b82f6` - 辅助蓝色
- `slate-900` / `#0f172a` - 深色背景

### 修改动画速度
```css
/* 扫描线 */
.scan-line { animation: scan 3s ... }  /* 改为 2s 更快 */

/* 角落脉冲 */
.corner-tl { animation: corner-pulse 2s ... }

/* 边框脉冲 */
.animate-pulse-slow { animation: pulse-slow 3s ... }
```

## 效果预览

### 居中欢迎卡片
- 卡片从中心弹性缩放进入
- 左上角图标渐变背景
- 右上角关闭按钮
- 底部进度指示器

### 定位提示卡片
- 聚光灯自动定位到目标元素
- 提示卡片出现在目标上方/下方/左侧/右侧
- 箭头指向目标
- 扫描线在聚光灯区域滑动

### 背景效果
- 半透明黑色遮罩 + 模糊
- 粒子从底部缓慢上升
- 全屏覆盖确保引导注意力

---

**开发状态**：✅ 动画系统完成，等待文案填充

**下一步**：将占位符替换为最终文案，测试各步骤流畅度
