<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { gameAPI, authAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import websocket from '../utils/websocket'
import { Beaker, Plus, Users, Shield, LogOut, Settings, Play, Info, X, Loader2, Database, MessageSquare, Trash2, Trophy } from 'lucide-vue-next'
import { cn } from '../utils/cn'

const props = defineProps<{
  // user props can be added if we pass from App.vue
}>()

const router = useRouter()
const { showAlert, showConfirm } = useDialog()
const user = ref(JSON.parse(localStorage.getItem('user') || '{}'))
const rooms = ref<any[]>([])
const decks = ref<any[]>([])
const pendingFeedbacks = ref<any[]>([])
const showCreateModal = ref(false)
const roomName = ref('')
const maxPlayers = ref(4)
const deckID = ref(0)
const isPointsMode = ref(false)
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
  loadPendingFeedbacks()
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

const loadPendingFeedbacks = async () => {
  try {
    const res = await authAPI.getMyFeedbacks()
    pendingFeedbacks.value = (res.data || []).filter((f: any) => f.status === 'unread')
  } catch (e) {
    console.error(e)
  }
}

const handleQuickWithdraw = async (id: number) => {
  const confirmed = await showConfirm('确认要直接撤回这条待处理的反馈吗？', '快速撤回')
  if (!confirmed) return
  try {
    await authAPI.withdrawFeedback(id)
    await loadPendingFeedbacks()
    showAlert('反馈已成功撤回', '已撤回')
  } catch (e: any) {
    showAlert(e.response?.data?.error || '撤回失败', '错误')
  }
}

const handleCreateRoom = async () => {
  loading.value = true

  try {
    const response = await gameAPI.createRoom(roomName.value, maxPlayers.value, deckID.value, isPointsMode.value)
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
                 <p class="text-[10px] text-blue-500/50 font-mono tracking-widest leading-none mt-1 uppercase">V1.0.0 Mendeleef</p>
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
                     Researcher_Mendeleef
                   </template>
                 </span>
               </div>
            </div>

            <div class="flex items-center gap-1.5">
              <router-link 
                to="/ranking" 
                class="flex items-center gap-2 px-4 py-2 hover:bg-amber-500/10 rounded-2xl transition-all text-amber-500/70 hover:text-amber-400 group" 
                title="积分排行榜"
              >
                <Trophy class="w-5 h-5 group-hover:scale-110 transition-transform" />
                <span class="text-[10px] font-black uppercase tracking-widest hidden md:block">全球排名</span>
              </router-link>
              <router-link to="/feedbacks" class="p-3 hover:bg-white/5 rounded-2xl transition-all text-slate-400 hover:text-white" title="消息中心">
                <MessageSquare class="w-5 h-5" />
              </router-link>
              <router-link to="/profile" class="p-3 hover:bg-white/5 rounded-2xl transition-all text-slate-400 hover:text-white" title="实验室档案">
                <Settings class="w-5 h-5" />
              </router-link>
              <router-link to="/reactions" class="p-3 hover:bg-blue-500/10 rounded-2xl transition-all text-blue-500/70 hover:text-blue-400" title="反应数据库">
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

             <!-- Pending Feedback Alert -->
             <div v-if="pendingFeedbacks.length > 0" class="hidden xl:flex items-center gap-4 px-6 py-3 bg-amber-500/10 border border-amber-500/20 rounded-[20px] animate-in slide-in-from-right-10 duration-500">
                <div class="flex flex-col">
                  <span class="text-[8px] font-black text-amber-600 dark:text-amber-500 uppercase tracking-widest">Pending_Comm</span>
                  <span class="text-[10px] font-bold text-slate-600 dark:text-slate-400">您有 {{ pendingFeedbacks.length }} 条待处理反馈</span>
                </div>
                <button 
                  @click="handleQuickWithdraw(pendingFeedbacks[0].id)"
                  class="p-2 hover:bg-amber-500/20 text-amber-600 rounded-lg transition-all"
                  title="撤回最新一条反馈"
                >
                  <Trash2 class="w-4 h-4" />
                </button>
             </div>

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

        <!-- Experimental Nodes (Room List Table) -->
        <div class="bg-white/80 dark:bg-[#121216]/60 backdrop-blur-xl border border-slate-200 dark:border-white/10 rounded-[32px] overflow-hidden">
          <div class="overflow-x-auto">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr class="border-b border-slate-200 dark:border-white/5 bg-slate-50/50 dark:bg-white/[0.02]">
                  <th class="px-8 py-5 text-[10px] font-black text-slate-500 uppercase tracking-widest">Experiment_Status</th>
                  <th class="px-6 py-5 text-[10px] font-black text-slate-500 uppercase tracking-widest">Node_Identifier</th>
                  <th class="px-6 py-5 text-[10px] font-black text-slate-500 uppercase tracking-widest">Protocol_Type</th>
                  <th class="px-6 py-5 text-[10px] font-black text-slate-500 uppercase tracking-widest">Participants</th>
                  <th class="px-8 py-5 text-[10px] font-black text-slate-500 uppercase tracking-widest text-right">Access_Control</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="rooms.length === 0">
                  <td colspan="5" class="py-32 text-center text-slate-400 dark:text-slate-600">
                    <div class="flex flex-col items-center justify-center">
                      <div class="w-20 h-20 bg-slate-100 dark:bg-white/5 rounded-full flex items-center justify-center mb-4">
                        <Info class="w-8 h-8 opacity-20" />
                      </div>
                      <p class="text-xl font-black tracking-tight uppercase">No_Active_Nodes</p>
                      <p class="text-[10px] font-mono mt-1 opacity-50 uppercase">等待核心激活...</p>
                    </div>
                  </td>
                </tr>
                <tr 
                  v-for="room in rooms" 
                  :key="room.id"
                  class="group border-b border-slate-200 dark:border-white/5 hover:bg-blue-500/[0.02] transition-colors"
                >
                  <td class="px-8 py-6">
                    <div class="flex items-center gap-3">
                      <div :class="cn(
                        'w-2 h-2 rounded-full',
                        room.status === 'waiting' ? 'bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]' : 
                        room.status === 'playing' ? 'bg-amber-500 shadow-[0_0_8px_rgba(245,158,11,0.5)] animate-pulse' : 
                        'bg-slate-500'
                      )"></div>
                      <span :class="cn(
                        'text-[10px] font-black uppercase tracking-widest',
                        room.status === 'waiting' ? 'text-emerald-500' : 
                        room.status === 'playing' ? 'text-amber-500' : 
                        'text-slate-500'
                      )">
                        {{ room.status === 'waiting' ? 'Ready' : room.status === 'playing' ? 'Active' : 'Closed' }}
                      </span>
                    </div>
                  </td>
                  <td class="px-6 py-6">
                    <div class="flex flex-col">
                      <span class="text-sm font-black text-slate-900 dark:text-white group-hover:text-blue-500 transition-colors">
                        {{ room.name }}
                      </span>
                      <span class="text-[9px] font-mono text-slate-400 dark:text-slate-600 mt-1 uppercase tracking-tighter">
                        ID: {{ room.id }}
                      </span>
                    </div>
                  </td>
                  <td class="px-6 py-6">
                    <div class="flex items-center gap-2">
                      <div v-if="room.is_points_mode" class="px-2 py-0.5 bg-amber-500/10 border border-amber-500/20 rounded-md text-amber-500 text-[8px] font-black uppercase">
                        Competitive
                      </div>
                      <div v-else class="px-2 py-0.5 bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-md text-slate-500 text-[8px] font-black uppercase">
                        Standard
                      </div>
                    </div>
                  </td>
                  <td class="px-6 py-6">
                    <div class="flex items-center gap-2">
                      <div class="flex -space-x-2">
                        <div v-for="i in Math.min(3, room.players?.length || 0)" :key="i" class="w-6 h-6 rounded-lg bg-slate-200 dark:bg-white/10 border-2 border-white dark:border-[#121216] flex items-center justify-center text-[10px] text-slate-600 font-bold">
                          {{ i === 1 ? (room.host_username?.[0]?.toUpperCase() || 'H') : 'R' }}
                        </div>
                        <div v-if="(room.players?.length || 0) > 3" class="w-6 h-6 rounded-lg bg-blue-500 border-2 border-white dark:border-[#121216] flex items-center justify-center text-[8px] text-white font-black">
                          +{{ room.players.length - 3 }}
                        </div>
                      </div>
                      <span class="text-[11px] font-black text-slate-900 dark:text-white ml-1">
                        {{ room.players?.length || 0 }} <span class="text-slate-400 dark:text-slate-600 font-normal">/ {{ room.max_players }}</span>
                      </span>
                    </div>
                  </td>
                  <td class="px-8 py-6 text-right">
                    <template v-if="room.status === 'waiting' && (room.players?.length || 0) < room.max_players">
                      <button 
                        @click="handleJoinRoom(room.id)" 
                        class="px-6 py-2.5 bg-blue-600 hover:bg-blue-500 text-white rounded-[14px] text-[10px] font-black uppercase tracking-widest transition-all hover:scale-105 active:scale-95 shadow-lg shadow-blue-500/20"
                      >
                        Join_Node
                      </button>
                    </template>
                    <template v-else>
                      <button 
                        @click="handleJoinRoom(room.id)"
                        class="px-6 py-2.5 bg-slate-100 dark:bg-white/5 hover:bg-white/10 text-slate-900 dark:text-white border border-slate-200 dark:border-white/10 rounded-[14px] text-[10px] font-black uppercase tracking-widest transition-all"
                      >
                        {{ room.status === 'playing' ? 'Spectate' : 'View' }}
                      </button>
                    </template>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
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
            &copy; 2026 MENDELEEF PROTCOL (PRODUCTION). ALL RIGHTS RESERVED.
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
               <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest">实验空间命名 (可选)</label>
               <span class="text-[9px] text-blue-500/40">IDENTIFIER_MENDELEEF</span>
            </div>
            <input
              v-model="roomName"
              type="text"
              autofocus
              placeholder="默认随机分配名称..."
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
               <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest">高级配置</label>
               <span class="text-[9px] text-blue-500/40">ADVANCED_PROTOCOL</span>
            </div>
            <div class="flex items-center gap-4 p-5 bg-white/5 border border-white/5 rounded-3xl group/toggle cursor-pointer transition-all hover:bg-white/10" @click="isPointsMode = !isPointsMode">
              <div :class="cn(
                'w-10 h-6 rounded-full relative transition-colors duration-300',
                isPointsMode ? 'bg-blue-600' : 'bg-slate-700'
              )">
                <div :class="cn(
                  'absolute top-1 left-1 w-4 h-4 bg-white rounded-full transition-transform duration-300',
                  isPointsMode ? 'translate-x-4' : 'translate-x-0'
                )"></div>
              </div>
              <div class="flex flex-col">
                <span :class="cn('text-[10px] font-black uppercase tracking-wider', isPointsMode ? 'text-blue-400' : 'text-slate-400')">
                  积分竞技模式
                </span>
                <span class="text-[9px] text-slate-500 mt-0.5">
                  开启后强制使用默认牌组，胜负将影响全球排名
                </span>
              </div>
            </div>
          </div>

          <div v-if="!isPointsMode" class="space-y-4">
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
