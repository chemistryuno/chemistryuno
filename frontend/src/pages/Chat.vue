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
const searchQuery = ref('') // 模态框搜索词
const scrollContainer = ref<HTMLElement | null>(null)

const showSearchModal = ref(false)
const showRequestsModal = ref(false)

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
  showSearchModal.value = true
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

        <div class="p-6">
          <div class="relative group">
            <div class="absolute inset-0 bg-blue-500/5 rounded-2xl blur group-focus-within:bg-blue-500/10 transition-all"></div>
            <Search class="absolute left-5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
            <input 
              v-model="searchTerm"
              placeholder="搜索联系人..."
              class="relative w-full h-12 bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl pl-12 pr-4 text-sm text-slate-700 dark:text-white focus:outline-none focus:border-blue-500/50 transition-all font-medium"
            />
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

        <!-- Footer Actions -->
        <div class="p-6 border-t border-slate-200 dark:border-white/5">
          <button 
            @click="showRequestsModal = true"
            class="w-full flex items-center justify-between px-6 h-14 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 rounded-2xl text-sm font-black uppercase tracking-widest text-slate-600 dark:text-white transition-all group relative overflow-hidden"
          >
            <div class="flex items-center gap-3 relative z-10">
              <UserPlus class="w-5 h-5 text-blue-500 group-hover:scale-110 transition-transform" />
              好友请求
            </div>
            <div v-if="pendingRequests.length > 0" class="relative z-10 flex items-center justify-center min-w-[24px] h-6 px-2 bg-rose-500 text-white text-[10px] rounded-full animate-bounce">
              {{ pendingRequests.length }}
            </div>
            <div class="absolute inset-0 bg-gradient-to-r from-blue-500/0 via-blue-500/5 to-blue-500/0 translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-1000"></div>
          </button>
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

    <!-- 搜索研究员模态框 -->
    <div v-if="showSearchModal" class="fixed inset-0 z-[100] flex items-center justify-center p-6 bg-slate-900/60 backdrop-blur-sm animate-in fade-in duration-300">
      <div @click.stop class="w-full max-w-xl bg-white dark:bg-[#0f0f12] rounded-[40px] border border-slate-200 dark:border-white/10 shadow-2xl flex flex-col overflow-hidden animate-in zoom-in-95 duration-300">
        <div class="p-8 border-b border-slate-100 dark:border-white/5 flex items-center justify-between">
          <div>
            <h3 class="text-xl font-black text-slate-800 dark:text-white uppercase tracking-tighter flex items-center gap-2">
              <Globe class="w-5 h-5 text-purple-500" />
              发现新研究员
            </h3>
            <p class="text-[10px] text-slate-400 font-mono uppercase tracking-widest mt-1">Cross-Server Researcher Search</p>
          </div>
          <button @click="showSearchModal = false" class="p-3 hover:bg-slate-100 dark:hover:bg-white/5 rounded-2xl text-slate-400 transition-all">
            <X class="w-6 h-6" />
          </button>
        </div>

        <div class="p-8 space-y-6">
          <div class="relative group">
            <div class="absolute inset-0 bg-blue-500/5 rounded-2xl blur group-focus-within:bg-blue-500/10 transition-all"></div>
            <Search class="absolute left-5 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
            <input 
              v-model="searchQuery"
              @keyup.enter="triggerSearch"
              autoFocus
              placeholder="输入 ID 或称号进行扫描..."
              class="relative w-full h-16 bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl pl-14 pr-16 text-lg text-slate-700 dark:text-white focus:outline-none focus:border-blue-500/50 transition-all font-medium"
            />
            <button 
              @click="triggerSearch"
              class="absolute right-4 top-1/2 -translate-y-1/2 p-3 text-slate-400 hover:text-blue-500 transition-colors"
            >
              <Send class="w-6 h-6 rotate-[-45deg]" />
            </button>
          </div>

          <div class="min-h-[300px] max-h-[450px] overflow-y-auto custom-scrollbar pr-2 space-y-3">
            <div v-if="searchLoading" class="flex flex-col items-center justify-center py-20 opacity-30">
              <Loader2 class="w-8 h-8 animate-spin text-blue-500 mb-4" />
              <span class="text-[10px] font-black uppercase tracking-widest">Scanning_Transmissions...</span>
            </div>
            
            <template v-else-if="globalSearchResults.length > 0">
              <div 
                v-for="result in globalSearchResults" 
                :key="result.uid"
                class="group p-4 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-[28px] flex items-center justify-between hover:border-purple-500/30 transition-all"
              >
                <div class="flex items-center gap-4">
                  <div class="relative">
                    <div class="w-14 h-14 rounded-2xl bg-white dark:bg-white/10 flex items-center justify-center text-3xl border border-slate-200 dark:border-white/10 shadow-sm">
                      {{ result.avatar || '🧪' }}
                    </div>
                    <div v-if="result.is_online" class="absolute -bottom-1 -right-1 w-4 h-4 bg-emerald-500 border-4 border-white dark:border-[#0f0f12] rounded-full shadow-lg shadow-emerald-500/20"></div>
                  </div>
                  <div>
                    <div class="text-base font-bold text-slate-700 dark:text-white flex items-center gap-2">
                      {{ result.username }}
                      <span v-if="result.is_online" class="px-2 py-0.5 bg-emerald-500/10 text-emerald-500 text-[8px] font-black rounded uppercase tracking-widest">Active</span>
                    </div>
                    <div class="flex items-center gap-3 mt-1.5 grayscale opacity-60">
                      <span class="text-[9px] text-slate-400 font-mono tracking-tight uppercase">ID: {{ result.uid }}</span>
                      <span class="text-[9px] text-blue-500 font-black uppercase tracking-widest">{{ result.points }}PT</span>
                      <span v-if="result.bounty > 0" class="text-[9px] text-rose-500 font-black uppercase tracking-widest">赏: {{ result.bounty }}</span>
                    </div>
                  </div>
                </div>
                
                <button 
                  v-if="!isFriend(result.uid)"
                  @click="sendRequest(result.uid)"
                  class="px-6 h-12 bg-blue-600 hover:bg-blue-500 text-white text-xs font-black uppercase tracking-widest rounded-2xl transition-all shadow-lg shadow-blue-500/10 active:scale-95"
                >
                  建立连接
                </button>
                <div v-else class="px-6 h-12 flex items-center gap-2 bg-emerald-500/10 text-emerald-500 text-xs font-black uppercase tracking-widest rounded-2xl">
                  <Check class="w-4 h-4" />
                  已是联络员
                </div>
              </div>
            </template>

            <div v-else-if="searchQuery" class="flex flex-col items-center justify-center py-20 opacity-20 grayscale">
              <FlaskConical class="w-16 h-16 mb-4" />
              <p class="text-sm font-black uppercase tracking-[0.2em]">Signal_Lost_404</p>
              <p class="text-[10px] mt-2 italic font-medium uppercase">未能发现目标频率</p>
            </div>
            <div v-else class="flex flex-col items-center justify-center py-20 opacity-10">
              <Search class="w-20 h-20 mb-4" />
              <p class="text-xs font-bold uppercase tracking-widest">Waiting_For_Target_Input</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 好友请求模态框 -->
    <div v-if="showRequestsModal" class="fixed inset-0 z-[100] flex items-center justify-center p-6 bg-slate-900/60 backdrop-blur-sm animate-in fade-in duration-300">
      <div @click.stop class="w-full max-w-xl bg-white dark:bg-[#0f0f12] rounded-[40px] border border-slate-200 dark:border-white/10 shadow-2xl flex flex-col overflow-hidden animate-in zoom-in-95 duration-300">
        <div class="p-8 border-b border-slate-100 dark:border-white/5 flex items-center justify-between">
          <div>
            <h3 class="text-xl font-black text-slate-800 dark:text-white uppercase tracking-tighter flex items-center gap-2">
              <UserPlus class="w-5 h-5 text-amber-500" />
              同步请求管理
            </h3>
            <p class="text-[10px] text-slate-400 font-mono uppercase tracking-widest mt-1">Pending Connection Requests</p>
          </div>
          <button @click="showRequestsModal = false" class="p-3 hover:bg-slate-100 dark:hover:bg-white/5 rounded-2xl text-slate-400 transition-all">
            <X class="w-6 h-6" />
          </button>
        </div>

        <div class="p-8 max-h-[600px] overflow-y-auto custom-scrollbar">
          <div v-if="pendingRequests.length === 0" class="flex flex-col items-center justify-center py-16 opacity-20 grayscale">
            <ShieldAlert class="w-16 h-16 mb-4" />
            <p class="text-sm font-black uppercase tracking-[0.2em]">All_Synced</p>
            <p class="text-[10px] mt-2 italic font-medium uppercase">暂无待处理的研究请求</p>
          </div>
          <div v-else class="space-y-4">
            <div 
              v-for="req in pendingRequests" 
              :key="req.id"
              class="p-5 bg-amber-500/5 dark:bg-amber-500/[0.02] border border-amber-500/10 rounded-[32px] flex items-center justify-between animate-in slide-in-from-bottom-4 transition-all"
            >
              <div class="flex items-center gap-4">
                <div class="w-14 h-14 rounded-2xl bg-amber-500/10 flex items-center justify-center text-3xl border border-amber-500/20">
                  {{ req.avatar || '🧪' }}
                </div>
                <div>
                  <div class="text-base font-bold text-slate-700 dark:text-white">{{ req.username }}</div>
                  <div v-if="req.hello_message" class="text-xs text-amber-600/80 font-medium italic mt-1 bg-amber-500/10 px-3 py-1 rounded-lg line-clamp-2 italic">
                    "{{ req.hello_message }}"
                  </div>
                  <div v-else class="text-[10px] text-amber-500/60 font-mono mt-1 px-2 py-0.5 bg-amber-500/5 rounded">REQ_CONNECT_P2P</div>
                </div>
              </div>
              <div class="flex gap-3">
                <button @click="handleRequest(req.id, 'accept')" class="w-12 h-12 rounded-2xl bg-emerald-500 hover:bg-emerald-600 text-white flex items-center justify-center transition-all shadow-lg shadow-emerald-500/20 active:scale-95">
                  <Check class="w-6 h-6" />
                </button>
                <button @click="handleRequest(req.id, 'decline')" class="w-12 h-12 rounded-2xl bg-rose-500 hover:bg-rose-600 text-white flex items-center justify-center transition-all shadow-lg shadow-rose-500/20 active:scale-95">
                  <X class="w-6 h-6" />
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
