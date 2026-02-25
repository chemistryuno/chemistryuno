# CSS 架构文档

## 概述

Chemistry UNO 采用模块化的 CSS 架构，将样式分离为独立的模块文件，提高可维护性和复用性。

## 架构设计

```text
styles/
├── main.css              # 主入口文件（导入所有模块）
├── variables.css         # CSS 变量和设计令牌
├── animations.css        # 动画库
├── components.css        # 组件样式库
├── mobile-game.css       # 移动端游戏样式（已存在）
└── chemical-keyboard.css # 化学键盘组件样式
```

## 模块说明

### 1. main.css - 主入口文件

**职责**：
- 导入所有 CSS 模块
- 全局重置样式
- Tailwind CSS 集成
- 辅助功能样式

**使用**：
```typescript
// main.ts
import './styles/main.css'
```

### 2. variables.css - CSS 变量

**职责**：
- 定义设计令牌（颜色、间距、字体等）
- 主题系统（亮色/暗色）
- 响应式断点
- Z-index 层级管理

**变量分类**：
- 颜色系统：`--color-*`
- 间距系统：`--spacing-*`
- 字体系统：`--text-*`, `--font-*`, `--leading-*`
- 圆角系统：`--radius-*`
- 阴影系统：`--shadow-*`
- Z-index 系统：`--z-*`
- 过渡系统：`--duration-*`, `--ease-*`

**使用示例**：
```css
.my-button {
  color: var(--color-primary);
  padding: var(--spacing-md);
  border-radius: var(--radius-lg);
  transition: all var(--duration-base) var(--ease-out);
}
```

### 3. animations.css - 动画库

**职责**：
- 统一管理所有动画效果
- 入场动画、循环动画、特效动画
- 工具类快速应用动画

**动画分类**：

#### 入场动画
- `slideUp` / `slideDown` - 滑入
- `slideInFromLeft` / `slideInFromRight` - 侧滑
- `fadeIn` - 淡入
- `zoomIn` - 缩放
- `bounceIn` - 弹入

#### 循环动画
- `pulse` - 脉冲
- `spin` - 旋转
- `bounce` - 弹跳
- `shake` - 抖动
- `glow` - 发光

#### 科技风动画
- `scanLine` - 扫描线
- `glowPulse` - 发光脉冲
- `float` - 浮动
- `orbit` - 轨道运动

#### 工具类
```html
<div class="animate-slide-up">滑入效果</div>
<div class="animate-pulse">脉冲效果</div>
<div class="animate-spin">旋转效果</div>
```

### 4. components.css - 组件样式库

**职责**：
- 可复用的组件样式类
- 按钮、卡片、输入框等基础组件
- 模态框、徽章、提示框等复合组件
- 滚动条、渐变、阴影等特效

**组件分类**：

#### 按钮
```html
<button class="btn-primary">主按钮</button>
<button class="btn-secondary">次要按钮</button>
<button class="btn-ghost">幽灵按钮</button>
<button class="btn-touch">触摸优化按钮</button>
```

#### 卡片
```html
<div class="card-base">基础卡片</div>
<div class="card-glass">玻璃态卡片</div>
<div class="card-hover">悬停效果卡片</div>
```

#### 输入框
```html
<input class="input-base" />
<input class="input-mobile" /> <!-- 防止iOS缩放 -->
<input class="input-error" /> <!-- 错误状态 -->
```

#### 模态框
```html
<div class="modal-overlay">
  <div class="modal-content">内容</div>
</div>
```

#### 徽章
```html
<span class="badge-primary">主要</span>
<span class="badge-success">成功</span>
<span class="badge-warning">警告</span>
<span class="badge-danger">危险</span>
```

#### 滚动条
```html
<div class="custom-scrollbar">美化滚动条</div>
<div class="custom-scrollbar-hidden">隐藏滚动条</div>
```

#### 特效
```html
<div class="glass">玻璃态</div>
<div class="gradient-blue">蓝色渐变</div>
<div class="shadow-glow-blue">蓝色发光</div>
```

### 5. chemical-keyboard.css - 化学键盘样式

**职责**：
- ChemicalKeyboard 组件的专用样式
- 元素按钮、数字键盘、括号按钮样式
- 拖拽手柄、横向滚动样式

**组件导入**：
```vue
<script setup>
import '@/styles/chemical-keyboard.css'
</script>
```

## 使用指南

### 方式一：使用预定义的工具类

推荐使用预定义的工具类，快速构建界面：

```html
<button class="btn-primary animate-bounce-in">
  点击我
</button>

<div class="card-glass shadow-lg">
  <h2 class="text-2xl font-bold mb-4">标题</h2>
  <p class="text-slate-600 dark:text-slate-300">内容</p>
</div>
```

### 方式二：使用 CSS 变量

在自定义样式中使用 CSS 变量：

```css
.my-component {
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
  padding: var(--spacing-lg);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-lg);
  transition: all var(--duration-base) var(--ease-out);
}
```

### 方式三：结合 Tailwind CSS

混合使用 Tailwind 和自定义类：

```html
<div class="p-4 rounded-lg card-glass animate-fade-in">
  <button class="btn-primary w-full">提交</button>
</div>
```

## 样式命名规范

### BEM 命名法

对于自定义组件样式，使用 BEM 命名：

```css
/* Block */
.chemical-keyboard { }

/* Element */
.chemical-keyboard__button { }

/* Modifier */
.chemical-keyboard__button--active { }
```

### 工具类命名

工具类使用简洁的语义化名称：

```css
.btn-primary { }     /* 主按钮 */
.card-glass { }      /* 玻璃卡片 */
.animate-spin { }    /* 旋转动画 */
```

## 最佳实践

### 1. 优先使用工具类

```html
<!-- ✅ 推荐 -->
<button class="btn-primary">按钮</button>

<!-- ❌ 不推荐 -->
<button style="background: blue; color: white;">按钮</button>
```

### 2. 使用 CSS 变量保持一致性

```css
/* ✅ 推荐 */
.my-button {
  color: var(--color-primary);
  padding: var(--spacing-md);
}

/* ❌ 不推荐 */
.my-button {
  color: #3b82f6;
  padding: 16px;
}
```

### 3. 避免深层嵌套

```css
/* ✅ 推荐 */
.card { }
.card-header { }
.card-body { }

/* ❌ 不推荐 */
.card .header .title span { }
```

### 4. 使用作用域样式

在 Vue 组件中使用 scoped 样式：

```vue
<style scoped>
.my-component {
  /* 仅影响当前组件 */
}
</style>
```

### 5. 组件样式优先顺序

```vue
<script setup>
// 1. 导入全局样式（main.css 已在 main.ts 中导入）
// 2. 导入组件专用样式
import '@/styles/chemical-keyboard.css'
</script>

<style scoped>
/* 3. 组件局部样式 */
.local-class {
  color: var(--color-primary);
}
</style>
```

## 响应式设计

### 使用 Tailwind 断点

```html
<div class="text-sm md:text-base lg:text-lg">
  响应式文本
</div>
```

### 使用 CSS 媒体查询

```css
.my-component {
  font-size: var(--text-base);
}

@media (min-width: 768px) {
  .my-component {
    font-size: var(--text-lg);
  }
}
```

## 性能优化

### 1. 避免昂贵的属性

```css
/* ❌ 性能差 */
.box {
  box-shadow: 0 0 50px rgba(0,0,0,0.5);
  filter: blur(10px);
}

/* ✅ 性能好 */
.box {
  box-shadow: var(--shadow-md);
  backdrop-filter: blur(10px); /* 配合 will-change 使用 */
  will-change: transform;
}
```

### 2. 使用 CSS 变量减少重复计算

```css
:root {
  --card-padding: 1rem;
  --card-shadow: 0 4px 6px rgba(0,0,0,0.1);
}

.card {
  padding: var(--card-padding);
  box-shadow: var(--card-shadow);
}
```

### 3. 优化动画性能

```css
/* ✅ 使用 transform 和 opacity（GPU 加速） */
.animate {
  transform: translateX(100px);
  opacity: 0.5;
  will-change: transform, opacity;
}

/* ❌ 避免动画 top、left、width、height */
.animate-bad {
  left: 100px;
  width: 200px;
}
```

## 暗色模式

### 使用 Tailwind 暗色类

```html
<div class="bg-white dark:bg-slate-900 text-slate-900 dark:text-white">
  自动适配暗色模式
</div>
```

### 使用 CSS 变量

```css
.my-component {
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
  border-color: var(--color-border);
}
```

## 迁移指南

### 从内联样式迁移

**步骤**：
1. 识别重复的样式模式
2. 在 components.css 中创建工具类
3. 替换内联样式为工具类

**示例**：
```html
<!-- 迁移前 -->
<button style="background: #3b82f6; color: white; padding: 0.5rem 1rem; border-radius: 0.5rem;">
  按钮
</button>

<!-- 迁移后 -->
<button class="btn-primary">
  按钮
</button>
```

### 从组件 scoped 样式迁移

**步骤**：
1. 识别可复用的样式
2. 提取到 components.css
3. 更新组件使用工具类

**示例**：
```vue
<!-- 迁移前 -->
<template>
  <div class="my-card">内容</div>
</template>

<style scoped>
.my-card {
  background: white;
  border-radius: 1rem;
  box-shadow: 0 4px 6px rgba(0,0,0,0.1);
  padding: 1.5rem;
}
</style>

<!-- 迁移后 -->
<template>
  <div class="card-base p-6">内容</div>
</template>
```

## 故障排查

### 样式未生效

1. **检查导入顺序**
   ```typescript
   // main.ts
   import './styles/main.css' // 必须在组件导入之前
   ```

2. **检查 CSS 作用域**
   ```vue
   <!-- 如果使用 scoped，确保类名正确 -->
   <style scoped>
   .local-class { } /* 仅影响当前组件 */
   </style>
   ```

3. **检查优先级**
   ```css
   /* 使用 !important 时要谨慎 */
   .override {
     color: red !important; /* 尽量避免 */
   }
   ```

### 暗色模式不工作

1. **确保 HTML 有 dark 类**
   ```typescript
   // 检查是否正确切换
   document.documentElement.classList.toggle('dark')
   ```

2. **使用正确的暗色类**
   ```html
   <!-- ✅ 正确 -->
   <div class="bg-white dark:bg-slate-900">

   <!-- ❌ 错误 -->
   <div class="bg-white bg-dark-slate-900">
   ```

## 未来扩展

可以继续添加的模块：
- `layouts.css` - 布局系统
- `typography.css` - 排版系统
- `forms.css` - 表单系统
- `tables.css` - 表格系统
- `print.css` - 打印样式

---

**版本**: 1.0.0
**最后更新**: 2026-02-25
**维护者**: Chemistry UNO Team
