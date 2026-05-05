<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
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
  reason?: string
  expires_at?: string
  created_at?: string
}

const loading = ref(true)
const submitting = ref(false)
const error = ref('')
const fieldError = ref('')
const appeals = ref<AppealRecord[]>([])
const sanctions = ref<SanctionRecord[]>([])
const roomId = ref('')
const riskScoreId = ref('')
const sanctionId = ref('')
const reason = ref('')
const evidence = ref('')

const user = computed(() => {
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
const canSubmit = computed(() => !submitting.value && !activeAppeal.value && reason.value.trim().length > 0)
const evidenceLimit = 1000

const normalizeList = (payload: any, key: string) => {
  if (Array.isArray(payload?.[key])) return payload[key]
  if (Array.isArray(payload?.data)) return payload.data
  return []
}

const loadPanel = async () => {
  loading.value = true
  error.value = ''
  try {
    const [appealResponse, sanctionResponse] = await Promise.all([
      authAPI.getPlayerAppeals(),
      authAPI.getPlayerSanctions().catch(() => ({ data: { sanctions: [] } })),
    ])
    appeals.value = normalizeList(appealResponse.data, 'appeals')
    sanctions.value = normalizeList(sanctionResponse.data, 'sanctions')

    const context = latestSanction.value || appeals.value[0]
    if (context) {
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

const validate = () => {
  fieldError.value = ''
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

const submitAppeal = async () => {
  if (!validate()) return
  submitting.value = true
  error.value = ''
  try {
    await authAPI.submitAppeal(roomId.value.trim() || 'account', {
      risk_score_id: Number(riskScoreId.value) || undefined,
      sanction_id: Number(sanctionId.value) || undefined,
      reason: reason.value.trim(),
      evidence: evidence.value.trim(),
    })
    reason.value = ''
    evidence.value = ''
    await loadPanel()
  } catch (err: any) {
    error.value = err.response?.data?.error || '提交申诉失败'
  } finally {
    submitting.value = false
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

onMounted(loadPanel)
</script>

<template>
  <div class="min-h-screen bg-slate-50 text-slate-900 dark:bg-[#08090b] dark:text-white">
    <main class="mx-auto flex w-full max-w-6xl flex-col gap-5 px-4 py-6 sm:px-6 lg:px-8">
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

      <div class="grid gap-5 lg:grid-cols-[360px_1fr]">
        <section class="space-y-5">
          <div class="rounded-lg border border-slate-200 bg-white p-5 shadow-sm dark:border-white/10 dark:bg-[#111318]">
            <div class="flex items-center gap-3">
              <div class="flex h-11 w-11 items-center justify-center rounded-lg bg-rose-50 text-rose-600 dark:bg-rose-500/10 dark:text-rose-300">
                <ShieldAlert class="h-5 w-5" />
              </div>
              <div>
                <h2 class="text-sm font-black">当前状态</h2>
                <p class="text-xs text-slate-500 dark:text-slate-400">账户与处罚信息</p>
              </div>
            </div>
            <div class="mt-5 space-y-3 text-sm">
              <div v-if="banState.isBanned" class="rounded-lg border border-rose-200 bg-rose-50 p-3 text-rose-700 dark:border-rose-500/20 dark:bg-rose-500/10 dark:text-rose-200">
                <div class="font-black">账号处于封禁中</div>
                <div class="mt-1 text-xs">截止时间：{{ formatBanUntil(banState.bannedUntil) }}</div>
                <div v-if="banState.banReason" class="mt-1 text-xs">原因：{{ banState.banReason }}</div>
              </div>
              <div v-else class="rounded-lg border border-emerald-200 bg-emerald-50 p-3 text-emerald-700 dark:border-emerald-500/20 dark:bg-emerald-500/10 dark:text-emerald-200">
                <div class="font-black">未检测到本地封禁状态</div>
                <div class="mt-1 text-xs">仍可针对异常检测或处罚记录提交说明。</div>
              </div>

              <div v-if="latestSanction" class="rounded-lg border border-slate-200 p-3 dark:border-white/10">
                <div class="text-xs font-black text-slate-500 dark:text-slate-400">最近处罚</div>
                <div class="mt-1 font-bold">{{ latestSanction.sanction_type || '未标注类型' }}</div>
                <div class="mt-1 text-xs text-slate-500 dark:text-slate-400">{{ latestSanction.reason || '未提供原因' }}</div>
              </div>
            </div>
          </div>

          <form class="rounded-lg border border-slate-200 bg-white p-5 shadow-sm dark:border-white/10 dark:bg-[#111318]" @submit.prevent="submitAppeal">
            <div class="flex items-center gap-3">
              <div class="flex h-11 w-11 items-center justify-center rounded-lg bg-sky-50 text-sky-600 dark:bg-sky-500/10 dark:text-sky-300">
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
                房间 ID
                <input v-model="roomId" class="h-10 rounded-lg border border-slate-200 bg-white px-3 text-sm text-slate-900 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20 dark:text-white" placeholder="不知道可留空" />
              </label>
              <div class="grid grid-cols-2 gap-3">
                <label class="grid gap-1 text-xs font-bold text-slate-500 dark:text-slate-400">
                  风险记录
                  <input v-model="riskScoreId" type="number" min="0" class="h-10 rounded-lg border border-slate-200 bg-white px-3 text-sm text-slate-900 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20 dark:text-white" />
                </label>
                <label class="grid gap-1 text-xs font-bold text-slate-500 dark:text-slate-400">
                  处罚记录
                  <input v-model="sanctionId" type="number" min="0" class="h-10 rounded-lg border border-slate-200 bg-white px-3 text-sm text-slate-900 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20 dark:text-white" />
                </label>
              </div>
              <label class="grid gap-1 text-xs font-bold text-slate-500 dark:text-slate-400">
                申诉理由
                <textarea v-model="reason" rows="4" class="resize-none rounded-lg border border-slate-200 bg-white p-3 text-sm text-slate-900 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20 dark:text-white" placeholder="请说明为什么该异常行为可能是误判" />
              </label>
              <label class="grid gap-1 text-xs font-bold text-slate-500 dark:text-slate-400">
                补充说明
                <textarea v-model="evidence" rows="4" class="resize-none rounded-lg border border-slate-200 bg-white p-3 text-sm text-slate-900 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20 dark:text-white" placeholder="网络波动、设备问题、回放时间点等" />
                <span class="text-right text-[11px]" :class="evidence.length > evidenceLimit ? 'text-rose-500' : 'text-slate-400'">{{ evidence.length }}/{{ evidenceLimit }}</span>
              </label>
              <div v-if="fieldError" class="text-xs font-bold text-rose-500">{{ fieldError }}</div>
              <button class="inline-flex h-11 items-center justify-center gap-2 rounded-lg bg-sky-600 px-4 text-sm font-black text-white transition-colors hover:bg-sky-500 disabled:cursor-not-allowed disabled:bg-slate-400" :disabled="!canSubmit">
                <Loader2 v-if="submitting" class="h-4 w-4 animate-spin" />
                <Send v-else class="h-4 w-4" />
                提交申诉
              </button>
            </div>
          </form>
        </section>

        <section class="rounded-lg border border-slate-200 bg-white shadow-sm dark:border-white/10 dark:bg-[#111318]">
          <div class="border-b border-slate-200 p-5 dark:border-white/10">
            <h2 class="text-sm font-black">申诉历史</h2>
            <p class="mt-1 text-xs text-slate-500 dark:text-slate-400">共 {{ appeals.length }} 条记录</p>
          </div>
          <div v-if="loading" class="flex min-h-[320px] items-center justify-center text-slate-400">
            <Loader2 class="h-6 w-6 animate-spin" />
          </div>
          <div v-else-if="appeals.length === 0" class="flex min-h-[320px] flex-col items-center justify-center p-8 text-center text-slate-400">
            <AlertTriangle class="mb-3 h-8 w-8" />
            <p class="text-sm font-bold">暂无申诉记录</p>
          </div>
          <div v-else class="divide-y divide-slate-200 dark:divide-white/10">
            <article v-for="appeal in appeals" :key="appeal.id" class="p-5">
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
              <div class="mt-4 rounded-lg bg-slate-50 p-3 text-sm leading-relaxed text-slate-700 dark:bg-black/20 dark:text-slate-200">
                {{ appeal.reason }}
              </div>
              <div v-if="appeal.evidence" class="mt-3 text-xs leading-relaxed text-slate-500 dark:text-slate-400">
                {{ appeal.evidence }}
              </div>
              <div v-if="appeal.review_remark" class="mt-3 rounded-lg border border-slate-200 p-3 text-xs text-slate-600 dark:border-white/10 dark:text-slate-300">
                审核备注：{{ appeal.review_remark }}
              </div>
              <div v-if="appeal.status === 'approved' && (appeal.compensation_amount || appeal.compensation_note)" class="mt-3 rounded-lg border border-emerald-200 bg-emerald-50 p-3 text-xs text-emerald-700 dark:border-emerald-500/20 dark:bg-emerald-500/10 dark:text-emerald-200">
                补偿：{{ appeal.compensation_amount || 0 }} 燃素
                <span v-if="appeal.compensation_status">，状态：{{ appeal.compensation_status }}</span>
                <div v-if="appeal.compensation_note" class="mt-1">{{ appeal.compensation_note }}</div>
              </div>
            </article>
          </div>
        </section>
      </div>
    </main>
  </div>
</template>
