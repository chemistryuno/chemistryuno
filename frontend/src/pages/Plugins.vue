<template>
  <div class="min-h-screen bg-slate-50 dark:bg-[#0a0a0c] text-slate-900 dark:text-white p-4 md:p-8 selection:bg-blue-500/30">
    <!-- Background Effects -->
    <div class="fixed inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-[-10%] right-[-10%] w-[45%] h-[45%] bg-purple-500/5 rounded-full blur-[140px]" />
      <div class="absolute bottom-[-10%] left-[-10%] w-[45%] h-[45%] bg-blue-500/5 rounded-full blur-[140px]" />
      <div class="absolute inset-0 bg-[url('/noise.svg')] opacity-20 brightness-50 contrast-150" />
    </div>

    <div class="max-w-6xl mx-auto relative z-10">
      <!-- Top Bar -->
      <div class="flex items-center justify-between mb-8">
        <button 
          @click="router.push('/')"
          class="group flex items-center gap-3 text-slate-400 hover:text-slate-900 dark:hover:text-white transition-all px-4 py-2 rounded-xl hover:bg-white dark:hover:bg-white/5 border border-transparent hover:border-slate-200 dark:hover:border-white/10"
        >
          <ArrowLeft class="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
          <span class="font-bold tracking-wider uppercase text-xs">返回大厅</span>
        </button>

        <button
          @click="loadPlugins"
          class="flex items-center gap-2 px-4 py-2 rounded-xl bg-blue-600/10 hover:bg-blue-600 text-blue-600 hover:text-white border border-blue-500/20 transition-all text-[10px] font-black uppercase tracking-widest"
        >
          <RefreshCw class="w-3.5 h-3.5" />
          刷新列表
        </button>
      </div>

      <!-- Title -->
      <div class="flex items-center justify-between mb-8 flex-wrap gap-4">
        <div class="flex items-center gap-4">
          <div class="p-3 bg-purple-500/10 rounded-2xl border border-purple-500/20">
            <Puzzle class="w-8 h-8 text-purple-500" />
          </div>
          <div>
            <h1 class="text-3xl font-black uppercase tracking-tighter">
              插件市场 / Plugins
            </h1>
            <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">
              已安装插件将为游戏注入全新的特殊卡牌能力
            </p>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <div class="bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl px-3 py-2 flex items-center gap-2 shadow-sm">
            <div class="w-7 h-7 bg-purple-500/10 border border-purple-500/20 rounded-lg flex items-center justify-center text-purple-500">
              <Puzzle class="w-3.5 h-3.5" />
            </div>
            <div class="flex flex-col">
              <span class="text-[7px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest leading-none mb-0.5">Plugins</span>
              <p class="text-[11px] font-black text-slate-700 dark:text-slate-200">{{ plugins.length }} 个</p>
            </div>
          </div>
          <div class="bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl px-3 py-2 flex items-center gap-2 shadow-sm">
            <div class="w-7 h-7 bg-emerald-500/10 border border-emerald-500/20 rounded-lg flex items-center justify-center text-emerald-500">
              <Sparkles class="w-3.5 h-3.5" />
            </div>
            <div class="flex flex-col">
              <span class="text-[7px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest leading-none mb-0.5">Active</span>
              <p class="text-[11px] font-black text-slate-700 dark:text-slate-200">{{ activeCount }} 个</p>
            </div>
          </div>
          <div class="bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl px-3 py-2 flex items-center gap-2 shadow-sm">
            <div class="w-7 h-7 bg-blue-500/10 border border-blue-500/20 rounded-lg flex items-center justify-center text-blue-500">
              <Layers class="w-3.5 h-3.5" />
            </div>
            <div class="flex flex-col">
              <span class="text-[7px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest leading-none mb-0.5">Cards</span>
              <p class="text-[11px] font-black text-slate-700 dark:text-slate-200">{{ totalCards }} 张</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex flex-col items-center justify-center py-20">
        <div class="w-10 h-10 border-4 border-purple-500/20 border-t-purple-500 rounded-full animate-spin mb-4"></div>
        <p class="text-slate-400 font-medium">正在加载插件数据...</p>
      </div>

      <!-- Error -->
      <div v-else-if="error" class="bg-white dark:bg-[#111114] border border-rose-500/20 rounded-[2rem] p-12 text-center shadow-sm">
        <div class="text-4xl mb-3">⚠️</div>
        <p class="text-rose-500">{{ error }}</p>
        <button @click="loadPlugins" class="mt-5 px-5 py-2.5 rounded-xl bg-rose-600 hover:bg-rose-500 text-white text-xs font-black uppercase tracking-widest transition-colors">
          重试加载
        </button>
      </div>

      <!-- Empty -->
      <div v-else-if="plugins.length === 0" class="bg-white dark:bg-[#111114] border-2 border-dashed border-slate-200 dark:border-white/10 rounded-[2.5rem] p-20 flex flex-col items-center justify-center text-center">
        <div class="w-16 h-16 rounded-2xl bg-purple-500/10 border border-purple-500/20 flex items-center justify-center mb-6">
          <Puzzle class="w-8 h-8 text-purple-500" />
        </div>
        <h3 class="text-xl font-bold text-slate-500">暂无已安装的插件</h3>
        <p class="text-slate-400 mt-2">管理员可在管理面板中安装 .cumod 插件文件</p>
      </div>

      <!-- Plugin List -->
      <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div
          v-for="plugin in plugins"
          :key="plugin.id"
          class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2rem] overflow-hidden shadow-sm"
        >
          <!-- Header -->
          <div class="p-5 border-b border-slate-100 dark:border-white/5 flex items-start gap-4">
            <div class="w-12 h-12 rounded-2xl bg-purple-500/10 border border-purple-500/20 flex items-center justify-center shrink-0">
              <Puzzle class="w-6 h-6 text-purple-500" />
            </div>
            <div class="min-w-0">
              <div class="flex items-center gap-2 flex-wrap">
                <h2 class="text-lg font-black text-slate-900 dark:text-white truncate">{{ plugin.name }}</h2>
                <span class="text-[10px] px-2 py-0.5 rounded-full bg-purple-500/10 text-purple-600 dark:text-purple-300 border border-purple-500/20 font-black uppercase tracking-widest">
                  v{{ plugin.version || '1.0.0' }}
                </span>
                <span
                  class="text-[10px] px-2 py-0.5 rounded-full border font-black uppercase tracking-widest"
                  :class="plugin.is_active
                    ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-300 border-emerald-500/20'
                    : 'bg-slate-200/60 dark:bg-white/5 text-slate-500 dark:text-slate-400 border-slate-200 dark:border-white/10'"
                >
                  {{ plugin.is_active ? '已激活' : '未激活' }}
                </span>
              </div>
              <p v-if="plugin.description" class="text-sm text-slate-500 dark:text-slate-400 mt-1 line-clamp-2">
                {{ plugin.description }}
              </p>
              <div class="flex items-center gap-3 mt-2 text-[10px] text-slate-400 dark:text-slate-500 font-mono uppercase tracking-widest">
                <span v-if="plugin.author">作者 {{ plugin.author }}</span>
                <span>卡牌 {{ plugin.cards?.length ?? 0 }}</span>
                <span>创建 {{ formatDate(plugin.created_at) }}</span>
              </div>
            </div>
          </div>

          <!-- Cards -->
          <div class="p-5">
            <div v-if="!plugin.cards || plugin.cards.length === 0" class="text-slate-500 text-sm text-center py-6">
              该插件暂无卡牌
            </div>
            <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div
                v-for="card in plugin.cards"
                :key="card.id"
                class="rounded-2xl border p-4 flex items-start gap-3 transition-all hover:-translate-y-0.5"
                :class="card.color ? '' : 'bg-slate-50/80 dark:bg-white/[0.03] border-slate-200 dark:border-white/10'"
                :style="card.color ? { backgroundColor: `${card.color}10`, borderColor: `${card.color}40` } : undefined"
              >
                <div
                  class="w-10 h-10 rounded-xl flex items-center justify-center text-[11px] font-black text-white flex-shrink-0 shadow-sm"
                  :style="card.color ? `background-color: ${card.color}` : 'background-color: #6366f1'"
                >
                  {{ card.symbol.slice(0, 3) }}
                </div>
                <div class="min-w-0">
                  <div class="font-black text-slate-900 dark:text-white text-sm truncate">
                    {{ card.display_name || card.symbol }}
                  </div>
                  <div class="text-[11px] text-slate-500 dark:text-slate-400 mt-0.5">
                    {{ effectLabel(card.effect_type) }}
                  </div>
                  <div class="text-[10px] text-slate-500 dark:text-slate-400 mt-1">
                    默认 {{ card.default_count }} 张 · 符号：{{ card.symbol }}
                  </div>
                  <div class="text-[10px] mt-1.5 font-mono text-slate-400 dark:text-slate-500 truncate" :title="card.effect_config">
                    {{ formatConfig(card.effect_config) }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowLeft, Puzzle, RefreshCw, Sparkles, Layers } from 'lucide-vue-next'
import { pluginAPI } from '../utils/api'

interface PluginCard {
  id: number
  plugin_id: number
  symbol: string
  display_name: string
  effect_type: string
  effect_config: string
  default_count: number
  color: string
  created_at: string
}

interface Plugin {
  id: number
  name: string
  description: string
  author: string
  version: string
  is_active: boolean
  created_at: string
  cards: PluginCard[]
}

const plugins = ref<Plugin[]>([])
const loading = ref(false)
const error = ref('')
const router = useRouter()

const activeCount = computed(() => plugins.value.filter((plugin) => plugin.is_active).length)
const totalCards = computed(() =>
  plugins.value.reduce((count, plugin) => count + (plugin.cards?.length ?? 0), 0)
)

async function loadPlugins() {
  loading.value = true
  error.value = ''
  try {
    const res = await pluginAPI.getPluginsWithCards()
    plugins.value = res.data ?? []
  } catch (e: any) {
    error.value = e?.response?.data?.error ?? '加载失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

function effectLabel(type: string): string {
  const map: Record<string, string> = {
    swap: '🔄 随机交换手牌',
    force_play: '⚡ 强制对手出牌',
    convert: '🔁 消耗换取新牌',
  }
  return map[type] ?? type
}

function formatConfig(raw: string): string {
  try {
    const obj = JSON.parse(raw)
    if (obj.count !== undefined) return `数量: ${obj.count}`
    if (obj.source_count !== undefined)
      return `消耗 ${obj.source_count} → 获得 ${obj.target_count}`
    return raw
  } catch {
    return raw
  }
}

function formatDate(iso: string): string {
  if (!iso) return ''
  return new Date(iso).toLocaleDateString('zh-CN')
}

onMounted(loadPlugins)
</script>
