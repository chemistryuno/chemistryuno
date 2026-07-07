/**
 * 连接状态 Composable
 * 将 WebSocket 服务的连接状态暴露为响应式数据，供 UI 展示重连反馈
 */

import { ref, onMounted, onUnmounted, computed } from 'vue'
import websocket, { type ConnectionState } from '../utils/websocket'

export function useConnectionState() {
  const connectionState = ref<ConnectionState>(websocket.getConnectionState())
  let unsubscribe: (() => void) | null = null

  // 是否应展示“连接中断”遮罩：重连中或彻底失败时
  const showConnectionOverlay = computed(
    () => connectionState.value === 'reconnecting' || connectionState.value === 'failed'
  )

  const isReconnecting = computed(() => connectionState.value === 'reconnecting')
  const isFailed = computed(() => connectionState.value === 'failed')
  const isConnected = computed(() => connectionState.value === 'connected')

  const statusText = computed(() => {
    switch (connectionState.value) {
      case 'connecting':
        return '正在建立连接…'
      case 'connected':
        return '已连接'
      case 'reconnecting':
        return '连接中断，正在重连…'
      case 'failed':
        return '连接失败'
      default:
        return '未连接'
    }
  })

  // 手动重试
  const retry = () => websocket.manualReconnect()

  onMounted(() => {
    unsubscribe = websocket.onConnectionStateChange((state) => {
      connectionState.value = state
    })
  })

  onUnmounted(() => {
    if (unsubscribe) unsubscribe()
  })

  return {
    connectionState,
    showConnectionOverlay,
    isReconnecting,
    isFailed,
    isConnected,
    statusText,
    retry,
  }
}
