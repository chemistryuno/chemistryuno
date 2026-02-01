<template>
  <div class="min-h-screen bg-[#0a0a0c] text-slate-300 font-sans selection:bg-blue-500/30">
    <div class="fixed inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-600/10 rounded-full blur-[120px] animate-pulse"></div>
      <div class="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-purple-600/5 rounded-full blur-[120px]"></div>
    </div>

    <div class="relative z-10 flex flex-col min-h-screen">
      <!-- Top Navigation -->
      <header class="px-8 py-6 border-b border-white/5 backdrop-blur-md bg-black/20 sticky top-0 z-50">
        <div class="max-w-[1400px] mx-auto flex justify-between items-center">
          <div class="flex items-center gap-6">
            <button @click="router.push('/')" class="p-3 hover:bg-white/5 rounded-2xl transition-all text-slate-400 hover:text-white group">
              <ArrowLeft class="w-6 h-6 group-hover:-translate-x-1 transition-transform" />
            </button>
            <div>
              <h1 class="text-2xl font-black text-white tracking-tighter flex items-center gap-3">
                <Trophy class="w-6 h-6 text-amber-500" />
                全球科研积分榜
              </h1>
              <p class="text-[10px] text-slate-500 font-mono uppercase tracking-[0.2em] mt-0.5">Global_Research_Leaderboard</p>
            </div>
          </div>
        </div>
      </header>

      <main class="flex-1 max-w-[1000px] mx-auto w-full px-6 py-12">
        <!-- Stats Overview -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-12">
          <div class="bg-white/5 border border-white/10 rounded-[32px] p-8 flex flex-col items-center text-center">
            <div class="w-12 h-12 bg-amber-500/10 border border-amber-500/20 rounded-2xl flex items-center justify-center text-amber-500 mb-4">
              <Target class="w-6 h-6" />
            </div>
            <span class="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-1">Decay_Protocol</span>
            <p class="text-sm font-bold text-slate-300">前10%每周衰减2%</p>
          </div>
          <div class="bg-white/5 border border-white/10 rounded-[32px] p-8 flex flex-col items-center text-center">
            <div class="w-12 h-12 bg-blue-500/10 border border-blue-500/20 rounded-2xl flex items-center justify-center text-blue-500 mb-4">
              <RefreshCw class="w-6 h-6" />
            </div>
            <span class="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-1">Season_Status</span>
            <p class="text-sm font-bold text-slate-300">本月活跃赛季中</p>
          </div>
          <div class="bg-white/5 border border-white/10 rounded-[32px] p-8 flex flex-col items-center text-center">
            <div class="w-12 h-12 bg-purple-500/10 border border-purple-500/20 rounded-2xl flex items-center justify-center text-purple-500 mb-4">
              <ShieldCheck class="w-6 h-6" />
            </div>
            <span class="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-1">Total_Competitors</span>
            <p class="text-sm font-bold text-slate-300">{{ leaderboard.length }} 名活跃研究员</p>
          </div>
        </div>

        <!-- Leaderboard Table -->
        <div class="bg-[#121216]/60 backdrop-blur-xl border border-white/10 rounded-[40px] overflow-hidden shadow-2xl">
          <div v-if="loading" class="py-32 flex flex-col items-center justify-center">
            <Loader2 class="w-10 h-10 animate-spin text-blue-500 mb-4" />
            <p class="text-xs font-black uppercase tracking-widest text-slate-500">Accessing_Database</p>
          </div>
          <div v-else class="overflow-x-auto">
            <table class="w-full border-collapse">
              <thead>
                <tr class="bg-white/[0.02] border-b border-white/5 text-left">
                  <th class="px-8 py-6 text-[10px] font-black text-slate-500 uppercase tracking-widest">Rank</th>
                  <th class="px-6 py-6 text-[10px] font-black text-slate-500 uppercase tracking-widest">Researcher</th>
                  <th class="px-6 py-6 text-[10px] font-black text-slate-500 uppercase tracking-widest">Points</th>
                  <th class="px-6 py-6 text-[10px] font-black text-slate-500 uppercase tracking-widest">Experimental_Bonus</th>
                  <th class="px-8 py-6 text-[10px] font-black text-slate-500 uppercase tracking-widest text-right">Actions</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-white/5">
                <tr 
                  v-for="(player, idx) in leaderboard" 
                  :key="player.uid"
                  :class="cn(
                    'group transition-colors',
                    player.uid === user.uid ? 'bg-blue-500/[0.03]' : 'hover:bg-white/[0.02]'
                  )"
                >
                  <td class="px-8 py-6">
                    <div class="flex items-center gap-4">
                      <span :class="cn(
                        'w-8 h-8 rounded-xl flex items-center justify-center text-xs font-black italic shadow-lg',
                        idx === 0 ? 'bg-amber-500 text-black' :
                        idx === 1 ? 'bg-slate-300 text-black' :
                        idx === 2 ? 'bg-amber-700 text-white' :
                        'bg-white/5 text-slate-500'
                      )">
                        {{ idx + 1 }}
                      </span>
                    </div>
                  </td>
                  <td class="px-6 py-6">
                    <div class="flex items-center gap-4">
                       <div class="w-10 h-10 rounded-xl bg-white/5 border border-white/10 flex items-center justify-center text-xl overflow-hidden">
                          <template v-if="player.avatar && player.avatar.startsWith('data:')">
                            <img :src="player.avatar" class="w-full h-full object-cover" />
                          </template>
                          <template v-else>
                            {{ player.avatar || '🧪' }}
                          </template>
                       </div>
                       <div class="flex flex-col">
                          <span class="text-sm font-black text-white group-hover:text-blue-400 transition-colors">
                            {{ player.username }}
                            <span v-if="player.uid === user.uid" class="ml-2 text-[8px] bg-blue-600 px-1.5 py-0.5 rounded uppercase font-black tracking-widest">You</span>
                          </span>
                          <span class="text-[9px] font-mono text-slate-500 uppercase tracking-tighter mt-1">ID: {{ player.uid }}</span>
                       </div>
                    </div>
                  </td>
                  <td class="px-6 py-6">
                    <div class="flex flex-col">
                       <span class="text-lg font-black text-white font-mono tracking-tighter">
                         {{ player.points }}
                       </span>
                       <span class="text-[8px] font-black text-slate-600 uppercase tracking-widest mt-0.5">Current_Score</span>
                    </div>
                  </td>
                  <td class="px-6 py-6">
                    <div v-if="player.bounty > 0" class="flex flex-col">
                       <div class="flex items-center gap-2 text-rose-500">
                          <Flame class="w-3 h-3" />
                          <span class="text-sm font-black font-mono tracking-tighter">{{ player.bounty }}</span>
                       </div>
                       <span class="text-[8px] font-black text-rose-500/60 uppercase tracking-widest mt-1 italic">Active_Bounty_Detected</span>
                    </div>
                    <div v-else class="text-[9px] font-bold text-slate-600 uppercase italic opacity-40">
                      Standard_Protocol
                    </div>
                  </td>
                  <td class="px-8 py-6 text-right">
                    <button 
                      v-if="player.uid !== user.uid"
                      @click="openBountyModal(player)"
                      class="px-4 py-2 bg-rose-600/10 hover:bg-rose-600 text-rose-500 hover:text-white border border-rose-500/20 rounded-xl text-[9px] font-black uppercase tracking-widest transition-all active:scale-95 flex items-center gap-2 ml-auto"
                    >
                      <Crosshair class="w-3 h-3" />
                      Issue_Bounty
                    </button>
                    <div v-else class="text-[9px] font-black text-blue-500 uppercase tracking-widest italic pr-4">
                      Protocol_Master
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
       <div class="absolute inset-0 bg-black/80 backdrop-blur-md" @click="showBountyModal = false"></div>
       <div class="relative w-full max-w-md bg-[#121216] border border-rose-500/30 rounded-[40px] shadow-2xl overflow-hidden animate-in zoom-in duration-300">
          <div class="p-8 border-b border-white/5 flex items-center justify-between bg-rose-500/5">
             <div class="flex items-center gap-4">
                <div class="w-12 h-12 bg-rose-500/10 border border-rose-500/20 rounded-2xl flex items-center justify-center text-rose-500">
                   <Target class="w-6 h-6" />
                </div>
                <div>
                   <h2 class="text-xl font-black text-white uppercase tracking-tight">发布悬赏令</h2>
                   <p class="text-[9px] text-rose-500/60 font-mono uppercase tracking-widest mt-1">Configure_Target_Bounty</p>
                </div>
             </div>
             <button @click="showBountyModal = false" class="text-slate-500 hover:text-white transition-colors">
                <X class="w-6 h-6" />
             </button>
          </div>
          
          <div class="p-10 space-y-8 text-center">
             <div class="flex flex-col items-center">
                <div class="w-20 h-20 rounded-3xl bg-white/5 border border-white/10 flex items-center justify-center text-4xl mb-6 relative group">
                   {{ selectedTarget?.avatar || '🧪' }}
                   <div class="absolute -top-3 -right-3 w-8 h-8 bg-rose-600 rounded-full flex items-center justify-center border-4 border-[#121216]">
                      <Crosshair class="w-4 h-4 text-white" />
                   </div>
                </div>
                <p class="text-[10px] font-black uppercase tracking-widest text-slate-500">目标研究员</p>
                <h3 class="text-2xl font-black text-white mt-1">{{ selectedTarget?.username }}</h3>
             </div>

             <div class="space-y-4">
                <div class="flex justify-between items-center px-2">
                   <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest">投入科研积分</label>
                   <span class="text-[10px] text-rose-500 font-mono font-bold">{{ userPoints }} AVAILABLE</span>
                </div>
                <div class="relative">
                   <input 
                      v-model="bountyAmount" 
                      type="number" 
                      placeholder="输入积分数值..."
                      class="w-full h-16 bg-black/40 border border-white/10 text-white px-6 py-5 rounded-2xl focus:ring-1 focus:ring-rose-500 outline-none transition-all font-mono text-lg"
                   />
                   <div class="absolute right-4 top-1/2 -translate-y-1/2 text-rose-500/30 font-mono uppercase text-[10px] font-black tracking-widest pointer-events-none">
                     Points_Alpha
                   </div>
                </div>
                <p class="text-[9px] text-slate-500 leading-relaxed text-left px-2">
                  <span class="text-rose-500">协议说明：</span>悬赏积分将立刻从您的账户中扣除。任何人在该目标的竞技对局中获胜均可平分此项奖赏。悬赏一旦发布，不可撤回。
                </p>
             </div>

             <div class="flex gap-4 pt-4">
                <button 
                   @click="showBountyModal = false"
                   class="flex-1 h-14 bg-white/5 hover:bg-white/10 text-slate-500 font-bold rounded-2xl transition-all uppercase tracking-widest text-xs"
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { pointsAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import { Trophy, ArrowLeft, Loader2, Target, RefreshCw, ShieldCheck, Crosshair, Flame, X } from 'lucide-vue-next'
import { cn } from '../utils/cn'

const router = useRouter()
const { showAlert } = useDialog()
const user = ref(JSON.parse(localStorage.getItem('user') || '{}'))

const leaderboard = ref<any[]>([])
const loading = ref(true)
const userPoints = ref(0)

const showBountyModal = ref(false)
const selectedTarget = ref<any>(null)
const bountyAmount = ref<number | null>(null)
const submitting = ref(false)

const loadLeaderboard = async () => {
  try {
    loading.value = true
    const response = await pointsAPI.getLeaderboard()
    leaderboard.value = response.data
    
    // 同时也尝试更新一下本地的用户分数实时显示（如果有的话，在此简化为从榜单里找自己）
    const self = leaderboard.value.find(p => p.uid === user.value.uid)
    if (self) userPoints.value = self.points
  } catch (error) {
    console.error('Failed to load leaderboard:', error)
  } finally {
    loading.value = false
  }
}

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

onMounted(() => {
  loadLeaderboard()
})
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.1);
}
</style>
