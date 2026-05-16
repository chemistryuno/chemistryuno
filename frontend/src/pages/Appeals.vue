<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { pageClassNames } from '@lib'
import {
  AlertTriangle,
  ArrowLeft,
  CheckCircle2,
  Clock3,
  FileText,
  Loader2,
  RefreshCw,
  Send,
  ShieldAlert,
  XCircle,
} from 'lucide-vue-next'
import { authAPI } from '../utils/api'
import { formatBanUntil, getBanState } from '../utils/banState'
import { useDialog } from '../utils/dialog'

type AppealStatus = 'pending' | 'under_review' | 'approved' | 'rejected' | string

interface AppealRecord {
  id: number
  room_id?: string
  risk_score_id?: number
  sanction_id?: number
  reason: string
  evidence?: string
  status: AppealStatus
  review_remark?: string
  compensation_amount?: number
  compensation_status?: string
  compensation_note?: string
  submitted_at?: string
  created_at?: string
}

interface SanctionRecord {
  id?: number
  room_id?: string
  risk_score_id?: number
  sanction_type?: string
  status?: string
  reason?: string
  effective_until?: string
  expires_at?: string
  created_at?: string
}

const loading = ref(true)
const router = useRouter()
const { showConfirm } = useDialog()
const submitting = ref(false)
const claimingId = ref<number | null>(null)
const error = ref('')
const fieldError = ref('')
const appeals = ref<AppealRecord[]>([])
const sanctions = ref<SanctionRecord[]>([])
const roomId = ref('')
const lockedRoomIds = ref<string[]>([])
const riskScoreId = ref('')
const sanctionId = ref('')
const reason = ref('')
const evidence = ref('')

const currentUser = ref<any>({})
const user = computed(() => {
  if (Object.keys(currentUser.value).length > 0) return currentUser.value
  try {
    return JSON.parse(localStorage.getItem('user') || '{}')
  } catch {
    return {}
  }
})

const banState = computed(() => getBanState(user.value))
const activeAppeal = computed(() =>
  appeals.value.find(appeal => appeal.status === 'pending' || appeal.status === 'under_review')
)
const latestSanction = computed(() => sanctions.value[0])
const serverBanState = ref<{ is_banned?: boolean; banned_until?: string; ban_reason?: string; can_submit?: boolean }>({})
const appealEntryLoaded = ref(false)
const banStatusQueryFailed = ref(false)
const isFutureOrOpenEnded = (value?: string) => {
  if (!value) return true
  const date = new Date(value)
  return Number.isNaN(date.getTime()) || date > new Date()
}
const activeBanSanction = computed(() =>
  sanctions.value.find(sanction =>
    sanction.sanction_type === 'ban' &&
    (!sanction.status || sanction.status === 'active') &&
    isFutureOrOpenEnded(sanction.effective_until || sanction.expires_at)
  )
)
const displayBanState = computed(() => {
  if (activeBanSanction.value) {
    return {
      isBanned: true,
      bannedUntil: activeBanSanction.value.effective_until || activeBanSanction.value.expires_at || serverBanState.value.banned_until || banState.value.bannedUntil,
      banReason: activeBanSanction.value.reason || serverBanState.value.ban_reason || banState.value.banReason || '',
    }
  }
  if (!appealEntryLoaded.value) return banState.value
  if (!serverBanState.value.is_banned) return { isBanned: false }
  return {
    isBanned: true,
    bannedUntil: serverBanState.value.banned_until || banState.value.bannedUntil,
    banReason: serverBanState.value.ban_reason || banState.value.banReason || '',
  }
})
const isBannedForAppeal = computed(() => {
  if (activeBanSanction.value) return true
  return appealEntryLoaded.value ? Boolean(serverBanState.value.is_banned) : banState.value.isBanned
})
const canSubmit = computed(() => !submitting.value && isBannedForAppeal.value && !activeAppeal.value && reason.value.trim().length > 0)
const evidenceLimit = 1000
const canClaimCompensation = (appeal: AppealRecord) =>
  appeal.status === 'approved' &&
  (appeal.compensation_amount || 0) > 0 &&
  appeal.compensation_status !== 'ok'

const normalizeList = (payload: any, key: string) => {
  if (Array.isArray(payload?.[key])) return payload[key]
  if (Array.isArray(payload?.data)) return payload.data
  return []
}

const findMatchingSanction = () => {
  const riskID = Number(riskScoreId.value) || undefined
  const rooms = lockedRoomIds.value.length ? lockedRoomIds.value : [roomId.value].filter(Boolean)
  return sanctions.value.find(sanction => {
    if (riskID && sanction.risk_score_id === riskID) return true
    if (sanction.room_id && rooms.includes(sanction.room_id)) return true
    return false
  })
}

const loadPanel = async () => {
  loading.value = true
  error.value = ''
  appealEntryLoaded.value = false
  serverBanState.value = {}
  banStatusQueryFailed.value = false
  try {
    const loadAppealEntry = async () => {
      try {
        return await authAPI.getAppealEntryStatus()
      } catch {
        banStatusQueryFailed.value = true
        return { data: null }
      }
    }
    const loadSanctions = async () => {
      try {
        return await authAPI.getPlayerSanctions()
      } catch {
        banStatusQueryFailed.value = true
        return { data: { sanctions: [] } }
      }
    }
    const [appealResponse, entryResponse] = await Promise.all([
      authAPI.getPlayerAppeals(),
      loadAppealEntry(),
    ])
    const [sanctionResponse, userResponse] = await Promise.all([
      loadSanctions(),
      authAPI.refreshUserInfo().catch(() => null),
    ])
    if (userResponse?.data) {
      currentUser.value = userResponse.data
      localStorage.setItem('user', JSON.stringify(userResponse.data))
    }
    appeals.value = normalizeList(appealResponse.data, 'appeals')
    sanctions.value = normalizeList(sanctionResponse.data, 'sanctions')
    if (entryResponse.data) {
      serverBanState.value = entryResponse.data
      appealEntryLoaded.value = true
      lockedRoomIds.value = Array.isArray(entryResponse.data.room_ids) ? entryResponse.data.room_ids : []
      roomId.value = lockedRoomIds.value[0] || roomId.value
      riskScoreId.value = entryResponse.data.latest_risk_score_id ? String(entryResponse.data.latest_risk_score_id) : riskScoreId.value
    }

    const matchedSanction = findMatchingSanction()
    if (matchedSanction?.id) {
      sanctionId.value = String(matchedSanction.id)
    }

    const context = latestSanction.value || appeals.value[0]
    if (context && lockedRoomIds.value.length === 0) {
      roomId.value = context.room_id || roomId.value
      riskScoreId.value = context.risk_score_id ? String(context.risk_score_id) : riskScoreId.value
      sanctionId.value = context.id ? String(context.id) : sanctionId.value
    }
  } catch (err: any) {
    error.value = err.response?.data?.error || '申诉记录加载失败'
  } finally {
    loading.value = false
  }
}

const validate = async () => {
  fieldError.value = ''
  if (!isBannedForAppeal.value) {
    await promptFeedbackRedirect()
    return false
  }
  if (!reason.value.trim()) {
    fieldError.value = '请填写申诉理由'
    return false
  }
  if (evidence.value.length > evidenceLimit) {
    fieldError.value = `补充说明不能超过 ${evidenceLimit} 个字符`
    return false
  }
  if (activeAppeal.value) {
    fieldError.value = '已有待处理申诉，请等待管理员审核'
    return false
  }
  return true
}

const promptFeedbackRedirect = async () => {
  const ok = await showConfirm(
    '当前账号未处于封禁状态，不能提交申诉。你可以前往反馈页面撰写反馈。',
    '申诉受限',
    '前往反馈',
    '留在本页'
  )
  if (ok) {
    router.push({ path: '/feedbacks', query: { compose: '1' } })
  }
}

const submitAppeal = async () => {
  if (!(await validate())) return
  submitting.value = true
  error.value = ''
  try {
    const payload: { risk_score_id?: number; sanction_id?: number; reason: string; evidence?: string } = {
      reason: reason.value.trim(),
    }
    const parsedRiskScoreID = Number(riskScoreId.value)
    if (parsedRiskScoreID) payload.risk_score_id = parsedRiskScoreID
    const parsedSanctionID = Number(sanctionId.value)
    if (parsedSanctionID) payload.sanction_id = parsedSanctionID
    const trimmedEvidence = evidence.value.trim()
    if (trimmedEvidence) payload.evidence = trimmedEvidence
    await authAPI.submitAppeal(lockedRoomIds.value[0] || roomId.value.trim() || 'account', payload)
    reason.value = ''
    evidence.value = ''
    await loadPanel()
  } catch (err: any) {
    error.value = err.response?.data?.error || '提交申诉失败'
  } finally {
    submitting.value = false
  }
}

const claimCompensation = async (appeal: AppealRecord) => {
  if (!canClaimCompensation(appeal)) return
  claimingId.value = appeal.id
  error.value = ''
  try {
    await authAPI.claimAppealCompensation(appeal.id)
    await loadPanel()
  } catch (err: any) {
    error.value = err.response?.data?.error || '领取补偿失败'
  } finally {
    claimingId.value = null
  }
}

const statusMeta = (status: AppealStatus) => {
  if (status === 'approved') return { label: '已通过', icon: CheckCircle2, cls: 'text-emerald-600 bg-emerald-50 border-emerald-200 dark:text-emerald-300 dark:bg-emerald-500/10 dark:border-emerald-500/20' }
  if (status === 'rejected') return { label: '已驳回', icon: XCircle, cls: 'text-rose-600 bg-rose-50 border-rose-200 dark:text-rose-300 dark:bg-rose-500/10 dark:border-rose-500/20' }
  if (status === 'under_review') return { label: '审核中', icon: Clock3, cls: 'text-sky-600 bg-sky-50 border-sky-200 dark:text-sky-300 dark:bg-sky-500/10 dark:border-sky-500/20' }
  return { label: '待审核', icon: Clock3, cls: 'text-amber-600 bg-amber-50 border-amber-200 dark:text-amber-300 dark:bg-amber-500/10 dark:border-amber-500/20' }
}

const formatDate = (value?: string) => {
  if (!value) return '未记录'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '未记录'
  return date.toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' })
}

const compensationStatusLabel = (status?: string) => {
  if (status === 'ok') return '已领取'
  if (status === 'failed') return '领取失败，可重试'
  return '待领取'
}

onMounted(loadPanel)
</script>

<template>
  <div :class="pageClassNames.appWhite">
    <!-- Background Effects -->
    <div class="fixed inset-0 overflow-hidden pointer-events-none z-0">
      <div class="absolute top-[-10%] right-[-10%] w-[50%] h-[50%] bg-blue-500/5 rounded-full blur-[120px]" />
      <div class="absolute bottom-[-10%] left-[-10%] w-[50%] h-[50%] bg-purple-500/5 rounded-full blur-[120px]" />
    </div>

    <main class="relative z-10 mx-auto flex w-full max-w-6xl flex-col gap-5 px-4 py-6 sm:px-6 lg:px-8">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <RouterLink to="/profile" class="mb-3 inline-flex items-center gap-2 text-xs font-bold text-slate-500 transition-colors hover:text-slate-900 dark:text-slate-400 dark:hover:text-white">
            <ArrowLeft class="h-4 w-4" />
            返回个人中心
          </RouterLink>
          <h1 class="text-2xl font-black tracking-tight sm:text-3xl">申诉中心</h1>
          <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">查看处罚状态，提交申诉，并追踪审核结果。</p>
        </div>
        <button
          class="inline-flex h-10 items-center justify-center gap-2 rounded-lg border border-slate-200 bg-white px-4 text-xs font-black text-slate-600 transition-colors hover:bg-slate-100 disabled:opacity-60 dark:border-white/10 dark:bg-white/5 dark:text-slate-200 dark:hover:bg-white/10"
          :disabled="loading"
          @click="loadPanel"
        >
          <Loader2 v-if="loading" class="h-4 w-4 animate-spin" />
          <RefreshCw v-else class="h-4 w-4" />
          刷新
        </button>
      </div>

      <div v-if="error" class="flex items-center justify-between gap-3 rounded-lg border border-rose-200 bg-rose-50 p-4 text-sm text-rose-700 dark:border-rose-500/20 dark:bg-rose-500/10 dark:text-rose-200">
        <span>{{ error }}</span>
        <button class="rounded-md border border-current px-3 py-1 text-xs font-bold" @click="loadPanel">重试</button>
      </div>

      <div class="grid gap-6 lg:grid-cols-[380px_1fr]">
        <section class="space-y-6">
          <div class="group relative overflow-hidden rounded-2xl border border-white/60 bg-white/60 p-6 shadow-xl shadow-blue-900/5 backdrop-blur-xl transition-all hover:shadow-2xl hover:shadow-blue-900/10 dark:border-white/10 dark:bg-black/40 dark:shadow-none dark:hover:bg-black/50">
            <div class="absolute -right-10 -top-10 h-32 w-32 rounded-full bg-rose-500/5 blur-3xl transition-all group-hover:bg-rose-500/10"></div>
            <div class="relative flex items-center gap-4">
              <div class="flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br from-rose-500/10 to-rose-500/5 text-rose-600 ring-1 ring-inset ring-rose-500/20 dark:text-rose-400">
                <ShieldAlert class="h-5 w-5" />
              </div>
              <div>
                <h2 class="text-sm font-black">当前状态</h2>
                <p class="text-xs text-slate-500 dark:text-slate-400">账户与处罚信息</p>
              </div>
            </div>
            <div class="mt-5 space-y-3 text-sm">
              <div v-if="displayBanState.isBanned" class="rounded-lg border border-rose-200 bg-rose-50 p-3 text-rose-700 dark:border-rose-500/20 dark:bg-rose-500/10 dark:text-rose-200">
                <div class="font-black">账号处于封禁中</div>
                <div class="mt-1 text-xs">截止时间：{{ formatBanUntil(displayBanState.bannedUntil) }}</div>
                <div v-if="displayBanState.banReason" class="mt-1 text-xs">原因：{{ displayBanState.banReason }}</div>
              </div>
              <div v-else-if="banStatusQueryFailed" class="rounded-lg border border-amber-200 bg-amber-50 p-3 text-amber-700 dark:border-amber-500/20 dark:bg-amber-500/10 dark:text-amber-200">
                <div class="font-black">账号封禁状态查询失败</div>
                <div class="mt-1 text-xs">请刷新后重试，或重新登录后再进入申诉中心。</div>
              </div>
              <div v-else class="rounded-lg border border-emerald-200 bg-emerald-50 p-3 text-emerald-700 dark:border-emerald-500/20 dark:bg-emerald-500/10 dark:text-emerald-200">
                <div class="font-black">账号未处于封禁状态</div>
                <div class="mt-1 text-xs">申诉仅面向封禁处罚，其他问题请前往反馈。</div>
              </div>

              <div v-if="latestSanction" class="rounded-lg border border-slate-200 p-3 dark:border-white/10">
                <div class="text-xs font-black text-slate-500 dark:text-slate-400">最近处罚</div>
                <div class="mt-1 font-bold">{{ latestSanction.sanction_type || '未标注类型' }}</div>
                <div class="mt-1 text-xs text-slate-500 dark:text-slate-400">{{ latestSanction.reason || '未提供原因' }}</div>
              </div>
            </div>
          </div>

          <form class="group relative overflow-hidden rounded-2xl border border-white/60 bg-white/60 p-6 shadow-xl shadow-blue-900/5 backdrop-blur-xl transition-all hover:shadow-2xl hover:shadow-blue-900/10 dark:border-white/10 dark:bg-black/40 dark:shadow-none dark:hover:bg-black/50" @submit.prevent="submitAppeal">
            <div class="absolute -right-10 -top-10 h-32 w-32 rounded-full bg-sky-500/5 blur-3xl transition-all group-hover:bg-sky-500/10"></div>
            <div class="relative flex items-center gap-4">
              <div class="flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br from-sky-500/10 to-sky-500/5 text-sky-600 ring-1 ring-inset ring-sky-500/20 dark:text-sky-400">
                <FileText class="h-5 w-5" />
              </div>
              <div>
                <h2 class="text-sm font-black">提交申诉</h2>
                <p class="text-xs text-slate-500 dark:text-slate-400">说明误判原因和证据</p>
              </div>
            </div>

            <div v-if="activeAppeal" class="mt-4 rounded-lg border border-amber-200 bg-amber-50 p-3 text-xs text-amber-700 dark:border-amber-500/20 dark:bg-amber-500/10 dark:text-amber-200">
              已有待处理申诉，当前不可重复提交。
            </div>

            <div class="mt-4 grid gap-3">
              <label class="grid gap-1 text-xs font-bold text-slate-500 dark:text-slate-400">
                自动关联房间 ID
                <div class="rounded-lg border border-slate-200/50 bg-slate-50/80 p-3 text-sm text-slate-700 dark:border-white/10 dark:bg-black/20 dark:text-slate-200">
                  <div v-if="lockedRoomIds.length" class="flex flex-wrap gap-2">
                    <span v-for="id in lockedRoomIds" :key="id" class="rounded-md bg-white px-2 py-1 text-xs font-black text-slate-700 ring-1 ring-slate-200 dark:bg-white/5 dark:text-slate-200 dark:ring-white/10">{{ id }}</span>
                  </div>
                  <span v-else class="text-slate-400">暂无可申诉房间列表</span>
                </div>
              </label>
              <div class="grid grid-cols-2 gap-3">
                <div class="grid gap-1">
                  <span class="text-xs font-bold text-slate-500 dark:text-slate-400">风险记录</span>
                  <div class="flex h-10 items-center rounded-lg bg-slate-50 px-3 text-sm font-medium text-slate-600 dark:bg-white/5 dark:text-slate-300">
                    {{ riskScoreId || '暂无相关记录' }}
                  </div>
                </div>
                <div class="grid gap-1">
                  <span class="text-xs font-bold text-slate-500 dark:text-slate-400">处罚记录</span>
                  <div class="flex h-10 items-center rounded-lg bg-slate-50 px-3 text-sm font-medium text-slate-600 dark:bg-white/5 dark:text-slate-300">
                    {{ sanctionId || '暂无相关记录' }}
                  </div>
                </div>
              </div>
              <label class="grid gap-1 text-xs font-bold text-slate-500 dark:text-slate-400">
                申诉理由
                <textarea v-model="reason" rows="4" class="resize-none rounded-lg border border-slate-200/50 bg-white/50 p-3 text-sm text-slate-900 outline-none transition-all placeholder:text-slate-400 focus:border-sky-500 focus:bg-white focus:ring-4 focus:ring-sky-500/10 dark:border-white/10 dark:bg-black/20 dark:text-white dark:focus:bg-black/40" placeholder="请说明为什么该异常行为可能是误判" />
              </label>
              <label class="grid gap-1 text-xs font-bold text-slate-500 dark:text-slate-400">
                补充说明
                <textarea v-model="evidence" rows="4" class="resize-none rounded-lg border border-slate-200/50 bg-white/50 p-3 text-sm text-slate-900 outline-none transition-all placeholder:text-slate-400 focus:border-sky-500 focus:bg-white focus:ring-4 focus:ring-sky-500/10 dark:border-white/10 dark:bg-black/20 dark:text-white dark:focus:bg-black/40" placeholder="网络波动、设备问题、回放时间点等" />
                <span class="text-right text-[11px]" :class="evidence.length > evidenceLimit ? 'text-rose-500' : 'text-slate-400'">{{ evidence.length }}/{{ evidenceLimit }}</span>
              </label>
              <div v-if="fieldError" class="text-xs font-bold text-rose-500">{{ fieldError }}</div>
              <button class="inline-flex h-11 items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-sky-600 to-blue-600 px-4 text-sm font-black text-white shadow-lg shadow-sky-500/20 transition-all hover:scale-[1.02] hover:from-sky-500 hover:to-blue-500 hover:shadow-sky-500/30 active:scale-[0.98] disabled:opacity-50 disabled:grayscale" :disabled="!canSubmit" @click.prevent="submitAppeal">
                <Loader2 v-if="submitting" class="h-4 w-4 animate-spin" />
                <Send v-else class="h-4 w-4" />
                提交申诉
              </button>
            </div>
          </form>
        </section>

        <div class="relative lg:h-full">
          <div class="flex h-full flex-col lg:absolute lg:inset-0">
            <section class="flex flex-1 flex-col overflow-hidden rounded-2xl border border-white/60 bg-white/60 shadow-xl shadow-blue-900/5 backdrop-blur-xl transition-all dark:border-white/10 dark:bg-black/40 dark:shadow-none">
              <div class="shrink-0 border-b border-slate-200/50 p-6 dark:border-white/10">
                <h2 class="text-base font-black">申诉历史</h2>
                <p class="mt-1 text-xs text-slate-500 dark:text-slate-400">共 {{ appeals.length }} 条记录</p>
              </div>
              <div class="flex-1 overflow-y-auto custom-scrollbar">
                <div v-if="loading" class="flex min-h-[320px] items-center justify-center text-slate-400">
                  <Loader2 class="h-6 w-6 animate-spin" />
                </div>
                <div v-else-if="appeals.length === 0" class="flex min-h-[320px] flex-col items-center justify-center p-8 text-center text-slate-400">
                  <AlertTriangle class="mb-3 h-8 w-8" />
                  <p class="text-sm font-bold">暂无申诉记录</p>
                </div>
                <div v-else class="divide-y divide-slate-200/50 dark:divide-white/10">
            <article v-for="appeal in appeals" :key="appeal.id" class="group/item p-6 transition-colors hover:bg-white/40 dark:hover:bg-white/[0.02]">
              <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <div class="flex items-center gap-2">
                    <component :is="statusMeta(appeal.status).icon" class="h-4 w-4" />
                    <h3 class="text-sm font-black">申诉 #{{ appeal.id }}</h3>
                  </div>
                  <p class="mt-1 text-xs text-slate-500 dark:text-slate-400">提交时间：{{ formatDate(appeal.submitted_at || appeal.created_at) }}</p>
                </div>
                <span class="inline-flex w-fit items-center rounded-full border px-3 py-1 text-xs font-black" :class="statusMeta(appeal.status).cls">
                  {{ statusMeta(appeal.status).label }}
                </span>
              </div>
              <div class="mt-4 rounded-xl border border-slate-200/50 bg-white/50 p-4 text-sm leading-relaxed text-slate-700 shadow-sm transition-all group-hover/item:bg-white dark:border-white/5 dark:bg-black/20 dark:text-slate-200 dark:group-hover/item:bg-black/40">
                {{ appeal.reason }}
              </div>
              <div v-if="appeal.evidence" class="mt-3 text-xs leading-relaxed text-slate-500 dark:text-slate-400">
                {{ appeal.evidence }}
              </div>
              <div v-if="appeal.review_remark" class="mt-3 rounded-xl border border-slate-200/50 bg-slate-50/50 p-4 text-xs leading-relaxed text-slate-600 dark:border-white/5 dark:bg-white/[0.02] dark:text-slate-300">
                审核备注：{{ appeal.review_remark }}
              </div>
              <div v-if="appeal.status === 'approved' && (appeal.compensation_amount || appeal.compensation_note)" class="mt-3 rounded-lg border border-emerald-200 bg-emerald-50 p-3 text-xs text-emerald-700 dark:border-emerald-500/20 dark:bg-emerald-500/10 dark:text-emerald-200">
                <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                  <div>
                    <div class="font-black">补偿：{{ appeal.compensation_amount || 0 }} 燃素</div>
                    <div class="mt-1">状态：{{ compensationStatusLabel(appeal.compensation_status) }}</div>
                    <div v-if="appeal.compensation_note" class="mt-1">{{ appeal.compensation_note }}</div>
                  </div>
                  <button
                    v-if="canClaimCompensation(appeal)"
                    class="inline-flex h-9 items-center justify-center gap-2 rounded-lg bg-emerald-600 px-4 text-xs font-black text-white transition-colors hover:bg-emerald-500 disabled:opacity-60"
                    :disabled="claimingId === appeal.id"
                    @click="claimCompensation(appeal)"
                  >
                    <Loader2 v-if="claimingId === appeal.id" class="h-4 w-4 animate-spin" />
                    <CheckCircle2 v-else class="h-4 w-4" />
                    领取补偿
                  </button>
                </div>
              </div>
                  </article>
                </div>
              </div>
            </section>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>
