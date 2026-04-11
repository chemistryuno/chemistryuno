<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { adminAPI, gameAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import { ArrowLeft, Ban, Eye, Shield, UserMinus } from 'lucide-vue-next'
import { cn } from '../utils/cn'

const route = useRoute()
const router = useRouter()
const { showAlert, showPrompt } = useDialog()

const replayData = ref<any>(null)
const loading = ref(true)
const loadError = ref('')
const selectedPerspectiveUID = ref<number | null>(null)
const focusPerspectiveOnly = ref(false)
const replayViewMode = ref<'game' | 'timeline'>('timeline')

const OPERATION_FAST_THRESHOLD_MS = 3000
const operationEventTypes = new Set(['play_card', 'double_play', 'draw_card'])

const currentUser = ref<any>({})
try {
  const raw = JSON.parse(localStorage.getItem('user') || '{}')
  if (raw.id && !raw.uid) {
    raw.uid = raw.id
  }
  currentUser.value = raw
} catch {
  currentUser.value = {}
}

const isAdmin = computed(() => !!currentUser.value?.is_admin)
const useAdminScope = computed(() => isAdmin.value && String(route.query.scope || '') === 'admin')
const replayReturnPath = computed(() => {
  const raw = String(route.query.from || '').trim()
  if (/^\/admin(?:\/[a-zA-Z0-9_-]+)?$/.test(raw)) {
    return raw
  }
  if (/^\/profile(?:\/[a-zA-Z0-9_-]+)?$/.test(raw)) {
    return raw
  }
  return useAdminScope.value ? '/admin/users' : '/profile/history'
})

const buildReplayTimelineURL = (targetHistoryID: number) => {
  const query = new URLSearchParams()
  if (useAdminScope.value) {
    query.set('scope', 'admin')
  }
  query.set('from', replayReturnPath.value)
  const queryStr = query.toString()
  return queryStr ? `/replay/${targetHistoryID}?${queryStr}` : `/replay/${targetHistoryID}`
}

const goBackToEntryPage = () => {
  router.push(replayReturnPath.value)
}

const isGameRoomReplayRoute = computed(() => {
  const queryId = Number(route.query.replay_history_id)
  return Number.isFinite(queryId) && queryId > 0 && typeof route.params.id === 'string'
})

const syncReplayViewModeByRoute = () => {
  if (isGameRoomReplayRoute.value) {
    replayViewMode.value = 'game'
    return
  }

  if (route.query.mode === 'game' || route.query.mode === 'timeline') {
    replayViewMode.value = route.query.mode
    return
  }

  replayViewMode.value = 'timeline'
}

syncReplayViewModeByRoute()
watch(() => route.fullPath, () => syncReplayViewModeByRoute())

const historyId = computed(() => {
  const fromParam = Number(route.params.historyId)
  if (Number.isFinite(fromParam) && fromParam > 0) {
    return fromParam
  }

  const fromQuery = Number(route.query.replay_history_id)
  if (Number.isFinite(fromQuery) && fromQuery > 0) {
    return fromQuery
  }

  return 0
})

const switchToGameView = () => {
  if (!historyId.value) return
  if (isGameRoomReplayRoute.value) {
    replayViewMode.value = 'game'
    return
  }

  const query = new URLSearchParams()
  query.set('replay_history_id', String(historyId.value))
  if (useAdminScope.value) {
    query.set('scope', 'admin')
  }
  query.set('from', replayReturnPath.value)
  router.push(`/room/replay?${query.toString()}`)
}

const switchToTimelineView = () => {
  if (!historyId.value) return
  if (!isGameRoomReplayRoute.value) {
    replayViewMode.value = 'timeline'
    return
  }

  router.push(buildReplayTimelineURL(historyId.value))
}

const participants = computed(() => {
  if (!Array.isArray(replayData.value?.player_profiles)) {
    return []
  }
  return replayData.value.player_profiles
})

const aiParticipantUIDSet = computed(() => {
  const set = new Set<number>()
  participants.value.forEach((profile: any) => {
    const uid = Number(profile?.uid)
    if (!Number.isFinite(uid)) return
    if (profile?.is_ai || uid < 0) {
      set.add(uid)
    }
  })
  return set
})

const currentPerspectiveName = computed(() => {
  const profile = participants.value.find((p: any) => Number(p.uid) === selectedPerspectiveUID.value)
  return profile?.nickname || profile?.username || '系统观察者'
})

const parseReplayTimeMs = (timeLike: any) => {
  if (!timeLike) return null
  const normalized = String(timeLike).includes('T') ? String(timeLike) : String(timeLike).replace(' ', 'T')
  const ms = new Date(normalized).getTime()
  return Number.isFinite(ms) ? ms : null
}

const events = computed(() => {
  const rawEvents = replayData.value?.replay?.events
  if (!Array.isArray(rawEvents)) {
    return []
  }

  const normalizedEvents = rawEvents.map((evt: any, index: number) => {
    const at = evt?.at || evt?.timestamp || ''
    return {
      ...evt,
      __index: index,
      eventType: evt?.type || evt?.event || '',
      actorUID: Number(evt?.actor_uid ?? evt?.uid ?? 0),
      at,
      atMs: parseReplayTimeMs(at),
      payload: evt?.payload || {}
    }
  })

  const sortedEvents = [...normalizedEvents].sort((a: any, b: any) => {
    if (a.atMs == null && b.atMs == null) return a.__index - b.__index
    if (a.atMs == null) return 1
    if (b.atMs == null) return -1
    if (a.atMs !== b.atMs) return a.atMs - b.atMs
    return a.__index - b.__index
  })

  const previousTimedByUID = new Map<number, number>()

  return sortedEvents.map((evt: any) => {
    const isAIActor = aiParticipantUIDSet.value.has(Number(evt.actorUID || 0))
    let operationMs: number | null = null
    let operationSource: 'payload' | 'delta' | null = null

    if (!isAIActor) {
      const payloadFastMs = Number(evt.payload?.fast_reaction_ms)
      if (Number.isFinite(payloadFastMs) && payloadFastMs > 0) {
        operationMs = payloadFastMs
        operationSource = 'payload'
      } else if (operationEventTypes.has(evt.eventType) && evt.atMs != null) {
        const previousMs = previousTimedByUID.get(Number(evt.actorUID || 0))
        if (previousMs != null) {
          const delta = evt.atMs - previousMs
          if (delta >= 0 && delta <= 10 * 60 * 1000) {
            operationMs = delta
            operationSource = 'delta'
          }
        }
      }
    }

    if (!isAIActor && evt.atMs != null && operationEventTypes.has(evt.eventType)) {
      previousTimedByUID.set(Number(evt.actorUID || 0), evt.atMs)
    }

    return {
      ...evt,
      operationMs,
      operationSource,
      isFastOperation: operationMs != null && operationMs < OPERATION_FAST_THRESHOLD_MS
    }
  })
})

const visibleEvents = computed(() => {
  if (!focusPerspectiveOnly.value || selectedPerspectiveUID.value == null) {
    return events.value
  }

  return events.value.filter((evt: any) => {
    if (evt.actorUID === selectedPerspectiveUID.value) {
      return true
    }
    return evt.eventType === 'game_start' || evt.eventType === 'game_finished' || evt.eventType === 'game_terminated_invalid'
  })
})

const operationStats = computed(() => {
  return participants.value
    .filter((profile: any) => {
      const uid = Number(profile.uid)
      return Number.isFinite(uid) && !aiParticipantUIDSet.value.has(uid)
    })
    .map((profile: any) => {
      const uid = Number(profile.uid)
      const samples = events.value.filter((evt: any) => evt.actorUID === uid && evt.operationMs != null)
      const opValues = samples.map((evt: any) => Number(evt.operationMs))
      const fastCount = samples.filter((evt: any) => evt.isFastOperation).length

      if (opValues.length === 0) {
        return {
          uid,
          nickname: profile.nickname || profile.username || `UID ${uid}`,
          avgMs: null,
          minMs: null,
          fastCount,
          totalCount: 0
        }
      }

      const sum = opValues.reduce((acc: number, cur: number) => acc + cur, 0)
      const min = Math.min(...opValues)

      return {
        uid,
        nickname: profile.nickname || profile.username || `UID ${uid}`,
        avgMs: Math.round(sum / opValues.length),
        minMs: Math.round(min),
        fastCount,
        totalCount: opValues.length
      }
    })
    .sort((a, b) => b.fastCount - a.fastCount)
})

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  try {
    return new Date(String(dateStr).replace(' ', 'T')).toLocaleString()
  } catch {
    return String(dateStr)
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
    fast_reaction: '快速反应'
  }
  return labels[type] || type
}

const formatOperationDuration = (ms: number | null) => {
  if (ms == null) return '-'
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

const resolveActorName = (uid: number) => {
  if (!uid) return '系统'
  if (selectedPerspectiveUID.value != null && uid === selectedPerspectiveUID.value) {
    return '你'
  }

  const profile = participants.value.find((p: any) => Number(p.uid) === Number(uid))
  if (profile) {
    return profile.nickname || profile.username || `UID ${uid}`
  }
  return `UID ${uid}`
}

const describeEvent = (evt: any) => {
  const actor = resolveActorName(Number(evt.actorUID || 0))
  const payload = evt.payload || {}
  const operationHint = evt.operationMs != null
    ? `（操作耗时 ${formatOperationDuration(evt.operationMs)}${evt.isFastOperation ? '，疑似过快' : ''}）`
    : ''

  if (evt.eventType === 'play_card') {
    const symbol = payload.card_symbol || payload.card_type || '未知卡'
    const substance = payload.substance || '未知物质'
    return `${actor} 使用 ${symbol} -> ${substance}${operationHint}`
  }

  if (evt.eventType === 'double_play') {
    const sub1 = payload.sub1 || payload.substance_1 || '?'
    const sub2 = payload.sub2 || payload.substance_2 || '?'
    return `${actor} 触发双联 ${sub1} + ${sub2}${operationHint}`
  }

  if (evt.eventType === 'draw_card') {
    return `${actor} 摸牌 ${payload.actual_count || payload.draw_count || 1} 张${operationHint}`
  }

  if (evt.eventType === 'timeout_auto_draw') {
    return `${actor} 超时自动摸牌 ${payload.draw_count || 1} 张`
  }

  if (evt.eventType === 'fast_reaction') {
    return `${actor} 触发快速反应 (${payload.interval_ms || '?'}ms)`
  }

  return `${actor} ${formatEventType(evt.eventType)}`
}

const chooseDefaultPerspective = () => {
  if (!participants.value.length) {
    selectedPerspectiveUID.value = null
    return
  }

  const preferred = participants.value.find((p: any) => Number(p.uid) === Number(currentUser.value?.uid))
  if (preferred) {
    selectedPerspectiveUID.value = Number(preferred.uid)
    return
  }

  const firstHuman = participants.value.find((p: any) => Number(p.uid) > 0)
  if (firstHuman) {
    selectedPerspectiveUID.value = Number(firstHuman.uid)
    return
  }

  selectedPerspectiveUID.value = Number(participants.value[0].uid)
}

const loadReplay = async () => {
  loading.value = true
  loadError.value = ''

  if (!historyId.value) {
    loadError.value = '无效的回放编号'
    loading.value = false
    return
  }

  try {
    let response
    if (useAdminScope.value) {
      response = await adminAPI.getGameReplay(historyId.value)
    } else {
      response = await gameAPI.getMyGameReplay(historyId.value)
    }

    replayData.value = response.data
    chooseDefaultPerspective()
  } catch (error: any) {
    if (!useAdminScope.value && isAdmin.value) {
      try {
        const fallback = await adminAPI.getGameReplay(historyId.value)
        replayData.value = fallback.data
        chooseDefaultPerspective()
        loading.value = false
        return
      } catch {
      }
    }

    loadError.value = error?.response?.data?.error || '回放加载失败'
  } finally {
    loading.value = false
  }
}

const enterLiveRoomAsSpectator = () => {
  const roomID = replayData.value?.room_id
  if (!roomID) return
  router.push(`/room/${roomID}?spectator=true`)
}

const banFromReplay = async (profile: any) => {
  if (!isAdmin.value || Number(profile.uid) <= 0) return

  const hourText = await showPrompt(`请输入封禁时长（小时），目标：${profile.nickname || profile.username || profile.uid}`, '24', '封禁设置')
  if (hourText === null) return

  const hours = Number(hourText)
  if (!Number.isFinite(hours) || hours <= 0) {
    await showAlert('封禁时长必须是大于0的数字', '参数错误')
    return
  }

  const reason = await showPrompt('请输入封禁原因', '回放审计判定违规', '封禁原因')
  if (reason === null) return

  const bannedUntil = new Date(Date.now() + hours * 60 * 60 * 1000).toISOString()
  try {
    await adminAPI.banUser(Number(profile.uid), bannedUntil, reason || '回放审计判定违规')
    await showAlert(`已封禁 ${profile.nickname || profile.username || profile.uid}`, '执行完成')
  } catch (error: any) {
    await showAlert(error?.response?.data?.error || '封禁失败', '错误')
  }
}

const kickFromReplay = async (profile: any) => {
  if (!isAdmin.value || Number(profile.uid) <= 0) return

  const reason = await showPrompt(`请输入踢出原因，目标：${profile.nickname || profile.username || profile.uid}`, '回放审计判定违规', '踢出原因')
  if (reason === null) return

  try {
    await adminAPI.kickPlayer(Number(profile.uid), reason || '回放审计判定违规')
    await showAlert(`已踢出 ${profile.nickname || profile.username || profile.uid}`, '执行完成')
  } catch (error: any) {
    await showAlert(error?.response?.data?.error || '踢出失败', '错误')
  }
}

onMounted(() => {
  loadReplay()
})
</script>

<template>
  <div class="h-screen w-full bg-slate-50 dark:bg-slate-900 text-slate-900 dark:text-white overflow-hidden flex flex-col font-sans selection:bg-blue-500/30">
    <header class="px-4 py-3 border-b border-slate-200 dark:border-white/10 bg-white/90 dark:bg-black/30 backdrop-blur-md flex items-center justify-between">
      <div class="flex items-center gap-3">
        <button @click="goBackToEntryPage" class="inline-flex items-center gap-2 px-3 py-2 rounded-xl border border-slate-200 dark:border-white/10 bg-white dark:bg-white/5 hover:bg-slate-100 dark:hover:bg-white/10 transition-all text-xs font-black uppercase tracking-widest">
          <ArrowLeft class="w-4 h-4" />
          返回
        </button>
        <div>
          <p class="text-[10px] font-black uppercase tracking-widest text-slate-400">Replay Room</p>
          <h1 class="text-sm md:text-base font-black uppercase tracking-wide">对局回放 #{{ String(historyId).padStart(4, '0') }}</h1>
        </div>
      </div>

      <div class="flex items-center gap-2">
        <div class="inline-flex items-center gap-1 p-1 rounded-xl border border-slate-200 dark:border-white/10 bg-white dark:bg-white/5">
          <button
            @click="switchToGameView"
            :class="cn('px-3 py-2 rounded-lg text-[10px] font-black uppercase tracking-widest transition-all', replayViewMode === 'game' ? 'bg-blue-600 text-white' : 'text-slate-500 hover:text-slate-900 dark:hover:text-white')"
          >
            游戏界面方式
          </button>
          <button
            @click="switchToTimelineView"
            :class="cn('px-3 py-2 rounded-lg text-[10px] font-black uppercase tracking-widest transition-all', replayViewMode === 'timeline' ? 'bg-cyan-600 text-white' : 'text-slate-500 hover:text-slate-900 dark:hover:text-white')"
          >
            当前回放方式
          </button>
        </div>

        <button
          v-if="replayData?.room_id"
          @click="enterLiveRoomAsSpectator"
          class="inline-flex items-center gap-2 px-3 py-2 rounded-xl border border-cyan-500/20 bg-cyan-500/10 text-cyan-600 dark:text-cyan-300 text-xs font-black uppercase tracking-widest hover:bg-cyan-500/20 transition-all"
        >
          <Eye class="w-4 h-4" />
          旁观当前房间
        </button>
      </div>
    </header>

    <div class="flex-1 min-h-0 grid grid-cols-1 lg:grid-cols-[300px_1fr] gap-0">
      <aside class="border-r border-slate-200 dark:border-white/10 bg-white dark:bg-black/20 p-4 overflow-y-auto space-y-4">
        <div class="rounded-xl border border-slate-200 dark:border-white/10 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-slate-400 mb-2">回放状态</p>
          <div class="flex flex-wrap gap-2">
            <span v-if="replayData?.cheat_detected" class="px-2 py-1 rounded-md text-[10px] font-black bg-rose-500/10 text-rose-500">CHEAT</span>
            <span v-if="replayData?.replay_permanent" class="px-2 py-1 rounded-md text-[10px] font-black bg-amber-500/10 text-amber-600 dark:text-amber-400">PERMANENT</span>
            <span v-else-if="replayData?.replay_expires_at" class="px-2 py-1 rounded-md text-[10px] font-black bg-cyan-500/10 text-cyan-600 dark:text-cyan-300">
              {{ formatDate(replayData.replay_expires_at) }} 到期
            </span>
          </div>
        </div>

        <div class="rounded-xl border border-slate-200 dark:border-white/10 p-3">
          <div class="flex items-center justify-between gap-2 mb-2">
            <p class="text-[10px] font-black uppercase tracking-widest text-slate-400">视角切换</p>
            <button
              @click="focusPerspectiveOnly = !focusPerspectiveOnly"
              :class="cn('px-2 py-1 rounded-md text-[10px] font-black uppercase tracking-wider', focusPerspectiveOnly ? 'bg-blue-600 text-white' : 'bg-slate-100 dark:bg-white/10 text-slate-500')"
            >
              {{ focusPerspectiveOnly ? '仅本视角' : '全局事件' }}
            </button>
          </div>
          <div class="space-y-2">
            <button
              v-for="profile in participants"
              :key="profile.uid"
              @click="selectedPerspectiveUID = Number(profile.uid)"
              :class="cn(
                'w-full text-left rounded-xl border px-3 py-2 transition-all',
                selectedPerspectiveUID === Number(profile.uid)
                  ? 'border-blue-500/40 bg-blue-500/10'
                  : 'border-slate-200 dark:border-white/10 bg-slate-50 dark:bg-white/5 hover:border-blue-500/30'
              )"
            >
              <p class="text-xs font-black text-slate-900 dark:text-white">{{ profile.nickname || profile.username || `UID ${profile.uid}` }}</p>
              <p class="text-[10px] text-slate-400 font-mono">UID: {{ profile.uid }}</p>
            </button>
          </div>
        </div>

        <div class="rounded-xl border border-slate-200 dark:border-white/10 p-3">
          <div class="flex items-center justify-between gap-2 mb-2">
            <p class="text-[10px] font-black uppercase tracking-widest text-slate-400">操作时间检测</p>
            <span class="px-2 py-1 rounded-md text-[10px] font-black bg-rose-500/10 text-rose-500">阈值 &lt; 3.00s</span>
          </div>
          <div class="space-y-2">
            <div
              v-for="stat in operationStats"
              :key="`op-stat-${stat.uid}`"
              class="rounded-lg border border-slate-200 dark:border-white/10 bg-slate-50 dark:bg-white/5 px-2 py-2"
            >
              <div class="flex items-center justify-between gap-2">
                <p class="text-[11px] font-semibold truncate">{{ stat.nickname }}</p>
                <span
                  :class="cn('px-2 py-0.5 rounded text-[10px] font-black', stat.fastCount > 0 ? 'bg-rose-500/15 text-rose-500' : 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400')"
                >
                  FAST {{ stat.fastCount }}
                </span>
              </div>
              <p class="text-[10px] text-slate-500 mt-1">
                平均 {{ formatOperationDuration(stat.avgMs) }} / 最快 {{ formatOperationDuration(stat.minMs) }} / 样本 {{ stat.totalCount }}
              </p>
            </div>
          </div>
        </div>

        <div v-if="isAdmin" class="rounded-xl border border-slate-200 dark:border-white/10 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-slate-400 mb-2">管理员处置</p>
          <div class="space-y-2">
            <div
              v-for="profile in participants.filter((p: any) => Number(p.uid) > 0)"
              :key="`action-${profile.uid}`"
              class="rounded-lg border border-slate-200 dark:border-white/10 px-2 py-2 bg-slate-50 dark:bg-white/5"
            >
              <div class="flex items-center justify-between gap-2">
                <span class="text-[11px] font-semibold">{{ profile.nickname || profile.username || profile.uid }}</span>
                <div class="flex items-center gap-1">
                  <button @click="kickFromReplay(profile)" class="p-1.5 rounded-md bg-amber-500/10 text-amber-600 dark:text-amber-400 hover:bg-amber-500/20">
                    <UserMinus class="w-3.5 h-3.5" />
                  </button>
                  <button @click="banFromReplay(profile)" class="p-1.5 rounded-md bg-rose-500/10 text-rose-500 hover:bg-rose-500/20">
                    <Ban class="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </aside>

      <main class="min-h-0 overflow-y-auto p-4 md:p-6 bg-slate-100/70 dark:bg-black/30">
        <div v-if="loading" class="h-full flex items-center justify-center">
          <div class="w-9 h-9 border-4 border-blue-500/20 border-t-blue-500 rounded-full animate-spin"></div>
        </div>

        <div v-else-if="loadError" class="max-w-2xl mx-auto rounded-2xl border border-rose-200 dark:border-rose-500/30 bg-rose-50 dark:bg-rose-500/10 p-5 text-rose-600 dark:text-rose-300">
          {{ loadError }}
        </div>

        <div v-else class="space-y-3">
          <template v-if="replayViewMode === 'game'">
            <div class="rounded-2xl border border-slate-200 dark:border-white/10 bg-white dark:bg-black/20 p-4">
              <div class="flex flex-wrap items-center justify-between gap-3 mb-4">
                <div>
                  <p class="text-[10px] font-black uppercase tracking-widest text-slate-400">游戏界面方式</p>
                  <p class="text-sm font-black">当前视角：{{ currentPerspectiveName }}</p>
                </div>
                <div class="text-right">
                  <p class="text-[10px] font-black uppercase tracking-widest text-slate-400">事件总量</p>
                  <p class="text-sm font-black text-blue-600 dark:text-blue-300">{{ visibleEvents.length }}</p>
                </div>
              </div>

              <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-2">
                <div
                  v-for="profile in participants"
                  :key="`player-overview-${profile.uid}`"
                  :class="cn(
                    'rounded-xl border px-3 py-2',
                    Number(profile.uid) === selectedPerspectiveUID
                      ? 'border-blue-500/40 bg-blue-500/10'
                      : 'border-slate-200 dark:border-white/10 bg-slate-50 dark:bg-white/5'
                  )"
                >
                  <p class="text-xs font-black text-slate-900 dark:text-white">{{ profile.nickname || profile.username || `UID ${profile.uid}` }}</p>
                  <p class="text-[10px] text-slate-400 font-mono">UID {{ profile.uid }}</p>
                  <p class="text-[10px] text-slate-500 mt-1">
                    快速操作 {{ operationStats.find((s: any) => s.uid === Number(profile.uid))?.fastCount || 0 }} 次
                  </p>
                </div>
              </div>
            </div>

            <div class="rounded-2xl border border-slate-200 dark:border-white/10 bg-white dark:bg-black/20 p-4">
              <div class="flex items-center justify-between gap-2 mb-3">
                <p class="text-[10px] font-black uppercase tracking-widest text-slate-400">操作流</p>
                <p class="text-[10px] text-slate-500">按游戏节奏展示操作，并检测时长</p>
              </div>

              <div class="space-y-2 max-h-[62vh] overflow-y-auto pr-1">
                <div
                  v-for="(evt, idx) in visibleEvents"
                  :key="`game-mode-${evt.at || idx}-${idx}`"
                  :class="cn(
                    'rounded-xl border px-3 py-2',
                    evt.isFastOperation ? 'border-rose-400/40 bg-rose-500/10' : 'border-slate-200 dark:border-white/10 bg-slate-50 dark:bg-white/5'
                  )"
                >
                  <div class="flex items-center justify-between gap-2 mb-1">
                    <span class="text-[10px] px-2 py-1 rounded-md bg-slate-100 dark:bg-white/10 text-slate-500 uppercase font-black tracking-wider">{{ formatEventType(evt.eventType) }}</span>
                    <div class="flex items-center gap-2">
                      <span
                        v-if="evt.operationMs != null"
                        :class="cn('text-[10px] px-2 py-1 rounded-md font-black', evt.isFastOperation ? 'bg-rose-500/20 text-rose-500' : 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400')"
                      >
                        {{ formatOperationDuration(evt.operationMs) }}
                      </span>
                      <span class="text-[10px] font-mono text-slate-400">{{ formatDate(evt.at) }}</span>
                    </div>
                  </div>
                  <p class="text-sm font-semibold text-slate-700 dark:text-slate-200">{{ idx + 1 }}. {{ describeEvent(evt) }}</p>
                </div>

                <div v-if="visibleEvents.length === 0" class="rounded-2xl border border-dashed border-slate-300 dark:border-white/20 bg-white/70 dark:bg-white/[0.03] p-8 text-center">
                  <p class="text-sm font-black text-slate-400 uppercase tracking-widest">NO EVENTS IN CURRENT PERSPECTIVE</p>
                </div>
              </div>
            </div>
          </template>

          <template v-else>
            <div class="rounded-2xl border border-slate-200 dark:border-white/10 bg-white dark:bg-black/20 p-4 flex items-center justify-between">
              <div>
                <p class="text-[10px] font-black uppercase tracking-widest text-slate-400">当前视角</p>
                <p class="text-sm font-black">{{ currentPerspectiveName }}</p>
              </div>
              <div class="text-right">
                <p class="text-[10px] font-black uppercase tracking-widest text-slate-400">事件数</p>
                <p class="text-sm font-black text-blue-600 dark:text-blue-300">{{ visibleEvents.length }}</p>
              </div>
            </div>

            <div
              v-for="(evt, idx) in visibleEvents"
              :key="`${evt.at || idx}-${idx}`"
              class="rounded-2xl border border-slate-200 dark:border-white/10 bg-white dark:bg-black/20 px-4 py-3"
            >
              <div class="flex items-center justify-between gap-3 mb-1">
                <span class="text-[10px] px-2 py-1 rounded-md bg-slate-100 dark:bg-white/10 text-slate-500 uppercase font-black tracking-wider">{{ formatEventType(evt.eventType) }}</span>
                <div class="flex items-center gap-2">
                  <span
                    v-if="evt.operationMs != null"
                    :class="cn('text-[10px] px-2 py-1 rounded-md font-black', evt.isFastOperation ? 'bg-rose-500/20 text-rose-500' : 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400')"
                  >
                    {{ formatOperationDuration(evt.operationMs) }}
                  </span>
                  <span class="text-[10px] font-mono text-slate-400">{{ formatDate(evt.at) }}</span>
                </div>
              </div>
              <p class="text-sm font-semibold text-slate-700 dark:text-slate-200">{{ idx + 1 }}. {{ describeEvent(evt) }}</p>
            </div>

            <div v-if="visibleEvents.length === 0" class="rounded-2xl border border-dashed border-slate-300 dark:border-white/20 bg-white/70 dark:bg-white/[0.03] p-8 text-center">
              <p class="text-sm font-black text-slate-400 uppercase tracking-widest">NO EVENTS IN CURRENT PERSPECTIVE</p>
            </div>
          </template>
        </div>
      </main>
    </div>
  </div>
</template>
