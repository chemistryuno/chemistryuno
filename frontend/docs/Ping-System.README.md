# Ping监控系统文档

## 概述

全局ping监控系统，在所有页面的右上角实时显示网络延迟，提供网络连接状态的可视化反馈。

## 功能特性

### 实时监控
- 每3秒自动发送ping请求
- 计算往返时间(RTT)
- 维护最近10次延迟记录
- 显示平均延迟

### 状态分级
根据平均延迟自动分级：

| 延迟范围 | 状态 | 颜色 | 文字 |
|---------|------|------|------|
| < 50ms | excellent | 翠绿 | 优秀 |
| 50-100ms | good | 绿色 | 良好 |
| 100-200ms | fair | 黄色 | 一般 |
| > 200ms | poor | 红色 | 较差 |
| 未连接 | disconnected | 灰色 | 未连接 |

### UI设计
- **位置**: 固定在右上角
- **样式**: 玻璃态设计 + 圆角边框
- **动画**: 渐入动画 + 脉冲指示器
- **悬停**: 显示详细信息工具提示
- **响应式**: 移动端自适应缩放

## 技术实现

### 前端架构

#### 1. usePing Composable
**文件**: [usePing.ts](../src/composables/usePing.ts)

**核心功能**:
```typescript
export function usePing() {
  // 状态管理
  const latency = ref<number>(0)
  const isConnected = ref<boolean>(false)
  const pingHistory = ref<number[]>([])

  // 计算属性
  const averageLatency = computed(() => {...})
  const pingStatus = computed<PingStatus>(() => {...})

  // 方法
  const sendPing = () => {...}
  const startPing = () => {...}
  const stopPing = () => {...}

  return {
    latency,
    isConnected,
    pingStatus,
    averageLatency,
    startPing,
    stopPing,
    resetPing
  }
}
```

**工作流程**:
1. 组件挂载后自动开始ping监控
2. 每3秒通过WebSocket发送ping消息
3. 接收pong响应后计算RTT
4. 更新延迟历史并计算平均值
5. 根据平均延迟计算状态等级

#### 2. PingIndicator 组件
**文件**: [PingIndicator.vue](../src/components/PingIndicator.vue)

**UI结构**:
```
┌─────────────────────────┐
│ 🛰️  42ms ⚫ │ ← 主显示
├─────────────────────────┤
│     网络状态            │ ← 悬停提示
│   延迟：42ms           │
│   状态：优秀            │
└─────────────────────────┘
```

**样式特性**:
- 玻璃态背景 (`backdrop-blur-md`)
- 圆角边框 (`rounded-bl-xl`)
- 阴影效果 (`shadow-lg`)
- 状态色彩动态绑定
- 脉冲动画指示器

#### 3. App.vue 集成
**文件**: [App.vue](../src/App.vue)

```vue
<template>
  <div>
    <AnnouncementTicker />
    <PingIndicator />  <!-- 全局ping指示器 -->
    <router-view></router-view>
  </div>
</template>
```

### 后端架构

#### WebSocket消息处理
**文件**: [backend/websocket/client.go](../../backend/websocket/client.go)

**Message结构扩展**:
```go
type Message struct {
    Type      string      `json:"type"`
    RoomID    string      `json:"room_id,omitempty"`
    Data      interface{} `json:"data,omitempty"`
    UID       int         `json:"uid,omitempty"`
    TargetUID int         `json:"target_uid,omitempty"`
    Message   string      `json:"message,omitempty"`
    Timestamp int64       `json:"timestamp,omitempty"` // 新增：时间戳字段
}
```

**ping/pong处理**:
```go
case "ping":
    // 接收到ping请求，立即返回pong响应
    c.Send(Message{
        Type:      "pong",
        Timestamp: msg.Timestamp, // 原样返回时间戳
    })
    return
```

**工作原理**:
1. 客户端发送 `{type: "ping", timestamp: 1708876543210}`
2. 服务器立即响应 `{type: "pong", timestamp: 1708876543210}`
3. 客户端接收pong，计算 `RTT = 当前时间 - timestamp`

## 使用指南

### 基本使用

系统自动启用，无需手动配置。所有页面自动显示ping指示器。

### 在其他组件中使用

如果需要在其他组件中访问ping数据：

```vue
<script setup lang="ts">
import { usePing } from '@/composables'

const { pingStatus, averageLatency, isConnected } = usePing()

// 使用ping数据
watchEffect(() => {
  if (pingStatus.value.status === 'poor') {
    console.warn('网络连接较差')
  }
})
</script>
```

### 自定义样式

PingIndicator使用scoped样式，如需自定义：

```vue
<!-- 在父组件中 -->
<style>
.ping-indicator {
  /* 自定义位置 */
  top: 60px;
  right: 20px;
}
</style>
```

## API参考

### usePing()

**返回值**:

| 属性 | 类型 | 描述 |
|-----|------|------|
| `latency` | `Ref<number>` | 当前延迟(ms) |
| `isConnected` | `Ref<boolean>` | WebSocket连接状态 |
| `pingHistory` | `Ref<number[]>` | 最近10次延迟记录 |
| `averageLatency` | `ComputedRef<number>` | 平均延迟 |
| `pingStatus` | `ComputedRef<PingStatus>` | 状态对象 |
| `startPing()` | `Function` | 开始ping监控 |
| `stopPing()` | `Function` | 停止ping监控 |
| `resetPing()` | `Function` | 重置ping数据 |

**PingStatus 接口**:

```typescript
interface PingStatus {
  latency: number              // 延迟毫秒数
  status: 'excellent' | 'good' | 'fair' | 'poor' | 'disconnected'
  statusText: string           // 状态文本
  statusColor: string          // Tailwind颜色类
}
```

## 配置选项

### 修改Ping间隔

在 `usePing.ts` 中修改：

```typescript
// 默认3秒
const PING_INTERVAL = 3000

// 改为5秒
const PING_INTERVAL = 5000
```

### 修改历史记录长度

```typescript
// 默认保留10次
const MAX_HISTORY = 10

// 改为20次
const MAX_HISTORY = 20
```

### 修改状态阈值

```typescript
const pingStatus = computed<PingStatus>(() => {
  const avgLatency = averageLatency.value

  if (avgLatency < 50) {        // 优秀阈值
    return { status: 'excellent', ... }
  } else if (avgLatency < 100) { // 良好阈值
    return { status: 'good', ... }
  } else if (avgLatency < 200) { // 一般阈值
    return { status: 'fair', ... }
  } else {                       // 较差阈值
    return { status: 'poor', ... }
  }
})
```

## 性能优化

### 1. 自动启动延迟
系统在组件挂载后延迟500ms启动，避免与其他初始化冲突：

```typescript
onMounted(() => {
  setTimeout(() => {
    if (websocket.isConnected()) {
      startPing()
    }
  }, 500)
})
```

### 2. 连接检测
如果WebSocket未连接，最多等待10秒：

```typescript
const checkInterval = setInterval(() => {
  if (websocket.isConnected()) {
    clearInterval(checkInterval)
    startPing()
  }
}, 1000)

setTimeout(() => {
  clearInterval(checkInterval)
}, 10000)
```

### 3. 清理机制
组件卸载时自动停止ping并清理事件监听器：

```typescript
onUnmounted(() => {
  stopPing()
})
```

## 故障排查

### Ping显示为0ms或未连接

**可能原因**:
1. WebSocket未连接
2. 后端未返回pong响应
3. 消息格式不正确

**解决方法**:
1. 检查浏览器控制台WebSocket连接状态
2. 检查后端日志是否有ping/pong消息
3. 使用浏览器开发者工具的网络面板查看WebSocket消息

### Ping值异常高

**可能原因**:
1. 网络延迟高
2. 服务器响应慢
3. WebSocket消息队列阻塞

**解决方法**:
1. 检查网络连接
2. 查看服务器负载
3. 检查是否有大量WebSocket消息发送

### 指示器不显示

**可能原因**:
1. 组件未正确导入
2. CSS被其他样式覆盖
3. Z-index层级问题

**解决方法**:
1. 检查 App.vue 是否导入 PingIndicator
2. 检查浏览器开发者工具的元素面板
3. 调整 z-index 值

## 移动端适配

### 尺寸调整
```css
@media (max-width: 640px) {
  .ping-content {
    @apply px-2 py-1 text-xs;
  }

  .ping-icon {
    @apply w-3 h-3;
  }

  .ping-status-dot {
    @apply w-1.5 h-1.5;
  }
}
```

### 触摸优化
- 移动端隐藏悬停提示
- 增大触摸区域
- 优化字体大小

## 最佳实践

### 1. 不要频繁修改Ping间隔
过于频繁的ping会增加服务器负担和网络流量。建议保持3-5秒间隔。

### 2. 监控状态变化
在关键操作前检查网络状态：

```typescript
const { pingStatus } = usePing()

const criticalOperation = async () => {
  if (pingStatus.value.status === 'disconnected') {
    alert('网络未连接，请检查网络')
    return
  }

  if (pingStatus.value.status === 'poor') {
    const confirmed = confirm('网络状况较差，是否继续？')
    if (!confirmed) return
  }

  // 执行关键操作
}
```

### 3. 日志记录
启用控制台日志以便调试：

```typescript
console.log(`[Ping] RTT: ${rtt}ms, 平均: ${averageLatency.value}ms`)
```

## 未来扩展

可以继续添加的功能：
- 📊 延迟趋势图表
- 📈 网络质量评分
- 🔔 延迟告警通知
- 📝 网络日志记录
- 🎯 区域Ping监控（多服务器）
- 📱 详细网络诊断工具

---

**版本**: 1.0.0
**最后更新**: 2026-02-25
**维护者**: Chemistry UNO Team
