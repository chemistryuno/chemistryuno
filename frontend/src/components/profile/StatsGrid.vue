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
  <div class="grid grid-cols-2 gap-2">
    <div class="bg-white dark:bg-white/5 border border-slate-200 dark:border-white/5 rounded-xl p-3 hover:bg-slate-50 dark:hover:bg-white/[0.08] transition-all shadow-sm dark:shadow-none group">
      <div class="flex items-center gap-1.5 mb-1">
        <Zap class="w-3 h-3 text-yellow-500 dark:text-yellow-400 group-hover:scale-110 transition-transform" />
        <span class="text-[8px] font-black uppercase text-slate-400 tracking-widest">Plays</span>
      </div>
      <div class="text-base font-black text-slate-900 dark:text-white font-mono leading-none">{{ stats?.totalGames || '--' }}</div>
    </div>
    <div class="bg-white dark:bg-white/5 border border-slate-200 dark:border-white/5 rounded-xl p-3 hover:bg-slate-50 dark:hover:bg-white/[0.08] transition-all shadow-sm dark:shadow-none group">
      <div class="flex items-center gap-1.5 mb-1">
        <Activity class="w-3 h-3 text-emerald-600 dark:text-emerald-400 group-hover:scale-110 transition-transform" />
        <span class="text-[8px] font-black uppercase text-slate-400 tracking-widest">Rate</span>
      </div>
      <div class="text-base font-black text-slate-900 dark:text-white font-mono leading-none">{{ stats?.winRate ? stats.winRate + '%' : '--%' }}</div>
    </div>
  </div>
</template>
