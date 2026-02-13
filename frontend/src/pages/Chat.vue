<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { friendAPI, authAPI, gameAPI } from '../utils/api'
import websocket from '../utils/websocket'
import {
  MessageCircle, User, UserPlus, Send,
  Search, ArrowLeft, X, Check,
  Trash2, ShieldAlert, FlaskConical, Globe, Loader2, Users, Trophy
} from 'lucide-vue-next'
import { cn } from '../utils/cn'
import { useDialog } from '../utils/dialog'

const router = useRouter()
const route = useRoute()
const { showAlert, showConfirm, showPrompt } = useDialog()

let initialUser: any = {}
try {
  initialUser = JSON.parse(localStorage.getItem('user') || '{}')
  // 兼容旧版本的 id 字段
  if (initialUser.id && !initialUser.uid) {
    initialUser.uid = initialUser.id
  }
} catch (e) {
  console.error('Failed to parse user in Chat:', e)
}
const currentUser = ref<any>(initialUser)

// 状态管理
const friends = ref<any[]>([])
const pendingRequests = ref<any[]>([])
const globalSearchResults = ref<any[]>([])
const loading = ref(true)
const searchLoading = ref(false)
const activeChat = ref<any>(null) // 当前选中的好友
const messages = ref<Record<number, any[]>>({}) // 缓存各好友的聊天记录
const newMessage = ref('')
const searchTerm = ref('')
const searchQuery = ref('') // 模态框搜索词
const scrollContainer = ref<HTMLElement | null>(null)

const showSearchModal = ref(false)
const showRequestsModal = ref(false)

// 房间状态缓存
const roomStatusCache = ref<Record<string, { status: string, checkedAt: number }>>({})

// 检查房间状态
const checkRoomStatus = async (roomId: string) => {
  // 检查缓存（缓存5秒）
  const cached = roomStatusCache.value[roomId]
  if (cached && Date.now() - cached.checkedAt < 5000) {
    return cached.status
  }

  try {
    const res = await gameAPI.checkRoomStatus(roomId)
    const status = res.data.exists ? res.data.status : 'closed'
    roomStatusCache.value[roomId] = {
      status,
      checkedAt: Date.now()
    }
    return status
  } catch (err) {
    // 如果出错，假设房间已关闭
    return 'closed'
  }
}

watch(showRequestsModal, (val) => {
  if (val) fetchRequests()
})

// 过滤后的好友列表
const filteredFriends = computed(() => {
  if (!searchTerm.value) return friends.value
  return friends.value.filter(f => 
    f.username.toLowerCase().includes(searchTerm.value.toLowerCase()) ||
    String(f.uid).includes(searchTerm.value)
  )
})

// 监听模态框搜索词变化
let searchTimeout: any = null
watch(searchQuery, (newVal) => {
  if (searchTimeout) clearTimeout(searchTimeout)
  if (!newVal.trim()) {
    globalSearchResults.value = []
    return
  }
  
  searchLoading.value = true
  searchTimeout = setTimeout(async () => {
    try {
      const res = await authAPI.searchUsers(newVal)
      globalSearchResults.value = (res.data || []).filter((u: any) => {
        return Number(u.uid) !== Number(currentUser.value.uid)
      })
    } catch (err) {
      console.error('全局搜索失败', err)
      globalSearchResults.value = []
    } finally {
      searchLoading.value = false
    }
  }, 500)
})

const triggerSearch = async () => {
  if (!searchQuery.value.trim()) return
  if (searchTimeout) clearTimeout(searchTimeout)
  
  searchLoading.value = true
  try {
    const res = await authAPI.searchUsers(searchQuery.value)
    globalSearchResults.value = (res.data || []).filter((u: any) => {
      return Number(u.uid) !== Number(currentUser.value.uid)
    })
  } catch (err) {
    console.error('搜索点击执行失败', err)
    globalSearchResults.value = []
  } finally {
    searchLoading.value = false
  }
}

const isFriend = (uid: number) => {
  return friends.value.some(f => Number(f.uid) === Number(uid))
}

const fetchFriends = async () => {
  try {
    const res = await friendAPI.getFriends()
    friends.value = res.data || []
    
    // 如果 URL 中有 uid 参数，自动选择该好友
    const targetUid = route.query.uid
    if (targetUid) {
      const friend = friends.value.find((f: any) => Number(f.uid) === Number(targetUid))
      if (friend) {
        selectChat(friend)
      }
    }
  } catch (err) {
    console.error('获取好友失败', err)
    friends.value = []
  }
}

const fetchRequests = async () => {
  try {
    const res = await friendAPI.getPendingRequests()
    pendingRequests.value = res.data || []
  } catch (err) {
    console.error('获取请求失败', err)
    pendingRequests.value = []
  }
}

const handleRequest = async (id: number, action: 'accept' | 'decline') => {
  try {
    await friendAPI.handleRequest(id, action)
    await fetchRequests()
    await fetchFriends()
    showAlert(action === 'accept' ? '已通过好友请求' : '已拒绝好友请求', '同步完成')
  } catch (err: any) {
    showAlert(err.response?.data?.error || '操作失败', '错误')
  }
}

const sendRequest = async (uid: number) => {
  const message = await showPrompt('请输入申请信息（可选）:', '你好，我想和你一起进行化学实验。', '建立研究连接')
  if (message === null) return // 用户取消
  
  try {
    await friendAPI.sendRequest(uid, message)
    showAlert('研究者连接请求已发出，等待对方同步波段。', '请求已发送')
  } catch (err: any) {
    showAlert(err.response?.data?.error || '请求发送失败', '错误')
  }
}

const deleteFriend = async (uid: number) => {
  const confirmed = await showConfirm('确定要删除这位研究员吗？所有加密通信记录将失效。', '解除关系')
  if (!confirmed) return
  try {
    await friendAPI.deleteFriend(uid)
    if (activeChat.value?.uid === uid) activeChat.value = null
    await fetchFriends()
    showAlert('已解除好友关系', '完成')
  } catch (err: any) {
    showAlert(err.response?.data?.error || '删除失败', '错误')
  }
}

const scrollToBottom = () => {
  nextTick(() => {
    if (scrollContainer.value) {
      scrollContainer.value.scrollTop = scrollContainer.value.scrollHeight
    }
  })
}

const focusSearch = () => {
  console.log('Chat: Opening search modal')
  showSearchModal.value = true
}

const openRequestsModal = () => {
  console.log('Chat: Opening requests modal')
  fetchRequests()
  showRequestsModal.value = true
}

// 选择聊天对象
const selectChat = async (friend: any) => {
  activeChat.value = friend
  if (!messages.value[friend.uid]) {
    messages.value[friend.uid] = []
  }

  // 加载历史消息
  try {
    const res = await authAPI.getPrivateChatHistory(friend.uid, 50)
    const historyMessages = await Promise.all((res.data || []).map(async (m: any) => {
      // 尝试解析游戏邀请消息
      let isGameInvite = false
      let gameInviteData = null

      try {
        const parsed = JSON.parse(m.message)
        if (parsed.type === 'game_invite') {
          isGameInvite = true
          gameInviteData = parsed

          // 检查房间状态
          if (gameInviteData.room_id) {
            gameInviteData.room_status = await checkRoomStatus(gameInviteData.room_id)
          }
        }
      } catch (e) {
        // 不是JSON或不是游戏邀请，按普通消息处理
      }

      return {
        uid: m.sender_uid,
        username: m.sender?.username || '研究员',
        text: isGameInvite ? '' : m.message,
        time: new Date(m.created_at),
        type: isGameInvite ? 'game_invite' : 'normal',
        gameInviteData: gameInviteData
      }
    }))
    messages.value[friend.uid] = historyMessages
  } catch (err) {
    console.error('加载聊天历史失败', err)
  }

  scrollToBottom()
}

const handleSend = () => {
  if (!newMessage.value.trim() || !activeChat.value) return
  
  websocket.send({
    type: 'private_chat',
    target_uid: activeChat.value.uid,
    message: newMessage.value
  })
  
  newMessage.value = ''
}

// 监听消息
const onPrivateMessage = async (msg: any) => {
  const otherUID = msg.uid === currentUser.value.uid ? msg.target_uid : msg.uid

  if (!messages.value[otherUID]) {
    messages.value[otherUID] = []
  }

  // 尝试解析游戏邀请消息
  let isGameInvite = false
  let gameInviteData = null

  try {
    const parsed = JSON.parse(msg.message)
    if (parsed.type === 'game_invite') {
      isGameInvite = true
      gameInviteData = parsed

      // 检查房间状态
      if (gameInviteData.room_id) {
        gameInviteData.room_status = await checkRoomStatus(gameInviteData.room_id)
      }
    }
  } catch (e) {
    // 不是JSON或不是游戏邀请，按普通消息处理
  }

  messages.value[otherUID].push({
    uid: msg.uid,
    username: msg.data?.username || '研究员',
    text: isGameInvite ? '' : msg.message,
    time: new Date(),
    type: isGameInvite ? 'game_invite' : 'normal',
    gameInviteData: gameInviteData
  })

  // 如果当前正在处理该对象的聊天，滚动
  if (activeChat.value?.uid === otherUID) {
    scrollToBottom()
  }
}

const handleJoinGame = (roomId: string) => {
  router.push(`/room/${roomId}`)
}

const onErrorMessage = (msg: any) => {
  showAlert(msg.message, '通信链路异常')
}

onMounted(() => {
  Promise.all([fetchFriends(), fetchRequests()]).finally(() => {
    loading.value = false
  })

  websocket.on('private_chat', onPrivateMessage)
  websocket.on('error', onErrorMessage)
  websocket.on('friend_request', handleIncomingFriendRequest)
  websocket.on('friend_request_handled', handleFriendRequestHandled)
})

onUnmounted(() => {
  websocket.off('private_chat', onPrivateMessage)
  websocket.off('error', onErrorMessage)
  websocket.off('friend_request', handleIncomingFriendRequest)
  websocket.off('friend_request_handled', handleFriendRequestHandled)
})

const handleIncomingFriendRequest = () => {
  fetchRequests()
}

const handleFriendRequestHandled = (data: any) => {
  if (data.action === 'accept') {
    fetchFriends()
  }
}

const formatTime = (date: Date) => {
  const d = new Date(date)
  return d.getHours().toString().padStart(2, '0') + ':' + 
         d.getMinutes().toString().padStart(2, '0')
}
</script>

<template>
  <div class="h-[100dvh] bg-slate-50 dark:bg-[#0a0a0c] text-white flex flex-col transition-colors duration-500 overflow-hidden">
    <!-- Header -->
    <header class="h-12 sm:h-14 bg-white/70 dark:bg-black/20 backdrop-blur-xl border-b border-slate-200 dark:border-white/5 flex items-center px-3 sm:px-4 shrink-0 relative z-30">
      <div class="flex items-center gap-2 sm:gap-3 w-full">
        <button @click="router.back()" class="p-1.5 sm:p-2 bg-slate-100 dark:bg-white/5 rounded-lg sm:rounded-xl text-slate-500 hover:bg-slate-200 dark:hover:bg-white/10 transition-all shrink-0">
          <ArrowLeft class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
        </button>
        <div class="min-w-0 flex-1">
          <h1 class="text-sm sm:text-base font-black text-slate-800 dark:text-white uppercase tracking-tighter flex items-center gap-1.5 sm:gap-2">
            <MessageCircle class="w-3.5 h-3.5 sm:w-4 sm:h-4 text-blue-500 shrink-0" />
            <span class="truncate">加密通讯链路</span>
          </h1>
          <p class="hidden sm:block text-[7px] sm:text-[8px] text-slate-400 font-mono uppercase tracking-widest mt-0.5">Secure_P2P_Messaging_System</p>
        </div>
      </div>
    </header>

    <main class="flex-1 flex overflow-hidden">
      <!-- Sidebar -->
      <aside :class="cn(
        'w-full lg:w-[300px] border-r border-slate-200 dark:border-white/5 flex flex-col bg-white/30 dark:bg-white/[0.01] shrink-0 transition-all duration-300',
        activeChat ? 'hidden lg:flex' : 'flex'
      )">
        <!-- Sidebar Header -->
        <div class="p-2.5 sm:p-3 pb-1.5 flex items-center justify-between">
          <div class="flex items-center gap-1.5">
            <User class="w-2.5 h-2.5 sm:w-3 sm:h-3 text-blue-500" />
            <span class="text-[7px] sm:text-[8px] font-black uppercase tracking-widest text-slate-500">研究员目录</span>
          </div>
          <button
            @click="focusSearch"
            class="p-1 sm:p-1.5 hover:bg-blue-500/10 text-blue-500 rounded-lg transition-all"
            title="添加新研究员"
          >
            <UserPlus class="w-3 h-3 sm:w-3.5 sm:h-3.5" />
          </button>
        </div>

        <div class="p-2.5 sm:p-3">
          <div class="relative group">
            <div class="absolute inset-0 bg-blue-500/5 rounded-lg sm:rounded-xl blur group-focus-within:bg-blue-500/10 transition-all"></div>
            <Search class="absolute left-2.5 sm:left-3 top-1/2 -translate-y-1/2 w-3 h-3 sm:w-3.5 sm:h-3.5 text-slate-400" />
            <input
              v-model="searchTerm"
              placeholder="搜索联系人..."
              class="relative w-full h-8 sm:h-9 bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-lg sm:rounded-xl pl-8 sm:pl-9 pr-2.5 sm:pr-3 text-[10px] sm:text-xs text-slate-700 dark:text-white focus:outline-none focus:border-blue-500/50 transition-all font-medium"
            />
          </div>
        </div>

        <!-- Friends List -->
        <div class="flex-1 overflow-y-auto custom-scrollbar p-2.5 sm:p-3 pt-0 space-y-1">
          <div class="flex items-center gap-1.5 mb-1.5 sm:mb-2 px-1.5">
            <User class="w-2.5 h-2.5 sm:w-3 sm:h-3 text-blue-500" />
            <span class="text-[7px] sm:text-[8px] font-black uppercase tracking-widest text-slate-500">活跃联系人</span>
          </div>

          <div v-if="loading" class="py-8 flex flex-col items-center opacity-30">
            <div class="w-6 h-6 border-2 border-blue-500/10 border-t-blue-500 rounded-full animate-spin mb-2"></div>
            <span class="text-[8px] font-black uppercase tracking-widest">Loading_Registry...</span>
          </div>
          <div v-else-if="filteredFriends.length === 0" class="py-8 flex flex-col items-center opacity-20 grayscale">
            <FlaskConical class="w-8 h-8 mb-2" />
            <span class="text-[8px] font-black uppercase tracking-widest">No_Connections_Found</span>
          </div>

          <div
            v-for="friend in filteredFriends"
            :key="friend.uid"
            @click="selectChat(friend)"
            :class="cn(
              'w-full p-2 sm:p-2.5 rounded-xl sm:rounded-2xl flex items-center gap-2 sm:gap-2.5 transition-all group relative overflow-hidden cursor-pointer',
              activeChat?.uid === friend.uid
                ? 'bg-blue-600 text-white shadow-lg shadow-blue-500/20 translate-x-0.5'
                : 'hover:bg-white dark:hover:bg-white/5 text-slate-700 dark:text-slate-400'
            )"
          >
            <div class="relative shrink-0">
              <div :class="cn(
                'w-8 h-8 sm:w-9 sm:h-9 rounded-lg sm:rounded-xl flex items-center justify-center text-base sm:text-lg border transition-all duration-300',
                activeChat?.uid === friend.uid ? 'bg-white/20 border-white/20' : 'bg-slate-100 dark:bg-white/5 border-slate-200 dark:border-white/10 group-hover:scale-110'
              )">
                {{ friend.avatar || '🧪' }}
              </div>
              <div v-if="friend.is_online" class="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 sm:w-3 sm:h-3 bg-emerald-500 border-2 border-white dark:border-[#0f0f12] rounded-full"></div>
            </div>
            <div class="flex-1 min-w-0 text-left">
              <div class="font-black text-[10px] sm:text-xs tracking-tight flex items-center gap-1 sm:gap-1.5">
                <span class="truncate">{{ friend.nickname || friend.username }}</span>
                <span class="text-[7px] sm:text-[8px] font-mono text-slate-400 group-hover:text-white/60 transition-colors shrink-0">UID:{{ friend.uid }}</span>
              </div>
              <div :class="cn(
                'text-[8px] sm:text-[9px] font-mono mt-0.5 truncate uppercase tracking-tighter opacity-60',
                activeChat?.uid === friend.uid ? 'text-white' : 'text-slate-400'
              )">
                {{ friend.is_online ? 'SYNC_ACTIVE' : 'OFFLINE_MODE' }}
              </div>
            </div>

            <button
              @click.stop="deleteFriend(friend.uid)"
              class="p-1 sm:p-1.5 hover:bg-rose-500/20 hover:text-rose-500 rounded-lg transition-all shrink-0"
            >
              <Trash2 class="w-3 h-3 sm:w-3.5 sm:h-3.5" />
            </button>
          </div>
        </div>

        <!-- Footer Actions -->
        <div class="p-2.5 sm:p-3 border-t border-slate-200 dark:border-white/5">
          <button
            @click="openRequestsModal"
            class="w-full flex items-center justify-between px-3 sm:px-4 h-10 sm:h-11 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 rounded-lg sm:rounded-xl text-[10px] sm:text-xs font-black uppercase tracking-widest text-slate-600 dark:text-white transition-all group relative overflow-hidden"
          >
            <div class="flex items-center gap-1.5 sm:gap-2 relative z-10">
              <UserPlus class="w-3 h-3 sm:w-4 sm:h-4 text-blue-500 group-hover:scale-110 transition-transform" />
              <span class="truncate">好友请求</span>
            </div>
            <div v-if="pendingRequests.length > 0" class="relative z-10 flex items-center justify-center min-w-[18px] sm:min-w-[20px] h-4 sm:h-5 px-1 sm:px-1.5 bg-rose-500 text-white text-[8px] sm:text-[9px] rounded-full animate-bounce">
              {{ pendingRequests.length }}
            </div>
            <div class="absolute inset-0 bg-gradient-to-r from-blue-500/0 via-blue-500/5 to-blue-500/0 translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-1000"></div>
          </button>
        </div>
      </aside>

      <!-- Chat Area -->
      <section :class="cn(
        'flex-1 flex flex-col bg-slate-50/50 dark:bg-black/40 relative',
        !activeChat && 'hidden lg:flex'
      )">
        <template v-if="activeChat">
          <!-- Chat Header -->
          <div class="h-12 sm:h-14 px-3 sm:px-4 border-b border-slate-200 dark:border-white/5 flex items-center justify-between shrink-0 bg-white/50 dark:bg-black/20 backdrop-blur-md">
            <div class="flex items-center gap-2 sm:gap-2.5 min-w-0 flex-1">
              <!-- Mobile Back Button -->
              <button
                @click="activeChat = null"
                class="lg:hidden p-1.5 bg-slate-100 dark:bg-white/5 rounded-lg text-slate-500 hover:bg-slate-200 dark:hover:bg-white/10 transition-all shrink-0"
              >
                <ArrowLeft class="w-3.5 h-3.5" />
              </button>

              <div class="w-7 h-7 sm:w-8 sm:h-8 rounded-lg bg-blue-600/10 flex items-center justify-center text-sm sm:text-base border border-blue-600/20 shrink-0">
                {{ activeChat.avatar || '🧪' }}
              </div>
              <div class="min-w-0 flex-1">
                <h2 class="text-[10px] sm:text-xs font-black text-slate-800 dark:text-white uppercase tracking-wider flex items-center gap-1 sm:gap-1.5">
                  <span class="truncate">{{ activeChat.username }}</span>
                  <span class="text-[7px] sm:text-[8px] font-mono text-slate-400 bg-slate-100 dark:bg-white/5 px-1 py-0.5 rounded shrink-0">UID:{{ activeChat.uid }}</span>
                </h2>
                <div class="flex items-center gap-1 mt-0.5">
                  <div :class="cn('w-1 h-1 rounded-full', activeChat.is_online ? 'bg-emerald-500 animate-pulse' : 'bg-slate-400')"></div>
                  <span class="text-[7px] sm:text-[8px] font-mono text-slate-500 uppercase tracking-widest truncate">{{ activeChat.is_online ? 'Encrypted Link Active' : 'Offline' }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Messages container -->
          <div
            ref="scrollContainer"
            class="flex-1 overflow-y-auto p-3 sm:p-4 space-y-3 sm:space-y-4 custom-scrollbar scroll-smooth"
          >
            <div v-if="(messages[activeChat.uid]?.length || 0) === 0" class="h-full flex flex-col items-center justify-center opacity-20 grayscale py-8 translate-y-[-10%] px-3">
              <div class="w-12 h-12 sm:w-14 sm:h-14 rounded-xl sm:rounded-2xl border-2 border-dashed border-slate-300 dark:border-white/20 flex items-center justify-center mb-3 sm:mb-4">
                <ShieldAlert class="w-5 h-5 sm:w-6 sm:h-6" />
              </div>
              <h3 class="text-xs sm:text-sm font-black uppercase tracking-widest mb-1 text-center">端到端加密对局室已建立</h3>
              <p class="text-[9px] sm:text-[10px] font-medium max-w-sm text-center leading-relaxed italic px-3">
                所有传输的数据均经过量子纠缠加密处理，指挥中心无法拦截此频率。
              </p>
            </div>

            <div
              v-for="(msg, idx) in messages[activeChat.uid]"
              :key="idx"
              :class="cn(
                'flex flex-col max-w-[85%] sm:max-w-[75%] animate-in fade-in duration-500',
                msg.uid === currentUser.uid ? 'ml-auto items-end' : 'mr-auto items-start'
              )"
            >
              <div class="flex items-center gap-1.5 sm:gap-2 px-0.5 mb-1">
                <span :class="cn(
                  'text-[7px] sm:text-[8px] font-black uppercase tracking-widest',
                  msg.uid === currentUser.uid ? 'order-last text-blue-500' : 'text-slate-400'
                )">{{ msg.username }}</span>
                <span class="text-[6px] sm:text-[7px] font-mono text-slate-500">{{ formatTime(msg.time) }}</span>
              </div>

              <!-- Game Invite Card -->
              <div v-if="msg.type === 'game_invite' && msg.gameInviteData"
                :class="cn(
                  'w-full max-w-xs p-3 sm:p-4 rounded-2xl sm:rounded-3xl bg-gradient-to-br border-2 shadow-lg backdrop-blur-sm transition-all',
                  msg.gameInviteData.room_status === 'finished' || msg.gameInviteData.room_status === 'closed'
                    ? 'from-slate-300/30 to-slate-400/30 border-slate-400/30 grayscale opacity-60'
                    : 'from-blue-500/10 to-purple-500/10 border-blue-500/20'
                )"
              >
                <div class="flex items-center gap-2 sm:gap-2.5 mb-2 sm:mb-2.5">
                  <div :class="cn(
                    'w-8 h-8 sm:w-9 sm:h-9 rounded-lg sm:rounded-xl flex items-center justify-center',
                    msg.gameInviteData.room_status === 'finished' || msg.gameInviteData.room_status === 'closed'
                      ? 'bg-slate-500/20'
                      : 'bg-blue-500/20'
                  )">
                    <FlaskConical :class="cn(
                      'w-4 h-4 sm:w-4.5 sm:h-4.5',
                      msg.gameInviteData.room_status === 'finished' || msg.gameInviteData.room_status === 'closed'
                        ? 'text-slate-500'
                        : 'text-blue-500'
                    )" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="text-[10px] sm:text-xs font-black text-slate-800 dark:text-white truncate">
                      {{ msg.gameInviteData.room_name }}
                    </div>
                    <div :class="cn(
                      'text-[8px] sm:text-[9px] font-mono uppercase tracking-wider',
                      msg.gameInviteData.room_status === 'finished' || msg.gameInviteData.room_status === 'closed'
                        ? 'text-slate-500'
                        : 'text-slate-500'
                    )">
                      {{ msg.gameInviteData.room_status === 'finished' || msg.gameInviteData.room_status === 'closed' ? '房间已关闭' : '实验室邀请' }}
                    </div>
                  </div>
                </div>

                <div v-if="msg.gameInviteData.room_status !== 'finished' && msg.gameInviteData.room_status !== 'closed'" class="flex items-center gap-3 mb-3 text-[9px] sm:text-[10px]">
                  <div class="flex items-center gap-1 text-slate-600 dark:text-slate-400">
                    <Users class="w-3 h-3 sm:w-3.5 sm:h-3.5" />
                    <span class="font-bold">{{ msg.gameInviteData.player_count }}/{{ msg.gameInviteData.max_players }}</span>
                  </div>
                  <div v-if="msg.gameInviteData.is_points_mode" class="flex items-center gap-1 px-2 py-0.5 bg-amber-500/10 text-amber-600 dark:text-amber-500 rounded-lg">
                    <Trophy class="w-3 h-3 sm:w-3.5 sm:h-3.5" />
                    <span class="font-black uppercase tracking-widest">积分模式</span>
                  </div>
                </div>

                <button
                  @click="msg.gameInviteData.room_status !== 'finished' && msg.gameInviteData.room_status !== 'closed' && handleJoinGame(msg.gameInviteData.room_id)"
                  :disabled="msg.gameInviteData.room_status === 'finished' || msg.gameInviteData.room_status === 'closed'"
                  :class="cn(
                    'w-full h-9 sm:h-10 rounded-lg sm:rounded-xl font-black text-[10px] sm:text-xs uppercase tracking-widest transition-all shadow-md active:scale-95 flex items-center justify-center gap-1.5',
                    msg.gameInviteData.room_status === 'finished' || msg.gameInviteData.room_status === 'closed'
                      ? 'bg-slate-400 text-slate-600 cursor-not-allowed'
                      : 'bg-blue-600 hover:bg-blue-500 text-white'
                  )"
                >
                  <FlaskConical class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
                  {{ msg.gameInviteData.room_status === 'finished' || msg.gameInviteData.room_status === 'closed' ? '房间已关闭' : '立即加入实验室' }}
                </button>
              </div>

              <!-- Normal Message -->
              <div v-else :class="cn(
                'px-3 py-2 sm:px-3.5 sm:py-2.5 rounded-xl sm:rounded-2xl text-[10px] sm:text-xs font-medium leading-relaxed shadow-sm break-words border',
                msg.uid === currentUser.uid
                  ? 'bg-blue-600 text-white rounded-tr-none border-blue-500'
                  : 'bg-white dark:bg-[#15151a] text-slate-700 dark:text-slate-200 rounded-tl-none border-slate-200 dark:border-white/5'
              )">
                {{ msg.text }}
              </div>
            </div>
          </div>

          <!-- Input Area -->
          <div class="p-2.5 sm:p-3 border-t border-slate-200 dark:border-white/5 bg-white/50 dark:bg-black/20 backdrop-blur-md">
            <div class="relative group">
              <div class="absolute -inset-1 bg-gradient-to-r from-blue-600 to-indigo-600 rounded-xl sm:rounded-2xl blur opacity-0 group-focus-within:opacity-10 transition duration-500"></div>
              <div class="relative flex items-center gap-2">
                <input
                  v-model="newMessage"
                  @keyup.enter="handleSend"
                  placeholder="注入数据流以开启通信频率..."
                  class="flex-1 h-10 sm:h-11 bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-xl sm:rounded-2xl px-3 sm:px-4 text-[10px] sm:text-xs focus:outline-none focus:border-blue-500/50 transition-all font-medium text-slate-700 dark:text-white"
                />
                <button
                  @click="handleSend"
                  :disabled="!newMessage.trim()"
                  class="w-10 h-10 sm:w-11 sm:h-11 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white rounded-xl sm:rounded-2xl flex items-center justify-center transition-all shadow-lg shadow-blue-500/20 active:scale-95 shrink-0"
                >
                  <Send class="w-3.5 h-3.5 sm:w-4 sm:h-4 -rotate-12 group-hover:rotate-0 transition-transform" />
                </button>
              </div>
            </div>
          </div>
        </template>

        <div v-else class="flex-1 flex flex-col items-center justify-center p-6 opacity-30">
          <div class="relative mb-6">
            <div class="absolute inset-0 bg-blue-500/10 rounded-full blur-3xl animate-pulse"></div>
            <MessageCircle class="w-12 h-12 sm:w-14 sm:h-14 text-blue-500 relative z-10" />
          </div>
          <h3 class="text-base sm:text-lg font-black text-slate-400 dark:text-white uppercase tracking-[0.2em] mb-2 text-center px-4">选择活跃波段</h3>
          <p class="text-[10px] sm:text-xs font-medium text-slate-500 dark:text-slate-400 max-w-sm text-center leading-relaxed mb-5 px-4">
            点击左侧活跃研究员成员，建立点对点（P2P）加密对话隧道。
          </p>
          <button
            @click="focusSearch"
            class="px-5 py-2.5 sm:px-6 sm:py-3 bg-blue-600 hover:bg-blue-500 text-white rounded-lg sm:rounded-xl font-black uppercase tracking-widest transition-all shadow-lg shadow-blue-500/20 active:scale-95 flex items-center gap-1.5 sm:gap-2 opacity-100 text-[10px] sm:text-xs"
          >
            <UserPlus class="w-3 h-3 sm:w-3.5 sm:h-3.5" />
            建立新研究连接
          </button>
        </div>
      </section>
    </main>

    <!-- 搜索研究员模态框 -->
    <div v-if="showSearchModal" @click="showSearchModal = false" class="fixed inset-0 z-[140] flex items-center justify-center p-3 sm:p-4 bg-slate-900/60 backdrop-blur-sm">
      <div @click.stop class="w-full max-w-md sm:max-w-lg bg-white dark:bg-[#0f0f12] rounded-2xl sm:rounded-3xl border border-slate-200 dark:border-white/10 shadow-2xl flex flex-col overflow-hidden max-h-[90vh]">
        <div class="p-3 sm:p-4 border-b border-slate-100 dark:border-white/5 flex items-center justify-between">
          <div class="min-w-0 flex-1">
            <h3 class="text-sm sm:text-base font-black text-slate-800 dark:text-white uppercase tracking-tighter flex items-center gap-1.5">
              <Globe class="w-3.5 h-3.5 sm:w-4 sm:h-4 text-purple-500 shrink-0" />
              <span class="truncate">发现新研究员</span>
            </h3>
            <p class="text-[7px] sm:text-[8px] text-slate-400 font-mono uppercase tracking-widest mt-0.5">Cross-Server Researcher Search</p>
          </div>
          <button @click="showSearchModal = false" class="p-1.5 sm:p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-lg sm:rounded-xl text-slate-400 transition-all shrink-0 ml-2">
            <X class="w-4 h-4 sm:w-5 sm:h-5" />
          </button>
        </div>

        <div class="p-3 sm:p-4 space-y-3 sm:space-y-4">
          <div class="relative group">
            <div class="absolute inset-0 bg-blue-500/5 rounded-lg sm:rounded-xl blur group-focus-within:bg-blue-500/10 transition-all"></div>
            <Search class="absolute left-3 sm:left-4 top-1/2 -translate-y-1/2 w-3.5 h-3.5 sm:w-4 sm:h-4 text-slate-400" />
            <input
              v-model="searchQuery"
              @keyup.enter="triggerSearch"
              autoFocus
              placeholder="输入 UID 或称号进行扫描..."
              class="relative w-full h-10 sm:h-11 bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-lg sm:rounded-xl pl-10 sm:pl-11 pr-11 sm:pr-12 text-xs sm:text-sm text-slate-700 dark:text-white focus:outline-none focus:border-blue-500/50 transition-all font-medium"
            />
            <button
              @click="triggerSearch"
              class="absolute right-2.5 sm:right-3 top-1/2 -translate-y-1/2 p-1.5 sm:p-2 text-slate-400 hover:text-blue-500 transition-colors"
            >
              <Send class="w-4 h-4 sm:w-5 sm:h-5 rotate-[-45deg]" />
            </button>
          </div>

          <div class="min-h-[200px] sm:min-h-[250px] max-h-[50vh] sm:max-h-[400px] overflow-y-auto custom-scrollbar pr-1 space-y-2">
            <div v-if="searchLoading" class="flex flex-col items-center justify-center py-16 opacity-30">
              <Loader2 class="w-6 h-6 animate-spin text-blue-500 mb-3" />
              <span class="text-[8px] font-black uppercase tracking-widest">Scanning_Transmissions...</span>
            </div>

            <template v-else-if="globalSearchResults.length > 0">
              <div
                v-for="result in globalSearchResults"
                :key="result.uid"
                class="group p-3 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl flex items-center justify-between hover:border-purple-500/30 transition-all"
              >
                <div class="flex items-center gap-3">
                  <div class="relative">
                    <div class="w-11 h-11 rounded-xl bg-white dark:bg-white/10 flex items-center justify-center text-2xl border border-slate-200 dark:border-white/10 shadow-sm">
                      {{ result.avatar || '🧪' }}
                    </div>
                    <div v-if="result.is_online" class="absolute -bottom-0.5 -right-0.5 w-3 h-3 bg-emerald-500 border-[3px] border-white dark:border-[#0f0f12] rounded-full shadow-md shadow-emerald-500/20"></div>
                  </div>
                  <div>
                    <div class="text-sm font-bold text-slate-700 dark:text-white flex items-center gap-1.5">
                      {{ result.nickname || result.username }}
                      <span v-if="result.is_online" class="px-1.5 py-0.5 bg-emerald-500/10 text-emerald-500 text-[7px] font-black rounded uppercase tracking-widest">Active</span>
                    </div>
                    <div class="flex items-center gap-2 mt-1 grayscale opacity-60">
                      <span class="text-[8px] text-slate-400 font-mono tracking-tight uppercase">UID: {{ result.uid }} | {{ result.username }}</span>
                      <span class="text-[8px] text-blue-500 font-black uppercase tracking-widest">{{ result.points }}PT</span>
                      <span v-if="result.bounty > 0" class="text-[8px] text-rose-500 font-black uppercase tracking-widest">赏: {{ result.bounty }}</span>
                    </div>
                  </div>
                </div>

                <button
                  v-if="!isFriend(result.uid)"
                  @click="sendRequest(result.uid)"
                  class="px-4 h-9 bg-blue-600 hover:bg-blue-500 text-white text-[10px] font-black uppercase tracking-widest rounded-xl transition-all shadow-lg shadow-blue-500/10 active:scale-95"
                >
                  建立连接
                </button>
                <div v-else class="px-4 h-9 flex items-center gap-1.5 bg-emerald-500/10 text-emerald-500 text-[10px] font-black uppercase tracking-widest rounded-xl">
                  <Check class="w-3.5 h-3.5" />
                  已是联络员
                </div>
              </div>
            </template>

            <div v-else-if="searchQuery" class="flex flex-col items-center justify-center py-16 opacity-20 grayscale">
              <FlaskConical class="w-12 h-12 mb-3" />
              <p class="text-xs font-black uppercase tracking-[0.2em]">Signal_Lost_404</p>
              <p class="text-[9px] mt-1.5 italic font-medium uppercase">未能发现目标频率</p>
            </div>
            <div v-else class="flex flex-col items-center justify-center py-16 opacity-10">
              <Search class="w-16 h-16 mb-3" />
              <p class="text-[10px] font-bold uppercase tracking-widest">Waiting_For_Target_Input</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 好友请求模态框 -->
    <div v-if="showRequestsModal" @click="showRequestsModal = false" class="fixed inset-0 z-[150] flex items-center justify-center p-3 sm:p-4 bg-slate-900/60 backdrop-blur-sm">
      <div @click.stop class="w-full max-w-md sm:max-w-lg bg-white dark:bg-[#0f0f12] rounded-2xl sm:rounded-3xl border border-slate-200 dark:border-white/10 shadow-2xl flex flex-col overflow-hidden max-h-[90vh]">
        <div class="p-3 sm:p-4 border-b border-slate-100 dark:border-white/5 flex items-center justify-between">
          <div>
            <h3 class="text-sm sm:text-base font-black text-slate-800 dark:text-white uppercase tracking-tighter flex items-center gap-1.5">
              <UserPlus class="w-3.5 h-3.5 sm:w-4 sm:h-4 text-amber-500" />
              同步请求管理
            </h3>
            <p class="text-[7px] sm:text-[8px] text-slate-400 font-mono uppercase tracking-widest mt-0.5">Pending Connection Requests</p>
          </div>
          <button @click="showRequestsModal = false" class="p-1.5 sm:p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-lg sm:rounded-xl text-slate-400 transition-all">
            <X class="w-4 h-4 sm:w-5 sm:h-5" />
          </button>
        </div>

        <div class="p-3 sm:p-4 max-h-[500px] overflow-y-auto custom-scrollbar">
          <div v-if="pendingRequests.length === 0" class="flex flex-col items-center justify-center py-12 opacity-20 grayscale">
            <ShieldAlert class="w-12 h-12 mb-3" />
            <p class="text-xs font-black uppercase tracking-[0.2em]">All_Synced</p>
            <p class="text-[9px] mt-1.5 italic font-medium uppercase">暂无待处理的研究请求</p>
          </div>
          <div v-else class="space-y-3">
            <div
              v-for="req in pendingRequests"
              :key="req.id"
              class="p-3.5 bg-amber-500/5 dark:bg-amber-500/[0.02] border border-amber-500/10 rounded-2xl flex items-center justify-between animate-in slide-in-from-bottom-4 transition-all"
            >
              <div class="flex items-center gap-3">
                <div class="w-11 h-11 rounded-xl bg-amber-500/10 flex items-center justify-center text-2xl border border-amber-500/20">
                  {{ req.avatar || '🧪' }}
                </div>
                <div>
                  <div class="text-sm font-bold text-slate-700 dark:text-white">{{ req.nickname || req.username }}</div>
                  <div v-if="req.hello_message" class="text-[10px] text-amber-600/80 font-medium italic mt-1 bg-amber-500/10 px-2.5 py-1 rounded-lg line-clamp-2 italic">
                    "{{ req.hello_message }}"
                  </div>
                  <div v-else class="text-[8px] text-amber-500/60 font-mono mt-0.5 px-1.5 py-0.5 bg-amber-500/5 rounded">REQ_CONNECT_P2P</div>
                </div>
              </div>
              <div class="flex gap-2">
                <button @click="handleRequest(req.id, 'accept')" class="w-10 h-10 rounded-xl bg-emerald-500 hover:bg-emerald-600 text-white flex items-center justify-center transition-all shadow-lg shadow-emerald-500/20 active:scale-95">
                  <Check class="w-5 h-5" />
                </button>
                <button @click="handleRequest(req.id, 'decline')" class="w-10 h-10 rounded-xl bg-rose-500 hover:bg-rose-600 text-white flex items-center justify-center transition-all shadow-lg shadow-rose-500/20 active:scale-95">
                  <X class="w-5 h-5" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 5px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(100, 116, 139, 0.1);
  border-radius: 10px;
}
.dark .custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.03);
}
.custom-scrollbar-hidden::-webkit-scrollbar {
  display: none;
}
</style>
