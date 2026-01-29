<script setup lang="ts">
import { ref, onMounted, computed, watch, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { 
  ArrowLeft, 
  Trophy, 
  MessageSquare, 
  Zap, 
  Info,
  User as UserIcon,
  Bot,
  Timer,
  RefreshCw,
  FlaskConical,
  Award
} from 'lucide-vue-next'
import { useDialog } from '../utils/dialog'
import { gameAPI } from '../utils/api'

const router = useRouter()
const { showAlert } = useDialog()

// --- 基础数据 ---
const CHEMIST_NAMES = ["门捷列夫", "拉瓦锡", "波义耳", "道尔顿", "居里夫人", "诺贝尔", "范霍夫", "阿伦尼乌斯", "勒沙特列", "鲍林", "舍勒", "海洛夫斯基", "维尔纳", "哈伯", "费舍尔", "普利斯特里", "卡文迪许", "侯德榜", "徐寿", "庄长恭"];

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

const SPECIAL_CARDS: Record<string, { effect: string, color: string, name: string }> = {
  'Skip': { effect: 'skip', color: 'bg-amber-400 text-white', name: '跳过' },
  'Reverse': { effect: 'reverse', color: 'bg-indigo-500 text-white', name: '转向' },
  '+2': { effect: 'draw2', color: 'bg-rose-500 text-white', name: '+2' },
  'Noble': { effect: 'wild', color: 'bg-gradient-to-br from-yellow-400 via-amber-500 to-yellow-600 text-white', name: '惰性气体' }
}

const PRECIPITATES = new Set(['AgCl', 'AgBr', 'AgI', 'Ag2CO3', 'Ag3PO4', 'BaSO4', 'BaCO3', 'Ba3(PO4)2', 'CaCO3', 'Ca3(PO4)2', 'Mg(OH)2', 'MgCO3', 'Cu(OH)2', 'Fe(OH)3', 'BaSO3', 'CaSO3']);

interface Card {
  type: string;
  id: string;
}

interface Player {
  id: number;
  name: string;
  hand: Card[];
  isBot: boolean;
}

// --- 响应式状态 ---
const deck = ref<Card[]>([])
const players = ref<Player[]>([])
const currentPlayerIndex = ref(0)
const direction = ref(1)
const lastPlayedSubstance = ref<{ formula: string, name: string } | null>(null)
const gameStatus = ref<'waiting' | 'playing' | 'finished'>('waiting')
const logs = ref<string[]>([])
const aiBubble = ref<{ name: string, text: string } | null>(null)
const timeRemaining = ref(30)
const substanceInput = ref('')
const anyCardMode = ref(false)
const manualInputMsg = ref('')

const exp = ref(Number(localStorage.getItem('chem_exp') || '0'))
const level = computed(() => Math.floor(exp.value / 100) + 1)
const achievements = ref<string[]>(JSON.parse(localStorage.getItem('chem_achievements') || '[]'))

const checkAchievements = (substance: string) => {
  if (substance === 'Au' && !achievements.value.includes('炼金术士')) {
    achievements.value.push('炼金术士')
    showAlert('获得成就：炼金术士 (合成单质金)', '成就达成！')
  }
  if (players.value[0].hand.length === 0 && !achievements.value.includes('实验达人')) {
    achievements.value.push('实验达人')
    showAlert('获得成就：实验达人 (赢得一场人机对战)', '成就达成！')
  }
  localStorage.setItem('chem_achievements', JSON.stringify(achievements.value))
}

const addExp = (amount: number) => {
  exp.value += amount
  localStorage.setItem('chem_exp', exp.value.toString())
}

const user = JSON.parse(localStorage.getItem('user') || '{"username": "玩家"}')

let timerInterval: any = null

// --- 工具函数 ---
const formatFormula = (formula: string) => {
  return formula.replace(/(\d+)/g, '<sub>$1</sub>')
}

const log = (msg: string) => {
  logs.value.unshift(msg)
  if (logs.value.length > 50) logs.value.pop()
}

const createDeck = () => {
  const newDeck: Card[] = []
  const elements = Object.keys(ELEMENTS_DATA)
  
  // 添加元素牌
  elements.forEach(el => {
    for (let i = 0; i < 6; i++) {
      newDeck.push({ type: el, id: `${el}-${i}-${Math.random()}` })
    }
  })
  
  // 添加特殊牌
  Object.keys(SPECIAL_CARDS).forEach(spec => {
    for (let i = 0; i < 4; i++) {
      newDeck.push({ type: spec, id: `${spec}-${i}-${Math.random()}` })
    }
  })
  
  return newDeck.sort(() => Math.random() - 0.5)
}

const drawCard = (count = 1) => {
  const cards: Card[] = []
  for (let i = 0; i < count; i++) {
    if (deck.value.length === 0) deck.value = createDeck()
    cards.push(deck.value.pop()!)
  }
  return cards
}

const startGame = () => {
  deck.value = createDeck()
  players.value = [
    { id: 0, name: user.username, hand: drawCard(7), isBot: false },
    { id: 1, name: CHEMIST_NAMES[Math.floor(Math.random() * CHEMIST_NAMES.length)], hand: drawCard(7), isBot: true },
    { id: 2, name: CHEMIST_NAMES[Math.floor(Math.random() * CHEMIST_NAMES.length)], hand: drawCard(7), isBot: true }
  ]
  
  currentPlayerIndex.value = 0
  direction.value = 1
  lastPlayedSubstance.value = { formula: 'H2O', name: '水' }
  gameStatus.value = 'playing'
  logs.value = []
  log('化学反应实验开始！当前底牌: H2O')
  startTimer()
}

const startTimer = () => {
  if (timerInterval) clearInterval(timerInterval)
  timeRemaining.value = 30
  timerInterval = setInterval(() => {
    if (timeRemaining.value > 0) {
      timeRemaining.value--
    } else {
      handleTimeout()
    }
  }, 1000)
}

const handleTimeout = () => {
  if (!isMyTurn.value) return
  log(`${user.username} 实验超时，自动摸牌`)
  const newCards = drawCard(1)
  players.value[0].hand.push(...newCards)
  nextTurn()
}

const nextTurn = () => {
  if (checkWinner()) return
  
  currentPlayerIndex.value = (currentPlayerIndex.value + direction.value + players.value.length) % players.value.length
  startTimer()
  
  if (players.value[currentPlayerIndex.value].isBot) {
    setTimeout(aiTurn, 1500)
  }
}

const checkWinner = () => {
  const winner = players.value.find(p => p.hand.length === 0)
  if (winner) {
    gameStatus.value = 'finished'
    clearInterval(timerInterval)
    showAlert(`${winner.name} 成功清空手牌，获得了实验胜利！`, '实验完成')
    setTimeout(() => {
      router.push('/')
    }, 2000)
    return true
  }
  return false
}

const isMyTurn = computed(() => currentPlayerIndex.value === 0)

// --- 化学逻辑 ---
const getIons = (comp: string) => {
  const ions: string[] = []
  if (comp.includes('H') && !comp.includes('OH') && comp !== 'H2O') ions.push('H+')
  if (comp.includes('OH')) ions.push('OH-')
  if (comp.includes('Na')) ions.push('Na+')
  if (comp.includes('K')) ions.push('K+')
  if (comp.includes('Ca')) ions.push('Ca2+')
  if (comp.includes('Ba')) ions.push('Ba2+')
  if (comp.includes('Mg')) ions.push('Mg2+')
  if (comp.includes('Cu')) ions.push('Cu2+')
  if (comp.includes('Fe')) ions.push('Fe3+')
  if (comp.includes('Cl')) ions.push('Cl-')
  if (comp.includes('SO4')) ions.push('SO4 2-')
  if (comp.includes('CO3')) ions.push('CO3 2-')
  if (comp.includes('NO3')) ions.push('NO3-')
  return ions
}

const checkReaction = async (formula1: string, formula2: string) => {
  if (formula1 === formula2) return true
  
  // 1. 尝试后端校验
  try {
    const res = await gameAPI.checkReaction(formula1, formula2)
    if (res.data.can_react) return true
  } catch (e) {
    console.error('后端校验失败，切换到本地逻辑:', e)
  }

  // 2. 本地兜底逻辑
  const ions1 = getIons(formula1)
  const ions2 = getIons(formula2)
  
  // 酸碱中和
  if ((ions1.includes('H+') && ions2.includes('OH-')) || (ions1.includes('OH-') && ions2.includes('H+'))) return true
  
  // 沉淀反应简化判断
  for (const i1 of ions1) {
    for (const i2 of ions2) {
      const combined = i1.replace('+', '') + i2.replace('-', '')
      for (const p of PRECIPITATES) {
        if (p.includes(i1.split(/[0-9]/)[0]) && p.includes(i2.split(/[0-9]/)[0])) return true
      }
    }
  }
  
  // 同元素
  const elements1: string[] = formula1.match(/[A-Z][a-z]?/g) || []
  const elements2: string[] = formula2.match(/[A-Z][a-z]?/g) || []
  if (elements1.some(e => elements2.includes(e))) return true

  return false
}

const handleManualSubmit = async () => {
  if (!isMyTurn.value) return
  const input = substanceInput.value.trim()
  if (!input) return
  
  const canItReact = await checkReaction(input, lastPlayedSubstance.value?.formula || '')
  
  if (canItReact) {
    // 检查手牌是否足够组成该物质
    const elementsNeeded: string[] = input.match(/[A-Z][a-z]?/g) || []
    const myHand = players.value[0].hand
    
    // 简化逻辑：只需持有一种构成元素即可出牌，或持有 Noble 卡
    const hasElement = myHand.some(c => elementsNeeded.includes(c.type) || c.type === 'Noble')
    
    if (hasElement || anyCardMode.value) {
      // 移除对应的牌
      const index = myHand.findIndex(c => elementsNeeded.includes(c.type) || c.type === 'Noble')
      if (index !== -1) myHand.splice(index, 1)
      
      lastPlayedSubstance.value = { formula: input, name: '新物质' }
      log(`${user.username} 合成了 ${input}`)
      addExp(10)
      checkAchievements(input)
      substanceInput.value = ''
      nextTurn()
    } else {
      manualInputMsg.value = '缺少构成该物质的元素牌'
    }
  } else {
    manualInputMsg.value = '该物质无法与当前底牌反应'
  }
}

const aiTurn = async () => {
  const bot = players.value[currentPlayerIndex.value]
  if (!bot.isBot) return
  
  const possibilities = [
    { formula: 'HCl', name: '盐酸' },
    { formula: 'NaOH', name: '氢氧化钠' },
    { formula: 'NaCl', name: '氯化钠' },
    { formula: 'CO2', name: '二氧化碳' },
    { formula: 'H2O', name: '水' },
    { formula: 'CuSO4', name: '硫酸铜' },
    { formula: 'Zn', name: '锌' },
    { formula: 'O2', name: '氧气' },
    { formula: 'H2', name: '氢气' },
    { formula: 'Fe', name: '铁' },
    { formula: 'AgNO3', name: '硝酸银' },
    { formula: 'CaCO3', name: '碳酸钙' },
    { formula: 'BaCl2', name: '氯化钡' },
    { formula: 'Na2SO4', name: '硫酸钠' },
    { formula: 'H2SO4', name: '硫酸' },
    { formula: 'MgO', name: '氧化镁' },
    { formula: 'CuO', name: '氧化铜' },
    { formula: 'NH3', name: '氨气' }
  ]
  
  const valid = []
  for (const p of possibilities) {
    if (await checkReaction(p.formula, lastPlayedSubstance.value?.formula || '')) {
      valid.push(p)
    }
  }
  
  if (valid.length > 0) {
    const choice = valid[Math.floor(Math.random() * valid.length)]
    lastPlayedSubstance.value = choice
    
    // 智能扣牌 logic
    const elementsNeeded: string[] = choice.formula.match(/[A-Z][a-z]?/g) || []
    const index = bot.hand.findIndex(c => elementsNeeded.includes(c.type) || c.type === 'Noble')
    if (index !== -1) {
      bot.hand.splice(index, 1)
    } else {
      bot.hand.pop() // 兜底扣牌
    }
    
    log(`${bot.name} 合成了 ${choice.name} (${choice.formula})`)
    
    // AI 气泡
    aiBubble.value = { name: bot.name, text: `我合成了 ${choice.formula}，该你进行了！` }
    setTimeout(() => aiBubble.value = null, 3000)
    
  } else {
    bot.hand.push(...drawCard(1))
    log(`${bot.name} 无法反应，摸一张牌`)
  }
  
  nextTurn()
}

const handleDraw = () => {
  if (!isMyTurn.value) return
  players.value[0].hand.push(...drawCard(1))
  log(`${user.username} 实验受阻，摸一张牌`)
  nextTurn()
}

onUnmounted(() => {
  if (timerInterval) clearInterval(timerInterval)
})
</script>

<template>
  <div class="h-screen bg-slate-50 dark:bg-[#0a0a0c] text-slate-900 dark:text-white flex flex-col font-sans overflow-hidden">
    <!-- 背景修饰 -->
    <div class="fixed inset-0 pointer-events-none opacity-20">
      <div class="absolute top-0 left-0 w-full h-full bg-[url('https://www.transparenttextures.com/patterns/carbon-fibre.png')]"></div>
      <div class="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-500/20 blur-[120px] rounded-full animate-pulse"></div>
      <div class="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-500/20 blur-[120px] rounded-full animate-pulse delay-700"></div>
    </div>

    <!-- 顶部状态栏 -->
    <header class="h-16 bg-white/50 dark:bg-black/20 backdrop-blur-xl border-b border-white/5 px-6 flex items-center justify-between relative z-50">
      <div class="flex items-center gap-4">
        <button @click="router.push('/')" class="p-2 hover:bg-white/10 rounded-xl transition-colors">
          <ArrowLeft class="w-5 h-5" />
        </button>
        <div class="flex flex-col">
          <span class="text-[10px] font-black text-blue-500 uppercase tracking-widest">Laboratory Mode</span>
          <h1 class="text-xs sm:text-sm font-black italic uppercase tracking-tighter">Human VS AI Battle</h1>
        </div>
        
        <div class="ml-6 hidden md:flex items-center gap-3">
          <div class="flex flex-col items-end">
            <span class="text-[8px] font-black text-slate-500 uppercase tracking-widest">Researcher Level</span>
            <span class="text-xs font-black italic text-blue-500">LV.{{ level }}</span>
          </div>
          <div class="w-24 h-1.5 bg-slate-200 dark:bg-white/5 rounded-full overflow-hidden">
            <div class="h-full bg-blue-500 transition-all duration-1000" :style="{ width: (exp % 100) + '%' }"></div>
          </div>
        </div>
      </div>

      <div class="flex items-center gap-4">
        <div class="px-4 py-2 bg-blue-500/10 border border-blue-500/20 rounded-xl flex items-center gap-3">
          <Timer :class="['w-4 h-4', timeRemaining < 10 ? 'text-red-500 animate-pulse' : 'text-blue-400']" />
          <span :class="['font-mono font-bold', timeRemaining < 10 ? 'text-red-500' : 'text-blue-400']">00:{{ timeRemaining.toString().padStart(2, '0') }}</span>
        </div>
      </div>
    </header>

    <main class="flex-1 relative flex flex-col items-center justify-center p-6 sm:p-12">
      <!-- 玩家信息（顶部和两侧） -->
      <div class="absolute inset-0 pointer-events-none p-8">
        <div v-for="(p, i) in players" :key="p.id" 
          :class="['absolute flex flex-col items-center gap-3 transition-all duration-500', 
            i === 0 ? 'bottom-8 left-1/2 -translate-x-1/2' : 
            i === 1 ? 'top-8 left-8' : 'top-8 right-8']"
        >
          <div :class="['w-16 h-16 rounded-2xl border flex items-center justify-center relative shadow-2xl transition-all',
            currentPlayerIndex === i ? 'border-blue-500 scale-110 shadow-blue-500/20' : 'border-white/10 bg-white/5']"
          >
            <UserIcon v-if="i === 0" class="w-8 h-8 text-blue-400" />
            <Bot v-else class="w-8 h-8 text-purple-400" />
            
            <!-- 手牌数提示 -->
             <div class="absolute -top-2 -right-2 w-7 h-7 bg-slate-800 border border-white/10 rounded-lg flex items-center justify-center text-[10px] font-black">
              {{ p.hand.length }}
             </div>
             
             <!-- AI 气泡 -->
             <div v-if="aiBubble && aiBubble.name === p.name" class="absolute left-full ml-4 top-0 w-48 bg-white dark:bg-slate-800 p-3 rounded-2xl border border-blue-500/30 shadow-2xl animate-in fade-in slide-in-from-left-2">
                <p class="text-[10px] text-blue-400 font-black mb-1">@PROCESSOR</p>
                <p class="text-xs font-bold leading-relaxed">{{ aiBubble.text }}</p>
             </div>
          </div>
          <p :class="['text-[10px] font-black uppercase tracking-widest', currentPlayerIndex === i ? 'text-blue-500' : 'text-slate-500']">
            {{ p.name }}
          </p>
        </div>
      </div>

      <!-- 中心反应区域 -->
      <div class="relative group">
        <!-- 反应堆光效 -->
        <div class="absolute inset-x-[-100px] inset-y-[-100px] bg-blue-500/5 blur-[100px] group-hover:bg-blue-500/10 transition-all duration-1000"></div>
        
        <div v-if="gameStatus === 'playing'" class="relative z-10 w-64 h-84 sm:w-80 sm:h-auto aspect-[4/5] bg-white dark:bg-slate-900 border-4 border-slate-200 dark:border-white/10 rounded-[3rem] shadow-2xl flex flex-col items-center justify-center p-8 transition-all hover:scale-105 active:scale-95 cursor-pointer">
          <div class="text-slate-400 uppercase text-[10px] font-black tracking-widest mb-4 flex items-center gap-2">
            <FlaskConical class="w-3 h-3" /> Latest Product
          </div>
          <h2 class="text-4xl sm:text-6xl font-black text-blue-600 dark:text-blue-400 mb-2 font-mono italic tracking-tighter" v-html="formatFormula(lastPlayedSubstance?.formula || 'H2O')"></h2>
          <p class="text-sm font-bold text-slate-500 uppercase tracking-widest italic">{{ lastPlayedSubstance?.name || '水' }}</p>
          
          <div class="mt-8 flex gap-2">
            <div class="px-3 py-1 bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 rounded-lg text-[8px] font-black uppercase">Verified</div>
            <div class="px-3 py-1 bg-blue-500/10 text-blue-500 border border-blue-500/20 rounded-lg text-[8px] font-black uppercase tracking-tight">Stable</div>
          </div>
        </div>
        
        <div v-else-if="gameStatus === 'waiting'" class="relative z-10 text-center space-y-8 animate-in fade-in zoom-in duration-700">
           <div class="w-32 h-32 bg-blue-500/10 border border-blue-500/30 rounded-[40px] flex items-center justify-center transform rotate-12 mx-auto">
             <FlaskConical class="w-16 h-16 text-blue-400" />
           </div>
           <div>
             <h2 class="text-4xl font-black text-white italic tracking-tighter uppercase mb-2">Ready to Start?</h2>
             <p class="text-slate-500 font-bold uppercase tracking-widest text-xs">Prepare your chemical formulas for the grand battle</p>
           </div>
           <button @click="startGame" class="bg-blue-600 hover:bg-blue-500 text-white px-10 py-4 rounded-2xl font-black uppercase tracking-widest shadow-2xl shadow-blue-500/30 transition-all active:scale-95 group">
             Initialize Process
           </button>
        </div>
      </div>

      <!-- 日志面板 -->
      <div class="absolute right-8 bottom-32 w-64 h-auto hidden xl:flex flex-col gap-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3 text-[10px] font-black text-slate-500 uppercase tracking-widest">
             <Zap class="w-3 h-3 text-amber-500" /> Reaction Logs
          </div>
        </div>
        <div class="h-48 bg-white/50 dark:bg-black/40 backdrop-blur-xl border border-white/5 rounded-3xl p-4 overflow-y-auto custom-scrollbar">
           <div v-for="(msg, i) in logs" :key="i" class="text-[10px] font-bold mb-2 last:mb-0 text-slate-500 dark:text-slate-400 animate-in fade-in slide-in-from-right-2">
              <span class="text-blue-500 mr-2">></span> {{ msg }}
           </div>
        </div>
      </div>
    </main>

    <!-- 底部控制栏 -->
    <footer v-if="gameStatus === 'playing'" class="bg-white/80 dark:bg-[#111114]/80 backdrop-blur-2xl border-t border-white/5 p-4 sm:p-8 relative z-50">
      <div class="max-w-6xl mx-auto flex flex-col gap-6">
        <!-- 输入框区域 -->
        <div class="flex flex-col sm:flex-row items-center gap-4">
          <div class="relative flex-1 group">
            <FlaskConical class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-600 group-focus-within:text-blue-400 transition-colors" />
            <input 
              v-model="substanceInput"
              @keyup.enter="handleManualSubmit"
              type="text" 
              placeholder="Enter Chemical Formula (e.g. HCl, NaOH...)"
              class="w-full bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl pl-12 pr-6 py-4 text-sm font-mono focus:outline-none focus:border-blue-500/50 transition-all placeholder:text-slate-500 placeholder:italic"
            />
          </div>
          <button @click="handleManualSubmit" class="bg-blue-600 hover:bg-blue-500 text-white px-8 py-4 rounded-xl font-black uppercase tracking-widest transition-all">
            Synthesize
          </button>
          <button @click="handleDraw" class="bg-slate-200 dark:bg-white/10 dark:hover:bg-white/20 text-slate-700 dark:text-white px-6 py-4 rounded-xl font-black uppercase tracking-widest transition-all flex items-center gap-2">
            <RefreshCw class="w-4 h-4" /> Draw
          </button>
        </div>

        <!-- 手牌展示 -->
        <div class="flex items-center gap-4 overflow-x-auto pb-2 custom-scrollbar no-scrollbar">
          <div v-for="card in players[0]?.hand" :key="card.id"
            class="flex-shrink-0 w-24 h-32 rounded-2xl border-2 flex flex-col items-center justify-center gap-2 transition-all hover:-translate-y-2 cursor-pointer shadow-xl"
            :class="[
              ELEMENTS_DATA[card.type] ? ELEMENTS_DATA[card.type].color : 
              SPECIAL_CARDS[card.type] ? SPECIAL_CARDS[card.type].color : 'bg-slate-300'
            ]"
          >
             <span class="text-xl font-black font-mono italic">{{ card.type }}</span>
             <span class="text-[9px] font-bold uppercase opacity-80">{{ ELEMENTS_DATA[card.type]?.name || SPECIAL_CARDS[card.type]?.name }}</span>
          </div>
        </div>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.no-scrollbar::-webkit-scrollbar {
  display: none;
}
.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
</style>
