# 控制台指令 - 新手指引调试

## 快速使用

在游戏房间页面打开浏览器控制台（F12），输入以下指令：

### 1. 显示新手指引
```javascript
showTutorial()
```
立即显示新手指引，无论是否已完成过教程。用于测试动画效果和文案。

### 2. 重置教程状态
```javascript
resetTutorial()
```
清除教程完成标记，刷新页面后将重新自动显示教程（首次用户体验）。

### 3. 查看教程状态
```javascript
checkTutorial()
```
查看当前教程完成状态和显示状态。

返回对象：
```javascript
{
  completed: true/false,  // 是否已完成教程
  showing: true/false     // 当前是否正在显示
}
```

## 使用场景

### 测试文案修改效果
1. 修改 [TutorialGuide.vue](./TutorialGuide.vue) 中的文案占位符
2. 刷新页面
3. 在控制台输入 `showTutorial()` 查看效果

### 测试首次用户体验
1. 在控制台输入 `resetTutorial()`
2. 刷新页面
3. 观察1秒后自动弹出的新手指引

### 调试动画效果
1. 输入 `showTutorial()` 启动教程
2. 使用浏览器开发者工具查看元素和动画
3. 可以多次调用测试不同步骤

## 控制台输出示例

进入游戏房间后，控制台会自动显示：

```
🎮 Chemistry UNO - 控制台指令
可用指令:
  showTutorial()  - 显示新手指引
  resetTutorial() - 重置教程状态（需刷新页面生效）
  checkTutorial() - 查看教程完成状态
```

## 技术实现

这些指令在 [GameRoom.vue:964-979](../pages/GameRoom.vue#L964-L979) 中注册到全局 `window` 对象，仅在开发和测试时使用。

生产环境中这些指令仍然可用，不会影响用户体验，因为：
- 普通用户不会打开控制台
- 指令名称清晰，不会意外调用
- 不会自动执行，需要手动输入
