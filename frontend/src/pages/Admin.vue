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
  ArrowUp,
  Plus,
  Star,
  MessageSquare,
  Trophy,
  Megaphone,
  Clock,
  Ban,
  UserMinus,
  X,
  Puzzle
} from 'lucide-vue-next'
import { cn } from '../utils/cn'

const router = useRouter()
const { showAlert, showConfirm, showPrompt } = useDialog()
const users = ref<any[]>([])
const gameHistory = ref<any[]>([])
const feedbacks = ref<any[]>([])
const announcements = ref<any[]>([])

const gameTimeConfigs = ref<any>({
  player_kick_timeout: '30',
  player_action_timeout: '30',
  auto_start_timeout: '10',
  half_ready_timeout: '60',
  reconnect_grace_period: '30',
  points_scaling_enabled: 'true'
})
const deckConfig = ref<any>(null)
const editingDeck = ref(false)
const deckCardsEdit = ref<{ key: string, value: number, id: string }[]>([])
const initialCardsEdit = ref(10)
const activeTab = ref('users')
const loading = ref(false)

const tabs = computed(() => {
  const allTabs = [
    { id: 'users', label: '研究员', icon: Users },
    { id: 'deck', label: '核心库存', icon: Layers },
    { id: 'special', label: '稀有元素', icon: Star },
    { id: 'announcements', label: '广播推送', icon: Megaphone },
    { id: 'feedbacks', label: '通讯报告', icon: MessageSquare },
    { id: 'game-time', label: '时间配置', icon: Clock },
    { id: 'history', label: '实验日志', icon: History }
  ]
  
  const user = JSON.parse(localStorage.getItem('user') || '{}')
  if (user.role === 'co-worker') {
    return allTabs.filter(tab => tab.id === 'users')
  }
  return allTabs
})

const searchTerm = ref('')
const showCreateAnnouncementModal = ref(false)
const showEditAnnouncementModal = ref(false)
const editingAnnouncement = ref<any>(null)
const newAnnouncement = ref({
  title: '',
  content: '',
  type: 'info',
  is_ticker: true,
  is_persistent: false,
  expires_in: '24h',
  on_join: false,
  cron_interval: 0,
  close_delay: 0
})

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
    } else if (activeTab.value === 'deck' || activeTab.value === 'special') {
      const response = await adminAPI.getGlobalDeckConfig()
      deckConfig.value = response.data
    } else if (activeTab.value === 'feedbacks') {
      const response = await adminAPI.getFeedbacks()
      feedbacks.value = (response.data || []).filter((fb: any) => fb.status === 'pending')
    } else if (activeTab.value === 'announcements') {
      const response = await adminAPI.getAnnouncements()
      announcements.value = response.data || []
    } else if (activeTab.value === 'game-time') {
      const response = await adminAPI.getGameTimeConfigs()
      if (response.data?.configs) {
        gameTimeConfigs.value = response.data.configs
      }
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
    // 如果用户点击取消，中断操作
    if (note === null) return

    await adminAPI.updateFeedbackStatus(id, 'dismissed', note || '')
    await showAlert('反馈已消除', '已处理')
    loadData()
  } catch (error: any) {
    await showAlert(error.response?.data?.error || '操作失败', '错误')
  }
}

const parseReportUID = (content: string): number | null => {
  const match = content.match(/UID:\s*(\d+)/)
  return match ? parseInt(match[1]) : null
}

const parseReportUsername = (content: string): string | null => {
  const match = content.match(/举报用户:\s*(.+?)\s*\(UID:/)
  return match ? match[1] : null
}

const handleBanReportedPlayer = (fb: any) => {
  const uid = parseReportUID(fb.content)
  const username = parseReportUsername(fb.content)
  if (!uid) {
    showAlert('无法从举报内容中解析被举报玩家的UID', '解析失败')
    return
  }
  openBanModal({ uid, username: username || `UID:${uid}` })
}

const handlePromoteUser = async (uid: string, currentRole: string) => {
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
    await adminAPI.promoteUser(uid, newRole)
    await showAlert('用户权限修改成功', '成功')
    loadData()
  } catch (error: any) {
    await showAlert(error.response?.data?.error || '修改权限失败', '错误')
  }
}

// Ban/Kick state
const showBanModal = ref(false)
const banTarget = ref<any>(null)
const banUntil = ref('')
const banReason = ref('您由于不正当游戏而被封禁')

// 生成默认封禁时间（当前时间 + 24小时），格式为 datetime-local 兼容格式
const formatDatetimeLocal = (d: Date) => {
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0') + 'T' + String(d.getHours()).padStart(2, '0') + ':' + String(d.getMinutes()).padStart(2, '0')
}

const getDefaultBanUntil = () => {
  return formatDatetimeLocal(new Date(Date.now() + 24 * 60 * 60 * 1000))
}

const banPresets = [
  { label: '1小时', hours: 1 },
  { label: '6小时', hours: 6 },
  { label: '24小时', hours: 24 },
  { label: '3天', hours: 72 },
  { label: '7天', hours: 168 },
  { label: '30天', hours: 720 },
  { label: '永久', hours: 87600 }, // ~10年
]

const selectedPreset = ref<number | null>(24)

const setBanDuration = (hours: number) => {
  selectedPreset.value = hours
  banUntil.value = formatDatetimeLocal(new Date(Date.now() + hours * 3600 * 1000))
}

const handleKickPlayer = async (user: any) => {
  const displayName = user.nickname || user.username
  const reason = await showPrompt(`踢出研究员 ${displayName} (UID: ${user.uid})\n请输入踢出原因:`, '违规行为...', '踢出玩家')
  if (reason === null) return

  try {
    await adminAPI.kickPlayer(user.uid, reason || '您由于不正当游戏而被踢出')
    await showAlert(`已将 ${displayName} 踢出服务器`, '操作完成')
  } catch (error: any) {
    await showAlert(error.response?.data?.error || '踢出失败', '操作失败')
  }
}

const openBanModal = (user: any) => {
  banTarget.value = user
  if (isBanned(user)) {
    // Already banned - pre-fill with existing ban data
    selectedPreset.value = null
    banUntil.value = formatDatetimeLocal(new Date(user.banned_until))
    banReason.value = user.ban_reason || '您由于不正当游戏而被封禁'
  } else {
    selectedPreset.value = 24
    banUntil.value = getDefaultBanUntil()
    banReason.value = '您由于不正当游戏而被封禁'
  }
  showBanModal.value = true
}

const handleBanUser = async () => {
  if (!banTarget.value) return
  if (!banUntil.value) {
    await showAlert('请选择封禁截止时间', '参数缺失')
    return
  }
  const until = new Date(banUntil.value)
  if (until <= new Date()) {
    await showAlert('封禁截止时间必须晚于当前时间', '时间无效')
    return
  }
  try {
    await adminAPI.banUser(banTarget.value.uid, until.toISOString(), banReason.value || '违规行为')
    await showAlert(`已封禁 ${banTarget.value.nickname || banTarget.value.username} 至 ${until.toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' })}（UTC+8）`, '封禁执行完成')
    showBanModal.value = false
    banTarget.value = null
  } catch (error: any) {
    await showAlert(error.response?.data?.error || '封禁失败', '操作失败')
  }
}

const handleCreateAnnouncement = async () => {
  if (!newAnnouncement.value.content) {
    await showAlert('请输入公告内容')
    return
  }
  try {
    await adminAPI.createAnnouncement(
      newAnnouncement.value.title,
      newAnnouncement.value.content,
      newAnnouncement.value.type,
      newAnnouncement.value.is_ticker,
      newAnnouncement.value.expires_in,
      newAnnouncement.value.on_join,
      newAnnouncement.value.cron_interval,
      newAnnouncement.value.close_delay,
      newAnnouncement.value.is_persistent
    )
    await showAlert('公告发布成功', '同步中...')
    showCreateAnnouncementModal.value = false
    newAnnouncement.value = { title: '', content: '', type: 'info', is_ticker: true, is_persistent: false, expires_in: '24h', on_join: false, cron_interval: 0, close_delay: 0 }
    loadData()
  } catch (err: any) {
    await showAlert(err.response?.data?.error || '发布失败')
  }
}

const handleToggleAnnouncement = async (id: number, active: boolean) => {
  try {
    await adminAPI.updateAnnouncementStatus(id, active)
    loadData()
  } catch (err: any) {
    await showAlert(err.response?.data?.error || '状态切换失败')
  }
}

const handleDeleteAnnouncement = async (id: number) => {
  const ok = await showConfirm('确定要永久删除这条公告吗？')
  if (!ok) return
  try {
    await adminAPI.deleteAnnouncement(id)
    loadData()
  } catch (err: any) {
    await showAlert(err.response?.data?.error || '删除失败')
  }
}

const handleEditAnnouncement = (ann: any) => {
  editingAnnouncement.value = {
    id: ann.id,
    title: ann.title || '',
    content: ann.content,
    type: ann.type,
    is_ticker: ann.is_ticker,
    is_persistent: ann.is_persistent,
    on_join: ann.on_join,
    cron_interval: ann.cron_interval,
    close_delay: ann.close_delay,
    expires_in: ann.expires_at ? '' : '24h'
  }
  showEditAnnouncementModal.value = true
}

const handleUpdateAnnouncement = async () => {
  if (!editingAnnouncement.value.content) {
    await showAlert('请输入公告内容')
    return
  }
  try {
    await adminAPI.updateAnnouncement(
      editingAnnouncement.value.id,
      editingAnnouncement.value.title,
      editingAnnouncement.value.content,
      editingAnnouncement.value.type,
      editingAnnouncement.value.is_ticker,
      editingAnnouncement.value.expires_in,
      editingAnnouncement.value.on_join,
      editingAnnouncement.value.cron_interval,
      editingAnnouncement.value.close_delay,
      editingAnnouncement.value.is_persistent
    )
    await showAlert('公告更新成功', '同步完成')
    showEditAnnouncementModal.value = false
    editingAnnouncement.value = null
    loadData()
  } catch (err: any) {
    await showAlert(err.response?.data?.error || '更新失败')
  }
}

const handleUpdateGameTimeConfig = async () => {
  try {
    const data = {
      player_kick_timeout: parseInt(gameTimeConfigs.value.player_kick_timeout),
      player_action_timeout: parseInt(gameTimeConfigs.value.player_action_timeout),
      auto_start_timeout: parseInt(gameTimeConfigs.value.auto_start_timeout),
      half_ready_timeout: parseInt(gameTimeConfigs.value.half_ready_timeout),
      reconnect_grace_period: parseInt(gameTimeConfigs.value.reconnect_grace_period),
      points_scaling_enabled: String(gameTimeConfigs.value.points_scaling_enabled)
    }

    await adminAPI.updateGameTimeConfig(data)
    await showAlert('游戏时间配置已更新，将在新游戏中生效', '成功')
    loadData()
  } catch (error: any) {
    await showAlert(error.response?.data?.error || '更新配置失败', '错误')
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
    initialCardsEdit.value = deckConfig.value.initial_cards || 10
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
    
    await adminAPI.updateGlobalDeckConfig(deckConfig.value.name, newCards, initialCardsEdit.value)
    await showAlert('配置已生效并同步至全球', '🌐 配置更新成功')
    
    // 更新本地显示并退出编辑
    deckConfig.value.cards = newCards
    deckConfig.value.initial_cards = initialCardsEdit.value
    editingDeck.value = false
  } catch (error: any) {
    await showAlert(error.response?.data?.error || '更新失败', '错误')
  }
}

// handleCardCountChange was removed because it was unused
const isBanned = (u: any) => {
  return u.banned_until && new Date(u.banned_until) > new Date()
}

const filteredUsers = computed(() => {
  return users.value.filter(u =>
    (u.nickname && u.nickname.includes(searchTerm.value)) ||
    (u.username && u.username.includes(searchTerm.value)) ||
    (u.uid && u.uid.toString().includes(searchTerm.value))
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
  const list = Array.isArray(gameHistory.value) ? [...gameHistory.value] : []
  return list
    .filter(game => 
      String(game.id).includes(searchTerm.value) || 
      (game.room_id && game.room_id.toLowerCase().includes(searchTerm.value.toLowerCase())) ||
      (game.winner_name && game.winner_name.toLowerCase().includes(searchTerm.value.toLowerCase()))
    )
    .sort((a, b) => {
      const dateB = new Date((b.finished_at || b.created_at || '').replace(' ', 'T')).getTime() || 0
      const dateA = new Date((a.finished_at || a.created_at || '').replace(' ', 'T')).getTime() || 0
      return dateB - dateA
    })
})
</script>

<template>
  <div class="min-h-screen bg-slate-50 dark:bg-[#070708] text-slate-900 dark:text-slate-200 p-3 lg:p-4 font-sans selection:bg-cyan-500/30">
    <div class="fixed inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] bg-cyan-500/5 rounded-full blur-[120px] animate-pulse" />
      <div class="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-orange-500/5 rounded-full blur-[120px]" />
      <div class="absolute inset-0 bg-[url('/noise.svg')] opacity-20 brightness-50 contrast-150 mix-blend-overlay" />
    </div>

    <div class="max-w-7xl mx-auto relative z-10">
      <header class="flex flex-col lg:flex-row items-center justify-between gap-4 mb-6">
        <div class="flex items-center gap-4">
          <div class="relative group">
            <div class="absolute inset-x-0 inset-y-0 bg-cyan-500 blur-2xl opacity-20 group-hover:opacity-40 transition-opacity" />
            <div class="w-16 h-16 rounded-xl bg-white dark:bg-[#111114] border border-cyan-500/40 flex items-center justify-center relative z-10 shadow-[0_0_30px_rgba(6,182,212,0.15)] group-hover:shadow-[0_0_40px_rgba(6,182,212,0.25)] transition-all">
              <Shield class="w-8 h-8 text-cyan-600 dark:text-cyan-400 group-hover:scale-110 transition-transform" />
            </div>
          </div>
          <div>
            <h1 class="text-xl font-black text-slate-900 dark:text-white italic tracking-tighter uppercase flex items-center gap-3">
              Core Protocol <span class="text-[10px] font-mono bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 px-2 py-1 rounded-sm border border-cyan-500/20 not-italic tracking-normal">LEVEL 4 SECURED</span>
            </h1>
            <p class="text-slate-400 dark:text-slate-500 text-[10px] font-black tracking-[0.2em] uppercase mt-1">实验室中枢神经系统 / Central Neural Console</p>
          </div>
        </div>

        <div class="flex items-center gap-4">
          <div class="px-4 py-2.5 bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/5 rounded-xl flex items-center gap-4 shadow-xl backdrop-blur-md">
            <div class="flex flex-col items-end">
              <span class="text-[9px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">Interface Status</span>
              <span class="text-xs font-bold text-cyan-600 dark:text-cyan-400 flex items-center gap-1.5">
                <span class="w-1.5 h-1.5 bg-cyan-500 rounded-full animate-ping" />
                ENCRYPTED / NODE-01
              </span>
            </div>
            <div class="w-px h-8 bg-slate-200 dark:bg-white/5" />
            <router-link 
              to="/ranking"
              class="flex items-center gap-2 text-slate-400 hover:text-cyan-600 dark:hover:text-cyan-400 transition-all group"
            >
              <Trophy class="w-4 h-4 group-hover:rotate-12 transition-transform" />
              <span class="text-[10px] font-black uppercase tracking-widest">Archive</span>
            </router-link>
            <div class="w-px h-8 bg-slate-200 dark:bg-white/5" />
            <button 
              @click="router.push('/')"
              class="flex items-center gap-2 text-slate-400 dark:text-slate-400 hover:text-red-500 dark:hover:text-red-400 transition-colors group"
            >
              <ArrowLeft class="w-4 h-4 group-hover:-translate-x-1 transition-transform" />
              <span class="text-[10px] font-black uppercase tracking-widest">BACK</span>
            </button>
          </div>
        </div>
      </header>

      <section class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 p-3 rounded-xl hover:border-cyan-500/40 transition-all shadow-lg dark:shadow-xl group">
          <div class="flex items-center justify-between mb-3">
            <div class="p-2 rounded-lg bg-cyan-500/10 text-cyan-600 dark:text-cyan-400">
              <Users class="w-5 h-5" />
            </div>
            <div class="text-[9px] font-black text-slate-400 dark:text-slate-600 uppercase tracking-[0.2em] group-hover:text-cyan-600 transition-colors">STAFF_INDEX</div>
          </div>
          <div class="text-xl font-black text-slate-900 dark:text-white italic tracking-tighter">{{ users.length }}</div>
          <div class="text-[10px] font-bold text-slate-400 dark:text-slate-500 uppercase mt-1 tracking-wider">在册研究员总数</div>
        </div>

        <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 p-3 rounded-xl hover:border-violet-500/40 transition-all shadow-lg dark:shadow-xl group">
          <div class="flex items-center justify-between mb-3">
            <div class="p-2 rounded-lg bg-violet-500/10 text-violet-600 dark:text-violet-400">
              <Cpu class="w-5 h-5" />
            </div>
            <div class="text-[9px] font-black text-slate-400 dark:text-slate-600 uppercase tracking-[0.2em] group-hover:text-violet-600 transition-colors">CORE_DRIVE</div>
          </div>
          <div class="text-xl font-black text-slate-900 dark:text-white italic tracking-tighter">{{ deckConfig ? Object.keys(deckConfig.cards).length : 0 }}</div>
          <div class="text-[10px] font-bold text-slate-400 dark:text-slate-500 uppercase mt-1 tracking-wider">核定反应元基数</div>
        </div>

        <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 p-3 rounded-xl hover:border-orange-500/40 transition-all shadow-lg dark:shadow-xl group">
          <div class="flex items-center justify-between mb-3">
            <div class="p-2 rounded-lg bg-orange-500/10 text-orange-600 dark:text-orange-400">
              <History class="w-5 h-5" />
            </div>
            <div class="text-[9px] font-black text-slate-400 dark:text-slate-600 uppercase tracking-[0.2em] group-hover:text-orange-600 transition-colors">LOG_BUFFER</div>
          </div>
          <div class="text-xl font-black text-slate-900 dark:text-white italic tracking-tighter">{{ gameHistory.length }}</div>
          <div class="text-[10px] font-bold text-slate-400 dark:text-slate-500 uppercase mt-1 tracking-wider">全域实验活动记录</div>
        </div>

        <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 p-3 rounded-xl hover:border-emerald-500/40 transition-all shadow-lg dark:shadow-xl group">
          <div class="flex items-center justify-between mb-3">
            <div class="p-2 rounded-lg bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
              <Database class="w-5 h-5" />
            </div>
            <div class="text-[9px] font-black text-slate-400 dark:text-slate-600 uppercase tracking-[0.2em] group-hover:text-emerald-600 transition-colors">SYSCAL_LOAD</div>
          </div>
          <div class="text-xl font-black text-slate-900 dark:text-white italic tracking-tighter">0.02%</div>
          <div class="text-[10px] font-bold text-slate-400 dark:text-slate-500 uppercase mt-1 tracking-wider">服务器负载指数</div>
        </div>
      </section>

      <main class="bg-white dark:bg-[#0c0c0e] border border-slate-200 dark:border-white/5 rounded-[1.25rem] shadow-[0_40px_100px_-20px_rgba(0,0,0,0.3)] overflow-hidden min-h-[600px] flex flex-col relative">
        <div class="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-cyan-500/50 to-transparent" />
        
        <nav class="flex border-b border-slate-100 dark:border-white/5 bg-slate-50/50 dark:bg-black/40 p-3 overflow-x-auto custom-scrollbar relative">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            @click="activeTab = tab.id"
            class="flex items-center gap-3 px-5 py-2.5 rounded-xl transition-all shrink-0 group relative"
            :class="[
              activeTab === tab.id 
                ? 'bg-cyan-500/10 text-cyan-600 dark:text-cyan-400' 
                : 'text-slate-400 dark:text-slate-500 hover:text-slate-900 dark:hover:text-slate-300'
            ]"
          >
            <component :is="tab.icon" :class="cn('w-4 h-4 transition-transform group-hover:scale-110', activeTab === tab.id ? 'text-cyan-500 animate-pulse' : '')" />
            <span class="font-black uppercase tracking-widest text-[10px]">{{ tab.label }}</span>
            <div v-if="activeTab === tab.id" class="absolute inset-x-0 bottom-1 px-5 z-10">
              <div class="h-0.5 bg-cyan-500 shadow-[0_0_15px_rgba(6,182,212,0.8)] rounded-full" />
            </div>
          </button>

          <div class="ml-auto flex items-center gap-2 pr-2">
            <router-link
              to="/admin/plugins"
              class="flex items-center gap-3 px-4 py-2.5 rounded-xl transition-all shrink-0 group border border-purple-500/20 bg-purple-500/10 text-purple-600 dark:text-purple-400 hover:bg-purple-500/20"
            >
              <Puzzle class="w-4 h-4 transition-transform group-hover:scale-110" />
              <span class="font-black uppercase tracking-widest text-[10px]">安装插件</span>
            </router-link>
          </div>
        </nav>

        <div class="p-4 flex-1">
          <div v-if="loading" class="h-full flex flex-col items-center justify-center text-slate-500 gap-4 py-20 relative overflow-hidden">
            <div class="absolute inset-0 bg-cyan-500/5 blur-[100px] opacity-20 animate-pulse" />
            <div class="relative">
              <div class="w-24 h-24 border-2 border-cyan-500/10 border-t-cyan-500 rounded-full animate-spin shadow-[0_0_20px_rgba(6,182,212,0.1)]" />
              <Terminal class="w-10 h-10 text-cyan-400 absolute inset-0 m-auto" />
            </div>
            <div class="flex flex-col items-center gap-2">
              <p class="font-mono text-[10px] uppercase tracking-[0.5em] text-cyan-500/60 animate-pulse">Establishing Secure Uplink...</p>
              <p class="font-mono text-[8px] uppercase tracking-widest text-slate-600">Syncing database layers. Please hold.</p>
            </div>
          </div>

          <div v-else class="animate-in fade-in slide-in-from-bottom-4 duration-500">
            <!-- Users Tab -->
            <div v-if="activeTab === 'users'" class="space-y-8">
              <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
                <h3 class="text-lg font-black italic uppercase text-slate-900 dark:text-white flex items-center gap-4">
                  <Terminal class="w-5 h-5 text-cyan-500 shrink-0" />
                  研究员全局索引录 <span class="text-slate-400 dark:text-slate-600 font-mono not-italic text-[10px] tracking-normal">/ STAFF@CORE --DIRECTORY</span>
                </h3>
                <div class="flex items-center gap-4">
                  <div class="relative group">
                    <SearchIcon class="w-4 h-4 absolute left-4 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-cyan-500 transition-colors" />
                    <input 
                      v-model="searchTerm"
                      type="text" 
                      placeholder="SEARCH UID / USERNAME..."
                      class="bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/5 rounded-xl pl-12 pr-6 py-2.5 text-[10px] font-black tracking-widest focus:outline-none focus:border-cyan-500/30 w-full md:w-64 transition-all placeholder:text-slate-400 dark:placeholder:text-slate-700 text-slate-900 dark:text-white"
                    />
                  </div>
                </div>
              </div>
              
              <div class="overflow-x-auto custom-scrollbar">
                <table class="w-full text-left">
                  <thead>
                    <tr class="text-slate-400 dark:text-slate-600 text-[9px] font-black uppercase tracking-[0.3em] border-b border-slate-100 dark:border-white/5">
                      <th class="px-4 py-2.5">Researcher Profile</th>
                      <th class="px-4 py-2.5">Recognition UID</th>
                      <th class="px-4 py-2.5">Auth Level</th>
                      <th class="px-4 py-2.5">Join Date</th>
                      <th class="px-4 py-2.5 text-right">Overrides</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-100 dark:divide-white/5 font-mono">
                    <tr v-for="u in filteredUsers" :key="u.uid" class="hover:bg-slate-50 dark:hover:bg-white/[0.02] transition-colors group">
                      <td class="px-4 py-2.5 text-xs font-bold text-slate-900 dark:text-white flex items-center gap-4">
                        <div class="w-10 h-10 bg-white dark:bg-black/40 rounded-xl flex items-center justify-center text-lg group-hover:scale-105 transition-transform overflow-hidden border border-slate-200 dark:border-white/10 shadow-sm">
                          <template v-if="u.avatar && u.avatar.startsWith('data:')">
                            <img :src="u.avatar" class="w-full h-full object-cover" />
                          </template>
                          <template v-else>
                            {{ u.avatar || '🧪' }}
                          </template>
                        </div>
                        <div class="flex flex-col">
                          <span class="group-hover:text-cyan-600 transition-colors uppercase tracking-tight text-[10px] font-black flex items-center gap-1.5">
                            {{ u.nickname || u.username }}
                            <span v-if="isBanned(u)" class="text-[7px] bg-rose-600 px-1.5 py-0.5 rounded uppercase font-black tracking-widest text-white animate-pulse">BANNED</span>
                          </span>
                          <span class="text-[8px] text-slate-400 font-mono tracking-tighter">ONLINE@OP-NODE</span>
                        </div>
                      </td>
                      <td class="px-4 py-2.5 text-[10px] text-slate-500 tracking-widest">{{ u.uid }}</td>
                      <td class="px-4 py-2.5">
                        <span v-if="u.role === 'admin'" class="text-[8px] px-2 py-0.5 bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 rounded-md border border-cyan-500/20 font-black tracking-widest uppercase shadow-[0_0_10px_rgba(6,182,212,0.1)] transition-all">LV.99 CORE</span>
                        <span v-else-if="u.role === 'co-worker'" class="text-[8px] px-2 py-0.5 bg-violet-500/10 text-violet-600 dark:text-violet-400 rounded-md border border-violet-500/20 font-black tracking-widest uppercase">LV.50 ASSIST</span>
                        <span v-else class="text-[8px] px-2 py-0.5 bg-slate-100 dark:bg-white/5 text-slate-400 dark:text-slate-500 rounded-md border border-slate-200 dark:border-white/10 font-black tracking-widest uppercase">LV.01 STAFF</span>
                      </td>
                      <td class="px-4 py-2.5 text-[9px] text-slate-500 uppercase font-bold">{{ new Date(u.created_at).toLocaleDateString() }}</td>
                      <td class="px-4 py-2.5 text-right">
                        <div v-if="!u.is_admin" class="flex items-center gap-2 justify-end transition-all">
                          <button
                            v-if="tabs.length > 1"
                            @click="handlePromoteUser(u.uid, u.role)"
                            class="p-2.5 bg-slate-100 dark:bg-white/5 hover:bg-violet-500/10 text-slate-400 hover:text-violet-600 rounded-xl transition-all border border-transparent hover:border-violet-500/20"
                            title="ELEVATE_AUTH"
                          >
                            <ArrowUp class="w-3.5 h-3.5" />
                          </button>
                          <button
                            @click="handleKickPlayer(u)"
                            class="p-2.5 bg-slate-100 dark:bg-white/5 hover:bg-amber-500/10 text-slate-400 hover:text-amber-600 rounded-xl transition-all border border-transparent hover:border-amber-500/20"
                            title="KICK_FROM_ROOM"
                          >
                            <UserMinus class="w-3.5 h-3.5" />
                          </button>
                          <button
                            @click="openBanModal(u)"
                            class="p-2.5 bg-slate-100 dark:bg-white/5 hover:bg-rose-500/10 text-slate-400 hover:text-rose-600 rounded-xl transition-all border border-transparent hover:border-rose-500/20"
                            title="BAN_USER"
                          >
                            <Ban class="w-3.5 h-3.5" />
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
                <h3 class="text-lg font-black italic uppercase text-slate-900 dark:text-white flex items-center gap-4">
                  <Layers class="w-5 h-5 text-cyan-500" />
                  全局卡组配置 <span class="text-slate-400 dark:text-slate-600 font-mono not-italic text-[10px] tracking-normal">/ DECK@GLOBAL</span>
                </h3>
                <div class="flex items-center gap-3">
                  <div v-if="!editingDeck" class="bg-white/50 dark:bg-white/5 backdrop-blur-sm px-4 py-2 border border-slate-200 dark:border-white/10 rounded-xl flex items-center gap-3">
                    <span class="text-[9px] font-black uppercase text-slate-400 tracking-widest leading-none">Init_Hand:</span>
                    <span class="text-sm font-black text-cyan-500 font-mono leading-none">{{ deckConfig.initial_cards || 10 }}</span>
                  </div>
                  <div v-if="!editingDeck" class="relative group">
                    <SearchIcon class="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-cyan-500 transition-colors" />
                    <input 
                      v-model="searchTerm"
                      type="text" 
                      placeholder="FILTER..."
                      class="bg-slate-50 dark:bg-black/30 border border-slate-200 dark:border-white/5 rounded-xl pl-9 pr-4 py-2 text-[10px] font-mono focus:outline-none focus:border-cyan-500/30 w-full md:w-48 transition-all text-slate-900 dark:text-white"
                    />
                  </div>
                  <button 
                    v-if="editingDeck"
                    @click="handleAddDeckItem"
                    class="px-4 py-2 bg-cyan-500/10 hover:bg-cyan-500/20 text-cyan-500 border border-cyan-500/20 rounded-xl font-black text-[10px] uppercase tracking-widest flex items-center gap-2 transition-all shadow-sm"
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
                        : 'bg-cyan-600 hover:bg-cyan-500 text-white shadow-cyan-500/10'
                    )"
                  >
                    <component :is="editingDeck ? Save : Edit2" class="w-3.5 h-3.5" />
                    {{ editingDeck ? "保存配置" : "编辑 / OVERRIDE" }}
                  </button>
                  <button 
                    v-if="editingDeck"
                    @click="editingDeck = false"
                    class="px-4 py-2 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-500 dark:text-slate-400 rounded-xl font-black text-[10px] uppercase tracking-widest transition-all border border-slate-200 dark:border-white/10"
                  >
                    取消
                  </button>
                </div>
              </div>

              <div class="overflow-x-auto custom-scrollbar border border-slate-200 dark:border-white/10 rounded-[1.5rem] bg-slate-50 dark:bg-black/20">
                <div v-if="editingDeck" class="p-4 border-b border-slate-200 dark:border-white/10 bg-white/50 dark:bg-white/[0.02] backdrop-blur-md flex flex-wrap items-center gap-4">
                  <div class="flex items-center gap-3">
                    <span class="text-[10px] font-black uppercase text-slate-500 tracking-widest flex items-center gap-2">
                       <Settings2 class="w-3 h-3" />
                       初始手牌数目:
                    </span>
                    <div class="flex items-center bg-white dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl overflow-hidden shadow-sm">
                      <button @click="initialCardsEdit = Math.max(1, initialCardsEdit - 1)" class="px-3 py-2 hover:bg-slate-50 dark:hover:bg-white/5 text-slate-500 transition-colors border-r border-slate-200 dark:border-white/10">-</button>
                      <input 
                        v-model.number="initialCardsEdit" 
                        type="number" 
                        min="1" 
                        max="20"
                        class="w-14 bg-transparent text-center text-sm font-black text-slate-900 dark:text-white outline-none font-mono"
                      />
                      <button @click="initialCardsEdit = Math.min(20, initialCardsEdit + 1)" class="px-3 py-2 hover:bg-slate-50 dark:hover:bg-white/5 text-slate-500 transition-colors border-l border-slate-200 dark:border-white/10">+</button>
                    </div>
                  </div>
                  <div class="h-4 w-px bg-slate-200 dark:bg-white/10 hidden md:block"></div>
                  <p class="text-[9px] text-slate-400 dark:text-slate-600 font-bold italic tracking-tight uppercase">/ Global_Initial_Hand_Allocation_Protocol</p>
                </div>
                <table class="w-full text-left table-fixed">
                  <thead>
                    <tr class="text-slate-400 dark:text-slate-600 text-[9px] font-black uppercase tracking-[0.3em] border-b border-slate-200 dark:border-white/10">
                      <th class="px-5 py-2.5 w-[45%]">Element / Key</th>
                      <th class="px-5 py-2.5 w-[30%]">Quantity</th>
                      <th v-if="editingDeck" class="px-5 py-2.5 text-right w-[25%]">Ops</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-100 dark:divide-white/5 font-mono">
                    <template v-if="editingDeck">
                      <tr v-for="item in deckCardsEdit.filter(i => i.key === '' || !specialElements.includes(i.key))" :key="item.id" class="hover:bg-white/40 dark:hover:bg-cyan-500/[0.02] transition-all">
                        <td class="px-5 py-2.5">
                          <input 
                            v-model="item.key"
                            type="text"
                            placeholder="元素符号..."
                            class="w-full bg-white dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2 focus:outline-none focus:border-cyan-500/50 text-slate-900 dark:text-white text-xs font-black tracking-tight"
                          />
                        </td>
                        <td class="px-5 py-2.5">
                          <input 
                            v-model.number="item.value"
                            type="number"
                            min="0"
                            class="w-full bg-white dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2 focus:outline-none focus:border-cyan-500/50 text-cyan-600 dark:text-cyan-400 text-xs font-black"
                          />
                        </td>
                        <td class="px-5 py-2.5 text-right">
                          <button @click="handleRemoveDeckItem(item.id)" class="p-2.5 bg-slate-100 dark:bg-white/5 hover:bg-red-500/10 text-slate-400 hover:text-red-500 rounded-xl transition-all border border-transparent hover:border-red-500/20">
                            <Trash2 class="w-4 h-4" />
                          </button>
                        </td>
                      </tr>
                    </template>
                    <template v-else>
                      <tr v-for="[type, count] in filteredDeck" :key="type" class="hover:bg-white/40 dark:hover:bg-cyan-500/[0.02] transition-all group">
                        <td class="px-5 py-2.5 text-[11px] font-black text-slate-900 dark:text-white flex items-center gap-3">
                           <div class="w-1.5 h-1.5 rounded-full bg-cyan-500/40 group-hover:bg-cyan-500 transition-colors" />
                           {{ type }}
                        </td>
                        <td class="px-5 py-2.5 font-black text-sm italic text-cyan-600 dark:text-cyan-400">{{ count }}</td>
                      </tr>
                    </template>
                    <tr v-if="(!editingDeck && filteredDeck.length === 0) || (editingDeck && deckCardsEdit.filter(i => !specialElements.includes(i.key)).length === 0)">
                      <td colspan="3" class="py-16 text-center text-slate-400 dark:text-slate-700 text-[10px] font-black uppercase tracking-[0.3em] italic">
                        / NO_ELEMENTS_LOADED_IN_MATRIX
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- Special Deck Tab -->
            <div v-if="activeTab === 'special' && deckConfig" class="space-y-6">
              <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
                <h3 class="text-lg font-black italic uppercase text-slate-900 dark:text-white flex items-center gap-4">
                  <Star class="w-5 h-5 text-violet-500" />
                  稀有元素配置 <span class="text-slate-400 dark:text-slate-600 font-mono not-italic text-[10px] tracking-normal">/ DECK@SPECIALS</span>
                </h3>
                <div class="flex items-center gap-3">
                  <div v-if="!editingDeck" class="relative group">
                    <SearchIcon class="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-violet-400 transition-colors" />
                    <input 
                      v-model="searchTerm"
                      type="text" 
                      placeholder="FILTER..."
                      class="bg-slate-50 dark:bg-black/30 border border-slate-200 dark:border-white/5 rounded-xl pl-9 pr-4 py-2 text-[10px] font-mono focus:outline-none focus:border-violet-500/30 w-full md:w-48 transition-all text-slate-900 dark:text-white"
                    />
                  </div>
                  <button 
                    v-if="editingDeck"
                    @click="handleAddDeckItem"
                    class="px-4 py-2 bg-violet-500/10 hover:bg-violet-500/20 text-violet-500 border border-violet-500/20 rounded-xl font-black text-[10px] uppercase tracking-widest flex items-center gap-2 transition-all shadow-sm"
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
                        : 'bg-violet-600 hover:bg-violet-500 text-white shadow-violet-500/10'
                    )"
                  >
                    <component :is="editingDeck ? Save : Edit2" class="w-3.5 h-3.5" />
                    {{ editingDeck ? "保存配置" : "编辑 / OVERRIDE" }}
                  </button>
                  <button 
                    v-if="editingDeck"
                    @click="editingDeck = false"
                    class="px-4 py-2 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-500 dark:text-slate-400 rounded-xl font-black text-[10px] uppercase tracking-widest transition-all border border-slate-200 dark:border-white/10"
                  >
                    取消
                  </button>
                </div>
              </div>

              <div class="overflow-x-auto custom-scrollbar border border-slate-200 dark:border-white/10 rounded-[1.5rem] bg-slate-50 dark:bg-black/20">
                <table class="w-full text-left table-fixed">
                  <thead>
                    <tr class="text-slate-400 dark:text-slate-600 text-[9px] font-black uppercase tracking-[0.3em] border-b border-slate-200 dark:border-white/10">
                      <th class="px-5 py-2.5 w-[45%]">Special Element</th>
                      <th class="px-5 py-2.5 w-[30%]">Quantity</th>
                      <th v-if="editingDeck" class="px-5 py-2.5 text-right w-[25%]">Ops</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-100 dark:divide-white/5 font-mono">
                    <template v-if="editingDeck">
                      <tr v-for="item in deckCardsEdit.filter(i => i.key === '' || specialElements.includes(i.key))" :key="item.id" class="hover:bg-white/40 dark:hover:bg-violet-500/[0.02] transition-all">
                        <td class="px-5 py-2.5">
                          <input 
                            v-model="item.key"
                            type="text"
                            placeholder="元素符号..."
                            class="w-full bg-white dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2 focus:outline-none focus:border-violet-500/50 text-slate-900 dark:text-white text-xs font-black tracking-tight"
                          />
                        </td>
                        <td class="px-5 py-2.5">
                          <input 
                            v-model.number="item.value"
                            type="number"
                            min="0"
                            class="w-full bg-white dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2 focus:outline-none focus:border-violet-500/50 text-violet-600 dark:text-violet-400 text-xs font-black"
                          />
                        </td>
                        <td class="px-5 py-2.5 text-right">
                          <button @click="handleRemoveDeckItem(item.id)" class="p-2.5 bg-slate-100 dark:bg-white/5 hover:bg-red-500/10 text-slate-400 hover:text-red-500 rounded-xl transition-all border border-transparent hover:border-red-500/20">
                            <Trash2 class="w-4 h-4" />
                          </button>
                        </td>
                      </tr>
                    </template>
                    <template v-else>
                      <tr v-for="[type, count] in filteredSpecialDeck" :key="type" class="hover:bg-white/40 dark:hover:bg-violet-500/[0.02] transition-colors group">
                        <td class="px-5 py-2.5 text-[11px] font-black text-violet-600 dark:text-violet-400 flex items-center gap-3">
                           <div class="w-1.5 h-1.5 rounded-full bg-violet-500/40 group-hover:bg-violet-500 transition-colors" />
                           {{ type }}
                        </td>
                        <td class="px-5 py-2.5 font-black text-sm italic text-violet-800 dark:text-violet-500">{{ count }}</td>
                      </tr>
                    </template>
                    <tr v-if="(!editingDeck && filteredSpecialDeck.length === 0) || (editingDeck && deckCardsEdit.filter(i => specialElements.includes(i.key)).length === 0)">
                      <td colspan="3" class="py-16 text-center text-slate-400 dark:text-slate-700 text-[10px] font-black uppercase tracking-[0.3em] italic">
                        / NO_SPECIAL_ELEMENTS_FOUND_IN_MATRIX
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- Game Time Config Tab -->
            <div v-if="activeTab === 'game-time'" class="space-y-8">
              <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
                <h3 class="text-lg font-black italic uppercase text-slate-900 dark:text-white flex items-center gap-4">
                  <Clock class="w-5 h-5 text-blue-500" />
                  游戏时间配置 <span class="text-slate-400 dark:text-slate-600 font-mono not-italic text-[10px] tracking-normal">/ GAME@TIMING --config</span>
                </h3>
              </div>

              <div class="space-y-4">
                <div class="flex items-center gap-2 mb-2">
                  <div class="h-px flex-1 bg-slate-200 dark:bg-white/5"></div>
                  <span class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">游戏时间参数 / GAME_TIME_PARAMS</span>
                  <div class="h-px flex-1 bg-slate-200 dark:bg-white/5"></div>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <!-- Player Kick Timeout -->
                  <div class="bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/5 p-4 rounded-[1.25rem] shadow-sm relative overflow-hidden">
                    <div class="absolute top-0 right-0 w-32 h-32 bg-blue-500/[0.03] blur-[50px] -mr-16 -mt-16" />
                    <div class="relative z-10">
                      <label class="text-[10px] font-mono text-blue-600 dark:text-blue-400 uppercase tracking-widest font-black flex items-center gap-2 mb-3">
                        <Clock class="w-4 h-4" /> 玩家离线踢出时间
                      </label>
                      <input
                        v-model="gameTimeConfigs.player_kick_timeout"
                        type="number"
                        min="10"
                        max="300"
                        class="w-full bg-white dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2.5 text-lg font-black text-slate-900 dark:text-white focus:outline-none focus:border-blue-500/50 transition-all"
                      />
                      <div class="text-[10px] font-bold text-slate-500 uppercase tracking-widest mt-2 opacity-70">秒 (范围: 10-300)</div>
                    </div>
                  </div>

                  <!-- Player Action Timeout -->
                  <div class="bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/5 p-4 rounded-[1.25rem] shadow-sm relative overflow-hidden">
                    <div class="absolute top-0 right-0 w-32 h-32 bg-green-500/[0.03] blur-[50px] -mr-16 -mt-16" />
                    <div class="relative z-10">
                      <label class="text-[10px] font-mono text-green-600 dark:text-green-400 uppercase tracking-widest font-black flex items-center gap-2 mb-3">
                        <Clock class="w-4 h-4" /> 玩家操作时间
                      </label>
                      <input
                        v-model="gameTimeConfigs.player_action_timeout"
                        type="number"
                        min="10"
                        max="300"
                        class="w-full bg-white dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2.5 text-lg font-black text-slate-900 dark:text-white focus:outline-none focus:border-green-500/50 transition-all"
                      />
                      <div class="text-[10px] font-bold text-slate-500 uppercase tracking-widest mt-2 opacity-70">秒 (范围: 10-300)</div>
                    </div>
                  </div>

                  <!-- Auto Start Timeout -->
                  <div class="bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/5 p-4 rounded-[1.25rem] shadow-sm relative overflow-hidden">
                    <div class="absolute top-0 right-0 w-32 h-32 bg-purple-500/[0.03] blur-[50px] -mr-16 -mt-16" />
                    <div class="relative z-10">
                      <label class="text-[10px] font-mono text-purple-600 dark:text-purple-400 uppercase tracking-widest font-black flex items-center gap-2 mb-3">
                        <Clock class="w-4 h-4" /> 自动开始倒计时
                      </label>
                      <input
                        v-model="gameTimeConfigs.auto_start_timeout"
                        type="number"
                        min="5"
                        max="60"
                        class="w-full bg-white dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2.5 text-lg font-black text-slate-900 dark:text-white focus:outline-none focus:border-purple-500/50 transition-all"
                      />
                      <div class="text-[10px] font-bold text-slate-500 uppercase tracking-widest mt-2 opacity-70">秒 (范围: 5-60)</div>
                    </div>
                  </div>

                  <!-- Half Ready Timeout -->
                  <div class="bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/5 p-4 rounded-[1.25rem] shadow-sm relative overflow-hidden">
                    <div class="absolute top-0 right-0 w-32 h-32 bg-orange-500/[0.03] blur-[50px] -mr-16 -mt-16" />
                    <div class="relative z-10">
                      <label class="text-[10px] font-mono text-orange-600 dark:text-orange-400 uppercase tracking-widest font-black flex items-center gap-2 mb-3">
                        <Clock class="w-4 h-4" /> 半数准备倒计时
                      </label>
                      <input
                        v-model="gameTimeConfigs.half_ready_timeout"
                        type="number"
                        min="30"
                        max="120"
                        class="w-full bg-white dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2.5 text-lg font-black text-slate-900 dark:text-white focus:outline-none focus:border-orange-500/50 transition-all"
                      />
                      <div class="text-[10px] font-bold text-slate-500 uppercase tracking-widest mt-2 opacity-70">秒 (范围: 30-120)</div>
                    </div>
                  </div>
                </div>

                <div class="flex items-center gap-2 mb-2 mt-6">
                  <div class="h-px flex-1 bg-slate-200 dark:bg-white/5"></div>
                  <span class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">进阶系统参数 / ADVANCED_SYSTEM_PARAMS</span>
                  <div class="h-px flex-1 bg-slate-200 dark:bg-white/5"></div>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <!-- Reconnect Grace Period -->
                  <div class="bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/5 p-4 rounded-[1.25rem] shadow-sm relative overflow-hidden">
                    <div class="absolute top-0 right-0 w-32 h-32 bg-rose-500/[0.03] blur-[50px] -mr-16 -mt-16" />
                    <div class="relative z-10">
                      <label class="text-[10px] font-mono text-rose-600 dark:text-rose-400 uppercase tracking-widest font-black flex items-center gap-2 mb-3">
                        <Activity class="w-4 h-4" /> 掉线重连宽限期
                      </label>
                      <input
                        v-model="gameTimeConfigs.reconnect_grace_period"
                        type="number"
                        min="0"
                        max="300"
                        class="w-full bg-white dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2.5 text-lg font-black text-slate-900 dark:text-white focus:outline-none focus:border-rose-500/50 transition-all"
                      />
                      <div class="text-[10px] font-bold text-slate-500 uppercase tracking-widest mt-2 opacity-70">秒 (0-300，目前作为预留参数)</div>
                    </div>
                  </div>

                  <!-- Points Scaling Enabled -->
                  <div class="bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/5 p-4 rounded-[1.25rem] shadow-sm relative overflow-hidden">
                    <div class="absolute top-0 right-0 w-32 h-32 bg-amber-500/[0.03] blur-[50px] -mr-16 -mt-16" />
                    <div class="relative z-10">
                      <label class="text-[10px] font-mono text-amber-600 dark:text-amber-400 uppercase tracking-widest font-black flex items-center gap-2 mb-3">
                        <Trophy class="w-4 h-4" /> 积分动态缩放系统
                      </label>
                      <select
                        v-model="gameTimeConfigs.points_scaling_enabled"
                        class="w-full bg-white dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2.5 text-lg font-black text-slate-900 dark:text-white focus:outline-none focus:border-amber-500/50 transition-all"
                      >
                        <option value="true">ENABLED / 启用</option>
                        <option value="false">DISABLED / 禁用</option>
                      </select>
                      <div class="text-[10px] font-bold text-slate-500 uppercase tracking-widest mt-2 opacity-70">根据房间人数和离线率自动调整积分获取量</div>
                    </div>
                  </div>
                </div>

                <!-- Save Button -->
                <div class="flex justify-end mt-6">
                  <button
                    @click="handleUpdateGameTimeConfig"
                    class="px-5 py-2.5 bg-gradient-to-r from-blue-500 to-purple-600 text-white font-black uppercase text-sm rounded-xl hover:shadow-2xl hover:shadow-blue-500/20 transition-all flex items-center gap-3 border border-white/10"
                  >
                    <Save class="w-4 h-4" />
                    保存配置
                  </button>
                </div>
              </div>

              <div class="p-5 rounded-[1.5rem] bg-blue-500/[0.03] border border-blue-500/10 flex flex-col sm:flex-row items-center sm:items-start gap-4 relative overflow-hidden group">
                <div class="absolute inset-0 bg-gradient-to-r from-blue-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
                <div class="p-5 rounded-xl bg-blue-500/10 text-blue-600 dark:text-blue-400 shrink-0 border border-blue-500/20 shadow-inner relative z-10">
                  <Activity class="w-7 h-7" />
                </div>
                <div class="space-y-2 relative z-10 text-center sm:text-left">
                  <h4 class="text-sm font-black text-slate-900 dark:text-white uppercase italic tracking-wider">游戏时间配置说明 / TIMING_MANUAL</h4>
                  <p class="text-xs text-slate-500 dark:text-slate-400 font-bold leading-relaxed max-w-2xl italic">
                    • 玩家离线踢出时间：玩家断线后多久会被踢出游戏<br>
                    • 玩家操作时间：每个玩家的回合操作时限<br>
                    • 自动开始倒计时：房间满员且全员准备后的倒计时<br>
                    • 半数准备倒计时：半数玩家准备后的倒计时<br>
                    ⚠️ 注意：配置更新后将在新创建的游戏中生效，不影响正在进行的游戏。
                  </p>
                </div>
              </div>
            </div>

            <!-- History Tab -->
            <div v-if="activeTab === 'history'" class="space-y-8">
              <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
                <h3 class="text-lg font-black italic uppercase text-slate-900 dark:text-white flex items-center gap-4">
                  <History class="w-5 h-5 text-cyan-500 shrink-0" />
                  全球实验追溯记录 <span class="text-slate-400 dark:text-slate-600 font-mono not-italic text-[10px] tracking-normal">/ SCAN@LOGS --ALL</span>
                </h3>
                <div class="flex items-center gap-4">
                  <div class="relative group">
                    <SearchIcon class="w-4 h-4 absolute left-4 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-cyan-500 transition-colors" />
                    <input 
                      v-model="searchTerm"
                      type="text" 
                      placeholder="SEARCH EXPERIMENT ID..."
                      class="bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/5 rounded-xl pl-12 pr-6 py-2.5 text-[10px] font-black tracking-widest focus:outline-none focus:border-cyan-500/30 w-full md:w-64 transition-all placeholder:text-slate-400 dark:placeholder:text-slate-700 text-slate-900 dark:text-white"
                    />
                  </div>
                </div>
              </div>

              <div class="overflow-x-auto custom-scrollbar border border-slate-200 dark:border-white/5 rounded-[1.5rem] bg-slate-50 dark:bg-black/20">
                <table class="w-full text-left">
                  <thead>
                    <tr class="text-slate-400 dark:text-slate-600 text-[9px] font-black uppercase tracking-[0.3em] border-b border-slate-200 dark:border-white/10">
                      <th class="px-5 py-2.5">Experiment ID</th>
                      <th class="px-5 py-2.5">Timestamp / Sync</th>
                      <th class="px-5 py-2.5">Subject Status</th>
                      <th class="px-5 py-2.5 text-right">Protocol Data</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-100 dark:divide-white/5 font-mono">
                    <tr v-for="game in filteredHistory" :key="game.id" class="hover:bg-white/40 dark:hover:bg-cyan-500/[0.03] transition-all group cursor-pointer">
                      <td class="px-5 py-2.5 font-black text-slate-900 dark:text-white group-hover:text-cyan-600 dark:group-hover:text-cyan-400 transition-colors text-xs tracking-tighter">
                        <span class="text-slate-400 dark:text-slate-600 font-normal opacity-50">STATION:</span>{{ String(game.id).padStart(4, '0') }}
                      </td>
                      <td class="px-5 py-2.5 text-[10px] text-slate-500 dark:text-slate-500 font-bold uppercase">
                        {{ new Date((game.finished_at || game.created_at || '').replace(' ', 'T')).toLocaleString() }}
                      </td>
                      <td class="px-5 py-2.5">
                        <div class="flex items-center gap-3">
                          <span class="text-[9px] px-3.5 py-1.5 bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 rounded-lg border border-cyan-500/20 font-black tracking-widest uppercase shadow-sm">
                            Winner: {{ game.winner_name || 'UNDEFINED' }}
                          </span>
                          <span class="text-[8px] text-slate-400 dark:text-slate-600 uppercase font-black opacity-40">/ COMPLETED</span>
                        </div>
                      </td>
                      <td class="px-5 py-2.5 text-right">
                        <div class="inline-flex items-center justify-center w-8 h-8 rounded-full border border-slate-200 dark:border-white/10 group-hover:border-cyan-500/50 transition-all">
                           <ChevronRight class="w-4 h-4 text-slate-300 dark:text-slate-800 group-hover:text-cyan-500 group-hover:translate-x-1 transition-all" />
                        </div>
                      </td>
                    </tr>
                    <tr v-if="filteredHistory.length === 0">
                      <td colspan="4" class="py-24 text-center text-slate-400 dark:text-slate-600 italic font-black uppercase tracking-[0.4em] text-[10px]">
                        / NO_HISTORY_DATA_FOUND_IN_BUFFER
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- Announcements Tab -->
            <div v-if="activeTab === 'announcements'" class="space-y-8">
              <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
                <h3 class="text-lg font-black italic uppercase text-slate-900 dark:text-white flex items-center gap-4">
                  <Megaphone class="w-5 h-5 text-purple-500 shrink-0" />
                  广播推送 / 播音指挥 <span class="text-slate-400 dark:text-slate-600 font-mono not-italic text-[10px] tracking-normal">/ COMMS@BROADCAST --ACTIVE</span>
                </h3>
                <button 
                  @click="showCreateAnnouncementModal = true"
                  class="bg-cyan-600 hover:bg-cyan-500 text-white px-4 py-2.5 rounded-xl text-[10px] font-black uppercase tracking-widest flex items-center gap-3 transition-all shadow-lg"
                >
                  <Plus class="w-4 h-4" />
                  NEW_BROADCAST
                </button>
              </div>

              <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
                <div v-for="ann in announcements" :key="ann.id" 
                  :class="cn(
                    'p-4 rounded-[1.5rem] border transition-all relative overflow-hidden group',
                    ann.active ? 'bg-cyan-500/[0.03] border-cyan-500/20 shadow-md' : 'bg-slate-50 dark:bg-black/20 border-slate-200 dark:border-white/5 opacity-60'
                  )"
                >
                  <div class="flex items-center justify-between mb-4 relative z-10">
                    <div class="flex items-center gap-3">
                      <div :class="cn(
                        'w-8 h-8 rounded-lg flex items-center justify-center border',
                        ann.type === 'emergency' ? 'bg-red-500/10 text-red-500 border-red-500/20' : 
                        ann.type === 'maintenance' ? 'bg-amber-500/10 text-amber-500 border-amber-500/20' :
                        'bg-cyan-500/10 text-cyan-500 border-cyan-500/20'
                      )">
                        <Megaphone class="w-4 h-4" />
                      </div>
                      <div>
                        <span class="text-[10px] font-black uppercase tracking-widest text-slate-500">{{ ann.type }}</span>
                        <div class="flex items-center gap-2">
                          <span class="text-xs font-black text-slate-900 dark:text-white" v-if="ann.title">{{ ann.title }}</span>
                          <span v-if="ann.is_ticker" class="text-[8px] px-1.5 py-0.5 bg-emerald-500/10 text-emerald-500 rounded border border-emerald-500/10 font-black uppercase tracking-tighter">TICKER</span>
                          <span v-else class="text-[8px] px-1.5 py-0.5 bg-blue-500/10 text-blue-500 rounded border border-blue-500/10 font-black uppercase tracking-tighter">ALERT</span>
                          <span v-if="ann.is_persistent" class="text-[8px] px-1.5 py-0.5 bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 rounded border border-emerald-500/20 font-black uppercase tracking-tighter">PERSISTENT</span>
                          <span v-if="ann.on_join" class="text-[8px] px-1.5 py-0.5 bg-purple-500/10 text-purple-500 rounded border border-purple-500/10 font-black uppercase tracking-tighter">ON_JOIN</span>
                          <span v-if="ann.cron_interval > 0" class="text-[8px] px-1.5 py-0.5 bg-amber-500/10 text-amber-500 rounded border border-amber-500/10 font-black uppercase tracking-tighter">CRON: {{ ann.cron_interval }}M</span>
                          <span v-if="!ann.is_ticker && ann.close_delay > 0" class="text-[8px] px-1.5 py-0.5 bg-red-500/10 text-red-500 rounded border border-red-500/10 font-black uppercase tracking-tighter">DELAY: {{ ann.close_delay }}S</span>
                        </div>
                      </div>
                    </div>
                    <div class="flex items-center gap-2">
                      <button @click="handleToggleAnnouncement(ann.id, !ann.active)"
                        :class="cn(
                          'px-3 py-1.5 rounded-lg text-[8px] font-black uppercase tracking-widest transition-all border',
                          ann.active ? 'bg-emerald-500 text-white border-emerald-600 shadow-lg shadow-emerald-500/20' : 'bg-slate-200 dark:bg-white/5 text-slate-500 border-slate-300 dark:border-white/10'
                        )"
                      >
                        {{ ann.active ? 'ACTIVE' : 'DISABLED' }}
                      </button>
                      <button @click="handleEditAnnouncement(ann)" class="p-2 bg-blue-500/10 text-blue-400 hover:bg-blue-500 hover:text-white rounded-lg transition-all border border-blue-500/10" title="编辑公告">
                        <Edit2 class="w-3.5 h-3.5" />
                      </button>
                      <button @click="handleDeleteAnnouncement(ann.id)" class="p-2 bg-red-500/10 text-red-400 hover:bg-red-500 hover:text-white rounded-lg transition-all border border-red-500/10">
                        <Trash2 class="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </div>

                  <div class="bg-white/40 dark:bg-black/40 p-5 rounded-xl relative z-10 border border-slate-100 dark:border-white/5 min-h-[60px] mb-4">
                    <p class="text-[11px] leading-relaxed text-slate-600 dark:text-slate-300 font-bold">{{ ann.content }}</p>
                  </div>

                  <div class="flex items-center justify-between text-[8px] font-mono text-slate-400 uppercase tracking-widest relative z-10">
                    <div class="flex items-center gap-2">
                      <Clock class="w-3 h-3" />
                      EXPIRES: {{ ann.expires_at ? new Date(ann.expires_at).toLocaleString() : 'NEVER' }}
                    </div>
                    <div class="flex items-center gap-1">
                       REV: {{ ann.id }}
                    </div>
                  </div>
                </div>
                <div v-if="announcements.length === 0" class="col-span-full py-20 text-center text-slate-400 italic font-black uppercase tracking-[0.3em] text-[10px]">
                  / NO_ACTIVE_BROADCASTS_DETECTED
                </div>
              </div>
            </div>

            <!-- Feedbacks Tab -->
            <div v-if="activeTab === 'feedbacks'" class="space-y-8">
              <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
                <h3 class="text-lg font-black italic uppercase text-slate-900 dark:text-white flex items-center gap-4">
                  <MessageSquare class="w-5 h-5 text-sky-500" />
                  外部反馈报告 <span class="text-slate-400 dark:text-slate-600 font-mono not-italic text-[10px] tracking-normal">/ INCOMING@COMMS --FILTERED</span>
                </h3>
              </div>

              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div v-for="fb in feedbacks" :key="fb.id" class="p-4 bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/5 rounded-[1.25rem] hover:border-sky-500/30 transition-all flex flex-col gap-4 shadow-sm hover:shadow-xl group relative overflow-hidden">
                   <div class="absolute top-0 right-0 w-32 h-32 bg-sky-500/5 blur-[50px] -mr-16 -mt-16 group-hover:bg-sky-500/10 transition-all" />
                   <div class="flex items-center justify-between relative z-10">
                      <div class="flex items-center gap-3">
                         <div class="w-10 h-10 rounded-xl bg-sky-500/10 flex items-center justify-center text-sky-600 dark:text-sky-400 font-black border border-sky-500/20">
                            {{ (fb.nickname || fb.username)[0].toUpperCase() }}
                         </div>
                         <div>
                            <p class="text-xs font-black text-slate-900 dark:text-white uppercase tracking-tight">{{ fb.nickname || fb.username }}</p>
                            <p class="text-[8px] text-slate-400 dark:text-slate-500 font-mono font-bold uppercase">{{ new Date(fb.created_at).toLocaleString() }}</p>
                         </div>
                      </div>
                      <span class="px-2 py-1 bg-slate-100 dark:bg-white/5 text-slate-400 dark:text-slate-500 text-[8px] font-black uppercase tracking-widest rounded-lg border border-slate-200 dark:border-white/5">{{ fb.page }}</span>
                   </div>
                   <div class="bg-white/50 dark:bg-black/40 p-5 rounded-xl relative z-10 border border-slate-100 dark:border-white/5 shadow-inner min-h-[80px]">
                     <p class="text-[11px] leading-relaxed text-slate-600 dark:text-slate-400 font-bold italic">“{{ fb.content }}”</p>
                   </div>
                   <div class="flex items-center justify-end gap-3 relative z-10">
                      <button v-if="fb.type === 'report'" @click="handleBanReportedPlayer(fb)" class="px-5 py-2.5 bg-rose-600 hover:bg-rose-500 text-white rounded-xl text-[10px] font-black uppercase tracking-widest transition-all shadow-lg hover:shadow-rose-500/20 active:scale-95 flex items-center gap-2">
                        <Ban class="w-3.5 h-3.5" />Ban
                      </button>
                      <button @click="handleAcceptFeedback(fb.id)" class="px-5 py-2.5 bg-cyan-600 hover:bg-cyan-500 text-white rounded-xl text-[10px] font-black uppercase tracking-widest transition-all shadow-lg hover:shadow-cyan-500/20 active:scale-95">Accept</button>
                      <button @click="handleDismissFeedback(fb.id)" class="px-5 py-2.5 bg-slate-100 dark:bg-white/5 text-slate-400 hover:text-red-500 rounded-xl text-[10px] font-black uppercase tracking-widest transition-all border border-slate-200 dark:border-white/10 hover:border-red-500/20 active:scale-95">Dismiss</button>
                   </div>
                </div>
                <div v-if="feedbacks.length === 0" class="col-span-full py-20 text-center text-slate-400 dark:text-slate-600 italic font-black uppercase tracking-[0.3em] text-[10px]">
                  / COMM_BUFF_EMPTY_STATUS_NOMINAL
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>

    <!-- 封禁用户模态框 -->
    <div v-if="showBanModal && banTarget" class="fixed inset-0 bg-slate-900/40 dark:bg-black/80 backdrop-blur-md z-50 flex items-center justify-center p-4">
      <div class="bg-white dark:bg-[#0c0c0e] border border-slate-200 dark:border-white/10 rounded-[1.5rem] p-4 max-w-md w-full shadow-[0_50px_100px_-20px_rgba(220,38,38,0.15)] animate-in zoom-in relative overflow-hidden">
        <div class="absolute top-0 right-0 w-40 h-40 bg-rose-500/5 blur-[60px] -mr-20 -mt-20" />

        <div class="flex items-center justify-between mb-8 relative z-10">
          <h3 class="text-lg font-black italic uppercase text-slate-900 dark:text-white flex items-center gap-4">
            <Ban class="w-6 h-6 text-rose-500" />
            {{ isBanned(banTarget) ? '修改封禁时间' : '封禁研究员' }} <span class="text-[10px] text-slate-400 font-mono not-italic tracking-normal">/ {{ isBanned(banTarget) ? 'MODIFY_BAN' : 'ACCESS_RESTRICT' }}</span>
          </h3>
          <button @click="showBanModal = false" class="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl transition-colors text-slate-400 hover:text-slate-900 dark:hover:text-white">
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="p-4 bg-rose-500/5 border border-rose-500/10 rounded-xl mb-6 relative z-10 flex items-center gap-4">
          <div class="w-12 h-12 bg-white dark:bg-black/40 rounded-xl flex items-center justify-center text-lg border border-slate-200 dark:border-white/10 shadow-sm overflow-hidden shrink-0">
            <template v-if="banTarget.avatar && banTarget.avatar.startsWith('data:')">
              <img :src="banTarget.avatar" class="w-full h-full object-cover" />
            </template>
            <template v-else>
              {{ banTarget.avatar || '🧪' }}
            </template>
          </div>
          <div>
            <p class="text-xs font-black text-slate-900 dark:text-white uppercase tracking-tight">{{ banTarget.nickname || banTarget.username }}</p>
            <p class="text-[9px] text-slate-400 font-mono">UID: {{ banTarget.uid }}</p>
          </div>
        </div>

        <div class="space-y-6 relative z-10">
          <div>
            <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-[0.2em] block mb-3 ml-1">Ban Duration / 封禁时长</label>
            <div class="grid grid-cols-4 gap-2 mb-4">
              <button
                v-for="preset in banPresets"
                :key="preset.hours"
                @click="setBanDuration(preset.hours)"
                :class="cn(
                  'px-3 py-2.5 rounded-xl text-[10px] font-black uppercase tracking-wider transition-all border active:scale-95',
                  selectedPreset === preset.hours
                    ? 'bg-rose-500/10 border-rose-500/30 text-rose-500 shadow-sm'
                    : 'bg-slate-50 dark:bg-black/40 border-slate-200 dark:border-white/10 text-slate-500 hover:border-rose-500/20 hover:text-rose-400'
                )"
              >
                {{ preset.label }}
              </button>
              <button
                @click="selectedPreset = null"
                :class="cn(
                  'px-3 py-2.5 rounded-xl text-[10px] font-black uppercase tracking-wider transition-all border active:scale-95',
                  selectedPreset === null
                    ? 'bg-rose-500/10 border-rose-500/30 text-rose-500 shadow-sm'
                    : 'bg-slate-50 dark:bg-black/40 border-slate-200 dark:border-white/10 text-slate-500 hover:border-rose-500/20 hover:text-rose-400'
                )"
              >
                自定义
              </button>
            </div>
            <div v-if="selectedPreset === null" class="animate-in slide-in-from-top-2 duration-200">
              <input
                v-model="banUntil"
                type="datetime-local"
                :min="formatDatetimeLocal(new Date())"
                class="w-full bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-5 py-2.5 text-sm font-black text-slate-900 dark:text-white focus:outline-none focus:border-rose-500/50 transition-all"
              />
            </div>
            <div class="flex items-center gap-2 ml-1 mt-2">
              <div class="w-1.5 h-1.5 rounded-full" :class="banUntil ? 'bg-rose-500 animate-pulse' : 'bg-slate-300 dark:bg-slate-600'"></div>
              <span class="text-[9px] font-bold text-slate-400 dark:text-slate-600 uppercase tracking-wider">
                封禁截止: {{ banUntil ? new Date(banUntil).toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' }) + '（UTC+8）' : '未设置' }}
              </span>
            </div>
          </div>

          <div>
            <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-[0.2em] block mb-3 ml-1">Reason / 封禁事由</label>
            <textarea
              v-model="banReason"
              rows="3"
              placeholder="请输入封禁原因..."
              class="w-full bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-5 py-2.5 text-xs font-bold text-slate-900 dark:text-white focus:outline-none focus:border-rose-500/50 transition-all resize-none"
            />
          </div>
        </div>

        <div class="flex gap-4 mt-10 relative z-10">
          <button
            @click="showBanModal = false"
            class="flex-1 px-4 py-2.5 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-500 dark:text-slate-400 rounded-[1.25rem] font-black text-[10px] uppercase tracking-widest transition-all border border-slate-200 dark:border-white/5"
          >
            Abort
          </button>
          <button
            @click="handleBanUser"
            class="flex-1 px-4 py-2.5 bg-rose-600 hover:bg-rose-500 text-white rounded-[1.25rem] font-black text-[10px] uppercase tracking-widest transition-all shadow-lg shadow-rose-500/20 active:scale-95 border border-rose-500/20"
          >
            {{ isBanned(banTarget) ? 'Update Ban' : 'Execute Ban' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 发布公告模态框 -->
    <div v-if="showCreateAnnouncementModal" class="fixed inset-0 bg-slate-900/40 dark:bg-black/80 backdrop-blur-md z-50 flex items-center justify-center p-4">
      <div class="bg-white dark:bg-[#0c0c0e] border border-slate-200 dark:border-white/10 rounded-[1.5rem] p-4 max-w-lg w-full shadow-[0_50px_100px_-20px_rgba(6,182,212,0.2)] animate-in zoom-in relative overflow-hidden">
        <div class="absolute top-0 right-0 w-40 h-40 bg-cyan-500/5 blur-[60px] -mr-20 -mt-20" />
        
        <h3 class="text-lg font-black italic uppercase text-slate-900 dark:text-white mb-8 flex items-center gap-4 relative z-10">
          <Megaphone class="w-6 h-6 text-cyan-500" />
          发起全域广播 <span class="text-[10px] text-slate-400 font-mono not-italic tracking-normal">/ NEW_SEC_BROADCAST</span>
        </h3>
        
        <div class="space-y-6 relative z-10">
          <div>
            <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-[0.2em] block mb-3 ml-1">Title / 标题 (可选)</label>
            <input 
              v-model="newAnnouncement.title"
              type="text" 
              placeholder="ENTER BROADCAST TITLE..."
              class="w-full bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-5 py-2.5 text-xs font-bold text-slate-900 dark:text-white focus:outline-none focus:border-cyan-500/50 transition-all"
            />
          </div>

          <div>
            <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-[0.2em] block mb-3 ml-1">Broadcast Content / 内容</label>
            <textarea 
              v-model="newAnnouncement.content"
              rows="3"
              placeholder="ENTER MESSAGE TO RESEARCHERS..."
              class="w-full bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-5 py-2.5 text-xs font-bold text-slate-900 dark:text-white focus:outline-none focus:border-cyan-500/50 transition-all resize-none"
            />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-[0.2em] block mb-3 ml-1">Priority / 优先级</label>
              <select 
                v-model="newAnnouncement.type"
                class="w-full bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2.5 text-[10px] font-black uppercase tracking-widest text-slate-900 dark:text-white focus:outline-none focus:border-cyan-500/30"
              >
                <option value="info">INFO / 公告</option>
                <option value="maintenance">MAINTENANCE / 维护</option>
                <option value="emergency">EMERGENCY / 紧急</option>
              </select>
            </div>
            <div>
              <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-[0.2em] block mb-3 ml-1">Format / 展现形式</label>
              <select 
                v-model="newAnnouncement.is_ticker"
                class="w-full bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2.5 text-[10px] font-black uppercase tracking-widest text-slate-900 dark:text-white focus:outline-none focus:border-cyan-500/30"
              >
                <option :value="true">TICKER / 跑马灯</option>
                <option :value="false">MODAL / 强制弹窗</option>
              </select>
            </div>
          </div>

          <div>
             <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-[0.2em] block mb-3 ml-1">TTL / 有效时长 (e.g. 24h, 7d, 30m)</label>
             <input 
              v-model="newAnnouncement.expires_in"
              type="text" 
              placeholder="24h"
              class="w-full bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-5 py-2.5 text-xs font-bold font-mono text-slate-900 dark:text-white focus:outline-none focus:border-cyan-500/50 tracking-widest uppercase"
            />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="flex items-center gap-3 p-4 bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl">
              <input 
                type="checkbox" 
                v-model="newAnnouncement.on_join"
                class="w-4 h-4 rounded border-slate-300 text-cyan-600 focus:ring-cyan-500"
              />
              <label class="text-[10px] font-black text-slate-600 dark:text-slate-400 uppercase tracking-widest">玩家加入时触发</label>
            </div>
            <div class="flex items-center gap-3 p-4 bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl">
              <input 
                type="checkbox" 
                v-model="newAnnouncement.is_persistent"
                class="w-4 h-4 rounded border-slate-300 text-cyan-600 focus:ring-cyan-500"
              />
              <label class="text-[10px] font-black text-slate-600 dark:text-slate-400 uppercase tracking-widest">常驻显示</label>
            </div>
          </div>

          <div>
              <label class="text-[8px] font-black text-slate-400 uppercase tracking-widest block mb-1">自动循环间隔 (分钟) / 0 = 禁止</label>
              <input 
                v-model.number="newAnnouncement.cron_interval"
                type="number" 
                placeholder="0"
                class="w-full bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2 text-[10px] font-black text-slate-900 dark:text-white focus:outline-none focus:border-cyan-500/30"
              />
          </div>

          <div v-if="!newAnnouncement.is_ticker">
             <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-[0.2em] block mb-3 ml-1">Force Close Delay / 强制等待时间 (秒)</label>
             <input 
              v-model.number="newAnnouncement.close_delay"
              type="number" 
              placeholder="0 = 立即关闭"
              class="w-full bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-5 py-2.5 text-xs font-bold font-mono text-slate-900 dark:text-white focus:outline-none focus:border-cyan-500/50 tracking-widest uppercase"
            />
          </div>
        </div>
        
        <div class="flex gap-4 mt-10 relative z-10">
          <button 
            @click="showCreateAnnouncementModal = false"
            class="flex-1 px-4 py-2.5 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-500 rounded-[1.25rem] font-black text-[10px] uppercase tracking-widest transition-all"
          >
            Cancel
          </button>
          <button 
            @click="handleCreateAnnouncement"
            class="flex-1 px-4 py-2.5 bg-cyan-600 hover:bg-cyan-500 text-white rounded-[1.25rem] font-black text-[10px] uppercase tracking-widest transition-all shadow-lg shadow-cyan-500/20 active:scale-95 border border-cyan-500/20"
          >
            Broadcast
          </button>
        </div>
      </div>
    </div>

    <!-- 编辑公告模态框 -->
    <div v-if="showEditAnnouncementModal && editingAnnouncement" class="fixed inset-0 bg-slate-900/40 dark:bg-black/80 backdrop-blur-md z-50 flex items-center justify-center p-4">
      <div class="bg-white dark:bg-[#0c0c0e] border border-slate-200 dark:border-white/10 rounded-[1.5rem] p-4 max-w-lg w-full shadow-[0_50px_100px_-20px_rgba(6,182,212,0.2)] animate-in zoom-in relative overflow-hidden">
        <div class="absolute top-0 right-0 w-40 h-40 bg-cyan-500/5 blur-[60px] -mr-20 -mt-20" />

        <h3 class="text-lg font-black italic uppercase text-slate-900 dark:text-white mb-8 flex items-center gap-4 relative z-10">
          <Edit2 class="w-6 h-6 text-cyan-500" />
          编辑广播 <span class="text-[10px] text-slate-400 font-mono not-italic tracking-normal">/ EDIT_BROADCAST</span>
        </h3>

        <div class="space-y-6 relative z-10">
          <div>
            <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-[0.2em] block mb-3 ml-1">Title / 标题 (可选)</label>
            <input
              v-model="editingAnnouncement.title"
              type="text"
              placeholder="ENTER BROADCAST TITLE..."
              class="w-full bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-5 py-2.5 text-xs font-bold text-slate-900 dark:text-white focus:outline-none focus:border-cyan-500/50 transition-all"
            />
          </div>

          <div>
            <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-[0.2em] block mb-3 ml-1">Broadcast Content / 内容</label>
            <textarea
              v-model="editingAnnouncement.content"
              rows="3"
              placeholder="ENTER MESSAGE TO RESEARCHERS..."
              class="w-full bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-5 py-2.5 text-xs font-bold text-slate-900 dark:text-white focus:outline-none focus:border-cyan-500/50 transition-all resize-none"
            />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-[0.2em] block mb-3 ml-1">Priority / 优先级</label>
              <select
                v-model="editingAnnouncement.type"
                class="w-full bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2.5 text-[10px] font-black uppercase tracking-widest text-slate-900 dark:text-white focus:outline-none focus:border-cyan-500/30"
              >
                <option value="info">INFO / 公告</option>
                <option value="maintenance">MAINTENANCE / 维护</option>
                <option value="emergency">EMERGENCY / 紧急</option>
              </select>
            </div>
            <div>
              <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-[0.2em] block mb-3 ml-1">Format / 展现形式</label>
              <select
                v-model="editingAnnouncement.is_ticker"
                class="w-full bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2.5 text-[10px] font-black uppercase tracking-widest text-slate-900 dark:text-white focus:outline-none focus:border-cyan-500/30"
              >
                <option :value="true">TICKER / 跑马灯</option>
                <option :value="false">MODAL / 强制弹窗</option>
              </select>
            </div>
          </div>

          <div>
             <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-[0.2em] block mb-3 ml-1">TTL / 有效时长 (e.g. 24h, 7d, 30m)</label>
             <input
              v-model="editingAnnouncement.expires_in"
              type="text"
              placeholder="24h"
              class="w-full bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-5 py-2.5 text-xs font-bold font-mono text-slate-900 dark:text-white focus:outline-none focus:border-cyan-500/50 tracking-widest uppercase"
            />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="flex items-center gap-3 p-4 bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl">
              <input
                type="checkbox"
                v-model="editingAnnouncement.on_join"
                class="w-4 h-4 rounded border-slate-300 text-cyan-600 focus:ring-cyan-500"
              />
              <label class="text-[10px] font-black text-slate-600 dark:text-slate-400 uppercase tracking-widest">玩家加入时触发</label>
            </div>
            <div class="flex items-center gap-3 p-4 bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl">
              <input
                type="checkbox"
                v-model="editingAnnouncement.is_persistent"
                class="w-4 h-4 rounded border-slate-300 text-cyan-600 focus:ring-cyan-500"
              />
              <label class="text-[10px] font-black text-slate-600 dark:text-slate-400 uppercase tracking-widest">常驻显示</label>
            </div>
          </div>

          <div>
              <label class="text-[8px] font-black text-slate-400 uppercase tracking-widest block mb-1">自动循环间隔 (分钟) / 0 = 禁止</label>
              <input
                v-model.number="editingAnnouncement.cron_interval"
                type="number"
                placeholder="0"
                class="w-full bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2 text-[10px] font-black text-slate-900 dark:text-white focus:outline-none focus:border-cyan-500/30"
              />
          </div>

          <div v-if="!editingAnnouncement.is_ticker">
             <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-[0.2em] block mb-3 ml-1">Force Close Delay / 强制等待时间 (秒)</label>
             <input
              v-model.number="editingAnnouncement.close_delay"
              type="number"
              placeholder="0 = 立即关闭"
              class="w-full bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-5 py-2.5 text-xs font-bold font-mono text-slate-900 dark:text-white focus:outline-none focus:border-cyan-500/50 tracking-widest uppercase"
            />
          </div>
        </div>

        <div class="flex gap-4 mt-10 relative z-10">
          <button
            @click="showEditAnnouncementModal = false"
            class="flex-1 px-4 py-2.5 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-500 rounded-[1.25rem] font-black text-[10px] uppercase tracking-widest transition-all"
          >
            Cancel
          </button>
          <button
            @click="handleUpdateAnnouncement"
            class="flex-1 px-4 py-2.5 bg-cyan-600 hover:bg-cyan-500 text-white rounded-[1.25rem] font-black text-[10px] uppercase tracking-widest transition-all shadow-lg shadow-cyan-500/20 active:scale-95 border border-cyan-500/20"
          >
            Update
          </button>
        </div>
      </div>
    </div>
  </div>
</template>


<style src="./Admin.css" scoped></style>
