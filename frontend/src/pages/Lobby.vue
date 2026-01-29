<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { gameAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import websocket from '../utils/websocket'
import { Beaker, Plus, Users, Shield, LogOut, Settings, Play, Info, X, Loader2, Database, Bot } from 'lucide-vue-next'
import { cn } from '../utils/cn'

const props = defineProps<{
  // user props can be added if we pass from App.vue
}>()

const router = useRouter()
const { showAlert } = useDialog()
const user = ref(JSON.parse(localStorage.getItem('user') || '{}'))
const rooms = ref<any[]>([])
const decks = ref<any[]>([])
const showCreateModal = ref(false)
const roomName = ref('')
const maxPlayers = ref(4)
const deckID = ref(0)
const loading = ref(false)
const currentTime = ref(new Date())

const loadDecks = async () => {
  try {
    const res = await gameAPI.getMyDecks()
    decks.value = res.data
    if (decks.value.length > 0) {
      deckID.value = decks.value[0].id
    }
  } catch (e) {
    console.error(e)
  }
}

let roomInterval: any
let timeInterval: any

onMounted(() => {
  loadRooms()
  loadDecks()
  websocket.connect()

  roomInterval = setInterval(loadRooms, 3000)
  timeInterval = setInterval(() => {
    currentTime.value = new Date()
  }, 1000)
})

onUnmounted(() => {
  if (roomInterval) clearInterval(roomInterval)
  if (timeInterval) clearInterval(timeInterval)
})

const loadRooms = async () => {
  try {
    const response = await gameAPI.getRooms()
    rooms.value = response.data || []
  } catch (error) {
    console.error('加载房间列表失败:', error)
  }
}

const handleCreateRoom = async () => {
  loading.value = true

  try {
    const response = await gameAPI.createRoom(roomName.value, maxPlayers.value, deckID.value)
    const room = response.data
    router.push(`/room/${room.id}`)
  } catch (error: any) {
    showAlert(error.response?.data?.error || '创建房间失败', '系统异常')
  } finally {
    loading.value = false
  }
}

const handleJoinRoom = async (roomId: string) => {
  try {
    await gameAPI.joinRoom(roomId)
    router.push(`/room/${roomId}`)
  } catch (error: any) {
    showAlert(error.response?.data?.error || '加入房间失败', '连接错误')
  }
}

const handleLogout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  websocket.disconnect()
  router.push('/login')
}

const activeNodesCount = computed(() => rooms.value.filter(r => r.status === 'playing').length)
</script>

<template>
  <div class="min-h-screen bg-slate-50 dark:bg-[#0a0a0c] text-slate-900 dark:text-slate-200 font-sans selection:bg-blue-500/30 overflow-x-hidden">
    <!-- Background Decor -->
    <div class="fixed inset-0 pointer-events-none overflow-hidden">
      <div class="absolute top-[-20%] left-[-10%] w-[60%] h-[60%] bg-blue-600/5 rounded-full blur-[150px]"></div>
      <div class="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-purple-600/5 rounded-full blur-[150px]"></div>
      <div class="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/carbon-fibre.png')] opacity-20"></div>
    </div>

    <!-- Main Layout Layer -->
    <div class="relative z-10 flex flex-col min-h-screen">
      
      <!-- Top Command Bar -->
      <header class="h-20 border-b border-slate-200 dark:border-white/5 bg-white/60 dark:bg-black/40 backdrop-blur-xl sticky top-0 z-50">
        <div class="max-w-[1400px] mx-auto h-full px-6 flex items-center justify-between">
          <div class="flex items-center gap-6">
            <div class="flex items-center gap-3 group px-4 py-2 bg-gradient-to-br from-blue-500/10 to-blue-600/5 border border-blue-500/20 rounded-2xl">
              <Beaker class="w-8 h-8 text-blue-500 group-hover:rotate-12 transition-transform" />
              <div>
                 <h1 class="text-lg font-black tracking-tighter text-slate-900 dark:text-white leading-none">CHEMISTRY <span class="text-blue-500">UNO</span></h1>
                 <p class="text-[10px] text-blue-500/50 font-mono tracking-widest leading-none mt-1 uppercase">Lab_Control_v4</p>
              </div>
            </div>

            <!-- Status Indicators (Desktop) -->
            <div class="hidden lg:flex items-center gap-6 text-[10px] font-mono tracking-[0.2em] text-slate-500 border-l border-slate-200 dark:border-white/10 pl-6 uppercase">
              <div class="flex items-center gap-2">
                <div class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></div>
                SERVER: STABLE
              </div>
              <div class="flex items-center gap-2">
                 <div class="w-1.5 h-1.5 rounded-full bg-blue-500"></div>
                 UP_TIME: {{ currentTime.toLocaleTimeString() }}
              </div>
            </div>
          </div>

          <div class="flex items-center gap-4">
            <!-- User Identity Chip -->
            <div class="hidden sm:flex items-center gap-3 pl-2 pr-4 py-1.5 bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl hover:bg-slate-200 dark:hover:bg-white/10 transition-all cursor-pointer group">
               <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-slate-200 to-slate-300 dark:from-slate-700 dark:to-slate-800 flex items-center justify-center text-xl shadow-inner group-hover:scale-105 transition-transform overflow-hidden">
                 <template v-if="user.avatar && user.avatar.startsWith('data:')">
                    <img :src="user.avatar" class="w-full h-full object-cover" />
                 </template>
                 <template v-else>
                    {{ user.avatar || '🧪' }}
                 </template>
               </div>
               <div class="flex flex-col">
                 <span class="text-xs font-black text-slate-900 dark:text-white">{{ user.username }}</span>
                 <span class="text-[9px] text-slate-500 font-mono flex items-center gap-1 uppercase">
                   <template v-if="user.is_admin">
                     <Shield class="w-2.5 h-2.5 text-yellow-500" /> Research_Lead
                   </template>
                   <template v-else>
                     Researcher_Alpha
                   </template>
                 </span>
               </div>
            </div>

            <div class="flex items-center gap-1.5">
              <router-link to="/profile" class="p-3 hover:bg-white/5 rounded-2xl transition-all text-slate-400 hover:text-white" title="实验室档案">
                <Settings class="w-5 h-5" />
              </router-link>
              <router-link v-if="user.role === 'admin' || user.role === 'co-worker'" to="/reactions" class="p-3 hover:bg-blue-500/10 rounded-2xl transition-all text-blue-500/70 hover:text-blue-400" title="反应数据库">
                <Database class="w-5 h-5" />
              </router-link>
              <router-link v-if="user.is_admin" to="/admin" class="p-3 hover:bg-yellow-500/10 rounded-2xl transition-all text-yellow-500/70 hover:text-yellow-400" title="科研管理">
                <Shield class="w-5 h-5" />
              </router-link>
              <div class="w-px h-6 bg-white/10 mx-1"></div>
              <button @click="handleLogout" class="p-3 hover:bg-red-500/10 rounded-2xl transition-all text-red-500/70 hover:text-red-400" title="切断连接">
                <LogOut class="w-5 h-5" />
              </button>
            </div>
          </div>
        </div>
      </header>

      <main class="flex-1 max-w-[1400px] mx-auto w-full px-6 py-8">
        <!-- Welcome & Global Actions -->
        <div class="flex flex-col lg:flex-row lg:items-end justify-between mb-12 gap-8">
          <div class="space-y-4">
            <div class="inline-flex items-center gap-2 px-3 py-1 bg-blue-500/10 border border-blue-500/20 rounded-full">
              <span class="w-1.5 h-1.5 bg-blue-500 rounded-full animate-ping"></span>
              <span class="text-[10px] font-bold text-blue-400 uppercase tracking-widest">Live Research Hall</span>
            </div>
            <h2 class="text-5xl font-black text-white tracking-tighter leading-none">
              实验大厅
            </h2>
            <p class="text-slate-500 dark:text-slate-400 max-w-lg font-medium leading-relaxed">
              欢迎回到元素实验室。目前有 <span class="text-slate-900 dark:text-white font-bold">{{ rooms.length }}</span> 个活跃实验，请加入现有队列或开启全新化学反应序列。
            </p>
          </div>

          <div class="flex items-center gap-6">
             <div class="hidden xl:flex items-center gap-8 px-8 py-5 bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/5 rounded-[32px]">
               <div class="text-center">
                 <p class="text-[10px] text-slate-500 uppercase font-bold tracking-widest mb-1">Total_Players</p>
                 <p class="text-2xl font-black text-slate-900 dark:text-white font-mono">1,248</p>
               </div>
               <div class="w-px h-8 bg-slate-200 dark:bg-white/5 font-mono"></div>
               <div class="text-center font-mono">
                 <p class="text-[10px] text-slate-500 uppercase font-bold tracking-widest mb-1">Active_Nodes</p>
                 <p class="text-2xl font-black text-blue-600 dark:text-blue-400">{{ activeNodesCount }}</p>
               </div>
             </div>

             <button 
              @click="router.push('/ai-battle')" 
              class="group relative flex items-center gap-3 bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 hover:border-purple-500/50 px-8 py-5 rounded-[24px] font-black text-slate-900 dark:text-white transition-all hover:scale-[1.02] hover:-translate-y-1 active:scale-95 overflow-hidden shadow-sm"
            >
              <Bot class="w-5 h-5 text-purple-500 group-hover:animate-bounce" />
              <span class="uppercase tracking-widest text-sm">人机实验室</span>
              <div class="absolute inset-0 w-full h-full bg-gradient-to-r from-transparent via-purple-500/5 to-transparent -translate-x-full group-hover:animate-shimmer"></div>
            </button>

            <button 
              @click="showCreateModal = true" 
              class="group relative flex items-center gap-3 bg-blue-600 hover:bg-blue-500 px-8 py-5 rounded-[24px] font-black text-white shadow-[0_20px_40px_rgba(37,99,235,0.2)] dark:shadow-[0_20px_40px_rgba(37,99,235,0.4)] transition-all hover:scale-[1.02] hover:-translate-y-1 active:scale-95 overflow-hidden"
            >
              <Plus class="w-5 h-5 group-hover:rotate-90 transition-transform duration-500" />
              <span class="uppercase tracking-widest text-sm">启动新实验</span>
              <div class="absolute inset-0 w-full h-full bg-gradient-to-r from-transparent via-white/10 to-transparent -translate-x-full group-hover:animate-shimmer"></div>
            </button>
          </div>
        </div>

        <!-- Experimental Nodes (Room List) -->
        <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
          <div v-if="rooms.length === 0" class="col-span-full py-32 flex flex-col items-center justify-center bg-slate-100/50 dark:bg-white/[0.02] border-2 border-dashed border-slate-200 dark:border-white/5 rounded-[40px] text-slate-400 dark:text-slate-600 transition-colors hover:bg-slate-100 dark:hover:bg-white/[0.03] hover:border-slate-300 dark:hover:border-white/10 group">
            <div class="w-24 h-24 bg-slate-200/50 dark:bg-white/5 rounded-full flex items-center justify-center mb-6 group-hover:scale-110 transition-transform">
              <Info class="w-10 h-10 opacity-30" />
            </div>
            <p class="text-2xl font-black text-slate-400 dark:text-slate-500 tracking-tight">NO_ACTIVE_EXPERIMENTS</p>
            <p class="text-sm mt-3 font-mono opacity-50 uppercase tracking-widest">请等待节点激活或手动创建</p>
          </div>
          <template v-else>
            <div 
              v-for="room in rooms"
              :key="room.id" 
              class="group relative bg-white/80 dark:bg-[#121216]/60 backdrop-blur-xl border border-slate-200 dark:border-white/10 rounded-[32px] p-1 transition-all hover:bg-white dark:hover:bg-[#16161c] hover:border-blue-500/30 hover:shadow-[0_20px_50px_rgba(0,0,0,0.1)] dark:hover:shadow-[0_20px_50px_rgba(0,0,0,0.5)] flex flex-col h-[320px]"
            >
              <div class="flex-1 p-6 flex flex-col">
                <div class="flex justify-between items-start mb-6">
                  <div class="flex flex-col">
                    <span class="text-[10px] font-mono text-blue-600 dark:text-blue-500/60 uppercase tracking-widest mb-1">Experiment_ID_{{ room.id.substring(0, 4) }}</span>
                    <h3 class="text-xl font-black text-slate-900 dark:text-white group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors truncate max-w-[180px] leading-tight">
                      {{ room.name }}
                    </h3>
                  </div>
                  <div :class="cn(
                    'px-2.5 py-1 rounded-lg text-[9px] font-black uppercase tracking-widest border',
                    room.status === 'waiting' ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20' : 
                    room.status === 'playing' ? 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20' : 
                    'bg-slate-500/10 text-slate-500 dark:text-slate-400 border-slate-500/20'
                  )">
                    {{ room.status === 'waiting' ? '● Ready' : room.status === 'playing' ? '○ Active' : 'End' }}
                  </div>
                </div>

                <div class="space-y-4 mb-auto">
                  <div class="flex items-center gap-3 p-3 bg-slate-50 dark:bg-white/5 rounded-2xl border border-slate-100 dark:border-white/5 group-hover:border-slate-200 dark:group-hover:border-white/10 transition-colors">
                    <div class="w-8 h-8 rounded-lg bg-blue-500/10 flex items-center justify-center">
                      <Users class="w-4 h-4 text-blue-600 dark:text-blue-400" />
                    </div>
                    <div class="flex flex-col">
                      <span class="text-[9px] text-slate-400 dark:text-slate-500 uppercase tracking-widest font-bold leading-none mb-1">Participants</span>
                      <span class="text-sm font-black text-slate-900 dark:text-white leading-none">
                        {{ room.players?.length || 0 }} <span class="text-slate-400 dark:text-slate-600 font-normal">/ {{ room.max_players }}</span>
                      </span>
                    </div>
                  </div>

                  <div class="flex items-center gap-3 p-3 bg-slate-50 dark:bg-white/5 rounded-2xl border border-slate-100 dark:border-white/5 group-hover:border-slate-200 dark:group-hover:border-white/10 transition-colors">
                    <div class="w-8 h-8 rounded-lg bg-purple-500/10 flex items-center justify-center">
                      <Shield class="w-4 h-4 text-purple-600 dark:text-purple-400" />
                    </div>
                    <div class="flex flex-col">
                      <span class="text-[9px] text-slate-400 dark:text-slate-500 uppercase tracking-widest font-bold leading-none mb-1">Safety_Level</span>
                      <span class="text-sm font-black text-slate-900 dark:text-white leading-none uppercase">Standard_Alpha</span>
                    </div>
                  </div>
                </div>

                <div class="mt-6">
                  <button 
                    v-if="room.status === 'waiting' && (room.players?.length || 0) < room.max_players"
                    @click="handleJoinRoom(room.id)" 
                    class="w-full h-14 bg-slate-100 dark:bg-white/5 hover:bg-blue-600 hover:text-white text-slate-900 dark:text-white border border-slate-200 dark:border-white/10 hover:border-blue-500 rounded-[20px] font-black transition-all flex items-center justify-center gap-2 group/btn relative overflow-hidden active:scale-95"
                  >
                    <Play class="w-4 h-4 fill-current group-hover/btn:translate-x-1 transition-transform" />
                    <span class="uppercase tracking-widest text-xs">执行初始化</span>
                  </button>
                  <div v-else class="w-full h-14 bg-slate-100 dark:bg-slate-800/20 border border-slate-200 dark:border-white/5 rounded-[20px] flex items-center justify-center gap-2 grayscale opacity-50 cursor-not-allowed">
                    <Loader2 class="w-4 h-4 animate-spin text-slate-400 dark:text-slate-500" />
                    <span class="uppercase tracking-widest text-xs font-bold text-slate-400 dark:text-slate-500">正在进行中</span>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </div>
      </main>

      <!-- Global Footer Terminal -->
      <footer class="mt-auto border-t border-white/5 bg-black/40 backdrop-blur-md p-4">
        <div class="max-w-[1400px] mx-auto flex flex-col md:flex-row justify-between items-center text-[10px] font-mono text-slate-500 uppercase tracking-[0.2em] gap-4">
          <div class="flex items-center gap-4">
            <span>System_Core_Ready</span>
            <span class="h-3 w-px bg-white/10"></span>
            <span class="text-blue-500">Secure_WebSocket_Active</span>
            <span class="h-3 w-px bg-white/10"></span>
            <span class="hidden sm:inline">AES_ENCRYPTION_ENABLED</span>
          </div>
          <div>
            &copy; 2024 LAB_V4-ALPHA PROTCOL. ALL RIGHTS RESERVED.
          </div>
        </div>
      </footer>
    </div>

    <!-- Modern Create Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-[#000]/80 backdrop-blur-md animate-in fade-in" @click="showCreateModal = false" />
      <div class="relative w-full max-w-lg bg-[#121216] border border-white/10 rounded-[40px] shadow-2xl overflow-hidden animate-in fade-in zoom-in slide-in-from-bottom-10 duration-500">
         <!-- Modal Header -->
         <div class="px-8 py-8 border-b border-white/5 flex items-center justify-between">
            <div class="flex items-center gap-4">
              <div class="w-12 h-12 bg-blue-500/10 border border-blue-500/20 rounded-2xl flex items-center justify-center text-blue-400">
                <Plus class="w-6 h-6" />
              </div>
              <div>
                <h2 class="text-2xl font-black text-white tracking-tight">开启新实验</h2>
                <p class="text-[10px] text-slate-500 font-mono uppercase tracking-widest">Setup_Experiment_Parameters</p>
              </div>
            </div>
            <button 
              @click="showCreateModal = false"
              class="p-3 hover:bg-white/5 rounded-2xl transition-colors text-slate-500 hover:text-white"
            >
              <X class="w-6 h-6" />
            </button>
         </div>

        <form @submit.prevent="handleCreateRoom" class="p-10 space-y-8">
          <div class="space-y-3">
            <div class="flex justify-between items-center px-1">
               <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest">实验空间命名</label>
               <span class="text-[9px] text-blue-500/40">IDENTIFIER_ALPHA</span>
            </div>
            <input
              v-model="roomName"
              type="text"
              required
              autofocus
              placeholder="请输入实验代号..."
              class="w-full bg-black/40 border border-white/5 text-white px-6 py-5 rounded-3xl focus:ring-1 focus:ring-blue-500/50 focus:border-blue-500/50 outline-none transition-all placeholder:text-slate-800 font-mono text-sm"
            />
          </div>

          <div class="space-y-4">
            <div class="flex justify-between items-center px-1">
               <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest">参与研究员人数</label>
               <span class="text-[9px] text-blue-500/40">CAPACITY_CONFIG</span>
            </div>
            <div class="grid grid-cols-4 gap-4">
              <button
                v-for="num in [2, 3, 4, 8]"
                :key="num"
                type="button"
                @click="maxPlayers = num"
                :class="cn(
                  'h-16 rounded-2xl text-sm font-black border transition-all flex items-center justify-center relative group/opt overflow-hidden',
                  maxPlayers === num 
                    ? 'bg-blue-500/10 border-blue-500/50 text-blue-400 ring-1 ring-blue-500/20 shadow-[0_0_20px_rgba(59,130,246,0.1)]' 
                    : 'bg-white/5 border-white/5 text-slate-600 hover:bg-white/10 hover:border-white/10'
                )"
              >
                <span class="relative z-10">{{ num }}P</span>
                <div v-if="maxPlayers === num" class="absolute inset-0 bg-blue-500/5 animate-pulse"></div>
              </button>
            </div>
          </div>

          <div class="space-y-4">
            <div class="flex justify-between items-center px-1">
               <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest">选择实验牌组</label>
               <span class="text-[9px] text-blue-500/40">DECK_PROTOCOL</span>
            </div>
            <div class="space-y-3">
              <button
                v-for="deck in decks"
                :key="deck.id"
                type="button"
                @click="deckID = deck.id"
                :class="cn(
                  'w-full flex items-center gap-4 p-4 rounded-3xl border transition-all text-left group/deck',
                  deckID === deck.id 
                    ? 'bg-blue-600/10 border-blue-500/50 shadow-[0_10px_30px_rgba(59,130,246,0.1)]' 
                    : 'bg-white/5 border-white/5 hover:border-white/10 hover:bg-white/10'
                )"
              >
                <div :class="cn(
                  'w-12 h-12 rounded-2xl flex items-center justify-center transition-colors',
                  deckID === deck.id ? 'bg-blue-500 text-white' : 'bg-white/5 text-slate-500'
                )">
                   <Beaker class="w-5 h-5" />
                </div>
                <div class="flex-1">
                  <p :class="cn('text-xs font-black uppercase tracking-wider', deckID === deck.id ? 'text-blue-400' : 'text-white')">
                    {{ deck.name }}
                  </p>
                  <p class="text-[10px] text-slate-500 mt-1 font-medium">
                    {{ deck.is_global ? 'Global Protocol' : 'Custom Sequence' }} • {{ Object.keys(deck.cards || {}).length }} elements
                  </p>
                </div>
                <div v-if="deckID === deck.id" class="w-2 h-2 rounded-full bg-blue-500 animate-pulse mr-2"></div>
              </button>
            </div>
          </div>

          <div class="flex gap-4 pt-6">
            <button 
              type="button" 
              @click="showCreateModal = false" 
              class="flex-1 h-14 bg-white/5 hover:bg-white/10 text-slate-400 font-bold rounded-2xl transition-all uppercase tracking-widest text-xs border border-white/5"
            >
              放弃设置
            </button>
            <button 
              type="submit" 
              :disabled="loading"
              class="flex-1 h-14 bg-blue-600 hover:bg-blue-500 text-white font-black rounded-2xl transition-all shadow-[0_15px_30px_rgba(37,99,235,0.3)] active:scale-95 disabled:grayscale flex items-center justify-center gap-2 group/sub relative overflow-hidden"
            >
              <template v-if="loading">
                <Loader2 class="w-5 h-5 animate-spin" />
              </template>
              <template v-else>
                <Play class="w-3.5 h-3.5 fill-current" />
                <span class="uppercase tracking-[0.2em] text-xs">执行初始化</span>
              </template>
              <div class="absolute inset-0 w-full h-full bg-gradient-to-r from-transparent via-white/10 to-transparent -translate-x-full group-hover/sub:animate-shimmer"></div>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
