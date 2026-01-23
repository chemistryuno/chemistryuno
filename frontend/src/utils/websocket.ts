interface WebSocketMessage {
  type: string
  [key: string]: any
}

class WebSocketService {
  private ws: WebSocket | null = null
  private listeners: { [key: string]: Array<(message: WebSocketMessage) => void> } = {}
  private reconnectAttempts: number = 0
  private readonly maxReconnectAttempts: number = 5

  connect(): void {
    const token = localStorage.getItem('token')
    if (!token) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.hostname}:8080/ws`
    
    this.ws = new WebSocket(wsUrl)

    this.ws.onopen = () => {
      console.log('WebSocket连接已建立')
      this.reconnectAttempts = 0
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
      this.attemptReconnect()
    }

    this.ws.onerror = (error: Event) => {
      console.error('WebSocket错误:', error)
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
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  send(message: WebSocketMessage): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    } else {
      console.error('WebSocket未连接')
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
      this.listeners[type].forEach(callback => callback(message))
    }
  }

  joinRoom(roomId: string): void {
    this.send({ type: 'join_room', room_id: roomId })
  }

  leaveRoom(): void {
    this.send({ type: 'leave_room' })
  }

  sendChat(message: string): void {
    this.send({ type: 'chat', message })
  }
}

export default new WebSocketService()
