interface WebSocketMessage {
  type: string
  [key: string]: any
}

class WebSocketService {
  private ws: WebSocket | null = null
  private listeners: { [key: string]: Array<(message: WebSocketMessage) => void> } = {}
  private reconnectAttempts: number = 0
  private readonly maxReconnectAttempts: number = 5
  private pendingMessages: WebSocketMessage[] = []
  private isConnecting: boolean = false

  connect(): void {
    const token = localStorage.getItem('token')
    if (!token) return

    // 避免重复连接
    if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
      return
    }

    this.isConnecting = true
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/ws?token=${token}`
    
    this.ws = new WebSocket(wsUrl)

    this.ws.onopen = () => {
      console.log('WebSocket连接已建立')
      this.reconnectAttempts = 0
      this.isConnecting = false
      
      // 发送待发送的消息
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
        console.error('消息解析失败:', error)
      }
    }

    this.ws.onclose = () => {
      console.log('WebSocket连接已关闭')
      this.isConnecting = false
      this.attemptReconnect()
    }

    this.ws.onerror = (error: Event) => {
      console.error('WebSocket错误:', error)
      this.isConnecting = false
    }
  }

  private attemptReconnect(): void {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      console.log(`尝试重连 (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`)
      setTimeout(() => this.connect(), 3000)
    }
  }

  disconnect(): void {
    this.reconnectAttempts = this.maxReconnectAttempts // 阻止自动重连
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.isConnecting = false
  }

  send(message: WebSocketMessage): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    } else {
      // 如果连接未建立，将消息加入队列
      console.log('WebSocket未连接，消息已加入队列')
      this.pendingMessages.push(message)
      // 尝试建立连接
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
