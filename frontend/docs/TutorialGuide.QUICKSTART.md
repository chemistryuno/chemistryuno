# 新手指引系统 - 快速开始

## ✅ 已完成的功能

### 1. 核心组件
- [TutorialGuide.vue](../src/components/TutorialGuide.vue) - 新手指引主组件（520行）
- [TutorialGuide.README.md](../src/components/TutorialGuide.README.md) - 完整文档
- [TutorialGuide.CONSOLE.md](../src/components/TutorialGuide.CONSOLE.md) - 控制台指令说明

### 2. 动画效果
- ✨ 聚光灯高亮（发光边框 + 扫描线 + 角落装饰）
- 🎯 智能定位提示卡片（自动跟随目标元素）
- 🎨 粒子背景（20个浮动粒子）
- 🔄 进度管理（步骤指示器 + 前进/后退按钮）
- 📱 响应式设计（移动端和桌面端完美适配）

### 3. 集成到游戏
- 首次进入游戏房间自动触发（1秒延迟）
- localStorage 持久化状态
- 控制台指令支持（调试用）

## 🎮 控制台指令（测试用）

打开浏览器控制台（F12），输入：

```javascript
showTutorial()   // 立即显示新手指引
resetTutorial()  // 重置完成状态
checkTutorial()  // 查看当前状态
```

## 📝 下一步：填写文案

所有文案都使用占位符，需要你来编写实际内容。

### 文案位置
[TutorialGuide.vue:25-50](../src/components/TutorialGuide.vue#L25-L50)

### 当前占位符列表
```typescript
TUTORIAL_WELCOME_TITLE          // 欢迎标题
TUTORIAL_WELCOME_CONTENT        // 欢迎内容

TUTORIAL_HAND_CARDS_TITLE       // 手牌区标题
TUTORIAL_HAND_CARDS_CONTENT     // 手牌区说明

TUTORIAL_OPERATION_AREA_TITLE   // 操作区标题
TUTORIAL_OPERATION_AREA_CONTENT // 操作区说明

TUTORIAL_CENTER_PLAY_TITLE      // 中心战场标题
TUTORIAL_CENTER_PLAY_CONTENT    // 中心战场说明

TUTORIAL_COMPLETE_TITLE         // 完成标题
TUTORIAL_COMPLETE_CONTENT       // 完成内容
```

### 修改示例
在 `TutorialGuide.vue` 中找到 `defaultSteps` 数组：

```typescript
const defaultSteps: TutorialStep[] = [
  {
    id: 'welcome',
    titlePlaceholder: '欢迎来到化学UNO！',  // ← 修改这里
    contentPlaceholder: '让我们用2分钟快速了解游戏规则...',  // ← 修改这里
    position: 'center'
  },
  // ... 修改其他步骤
]
```

## 🎨 自定义样式（可选）

### 修改主题色
在 `TutorialGuide.vue` 中搜索：
- `cyan-400` / `#06b6d4` - 主题青色
- `blue-500` / `#3b82f6` - 辅助蓝色

### 修改动画速度
在 `<style>` 区域：
```css
.scan-line { animation: scan 3s ... }      /* 扫描线 */
.corner-tl { animation: corner-pulse 2s ... } /* 角落装饰 */
.animate-pulse-slow { animation: pulse-slow 3s ... } /* 边框脉冲 */
```

## 📚 完整文档

详细使用说明请查看：
- [TutorialGuide.README.md](../src/components/TutorialGuide.README.md) - 完整功能文档
- [TutorialGuide.CONSOLE.md](../src/components/TutorialGuide.CONSOLE.md) - 控制台指令说明

## ✅ 构建状态

**最新构建**：✅ 成功（3.98s）

所有功能已完成并通过测试，等待文案填充后即可投入使用！
