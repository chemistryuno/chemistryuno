# 新手指引系统 - 精确定位优化

## 🎯 优化内容

### 1. 聚光灯精确匹配
**改进前**：使用固定的 `spotlightRadius` 参数（150-250px），导致高亮区域过大
**改进后**：自动匹配目标元素的实际尺寸，仅添加 16px 的适度 padding

#### 代码对比
```typescript
// ❌ 旧版本 - 固定半径
const radius = step.spotlightRadius || 150
spotlightStyle.value = {
  width: `${Math.max(rect.width, radius * 2)}px`,
  height: `${Math.max(rect.height, radius * 2)}px`,
}

// ✅ 新版本 - 精确匹配
const padding = 16
spotlightStyle.value = {
  width: `${rect.width + padding * 2}px`,
  height: `${rect.height + padding * 2}px`,
}
```

### 2. 提示框智能边界检测
**新增功能**：自动检测提示框是否会超出屏幕，如果超出则自动调整位置

#### 边界检测逻辑
- **顶部溢出** → 自动切换到 `bottom` 位置
- **底部溢出** → 自动切换到 `top` 位置
- **左侧溢出** → 自动切换到 `right` 位置
- **右侧溢出** → 自动切换到 `left` 位置
- **水平居中溢出** → 限制在屏幕边缘内（保持 16px 边距）

#### 关键参数
```typescript
const tooltipEstimatedWidth = 320   // 提示框预估宽度
const tooltipEstimatedHeight = 200  // 提示框预估高度
const edgePadding = 16              // 距离屏幕边缘的最小距离
```

### 3. 步骤配置简化
**移除参数**：不再需要手动指定 `spotlightRadius`

#### 配置对比
```typescript
// ❌ 旧版本 - 需要手动设置半径
{
  id: 'create-room',
  targetSelector: '[data-tutorial="create-room"]',
  position: 'bottom',
  spotlightRadius: 150  // ← 需要手动调整
}

// ✅ 新版本 - 自动计算
{
  id: 'create-room',
  targetSelector: '[data-tutorial="create-room"]',
  position: 'bottom'  // ← 简洁清晰
}
```

## 📐 视觉效果

### 聚光灯高亮
- 紧贴按钮边缘
- 16px 的视觉呼吸空间
- 发光边框清晰可见
- 角落装饰精准定位

### 提示框定位
- 首选用户指定的方向（top/bottom/left/right）
- 检测到溢出自动翻转到相反方向
- 始终保持在视口内
- 与聚光灯协调对齐

## 🎨 动画效果（保持不变）
- ✨ 发光边框脉冲（3秒周期）
- 📡 扫描线上下滑动（3秒周期）
- 🔷 四角装饰呼吸（2秒周期）
- ✨ 20个粒子背景浮动

## 🚀 使用方法

### 控制台测试（游戏房间）
```javascript
showTutorial()   // 显示新手指引
resetTutorial()  // 重置状态
checkTutorial()  // 查看状态
```

### 控制台测试（大厅）
```javascript
showLobbyTutorial()   // 显示大厅指引
resetLobbyTutorial()  // 重置状态
checkLobbyTutorial()  // 查看状态
```

## 📊 技术细节

### 坐标计算流程
1. **获取目标元素** - `document.querySelector(targetSelector)`
2. **获取尺寸位置** - `element.getBoundingClientRect()`
3. **计算聚光灯中心** - `(rect.left + rect.width / 2, rect.top + rect.height / 2)`
4. **设置精确尺寸** - `rect.width + 32px, rect.height + 32px`
5. **计算提示框初始位置** - 根据 `position` 参数
6. **边界检测** - 检查是否超出 viewport
7. **自动调整** - 超出则切换到相反方向
8. **最终修正** - 水平居中时限制在屏幕内

### 响应式支持
- 监听窗口 resize 事件
- 自动重新计算所有坐标
- 确保在任意分辨率下正确显示

## 🔧 后续可优化项

1. **动态尺寸检测** - 通过 ref 获取提示框真实宽高，替代预估值
2. **移动端适配** - 移动端可能需要不同的 padding 和边距
3. **过渡动画** - 位置切换时添加平滑过渡效果
4. **Z-index 管理** - 确保在复杂布局中正确显示

## ✅ 构建状态

**最新构建**：✅ 成功（3.98s）
**文件大小**：前端 247.45 KB，后端 27.01 MB
**状态**：生产就绪

---

**更新时间**：2026-02-24
**版本**：Chemistry UNO V1.2.0 Mendeleef
