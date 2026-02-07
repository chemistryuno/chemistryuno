<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { gameAPI, adminAPI, friendAPI, authAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import websocket from '../utils/websocket'
import { ArrowLeft, Play, RefreshCw, Zap, Activity, FlaskConical, Trophy, ChevronRight, Loader2, Users, Timer, Plus, QrCode, Copy, Sparkles, ShieldAlert, Ban, UserMinus, X, MessageCircle, UserPlus, Flag } from 'lucide-vue-next'
import { cn } from '../utils/cn'
import ChatBox from '../components/ChatBox.vue'

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

const loading = ref(false)
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

const isReady = computed(() => {
  return roomInfo.value?.ready_uids?.includes(Number(user.value.uid))
})

const handleToggleReady = async () => {
  if (!roomInfo.value || !user.value.uid) return
  
  // 乐观更新：立即切换状态
  const uidNum = Number(user.value.uid)
  const isCurrentlyReady = roomInfo.value.ready_uids.includes(uidNum)
  
  if (isCurrentlyReady) {
    roomInfo.value.ready_uids = roomInfo.value.ready_uids.filter((id: number) => id !== uidNum)
  } else {
    roomInfo.value.ready_uids = [...roomInfo.value.ready_uids, uidNum]
  }

  try {
    await gameAPI.ready(id)
    // 状态也会通过 WebSocket 更新，但手动标记一下提高体验
  } catch (error: any) {
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
  return friendsList.value.some(f => Number(f.uid) === Number(uid))
}

const handleAddFriend = async (player: any) => {
  try {
    await friendAPI.sendRequest(player.uid)
    showAlert(`已向研究员 ${player.username} 发送同步请求，等待量子握手。`, '请求已发送')
  } catch (error: any) {
    showAlert(error.response?.data?.error || '请求发送失败', '链路故障')
  }
}

// Chat system
const showChat = ref(false)
const hasNewMessage = ref(false)
const showQrModal = ref(false)

const startPrivateChat = (player: any) => {
  if (!isFriend(player.uid)) {
    showAlert('只有互为好友的研究员才能开启单向加密传输。', '权限受限')
    return
  }
  showChat.value = true
  hasNewMessage.value = false
  window.dispatchEvent(new CustomEvent('start-private-chat', {
    detail: { uid: player.uid, username: player.username }
  }))
}

// Admin management state
const showAdminModal = ref(false)
const adminTargetUser = ref<any>(null)
const adminActionType = ref<'kick' | 'ban'>('kick')
const banHours = ref(24)
const banReason = ref('你由于违规游戏而被踢出')

watch(showChat, (val) => {
  if (val) hasNewMessage.value = false
})

const openAdminAction = (player: any) => {
  if (!user.value.is_admin || player.uid === user.value.uid) return
  adminTargetUser.value = player
  adminActionType.value = 'kick'
  banReason.value = '你由于违规游戏而被踢出'
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
      await adminAPI.kickPlayer(id, adminTargetUser.value.uid, banReason.value)
      showAlert('已踢出该玩家', '成功')
    } else {
      await adminAPI.banUser(adminTargetUser.value.uid, banHours.value, banReason.value)
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
  'Cl': { name: '氯', class: 'element-Cl' },
  'Na': { name: '钠', class: 'element-Na' },
  'Mg': { name: '镁', class: 'element-Mg' },
  'Al': { name: '铝', class: 'element-Al' },
  'Cu': { name: '铜', class: 'element-Cu' },
  'Fe': { name: '铁', class: 'element-Fe' },
  'Zn': { name: '锌', class: 'element-Zn' },
  'Ag': { name: '银', class: 'element-Ag' },
  'K': { name: '钾', class: 'element-K' },
  'Ca': { name: '钙', class: 'element-Ca' },
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

const showLogs = ref(false)
const showHints = ref(true)

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
})

const handleGameUpdate = (message: any) => {
  // 如果收到的是完整的游戏状态对象
  if (message.data && typeof message.data === 'object') {
    gameState.value = message.data
    if (isMyTurn.value) {
      fetchTurnSubstances()
    }
  } else {
    // 如果收到的是房间ID字符串，则重新拉取完整状态
    loadGameState().then(() => {
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

const loadGameState = async () => {
  if (isRedirecting.value) return
  try {
    loading.value = true
    const response = await gameAPI.getRoomState(id)
    const data = response.data
    
    roomInfo.value = {
      id: data.id,
      name: data.name,
      players: data.players,
      ready_uids: data.ready_uids || [],
      countdown: data.countdown || 0,
      max_players: data.max_players,
      status: data.status,
      is_points_mode: data.is_points_mode,
      deck_config: data.deck_config
    }
    
    playersInfo.value = data.players_info || []
    
    if (data.game_state) {
      gameState.value = data.game_state
    }
    
    loading.value = false
  } catch (error: any) {
    console.error('加载游戏状态失败:', error)
    loading.value = false
    
    if (error.response?.status === 404) {
      isRedirecting.value = true
      showAlert('房间不存在或已被关闭', '未知实验室')
      router.push('/')
    } else if (error.response?.status === 401) {
      isRedirecting.value = true
      showAlert('身份验证失败，请重新登录', '准入失败')
      router.push('/login')
    } else if (error.response?.status === 403) {
      isRedirecting.value = true
      showAlert('您不在该房间中', '准入失败')
      router.push('/')
    } else {
      // 这里的 400 错误通常也是房间不存在（如果后端还没改完）
      isRedirecting.value = true
      showAlert('实验环境加载异常', '系统错误')
      router.push('/')
    }
  }
}

onMounted(() => {
  friendAPI.getFriends().then(res => friendsList.value = res.data)
  loadGameState().then(() => {
    websocket.joinRoom(id)
    websocket.on('game_update', handleGameUpdate)
    websocket.on('player_joined', loadGameState)
    websocket.on('player_left', loadGameState)
    websocket.on('action_toast', handleActionToast)
    websocket.on('room_terminated', handleRoomTerminated)
    websocket.on('player_kicked', handlePlayerKicked)
    websocket.on('chat', handleChatNotify)
    websocket.on('private_chat', handleChatNotify)
  })
})

onUnmounted(() => {
  if (timerInterval) clearInterval(timerInterval)
  websocket.leaveRoom()
  websocket.off('game_update', handleGameUpdate)
  websocket.off('player_joined', loadGameState)
  websocket.off('player_left', loadGameState)
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

    if (matchingSubs.length === 1) {
      // 只有一种可能，直接出
      await gameAPI.playCard(id, card, matchingSubs[0])
      selectedCard.value = null
      selectedSubstance.value = null
      availableSubstances.value = []
    } else {
      // 多种可能，显示选择器
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
    const confirmed = await showConfirm('确定要离开当前实验房间吗？', '中断实验')
    if (confirmed) {
      await gameAPI.leaveRoom(id)
      router.push('/')
    }
  } catch (error) {
    console.error('离开房间失败:', error)
    router.push('/')
  }
}

const shareLink = computed(() => window.location.href)
const handleCopyLink = async () => {
  try {
    await navigator.clipboard.writeText(shareLink.value)
    showAlert('实验邀请链接已复制到剪贴板，快发送给你的科研伙伴吧！', '任务下达')
  } catch (err) {
    showAlert('链接复制失败，请手动复制浏览器地址栏', '设备故障')
  }
}

const getCardStyle = (card: any) => {
  if (!card) return ''
  const nobleGases = ['He', 'Ne', 'Ar', 'Kr']
  if (nobleGases.includes(card.type)) return 'noble'
  if (card.effect === 'Au' || card.type === 'Au') return 'gold' // Au 特效
  if (card.effect === '+2' || card.effect === '+4') return 'special'
  
  // 如果在 ELEMENTS_DATA 中有，返回对应的颜色类
  if (ELEMENTS_DATA[card.type]) return '' 
  
  return 'element'
}

const getDynamicCardClass = (card: any) => {
  if (ELEMENTS_DATA[card.type]) return ELEMENTS_DATA[card.type].class
  const style = getCardStyle(card)
  if (style === 'noble') return 'card-noble'
  if (style === 'gold') return 'card-gold'
  if (style === 'special') return 'card-special'
  return ''
}

const isMobile = ref(false)
const handContainer = ref<HTMLElement | null>(null)
const substancesContainer = ref<HTMLElement | null>(null)

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
  const handleResize = () => {
    isMobile.value = window.innerWidth < 640
  }
  window.addEventListener('resize', handleResize)
  
  // 初始化拖拽滑动
  setTimeout(() => {
    setupDraggable(handContainer.value)
    setupDraggable(substancesContainer.value)
  }, 500)

  onUnmounted(() => window.removeEventListener('resize', handleResize))
})
</script>

<template>
  <div class="h-screen bg-slate-50 dark:bg-[#0a0a0c] text-slate-900 dark:text-white overflow-hidden flex flex-col font-sans selection:bg-blue-500/30">
    <!-- Loading State -->
    <div v-if="loading" class="h-screen bg-slate-50 dark:bg-[#0a0a0c] flex flex-col items-center justify-center p-4 relative overflow-hidden">
      <!-- Background Elements -->
      <div class="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] bg-blue-600/10 rounded-full blur-[120px] animate-pulse"></div>
      <div class="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-purple-600/10 rounded-full blur-[120px]"></div>
      <div class="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/carbon-fibre.png')] opacity-20"></div>

      <div class="relative z-10 flex flex-col items-center gap-6 animate-in fade-in zoom-in duration-700">
        <div class="relative group">
          <div class="w-24 h-24 bg-blue-500/10 border border-blue-500/30 rounded-[32px] flex items-center justify-center transform rotate-12 group-hover:rotate-0 transition-all duration-700">
            <FlaskConical class="w-12 h-12 text-blue-400 group-hover:scale-110 transition-transform" />
          </div>
          <div class="absolute -top-2 -right-2 w-8 h-8 bg-blue-500 rounded-xl flex items-center justify-center animate-bounce shadow-[0_0_20px_rgba(59,130,246,0.5)]">
             <Zap class="w-4 h-4 text-white fill-current" />
          </div>
        </div>
        <div class="text-center space-y-2">
          <h2 class="text-2xl font-black text-slate-800 dark:text-white tracking-widest uppercase">Initializing Lab</h2>
          <div class="flex items-center gap-1 justify-center">
             <span class="w-1.5 h-1.5 bg-blue-500 rounded-full animate-bounce [animation-delay:-0.3s]"></span>
             <span class="w-1.5 h-1.5 bg-blue-500 rounded-full animate-bounce [animation-delay:-0.15s]"></span>
             <span class="w-1.5 h-1.5 bg-blue-500 rounded-full animate-bounce"></span>
          </div>
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

      <!-- Compressed Header -->
      <header class="h-[56px] sm:h-[64px] bg-white/70 dark:bg-black/60 backdrop-blur-3xl border-b border-slate-200 dark:border-white/5 px-3 sm:px-6 flex items-center gap-3 z-50 sticky top-0 overflow-x-auto custom-scrollbar-hidden">
        <div class="flex items-center gap-2 sm:gap-4 shrink-0">
          <button 
            @click="handleLeaveRoom" 
            class="w-8 h-8 flex items-center justify-center hover:bg-slate-100 dark:hover:bg-white/10 rounded-xl text-slate-500 hover:text-blue-500 transition-all"
          >
            <ArrowLeft class="w-4 h-4 sm:w-5 sm:h-5" />
          </button>
          <div class="hidden xs:block">
            <h2 class="text-[9px] font-black tracking-widest uppercase font-mono text-slate-400">Node: {{ id.substring(0, 6) }}</h2>
            <div class="flex items-center gap-1">
               <div :class="cn('w-1 h-1 rounded-full animate-pulse', roomInfo?.status === 'waiting' ? 'bg-amber-500' : 'bg-emerald-500')"></div>
               <span class="text-[7px] font-black uppercase text-slate-500 tracking-tighter">{{ roomInfo?.status === 'waiting' ? 'Idle' : 'Active' }}</span>
            </div>
          </div>
        </div>

        <!-- Players Horizontal Bar -->
        <div class="flex-1 flex items-center gap-1.5 sm:gap-3 overflow-x-auto custom-scrollbar-hidden py-1">
          <template v-if="allPlayers.length > 0">
            <div 
              v-for="(player, index) in allPlayers"
              :key="player.uid || index"
              :class="cn(
                'flex items-center gap-1.5 sm:gap-2 px-2 py-1 rounded-xl border transition-all shrink-0',
                gameState?.current_player === index 
                  ? 'bg-blue-600 shadow-md shadow-blue-500/10 ring-1 ring-blue-500/20 border-blue-500' 
                  : (gameState ? 'bg-slate-100 dark:bg-white/5 border-slate-200 dark:border-white/5 opacity-60' : 'bg-slate-100 dark:bg-white/5 border-slate-200 dark:border-white/10')
              )"
            >
              <div class="relative w-6 h-6 sm:w-8 sm:h-8 shrink-0">
                <div :class="cn(
                  'w-full h-full rounded-lg flex items-center justify-center text-xs border overflow-hidden relative',
                   gameState?.current_player === index ? 'bg-white text-blue-600 border-white/20' : 'bg-slate-100 dark:bg-slate-800 border-slate-200 dark:border-white/10'
                )">
                   <img v-if="player.avatar && player.avatar.startsWith('data:')" :src="player.avatar" class="w-full h-full object-cover" />
                   <span v-else>{{ player.avatar || '🧪' }}</span>
                   
                   <!-- Offline Overlay -->
                   <div v-if="player.is_offline" class="absolute inset-0 bg-red-500/40 flex items-center justify-center backdrop-blur-[1px]">
                      <Activity class="w-3.5 h-3.5 text-white animate-pulse" />
                   </div>
                </div>
                <!-- Action Progress Dots (Only during gameplay) -->
                <div v-if="gameState" class="absolute -bottom-0.5 -right-0.5 flex gap-0.5">
                  <div v-for="i in 2" :key="i" :class="cn('w-1.5 h-1.5 rounded-full border border-black/20', i <= (player.action_progress || 0) ? (gameState?.current_player === index ? 'bg-white' : 'bg-blue-500') : 'bg-slate-500')"></div>
                </div>
              </div>
              <div class="flex flex-col min-w-0">
                <div class="flex items-center gap-1 leading-none">
                  <span class="text-[8px] font-black truncate max-w-[40px] sm:max-w-[60px] tracking-tight" :class="gameState?.current_player === index ? 'text-white' : 'text-slate-500'">{{ player.username }}</span>
                  <span class="text-[6px] font-mono opacity-40 shrink-0" :class="gameState?.current_player === index ? 'text-white' : 'text-slate-500'">#{{ player.uid }}</span>
                  <Zap v-if="player.double_action_available" :class="cn('w-2 h-2 fill-current', gameState?.current_player === index ? 'text-amber-300' : 'text-amber-500')" />
                  <!-- Player Actions -->
                  <div class="flex items-center gap-0.5 ml-auto">
                    <button v-if="Number(player.uid) !== Number(user.uid) && !isFriend(player.uid)" 
                            @click.stop="handleAddFriend(player)"
                            :class="cn('p-0.5 rounded transition-colors', gameState?.current_player === index ? 'hover:bg-white/20 text-white' : 'hover:bg-amber-500/20 text-amber-500')"
                            title="添加好友"
                    >
                      <UserPlus class="w-2.5 h-2.5" />
                    </button>
                    <button v-if="Number(player.uid) !== Number(user.uid) && isFriend(player.uid)" 
                            @click.stop="startPrivateChat(player)"
                            :class="cn('p-0.5 rounded transition-colors', gameState?.current_player === index ? 'hover:bg-white/20 text-white' : 'hover:bg-blue-500/20 text-blue-500')"
                            title="私聊"
                    >
                      <MessageCircle class="w-2.5 h-2.5" />
                    </button>
                    <!-- Report Player -->
                    <button v-if="Number(player.uid) !== Number(user.uid)"
                            @click.stop="handleReportPlayer(player)"
                            :class="cn('p-0.5 rounded transition-colors', gameState?.current_player === index ? 'hover:bg-white/20 text-white' : 'hover:bg-rose-500/20 text-rose-500')"
                            title="举报玩家"
                    >
                      <Flag class="w-2.5 h-2.5" />
                    </button>
                    <!-- Admin Actions -->
                    <button v-if="user.is_admin && Number(player.uid) !== Number(user.uid)" 
                            @click.stop="openAdminAction(player)" 
                            :class="cn('p-0.5 rounded transition-colors', gameState?.current_player === index ? 'hover:bg-white/20 text-white' : 'hover:bg-red-500/20 text-red-500')"
                            title="管理玩家"
                    >
                      <ShieldAlert class="w-2.5 h-2.5" />
                    </button>
                  </div>
                </div>
                <!-- Status/Card Count -->
                <div class="flex items-center gap-1">
                  <template v-if="gameState">
                    <Trophy v-if="!player.is_offline" :class="cn('w-2 h-2', gameState?.current_player === index ? 'text-white' : 'text-slate-400')" />
                    <span v-if="!player.is_offline" :class="cn('text-[7px] font-mono font-bold', gameState?.current_player === index ? 'text-white/80' : 'text-slate-400')">{{ player.card_count || 0 }}</span>
                    <span v-else class="text-[6px] font-black uppercase text-red-500 animate-pulse tracking-tighter">OFFLINE</span>
                  </template>
                  <template v-else>
                    <span :class="cn('text-[6px] font-black uppercase tracking-widest', player.is_ready ? 'text-emerald-500' : 'text-slate-400')">
                       {{ player.is_ready ? 'READY' : 'WAIT' }}
                    </span>
                  </template>
                </div>
              </div>
            </div>

            <!-- Empty Slots in Top Bar -->
            <div 
              v-for="i in (roomInfo?.max_players || 0) - allPlayers.length" 
              :key="'empty-top-' + i"
              class="flex items-center gap-1.5 px-2 py-1 rounded-xl border border-dashed border-slate-200 dark:border-white/5 opacity-30 shrink-0"
            >
              <div class="w-6 h-6 sm:w-8 sm:h-8 rounded-lg border border-dashed border-slate-300 dark:border-white/10 flex items-center justify-center">
                 <Plus class="w-3 h-3 text-slate-400" />
              </div>
              <div class="hidden sm:flex flex-col">
                 <span class="text-[7px] font-black uppercase tracking-tighter text-slate-400">EMPTY_SLOT</span>
              </div>
            </div>
          </template>
          <div v-else class="flex items-center gap-1.5 opacity-30 px-3">
             <Loader2 class="w-3.5 h-3.5 animate-spin" />
             <span class="text-[9px] font-black uppercase tracking-widest italic">Awaiting Peers...</span>
          </div>
        </div>

        <!-- Global Status -->
        <div class="flex items-center gap-2 pl-3 border-l border-slate-200 dark:border-white/10 shrink-0">
          <div v-if="gameState?.status === 'playing'" class="flex items-center gap-1.5 px-2 py-1 bg-blue-500/10 border border-blue-500/20 rounded-lg">
             <Activity class="w-3 h-3 text-blue-500" :class="timeRemaining <= 10 && 'animate-pulse'" />
             <span class="font-mono font-black text-[10px] text-blue-500">{{ timeRemaining }}S</span>
          </div>

          <button v-if="!roomInfo?.is_points_mode" @click="showHints = !showHints" class="w-8 h-8 flex items-center justify-center bg-slate-100 dark:bg-white/5 rounded-lg border border-slate-200 dark:border-white/10 text-slate-500 hover:text-blue-500">
             <Sparkles class="w-3.5 h-3.5" :class="showHints && 'fill-current text-blue-500'" />
          </button>
          <button @click="showLogs = !showLogs" class="w-8 h-8 flex items-center justify-center bg-slate-100 dark:bg-white/5 rounded-lg border border-slate-200 dark:border-white/10 text-slate-500 hover:text-blue-500">
             <Zap class="w-3.5 h-3.5" :class="showLogs && 'fill-current text-blue-500'" />
          </button>
          
          <button @click="showChat = !showChat" class="w-8 h-8 relative flex items-center justify-center bg-slate-100 dark:bg-white/5 rounded-lg border border-slate-200 dark:border-white/10 text-slate-500 hover:text-blue-500">
             <MessageCircle class="w-3.5 h-3.5" :class="showChat && 'fill-current text-blue-500'" />
             <div v-if="hasNewMessage" class="absolute -top-0.5 -right-0.5 w-2.5 h-2.5 bg-rose-500 border-2 border-white dark:border-[#0d0d10] rounded-full animate-pulse"></div>
          </button>
        </div>
      </header>

      <!-- Main Action Focus Area -->
      <div class="flex-1 relative flex flex-col items-center justify-center p-4 mb-20 overflow-hidden">
          <!-- Left Sidebar: Hint & Status -->
          <div :class="cn(
            'absolute left-6 top-6 bottom-6 w-72 z-[60] bg-white/80 dark:bg-black/80 backdrop-blur-3xl border border-slate-200 dark:border-white/10 rounded-[40px] shadow-3xl transition-all duration-700 flex flex-col overflow-hidden',
            showHints ? 'translate-x-0 opacity-100' : 'translate-x-[calc(-100%-3rem)] opacity-0 pointer-events-none'
          )">
             <div class="p-6 border-b border-slate-200 dark:border-white/10 flex items-center justify-between">
                <div class="flex items-center gap-2">
                   <Trophy class="w-4 h-4 text-blue-500" />
                   <span class="text-xs font-black uppercase tracking-widest text-slate-500">实验辅助情报</span>
                </div>
                <button @click="showHints = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-white transition-colors">
                   <ArrowLeft class="w-4 h-4" />
                </button>
             </div>
             
             <div class="flex-1 overflow-y-auto p-5 custom-scrollbar space-y-6">
                <!-- Status Banners -->
                <div class="space-y-3">
                   <div v-if="allowedAny" class="bg-amber-500/10 border border-amber-500/20 p-3 rounded-2xl animate-pulse">
                      <div class="flex items-center gap-2 text-amber-500 mb-1">
                         <Zap class="w-3.5 h-3.5 fill-current" />
                         <span class="text-[10px] font-black uppercase tracking-wider">AU 特权激活</span>
                      </div>
                      <p class="text-[9px] font-bold text-slate-500">已跳过所有反应规则限制</p>
                   </div>

                   <div v-if="gameState?.pending_draw_count > 0" class="bg-red-500/10 border border-red-500/20 p-3 rounded-2xl animate-bounce">
                      <div class="flex items-center gap-2 text-red-500 mb-1">
                         <RefreshCw class="w-3.5 h-3.5 animate-spin-slow" />
                         <span class="text-[10px] font-black uppercase tracking-wider">加牌预演中</span>
                      </div>
                      <p class="text-[9px] font-bold text-slate-500">需结算或叠加累计: {{ gameState.pending_draw_count }}</p>
                   </div>
                </div>

                <!-- Turn Hints -->
                <div v-if="isMyTurn">
                   <div class="flex items-center gap-2 mb-3">
                      <FlaskConical class="w-3.5 h-3.5 text-blue-500" />
                      <span class="text-[10px] font-black uppercase tracking-widest text-slate-500">可用合成路径</span>
                   </div>
                   
                   <div v-if="turnReadySubstances.length > 0" class="space-y-2">
                      <button 
                         v-for="sub in turnReadySubstances" 
                         :key="sub"
                         @click="selectedSubstance = sub; handlePlayCard()"
                         class="w-full text-left px-4 py-3 bg-white/50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl hover:border-blue-500 hover:bg-blue-500/5 transition-all group"
                      >
                         <div class="flex items-center justify-between">
                            <span class="text-xs font-black dark:text-white" v-html="formatFormula(sub)"></span>
                            <div class="w-1.5 h-1.5 rounded-full bg-emerald-500 group-hover:scale-125 transition-transform shadow-[0_0_8px_rgba(16,185,129,0.5)]"></div>
                         </div>
                         <p class="text-[9px] font-bold text-slate-400 mt-1 tracking-tighter">{{ getSubstanceName(sub) }}</p>
                      </button>
                   </div>
                   <div v-else class="py-10 flex flex-col items-center justify-center opacity-30 text-center">
                      <Zap class="w-8 h-8 mb-3" />
                      <p class="text-[10px] font-black uppercase tracking-widest">目前无可用反应</p>
                      <p class="text-[9px] font-bold mt-1">请尝试摸牌补充底物</p>
                   </div>
                </div>
                
                <div v-else-if="roomInfo?.status === 'waiting'" class="space-y-4">
                   <!-- 积分模式提示 -->
                   <div v-if="roomInfo?.is_points_mode" class="p-4 bg-amber-500/10 border border-amber-500/20 rounded-2xl flex items-center gap-3">
                      <Trophy class="w-5 h-5 text-amber-500 shrink-0" />
                      <div class="text-left">
                         <p class="text-[10px] font-black uppercase tracking-widest text-amber-600 dark:text-amber-500">Competitive Mode</p>
                         <p class="text-[9px] font-bold text-slate-500 mt-0.5">积分竞技模式：胜者将获得积分，败者扣除积分。强制使用默认牌组。</p>
                      </div>
                   </div>

                   <div class="p-4 bg-blue-500/5 border border-blue-500/10 rounded-2xl flex flex-col items-center text-center">
                      <Users class="w-6 h-6 text-blue-500 mb-2" />
                      <span class="text-[10px] font-black uppercase tracking-widest text-blue-500">准备就绪?</span>
                      <p class="text-[9px] font-bold text-slate-500 mt-1">当前由于连接数 {{ allPlayers.length }} / {{ roomInfo?.max_players }}，等待就绪的人数达标后，实验室将自动开启。</p>
                   </div>
                   <div class="p-4 bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl">
                      <div class="flex items-center gap-2 mb-2">
                         <QrCode class="w-3.5 h-3.5 text-blue-500" />
                         <span class="text-[10px] font-black uppercase tracking-widest text-slate-500">快速邀请</span>
                      </div>
                      <p class="text-[8px] font-bold text-slate-400 leading-relaxed uppercase">
                         点击中间区域的“招募伙伴”按钮可快速复制链接，或点击二维码图标让好友扫码加入此反应室。
                      </p>
                   </div>
                </div>
                
                <div v-else class="py-10 flex flex-col items-center justify-center opacity-20 text-center">
                   <Timer class="w-8 h-8 mb-3" />
                   <p class="text-[10px] font-black uppercase tracking-widest">等待其他研究员行动</p>
                </div>
             </div>
          </div>

          <!-- Latest Reaction Display -->
          <div v-if="gameState?.last_card" class="relative group scale-90 sm:scale-100 flex flex-col items-center justify-center">
             <div class="absolute -inset-16 bg-blue-600/10 rounded-full blur-[100px] opacity-50 group-hover:opacity-80 transition-opacity animate-pulse"></div>
             
             <!-- Double Play Display (Side by Side) -->
             <div v-if="gameState?.last_card?.reactants?.length > 0" class="flex items-center gap-6 sm:gap-10 relative z-10">
                <div v-for="(sub, idx) in gameState.last_card.reactants" :key="idx" class="relative group/card">
                   <div :class="cn(
                      'w-28 h-40 sm:w-32 h-48 rounded-[32px] border-4 border-white/30 flex flex-col items-center justify-center gap-4 shadow-2xl transition-all hover:scale-105',
                      getDynamicCardClass(gameState?.last_card?.card)
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
               'w-40 h-56 sm:w-48 h-64 rounded-[32px] border-4 border-white/30 flex flex-col items-center justify-center gap-4 sm:gap-6 shadow-2xl transition-all hover:scale-105 relative overflow-hidden',
               getDynamicCardClass(gameState?.last_card?.card)
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
                <div class="w-24 h-24 sm:w-32 sm:h-32 rounded-[32px] sm:rounded-[40px] border-4 border-emerald-500/30 flex items-center justify-center relative z-10">
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
                <div class="w-24 h-24 sm:w-32 sm:h-32 rounded-[32px] sm:rounded-[40px] border-4 border-dashed border-blue-500/30 flex items-center justify-center rotate-45 group hover:rotate-0 transition-all duration-700">
                   <FlaskConical class="w-10 h-10 sm:w-14 sm:h-14 text-blue-500/40 -rotate-45 group-hover:rotate-0 transition-all" />
                </div>
                <div v-if="roomInfo?.countdown > 0" class="absolute -top-3 -right-3 bg-red-500 text-white px-4 py-1.5 rounded-xl text-lg font-black shadow-lg animate-bounce">
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
                  <p v-if="roomInfo?.countdown > 0" class="text-[8px] font-black uppercase tracking-[0.2em] text-blue-500 animate-pulse mt-2">
                    实验室压力充盈中，即将开启研究循环...
                  </p>
                </div>

                <div class="flex flex-col items-center gap-3 bg-white/50 dark:bg-white/5 backdrop-blur-xl p-4 sm:p-5 rounded-[24px] border border-slate-200 dark:border-white/10 shadow-sm w-full max-w-sm">
                  <div class="flex flex-wrap justify-center gap-2 sm:gap-3">
                    <div class="flex items-center gap-2 px-3 py-1.5 bg-slate-100 dark:bg-white/5 rounded-xl border border-slate-200 dark:border-white/10">
                      <Users class="w-3 h-3 text-blue-500" />
                      <span class="text-[8px] font-black uppercase tracking-widest text-slate-600 dark:text-slate-400">
                        研究员: {{ allPlayers.length }} / {{ roomInfo?.max_players }}
                      </span>
                    </div>
                    <div class="flex items-center gap-2 px-3 py-1.5 bg-slate-100 dark:bg-white/5 rounded-xl border border-slate-200 dark:border-white/10">
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
      <div class="fixed bottom-0 left-0 right-0 z-[70] bg-white/60 dark:bg-black/60 backdrop-blur-2xl border-t border-slate-200 dark:border-white/5 flex flex-col items-center">
        <!-- Turn-related buttons and timer -->
        <div class="h-0 relative w-full flex justify-center">
           <div v-if="isMyTurn" class="absolute bottom-full mb-4 flex flex-col items-center gap-3 animate-in slide-in-from-bottom-4">
              <div class="flex items-center bg-white/90 dark:bg-black/80 backdrop-blur-xl border border-slate-200 dark:border-white/10 rounded-xl sm:rounded-2xl p-0.5 shadow-2xl">
                <input 
                  v-model="substanceInput" 
                  @keyup.enter="handleInputPlay"
                  placeholder="手动注入化学式" 
                  class="bg-transparent border-none outline-none text-[10px] sm:text-xs px-3 py-1 w-32 sm:w-48 font-black tracking-widest placeholder:text-slate-400 text-slate-900 dark:text-white"
                />
                
                <div class="flex items-center gap-1">
                   <button 
                      @click="handleInputPlay"
                      class="bg-blue-600 hover:bg-blue-500 w-7 h-7 rounded-lg flex items-center justify-center transition-all active:scale-90 shadow-lg group"
                      title="执行反应"
                   >
                      <ChevronRight class="w-4 h-4 text-white group-hover:translate-x-0.5 transition-transform" />
                   </button>
                   
                   <div class="w-px h-5 bg-slate-200 dark:bg-white/10 mx-0.5"></div>

                   <button 
                      @click="handleDrawCard"
                      :disabled="!isMyTurn"
                      :class="cn(
                        'px-3 h-7 rounded-lg flex items-center justify-center gap-1.5 transition-all active:scale-95 shadow-lg group relative overflow-hidden',
                        isMyTurn ? (gameState?.pending_draw_count > 0 ? 'bg-red-600 hover:bg-red-500 text-white' : 'bg-slate-800 dark:bg-white/10 hover:bg-slate-700 dark:hover:bg-white/20 text-white') : 'bg-slate-200 dark:bg-slate-800 text-slate-400 cursor-not-allowed grayscale'
                      )"
                   >
                      <Plus v-if="!(gameState?.pending_draw_count > 0)" class="w-3 h-3" />
                      <RefreshCw v-else class="w-3 h-3 animate-spin-slow" />
                      <span class="text-[9px] font-black uppercase tracking-widest whitespace-nowrap">
                        摸牌{{ gameState?.pending_draw_count > 0 ? gameState.pending_draw_count : '1' }}张
                      </span>
                   </button>
                </div>
              </div>
              
              <div class="flex items-center gap-2">
                <div class="bg-blue-600/90 backdrop-blur-md px-4 py-1.5 rounded-full border border-white/20 shadow-lg flex items-center gap-2.5 animate-slide-in-bottom">
                  <Zap class="w-3 h-3 fill-current animate-pulse text-white" />
                  <span class="text-[9px] font-black uppercase tracking-widest text-white">科研操作 ({{ timeRemaining }}s)</span>
                  
                  <!-- 双联行动按钮 -->
                  <button 
                    v-if="myData?.double_action_available"
                    @click.stop="toggleDoubleMode"
                    :class="cn(
                      'px-3 py-1 rounded-xl border border-white/20 transition-all flex items-center gap-2 relative overflow-hidden',
                      doubleMode ? 'bg-amber-500 text-white border-amber-400 shadow-md' : 'bg-black/40 text-white/60 hover:text-white hover:bg-black/60'
                    )"
                  >
                     <div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/10 to-transparent -translate-x-full group-hover/btn:animate-shimmer"></div>
                     <Activity :class="cn('w-3.5 h-3.5', doubleMode && 'animate-spin')" />
                     <span class="text-[8px] font-black uppercase tracking-tighter">{{ doubleMode ? '解除超限' : '超限双联' }}</span>
                  </button>
                </div>
              </div>

              <!-- 双联模式提示状态 -->
              <div v-if="doubleMode" class="mt-2 flex items-center gap-4 animate-in slide-in-from-top-4 duration-500">
                <div class="flex items-center gap-3">
                  <div :class="cn('w-10 h-10 rounded-xl flex items-center justify-center border-2 transition-all duration-500', firstDoubleSubstance ? 'bg-blue-500/20 border-blue-500 shadow-lg' : 'bg-slate-800/50 border-white/10 opacity-50')">
                    <span v-if="firstDoubleSubstance" class="text-[10px] font-black" v-html="formatFormula(firstDoubleSubstance)"></span>
                    <FlaskConical v-else class="w-4 h-4 text-slate-500" />
                  </div>
                  <div class="w-4 h-0.5 bg-blue-500/30"></div>
                  <div :class="cn('w-10 h-10 rounded-xl flex items-center justify-center border-2 transition-all duration-500', secondDoubleSubstance ? 'bg-blue-500/20 border-blue-500 shadow-lg' : 'bg-slate-800/50 border-white/10 opacity-50')">
                    <span v-if="secondDoubleSubstance" class="text-[10px] font-black" v-html="formatFormula(secondDoubleSubstance)"></span>
                    <FlaskConical v-else class="w-4 h-4 text-slate-500" />
                  </div>
                </div>

                <button 
                  v-if="firstDoubleSubstance && secondDoubleSubstance"
                  @click="handleDoublePlay"
                  class="bg-emerald-600 hover:bg-emerald-500 text-white px-6 py-2 rounded-2xl flex items-center gap-2 shadow-lg animate-in zoom-in duration-300 group"
                >
                  <span class="text-[10px] font-black uppercase tracking-widest">启动双联反应</span>
                  <Play class="w-3.5 h-3.5 fill-current group-hover:translate-x-0.5 transition-transform" />
                </button>
              </div>
           </div>
        </div>

        <div class="w-full max-w-6xl flex justify-center items-end py-1 sm:py-2">
           <div ref="handContainer" class="flex items-end gap-1 sm:gap-1.5 px-3 sm:px-6 overflow-x-auto custom-scrollbar-hidden py-2 sm:py-4 min-h-[110px] sm:min-h-[150px] w-full max-w-7xl">
            <div v-if="roomInfo?.status === 'waiting'" class="flex flex-col items-center justify-center opacity-30 pb-2 min-w-full">
              <Loader2 class="w-8 h-8 sm:w-12 sm:h-12 mb-1 animate-spin text-blue-500" />
              <p class="font-black uppercase tracking-widest text-[8px] sm:text-xs text-slate-500 text-center">正在同步量子状态并等待开场就绪...</p>
            </div>
            <template v-else-if="myData?.hand_cards?.length > 0">
              <div
                v-for="(card, index) in myData.hand_cards"
                :key="index"
                @click="isMyTurn && handleCardClick(card)"
                :class="cn(
                  'relative w-16 sm:w-24 h-22 sm:h-34 rounded-xl border-4 flex flex-col items-center justify-center cursor-pointer transition-all duration-300 shadow-lg overflow-hidden shrink-0',
                  getDynamicCardClass(card),
                  selectedCard === card && 'selected',
                  !isMyTurn && 'disabled'
                )"
                :style="{
                  transform: selectedCard === card ? (isMobile ? 'translateY(-12px)' : 'translateY(-20px)') : 'none'
                }"
              >
                <div class="absolute top-0.5 sm:top-1.5 left-0.5 sm:left-1.5 text-[5px] sm:text-[7px] font-black opacity-30 uppercase tracking-tighter">{{ ELEMENTS_DATA[card.type] ? 'Elem' : 'Spec' }}</div>
                <div class="flex flex-col items-center justify-center">
                  <div class="text-lg sm:text-xl font-black font-mono italic tracking-tighter leading-none">{{ card.type }}</div>
                  <div v-if="card.effect || ['He','Ne','Ar','Kr'].includes(card.type)" class="mt-0.5 sm:mt-1.5 px-1 sm:px-1.5 py-0.5 bg-black/10 rounded-lg text-[8px] sm:text-[10px] font-black uppercase tracking-tighter">
                    {{ ['He','Ne','Ar','Kr'].includes(card.type) ? '转向' : card.effect === 'Au' ? '跳过' : card.effect === '+2' ? '+2' : card.effect === '+4' ? '+4' : card.effect }}
                  </div>
                  <div v-else-if="ELEMENTS_DATA[card.type]" class="text-[7px] sm:text-[9px] font-bold opacity-80 mt-0.5 sm:mt-1 uppercase tracking-tighter font-serif italic text-black/40">
                    {{ ELEMENTS_DATA[card.type].name }}
                  </div>
                </div>
                <div class="absolute bottom-0.5 sm:bottom-1.5 right-0.5 sm:right-1.5 text-[5px] sm:text-[6px] font-mono opacity-40 uppercase tracking-tighter">
                  {{ card.effect ? 'Func' : 'Pass' }}
                </div>
              </div>
            </template>
            <div v-else class="flex flex-col items-center justify-center opacity-10 pb-4 sm:pb-10">
              <FlaskConical class="w-10 h-10 sm:w-16 sm:h-16 mb-2" />
              <p class="font-black uppercase tracking-widest text-[8px] sm:text-xs">Inventory_Empty</p>
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
                  <div :class="cn('relative w-px h-px flex items-center justify-center scale-110 !cursor-default', getDynamicCardClass(selectedCard))">
                     <!-- Dummy container to hold getDynamicCardClass utility styles -->
                     <div class="w-16 sm:w-24 h-22 sm:h-34 rounded-xl border-4 flex flex-col items-center justify-center">
                        <div class="text-xl sm:text-2xl font-black tracking-tighter">{{ selectedCard.type }}</div>
                     </div>
                  </div>
               </div>
             </div>

             <div class="grid grid-cols-3 gap-2 sm:gap-3 mb-6 sm:mb-8 max-h-[160px] sm:max-h-[240px] overflow-y-auto pr-2 custom-scrollbar">
                <button
                  v-for="(substance, index) in availableSubstances"
                  :key="index"
                  @click="selectedSubstance = substance"
                  :class="cn(
                    'relative p-2.5 sm:p-4 rounded-xl sm:rounded-2xl border transition-all flex flex-col items-center justify-center gap-1.5 sm:gap-2 overflow-hidden',
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
            <div class="grid grid-cols-3 gap-2">
              <button 
                v-for="h in [1, 24, 72, 168, 720, -1]" 
                :key="h"
                @click="banHours = h"
                :class="cn(
                  'py-2.5 rounded-xl text-[10px] font-bold border transition-all',
                  banHours === h ? 'bg-red-500 text-white border-red-500' : 'bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/10 text-slate-500'
                )"
              >
                {{ h === -1 ? '永久' : (h < 24 ? h + 'h' : Math.floor(h/24) + 'd') }}
              </button>
            </div>
          </div>

          <div class="space-y-4">
            <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest block">操作事由</label>
            <div class="relative group">
              <div class="absolute inset-0 bg-red-500/5 rounded-2xl blur-lg group-focus-within:bg-red-500/10 transition-all"></div>
              <textarea 
                v-model="banReason"
                placeholder="请输入详细的违规事由..."
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

    <!-- Chat Sidebar/Modal -->
    <div 
      v-if="showChat"
      class="fixed inset-y-0 right-0 w-full sm:w-96 bg-white dark:bg-[#0d0d10]/95 backdrop-blur-2xl border-l border-slate-200 dark:border-white/5 z-[90] shadow-3xl animate-slide-in-right"
    >
      <ChatBox title="实验内通信线程" />
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
