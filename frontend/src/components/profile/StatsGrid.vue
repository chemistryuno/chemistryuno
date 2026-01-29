<script setup lang="ts">
import { Zap, Activity } from 'lucide-vue-next'

defineProps<{
  stats?: {
    totalGames: number | string
    winRate: number | string
  }
}>()
</script>

<template>
  <div class="grid grid-cols-2 gap-4">
    <div class="bg-white dark:bg-white/5 border border-slate-200 dark:border-white/5 rounded-3xl p-5 hover:bg-slate-50 dark:hover:bg-white/[0.08] transition-colors shadow-sm dark:shadow-none">
      <div class="flex items-center gap-3 mb-2">
        <Zap class="w-4 h-4 text-yellow-500 dark:text-yellow-400" />
        <span class="text-[10px] font-black uppercase text-slate-500">总场次</span>
      </div>
      <div class="text-2xl font-black text-slate-900 dark:text-white">{{ stats?.totalGames || '--' }}</div>
    </div>
    <div class="bg-white dark:bg-white/5 border border-slate-200 dark:border-white/5 rounded-3xl p-5 hover:bg-slate-50 dark:hover:bg-white/[0.08] transition-colors shadow-sm dark:shadow-none">
      <div class="flex items-center gap-3 mb-2">
        <Activity class="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
        <span class="text-[10px] font-black uppercase text-slate-500">胜率</span>
      </div>
      <div class="text-2xl font-black text-slate-900 dark:text-white">{{ stats?.winRate ? stats.winRate + '%' : '--%' }}</div>
    </div>
  </div>
</template>
