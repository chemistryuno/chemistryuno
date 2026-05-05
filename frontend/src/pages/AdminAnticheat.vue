<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useDialog } from '../utils/dialog'
import { adminAPI } from '../utils/api'
import {
  Shield,
  Search as SearchIcon,
  ArrowLeft,
  ChevronLeft,
  ChevronRight,
  CheckCircle,
  XCircle,
  Eye,
  Save,
  Settings,
  Download,
  Filter,
} from 'lucide-vue-next'

const { showAlert, showPrompt } = useDialog()
const activeTab = ref<'detection' | 'appeals' | 'config' | 'audit'>('detection')
const loading = ref(false)

// ==================== Detection List ====================
const detectionList = ref<any[]>([])
const detectionSearchTerm = ref('')
const detectionStatusFilter = ref<'all' | 'observe' | 'warning' | 'mute' | 'ban'>('all')
const detectionPage = ref(1)
const detectionLimit = ref(20)
const detectionTotal = ref(0)

const filteredDetections = computed(() => {
  let items = detectionList.value
  if (detectionSearchTerm.value) {
    const term = detectionSearchTerm.value.toLowerCase()
    items = items.filter(d => 
      d.player_id?.toString().includes(term) || 
      d.room_id?.toString().includes(term)
    )
  }
  if (detectionStatusFilter.value !== 'all') {
    items = items.filter(d => d.sanction_type === detectionStatusFilter.value)
  }
  return items
})

const loadDetections = async () => {
  loading.value = true
  try {
    const response = await adminAPI.getDetectionList({
      page: detectionPage.value,
      limit: detectionLimit.value,
      status: detectionStatusFilter.value !== 'all' ? detectionStatusFilter.value : undefined,
    })
    detectionList.value = response.data?.detections || response.data?.data || []
    detectionTotal.value = response.data?.total || response.data?.count || detectionList.value.length
  } catch (error: any) {
    showAlert(error.response?.data?.error || '加载检测列表失败', '错误')
  } finally {
    loading.value = false
  }
}

// ==================== Detection Details & Review ====================
const showDetailModal = ref(false)
const selectedDetection = ref<any>(null)
const reviewDecision = ref<'confirm' | 'override'>('confirm')
const reviewNote = ref('')
const enforcementReason = ref('Anticheat panel manual enforcement')
const enforcementUntil = ref('')
const enforcementLoading = ref(false)

const normalizeDetectionDetail = (payload: any, fallback: any) => {
  const risk = payload?.risk_score || payload || {}
  const sanctions = payload?.sanctions || []
  return {
    ...fallback,
    ...risk,
    id: risk.id || fallback.id,
    player_id: risk.player_id || risk.player_uid || fallback.player_id,
    player_uid: risk.player_uid || risk.player_id || fallback.player_uid || fallback.player_id,
    room_id: risk.room_id || fallback.room_id,
    risk_score: risk.risk_score ?? fallback.risk_score ?? 0,
    sanction_type: fallback.sanction_type || sanctions[0]?.sanction_type || 'observe',
    sanctions,
  }
}

const defaultBanUntil = () => new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString().slice(0, 16)

const openDetectionDetail = async (detection: any) => {
  try {
    const response = await adminAPI.getDetectionDetail(detection.id)
    selectedDetection.value = normalizeDetectionDetail(response.data, detection)
    enforcementReason.value = 'Anticheat panel manual enforcement'
    enforcementUntil.value = defaultBanUntil()
    showDetailModal.value = true
  } catch (error: any) {
    showAlert(error.response?.data?.error || '加载检测详情失败', '错误')
  }
}

const submitReview = async () => {
  if (!selectedDetection.value) return
  
  try {
    await adminAPI.reviewDetection(selectedDetection.value.id, {
      decision: reviewDecision.value,
      note: reviewNote.value,
    })
    showAlert('审核已提交', '成功')
    showDetailModal.value = false
    reviewDecision.value = 'confirm'
    reviewNote.value = ''
    loadDetections()
  } catch (error: any) {
    showAlert(error.response?.data?.error || '提交审核失败', '错误')
  }
}

const handlePanelBan = async () => {
  if (!selectedDetection.value) return
  if (!enforcementUntil.value || !enforcementReason.value.trim()) {
    showAlert('请填写封禁截止时间和原因', '参数缺失')
    return
  }
  const until = new Date(enforcementUntil.value)
  if (Number.isNaN(until.getTime()) || until <= new Date()) {
    showAlert('封禁截止时间必须晚于当前时间', '时间无效')
    return
  }

  enforcementLoading.value = true
  try {
    await adminAPI.banFromAnticheatPanel({
      player_uid: Number(selectedDetection.value.player_uid || selectedDetection.value.player_id),
      banned_until: until.toISOString(),
      reason: enforcementReason.value.trim(),
      room_id: selectedDetection.value.room_id,
      risk_score_id: Number(selectedDetection.value.id) || undefined,
    })
    showAlert('已执行封禁并写入审计日志', '操作完成')
    await Promise.all([loadDetections(), loadAuditLog()])
  } catch (error: any) {
    showAlert(error.response?.data?.error || '封禁失败', '操作失败')
  } finally {
    enforcementLoading.value = false
  }
}

const handlePanelUnban = async () => {
  if (!selectedDetection.value) return
  enforcementLoading.value = true
  try {
    await adminAPI.unbanFromAnticheatPanel({
      player_uid: Number(selectedDetection.value.player_uid || selectedDetection.value.player_id),
      reason: enforcementReason.value.trim() || 'Manual unban from anticheat panel',
      room_id: selectedDetection.value.room_id,
    })
    showAlert('已解除封禁并写入审计日志', '操作完成')
    await Promise.all([loadDetections(), loadAuditLog()])
  } catch (error: any) {
    showAlert(error.response?.data?.error || '解封失败', '操作失败')
  } finally {
    enforcementLoading.value = false
  }
}

// ==================== Appeals Management ====================
const appealsList = ref<any[]>([])
const appealsSearchTerm = ref('')
const appealsStatusFilter = ref<'all' | 'pending' | 'approved' | 'rejected'>('all')
const appealsPage = ref(1)
const appealsLimit = ref(20)
const appealsTotal = ref(0)

const filteredAppeals = computed(() => {
  let items = appealsList.value
  if (appealsSearchTerm.value) {
    const term = appealsSearchTerm.value.toLowerCase()
    items = items.filter(a => 
      (a.player_id || a.player_uid)?.toString().includes(term) || 
      a.room_id?.toString().includes(term)
    )
  }
  if (appealsStatusFilter.value !== 'all') {
    items = items.filter(a => a.status === appealsStatusFilter.value)
  }
  return items
})

const loadAppeals = async () => {
  loading.value = true
  try {
    const response = await adminAPI.getAppealsList({
      page: appealsPage.value,
      limit: appealsLimit.value,
      status: appealsStatusFilter.value !== 'all' ? appealsStatusFilter.value : undefined,
    })
    appealsList.value = response.data?.appeals || response.data?.data || []
    appealsTotal.value = response.data?.total || response.data?.count || appealsList.value.length
  } catch (error: any) {
    showAlert(error.response?.data?.error || '加载申诉列表失败', '错误')
  } finally {
    loading.value = false
  }
}

// ==================== Appeal Approval with Compensation ====================
const showApprovalModal = ref(false)
const pendingAppealId = ref<string>('')
const approvalNote = ref('')
const compensationAmount = ref(100) // 默认补偿金额
const compensationMessage = ref('') // 自定义补偿文案
const defaultCompensationMessage = ref('由于反作弊系统将您误封，在此，ChemistryUNO开发组向受到影响的研究员提供燃素补偿，感谢研究员对维护纯净游戏环境做出的贡献')

const handleApproveAppeal = async (appealId: string) => {
  pendingAppealId.value = appealId
  approvalNote.value = ''
  compensationAmount.value = 100
  compensationMessage.value = defaultCompensationMessage.value
  showApprovalModal.value = true
}

const confirmApproval = async () => {
  if (!pendingAppealId.value) return
  
  try {
    await adminAPI.approveAppeal(pendingAppealId.value, { 
      note: approvalNote.value || '通过审核',
      compensation_amount: compensationAmount.value,
      compensation_message: compensationMessage.value || defaultCompensationMessage.value
    })
    showAlert('申诉已批准，补偿已发放', '成功')
    showApprovalModal.value = false
    await Promise.all([loadAppeals(), loadAuditLog()])
  } catch (error: any) {
    showAlert(error.response?.data?.error || '批准申诉失败', '错误')
  }
}

const cancelApproval = () => {
  showApprovalModal.value = false
  pendingAppealId.value = ''
}

const handleRejectAppeal = async (appealId: string) => {
  const note = await showPrompt('请输入拒绝理由（可选）：', '拒绝申诉')
  if (note === null) return
  
  try {
    await adminAPI.rejectAppeal(appealId, { note })
    showAlert('申诉已拒绝', '成功')
    await Promise.all([loadAppeals(), loadAuditLog()])
  } catch (error: any) {
    showAlert(error.response?.data?.error || '拒绝申诉失败', '错误')
  }
}

// ==================== Config Management ====================
const configData = ref<any>(null)
const editingConfig = ref(false)
const tempConfig = ref<any>(null)

const loadConfig = async () => {
  loading.value = true
  try {
    const response = await adminAPI.getAnticheatConfig()
    configData.value = response.data
  } catch (error: any) {
    showAlert(error.response?.data?.error || '加载配置失败', '错误')
  } finally {
    loading.value = false
  }
}

const startEditConfig = () => {
  tempConfig.value = JSON.parse(JSON.stringify(configData.value))
  editingConfig.value = true
}

const cancelEditConfig = () => {
  editingConfig.value = false
  tempConfig.value = null
}

const saveConfig = async () => {
  if (!tempConfig.value) return
  
  try {
    await adminAPI.updateAnticheatConfig(tempConfig.value)
    showAlert('配置已更新', '成功')
    configData.value = tempConfig.value
    editingConfig.value = false
    tempConfig.value = null
  } catch (error: any) {
    showAlert(error.response?.data?.error || '更新配置失败', '错误')
  }
}

// ==================== Audit Log ====================
const auditLogs = ref<any[]>([])
const auditSearchTerm = ref('')
const auditPage = ref(1)
const auditLimit = ref(20)
const auditTotal = ref(0)
const auditStartDate = ref('')
const auditEndDate = ref('')

const loadAuditLog = async () => {
  loading.value = true
  try {
    const response = await adminAPI.getAuditLog({
      page: auditPage.value,
      limit: auditLimit.value,
      player_id: auditSearchTerm.value || undefined,
      start_date: auditStartDate.value || undefined,
      end_date: auditEndDate.value || undefined,
    })
    auditLogs.value = response.data?.logs || []
    auditTotal.value = response.data?.total || 0
  } catch (error: any) {
    showAlert(error.response?.data?.error || '加载审计日志失败', '错误')
  } finally {
    loading.value = false
  }
}

const exportAuditLog = async () => {
  try {
    const response = await adminAPI.exportAuditLog({
      start_date: auditStartDate.value || undefined,
      end_date: auditEndDate.value || undefined,
    })
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `anticheat_audit_log_${new Date().toISOString().split('T')[0]}.xlsx`)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  } catch (error: any) {
    showAlert(error.response?.data?.error || '导出日志失败', '错误')
  }
}

// ==================== Lifecycle ====================
onMounted(() => {
  loadDetections()
})

watch(() => activeTab.value, (newTab) => {
  if (newTab === 'detection') {
    loadDetections()
  } else if (newTab === 'appeals') {
    loadAppeals()
  } else if (newTab === 'config') {
    loadConfig()
  } else if (newTab === 'audit') {
    loadAuditLog()
  }
})

// Utility functions
const getRiskColor = (score: number) => {
  if (score < 20) return 'text-green-600'
  if (score < 40) return 'text-blue-600'
  if (score < 60) return 'text-yellow-600'
  if (score < 80) return 'text-orange-600'
  return 'text-red-600'
}

const getSanctionBadge = (type: string) => {
  const badges: Record<string, { color: string; label: string }> = {
    observe: { color: 'bg-blue-100 text-blue-800', label: '观察' },
    warning: { color: 'bg-yellow-100 text-yellow-800', label: '警告' },
    mute: { color: 'bg-orange-100 text-orange-800', label: '禁言' },
    ban: { color: 'bg-red-100 text-red-800', label: '封号' },
  }
  return badges[type] || { color: 'bg-gray-100 text-gray-800', label: type }
}

const getStatusBadge = (status: string) => {
  const badges: Record<string, { color: string; label: string }> = {
    pending: { color: 'bg-gray-100 text-gray-800', label: '待审核' },
    approved: { color: 'bg-green-100 text-green-800', label: '已批准' },
    rejected: { color: 'bg-red-100 text-red-800', label: '已拒绝' },
  }
  return badges[status] || { color: 'bg-gray-100 text-gray-800', label: status }
}

const getCompensationBadge = (status: string) => {
  const badges: Record<string, { color: string; label: string }> = {
    pending: { color: 'bg-yellow-100 text-yellow-800', label: '待发放' },
    ok: { color: 'bg-green-100 text-green-800', label: '已发放' },
    failed: { color: 'bg-red-100 text-red-800', label: '发放失败' },
  }
  return badges[status] || { color: 'bg-gray-100 text-gray-800', label: status }
}
</script>

<template>
  <div class="min-h-screen bg-slate-50 text-slate-900 dark:bg-[#0a0a0c] dark:text-white selection:bg-blue-500/30">
    <!-- Background Effects -->
    <div class="fixed inset-0 overflow-hidden pointer-events-none z-0">
      <div class="absolute top-[-10%] right-[-10%] w-[50%] h-[50%] bg-blue-500/5 rounded-full blur-[120px]" />
      <div class="absolute bottom-[-10%] left-[-10%] w-[50%] h-[50%] bg-purple-500/5 rounded-full blur-[120px]" />
    </div>

    <main class="relative z-10 mx-auto flex w-full max-w-6xl flex-col gap-5 px-4 py-6 sm:px-6 lg:px-8">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <RouterLink to="/lobby" class="mb-3 inline-flex items-center gap-2 text-xs font-bold text-slate-500 transition-colors hover:text-slate-900 dark:text-slate-400 dark:hover:text-white">
            <ArrowLeft class="h-4 w-4" />
            返回大厅
          </RouterLink>
          <h1 class="text-2xl font-black tracking-tight sm:text-3xl flex items-center gap-2">
            <Shield class="w-6 h-6 text-sky-500" />
            反作弊系统管理
          </h1>
          <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">检测记录管理、申诉处理及系统配置。</p>
        </div>
      </div>

      <!-- Tabs -->
      <div class="flex gap-2 overflow-x-auto border-b border-slate-200 pb-2 dark:border-white/10 custom-scrollbar">
        <button
          v-for="tab in ['detection', 'appeals', 'config', 'audit']"
          :key="tab"
          :class="['whitespace-nowrap rounded-lg px-4 py-2 text-sm font-bold transition-colors', activeTab === tab ? 'bg-sky-600 text-white shadow-sm dark:bg-sky-500' : 'text-slate-600 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-white/5']"
          @click="activeTab = tab as any"
        >
          {{ 
            tab === 'detection' ? '检测管理' :
            tab === 'appeals' ? '申诉管理' :
            tab === 'config' ? '配置管理' :
            '审计日志'
          }}
        </button>
      </div>

      <!-- Detection Tab -->
      <div v-if="activeTab === 'detection'" class="space-y-5 animate-in fade-in slide-in-from-bottom-4 duration-500">
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex-1 flex items-center gap-2 h-10 rounded-lg border border-slate-200 bg-white px-3 focus-within:border-sky-400 dark:border-white/10 dark:bg-[#111318]">
            <SearchIcon class="h-4 w-4 text-slate-400" />
            <input 
              v-model="detectionSearchTerm"
              type="text"
              placeholder="搜索玩家ID或房间ID"
              class="w-full bg-transparent text-sm text-slate-900 outline-none dark:text-white placeholder:text-slate-400"
              @input="detectionPage = 1"
            />
          </div>
          <select v-model="detectionStatusFilter" @change="detectionPage = 1" class="h-10 rounded-lg border border-slate-200 bg-white px-3 text-sm text-slate-900 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-[#111318] dark:text-white">
            <option value="all">所有状态</option>
            <option value="observe">观察</option>
            <option value="warning">警告</option>
            <option value="mute">禁言</option>
            <option value="ban">封号</option>
          </select>
        </div>

        <div class="rounded-lg border border-slate-200 bg-white shadow-sm dark:border-white/10 dark:bg-[#111318] overflow-x-auto">
          <table class="w-full text-left text-sm whitespace-nowrap">
            <thead class="border-b border-slate-200 bg-slate-50 dark:border-white/10 dark:bg-black/20">
              <tr>
                <th class="p-4 font-black">玩家ID</th>
                <th class="p-4 font-black">房间ID</th>
                <th class="p-4 font-black">风险分数</th>
                <th class="p-4 font-black">处罚类型</th>
                <th class="p-4 font-black">检测时间</th>
                <th class="p-4 font-black text-right">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-200 dark:divide-white/10">
              <tr v-if="filteredDetections.length === 0">
                <td colspan="6" class="p-8 text-center text-slate-400 font-bold">暂无检测记录</td>
              </tr>
              <tr v-for="detection in filteredDetections" :key="detection.id" class="hover:bg-slate-50 dark:hover:bg-white/5 transition-colors">
                <td class="p-4">{{ detection.player_id }}</td>
                <td class="p-4">{{ detection.room_id }}</td>
                <td class="p-4">
                  <span :class="['font-black', getRiskColor(detection.risk_score)]">
                    {{ detection.risk_score.toFixed(1) }}
                  </span>
                </td>
                <td class="p-4">
                  <span :class="['inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-black', getSanctionBadge(detection.sanction_type).color]">
                    {{ getSanctionBadge(detection.sanction_type).label }}
                  </span>
                </td>
                <td class="p-4 text-slate-500">{{ new Date(detection.created_at).toLocaleString('zh-CN') }}</td>
                <td class="p-4 text-right">
                  <button class="inline-flex items-center gap-1 rounded-lg bg-sky-50 px-3 py-1.5 text-xs font-black text-sky-600 transition-colors hover:bg-sky-100 dark:bg-sky-500/10 dark:text-sky-300 dark:hover:bg-sky-500/20" @click="openDetectionDetail(detection)">
                    <Eye class="h-3.5 w-3.5" />
                    查看
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div class="flex items-center justify-center gap-4 text-sm font-bold">
          <button :disabled="detectionPage === 1" @click="detectionPage--" class="flex items-center gap-1 rounded-lg border border-slate-200 bg-white px-3 py-1.5 transition-colors hover:bg-slate-50 disabled:opacity-50 dark:border-white/10 dark:bg-[#111318] dark:hover:bg-white/5">
            <ChevronLeft class="h-4 w-4" /> 上一页
          </button>
          <span class="text-slate-500">第 {{ detectionPage }} 页，共 {{ Math.ceil(detectionTotal / detectionLimit) || 1 }} 页</span>
          <button :disabled="detectionPage * detectionLimit >= detectionTotal" @click="detectionPage++" class="flex items-center gap-1 rounded-lg border border-slate-200 bg-white px-3 py-1.5 transition-colors hover:bg-slate-50 disabled:opacity-50 dark:border-white/10 dark:bg-[#111318] dark:hover:bg-white/5">
            下一页 <ChevronRight class="h-4 w-4" />
          </button>
        </div>
      </div>

      <!-- Appeals Tab -->
      <div v-if="activeTab === 'appeals'" class="space-y-5 animate-in fade-in slide-in-from-bottom-4 duration-500">
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex-1 flex items-center gap-2 h-10 rounded-lg border border-slate-200 bg-white px-3 focus-within:border-sky-400 dark:border-white/10 dark:bg-[#111318]">
            <SearchIcon class="h-4 w-4 text-slate-400" />
            <input 
              v-model="appealsSearchTerm"
              type="text"
              placeholder="搜索玩家ID或房间ID"
              class="w-full bg-transparent text-sm text-slate-900 outline-none dark:text-white placeholder:text-slate-400"
              @input="appealsPage = 1"
            />
          </div>
          <select v-model="appealsStatusFilter" @change="appealsPage = 1" class="h-10 rounded-lg border border-slate-200 bg-white px-3 text-sm text-slate-900 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-[#111318] dark:text-white">
            <option value="all">所有状态</option>
            <option value="pending">待审核</option>
            <option value="approved">已批准</option>
            <option value="rejected">已拒绝</option>
          </select>
        </div>

        <div class="rounded-lg border border-slate-200 bg-white shadow-sm dark:border-white/10 dark:bg-[#111318] overflow-x-auto">
          <table class="w-full text-left text-sm whitespace-nowrap">
            <thead class="border-b border-slate-200 bg-slate-50 dark:border-white/10 dark:bg-black/20">
              <tr>
                <th class="p-4 font-black">玩家ID</th>
                <th class="p-4 font-black max-w-[200px]">申诉理由</th>
                <th class="p-4 font-black">状态</th>
                <th class="p-4 font-black">申诉时间</th>
                <th class="p-4 font-black text-right">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-200 dark:divide-white/10">
              <tr v-if="filteredAppeals.length === 0">
                <td colspan="5" class="p-8 text-center text-slate-400 font-bold">暂无申诉记录</td>
              </tr>
              <tr v-for="appeal in filteredAppeals" :key="appeal.id" class="hover:bg-slate-50 dark:hover:bg-white/5 transition-colors">
                <td class="p-4">{{ appeal.player_id || appeal.player_uid }}</td>
                <td class="p-4 max-w-[200px] truncate" :title="appeal.reason">{{ appeal.reason }}</td>
                <td class="p-4">
                  <span :class="['inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-black', getStatusBadge(appeal.status).color]">
                    {{ getStatusBadge(appeal.status).label }}
                  </span>
                </td>
                <td class="p-4 text-slate-500">{{ new Date(appeal.created_at).toLocaleString('zh-CN') }}</td>
                <td class="p-4 text-right">
                  <div v-if="appeal.status === 'pending'" class="flex items-center justify-end gap-2">
                    <button class="inline-flex items-center gap-1 rounded-lg bg-emerald-50 px-3 py-1.5 text-xs font-black text-emerald-600 transition-colors hover:bg-emerald-100 dark:bg-emerald-500/10 dark:text-emerald-300 dark:hover:bg-emerald-500/20" @click="handleApproveAppeal(appeal.id)">
                      <CheckCircle class="h-3.5 w-3.5" /> 批准
                    </button>
                    <button class="inline-flex items-center gap-1 rounded-lg bg-rose-50 px-3 py-1.5 text-xs font-black text-rose-600 transition-colors hover:bg-rose-100 dark:bg-rose-500/10 dark:text-rose-300 dark:hover:bg-rose-500/20" @click="handleRejectAppeal(appeal.id)">
                      <XCircle class="h-3.5 w-3.5" /> 拒绝
                    </button>
                  </div>
                  <span v-else class="text-slate-400">-</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div class="flex items-center justify-center gap-4 text-sm font-bold">
          <button :disabled="appealsPage === 1" @click="appealsPage--" class="flex items-center gap-1 rounded-lg border border-slate-200 bg-white px-3 py-1.5 transition-colors hover:bg-slate-50 disabled:opacity-50 dark:border-white/10 dark:bg-[#111318] dark:hover:bg-white/5">
            <ChevronLeft class="h-4 w-4" /> 上一页
          </button>
          <span class="text-slate-500">第 {{ appealsPage }} 页，共 {{ Math.ceil(appealsTotal / appealsLimit) || 1 }} 页</span>
          <button :disabled="appealsPage * appealsLimit >= appealsTotal" @click="appealsPage++" class="flex items-center gap-1 rounded-lg border border-slate-200 bg-white px-3 py-1.5 transition-colors hover:bg-slate-50 disabled:opacity-50 dark:border-white/10 dark:bg-[#111318] dark:hover:bg-white/5">
            下一页 <ChevronRight class="h-4 w-4" />
          </button>
        </div>
      </div>

      <!-- Config Tab -->
      <div v-if="activeTab === 'config'" class="space-y-5 animate-in fade-in slide-in-from-bottom-4 duration-500">
        <div v-if="!editingConfig && configData" class="rounded-lg border border-slate-200 bg-white p-5 shadow-sm dark:border-white/10 dark:bg-[#111318]">
          <div class="mb-5 flex items-center justify-between">
            <h2 class="text-lg font-black">反作弊策略配置</h2>
            <button class="inline-flex items-center gap-2 rounded-lg bg-sky-600 px-4 py-2 text-sm font-black text-white transition-colors hover:bg-sky-500" @click="startEditConfig">
              <Settings class="h-4 w-4" /> 编辑配置
            </button>
          </div>
          <div class="grid gap-6 md:grid-cols-2">
            <!-- Details -->
            <div class="rounded-xl border border-slate-100 bg-slate-50 p-4 dark:border-white/5 dark:bg-black/20">
              <h3 class="mb-3 text-sm font-black text-slate-500 dark:text-slate-400">检测权重</h3>
              <div class="space-y-2 text-sm">
                <div class="flex justify-between py-1"><span class="text-slate-500">响应时间权重:</span> <span class="font-bold">{{ configData.dimensions?.response_time?.weight || 0.25 }}</span></div>
                <div class="flex justify-between py-1"><span class="text-slate-500">操作频率权重:</span> <span class="font-bold">{{ configData.dimensions?.frequency?.weight || 0.25 }}</span></div>
                <div class="flex justify-between py-1"><span class="text-slate-500">胜率异常权重:</span> <span class="font-bold">{{ configData.dimensions?.win_rate?.weight || 0.20 }}</span></div>
                <div class="flex justify-between py-1"><span class="text-slate-500">操作模式权重:</span> <span class="font-bold">{{ configData.dimensions?.pattern?.weight || 0.15 }}</span></div>
                <div class="flex justify-between py-1"><span class="text-slate-500">账号年龄权重:</span> <span class="font-bold">{{ configData.dimensions?.account_age?.weight || 0.15 }}</span></div>
              </div>
            </div>
            
            <div class="rounded-xl border border-slate-100 bg-slate-50 p-4 dark:border-white/5 dark:bg-black/20">
              <h3 class="mb-3 text-sm font-black text-slate-500 dark:text-slate-400">处罚阈值</h3>
              <div class="space-y-2 text-sm">
                <div class="flex justify-between py-1"><span class="text-slate-500">观察阈值:</span> <span class="font-bold">{{ configData.sanctions?.observe || 20 }}-40</span></div>
                <div class="flex justify-between py-1"><span class="text-slate-500">警告阈值:</span> <span class="font-bold">{{ configData.sanctions?.warning || 40 }}-60</span></div>
                <div class="flex justify-between py-1"><span class="text-slate-500">禁言阈值:</span> <span class="font-bold">{{ configData.sanctions?.mute || 60 }}-80</span></div>
                <div class="flex justify-between py-1"><span class="text-slate-500">封号阈值:</span> <span class="font-bold">{{ configData.sanctions?.ban || 80 }}-100</span></div>
              </div>
            </div>

            <div class="rounded-xl border border-slate-100 bg-slate-50 p-4 dark:border-white/5 dark:bg-black/20 md:col-span-2">
              <h3 class="mb-3 text-sm font-black text-slate-500 dark:text-slate-400">解封补偿</h3>
              <div class="space-y-2 text-sm">
                <div class="flex justify-between py-1"><span class="text-slate-500">补偿金额:</span> <span class="font-bold text-emerald-600">{{ configData.unban?.compensation_amount || 100 }} 燃素</span></div>
                <div class="py-1">
                  <span class="block text-slate-500 mb-2">默认文案:</span> 
                  <div class="rounded-lg border border-slate-200 bg-white p-3 text-slate-700 dark:border-white/10 dark:bg-[#111318] dark:text-slate-300">
                    {{ configData.unban?.default_message || '由于反作弊系统将您误封，在此，ChemistryUNO开发组向受到影响的研究员提供燃素补偿，感谢研究员对维护纯净游戏环境做出的贡献' }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-if="editingConfig && tempConfig" class="rounded-lg border border-slate-200 bg-white p-5 shadow-sm dark:border-white/10 dark:bg-[#111318]">
          <h2 class="mb-5 text-lg font-black">编辑配置</h2>
          
          <div class="grid gap-6 md:grid-cols-2">
            <!-- Edit Forms -->
            <div class="space-y-4">
              <h4 class="font-bold border-b border-slate-100 pb-2 dark:border-white/10">响应时间检测</h4>
              <label class="block text-sm font-bold text-slate-500">权重: <input v-model.number="tempConfig.dimensions.response_time.weight" type="number" step="0.01" min="0" max="1" class="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20" /></label>
              <label class="block text-sm font-bold text-slate-500">阈值 (ms): <input v-model.number="tempConfig.dimensions.response_time.threshold" type="number" class="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20" /></label>
            </div>

            <div class="space-y-4">
              <h4 class="font-bold border-b border-slate-100 pb-2 dark:border-white/10">操作频率检测</h4>
              <label class="block text-sm font-bold text-slate-500">权重: <input v-model.number="tempConfig.dimensions.frequency.weight" type="number" step="0.01" min="0" max="1" class="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20" /></label>
              <label class="block text-sm font-bold text-slate-500">阈值 (每10秒): <input v-model.number="tempConfig.dimensions.frequency.threshold" type="number" class="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20" /></label>
            </div>

            <div class="space-y-4">
              <h4 class="font-bold border-b border-slate-100 pb-2 dark:border-white/10">胜率异常检测</h4>
              <label class="block text-sm font-bold text-slate-500">权重: <input v-model.number="tempConfig.dimensions.win_rate.weight" type="number" step="0.01" min="0" max="1" class="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20" /></label>
              <label class="block text-sm font-bold text-slate-500">胜率阈值: <input v-model.number="tempConfig.dimensions.win_rate.threshold" type="number" step="0.01" min="0" max="1" class="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20" /></label>
            </div>

            <div class="space-y-4">
              <h4 class="font-bold border-b border-slate-100 pb-2 dark:border-white/10">处罚阈值</h4>
              <label class="block text-sm font-bold text-slate-500">观察下界: <input v-model.number="tempConfig.sanctions.observe" type="number" min="0" max="100" class="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20" /></label>
              <label class="block text-sm font-bold text-slate-500">警告下界: <input v-model.number="tempConfig.sanctions.warning" type="number" min="0" max="100" class="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20" /></label>
              <label class="block text-sm font-bold text-slate-500">禁言下界: <input v-model.number="tempConfig.sanctions.mute" type="number" min="0" max="100" class="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20" /></label>
              <label class="block text-sm font-bold text-slate-500">封号下界: <input v-model.number="tempConfig.sanctions.ban" type="number" min="0" max="100" class="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20" /></label>
            </div>
            
            <div class="space-y-4 md:col-span-2">
              <h4 class="font-bold border-b border-slate-100 pb-2 dark:border-white/10">解封补偿配置</h4>
              <label class="block text-sm font-bold text-slate-500">补偿金额（燃素）: <input v-model.number="tempConfig.unban.compensation_amount" type="number" min="0" step="1" class="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20" /></label>
              <label class="block text-sm font-bold text-slate-500">默认补偿文案: <textarea v-model="tempConfig.unban.default_message" rows="4" class="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20 resize-none"></textarea></label>
            </div>
          </div>

          <div class="mt-6 flex gap-3 justify-end">
            <button class="inline-flex items-center gap-2 rounded-lg bg-sky-600 px-4 py-2 text-sm font-black text-white hover:bg-sky-500" @click="saveConfig">
              <Save class="h-4 w-4" /> 保存配置
            </button>
            <button class="rounded-lg border border-slate-200 px-4 py-2 text-sm font-bold hover:bg-slate-50 dark:border-white/10 dark:hover:bg-white/5" @click="cancelEditConfig">取消</button>
          </div>
        </div>
      </div>

      <!-- Audit Log Tab -->
      <div v-if="activeTab === 'audit'" class="space-y-5 animate-in fade-in slide-in-from-bottom-4 duration-500">
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex-1 flex items-center gap-2 h-10 rounded-lg border border-slate-200 bg-white px-3 focus-within:border-sky-400 dark:border-white/10 dark:bg-[#111318]">
            <SearchIcon class="h-4 w-4 text-slate-400" />
            <input 
              v-model="auditSearchTerm"
              type="text"
              placeholder="搜索玩家ID"
              class="w-full bg-transparent text-sm text-slate-900 outline-none dark:text-white placeholder:text-slate-400"
            />
          </div>
          <input v-model="auditStartDate" type="date" class="h-10 rounded-lg border border-slate-200 bg-white px-3 text-sm text-slate-900 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-[#111318] dark:text-white" />
          <input v-model="auditEndDate" type="date" class="h-10 rounded-lg border border-slate-200 bg-white px-3 text-sm text-slate-900 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-[#111318] dark:text-white" />
          <button class="inline-flex items-center gap-2 h-10 rounded-lg bg-sky-600 px-4 text-sm font-black text-white hover:bg-sky-500" @click="loadAuditLog">
            <Filter class="h-4 w-4" /> 查询
          </button>
          <button class="inline-flex items-center gap-2 h-10 rounded-lg border border-slate-200 bg-white px-4 text-sm font-bold hover:bg-slate-50 dark:border-white/10 dark:bg-[#111318] dark:hover:bg-white/5" @click="exportAuditLog">
            <Download class="h-4 w-4" /> 导出
          </button>
        </div>

        <div class="rounded-lg border border-slate-200 bg-white shadow-sm dark:border-white/10 dark:bg-[#111318] overflow-x-auto">
          <table class="w-full text-left text-sm whitespace-nowrap">
            <thead class="border-b border-slate-200 bg-slate-50 dark:border-white/10 dark:bg-black/20">
              <tr>
                <th class="p-4 font-black">玩家ID</th>
                <th class="p-4 font-black">操作类型</th>
                <th class="p-4 font-black max-w-[300px]">详情</th>
                <th class="p-4 font-black">补偿状态</th>
                <th class="p-4 font-black">时间</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-200 dark:divide-white/10">
              <tr v-if="auditLogs.length === 0">
                <td colspan="5" class="p-8 text-center text-slate-400 font-bold">暂无审计日志</td>
              </tr>
              <tr v-for="log in auditLogs" :key="log.id" class="hover:bg-slate-50 dark:hover:bg-white/5 transition-colors">
                <td class="p-4">{{ log.player_id }}</td>
                <td class="p-4 font-medium">{{ log.action_type }}</td>
                <td class="p-4 max-w-[300px] truncate" :title="log.details">{{ log.details }}</td>
                <td class="p-4">
                  <div v-if="log.compensation_status" class="flex flex-col gap-1">
                    <span :class="['inline-flex w-fit items-center rounded-full px-2 py-0.5 text-[10px] font-black uppercase tracking-widest', getCompensationBadge(log.compensation_status).color]">
                      {{ getCompensationBadge(log.compensation_status).label }}
                    </span>
                    <span v-if="log.compensation_amount" class="text-[10px] text-emerald-600 dark:text-emerald-400 font-bold">
                      {{ log.compensation_amount }}燃素
                    </span>
                  </div>
                  <span v-else class="text-slate-400">-</span>
                </td>
                <td class="p-4 text-slate-500">{{ new Date(log.created_at).toLocaleString('zh-CN') }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div class="flex items-center justify-center gap-4 text-sm font-bold">
          <button :disabled="auditPage === 1" @click="auditPage--" class="flex items-center gap-1 rounded-lg border border-slate-200 bg-white px-3 py-1.5 transition-colors hover:bg-slate-50 disabled:opacity-50 dark:border-white/10 dark:bg-[#111318] dark:hover:bg-white/5">
            <ChevronLeft class="h-4 w-4" /> 上一页
          </button>
          <span class="text-slate-500">第 {{ auditPage }} 页，共 {{ Math.ceil(auditTotal / auditLimit) || 1 }} 页</span>
          <button :disabled="auditPage * auditLimit >= auditTotal" @click="auditPage++" class="flex items-center gap-1 rounded-lg border border-slate-200 bg-white px-3 py-1.5 transition-colors hover:bg-slate-50 disabled:opacity-50 dark:border-white/10 dark:bg-[#111318] dark:hover:bg-white/5">
            下一页 <ChevronRight class="h-4 w-4" />
          </button>
        </div>
      </div>
    </main>

    <!-- Modals -->
    <Teleport to="body">
      <!-- Detection Detail Modal -->
      <div v-if="showDetailModal" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 p-4 backdrop-blur-sm" @click.self="showDetailModal = false">
        <div class="w-full max-w-2xl max-h-[90vh] overflow-y-auto custom-scrollbar rounded-2xl bg-white p-6 shadow-2xl dark:bg-[#111318] border border-slate-200 dark:border-white/10 animate-in zoom-in-95 duration-200">
          <h2 class="mb-6 text-xl font-black">检测详情</h2>
          <div v-if="selectedDetection" class="space-y-6">
            <div class="grid grid-cols-2 gap-4 rounded-xl border border-slate-100 bg-slate-50 p-4 dark:border-white/5 dark:bg-black/20">
              <div><span class="text-xs font-bold uppercase tracking-widest text-slate-500 block mb-1">玩家ID</span><span class="font-medium">{{ selectedDetection.player_id }}</span></div>
              <div><span class="text-xs font-bold uppercase tracking-widest text-slate-500 block mb-1">房间ID</span><span class="font-medium">{{ selectedDetection.room_id }}</span></div>
              <div><span class="text-xs font-bold uppercase tracking-widest text-slate-500 block mb-1">风险分数</span><span :class="['font-black', getRiskColor(selectedDetection.risk_score)]">{{ selectedDetection.risk_score.toFixed(1) }}</span></div>
              <div><span class="text-xs font-bold uppercase tracking-widest text-slate-500 block mb-1">处罚类型</span><span :class="['inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-black', getSanctionBadge(selectedDetection.sanction_type).color]">{{ getSanctionBadge(selectedDetection.sanction_type).label }}</span></div>
            </div>

            <div>
              <h3 class="mb-3 text-sm font-black border-b border-slate-100 pb-2 dark:border-white/10">维度分数</h3>
              <div class="grid grid-cols-2 gap-3 text-sm">
                <div class="flex justify-between rounded-lg bg-slate-50 px-3 py-2 dark:bg-black/20"><span class="text-slate-500">响应时间</span><span class="font-bold">{{ selectedDetection.response_time_score?.toFixed(1) || 'N/A' }}</span></div>
                <div class="flex justify-between rounded-lg bg-slate-50 px-3 py-2 dark:bg-black/20"><span class="text-slate-500">操作频率</span><span class="font-bold">{{ selectedDetection.frequency_score?.toFixed(1) || 'N/A' }}</span></div>
                <div class="flex justify-between rounded-lg bg-slate-50 px-3 py-2 dark:bg-black/20"><span class="text-slate-500">胜率异常</span><span class="font-bold">{{ selectedDetection.win_rate_score?.toFixed(1) || 'N/A' }}</span></div>
                <div class="flex justify-between rounded-lg bg-slate-50 px-3 py-2 dark:bg-black/20"><span class="text-slate-500">操作模式</span><span class="font-bold">{{ selectedDetection.pattern_score?.toFixed(1) || 'N/A' }}</span></div>
                <div class="flex justify-between rounded-lg bg-slate-50 px-3 py-2 dark:bg-black/20"><span class="text-slate-500">账号年龄</span><span class="font-bold">{{ selectedDetection.account_age_score?.toFixed(1) || 'N/A' }}</span></div>
              </div>
            </div>

            <div>
              <h3 class="mb-3 text-sm font-black border-b border-slate-100 pb-2 dark:border-white/10">人工审核</h3>
              <div class="space-y-3">
                <label class="block text-sm font-bold text-slate-500">审核决策:
                  <select v-model="reviewDecision" class="mt-1 block w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20">
                    <option value="confirm">确认处罚</option>
                    <option value="override">推翻处罚</option>
                  </select>
                </label>
                <label class="block text-sm font-bold text-slate-500">审核备注:
                  <textarea v-model="reviewNote" rows="3" class="mt-1 block w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20 resize-none" placeholder="输入审核备注..."></textarea>
                </label>
              </div>
            </div>

            <div>
              <h3 class="mb-3 text-sm font-black border-b border-slate-100 pb-2 dark:border-white/10 text-rose-600">封禁处置</h3>
              <div class="space-y-3">
                <label class="block text-sm font-bold text-slate-500">封禁截止时间:
                  <input v-model="enforcementUntil" type="datetime-local" class="mt-1 block w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-rose-400 dark:border-white/10 dark:bg-black/20" />
                </label>
                <label class="block text-sm font-bold text-slate-500">处置原因:
                  <textarea v-model="enforcementReason" rows="2" class="mt-1 block w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-rose-400 dark:border-white/10 dark:bg-black/20 resize-none" placeholder="请输入封禁或解封原因"></textarea>
                </label>
                <div class="flex gap-2 justify-end pt-2">
                  <button class="rounded-lg bg-rose-600 px-4 py-2 text-sm font-black text-white hover:bg-rose-500 disabled:opacity-50" :disabled="enforcementLoading" @click="handlePanelBan">执行封禁</button>
                  <button class="rounded-lg bg-slate-200 px-4 py-2 text-sm font-black text-slate-700 hover:bg-slate-300 disabled:opacity-50 dark:bg-white/10 dark:text-slate-300 dark:hover:bg-white/20" :disabled="enforcementLoading" @click="handlePanelUnban">解除封禁</button>
                </div>
              </div>
            </div>
          </div>
          <div class="mt-8 flex gap-3 justify-end border-t border-slate-100 pt-4 dark:border-white/10">
            <button class="inline-flex items-center gap-2 rounded-lg bg-sky-600 px-4 py-2 text-sm font-black text-white hover:bg-sky-500" @click="submitReview">
              <Save class="h-4 w-4" /> 提交审核
            </button>
            <button class="rounded-lg border border-slate-200 px-4 py-2 text-sm font-bold hover:bg-slate-50 dark:border-white/10 dark:hover:bg-white/5" @click="showDetailModal = false">关闭</button>
          </div>
        </div>
      </div>

      <!-- Approval Modal with Compensation -->
      <div v-if="showApprovalModal" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 p-4 backdrop-blur-sm" @click.self="cancelApproval">
        <div class="w-full max-w-lg rounded-2xl bg-white p-6 shadow-2xl dark:bg-[#111318] border border-slate-200 dark:border-white/10 animate-in zoom-in-95 duration-200">
          <h2 class="mb-6 text-xl font-black">批准申诉并发放补偿</h2>
          
          <div class="space-y-6">
            <div class="rounded-xl border border-emerald-100 bg-emerald-50 p-4 dark:border-emerald-500/10 dark:bg-emerald-500/5">
              <h3 class="mb-2 text-sm font-bold text-emerald-700 dark:text-emerald-400">向玩家发送的消息</h3>
              <div class="rounded-lg bg-white p-3 text-sm text-slate-700 dark:bg-black/20 dark:text-slate-300 border border-emerald-100 dark:border-emerald-500/20">
                {{ compensationMessage }}
              </div>
              <h3 class="mt-4 mb-2 text-sm font-bold text-emerald-700 dark:text-emerald-400">补偿数额</h3>
              <div class="flex items-center gap-1 font-black text-emerald-600 dark:text-emerald-400">
                <span class="text-2xl">{{ compensationAmount }}</span>
                <span class="text-sm">燃素</span>
              </div>
            </div>

            <details class="group rounded-xl border border-slate-200 bg-slate-50 dark:border-white/10 dark:bg-black/20">
              <summary class="cursor-pointer p-4 text-sm font-bold text-slate-600 outline-none dark:text-slate-300 flex items-center justify-between">
                调整补偿配置
                <ChevronRight class="h-4 w-4 transition-transform group-open:rotate-90" />
              </summary>
              <div class="border-t border-slate-200 p-4 space-y-4 dark:border-white/10">
                <label class="block text-sm font-bold text-slate-500">补偿金额（燃素）:
                  <div class="flex items-center gap-2 mt-1">
                    <input v-model.number="compensationAmount" type="number" min="0" step="1" class="flex-1 rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/40" />
                    <span class="text-xs text-slate-400">默认: 100</span>
                  </div>
                </label>
                <label class="block text-sm font-bold text-slate-500">补偿文案:
                  <textarea v-model="compensationMessage" rows="3" class="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/40 resize-none" placeholder="输入自定义补偿文案..."></textarea>
                </label>
                <button class="text-xs font-bold text-sky-600 hover:text-sky-500 dark:text-sky-400" @click="compensationMessage = defaultCompensationMessage">恢复默认文案</button>
              </div>
            </details>

            <label class="block text-sm font-bold text-slate-500">审核备注（可选）:
              <textarea v-model="approvalNote" rows="2" class="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20 resize-none" placeholder="输入审核备注..."></textarea>
            </label>
          </div>

          <div class="mt-8 flex gap-3 justify-end border-t border-slate-100 pt-4 dark:border-white/10">
            <button class="inline-flex items-center gap-2 rounded-lg bg-emerald-600 px-4 py-2 text-sm font-black text-white hover:bg-emerald-500" @click="confirmApproval">
              <CheckCircle class="h-4 w-4" /> 确认批准
            </button>
            <button class="rounded-lg border border-slate-200 px-4 py-2 text-sm font-bold hover:bg-slate-50 dark:border-white/10 dark:hover:bg-white/5" @click="cancelApproval">取消</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>


