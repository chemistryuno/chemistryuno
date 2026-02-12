<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { gameAPI, adminAPI, friendAPI, authAPI, commonAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import websocket from '../utils/websocket'
import { ArrowLeft, Play, RefreshCw, Zap, Activity, FlaskConical, Trophy, ChevronRight, Loader2, Users, Timer, Plus, QrCode, Copy, Sparkles, ShieldAlert, Ban, UserMinus, X, MessageCircle, UserPlus, Flag, Send } from 'lucide-vue-next'
import { cn } from '../utils/cn'
import ChatBox from '../components/ChatBox.vue'
import '../styles/mobile-game.css'

const route = useRoute()
const router = useRouter()
const { showAlert, showConfirm, showPrompt } = useDialog()
const id = route.params.id as string

const user = ref<any>({})
try {
  const userData = JSON.parse(localStorage.getItem('user') || '{}')
  // 兼容旧版本的 id 字段
  if (userData.id && !userData.uid) {
    userData.uid = userData.id
  }
  user.value = userData
} catch (e) {
  console.error('Failed to parse user in GameRoom:', e)
}

const gameState = ref<any>(null)
const roomInfo = ref<any>(null)
const playersInfo = ref<any[]>([])
const friendsList = ref<any[]>([])
const availableSubstances = ref<string[]>([])

const loading = ref(true)
const loadError = ref<string | null>(null)
const isRedirecting = ref(false)
const timeRemaining = ref(0)
let timerInterval: any = null
const selectedCard = ref<any>(null)
const selectedSubstance = ref<string | null>(null)
const turnReadySubstances = ref<string[]>([])
const doubleMode = ref(false)
const firstDoubleSubstance = ref<string | null>(null)
const secondDoubleSubstance = ref<string | null>(null)
const substanceInput = ref('')
const randomHints = ref<any[]>([])
const reactionHints = ref<any[]>([])

// UI State
const isMobile = ref(false)
const showLogs = ref(false)
const showHints = ref(true)
const showDeckDetailModal = ref(false)
const handContainer = ref<HTMLElement | null>(null)
const substancesContainer = ref<HTMLElement | null>(null)
const playersContainer = ref<HTMLElement | null>(null)

// 移动端自动全屏
const requestFullscreen = () => {
  if (!isMobile.value) return
  const el = document.documentElement as any
  const rfs = el.requestFullscreen || el.webkitRequestFullscreen || el.mozRequestFullScreen || el.msRequestFullscreen
  if (rfs) {
    rfs.call(el).catch(() => {})
  }
}

// 移动端退出全屏
const exitFullscreen = () => {
  if (!isMobile.value) return
  const doc = document as any
  if (doc.fullscreenElement || doc.webkitFullscreenElement || doc.mozFullScreenElement || doc.msFullscreenElement) {
    const efs = doc.exitFullscreen || doc.webkitExitFullscreen || doc.mozCancelFullScreen || doc.msExitFullscreen
    if (efs) {
      efs.call(doc).catch(() => {})
    }
  }
}

// 输入框焦点管理：聚焦时退出全屏，失焦时恢复全屏
const handleInputFocus = () => exitFullscreen()
const handleInputBlur = () => requestFullscreen()

// 自动滚动到当前行动玩家
const scrollToActivePlayer = () => {
  if (!playersContainer.value || gameState.value?.current_player == null) return
  const container = playersContainer.value
  const playerCards = container.querySelectorAll('[data-player-card]')
  const activeIndex = gameState.value.current_player
  const activeCard = playerCards[activeIndex] as HTMLElement
  if (!activeCard) return
  const containerRect = container.getBoundingClientRect()
  const cardRect = activeCard.getBoundingClientRect()
  const scrollLeft = activeCard.offsetLeft - containerRect.width / 2 + cardRect.width / 2
  container.scrollTo({ left: scrollLeft, behavior: 'smooth' })
}

const fetchRandomHints = async () => {
  try {
    const res = await commonAPI.getHints()
    randomHints.value = res.data || []
  } catch (error) {
    console.error('Failed to fetch hints from labs:', error)
  }
}

const fetchReactionHints = async () => {
  try {
    const res = await gameAPI.getReactionHints(id)
    reactionHints.value = res.data || []
  } catch (error) {
    console.error('Failed to fetch reaction hints:', error)
    reactionHints.value = []
  }
}

const viewCurrentDeckConfig = () => {
  if (roomInfo.value?.deck_config) {
    showDeckDetailModal.value = true
  }
}

const isReady = computed(() => {
  return roomInfo.value?.ready_uids?.includes(Number(user.value.uid))
})

const handleToggleReady = async () => {
  console.log('handleToggleReady called')
  console.log('roomInfo:', roomInfo.value)
  console.log('user.uid:', user.value?.uid)

  if (!roomInfo.value || !user.value.uid) {
    console.error('Cannot toggle ready - missing roomInfo or user.uid')
    if (!roomInfo.value) {
      showAlert('房间信息未加载，请刷新页面', '错误')
    } else if (!user.value.uid) {
      showAlert('用户信息异常，请重新登录', '错误')
    }
    return
  }

  // 乐观更新：立即切换状态
  const uidNum = Number(user.value.uid)
  const isCurrentlyReady = roomInfo.value.ready_uids.includes(uidNum)

  console.log('uidNum:', uidNum)
  console.log('isCurrentlyReady:', isCurrentlyReady)
  console.log('ready_uids before:', roomInfo.value.ready_uids)

  if (isCurrentlyReady) {
    roomInfo.value.ready_uids = roomInfo.value.ready_uids.filter((id: number) => id !== uidNum)
  } else {
    roomInfo.value.ready_uids = [...roomInfo.value.ready_uids, uidNum]
  }

  console.log('ready_uids after:', roomInfo.value.ready_uids)

  try {
    const response = await gameAPI.ready(id)
    console.log('Ready API response:', response)
    // 状态也会通过 WebSocket 更新，但手动标记一下提高体验
    await loadGameState(true)
  } catch (error: any) {
    console.error('Ready API error:', error)
    // 恢复状态
    if (isCurrentlyReady) {
      if (!roomInfo.value.ready_uids.includes(uidNum)) {
        roomInfo.value.ready_uids.push(uidNum)
      }
    } else {
      roomInfo.value.ready_uids = roomInfo.value.ready_uids.filter((id: number) => id !== uidNum)
    }
    showAlert(error.response?.data?.error || '操作失败', '错误')
  }
}

const isFriend = (uid: number) => {
  return friendsList.value?.some(f => Number(f.uid) === Number(uid)) ?? false
}

const handleAddFriend = async (player: any) => {
  try {
    const displayName = player.nickname || player.username
    await friendAPI.sendRequest(player.uid)
    showAlert(`已向研究员 ${displayName} 发送同步请求，等待量子握手。`, '请求已发送')
  } catch (error: any) {
    showAlert(error.response?.data?.error || '请求发送失败', '链路故障')
  }
}

// Chat system
const showPlayers = ref(false)
const showChat = ref(false)
const hasNewMessage = ref(false)
const showQrModal = ref(false)
const showInviteFriendsModal = ref(false)

const startPrivateChat = (player: any) => {
  if (!isFriend(player.uid)) {
    showAlert('只有互为好友的研究员才能开启单向加密传输。', '权限受限')
    return
  }
  showChat.value = true
  hasNewMessage.value = false
  const displayName = player.nickname || player.username
  window.dispatchEvent(new CustomEvent('start-private-chat', {
    detail: { uid: player.uid, username: displayName }
  }))
}

const sendGameInvite = async (friend: any) => {
  const inviteData = {
    type: 'game_invite',
    room_id: id,
    room_name: roomInfo.value?.name || '实验室',
    player_count: allPlayers.value.length,
    max_players: roomInfo.value?.max_players || 0,
    is_points_mode: roomInfo.value?.is_points_mode || false,
    is_private: roomInfo.value?.is_private || false,
    access_key: roomInfo.value?.access_key || ''
  }

  websocket.send({
    type: 'private_chat',
    target_uid: friend.uid,
    message: JSON.stringify(inviteData),
    is_game_invite: true
  })

  showAlert(`游戏邀请已发送给 ${friend.username}`, '邀请已发送')
  showInviteFriendsModal.value = false
}

// Admin management state
const showAdminModal = ref(false)
const adminTargetUser = ref<any>(null)
const adminActionType = ref<'kick' | 'ban'>('kick')
const banUntil = ref('')
const banReason = ref('你由于违规游戏而被踢出')
const selectedBanPreset = ref<number | null>(24)

const formatDatetimeLocal = (d: Date) => {
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0') + 'T' + String(d.getHours()).padStart(2, '0') + ':' + String(d.getMinutes()).padStart(2, '0')
}

const getDefaultBanUntil = () => {
  return formatDatetimeLocal(new Date(Date.now() + 24 * 60 * 60 * 1000))
}

const banPresets = [
  { label: '1小时', hours: 1 },
  { label: '6小时', hours: 6 },
  { label: '24小时', hours: 24 },
  { label: '3天', hours: 72 },
  { label: '7天', hours: 168 },
  { label: '30天', hours: 720 },
  { label: '永久', hours: 87600 },
]

const setBanDuration = (hours: number) => {
  selectedBanPreset.value = hours
  banUntil.value = formatDatetimeLocal(new Date(Date.now() + hours * 3600 * 1000))
}

watch(showChat, (val) => {
  if (val) {
    hasNewMessage.value = false
    if (isMobile.value) {
      showHints.value = false
      showPlayers.value = false
    }
  }
})

watch(showPlayers, (val) => {
  if (val && isMobile.value) {
    showHints.value = false
    showChat.value = false
  }
})

watch(showHints, (val) => {
  if (val) {
    if (isMobile.value) {
      showChat.value = false
      showPlayers.value = false
    }
    if (randomHints.value.length === 0) {
      fetchRandomHints()
    }
    // 如果是玩家回合且提示为空，尝试获取提示（延迟检查，避免 computed 未初始化）
    nextTick(() => {
      if (isMyTurn.value && turnReadySubstances.value.length === 0) {
        fetchTurnSubstances()
      }
    })
  }
}, { immediate: true })

// 移动端自动关闭提示面板
watch(isMobile, (val) => {
  if (val) {
    showHints.value = false
  }
})

const openAdminAction = (player: any) => {
  if (!user.value.is_admin || player.uid === user.value.uid) return
  adminTargetUser.value = player
  adminActionType.value = 'kick'
  banReason.value = '你由于违规游戏而被踢出'
  selectedBanPreset.value = 24
  banUntil.value = getDefaultBanUntil()
  showAdminModal.value = true
}

const handleReportPlayer = async (player: any) => {
  const reason = await showPrompt(`举报研究员 ${player.username} (UID: ${player.uid})`, '请输入举报原因', '违规行为举报')
  if (!reason) return
  
  try {
    await authAPI.submitFeedback(`举报用户: ${player.username} (UID: ${player.uid})\n原因: ${reason}`, 'report')
    showAlert('举报已提交，系统正在量子分析中。', '已收到报告')
  } catch (err: any) {
    showAlert(err.response?.data?.error || '无法建立举报链路', '网络干扰')
  }
}

const handleAdminAction = async () => {
  if (!adminTargetUser.value) return
  try {
    if (adminActionType.value === 'kick') {
      await adminAPI.kickPlayer(adminTargetUser.value.uid, banReason.value)
      showAlert('该玩家已被强制下线并清除登录状态', '成功')
    } else {
      if (!banUntil.value) {
        showAlert('请选择封禁截止时间', '参数缺失')
        return
      }
      const until = new Date(banUntil.value)
      if (until <= new Date()) {
        showAlert('封禁截止时间必须晚于当前时间', '时间无效')
        return
      }
      await adminAPI.banUser(adminTargetUser.value.uid, until.toISOString(), banReason.value)
      showAlert('该玩家已被封禁', '成功')
    }
    showAdminModal.value = false
  } catch (e: any) {
    showAlert(e.response?.data?.error || '操作失败', '错误')
  }
}

const allPlayers = computed(() => {
  if (gameState.value?.players) {
    return gameState.value.players.map((p: any) => {
      const baseInfo = playersInfo.value.find(b => Number(b.uid) === Number(p.uid))
      return {
        ...p,
        avatar: p.avatar || baseInfo?.avatar,
        username: p.username || baseInfo?.username,
        is_ready: roomInfo.value?.ready_uids?.includes(Number(p.uid)),
        is_offline: baseInfo?.is_offline
      }
    })
  }
  return playersInfo.value.map(p => ({
    ...p,
    is_ready: roomInfo.value?.ready_uids?.includes(Number(p.uid)),
    is_offline: p.is_offline
  }))
})

const currentPlayerObj = computed(() => {
  if (!gameState.value) return null
  return gameState.value.players?.[gameState.value.current_player]
})
const isMyTurn = computed(() => {
  if (!currentPlayerObj.value || !user.value) return false
  return Number(currentPlayerObj.value.uid) === Number(user.value.uid)
})
const myData = computed(() => {
  if (!gameState.value || !user.value) return null
  return (gameState.value.players || []).find((p: any) => Number(p.uid) === Number(user.value.uid))
})
const myIndex = computed(() => {
  if (!gameState.value || !user.value) return -1
  return (gameState.value.players || []).findIndex((p: any) => Number(p.uid) === Number(user.value.uid))
})
const allowedAny = computed(() => {
  if (!gameState.value) return false
  return typeof gameState.value?.allowed_any_player !== 'undefined' && gameState.value?.allowed_any_player === myIndex.value
})
const winner = computed(() => gameState.value?.players?.find((p: any) => p.card_count === 0))

const ELEMENTS_DATA: Record<string, { name: string, class: string }> = {
  'H': { name: '氢', class: 'element-H' },
  'O': { name: '氧', class: 'element-O' },
  'C': { name: '碳', class: 'element-C' },
  'N': { name: '氮', class: 'element-N' },
  'S': { name: '硫', class: 'element-S' },
  'F': { name: '氟', class: 'element-F' },
  'P': { name: '磷', class: 'element-P' },
  'Cl': { name: '氯', class: 'element-Cl' },
  'Br': { name: '溴', class: 'element-Br' },
  'I': { name: '碘', class: 'element-I' },
  'Na': { name: '钠', class: 'element-Na' },
  'K': { name: '钾', class: 'element-K' },
  'Mg': { name: '镁', class: 'element-Mg' },
  'Ca': { name: '钙', class: 'element-Ca' },
  'Ba': { name: '钡', class: 'element-Ba' },
  'Al': { name: '铝', class: 'element-Al' },
  'Fe': { name: '铁', class: 'element-Fe' },
  'Zn': { name: '锌', class: 'element-Zn' },
  'Ag': { name: '银', class: 'element-Ag' },
  'Hg': { name: '汞', class: 'element-Hg' },
  'Cu': { name: '铜', class: 'element-Cu' },
}

const SUBSTANCE_NAMES: Record<string, string> = {
  'H2O': '水', 'H2': '氢气', 'O2': '氧气', 'HCl': '盐酸', 'H2SO4': '硫酸',
  'NaOH': '氢氧化钠', 'NaCl': '氯化钠', 'CO2': '二氧化碳', 'CaO': '氧化钙',
  'CuO': '氧化铜', 'Fe2O3': '氧化铁', 'Fe': '铁', 'Cu': '铜', 'Zn': '锌',
  'Mg': '镁', 'Al': '铝', 'C': '碳', 'S': '硫', 'Cl2': '氯气', 'AgNO3': '硝酸银'
}

const formatFormula = (formula: string) => {
  if (!formula) return ''
  return formula.replace(/(\d+)/g, '<sub>$1</sub>')
}

const getSubstanceName = (formula: string) => {
  if (SUBSTANCE_NAMES[formula]) return SUBSTANCE_NAMES[formula]
  return formula
}

// 解析化学式，返回元素及其数量（与后端 parseSubstance 逻辑一致）
const parseSubstanceElements = (substance: string): Record<string, number> => {
  const result: Record<string, number> = {}
  const stack: Record<string, number>[] = [result]
  let i = 0
  while (i < substance.length) {
    const c = substance[i]
    if (c === '(') {
      stack.push({})
      i++
    } else if (c === ')') {
      i++
      let count = 0
      while (i < substance.length && substance[i] >= '0' && substance[i] <= '9') {
        count = count * 10 + (substance.charCodeAt(i) - 48)
        i++
      }
      if (count === 0) count = 1
      const top = stack.pop()!
      const parent = stack[stack.length - 1]
      for (const [k, v] of Object.entries(top)) {
        parent[k] = (parent[k] || 0) + v * count
      }
    } else if (c >= 'A' && c <= 'Z') {
      const start = i
      i++
      while (i < substance.length && substance[i] >= 'a' && substance[i] <= 'z') i++
      const element = substance.slice(start, i)
      let count = 0
      while (i < substance.length && substance[i] >= '0' && substance[i] <= '9') {
        count = count * 10 + (substance.charCodeAt(i) - 48)
        i++
      }
      if (count === 0) count = 1
      const current = stack[stack.length - 1]
      current[element] = (current[element] || 0) + count
    } else {
      i++
    }
  }
  return result
}

// 检查玩家手牌是否包含合成该物质所需的所有元素
const canPlayerMakeSubstance = (substance: string): boolean => {
  if (!myData.value?.hand_cards) return false
  const required = parseSubstanceElements(substance)
  // 统计手牌中各元素数量
  const handElements: Record<string, number> = {}
  for (const card of myData.value.hand_cards) {
    handElements[card.type] = (handElements[card.type] || 0) + 1
  }
  for (const [elem, count] of Object.entries(required)) {
    if ((handElements[elem] || 0) < count) return false
  }
  return true
}

// 过滤并随机取最多3个可接续反应物提示
const filteredReactionHints = computed(() => {
  if (!reactionHints.value.length || !isMyTurn.value) return []
  const eligible = reactionHints.value.filter((hint: any) => canPlayerMakeSubstance(hint.substance))
  if (eligible.length <= 3) return eligible
  // 随机打乱后取前3个
  const shuffled = [...eligible]
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]]
  }
  return shuffled.slice(0, 3)
})

const exp = ref(Number(localStorage.getItem('chem_exp') || '0'))
const achievements = ref<string[]>(JSON.parse(localStorage.getItem('chem_achievements') || '[]'))

const checkAchievements = (substance: string) => {
  if (!substance) return
  if (substance.includes('Au') && !achievements.value.includes('炼金术士')) {
    achievements.value.push('炼金术士')
    showAlert('获得成就：炼金术士 (合成单质金)', '成就达成！')
  }
  localStorage.setItem('chem_achievements', JSON.stringify(achievements.value))
}

const addExp = (amount: number) => {
  exp.value += amount
  localStorage.setItem('chem_exp', exp.value.toString())
}

// 如果是积分赛，强制关闭提示并锁定
watch(() => roomInfo.value?.is_points_mode, (val) => {
  if (val) {
    showHints.value = false
  }
})
// --- 移植结束 ---

const startTimer = () => {
  if (timerInterval) clearInterval(timerInterval)
  timerInterval = setInterval(() => {
    if (!gameState.value || !gameState.value.turn_end_time) return
    const now = Date.now()
    const diff = Math.max(0, Math.floor((gameState.value.turn_end_time - now) / 1000))
    timeRemaining.value = diff
  }, 1000)
}

watch(() => gameState.value?.turn_end_time, () => {
  startTimer()
})

// 场上物质变化时，刷新反应提示
watch(() => gameState.value?.last_card?.substance, () => {
  if (gameState.value?.status === 'playing') {
    fetchReactionHints()
  }
}, { immediate: true })

const fetchTurnSubstances = async () => {
  if (!isMyTurn.value) {
    turnReadySubstances.value = []
    return
  }
  try {
    const response = await gameAPI.getAvailableSubstances(id)
    turnReadySubstances.value = response.data || []
  } catch (error) {
    console.error('获取回合可用物质失败:', error)
  }
}

watch(() => isMyTurn.value, (val) => {
  if (val) {
    fetchTurnSubstances()
  } else {
    turnReadySubstances.value = []
  }
}, { immediate: true })

const handleGameUpdate = (message: any) => {
  console.log('[GameRoom] handleGameUpdate called, message:', message)
  // 如果收到的是完整的游戏状态对象
  if (message.data && typeof message.data === 'object') {
    console.log('[GameRoom] Received full game state object')
    gameState.value = message.data
    if (isMyTurn.value) {
      fetchTurnSubstances()
    }
  } else {
    console.log('[GameRoom] Received game_update event, reloading state')
    // 如果收到的是房间ID字符串，则重新拉取完整状态
    loadGameState(true).then(() => {
      console.log('[GameRoom] State reloaded after game_update')
      if (isMyTurn.value) {
        fetchTurnSubstances()
      }
    })
  }
}

const handleActionToast = (msg: any) => {
  showAlert(msg.data, '实验状态变更')
}

const handleRoomTerminated = async (msg: any) => {
  isRedirecting.value = true
  const reason = msg.message || '由于连接中断，实验室已关闭'
  await showAlert(reason, '实验结束')
  router.push('/')
}

const handlePlayerKicked = async (msg: any) => {
  isRedirecting.value = true
  await showAlert(msg.message || '由于消极游戏，您已被踢出', '权限移除')
  router.push('/')
}

const handleChatNotify = () => {
  if (!showChat.value) {
    hasNewMessage.value = true
  }
}

// 为 WebSocket 事件创建包装函数，确保类型匹配
const handlePlayerJoined = () => {
  console.log('[GameRoom] Player joined event received, reloading game state')
  loadGameState(true).then(() => {
    console.log('[GameRoom] Game state reloaded after player joined, players:', playersInfo.value.length)
  })
}

const handlePlayerLeft = () => {
  console.log('[GameRoom] Player left event received, reloading game state')
  loadGameState(true).then(() => {
    console.log('[GameRoom] Game state reloaded after player left, players:', playersInfo.value.length)
  })
}

const loadGameState = async (silent = false) => {
  if (isRedirecting.value) {
    loading.value = false
    return
  }
  try {
    if (!silent && !roomInfo.value) {
      loading.value = true
    }
    console.log('[GameRoom] Loading game state for room:', id)

    // 只在首次加载时尝试加入房间（silent=false），避免重复调用
    if (!silent) {
      try {
        // 从 URL 查询参数中获取访问密钥
        const accessKey = route.query.key as string | undefined
        await gameAPI.joinRoom(id, accessKey)
        console.log('[GameRoom] Successfully joined/rejoined room')
      } catch (joinError: any) {
        // 如果加入失败（例如房间已满、被封禁等），显示错误并返回
        console.error('[GameRoom] Failed to join room:', joinError)
        const errorMsg = joinError.response?.data?.error || '无法加入该房间'
        loadError.value = errorMsg
        showAlert(errorMsg, '加入失败')
        loading.value = false
        router.push('/')
        return
      }
    }

    const response = await gameAPI.getRoomState(id)
    const data = response.data
    console.log('[GameRoom] Game state loaded:', {
      status: data.status,
      players: data.players,
      players_info: data.players_info,
      ready_uids: data.ready_uids,
      has_game_state: !!data.game_state
    })

    roomInfo.value = {
      id: data.id,
      name: data.name,
      players: data.players,
      ready_uids: data.ready_uids || [],
      countdown: data.countdown || 0,
      max_players: data.max_players,
      status: data.status,
      is_points_mode: data.is_points_mode,
      deck_config: data.deck_config,
      is_private: data.is_private,
      access_key: data.access_key
    }

    console.log('[GameRoom] roomInfo updated, status:', roomInfo.value.status)

    playersInfo.value = data.players_info || []

    if (data.game_state) {
      gameState.value = data.game_state
      console.log('[GameRoom] Game state updated:', {
        current_player: gameState.value.current_player,
        players_count: gameState.value.players?.length,
        status: gameState.value.status,
        deck_count: gameState.value.deck_count
      })
    } else {
      console.log('[GameRoom] No game_state in response, room status:', data.status)
    }
    
    loading.value = false
  } catch (error: any) {
    console.error('加载游戏状态失败:', error)
    loading.value = false

    if (error.response?.status === 404) {
      loadError.value = '房间不存在或已被关闭'
      isRedirecting.value = true
      showAlert('房间不存在或已被关闭', '未知实验室')
      router.push('/')
    } else if (error.response?.status === 401) {
      loadError.value = '身份验证失败，请重新登录'
      isRedirecting.value = true
      showAlert('身份验证失败，请重新登录', '准入失败')
      router.push('/login')
    } else if (error.response?.status === 403) {
      loadError.value = '您不在该房间中'
      isRedirecting.value = true
      showAlert('您不在该房间中', '准入失败')
      router.push('/')
    } else {
      loadError.value = '实验环境加载异常，请重试'
      // 非致命错误不自动跳转，允许用户重试
      if (!silent) {
        isRedirecting.value = true
        router.push('/')
      }
    }
  }
}

onMounted(() => {
  // 重置状态，防止之前的错误状态影响
  isRedirecting.value = false

  // 设置一个安全超时，如果15秒后还在loading状态，强制重置
  const safetyTimeout = setTimeout(() => {
    if (loading.value) {
      console.error('Loading timeout - forcing reset')
      loading.value = false
      loadError.value = '实验室初始化超时，请检查网络连接后重试'
      showAlert('实验室初始化超时，请检查网络连接后重试', '连接超时')
      router.push('/')
    }
  }, 15000)

  // 加载好友列表，添加错误处理
  friendAPI.getFriends()
    .then(res => {
      friendsList.value = res.data || []
      console.log('[GameRoom] Friends list loaded:', friendsList.value.length, 'friends')
    })
    .catch(err => {
      console.error('Failed to load friends list:', err)
      friendsList.value = [] // 确保失败时也初始化为空数组
      // 继续加载游戏状态，即使好友列表加载失败
    })

  loadGameState()
    .then(() => {
      clearTimeout(safetyTimeout) // 成功加载后清除超时

      // 游戏状态加载成功后，获取提示信息（避免 setup 阶段的 API 调用失败导致提示为空）
      if (showHints.value && randomHints.value.length === 0) {
        fetchRandomHints()
      }

      // 确保WebSocket已连接
      if (!websocket.isConnected()) {
        websocket.connect()
      }

      websocket.joinRoom(id)
      websocket.on('game_update', handleGameUpdate)
      websocket.on('player_joined', handlePlayerJoined)
      websocket.on('player_left', handlePlayerLeft)
      websocket.on('action_toast', handleActionToast)
      websocket.on('room_terminated', handleRoomTerminated)
      websocket.on('player_kicked', handlePlayerKicked)
      websocket.on('chat', handleChatNotify)
      websocket.on('private_chat', handleChatNotify)
    })
    .catch(err => {
      clearTimeout(safetyTimeout) // 捕获错误后也清除超时
      // loadGameState 内部已经处理了错误，这里只是确保不会有未处理的promise rejection
      console.error('Failed to initialize game room:', err)
      loading.value = false
    })
})

onUnmounted(() => {
  if (timerInterval) clearInterval(timerInterval)
  websocket.leaveRoom()
  websocket.off('game_update', handleGameUpdate)
  websocket.off('player_joined', handlePlayerJoined)
  websocket.off('player_left', handlePlayerLeft)
  websocket.off('action_toast', handleActionToast)
  websocket.off('room_terminated', handleRoomTerminated)
  websocket.off('player_kicked', handlePlayerKicked)
  websocket.off('chat', handleChatNotify)
  websocket.off('private_chat', handleChatNotify)
})

const handleCardClick = async (card: any) => {
  if (!isMyTurn.value) return

  // 功能牌直接打出，元素牌需检查能否反应
  const specialTypes = ['+2', '+4', 'Au', 'He', 'Ne', 'Ar', 'Kr']
  if (specialTypes.includes(card.type) || card.effect) {
    // 功能牌直接打出
    try {
      await gameAPI.playCard(id, card, card.type)
      selectedCard.value = null
      selectedSubstance.value = null
      availableSubstances.value = []
      return
    } catch (error: any) {
      showAlert(error.response?.data?.error || '出牌失败', '反应中断')
      return
    }
  }

  // 元素牌，先查可用substance
  try {
    const response = await gameAPI.getAvailableSubstances(id)
    const subs = response.data || []
    
    // 找出包含该元素的可用物质
    const matchingSubs = subs.filter((s: string) => {
      // 匹配物质中的元素
      const regex = /[A-Z][a-z]?/g
      const elementsInSub = (s.match(regex) || []) as string[]
      return elementsInSub.includes(card.type)
    })

    if (matchingSubs.length === 0) {
      showAlert('该元素当前无法参与任何反应', '反应受阻')
      return
    }

    // 优先寻找由该元素单个原子组成的单质（即物质名等于元素符号）
    const defaultSubstance = card.type
    if (matchingSubs.includes(defaultSubstance)) {
      await gameAPI.playCard(id, card, defaultSubstance)
      selectedCard.value = null
      selectedSubstance.value = null
      availableSubstances.value = []
      // 增加经验值
      addExp(10)
      checkAchievements(defaultSubstance)
      return
    }

    if (matchingSubs.length === 1) {
      // 只有一种可能，直接出
      await gameAPI.playCard(id, card, matchingSubs[0])
      selectedCard.value = null
      selectedSubstance.value = null
      availableSubstances.value = []
      // 增加经验值
      addExp(10)
      checkAchievements(matchingSubs[0])
    } else {
      // 多种可能（且不含默认单质），显示选择器
      selectedCard.value = card
      availableSubstances.value = matchingSubs
    }
  } catch (error: any) {
    showAlert(error.response?.data?.error || '出牌失败', '反应中断')
  }
}

const handlePlayCard = async () => {
  if (!selectedSubstance.value) {
    showAlert('请选择要合成或放置的化学物质', '未选择目标')
    return
  }

  if (doubleMode.value) {
    if (!firstDoubleSubstance.value) {
      firstDoubleSubstance.value = selectedSubstance.value
    } else if (!secondDoubleSubstance.value) {
      secondDoubleSubstance.value = selectedSubstance.value
    }
    selectedCard.value = null
    selectedSubstance.value = null
    availableSubstances.value = []
    return
  }

  try {
    // 如果没有选中的卡片，则传递一个带类型的占位符，后端会根据物质消耗手牌
    const cardToPlay = selectedCard.value || { type: selectedSubstance.value, count: 1, effect: '' }
    await gameAPI.playCard(id, cardToPlay, selectedSubstance.value)
    
    // 增加经验值并检查成就
    addExp(10)
    checkAchievements(selectedSubstance.value)
    
    selectedCard.value = null
    selectedSubstance.value = null
    availableSubstances.value = []
  } catch (error: any) {
    showAlert(error.response?.data?.error || '出牌失败', '反应中断')
  }
}

const handleDoublePlay = async () => {
  if (!firstDoubleSubstance.value || !secondDoubleSubstance.value) {
    showAlert('请选择参与双联反应的两种物质', '未就绪')
    return
  }

  try {
    await gameAPI.playDouble(id, firstDoubleSubstance.value, secondDoubleSubstance.value)
    
    // 增加经验值
    addExp(25)
    checkAchievements(firstDoubleSubstance.value)
    checkAchievements(secondDoubleSubstance.value)

    firstDoubleSubstance.value = null
    secondDoubleSubstance.value = null
    doubleMode.value = false
    selectedCard.value = null
    selectedSubstance.value = null
    availableSubstances.value = []
  } catch (error: any) {
    showAlert(error.response?.data?.error || '双联行动失败', '反应中断')
  }
}

const toggleDoubleMode = () => {
  if (!myData.value?.double_action_available) {
    showAlert('双联反应尚未就绪，请先进行普通实验（行动）', '无法发动')
    return
  }
  doubleMode.value = !doubleMode.value
  firstDoubleSubstance.value = null
  secondDoubleSubstance.value = null
  selectedSubstance.value = null
}

const removeSubstance = (pos: number) => {
  if (pos === 1) {
    firstDoubleSubstance.value = secondDoubleSubstance.value
    secondDoubleSubstance.value = null
  } else {
    secondDoubleSubstance.value = null
  }
}

const handleInputPlay = async () => {
  if (!substanceInput.value) return

  if (doubleMode.value) {
    const sub = substanceInput.value
    if (!firstDoubleSubstance.value) {
      firstDoubleSubstance.value = sub
    } else if (!secondDoubleSubstance.value) {
      secondDoubleSubstance.value = sub
    }
    substanceInput.value = ''
    return
  }

  try {
    // 为兼容原API，传一个空Card对象
    await gameAPI.playCard(id, { type: '', count: 0, effect: '' }, substanceInput.value)
    
    // 增加经验值并检查成就
    addExp(10)
    checkAchievements(substanceInput.value)

    substanceInput.value = ''
    selectedCard.value = null
    selectedSubstance.value = null
    availableSubstances.value = []
  } catch (error: any) {
    showAlert(error.response?.data?.error || '出牌失败', '反应中断')
  }
}

const handleDrawCard = async () => {
  try {
    await gameAPI.drawCard(id)
  } catch (error: any) {
    showAlert(error.response?.data?.error || '摸牌失败', '系统异常')
  }
}

const handleLeaveRoom = async () => {
  try {
    const confirmed = await showConfirm('暂时离开实验室？你可以在被踢出前随时返回继续实验', '暂离实验')
    if (confirmed) {
      router.push('/')
    }
  } catch (error) {
    console.error('离开房间失败:', error)
    router.push('/')
  }
}

// 生成分享链接（私密房间自动带密钥）
const shareLink = computed(() => {
  const currentUrl = window.location.href
  // 如果是私密房间且有访问密钥，自动添加key参数
  if (roomInfo.value?.is_private && roomInfo.value?.access_key) {
    // 检查URL中是否已经包含key参数
    if (!currentUrl.includes('?key=')) {
      const separator = currentUrl.includes('?') ? '&' : '?'
      return `${currentUrl}${separator}key=${roomInfo.value.access_key}`
    }
  }
  return currentUrl
})

const handleCopyLink = async () => {
  try {
    await navigator.clipboard.writeText(shareLink.value)
    if (roomInfo.value?.is_private) {
      showAlert('私密房间邀请链接已复制（含访问密钥），快发送给你的科研伙伴吧！', '任务下达')
    } else {
      showAlert('实验邀请链接已复制到剪贴板，快发送给你的科研伙伴吧！', '任务下达')
    }
  } catch (err) {
    showAlert('链接复制失败，请手动复制浏览器地址栏', '设备故障')
  }
}

const getDynamicCardClass = (card: any, formula?: string) => {
  if (!card) {
    if (formula) {
      const elements = formula.match(/[A-Z][a-z]?/g) || []
      if (elements.length > 1) return 'card-reaction'
      if (elements.length === 1 && ELEMENTS_DATA[elements[0]]) return ELEMENTS_DATA[elements[0]].class
    }
    return ''
  }

  // 特殊性质卡牌优先
  const nobleGases = ['He', 'Ne', 'Ar', 'Kr']
  if (nobleGases.includes(card.type)) return 'card-noble'
  if (card.effect || card.type === 'Au') return 'card-func'

  // 如果提供了分子式（通常是反应结果）
  if (formula) {
    const elements = formula.match(/[A-Z][a-z]?/g) || []
    // 判读是否为化合物（包含多种元素）
    if (elements.length > 1) return 'card-reaction'
    // 单质则使用该元素的颜色
    if (elements.length === 1 && ELEMENTS_DATA[elements[0]]) return ELEMENTS_DATA[elements[0]].class
  }

  // 基础元素颜色
  if (ELEMENTS_DATA[card.type]) return ELEMENTS_DATA[card.type].class
  
  return ''
}

const setupDraggable = (el: HTMLElement | null) => {
  if (!el) return
  let isDown = false
  let startX: number
  let scrollLeft: number

  el.addEventListener('mousedown', (e) => {
    isDown = true
    el.style.cursor = 'grabbing'
    startX = e.pageX - el.offsetLeft
    scrollLeft = el.scrollLeft
  })

  el.addEventListener('mouseleave', () => {
    isDown = false
    el.style.cursor = 'grab'
  })

  el.addEventListener('mouseup', () => {
    isDown = false
    el.style.cursor = 'grab'
  })

  el.addEventListener('mousemove', (e) => {
    if (!isDown) return
    e.preventDefault()
    const x = e.pageX - el.offsetLeft
    const walk = (x - startX) * 2
    el.scrollLeft = scrollLeft - walk
  })
}

onMounted(() => {
  isMobile.value = window.innerWidth < 640

  // 移动端自动关闭提示面板
  if (isMobile.value) {
    showHints.value = false
  }

  const handleResize = () => {
    isMobile.value = window.innerWidth < 640
  }
  window.addEventListener('resize', handleResize)

  // 移动端自动全屏
  if (isMobile.value) {
    // 用户首次交互后请求全屏
    const onFirstInteraction = () => {
      requestFullscreen()
      document.removeEventListener('touchstart', onFirstInteraction)
      document.removeEventListener('click', onFirstInteraction)
    }
    document.addEventListener('touchstart', onFirstInteraction, { once: true })
    document.addEventListener('click', onFirstInteraction, { once: true })
  }

  // 初始化拖拽滑动
  setTimeout(() => {
    setupDraggable(handContainer.value)
    setupDraggable(substancesContainer.value)
  }, 500)

  onUnmounted(() => window.removeEventListener('resize', handleResize))
})

// 监听当前玩家变化，自动滚动到行动玩家
watch(() => gameState.value?.current_player, () => {
  nextTick(() => scrollToActivePlayer())
})
</script>

<template>
  <div class="h-screen bg-slate-50 dark:bg-slate-900 text-slate-900 dark:text-white overflow-hidden flex flex-col font-sans selection:bg-blue-500/30 max-w-[1920px] mx-auto">
    <!-- Loading State -->
    <div v-if="loading" class="h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 flex flex-col items-center justify-center p-4 relative overflow-hidden">
      <!-- Background Elements -->
      <div class="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] bg-blue-600/20 dark:bg-blue-500/30 rounded-full blur-[120px] animate-pulse"></div>
      <div class="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-purple-600/20 dark:bg-purple-500/30 rounded-full blur-[120px]"></div>
      <div class="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/carbon-fibre.png')] opacity-20"></div>

      <div class="relative z-10 flex flex-col items-center gap-6 animate-in fade-in zoom-in duration-700">
        <div class="relative group">
          <div class="w-24 h-24 bg-blue-500/20 dark:bg-blue-500/30 border-2 border-blue-500/50 dark:border-blue-400/50 rounded-[32px] flex items-center justify-center transform rotate-12 group-hover:rotate-0 transition-all duration-700 shadow-lg shadow-blue-500/20">
            <FlaskConical class="w-12 h-12 text-blue-600 dark:text-blue-400 group-hover:scale-110 transition-transform drop-shadow-lg" />
          </div>
          <div class="absolute -top-2 -right-2 w-8 h-8 bg-blue-500 dark:bg-blue-400 rounded-xl flex items-center justify-center animate-bounce shadow-[0_0_20px_rgba(59,130,246,0.5)]">
             <Zap class="w-4 h-4 text-white fill-current" />
          </div>
        </div>
        <div class="text-center space-y-3">
          <h2 class="text-2xl font-black text-slate-800 dark:text-white tracking-widest uppercase drop-shadow-lg">Initializing Lab</h2>
          <p class="text-sm text-slate-600 dark:text-slate-300 font-medium">正在连接实验室...</p>
          <div class="flex items-center gap-1 justify-center">
             <span class="w-2 h-2 bg-blue-500 dark:bg-blue-400 rounded-full animate-bounce [animation-delay:-0.3s] shadow-lg shadow-blue-500/50"></span>
             <span class="w-2 h-2 bg-blue-500 dark:bg-blue-400 rounded-full animate-bounce [animation-delay:-0.15s] shadow-lg shadow-blue-500/50"></span>
             <span class="w-2 h-2 bg-blue-500 dark:bg-blue-400 rounded-full animate-bounce shadow-lg shadow-blue-500/50"></span>
          </div>
        </div>
      </div>
    </div>

    <!-- Error / No Data State - 防止黑屏 -->
    <div v-else-if="!roomInfo" class="h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 flex flex-col items-center justify-center p-4 relative overflow-hidden">
      <div class="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] bg-red-600/10 dark:bg-red-500/20 rounded-full blur-[120px]"></div>
      <div class="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-purple-600/10 dark:bg-purple-500/20 rounded-full blur-[120px]"></div>

      <div class="relative z-10 flex flex-col items-center gap-6">
        <div class="w-24 h-24 bg-red-500/10 dark:bg-red-500/20 border-2 border-red-500/30 rounded-[32px] flex items-center justify-center shadow-lg">
          <Activity class="w-12 h-12 text-red-500 dark:text-red-400" />
        </div>
        <div class="text-center space-y-3">
          <h2 class="text-2xl font-black text-slate-800 dark:text-white tracking-widest uppercase">Connection Lost</h2>
          <p class="text-sm text-slate-600 dark:text-slate-300 font-medium">{{ loadError || '实验室连接异常' }}</p>
        </div>
        <div class="flex items-center gap-3 mt-4">
          <button
            @click="loadError = null; loading = true; loadGameState()"
            class="px-6 py-3 bg-blue-600 hover:bg-blue-500 text-white font-black rounded-xl transition-all shadow-lg active:scale-95 uppercase tracking-widest text-xs flex items-center gap-2"
          >
            <RefreshCw class="w-4 h-4" />
            重新连接
          </button>
          <button
            @click="router.push('/')"
            class="px-6 py-3 bg-slate-200 dark:bg-white/10 hover:bg-slate-300 dark:hover:bg-white/20 text-slate-700 dark:text-white font-black rounded-xl transition-all shadow-lg active:scale-95 uppercase tracking-widest text-xs flex items-center gap-2"
          >
            <ArrowLeft class="w-4 h-4" />
            返回大厅
          </button>
        </div>
      </div>
    </div>

    <template v-else>
      <!-- Dynamic Background -->
      <div class="fixed inset-0 pointer-events-none">
        <div class="absolute top-1/4 left-1/4 w-[50%] h-[50%] bg-blue-600/5 rounded-full blur-[150px] animate-pulse"></div>
        <div class="absolute bottom-1/4 right-1/4 w-[50%] h-[50%] bg-purple-600/5 rounded-full blur-[150px] animate-pulse delay-1000"></div>
        <div class="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/carbon-fibre.png')] opacity-20"></div>
        <!-- Scanning Line -->
        <div class="absolute top-0 left-0 w-full h-px bg-blue-500/20 shadow-[0_0_15px_rgba(59,130,246,0.5)] animate-scan"></div>
      </div>

      <!-- Compressed Header - 移动端优化 -->
      <header class="h-11 sm:h-16 bg-white/70 dark:bg-black/60 backdrop-blur-3xl border-b border-slate-200 dark:border-white/5 px-2 sm:px-6 flex items-center gap-2 sm:gap-3 z-50 sticky top-0 overflow-x-auto custom-scrollbar-hidden">
        <div class="flex items-center gap-2 sm:gap-4 shrink-0">
          <button
            @click="handleLeaveRoom"
            class="btn-touch flex items-center justify-center hover:bg-slate-100 dark:hover:bg-white/10 rounded-xl text-slate-500 hover:text-blue-500 transition-all touch-feedback"
          >
            <ArrowLeft class="icon-touch" />
          </button>
          <div class="hidden xs:block">
            <h2 class="text-xs-mobile font-black tracking-widest uppercase font-mono text-slate-400">Node: {{ id.substring(0, 6) }}</h2>
            <div class="flex items-center gap-1">
               <div :class="cn('w-1.5 h-1.5 sm:w-1 sm:h-1 rounded-full animate-pulse', roomInfo?.status === 'waiting' ? 'bg-amber-500' : 'bg-emerald-500')"></div>
               <span class="text-xs-mobile font-black uppercase text-slate-500 tracking-tighter">{{ roomInfo?.status === 'waiting' ? 'Idle' : 'Active' }}</span>
            </div>
          </div>
        </div>

        <!-- Spacer to push status buttons to right -->
        <div class="flex-1"></div>

        <!-- Global Status -->
        <div class="flex items-center gap-2 sm:gap-1.5 pl-3 border-l border-slate-200 dark:border-white/10 shrink-0">
          <div v-if="gameState?.status === 'playing'" class="flex items-center gap-1.5 px-2.5 sm:px-2 py-1 bg-blue-500/10 border border-blue-500/20 rounded-lg">
             <Activity class="w-3.5 h-3.5 sm:w-3 sm:h-3 text-blue-500" :class="timeRemaining <= 10 && 'animate-pulse'" />
             <span class="font-mono font-black text-xs-mobile text-blue-500">{{ timeRemaining }}S</span>
          </div>

          <button @click="showPlayers = !showPlayers" class="btn-touch relative flex items-center justify-center gap-1 bg-slate-100 dark:bg-white/5 rounded-lg border border-slate-200 dark:border-white/10 text-slate-500 hover:text-blue-500 touch-feedback">
             <Users class="icon-touch" :class="showPlayers && 'fill-current text-blue-500'" />
             <span class="text-[10px] sm:text-xs-mobile font-black text-slate-400">{{ allPlayers.length }}</span>
          </button>

          <button v-if="!roomInfo?.is_points_mode" @click="showHints = !showHints" class="btn-touch flex items-center justify-center bg-slate-100 dark:bg-white/5 rounded-lg border border-slate-200 dark:border-white/10 text-slate-500 hover:text-blue-500 touch-feedback">
             <Sparkles class="icon-touch" :class="showHints && 'fill-current text-blue-500'" />
          </button>

          <button @click="showChat = !showChat; hasNewMessage = false" class="btn-touch relative flex items-center justify-center bg-slate-100 dark:bg-white/5 rounded-lg border border-slate-200 dark:border-white/10 text-slate-500 hover:text-blue-500 touch-feedback">
             <MessageCircle class="icon-touch" :class="showChat && 'fill-current text-blue-500'" />
             <div v-if="hasNewMessage" class="absolute -top-1 -right-1 w-3 h-3 sm:w-2.5 sm:h-2.5 bg-rose-500 border-2 border-white dark:border-[#0d0d10] rounded-full animate-pulse"></div>
          </button>
        </div>
      </header>

      <!-- Main Action Focus Area -->
      <div class="flex-1 relative flex flex-col items-center justify-center p-2 sm:p-4 mb-16 sm:mb-20 overflow-hidden">
          <!-- Left Sidebar: Hint & Status -->
          <div :class="cn(
            'fixed left-0 top-0 bottom-0 w-full lg:w-80 z-[100] bg-white/95 dark:bg-slate-900/60 backdrop-blur-3xl border-r lg:border border-slate-200 dark:border-white/10 lg:rounded-[40px] lg:top-6 lg:bottom-52 lg:left-6 shadow-3xl transition-all duration-500 flex flex-col overflow-hidden',
            showHints ? 'translate-x-0 opacity-100' : '-translate-x-full opacity-0 pointer-events-none'
          )">
             <div class="p-4 py-3 border-b border-slate-200 dark:border-white/10 flex items-center justify-between bg-slate-50/50 dark:bg-white/[0.02]">
                <div class="flex items-center gap-2">
                   <div class="w-6 h-6 rounded-lg bg-blue-500/10 flex items-center justify-center">
                      <Trophy class="w-3.5 h-3.5 text-blue-500" />
                   </div>
                   <div>
                      <h3 class="text-[10px] font-black uppercase tracking-widest text-slate-800 dark:text-white">实验辅助情报</h3>
                      <p class="text-[8px] font-mono text-slate-400 uppercase tracking-tighter">Intelligence_Protocol</p>
                   </div>
                </div>
                <button @click="showHints = false" class="p-1 hover:bg-slate-200 dark:hover:bg-white/10 rounded-lg transition-colors text-slate-400 hover:text-slate-600 dark:hover:text-white">
                   <ArrowLeft class="w-4 h-4" />
                </button>
             </div>
             
             <div class="flex-1 overflow-y-auto p-3 custom-scrollbar space-y-4">
                <!-- Status Banners -->
                <div class="space-y-2">
                   <div v-if="allowedAny" class="bg-amber-500/10 border border-amber-500/20 p-2.5 rounded-xl animate-pulse">
                      <div class="flex items-center gap-1.5 text-amber-500 mb-0.5">
                         <Zap class="w-3 h-3 fill-current" />
                         <span class="text-[9px] font-black uppercase tracking-wider">AU 特权激活</span>
                      </div>
                      <p class="text-[8px] font-bold text-slate-500">已跳过所有反应规则限制</p>
                   </div>

                   <div v-if="gameState?.pending_draw_count > 0" class="bg-red-500/10 border border-red-500/20 p-2.5 rounded-xl animate-bounce">
                      <div class="flex items-center gap-1.5 text-red-500 mb-0.5">
                         <RefreshCw class="w-3 h-3 animate-spin-slow" />
                         <span class="text-[9px] font-black uppercase tracking-wider">正在加牌</span>
                      </div>
                      <p class="text-[8px] font-bold text-slate-500">需结算或叠加累计: {{ gameState.pending_draw_count }}</p>
                   </div>
                </div>

                <div v-if="roomInfo?.status === 'waiting'" class="space-y-3">
                   <!-- 积分模式提示 -->
                   <div v-if="roomInfo?.is_points_mode" class="p-3 bg-amber-500/10 border border-amber-500/20 rounded-xl flex items-center gap-2">
                      <Trophy class="w-4 h-4 text-amber-500 shrink-0" />
                      <div class="text-left">
                         <p class="text-[9px] font-black uppercase tracking-widest text-amber-600 dark:text-amber-500">Competitive Mode</p>
                         <p class="text-[8px] font-bold text-slate-500 mt-0.5">积分竞技模式：胜者获得积分，败者扣除积分。</p>
                      </div>
                   </div>

                   <div class="p-3 bg-blue-500/5 border border-blue-500/10 rounded-xl flex flex-col items-center text-center">
                      <Users class="w-5 h-5 text-blue-500 mb-1.5" />
                      <span class="text-[9px] font-black uppercase tracking-widest text-blue-500">准备就绪?</span>
                      <p class="text-[8px] font-bold text-slate-500 mt-0.5">当前连接数 {{ allPlayers.length }}/{{ roomInfo?.max_players }}，等待就绪后自动开启。</p>
                   </div>
                   <div class="p-3 bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl">
                      <div class="flex items-center gap-1.5 mb-1.5">
                         <QrCode class="w-3 h-3 text-blue-500" />
                         <span class="text-[9px] font-black uppercase tracking-widest text-slate-500">快速邀请</span>
                      </div>
                      <p class="text-[7px] font-bold text-slate-400 leading-relaxed uppercase">
                         点击中间区域的"招募伙伴"按钮可快速复制链接，或点击二维码图标让好友扫码加入反应室。
                      </p>
                   </div>
                </div>
                
                <div v-else class="py-8 flex flex-col items-center justify-center opacity-20 text-center">
                   <Timer class="w-6 h-6 mb-2" />
                   <p class="text-[9px] font-black uppercase tracking-widest">等待其他研究员行动</p>
                </div>

                <!-- Reaction-based Hints (场上物质反应提示) -->
                <div v-if="filteredReactionHints.length > 0 && gameState?.status === 'playing' && isMyTurn" class="pt-3 border-t border-slate-200 dark:border-white/10">
                   <div class="flex items-center justify-between mb-3">
                      <div class="flex items-center gap-1.5">
                         <Activity class="w-3 h-3 text-emerald-500" />
                         <span class="text-[9px] font-black uppercase tracking-widest text-slate-500">可接续反应物</span>
                      </div>
                      <button @click="fetchReactionHints" class="p-1 hover:bg-slate-100 dark:hover:bg-white/5 rounded-full transition-colors text-slate-400 hover:text-emerald-500">
                         <RefreshCw class="w-2.5 h-2.5" />
                      </button>
                   </div>
                   <div class="space-y-1.5">
                      <button
                         v-for="(hint, idx) in filteredReactionHints"
                         :key="idx"
                         @click="selectedSubstance = hint.substance; handlePlayCard()"
                         class="w-full text-left px-3 py-2 bg-white/50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl hover:border-emerald-500 hover:bg-emerald-500/5 transition-all group cursor-pointer"
                      >
                         <div class="flex items-center justify-between">
                            <span class="text-[10px] font-black dark:text-white" v-html="formatFormula(hint.substance)"></span>
                            <div class="flex items-center gap-1.5">
                              <span v-if="hint.name" class="text-[8px] font-bold text-slate-400 tracking-tighter">{{ hint.name }}</span>
                              <span v-else-if="hint.source" class="text-[7px] font-mono text-emerald-500/60 tracking-tighter">← <span v-html="formatFormula(hint.source)"></span></span>
                              <div class="w-1.5 h-1.5 rounded-full bg-emerald-500 group-hover:scale-125 transition-transform shadow-[0_0_8px_rgba(16,185,129,0.5)]"></div>
                            </div>
                         </div>
                      </button>
                   </div>
                </div>

                <!-- Database Trivia Hints -->
                <div v-if="randomHints.length > 0" class="pt-3 border-t border-slate-200 dark:border-white/10">
                   <div class="flex items-center justify-between mb-3">
                      <div class="flex items-center gap-1.5">
                         <Sparkles class="w-3 h-3 text-blue-500" />
                         <span class="text-[9px] font-black uppercase tracking-widest text-slate-500">实验小贴士</span>
                      </div>
                      <button @click="fetchRandomHints" class="p-1 hover:bg-slate-100 dark:hover:bg-white/5 rounded-full transition-colors text-slate-400 hover:text-blue-500">
                         <RefreshCw class="w-2.5 h-2.5" />
                      </button>
                   </div>
                   <div class="space-y-3">
                      <div v-for="hint in randomHints" :key="hint.id" class="relative pl-3">
                         <div class="absolute left-0 top-1 bottom-1 w-0.5 bg-blue-500/30 rounded-full"></div>
                         <h4 v-if="hint.title" class="text-[9px] font-black text-slate-600 dark:text-slate-300 mb-0.5">{{ hint.title }}</h4>
                         <p class="text-[8px] font-bold text-slate-400 leading-relaxed">{{ hint.content }}</p>
                      </div>
                   </div>
                </div>
             </div>
          </div>

          <!-- Latest Reaction Display -->
          <div v-if="gameState?.last_card" class="relative group scale-75 sm:scale-100 flex flex-col items-center justify-center">
             <div class="absolute -inset-16 bg-blue-600/10 rounded-full blur-[100px] opacity-50 group-hover:opacity-80 transition-opacity animate-pulse"></div>
             
             <!-- Double Play Display (Side by Side) -->
             <div v-if="gameState?.last_card?.reactants?.length > 0" class="flex items-center gap-6 sm:gap-10 relative z-10">
                <div v-for="(sub, idx) in gameState.last_card.reactants" :key="idx" class="relative group/card">
                   <div :class="cn(
                      'uno-card w-28 h-40 sm:w-32 h-48 rounded-[32px] flex flex-col items-center justify-center gap-4 hover:scale-105',
                      getDynamicCardClass(gameState?.last_card?.card, sub)
                   )">
                      <span class="text-[28px] sm:text-[36px] font-black font-mono italic drop-shadow-lg" v-html="formatFormula(sub)"></span>
                      <div class="px-3 py-1 bg-white/10 backdrop-blur-md rounded-lg border border-white/20 max-w-[85%]">
                         <span class="text-[8px] font-black tracking-widest truncate block text-center">{{ getSubstanceName(sub) }}</span>
                      </div>
                   </div>
                </div>
                <!-- Plus Operator -->
                <div class="w-10 h-10 rounded-xl bg-white/10 backdrop-blur-md border border-white/20 flex items-center justify-center text-white shadow-lg">
                   <Plus class="w-4 h-4 stroke-[4px]" />
                </div>
             </div>

             <!-- Single Play Display -->
             <div v-else :class="cn(
                'uno-card w-40 h-56 sm:w-48 h-64 rounded-[32px] flex flex-col items-center justify-center gap-4 sm:gap-6 hover:scale-105',
                getDynamicCardClass(gameState?.last_card?.card, gameState?.last_card?.substance)
              )">
                <div class="absolute top-4 left-4 opacity-20 text-[8px] uppercase font-black tracking-widest leading-none">Result</div>
                <span class="text-[32px] sm:text-[44px] font-black font-mono italic drop-shadow-lg leading-none" v-html="formatFormula(gameState?.last_card?.substance)"></span>
                <div class="px-4 py-1.5 bg-white/10 backdrop-blur-md rounded-xl border border-white/20 max-w-[85%]">
                   <span class="text-[9px] sm:text-[10px] font-black tracking-widest text-center block leading-tight">{{ getSubstanceName(gameState?.last_card?.substance) }}</span>
                </div>
                <div class="absolute bottom-4 right-4 opacity-30">
                   <FlaskConical class="w-4 h-4 fill-current" />
                </div>
             </div>

             <!-- Direction Ring -->
             <div class="absolute -inset-12 pointer-events-none">
                <div :class="cn(
                   'absolute -inset-12 pointer-events-none border-2 border-blue-500/10 rounded-full',
                   gameState?.direction === 1 ? 'animate-spin-slow' : 'animate-reverse-spin-slow'
                )"></div>
             </div>
          </div>

          <!-- Waiting for play state (Au triggered or Initial) -->
          <div v-else-if="gameState?.status === 'playing' && !gameState?.last_card" class="flex flex-col items-center gap-4 sm:gap-6 animate-in fade-in zoom-in duration-700">
             <div class="relative group">
                <div class="absolute -inset-8 bg-emerald-500/10 rounded-full blur-[60px] group-hover:bg-emerald-500/20 transition-all animate-pulse"></div>
                <div class="w-24 h-24 sm:w-32 sm:h-32 rounded-[32px] sm:rounded-[40px] border-2 border-emerald-500/30 flex items-center justify-center relative z-10 backdrop-blur-md bg-emerald-500/5">
                   <Zap class="w-10 h-10 sm:w-14 sm:h-14 text-emerald-500/40" />
                </div>
             </div>
             <div class="text-center relative z-10">
                <h3 class="text-lg sm:text-xl font-black text-slate-800 dark:text-white uppercase tracking-[0.2em]">
                   等待 {{ allPlayers[gameState?.current_player]?.username || '研究员' }} 出牌
                </h3>
                <p class="text-[8px] font-bold text-slate-500 mt-1 uppercase italic tracking-tighter">
                   Reaction Reactor Reseted _ New Deployment Window Open
                </p>
             </div>
          </div>
          
          <div v-else-if="roomInfo?.status === 'waiting'" class="flex flex-col items-center gap-6 sm:gap-10 animate-in fade-in zoom-in duration-1000">
             <div class="relative">
                <div class="absolute inset-0 bg-blue-500/10 rounded-full blur-[60px] animate-pulse"></div>
                <div class="w-24 h-24 sm:w-32 sm:h-32 rounded-[32px] sm:rounded-[40px] border-2 border-dashed border-blue-500/30 flex items-center justify-center rotate-45 group hover:rotate-0 transition-all duration-700 backdrop-blur-md bg-blue-500/5">
                   <FlaskConical class="w-10 h-10 sm:w-14 sm:h-14 text-blue-500/40 -rotate-45 group-hover:rotate-0 transition-all" />
                </div>
                <div v-if="roomInfo?.countdown > 0" class="absolute inset-0 flex items-center justify-center z-50 pointer-events-none">
                   <div class="scale-[3] sm:scale-[5] opacity-20 font-black italic select-none animate-ping text-blue-500">
                      {{ roomInfo.countdown }}
                   </div>
                </div>

                <div v-if="roomInfo?.countdown > 0" class="absolute -top-3 -right-3 bg-red-500 text-white px-4 py-1.5 rounded-xl text-lg font-black shadow-lg animate-bounce z-10">
                   {{ roomInfo.countdown }}
                </div>
                <div v-else class="absolute -top-3 -right-3 bg-amber-500 text-white px-3 py-1 rounded-lg text-[8px] font-black uppercase tracking-widest shadow-lg animate-pulse">
                   Ready Check
                </div>
             </div>

             <div class="flex flex-col items-center gap-6">
                <div class="flex flex-col items-center gap-3">
                  <h3 class="text-xl sm:text-2xl font-black text-slate-800 dark:text-white uppercase tracking-[0.1em] text-center">{{ roomInfo?.name || '实验室准备中' }}</h3>
                  
                  <!-- Compact Ready Button -->
                  <button 
                    @click="handleToggleReady"
                    :class="cn(
                      'px-8 sm:px-12 py-3 sm:py-5 rounded-2xl text-sm sm:text-lg font-black uppercase tracking-[0.2em] transition-all duration-500 shadow-xl relative overflow-hidden active:scale-95 text-white',
                      isReady ? 'bg-emerald-500 shadow-emerald-500/40' : 'bg-blue-600 shadow-blue-500/40'
                    )"
                  >
                    <div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 to-transparent -translate-x-full group-hover:animate-shimmer"></div>
                    <div class="flex items-center gap-3">
                      <Zap :class="cn('w-4 h-4 sm:w-6 sm:h-6', isReady ? 'fill-current' : 'animate-pulse')" />
                      <span>{{ isReady ? '已就绪' : '手动准备' }}</span>
                    </div>
                  </button>

                  <!-- Countdown Tip -->
                  <div v-if="roomInfo?.countdown > 0" class="flex flex-col items-center gap-1 mt-2">
                    <p class="text-[10px] font-black uppercase tracking-[0.2em] text-blue-500 animate-pulse">
                      实验即将开始: <span class="text-lg">{{ roomInfo.countdown }}</span>S
                    </p>
                    <p class="text-[7px] font-bold text-slate-400 dark:text-slate-600 uppercase tracking-tighter italic">
                      实验室压力充盈中，即将开启研究循环...
                    </p>
                  </div>
                </div>

                <div class="flex flex-col items-center gap-3 bg-white/50 dark:bg-white/5 backdrop-blur-xl p-4 sm:p-5 rounded-[24px] border border-slate-200 dark:border-white/10 shadow-sm w-full max-w-sm">
                  <div class="flex flex-wrap justify-center gap-2 sm:gap-3">
                    <div class="flex items-center gap-2 px-3 py-1.5 bg-slate-100 dark:bg-white/5 rounded-xl border border-slate-200 dark:border-white/10">
                      <Users class="w-3 h-3 text-blue-500" />
                      <span class="text-[8px] font-black uppercase tracking-widest text-slate-600 dark:text-slate-400">
                        研究员: {{ allPlayers.length }} / {{ roomInfo?.max_players }}
                      </span>
                    </div>
                    <div
                      @click="viewCurrentDeckConfig"
                      class="flex items-center gap-2 px-3 py-1.5 bg-slate-100 dark:bg-white/5 rounded-xl border border-slate-200 dark:border-white/10 cursor-pointer hover:bg-slate-200 dark:hover:bg-white/10 transition-colors"
                      title="点击查看牌组详情"
                    >
                      <FlaskConical class="w-3 h-3 text-emerald-500" />
                      <span class="text-[8px] font-black uppercase tracking-widest text-slate-600 dark:text-slate-400">
                        方案: {{ roomInfo?.deck_config?.name || '基础协议' }}
                      </span>
                    </div>
                  </div>

                  <div class="flex items-center gap-2 w-full">
                    <button
                        @click="handleCopyLink"
                        class="flex-1 flex items-center justify-center gap-2 py-2.5 bg-slate-800 dark:bg-white/10 hover:bg-slate-700 text-white rounded-xl transition-all active:scale-95 group shadow-md"
                    >
                        <Copy class="w-3 h-3 group-hover:rotate-12 transition-transform" />
                        <span class="text-[9px] font-black uppercase tracking-widest">招募成员</span>
                    </button>
                    <button
                        @click="showInviteFriendsModal = true"
                        class="flex-1 flex items-center justify-center gap-2 py-2.5 bg-blue-600 hover:bg-blue-500 text-white rounded-xl transition-all active:scale-95 group shadow-md"
                    >
                        <UserPlus class="w-3 h-3 group-hover:scale-110 transition-transform" />
                        <span class="text-[9px] font-black uppercase tracking-widest">邀请好友</span>
                    </button>
                    <button
                        @click="showQrModal = !showQrModal"
                        class="w-10 h-10 flex items-center justify-center bg-white dark:bg-white/10 border border-slate-200 dark:border-white/10 rounded-xl text-slate-500 hover:text-blue-500 transition-all active:scale-90 shadow-md"
                    >
                        <QrCode class="w-5 h-5" />
                    </button>
                  </div>

                  <!-- QR Code 浮窗 -->
                  <div v-if="showQrModal" class="mt-2 p-3 bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[24px] shadow-2xl animate-in zoom-in duration-300 flex flex-col items-center gap-3">
                     <div class="p-2 bg-white rounded-xl border-2 border-blue-500/20">
                        <img 
                          :src="`https://api.qrserver.com/v1/create-qr-code/?size=120x120&data=${encodeURIComponent(shareLink)}`" 
                          alt="Join QR Code"
                          class="w-32 h-32"
                        />
                     </div>
                     <div class="text-center pb-1">
                        <p class="text-[9px] font-black uppercase tracking-widest text-blue-500">实验室快传</p>
                     </div>
                  </div>
                </div>
             </div>
          </div>
        <!-- Table Console Background Removed or Simplified -->
        <div class="absolute inset-0 pointer-events-none overflow-hidden">
           <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-full h-full bg-blue-500/[0.02] dark:bg-blue-500/[0.05] rounded-full blur-[120px]"></div>
        </div>
      </div>

      <!-- Hand / Deck Area -->
      <div class="fixed bottom-0 left-0 right-0 z-[70] bg-white/70 dark:bg-black/60 backdrop-blur-2xl border-t border-slate-200 dark:border-white/5 flex flex-col items-center">
        <!-- Turn-related buttons and timer - 移动端优化 -->
        <div class="h-0 relative w-full flex justify-center">
           <div v-if="isMyTurn" class="absolute bottom-full mb-2 sm:mb-2 flex flex-col items-center gap-2 sm:gap-2 animate-in slide-in-from-bottom-4">
              <div class="flex items-center bg-white/90 dark:bg-black/80 backdrop-blur-xl border border-slate-200 dark:border-white/10 rounded-xl sm:rounded-lg p-1 sm:p-0.5 shadow-xl">
                <input
                  v-model="substanceInput"
                  @keyup.enter="handleInputPlay"
                  @focus="handleInputFocus"
                  @blur="handleInputBlur"
                  placeholder="手动注入化学式"
                  class="bg-transparent border-none outline-none text-sm sm:text-xs-mobile px-3 sm:px-2 py-1.5 sm:py-0.5 w-32 sm:w-40 font-black tracking-widest placeholder:text-slate-400 text-slate-900 dark:text-white"
                />

                <div class="flex items-center gap-1">
                   <button
                      @click="handleInputPlay"
                      class="btn-touch bg-blue-600 hover:bg-blue-500 rounded-lg sm:rounded-md flex items-center justify-center transition-all touch-feedback shadow-md group"
                      title="执行反应"
                   >
                      <ChevronRight class="w-4 h-4 sm:w-3.5 sm:h-3.5 text-white group-hover:translate-x-0.5 transition-transform" />
                   </button>

                   <div class="w-px h-5 sm:h-4 bg-slate-200 dark:bg-white/10 mx-1 sm:mx-0.5"></div>

                   <button
                      @click="handleDrawCard"
                      :disabled="!isMyTurn"
                      :class="cn(
                        'px-3 sm:px-2 btn-touch rounded-lg sm:rounded-md flex items-center justify-center gap-1.5 sm:gap-1 transition-all touch-feedback shadow-md group relative overflow-hidden',
                        isMyTurn ? (gameState?.pending_draw_count > 0 ? 'bg-red-600 hover:bg-red-500 text-white' : 'bg-slate-800 dark:bg-white/10 hover:bg-slate-700 dark:hover:bg-white/20 text-white') : 'bg-slate-200 dark:bg-slate-800 text-slate-400 cursor-not-allowed grayscale'
                      )"
                   >
                      <Plus v-if="!(gameState?.pending_draw_count > 0)" class="w-3 h-3 sm:w-2.5 sm:h-2.5" />
                      <RefreshCw v-else class="w-3 h-3 sm:w-2.5 sm:h-2.5 animate-spin-slow" />
                      <span class="text-xs-mobile font-black uppercase tracking-widest whitespace-nowrap">
                        摸牌{{ gameState?.pending_draw_count > 0 ? gameState.pending_draw_count : '2' }}张
                      </span>
                   </button>
                </div>
              </div>

              <div class="flex items-center gap-2 sm:gap-1.5">
                <div class="bg-blue-600/90 backdrop-blur-md px-4 sm:px-3 py-2 sm:py-1 rounded-full border border-white/20 shadow-md flex items-center gap-2.5 sm:gap-2 animate-slide-in-bottom">
                  <Zap class="w-3 h-3 sm:w-2.5 sm:h-2.5 fill-current animate-pulse text-white" />
                  <span class="text-xs-mobile font-black uppercase tracking-widest text-white">操作 ({{ timeRemaining }}s)</span>

                  <!-- 双联行动按钮 -->
                  <button
                    v-if="myData?.double_action_available"
                    @click.stop="toggleDoubleMode"
                    :class="cn(
                      'px-2.5 sm:px-2 py-1 sm:py-0.5 rounded-lg border border-white/20 transition-all flex items-center gap-1.5 relative overflow-hidden touch-feedback',
                      doubleMode ? 'bg-amber-500 text-white border-amber-400 shadow-sm' : 'bg-black/40 text-white/60 hover:text-white hover:bg-black/60'
                    )"
                  >
                     <div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/10 to-transparent -translate-x-full group-hover/btn:animate-shimmer"></div>
                     <Activity :class="cn('w-3.5 h-3.5 sm:w-3 sm:h-3', doubleMode && 'animate-spin')" />
                     <span class="text-xs-mobile font-black uppercase tracking-tighter">{{ doubleMode ? '解除超限' : '超限双联' }}</span>
                  </button>
                </div>
              </div>

              <!-- 双联模式提示状态 -->
              <div v-if="doubleMode" class="mt-1 flex flex-wrap items-center justify-center gap-3 animate-in slide-in-from-top-4 duration-500">
                <div class="flex items-center gap-2">
                  <div
                    @click="firstDoubleSubstance && removeSubstance(1)"
                    :class="cn(
                      'w-8 h-8 rounded-lg flex items-center justify-center border-2 transition-all duration-300 relative group/sub',
                      firstDoubleSubstance ? 'bg-blue-500/20 border-blue-500 shadow-md cursor-pointer hover:border-red-500/50' : 'bg-slate-800/50 border-white/10 opacity-50'
                    )"
                  >
                    <span v-if="firstDoubleSubstance" class="text-[9px] font-black group-hover/sub:opacity-20 transition-opacity" v-html="formatFormula(firstDoubleSubstance)"></span>
                    <X v-if="firstDoubleSubstance" class="w-3 h-3 text-red-500 absolute opacity-0 group-hover/sub:opacity-100 transition-opacity" />
                    <FlaskConical v-else class="w-3 h-3 text-slate-500" />
                  </div>
                  <div class="w-3 h-0.5 bg-blue-500/30"></div>
                  <div
                    @click="secondDoubleSubstance && removeSubstance(2)"
                    :class="cn(
                      'w-8 h-8 rounded-lg flex items-center justify-center border-2 transition-all duration-300 relative group/sub',
                      secondDoubleSubstance ? 'bg-blue-500/20 border-blue-500 shadow-md cursor-pointer hover:border-red-500/50' : 'bg-slate-800/50 border-white/10 opacity-50'
                    )"
                  >
                    <span v-if="secondDoubleSubstance" class="text-[9px] font-black group-hover/sub:opacity-20 transition-opacity" v-html="formatFormula(secondDoubleSubstance)"></span>
                    <X v-if="secondDoubleSubstance" class="w-3 h-3 text-red-500 absolute opacity-0 group-hover/sub:opacity-100 transition-opacity" />
                    <FlaskConical v-else class="w-3 h-3 text-slate-500" />
                  </div>
                </div>

                <div class="flex items-center gap-1.5">
                  <button
                    v-if="firstDoubleSubstance && secondDoubleSubstance"
                    @click="handleDoublePlay"
                    class="bg-emerald-600 hover:bg-emerald-500 text-white px-3 py-1.5 rounded-xl flex items-center gap-1.5 shadow-md animate-in zoom-in duration-300 group"
                  >
                    <span class="text-[9px] font-black uppercase tracking-widest">启动反应</span>
                    <Play class="w-3 h-3 fill-current group-hover:translate-x-0.5 transition-transform" />
                  </button>

                  <button
                    @click="toggleDoubleMode"
                    class="bg-slate-800/80 hover:bg-slate-700 text-white/80 px-3 py-1.5 rounded-xl flex items-center gap-1.5 border border-white/10 shadow-md transition-all"
                  >
                    <span class="text-[9px] font-black uppercase tracking-widest">取消</span>
                  </button>
                </div>
              </div>
           </div>
        </div>

        <div class="w-full max-w-7xl mx-auto flex justify-center items-end py-2 sm:py-1">
           <div ref="handContainer" class="hand-container-mobile w-full custom-scrollbar-hidden">
            <div v-if="roomInfo?.status === 'waiting'" class="flex flex-col items-center justify-center opacity-30 pb-1 min-w-full">
              <Loader2 class="w-8 h-8 sm:w-6 sm:h-6 mb-1 animate-spin text-blue-500" />
              <p class="font-black uppercase tracking-widest text-xs-mobile text-slate-500 text-center">正在同步量子状态并等待开场就绪...</p>
            </div>
            <template v-else-if="myData?.hand_cards?.length > 0">
              <div
                v-for="(card, index) in myData.hand_cards"
                :key="index"
                @click="isMyTurn && handleCardClick(card)"
                :class="cn(
                  'uno-card card-mobile flex flex-col items-center justify-center cursor-pointer shrink-0 touch-feedback',
                  getDynamicCardClass(card),
                  selectedCard === card && 'ring-2 ring-blue-500 scale-105 z-10',
                  !isMyTurn && 'opacity-60 grayscale cursor-not-allowed'
                )"
                :style="{
                  transform: selectedCard === card ? (isMobile ? 'translateY(-10px)' : 'translateY(-12px)') : 'none'
                }"
              >
                <div class="absolute top-1 left-1 text-xs-mobile sm:text-[6px] font-black opacity-30 uppercase tracking-tighter">{{ ELEMENTS_DATA[card.type] ? 'Elem' : 'Spec' }}</div>
                <div class="flex flex-col items-center justify-center">
                  <div class="text-base sm:text-base font-black font-mono italic tracking-tighter leading-none">{{ card.type }}</div>
                  <div v-if="card.effect || ['He','Ne','Ar','Kr'].includes(card.type)" class="mt-1 px-1.5 sm:px-1 py-0.5 bg-black/10 rounded-md text-xs-mobile sm:text-[8px] font-black uppercase tracking-tighter">
                    {{ ['He','Ne','Ar','Kr'].includes(card.type) ? '转向' : card.effect === 'Au' ? '跳过' : card.effect === '+2' ? '+2' : card.effect === '+4' ? '+4' : card.effect }}
                  </div>
                  <div v-else-if="ELEMENTS_DATA[card.type]" class="text-xs-mobile sm:text-[8px] font-bold opacity-80 mt-0.5 uppercase tracking-tighter font-serif italic text-black/40">
                    {{ ELEMENTS_DATA[card.type].name }}
                  </div>
                </div>
                <div class="absolute bottom-1 right-1 text-xs-mobile sm:text-[6px] font-mono opacity-40 uppercase tracking-tighter">
                  {{ card.effect ? 'Func' : 'Pass' }}
                </div>
              </div>
            </template>
            <div v-else class="flex flex-col items-center justify-center opacity-10 pb-3 sm:pb-4">
              <FlaskConical class="w-10 h-10 sm:w-12 sm:h-12 mb-1" />
              <p class="font-black uppercase tracking-widest text-xs-mobile">Inventory_Empty</p>
            </div>
           </div>
        </div>
      </div>

      <!-- Modern Substance Recombinator (Selection Modal) -->
      <div v-if="selectedCard && availableSubstances.length > 0" class="fixed inset-0 z-[100] flex items-center justify-center p-4 sm:p-6">
        <div class="absolute inset-0 bg-slate-900/40 dark:bg-black/90 backdrop-blur-xl animate-in fade-in" @click="selectedCard = null" />
        <div class="relative w-full max-w-2xl max-h-[90vh] bg-white dark:bg-[#0d0d10] border border-slate-200 dark:border-white/10 rounded-[32px] sm:rounded-[48px] shadow-2xl flex flex-col">
           <div class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-blue-600 via-purple-600 to-blue-600 opacity-50"></div>
           
           <div class="p-5 sm:p-8">
             <div class="flex flex-col md:flex-row justify-between items-start gap-4 sm:gap-6 mb-6 sm:mb-8">
               <div class="space-y-1.5 sm:space-y-3">
                  <div class="inline-flex items-center gap-1.5 px-2 py-0.5 bg-blue-500/10 border border-blue-500/20 rounded-full">
                     <Zap class="w-2.5 h-2.5 text-blue-600 dark:text-blue-400" />
                     <span class="text-[7px] sm:text-[9px] font-bold text-blue-600 dark:text-blue-400 uppercase tracking-widest">Protocol_Active</span>
                  </div>
                  <h3 class="text-xl sm:text-2xl font-black text-slate-900 dark:text-white tracking-tighter leading-none">
                    化学物质重组
                  </h3>
                  <p class="text-[10px] sm:text-xs text-slate-500 dark:text-slate-400 max-w-xs font-medium leading-relaxed">
                    请选择一个与 <span class="text-slate-900 dark:text-white font-black underline decoration-blue-500 underline-offset-4">{{ selectedCard.type }}</span> 兼容的目标物质。
                  </p>
               </div>
               
               <div class="relative group self-center md:self-auto hidden sm:block text-white">
                  <div class="absolute -inset-6 bg-blue-600/10 rounded-full blur-xl group-hover:bg-blue-600/20 transition-all"></div>
                  <div :class="cn('uno-card w-16 sm:w-24 h-22 sm:h-34 rounded-xl flex flex-col items-center justify-center scale-110 !cursor-default', getDynamicCardClass(selectedCard))">
                     <div class="text-xl sm:text-2xl font-black tracking-tighter">{{ selectedCard.type }}</div>
                  </div>
               </div>
             </div>

             <div class="substances-container-mobile custom-scrollbar mb-6 sm:mb-8">
                <button
                  v-for="(substance, index) in availableSubstances"
                  :key="index"
                  @click="selectedSubstance = substance"
                  :class="cn(
                    'substance-card-mobile border overflow-hidden',
                    selectedSubstance === substance ? 'bg-blue-600/10 border-blue-400 dark:border-blue-500 text-blue-600 dark:text-white shadow-md' : 'bg-slate-50/80 dark:bg-white/[0.03] border-slate-200 dark:border-white/5 text-slate-500 hover:bg-blue-50/50 dark:hover:bg-white/[0.05]'
                  )"
                >
                  <div :class="cn(
                    'w-7 h-7 sm:w-10 sm:h-10 rounded-lg sm:rounded-xl flex items-center justify-center border transition-all duration-500',
                    selectedSubstance === substance ? 'bg-blue-500/20 border-blue-500/30 rotate-12' : 'bg-slate-100 dark:bg-black/40 border-slate-200 dark:border-white/5 opacity-60'
                  )">
                    <FlaskConical :class="cn('w-3.5 h-3.5 sm:w-5 h-5', selectedSubstance === substance ? 'text-blue-600 dark:text-blue-400' : 'text-slate-400 dark:text-slate-600')" />
                  </div>
                  <span :class="cn('font-black tracking-widest text-[8px] sm:text-[10px] truncate w-full text-center', selectedSubstance === substance ? 'text-blue-600 dark:text-white' : 'text-slate-500')">{{ substance }}</span>
                  <div v-if="selectedSubstance === substance" class="absolute inset-0 bg-blue-500/5 animate-pulse"></div>
                </button>
             </div>

             <div class="flex gap-2 sm:gap-3">
                <button 
                  @click="selectedCard = null; selectedSubstance = null;" 
                  class="flex-1 h-10 sm:h-12 bg-slate-50 dark:bg-white/5 hover:bg-slate-100 dark:hover:bg-white/10 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white font-black rounded-xl transition-all uppercase tracking-widest text-[8px] sm:text-[10px] border border-slate-200 dark:border-white/5"
                >
                  终止
                </button>
                <button 
                  @click="handlePlayCard"
                  :disabled="!selectedSubstance"
                  class="flex-[2] h-10 sm:h-12 bg-blue-600 hover:bg-blue-500 text-white font-black rounded-xl transition-all shadow-lg disabled:opacity-50 disabled:grayscale flex items-center justify-center gap-2 group/confirm relative overflow-hidden"
                >
                  <span class="uppercase tracking-widest sm:tracking-[0.1em] text-[10px] sm:text-xs">执行反应</span>
                  <ChevronRight class="w-4 h-4 text-white group-hover/confirm:translate-x-1 transition-transform" />
                </button>
             </div>
           </div>
        </div>
      </div>

      <!-- Experimental Victory / Failure Protocol -->
      <div v-if="gameState?.status === 'finished'" class="fixed inset-0 z-[100] flex items-center justify-center p-4 overflow-hidden bg-slate-900/40 backdrop-blur-2xl">
        <div class="relative w-full max-w-lg bg-white dark:bg-[#0d0d10] border border-slate-200 dark:border-white/10 rounded-[32px] sm:rounded-[40px] shadow-2xl flex flex-col items-center text-center overflow-hidden animate-zoom-in p-6 sm:p-12">
           <div class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-transparent via-yellow-400 to-transparent animate-shimmer"></div>

           <div class="relative mb-6 sm:mb-10 transform-gpu">
              <div class="absolute inset-0 bg-blue-500/30 rounded-full blur-3xl animate-pulse"></div>
              <div class="w-20 h-20 sm:w-24 h-24 bg-gradient-to-br from-blue-500 to-blue-700 rounded-[24px] sm:rounded-[32px] flex items-center justify-center shadow-lg rotate-12">
                 <Trophy class="w-10 h-10 sm:w-12 sm:h-12 text-white" />
              </div>
              <div class="absolute -bottom-2 -right-2 w-8 h-8 bg-white rounded-lg flex items-center justify-center shadow-lg animate-bounce">
                 <Zap class="w-4 h-4 text-blue-600 fill-current" />
              </div>
           </div>

           <div class="space-y-2 sm:space-y-3 mb-8 sm:mb-12 px-2">
              <div class="inline-flex items-center gap-1.5 px-3 py-1 bg-blue-500/10 border border-blue-500/20 rounded-full">
                 <span class="w-1.5 h-1.5 bg-blue-500 rounded-full animate-ping"></span>
                 <span class="text-[7px] sm:text-[9px] font-black text-blue-600 dark:text-blue-400 uppercase tracking-widest font-mono">Mission_Success</span>
              </div>
              <template v-if="winner?.uid === user.uid">
                <h2 class="text-3xl sm:text-5xl font-black text-slate-900 dark:text-white tracking-tighter leading-none">
                  实验成功
                </h2>
                <p class="text-[10px] sm:text-xs text-slate-500 dark:text-slate-400 font-medium leading-relaxed max-w-xs mx-auto">
                  恭喜研究员！你已成功稳定了反应核心。
                </p>
              </template>
              <template v-else>
                <h2 class="text-3xl sm:text-5xl font-black text-slate-900 dark:text-white tracking-tighter leading-none">
                  反应终止
                </h2>
                <p class="text-[10px] sm:text-xs text-slate-500 dark:text-slate-400 font-medium leading-relaxed max-w-xs mx-auto">
                  实验由 <span class="text-slate-900 dark:text-white font-black">{{ winner?.username }}</span> 成功收官。
                </p>
              </template>

              <div v-if="gameState?.points_changes" class="w-full mt-4 p-3 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl">
                 <div class="flex items-center justify-between mb-2 border-b border-slate-200 dark:border-white/5 pb-1.5">
                    <span class="text-[8px] font-black uppercase tracking-widest text-slate-500">Rankings</span>
                    <span class="text-[8px] font-black uppercase tracking-widest text-blue-500">Points_Δ</span>
                 </div>
                 <div class="space-y-2">
                    <div 
                      v-for="(val, uid) in gameState.points_changes" 
                      :key="uid"
                      class="flex items-center justify-between group"
                    >
                       <div class="flex items-center gap-2">
                          <div class="w-1.5 h-1.5 rounded-full bg-slate-400"></div>
                          <span class="text-xs font-bold text-slate-600 dark:text-slate-300 text-left">
                            {{ gameState.players.find((p: any) => String(p.uid) === String(uid))?.username || 'User' }}
                          </span>
                       </div>
                       <span :class="cn(
                         'text-xs font-black font-mono',
                         val >= 0 ? 'text-emerald-500' : 'text-rose-500'
                       )">
                         {{ val >= 0 ? '+' : '' }}{{ val }}
                       </span>
                    </div>
                 </div>
              </div>
           </div>

           <div class="w-full space-y-4">
              <button 
                @click="router.push('/')"
                class="w-full h-14 sm:h-18 bg-blue-600 hover:bg-blue-500 text-white font-black rounded-[20px] sm:rounded-3xl transition-all shadow-xl hover:scale-105 active:scale-95 flex items-center justify-center gap-2 sm:gap-3 group relative overflow-hidden"
              >
                 <span class="uppercase tracking-widest sm:tracking-[0.3em] text-[11px] sm:text-sm">返回指挥大厅</span>
                 <ChevronRight class="w-5 h-5 sm:w-6 sm:h-6 group-hover:translate-x-1 transition-transform" />
              </button>
           </div>
        </div>
      </div>
    </template>
    <!-- Admin Management Modal -->
    <div v-if="showAdminModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/80 backdrop-blur-md" @click="showAdminModal = false"></div>
      <div class="relative w-full max-w-lg bg-white dark:bg-[#121216] border border-slate-200 dark:border-white/10 rounded-[40px] shadow-2xl overflow-hidden animate-in zoom-in duration-300">
        <div class="p-8 border-b border-slate-200 dark:border-white/5 bg-slate-50/50 dark:bg-white/[0.02]">
          <div class="flex items-center justify-between mb-2">
            <h3 class="text-2xl font-black text-slate-900 dark:text-white tracking-tighter flex items-center gap-3">
              <ShieldAlert class="w-6 h-6 text-red-500" />
              权限执行控制
            </h3>
            <button @click="showAdminModal = false" class="p-2 hover:bg-slate-200 dark:hover:bg-white/5 rounded-full transition-colors">
              <X class="w-6 h-6 text-slate-400" />
            </button>
          </div>
          <p class="text-[10px] text-slate-500 font-mono uppercase tracking-[0.2em]">Target: {{ adminTargetUser?.username }} (UID: {{ adminTargetUser?.uid }})</p>
        </div>

        <div class="p-8 space-y-8">
          <div class="grid grid-cols-2 gap-4">
            <button 
              @click="adminActionType = 'kick'; banReason = '你由于违规游戏而被踢出'"
              :class="cn(
                'flex flex-col items-center gap-3 p-6 rounded-3xl border transition-all group',
                adminActionType === 'kick' ? 'bg-amber-500/10 border-amber-500/50 text-amber-500' : 'bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/10 text-slate-500'
              )"
            >
              <UserMinus class="w-8 h-8 group-hover:scale-110 transition-transform" />
              <span class="text-xs font-black uppercase tracking-widest">驱逐出场</span>
            </button>
            <button 
              @click="adminActionType = 'ban'; banReason = '你由于违规游戏而被封禁'"
              :class="cn(
                'flex flex-col items-center gap-3 p-6 rounded-3xl border transition-all group',
                adminActionType === 'ban' ? 'bg-red-500/10 border-red-500/50 text-red-500' : 'bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/10 text-slate-500'
              )"
            >
              <Ban class="w-8 h-8 group-hover:scale-110 transition-transform" />
              <span class="text-xs font-black uppercase tracking-widest">限制访问</span>
            </button>
          </div>

          <div v-if="adminActionType === 'ban'" class="space-y-4 animate-in slide-in-from-top-4 duration-300">
            <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest block">封禁时长</label>
            <div class="grid grid-cols-4 gap-2">
              <button
                v-for="preset in banPresets"
                :key="preset.hours"
                @click="setBanDuration(preset.hours)"
                :class="cn(
                  'px-2 py-2 rounded-xl text-[10px] font-black uppercase tracking-wider transition-all border active:scale-95',
                  selectedBanPreset === preset.hours
                    ? 'bg-red-500/10 border-red-500/30 text-red-500 shadow-sm'
                    : 'bg-slate-50 dark:bg-black/20 border-slate-200 dark:border-white/10 text-slate-500 hover:border-red-500/20 hover:text-red-400'
                )"
              >
                {{ preset.label }}
              </button>
              <button
                @click="selectedBanPreset = null"
                :class="cn(
                  'px-2 py-2 rounded-xl text-[10px] font-black uppercase tracking-wider transition-all border active:scale-95',
                  selectedBanPreset === null
                    ? 'bg-red-500/10 border-red-500/30 text-red-500 shadow-sm'
                    : 'bg-slate-50 dark:bg-black/20 border-slate-200 dark:border-white/10 text-slate-500 hover:border-red-500/20 hover:text-red-400'
                )"
              >
                自定义
              </button>
            </div>
            <div v-if="selectedBanPreset === null" class="animate-in slide-in-from-top-2 duration-200">
              <input
                v-model="banUntil"
                type="datetime-local"
                :min="formatDatetimeLocal(new Date())"
                @focus="handleInputFocus"
                @blur="handleInputBlur"
                class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 rounded-2xl px-5 py-3 text-sm font-bold text-slate-700 dark:text-white focus:outline-none focus:border-red-500/50 transition-all"
              />
            </div>
            <div class="flex items-center gap-2 ml-1 mt-1">
              <div class="w-1.5 h-1.5 rounded-full" :class="banUntil ? 'bg-red-500 animate-pulse' : 'bg-slate-300 dark:bg-slate-600'"></div>
              <span class="text-[9px] font-bold text-slate-400 dark:text-slate-600 uppercase tracking-wider">
                截止: {{ banUntil ? new Date(banUntil).toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' }) + '（UTC+8）' : '未设置' }}
              </span>
            </div>
          </div>

          <div class="space-y-4">
            <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest block">操作事由</label>
            <div class="relative group">
              <div class="absolute inset-0 bg-red-500/5 rounded-2xl blur-lg group-focus-within:bg-red-500/10 transition-all"></div>
              <textarea
                v-model="banReason"
                placeholder="请输入详细的违规事由..."
                @focus="handleInputFocus"
                @blur="handleInputBlur"
                class="relative w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 rounded-2xl px-6 py-4 text-sm font-medium text-slate-700 dark:text-white focus:outline-none focus:border-red-500/50 min-h-[100px] transition-all"
              ></textarea>
            </div>
          </div>

          <button 
            @click="handleAdminAction"
            :class="cn(
              'w-full h-16 rounded-[24px] font-black uppercase tracking-[0.2em] text-xs transition-all shadow-xl active:scale-95',
              adminActionType === 'kick' ? 'bg-amber-500 hover:bg-amber-400 text-white shadow-amber-500/20' : 'bg-red-600 hover:bg-red-500 text-white shadow-red-500/20'
            )"
          >
            确认执行操作
          </button>
        </div>
      </div>
    </div>

    <!-- Invite Friends Modal -->
    <div v-if="showInviteFriendsModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/80 backdrop-blur-md" @click="showInviteFriendsModal = false"></div>
      <div class="relative w-full max-w-lg bg-white dark:bg-[#121216] border border-slate-200 dark:border-white/10 rounded-[40px] shadow-2xl overflow-hidden animate-in zoom-in duration-300">
        <div class="p-8 border-b border-slate-200 dark:border-white/5 bg-slate-50/50 dark:bg-white/[0.02]">
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-2xl font-black text-slate-900 dark:text-white tracking-tighter flex items-center gap-3">
                <UserPlus class="w-6 h-6 text-blue-500" />
                邀请好友加入
              </h3>
              <p class="text-[10px] text-slate-500 font-mono uppercase tracking-[0.2em] mt-2">选择一位好友发送游戏邀请</p>
            </div>
            <button @click="showInviteFriendsModal = false" class="p-2 hover:bg-slate-200 dark:hover:bg-white/5 rounded-full transition-colors">
              <X class="w-6 h-6 text-slate-400" />
            </button>
          </div>
        </div>

        <div class="p-8 max-h-[500px] overflow-y-auto custom-scrollbar">
          <div v-if="friendsList.length === 0" class="flex flex-col items-center justify-center py-16 opacity-20 grayscale">
            <Users class="w-16 h-16 mb-4" />
            <p class="text-sm font-black uppercase tracking-[0.2em]">暂无好友</p>
            <p class="text-[10px] mt-2 italic font-medium uppercase">请先添加好友后再邀请</p>
          </div>
          <div v-else class="space-y-3">
            <button
              v-for="friend in friendsList"
              :key="friend.uid"
              @click="sendGameInvite(friend)"
              class="w-full p-4 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-[24px] flex items-center justify-between hover:border-blue-500/30 hover:bg-blue-500/5 transition-all group"
            >
              <div class="flex items-center gap-4">
                <div class="relative">
                  <div class="w-12 h-12 rounded-xl bg-white dark:bg-white/10 flex items-center justify-center text-2xl border border-slate-200 dark:border-white/10 shadow-sm">
                    {{ friend.avatar || '🧪' }}
                  </div>
                  <div v-if="friend.is_online" class="absolute -bottom-1 -right-1 w-3.5 h-3.5 bg-emerald-500 border-3 border-white dark:border-[#121216] rounded-full shadow-lg shadow-emerald-500/20"></div>
                </div>
                <div class="text-left">
                  <div class="text-base font-bold text-slate-700 dark:text-white flex items-center gap-2">
                    {{ friend.nickname || friend.username }}
                    <span v-if="friend.is_online" class="px-2 py-0.5 bg-emerald-500/10 text-emerald-500 text-[8px] font-black rounded uppercase tracking-widest">Online</span>
                  </div>
                  <div class="text-[9px] text-slate-400 font-mono mt-1">UID: {{ friend.uid }}</div>
                </div>
              </div>
              <div class="w-8 h-8 rounded-lg bg-blue-500/10 flex items-center justify-center group-hover:bg-blue-500 transition-all">
                <Send class="w-4 h-4 text-blue-500 group-hover:text-white group-hover:translate-x-0.5 transition-all" />
              </div>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Players Floating Panel -->
    <div
      :class="cn(
        'fixed right-0 top-0 bottom-0 w-full lg:w-80 z-[100] bg-white/95 dark:bg-slate-900/60 backdrop-blur-3xl border-l lg:border border-slate-200 dark:border-white/10 lg:rounded-[40px] lg:top-6 lg:bottom-52 lg:right-6 shadow-3xl transition-all duration-500 flex flex-col overflow-hidden',
        showPlayers ? 'translate-x-0 opacity-100' : 'translate-x-full opacity-0 pointer-events-none'
      )"
    >
      <div class="p-4 py-3 border-b border-slate-200 dark:border-white/10 flex items-center justify-between bg-slate-50/50 dark:bg-white/[0.02]">
        <div class="flex items-center gap-2">
          <div class="w-6 h-6 rounded-lg bg-blue-500/10 flex items-center justify-center">
            <Users class="w-3.5 h-3.5 text-blue-500" />
          </div>
          <div>
            <h3 class="text-[10px] font-black uppercase tracking-widest text-slate-800 dark:text-white">研究员列表</h3>
            <p class="text-[8px] font-mono text-slate-400 uppercase tracking-tighter">Players_{{ allPlayers.length }}/{{ roomInfo?.max_players }}</p>
          </div>
        </div>
        <button @click="showPlayers = false" class="p-1 hover:bg-slate-200 dark:hover:bg-white/10 rounded-lg transition-colors text-slate-400 hover:text-slate-600 dark:hover:text-white">
          <X class="w-4 h-4" />
        </button>
      </div>

      <div ref="playersContainer" class="flex-1 overflow-y-auto p-3 custom-scrollbar space-y-2">
        <template v-if="allPlayers.length > 0">
          <div
            v-for="(player, index) in allPlayers"
            :key="player.uid || index"
            data-player-card
            :class="cn(
              'flex items-center gap-2 px-3 py-2.5 rounded-xl border transition-all',
              gameState?.current_player === index
                ? 'bg-blue-600 shadow-md shadow-blue-500/10 ring-1 ring-blue-500/20 border-blue-500'
                : (gameState ? 'bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/5 opacity-70 hover:opacity-100' : 'bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/10')
            )"
          >
            <div class="relative w-8 h-8 shrink-0">
              <div :class="cn(
                'w-full h-full rounded-lg flex items-center justify-center text-sm border overflow-hidden relative',
                gameState?.current_player === index ? 'bg-white text-blue-600 border-white/20' : 'bg-slate-100 dark:bg-slate-800 border-slate-200 dark:border-white/10'
              )">
                <img v-if="player.avatar && player.avatar.startsWith('data:')" :src="player.avatar" class="w-full h-full object-cover" />
                <span v-else>{{ player.avatar || '🧪' }}</span>

                <!-- Offline Overlay -->
                <div v-if="player.is_offline" class="absolute inset-0 bg-red-500/40 flex items-center justify-center backdrop-blur-[1px]">
                  <Activity class="w-3.5 h-3.5 text-white animate-pulse" />
                </div>
              </div>
              <!-- Action Progress Dots -->
              <div v-if="gameState" class="absolute -bottom-0.5 -right-0.5 flex gap-0.5">
                <div v-for="i in 2" :key="i" :class="cn('w-1.5 h-1.5 rounded-full border border-black/20', i <= (player.action_progress || 0) ? (gameState?.current_player === index ? 'bg-white' : 'bg-blue-500') : 'bg-slate-500')"></div>
              </div>
            </div>
            <div class="flex flex-col min-w-0 flex-1">
              <div class="flex items-center gap-1 leading-none">
                <span class="text-[11px] font-black truncate max-w-[80px] tracking-tight" :class="gameState?.current_player === index ? 'text-white' : 'text-slate-700 dark:text-slate-300'">{{ player.nickname || player.username }}</span>
                <span class="text-[8px] font-mono opacity-40 shrink-0" :class="gameState?.current_player === index ? 'text-white' : 'text-slate-500'">#{{ player.uid }}</span>
                <Zap v-if="player.double_action_available" :class="cn('w-2.5 h-2.5 fill-current', gameState?.current_player === index ? 'text-amber-300' : 'text-amber-500')" />
              </div>
              <!-- Status/Card Count -->
              <div class="flex items-center gap-1 mt-0.5">
                <template v-if="gameState">
                  <Trophy v-if="!player.is_offline" :class="cn('w-2.5 h-2.5', gameState?.current_player === index ? 'text-white' : 'text-slate-400')" />
                  <span v-if="!player.is_offline" :class="cn('text-[10px] font-mono font-bold', gameState?.current_player === index ? 'text-white/80' : 'text-slate-400')">{{ player.card_count || 0 }} 张</span>
                  <span v-else class="text-[9px] font-black uppercase text-red-500 animate-pulse tracking-tighter">OFFLINE</span>
                </template>
                <template v-else>
                  <span :class="cn('text-[10px] font-black uppercase tracking-widest', player.is_ready ? 'text-emerald-500' : 'text-slate-400')">
                    {{ player.is_ready ? 'READY' : 'WAIT' }}
                  </span>
                </template>
              </div>
            </div>
            <!-- Player Actions -->
            <div class="flex items-center gap-1 shrink-0">
              <button v-if="Number(player.uid) !== Number(user.uid) && !isFriend(player.uid)"
                      @click.stop="handleAddFriend(player)"
                      :class="cn('p-1 rounded-lg transition-colors touch-feedback', gameState?.current_player === index ? 'hover:bg-white/20 text-white' : 'hover:bg-amber-500/20 text-amber-500')"
                      title="添加好友"
              >
                <UserPlus class="w-3 h-3" />
              </button>
              <button v-if="Number(player.uid) !== Number(user.uid) && isFriend(player.uid)"
                      @click.stop="startPrivateChat(player)"
                      :class="cn('p-1 rounded-lg transition-colors touch-feedback', gameState?.current_player === index ? 'hover:bg-white/20 text-white' : 'hover:bg-blue-500/20 text-blue-500')"
                      title="私聊"
              >
                <MessageCircle class="w-3 h-3" />
              </button>
              <button v-if="Number(player.uid) !== Number(user.uid)"
                      @click.stop="handleReportPlayer(player)"
                      :class="cn('p-1 rounded-lg transition-colors touch-feedback', gameState?.current_player === index ? 'hover:bg-white/20 text-white' : 'hover:bg-rose-500/20 text-rose-500')"
                      title="举报玩家"
              >
                <Flag class="w-3 h-3" />
              </button>
              <button v-if="user.is_admin && Number(player.uid) !== Number(user.uid)"
                      @click.stop="openAdminAction(player)"
                      :class="cn('p-1 rounded-lg transition-colors touch-feedback', gameState?.current_player === index ? 'hover:bg-white/20 text-white' : 'hover:bg-red-500/20 text-red-500')"
                      title="管理玩家"
              >
                <ShieldAlert class="w-3 h-3" />
              </button>
            </div>
          </div>

          <!-- Empty Slots -->
          <div
            v-for="i in (roomInfo?.max_players || 0) - allPlayers.length"
            :key="'empty-slot-' + i"
            class="flex items-center gap-2 px-3 py-2.5 rounded-xl border border-dashed border-slate-200 dark:border-white/5 opacity-30"
          >
            <div class="w-8 h-8 rounded-lg border border-dashed border-slate-300 dark:border-white/10 flex items-center justify-center">
              <Plus class="w-3.5 h-3.5 text-slate-400" />
            </div>
            <span class="text-[10px] font-black uppercase tracking-tighter text-slate-400">EMPTY_SLOT</span>
          </div>
        </template>
        <div v-else class="flex flex-col items-center justify-center py-8 opacity-30">
          <Loader2 class="w-5 h-5 animate-spin mb-2" />
          <span class="text-[10px] font-black uppercase tracking-widest italic">Awaiting Peers...</span>
        </div>
      </div>
    </div>

    <!-- Mobile Overlay for Players Panel -->
    <div
      v-if="showPlayers"
      class="fixed inset-0 bg-white/10 dark:bg-black/20 backdrop-blur-[2px] z-[95] lg:hidden"
      @click="showPlayers = false"
    ></div>

    <!-- Chat Floating Sidebar -->
    <div
      :class="cn(
        'fixed right-0 top-0 bottom-0 w-full lg:w-80 z-[100] lg:top-6 lg:bottom-52 lg:right-6 transition-all duration-500 ease-[cubic-bezier(0.23,1,0.32,1)] flex flex-col',
        showChat ? 'translate-x-0 opacity-100' : 'translate-x-full opacity-0 pointer-events-none'
      )"
    >
      <ChatBox
        :roomId="id"
        title="实验内通信线程"
        maxHeight="100%"
        class="h-full !bg-white/95 dark:!bg-slate-900/60 backdrop-blur-3xl shadow-3xl lg:rounded-[40px] border-l lg:border border-slate-200 dark:border-white/10"
        @close="showChat = false"
        @input-focus="handleInputFocus"
        @input-blur="handleInputBlur"
      />
    </div>

    <!-- Mobile Overlay for Chat -->
    <div
      v-if="showChat"
      class="fixed inset-0 bg-white/10 dark:bg-black/20 backdrop-blur-[2px] z-[95] lg:hidden"
      @click="showChat = false"
    ></div>

    <!-- 牌组详情查看模态框 -->
    <div v-if="showDeckDetailModal && roomInfo?.deck_config" class="fixed inset-0 z-[200] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-900/40 dark:bg-black/80 backdrop-blur-md" @click="showDeckDetailModal = false" />
      <div class="relative w-full max-w-2xl bg-white dark:bg-[#121216] border border-slate-200 dark:border-white/10 rounded-[32px] shadow-2xl overflow-hidden">
         <div class="px-5 py-4 border-b border-slate-100 dark:border-white/5 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 bg-emerald-500/10 border border-emerald-500/20 rounded-xl flex items-center justify-center text-emerald-500">
                <FlaskConical class="w-4 h-4" />
              </div>
              <div>
                <h2 class="text-base font-black text-slate-800 dark:text-white tracking-tight leading-none">{{ roomInfo.deck_config.name }}</h2>
                <p class="text-[8px] text-slate-400 dark:text-slate-500 font-mono uppercase tracking-widest mt-1">Deck_Configuration</p>
              </div>
            </div>
            <button @click="showDeckDetailModal = false" class="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl transition-colors text-slate-400 hover:text-slate-900 dark:hover:text-white">
              <X class="w-4 h-4" />
            </button>
         </div>
         <div class="p-5 max-h-[60vh] overflow-y-auto custom-scrollbar space-y-4">
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-2">
              <div class="p-3 bg-slate-50 dark:bg-white/5 rounded-xl border border-slate-200 dark:border-white/10">
                <p class="text-[8px] text-slate-400 mb-1 uppercase tracking-wider">牌组名称</p>
                <p class="text-[11px] font-black text-slate-900 dark:text-white">{{ roomInfo.deck_config.name }}</p>
              </div>
              <div class="p-3 bg-slate-50 dark:bg-white/5 rounded-xl border border-slate-200 dark:border-white/10">
                <p class="text-[8px] text-slate-400 mb-1 uppercase tracking-wider">元素种类</p>
                <p class="text-[11px] font-black text-blue-600 dark:text-blue-400">{{ Object.keys(roomInfo.deck_config.cards || {}).length }} 种</p>
              </div>
              <div class="p-3 bg-slate-50 dark:bg-white/5 rounded-xl border border-slate-200 dark:border-white/10">
                <p class="text-[8px] text-slate-400 mb-1 uppercase tracking-wider">总卡牌数</p>
                <p class="text-[11px] font-black text-slate-900 dark:text-white">{{ (Object.values(roomInfo.deck_config.cards || {}) as number[]).reduce((a, b) => a + b, 0) }} 张</p>
              </div>
              <div class="p-3 bg-slate-50 dark:bg-white/5 rounded-xl border border-slate-200 dark:border-white/10">
                <p class="text-[8px] text-slate-400 mb-1 uppercase tracking-wider">起始手牌</p>
                <p class="text-[11px] font-black text-slate-900 dark:text-white">{{ roomInfo.deck_config.initial_cards || 7 }} 张</p>
              </div>
            </div>
            <div class="p-4 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl">
              <div class="flex items-center justify-between mb-3">
                <span class="text-[9px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">卡牌配置</span>
                <span class="text-[8px] text-blue-500/40 font-mono">CARD_LIST</span>
              </div>
              <div class="grid grid-cols-2 sm:grid-cols-3 gap-2 max-h-64 overflow-y-auto custom-scrollbar pr-1">
                <div
                  v-for="(count, formula) in roomInfo.deck_config.cards"
                  :key="formula"
                  class="p-2.5 bg-white dark:bg-black/20 rounded-lg border border-slate-200 dark:border-white/10"
                >
                  <div class="flex items-center justify-between">
                    <span class="text-[10px] font-black text-slate-900 dark:text-white font-mono" v-html="String(formula).replace(/(\d+)/g, '<sub>$1</sub>')"></span>
                    <span class="text-[9px] font-black text-blue-600 dark:text-blue-400">×{{ count }}</span>
                  </div>
                </div>
              </div>
            </div>
         </div>
         <div class="px-5 py-3 border-t border-slate-100 dark:border-white/5 flex justify-end">
            <button @click="showDeckDetailModal = false" class="px-4 py-2 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-700 dark:text-slate-300 font-bold rounded-xl transition-all uppercase tracking-widest text-[10px] border border-slate-200 dark:border-white/5">
              关闭
            </button>
         </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 游戏内特定滚动条隐藏 */
:deep(.custom-scrollbar-hidden::-webkit-scrollbar) {
  display: none;
}
:deep(.custom-scrollbar-hidden) {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
</style>
