<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { adminAPI } from '../utils/api'
import { 
  Shield, 
  ArrowLeft, 
  Users, 
  Layers, 
  History, 
  Trash2, 
  Edit2, 
  Save, 
  ChevronRight, 
  Terminal,
  Activity,
  Cpu,
  Database,
  Search as SearchIcon
} from 'lucide-vue-next'
import { cn } from '../utils/cn'

const router = useRouter()
const users = ref<any[]>([])
const gameHistory = ref<any[]>([])
const deckConfig = ref<any>(null)
const editingDeck = ref(false)
const activeTab = ref('users')
const loading = ref(false)
const searchTerm = ref('')

const loadData = async () => {
  loading.value = true
  try {
    if (activeTab.value === 'users') {
      const response = await adminAPI.getAllUsers()
      users.value = response.data || []
    } else if (activeTab.value === 'history') {
      const response = await adminAPI.getGameHistory()
      gameHistory.value = response.data || []
    } else if (activeTab.value === 'deck') {
      const response = await adminAPI.getGlobalDeckConfig()
      deckConfig.value = response.data
    }
  } catch (error) {
    console.error('加载数据失败:', error)
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

watch(activeTab, loadData)

const handleDeleteUser = async (userId: string) => {
  if (!window.confirm('确定要永久删除该研究员吗？此操作不可逆！')) return
  try {
    await adminAPI.deleteUser(userId)
    alert('用户已从实验室数据库抹除')
    loadData()
  } catch (error: any) {
    alert(error.response?.data?.error || '删除失败')
  }
}

const handleUpdateDeck = async () => {
  try {
    await adminAPI.updateGlobalDeckConfig(deckConfig.value.name, deckConfig.value.cards)
    alert('配置已生效并同步至全球')
    editingDeck.value = false
  } catch (error: any) {
    alert(error.response?.data?.error || '更新失败')
  }
}

const handleCardCountChange = (cardType: string, value: string) => {
  deckConfig.value = {
    ...deckConfig.value,
    cards: {
      ...deckConfig.value.cards,
      [cardType]: parseInt(value) || 0,
    },
  }
}

const filteredUsers = computed(() => {
  return users.value.filter(u => u.username.toLowerCase().includes(searchTerm.value.toLowerCase()))
})
</script>

<script lang="ts">
import { computed } from 'vue'
</script>

<template>
  <div class="min-h-screen bg-[#0a0a0c] text-slate-200 p-4 lg:p-10 font-sans selection:bg-orange-500/30">
    <div class="fixed inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] bg-orange-500/5 rounded-full blur-[120px]" />
      <div class="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-blue-500/5 rounded-full blur-[120px]" />
      <div class="absolute inset-0 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-20 brightness-50 contrast-150" />
    </div>

    <div class="max-w-7xl mx-auto relative z-10">
      <header class="flex flex-col lg:flex-row items-center justify-between gap-8 mb-12">
        <div class="flex items-center gap-6">
          <div class="relative group">
            <div class="absolute inset-x-0 inset-y-0 bg-orange-500 blur-2xl opacity-20 group-hover:opacity-40 transition-opacity" />
            <div class="w-16 h-16 rounded-2xl bg-[#111114] border border-orange-500/40 flex items-center justify-center relative z-10 shadow-2xl">
              <Shield class="w-8 h-8 text-orange-400 group-hover:scale-110 transition-transform" />
            </div>
          </div>
          <div>
            <h1 class="text-3xl font-black text-white italic tracking-tighter uppercase flex items-center gap-3">
              System Override <span class="text-xs font-mono bg-orange-500/20 text-orange-400 px-2 py-1 rounded border border-orange-500/30 not-italic">v4.0.2</span>
            </h1>
            <p class="text-slate-500 text-sm font-bold tracking-widest uppercase mt-1">实验室核心控制台 / Core Admin Console</p>
          </div>
        </div>

        <div class="flex items-center gap-4">
          <div class="px-6 py-3 bg-[#111114] border border-white/5 rounded-2xl flex items-center gap-4 shadow-xl">
            <div class="flex flex-col items-end">
              <span class="text-[10px] font-black text-slate-500 uppercase">Server Status</span>
              <span class="text-xs font-bold text-emerald-400 flex items-center gap-1.5">
                <span class="w-2 h-2 bg-emerald-500 rounded-full animate-pulse" />
                STABLE / OP-CON 1
              </span>
            </div>
            <div class="w-px h-8 bg-white/5" />
            <button 
              @click="router.push('/')"
              class="flex items-center gap-2 text-slate-400 hover:text-white transition-colors group"
            >
              <ArrowLeft class="w-4 h-4 group-hover:-translate-x-1 transition-transform" />
              <span class="text-xs font-black uppercase tracking-widest">Exit</span>
            </button>
          </div>
        </div>
      </header>

      <section class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-12">
        <div class="bg-[#111114] border border-white/10 p-6 rounded-[2rem] hover:border-white/20 transition-all shadow-xl group">
          <div class="flex items-center justify-between mb-4">
            <div class="p-3 rounded-2xl bg-blue-500/10 text-blue-400">
              <Users class="w-6 h-6" />
            </div>
            <div class="text-[10px] font-black text-slate-600 uppercase tracking-[0.2em] group-hover:text-slate-400 transition-colors">实时数据 / Live</div>
          </div>
          <div class="text-3xl font-black text-white italic">{{ users.length }}</div>
          <div class="text-xs font-bold text-slate-500 uppercase mt-1 tracking-wider">活跃研究员</div>
        </div>

        <div class="bg-[#111114] border border-white/10 p-6 rounded-[2rem] hover:border-white/20 transition-all shadow-xl group">
          <div class="flex items-center justify-between mb-4">
            <div class="p-3 rounded-2xl bg-purple-500/10 text-purple-400">
              <Cpu class="w-6 h-6" />
            </div>
            <div class="text-[10px] font-black text-slate-600 uppercase tracking-[0.2em] group-hover:text-slate-400 transition-colors">实时数据 / Live</div>
          </div>
          <div class="text-3xl font-black text-white italic">{{ deckConfig ? Object.keys(deckConfig.cards).length : 0 }}</div>
          <div class="text-xs font-bold text-slate-500 uppercase mt-1 tracking-wider">核心元素种类</div>
        </div>

        <div class="bg-[#111114] border border-white/10 p-6 rounded-[2rem] hover:border-white/20 transition-all shadow-xl group">
          <div class="flex items-center justify-between mb-4">
            <div class="p-3 rounded-2xl bg-orange-500/10 text-orange-400">
              <History class="w-6 h-6" />
            </div>
            <div class="text-[10px] font-black text-slate-600 uppercase tracking-[0.2em] group-hover:text-slate-400 transition-colors">实时数据 / Live</div>
          </div>
          <div class="text-3xl font-black text-white italic">{{ gameHistory.length }}</div>
          <div class="text-xs font-bold text-slate-500 uppercase mt-1 tracking-wider">实验记录总额</div>
        </div>

        <div class="bg-[#111114] border border-white/10 p-6 rounded-[2rem] hover:border-white/20 transition-all shadow-xl group">
          <div class="flex items-center justify-between mb-4">
            <div class="p-3 rounded-2xl bg-emerald-500/10 text-emerald-400">
              <Database class="w-6 h-6" />
            </div>
            <div class="text-[10px] font-black text-slate-600 uppercase tracking-[0.2em] group-hover:text-slate-400 transition-colors">实时数据 / Live</div>
          </div>
          <div class="text-3xl font-black text-white italic">12%</div>
          <div class="text-xs font-bold text-slate-500 uppercase mt-1 tracking-wider">系统负载</div>
        </div>
      </section>

      <main class="bg-[#111114] border border-white/10 rounded-[2.5rem] shadow-2xl overflow-hidden min-h-[600px] flex flex-col">
        <nav class="flex border-b border-white/5 bg-black/20 p-2">
          <button
            v-for="tab in [
              { id: 'users', label: '研究员名单 / PERSONNEL', icon: Users },
              { id: 'deck', label: '核心库存配置 / REDUCTION', icon: Layers },
              { id: 'history', label: '历史实验记录 / TRACING', icon: History }
            ]"
            :key="tab.id"
            @click="activeTab = tab.id"
            :class="cn(
              'flex items-center gap-3 px-8 py-5 text-xs font-black uppercase tracking-[0.1em] transition-all rounded-2xl relative',
              activeTab === tab.id 
                ? 'text-orange-400 bg-white/5' 
                : 'text-slate-500 hover:text-slate-300 hover:bg-white/5'
            )"
          >
            <component :is="tab.icon" class="w-4 h-4" />
            {{ tab.label }}
            <div v-if="activeTab === tab.id" class="absolute inset-x-0 bottom-2 px-8">
              <div class="h-0.5 bg-orange-500 shadow-[0_0_10px_rgba(249,115,22,0.5)] rounded-full" />
            </div>
          </button>
        </nav>

        <div class="p-10 flex-1">
          <div v-if="loading" class="h-full flex flex-col items-center justify-center text-slate-500 gap-6 py-20">
            <div class="relative">
              <div class="w-20 h-20 border-4 border-orange-500/20 border-t-orange-500 rounded-full animate-spin" />
              <Terminal class="w-8 h-8 text-orange-400 absolute inset-0 m-auto" />
            </div>
            <p class="font-mono text-sm uppercase tracking-widest animate-pulse">Synchronizing Database Layers...</p>
          </div>

          <div v-else class="animate-in fade-in slide-in-from-bottom-4 duration-500">
            <!-- Users Tab -->
            <div v-if="activeTab === 'users'" class="space-y-8">
              <div class="flex flex-col md:flex-row md:items-center justify-between gap-6">
                <h3 class="text-xl font-black italic uppercase text-white flex items-center gap-4">
                  <Terminal class="w-5 h-5 text-orange-400" />
                  研究员全局索引录 <span class="text-slate-600 font-mono not-italic text-xs">/ ROOT@ADMIN:~# list --all</span>
                </h3>
                <div class="relative group">
                  <SearchIcon class="w-4 h-4 absolute left-4 top-1/2 -translate-y-1/2 text-slate-600 group-focus-within:text-orange-400 transition-colors" />
                  <input 
                    v-model="searchTerm"
                    type="text" 
                    placeholder="SEARCH UID / USERNAME..."
                    class="bg-black/40 border border-white/5 rounded-2xl pl-12 pr-6 py-3 text-xs font-mono focus:outline-none focus:border-orange-500/30 w-full md:w-80 transition-all placeholder:text-slate-700"
                  />
                </div>
              </div>
              
              <div class="overflow-x-auto custom-scrollbar">
                <table class="w-full text-left">
                  <thead>
                    <tr class="text-slate-600 text-[10px] font-black uppercase tracking-[0.2em] border-b border-white/5">
                      <th class="px-6 py-4">Researcher Profile</th>
                      <th class="px-6 py-4">Recognition UID</th>
                      <th class="px-6 py-4">Auth Level</th>
                      <th class="px-6 py-4">Join Date</th>
                      <th class="px-6 py-4 text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-white/5 font-mono">
                    <tr v-for="u in filteredUsers" :key="u.id" class="hover:bg-white/5 transition-colors group">
                      <td class="px-6 py-6 text-sm font-bold text-white flex items-center gap-4">
                        <div class="w-10 h-10 bg-white/5 rounded-xl flex items-center justify-center text-xl group-hover:scale-110 transition-transform overflow-hidden">
                          <template v-if="u.avatar && u.avatar.startsWith('data:')">
                            <img :src="u.avatar" class="w-full h-full object-cover" />
                          </template>
                          <template v-else>
                            {{ u.avatar || '🧪' }}
                          </template>
                        </div>
                        {{ u.username }}
                      </td>
                      <td class="px-6 py-6 text-xs text-slate-500">{{ u.id }}</td>
                      <td class="px-6 py-6">
                        <span v-if="u.is_admin" class="text-[10px] px-3 py-1 bg-orange-500/10 text-orange-400 rounded-full border border-orange-500/20 font-black tracking-widest">LV.99 CORE</span>
                        <span v-else class="text-[10px] px-3 py-1 bg-white/5 text-slate-400 rounded-full border border-white/10 font-black tracking-widest">LV.01 STAFF</span>
                      </td>
                      <td class="px-6 py-6 text-xs text-slate-500">{{ new Date(u.created_at).toLocaleDateString() }}</td>
                      <td class="px-6 py-6 text-right">
                        <button 
                          v-if="!u.is_admin"
                          @click="handleDeleteUser(u.id)"
                          class="p-3 hover:bg-red-500/20 text-slate-600 hover:text-red-400 rounded-xl transition-all"
                          title="抹除权限"
                        >
                          <Trash2 class="w-5 h-5" />
                        </button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- Deck Tab -->
            <div v-if="activeTab === 'deck' && deckConfig" class="max-w-4xl mx-auto space-y-10">
              <div class="flex items-center justify-between mb-8">
                 <h3 class="text-xl font-black italic uppercase text-white flex items-center gap-4">
                  <Cpu class="w-5 h-5 text-blue-400" />
                  全局卡组配置面板 <span class="text-slate-600 font-mono not-italic text-xs">/ CONFIG@REDUCTION --GLOBAL</span>
                </h3>
                <button 
                  @click="editingDeck ? handleUpdateDeck() : (editingDeck = true)"
                  :class="cn(
                    'px-6 py-3 rounded-2xl font-black text-xs uppercase tracking-widest flex items-center gap-2 transition-all shadow-xl',
                    editingDeck 
                      ? 'bg-emerald-600 hover:bg-emerald-500 text-white' 
                      : 'bg-blue-600 hover:bg-blue-500 text-white'
                  )"
                >
                  <component :is="editingDeck ? Save : Edit2" class="w-4 h-4" />
                  {{ editingDeck ? "上传全局配置" : "进入编辑模式" }}
                </button>
              </div>

              <div class="grid grid-cols-2 lg:grid-cols-3 gap-6">
                <div v-for="[type, count] in Object.entries(deckConfig.cards)" :key="type" class="bg-[#1a1c1e] border border-white/5 p-6 rounded-3xl hover:border-blue-500/20 transition-all">
                  <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest block mb-4">{{ type }}</label>
                  <input
                    v-if="editingDeck"
                    type="number"
                    :value="count"
                    @input="(e: any) => handleCardCountChange(type, e.target.value)"
                    class="w-full bg-black/40 border border-blue-500/20 rounded-xl px-4 py-3 font-mono text-blue-400 focus:outline-none focus:border-blue-500"
                  />
                  <div v-else class="text-3xl font-black text-white italic">{{ count }}</div>
                </div>
              </div>
            </div>

            <!-- History Tab -->
            <div v-if="activeTab === 'history'" class="space-y-8">
              <h3 class="text-xl font-black italic uppercase text-white flex items-center gap-4">
                <History class="w-5 h-5 text-purple-400" />
                全球实验追溯记录 <span class="text-slate-600 font-mono not-italic text-xs">/ SCAN@LOGS --ALL</span>
              </h3>
              
              <div class="grid gap-4">
                <div v-if="gameHistory.length === 0" class="py-20 flex flex-col items-center justify-center border-2 border-dashed border-white/5 rounded-[2.5rem] text-slate-600">
                  <Activity class="w-12 h-12 mb-4 opacity-10" />
                  <p class="italic font-bold">目前尚未检索到任何历史实验数据</p>
                </div>
                <template v-else>
                  <div v-for="(game, i) in gameHistory" :key="i" class="bg-[#1a1c1e] border border-white/5 p-6 rounded-3xl hover:bg-white/5 transition-all group flex items-center justify-between">
                    <div class="flex items-center gap-6">
                      <div class="w-12 h-12 bg-white/5 rounded-2xl flex items-center justify-center font-mono text-purple-400 font-black">
                        #{{ i+1 }}
                      </div>
                      <div>
                        <div class="text-sm font-bold text-white group-hover:text-purple-400 transition-colors">实验反应堆: {{ game.id.substring(0, 8).toUpperCase() }}</div>
                        <div class="text-[10px] text-slate-500 uppercase font-black tracking-widest mt-1">
                          {{ new Date(game.created_at).toLocaleString() }} | 状态: 已结束
                        </div>
                      </div>
                    </div>
                    <ChevronRight class="w-5 h-5 text-slate-700 group-hover:text-purple-400 transition-all group-hover:translate-x-1" />
                  </div>
                </template>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>
