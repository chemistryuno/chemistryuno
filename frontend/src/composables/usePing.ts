/**
 * Ping监控 Composable
 * 负责监控WebSocket连接延迟和网络状态
 */

import { ref, computed, onMounted, onUnmounted } from 'vue'
import websocket from '../utils/websocket'

export interface PingStatus {
  latency: number // 延迟毫秒数
  status: 'excellent' | 'good' | 'fair' | 'poor' | 'disconnected' // 连接状态
  statusText: string // 状态文本
  statusColor: string // 状态颜色
}

export function usePing() {
  // 状态
  const latency = ref<number>(0) // 当前延迟
  const isConnected = ref<boolean>(false) // 连接状态
  const pingHistory = ref<number[]>([]) // 延迟历史记录（最近10次）
  const lastPingTime = ref<number>(0) // 上次ping时间戳

  // Ping间隔（毫秒）
  const PING_INTERVAL = 3000
  const MAX_HISTORY = 10
  const HIGH_PING_THRESHOLD_MS = 1000

  // 定时器
  let pingTimer: number | null = null

  // 计算平均延迟
  const averageLatency = computed(() => {
    if (pingHistory.value.length === 0) return 0
    const sum = pingHistory.value.reduce((a, b) => a + b, 0)
    return Math.round(sum / pingHistory.value.length)
  })

  // 计算状态
  const pingStatus = computed<PingStatus>(() => {
    if (!isConnected.value) {
      return {
        latency: 0,
        status: 'disconnected',
        statusText: '未连接',
        statusColor: 'text-slate-400'
      }
    }

    const avgLatency = averageLatency.value

    if (avgLatency === 0) {
      return {
        latency: 0,
        status: 'disconnected',
        statusText: '检测中',
        statusColor: 'text-slate-400'
      }
    }

    if (avgLatency < 50) {
      return {
        latency: avgLatency,
        status: 'excellent',
        statusText: '优秀',
        statusColor: 'text-emerald-500'
      }
    } else if (avgLatency < 100) {
      return {
        latency: avgLatency,
        status: 'good',
        statusText: '良好',
        statusColor: 'text-green-500'
      }
    } else if (avgLatency < 200) {
      return {
        latency: avgLatency,
        status: 'fair',
        statusText: '一般',
        statusColor: 'text-yellow-500'
      }
    } else {
      return {
        latency: avgLatency,
        status: 'poor',
        statusText: '较差',
        statusColor: 'text-red-500'
      }
    }
  })

  // 发送ping请求
  const sendPing = () => {
    if (!websocket.isConnected()) {
      isConnected.value = false
      return
    }

    isConnected.value = true
    lastPingTime.value = Date.now()

    // 发送ping消息
    websocket.send({
      type: 'ping',
      timestamp: lastPingTime.value
    })
  }

  // 处理pong响应
  const handlePong = (message: any) => {
    if (message.type !== 'pong') return

    const now = Date.now()
    const rtt = now - (message.timestamp || lastPingTime.value)

    // 更新延迟
    latency.value = rtt

    // 更新历史记录
    pingHistory.value.push(rtt)
    if (pingHistory.value.length > MAX_HISTORY) {
      pingHistory.value.shift()
    }

    if (rtt >= HIGH_PING_THRESHOLD_MS) {
      console.warn(`[Ping] High latency: ${rtt}ms, average: ${averageLatency.value}ms`)
    }
  }

  // 开始ping监控
  const startPing = () => {
    // 立即发送一次
    sendPing()

    // 设置定时器
    pingTimer = window.setInterval(() => {
      sendPing()
    }, PING_INTERVAL)

    // 监听pong响应
    websocket.on('pong', handlePong)
  }

  // 停止ping监控
  const stopPing = () => {
    if (pingTimer) {
      clearInterval(pingTimer)
      pingTimer = null
    }
    websocket.off('pong', handlePong)
  }

  // 重置ping数据
  const resetPing = () => {
    latency.value = 0
    pingHistory.value = []
    lastPingTime.value = 0
    isConnected.value = false
  }

  // 生命周期
  onMounted(() => {
    // 等待WebSocket连接建立后开始ping
    setTimeout(() => {
      if (websocket.isConnected()) {
        startPing()
      } else {
        // 如果未连接，继续等待
        const checkInterval = setInterval(() => {
          if (websocket.isConnected()) {
            clearInterval(checkInterval)
            startPing()
          }
        }, 1000)

        // 10秒后放弃等待
        setTimeout(() => {
          clearInterval(checkInterval)
        }, 10000)
      }
    }, 500)
  })

  onUnmounted(() => {
    stopPing()
  })

  return {
    // 状态
    latency,
    isConnected,
    pingHistory,
    averageLatency,
    pingStatus,

    // 方法
    startPing,
    stopPing,
    resetPing,
  }
}
