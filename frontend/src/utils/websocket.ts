import { WS_URL } from './runtimeConfig'

interface WebSocketMessage {
  type: string
  [key: string]: any
}

// 连接状态，供 UI 展示重连反馈
export type ConnectionState = 'connecting' | 'connected' | 'reconnecting' | 'disconnected' | 'failed'

type ConnectionStateListener = (state: ConnectionState) => void

class WebSocketService {
  private ws: WebSocket | null = null
  private listeners: { [key: string]: Array<(message: WebSocketMessage) => void> } = {}
  private reconnectAttempts: number = 0
  private readonly maxReconnectAttempts: number = 5
  private pendingMessages: WebSocketMessage[] = []
  private isConnecting: boolean = false
  private networkEventsBound: boolean = false
  private reconnectTimer: number | null = null
  private connectionState: ConnectionState = 'disconnected'
  private stateListeners: ConnectionStateListener[] = []

  // 订阅连接状态变化（返回取消订阅函数）
  onConnectionStateChange(listener: ConnectionStateListener): () => void {
    this.stateListeners.push(listener)
    // 立即回调当前状态
    listener(this.connectionState)
    return () => {
      this.stateListeners = this.stateListeners.filter(l => l !== listener)
    }
  }

  getConnectionState(): ConnectionState {
    return this.connectionState
  }

  private setConnectionState(state: ConnectionState): void {
    if (this.connectionState === state) return
    this.connectionState = state
    this.stateListeners.forEach(l => {
      try {
        l(state)
      } catch (e) {
        console.error('[WebSocket] connection-state listener error:', e)
      }
    })
  }

  private bindNetworkEvents(): void {
    if (this.networkEventsBound) return
    this.networkEventsBound = true

    window.addEventListener('offline', () => {
      console.log('[WebSocket] Browser offline, connection will retry when network returns')
    })

    window.addEventListener('online', () => {
      console.log('[WebSocket] Browser back online')
      // 网络恢复后重置退避并立即重连（即便此前已进入 failed 状态）
      if (!this.isConnected()) {
        this.manualReconnect()
      }
    })
  }

  connect(): void {
    if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
      return
    }

    this.bindNetworkEvents()
    this.isConnecting = true
    // 首次连接显示 connecting，重连过程中保持 reconnecting
    this.setConnectionState(this.reconnectAttempts > 0 ? 'reconnecting' : 'connecting')
    const wsUrl = WS_URL

    this.ws = new WebSocket(wsUrl)

    this.ws.onopen = () => {
      console.log('[WebSocket] connected')
      this.reconnectAttempts = 0
      this.isConnecting = false
      this.setConnectionState('connected')

      while (this.pendingMessages.length > 0) {
        const msg = this.pendingMessages.shift()
        if (msg) this.send(msg)
      }
    }

    this.ws.onmessage = (event: MessageEvent) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data)
        this.handleMessage(message)
      } catch (error) {
        console.error('[WebSocket] failed to parse message:', error)
      }
    }

    this.ws.onclose = () => {
      console.log('[WebSocket] disconnected')
      this.isConnecting = false
      this.attemptReconnect()
    }

    this.ws.onerror = (error: Event) => {
      console.error('[WebSocket] error:', error)
      this.isConnecting = false
    }
  }

  private attemptReconnect(): void {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      // 指数退避：1s, 2s, 4s, 8s, 16s（上限 16s），带少量抖动
      const base = Math.min(1000 * 2 ** (this.reconnectAttempts - 1), 16000)
      const delay = base + Math.floor(Math.random() * 500)
      console.log(`[WebSocket] reconnect attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts} in ${delay}ms`)
      this.setConnectionState('reconnecting')
      if (this.reconnectTimer !== null) {
        clearTimeout(this.reconnectTimer)
      }
      this.reconnectTimer = window.setTimeout(() => {
        this.reconnectTimer = null
        this.connect()
      }, delay)
    } else {
      // 重连次数耗尽，进入 failed 状态，等待用户手动重试或网络恢复
      console.warn('[WebSocket] max reconnect attempts reached, giving up')
      this.setConnectionState('failed')
    }
  }

  // 手动重连（供 UI 的“重试”按钮调用），重置退避计数
  manualReconnect(): void {
    if (this.reconnectTimer !== null) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    this.reconnectAttempts = 0
    this.connect()
  }

  disconnect(): void {
    this.reconnectAttempts = this.maxReconnectAttempts
    this.pendingMessages = []
    if (this.reconnectTimer !== null) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.isConnecting = false
    this.setConnectionState('disconnected')
  }

  send(message: WebSocketMessage): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    } else {
      console.log('[WebSocket] queueing message until the connection is ready')
      this.pendingMessages.push(message)
      if (!this.isConnecting) {
        this.connect()
      }
    }
  }

  on(event: string, callback: (message: WebSocketMessage) => void): void {
    if (!this.listeners[event]) {
      this.listeners[event] = []
    }
    this.listeners[event].push(callback)
  }

  off(event: string, callback: (message: WebSocketMessage) => void): void {
    if (this.listeners[event]) {
      this.listeners[event] = this.listeners[event].filter(cb => cb !== callback)
    }
  }

  private handleMessage(message: WebSocketMessage): void {
    const { type } = message
    if (this.listeners[type]) {
      this.listeners[type].forEach(callback => {
        try {
          callback(message)
        } catch (e) {
          console.error(`WebSocket handler error for "${type}":`, e)
        }
      })
    }
  }

  joinRoom(roomId: string): void {
    console.log('[WebSocket] Joining room:', roomId)
    this.send({ type: 'join_room', room_id: roomId })
  }

  leaveRoom(): void {
    console.log('[WebSocket] Leaving room')
    this.send({ type: 'leave_room' })
  }

  sendChat(message: string): void {
    this.send({ type: 'chat', message })
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN
  }
}

export default new WebSocketService()
