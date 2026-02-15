# 化学反应浮窗通知系统

## 🎨 设计理念

基于**实验室科技风 + 现代玩具化设计**，创造了一个充满科学感的反应通知系统。

### 核心特色
- 🔬 **原子模型动画** - 旋转的电子轨道和脉冲核心
- ✨ **粒子系统** - 模拟化学反应中的分子运动
- 💫 **玻璃态效果** - 毛玻璃背景 + 发光边框
- 🌈 **科学配色** - 5种反应类型，每种独特的颜色方案
- ⚡ **能量条动画** - 展示反应强度

---

## 📦 组件使用

### 1. 引入组件

```vue
<template>
  <div>
    <!-- 你的游戏内容 -->
    <ReactionToast ref="toastRef" />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import ReactionToast from '@/components/ReactionToast.vue'

const toastRef = ref(null)
</script>
```

### 2. 触发通知

```javascript
// 合成反应 - 蓝绿色科技感
toastRef.value?.showToast(
  'synthesis',
  'H₂ + O₂ → H₂O',
  'Synthesis Reaction',
  85
)

// 分解反应 - 橙红色能量爆发
toastRef.value?.showToast(
  'decomposition',
  'H₂O → H₂ + O₂',
  'Decomposition',
  75
)

// 置换反应 - 紫色魔幻感
toastRef.value?.showToast(
  'displacement',
  'Zn + CuSO₄ → ZnSO₄ + Cu',
  'Displacement Reaction',
  90
)

// 燃烧反应 - 火焰橙黄色
toastRef.value?.showToast(
  'combustion',
  'CH₄ + O₂ → CO₂ + H₂O',
  'Combustion Reaction',
  95
)

// 中和反应 - 青绿色平衡感
toastRef.value?.showToast(
  'neutralization',
  'HCl + NaOH → NaCl + H₂O',
  'Neutralization',
  70
)
```

---

## 🎯 参数说明

### showToast(type, equation, name, energy)

| 参数 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `type` | string | 反应类型 | 'synthesis' \| 'decomposition' \| 'displacement' \| 'combustion' \| 'neutralization' |
| `equation` | string | 化学方程式 | 'H₂ + O₂ → H₂O' |
| `name` | string | 反应名称 | 'Synthesis Reaction' |
| `energy` | number | 能量百分比(0-100) | 85 |

---

## 🌈 反应类型配色

### 1. Synthesis (合成反应) - 蓝绿科技
```css
颜色: #06b6d4 → #3b82f6
场景: 元素结合，构建新物质
视觉: 蓝绿渐变 + 科技光效
```

### 2. Decomposition (分解反应) - 橙红爆发
```css
颜色: #f97316 → #dc2626
场景: 物质分解，能量释放
视觉: 橙红渐变 + 爆炸效果
```

### 3. Displacement (置换反应) - 紫色魔幻
```css
颜色: #a855f7 → #7e22ce
场景: 元素交换，魔法般转换
视觉: 紫色渐变 + 神秘光晕
```

### 4. Combustion (燃烧反应) - 火焰橙黄
```css
颜色: #fb923c → #f59e0b
场景: 燃烧放热，火焰效果
视觉: 橙黄渐变 + 火焰闪烁
```

### 5. Neutralization (中和反应) - 青绿平衡
```css
颜色: #10b981 → #047857
场景: 酸碱中和，恢复平衡
视觉: 青绿渐变 + 平静波纹
```

---

## ✨ 动效清单

### 入场动画
- **弹性飞入** - 从右侧弹跳进入
- **缩放脉冲** - 0.8 → 1.05 → 1
- **旋转修正** - 5° → -2° → 0°

### 持续动画
- **发光边框** - 4秒循环旋转
- **粒子上升** - 12个粒子随机浮动
- **原子旋转** - 8秒完整旋转
- **电子轨道** - 2秒环绕运动
- **核心脉冲** - 2秒呼吸效果
- **文字发光** - 2秒光晕循环
- **能量填充** - 4秒渐进填充
- **能量闪光** - 1.5秒光带滑过

### 离场动画
- **飞出消失** - 向右旋转飞出
- **透明度** - 1 → 0
- **缩放** - 1 → 0.8

---

## 🎮 游戏中集成示例

### GameRoom.vue 中使用

```vue
<template>
  <div class="game-room">
    <!-- 游戏界面 -->
    <div class="game-board">
      <!-- ... 游戏内容 ... -->
    </div>

    <!-- 反应通知系统 -->
    <ReactionToast ref="reactionToast" />
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import ReactionToast from '@/components/ReactionToast.vue'
import { useGameStore } from '@/stores/game'

const reactionToast = ref(null)
const gameStore = useGameStore()

// 监听游戏事件
watch(() => gameStore.lastReaction, (reaction) => {
  if (reaction) {
    reactionToast.value?.showToast(
      reaction.type,
      reaction.equation,
      reaction.name,
      reaction.energy
    )
  }
})

// 或者通过 WebSocket 监听反应事件
import { websocket } from '@/utils/websocket'

websocket.on('reaction_occurred', (data) => {
  reactionToast.value?.showToast(
    data.type,
    data.equation,
    data.name,
    data.energy || 75
  )
})
</script>
```

---

## 📱 响应式优化

### 移动端适配
- ✅ 自适应宽度 (最大100vw - 24px)
- ✅ 字体缩小 (16px → 14px)
- ✅ 图标缩小 (56px → 48px)
- ✅ Padding 减小 (20px → 16px)
- ✅ 位置调整 (top: 60px, 左右12px)

### 暗色模式
- ✅ 边框透明度调整
- ✅ 能量条背景优化
- ✅ 自动适配系统主题

---

## 🎨 字体选择

### Orbitron (反应方程式)
- **风格**: 未来科技感，几何化
- **用途**: 化学方程式显示
- **权重**: 900 (黑体)

### JetBrains Mono (反应名称)
- **风格**: 等宽字体，代码感
- **用途**: 反应类型标签
- **权重**: 700 (粗体)

---

## 🚀 性能优化

### CSS优化
- ✅ 使用 `transform` 和 `opacity` 实现动画
- ✅ GPU加速 (`will-change` 自动应用)
- ✅ 减少重绘和回流

### 内存管理
- ✅ 4秒后自动销毁toast
- ✅ 使用 TransitionGroup 优化列表动画
- ✅ 粒子数量控制在12个以内

### 可访问性
- ✅ 使用语义化标签
- ✅ 适当的对比度
- ✅ 支持键盘导航（可扩展）

---

## 🎯 最佳实践

### 1. 控制频率
```javascript
// 避免短时间内大量通知
let lastToastTime = 0
const MIN_TOAST_INTERVAL = 1000 // 最小间隔1秒

function showReactionToast(type, equation, name, energy) {
  const now = Date.now()
  if (now - lastToastTime < MIN_TOAST_INTERVAL) {
    return // 跳过过于频繁的通知
  }
  lastToastTime = now
  reactionToast.value?.showToast(type, equation, name, energy)
}
```

### 2. 能量值映射
```javascript
// 根据反应类型设置合理的能量值
const energyLevels = {
  synthesis: 70-85,      // 合成反应中等能量
  decomposition: 75-90,  // 分解需要能量输入
  displacement: 65-80,   // 置换中等能量
  combustion: 90-100,    // 燃烧高能反应
  neutralization: 60-75  // 中和低能反应
}
```

### 3. 方程式格式化
```javascript
// 使用Unicode上下标字符
const formatEquation = (equation) => {
  return equation
    .replace(/(\d+)/g, (match) => {
      // 将数字转换为下标
      const subscripts = '₀₁₂₃₄₅₆₇₈₉'
      return match.split('').map(d => subscripts[d]).join('')
    })
}

// 示例
formatEquation('H2 + O2 → H2O') // 输出: H₂ + O₂ → H₂O
```

---

## 🔧 扩展建议

### 未来可添加功能
1. **音效支持** - 不同反应播放不同音效
2. **振动反馈** - 移动端触觉反馈
3. **可关闭按钮** - 允许用户手动关闭
4. **堆叠限制** - 最多显示3-4个toast
5. **优先级系统** - 重要反应可覆盖普通通知
6. **自定义位置** - 支持左上、右上、中央等
7. **反应历史** - 点击查看详细历史记录

---

## 📸 效果预览

### 视觉特点
- 🔵 **玻璃态质感** - 毛玻璃背景 + 高斯模糊
- ⚡ **发光边框** - 渐变色彩 + 循环动画
- ✨ **粒子特效** - 12个粒子模拟分子运动
- 🌀 **原子模型** - 旋转的电子轨道
- 📊 **能量可视化** - 动态填充的进度条
- 💫 **波纹效果** - 入场时的冲击波

### 动效时序
```
0.0s - Toast 弹性飞入
0.2s - 粒子开始上升
0.4s - 能量条开始填充
0.6s - 入场动画完成
...持续动画...
3.6s - 准备离场
4.0s - Toast 飞出消失
```

---

## 🎉 总结

这是一个充满科学感和视觉冲击力的化学反应通知系统，完美融合了：
- 🧪 实验室科技风格
- 🎮 现代游戏设计
- ⚛️ 物理化学美学
- ✨ 流畅的动效体验

**立即集成到你的Chemistry UNO游戏中，让每一次化学反应都成为视觉盛宴！** 🚀
