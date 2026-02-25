/**
 * WebSocket 通信 Composable
 * 负责处理游戏房间的 WebSocket 连接和消息
 */

import { onMounted, onUnmounted, type Ref } from 'vue'
import websocket from '../utils/websocket'
import feedback from '../utils/feedback'

export function useGameSocket(
  roomId: string,
  callbacks: {
    onGameUpdate?: (data: any) => void
    onPlayerJoined?: (data: any) => void
    onPlayerLeft?: (data: any) => void
    onActionToast?: (data: any) => void
    onRoomTerminated?: (data: any) => void
    onPlayerKicked?: (data: any) => void
    onChatMessage?: (data: any) => void
  }
) {
  // 游戏更新
  const handleGameUpdate = (data: any) => {
    console.log('收到游戏状态更新:', data)
    callbacks.onGameUpdate?.(data)
  }

  // 玩家加入
  const handlePlayerJoined = (data: any) => {
    console.log('玩家加入:', data)
    callbacks.onPlayerJoined?.(data)
  }

  // 玩家离开
  const handlePlayerLeft = (data: any) => {
    console.log('玩家离开:', data)
    callbacks.onPlayerLeft?.(data)
  }

  // 操作提示
  const handleActionToast = (data: any) => {
    console.log('操作提示:', data)
    callbacks.onActionToast?.(data)
  }

  // 房间关闭
  const handleRoomTerminated = (data: any) => {
    console.log('房间已关闭:', data)
    callbacks.onRoomTerminated?.(data)
  }

  // 玩家被踢
  const handlePlayerKicked = (data: any) => {
    console.log('玩家被踢:', data)
    callbacks.onPlayerKicked?.(data)
  }

  // 聊天消息
  const handleChatMessage = (data: any) => {
    console.log('聊天消息:', data)
    callbacks.onChatMessage?.(data)
  }

  // 发送消息
  const sendMessage = (type: string, data: any) => {
    websocket.send({
      type,
      ...data
    })
  }

  // 连接和断开
  const connect = () => {
    console.log('[WebSocket] 加入房间:', roomId)
    websocket.joinRoom(roomId)

    // 注册事件监听
    websocket.on('game_update', handleGameUpdate)
    websocket.on('player_joined', handlePlayerJoined)
    websocket.on('player_left', handlePlayerLeft)
    websocket.on('action_toast', handleActionToast)
    websocket.on('room_terminated', handleRoomTerminated)
    websocket.on('player_kicked', handlePlayerKicked)
    websocket.on('chat', handleChatMessage)
    websocket.on('private_chat', handleChatMessage)
  }

  const disconnect = () => {
    console.log('[WebSocket] 离开房间:', roomId)
    websocket.leaveRoom()

    // 移除事件监听
    websocket.off('game_update', handleGameUpdate)
    websocket.off('player_joined', handlePlayerJoined)
    websocket.off('player_left', handlePlayerLeft)
    websocket.off('action_toast', handleActionToast)
    websocket.off('room_terminated', handleRoomTerminated)
    websocket.off('player_kicked', handlePlayerKicked)
    websocket.off('chat', handleChatMessage)
    websocket.off('private_chat', handleChatMessage)
  }

  // 生命周期钩子
  onMounted(() => {
    connect()
  })

  onUnmounted(() => {
    disconnect()
  })

  return {
    sendMessage,
    connect,
    disconnect,
  }
}
