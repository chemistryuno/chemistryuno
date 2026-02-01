<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { gameAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import websocket from '../utils/websocket'
import { ArrowLeft, Play, RefreshCw, Zap, Activity, FlaskConical, Trophy, ChevronRight, Loader2, Users, Timer } from 'lucide-vue-next'
import { cn } from '../utils/cn'

const route = useRoute()
const router = useRouter()
const { showAlert, showConfirm } = useDialog()
const id = route.params.id as string

const user = ref(JSON.parse(localStorage.getItem('user') || '{}'))
const gameState = ref<any>(null)
const roomInfo = ref<any>(null)
const availableSubstances = ref<string[]>([])
const turnReadySubstances = ref<string[]>([])
const selectedCard = ref<any>(null)
const selectedSubstance = ref<string | null>(null)
const doubleMode = ref(false)
const firstDoubleSubstance = ref<string | null>(null)
const secondDoubleSubstance = ref<string | null>(null)
const substanceInput = ref('')
const loading = ref(true)
const timeRemaining = ref(30)
let timerInterval: any = null

const currentPlayerObj = computed(() => {
  if (!gameState.value) return null
  return gameState.value.players?.[gameState.value.current_player]
})
const isMyTurn = computed(() => {
  if (!currentPlayerObj.value || !user.value) return false
  return currentPlayerObj.value.uid === user.value.uid
})
const myData = computed(() => {
  if (!gameState.value || !user.value) return null
  return (gameState.value.players || []).find((p: any) => p.uid === user.value.uid)
})
const myIndex = computed(() => {
  if (!gameState.value || !user.value) return -1
  return (gameState.value.players || []).findIndex((p: any) => p.uid === user.value.uid)
})
const allowedAny = computed(() => {
  if (!gameState.value) return false
  return typeof gameState.value?.allowed_any_player !== 'undefined' && gameState.value?.allowed_any_player === myIndex.value
})
const winner = computed(() => gameState.value?.players?.find((p: any) => p.card_count === 0))

// --- 人机实验室移植功能 ---
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
      status: data.status
    }
    
    if (data.game_state) {
      gameState.value = data.game_state
    }
    
    loading.value = false
  } catch (error: any) {
    console.error('加载游戏状态失败:', error)
    loading.value = false
    
    if (error.response?.status === 404) {
      showAlert('房间不存在或已被关闭', '未知实验室')
      router.push('/')
    } else if (error.response?.status === 401) {
      showAlert('身份验证失败，请重新登录', '准入失败')
      router.push('/login')
    } else {
      showAlert('加载房间失败，请稍后重试', '系统错误')
    }
  }
}

onMounted(() => {
  loadGameState().then(() => {
    websocket.joinRoom(id)
    websocket.on('game_update', handleGameUpdate)
    websocket.on('player_joined', loadGameState)
    websocket.on('player_left', loadGameState)
  })
})

onUnmounted(() => {
  if (timerInterval) clearInterval(timerInterval)
  websocket.leaveRoom()
  websocket.off('game_update', handleGameUpdate)
  websocket.off('player_joined', loadGameState)
  websocket.off('player_left', loadGameState)
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
  'bottom-2 sm:bottom-0 translate-y-0 sm:translate-y-1/2 translate-x-1/2 right-1/2',
  'left-2 sm:left-0 -translate-x-0 sm:-translate-x-1/2 top-1/2 -translate-y-1/2',
  'top-2 sm:top-0 -translate-y-0 sm:-translate-y-1/2 translate-x-1/2 right-1/2',
  'right-2 sm:right-0 translate-x-0 sm:translate-x-1/2 top-1/2 -translate-y-1/2',
]

const isMobile = ref(false)
onMounted(() => {
  isMobile.value = window.innerWidth < 640
  const handleResize = () => {
    isMobile.value = window.innerWidth < 640
  }
  window.addEventListener('resize', handleResize)
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

      <!-- Experimental Header -->
      <header class="h-[60px] sm:h-[72px] bg-white/60 dark:bg-black/40 backdrop-blur-2xl border-b border-slate-200 dark:border-white/5 px-4 sm:px-6 flex justify-between items-center z-50 sticky top-0">
        <div class="flex items-center gap-3 sm:gap-6">
          <button 
            @click="handleLeaveRoom" 
            class="w-8 h-8 sm:w-10 sm:h-10 flex items-center justify-center hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl sm:rounded-2xl text-slate-500 hover:text-slate-900 dark:hover:text-white transition-all group"
          >
            <ArrowLeft class="w-4 h-4 sm:w-5 sm:h-5 group-hover:-translate-x-1 transition-transform" />
          </button>
          
          <div class="hidden md:flex items-center gap-4 bg-slate-100/50 dark:bg-white/5 px-4 py-1.5 rounded-2xl border border-slate-200 dark:border-white/10">
            <div class="flex flex-col items-end">
              <span class="text-[7px] font-black text-slate-500 uppercase tracking-widest leading-none mb-1">Expertise</span>
              <span class="text-[10px] font-black italic text-blue-500 leading-none">LV.{{ level }}</span>
            </div>
            <div class="w-16 h-1.5 bg-slate-200 dark:bg-white/10 rounded-full overflow-hidden">
              <div class="h-full bg-blue-500 transition-all duration-1000" :style="{ width: (exp % 100) + '%' }"></div>
            </div>
          </div>

          <div class="h-6 sm:h-8 w-px bg-slate-200 dark:bg-white/10"></div>
          <div>
            <div class="flex items-center gap-2">
               <span class="text-[8px] sm:text-[10px] font-mono text-blue-600 dark:text-blue-500/50 uppercase tracking-widest hidden xs:block">Node:</span>
               <h2 class="text-xs sm:text-sm font-black tracking-widest uppercase font-mono text-slate-900 dark:text-white">{{ id.substring(0, 8) }}</h2>
            </div>
            <div class="flex items-center gap-2 mt-0.5">
               <div :class="cn('w-1 h-1 sm:w-1.5 sm:h-1.5 rounded-full animate-pulse', roomInfo?.status === 'waiting' ? 'bg-amber-500' : 'bg-emerald-500')"></div>
               <p class="text-[8px] sm:text-[10px] text-slate-500 dark:text-slate-400 font-bold uppercase tracking-widest truncate max-w-[80px] sm:max-w-none">
                {{ roomInfo?.status === 'waiting' ? 'Waiting' : 'Processing' }}
               </p>
            </div>
          </div>
        </div>

        <div class="flex items-center gap-2 sm:gap-4">
          <button 
            v-if="roomInfo?.status === 'waiting' && user.uid === roomInfo?.host_uid"
            @click="handleStartGame" 
            class="bg-blue-600 hover:bg-blue-500 px-3 sm:px-6 h-9 sm:h-11 rounded-xl sm:rounded-2xl font-black text-[10px] sm:text-xs uppercase tracking-wider sm:tracking-[0.2em] shadow-[0_10px_20px_rgba(37,99,235,0.2)] transition-all active:scale-95 flex items-center gap-2 sm:gap-3 group overflow-hidden relative text-white"
          >
            <div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/10 to-transparent -translate-x-full group-hover:animate-shimmer"></div>
            <Play class="w-3 h-3 sm:w-3.5 sm:h-3.5 fill-current" />
            <span class="hidden sm:inline">启动反应堆</span>
            <span class="sm:hidden">启动</span>
          </button>
          
          <div class="px-3 sm:px-5 h-9 sm:h-11 bg-slate-100 dark:bg-white/5 rounded-xl sm:rounded-2xl border border-slate-200 dark:border-white/10 flex items-center gap-2 sm:gap-4 font-mono">
             <div class="flex flex-col items-end">
               <span class="text-[7px] sm:text-[9px] text-slate-500 font-bold uppercase tracking-tight">Users</span>
               <span class="text-[10px] sm:text-xs font-black text-slate-900 dark:text-white leading-none">
                 {{ Array.isArray(roomInfo?.players) ? roomInfo.players.length : 0 }} / {{ roomInfo?.max_players || 4 }}
               </span>
             </div>
             <Users class="w-3 h-3 sm:w-4 sm:h-4 text-blue-600 dark:text-blue-400 opacity-50" />
          </div>
        </div>
      </header>

      <!-- Au Allow Banner -->
      <div v-if="allowedAny" class="w-full flex justify-center mt-3 z-50">
        <div class="bg-yellow-100 dark:bg-yellow-900/30 border border-yellow-300 dark:border-yellow-700 text-yellow-800 dark:text-yellow-200 px-4 py-2 rounded-xl font-bold text-sm shadow-md">
          Au: 跳过 — 你可无视反应条件出牌
        </div>
      </div>

      <!-- Reaction Chamber (Main Table) -->
      <div class="flex-1 relative flex items-center justify-center p-2 sm:p-12 overflow-hidden">          
          <!-- Reaction Logs (Moved from AI Lab) -->
          <div :class="cn(
            'absolute right-4 top-4 bottom-4 w-64 z-[60] bg-white/80 dark:bg-black/80 backdrop-blur-2xl border border-slate-200 dark:border-white/10 rounded-[32px] shadow-2xl transition-all duration-500 flex flex-col overflow-hidden',
            showLogs ? 'translate-x-0 opacity-100' : 'translate-x-[calc(100%+2rem)] opacity-0 pointer-events-none'
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
                         <span class="opacity-40">@{{ gameState?.players?.find(p => p.uid === play.player_uid)?.username || 'User' }}</span>
                      </div>
                   </div>
                </div>
             </div>
          </div>

          <!-- Log Toggle Button -->
          <button 
            @click="showLogs = !showLogs"
            :class="cn(
              'absolute right-6 top-6 z-50 p-3 rounded-2xl border transition-all active:scale-95 group',
              showLogs ? 'bg-blue-600 border-blue-500 text-white' : 'bg-white/50 dark:bg-black/40 border-slate-200 dark:border-white/10 text-slate-500 hover:text-blue-500'
            )"
          >
             <Zap :class="cn('w-5 h-5 group-hover:animate-pulse', showLogs && 'fill-current')" />
          </button>
        <!-- Table Console Background -->
        <div class="absolute w-full max-w-5xl aspect-square sm:aspect-[16/10] max-h-[90%] bg-slate-200/20 dark:bg-[#121216]/20 rounded-[40px] sm:rounded-[80px] border border-slate-200 dark:border-white/5 shadow-[inset_0_0_100px_rgba(0,0,0,0.05)] dark:shadow-[inset_0_0_100px_rgba(0,0,0,0.5)] pointer-events-none">
           <div class="absolute top-8 left-1/2 -translate-x-1/2 flex gap-12 opacity-30">
              <div class="w-16 sm:w-32 h-1 bg-gradient-to-r from-transparent via-blue-500 to-transparent"></div>
           </div>
        </div>

        <!-- Players Radial Layout -->
        <div class="w-full max-w-5xl aspect-square sm:aspect-[16/10] max-h-full relative flex items-center justify-center">
          
          <!-- Direction Indicator -->
          <div class="absolute inset-0 flex items-center justify-center pointer-events-none">
             <div :class="cn(
                'w-[300px] h-[300px] sm:w-[500px] sm:h-[500px] border-2 border-dashed border-blue-500/10 rounded-full transition-all duration-1000',
                gameState?.direction === 1 ? 'animate-spin-slow' : 'animate-reverse-spin-slow'
             )"></div>
          </div>
          
          <!-- Reaction Core (Center Pile) -->
          <div class="relative z-20 flex flex-col sm:flex-row items-center gap-4 sm:gap-12 lg:gap-20 scale-75 sm:scale-90 lg:scale-100">
            <!-- Draw Pile -->
            <div class="group relative">
               <div class="absolute -inset-4 bg-blue-500/10 dark:bg-blue-500/20 rounded-[40px] blur-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>
               <div 
                  @click="isMyTurn ? handleDrawCard() : undefined"
                  :class="cn(
                    'w-20 h-28 sm:w-28 sm:h-40 bg-gradient-to-br from-slate-100 to-slate-200 dark:from-slate-800 dark:to-slate-900 rounded-[20px] sm:rounded-[28px] border-2 border-slate-200 dark:border-white/5 flex flex-col items-center justify-center gap-1 sm:gap-3 shadow-xl dark:shadow-2xl transition-all relative overflow-hidden',
                    isMyTurn ? 'cursor-pointer hover:border-blue-400 dark:hover:border-blue-500/50 hover:-translate-y-2 active:scale-95 group' : 'grayscale opacity-40 cursor-not-allowed'
                  )"
                >
                  <div class="absolute top-2 left-2 right-2 flex justify-between items-center opacity-40 dark:opacity-20 text-slate-500">
                     <span class="text-[5px] sm:text-[7px] font-mono uppercase">Stack</span>
                  </div>
                  <div class="w-8 h-8 sm:w-12 sm:h-12 bg-slate-900/5 dark:bg-white/5 rounded-xl sm:rounded-2xl flex items-center justify-center border border-slate-900/5 dark:border-white/10 group-hover:scale-110 group-hover:bg-blue-500/10 transition-all duration-500">
                    <RefreshCw class="w-4 h-4 sm:w-6 sm:h-6 text-slate-400 dark:text-slate-500 group-hover:text-blue-600 dark:group-hover:text-blue-400 group-hover:rotate-180 transition-all duration-700" />
                  </div>
                  <div class="text-center px-1">
                    <span class="text-[7px] sm:text-[9px] font-black uppercase tracking-widest text-slate-400 dark:text-slate-500 group-hover:text-blue-600 dark:group-hover:text-blue-500/70 transition-colors">DRAW</span>
                  </div>
                </div>
            </div>

            <!-- Discard / Current reaction -->
            <div class="relative flex flex-col items-center gap-4 sm:gap-6">
              <!-- Turn Timer -->
              <div v-if="gameState?.status === 'playing'" 
                :class="cn(
                  'absolute -top-16 sm:-top-24 flex items-center gap-2 px-4 py-1.5 rounded-full border backdrop-blur-md transition-all duration-500 z-50',
                  timeRemaining <= 10 ? 'bg-red-500/20 border-red-500 text-red-400 animate-pulse' : 'bg-blue-500/10 border-blue-500/30 text-blue-600 dark:text-blue-400 border-blue-200 dark:border-blue-500/30'
                )"
              >
                <Timer :class="cn('w-3 h-3 sm:w-4 sm:h-4', timeRemaining <= 10 && 'animate-bounce')" />
                <span class="font-mono font-black text-xs sm:text-sm tracking-widest">{{ timeRemaining }}S</span>
              </div>

              <div v-if="gameState?.last_card" class="relative">
                <div class="absolute -inset-8 sm:-inset-12 bg-blue-600/10 rounded-full blur-3xl animate-pulse"></div>
                <div :class="cn(
                  'game-card scale-[1.1] sm:scale-[1.3] pointer-events-none shadow-[0_20px_40px_rgba(0,0,0,0.1)] dark:shadow-[0_20px_40px_rgba(0,0,0,0.8)] z-10 text-white', 
                  getCardStyle(gameState?.last_card?.card),
                  getDynamicCardClass(gameState?.last_card?.card)
                )">
                    <div class="absolute top-1 sm:top-2 left-1 sm:left-2 text-[6px] sm:text-[8px] uppercase font-black opacity-30 tracking-widest leading-none">Latest</div>
                    <div class="flex flex-col items-center justify-center">
                      <div class="text-2xl sm:text-3xl tracking-tighter font-black font-mono italic">{{ gameState?.last_card?.card?.type }}</div>
                      <div v-if="gameState?.last_card?.card?.effect || ['He','Ne','Ar','Kr'].includes(gameState?.last_card?.card?.type)" class="text-[8px] sm:text-[10px] font-bold bg-white/20 px-1.5 py-0.5 rounded-full mt-1 uppercase tracking-tighter">
                        {{ ['He','Ne','Ar','Kr'].includes(gameState?.last_card?.card?.type) ? '转向' : gameState?.last_card?.card?.effect === 'Au' ? '跳过' : gameState?.last_card?.card?.effect === '+2' ? '+2' : gameState?.last_card?.card?.effect === '+4' ? '+4' : gameState?.last_card?.card?.effect }}
                      </div>
                    </div>
                </div>
              </div>
              <div v-else class="w-28 h-40 sm:w-36 sm:h-52 rounded-[24px] sm:rounded-[32px] border-2 border-dashed border-slate-200 dark:border-white/5 flex flex-col items-center justify-center gap-2 opacity-30">
                <Loader2 class="w-5 h-5 sm:w-6 sm:h-6 animate-spin text-slate-500 dark:text-slate-600" />
                <p class="text-[8px] sm:text-[10px] text-slate-500 dark:text-slate-600 font-black uppercase tracking-widest">Awaiting</p>
              </div>
               
              <div v-if="gameState?.last_card" class="flex flex-col items-center gap-2 mt-4 sm:mt-8">
                <div class="bg-white/70 dark:bg-blue-600/10 backdrop-blur-3xl px-4 sm:px-8 py-2 sm:py-3 rounded-2xl sm:rounded-[24px] border border-blue-200 dark:border-blue-500/20 shadow-[0_10px_30px_rgba(0,0,0,0.05)] dark:shadow-[0_10px_30px_rgba(0,0,0,0.3)] flex flex-col items-center animate-in fade-in slide-in-from-top-4">
                  <div class="flex items-center gap-2 mb-1">
                    <FlaskConical class="w-3 h-3 text-blue-500" />
                    <span class="uppercase tracking-widest text-[7px] sm:text-[8px] font-black text-slate-400 dark:text-slate-500">Stability Verified</span>
                  </div>
                  <h3 class="text-xl sm:text-3xl font-black text-blue-600 dark:text-blue-400 tracking-tighter font-mono italic" v-html="formatFormula(gameState?.last_card?.substance)"></h3>
                  <p class="text-[9px] sm:text-[11px] font-bold text-slate-500 uppercase tracking-[0.2em] italic">{{ getSubstanceName(gameState?.last_card?.substance) }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Player badges positions (radial) -->
          <template v-if="(gameState?.players || []).length > 0">
            <div 
              v-for="(player, index) in (gameState?.players || [])"
              :key="player.uid"
              :class="cn(
                'absolute transition-all duration-700 z-30',
                playerPositions[Number(index) % playerPositions.length]
              )"
            >
              <div :class="cn(
                'p-1.5 sm:p-2 rounded-[20px] sm:rounded-[32px] transition-all duration-500 group/player border backdrop-blur-xl shadow-2xl flex items-center gap-2 sm:gap-4 min-w-[140px] sm:min-w-[180px]',
                gameState?.current_player === index 
                  ? 'bg-blue-600/10 border-blue-400 dark:border-blue-500 shadow-[0_0_40px_rgba(59,130,246,0.1)] dark:shadow-[0_0_40px_rgba(59,130,246,0.3)] scale-105 sm:scale-110' 
                  : 'bg-white/80 dark:bg-black/60 border-slate-200 dark:border-white/5 hover:border-blue-400 dark:hover:border-white/20'
              )">
                <div class="relative flex-shrink-0">
                  <div class="w-10 h-10 sm:w-14 sm:h-14 bg-slate-100 dark:bg-slate-800 rounded-xl sm:rounded-2xl flex items-center justify-center text-xl sm:text-2xl border border-slate-200 dark:border-white/10 group-hover/player:scale-105 transition-transform overflow-hidden shadow-inner">
                    <template v-if="player.avatar && player.avatar.startsWith('data:')">
                       <img :src="player.avatar" class="w-full h-full object-cover" />
                    </template>
                    <template v-else>
                       {{ player.avatar || '🧪' }}
                    </template>
                  </div>
                  <div v-if="gameState?.current_player === index" class="absolute -top-1 -right-1 bg-blue-600 dark:bg-blue-500 p-1 rounded-md sm:rounded-lg shadow-lg animate-bounce">
                    <Zap class="w-2.5 h-2.5 sm:w-3 sm:h-3 text-white fill-current" />
                  </div>
                </div>
                <div class="flex flex-col pr-1 sm:pr-2 min-w-0">
                  <div class="flex items-center gap-1 sm:gap-2">
                     <span :class="cn('text-[9px] sm:text-[10px] font-bold uppercase tracking-tight truncate max-w-[60px] sm:max-w-[80px]', gameState?.current_player === index ? 'text-blue-600 dark:text-blue-400' : 'text-slate-900 dark:text-white')">{{ player.username }}</span>
                     <span v-if="player.uid === user.uid" class="text-[7px] sm:text-[8px] bg-blue-500/20 text-blue-400 px-1 rounded font-mono">YOU</span>
                  </div>
                  <div class="flex items-center gap-2 sm:gap-3 mt-1 sm:mt-1.5">
                     <div class="flex items-center gap-1">
                        <div class="grid grid-cols-3 sm:grid-cols-2 gap-0.5 sm:gap-0.5">
                           <div v-for="i in 3" :key="i" :class="cn('w-0.5 h-0.5 sm:w-1 sm:h-1 rounded-full', (i-1) < player.card_count ? 'bg-blue-400' : 'bg-slate-700')"></div>
                        </div>
                        <span class="text-[9px] sm:text-[10px] font-mono text-slate-400">{{ player.card_count }} 张</span>
                     </div>
                     <!-- 行动进度 / 双联反应就绪标志 -->
                     <div v-if="player.double_action_available" class="flex items-center gap-1 bg-blue-500/10 px-1.5 py-0.5 rounded-full border border-blue-500/20">
                        <Zap class="w-2 h-2 text-blue-400 fill-current" />
                        <span class="text-[7px] sm:text-[8px] font-black text-blue-400 uppercase tracking-tighter">Ready</span>
                     </div>
                     <div v-else class="flex items-center gap-0.5 opacity-40">
                        <div v-for="i in 2" :key="i" :class="cn('w-2 h-0.5 rounded-full', i <= (player.action_progress || 0) ? 'bg-blue-500' : 'bg-slate-700')"></div>
                     </div>
                  </div>
                </div>
                <div v-if="gameState?.current_player === index" class="absolute inset-0 rounded-[20px] sm:rounded-[32px] border-2 border-blue-500/50 animate-pulse pointer-events-none"></div>
              </div>
            </div>
          </template>
          <div v-else-if="roomInfo?.status === 'waiting'" class="absolute bottom-0 translate-y-1/2 translate-x-1/2 right-1/2 z-30">
            <div class="p-3 sm:p-4 rounded-[20px] sm:rounded-[32px] bg-black/60 border border-white/5 backdrop-blur-xl shadow-2xl flex items-center gap-2 sm:gap-3 min-w-[160px] sm:min-w-[200px]">
              <Loader2 class="w-6 h-6 sm:w-8 sm:h-8 text-blue-500 animate-spin" />
              <div class="flex flex-col">
                <span class="text-[10px] sm:text-xs font-bold text-white uppercase tracking-tight">等待玩家加入</span>
                <span class="text-[8px] sm:text-[10px] text-slate-400">{{ roomInfo.players?.length || 0 }} / {{ roomInfo.max_players || 4 }} 玩家</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Hand / Deck Area -->
      <div class="h-[140px] sm:h-[240px] bg-gradient-to-t from-blue-900/5 dark:from-blue-900/10 to-transparent relative mt-auto px-4 sm:px-8 flex flex-col items-center">
        <!-- Turn Ready Substances -->
        <div v-if="isMyTurn" class="absolute -top-24 left-4 right-4 sm:left-12 sm:right-12 flex flex-col gap-2">
          <div v-if="turnReadySubstances.length > 0">
            <div class="flex items-center gap-2 mb-1">
               <FlaskConical class="w-3 h-3 text-blue-500" />
               <span class="text-[8px] sm:text-[10px] font-black uppercase tracking-widest text-slate-500">可进行反应 (点击即可出牌)</span>
            </div>
            <div class="flex gap-2 overflow-x-auto custom-scrollbar-hidden pb-4">
               <button 
                  v-for="sub in turnReadySubstances" 
                  :key="sub"
                  @click="selectedSubstance = sub; handlePlayCard()"
                  class="flex-shrink-0 px-3 py-1.5 sm:px-4 sm:py-2 bg-white/80 dark:bg-white/5 backdrop-blur-md border border-slate-200 dark:border-white/10 rounded-xl sm:rounded-2xl text-[10px] sm:text-xs font-black hover:border-blue-500 hover:text-blue-500 transition-all active:scale-95 flex items-center gap-2 group"
               >
                  <div class="w-1.5 h-1.5 rounded-full bg-blue-500 group-hover:scale-150 transition-transform"></div>
                  {{ sub }}
               </button>
            </div>
          </div>
          <div v-else-if="!loading" class="flex items-center justify-center p-4 bg-amber-500/10 border border-amber-500/20 rounded-2xl backdrop-blur-md animate-in fade-in slide-in-from-top-2">
             <div class="flex items-center gap-3">
                <div class="w-8 h-8 rounded-full bg-amber-500/20 flex items-center justify-center">
                   <Zap class="w-4 h-4 text-amber-500" />
                </div>
                <div class="flex flex-col">
                   <span class="text-[10px] sm:text-xs font-black text-amber-500 uppercase tracking-widest">能量不足 / 无法反应</span>
                   <span class="text-[8px] sm:text-[9px] text-slate-500 dark:text-slate-400">请抽取额外反应底物 (摸2张牌并结束回合)</span>
                </div>
             </div>
          </div>
        </div>

        <!-- Turn Tip -->
        <div class="h-0 relative w-full flex justify-center">
           <div v-if="isMyTurn" class="absolute -top-4 sm:-top-6 translate-y-[-100%] flex flex-col items-center gap-1 sm:gap-2 animate-in fade-in slide-in-from-bottom-2">
              <div class="flex items-center gap-2 bg-white/90 dark:bg-slate-900/80 backdrop-blur-xl border border-slate-200 dark:border-white/10 p-1.5 rounded-2xl shadow-2xl mb-2">
                <input 
                  v-model="substanceInput" 
                  @keyup.enter="handleInputPlay"
                  placeholder="输入化学式 (如 H2O)" 
                  class="bg-transparent border-none outline-none text-[10px] sm:text-xs px-3 py-1 w-32 sm:w-48 font-black tracking-widest uppercase placeholder:text-slate-400 text-slate-900 dark:text-white"
                />
                <button 
                  @click="handleInputPlay"
                  class="bg-blue-600 hover:bg-blue-500 w-8 h-8 rounded-xl flex items-center justify-center transition-all active:scale-90"
                >
                  <ChevronRight class="w-4 h-4 text-white" />
                </button>
              </div>
              <div class="bg-blue-600 px-4 sm:px-8 py-1.5 sm:py-2.5 rounded-full shadow-[0_15px_30px_rgba(37,99,235,0.2)] dark:shadow-[0_15px_30px_rgba(37,99,235,0.4)] flex items-center gap-2 sm:gap-3 active:scale-95 transition-transform relative group">
                <Zap class="w-3 h-3 sm:w-4 sm:h-4 fill-current animate-pulse text-white" />
                <span class="text-[9px] sm:text-xs font-black uppercase tracking-widest sm:tracking-[0.3em] text-white">Your_Turn_Active ({{ timeRemaining }}s)</span>
                
                <!-- 双联行动按钮 -->
                <button 
                  v-if="myData?.double_action_available"
                  @click.stop="toggleDoubleMode"
                  :class="cn(
                    'absolute -right-2 top-1/2 -translate-y-1/2 translate-x-full ml-4 px-4 py-2 rounded-2xl border-2 transition-all flex items-center gap-2 whitespace-nowrap overflow-hidden group/btn',
                    doubleMode 
                      ? 'bg-amber-500 border-amber-400 text-white shadow-[0_0_20px_rgba(245,158,11,0.4)]' 
                      : 'bg-white/10 backdrop-blur-md border-white/20 text-white hover:bg-white/20'
                  )"
                >
                   <div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/10 to-transparent -translate-x-full group-hover/btn:animate-shimmer"></div>
                   <Activity :class="cn('w-4 h-4', doubleMode && 'animate-spin')" />
                   <div class="flex flex-col items-start leading-none">
                      <span class="text-[10px] font-black uppercase tracking-tighter">{{ doubleMode ? '取消双联' : '双联反应' }}</span>
                      <span class="text-[7px] font-bold opacity-70 uppercase tracking-tighter">Double Action</span>
                   </div>
                </button>
              </div>

              <!-- 双联模式提示状态 -->
              <div v-if="doubleMode" class="mt-4 flex flex-col items-center gap-2 animate-in slide-in-from-top-4 duration-500">
                <div class="flex items-center gap-3">
                  <div :class="cn('w-12 h-12 rounded-2xl flex items-center justify-center border-2 transition-all duration-500', firstDoubleSubstance ? 'bg-blue-500/20 border-blue-500 shadow-lg' : 'bg-slate-800/50 border-white/10 opacity-50')">
                    <span v-if="firstDoubleSubstance" class="text-xs font-black">{{ firstDoubleSubstance }}</span>
                    <FlaskConical v-else class="w-5 h-5 text-slate-500" />
                  </div>
                  <div class="w-4 h-0.5 bg-blue-500/30"></div>
                  <div :class="cn('w-12 h-12 rounded-2xl flex items-center justify-center border-2 transition-all duration-500', secondDoubleSubstance ? 'bg-blue-500/20 border-blue-500 shadow-lg' : 'bg-slate-800/50 border-white/10 opacity-50')">
                    <span v-if="secondDoubleSubstance" class="text-xs font-black">{{ secondDoubleSubstance }}</span>
                    <FlaskConical v-else class="w-5 h-5 text-slate-500" />
                  </div>
                </div>
                <div class="text-[8px] font-black text-amber-500 uppercase tracking-[0.2em] animate-pulse">
                   {{ !firstDoubleSubstance ? '请选择/输入第一种参与物质' : !secondDoubleSubstance ? '请选择/输入第二种参与物质' : '反应就绪，点击执行' }}
                </div>
                <button 
                  v-if="firstDoubleSubstance && secondDoubleSubstance"
                  @click="handleDoublePlay"
                  class="mt-2 bg-gradient-to-r from-amber-500 to-orange-600 px-6 py-2 rounded-xl text-white font-black text-[10px] uppercase tracking-widest shadow-xl hover:scale-105 active:scale-95 transition-all"
                >
                  启动双联反应仪式
                </button>
              </div>
           </div>
        </div>

        <div class="flex-1 w-full max-w-6xl flex justify-center items-end pb-2 sm:pb-6">
           <div class="flex flex-nowrap justify-center gap-x-1 sm:gap-x-4 px-4 sm:px-12 h-[120px] sm:h-[160px] w-full overflow-x-auto custom-scrollbar-hidden py-1 sm:py-3 translate-y-4 hover:translate-y-0 transition-all duration-500">
            <div v-if="roomInfo?.status === 'waiting'" class="flex flex-col items-center justify-center opacity-30 pb-4 sm:pb-10">
              <Loader2 class="w-10 h-10 sm:w-16 sm:h-16 mb-2 sm:mb-4 animate-spin text-blue-500" />
              <p class="font-black uppercase tracking-widest text-[10px] sm:text-sm text-slate-500">等待房主启动反应堆</p>
            </div>
            <template v-else-if="myData?.hand_cards?.length > 0">
              <div
                v-for="(card, index) in myData.hand_cards"
                :key="index"
                @click="isMyTurn && handleCardClick(card)"
                :class="cn(
                  'game-card flex-shrink-0 cursor-pointer transition-all duration-300 transform-gpu origin-bottom text-white',
                  getCardStyle(card),
                  getDynamicCardClass(card),
                  selectedCard === card ? 'selected -translate-y-4 sm:-translate-y-8 scale-110 shadow-[0_20px_40px_rgba(0,0,0,0.5)] z-50 ring-2 ring-blue-500/50' : 'hover:-translate-y-6 hover:rotate-2 hover:z-40',
                  !isMyTurn && 'opacity-40 grayscale-[0.8] cursor-not-allowed pointer-events-none translate-y-8 sm:translate-y-12'
                )"
                :style="{
                  transform: `rotate(${(Number(index) - ((myData?.hand_cards?.length || 0) - 1) / 2) * (isMobile ? 1 : 2)}deg)`
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
