class WebSocketService {
  constructor() {
    this.ws = null
    this.listeners = {}
    this.reconnectAttempts = 0
    this.maxReconnectAttempts = 5
  }

  connect() {
    const token = localStorage.getItem('token')
    if (!token) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.hostname}:8080/ws`
    
    this.ws = new WebSocket(wsUrl)

    this.ws.onopen = () => {
      console.log('WebSocket连接已建立')
      this.reconnectAttempts = 0
    }

    this.ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data)
        this.handleMessage(message)
      } catch (error) {
        console.error('消息解析失败:', error)
      }
    }

    this.ws.onclose = () => {
      console.log('WebSocket连接已关闭')
      this.attemptReconnect()
    }

    this.ws.onerror = (error) => {
      console.error('WebSocket错误:', error)
    }
  }

  attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      console.log(`尝试重连 (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`)
      setTimeout(() => this.connect(), 3000)
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  send(message) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    } else {
      console.error('WebSocket未连接')
    }
  }

  on(event, callback) {
    if (!this.listeners[event]) {
      this.listeners[event] = []
    }
    this.listeners[event].push(callback)
  }

  off(event, callback) {
    if (this.listeners[event]) {
      this.listeners[event] = this.listeners[event].filter(cb => cb !== callback)
    }
  }

  handleMessage(message) {
    const { type } = message
    if (this.listeners[type]) {
      this.listeners[type].forEach(callback => callback(message))
    }
  }

  joinRoom(roomId) {
    this.send({ type: 'join_room', room_id: roomId })
  }

  leaveRoom() {
    this.send({ type: 'leave_room' })
  }

  sendChat(message) {
    this.send({ type: 'chat', message })
  }
}

export default new WebSocketService()
