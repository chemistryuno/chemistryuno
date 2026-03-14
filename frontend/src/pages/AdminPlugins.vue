<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { adminAPI, pluginAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import { listRegisteredPluginRoutes, refreshPluginConfiguredRoutes } from '../utils/plugin-runtime'
import { ArrowLeft, Power, Puzzle, RefreshCw, Save, ToggleLeft, ToggleRight, Trash2, Upload } from 'lucide-vue-next'

type Plugin = {
  id: number
  name: string
  description?: string
  version?: string
  is_active: boolean
  card_count?: number
  has_client_script?: boolean
  has_server_script?: boolean
  config_schema?: string
}

type SchemaField = {
  key: string
  label?: string
  type?: string
  description?: string
  required?: boolean
  read_only?: boolean
  accept?: string
  max_size_kb?: number
}

type RoutePage = {
  path: string
  title: string
  description: string
  content_html: string
  requires_auth: boolean
  admin_only: boolean
  co_worker_only: boolean
}

type SettingsSnapshot = {
  id: string
  created_at: string
  created_by: number
  settings: Record<string, string>
}

const router = useRouter()
const { showAlert, showConfirm } = useDialog()

const loading = ref(false)
const saving = ref(false)
const settingsLoading = ref(false)
const plugins = ref<Plugin[]>([])
const selectedPluginId = ref<number | null>(null)
const schemaFields = ref<SchemaField[]>([])
const settingValues = ref<Record<string, string>>({})
const routePagesByKey = ref<Record<string, RoutePage[]>>({})
const installFileInput = ref<HTMLInputElement | null>(null)
const settingsHistory = ref<SettingsSnapshot[]>([])
const selectedSnapshotId = ref('')

const showRestartModal = ref(false)
const restartDelay = ref(30)
const restartReason = ref('')
const restartScheduled = ref(false)
const restartCountdown = ref(0)
let restartTimer: ReturnType<typeof setInterval> | null = null

const selectedPlugin = computed(() => plugins.value.find((p) => p.id === selectedPluginId.value) ?? null)
const runtimeRoutes = computed(() => selectedPlugin.value ? listRegisteredPluginRoutes(selectedPlugin.value.id) : [])

function parseSchema(raw?: string) {
  if (!raw || !raw.trim()) return []
  try {
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed.filter((item) => item && typeof item.key === 'string') : []
  } catch {
    return []
  }
}

function parseRoutePages(raw: string): RoutePage[] {
  if (!raw || !raw.trim()) return []
  try {
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return parsed
      .filter((item) => item && typeof item.path === 'string')
      .map((item) => ({
        path: String(item.path || ''),
        title: String(item.title || ''),
        description: String(item.description || ''),
        content_html: String(item.content_html || ''),
        requires_auth: item.requires_auth !== false,
        admin_only: Boolean(item.admin_only),
        co_worker_only: Boolean(item.co_worker_only)
      }))
  } catch {
    return []
  }
}

function stringifyRoutePages(pages: RoutePage[]) {
  return JSON.stringify(pages.map((page) => ({
    path: page.path,
    title: page.title,
    description: page.description,
    content_html: page.content_html,
    requires_auth: page.requires_auth,
    admin_only: page.admin_only,
    co_worker_only: page.co_worker_only
  })))
}

function fieldType(field: SchemaField) {
  return (field.type || 'text').toLowerCase()
}

function fieldLabel(field: SchemaField) {
  return field.label || field.key
}

function addRoutePage(key: string) {
  if (!routePagesByKey.value[key]) routePagesByKey.value[key] = []
  routePagesByKey.value[key].push({
    path: '',
    title: '',
    description: '',
    content_html: '',
    requires_auth: true,
    admin_only: false,
    co_worker_only: false
  })
}

function removeRoutePage(key: string, index: number) {
  const pages = routePagesByKey.value[key] || []
  routePagesByKey.value[key] = pages.filter((_, i) => i !== index)
}

async function onUploadFile(field: SchemaField, e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  const maxSize = (field.max_size_kb || 1024) * 1024
  if (file.size > maxSize) {
    await showAlert(`文件超过 ${(field.max_size_kb || 1024)}KB`, '上传失败')
    input.value = ''
    return
  }
  const reader = new FileReader()
  reader.onload = () => {
    settingValues.value[field.key] = typeof reader.result === 'string' ? reader.result : ''
  }
  reader.readAsDataURL(file)
}

async function loadPlugins() {
  loading.value = true
  try {
    const res = await pluginAPI.getPlugins()
    plugins.value = res.data || []
    if (!plugins.value.length) return
    if (!plugins.value.some((p) => p.id === selectedPluginId.value)) {
      selectedPluginId.value = plugins.value[0].id
    }
    if (selectedPluginId.value) await loadSettings(selectedPluginId.value)
  } catch (e: any) {
    await showAlert(e?.response?.data?.error || '加载插件失败', '错误')
  } finally {
    loading.value = false
  }
}

async function selectPlugin(plugin: Plugin) {
  selectedPluginId.value = plugin.id
  await loadSettings(plugin.id)
}

async function loadSettings(pluginId: number) {
  settingsLoading.value = true
  try {
    const plugin = plugins.value.find((item) => item.id === pluginId)
    const schemaFromPlugin = parseSchema(plugin?.config_schema)
    const res = await pluginAPI.getPluginSettings(pluginId)
    const schemaFromAPI = Array.isArray(res?.data?.schema) ? res.data.schema : []
    schemaFields.value = schemaFromAPI.length ? schemaFromAPI : schemaFromPlugin
    settingValues.value = { ...(res?.data?.settings || {}) }
    routePagesByKey.value = {}
    for (const field of schemaFields.value) {
      if (fieldType(field) !== 'route_list') continue
      routePagesByKey.value[field.key] = parseRoutePages(settingValues.value[field.key] || '[]')
    }
    await loadSettingsHistory(pluginId)
  } catch (e: any) {
    await showAlert(e?.response?.data?.error || '加载配置失败', '错误')
  } finally {
    settingsLoading.value = false
  }
}

async function loadSettingsHistory(pluginId: number) {
  try {
    const res = await pluginAPI.getPluginSettingsHistory(pluginId)
    settingsHistory.value = Array.isArray(res?.data?.history) ? res.data.history : []
    const latest = settingsHistory.value[settingsHistory.value.length - 1]
    selectedSnapshotId.value = latest?.id || ''
  } catch {
    settingsHistory.value = []
    selectedSnapshotId.value = ''
  }
}

function validateRoutePagesBeforeSave() {
  const seen = new Set<string>()
  const existingPaths = new Set(router.getRoutes().map((r) => r.path))
  const currentPluginPaths = new Set(runtimeRoutes.value.map((route) => route.path))

  for (const [key, pages] of Object.entries(routePagesByKey.value)) {
    for (let i = 0; i < pages.length; i++) {
      const path = String(pages[i].path || '').trim()
      if (!path) {
        throw new Error(`route_list(${key}) 第 ${i + 1} 项 path 不能为空`)
      }
      if (!path.startsWith('/')) {
        throw new Error(`route_list(${key}) 第 ${i + 1} 项 path 必须以 / 开头`)
      }
      if (path.includes(' ')) {
        throw new Error(`route_list(${key}) 第 ${i + 1} 项 path 不能包含空格`)
      }
      if (seen.has(path)) {
        throw new Error(`route_list 存在重复 path: ${path}`)
      }
      seen.add(path)
      if (existingPaths.has(path) && !currentPluginPaths.has(path)) {
        throw new Error(`路由冲突：${path} 已被系统或其他插件占用`)
      }
    }
  }
}

async function saveSettings() {
  if (!selectedPlugin.value) return
  const payload: Record<string, string> = {}
  try {
    validateRoutePagesBeforeSave()
  } catch (err: any) {
    await showAlert(err?.message || '路由校验失败', '错误')
    return
  }
  for (const field of schemaFields.value) {
    if (fieldType(field) === 'route_list') {
      payload[field.key] = stringifyRoutePages(routePagesByKey.value[field.key] || [])
    } else if (fieldType(field) === 'switch') {
      payload[field.key] = settingValues.value[field.key] === 'true' ? 'true' : 'false'
    } else {
      payload[field.key] = settingValues.value[field.key] || ''
    }
  }
  saving.value = true
  try {
    await pluginAPI.updatePluginSettings(selectedPlugin.value.id, payload)
    await refreshPluginConfiguredRoutes(selectedPlugin.value.id)
    await loadSettings(selectedPlugin.value.id)
    await showAlert('配置保存成功', '完成')
  } catch (e: any) {
    await showAlert(e?.response?.data?.error || '保存失败', '错误')
  } finally {
    saving.value = false
  }
}

async function rollbackSettings() {
  if (!selectedPlugin.value) return
  if (!selectedSnapshotId.value) {
    await showAlert('请先选择一个历史快照', '提示')
    return
  }
  try {
    await pluginAPI.rollbackPluginSettings(selectedPlugin.value.id, selectedSnapshotId.value)
    await refreshPluginConfiguredRoutes(selectedPlugin.value.id)
    await loadSettings(selectedPlugin.value.id)
    await showAlert('已回滚到指定快照', '成功')
  } catch (e: any) {
    await showAlert(e?.response?.data?.error || '回滚失败', '错误')
  }
}

async function togglePlugin(plugin: Plugin) {
  try {
    await pluginAPI.updatePlugin(plugin.id, { is_active: !plugin.is_active })
    await loadPlugins()
  } catch (e: any) {
    await showAlert(e?.response?.data?.error || '操作失败', '错误')
  }
}

async function deletePlugin(plugin: Plugin) {
  const confirmed = await showConfirm(`确定删除插件「${plugin.name}」吗？该插件配置与卡牌将一起删除。`, '删除插件')
  if (!confirmed) return
  try {
    await pluginAPI.deletePlugin(plugin.id)
    await showAlert('插件已删除', '完成')
    await loadPlugins()
  } catch (e: any) {
    await showAlert(e?.response?.data?.error || '删除插件失败', '错误')
  }
}

async function installPlugin(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try {
    await pluginAPI.installPlugin(file)
    await showAlert('安装成功', '完成')
    await loadPlugins()
  } catch (err: any) {
    await showAlert(err?.response?.data?.error || '安装失败', '错误')
  } finally {
    if (installFileInput.value) installFileInput.value.value = ''
  }
}

async function reloadPlugins() {
  try {
    await pluginAPI.reloadPlugins()
    await loadPlugins()
    await showAlert('热重载完成', '完成')
  } catch (e: any) {
    await showAlert(e?.response?.data?.error || '热重载失败', '错误')
  }
}

async function restoreDefaultDeck() {
  const confirmed = await showConfirm('确定恢复全局卡组默认配置吗？这会覆盖当前全局卡组设置。', '恢复默认卡组')
  if (!confirmed) return
  try {
    await adminAPI.resetGlobalDeckConfig()
    await showAlert('全局卡组已恢复默认配置', '完成')
  } catch (e: any) {
    await showAlert(e?.response?.data?.error || '恢复默认卡组失败', '错误')
  }
}

async function scheduleRestart() {
  try {
    await pluginAPI.scheduleRestart(restartDelay.value, restartReason.value || undefined)
    showRestartModal.value = false
    restartScheduled.value = true
    restartCountdown.value = restartDelay.value
    if (restartTimer) clearInterval(restartTimer)
    restartTimer = setInterval(() => {
      restartCountdown.value--
      if (restartCountdown.value <= 0) {
        if (restartTimer) clearInterval(restartTimer)
        restartTimer = null
        restartScheduled.value = false
      }
    }, 1000)
  } catch (e: any) {
    await showAlert(e?.response?.data?.error || '安排重启失败', '错误')
  }
}

async function cancelRestart() {
  await pluginAPI.cancelRestart()
  if (restartTimer) clearInterval(restartTimer)
  restartTimer = null
  restartScheduled.value = false
  restartCountdown.value = 0
}

onMounted(loadPlugins)
</script>

<template>
  <div class="min-h-screen bg-slate-950 text-white">
    <div class="sticky top-0 z-30 bg-slate-950/90 backdrop-blur border-b border-white/5 px-3 py-2 flex items-center gap-2 flex-wrap">
      <button @click="router.back()" class="p-1.5 hover:bg-white/10 rounded-md"><ArrowLeft class="w-4 h-4" /></button>
      <Puzzle class="w-5 h-5 text-purple-400" />
      <h1 class="text-xs font-black uppercase tracking-widest">插件系统管理</h1>
      <span v-if="restartScheduled" class="text-[11px] px-2 py-0.5 rounded-full bg-orange-500/20 border border-orange-500/40 text-orange-300">重启 {{ restartCountdown }}s</span>
      <div class="ml-auto flex items-center gap-2">
        <button @click="reloadPlugins" class="px-4 py-1.5 bg-blue-600 hover:bg-blue-500 rounded-2xl text-[10px] font-black uppercase tracking-widest flex items-center gap-1.5 shadow-lg shadow-blue-500/20 active:scale-95 transition-all"><RefreshCw class="w-3 h-3" />热重载</button>
        <button @click="restoreDefaultDeck" class="px-4 py-1.5 bg-slate-800 hover:bg-slate-700 border border-white/10 rounded-2xl text-[10px] font-black uppercase tracking-widest active:scale-95 transition-all">恢复默认卡组</button>
        <button @click="installFileInput?.click()" class="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-500 rounded-2xl text-[10px] font-black uppercase tracking-widest flex items-center gap-1.5 shadow-lg shadow-indigo-500/20 active:scale-95 transition-all"><Upload class="w-3 h-3" />安装.CUMOD</button>
        <button @click="restartScheduled ? cancelRestart() : (showRestartModal = true)" :class="cn('px-4 py-1.5 rounded-2xl text-[10px] font-black uppercase tracking-widest flex items-center gap-1.5 shadow-lg active:scale-95 transition-all', restartScheduled ? 'bg-amber-600 hover:bg-amber-500 shadow-amber-500/20' : 'bg-red-600 hover:bg-red-500 shadow-red-500/20')"><Power class="w-3 h-3" />{{ restartScheduled ? '取消重启' : '重启服务器' }}</button>
      </div>
      <input ref="installFileInput" type="file" accept=".cumod" class="hidden" @change="installPlugin" />
    </div>

    <div class="flex h-[calc(100vh-49px)]">
      <div class="w-72 border-r border-white/5 overflow-y-auto p-2 space-y-1.5">
        <div v-if="loading" class="text-center text-slate-500 text-xs py-8">加载中...</div>
        <div v-else-if="!plugins.length" class="text-center text-slate-500 text-xs py-8">暂无插件</div>
        <div v-for="plugin in plugins" :key="plugin.id" @click="selectPlugin(plugin)"
             class="p-2 rounded-lg border cursor-pointer transition-all group"
             :class="selectedPluginId === plugin.id ? 'bg-purple-600/20 border-purple-500/50' : 'bg-white/3 border-white/5 hover:bg-white/5'">
          <div class="flex items-center justify-between gap-2">
            <div class="min-w-0">
              <p class="text-[12px] font-semibold truncate">{{ plugin.name }}</p>
              <p class="text-[10px] text-slate-500 truncate">v{{ plugin.version || '1.0.0' }} · 卡牌 {{ plugin.card_count ?? 0 }}</p>
            </div>
            <button @click.stop="togglePlugin(plugin)" class="p-1 hover:bg-white/10 rounded">
              <ToggleRight v-if="plugin.is_active" class="w-4 h-4 text-emerald-400" />
              <ToggleLeft v-else class="w-4 h-4 text-slate-400" />
            </button>
            <button @click.stop="deletePlugin(plugin)" class="p-1 hover:bg-red-500/20 rounded">
              <Trash2 class="w-4 h-4 text-red-400" />
            </button>
          </div>
        </div>
      </div>

      <div class="flex-1 overflow-y-auto p-4 space-y-4">
        <div v-if="!selectedPlugin" class="h-full flex items-center justify-center text-slate-500 text-sm">← 选择插件配置</div>

        <template v-else>
          <section class="rounded-2xl border border-white/10 bg-white/[0.03] p-4">
            <h2 class="text-lg font-black">{{ selectedPlugin.name }}</h2>
            <p class="text-xs text-slate-400 mt-1">{{ selectedPlugin.description || '暂无描述' }}</p>
            <div class="mt-2 text-[10px] flex gap-2 flex-wrap">
              <span class="px-2 py-0.5 rounded-full bg-white/10 border border-white/10">卡牌 {{ selectedPlugin.card_count ?? 0 }}</span>
              <span class="px-2 py-0.5 rounded-full border" :class="selectedPlugin.has_client_script ? 'bg-emerald-500/20 border-emerald-500/30' : 'bg-slate-500/20 border-slate-500/30'">Client {{ selectedPlugin.has_client_script ? 'ON' : 'OFF' }}</span>
              <span class="px-2 py-0.5 rounded-full border" :class="selectedPlugin.has_server_script ? 'bg-blue-500/20 border-blue-500/30' : 'bg-slate-500/20 border-slate-500/30'">Server {{ selectedPlugin.has_server_script ? 'ON' : 'OFF' }}</span>
            </div>
          </section>

          <section class="rounded-2xl border border-white/10 bg-white/[0.03] p-4">
            <div class="flex items-center justify-between mb-3">
              <h3 class="text-sm font-black">统一配置（Schema 驱动）</h3>
              <button @click="saveSettings" :disabled="saving" class="px-2.5 py-1 rounded bg-amber-600 hover:bg-amber-500 disabled:opacity-60 text-[11px] font-bold"><Save class="w-3 h-3 inline mr-1" />{{ saving ? '保存中' : '保存配置' }}</button>
            </div>
            <div v-if="settingsLoading" class="text-xs text-slate-500 py-6 text-center">正在加载配置...</div>
            <div v-else class="space-y-3">
              <div v-for="field in schemaFields" :key="field.key" class="rounded-xl border border-white/10 bg-white/[0.02] p-3">
                <p class="text-xs font-bold">{{ fieldLabel(field) }}</p>
                <p class="text-[10px] text-slate-500">{{ field.key }} · {{ fieldType(field) }}</p>
                <p v-if="field.description" class="text-[11px] text-slate-400 mt-1">{{ field.description }}</p>
                <div class="mt-2">
                  <input v-if="fieldType(field) === 'text'" v-model="settingValues[field.key]" :disabled="field.read_only" type="text" class="w-full bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm" />
                  <textarea v-else-if="fieldType(field) === 'textarea' || fieldType(field) === 'json'" v-model="settingValues[field.key]" :disabled="field.read_only" rows="4" class="w-full bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm font-mono resize-none" />
                  <input v-else-if="fieldType(field) === 'number'" v-model="settingValues[field.key]" :disabled="field.read_only" type="number" class="w-full bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm" />
                  <label v-else-if="fieldType(field) === 'switch'" class="inline-flex items-center gap-2 text-sm">
                    <input type="checkbox" :checked="settingValues[field.key] === 'true'" :disabled="field.read_only" @change="settingValues[field.key] = ($event.target as HTMLInputElement).checked ? 'true' : 'false'" />
                    {{ settingValues[field.key] === 'true' ? '开启' : '关闭' }}
                  </label>
                  <div v-else-if="fieldType(field) === 'image' || fieldType(field) === 'file'" class="space-y-2">
                    <input type="file" :disabled="field.read_only" :accept="field.accept || (fieldType(field) === 'image' ? 'image/*' : '*/*')" class="w-full text-xs text-slate-300" @change="onUploadFile(field, $event)" />
                    <img v-if="fieldType(field) === 'image' && settingValues[field.key]" :src="settingValues[field.key]" class="w-32 rounded-xl border border-white/10 bg-white p-2" />
                  </div>
                  <div v-else-if="fieldType(field) === 'route_list'" class="space-y-2">
                    <button @click="addRoutePage(field.key)" class="px-2 py-1 rounded bg-blue-600 hover:bg-blue-500 text-[10px] font-bold">新增页面</button>
                    <div v-for="(page, idx) in routePagesByKey[field.key] || []" :key="`${field.key}-${idx}`" class="rounded-lg border border-white/10 p-2 space-y-2">
                      <div class="grid grid-cols-2 gap-2">
                        <input v-model="page.path" placeholder="/plugin/page" class="bg-white/5 border border-white/10 rounded-lg px-2 py-1.5 text-xs font-mono" />
                        <input v-model="page.title" placeholder="标题" class="bg-white/5 border border-white/10 rounded-lg px-2 py-1.5 text-xs" />
                      </div>
                      <input v-model="page.description" placeholder="描述" class="w-full bg-white/5 border border-white/10 rounded-lg px-2 py-1.5 text-xs" />
                      <textarea v-model="page.content_html" rows="3" placeholder="HTML 内容" class="w-full bg-white/5 border border-white/10 rounded-lg px-2 py-1.5 text-xs font-mono resize-none" />
                      <div class="flex gap-2 text-[10px]">
                        <label><input v-model="page.requires_auth" type="checkbox" />登录</label>
                        <label><input v-model="page.admin_only" type="checkbox" />管理员</label>
                        <label><input v-model="page.co_worker_only" type="checkbox" />协作者</label>
                      </div>
                      <button @click="removeRoutePage(field.key, idx)" class="px-2 py-1 rounded bg-red-500/15 hover:bg-red-500/25 text-[10px] text-red-400 font-bold">移除</button>
                    </div>
                  </div>
                  <input v-else v-model="settingValues[field.key]" :disabled="field.read_only" type="text" class="w-full bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm" />
                </div>
              </div>
              <div v-if="!schemaFields.length" class="text-xs text-slate-500">该插件未提供 config_schema，建议插件作者补充 schema。</div>
            </div>
          </section>

          <section class="rounded-2xl border border-white/10 bg-white/[0.03] p-4">
            <h3 class="text-sm font-black mb-3">路由可视化</h3>
            <div class="mb-3">
              <p class="text-[11px] text-slate-400 mb-1">配置路由（route_list）</p>
              <div v-if="Object.values(routePagesByKey).flat().length === 0" class="text-xs text-slate-500">暂无配置路由</div>
              <div
                v-for="(page, idx) in Object.values(routePagesByKey).flat()"
                :key="`configured-${idx}-${page.path}`"
                class="rounded-lg border border-blue-500/20 bg-blue-500/10 p-2 mb-2"
              >
                <p class="text-xs font-mono text-blue-300">{{ page.path || '(未填写路径)' }}</p>
                <p class="text-[11px] text-slate-300">{{ page.title || '未命名页面' }}</p>
              </div>
            </div>
            <div v-if="runtimeRoutes.length === 0" class="text-xs text-slate-500">暂无已注册路由，保存 route_list 后会自动刷新</div>
            <div v-for="route in runtimeRoutes" :key="route.name" class="rounded-lg border border-emerald-500/20 bg-emerald-500/10 p-2 mb-2">
              <p class="text-xs font-mono text-emerald-300">{{ route.path }}</p>
              <p class="text-[11px] text-slate-300">{{ route.name }} · {{ route.source }}</p>
            </div>
          </section>

          <section class="rounded-2xl border border-white/10 bg-white/[0.03] p-4">
            <h3 class="text-sm font-black mb-3">配置历史与回滚</h3>
            <div class="flex items-center gap-2">
              <select v-model="selectedSnapshotId" class="flex-1 bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-xs">
                <option value="">请选择历史快照</option>
                <option v-for="snap in settingsHistory" :key="snap.id" :value="snap.id">
                  {{ new Date(snap.created_at).toLocaleString('zh-CN') }} · by {{ snap.created_by }} · {{ Object.keys(snap.settings || {}).length }} 项
                </option>
              </select>
              <button @click="rollbackSettings" class="px-3 py-2 bg-red-600 hover:bg-red-500 rounded-lg text-xs font-bold">回滚</button>
            </div>
            <p class="text-[11px] text-slate-500 mt-2">保存后会自动生成快照，最多保留 20 条历史。</p>
          </section>
        </template>
      </div>
    </div>

    <div v-if="showRestartModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div class="bg-slate-900 border border-red-500/20 rounded-2xl w-full max-w-md">
        <div class="p-4 border-b border-white/5 text-sm font-black text-red-300 uppercase tracking-widest">安排服务器重启</div>
        <div class="p-4 space-y-3">
          <input v-model.number="restartDelay" type="number" min="10" max="300" class="w-full bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm" />
          <input v-model="restartReason" type="text" placeholder="重启原因（可选）" class="w-full bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm" />
        </div>
        <div class="flex gap-2 p-4 border-t border-white/5">
          <button @click="showRestartModal = false" class="flex-1 py-2 bg-white/5 hover:bg-white/10 rounded-lg text-sm font-bold">取消</button>
          <button @click="scheduleRestart" class="flex-1 py-2 bg-red-600 hover:bg-red-500 rounded-lg text-sm font-bold">确认</button>
        </div>
      </div>
    </div>
  </div>
</template>
