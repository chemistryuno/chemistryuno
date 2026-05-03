<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { AlertTriangle, Info, RotateCw, ShieldCheck, TimerReset } from 'lucide-vue-next'
import { authAPI } from '../../utils/api'

type AnticheatStats = {
  bans_today: number
  system_uptime_days: number
}

const REFRESH_INTERVAL_MS = 5 * 60 * 1000

const stats = ref<AnticheatStats | null>(null)
const cachedStats = ref<AnticheatStats | null>(null)
const loading = ref(true)
const refreshing = ref(false)
const error = ref('')
const usingCached = ref(false)
const lastUpdated = ref<Date | null>(null)
let refreshTimer: number | undefined

const displayStats = computed(() => stats.value || cachedStats.value)
const hasStats = computed(() => Boolean(displayStats.value))

const normalizeStat = (value: unknown) => {
  const parsed = Number(value)
  return Number.isFinite(parsed) && parsed >= 0 ? Math.floor(parsed) : 0
}

const fetchStats = async () => {
  const isInitialLoad = !stats.value && !cachedStats.value
  loading.value = isInitialLoad
  refreshing.value = !isInitialLoad
  error.value = ''

  try {
    const response = await authAPI.getPlayerAnticheatStats()
    const nextStats = {
      bans_today: normalizeStat(response.data?.bans_today),
      system_uptime_days: normalizeStat(response.data?.system_uptime_days),
    }

    stats.value = nextStats
    cachedStats.value = nextStats
    usingCached.value = false
    lastUpdated.value = new Date()
  } catch (err: any) {
    if (err?.response?.status === 429 && cachedStats.value) {
      stats.value = cachedStats.value
      usingCached.value = true
      return
    }

    if (cachedStats.value) {
      stats.value = cachedStats.value
      usingCached.value = true
    } else {
      error.value = err?.response?.data?.error || '统计暂不可用'
    }
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

onMounted(() => {
  fetchStats()
  refreshTimer = window.setInterval(fetchStats, REFRESH_INTERVAL_MS)
})

onUnmounted(() => {
  if (refreshTimer) {
    window.clearInterval(refreshTimer)
  }
})
</script>

<template>
  <section class="bg-white dark:bg-white/5 border border-slate-200 dark:border-white/5 rounded-xl p-4 shadow-sm dark:shadow-none">
    <div class="flex items-center justify-between gap-3 mb-3">
      <div class="flex items-center gap-2 min-w-0">
        <div class="w-8 h-8 rounded-xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 flex items-center justify-center shrink-0">
          <ShieldCheck class="w-4 h-4" />
        </div>
        <div class="min-w-0">
          <h3 class="text-[10px] font-black uppercase tracking-widest text-slate-500 dark:text-slate-400 truncate">Anticheat</h3>
          <p class="text-[9px] font-mono uppercase text-slate-400 truncate">Transparency</p>
        </div>
      </div>
      <RotateCw
        class="w-3.5 h-3.5 text-slate-300 dark:text-slate-600 shrink-0"
        :class="{ 'animate-spin text-blue-500 dark:text-blue-400': refreshing }"
      />
    </div>

    <div v-if="loading" class="grid grid-cols-2 gap-2">
      <div class="h-16 rounded-xl bg-slate-100 dark:bg-white/5 animate-pulse" />
      <div class="h-16 rounded-xl bg-slate-100 dark:bg-white/5 animate-pulse" />
    </div>

    <div v-else-if="hasStats" class="grid grid-cols-2 gap-2">
      <div class="rounded-xl border border-slate-100 dark:border-white/5 bg-slate-50 dark:bg-black/10 p-3">
        <div class="flex items-center justify-between gap-2 mb-1">
          <span class="text-[8px] font-black uppercase tracking-widest text-slate-400">Bans Today</span>
          <Info class="w-3 h-3 text-slate-300 shrink-0" title="Total bans issued in the last 24 hours. Refreshed every 5 minutes." />
        </div>
        <div class="font-mono text-lg font-black leading-none text-slate-900 dark:text-white">
          {{ displayStats?.bans_today ?? '--' }}
        </div>
      </div>

      <div class="rounded-xl border border-slate-100 dark:border-white/5 bg-slate-50 dark:bg-black/10 p-3">
        <div class="flex items-center justify-between gap-2 mb-1">
          <span class="text-[8px] font-black uppercase tracking-widest text-slate-400">System Running</span>
          <TimerReset class="w-3 h-3 text-slate-300 shrink-0" title="Days since the anticheat service last reset." />
        </div>
        <div class="font-mono text-lg font-black leading-none text-slate-900 dark:text-white">
          {{ displayStats?.system_uptime_days ?? '--' }}<span class="ml-1 text-[9px] text-slate-400">days</span>
        </div>
      </div>
    </div>

    <div v-else class="flex items-center gap-2 rounded-xl border border-amber-500/20 bg-amber-500/10 px-3 py-3 text-amber-600 dark:text-amber-400">
      <AlertTriangle class="w-4 h-4 shrink-0" />
      <span class="text-[10px] font-bold uppercase tracking-widest">{{ error }}</span>
    </div>

    <p v-if="usingCached && hasStats" class="mt-2 text-[9px] font-mono uppercase tracking-widest text-slate-400">
      Cached{{ lastUpdated ? ` / ${lastUpdated.toLocaleTimeString()}` : '' }}
    </p>
  </section>
</template>
