<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { gameAPI } from '../../utils/api'
import { History, ChevronRight, Activity, Trophy, AlertTriangle, Eye } from 'lucide-vue-next'

const history = ref<any[]>([])
const loading = ref(true)
const router = useRouter()
const route = useRoute()

const fetchHistory = async () => {
  try {
    const response = await gameAPI.getMyGameHistory()
    history.value = response.data
  } catch (error) {
    console.error('获取对局记录失败:', error)
  } finally {
    loading.value = false
  }
}

onMounted(fetchHistory)

const formatDate = (dateStr: string) => {
  if (!dateStr) return '未知时间'
  try {
    // 兼容 Go 的时间格式
    return new Date(dateStr.replace(' ', 'T')).toLocaleString()
  } catch (e) {
    return dateStr
  }
}

const openReplay = (game: any) => {
  const query = new URLSearchParams()
  query.set('from', route.path.startsWith('/profile') ? route.path : '/profile/history')
  router.push(`/replay/${game.id}?${query.toString()}`)
}
</script>

<template>
  <div class="space-y-5">
    <div class="flex items-center justify-between">
      <h3 class="text-base font-black italic uppercase text-slate-800 dark:text-white flex items-center gap-4">
        <History class="w-5 h-5 text-cyan-500" />
        对局历史记录 <span class="text-slate-400 dark:text-slate-600 font-mono not-italic text-[10px] tracking-normal">/ MATCH@HISTORY</span>
      </h3>
    </div>

    <div v-if="loading" class="py-10 flex flex-col items-center justify-center gap-4">
      <div class="w-8 h-8 border-4 border-cyan-500/10 border-t-cyan-500 rounded-full animate-spin"></div>
      <p class="text-[10px] font-black uppercase tracking-widest text-slate-400 animate-pulse">Retrieving combat logs...</p>
    </div>

    <div v-else-if="history.length === 0" class="py-10 flex flex-col items-center justify-center border-2 border-dashed border-slate-200 dark:border-white/5 rounded-2xl bg-slate-50 dark:bg-white/[0.02]">
      <Activity class="w-8 h-8 text-slate-300 dark:text-slate-700 mb-4 opacity-40" />
      <p class="text-slate-400 font-black italic uppercase tracking-widest text-xs">/ NO_GAME_DATA_FOUND</p>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div
        v-for="game in history"
        :key="game.id"
        class="group relative overflow-hidden p-4 bg-white dark:bg-black/20 border border-slate-200 dark:border-white/10 rounded-2xl hover:border-cyan-500/50 transition-all hover:shadow-xl hover:shadow-cyan-500/5"
      >
        <div class="absolute top-0 right-0 w-32 h-32 bg-cyan-500/[0.03] blur-3xl -mr-16 -mt-16 group-hover:bg-cyan-500/[0.08] transition-all" />

        <div class="flex items-center justify-between relative z-10 mb-3">
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 rounded-xl bg-cyan-500/10 flex items-center justify-center text-cyan-600 dark:text-cyan-400">
              <span class="font-mono text-[10px] font-black">#{{ String(game.id).padStart(4, '0') }}</span>
            </div>
            <div>
              <p class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">Protocol ID</p>
              <p class="text-[10px] font-black text-slate-900 dark:text-white truncate max-w-[120px]">{{ game.room_id }}</p>
            </div>
          </div>
          <div class="text-right">
            <p class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">Timestamp</p>
            <p class="text-[10px] font-mono font-bold text-slate-500 dark:text-slate-400">{{ formatDate(game.finished_at) }}</p>
          </div>
        </div>

        <div class="bg-slate-50 dark:bg-white/5 rounded-xl p-3 relative z-10 border border-slate-100 dark:border-white/5">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <Trophy class="w-4 h-4 text-amber-500" v-if="!game.is_invalid && game.winner_name !== '未结算'" />
              <AlertTriangle class="w-4 h-4 text-rose-500" v-else-if="game.is_invalid" />
              <div>
                <p class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">Winner / 优胜者</p>
                <p :class="[
                  'text-xs font-black italic uppercase tracking-tighter',
                  game.is_invalid ? 'text-rose-500' : (game.winner_name === '未结算' ? 'text-slate-400' : 'text-amber-500')
                ]">
                  {{ game.is_invalid ? '无效对局' : game.winner_name }}
                </p>
                <p v-if="game.is_invalid && game.invalid_reason" class="text-[9px] font-bold text-rose-400 mt-1 max-w-[220px] leading-relaxed">
                  {{ game.invalid_reason }}
                </p>
              </div>
            </div>
            <div class="text-right">
              <p class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">Players / 规模</p>
              <p class="text-xs font-black text-slate-900 dark:text-white">{{ game.players?.length || 0 }} 研究员</p>
            </div>
          </div>
        </div>

        <div class="mt-3 relative z-10 flex items-center justify-between gap-2 flex-wrap">
          <div class="flex items-center gap-2 flex-wrap">
            <span v-if="game.replay_expires_at" class="px-2 py-1 rounded-lg bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 text-[10px] font-black tracking-wider">7天回放</span>
          </div>

          <button
            v-if="game.has_replay"
            @click="openReplay(game)"
            class="inline-flex items-center gap-1 px-3 py-1.5 rounded-xl bg-cyan-500/10 text-cyan-600 dark:text-cyan-300 text-[11px] font-black hover:bg-cyan-500/20 transition-colors"
          >
            <Eye class="w-3.5 h-3.5" />
            查看回放
            <ChevronRight class="w-3.5 h-3.5" />
          </button>
          <span v-else class="text-[10px] text-slate-400 font-black">回放不可用</span>
        </div>
      </div>
    </div>
  </div>
</template>
