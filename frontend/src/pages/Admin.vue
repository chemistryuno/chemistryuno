<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { adminAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
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
  Search as SearchIcon,
  Key,
  ArrowUp,
  Plus,
  Star,
  MessageSquare
} from 'lucide-vue-next'
import { cn } from '../utils/cn'

const router = useRouter()
const { showAlert, showConfirm, showPrompt } = useDialog()
const users = ref<any[]>([])
const gameHistory = ref<any[]>([])
const feedbacks = ref<any[]>([])
const deckConfig = ref<any>(null)
const editingDeck = ref(false)
const deckCardsEdit = ref<{ key: string, value: number, id: string }[]>([])
const activeTab = ref('users')
const loading = ref(false)
const searchTerm = ref('')
const showCreateUserModal = ref(false)
const newUser = ref({ username: '', password: '' })

const specialElements = ['He', 'Ne', 'Ar', 'Kr', 'Au', '+2', '+4']

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
    } else if (activeTab.value === 'feedbacks') {
      const response = await adminAPI.getFeedbacks()
      feedbacks.value = response.data || []
    }
  } catch (error) {
    console.error('加载数据失败:', error)
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

watch(activeTab, () => {
  loadData()
  searchTerm.value = ''
})

const handleCreateUser = async () => {
  if (!newUser.value.username || !newUser.value.password) {
    await showAlert('请填写完整的用户信息', '输入错误')
    return
  }
  
  if (newUser.value.password.length < 6) {
    await showAlert('密码长度至少为6位', '密码错误')
    return
  }
  
  try {
    await adminAPI.createUser(newUser.value.username, newUser.value.password)
    await showAlert('用户创建成功', '成功')
    showCreateUserModal.value = false
    newUser.value = { username: '', password: '' }
    loadData()
  } catch (error: any) {
    await showAlert(error.response?.data?.error || '创建用户失败', '错误')
  }
}

const handleAcceptFeedback = async (id: number) => {
  try {
    const note = await showPrompt('处理说明（可留空，将使用默认文本）:', '输入说明', '处理反馈')
    await adminAPI.updateFeedbackStatus(id, 'accepted', note || '')
    await showAlert('反馈已接受', '已处理')
    loadData()
  } catch (error: any) {
    await showAlert(error.response?.data?.error || '操作失败', '错误')
  }
}

const handleDismissFeedback = async (id: number) => {
  try {
    const note = await showPrompt('处理说明（可留空，将使用默认文本）:', '输入说明', '处理反馈')
    await adminAPI.updateFeedbackStatus(id, 'dismissed', note || '')
    await showAlert('反馈已消除', '已处理')
    loadData()
  } catch (error: any) {
    await showAlert(error.response?.data?.error || '操作失败', '错误')
  }
}

const handleDeleteUser = async (userId: string) => {
  const confirmed = await showConfirm('确定要永久删除该研究员吗？此操作不可逆！', '⚠️ 危险操作')
  if (!confirmed) return
  
  try {
    await adminAPI.deleteUser(userId)
    await showAlert('用户已从实验室数据库抹除', '删除完成')
    loadData()
  } catch (error: any) {
    await showAlert(error.response?.data?.error || '删除失败', '错误')
  }
}

const handleChangePassword = async (userId: string) => {
  const newPassword = await showPrompt('请输入新密码（至少6位）:', '输入新密码...', '🔐 修改密码')
  if (!newPassword || newPassword.length < 6) {
    if (newPassword !== null) {
      await showAlert('密码长度至少为6位', '密码错误')
    }
    return
  }
  
  try {
    await adminAPI.changeUserPassword(userId, newPassword)
    await showAlert('密码修改成功', '成功')
  } catch (error: any) {
    await showAlert(error.response?.data?.error || '修改密码失败', '错误')
  }
}

const handlePromoteUser = async (userId: string, currentRole: string) => {
  const roles = ['user', 'co-worker', 'admin']
  const roleLabels = {
    'user': 'LV.01 STAFF (普通用户)',
    'co-worker': 'LV.50 CO-WORKER (化学助理)',
    'admin': 'LV.99 CORE (管理员)'
  }
  
  let message = '请选择新角色:\n\n'
  roles.forEach((role, index) => {
    message += `${index + 1}. ${roleLabels[role as keyof typeof roleLabels]}\n`
  })
  
  const choice = await showPrompt(message, '请输入数字(1-3)', '🚀 权限提升')
  if (!choice || !['1', '2', '3'].includes(choice)) {
    return
  }
  
  const newRole = roles[parseInt(choice) - 1]
  if (newRole === currentRole) {
    await showAlert('用户已是该角色', '提示')
    return
  }
  
  const confirmed = await showConfirm(`确定要将用户角色修改为 ${roleLabels[newRole as keyof typeof roleLabels]} 吗？`, '确认权限修改')
  if (!confirmed) {
    return
  }
  
  try {
    await adminAPI.promoteUser(userId, newRole)
    await showAlert('用户权限修改成功', '成功')
    loadData()
  } catch (error: any) {
    await showAlert(error.response?.data?.error || '修改权限失败', '错误')
  }
}

const toggleDeckEdit = () => {
  if (!editingDeck.value) {
    // 进入编辑模式：将当前卡组配置转换为可编辑数组
    deckCardsEdit.value = Object.entries(deckConfig.value.cards).map(([key, value]) => ({
      key,
      value: value as number,
      id: Math.random().toString(36).substr(2, 9)
    }))
    editingDeck.value = true
  } else {
    // 退出编辑模式且不保存
    editingDeck.value = false
  }
}

const handleAddDeckItem = () => {
  deckCardsEdit.value.push({
    key: '',
    value: 1,
    id: Math.random().toString(36).substr(2, 9)
  })
}

const handleRemoveDeckItem = (id: string) => {
  deckCardsEdit.value = deckCardsEdit.value.filter(item => item.id !== id)
}

const handleUpdateDeck = async () => {
  try {
    // 将数组重新转换为映射对象，同时去除空键名
    const newCards: Record<string, number> = {}
    deckCardsEdit.value.forEach(item => {
      if (item.key.trim()) {
        newCards[item.key.trim()] = item.value
      }
    })
    
    await adminAPI.updateGlobalDeckConfig(deckConfig.value.name, newCards)
    await showAlert('配置已生效并同步至全球', '🌐 配置更新成功')
    
    // 更新本地显示并退出编辑
    deckConfig.value.cards = newCards
    editingDeck.value = false
  } catch (error: any) {
    await showAlert(error.response?.data?.error || '更新失败', '错误')
  }
}

const handleCardCountChange = (cardType: string, value: string) => {
  // 此方法在新的数组驱动模式下可能不再直接使用，但保留以防万一
  if (deckConfig.value && !editingDeck.value) {
    deckConfig.value = {
      ...deckConfig.value,
      cards: {
        ...deckConfig.value.cards,
        [cardType]: parseInt(value) || 0,
      },
    }
  }
}

const filteredUsers = computed(() => {
  return users.value.filter(u => 
    u.username.includes(searchTerm.value) ||
    u.uid.toString().includes(searchTerm.value)
  )
})

const filteredDeck = computed(() => {
  if (!deckConfig.value) return []
  return Object.entries(deckConfig.value.cards)
    .filter(([type]) => type.includes(searchTerm.value))
    .filter(([type]) => !specialElements.includes(type))
    .sort((a, b) => a[0].localeCompare(b[0]))
})

const filteredSpecialDeck = computed(() => {
  if (!deckConfig.value) return []
  return Object.entries(deckConfig.value.cards)
    .filter(([type]) => type.includes(searchTerm.value))
    .filter(([type]) => specialElements.includes(type))
    .sort((a, b) => a[0].localeCompare(b[0]))
})

const filteredHistory = computed(() => {
  return [...gameHistory.value]
    .filter(game => game.id.includes(searchTerm.value))
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
})
</script>

<template>
  <div class="min-h-screen bg-slate-50 dark:bg-[#0a0a0c] text-slate-900 dark:text-slate-200 p-4 lg:p-10 font-sans selection:bg-orange-500/30">
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
              { id: 'special', label: '稀有元素配置 / SPECIALS', icon: Star },
              { id: 'feedbacks', label: '反馈报告 / FEEDBACK', icon: MessageSquare },
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
              <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-6">
                <h3 class="text-xl font-black italic uppercase text-white flex items-center gap-4">
                  <Terminal class="w-5 h-5 text-orange-400" />
                  研究员全局索引录 <span class="text-slate-600 font-mono not-italic text-xs">/ ROOT@ADMIN:~# list --all</span>
                </h3>
                <div class="flex items-center gap-4">
                  <div class="relative group">
                    <SearchIcon class="w-4 h-4 absolute left-4 top-1/2 -translate-y-1/2 text-slate-600 group-focus-within:text-orange-400 transition-colors" />
                    <input 
                      v-model="searchTerm"
                      type="text" 
                      placeholder="SEARCH UID / USERNAME..."
                      class="bg-black/40 border border-white/5 rounded-2xl pl-12 pr-6 py-3 text-xs font-mono focus:outline-none focus:border-orange-500/30 w-full md:w-64 transition-all placeholder:text-slate-700"
                    />
                  </div>
                  <button 
                    @click="showCreateUserModal = true"
                    class="px-6 py-3 bg-emerald-600 hover:bg-emerald-500 text-white rounded-2xl font-black text-xs uppercase tracking-widest flex items-center gap-2 transition-all shadow-xl whitespace-nowrap"
                  >
                    <Users class="w-4 h-4" />
                    添加用户
                  </button>
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
                    <tr v-for="u in filteredUsers" :key="u.uid" class="hover:bg-white/5 transition-colors group">
                      <td class="px-6 py-4 text-xs font-bold text-white flex items-center gap-3">
                        <div class="w-8 h-8 bg-white/5 rounded-lg flex items-center justify-center text-lg group-hover:scale-110 transition-transform overflow-hidden">
                          <template v-if="u.avatar && u.avatar.startsWith('data:')">
                            <img :src="u.avatar" class="w-full h-full object-cover" />
                          </template>
                          <template v-else>
                            {{ u.avatar || '🧪' }}
                          </template>
                        </div>
                        {{ u.username }}
                      </td>
                      <td class="px-6 py-4 text-[10px] text-slate-500">{{ u.uid }}</td>
                      <td class="px-6 py-4">
                        <span v-if="u.role === 'admin'" class="text-[9px] px-2 py-0.5 bg-orange-500/10 text-orange-400 rounded-md border border-orange-500/20 font-black tracking-widest">LV.99 CORE</span>
                        <span v-else-if="u.role === 'co-worker'" class="text-[9px] px-2 py-0.5 bg-blue-500/10 text-blue-400 rounded-md border border-blue-500/20 font-black tracking-widest">LV.50 CO-WORKER</span>
                        <span v-else class="text-[9px] px-2 py-0.5 bg-white/5 text-slate-400 rounded-md border border-white/10 font-black tracking-widest">LV.01 STAFF</span>
                      </td>
                      <td class="px-6 py-4 text-[10px] text-slate-500">{{ new Date(u.created_at).toLocaleDateString() }}</td>
                      <td class="px-6 py-4 text-right">
                        <div v-if="!u.is_admin" class="flex items-center gap-1 justify-end">
                          <button 
                            @click="handleChangePassword(u.uid)"
                            class="p-2 hover:bg-blue-500/20 text-slate-600 hover:text-blue-400 rounded-lg transition-all"
                            title="修改密码"
                          >
                            <Key class="w-3.5 h-3.5" />
                          </button>
                          <button 
                            @click="handlePromoteUser(u.uid, u.role)"
                            class="p-2 hover:bg-green-500/20 text-slate-600 hover:text-green-400 rounded-lg transition-all"
                            title="修改权限"
                          >
                            <ArrowUp class="w-3.5 h-3.5" />
                          </button>
                          <button 
                            @click="handleDeleteUser(u.uid)"
                            class="p-2 hover:bg-red-500/20 text-slate-600 hover:text-red-400 rounded-lg transition-all"
                            title="删除用户"
                          >
                            <Trash2 class="w-3.5 h-3.5" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- Deck Tab -->
            <div v-if="activeTab === 'deck' && deckConfig" class="space-y-6">
              <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
                <h3 class="text-lg font-black italic uppercase text-white flex items-center gap-3">
                  <Cpu class="w-5 h-5 text-blue-400" />
                  全局卡组配置 <span class="text-slate-600 font-mono not-italic text-[10px]">/ DECK@GLOBAL</span>
                </h3>
                <div class="flex items-center gap-3">
                  <div v-if="!editingDeck" class="relative group">
                    <SearchIcon class="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-slate-600 group-focus-within:text-blue-400 transition-colors" />
                    <input 
                      v-model="searchTerm"
                      type="text" 
                      placeholder="FILTER..."
                      class="bg-black/30 border border-white/5 rounded-xl pl-9 pr-4 py-2 text-[10px] font-mono focus:outline-none focus:border-blue-500/30 w-full md:w-48 transition-all"
                    />
                  </div>
                  <button 
                    v-if="editingDeck"
                    @click="handleAddDeckItem"
                    class="px-4 py-2 bg-blue-500/10 hover:bg-blue-500/20 text-blue-400 border border-blue-500/20 rounded-xl font-black text-[10px] uppercase tracking-widest flex items-center gap-2 transition-all"
                  >
                    <Plus class="w-3 h-3" />
                    添加元素
                  </button>
                  <button 
                    @click="editingDeck ? handleUpdateDeck() : toggleDeckEdit()"
                    :class="cn(
                      'px-4 py-2 rounded-xl font-black text-[10px] uppercase tracking-widest flex items-center gap-2 transition-all shadow-lg',
                      editingDeck 
                        ? 'bg-emerald-600 hover:bg-emerald-500 text-white' 
                        : 'bg-blue-600 hover:bg-blue-500 text-white'
                    )"
                  >
                    <component :is="editingDeck ? Save : Edit2" class="w-3.5 h-3.5" />
                    {{ editingDeck ? "保存配置" : "编辑" }}
                  </button>
                  <button 
                    v-if="editingDeck"
                    @click="editingDeck = false"
                    class="px-4 py-2 bg-white/5 hover:bg-white/10 text-slate-400 rounded-xl font-black text-[10px] uppercase tracking-widest transition-all"
                  >
                    取消
                  </button>
                </div>
              </div>

              <div class="overflow-x-auto custom-scrollbar border border-white/5 rounded-2xl bg-black/20">
                <table class="w-full text-left table-fixed">
                  <thead>
                    <tr class="text-slate-600 text-[9px] font-black uppercase tracking-[0.2em] border-b border-white/5">
                      <th class="px-6 py-3 w-[45%]">Element / Key</th>
                      <th class="px-6 py-3 w-[30%]">Quantity</th>
                      <th v-if="editingDeck" class="px-6 py-3 text-right w-[25%]">Ops</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-white/5 font-mono">
                    <template v-if="editingDeck">
                      <tr v-for="item in deckCardsEdit.filter(i => i.key === '' || !specialElements.includes(i.key))" :key="item.id" class="hover:bg-white/5 transition-colors">
                        <td class="px-6 py-3">
                          <input 
                            v-model="item.key"
                            type="text"
                            placeholder="元素符号..."
                            class="w-full bg-black/40 border border-white/10 rounded-lg px-3 py-1.5 focus:outline-none focus:border-blue-500/50 text-white text-xs"
                          />
                        </td>
                        <td class="px-6 py-3">
                          <input 
                            v-model.number="item.value"
                            type="number"
                            min="0"
                            class="w-full bg-black/40 border border-white/10 rounded-lg px-3 py-1.5 focus:outline-none focus:border-blue-500/50 text-blue-400 text-xs font-bold"
                          />
                        </td>
                        <td class="px-6 py-3 text-right">
                          <button @click="handleRemoveDeckItem(item.id)" class="p-2 hover:bg-red-500/20 text-slate-600 hover:text-red-400 rounded-lg transition-all">
                            <Trash2 class="w-3.5 h-3.5" />
                          </button>
                        </td>
                      </tr>
                    </template>
                    <template v-else>
                      <tr v-for="[type, count] in filteredDeck" :key="type" class="hover:bg-white/5 transition-colors group">
                        <td class="px-6 py-3 text-xs font-bold text-white">{{ type }}</td>
                        <td class="px-6 py-3 font-black text-sm italic text-blue-400">{{ count }}</td>
                      </tr>
                    </template>
                    <tr v-if="(!editingDeck && filteredDeck.length === 0) || (editingDeck && deckCardsEdit.filter(i => !specialElements.includes(i.key)).length === 0)">
                      <td colspan="3" class="py-12 text-center text-slate-700 text-[10px] font-bold uppercase tracking-widest italic">
                        No active elements found in matrix
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- Special Deck Tab -->
            <div v-if="activeTab === 'special' && deckConfig" class="space-y-6">
              <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
                <h3 class="text-lg font-black italic uppercase text-white flex items-center gap-3">
                  <Star class="w-5 h-5 text-yellow-400" />
                  稀有元素配置 <span class="text-slate-600 font-mono not-italic text-[10px]">/ DECK@SPECIALS</span>
                </h3>
                <div class="flex items-center gap-3">
                  <div v-if="!editingDeck" class="relative group">
                    <SearchIcon class="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-slate-600 group-focus-within:text-yellow-400 transition-colors" />
                    <input 
                      v-model="searchTerm"
                      type="text" 
                      placeholder="FILTER..."
                      class="bg-black/30 border border-white/5 rounded-xl pl-9 pr-4 py-2 text-[10px] font-mono focus:outline-none focus:border-yellow-500/30 w-full md:w-48 transition-all"
                    />
                  </div>
                  <button 
                    v-if="editingDeck"
                    @click="handleAddDeckItem"
                    class="px-4 py-2 bg-yellow-500/10 hover:bg-yellow-500/20 text-yellow-400 border border-yellow-500/20 rounded-xl font-black text-[10px] uppercase tracking-widest flex items-center gap-2 transition-all"
                  >
                    <Plus class="w-3 h-3" />
                    添加元素
                  </button>
                  <button 
                    @click="editingDeck ? handleUpdateDeck() : toggleDeckEdit()"
                    :class="cn(
                      'px-4 py-2 rounded-xl font-black text-[10px] uppercase tracking-widest flex items-center gap-2 transition-all shadow-lg',
                      editingDeck 
                        ? 'bg-emerald-600 hover:bg-emerald-500 text-white' 
                        : 'bg-yellow-600 hover:bg-yellow-500 text-white'
                    )"
                  >
                    <component :is="editingDeck ? Save : Edit2" class="w-3.5 h-3.5" />
                    {{ editingDeck ? "保存配置" : "编辑" }}
                  </button>
                  <button 
                    v-if="editingDeck"
                    @click="editingDeck = false"
                    class="px-4 py-2 bg-white/5 hover:bg-white/10 text-slate-400 rounded-xl font-black text-[10px] uppercase tracking-widest transition-all"
                  >
                    取消
                  </button>
                </div>
              </div>

              <div class="overflow-x-auto custom-scrollbar border border-white/5 rounded-2xl bg-black/20">
                <table class="w-full text-left table-fixed">
                  <thead>
                    <tr class="text-slate-600 text-[9px] font-black uppercase tracking-[0.2em] border-b border-white/5">
                      <th class="px-6 py-3 w-[45%]">Special Element</th>
                      <th class="px-6 py-3 w-[30%]">Quantity</th>
                      <th v-if="editingDeck" class="px-6 py-3 text-right w-[25%]">Ops</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-white/5 font-mono">
                    <template v-if="editingDeck">
                      <tr v-for="item in deckCardsEdit.filter(i => i.key === '' || specialElements.includes(i.key))" :key="item.id" class="hover:bg-white/5 transition-colors">
                        <td class="px-6 py-3">
                          <input 
                            v-model="item.key"
                            type="text"
                            placeholder="元素符号..."
                            class="w-full bg-black/40 border border-white/10 rounded-lg px-3 py-1.5 focus:outline-none focus:border-yellow-500/50 text-white text-xs font-bold"
                          />
                        </td>
                        <td class="px-6 py-3">
                          <input 
                            v-model.number="item.value"
                            type="number"
                            min="0"
                            class="w-full bg-black/40 border border-white/10 rounded-lg px-3 py-1.5 focus:outline-none focus:border-yellow-500/50 text-yellow-400 text-xs font-black"
                          />
                        </td>
                        <td class="px-6 py-3 text-right">
                          <button @click="handleRemoveDeckItem(item.id)" class="p-2 hover:bg-red-500/20 text-slate-600 hover:text-red-400 rounded-lg transition-all">
                            <Trash2 class="w-3.5 h-3.5" />
                          </button>
                        </td>
                      </tr>
                    </template>
                    <template v-else>
                      <tr v-for="[type, count] in filteredSpecialDeck" :key="type" class="hover:bg-white/5 transition-colors group">
                        <td class="px-6 py-3 text-xs font-bold text-yellow-400">{{ type }}</td>
                        <td class="px-6 py-3 font-black text-sm italic text-yellow-500">{{ count }}</td>
                      </tr>
                    </template>
                    <tr v-if="(!editingDeck && filteredSpecialDeck.length === 0) || (editingDeck && deckCardsEdit.filter(i => specialElements.includes(i.key)).length === 0)">
                      <td colspan="3" class="py-12 text-center text-slate-700 text-[10px] font-bold uppercase tracking-widest italic">
                        No special elements found in matrix
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- History Tab -->
            <div v-if="activeTab === 'history'" class="space-y-8">
              <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-6">
                <h3 class="text-xl font-black italic uppercase text-white flex items-center gap-4">
                  <History class="w-5 h-5 text-purple-400" />
                  全球实验追溯记录 <span class="text-slate-600 font-mono not-italic text-xs">/ SCAN@LOGS --ALL</span>
                </h3>
                <div class="flex items-center gap-4">
                  <div class="relative group">
                    <SearchIcon class="w-4 h-4 absolute left-4 top-1/2 -translate-y-1/2 text-slate-600 group-focus-within:text-purple-400 transition-colors" />
                    <input 
                      v-model="searchTerm"
                      type="text" 
                      placeholder="SEARCH EXPERIMENT ID..."
                      class="bg-black/40 border border-white/5 rounded-2xl pl-12 pr-6 py-3 text-xs font-mono focus:outline-none focus:border-purple-500/30 w-full md:w-64 transition-all placeholder:text-slate-700"
                    />
                  </div>
                </div>
              </div>

              <div class="overflow-x-auto custom-scrollbar">
                <table class="w-full text-left">
                  <thead>
                    <tr class="text-slate-600 text-[10px] font-black uppercase tracking-[0.2em] border-b border-white/5">
                      <th class="px-6 py-4">Experiment ID</th>
                      <th class="px-6 py-4">Timestamp / Logs</th>
                      <th class="px-6 py-4">Status</th>
                      <th class="px-6 py-4 text-right">Details</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-white/5 font-mono">
                    <tr v-for="game in filteredHistory" :key="game.id" class="hover:bg-white/5 transition-colors group cursor-pointer">
                      <td class="px-6 py-6 font-bold text-white group-hover:text-purple-400 transition-colors">
                        REACTOR-{{ game.id.substring(0, 8).toUpperCase() }}
                      </td>
                      <td class="px-6 py-6 text-xs text-slate-500">
                        {{ new Date(game.created_at).toLocaleString() }}
                      </td>
                      <td class="px-6 py-6">
                        <span class="text-[10px] px-3 py-1 bg-purple-500/10 text-purple-400 rounded-full border border-purple-500/20 font-black tracking-widest uppercase">Terminated</span>
                      </td>
                      <td class="px-6 py-6 text-right">
                        <ChevronRight class="w-5 h-5 text-slate-700 group-hover:text-purple-400 transition-all group-hover:translate-x-1 inline-block" />
                      </td>
                    </tr>
                    <tr v-if="filteredHistory.length === 0">
                      <td colspan="4" class="py-20 text-center text-slate-600 italic font-bold">
                        目前尚未检索到任何匹配的实验数据
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- Feedback Tab -->
            <div v-if="activeTab === 'feedbacks'" class="space-y-8">
              <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-6">
                <h3 class="text-xl font-black italic uppercase text-white flex items-center gap-4">
                  <MessageSquare class="w-5 h-5 text-blue-400" />
                  外部反馈报告 <span class="text-slate-600 font-mono not-italic text-xs">/ INCOMING@COMMS --FILTERED</span>
                </h3>
              </div>

              <div class="grid grid-cols-1 gap-6">
                <div v-for="fb in feedbacks" :key="fb.id" class="p-6 bg-black/20 border border-white/5 rounded-3xl hover:border-blue-500/30 transition-all flex flex-col gap-4">
                   <div class="flex items-center justify-between">
                      <div class="flex items-center gap-3">
                         <div class="w-10 h-10 rounded-xl bg-blue-500/10 flex items-center justify-center text-blue-400">
                            {{ fb.username[0].toUpperCase() }}
                         </div>
                         <div>
                            <p class="text-sm font-black text-white uppercase">{{ fb.username }}</p>
                            <p class="text-[10px] text-slate-500 font-mono">{{ new Date(fb.created_at).toLocaleString() }}</p>
                         </div>
                      </div>
                      <span class="px-3 py-1 bg-white/5 text-slate-500 text-[9px] font-black uppercase tracking-widest rounded-full border border-white/5">{{ fb.page }}</span>
                   </div>
                   <p class="text-sm leading-relaxed text-slate-300 font-medium bg-white/5 p-4 rounded-2xl italic">“{{ fb.content }}”</p>
                   <div class="flex items-center justify-end gap-3">
                      <button @click="handleAcceptFeedback(fb.id)" class="px-3 py-1 bg-emerald-500 text-white rounded-md text-sm">接受</button>
                      <button @click="handleDismissFeedback(fb.id)" class="px-3 py-1 bg-red-500 text-white rounded-md text-sm">消除</button>
                   </div>
                </div>
                <div v-if="feedbacks.length === 0" class="py-20 text-center text-slate-600 italic font-bold">
                  目前尚未收到任何外部反馈报告
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>

    <!-- 创建用户模态框 -->
    <div v-if="showCreateUserModal" class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div class="bg-[#111114] border border-white/10 rounded-[2rem] p-8 max-w-md w-full shadow-2xl">
        <h3 class="text-xl font-black text-white mb-6 flex items-center gap-3">
          <Users class="w-5 h-5 text-emerald-400" />
          添加新研究员
        </h3>
        
        <div class="space-y-6">
          <div>
            <label class="text-xs font-black text-slate-500 uppercase tracking-widest block mb-2">用户名</label>
            <input 
              v-model="newUser.username"
              type="text" 
              placeholder="输入用户名..."
              class="w-full bg-black/40 border border-white/5 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-emerald-500/30 transition-all"
            />
          </div>
          
          <div>
            <label class="text-xs font-black text-slate-500 uppercase tracking-widest block mb-2">密码</label>
            <input 
              v-model="newUser.password"
              type="password" 
              placeholder="输入密码（至少6位）..."
              class="w-full bg-black/40 border border-white/5 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-emerald-500/30 transition-all"
            />
          </div>
        </div>
        
        <div class="flex gap-4 mt-8">
          <button 
            @click="showCreateUserModal = false; newUser = { username: '', password: '' }"
            class="flex-1 px-6 py-3 bg-white/5 hover:bg-white/10 text-slate-400 hover:text-white rounded-xl font-black text-xs uppercase tracking-widest transition-all"
          >
            取消
          </button>
          <button 
            @click="handleCreateUser"
            class="flex-1 px-6 py-3 bg-emerald-600 hover:bg-emerald-500 text-white rounded-xl font-black text-xs uppercase tracking-widest transition-all"
          >
            创建用户
          </button>
        </div>
      </div>
    </div>
  </div>
</template>


<style scoped>
.animate-in {
  animation: fadeInScale 0.3s ease-out;
}

@keyframes fadeInScale {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.zoom-in {
  animation: zoomIn 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes zoomIn {
  from {
    opacity: 0;
    transform: scale(0.3);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 3px;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.2);
}
</style>
