<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { pluginAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import {
  ArrowLeft,
  Plus,
  Trash2,
  Edit2,
  Save,
  X,
  RefreshCw,
  Puzzle,
  Upload,
  Power,
  ToggleLeft,
  ToggleRight
} from 'lucide-vue-next'
import { cn } from '../utils/cn'

const router = useRouter()
const { showAlert, showConfirm } = useDialog()

// --- State ---
const loading = ref(false)
const plugins = ref<any[]>([])
const selectedPlugin = ref<any>(null)
const pluginCards = ref<any[]>([])

// Plugin form
const showPluginForm = ref(false)
const editingPlugin = ref<any>(null)
const pluginForm = ref({ name: '', description: '' })

// Card form
const showCardForm = ref(false)
const editingCard = ref<any>(null)
const cardForm = ref({
  symbol: '',
  display_name: '',
  effect_type: 'swap',
  effect_config_raw: '{"count": 2}',
  default_count: 2,
  color: '#06b6d4'
})

// .cumod 安装
const showInstallModal = ref(false)
const installFile = ref<File | null>(null)
const installing = ref(false)
const installFileInput = ref<HTMLInputElement | null>(null)

// 服务器重启
const showRestartModal = ref(false)
const restartDelay = ref(30)
const restartReason = ref('')
const restartScheduled = ref(false)
const restartCountdown = ref(0)
let restartTimer: ReturnType<typeof setInterval> | null = null

// Effect type options
const effectTypes = [
  { value: 'swap', label: '交换 (swap)', hint: '{"count": 3}  —— 随机换 N 张手牌' },
  { value: 'force_play', label: '强制出牌 (force_play)', hint: '{"count": 2}  —— 下一位必须额外打 N 张' },
  { value: 'convert', label: '卡牌转换 (convert)', hint: '{"source_count": 2, "target_count": 4}  —— 消耗自身 N 张，摸 M 张' },
]

const effectHint = computed(() => {
  return effectTypes.find(e => e.value === cardForm.value.effect_type)?.hint || ''
})

// --- API calls ---
const loadPlugins = async () => {
  loading.value = true
  try {
    const res = await pluginAPI.getPlugins()
    plugins.value = res.data || []
  } catch (e: any) {
    await showAlert(e.response?.data?.error || '加载插件列表失败', '错误')
  } finally {
    loading.value = false
  }
}

const loadCards = async (pluginId: number) => {
  try {
    const res = await pluginAPI.getPluginCardsByPlugin(pluginId)
    pluginCards.value = res.data || []
  } catch (e: any) {
    await showAlert(e.response?.data?.error || '加载卡牌失败', '错误')
  }
}

const selectPlugin = async (plugin: any) => {
  selectedPlugin.value = plugin
  await loadCards(plugin.id)
}

// --- Plugin CRUD ---
const openNewPlugin = () => {
  editingPlugin.value = null
  pluginForm.value = { name: '', description: '' }
  showPluginForm.value = true
}

const openEditPlugin = (plugin: any) => {
  editingPlugin.value = plugin
  pluginForm.value = { name: plugin.name, description: plugin.description || '' }
  showPluginForm.value = true
}

const savePlugin = async () => {
  if (!pluginForm.value.name.trim()) {
    await showAlert('插件名称不能为空', '提示')
    return
  }
  try {
    if (editingPlugin.value) {
      await pluginAPI.updatePlugin(editingPlugin.value.id, pluginForm.value)
    } else {
      await pluginAPI.createPlugin(pluginForm.value)
    }
    showPluginForm.value = false
    await loadPlugins()
  } catch (e: any) {
    await showAlert(e.response?.data?.error || '保存失败', '错误')
  }
}

const togglePluginActive = async (plugin: any) => {
  try {
    await pluginAPI.updatePlugin(plugin.id, { is_active: !plugin.is_active })
    await loadPlugins()
    if (selectedPlugin.value?.id === plugin.id) {
      selectedPlugin.value = plugins.value.find(p => p.id === plugin.id)
    }
  } catch (e: any) {
    await showAlert(e.response?.data?.error || '操作失败', '错误')
  }
}

const deletePlugin = async (plugin: any) => {
  const confirmed = await showConfirm(
    `确定删除插件「${plugin.name}」及其所有卡牌吗？此操作不可撤销。`,
    '删除插件'
  )
  if (!confirmed) return
  try {
    await pluginAPI.deletePlugin(plugin.id)
    if (selectedPlugin.value?.id === plugin.id) {
      selectedPlugin.value = null
      pluginCards.value = []
    }
    await loadPlugins()
  } catch (e: any) {
    await showAlert(e.response?.data?.error || '删除失败', '错误')
  }
}

// --- Card CRUD ---
const openNewCard = () => {
  if (!selectedPlugin.value) return
  editingCard.value = null
  cardForm.value = {
    symbol: '',
    display_name: '',
    effect_type: 'swap',
    effect_config_raw: '{"count": 2}',
    default_count: 2,
    color: '#06b6d4'
  }
  showCardForm.value = true
}

const openEditCard = (card: any) => {
  editingCard.value = card
  cardForm.value = {
    symbol: card.symbol,
    display_name: card.display_name || '',
    effect_type: card.effect_type,
    effect_config_raw: card.effect_config,
    default_count: card.default_count,
    color: card.color || '#06b6d4'
  }
  showCardForm.value = true
}

const saveCard = async () => {
  if (!cardForm.value.symbol.trim()) {
    await showAlert('卡牌符号不能为空', '提示')
    return
  }
  let effectConfig: object
  try {
    effectConfig = JSON.parse(cardForm.value.effect_config_raw)
  } catch {
    await showAlert('效果配置 JSON 格式错误，请检查', '格式错误')
    return
  }
  const payload = {
    symbol: cardForm.value.symbol.trim().toUpperCase(),
    display_name: cardForm.value.display_name,
    effect_type: cardForm.value.effect_type,
    effect_config: effectConfig,
    default_count: Number(cardForm.value.default_count) || 2,
    color: cardForm.value.color
  }
  try {
    if (editingCard.value) {
      await pluginAPI.updateCard(editingCard.value.id, payload)
    } else {
      await pluginAPI.createCard(selectedPlugin.value.id, payload)
    }
    showCardForm.value = false
    await loadCards(selectedPlugin.value.id)
  } catch (e: any) {
    await showAlert(e.response?.data?.error || '保存卡牌失败', '错误')
  }
}

const deleteCard = async (card: any) => {
  const confirmed = await showConfirm(`确定删除卡牌「${card.symbol}」吗？`, '删除卡牌')
  if (!confirmed) return
  try {
    await pluginAPI.deleteCard(card.id)
    await loadCards(selectedPlugin.value.id)
  } catch (e: any) {
    await showAlert(e.response?.data?.error || '删除失败', '错误')
  }
}

const handleReload = async () => {
  try {
    const res = await pluginAPI.reloadPlugins()
    await showAlert(`已重载，共 ${res.data.count} 张插件卡牌`, '热重载完成')
  } catch (e: any) {
    await showAlert(e.response?.data?.error || '重载失败', '错误')
  }
}

// --- .cumod 安装 ---
const onInstallFileChange = (e: Event) => {
  const input = e.target as HTMLInputElement
  installFile.value = input.files?.[0] ?? null
}

const handleInstall = async () => {
  if (!installFile.value) {
    await showAlert('请先选择 .cumod 文件', '提示')
    return
  }
  installing.value = true
  try {
    const res = await pluginAPI.installPlugin(installFile.value)
    const d = res.data
    showInstallModal.value = false
    installFile.value = null
    await loadPlugins()
    await showAlert(
      `插件「${d.plugin?.name}」v${d.plugin?.version} 安装成功，共载入 ${d.count} 张卡牌。`,
      '安装成功 ✅'
    )
  } catch (e: any) {
    await showAlert(e.response?.data?.error || '安装失败', '安装错误')
  } finally {
    installing.value = false
    if (installFileInput.value) installFileInput.value.value = ''
  }
}

// --- 服务器重启 ---
const handleScheduleRestart = async () => {
  if (restartDelay.value < 10 || restartDelay.value > 300) {
    await showAlert('重启延迟必须在 10 ~ 300 秒之间', '参数错误')
    return
  }
  try {
    await pluginAPI.scheduleRestart(restartDelay.value, restartReason.value || undefined)
    restartScheduled.value = true
    restartCountdown.value = restartDelay.value
    showRestartModal.value = false
    // 本地倒计时（仅显示用，实际由服务端控制）
    restartTimer = setInterval(() => {
      restartCountdown.value--
      if (restartCountdown.value <= 0) {
        clearInterval(restartTimer!)
        restartTimer = null
        restartScheduled.value = false
      }
    }, 1000)
  } catch (e: any) {
    await showAlert(e.response?.data?.error || '安排重启失败', '错误')
  }
}

const handleCancelRestart = async () => {
  try {
    await pluginAPI.cancelRestart()
    if (restartTimer) { clearInterval(restartTimer); restartTimer = null }
    restartScheduled.value = false
    restartCountdown.value = 0
    await showAlert('服务器重启已取消', '已取消')
  } catch (e: any) {
    await showAlert(e.response?.data?.error || '取消失败', '错误')
  }
}

onMounted(loadPlugins)
</script>

<template>
  <div class="min-h-screen bg-slate-950 text-white">
    <!-- Header -->
    <div class="sticky top-0 z-30 bg-slate-950/90 backdrop-blur border-b border-white/5 px-3 py-2 flex items-center gap-2 flex-wrap">
      <button @click="router.back()" class="p-1.5 hover:bg-white/10 rounded-md transition-colors">
        <ArrowLeft class="w-4 h-4" />
      </button>
      <Puzzle class="w-5 h-5 text-purple-400" />
      <h1 class="text-xs font-black uppercase tracking-widest text-white">插件系统管理</h1>

      <!-- 重启倒计时徽标 -->
      <span
        v-if="restartScheduled"
        class="flex items-center gap-1 px-2.5 py-0.5 rounded-full bg-orange-500/20 border border-orange-500/40 text-orange-300 text-[11px] font-bold animate-pulse"
      >
        ⏱ 重启倒计时 {{ restartCountdown }}s
        <button @click="handleCancelRestart" class="ml-1 text-red-400 hover:text-red-300 font-black">✕</button>
      </span>

      <div class="ml-auto flex items-center gap-2 flex-wrap">
        <button @click="handleReload" class="flex items-center gap-1 px-2.5 py-1 bg-emerald-600 hover:bg-emerald-500 rounded-md text-[11px] font-bold transition-colors">
          <RefreshCw class="w-3 h-3" />
          热重载
        </button>
        <button @click="showInstallModal = true" class="flex items-center gap-1 px-2.5 py-1 bg-indigo-600 hover:bg-indigo-500 rounded-md text-[11px] font-bold transition-colors">
          <Upload class="w-3 h-3" />
          安装 .cumod
        </button>
        <button @click="openNewPlugin" class="flex items-center gap-1 px-2.5 py-1 bg-purple-600 hover:bg-purple-500 rounded-md text-[11px] font-bold transition-colors">
          <Plus class="w-3 h-3" />
          新建插件
        </button>
        <button
          @click="restartScheduled ? handleCancelRestart() : (showRestartModal = true)"
          :class="cn(
            'flex items-center gap-1 px-2.5 py-1 rounded-md text-[11px] font-bold transition-colors',
            restartScheduled ? 'bg-red-700 hover:bg-red-600' : 'bg-red-600/70 hover:bg-red-600'
          )"
        >
          <Power class="w-3 h-3" />
          {{ restartScheduled ? '取消重启' : '重启服务器' }}
        </button>
      </div>
    </div>

    <div class="flex h-[calc(100vh-49px)]">
      <!-- Left: Plugin List -->
      <div class="w-64 shrink-0 border-r border-white/5 overflow-y-auto">
        <div class="p-2 space-y-1.5">
          <div v-if="loading" class="text-center text-slate-500 text-xs py-8">加载中...</div>
          <div v-else-if="!plugins.length" class="text-center text-slate-500 text-xs py-8">暂无插件，点击「安装 .cumod」或「新建插件」</div>

          <div
            v-for="plugin in plugins"
            :key="plugin.id"
            @click="selectPlugin(plugin)"
            :class="cn(
              'p-2 rounded-lg border cursor-pointer transition-all group',
              selectedPlugin?.id === plugin.id
                ? 'bg-purple-600/20 border-purple-500/50'
                : 'bg-white/3 border-white/5 hover:bg-white/5'
            )"
          >
            <div class="flex items-start justify-between gap-2">
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <span class="text-[12px] font-semibold truncate leading-tight">{{ plugin.name }}</span>
                  <span
                    :class="cn('text-[9px] font-black uppercase px-1 py-0.5 rounded-full', plugin.is_active ? 'bg-emerald-500/20 text-emerald-400' : 'bg-slate-500/20 text-slate-400')"
                  >{{ plugin.is_active ? 'ON' : 'OFF' }}</span>
                </div>
                <p v-if="plugin.version" class="text-[9px] text-slate-500 font-mono mt-0.5 truncate">v{{ plugin.version }}<span v-if="plugin.author"> · {{ plugin.author }}</span></p>
                <p v-if="plugin.description" class="text-[10px] text-slate-400 mt-0.5 truncate leading-tight">{{ plugin.description }}</p>
              </div>
              <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity shrink-0">
                <button @click.stop="openEditPlugin(plugin)" class="p-1 hover:bg-white/10 rounded-lg" title="编辑">
                  <Edit2 class="w-3 h-3 text-slate-400" />
                </button>
                <button @click.stop="togglePluginActive(plugin)" class="p-1 hover:bg-white/10 rounded-lg" :title="plugin.is_active ? '停用' : '启用'">
                  <ToggleRight v-if="plugin.is_active" class="w-4 h-4 text-emerald-400" />
                  <ToggleLeft v-else class="w-4 h-4 text-slate-400" />
                </button>
                <button @click.stop="deletePlugin(plugin)" class="p-1 hover:bg-red-500/20 rounded-lg" title="删除">
                  <Trash2 class="w-3 h-3 text-red-400" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Right: Cards Panel -->
      <div class="flex-1 overflow-y-auto">
        <div v-if="!selectedPlugin" class="flex items-center justify-center h-full text-slate-500 text-sm">
          ← 选择一个插件查看卡牌
        </div>

        <div v-else class="p-3">
          <!-- Plugin Header -->
          <div class="flex items-center justify-between mb-2">
            <div>
              <h2 class="text-sm font-black">{{ selectedPlugin.name }}</h2>
              <p class="text-[11px] text-slate-400">{{ pluginCards.length }} 张卡牌</p>
            </div>
            <button @click="openNewCard" class="flex items-center gap-1 px-2.5 py-1 bg-blue-600 hover:bg-blue-500 rounded-md text-[11px] font-bold transition-colors">
              <Plus class="w-3 h-3" />
              添加卡牌
            </button>
          </div>

          <!-- Cards Grid -->
          <div v-if="!pluginCards.length" class="text-center text-slate-500 text-xs py-12">暂无卡牌，点击「添加卡牌」</div>
          <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-2">
            <div
              v-for="card in pluginCards"
              :key="card.id"
              class="bg-white/3 border border-white/5 rounded-lg p-2 group hover:bg-white/5 transition-colors"
            >
              <!-- Card visual -->
              <div
                class="w-8 h-8 rounded-md flex items-center justify-center text-[10px] font-black mb-1.5 mx-auto"
                :style="{ backgroundColor: (card.color || '#06b6d4') + '33', border: '1px solid ' + (card.color || '#06b6d4') + '66', color: card.color || '#06b6d4' }"
              >
                {{ card.symbol.slice(0, 4) }}
              </div>
              <div class="text-center">
                <div class="text-[11px] font-bold text-white truncate leading-tight">{{ card.display_name || card.symbol }}</div>
                <div class="text-[9px] text-slate-400 mt-0.5">{{ card.effect_type }}</div>
                <div class="text-[9px] text-slate-500 mt-0.5 font-mono">×{{ card.default_count }}</div>
              </div>
              <!-- Actions -->
              <div class="flex items-center justify-center gap-1.5 mt-1.5 opacity-0 group-hover:opacity-100 transition-opacity">
                <button @click="openEditCard(card)" class="p-1 hover:bg-white/10 rounded-lg" title="编辑">
                  <Edit2 class="w-3 h-3 text-slate-400" />
                </button>
                <button @click="deleteCard(card)" class="p-1 hover:bg-red-500/20 rounded-lg" title="删除">
                  <Trash2 class="w-3 h-3 text-red-400" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- .cumod 安装 Modal -->
    <div v-if="showInstallModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div class="bg-slate-900 border border-white/10 rounded-2xl w-full max-w-md shadow-2xl">
        <div class="flex items-center justify-between p-4 border-b border-white/5">
          <div class="flex items-center gap-2">
            <Upload class="w-4 h-4 text-indigo-400" />
            <h3 class="font-black text-sm uppercase tracking-widest">安装 .cumod 插件</h3>
          </div>
          <button @click="showInstallModal = false; installFile = null" class="p-1 hover:bg-white/10 rounded-lg">
            <X class="w-4 h-4" />
          </button>
        </div>
        <div class="p-4 space-y-4">
          <p class="text-xs text-slate-400 leading-relaxed">
            .cumod 文件是 Chemistry UNO 插件格式（ZIP 压缩包），包含 <code class="text-indigo-300">manifest.json</code>、可选的 <code class="text-indigo-300">cards.json</code>，以及可选的 <code class="text-indigo-300">client.js</code>/<code class="text-indigo-300">server.js</code>。
          </p>
          <div
            class="border-2 border-dashed border-white/10 rounded-xl p-6 text-center cursor-pointer hover:border-indigo-500/50 hover:bg-indigo-500/5 transition-all"
            @click="installFileInput?.click()"
          >
            <Upload class="w-8 h-8 text-slate-500 mx-auto mb-2" />
            <p class="text-sm text-slate-400">
              {{ installFile ? installFile.name : '点击选择 .cumod 文件' }}
            </p>
            <p class="text-xs text-slate-600 mt-1">仅支持 .cumod 格式</p>
          </div>
          <input
            ref="installFileInput"
            type="file"
            accept=".cumod"
            class="hidden"
            @change="onInstallFileChange"
          />
        </div>
        <div class="flex gap-2 p-4 border-t border-white/5">
          <button @click="showInstallModal = false; installFile = null" class="flex-1 py-2 bg-white/5 hover:bg-white/10 rounded-lg text-sm font-bold transition-colors">取消</button>
          <button
            @click="handleInstall"
            :disabled="!installFile || installing"
            class="flex-1 py-2 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed rounded-lg text-sm font-bold transition-colors flex items-center justify-center gap-1.5"
          >
            <svg v-if="installing" class="animate-spin w-3.5 h-3.5" viewBox="0 0 24 24" fill="none">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
            </svg>
            <Upload v-else class="w-3.5 h-3.5" />
            {{ installing ? '安装中…' : '安装插件' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 服务器重启 Modal -->
    <div v-if="showRestartModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div class="bg-slate-900 border border-red-500/20 rounded-2xl w-full max-w-md shadow-2xl">
        <div class="flex items-center justify-between p-4 border-b border-white/5">
          <div class="flex items-center gap-2">
            <Power class="w-4 h-4 text-red-400" />
            <h3 class="font-black text-sm uppercase tracking-widest text-red-300">安排服务器重启</h3>
          </div>
          <button @click="showRestartModal = false" class="p-1 hover:bg-white/10 rounded-lg">
            <X class="w-4 h-4" />
          </button>
        </div>
        <div class="p-4 space-y-4">
          <div class="rounded-lg bg-red-500/10 border border-red-500/20 p-3 text-xs text-red-300">
            ⚠️ 重启将断开所有玩家连接并终止进行中的游戏。请确保在非高峰时段操作。
          </div>
          <div>
            <label class="text-[11px] font-bold text-slate-400 uppercase tracking-wider block mb-1">倒计时（秒）10 ~ 300</label>
            <input
              v-model.number="restartDelay"
              type="number"
              min="10"
              max="300"
              class="w-full bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-red-500 transition-colors"
            />
          </div>
          <div>
            <label class="text-[11px] font-bold text-slate-400 uppercase tracking-wider block mb-1">重启原因（广播给玩家）</label>
            <input
              v-model="restartReason"
              type="text"
              placeholder="如：服务器维护，请稍候"
              class="w-full bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-red-500 transition-colors"
            />
          </div>
        </div>
        <div class="flex gap-2 p-4 border-t border-white/5">
          <button @click="showRestartModal = false" class="flex-1 py-2 bg-white/5 hover:bg-white/10 rounded-lg text-sm font-bold transition-colors">取消</button>
          <button
            @click="handleScheduleRestart"
            class="flex-1 py-2 bg-red-600 hover:bg-red-500 rounded-lg text-sm font-bold transition-colors flex items-center justify-center gap-1.5"
          >
            <Power class="w-3.5 h-3.5" />
            确认安排重启
          </button>
        </div>
      </div>
    </div>

    <!-- Plugin Form Modal -->
    <div v-if="showPluginForm" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div class="bg-slate-900 border border-white/10 rounded-2xl w-full max-w-md shadow-2xl">
        <div class="flex items-center justify-between p-4 border-b border-white/5">
          <h3 class="font-black text-sm uppercase tracking-widest">{{ editingPlugin ? '编辑插件' : '新建插件' }}</h3>
          <button @click="showPluginForm = false" class="p-1 hover:bg-white/10 rounded-lg">
            <X class="w-4 h-4" />
          </button>
        </div>
        <div class="p-4 space-y-3">
          <div>
            <label class="text-[11px] font-bold text-slate-400 uppercase tracking-wider block mb-1">插件名称 *</label>
            <input v-model="pluginForm.name" type="text" placeholder="如：交换卡包" class="w-full bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-purple-500 transition-colors" />
          </div>
          <div>
            <label class="text-[11px] font-bold text-slate-400 uppercase tracking-wider block mb-1">描述</label>
            <textarea v-model="pluginForm.description" placeholder="插件功能说明（可选）" rows="3" class="w-full bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-purple-500 transition-colors resize-none" />
          </div>
        </div>
        <div class="flex gap-2 p-4 border-t border-white/5">
          <button @click="showPluginForm = false" class="flex-1 py-2 bg-white/5 hover:bg-white/10 rounded-lg text-sm font-bold transition-colors">取消</button>
          <button @click="savePlugin" class="flex-1 py-2 bg-purple-600 hover:bg-purple-500 rounded-lg text-sm font-bold transition-colors flex items-center justify-center gap-1.5">
            <Save class="w-3.5 h-3.5" />
            保存
          </button>
        </div>
      </div>
    </div>

    <!-- Card Form Modal -->
    <div v-if="showCardForm" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div class="bg-slate-900 border border-white/10 rounded-2xl w-full max-w-lg shadow-2xl">
        <div class="flex items-center justify-between p-4 border-b border-white/5">
          <h3 class="font-black text-sm uppercase tracking-widest">{{ editingCard ? '编辑卡牌' : '添加卡牌' }}</h3>
          <button @click="showCardForm = false" class="p-1 hover:bg-white/10 rounded-lg">
            <X class="w-4 h-4" />
          </button>
        </div>
        <div class="p-4 space-y-3 max-h-[70vh] overflow-y-auto">
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="text-[11px] font-bold text-slate-400 uppercase tracking-wider block mb-1">符号 (Symbol) *</label>
              <input
                v-model="cardForm.symbol"
                type="text"
                placeholder="如：SWAP3"
                :disabled="!!editingCard"
                class="w-full bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm font-mono focus:outline-none focus:border-blue-500 transition-colors disabled:opacity-50"
              />
            </div>
            <div>
              <label class="text-[11px] font-bold text-slate-400 uppercase tracking-wider block mb-1">显示名称</label>
              <input v-model="cardForm.display_name" type="text" placeholder="如：三重交换" class="w-full bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-blue-500 transition-colors" />
            </div>
          </div>

          <div>
            <label class="text-[11px] font-bold text-slate-400 uppercase tracking-wider block mb-1">效果类型 *</label>
            <select v-model="cardForm.effect_type" class="w-full bg-slate-800 border border-white/10 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-blue-500 transition-colors">
              <option v-for="et in effectTypes" :key="et.value" :value="et.value">{{ et.label }}</option>
            </select>
            <p class="text-[10px] text-slate-500 mt-1 font-mono">{{ effectHint }}</p>
          </div>

          <div>
            <label class="text-[11px] font-bold text-slate-400 uppercase tracking-wider block mb-1">效果配置 (JSON) *</label>
            <textarea
              v-model="cardForm.effect_config_raw"
              rows="3"
              class="w-full bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm font-mono focus:outline-none focus:border-blue-500 transition-colors resize-none"
            />
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="text-[11px] font-bold text-slate-400 uppercase tracking-wider block mb-1">默认牌组数量</label>
              <input v-model.number="cardForm.default_count" type="number" min="1" max="10" class="w-full bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-blue-500 transition-colors" />
            </div>
            <div>
              <label class="text-[11px] font-bold text-slate-400 uppercase tracking-wider block mb-1">颜色 (HEX)</label>
              <div class="flex items-center gap-2">
                <input v-model="cardForm.color" type="color" class="w-9 h-9 bg-transparent border-0 cursor-pointer rounded p-0" />
                <input v-model="cardForm.color" type="text" class="flex-1 bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm font-mono focus:outline-none focus:border-blue-500 transition-colors" />
              </div>
            </div>
          </div>
        </div>
        <div class="flex gap-2 p-4 border-t border-white/5">
          <button @click="showCardForm = false" class="flex-1 py-2 bg-white/5 hover:bg-white/10 rounded-lg text-sm font-bold transition-colors">取消</button>
          <button @click="saveCard" class="flex-1 py-2 bg-blue-600 hover:bg-blue-500 rounded-lg text-sm font-bold transition-colors flex items-center justify-center gap-1.5">
            <Save class="w-3.5 h-3.5" />
            保存卡牌
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
