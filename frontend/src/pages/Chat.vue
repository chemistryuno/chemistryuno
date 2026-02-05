<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { friendAPI, authAPI } from '../utils/api'
import websocket from '../utils/websocket'
import { 
  MessageCircle, User, UserPlus, Send, 
  Search, ArrowLeft, MoreVertical, X, Check,
  Trash2, ShieldAlert, FlaskConical, Globe, Loader2
} from 'lucide-vue-next'
import { cn } from '../utils/cn'
import { useDialog } from '../utils/dialog'

const router = useRouter()
const route = useRoute()
const { showAlert, showConfirm, showPrompt } = useDialog()
const currentUser = ref(JSON.parse(localStorage.getItem('user') || '{}'))

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
const scrollContainer = ref<HTMLElement | null>(null)

// 过滤后的好友列表
const filteredFriends = computed(() => {
  if (!searchTerm.value) return friends.value
  return friends.value.filter(f => 
    f.username.toLowerCase().includes(searchTerm.value.toLowerCase()) ||
    String(f.uid).includes(searchTerm.value)
  )
})

// 监听搜索词变化，进行全局搜索
let searchTimeout: any = null
watch(searchTerm, (newVal) => {
  if (searchTimeout) clearTimeout(searchTimeout)
  if (!newVal.trim()) {
    globalSearchResults.value = []
    return
  }
  
  searchLoading.value = true
  searchTimeout = setTimeout(async () => {
    try {
      const res = await authAPI.searchUsers(newVal)
      // 排除掉自己
      globalSearchResults.value = res.data.filter((u: any) => {
        return Number(u.uid) !== Number(currentUser.value.uid)
      })
    } catch (err) {
      console.error('全局搜索失败', err)
    } finally {
      searchLoading.value = false
    }
  }, 500)
})

const isSearchingDetailed = ref(false)

const triggerSearch = async () => {
  if (!searchTerm.value.trim()) return
  if (searchTimeout) clearTimeout(searchTimeout)
  
  searchLoading.value = true
  try {
    const res = await authAPI.searchUsers(searchTerm.value)
    globalSearchResults.value = res.data.filter((u: any) => {
      return Number(u.uid) !== Number(currentUser.value.uid)
    })
  } catch (err) {
    console.error('搜索点击执行失败', err)
  } finally {
    searchLoading.value = false
  }
}

const isFriend = (uid: number) => {
  return friends.value.some(f => Number(f.uid) === Number(uid))
}

const handleSearchClick = (user: any) => {
  if (isFriend(user.uid)) {
    const friend = friends.value.find(f => Number(f.uid) === Number(user.uid))
    if (friend) selectChat(friend)
  } else {
    sendRequest(user.uid)
  }
}

const fetchFriends = async () => {
  try {
    const res = await friendAPI.getFriends()
    friends.value = res.data
    
    // 如果 URL 中有 uid 参数，自动选择该好友
    const targetUid = route.query.uid
    if (targetUid) {
      const friend = friends.value.find(f => Number(f.uid) === Number(targetUid))
      if (friend) {
        selectChat(friend)
      }
    }
  } catch (err) {
    console.error('获取好友失败', err)
  }
}

const fetchRequests = async () => {
  try {
    const res = await friendAPI.getPendingRequests()
    pendingRequests.value = res.data
  } catch (err) {
    console.error('获取请求失败', err)
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
  searchTerm.value = ''
  nextTick(() => {
    const el = document.getElementById('search-input')
    if (el) el.focus()
  })
}

// 选择聊天对象
const selectChat = (friend: any) => {
  activeChat.value = friend
  if (!messages.value[friend.uid]) {
    messages.value[friend.uid] = []
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
const onPrivateMessage = (msg: any) => {
  const otherUID = msg.uid === currentUser.value.uid ? msg.target_uid : msg.uid
  
  if (!messages.value[otherUID]) {
    messages.value[otherUID] = []
  }
  
  messages.value[otherUID].push({
    uid: msg.uid,
    username: msg.data?.username || '研究员',
    text: msg.message,
    time: new Date()
  })

  // 如果当前正在处理该对象的聊天，滚动
  if (activeChat.value?.uid === otherUID) {
    scrollToBottom()
  }
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
})

onUnmounted(() => {
  websocket.off('private_chat', onPrivateMessage)
  websocket.off('error', onErrorMessage)
})

const formatTime = (date: Date) => {
  const d = new Date(date)
  return d.getHours().toString().padStart(2, '0') + ':' + 
         d.getMinutes().toString().padStart(2, '0')
}
</script>

<template>
  <div class="min-h-screen bg-slate-50 dark:bg-[#0a0a0c] text-white flex flex-col transition-colors duration-500">
    <!-- Header -->
    <header class="h-20 bg-white/70 dark:bg-black/20 backdrop-blur-xl border-b border-slate-200 dark:border-white/5 flex items-center px-8 shrink-0 relative z-30">
      <div class="flex items-center gap-6">
        <button @click="router.back()" class="p-3 bg-slate-100 dark:bg-white/5 rounded-2xl text-slate-500 hover:bg-slate-200 dark:hover:bg-white/10 transition-all">
          <ArrowLeft class="w-6 h-6" />
        </button>
        <div>
          <h1 class="text-2xl font-black text-slate-800 dark:text-white uppercase tracking-tighter flex items-center gap-3">
            <MessageCircle class="w-6 h-6 text-blue-500" />
            加密通讯链路
          </h1>
          <p class="text-[10px] text-slate-400 font-mono uppercase tracking-widest mt-0.5">Secure_P2P_Messaging_System</p>
        </div>
      </div>
    </header>

    <main class="flex-1 flex overflow-hidden">
      <!-- Sidebar -->
      <aside class="w-[380px] border-r border-slate-200 dark:border-white/5 flex flex-col bg-white/30 dark:bg-white/[0.01] shrink-0">
        <!-- Sidebar Header -->
        <div class="p-6 pb-2 flex items-center justify-between">
          <div class="flex items-center gap-2">
            <User class="w-4 h-4 text-blue-500" />
            <span class="text-[10px] font-black uppercase tracking-widest text-slate-500">研究员目录 / Registry</span>
          </div>
          <button 
            @click="focusSearch"
            class="p-2 hover:bg-blue-500/10 text-blue-500 rounded-xl transition-all"
            title="添加新研究员"
          >
            <UserPlus class="w-4 h-4" />
          </button>
        </div>

        <!-- Search -->
        <div class="p-6">
          <div class="relative group">
            <div class="absolute inset-0 bg-blue-500/5 rounded-2xl blur group-focus-within:bg-blue-500/10 transition-all"></div>
            <Search class="absolute left-5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
            <input 
              id="search-input"
              v-model="searchTerm"
              @keyup.enter="triggerSearch"
              placeholder="搜索 ID 或称号以建立链接..."
              class="relative w-full h-14 bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl pl-12 pr-12 text-sm text-slate-700 dark:text-white focus:outline-none focus:border-blue-500/50 transition-all font-medium"
            />
            <button 
              @click="triggerSearch"
              class="absolute right-4 top-1/2 -translate-y-1/2 p-2 text-slate-400 hover:text-blue-500 transition-colors"
            >
              <Send class="w-4 h-4 rotate-[-45deg]" />
            </button>
          </div>
        </div>

        <!-- Requests Section -->
        <div v-if="pendingRequests.length > 0" class="px-6 mb-6">
          <div class="flex items-center gap-2 mb-4 px-2">
            <UserPlus class="w-4 h-4 text-amber-500" />
            <span class="text-[10px] font-black uppercase tracking-widest text-slate-500">同步请求 / Requests</span>
            <span class="px-2 py-0.5 bg-amber-500 text-white text-[10px] font-black rounded-lg ml-auto animate-pulse">{{ pendingRequests.length }}</span>
          </div>
          <div class="space-y-3">
            <div 
              v-for="req in pendingRequests" 
              :key="req.id"
              class="p-4 bg-amber-500/5 dark:bg-amber-500/[0.02] border border-amber-500/10 rounded-2xl flex items-center justify-between animate-in slide-in-from-left-4"
            >
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-xl bg-amber-500/10 flex items-center justify-center text-lg border border-amber-500/20">
                  {{ req.avatar || '🧪' }}
                </div>
                <div class="min-w-0 flex-1">
                  <div class="text-sm font-bold text-slate-700 dark:text-white truncate">{{ req.username }}</div>
                  <div v-if="req.hello_message" class="text-[9px] text-amber-500/80 font-medium italic mt-0.5 line-clamp-1">"{{ req.hello_message }}"</div>
                  <div v-else class="text-[9px] text-amber-500/60 font-mono">REQ_CONNECT</div>
                </div>
              </div>
              <div class="flex gap-2">
                <button @click="handleRequest(req.id, 'accept')" class="w-8 h-8 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-white flex items-center justify-center transition-all shadow-lg shadow-emerald-500/10 active:scale-95">
                  <Check class="w-4 h-4" />
                </button>
                <button @click="handleRequest(req.id, 'decline')" class="w-8 h-8 rounded-lg bg-rose-500 hover:bg-rose-600 text-white flex items-center justify-center transition-all shadow-lg shadow-rose-500/10 active:scale-95">
                  <X class="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Global Search Results -->
        <div v-if="searchTerm && (globalSearchResults.length > 0 || searchLoading)" class="px-6 mb-6">
          <div class="flex items-center gap-2 mb-4 px-2">
            <Globe class="w-4 h-4 text-purple-500" />
            <span class="text-[10px] font-black uppercase tracking-widest text-slate-500">发现研究员 / DISCOVERY</span>
            <Loader2 v-if="searchLoading" class="w-3 h-3 animate-spin text-slate-400 ml-auto" />
          </div>
          <div class="space-y-2">
            <div 
              v-for="result in globalSearchResults" 
              :key="result.uid"
              @click="handleSearchClick(result)"
              class="group p-3 bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl flex items-center justify-between hover:border-purple-500/30 transition-all cursor-pointer"
            >
              <div class="flex items-center gap-3">
                <div class="relative">
                  <div class="w-9 h-9 rounded-xl bg-slate-100 dark:bg-white/10 flex items-center justify-center text-lg">
                    {{ result.avatar || '🧪' }}
                  </div>
                  <div v-if="result.is_online" class="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 bg-emerald-500 border-2 border-white dark:border-[#0a0a0c] rounded-full"></div>
                </div>
                <div class="min-w-0 flex-1">
                  <div class="text-xs font-bold text-slate-700 dark:text-white truncate flex items-center gap-1.5">
                    {{ result.username }}
                    <span v-if="result.is_online" class="w-1 h-1 rounded-full bg-emerald-500 animate-pulse"></span>
                  </div>
                  <div class="flex items-center gap-2 mt-0.5">
                    <span class="text-[8px] text-slate-400 font-mono tracking-tighter uppercase">ID: {{ result.uid }}</span>
                    <span class="text-[8px] text-blue-500/60 font-black uppercase tracking-tighter">{{ result.points }}PT</span>
                    <span class="text-[8px] text-amber-500/60 font-black uppercase tracking-tighter">WIN: {{ result.win_count }}</span>
                  </div>
                </div>
              </div>
              
              <div class="flex items-center gap-2">
                <span v-if="isFriend(result.uid)" class="text-[8px] font-black text-emerald-500 uppercase tracking-widest bg-emerald-500/10 px-2 py-1 rounded-lg">已建立链接</span>
                <button 
                  v-else
                  @click.stop="sendRequest(result.uid)"
                  class="w-8 h-8 rounded-lg bg-blue-600/10 hover:bg-blue-600 text-blue-600 hover:text-white flex items-center justify-center transition-all"
                  title="添加研究员"
                >
                  <UserPlus class="w-4 h-4" />
                </button>
              </div>
            </div>
            <div v-if="!searchLoading && globalSearchResults.length === 0" class="text-center py-4 opacity-30">
               <p class="text-[9px] font-black uppercase tracking-[0.2em]">End_Of_Transmission</p>
            </div>
          </div>
        </div>

        <!-- Friends List -->
        <div class="flex-1 overflow-y-auto custom-scrollbar p-6 pt-0 space-y-2">
          <div class="flex items-center gap-2 mb-4 px-2">
            <User class="w-4 h-4 text-blue-500" />
            <span class="text-[10px] font-black uppercase tracking-widest text-slate-500">活跃联系人 / CONTACTS</span>
          </div>
          
          <div v-if="loading" class="py-12 flex flex-col items-center opacity-30">
            <div class="w-8 h-8 border-2 border-blue-500/10 border-t-blue-500 rounded-full animate-spin mb-3"></div>
            <span class="text-[9px] font-black uppercase tracking-widest">Loading_Registry...</span>
          </div>
          <div v-else-if="filteredFriends.length === 0" class="py-12 flex flex-col items-center opacity-20 grayscale">
            <FlaskConical class="w-12 h-12 mb-4" />
            <span class="text-[10px] font-black uppercase tracking-widest">No_Connections_Found</span>
          </div>

          <div 
            v-for="friend in filteredFriends" 
            :key="friend.uid"
            @click="selectChat(friend)"
            :class="cn(
              'w-full p-4 rounded-3xl flex items-center gap-4 transition-all group relative overflow-hidden cursor-pointer',
              activeChat?.uid === friend.uid 
                ? 'bg-blue-600 text-white shadow-xl shadow-blue-500/20 translate-x-1' 
                : 'hover:bg-white dark:hover:bg-white/5 text-slate-700 dark:text-slate-400'
            )"
          >
            <div class="relative">
              <div :class="cn(
                'w-12 h-12 rounded-[20px] flex items-center justify-center text-2xl border transition-all duration-300',
                activeChat?.uid === friend.uid ? 'bg-white/20 border-white/20' : 'bg-slate-100 dark:bg-white/5 border-slate-200 dark:border-white/10 group-hover:scale-110'
              )">
                {{ friend.avatar || '🧪' }}
              </div>
              <div v-if="friend.is_online" class="absolute -bottom-0.5 -right-0.5 w-3.5 h-3.5 bg-emerald-500 border-[3px] border-white dark:border-[#0f0f12] rounded-full"></div>
            </div>
            <div class="flex-1 min-w-0 text-left">
              <div class="font-black text-sm tracking-tight flex items-center gap-2">
                <span class="truncate">{{ friend.username }}</span>
                <span class="text-[9px] font-mono text-slate-400 group-hover:text-white/60 transition-colors">ID:{{ friend.uid }}</span>
              </div>
              <div :class="cn(
                'text-[10px] font-mono mt-0.5 truncate uppercase tracking-tighter opacity-60',
                activeChat?.uid === friend.uid ? 'text-white' : 'text-slate-400'
              )">
                {{ friend.is_online ? 'SYNC_ACTIVE' : 'OFFLINE_MODE' }}
              </div>
            </div>
            
            <!-- Context Action -->
            <button 
              @click.stop="deleteFriend(friend.uid)" 
              class="opacity-0 group-hover:opacity-100 p-2 hover:bg-rose-500/20 hover:text-rose-500 rounded-xl transition-all"
            >
              <Trash2 class="w-4 h-4" />
            </button>
          </div>
        </div>
      </aside>

      <!-- Chat Area -->
      <section class="flex-1 flex flex-col bg-slate-50/50 dark:bg-black/40 relative">
        <template v-if="activeChat">
          <!-- Chat Header -->
          <div class="h-20 px-10 border-b border-slate-200 dark:border-white/5 flex items-center justify-between shrink-0 bg-white/50 dark:bg-black/20 backdrop-blur-md">
            <div class="flex items-center gap-4">
              <div class="w-10 h-10 rounded-xl bg-blue-600/10 flex items-center justify-center text-xl border border-blue-600/20">
                {{ activeChat.avatar || '🧪' }}
              </div>
              <div>
                <h2 class="text-sm font-black text-slate-800 dark:text-white uppercase tracking-wider flex items-center gap-2">
                  {{ activeChat.username }}
                  <span class="text-[9px] font-mono text-slate-400 bg-slate-100 dark:bg-white/5 px-1.5 py-0.5 rounded">ID:{{ activeChat.uid }}</span>
                </h2>
                <div class="flex items-center gap-1.5 mt-0.5">
                  <div :class="cn('w-1.5 h-1.5 rounded-full', activeChat.is_online ? 'bg-emerald-500 animate-pulse' : 'bg-slate-400')"></div>
                  <span class="text-[9px] font-mono text-slate-500 uppercase tracking-widest">{{ activeChat.is_online ? 'Encrypted Link Active' : 'Offline' }}</span>
                </div>
              </div>
            </div>
            <div class="flex gap-2">
              <button class="p-3 bg-slate-100 dark:bg-white/5 rounded-2xl text-slate-400 hover:text-blue-500 transition-all">
                <Search class="w-5 h-5" />
              </button>
              <button class="p-3 bg-slate-100 dark:bg-white/5 rounded-2xl text-slate-400 hover:text-blue-500 transition-all">
                <MoreVertical class="w-5 h-5" />
              </button>
            </div>
          </div>

          <!-- Messages container -->
          <div 
            ref="scrollContainer"
            class="flex-1 overflow-y-auto p-10 space-y-8 custom-scrollbar scroll-smooth"
          >
            <div v-if="(messages[activeChat.uid]?.length || 0) === 0" class="h-full flex flex-col items-center justify-center opacity-20 grayscale py-20 translate-y-[-10%]">
              <div class="w-24 h-24 rounded-[40px] border-4 border-dashed border-slate-300 dark:border-white/20 flex items-center justify-center mb-8">
                <ShieldAlert class="w-10 h-10" />
              </div>
              <h3 class="text-xl font-black uppercase tracking-widest mb-2">端到端加密对局室已建立</h3>
              <p class="text-xs font-medium max-w-sm text-center leading-relaxed italic">
                所有传输的数据均经过量子纠缠加密处理，指挥中心无法拦截此频率。
              </p>
            </div>

            <div 
              v-for="(msg, idx) in messages[activeChat.uid]" 
              :key="idx"
              :class="cn(
                'flex flex-col max-w-[70%] animate-in fade-in duration-500',
                msg.uid === currentUser.uid ? 'ml-auto items-end' : 'mr-auto items-start'
              )"
            >
              <div class="flex items-center gap-3 px-1 mb-2">
                <span :class="cn(
                  'text-[9px] font-black uppercase tracking-widest',
                  msg.uid === currentUser.uid ? 'order-last text-blue-500' : 'text-slate-400'
                )">{{ msg.username }}</span>
                <span class="text-[8px] font-mono text-slate-500">{{ formatTime(msg.time) }}</span>
              </div>
              
              <div :class="cn(
                'px-6 py-4 rounded-[32px] text-sm font-medium leading-relaxed shadow-sm break-words border',
                msg.uid === currentUser.uid 
                  ? 'bg-blue-600 text-white rounded-tr-none border-blue-500' 
                  : 'bg-white dark:bg-[#15151a] text-slate-700 dark:text-slate-200 rounded-tl-none border-slate-200 dark:border-white/5'
              )">
                {{ msg.text }}
              </div>
            </div>
          </div>

          <!-- Input Area -->
          <div class="p-8 border-t border-slate-200 dark:border-white/5 bg-white/50 dark:bg-black/20 backdrop-blur-md">
            <div class="relative group">
              <div class="absolute -inset-1 bg-gradient-to-r from-blue-600 to-indigo-600 rounded-[32px] blur opacity-0 group-focus-within:opacity-10 transition duration-500"></div>
              <div class="relative flex items-center gap-4">
                <input 
                  v-model="newMessage"
                  @keyup.enter="handleSend"
                  placeholder="注入数据流以开启通信频率..."
                  class="flex-1 h-16 bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[28px] px-8 text-sm focus:outline-none focus:border-blue-500/50 transition-all font-medium text-slate-700 dark:text-white"
                />
                <button 
                  @click="handleSend"
                  :disabled="!newMessage.trim()"
                  class="w-16 h-16 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white rounded-[28px] flex items-center justify-center transition-all shadow-xl shadow-blue-500/20 active:scale-95 shrink-0"
                >
                  <Send class="w-6 h-6 -rotate-12 group-hover:rotate-0 transition-transform" />
                </button>
              </div>
            </div>
          </div>
        </template>

        <div v-else class="flex-1 flex flex-col items-center justify-center p-20 opacity-30">
          <div class="relative mb-12">
            <div class="absolute inset-0 bg-blue-500/10 rounded-full blur-3xl animate-pulse"></div>
            <MessageCircle class="w-24 h-24 text-blue-500 relative z-10" />
          </div>
          <h3 class="text-2xl font-black text-slate-400 dark:text-white uppercase tracking-[0.2em] mb-4">选择活跃波段</h3>
          <p class="text-sm font-medium text-slate-500 dark:text-slate-400 max-w-sm text-center leading-relaxed mb-8">
            点击左侧活跃研究员成员，建立点对点（P2P）加密对话隧道。
          </p>
          <button 
            @click="focusSearch"
            class="px-8 py-4 bg-blue-600 hover:bg-blue-500 text-white rounded-2xl font-black uppercase tracking-widest transition-all shadow-xl shadow-blue-500/20 active:scale-95 flex items-center gap-3 opacity-100"
          >
            <UserPlus class="w-4 h-4" />
            建立新研究连接
          </button>
        </div>
      </section>
    </main>
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
