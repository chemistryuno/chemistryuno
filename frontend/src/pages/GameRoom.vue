<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { gameAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import websocket from '../utils/websocket'
import { ArrowLeft, Play, RefreshCw, Zap, Activity, FlaskConical, Trophy, ChevronRight, Loader2, Users, Timer, Plus, QrCode, Copy, ExternalLink, Sparkles } from 'lucide-vue-next'
import { cn } from '../utils/cn'

const route = useRoute()
const router = useRouter()
const { showAlert, showConfirm } = useDialog()
const id = route.params.id as string

const user = ref(JSON.parse(localStorage.getItem('user') || '{}'))
const gameState = ref<any>(null)
const roomInfo = ref<any>(null)
const playersInfo = ref<any[]>([])
const availableSubstances = ref<string[]>([])
const turnReadySubstances = ref<string[]>([])
const selectedCard = ref<any>(null)
const selectedSubstance = ref<string | null>(null)
const doubleMode = ref(false)
const firstDoubleSubstance = ref<string | null>(null)
const secondDoubleSubstance = ref<string | null>(null)
const substanceInput = ref('')
const loading = ref(true)
const isRedirecting = ref(false)
const showQrModal = ref(false)
const timeRemaining = ref(30)
let timerInterval: any = null

const allPlayers = computed(() => {
  if (gameState.value?.players) {
    return gameState.value.players.map((p: any) => {
      const baseInfo = playersInfo.value.find(b => Number(b.uid) === Number(p.uid))
      return {
        ...p,
        avatar: p.avatar || baseInfo?.avatar,
        username: p.username || baseInfo?.username,
        is_host: Number(p.uid) === Number(roomInfo.value?.host_uid)
      }
    })
  }
  return playersInfo.value
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

const ELEMENTS_DATA: Record<string, { name: string, color: string }> = {
  'H': { name: '氢', color: 'bg-blue-100 dark:bg-blue-900/30 text-blue-600 border-blue-200' },
  'O': { name: '氧', color: 'bg-red-100 dark:bg-red-900/30 text-red-600 border-red-200' },
  'C': { name: '碳', color: 'bg-slate-700 text-white border-slate-800' },
  'N': { name: '氮', color: 'bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 border-indigo-200' },
  'S': { name: '硫', color: 'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-600 border-yellow-200' },
  'Cl': { name: '氯', color: 'bg-emerald-100 dark:bg-emerald-900/30 text-emerald-600 border-emerald-200' },
  'Na': { name: '钠', color: 'bg-orange-100 dark:bg-orange-900/30 text-orange-600 border-orange-200' },
  'Mg': { name: '镁', color: 'bg-cyan-100 dark:bg-cyan-900/30 text-cyan-600 border-cyan-200' },
  'Al': { name: '铝', color: 'bg-zinc-100 dark:bg-zinc-900/30 text-zinc-600 border-zinc-200' },
  'Cu': { name: '铜', color: 'bg-orange-200 dark:bg-orange-950/30 text-orange-700 border-orange-300' },
  'Fe': { name: '铁', color: 'bg-stone-200 dark:bg-stone-900/30 text-stone-600 border-stone-300' },
  'Zn': { name: '锌', color: 'bg-teal-100 dark:bg-teal-900/30 text-teal-600 border-teal-200' },
  'Ag': { name: '银', color: 'bg-slate-100 dark:bg-slate-800 text-slate-500 border-slate-200' },
  'K': { name: '钾', color: 'bg-purple-100 dark:bg-purple-900/30 text-purple-600 border-purple-200' },
  'Ca': { name: '钙', color: 'bg-amber-100 dark:bg-amber-900/30 text-amber-600 border-amber-200' },
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
const level = computed(() => Math.floor(exp.value / 100) + 1)
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
  if (message.data) {
    gameState.value = message.data
    if (isMyTurn.value) {
      fetchTurnSubstances()
    }
  } else {
    loadGameState().then(() => {
      if (isMyTurn.value) {
        fetchTurnSubstances()
      }
    })
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
      host_uid: data.host_uid,
      players: data.players,
      max_players: data.max_players,
      status: data.status,
      is_points_mode: data.is_points_mode
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
  loadGameState().then(() => {
    websocket.joinRoom(id)
    websocket.on('game_update', handleGameUpdate)
    websocket.on('player_joined', loadGameState)
    websocket.on('player_left', loadGameState)
    websocket.on('room_terminated', async (msg: any) => {
      isRedirecting.value = true
      const reason = msg.message || '房主已中断连接，实验室已关闭'
      await showAlert(reason, '实验结束')
      router.push('/')
    })
    websocket.on('player_kicked', async (msg: any) => {
      isRedirecting.value = true
      await showAlert(msg.message || '由于消极游戏，您已被踢出', '权限移除')
      router.push('/')
    })
  })
})

onUnmounted(() => {
  if (timerInterval) clearInterval(timerInterval)
  websocket.leaveRoom()
  websocket.off('game_update', handleGameUpdate)
  websocket.off('player_joined', loadGameState)
  websocket.off('player_left', loadGameState)
  websocket.off('room_terminated', () => {})
})

const handleStartGame = async () => {
  try {
    await gameAPI.startGame(id)
    await loadGameState()
  } catch (error: any) {
    showAlert(error.response?.data?.error || '开始游戏失败', '启动失败')
  }
}

const handleCardClick = async (card: any) => {
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
    const canPlay = (response.data || []).includes(card.type)
    if (canPlay) {
      await gameAPI.playCard(id, card, card.type)
      selectedCard.value = null
      selectedSubstance.value = null
      availableSubstances.value = []
    } else {
      showAlert('该元素当前无法与上一张牌反应，不能打出', '出牌失败')
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
    const sub = substanceInput.value.toUpperCase()
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
  if (ELEMENTS_DATA[card.type]) return ELEMENTS_DATA[card.type].color
  const style = getCardStyle(card)
  if (style === 'noble') return 'bg-indigo-600 border-indigo-400'
  if (style === 'gold') return 'bg-amber-500 border-amber-300'
  if (style === 'special') return 'bg-rose-600 border-rose-400'
  return ''
}

const playerPositions = [
  'bottom-2 sm:bottom-[-20px] translate-y-0 sm:translate-y-1/2 translate-x-1/2 right-1/2 scale-100 sm:scale-105',
  'left-2 sm:left-[-30px] -translate-x-0 sm:-translate-x-1/4 top-1/2 -translate-y-1/2 scale-90 sm:scale-100',
  'top-2 sm:top-[-20px] -translate-y-0 sm:-translate-y-1/2 translate-x-1/2 right-1/2 scale-90 sm:scale-100',
  'right-2 sm:right-[-30px] translate-x-0 sm:translate-x-1/4 top-1/2 -translate-y-1/2 scale-90 sm:scale-100',
]

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
          <h2 class="text-2xl font-black text-white tracking-widest uppercase">Initializing Lab</h2>
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
      <header class="h-[64px] sm:h-[80px] bg-white/70 dark:bg-black/60 backdrop-blur-3xl border-b border-slate-200 dark:border-white/5 px-4 sm:px-6 flex items-center gap-4 z-50 sticky top-0 overflow-x-auto custom-scrollbar-hidden">
        <div class="flex items-center gap-3 pr-4 border-r border-slate-200 dark:border-white/10 shrink-0">
          <button 
            @click="handleLeaveRoom" 
            class="w-8 h-8 sm:w-10 sm:h-10 flex items-center justify-center hover:bg-white/10 rounded-2xl text-slate-500 hover:text-blue-500 transition-all"
          >
            <ArrowLeft class="w-4 h-4 sm:w-5 sm:h-5" />
          </button>
          <div class="hidden xs:block">
            <h2 class="text-[10px] font-black tracking-widest uppercase font-mono text-slate-400">Node: {{ id.substring(0, 6) }}</h2>
            <div class="flex items-center gap-1.5">
               <div :class="cn('w-1.5 h-1.5 rounded-full animate-pulse', roomInfo?.status === 'waiting' ? 'bg-amber-500' : 'bg-emerald-500')"></div>
               <span class="text-[8px] font-black uppercase text-slate-500 tracking-tighter">{{ roomInfo?.status === 'waiting' ? 'Idle' : 'Active' }}</span>
            </div>
          </div>
        </div>

        <!-- Players Horizontal Bar -->
        <div class="flex flex-1 items-center gap-2 sm:gap-4 px-2 overflow-x-auto custom-scrollbar-hidden py-1">
          <template v-if="allPlayers.length > 0">
            <div 
              v-for="(player, index) in allPlayers"
              :key="player.uid"
              :class="cn(
                'flex items-center gap-2 sm:gap-3 px-3 py-1.5 rounded-2xl border transition-all shrink-0',
                gameState?.current_player === index 
                  ? 'bg-blue-600 shadow-lg shadow-blue-500/20 ring-1 ring-blue-500/20 border-blue-500' 
                  : (gameState ? 'bg-white/5 border-slate-200 dark:border-white/5 opacity-60' : 'bg-white/5 border-white/10')
              )"
            >
              <div class="relative w-7 h-7 sm:w-9 sm:h-9 shrink-0">
                <div :class="cn(
                  'w-full h-full rounded-lg flex items-center justify-center text-sm border overflow-hidden',
                   gameState?.current_player === index ? 'bg-white text-blue-600 border-white/20' : 'bg-slate-100 dark:bg-slate-800 border-slate-200 dark:border-white/10'
                )">
                   <img v-if="player.avatar && player.avatar.startsWith('data:')" :src="player.avatar" class="w-full h-full object-cover" />
                   <span v-else>{{ player.avatar || '🧪' }}</span>
                </div>
                <!-- Action Progress Dots (Only during gameplay) -->
                <div v-if="gameState" class="absolute -bottom-1 -right-1 flex gap-0.5">
                  <div v-for="i in 2" :key="i" :class="cn('w-1.5 h-1.5 rounded-full border border-black/20', i <= (player.action_progress || 0) ? (gameState?.current_player === index ? 'bg-white' : 'bg-blue-500') : 'bg-slate-500')"></div>
                </div>
              </div>
              <div class="flex flex-col min-w-0">
                <div class="flex items-center gap-1.5">
                  <span class="text-[10px] font-bold truncate max-w-[60px] tracking-tight" :class="gameState?.current_player === index ? 'text-white' : 'text-slate-500'">{{ player.username }}</span>
                  <Zap v-if="player.double_action_available" :class="cn('w-2.5 h-2.5 fill-current', gameState?.current_player === index ? 'text-amber-300' : 'text-amber-500')" />
                  <span v-if="player.is_host" :class="cn('w-2 h-2 rounded-full ring-2', gameState?.current_player === index ? 'bg-amber-300 ring-amber-300/20' : 'bg-amber-500 ring-amber-500/20')" title="房主"></span>
                </div>
                <!-- Status/Card Count -->
                <div class="flex items-center gap-1">
                  <template v-if="gameState">
                    <Trophy :class="cn('w-2 h-2', gameState?.current_player === index ? 'text-white' : 'text-slate-400')" />
                    <span :class="cn('text-[8px] font-mono font-bold', gameState?.current_player === index ? 'text-white/80' : 'text-slate-400')">{{ player.card_count || 0 }}</span>
                  </template>
                  <template v-else>
                    <span class="text-[8px] font-black uppercase text-slate-400 tracking-widest">{{ player.is_host ? 'Host' : 'Guest' }}</span>
                  </template>
                </div>
              </div>
            </div>

            <!-- Empty Slots in Top Bar -->
            <div 
              v-for="i in (roomInfo?.max_players || 0) - allPlayers.length" 
              :key="'empty-top-' + i"
              class="flex items-center gap-2 px-3 py-1.5 rounded-2xl border border-dashed border-slate-200 dark:border-white/5 opacity-30 shrink-0"
            >
              <div class="w-7 h-7 sm:w-9 sm:h-9 rounded-lg border border-dashed border-slate-300 dark:border-white/10 flex items-center justify-center">
                 <Plus class="w-3 h-3 text-slate-400" />
              </div>
              <div class="hidden sm:flex flex-col">
                 <span class="text-[8px] font-black uppercase tracking-tighter text-slate-400">EMPTY_SLOT</span>
              </div>
            </div>
          </template>
          <div v-else class="flex items-center gap-2 opacity-30 px-4">
             <Loader2 class="w-4 h-4 animate-spin" />
             <span class="text-[10px] font-black uppercase tracking-widest italic">Awaiting Peers...</span>
          </div>
        </div>

        <!-- Global Status -->
        <div class="flex items-center gap-3 pl-4 border-l border-slate-200 dark:border-white/10 shrink-0">
          <div v-if="gameState?.status === 'playing'" class="flex items-center gap-2 px-3 py-1.5 bg-blue-500/10 border border-blue-500/20 rounded-xl">
             <Timer class="w-3 h-3 text-blue-500" :class="timeRemaining <= 10 && 'animate-pulse'" />
             <span class="font-mono font-black text-xs text-blue-500">{{ timeRemaining }}S</span>
          </div>
          <button 
            v-if="roomInfo?.status === 'waiting' && user.uid === roomInfo?.host_uid"
            @click="handleStartGame" 
            class="bg-blue-600 hover:bg-blue-500 px-4 h-9 rounded-xl font-black text-[10px] uppercase text-white shadow-xl flex items-center gap-2"
          >
            <Play class="w-3 h-3 fill-current" />
            <span>启动</span>
          </button>

          <button @click="showHints = !showHints" class="w-9 h-9 flex items-center justify-center bg-slate-100 dark:bg-white/5 rounded-xl border border-slate-200 dark:border-white/10 text-slate-500 hover:text-blue-500">
             <Sparkles class="w-4 h-4" :class="showHints && 'fill-current text-blue-500'" />
          </button>
          <button @click="showLogs = !showLogs" class="w-9 h-9 flex items-center justify-center bg-slate-100 dark:bg-white/5 rounded-xl border border-slate-200 dark:border-white/10 text-slate-500 hover:text-blue-500">
             <Zap class="w-4 h-4" :class="showLogs && 'fill-current text-blue-500'" />
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
                         <p class="text-[9px] font-bold text-slate-400 mt-1 uppercase tracking-tighter">{{ getSubstanceName(sub) }}</p>
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
                      <p class="text-[9px] font-bold text-slate-500 mt-1">当前由于连接数 {{ allPlayers.length }} / {{ roomInfo?.max_players }}，等待人数达标后，房主可通过顶部“启动”按钮开启实验。</p>
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
          <div v-if="gameState?.last_card" class="relative group mb-8">
             <div class="absolute -inset-16 bg-blue-600/10 rounded-full blur-[100px] opacity-50 group-hover:opacity-80 transition-opacity animate-pulse"></div>
             
             <!-- Double Play Display (Side by Side) -->
             <div v-if="gameState?.last_card?.reactants?.length > 0" class="flex items-center gap-6 sm:gap-10 relative z-10 scale-95 sm:scale-105">
                <div v-for="(sub, idx) in gameState.last_card.reactants" :key="idx" class="relative group/card">
                   <div :class="cn(
                      'w-28 h-40 sm:w-36 sm:h-52 rounded-[32px] border-2 shadow-2xl flex flex-col items-center justify-center gap-4 text-white transition-all',
                      getCardStyle(gameState?.last_card?.card),
                      getDynamicCardClass(gameState?.last_card?.card)
                   )">
                      <span class="text-[28px] sm:text-[36px] font-black font-mono italic drop-shadow-lg" v-html="formatFormula(sub)"></span>
                      <div class="px-3 py-1 bg-white/10 backdrop-blur-md rounded-lg border border-white/20 max-w-[85%]">
                         <span class="text-[8px] font-black tracking-widest uppercase truncate block text-center">{{ getSubstanceName(sub) }}</span>
                      </div>
                   </div>
                </div>
                <!-- Plus Operator -->
                <div class="flex items-center justify-center w-10 h-10 rounded-full bg-blue-600 text-white border-4 border-[#0d0d10] shadow-[0_0_20px_rgba(37,99,235,0.4)] z-20">
                   <Plus class="w-5 h-5 stroke-[4px]" />
                </div>
             </div>

             <!-- Single Play Display -->
             <div v-else :class="cn(
               'w-32 h-48 sm:w-40 sm:h-60 rounded-[40px] border-2 shadow-2xl flex flex-col items-center justify-center gap-5 text-white transition-all relative z-10',
               getCardStyle(gameState?.last_card?.card),
               getDynamicCardClass(gameState?.last_card?.card)
             )">
                <div class="absolute top-5 left-5 opacity-20 text-[10px] uppercase font-black tracking-widest">Reaction Result</div>
                <span class="text-[40px] sm:text-[52px] font-black font-mono italic drop-shadow-lg" v-html="formatFormula(gameState?.last_card?.substance)"></span>
                <div class="px-6 py-2 bg-white/10 backdrop-blur-md rounded-2xl border border-white/20 max-w-[85%]">
                   <span class="text-[10px] sm:text-xs font-black tracking-widest uppercase text-center block leading-tight">{{ getSubstanceName(gameState?.last_card?.substance) }}</span>
                </div>
                <div class="absolute bottom-5 right-5 opacity-30">
                   <FlaskConical class="w-5 h-5 fill-current" />
                </div>
             </div>

             <!-- Stability & Info Label Removed as it's now integrated or redundant -->
             
             <!-- Direction Ring -->
             <div class="absolute -inset-12 pointer-events-none">
                <div :class="cn(
                   'w-full h-full border-2 border-blue-500/10 rounded-full',
                   gameState?.direction === 1 ? 'animate-spin-slow' : 'animate-reverse-spin-slow'
                )" style="border-style: double;"></div>
             </div>
          </div>
          
          <div v-else class="flex flex-col items-center gap-10 sm:gap-14 animate-in fade-in zoom-in duration-1000">
             <div class="relative">
                <div class="absolute inset-0 bg-blue-500/10 rounded-full blur-[80px] animate-pulse"></div>
                <div class="w-32 h-32 sm:w-40 sm:h-40 rounded-[48px] sm:rounded-[64px] border-4 border-dashed border-blue-500/30 flex items-center justify-center rotate-45 group hover:rotate-0 transition-all duration-700">
                   <FlaskConical class="w-12 h-12 sm:w-16 sm:h-16 text-blue-500/40 -rotate-45 group-hover:rotate-0 transition-all" />
                </div>
                <div class="absolute -top-4 -right-4 bg-amber-500 text-white px-4 py-1.5 rounded-xl text-[10px] font-black uppercase tracking-widest shadow-xl animate-bounce">
                   Ready Check
                </div>
             </div>

             <div class="flex flex-col items-center gap-4">
                <h3 class="text-2xl sm:text-3xl font-black text-slate-800 dark:text-white uppercase tracking-[0.2em]">{{ roomInfo?.name || '实验室准备中' }}</h3>
                <div class="flex items-center gap-3">
                   <div class="flex items-center gap-2 px-4 py-2 bg-slate-100 dark:bg-white/5 rounded-2xl border border-slate-200 dark:border-white/10 shadow-sm">
                      <Users class="w-4 h-4 text-blue-500" />
                      <span class="text-xs font-bold text-slate-500 uppercase tracking-widest">{{ allPlayers.length }} / {{ roomInfo?.max_players }} 研究员已就位</span>
                   </div>
                   <!-- 分享链接按钮 -->
                   <button 
                      @click="handleCopyLink"
                      class="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-2xl shadow-lg shadow-blue-500/20 transition-all active:scale-95 group"
                   >
                      <Copy class="w-4 h-4 group-hover:rotate-12 transition-transform" />
                      <span class="text-[10px] font-black uppercase tracking-widest italic">招募伙伴</span>
                   </button>
                   <!-- QR Code 按钮 -->
                   <button 
                      @click="showQrModal = !showQrModal"
                      class="w-10 h-10 flex items-center justify-center bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl text-slate-500 hover:text-blue-500 shadow-sm transition-all active:scale-90"
                      title="显示二维码"
                   >
                      <QrCode class="w-5 h-5" />
                   </button>
                </div>

                <!-- QR Code 浮窗（仅在展开时显示） -->
                <div v-if="showQrModal" class="mt-4 p-4 bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[32px] shadow-2xl animate-in zoom-in duration-300 flex flex-col items-center gap-4">
                   <div class="p-3 bg-white rounded-2xl border-4 border-blue-500/20 shadow-inner">
                      <img 
                        :src="`https://api.qrserver.com/v1/create-qr-code/?size=160x160&data=${encodeURIComponent(shareLink)}`" 
                        alt="Join QR Code"
                        class="w-40 h-40"
                      />
                   </div>
                   <div class="text-center">
                      <p class="text-[10px] font-black uppercase tracking-widest text-blue-500">实验室准入码</p>
                      <p class="text-[8px] font-bold text-slate-400 mt-1 uppercase italic">请扫码进入该科研区域</p>
                   </div>
                </div>
             </div>

             <!-- Player Grid Removed from center to rely on Top Bar -->
          </div>
          
          <div :class="cn(
            'absolute right-6 top-6 bottom-6 w-72 z-[60] bg-white/80 dark:bg-black/80 backdrop-blur-3xl border border-slate-200 dark:border-white/10 rounded-[40px] shadow-3xl transition-all duration-700 flex flex-col overflow-hidden',
            showLogs ? 'translate-x-0 opacity-100' : 'translate-x-[calc(100%+3rem)] opacity-0 pointer-events-none'
          )">
             <div class="p-6 border-b border-slate-200 dark:border-white/10 flex items-center justify-between">
                <div class="flex items-center gap-2">
                   <Activity class="w-4 h-4 text-blue-500" />
                   <span class="text-xs font-black uppercase tracking-widest text-slate-500">Reaction Logs</span>
                </div>
                <button @click="showLogs = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-white transition-colors">
                   <ArrowLeft class="w-4 h-4 rotate-180" />
                </button>
             </div>
             <div class="flex-1 overflow-y-auto p-4 custom-scrollbar">
                <div v-if="!gameState?.discard_pile?.length" class="h-full flex flex-col items-center justify-center opacity-20 gap-3">
                   <FlaskConical class="w-8 h-8" />
                   <p class="text-[10px] font-black uppercase tracking-widest">No Records</p>
                </div>
                <div v-for="(play, idx) in [...(gameState?.discard_pile || [])].reverse()" :key="idx" class="mb-4 last:mb-0 group animate-in slide-in-from-right-4 duration-300">
                   <div class="flex items-center gap-3 mb-1">
                      <div class="w-1.5 h-1.5 rounded-full bg-blue-500 group-first:animate-pulse"></div>
                      <span class="text-[9px] font-black text-slate-400 uppercase tracking-tighter">Turn #{{ (gameState?.discard_pile?.length || 0) - idx }}</span>
                   </div>
                   <div class="bg-slate-50 dark:bg-white/5 p-3 rounded-2xl border border-slate-200 dark:border-white/5">
                      <p class="text-[11px] font-black text-blue-600 dark:text-blue-400 mb-1" v-html="formatFormula(play.substance)"></p>
                      <div class="flex items-center justify-between text-[8px] font-bold text-slate-500 uppercase tracking-widest">
                         <span>{{ getSubstanceName(play.substance) }}</span>
                         <span class="opacity-40">@{{ gameState?.players?.find((p: any) => p.uid === play.player_uid)?.username || 'User' }}</span>
                      </div>
                   </div>
                </div>
             </div>
          </div>

          <!-- Log Toggle Button Removed from here -->
          
        <!-- Table Console Background Removed or Simplified -->
        <div class="absolute inset-0 pointer-events-none overflow-hidden">
           <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-full h-full bg-blue-500/[0.02] dark:bg-blue-500/[0.05] rounded-full blur-[120px]"></div>
        </div>
      </div>

      <!-- Hand / Deck Area -->
      <div class="h-auto bg-gradient-to-t from-blue-900/10 dark:from-blue-900/20 to-transparent relative mt-auto px-4 sm:px-12 flex flex-col items-center pb-2 sm:pb-6">
        <!-- Turn Ready Substances Removed: Now in Sidebar -->

        <!-- Turn Tip -->
        <div class="h-0 relative w-full flex justify-center">
           <div v-if="isMyTurn" class="absolute -top-4 sm:-top-5 translate-y-[-100%] flex flex-col items-center gap-2 animate-in fade-in slide-in-from-bottom-4">
              <div class="flex items-center gap-2 bg-white/95 dark:bg-slate-900/90 backdrop-blur-2xl border border-slate-200 dark:border-white/10 p-1.5 rounded-[20px] shadow-[0_10px_30px_rgba(0,0,0,0.1)] mb-2">
                <input 
                  v-model="substanceInput" 
                  @keyup.enter="handleInputPlay"
                  placeholder="手动注入化学式 (如 H2O)" 
                  class="bg-transparent border-none outline-none text-xs sm:text-sm px-4 py-1.5 w-40 sm:w-60 font-black tracking-widest uppercase placeholder:text-slate-400 text-slate-900 dark:text-white"
                />
                
                <div class="flex items-center gap-1.5">
                   <button 
                      @click="handleInputPlay"
                      class="bg-blue-600 hover:bg-blue-500 w-9 h-9 rounded-xl flex items-center justify-center transition-all active:scale-90 shadow-lg group"
                      title="执行反应"
                   >
                      <ChevronRight class="w-5 h-5 text-white group-hover:translate-x-0.5 transition-transform" />
                   </button>
                   
                   <div class="w-px h-6 bg-slate-200 dark:bg-white/10 mx-0.5"></div>

                   <button 
                      @click="handleDrawCard"
                      :disabled="!isMyTurn"
                      :class="cn(
                        'px-4 h-9 rounded-xl flex items-center justify-center gap-2 transition-all active:scale-95 shadow-lg group relative overflow-hidden',
                        isMyTurn ? (gameState?.pending_draw_count > 0 ? 'bg-red-600 hover:bg-red-500 text-white' : 'bg-slate-800 dark:bg-white/10 hover:bg-slate-700 dark:hover:bg-white/20 text-white') : 'bg-slate-200 dark:bg-slate-800 text-slate-400 cursor-not-allowed grayscale'
                      )"
                   >
                      <Plus v-if="!(gameState?.pending_draw_count > 0)" class="w-3.5 h-3.5" />
                      <RefreshCw v-else class="w-3.5 h-3.5 animate-spin-slow" />
                      <span class="text-[10px] font-black uppercase tracking-widest whitespace-nowrap">
                        摸牌{{ gameState?.pending_draw_count > 0 ? gameState.pending_draw_count : '1' }}张
                      </span>
                   </button>
                </div>
              </div>
              
              <div class="flex items-center gap-4">
                <div class="bg-blue-600 px-6 sm:px-10 py-2 sm:py-3 rounded-[32px] shadow-[0_15px_30px_rgba(37,99,235,0.3)] flex items-center gap-3 active:scale-95 transition-transform relative group">
                  <Zap class="w-4 h-4 fill-current animate-pulse text-white" />
                  <span class="text-[10px] sm:text-xs font-black uppercase tracking-widest sm:tracking-[0.4em] text-white">科研人员操作中 ({{ timeRemaining }}s)</span>
                  
                  <!-- 双联行动按钮 -->
                  <button 
                    v-if="myData?.double_action_available"
                    @click.stop="toggleDoubleMode"
                    :class="cn(
                      'absolute -right-4 top-1/2 -translate-y-1/2 translate-x-full ml-4 px-5 py-2.5 rounded-[24px] border-2 transition-all flex items-center gap-2 whitespace-nowrap overflow-hidden group/btn shadow-xl',
                      doubleMode 
                        ? 'bg-amber-500 border-amber-400 text-white shadow-[0_0_30px_rgba(245,158,11,0.5)]' 
                        : 'bg-white/10 backdrop-blur-md border-white/20 text-white hover:bg-white/20'
                    )"
                  >
                     <div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/10 to-transparent -translate-x-full group-hover/btn:animate-shimmer"></div>
                     <Activity :class="cn('w-5 h-5', doubleMode && 'animate-spin')" />
                     <div class="flex flex-col items-start leading-none">
                        <span class="text-[10px] font-black uppercase tracking-tighter">{{ doubleMode ? '解除超限' : '超限双联反应' }}</span>
                        <span class="text-[7px] font-bold opacity-70 uppercase tracking-tighter">Overload Mode</span>
                     </div>
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

        <div class="w-full max-w-6xl flex justify-center items-end py-2 sm:py-4">
           <div ref="handContainer" class="flex flex-nowrap justify-start sm:justify-center gap-x-2 sm:gap-x-4 px-6 sm:px-12 h-[130px] sm:h-[180px] w-full overflow-x-auto custom-scrollbar py-1 transition-all duration-500 cursor-grab select-none">
            <div v-if="roomInfo?.status === 'waiting'" class="flex flex-col items-center justify-center opacity-30 pb-4 min-w-full">
              <Loader2 class="w-10 h-10 sm:w-16 sm:h-16 mb-2 animate-spin text-blue-500" />
              <p class="font-black uppercase tracking-widest text-[10px] sm:text-sm text-slate-500">等待房主启动反应堆</p>
            </div>
            <template v-else-if="myData?.hand_cards?.length > 0">
              <div
                v-for="(card, index) in myData.hand_cards"
                :key="index"
                @click="isMyTurn && handleCardClick(card)"
                :class="cn(
                  'game-card flex-shrink-0 cursor-pointer transition-all duration-500 transform-gpu origin-bottom text-white',
                  getCardStyle(card),
                  getDynamicCardClass(card),
                  selectedCard === card ? 'selected -translate-y-6 sm:-translate-y-10 scale-110 shadow-[0_30px_60px_rgba(0,0,0,0.6)] z-50 ring-4 ring-blue-500/30' : 'hover:-translate-y-6 hover:z-40',
                  !isMyTurn && 'opacity-40 grayscale-[0.8] cursor-not-allowed pointer-events-none translate-y-8 sm:translate-y-12'
                )"
                :style="{
                  transform: selectedCard === card ? (isMobile ? 'translateY(-16px)' : 'translateY(-24px)') : 'none'
                }"
              >
                <div class="absolute top-1 sm:top-2 left-1 sm:left-2 text-[6px] sm:text-[8px] font-black uppercase opacity-30 tracking-widest">{{ ELEMENTS_DATA[card.type] ? 'Element' : 'Spec' }}</div>
                <div class="flex flex-col items-center justify-center">
                  <div class="text-xl sm:text-2xl font-black font-mono italic tracking-tighter">{{ card.type }}</div>
                  <div v-if="card.effect || ['He','Ne','Ar','Kr'].includes(card.type)" class="text-[7px] sm:text-[9px] font-bold bg-white/20 px-1.5 py-0.5 rounded-full mt-1 uppercase tracking-tighter">
                    {{ ['He','Ne','Ar','Kr'].includes(card.type) ? '转向' : card.effect === 'Au' ? '跳过' : card.effect === '+2' ? '+2' : card.effect === '+4' ? '+4' : card.effect }}
                  </div>
                  <div v-else-if="ELEMENTS_DATA[card.type]" class="text-[8px] sm:text-[10px] font-bold opacity-80 mt-1 uppercase tracking-tighter font-serif italic text-black/40">
                    {{ ELEMENTS_DATA[card.type].name }}
                  </div>
                </div>
                <div class="absolute bottom-1 sm:bottom-2 right-1 sm:right-2 text-[5px] sm:text-[7px] font-mono opacity-40 uppercase tracking-tighter">
                  {{ card.effect ? 'Function' : 'Passive' }}
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
        <div class="relative w-full max-w-2xl max-h-[90vh] bg-white dark:bg-[#0d0d10] border border-slate-200 dark:border-white/10 rounded-[32px] sm:rounded-[48px] shadow-[0_0_100px_rgba(0,0,0,0.1)] dark:shadow-[0_0_100px_rgba(0,0,0,0.8)] overflow-y-auto animate-in fade-in zoom-in slide-in-from-bottom-12 duration-500">
           <!-- Modal Header Decor -->
           <div class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-blue-600 via-purple-600 to-blue-600 opacity-50"></div>
           
           <div class="p-6 sm:p-12">
             <div class="flex flex-col md:flex-row justify-between items-start gap-6 sm:gap-10 mb-8 sm:mb-12">
               <div class="space-y-2 sm:space-y-4">
                  <div class="inline-flex items-center gap-2 px-2 sm:px-3 py-1 bg-blue-500/10 border border-blue-500/20 rounded-full">
                     <Zap class="w-3 h-3 text-blue-600 dark:text-blue-400" />
                     <span class="text-[8px] sm:text-[10px] font-bold text-blue-600 dark:text-blue-400 uppercase tracking-widest">Protocol_Active</span>
                  </div>
                  <h3 class="text-2xl sm:text-4xl font-black text-slate-900 dark:text-white tracking-tighter leading-none">
                    化学物质重组
                  </h3>
                  <p class="text-xs sm:text-sm text-slate-500 dark:text-slate-400 max-w-sm font-medium leading-relaxed">
                    请选择一个与 <span class="text-slate-900 dark:text-white font-black underline decoration-blue-500 underline-offset-4">{{ selectedCard.type }}</span> 兼容的目标物质。
                  </p>
               </div>
               
               <div class="relative group self-center md:self-auto hidden sm:block text-white">
                  <div class="absolute -inset-8 bg-blue-600/10 rounded-full blur-2xl group-hover:bg-blue-600/20 transition-all"></div>
                  <div :class="cn('game-card scale-110 sm:scale-125 !cursor-default', getCardStyle(selectedCard))">
                     <div class="text-2xl sm:text-3xl font-black tracking-tighter">{{ selectedCard.type }}</div>
                  </div>
               </div>
             </div>

             <div class="grid grid-cols-2 sm:grid-cols-3 gap-3 sm:gap-4 mb-8 sm:mb-12 max-h-[200px] sm:max-h-[300px] overflow-y-auto pr-2 sm:pr-4 custom-scrollbar">
                <button
                  v-for="(substance, index) in availableSubstances"
                  :key="index"
                  @click="selectedSubstance = substance"
                  :class="cn(
                    'group relative p-3 sm:p-6 rounded-2xl sm:rounded-3xl border transition-all flex flex-col items-center justify-center gap-2 sm:gap-3 overflow-hidden',
                    selectedSubstance === substance 
                      ? 'bg-blue-600/10 border-blue-400 dark:border-blue-500 text-blue-600 dark:text-white shadow-xl dark:shadow-[0_15px_35px_rgba(59,130,246,0.15)]' 
                      : 'bg-slate-50 dark:bg-white/[0.03] border-slate-200 dark:border-white/5 text-slate-500 hover:bg-slate-100 dark:hover:bg-white/[0.05] hover:border-blue-300 dark:hover:border-white/10'
                  )"
                >
                  <div :class="cn(
                    'w-8 h-8 sm:w-12 sm:h-12 rounded-xl sm:rounded-2xl flex items-center justify-center border transition-all duration-500',
                    selectedSubstance === substance ? 'bg-blue-500/20 border-blue-500/30 rotate-12' : 'bg-slate-500/10 dark:bg-black/40 border-slate-200 dark:border-white/5 opacity-40 group-hover:rotate-12'
                  )">
                    <FlaskConical :class="cn('w-4 h-4 sm:w-6 sm:h-6', selectedSubstance === substance ? 'text-blue-600 dark:text-blue-400' : 'text-slate-400 dark:text-slate-600')" />
                  </div>
                  <span :class="cn('font-black tracking-widest text-[9px] sm:text-[11px] uppercase truncate w-full text-center', selectedSubstance === substance ? 'text-blue-600 dark:text-white' : 'text-slate-500')">{{ substance }}</span>
                  <div v-if="selectedSubstance === substance" class="absolute inset-0 bg-blue-500/5 animate-pulse"></div>
                </button>
             </div>

             <div class="flex gap-3 sm:gap-4">
                <button 
                  @click="selectedCard = null; selectedSubstance = null;" 
                  class="flex-1 h-12 sm:h-16 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white font-black rounded-xl sm:rounded-2xl transition-all uppercase tracking-widest text-[9px] sm:text-[11px] border border-slate-200 dark:border-white/5"
                >
                  终止
                </button>
                <button 
                  @click="handlePlayCard"
                  :disabled="!selectedSubstance"
                  class="flex-[2] h-12 sm:h-16 bg-blue-600 hover:bg-blue-500 text-white font-black rounded-xl sm:rounded-2xl transition-all shadow-[0_20px_40px_rgba(37,99,235,0.2)] dark:shadow-[0_20px_40px_rgba(37,99,235,0.3)] disabled:opacity-50 disabled:grayscale flex items-center justify-center gap-2 sm:gap-3 group/confirm relative overflow-hidden"
                >
                  <span class="uppercase tracking-widest sm:tracking-[0.2em] text-[10px] sm:text-xs">执行反应</span>
                  <ChevronRight class="w-4 h-4 sm:w-5 sm:h-5 group-hover/confirm:translate-x-1 transition-transform" />
                </button>
             </div>
           </div>
        </div>
      </div>

      <!-- Experimental Victory / Failure Protocol -->
      <div v-if="gameState?.status === 'finished'" class="fixed inset-0 z-[200] flex items-center justify-center p-4 sm:p-6">
        <div class="absolute inset-0 bg-slate-900/60 dark:bg-black/95 backdrop-blur-2xl animate-in fade-in duration-1000" />
        
        <div class="relative w-full max-w-xl bg-white dark:bg-[#0a0a0c] border border-blue-500/20 rounded-[48px] sm:rounded-[64px] p-8 sm:p-16 flex flex-col items-center text-center overflow-hidden animate-in fade-in zoom-in spin-in-1 duration-1000 shadow-2xl">
           <!-- Background Glow -->
           <div class="absolute -top-32 -left-32 w-64 h-64 bg-blue-500/10 dark:bg-blue-500/20 rounded-full blur-[100px]"></div>
           <div class="absolute -bottom-32 -right-32 w-64 h-64 bg-purple-500/5 dark:bg-purple-500/10 rounded-full blur-[100px]"></div>

           <div class="relative mb-8 sm:mb-12 transform-gpu">
              <div class="absolute inset-0 bg-blue-500/30 rounded-full blur-3xl animate-pulse"></div>
              <div class="w-24 h-24 sm:w-32 h-32 bg-gradient-to-br from-blue-500 to-blue-700 rounded-[30px] sm:rounded-[40px] flex items-center justify-center shadow-[0_20px_60px_rgba(59,130,246,0.3)] dark:shadow-[0_20px_60px_rgba(59,130,246,0.5)] rotate-12">
                 <Trophy class="w-12 h-12 sm:w-16 sm:h-16 text-white" />
              </div>
              <div class="absolute -bottom-4 -right-4 w-10 h-10 sm:w-12 sm:h-12 bg-white rounded-xl flex items-center justify-center shadow-2xl animate-bounce">
                 <Zap class="w-5 h-5 sm:w-6 sm:h-6 text-blue-600 fill-current" />
              </div>
           </div>

           <div class="space-y-3 sm:space-y-4 mb-10 sm:mb-16 px-4">
              <div class="inline-flex items-center gap-2 px-3 sm:px-4 py-1 sm:py-1.5 bg-blue-500/10 border border-blue-500/20 rounded-full">
                 <span class="w-2 h-2 bg-blue-500 rounded-full animate-ping"></span>
                 <span class="text-[8px] sm:text-[10px] font-black text-blue-600 dark:text-blue-400 uppercase tracking-widest font-mono">Mission_Success</span>
              </div>
              <template v-if="winner?.uid === user.uid">
                <h2 class="text-4xl sm:text-6xl font-black text-slate-900 dark:text-white tracking-tighter leading-none">
                  实验大获成功
                </h2>
                <p class="text-xs sm:text-sm text-slate-500 dark:text-slate-400 font-medium leading-relaxed max-w-sm mx-auto">
                  恭喜研究员！你已成功稳定了反应核心。此项成果将被载入实验室历史。
                </p>
              </template>
              <template v-else>
                <h2 class="text-4xl sm:text-6xl font-black text-slate-900 dark:text-white tracking-tighter leading-none">
                  反应链终止
                </h2>
                <p class="text-xs sm:text-sm text-slate-500 dark:text-slate-400 font-medium leading-relaxed max-w-sm mx-auto">
                  实验由 <span class="text-slate-900 dark:text-white font-black">{{ winner?.username }}</span> 成功收官。
                </p>
              </template>

              <!-- 积分变动显示 (如有) -->
              <div v-if="gameState?.points_changes" class="w-full mt-6 p-4 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl">
                 <div class="flex items-center justify-between mb-3 border-b border-slate-200 dark:border-white/5 pb-2">
                    <span class="text-[10px] font-black uppercase tracking-widest text-slate-500">Player_Rankings</span>
                    <span class="text-[10px] font-black uppercase tracking-widest text-blue-500">Points_Δ</span>
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
                class="w-full h-14 sm:h-18 bg-blue-600 hover:bg-blue-500 text-white font-black rounded-[20px] sm:rounded-3xl transition-all shadow-[0_20px_40px_rgba(37,99,235,0.2)] dark:shadow-[0_20px_40px_rgba(37,99,235,0.3)] hover:scale-105 active:scale-95 flex items-center justify-center gap-2 sm:gap-3 group relative overflow-hidden"
              >
                 <span class="uppercase tracking-widest sm:tracking-[0.3em] text-[11px] sm:text-sm">返回指挥大厅</span>
                 <ChevronRight class="w-5 h-5 sm:w-6 sm:h-6 group-hover:translate-x-1 transition-transform" />
              </button>
           </div>
        </div>
      </div>
    </template>
  </div>
</template>
