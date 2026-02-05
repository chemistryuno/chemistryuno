<template>
  <div class="min-h-screen bg-slate-50 dark:bg-[#0a0a0c] text-slate-800 dark:text-slate-300 font-sans selection:bg-blue-500/30 transition-colors duration-500">
    <div class="fixed inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-600/10 rounded-full blur-[120px] animate-pulse"></div>
      <div class="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-purple-600/5 rounded-full blur-[120px]"></div>
    </div>

    <div class="relative z-10 flex flex-col min-h-screen">
      <!-- Top Navigation -->
      <header class="px-8 py-6 border-b border-slate-200 dark:border-white/5 backdrop-blur-md bg-white/70 dark:bg-black/20 sticky top-0 z-50">
        <div class="max-w-[1400px] mx-auto flex justify-between items-center">
          <div class="flex items-center gap-6">
            <button @click="router.push('/')" class="p-3 hover:bg-slate-200 dark:hover:bg-white/5 rounded-2xl transition-all text-slate-500 dark:text-slate-400 hover:text-blue-600 dark:hover:text-white group">
              <ArrowLeft class="w-6 h-6 group-hover:-translate-x-1 transition-transform" />
            </button>
            <div>
              <h1 class="text-2xl font-black text-slate-900 dark:text-white tracking-tighter flex items-center gap-3">
                <Trophy class="w-6 h-6 text-amber-500" />
                全球科研积分榜
              </h1>
              <p class="text-[10px] text-slate-500 font-mono uppercase tracking-[0.2em] mt-0.5">Global_Research_Leaderboard</p>
            </div>
          </div>
        </div>
      </header>

      <main class="flex-1 max-w-[1100px] mx-auto w-full px-6 py-8">
        <!-- Top Section: Mode Switch & Stats -->
        <div class="flex flex-col md:flex-row md:items-center justify-between gap-6 mb-8">
          <!-- Mode Switch Tabs -->
          <div class="flex items-center gap-2 bg-slate-200/50 dark:bg-white/5 p-1.5 rounded-[20px] border border-slate-200 dark:border-white/5 w-fit">
            <button 
              @click="rankingMode = 'total'"
              :class="cn(
                'px-6 py-2.5 rounded-[15px] text-[10px] font-black uppercase tracking-widest transition-all',
                rankingMode === 'total' 
                  ? 'bg-blue-600/10 text-blue-600 dark:text-blue-400 border border-blue-500/20 shadow-sm' 
                  : 'text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'
              )"
            >
              全量积分
            </button>

            <button 
              @click="rankingMode = 'monthly'"
              :class="cn(
                'px-6 py-2.5 rounded-[15px] text-[10px] font-black uppercase tracking-widest transition-all',
                rankingMode === 'monthly' 
                  ? 'bg-indigo-600/10 text-indigo-600 dark:text-indigo-400 border border-indigo-500/20 shadow-sm' 
                  : 'text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'
              )"
            >
              本月活跃 (月榜)
            </button>
          </div>

          <!-- Search Bar -->
          <div class="flex-1 max-w-sm md:ml-4">
             <div class="relative group">
                <div class="absolute inset-0 bg-blue-500/5 rounded-2xl blur group-focus-within:bg-blue-500/10 transition-all"></div>
                <Search class="absolute left-5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 group-focus-within:text-blue-500 transition-colors" />
                <input 
                  v-model="searchTerm"
                  placeholder="搜索研究员 ID 或称号..."
                  class="relative w-full h-12 bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-[20px] pl-12 pr-6 text-xs text-slate-700 dark:text-white focus:outline-none focus:border-blue-500/50 transition-all font-medium"
                />
                <div v-if="isSearching" class="absolute right-4 top-1/2 -translate-y-1/2">
                  <Loader2 class="w-4 h-4 text-blue-500 animate-spin" />
                </div>
             </div>
          </div>

          <!-- Compact Stats Overview -->
          <div class="flex flex-wrap items-center gap-3">
            <div class="bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl px-4 py-3 flex items-center gap-3 shadow-sm group hover:border-amber-500/30 transition-colors">
              <div class="w-8 h-8 bg-amber-500/10 border border-amber-500/20 rounded-lg flex items-center justify-center text-amber-500 shrink-0">
                <Target class="w-4 h-4" />
              </div>
              <div class="flex flex-col">
                <span class="text-[8px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest leading-none mb-1">Decay</span>
                <p class="text-[11px] font-bold text-slate-600 dark:text-slate-300 leading-tight">前10%每周衰减2%</p>
              </div>
            </div>
            <div class="bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl px-4 py-3 flex items-center gap-3 shadow-sm group hover:border-blue-500/30 transition-colors">
              <div class="w-8 h-8 bg-blue-500/10 border border-blue-500/20 rounded-lg flex items-center justify-center text-blue-500 shrink-0">
                <RefreshCw class="w-4 h-4" />
              </div>
              <div class="flex flex-col">
                <span class="text-[8px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest leading-none mb-1">Status</span>
                <p class="text-[11px] font-bold text-slate-600 dark:text-slate-300 leading-tight">赛季活跃中</p>
              </div>
            </div>
            <div class="bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl px-4 py-3 flex items-center gap-3 shadow-sm group hover:border-purple-500/30 transition-colors">
              <div class="w-8 h-8 bg-purple-500/10 border border-purple-500/20 rounded-lg flex items-center justify-center text-purple-500 shrink-0">
                <ShieldCheck class="w-4 h-4" />
              </div>
              <div class="flex flex-col">
                <span class="text-[8px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest leading-none mb-1">Total</span>
                <p class="text-[11px] font-bold text-slate-600 dark:text-slate-300 leading-tight">{{ leaderboard.length }} 名研究员</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Leaderboard Table -->
        <div class="bg-white/80 dark:bg-[#121216]/60 backdrop-blur-xl border border-slate-200 dark:border-white/10 rounded-[32px] overflow-hidden shadow-2xl dark:shadow-none">
          <div v-if="loading" class="py-32 flex flex-col items-center justify-center">
            <Loader2 class="w-10 h-10 animate-spin text-blue-500 mb-4" />
            <p class="text-xs font-black uppercase tracking-widest text-slate-500">Accessing_Database</p>
          </div>
          <div v-else class="overflow-x-auto">
            <table class="w-full border-collapse">
              <thead>
                <tr class="bg-slate-50/50 dark:bg-white/[0.02] border-b border-slate-100 dark:border-white/5 text-left">
                  <th class="px-8 py-5 text-[10px] font-black text-slate-500 uppercase tracking-widest">{{ searchTerm ? 'Match' : 'Rank' }}</th>
                  <th class="px-6 py-5 text-[10px] font-black text-slate-500 uppercase tracking-widest">Researcher</th>
                  <th class="px-6 py-5 text-[10px] font-black text-slate-500 uppercase tracking-widest">Points</th>
                  <th class="px-6 py-5 text-[10px] font-black text-slate-500 uppercase tracking-widest">Experimental_Bonus</th>
                  <th class="px-8 py-5 text-[10px] font-black text-slate-500 uppercase tracking-widest text-right">Actions</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 dark:divide-white/5">
                <tr 
                  v-for="(player, idx) in (searchTerm ? searchResults : leaderboard)" 
                  :key="player.uid"
                  :class="cn(
                    'group transition-colors',
                    player.uid === user.uid ? 'bg-blue-50/70 dark:bg-blue-500/[0.03]' : 'hover:bg-slate-50/50 dark:hover:bg-white/[0.02]'
                  )"
                >
                  <td class="px-8 py-4">
                    <div class="flex items-center gap-4">
                      <template v-if="!searchTerm">
                        <span :class="cn(
                          'w-7 h-7 rounded-lg flex items-center justify-center text-[11px] font-black italic shadow-lg',
                          idx === 0 ? 'bg-amber-500 text-amber-950 dark:text-black' :
                          idx === 1 ? 'bg-slate-300 text-slate-900 dark:text-black' :
                          idx === 2 ? 'bg-amber-700 text-white' :
                          'bg-slate-100 dark:bg-white/5 text-slate-500'
                        )">
                          {{ idx + 1 }}
                        </span>
                      </template>
                      <template v-else>
                        <div class="w-7 h-7 bg-blue-500/10 text-blue-500 rounded-lg flex items-center justify-center">
                          <Target class="w-3 h-3" />
                        </div>
                      </template>
                    </div>
                  </td>
                  <td class="px-6 py-4">
                    <div class="flex items-center gap-4">
                       <div class="relative w-9 h-9 rounded-lg bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 flex items-center justify-center text-lg overflow-hidden shrink-0">
                          <template v-if="player.avatar && player.avatar.startsWith('data:')">
                            <img :src="player.avatar" class="w-full h-full object-cover" />
                          </template>
                          <template v-else>
                            {{ player.avatar || '🧪' }}
                          </template>
                          <div v-if="player.is_online" class="absolute bottom-0 right-0 w-2.5 h-2.5 bg-emerald-500 border-2 border-white dark:border-[#121216] rounded-full"></div>
                       </div>
                       <div class="flex flex-col">
                          <span class="text-sm font-black text-slate-900 dark:text-white group-hover:text-blue-500 transition-colors flex items-center gap-2">
                            {{ player.username }}
                            <span v-if="player.is_online" class="text-[7px] text-emerald-500 font-black uppercase tracking-tighter">Online</span>
                            <span v-if="player.uid === user.uid" class="text-[8px] bg-blue-600 px-1.5 py-0.5 rounded uppercase font-black tracking-widest text-white">You</span>
                          </span>
                          <span class="text-[8px] font-mono text-slate-400 dark:text-slate-500 uppercase tracking-tighter mt-0.5">ID: {{ player.uid }}</span>
                       </div>
                    </div>
                  </td>
                  <td class="px-6 py-4">
                    <div class="flex flex-col">
                       <span class="text-base font-black text-slate-900 dark:text-white font-mono tracking-tighter">
                         {{ rankingMode === 'monthly' ? player.monthly_points : player.points }}
                       </span>
                    </div>
                  </td>
                  <td class="px-6 py-4">
                    <div v-if="player.bounty > 0" class="flex flex-col">
                       <div class="flex items-center gap-1.5 text-rose-500">
                          <Flame class="w-3 h-3" />
                          <span class="text-sm font-black font-mono tracking-tighter">{{ player.bounty }}</span>
                       </div>
                    </div>
                    <div v-else class="text-[8px] font-bold text-slate-400 dark:text-slate-600 uppercase italic opacity-40 leading-none">
                      Standard
                    </div>
                  </td>
                  <td class="px-8 py-4 text-right">
                    <div v-if="player.uid !== user.uid" class="flex items-center justify-end gap-2">
                      <button 
                        v-if="player.is_online"
                        @click="handleDuel(player)"
                        title="Duel Protocol"
                        class="p-2.5 bg-blue-600/10 hover:bg-blue-600 text-blue-600 hover:text-white border border-blue-500/20 rounded-xl transition-all active:scale-95 shadow-sm"
                      >
                        <Swords class="w-3.5 h-3.5" />
                      </button>
                      <button 
                        @click="openBountyModal(player)"
                        title="Issue Bounty"
                        class="p-2.5 bg-rose-600/10 hover:bg-rose-600 text-rose-600 hover:text-white border border-rose-500/20 rounded-xl transition-all active:scale-95 shadow-sm"
                      >
                        <Crosshair class="w-3.5 h-3.5" />
                      </button>
                      <button 
                        v-if="!isFriend(player.uid)"
                        @click="handleAddFriend(player)"
                        title="Add Friend"
                        class="p-2.5 bg-amber-600/10 hover:bg-amber-600 text-amber-600 hover:text-white border border-amber-500/20 rounded-xl transition-all active:scale-95 shadow-sm"
                      >
                        <UserPlus class="w-3.5 h-3.5" />
                      </button>
                      <button 
                        v-else
                        @click="startPrivateChat(player)"
                        title="Private Message"
                        class="p-2.5 bg-emerald-600/10 hover:bg-emerald-600 text-emerald-600 hover:text-white border border-emerald-500/20 rounded-xl transition-all active:scale-95 shadow-sm"
                      >
                        <MessageCircle class="w-3.5 h-3.5" />
                      </button>
                    </div>
                    <div v-else class="text-[9px] font-black text-blue-600 dark:text-blue-500 uppercase tracking-widest italic pr-2">
                      Master
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </main>
    </div>

    <!-- Bounty Issue Modal -->
    <div v-if="showBountyModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
       <div class="absolute inset-0 bg-slate-900/40 dark:bg-black/80 backdrop-blur-md" @click="showBountyModal = false"></div>
       <div class="relative w-full max-w-md bg-white dark:bg-[#121216] border border-slate-200 dark:border-rose-500/30 rounded-[40px] shadow-2xl overflow-hidden animate-in zoom-in duration-300">
          <div class="p-8 border-b border-slate-100 dark:border-white/5 flex items-center justify-between bg-rose-500/5">
             <div class="flex items-center gap-4">
                <div class="w-12 h-12 bg-rose-500/10 border border-rose-500/20 rounded-2xl flex items-center justify-center text-rose-500">
                   <Target class="w-6 h-6" />
                </div>
                <div>
                   <h2 class="text-xl font-black text-slate-900 dark:text-white uppercase tracking-tight">发布悬赏令</h2>
                   <p class="text-[9px] text-rose-500/60 font-mono uppercase tracking-widest mt-1">Configure_Target_Bounty</p>
                </div>
             </div>
             <button @click="showBountyModal = false" class="text-slate-400 hover:text-slate-900 dark:hover:text-white transition-colors">
                <X class="w-6 h-6" />
             </button>
          </div>
          
          <div class="p-10 space-y-8 text-center">
             <div class="flex flex-col items-center">
                <div class="w-20 h-20 rounded-3xl bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 flex items-center justify-center text-4xl mb-6 relative group">
                   {{ selectedTarget?.avatar || '🧪' }}
                   <div class="absolute -top-3 -right-3 w-8 h-8 bg-rose-600 rounded-full flex items-center justify-center border-4 border-white dark:border-[#121216]">
                      <Crosshair class="w-4 h-4 text-white" />
                   </div>
                </div>
                <p class="text-[10px] font-black uppercase tracking-widest text-slate-400 dark:text-slate-500">目标研究员</p>
                <h3 class="text-2xl font-black text-slate-900 dark:text-white mt-1">{{ selectedTarget?.username }}</h3>
             </div>

             <div class="space-y-4">
                <div class="flex justify-between items-center px-2">
                   <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">投入科研积分</label>
                   <span class="text-[10px] text-rose-600 dark:text-rose-500 font-mono font-bold">{{ userPoints }} AVAILABLE</span>
                </div>
                <div class="relative">
                   <input 
                      v-model="bountyAmount" 
                      type="number" 
                      placeholder="输入积分数值..."
                      class="w-full h-16 bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-white px-6 py-5 rounded-2xl focus:ring-1 focus:ring-rose-500 outline-none transition-all font-mono text-lg"
                   />
                   <div class="absolute right-4 top-1/2 -translate-y-1/2 text-rose-500/30 font-mono uppercase text-[10px] font-black tracking-widest pointer-events-none">
                     Points_Mendeleef
                   </div>
                </div>
                <p class="text-[9px] text-slate-500 leading-relaxed text-left px-2">
                  <span class="text-rose-600 dark:text-rose-500">协议说明：</span>悬赏积分将立刻从您的账户中扣除。任何人在该目标的竞技对局中获胜均可平分此项奖赏。悬赏一旦发布，不可撤回。
                </p>
             </div>

             <div class="flex gap-4 pt-4">
                <button 
                   @click="showBountyModal = false"
                   class="flex-1 h-14 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-500 font-bold rounded-2xl transition-all uppercase tracking-widest text-xs"
                >
                   放弃发布
                </button>
                <button 
                   @click="handleCreateBounty"
                   :disabled="submitting || !bountyAmount || bountyAmount > userPoints"
                   class="flex-1 h-14 bg-rose-600 hover:bg-rose-500 text-white font-black rounded-2xl transition-all shadow-[0_15px_30px_rgba(225,29,72,0.3)] disabled:grayscale flex items-center justify-center gap-2 group/btn"
                >
                   <template v-if="submitting">
                      <Loader2 class="w-5 h-5 animate-spin" />
                   </template>
                   <template v-else>
                      <Target class="w-4 h-4 group-hover:scale-125 transition-transform" />
                      <span class="uppercase tracking-widest text-xs">执行发布</span>
                   </template>
                </button>
             </div>
          </div>
       </div>
    </div>

    <!-- Floating Chat Toggle -->
    <button 
      @click="showChat = !showChat" 
      class="fixed bottom-6 right-6 z-50 w-14 h-14 bg-blue-600 hover:bg-blue-500 text-white rounded-[24px] shadow-2xl shadow-blue-500/30 flex items-center justify-center transition-all hover:scale-110 active:scale-95 group"
    >
      <MessageCircle class="w-6 h-6 group-hover:rotate-12 transition-transform" />
      <div v-if="hasNewMessage" class="absolute -top-1 -right-1 w-4 h-4 bg-rose-500 border-2 border-white dark:border-[#0a0a0c] rounded-full animate-pulse"></div>
    </button>

    <!-- Chat Sidebar/Modal -->
    <div 
      v-if="showChat"
      class="fixed bottom-24 right-6 z-50 w-[calc(100vw-3rem)] sm:w-[400px] shadow-2xl animate-in slide-in-from-bottom-10 duration-300 pointer-events-auto"
    >
      <ChatBox title="全球通信频率" maxHeight="500px" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { pointsAPI, gameAPI, friendAPI, authAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import { Trophy, ArrowLeft, Loader2, Target, RefreshCw, ShieldCheck, Crosshair, Flame, X, Swords, MessageCircle, MessageSquare, UserPlus, Search } from 'lucide-vue-next'
import { cn } from '../utils/cn'
import ChatBox from '../components/ChatBox.vue'
import websocket from '../utils/websocket'

const router = useRouter()
const { showAlert, showPrompt } = useDialog()
const user = ref(JSON.parse(localStorage.getItem('user') || '{}'))

const leaderboard = ref<any[]>([])
const friendsList = ref<any[]>([])
const loading = ref(true)
const userPoints = ref(0)
const rankingMode = ref<'total' | 'monthly'>('total')

const searchTerm = ref('')
const searchResults = ref<any[]>([])
const isSearching = ref(false)

// 监听搜索词
let searchTimeout: any = null
watch(searchTerm, (newVal) => {
  if (searchTimeout) clearTimeout(searchTimeout)
  if (!newVal.trim()) {
    searchResults.value = []
    isSearching.value = false
    return
  }
  
  isSearching.value = true
  searchTimeout = setTimeout(async () => {
    try {
      const res = await authAPI.searchUsers(newVal)
      searchResults.value = res.data || []
    } catch (err) {
      console.error('搜索失败:', err)
      searchResults.value = []
    } finally {
      isSearching.value = false
    }
  }, 500)
})

const isFriend = (uid: number) => {
  return friendsList.value.some(f => f.uid === uid)
}

const handleAddFriend = async (player: any) => {
  const message = await showPrompt('请输入申请信息（可选）:', '你好，我想和你一起进行化学实验。', '发送好友请求')
  if (message === null) return

  try {
    await friendAPI.sendRequest(player.uid, message)
    showAlert(`已向研究员 ${player.username} 发送同步请求，等待量子握手。`, '请求已发送')
  } catch (error: any) {
    showAlert(error.response?.data?.error || '请求发送失败', '链路故障')
  }
}

const showChat = ref(false)
const hasNewMessage = ref(false)

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

const loadLeaderboard = async () => {
  try {
    loading.value = true
    const [leaderRes, friendsRes] = await Promise.all([
      pointsAPI.getLeaderboard(rankingMode.value),
      friendAPI.getFriends()
    ])
    leaderboard.value = leaderRes.data
    friendsList.value = friendsRes.data
    
    // 同时也尝试更新一下本地的用户分数实时显示
    const self = leaderboard.value.find(p => p.uid === user.value.uid)
    if (self) userPoints.value = self.points
  } catch (error) {
    console.error('Failed to load ranking data:', error)
  } finally {
    loading.value = false
  }
}

import { watch } from 'vue'
watch(rankingMode, () => {
  loadLeaderboard()
})

const openBountyModal = (player: any) => {
  selectedTarget.value = player
  bountyAmount.value = 100
  showBountyModal.value = true
}

const handleCreateBounty = async () => {
  if (!selectedTarget.value || !bountyAmount.value) return
  if (bountyAmount.value > userPoints.value) {
    showAlert('科研积分余额不足，无法发起此项悬赏', '核心功率受限')
    return
  }

  try {
    submitting.value = true
    await pointsAPI.createBounty(selectedTarget.value.uid, bountyAmount.value)
    showAlert(`已成功对研究员 ${selectedTarget.value.username} 发布悬赏。`, '目标已锁定')
    showBountyModal.value = false
    loadLeaderboard() // 刷新列表
  } catch (error: any) {
    showAlert(error.response?.data?.error || '发布悬赏失败', '系统通讯故障')
  } finally {
    submitting.value = false
  }
}

const handleDuel = async (player: any) => {
  try {
    const res = await gameAPI.initiateDuel(player.uid)
    // 后端会通过 WebSocket 广播 duel_start，这里只需提示
    showAlert(`已向 ${player.username} 发起单挑协议，正在建立量子隧道...`, '协议启动')
  } catch (error: any) {
    showAlert(error.response?.data?.error || '发起单挑失败', '系统通讯故障')
  }
}

const onChatMessage = () => {
  if (!showChat.value) {
    hasNewMessage.value = true
  }
}

onMounted(() => {
  loadLeaderboard()
  websocket.on('chat', onChatMessage)
  websocket.on('private_chat', onChatMessage)
})

onUnmounted(() => {
  websocket.off('chat', onChatMessage)
  websocket.off('private_chat', onChatMessage)
})
</script>

<style src="./Ranking.css" scoped></style>
