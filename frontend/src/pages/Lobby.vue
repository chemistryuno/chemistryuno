<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { gameAPI, authAPI, commonAPI, friendAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import websocket from '../utils/websocket'
import { Beaker, Plus, Shield, LogOut, Settings, Play, X, Loader2, Database, MessageCircle, Trophy, Megaphone, Menu } from 'lucide-vue-next'
import { cn } from '../utils/cn'
import ChatBox from '../components/ChatBox.vue'

const props = defineProps<{
  // user props can be added if we pass from App.vue
}>()

const router = useRouter()
const { showAlert, showConfirm } = useDialog()
const user = ref<any>({})
try {
  const userData = JSON.parse(localStorage.getItem('user') || '{}')
  // 兼容旧版本的 id 字段
  if (userData.id && !userData.uid) {
    userData.uid = userData.id
  }
  user.value = userData
} catch (e) {
  console.error('Failed to parse user in Lobby:', e)
}

const friendsList = ref<any[]>([])

const loadFriends = async () => {
  try {
    const res = await friendAPI.getFriends()
    friendsList.value = res.data
  } catch (err) {
    console.error(err)
  }
}

const rooms = ref<any[]>([])
const decks = ref<any[]>([])
const pendingFeedbacks = ref<any[]>([])
const persistentAnnouncements = ref<any[]>([])
const showCreateModal = ref(false)
const showAIArenaModal = ref(false)
const showDeckDetailModal = ref(false)
const showAccessKeyModal = ref(false)
const createdRoomInfo = ref<any>(null)
const selectedDeckConfig = ref<any>(null)
const roomName = ref('')
const maxPlayers = ref(4)
const deckID = ref(0)
const isPointsMode = ref(false)
const isPrivate = ref(false)

const pveDifficulty = ref(50)
const aiCount = ref(1)
const customAccessKey = ref('') // 自定义访问密钥
const loading = ref(false)
const currentTime = ref(new Date())
const onlineCount = ref(0)
const isMobileMenuOpen = ref(false)
const appVersion = ref('V1.2.1 Mendeleef') // 默认值

// 获取当前域名用于生成分享链接
const currentOrigin = window.location.origin

const activeRoom = computed(() => {
  return rooms.value.find(r => 
    (r.status === 'playing' || r.status === 'waiting') && 
    r.players && 
    r.players.includes(Number(user.value.uid))
  )
})

const loadDecks = async () => {
  try {
    const res = await gameAPI.getMyDecks()
    const allDecks = res.data || []
    // 排序：全局优先
    allDecks.sort((a: any, b: any) => {
      if (a.is_global && !b.is_global) return -1
      if (!a.is_global && b.is_global) return 1
      return 0
    })
    decks.value = allDecks
    if (decks.value.length > 0) {
      // 默认选择全局牌组，如果没有则选择第一个
      const globalDeck = decks.value.find((d: any) => d.is_global)
      deckID.value = globalDeck ? globalDeck.id : decks.value[0].id
    }
  } catch (e) {
    console.error(e)
  }
}

let roomInterval: any
let timeInterval: any

const handleOnlineCountUpdate = (msg: any) => {
  onlineCount.value = msg.data || 0
}

const handleSystemAnnouncement = (msg: any) => {
  const ann = msg.data
  if (ann && ann.is_persistent) {
    const exists = persistentAnnouncements.value.some(a => a.id === ann.id)
    if (!exists) {
      persistentAnnouncements.value.unshift(ann)
    } else {
      const idx = persistentAnnouncements.value.findIndex(a => a.id === ann.id)
      persistentAnnouncements.value[idx] = ann
    }
  }
}

onMounted(() => {
  loadRooms()
  loadDecks()
  loadPendingFeedbacks()
  loadPersistentAnnouncements()
  loadFriends()
  loadVersion()
  websocket.connect()
  websocket.on('online_count', handleOnlineCountUpdate)
  websocket.on('system_announcement', handleSystemAnnouncement)

  roomInterval = setInterval(loadRooms, 3000)
  timeInterval = setInterval(() => {
    currentTime.value = new Date()
  }, 1000)
})

onUnmounted(() => {
  if (roomInterval) clearInterval(roomInterval)
  if (timeInterval) clearInterval(timeInterval)
  websocket.off('online_count', handleOnlineCountUpdate)
  websocket.off('system_announcement', handleSystemAnnouncement)
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

const loadVersion = async () => {
  try {
    const res = await authAPI.getVersion()
    if (res.data && res.data.fullVersion) {
      appVersion.value = res.data.fullVersion
    }
  } catch (e) {
    console.error('获取版本信息失败:', e)
  }
}

const loadPersistentAnnouncements = async () => {
  try {
    const res = await commonAPI.getAnnouncements()
    persistentAnnouncements.value = (res.data || []).filter((a: any) => a.is_persistent)
  } catch (e) {
    console.error(e)
  }
}

// 查看牌组配置详情
const handleViewDeckConfig = (deckConfig: any) => {
  if (!deckConfig) return
  selectedDeckConfig.value = deckConfig
  showDeckDetailModal.value = true
}

const handleCreateRoom = async () => {
  loading.value = true

  try {
    const response = await gameAPI.createRoom(
      roomName.value,
      maxPlayers.value,
      deckID.value,
      isPointsMode.value,
      isPrivate.value,
      customAccessKey.value || undefined
    )
    const room = response.data
    // 重置状态
    showCreateModal.value = false
    roomName.value = ''
    customAccessKey.value = ''

    // 如果是私密房间且有访问密钥，显示密钥模态框
    if (isPrivate.value && room.access_key) {
      createdRoomInfo.value = {
        id: room.id,
        access_key: room.access_key,
        name: room.name
      }
      showAccessKeyModal.value = true
      isPrivate.value = false
    } else {
      isPrivate.value = false
      router.push(`/room/${room.id}`)
    }
  } catch (error: any) {
    showAlert(error.response?.data?.error || '创建房间失败', '系统异常')
  } finally {
    loading.value = false
  }
}

const handleCreateAIRoom = async () => {
  loading.value = true
  try {
    const response = await gameAPI.createRoom(
      roomName.value || `AI Challenge - ${pveDifficulty.value}`,
      1 + aiCount.value, // MaxPlayers = 1 (Human) + AI Count
      deckID.value,
      isPointsMode.value,
      true, // AI 房间强制私密/不公开
      undefined,
      true, // IsPvE
      pveDifficulty.value,
      aiCount.value
    )
    const room = response.data
    showAIArenaModal.value = false
    router.push(`/room/${room.id}`)
  } catch (error: any) {
    showAlert(error.response?.data?.error || '创建AI对战失败', '系统异常')
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

const handleLeaveRoom = async (roomId: string) => {
  const ok = await showConfirm('确定要中止并退出当前的实验吗？如果你正在游戏中，系统将按逃跑处理。')
  if (!ok) return

  try {
    await gameAPI.leaveRoom(roomId)
    loadRooms()
    showAlert('已成功从核心节点撤离。', '实验中止')
  } catch (error: any) {
    showAlert(error.response?.data?.error || '退出失败', '链路故障')
  }
}

const handleLogout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  websocket.disconnect()
  router.push('/login')
}

const activeNodesCount = computed(() => rooms.value.filter(r => r.status === 'playing').length)

const copyToClipboard = (text: string) => {
  if (window.navigator && window.navigator.clipboard) {
    window.navigator.clipboard.writeText(text)
    showAlert('复制成功', '成功')
  } else {
    showAlert('当前环境不支持剪贴板操作', '错误')
  }
}
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
    <div class="relative z-10 flex flex-col xl:h-screen min-h-screen xl:overflow-hidden">
      
      <!-- Top Command Bar - 移动端优化 -->
      <header class="h-14 sm:h-16 border-b border-slate-200 dark:border-white/5 bg-white/60 dark:bg-black/40 backdrop-blur-xl sticky top-0 z-50 shrink-0">
        <div class="max-w-[1400px] mx-auto h-full px-4 sm:px-6 flex items-center justify-between">
          <div class="flex items-center gap-3 sm:gap-4">
            <div class="flex items-center gap-2 sm:gap-2.5 group px-2.5 sm:px-3 py-1.5 bg-gradient-to-br from-blue-500/10 to-blue-600/5 border border-blue-500/20 rounded-xl">
              <Beaker class="w-5 h-5 sm:w-6 sm:h-6 text-blue-500 group-hover:rotate-12 transition-transform" />
              <div>
                 <h1 class="text-sm sm:text-base font-black tracking-tighter text-slate-900 dark:text-white leading-none">CHEMISTRY <span class="text-blue-500">UNO</span></h1>
                 <p class="text-xs-mobile text-blue-500/50 font-mono tracking-widest leading-none mt-0.5 sm:mt-1 uppercase">{{ appVersion }}</p>
              </div>
            </div>

            <!-- Status Indicators (Desktop) -->
            <div class="hidden lg:flex items-center gap-4 text-xs-mobile font-mono tracking-widest text-slate-500 border-l border-slate-200 dark:border-white/10 pl-6 uppercase">
              <div class="flex items-center gap-2">
                <div class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></div>
                STABLE
              </div>
              <div class="flex items-center gap-2">
                 UP_TIME: {{ currentTime.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) }}
              </div>
            </div>
          </div>

          <div class="flex items-center gap-2 sm:gap-3">
            <!-- User Identity Chip -->
            <div @click="router.push('/profile')" class="flex items-center gap-2 sm:gap-2.5 pl-1.5 pr-2 sm:pr-3 py-1 bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl hover:bg-slate-200 dark:hover:bg-white/10 transition-all cursor-pointer group touch-feedback">
               <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-slate-200 to-slate-300 dark:from-slate-700 dark:to-slate-800 flex items-center justify-center text-base shadow-inner group-hover:scale-105 transition-transform overflow-hidden">
                 <template v-if="user.avatar && user.avatar.startsWith('data:')">
                    <img :src="user.avatar" class="w-full h-full object-cover" />
                 </template>
                 <template v-else>
                    {{ user.avatar || '🧪' }}
                 </template>
               </div>
               <div class="hidden sm:flex flex-col">
                 <span class="text-xs-mobile font-black text-slate-900 dark:text-white">{{ user.nickname }}</span>
                 <span class="text-[9px] sm:text-[8px] text-slate-500 font-mono uppercase">
                   {{ user.is_admin ? 'Lead' : 'Researcher' }}
                 </span>
               </div>
            </div>

            <!-- Desktop Navigation -->
            <div class="hidden lg:flex items-center gap-0.5">
              <router-link
                to="/ranking"
                class="flex items-center gap-1.5 px-2.5 sm:px-3 py-2 hover:bg-amber-500/10 rounded-xl transition-all text-amber-500/70 hover:text-amber-400 group touch-feedback"
                title="积分排行榜"
              >
                <Trophy class="w-4 h-4 group-hover:scale-110 transition-transform" />
                <span class="text-xs-mobile font-black uppercase tracking-widest hidden md:block">排位</span>
              </router-link>
              <router-link to="/feedbacks" class="p-2 hover:bg-amber-500/10 rounded-xl transition-all text-amber-500/70 hover:text-amber-400 touch-feedback" title="反馈与公告">
                <Megaphone class="w-4 h-4" />
              </router-link>
              <router-link to="/profile" class="p-2 hover:bg-white/5 rounded-xl transition-all text-slate-400 hover:text-white touch-feedback" title="个人主页">
                <Settings class="w-4 h-4" />
              </router-link>
              <router-link to="/data" class="p-2 hover:bg-blue-500/10 rounded-xl transition-all text-blue-500/70 hover:text-blue-400 touch-feedback" title="数据库">
                <Database class="w-4 h-4" />
              </router-link>
              <router-link to="/chat" class="p-2 hover:bg-indigo-500/10 rounded-xl transition-all text-indigo-500/70 hover:text-indigo-400 touch-feedback" title="公共频道">
                <MessageCircle class="w-4 h-4" />
              </router-link>
              <router-link v-if="user.is_admin" to="/admin" class="p-2 hover:bg-yellow-500/10 rounded-xl transition-all text-yellow-500/70 hover:text-yellow-400 touch-feedback" title="管理面板">
                <Shield class="w-4 h-4" />
              </router-link>
              <div class="w-px h-5 bg-white/10 mx-1"></div>
              <button @click="handleLogout" class="p-2 hover:bg-red-500/10 rounded-xl transition-all text-red-500/70 hover:text-red-400 touch-feedback" title="退出登录">
                <LogOut class="w-4 h-4" />
              </button>
            </div>

            <!-- Mobile Menu Toggle -->
            <button
              @click="isMobileMenuOpen = !isMobileMenuOpen"
              class="lg:hidden p-2.5 bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl text-slate-500 touch-feedback transition-all"
            >
              <Menu v-if="!isMobileMenuOpen" class="w-5 h-5" />
              <X v-else class="w-5 h-5" />
            </button>
          </div>
        </div>
      </header>

      <!-- Mobile Menu Overlay -->
      <transition
        enter-active-class="transition duration-200 ease-out"
        enter-from-class="opacity-0 -translate-y-4"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition duration-150 ease-in"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 -translate-y-4"
      >
        <div v-if="isMobileMenuOpen" class="lg:hidden fixed inset-0 z-[45] pt-16 bg-white/95 dark:bg-slate-900/95 backdrop-blur-xl">
          <div class="p-6 space-y-4">
            <div class="grid grid-cols-2 gap-3">
              <router-link @click="isMobileMenuOpen = false" to="/ranking" class="flex flex-col items-center justify-center p-4 bg-slate-100 dark:bg-white/5 rounded-2xl border border-slate-200 dark:border-white/10">
                <Trophy class="w-6 h-6 text-amber-500 mb-2" />
                <span class="text-xs font-black uppercase tracking-widest">排位榜单</span>
              </router-link>
              <router-link @click="isMobileMenuOpen = false" to="/profile" class="flex flex-col items-center justify-center p-4 bg-slate-100 dark:bg-white/5 rounded-2xl border border-slate-200 dark:border-white/10">
                <Settings class="w-6 h-6 text-slate-500 mb-2" />
                <span class="text-xs font-black uppercase tracking-widest">个人主页</span>
              </router-link>
              <router-link @click="isMobileMenuOpen = false" to="/data" class="flex flex-col items-center justify-center p-4 bg-slate-100 dark:bg-white/5 rounded-2xl border border-slate-200 dark:border-white/10">
                <Database class="w-6 h-6 text-blue-500 mb-2" />
                <span class="text-xs font-black uppercase tracking-widest">物质百科</span>
              </router-link>
              <router-link @click="isMobileMenuOpen = false" to="/chat" class="flex flex-col items-center justify-center p-4 bg-slate-100 dark:bg-white/5 rounded-2xl border border-slate-200 dark:border-white/10">
                <MessageCircle class="w-6 h-6 text-indigo-500 mb-2" />
                <span class="text-xs font-black uppercase tracking-widest">公共频道</span>
              </router-link>
              <router-link @click="isMobileMenuOpen = false" to="/feedbacks" class="flex flex-col items-center justify-center p-4 bg-slate-100 dark:bg-white/5 rounded-2xl border border-slate-200 dark:border-white/10">
                <Megaphone class="w-6 h-6 text-emerald-500 mb-2" />
                <span class="text-xs font-black uppercase tracking-widest">反馈中心</span>
              </router-link>
              <router-link v-if="user.is_admin" @click="isMobileMenuOpen = false" to="/admin" class="flex flex-col items-center justify-center p-4 bg-slate-100 dark:bg-white/5 rounded-2xl border border-slate-200 dark:border-white/10">
                <Shield class="w-6 h-6 text-yellow-500 mb-2" />
                <span class="text-xs font-black uppercase tracking-widest">管理面板</span>
              </router-link>
            </div>
            
            <button 
              @click="handleLogout" 
              class="w-full flex items-center justify-center gap-3 py-4 bg-red-500/10 hover:bg-red-500/20 text-red-500 rounded-2xl border border-red-500/20 transition-all font-black uppercase tracking-[0.2em] text-sm"
            >
              <LogOut class="w-5 h-5" />
              中止实验并注销
            </button>
          </div>
        </div>
      </transition>

      <main class="flex-1 max-w-[1400px] mx-auto w-full px-4 sm:px-5 py-4 flex flex-col min-h-0">
        <!-- Welcome & Global Actions -->
        <div class="flex flex-col lg:flex-row lg:items-end justify-between mb-4 gap-4 shrink-0">
          <div class="space-y-1.5 font-bold">
            <div class="inline-flex items-center gap-1.5 px-2 py-0.5 bg-blue-500/10 border border-blue-500/20 rounded-lg">
              <span class="w-1 h-1 bg-blue-500 rounded-full animate-ping"></span>
              <span class="text-[8px] font-black text-blue-500 uppercase tracking-widest">Research_Lobby</span>
            </div>
            <h2 class="text-xl font-black text-slate-800 dark:text-white tracking-tighter leading-none">
              试验场枢纽
            </h2>
            <p class="text-[10px] text-slate-500 font-medium max-w-md leading-none">当前有 <span class="text-blue-500 font-black">{{ onlineCount }}</span> 名研究员在线进行博弈。</p>
          </div>

          <div class="flex items-center gap-3">
             <div class="hidden xl:flex items-center gap-3 px-3 py-2 bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/5 rounded-xl">
               <div class="text-center min-w-[60px]">
                 <p class="text-[7px] text-slate-400 uppercase font-black tracking-widest mb-0.5">Online_Staff</p>
                 <p class="text-sm font-black text-slate-900 dark:text-white font-mono leading-none">{{ onlineCount }}</p>
               </div>
               <div class="w-px h-4 bg-slate-200 dark:bg-white/5"></div>
               <div class="text-center min-w-[60px]">
                 <p class="text-[7px] text-slate-400 uppercase font-black tracking-widest mb-0.5">Core_Nodes</p>
                 <p class="text-sm font-black text-blue-600 dark:text-blue-400 font-mono leading-none">{{ activeNodesCount }}</p>
               </div>
             </div>

            <button 
              @click="showCreateModal = true" 
              class="group relative flex items-center gap-2 bg-blue-600 hover:bg-blue-500 px-4 py-2.5 rounded-xl font-black text-white shadow-lg shadow-blue-500/10 transition-all active:scale-95 overflow-hidden"
            >
              <Plus class="w-3.5 h-3.5 group-hover:rotate-90 transition-transform duration-500" />
              <span class="uppercase tracking-widest text-[9px]">开启实验</span>
            </button>

            <button 
              @click="showAIArenaModal = true; isPointsMode = true; isPrivate = true" 
              class="group relative flex items-center gap-2 bg-purple-600 hover:bg-purple-500 px-4 py-2.5 rounded-xl font-black text-white shadow-lg shadow-purple-500/10 transition-all active:scale-95 overflow-hidden ml-2"
            >
              <div class="relative">
                <Beaker class="w-3.5 h-3.5" />
                <div class="absolute -top-1 -right-1 w-2 h-2 bg-yellow-400 rounded-full animate-pulse border-2 border-purple-600"></div>
              </div>
              <span class="uppercase tracking-widest text-[9px]">AI 竞技场</span>
            </button>
          </div>
        </div>

        <!-- Main Layout Grid -->
        <div class="grid grid-cols-1 xl:grid-cols-12 gap-6 items-stretch flex-1 min-h-0">
          <!-- Left Column: Notifications & Room List -->
          <div class="xl:col-span-9 space-y-6 xl:overflow-y-auto xl:pr-3 custom-scrollbar min-h-0">
            <!-- Persistent Announcements -->
            <div v-if="persistentAnnouncements.length > 0" class="space-y-3 animate-in fade-in duration-700">
               <div v-for="ann in persistentAnnouncements" :key="ann.id" 
                    :class="cn(
                      'relative overflow-hidden p-5 rounded-2xl border transition-all hover:shadow-md',
                      ann.type === 'emergency' ? 'bg-red-500/5 border-red-500/20 shadow-red-500/5' : 
                      ann.type === 'maintenance' ? 'bg-amber-500/5 border-amber-500/20 shadow-amber-500/5' : 
                      'bg-blue-500/5 border-blue-500/20 shadow-blue-500/5'
                    )">
                  <div class="relative z-10 flex flex-col md:flex-row md:items-center gap-4">
                    <div :class="cn(
                       'w-10 h-10 rounded-xl flex items-center justify-center shrink-0 border',
                       ann.type === 'emergency' ? 'bg-red-500/10 text-red-500 border-red-500/20' : 
                       ann.type === 'maintenance' ? 'bg-amber-500/10 text-amber-500 border-amber-500/20' : 
                       'bg-blue-500/10 text-blue-500 border-blue-500/20'
                    )">
                      <Megaphone class="w-5 h-5" />
                    </div>
                    <div class="flex-1">
                       <div class="flex items-center gap-2 mb-0.5">
                          <span :class="cn(
                            'text-[9px] font-black uppercase tracking-widest',
                            ann.type === 'emergency' ? 'text-red-500' : 
                            ann.type === 'maintenance' ? 'text-amber-500' : 
                            'text-blue-500'
                          )">
                            {{ ann.type }}
                          </span>
                       </div>
                       <h3 class="text-base font-black text-slate-900 dark:text-white mb-1" v-if="ann.title">{{ ann.title }}</h3>
                       <p class="text-xs font-medium text-slate-600 dark:text-slate-400 leading-relaxed">{{ ann.content }}</p>
                    </div>
                  </div>
               </div>
            </div>

            <!-- Rejoin Banner -->
            <div v-if="activeRoom" class="p-5 bg-blue-600 rounded-2xl shadow-xl shadow-blue-500/20 flex items-center justify-between group overflow-hidden relative animate-in slide-in-from-top-4 duration-500">
               <div class="flex items-center gap-5">
                  <div class="w-12 h-12 bg-white/20 rounded-xl flex items-center justify-center animate-[pulse_2s_infinite]">
                     <Beaker class="w-6 h-6 text-white" />
                  </div>
                  <div class="flex flex-col">
                     <span class="text-[9px] font-black uppercase text-blue-200 tracking-widest leading-none mb-1">
                        {{ activeRoom.status === 'waiting' && activeRoom.countdown > 0 ? `实验启动中: ${activeRoom.countdown}S` : '活跃实验中' }}
                     </span>
                     <h3 class="text-lg font-black text-white uppercase tracking-wider">{{ activeRoom.name }}</h3>
                  </div>
               </div>
               <div class="flex items-center gap-3">
                  <button 
                    @click="handleLeaveRoom(activeRoom.id)"
                    class="px-5 py-3 bg-red-500/10 hover:bg-red-500 text-red-500 hover:text-white rounded-xl text-[11px] font-black uppercase tracking-widest transition-all border border-red-500/20 active:scale-95 flex items-center gap-2"
                  >
                    <X class="w-4 h-4" />
                    中止实验
                  </button>
                  <button 
                    @click="router.push(`/room/${activeRoom.id}`)"
                    class="px-6 py-3 bg-white text-blue-600 rounded-xl text-[11px] font-black uppercase tracking-widest hover:scale-105 hover:bg-blue-50 transition-all shadow-lg active:scale-95 flex items-center gap-2"
                  >
                    <Play class="w-4 h-4 fill-current" />
                    重连
                  </button>
               </div>
            </div>

            <!-- Experimental Nodes (Room List Table) -->
            <div class="bg-white/80 dark:bg-[#121216]/60 backdrop-blur-xl border border-slate-200 dark:border-white/10 rounded-2xl overflow-hidden">
          <div class="overflow-x-auto">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr class="border-b border-slate-200 dark:border-white/5 bg-slate-50/50 dark:bg-white/[0.02]">
                  <th class="px-5 py-3 text-[8px] font-black text-slate-500 uppercase tracking-widest">Status</th>
                  <th class="px-5 py-3 text-[8px] font-black text-slate-500 uppercase tracking-widest">Identifier</th>
                  <th class="px-5 py-3 text-[8px] font-black text-slate-500 uppercase tracking-widest">Config</th>
                  <th class="px-5 py-3 text-[8px] font-black text-slate-500 uppercase tracking-widest">Players</th>
                  <th class="px-5 py-3 text-[8px] font-black text-slate-500 uppercase tracking-widest text-right">Access</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="rooms.length === 0">
                  <td colspan="5" class="py-16 text-center text-slate-400 dark:text-slate-600">
                    <div class="flex flex-col items-center justify-center">
                      <p class="text-sm font-black tracking-tight uppercase">No_Active_Nodes</p>
                      <p class="text-[8px] font-mono mt-1 opacity-50 uppercase tracking-widest">Waiting_for_Core...</p>
                    </div>
                  </td>
                </tr>
                <tr 
                  v-for="room in rooms" 
                  :key="room.id"
                  class="group border-b border-slate-200 dark:border-white/5 hover:bg-blue-500/[0.02] transition-colors"
                >
                  <td class="px-5 py-3">
                    <div class="flex items-center gap-2">
                      <div :class="cn(
                        'w-1.5 h-1.5 rounded-full',
                        room.status === 'waiting' ? (room.countdown > 0 ? 'bg-blue-500 animate-ping' : 'bg-emerald-500 shadow-[0_0_6px_rgba(16,185,129,0.5)]') : 
                        room.status === 'playing' ? 'bg-amber-500 shadow-[0_0_6px_rgba(245,158,11,0.5)] animate-pulse' : 
                        'bg-slate-500'
                      )"></div>
                      <span :class="cn(
                        'text-[9px] font-black uppercase tracking-widest',
                        room.status === 'waiting' ? (room.countdown > 0 ? 'text-blue-500' : 'text-emerald-500') : 
                        room.status === 'playing' ? 'text-amber-500' : 
                        'text-slate-500'
                      )">
                        {{ room.status === 'waiting' ? (room.countdown > 0 ? room.countdown + 'S' : 'Ready') : room.status === 'playing' ? 'Active' : 'Closed' }}
                      </span>
                    </div>
                  </td>
                  <td class="px-5 py-3">
                    <div class="flex flex-col">
                      <span class="text-xs font-black text-slate-900 dark:text-white group-hover:text-blue-500 transition-colors">
                        {{ room.name }}
                      </span>
                      <span class="text-[8px] font-mono text-slate-400 dark:text-slate-600 uppercase tracking-tighter">
                        ID: {{ room.id }}
                      </span>
                    </div>
                  </td>
                  <td class="px-5 py-3">
                    <div class="flex flex-col gap-1">
                      <div class="flex items-center gap-1.5">
                        <div v-if="room.is_points_mode" class="px-1.5 py-0.5 bg-amber-500/10 border border-amber-500/20 rounded text-amber-500 text-[8px] font-black uppercase">
                          Ranked
                        </div>
                        <div v-if="room.deck_config"
                          @click.stop="handleViewDeckConfig(room.deck_config)"
                          class="px-1.5 py-0.5 bg-blue-500/10 border border-blue-500/20 rounded text-blue-500 text-[8px] font-black uppercase cursor-pointer hover:bg-blue-500/20 transition-colors"
                          title="点击查看牌组详情"
                        >
                          {{ room.deck_config.name }}
                        </div>
                      </div>
                    </div>
                  </td>
                  <td class="px-5 py-3">
                    <span class="text-[10px] font-black text-slate-900 dark:text-white font-mono">
                      {{ room.players?.length || 0 }}<span class="text-slate-400">/{{ room.max_players }}</span>
                    </span>
                  </td>
                  <td class="px-5 py-3 text-right">
                    <button 
                      @click="handleJoinRoom(room.id)"
                      class="px-4 py-1.5 border border-slate-200 dark:border-white/10 rounded-lg text-[9px] font-black uppercase tracking-widest transition-all hover:bg-blue-600 hover:text-white"
                    >
                       {{ room.status === 'playing' ? 'Spectate' : 'Enter' }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Right Column: World Chat -->
      <div class="xl:col-span-3 h-full min-h-0 bg-white/40 dark:bg-black/20 rounded-2xl overflow-hidden border border-slate-200 dark:border-white/5">
         <ChatBox title="全域通信频率" placeholder="发送..." maxHeight="100%" class="h-full" />
      </div>
    </div>
  </main>

      <!-- Global Footer Terminal -->
      <footer class="mt-auto border-t border-white/5 bg-black/40 backdrop-blur-md p-4 shrink-0">
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
      <div class="absolute inset-0 bg-slate-900/40 dark:bg-black/80 backdrop-blur-md animate-in fade-in" @click="showCreateModal = false" />
      <div class="relative w-full max-w-lg bg-white dark:bg-[#121216] border border-slate-200 dark:border-white/10 rounded-[40px] shadow-2xl overflow-hidden flex flex-col max-h-[90vh] animate-in fade-in zoom-in slide-in-from-bottom-10 duration-500">
         <!-- Modal Header -->
         <div class="px-6 py-5 border-b border-slate-100 dark:border-white/5 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 bg-blue-500/10 border border-blue-500/20 rounded-xl flex items-center justify-center text-blue-500 dark:text-blue-400">
                <Plus class="w-5 h-5" />
              </div>
              <div>
                <h2 class="text-lg font-black text-slate-800 dark:text-white tracking-tight leading-none">开启新实验</h2>
                <p class="text-[9px] text-slate-400 dark:text-slate-500 font-mono uppercase tracking-widest mt-1">Setup_Experiment_Parameters</p>
              </div>
            </div>
            <button 
              @click="showCreateModal = false"
              class="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl transition-colors text-slate-400 hover:text-slate-900 dark:hover:text-white"
            >
              <X class="w-5 h-5" />
            </button>
         </div>

        <form @submit.prevent="handleCreateRoom" class="flex flex-col min-h-0">
          <div class="p-6 space-y-5 overflow-y-auto custom-scrollbar flex-1">
            <div class="space-y-2">
            <div class="flex justify-between items-center px-1">
               <label class="text-[9px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">实验空间命名</label>
               <span class="text-[8px] text-blue-500/40 font-mono">IDENTIFIER</span>
            </div>
            <input
              v-model="roomName"
              type="text"
              autofocus
              placeholder="默认随机分配名称..."
              class="w-full bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/5 text-slate-900 dark:text-white px-4 py-3 rounded-xl focus:ring-1 focus:ring-blue-500/50 focus:border-blue-500/50 outline-none transition-all placeholder:text-slate-300 dark:placeholder:text-slate-700 font-mono text-xs"
            />
          </div>

          <div class="space-y-2">
            <div class="flex justify-between items-center px-1">
               <label class="text-[9px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">研究员容量</label>
               <span class="text-[8px] text-blue-500/40 font-mono">CAPACITY</span>
            </div>
            <div class="grid grid-cols-4 gap-3">
              <button
                v-for="num in [2, 3, 4, 8]"
                :key="num"
                type="button"
                @click="maxPlayers = num"
                :class="cn(
                  'h-10 rounded-xl text-[11px] font-black border transition-all flex items-center justify-center relative group/opt overflow-hidden',
                  maxPlayers === num 
                    ? 'bg-blue-500/10 border-blue-500/50 text-blue-600 dark:text-blue-400 ring-1 ring-blue-500/20 shadow-[0_4px_12px_rgba(59,130,246,0.1)]' 
                    : 'bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/5 text-slate-400 dark:text-slate-600 hover:bg-slate-100 dark:hover:bg-white/10'
                )"
              >
                <span class="relative z-10">{{ num }}P</span>
                <div v-if="maxPlayers === num" class="absolute inset-0 bg-blue-500/5 animate-pulse"></div>
              </button>
            </div>
          </div>

          <div class="space-y-2">
            <div class="flex justify-between items-center px-1">
               <label class="text-[9px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">核心配置</label>
               <span class="text-[8px] text-blue-500/40 font-mono">PROTOCOL</span>
            </div>
            <div class="flex items-center gap-3 p-3 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/5 rounded-xl group/toggle cursor-pointer transition-all hover:bg-slate-100 dark:hover:bg-white/10" @click="isPointsMode = !isPointsMode; if(isPointsMode) isPrivate = false">
              <div :class="cn(
                'w-8 h-4.5 rounded-full relative transition-colors duration-300',
                isPointsMode ? 'bg-blue-600' : 'bg-slate-300 dark:bg-slate-700'
              )">
                <div :class="cn(
                  'absolute top-0.75 left-0.75 w-3 h-3 bg-white rounded-full transition-transform duration-300',
                  isPointsMode ? 'translate-x-3.5' : 'translate-x-0'
                )"></div>
              </div>
              <div class="flex flex-col">
                <span :class="cn('text-[9px] font-black uppercase tracking-wider', isPointsMode ? 'text-blue-600 dark:text-blue-400' : 'text-slate-400 dark:text-slate-400')">
                  积分竞技模式
                </span>
                <span class="text-[8px] text-slate-400 dark:text-slate-500 mt-0.5 leading-tight">
                  胜负将影响全球排名
                </span>
              </div>
            </div>

            <div class="flex items-center gap-3 p-3 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/5 rounded-xl group/toggle cursor-pointer transition-all hover:bg-slate-100 dark:hover:bg-white/10" @click="isPrivate = !isPrivate; if(isPrivate) isPointsMode = false">
              <div :class="cn(
                'w-8 h-4.5 rounded-full relative transition-colors duration-300',
                isPrivate ? 'bg-amber-600' : 'bg-slate-300 dark:bg-slate-700'
              )">
                <div :class="cn(
                  'absolute top-0.75 left-0.75 w-3 h-3 bg-white rounded-full transition-transform duration-300',
                  isPrivate ? 'translate-x-3.5' : 'translate-x-0'
                )"></div>
              </div>
              <div class="flex flex-col">
                <span :class="cn('text-[9px] font-black uppercase tracking-wider', isPrivate ? 'text-amber-600 dark:text-amber-400' : 'text-slate-400 dark:text-slate-400')">
                  私密隐藏频道
                </span>
                <span class="text-[8px] text-slate-400 dark:text-slate-500 mt-0.5 leading-tight">
                  不显示在大厅，仅能通过二维码/链接加入
                </span>
              </div>
            </div>
          </div>

          <!-- 私密房间密钥输入框 -->
          <div v-if="isPrivate" class="space-y-2 animate-in slide-in-from-top-2 fade-in duration-300">
            <div class="flex justify-between items-center px-1">
               <label class="text-[9px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">访问密钥</label>
               <span class="text-[8px] text-amber-500/40 font-mono">ACCESS_KEY</span>
            </div>
            <input
              v-model="customAccessKey"
              type="text"
              placeholder="留空自动生成8位密钥..."
              maxlength="20"
              class="w-full bg-amber-50 dark:bg-amber-500/5 border border-amber-200 dark:border-amber-500/20 text-slate-900 dark:text-white px-4 py-3 rounded-xl focus:ring-1 focus:ring-amber-500/50 focus:border-amber-500/50 outline-none transition-all placeholder:text-slate-300 dark:placeholder:text-slate-700 font-mono text-xs"
            />
            <p class="text-[8px] text-amber-600 dark:text-amber-500 px-1 leading-relaxed">
              💡 <span class="font-bold">提示：</span>可自定义4-20位密钥，或留空由系统自动生成
            </p>
          </div>

          <div v-if="!isPointsMode" class="space-y-2">
            <div class="flex justify-between items-center px-1">
               <label class="text-[9px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">选择牌组</label>
               <span class="text-[8px] text-blue-500/40 font-mono">DECK</span>
            </div>
            <div v-if="decks.length === 0" class="p-4 bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl text-center">
              <p class="text-[10px] text-slate-500 dark:text-slate-400 font-mono">暂无可用牌组，将使用默认全局牌组</p>
            </div>
            <div v-else class="space-y-2">
              <button
                v-for="deck in decks"
                :key="deck.id"
                type="button"
                @click="deckID = deck.id"
                :class="cn(
                  'w-full flex items-center gap-3 p-3 rounded-xl border transition-all text-left group/deck',
                  deckID === deck.id
                    ? 'bg-blue-600/5 dark:bg-blue-600/10 border-blue-500/50 shadow-[0_4px_12px_rgba(59,130,246,0.05)]'
                    : 'bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/5 hover:border-slate-300 dark:hover:border-white/10'
                )"
              >
                <div :class="cn(
                  'w-9 h-9 rounded-lg flex items-center justify-center transition-colors',
                  deckID === deck.id ? 'bg-blue-500 text-white' : 'bg-slate-200 dark:bg-white/5 text-slate-400 dark:text-slate-500'
                )">
                   <Beaker class="w-4 h-4" />
                </div>
                <div class="flex-1">
                  <p :class="cn('text-[11px] font-black uppercase tracking-wider', deckID === deck.id ? 'text-blue-600 dark:text-blue-400' : 'text-slate-700 dark:text-white')">
                    {{ deck.name }}
                  </p>
                  <p class="text-[8px] text-slate-400 dark:text-slate-500 mt-0.5 font-mono uppercase tracking-tighter">
                    {{ Object.keys(deck.cards || {}).length }} Elements
                  </p>
                </div>
                <div v-if="deckID === deck.id" class="w-1.5 h-1.5 rounded-full bg-blue-500 animate-pulse mr-1"></div>
              </button>
            </div>
          </div>

          </div>
          <div class="p-6 border-t border-slate-100 dark:border-white/5 bg-slate-50/50 dark:bg-white/[0.02] flex gap-3 shrink-0">
            <button 
              type="button" 
              @click="showCreateModal = false" 
              class="flex-1 h-11 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-500 dark:text-slate-400 font-bold rounded-xl transition-all uppercase tracking-widest text-[10px] border border-slate-200 dark:border-white/5"
            >
              放弃
            </button>
            <button 
              type="submit" 
              :disabled="loading"
              class="flex-1 h-11 bg-blue-600 hover:bg-blue-500 text-white font-black rounded-xl transition-all shadow-[0_8px_20px_rgba(37,99,235,0.2)] active:scale-95 disabled:grayscale flex items-center justify-center gap-2 group/sub relative overflow-hidden"
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

    <!-- AI Arena Modal -->
    <div v-if="showAIArenaModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md animate-in fade-in" @click="showAIArenaModal = false" />
      <div class="relative w-full max-w-lg bg-white dark:bg-[#121216] border border-purple-500/30 rounded-[40px] shadow-2xl shadow-purple-500/10 overflow-hidden flex flex-col max-h-[90vh] animate-in fade-in zoom-in slide-in-from-bottom-10 duration-500">
         <!-- Modal Header -->
         <div class="px-6 py-5 border-b border-slate-100 dark:border-white/5 flex items-center justify-between bg-purple-500/5">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 bg-purple-500/10 border border-purple-500/20 rounded-xl flex items-center justify-center text-purple-600 dark:text-purple-400">
                <Beaker class="w-5 h-5" />
              </div>
              <div>
                <h2 class="text-lg font-black text-slate-800 dark:text-white tracking-tight leading-none">人机竞技场</h2>
                <p class="text-[9px] text-purple-500/60 font-mono uppercase tracking-widest mt-1">AI_Challenge_Mode</p>
              </div>
            </div>
            <button 
              @click="showAIArenaModal = false"
              class="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl transition-colors text-slate-400 hover:text-slate-900 dark:hover:text-white"
            >
              <X class="w-5 h-5" />
            </button>
         </div>

        <form @submit.prevent="handleCreateAIRoom" class="flex flex-col min-h-0">
          <div class="p-6 space-y-6 overflow-y-auto custom-scrollbar flex-1">
            
            <!-- Difficulty Slider -->
            <div class="space-y-4">
              <div class="flex justify-between items-end px-1">
                 <div>
                   <label class="text-[9px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest block mb-1">AI 智能等级</label>
                   <h3 class="text-2xl font-black text-slate-800 dark:text-white">{{ pveDifficulty }}<span class="text-sm text-slate-400 ml-1">%</span></h3>
                 </div>
                 <span class="text-[8px] text-purple-500/40 font-mono">DIFFICULTY</span>
              </div>
              <div class="relative h-6 flex items-center">
                <input 
                  type="range" 
                  v-model.number="pveDifficulty" 
                  min="1" 
                  max="100" 
                  class="w-full h-2 bg-slate-200 dark:bg-white/10 rounded-lg appearance-none cursor-pointer accent-purple-600"
                />
              </div>
              <div class="flex justify-between text-[10px] text-slate-400 font-mono">
                <span>Rookie (1%)</span>
                <span>Grandmaster (100%)</span>
              </div>
            </div>

            <!-- AI Count -->
            <div class="space-y-3">
              <div class="flex justify-between items-center px-1">
                 <label class="text-[9px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">AI 对手数量</label>
                 <span class="text-[8px] text-purple-500/40 font-mono">OPPONENTS</span>
              </div>
              <div class="grid grid-cols-4 gap-3">
                <button
                  v-for="num in [1, 2, 3, 7]"
                  :key="num"
                  type="button"
                  @click="aiCount = num"
                  :class="cn(
                    'h-12 rounded-xl text-sm font-black border transition-all flex flex-col items-center justify-center relative group/opt overflow-hidden',
                    aiCount === num 
                      ? 'bg-purple-500/10 border-purple-500/50 text-purple-600 dark:text-purple-400 ring-1 ring-purple-500/20 shadow-[0_4px_12px_rgba(168,85,247,0.1)]' 
                      : 'bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/5 text-slate-400 dark:text-slate-600 hover:bg-slate-100 dark:hover:bg-white/10'
                  )"
                >
                  <span class="relative z-10">{{ num }} AI</span>
                  <div v-if="aiCount === num" class="absolute inset-0 bg-purple-500/5 animate-pulse"></div>
                </button>
              </div>
            </div>

            <!-- Points Mode Toggle -->
            <div class="flex items-center gap-3 p-4 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/5 rounded-xl group/toggle cursor-pointer transition-all hover:bg-slate-100 dark:hover:bg-white/10" @click="isPointsMode = !isPointsMode">
              <div :class="cn(
                'w-10 h-6 rounded-full relative transition-colors duration-300',
                isPointsMode ? 'bg-purple-600' : 'bg-slate-300 dark:bg-slate-700'
              )">
                <div :class="cn(
                  'absolute top-1 left-1 w-4 h-4 bg-white rounded-full transition-transform duration-300',
                  isPointsMode ? 'translate-x-4' : 'translate-x-0'
                )"></div>
              </div>
              <div class="flex flex-col">
                <span :class="cn('text-[10px] font-black uppercase tracking-wider', isPointsMode ? 'text-purple-600 dark:text-purple-400' : 'text-slate-400 dark:text-slate-400')">
                  积分结算
                </span>
                <span class="text-[9px] text-slate-400 dark:text-slate-500 mt-0.5 leading-tight">
                  难度 >= 50% 时可获得积分奖励，否则仅供练习
                </span>
              </div>
            </div>

          </div>
          <div class="p-6 border-t border-slate-100 dark:border-white/5 bg-slate-50/50 dark:bg-white/[0.02] flex gap-3 shrink-0">
            <button 
              type="button"
              @click="showAIArenaModal = false"
              class="flex-1 px-4 py-3 border border-slate-200 dark:border-white/10 rounded-xl text-[11px] font-black uppercase tracking-widest hover:bg-slate-100 dark:hover:bg-white/5 transition-colors text-slate-500"
            >
              取消
            </button>
            <button 
              type="submit" 
              :disabled="loading"
              class="flex-[2] px-6 py-3 bg-purple-600 hover:bg-purple-500 text-white rounded-xl text-[11px] font-black uppercase tracking-widest shadow-lg shadow-purple-500/20 hover:shadow-purple-500/40 hover:-translate-y-0.5 active:translate-y-0 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              <Loader2 v-if="loading" class="w-4 h-4 animate-spin" />
              <span>开始挑战</span>
            </button>
          </div>
        </form>
      </div>
    </div>
    </div>

    <!-- 牌组详情查看模态框 -->
    <div v-if="showDeckDetailModal && selectedDeckConfig" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-900/40 dark:bg-black/80 backdrop-blur-md animate-in fade-in" @click="showDeckDetailModal = false" />
      <div class="relative w-full max-w-2xl bg-white dark:bg-[#121216] border border-slate-200 dark:border-white/10 rounded-[40px] shadow-2xl overflow-hidden animate-in fade-in zoom-in slide-in-from-bottom-10 duration-500">
         <!-- Modal Header -->
         <div class="px-6 py-5 border-b border-slate-100 dark:border-white/5 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 bg-blue-500/10 border border-blue-500/20 rounded-xl flex items-center justify-center text-blue-500 dark:text-blue-400">
                <Database class="w-5 h-5" />
              </div>
              <div>
                <h2 class="text-lg font-black text-slate-800 dark:text-white tracking-tight leading-none">{{ selectedDeckConfig.name }}</h2>
                <p class="text-[9px] text-slate-400 dark:text-slate-500 font-mono uppercase tracking-widest mt-1">Deck_Configuration_Details</p>
              </div>
            </div>
            <button
              @click="showDeckDetailModal = false"
              class="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl transition-colors text-slate-400 hover:text-slate-900 dark:hover:text-white"
            >
              <X class="w-5 h-5" />
            </button>
         </div>

         <!-- Modal Content -->
         <div class="p-6 max-h-[60vh] overflow-y-auto custom-scrollbar">
            <div class="space-y-4">
              <!-- 基础信息 -->
              <div class="p-4 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl">
                <div class="flex items-center justify-between mb-3">
                  <span class="text-[9px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">基础信息</span>
                  <span class="text-[8px] text-blue-500/40 font-mono">BASIC_INFO</span>
                </div>
                <div class="grid grid-cols-2 gap-3">
                  <div class="p-3 bg-white dark:bg-black/20 rounded-lg">
                    <p class="text-[8px] text-slate-400 mb-1 uppercase tracking-wider">牌组名称</p>
                    <p class="text-[11px] font-black text-slate-900 dark:text-white">{{ selectedDeckConfig.name }}</p>
                  </div>
                  <div class="p-3 bg-white dark:bg-black/20 rounded-lg">
                    <p class="text-[8px] text-slate-400 mb-1 uppercase tracking-wider">元素种类</p>
                    <p class="text-[11px] font-black text-blue-600 dark:text-blue-400">{{ Object.keys(selectedDeckConfig.cards || {}).length }} 种</p>
                  </div>
                  <div class="p-3 bg-white dark:bg-black/20 rounded-lg">
                    <p class="text-[8px] text-slate-400 mb-1 uppercase tracking-wider">总卡牌数</p>
                    <p class="text-[11px] font-black text-slate-900 dark:text-white">{{ (Object.values(selectedDeckConfig.cards || {}) as number[]).reduce((a, b) => a + b, 0) }} 张</p>
                  </div>
                  <div class="p-3 bg-white dark:bg-black/20 rounded-lg">
                    <p class="text-[8px] text-slate-400 mb-1 uppercase tracking-wider">起始手牌</p>
                    <p class="text-[11px] font-black text-slate-900 dark:text-white">{{ selectedDeckConfig.initial_cards || 7 }} 张</p>
                  </div>
                </div>
              </div>

              <!-- 卡牌列表 -->
              <div class="p-4 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl">
                <div class="flex items-center justify-between mb-3">
                  <span class="text-[9px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">卡牌配置</span>
                  <span class="text-[8px] text-blue-500/40 font-mono">CARD_LIST</span>
                </div>
                <div class="grid grid-cols-2 sm:grid-cols-3 gap-2 max-h-64 overflow-y-auto custom-scrollbar pr-2">
                  <div
                    v-for="(count, formula) in selectedDeckConfig.cards"
                    :key="formula"
                    class="p-2.5 bg-white dark:bg-black/20 rounded-lg border border-slate-200 dark:border-white/10 hover:border-blue-500/50 transition-colors"
                  >
                    <div class="flex items-center justify-between">
                      <span class="text-[10px] font-black text-slate-900 dark:text-white font-mono" v-html="String(formula).replace(/(\d+)/g, '<sub>$1</sub>')"></span>
                      <span class="text-[9px] font-black text-blue-600 dark:text-blue-400">×{{ count }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
         </div>

         <!-- Modal Footer -->
         <div class="px-6 py-4 border-t border-slate-100 dark:border-white/5 flex justify-end">
            <button
              @click="showDeckDetailModal = false"
              class="px-5 py-2.5 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-700 dark:text-slate-300 font-bold rounded-xl transition-all uppercase tracking-widest text-[10px] border border-slate-200 dark:border-white/5"
            >
              关闭
            </button>
         </div>
      </div>
    </div>

    <!-- 私密房间密钥模态框 -->
    <div v-if="showAccessKeyModal && createdRoomInfo" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-900/40 dark:bg-black/80 backdrop-blur-md animate-in fade-in" @click="showAccessKeyModal = false" />
      <div class="relative w-full max-w-md bg-white dark:bg-[#121216] border border-slate-200 dark:border-white/10 rounded-[40px] shadow-2xl overflow-hidden animate-in fade-in zoom-in slide-in-from-bottom-10 duration-500">
         <!-- Modal Header -->
         <div class="px-6 py-5 border-b border-slate-100 dark:border-white/5 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 bg-green-500/10 border border-green-500/20 rounded-xl flex items-center justify-center text-green-500 dark:text-green-400">
                <Shield class="w-5 h-5" />
              </div>
              <div>
                <h2 class="text-lg font-black text-slate-800 dark:text-white tracking-tight leading-none">私密房间已创建</h2>
                <p class="text-[9px] text-slate-400 dark:text-slate-500 font-mono uppercase tracking-widest mt-1">Private_Room_Created</p>
              </div>
            </div>
            <button
              @click="showAccessKeyModal = false; router.push(`/room/${createdRoomInfo.id}`)"
              class="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl transition-colors text-slate-400 hover:text-slate-900 dark:hover:text-white"
            >
              <X class="w-5 h-5" />
            </button>
         </div>

         <!-- Modal Body -->
         <div class="p-6 space-y-6">
            <div class="bg-green-500/5 border border-green-500/20 rounded-2xl p-5">
              <p class="text-xs text-slate-600 dark:text-slate-400 mb-4 leading-relaxed">
                ✅ 房间已成功创建！以下是访问密钥和分享链接，请妥善保管。
              </p>

              <!-- 访问密钥 -->
              <div class="mb-4">
                <label class="text-[9px] font-black text-slate-500 dark:text-slate-500 uppercase tracking-[0.2em] block mb-2">Access Key / 访问密钥</label>
                <div class="flex gap-2">
                  <div class="flex-1 bg-white dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-3 font-mono text-lg font-black text-green-600 dark:text-green-400 tracking-widest text-center select-all">
                    {{ createdRoomInfo.access_key }}
                  </div>
                  <button
                    @click="copyToClipboard(createdRoomInfo.access_key)"
                    class="px-4 py-3 bg-green-500/10 hover:bg-green-500/20 border border-green-500/20 text-green-600 dark:text-green-400 rounded-xl transition-all font-black text-[10px] uppercase tracking-widest"
                    title="复制密钥"
                  >
                    复制
                  </button>
                </div>
              </div>

              <!-- 分享链接 -->
              <div>
                <label class="text-[9px] font-black text-slate-500 dark:text-slate-500 uppercase tracking-[0.2em] block mb-2">Share Link / 分享链接</label>
                <div class="flex gap-2">
                  <div class="flex-1 bg-white dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-3 font-mono text-[10px] text-slate-600 dark:text-slate-400 truncate select-all">
                    {{ currentOrigin }}/room/{{ createdRoomInfo.id }}?key={{ createdRoomInfo.access_key }}
                  </div>
                  <button
                    @click="copyToClipboard(`${currentOrigin}/room/${createdRoomInfo.id}?key=${createdRoomInfo.access_key}`)"
                    class="px-4 py-3 bg-blue-500/10 hover:bg-blue-500/20 border border-blue-500/20 text-blue-600 dark:text-blue-400 rounded-xl transition-all font-black text-[10px] uppercase tracking-widest"
                    title="复制链接"
                  >
                    复制
                  </button>
                </div>
              </div>
            </div>

            <div class="bg-amber-500/5 border border-amber-500/20 rounded-xl p-4">
              <p class="text-[10px] text-amber-700 dark:text-amber-400 leading-relaxed">
                ⚠️ <span class="font-black">提示：</span>其他玩家需要使用访问密钥或完整链接才能加入此私密房间。
              </p>
            </div>
         </div>

         <!-- Modal Footer -->
         <div class="px-6 py-4 border-t border-slate-100 dark:border-white/5 flex gap-3">
            <button
              @click="showAccessKeyModal = false"
              class="flex-1 px-5 py-2.5 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-700 dark:text-slate-300 font-bold rounded-xl transition-all uppercase tracking-widest text-[10px] border border-slate-200 dark:border-white/5"
            >
              稍后进入
            </button>
            <button
              @click="showAccessKeyModal = false; router.push(`/room/${createdRoomInfo.id}`)"
              class="flex-1 px-5 py-2.5 bg-green-600 hover:bg-green-500 text-white font-black rounded-xl transition-all shadow-lg shadow-green-500/20 active:scale-95 uppercase tracking-widest text-[10px]"
            >
              立即进入房间
            </button>
         </div>
      </div>
    </div>
  </div>
</template>
