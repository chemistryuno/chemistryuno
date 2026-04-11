<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { gameAPI } from '../../utils/api'
import { History, ChevronRight, Activity, Trophy, AlertTriangle, Eye } from 'lucide-vue-next'

const history = ref<any[]>([])
const loading = ref(true)
const replayVisible = ref(false)
const replayLoading = ref(false)
const replayError = ref('')
const replayData = ref<any>(null)
const activeHistory = ref<any>(null)

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

const formatEventType = (type: string) => {
  const labels: Record<string, string> = {
    game_start: '对局开始',
    play_card: '出牌',
    double_play: '双联',
    draw_card: '摸牌',
    timeout_auto_draw: '超时自动摸牌',
    game_finished: '对局结束',
    game_terminated_invalid: '无效结算',
    fast_reaction: '快速反应',
  }
  return labels[type] || type
}

const replayEvents = computed(() => {
  const replay = replayData.value?.replay
  if (!replay || !Array.isArray(replay.events)) {
    return []
  }
  return replay.events
})

const describeEvent = (evt: any) => {
  const actorUID = evt?.actor_uid ?? evt?.uid
  const actor = actorUID ? `UID ${actorUID}` : '系统'
  const eventType = formatEventType(evt?.type || evt?.event || '')
  const payload = evt?.payload || {}

  if (eventType === '出牌') {
    const symbol = payload.card_symbol || payload.card_type || '未知卡'
    const substance = payload.substance || '未知物质'
    const speed = payload.fast_reaction_ms ? ` (${payload.fast_reaction_ms}ms)` : ''
    return `${actor} 使用 ${symbol} -> ${substance}${speed}`
  }

  if (eventType === '双联') {
    const sub1 = payload.sub1 || payload.substance_1 || '?'
    const sub2 = payload.sub2 || payload.substance_2 || '?'
    const speed = payload.fast_reaction_ms ? ` (${payload.fast_reaction_ms}ms)` : ''
    return `${actor} 触发双联 ${sub1} + ${sub2}${speed}`
  }

  if (eventType === '摸牌') {
    const count = payload.actual_count || payload.draw_count || 1
    return `${actor} 摸牌 ${count} 张`
  }

  if (eventType === '超时自动摸牌') {
    const count = payload.draw_count || 1
    return `${actor} 超时，系统自动摸牌 ${count} 张`
  }

  return `${actor} ${eventType}`
}

const openReplay = async (game: any) => {
  replayVisible.value = true
  replayLoading.value = true
  replayError.value = ''
  replayData.value = null
  activeHistory.value = game

  try {
    const response = await gameAPI.getMyGameReplay(game.id)
    replayData.value = response.data
  } catch (error: any) {
    replayError.value = error?.response?.data?.error || '回放加载失败'
  } finally {
    replayLoading.value = false
  }
}

const closeReplay = () => {
  replayVisible.value = false
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
            <span v-if="game.cheat_detected" class="px-2 py-1 rounded-lg bg-rose-500/10 text-rose-500 text-[10px] font-black tracking-wider">CHEAT</span>
            <span v-if="game.replay_permanent" class="px-2 py-1 rounded-lg bg-amber-500/10 text-amber-600 dark:text-amber-400 text-[10px] font-black tracking-wider">永久保留</span>
            <span v-else-if="game.replay_expires_at" class="px-2 py-1 rounded-lg bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 text-[10px] font-black tracking-wider">7天回放</span>
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

    <div v-if="replayVisible" class="fixed inset-0 z-[110] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="closeReplay"></div>

      <div class="relative w-full max-w-3xl max-h-[85vh] overflow-hidden rounded-2xl border border-white/10 bg-white dark:bg-slate-950 shadow-2xl">
        <div class="flex items-center justify-between px-5 py-4 border-b border-slate-200 dark:border-white/10">
          <div>
            <p class="text-xs font-black tracking-widest text-slate-400">Replay / 回放</p>
            <h4 class="text-sm font-black text-slate-900 dark:text-white">#{{ String(activeHistory?.id || '').padStart(4, '0') }} 对局回放</h4>
          </div>
          <button @click="closeReplay" class="px-3 py-1.5 rounded-lg text-xs font-black text-slate-500 hover:bg-slate-100 dark:hover:bg-white/10">关闭</button>
        </div>

        <div class="p-5 overflow-y-auto max-h-[calc(85vh-72px)] space-y-4">
          <div v-if="replayLoading" class="py-8 flex justify-center">
            <div class="w-7 h-7 border-4 border-cyan-500/20 border-t-cyan-500 rounded-full animate-spin"></div>
          </div>

          <div v-else-if="replayError" class="rounded-xl border border-rose-200 bg-rose-50 dark:bg-rose-500/10 px-4 py-3 text-sm text-rose-600 dark:text-rose-300">
            {{ replayError }}
          </div>

          <template v-else>
            <div class="flex items-center gap-2 flex-wrap">
              <span v-if="replayData?.cheat_detected" class="px-2 py-1 rounded-lg bg-rose-500/10 text-rose-500 text-[11px] font-black">CHEAT 检测</span>
              <span v-if="replayData?.replay_permanent" class="px-2 py-1 rounded-lg bg-amber-500/10 text-amber-600 dark:text-amber-400 text-[11px] font-black">永久回放</span>
              <span v-else-if="replayData?.replay_expires_at" class="px-2 py-1 rounded-lg bg-cyan-500/10 text-cyan-600 dark:text-cyan-300 text-[11px] font-black">到期 {{ formatDate(replayData.replay_expires_at) }}</span>
            </div>

            <div v-if="replayData?.cheat_uids?.length" class="rounded-xl border border-rose-200/70 dark:border-rose-500/30 bg-rose-50 dark:bg-rose-500/10 p-3">
              <p class="text-xs font-black text-rose-500 mb-2">可疑 UID</p>
              <div class="flex gap-2 flex-wrap">
                <span v-for="uid in replayData.cheat_uids" :key="uid" class="px-2 py-1 rounded-md text-xs font-black bg-rose-500/15 text-rose-600 dark:text-rose-300">
                  UID {{ uid }}
                </span>
              </div>
            </div>

            <div class="rounded-xl border border-slate-200 dark:border-white/10 p-3">
              <p class="text-xs font-black text-slate-500 mb-2">参与者</p>
              <div class="flex flex-wrap gap-2">
                <span
                  v-for="profile in (replayData?.player_profiles || [])"
                  :key="profile.uid"
                  class="px-2 py-1 rounded-md text-xs font-semibold bg-slate-100 dark:bg-white/10 text-slate-700 dark:text-slate-200"
                >
                  {{ profile.nickname }} (UID {{ profile.uid }})
                </span>
              </div>
            </div>

            <div class="rounded-xl border border-slate-200 dark:border-white/10 p-3">
              <p class="text-xs font-black text-slate-500 mb-2">事件时间线</p>
              <div v-if="replayEvents.length === 0" class="text-xs text-slate-400">暂无事件记录</div>
              <div v-else class="space-y-2">
                <div
                  v-for="(event, idx) in replayEvents"
                  :key="`${event.at || event.timestamp || 't'}-${idx}`"
                  class="rounded-lg border border-slate-100 dark:border-white/10 px-3 py-2 bg-slate-50 dark:bg-white/5"
                >
                  <div class="flex items-center justify-between gap-3">
                    <p class="text-xs font-semibold text-slate-700 dark:text-slate-200">{{ idx + 1 }}. {{ describeEvent(event) }}</p>
                    <p class="text-[10px] font-mono text-slate-400">{{ formatDate(event.at || event.timestamp) }}</p>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>
