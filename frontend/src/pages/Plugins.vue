<template>
  <div :class="pageClassNames.appPadded">
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
          class="group flex items-center gap-3 text-slate-500 dark:text-slate-400 hover:text-blue-600 dark:hover:text-blue-400 transition-all px-4 py-2.5 rounded-2xl bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 shadow-sm active:scale-95"
        >
          <ArrowLeft class="w-4 h-4 group-hover:-translate-x-1 transition-transform" />
          <span class="font-black tracking-widest uppercase text-[10px]">返回大厅</span>
        </button>

        <button
          @click="loadPlugins"
          class="flex items-center gap-2 px-5 py-2.5 rounded-2xl bg-blue-600 hover:bg-blue-500 dark:bg-blue-500 dark:hover:bg-blue-400 text-white shadow-lg shadow-blue-500/20 transition-all text-[10px] font-black uppercase tracking-widest active:scale-95"
        >
          <RefreshCw class="w-3.5 h-3.5" :class="{ 'animate-spin': loading }" />
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
      <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <div
          v-for="plugin in plugins"
          :key="plugin.id"
          class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2.5rem] overflow-hidden shadow-sm hover:shadow-xl transition-all duration-500 group/card relative"
        >
          <!-- Plugin Active Glow -->
          <div v-if="plugin.is_active" class="absolute -right-12 -top-12 w-32 h-32 bg-emerald-500/10 rounded-full blur-3xl group-hover/card:bg-emerald-500/20 transition-all"></div>

          <!-- Header -->
          <div class="p-6 border-b border-slate-100 dark:border-white/5 flex items-start gap-5 relative z-10">
            <div class="w-14 h-14 rounded-2xl bg-gradient-to-br from-purple-500/20 to-indigo-500/20 border border-purple-500/20 flex items-center justify-center shrink-0 shadow-inner group-hover/card:scale-110 transition-transform duration-500">
              <Puzzle class="w-7 h-7 text-purple-500" />
            </div>
            <div class="min-w-0 flex-1">
              <div class="flex items-center justify-between gap-2 flex-wrap mb-1.5">
                <div class="flex items-center gap-2 max-w-full">
                   <h2 class="text-xl font-black text-slate-900 dark:text-white truncate tracking-tight">{{ plugin.name }}</h2>
                   <span class="text-[9px] px-2 py-0.5 rounded-lg bg-slate-100 dark:bg-white/5 text-slate-500 dark:text-slate-400 border border-slate-200 dark:border-white/10 font-black uppercase tracking-widest">
                     v{{ plugin.version || '1.0.0' }}
                   </span>
                </div>
                <div
                  class="text-[9px] px-2.5 py-1 rounded-full border font-black uppercase tracking-[0.15em] flex items-center gap-1.5"
                  :class="plugin.is_active
                    ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20 shadow-[0_0_12px_rgba(16,185,129,0.1)]'
                    : 'bg-slate-100 dark:bg-white/5 text-slate-400 border-slate-200 dark:border-white/10'"
                >
                  <div v-if="plugin.is_active" class="w-1 h-1 rounded-full bg-emerald-500 animate-pulse"></div>
                  {{ plugin.is_active ? 'Active' : 'Standby' }}
                </div>
              </div>
              <p v-if="plugin.description" class="text-xs text-slate-500 dark:text-slate-400 leading-relaxed max-w-md">
                {{ plugin.description }}
              </p>
              <div class="flex items-center gap-4 mt-3 text-[9px] text-slate-400 dark:text-slate-500 font-mono uppercase tracking-widest">
                <div class="flex items-center gap-1.5">
                   <UserIcon class="w-3 h-3 opacity-50" />
                   <span>{{ plugin.author || 'Mendeleev' }}</span>
                </div>
                <div class="flex items-center gap-1.5">
                   <Layers class="w-3 h-3 opacity-50" />
                   <span>{{ plugin.cards?.length ?? 0 }} Assets</span>
                </div>
                <div class="flex items-center gap-1.5 hidden sm:flex">
                   <Clock class="w-3 h-3 opacity-50" />
                   <span>{{ formatDate(plugin.created_at) }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Cards Scrollable Area -->
          <div class="bg-slate-50/50 dark:bg-black/20 p-6">
            <div v-if="!plugin.cards || plugin.cards.length === 0" class="flex flex-col items-center justify-center py-10 text-slate-400">
               <Layers class="w-8 h-8 opacity-10 mb-2" />
               <p class="text-[10px] font-black uppercase tracking-widest italic">No Internal Card Assets</p>
            </div>
            <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
              <div
                v-for="card in plugin.cards"
                :key="card.id"
                class="group/card-item bg-white dark:bg-[#16161a] rounded-[1.5rem] border border-slate-200 dark:border-white/5 p-4 flex items-start gap-4 transition-all hover:bg-white dark:hover:bg-white/10 hover:shadow-lg hover:border-blue-500/30 dark:hover:border-blue-500/20"
              >
                <!-- Card Symbol Hexagon-like shape -->
                <div
                  class="w-12 h-12 rounded-2xl flex items-center justify-center text-xs font-black text-white shrink-0 shadow-lg relative overflow-hidden transition-transform group-hover/card-item:scale-105"
                  :style="card.color ? `background-color: ${card.color}` : 'background-color: #6366f1'"
                >
                  <div class="absolute inset-0 bg-white/10 opacity-0 group-hover/card-item:opacity-100 transition-opacity"></div>
                  <span class="relative z-10">{{ card.symbol.slice(0, 3).toUpperCase() }}</span>
                </div>

                <div class="min-w-0 flex-1">
                  <div class="flex items-center justify-between gap-2">
                    <span class="font-black text-slate-900 dark:text-white text-[13px] truncate tracking-tight">
                      {{ card.display_name || card.symbol }}
                    </span>
                    <span class="text-[8px] font-mono text-blue-500/60 font-black">×{{ card.default_count }}</span>
                  </div>
                  
                  <!-- Type Badge -->
                  <div class="inline-flex items-center px-1.5 py-0.5 rounded-md bg-slate-100 dark:bg-white/5 text-[9px] font-bold text-slate-500 mt-1 uppercase tracking-tighter">
                    {{ effectLabel(card.effect_type) }}
                  </div>

                  <!-- Config Summary -->
                  <div class="flex items-center gap-1.5 mt-2 overflow-hidden">
                    <div class="flex-1 h-px bg-slate-100 dark:bg-white/5"></div>
                    <span class="text-[8px] font-black font-mono text-slate-400 dark:text-slate-500 uppercase shrink-0">
                      {{ formatConfig(card.effect_config) }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
          
          <!-- Footer Actions -->
          <div class="px-6 py-4 bg-white dark:bg-[#111114] border-t border-slate-100 dark:border-white/5 flex items-center justify-between">
             <div class="flex items-center gap-1.5">
                <span class="w-1.5 h-1.5 rounded-full" :class="plugin.is_active ? 'bg-emerald-500' : 'bg-slate-300'"></span>
                <span class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Protocol OK</span>
             </div>
             <div class="text-[8px] font-mono text-slate-400 opacity-50">UID_{{ plugin.id }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { pageClassNames } from '@lib'
import { ArrowLeft, Puzzle, RefreshCw, Sparkles, Layers, User as UserIcon, Clock } from 'lucide-vue-next'
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
