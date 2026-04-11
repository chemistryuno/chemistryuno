import { WS_URL } from './runtimeConfig'

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
  private networkEventsBound: boolean = false

  private bindNetworkEvents(): void {
    if (this.networkEventsBound) return
    this.networkEventsBound = true

    window.addEventListener('offline', () => {
      console.log('[WebSocket] Browser offline, connection will retry when network returns')
    })

    window.addEventListener('online', () => {
      console.log('[WebSocket] Browser back online')
      if (!this.isConnected() && this.reconnectAttempts < this.maxReconnectAttempts) {
        this.connect()
      }
    })
  }

  connect(): void {
    if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
      return
    }

    this.bindNetworkEvents()
    this.isConnecting = true
    const wsUrl = WS_URL

    this.ws = new WebSocket(wsUrl)

    this.ws.onopen = () => {
      console.log('[WebSocket] connected')
      this.reconnectAttempts = 0
      this.isConnecting = false

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
      console.log(`[WebSocket] reconnect attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts}`)
      setTimeout(() => this.connect(), 3000)
    }
  }

  disconnect(): void {
    this.reconnectAttempts = this.maxReconnectAttempts
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
