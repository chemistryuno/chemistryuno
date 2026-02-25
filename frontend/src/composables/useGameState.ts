/**
 * 游戏状态管理 Composable
 * 负责管理游戏状态、房间信息和玩家数据
 */

import { ref, computed, type Ref } from 'vue'
import { gameAPI } from '../utils/api'

export function useGameState(roomId: string) {
  // 状态
  const gameState = ref<any>(null)
  const roomInfo = ref<any>(null)
  const playersInfo = ref<any[]>([])
  const loading = ref(true)
  const loadError = ref<string | null>(null)

  // 计算属性
  const allPlayers = computed(() => {
    if (gameState.value?.players) {
      return gameState.value.players.map((p: any) => {
        const baseInfo = playersInfo.value.find(b => Number(b.uid) === Number(p.uid))
        return {
          ...p,
          avatar: p.avatar || baseInfo?.avatar,
          nickname: p.nickname || baseInfo?.nickname || baseInfo?.username,
          username: p.username || baseInfo?.username,
        }
      })
    }
    return []
  })

  const isMyTurn = computed(() => {
    const user = JSON.parse(localStorage.getItem('user') || '{}')
    return gameState.value?.current_player === user.uid
  })

  const currentPlayerObj = computed(() => {
    return allPlayers.value.find(
      (p: any) => p.uid === gameState.value?.current_player
    )
  })

  // 方法
  const loadGameState = async (silent = false) => {
    if (!silent) {
      loading.value = true
      loadError.value = null
    }

    try {
      const response = await gameAPI.getGameState(roomId)
      gameState.value = response.data.game_state
      roomInfo.value = response.data.room_info

      // 获取玩家基础信息
      if (gameState.value?.players) {
        const uids = gameState.value.players.map((p: any) => p.uid)
        const playersResponse = await gameAPI.getPlayersInfo(uids)
        playersInfo.value = playersResponse.data.players || []
      }

      loading.value = false
      return true
    } catch (error: any) {
      console.error('加载游戏状态失败:', error)
      loadError.value = error.response?.data?.error || '加载失败'
      loading.value = false
      return false
    }
  }

  const updateGameState = (newState: any) => {
    gameState.value = newState
  }

  const updateRoomInfo = (newInfo: any) => {
    roomInfo.value = newInfo
  }

  return {
    // 状态
    gameState,
    roomInfo,
    playersInfo,
    loading,
    loadError,

    // 计算属性
    allPlayers,
    isMyTurn,
    currentPlayerObj,

    // 方法
    loadGameState,
    updateGameState,
    updateRoomInfo,
  }
}
