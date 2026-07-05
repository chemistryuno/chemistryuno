<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { pageClassNames } from '@lib'
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
  Gavel,
  Ban,
  AlertTriangle,
  VolumeX,
  UserCheck,
  Clock,
  FileText,
  ListChecks,
  RefreshCw,
} from 'lucide-vue-next'

type ReplayEvidenceAnchor = {
  room_id?: string
  game_history_id?: number
  replay_id?: string
  event_index?: number
  event_id?: string
  event_type?: string
  player_uid?: number
  event_timestamp_ms?: number
  turn_number?: number
  action_summary?: string
  evidence_precision?: string
  compatibility_level?: string
  navigation_url?: string
  has_replay?: boolean
  replay_available?: boolean
}

const { showAlert, showPrompt, showConfirm } = useDialog()
const activeTab = ref<'detection' | 'appeals' | 'config' | 'audit'>('detection')
const loading = ref(false)
const operating = ref(false)

// ==================== 快速查询 ====================
const quickPlayerId = ref('')
const quickSearching = ref(false)
const quickSearchResult = ref<any>(null)

const quickSearchPlayer = async () => {
  if (!quickPlayerId.value.trim()) return
  quickSearching.value = true
  quickSearchResult.value = null
  try {
    // 使用审核日志查询玩家信息
    const res = await adminAPI.getAuditLog({
      player_id: quickPlayerId.value.trim(),
      limit: 5,
    })
    const logs = res.data?.logs || []
    if (logs.length > 0) {
      quickSearchResult.value = { player_id: quickPlayerId.value.trim(), logs }
    } else {
      quickSearchResult.value = { player_id: quickPlayerId.value.trim(), logs: [], no_data: true }
    }
  } catch (e: any) {
    showAlert(e.response?.data?.error || '查询失败', '错误')
  } finally {
    quickSearching.value = false
  }
}

// ==================== 快速处置 ====================
const showQuickBanModal = ref(false)
const quickBanTarget = ref({ uid: 0, reason: '', until: defaultBanUntil() })

function defaultBanUntil() {
  const d = new Date(Date.now() + 24 * 60 * 60 * 1000)
  return d.toISOString().slice(0, 16)
}

const openQuickBan = (uid: number) => {
  quickBanTarget.value = { uid, reason: '反作弊系统检测到异常行为', until: defaultBanUntil() }
  showQuickBanModal.value = true
}

const executeQuickBan = async () => {
  if (!quickBanTarget.value.uid) return
  const until = new Date(quickBanTarget.value.until)
  if (Number.isNaN(until.getTime()) || until <= new Date()) {
    showAlert('封禁截止时间必须晚于当前时间', '时间无效')
    return
  }
  operating.value = true
  try {
    await adminAPI.banFromAnticheatPanel({
      player_uid: quickBanTarget.value.uid,
      banned_until: until.toISOString(),
      reason: quickBanTarget.value.reason.trim() || '反作弊系统检测到异常行为',
    })
    showAlert('封禁已执行', '操作完成')
    showQuickBanModal.value = false
    loadDetections()
    loadAuditLog()
  } catch (e: any) {
    showAlert(e.response?.data?.error || '封禁失败', '错误')
  } finally {
    operating.value = false
  }
}

// ==================== 检测列表 ====================
const detectionList = ref<any[]>([])
const detectionSearchTerm = ref('')
const detectionStatusFilter = ref<'all' | 'observe' | 'warning' | 'mute' | 'ban'>('all')
const detectionPage = ref(1)
const detectionLimit = ref(20)
const detectionTotal = ref(0)
const archivedIds = ref<Set<number>>(new Set())
const showArchived = ref(false)

const archiveDetection = (id: number) => {
  archivedIds.value = new Set([...archivedIds.value, id])
}

// ==================== 批量处置 ====================
const selectedDetectionIds = ref<Set<number | string>>(new Set())
const batchReviewLoading = ref(false)

const toggleDetectionSelected = (id: number | string) => {
  const next = new Set(selectedDetectionIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  selectedDetectionIds.value = next
}

const clearDetectionSelection = () => {
  selectedDetectionIds.value = new Set()
}

const batchReviewSelected = async () => {
  const ids = Array.from(selectedDetectionIds.value)
  if (ids.length === 0) return
  batchReviewLoading.value = true
  try {
    const res = await adminAPI.batchReviewDetections({ ids, decision: 'confirm', remark: '批量处置' })
    const data = res.data || {}
    showAlert(`批量处置完成：成功 ${data.success ?? 0} 条，失败 ${data.failed ?? 0} 条`, '完成')
    // Archive successfully processed entries and refresh.
    for (const item of (data.results || [])) {
      if (item.success) archiveDetection(Number(item.id))
    }
    clearDetectionSelection()
    await loadDetections()
  } catch (error: any) {
    showAlert(error.response?.data?.error || '批量处置失败', '错误')
  } finally {
    batchReviewLoading.value = false
  }
}

const filteredDetections = computed(() => {
  let items = detectionList.value
  if (!showArchived.value) {
    items = items.filter(d => !archivedIds.value.has(d.id))
  }
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

// ==================== 树形分组 ====================
const expandedRooms = ref<Record<string, boolean>>({})
const expandedPlayers = ref<Record<string, boolean>>({})

const toggleRoom = (roomId: string) => {
  expandedRooms.value[roomId] = !expandedRooms.value[roomId]
}

const togglePlayer = (playerKey: string) => {
  expandedPlayers.value[playerKey] = !expandedPlayers.value[playerKey]
}

const groupedDetections = computed(() => {
  const groups: Record<string, {
    roomId: string
    players: Record<string, {
      uid: number
      detections: any[]
    }>
    allDetections: any[]
  }> = {}

  for (const d of filteredDetections.value) {
    const roomId = d.room_id || 'unknown'
    if (!groups[roomId]) {
      groups[roomId] = { roomId, players: {}, allDetections: [] }
    }
    groups[roomId].allDetections.push(d)

    const uid = d.player_id ?? d.player_uid ?? 0
    const playerKey = String(uid)
    if (!groups[roomId].players[playerKey]) {
      groups[roomId].players[playerKey] = { uid, detections: [] }
    }
    groups[roomId].players[playerKey].detections.push(d)
  }

  // Sort rooms by newest detection, players by highest avg risk
  const sorted = Object.values(groups).sort((a, b) => {
    const aMax = Math.max(...a.allDetections.map(d => new Date(d.created_at).getTime()))
    const bMax = Math.max(...b.allDetections.map(d => new Date(d.created_at).getTime()))
    return bMax - aMax
  })

  for (const room of sorted) {
    const playerList = Object.values(room.players)
    playerList.sort((a, b) => {
      const aAvg = a.detections.reduce((s, d) => s + (d.risk_score || 0), 0) / a.detections.length
      const bAvg = b.detections.reduce((s, d) => s + (d.risk_score || 0), 0) / b.detections.length
      return bAvg - aAvg
    })
    room.players = Object.fromEntries(playerList.map(p => [String(p.uid), p]))
  }

  return sorted
})

const roomRiskSummary = (detections: any[]) => {
  const scores = detections.map(d => d.risk_score || 0)
  const max = Math.max(...scores)
  const avg = scores.reduce((a, b) => a + b, 0) / scores.length
  return { max, avg }
}

const uniquePlayersInRoom = (detections: any[]) => {
  const uids = new Set(detections.map(d => d.player_id ?? d.player_uid))
  return uids.size
}

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

// ==================== 检测详情 ====================
const showDetailModal = ref(false)
const selectedDetection = ref<any>(null)
const reviewNote = ref('')
const enforcementReason = ref('反作弊面板人工处置')
const enforcementUntil = ref('')
const enforcementLoading = ref(false)
const punishmentDecision = ref<'observe' | 'warning' | 'mute' | 'ban'>('ban')
const punishmentChangeLoading = ref(false)

const decodeJSONMaybe = (value: any) => {
  if (typeof value !== 'string') return value
  const trimmed = value.trim()
  if (!trimmed || (!trimmed.startsWith('{') && !trimmed.startsWith('['))) return value
  try { return JSON.parse(trimmed) } catch { return value }
}

const normalizeAnchor = (value: any): ReplayEvidenceAnchor | null => {
  const anchor = decodeJSONMaybe(value)
  if (!anchor || typeof anchor !== 'object' || Array.isArray(anchor)) return null
  if (!anchor.game_history_id && !anchor.replay_id && !anchor.room_id) return null
  return {
    ...anchor,
    event_index: Number(anchor.event_index || 0) || undefined,
    game_history_id: Number(anchor.game_history_id || 0) || undefined,
    player_uid: Number(anchor.player_uid || 0) || undefined,
    event_timestamp_ms: Number(anchor.event_timestamp_ms || 0) || undefined,
  }
}

const anchorKey = (anchor: ReplayEvidenceAnchor) =>
  [anchor.game_history_id || '', anchor.replay_id || '', anchor.event_id || '', anchor.event_index || '', anchor.event_timestamp_ms || ''].join(':')

const normalizeAnchorList = (value: any): ReplayEvidenceAnchor[] => {
  const decoded = decodeJSONMaybe(value)
  if (!decoded) return []
  if (Array.isArray(decoded)) return decoded.map(normalizeAnchor).filter(Boolean) as ReplayEvidenceAnchor[]
  const one = normalizeAnchor(decoded)
  return one ? [one] : []
}

const anchorsFromIndicator = (indicator: any): ReplayEvidenceAnchor[] => {
  const anchors = normalizeAnchorList(indicator?.evidence_anchors)
  if (anchors.length) return anchors
  const anchor = normalizeAnchor(indicator?.primary_evidence || indicator?.replay_anchor)
  return anchor ? [anchor] : []
}

const anchorsFromReportContribution = (report: any): ReplayEvidenceAnchor[] =>
  normalizeAnchorList(report?.evidence_anchors || report?.anchors || report?.primary_evidence)

const primaryEvidenceAnchor = computed(() =>
  normalizeAnchor(selectedDetection.value?.primary_evidence || selectedDetection.value?.replay_navigation)
)

const relatedEvidenceAnchors = computed(() => {
  const anchors = normalizeAnchorList(selectedDetection.value?.related_evidence)
  const primary = primaryEvidenceAnchor.value
  if (primary && !anchors.some(a => anchorKey(a) === anchorKey(primary))) {
    anchors.unshift(primary)
  }
  return anchors
})

const parseReplayTargetID = (anchor: ReplayEvidenceAnchor) => {
  const rawTarget = anchor.game_history_id || anchor.replay_id
  if (rawTarget == null || rawTarget === '') return ''
  const target = String(rawTarget).trim()
  return /^\d+$/.test(target) ? target : ''
}

const hasReplayForAnchor = (anchorLike: any) => {
  const anchor = normalizeAnchor(anchorLike)
  if (!anchor) return false
  if (anchor.has_replay === false || anchor.replay_available === false) return false
  return Boolean(parseReplayTargetID(anchor))
}

const replayRouteForAnchor = (anchorLike: any) => {
  const anchor = normalizeAnchor(anchorLike)
  if (!anchor || !hasReplayForAnchor(anchor)) return ''
  const replayTarget = parseReplayTargetID(anchor)
  if (!replayTarget) return ''
  const query = new URLSearchParams()
  query.set('scope', 'admin')
  query.set('from', '/admin/anticheat')
  if (anchor.event_index) query.set('event_index', String(anchor.event_index))
  if (anchor.event_id) query.set('event_id', anchor.event_id)
  if (anchor.event_timestamp_ms) query.set('timestamp_ms', String(anchor.event_timestamp_ms))
  if (anchor.player_uid) query.set('uid', String(anchor.player_uid))
  return `/replay/${replayTarget}?${query.toString()}`
}

const formatAnchorPosition = (anchorLike: any) => {
  const anchor = normalizeAnchor(anchorLike)
  if (!anchor) return '-'
  const parts = []
  if (anchor.event_index) parts.push(`#${anchor.event_index}`)
  if (anchor.event_id) parts.push(anchor.event_id)
  if (!parts.length && anchor.event_timestamp_ms) parts.push(`${anchor.event_timestamp_ms}ms`)
  return parts.join(' / ') || '房间级'
}

const formatAnchorTime = (anchorLike: any) => {
  const anchor = normalizeAnchor(anchorLike)
  if (!anchor?.event_timestamp_ms) return '-'
  const date = new Date(anchor.event_timestamp_ms)
  return Number.isNaN(date.getTime()) ? '-' : date.toLocaleString('zh-CN')
}

const precisionLabel = (precision?: string) => {
  if (precision === 'operation') return '操作级'
  if (precision === 'room') return '房间级'
  return precision || '未知'
}

const normalizeDetectionDetail = (payload: any, fallback: any) => {
  const risk = payload?.risk_score || payload || {}
  const sanctions = payload?.sanctions || []
  return {
    ...fallback, ...risk,
    id: risk.id || fallback.id,
    player_id: risk.player_id || risk.player_uid || fallback.player_id,
    player_uid: risk.player_uid || risk.player_id || fallback.player_uid || fallback.player_id,
    room_id: risk.room_id || fallback.room_id,
    risk_score: risk.risk_score ?? fallback.risk_score ?? 0,
    sanction_type: fallback.sanction_type || sanctions[0]?.sanction_type || 'observe',
    suggested_action: risk.suggested_action || fallback.suggested_action || fallback.sanction_type || 'observe',
    review_status: risk.review_status || fallback.review_status || 'pending',
    punishment_decision: risk.punishment_decision || fallback.punishment_decision || fallback.sanction_type || 'observe',
    sanction_id: risk.sanction_id || fallback.sanction_id || sanctions[0]?.id,
    replay_id: risk.replay_id || fallback.replay_id || '',
    operation_index: risk.operation_index ?? fallback.operation_index ?? 0,
    operation_timestamp: risk.operation_timestamp || fallback.operation_timestamp,
    indicator_details: risk.indicator_details || fallback.indicator_details || [],
    report_contribution: risk.report_contribution || fallback.report_contribution || null,
    primary_evidence: risk.primary_evidence || fallback.primary_evidence || risk.replay_navigation || fallback.replay_navigation || null,
    related_evidence: risk.related_evidence || fallback.related_evidence || [],
    replay_navigation: risk.replay_navigation || fallback.replay_navigation || risk.primary_evidence || fallback.primary_evidence || null,
    sanctions,
  }
}

const openDetectionDetail = async (detection: any) => {
  try {
    const response = await adminAPI.getDetectionDetail(detection.id)
    selectedDetection.value = normalizeDetectionDetail(response.data, detection)
    enforcementReason.value = '反作弊面板人工处置'
    enforcementUntil.value = defaultBanUntil()
    punishmentDecision.value = (selectedDetection.value.punishment_decision || selectedDetection.value.suggested_action || 'ban') as any
    showDetailModal.value = true
  } catch (error: any) {
    showAlert(error.response?.data?.error || '加载检测详情失败', '错误')
  }
}

const changePunishmentDecision = async () => {
  if (!selectedDetection.value) return
  punishmentChangeLoading.value = true
  try {
    await adminAPI.changeDetectionPunishment(selectedDetection.value.id, {
      punishment_decision: punishmentDecision.value,
      sanction_id: Number(selectedDetection.value.sanction_id) || undefined,
      reason: enforcementReason.value.trim() || '审核后调整处罚决定',
    })
    showAlert('处罚决定已更新', '操作完成')
    archiveDetection(selectedDetection.value.id)
    await Promise.all([loadDetections(), loadAuditLog()])
    selectedDetection.value.punishment_decision = punishmentDecision.value
  } catch (error: any) {
    showAlert(error.response?.data?.error || '处罚决定更新失败', '操作失败')
  } finally {
    punishmentChangeLoading.value = false
  }
}

const submitReview = async () => {
  if (!selectedDetection.value) return
  const id = selectedDetection.value.id
  try {
    await adminAPI.reviewDetection(id, {
      decision: 'confirm',
      note: reviewNote.value,
    })
    showAlert('审核已提交', '成功')
    archiveDetection(id)
    showDetailModal.value = false
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
    archiveDetection(selectedDetection.value.id)
    await Promise.all([loadDetections(), loadAuditLog()])
  } catch (error: any) {
    showAlert(error.response?.data?.error || '封禁失败', '操作失败')
  } finally {
    enforcementLoading.value = false
  }
}

// ==================== 申诉管理 ====================
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

// ==================== 申诉审批 ====================
const showApprovalModal = ref(false)
const pendingAppealId = ref<string>('')
const approvalNote = ref('')
const compensationAmount = ref(100)
const compensationMessage = ref('')

const defaultCompensationMessage = '由于反作弊系统将您误封，ChemistryUNO开发组向您提供燃素补偿，感谢您对维护纯净游戏环境做出的贡献'

const handleApproveAppeal = async (appealId: string) => {
  pendingAppealId.value = appealId
  approvalNote.value = ''
  compensationAmount.value = 100
  compensationMessage.value = defaultCompensationMessage
  showApprovalModal.value = true
}

const confirmApproval = async () => {
  if (!pendingAppealId.value) return
  try {
    await adminAPI.approveAppeal(pendingAppealId.value, {
      note: approvalNote.value || '通过审核',
      compensation_amount: compensationAmount.value,
      compensation_message: compensationMessage.value || defaultCompensationMessage,
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

// ==================== 配置管理 ====================
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

// ==================== 规则离线测试（沙盒） ====================
const ruleTestLoading = ref(false)
const ruleTestResult = ref<any>(null)
const ruleTestSampleLimit = ref(100)

const runRuleTest = async () => {
  ruleTestLoading.value = true
  ruleTestResult.value = null
  try {
    // 使用当前正在编辑的草拟配置（若无则后端回退到线上配置）。
    const draft = tempConfig.value ? JSON.parse(JSON.stringify(tempConfig.value)) : undefined
    const res = await adminAPI.runRuleTest({ draft, sample_limit: ruleTestSampleLimit.value, note: 'admin panel rule test' })
    ruleTestResult.value = res.data?.result || null
  } catch (error: any) {
    showAlert(error.response?.data?.error || '规则测试失败', '错误')
  } finally {
    ruleTestLoading.value = false
  }
}

// ==================== 审计日志 ====================
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
    link.setAttribute('download', `反作弊审计日志_${new Date().toISOString().split('T')[0]}.xlsx`)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  } catch (error: any) {
    showAlert(error.response?.data?.error || '导出日志失败', '错误')
  }
}

// ==================== 生命周期 ====================
onMounted(loadDetections)

watch(activeTab, (newTab) => {
  if (newTab === 'detection') loadDetections()
  else if (newTab === 'appeals') loadAppeals()
  else if (newTab === 'config') loadConfig()
  else if (newTab === 'audit') loadAuditLog()
})

// ==================== 工具函数 ====================
const getRiskColor = (score: number) => {
  if (score < 20) return 'text-green-600'
  if (score < 40) return 'text-blue-600'
  if (score < 60) return 'text-yellow-600'
  if (score < 80) return 'text-orange-600'
  return 'text-red-600'
}

const getRiskBg = (score: number) => {
  if (score < 20) return 'bg-green-500/10 text-green-600 dark:text-green-400'
  if (score < 40) return 'bg-blue-500/10 text-blue-600 dark:text-blue-400'
  if (score < 60) return 'bg-yellow-500/10 text-yellow-600 dark:text-yellow-400'
  if (score < 80) return 'bg-orange-500/10 text-orange-600 dark:text-orange-400'
  return 'bg-red-500/10 text-red-600 dark:text-red-400'
}

const getRiskLevel = (score: number) => {
  if (score < 20) return '正常'
  if (score < 40) return '低风险'
  if (score < 60) return '中风险'
  if (score < 80) return '高风险'
  return '严重'
}

const getSanctionBadge = (type: string) => {
  const badges: Record<string, { color: string; label: string }> = {
    observe: { color: 'bg-blue-100 text-blue-800 dark:bg-blue-500/10 dark:text-blue-400', label: '观察' },
    warning: { color: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-500/10 dark:text-yellow-400', label: '警告' },
    mute: { color: 'bg-orange-100 text-orange-800 dark:bg-orange-500/10 dark:text-orange-400', label: '禁言' },
    ban: { color: 'bg-red-100 text-red-800 dark:bg-red-500/10 dark:text-red-400', label: '封号' },
  }
  return badges[type] || { color: 'bg-gray-100 text-gray-800', label: type }
}

const formatTime = (value?: string) => {
  if (!value) return '-'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '-' : date.toLocaleString('zh-CN')
}

const getStatusBadge = (status: string) => {
  const badges: Record<string, { color: string; label: string }> = {
    pending: { color: 'bg-gray-100 text-gray-800 dark:bg-white/10 dark:text-gray-300', label: '待审核' },
    approved: { color: 'bg-green-100 text-green-800 dark:bg-green-500/10 dark:text-green-400', label: '已批准' },
    rejected: { color: 'bg-red-100 text-red-800 dark:bg-red-500/10 dark:text-red-400', label: '已拒绝' },
  }
  return badges[status] || { color: 'bg-gray-100 text-gray-800', label: status }
}

const getCompensationBadge = (status: string) => {
  const badges: Record<string, { color: string; label: string }> = {
    pending: { color: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-500/10 dark:text-yellow-400', label: '待发放' },
    ok: { color: 'bg-green-100 text-green-800 dark:bg-green-500/10 dark:text-green-400', label: '已发放' },
    failed: { color: 'bg-red-100 text-red-800 dark:bg-red-500/10 dark:text-red-400', label: '发放失败' },
  }
  return badges[status] || { color: 'bg-gray-100 text-gray-800', label: status }
}

const translateIndicatorName = (name: string) => {
  const map: Record<string, string> = {
    response_time: '响应时间',
    frequency: '操作频率',
    win_rate: '胜率异常',
    pattern: '操作模式',
    account_age: '账号年龄',
    variance: '方差分析',
    optimality: '出牌最优度',
    draw_randomness: '抽牌随机性',
    rhythm: '回合节奏',
    collusion: '合谋检测',
    multi_account: '多开检测',
    timestamp: '时间戳校验',
    report_contribution: '举报贡献',
  }
  return map[name] || map[name.toLowerCase()] || name
}

const translateEventType = (type: string) => {
  const map: Record<string, string> = {
    game_start: '对局开始',
    play_card: '出牌',
    double_play: '双联',
    draw_card: '摸牌',
    timeout_auto_draw: '超时自动摸牌',
    game_finished: '对局结束',
    game_terminated_invalid: '无效结算',
    game_terminated_all_humans_exited: '玩家全部退出',
    fast_reaction: '快速反应',
    room: '房间',
    report: '举报',
  }
  return map[type] || map[type.toLowerCase()] || type
}
</script>

<template>
  <div :class="pageClassNames.appWhite">
    <div class="fixed inset-0 overflow-hidden pointer-events-none z-0">
      <div class="absolute top-[-10%] right-[-10%] w-[50%] h-[50%] bg-blue-500/5 rounded-full blur-[120px]" />
      <div class="absolute bottom-[-10%] left-[-10%] w-[50%] h-[50%] bg-purple-500/5 rounded-full blur-[120px]" />
    </div>

    <main class="relative z-10 mx-auto flex w-full max-w-6xl flex-col gap-5 px-4 py-6 sm:px-6 lg:px-8 overflow-x-hidden">
      <!-- 页头 -->
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <RouterLink to="/lobby" class="mb-3 inline-flex items-center gap-2 text-xs font-bold text-slate-500 transition-colors hover:text-slate-900 dark:text-slate-400 dark:hover:text-white">
            <ArrowLeft class="h-4 w-4" /> 返回大厅
          </RouterLink>
          <h1 class="text-2xl font-black tracking-tight sm:text-3xl flex items-center gap-2">
            <Shield class="w-6 h-6 text-sky-500" />
            反作弊管理中心
          </h1>
          <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">检测记录 · 申诉审批 · 即时处置 · 系统配置</p>
        </div>
      </div>

      <!-- 标签页 -->
      <div class="flex gap-2 overflow-x-auto border-b border-slate-200 pb-2 dark:border-white/10 custom-scrollbar">
        <button v-for="tab in ['detection', 'appeals', 'config', 'audit']" :key="tab"
          :class="['whitespace-nowrap rounded-lg px-4 py-2 text-sm font-bold transition-colors', activeTab === tab ? 'bg-sky-600 text-white shadow-sm dark:bg-sky-500' : 'text-slate-600 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-white/5']"
          @click="activeTab = tab as any">
          {{ tab === 'detection' ? '检测管理' : tab === 'appeals' ? '申诉管理' : tab === 'config' ? '配置管理' : '审计日志' }}
        </button>
      </div>

      <!-- ======================== 检测管理 ======================== -->
      <div v-if="activeTab === 'detection'" class="space-y-5 animate-in fade-in slide-in-from-bottom-4 duration-500">
        <!-- 快捷操作栏 -->
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <div class="rounded-xl border border-slate-200 bg-white p-4 dark:border-white/10 dark:bg-[#111318]">
            <div class="flex items-center gap-2 mb-2">
              <SearchIcon class="h-4 w-4 text-slate-400" />
              <span class="text-xs font-black text-slate-500 uppercase tracking-widest">玩家查询</span>
            </div>
            <div class="flex gap-2">
              <input v-model="quickPlayerId" type="text" placeholder="输入玩家UID" class="flex-1 h-9 rounded-lg border border-slate-200 bg-white px-3 text-sm outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20 dark:text-white"
                @keyup.enter="quickSearchPlayer" />
              <button class="h-9 rounded-lg bg-sky-600 px-3 text-xs font-black text-white hover:bg-sky-500 disabled:opacity-50" :disabled="quickSearching" @click="quickSearchPlayer">
                <template v-if="quickSearching">查询中...</template>
                <template v-else>查询</template>
              </button>
            </div>
            <div v-if="quickSearchResult" class="mt-2 text-xs">
              <div v-if="quickSearchResult.no_data" class="text-slate-400">未找到该玩家的相关记录</div>
              <div v-else>
                <div class="flex items-center justify-between mb-1">
                  <span class="font-bold text-slate-600 dark:text-slate-300">玩家 UID {{ quickSearchResult.player_id }}</span>
                  <div class="flex gap-1">
                    <button class="px-2 py-0.5 rounded bg-rose-500/10 text-rose-600 font-black text-[10px] hover:bg-rose-500/20" @click="openQuickBan(Number(quickSearchResult.player_id))">
                      快速封禁
                    </button>
                  </div>
                </div>
                <div class="text-slate-400">最近 {{ quickSearchResult.logs.length }} 条操作记录</div>
              </div>
            </div>
          </div>

          <div class="rounded-xl border border-slate-200 bg-white p-4 dark:border-white/10 dark:bg-[#111318]">
            <div class="flex items-center gap-2 mb-2">
              <Gavel class="h-4 w-4 text-slate-400" />
              <span class="text-xs font-black text-slate-500 uppercase tracking-widest">快捷处置</span>
            </div>
            <div class="flex flex-wrap gap-2">
              <button class="inline-flex items-center gap-1 rounded-lg bg-rose-500/10 px-3 py-1.5 text-xs font-black text-rose-600 hover:bg-rose-500/20 dark:text-rose-400" @click="openQuickBan(0)">
                <Ban class="h-3.5 w-3.5" /> 快速封禁
              </button>
              <button class="rounded-lg bg-yellow-500/10 px-3 py-1.5 text-xs font-black text-yellow-600 hover:bg-yellow-500/20 dark:text-yellow-400" @click="loadDetections(); showAlert('已刷新检测列表', '提示')">
                <RefreshCw class="h-3.5 w-3.5" /> 刷新列表
              </button>
            </div>
          </div>

          <div class="rounded-xl border border-slate-200 bg-white p-4 dark:border-white/10 dark:bg-[#111318]">
            <div class="flex items-center gap-2 mb-2">
              <ListChecks class="h-4 w-4 text-slate-400" />
              <span class="text-xs font-black text-slate-500 uppercase tracking-widest">统计概览</span>
            </div>
            <div class="text-sm flex flex-wrap gap-x-3 gap-y-1">
              <span class="text-slate-500">待处理：</span>
              <span class="font-black text-rose-500">{{ detectionList.filter(d => d.review_status !== 'processed').length }}</span>
              <span class="text-slate-300">|</span>
              <span class="text-slate-500">已归档：</span>
              <span class="font-black text-slate-600 dark:text-slate-300">{{ archivedIds.size }}</span>
            </div>
          </div>
        </div>

        <!-- 批量处置操作栏 -->
        <div v-if="selectedDetectionIds.size > 0" class="flex flex-wrap items-center gap-3 rounded-xl border border-sky-200 bg-sky-50 px-4 py-3 dark:border-sky-500/30 dark:bg-sky-500/10">
          <span class="text-sm font-black text-sky-700 dark:text-sky-300">已选 {{ selectedDetectionIds.size }} 条</span>
          <button
            :disabled="batchReviewLoading"
            class="inline-flex items-center gap-1.5 rounded-lg bg-sky-600 px-3 py-1.5 text-xs font-black text-white hover:bg-sky-500 disabled:opacity-50"
            @click="batchReviewSelected"
          >
            <ListChecks class="h-3.5 w-3.5" /> {{ batchReviewLoading ? '处理中…' : '批量标记已处理' }}
          </button>
          <button class="rounded-lg border border-slate-200 px-3 py-1.5 text-xs font-bold hover:bg-white dark:border-white/10 dark:hover:bg-white/5" @click="clearDetectionSelection">取消选择</button>
        </div>

        <!-- 搜索与筛选 -->
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex-1 flex items-center gap-2 h-10 rounded-lg border border-slate-200 bg-white px-3 focus-within:border-sky-400 dark:border-white/10 dark:bg-[#111318]">
            <SearchIcon class="h-4 w-4 text-slate-400" />
            <input v-model="detectionSearchTerm" type="text" placeholder="搜索玩家ID或房间ID" class="w-full bg-transparent text-sm text-slate-900 outline-none dark:text-white placeholder:text-slate-400" @input="detectionPage = 1" />
          </div>
          <select v-model="detectionStatusFilter" @change="detectionPage = 1" class="h-10 rounded-lg border border-slate-200 bg-white px-3 text-sm text-slate-900 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-[#111318] dark:text-white">
            <option value="all">所有状态</option>
            <option value="observe">观察</option>
            <option value="warning">警告</option>
            <option value="mute">禁言</option>
            <option value="ban">封号</option>
          </select>
          <button
            @click="showArchived = !showArchived"
            :class="['inline-flex items-center gap-1.5 h-10 rounded-lg border px-3 text-xs font-black transition-colors', showArchived ? 'bg-slate-200 text-slate-700 border-slate-300 dark:bg-white/15 dark:text-white dark:border-white/20' : 'border-slate-200 text-slate-500 hover:bg-slate-50 dark:border-white/10 dark:text-slate-400 dark:hover:bg-white/5']"
          >
            <FileText class="h-3.5 w-3.5" /> {{ showArchived ? '隐藏已归档' : '显示已归档' }}
          </button>
        </div>

        <!-- 检测列表（树形分组） -->
        <div class="space-y-3">
          <div v-if="groupedDetections.length === 0" class="rounded-xl border border-slate-200 bg-white p-8 text-center text-sm font-bold text-slate-400 dark:border-white/10 dark:bg-[#111318]">
            暂无检测记录
          </div>

          <div v-for="room in groupedDetections" :key="room.roomId" class="rounded-xl border border-slate-200 bg-white shadow-sm dark:border-white/10 dark:bg-[#111318] overflow-hidden">
            <!-- 一级：房间 -->
            <div
              class="flex flex-wrap items-center gap-2 px-4 py-3 cursor-pointer select-none transition-colors hover:bg-slate-50 dark:hover:bg-white/5 border-b border-slate-200 dark:border-white/10"
              @click="toggleRoom(room.roomId)"
            >
              <div class="flex items-center gap-2 min-w-0 flex-1">
                <ChevronRight
                  class="h-4 w-4 flex-shrink-0 text-slate-400 transition-transform duration-200"
                  :class="expandedRooms[room.roomId] ? 'rotate-90' : ''"
                />
                <div class="flex items-center gap-2 min-w-0">
                  <span class="text-[10px] font-black uppercase tracking-widest text-slate-400 flex-shrink-0">房间</span>
                  <code class="text-sm font-mono font-bold truncate">{{ room.roomId }}</code>
                </div>
              </div>
              <div class="flex items-center gap-2 flex-shrink-0">
                <span class="text-[11px] text-slate-500 whitespace-nowrap">
                  {{ uniquePlayersInRoom(room.allDetections) }} 名玩家 · {{ room.allDetections.length }} 条记录
                </span>
                <span :class="['inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-black whitespace-nowrap', getRiskBg(roomRiskSummary(room.allDetections).max)]">
                  最高 {{ roomRiskSummary(room.allDetections).max.toFixed(0) }}
                </span>
              </div>
            </div>

            <!-- 二级：玩家列表 -->
            <div v-if="expandedRooms[room.roomId]" class="divide-y divide-slate-100 dark:divide-white/5">
              <div v-for="player in Object.values(room.players)" :key="player.uid" class="transition-colors">
                <!-- 玩家头 -->
                <div
                  class="flex items-center justify-between gap-3 px-4 py-2.5 pl-10 cursor-pointer select-none transition-colors hover:bg-slate-50 dark:hover:bg-white/5"
                  @click="togglePlayer(`${room.roomId}:${player.uid}`)"
                >
                  <div class="flex items-center gap-2 min-w-0">
                    <ChevronRight
                      class="h-3.5 w-3.5 flex-shrink-0 text-slate-400 transition-transform duration-200"
                      :class="expandedPlayers[`${room.roomId}:${player.uid}`] ? 'rotate-90' : ''"
                    />
                    <span class="text-[10px] font-black uppercase tracking-widest text-slate-400">玩家</span>
                    <span class="text-sm font-bold">{{ player.uid }}</span>
                  </div>
                  <div class="flex items-center gap-2 flex-shrink-0">
                    <span class="text-[11px] text-slate-500">{{ player.detections.length }} 条</span>
                    <button
                      class="rounded-lg bg-rose-500/10 px-2 py-1 text-[10px] font-black text-rose-600 hover:bg-rose-500/20 dark:text-rose-400"
                      @click.stop="openQuickBan(player.uid)"
                    >
                      封禁
                    </button>
                  </div>
                </div>

                <!-- 三级：检测记录 -->
                <div v-if="expandedPlayers[`${room.roomId}:${player.uid}`]" class="bg-slate-50/50 dark:bg-black/10">
                  <div
                    v-for="detection in player.detections"
                    :key="detection.id"
                    class="flex flex-wrap items-center gap-2 px-4 py-2.5 pl-16 border-t border-slate-100 dark:border-white/5 hover:bg-slate-100/50 dark:hover:bg-white/5 transition-colors"
                  >
                    <input
                      type="checkbox"
                      class="h-4 w-4 flex-shrink-0 accent-sky-600"
                      :checked="selectedDetectionIds.has(detection.id)"
                      @click.stop="toggleDetectionSelected(detection.id)"
                    />
                    <div class="flex items-center gap-2 min-w-0 flex-1 flex-wrap">
                      <span :class="['inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-black whitespace-nowrap', getRiskBg(detection.risk_score)]">
                        {{ detection.risk_score.toFixed(0) }} · {{ getRiskLevel(detection.risk_score) }}
                      </span>
                      <span :class="['inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-black whitespace-nowrap', getSanctionBadge(detection.suggested_action || detection.sanction_type).color]">
                        {{ getSanctionBadge(detection.suggested_action || detection.sanction_type).label }}
                      </span>
                      <span class="text-[11px] font-bold whitespace-nowrap" :class="detection.review_status === 'processed' ? 'text-emerald-600 dark:text-emerald-400' : 'text-amber-600 dark:text-amber-400'">
                        {{ detection.review_status === 'processed' ? '已处理' : '待处理' }}
                      </span>
                      <span class="text-[11px] text-slate-500 whitespace-nowrap">{{ formatTime(detection.created_at) }}</span>
                    </div>
                    <div class="flex items-center gap-1 flex-shrink-0">
                      <button
                        class="inline-flex items-center gap-1 rounded-lg bg-sky-50 px-2 py-1.5 text-[10px] font-black text-sky-600 transition-colors hover:bg-sky-100 dark:bg-sky-500/10 dark:text-sky-300 dark:hover:bg-sky-500/20"
                        @click="openDetectionDetail(detection)"
                      >
                        <Eye class="h-3.5 w-3.5" /> 详情
                      </button>
                      <button
                        class="inline-flex items-center gap-1 rounded-lg bg-rose-50 px-2 py-1.5 text-[10px] font-black text-rose-600 transition-colors hover:bg-rose-100 dark:bg-rose-500/10 dark:text-rose-300 dark:hover:bg-rose-500/20"
                        @click="openQuickBan(detection.player_id ?? detection.player_uid)"
                      >
                        <Ban class="h-3.5 w-3.5" /> 封禁
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 分页 -->
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

      <!-- ======================== 申诉管理 ======================== -->
      <div v-if="activeTab === 'appeals'" class="space-y-5 animate-in fade-in slide-in-from-bottom-4 duration-500">
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex-1 flex items-center gap-2 h-10 rounded-lg border border-slate-200 bg-white px-3 focus-within:border-sky-400 dark:border-white/10 dark:bg-[#111318]">
            <SearchIcon class="h-4 w-4 text-slate-400" />
            <input v-model="appealsSearchTerm" type="text" placeholder="搜索玩家ID或房间ID" class="w-full bg-transparent text-sm text-slate-900 outline-none dark:text-white placeholder:text-slate-400" @input="appealsPage = 1" />
          </div>
          <select v-model="appealsStatusFilter" @change="appealsPage = 1" class="h-10 rounded-lg border border-slate-200 bg-white px-3 text-sm text-slate-900 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-[#111318] dark:text-white">
            <option value="all">所有状态</option>
            <option value="pending">待审核</option>
            <option value="approved">已批准</option>
            <option value="rejected">已拒绝</option>
          </select>
        </div>

        <div class="rounded-xl border border-slate-200 bg-white shadow-sm dark:border-white/10 dark:bg-[#111318]">
          <table class="w-full text-left text-sm">
            <thead>
              <tr class="border-b border-slate-200 bg-slate-50 dark:border-white/10 dark:bg-black/20">
                <th class="p-2.5 font-black text-[10px] uppercase tracking-widest text-slate-500">玩家</th>
                <th class="p-2.5 font-black text-[10px] uppercase tracking-widest text-slate-500">申诉理由</th>
                <th class="p-2.5 font-black text-[10px] uppercase tracking-widest text-slate-500">状态</th>
                <th class="p-2.5 font-black text-[10px] uppercase tracking-widest text-slate-500">时间</th>
                <th class="p-2.5 font-black text-[10px] uppercase tracking-widest text-slate-500 text-right">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-200 dark:divide-white/10">
              <tr v-if="filteredAppeals.length === 0">
                <td colspan="5" class="p-8 text-center text-slate-400 font-bold">暂无申诉记录</td>
              </tr>
              <tr v-for="appeal in filteredAppeals" :key="appeal.id" class="hover:bg-slate-50 dark:hover:bg-white/5 transition-colors">
                <td class="p-2.5 font-bold whitespace-nowrap">{{ appeal.player_id || appeal.player_uid }}</td>
                <td class="p-2.5 max-w-[180px]">
                  <div class="truncate text-sm" :title="appeal.reason">{{ appeal.reason }}</div>
                </td>
                <td class="p-2.5">
                  <span :class="['inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-black whitespace-nowrap', getStatusBadge(appeal.status).color]">
                    {{ getStatusBadge(appeal.status).label }}
                  </span>
                </td>
                <td class="p-2.5 text-slate-500 text-[11px] whitespace-nowrap">{{ formatTime(appeal.created_at) }}</td>
                <td class="p-2.5 text-right whitespace-nowrap">
                  <div v-if="appeal.status === 'pending'" class="flex items-center justify-end gap-1">
                    <button class="inline-flex items-center gap-1 rounded-lg bg-emerald-50 px-2 py-1.5 text-[10px] font-black text-emerald-600 transition-colors hover:bg-emerald-100 dark:bg-emerald-500/10 dark:text-emerald-300 dark:hover:bg-emerald-500/20" @click="handleApproveAppeal(appeal.id)">
                      <CheckCircle class="h-3.5 w-3.5" /> 批准
                    </button>
                    <button class="inline-flex items-center gap-1 rounded-lg bg-rose-50 px-2 py-1.5 text-[10px] font-black text-rose-600 transition-colors hover:bg-rose-100 dark:bg-rose-500/10 dark:text-rose-300 dark:hover:bg-rose-500/20" @click="handleRejectAppeal(appeal.id)">
                      <XCircle class="h-3.5 w-3.5" /> 拒绝
                    </button>
                  </div>
                  <span v-else class="text-slate-400">-</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 分页 -->
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

      <!-- ======================== 配置管理 ======================== -->
      <div v-if="activeTab === 'config'" class="space-y-5 animate-in fade-in slide-in-from-bottom-4 duration-500">
        <div v-if="!editingConfig && configData" class="rounded-xl border border-slate-200 bg-white p-5 shadow-sm dark:border-white/10 dark:bg-[#111318]">
          <div class="mb-5 flex items-center justify-between">
            <h2 class="text-lg font-black">反作弊策略配置</h2>
            <button class="inline-flex items-center gap-2 rounded-lg bg-sky-600 px-4 py-2 text-sm font-black text-white transition-colors hover:bg-sky-500" @click="startEditConfig">
              <Settings class="h-4 w-4" /> 编辑配置
            </button>
          </div>
          <div class="grid gap-6 md:grid-cols-2">
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
                  <div class="rounded-lg border border-slate-200 bg-white p-3 text-slate-700 dark:border-white/10 dark:bg-[#111318] dark:text-slate-300 text-sm">
                    {{ configData.unban?.default_message || defaultCompensationMessage }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-if="editingConfig && tempConfig" class="rounded-xl border border-slate-200 bg-white p-5 shadow-sm dark:border-white/10 dark:bg-[#111318]">
          <h2 class="mb-5 text-lg font-black">编辑配置</h2>
          <div class="grid gap-6 md:grid-cols-2">
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

            <!-- 优化特性开关（默认关闭，灰度启用） -->
            <div v-if="tempConfig.optimization" class="space-y-4 md:col-span-2">
              <h4 class="font-bold border-b border-slate-100 pb-2 dark:border-white/10">优化特性（灰度，默认关闭）</h4>
              <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div class="rounded-lg border border-slate-200 p-3 dark:border-white/10">
                  <label class="flex items-center gap-2 text-sm font-black"><input v-model="tempConfig.optimization.adaptive_threshold.enabled" type="checkbox" class="h-4 w-4 accent-sky-600" /> 自适应阈值</label>
                  <label class="mt-2 block text-xs font-bold text-slate-500">个人基线权重: <input v-model.number="tempConfig.optimization.adaptive_threshold.personal_weight" type="number" step="0.05" min="0" max="1" class="mt-1 w-full rounded-lg border border-slate-200 px-2 py-1.5 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20" /></label>
                  <label class="mt-2 block text-xs font-bold text-slate-500">全局超人阈值(z): <input v-model.number="tempConfig.optimization.adaptive_threshold.global_superhuman_z" type="number" step="0.5" min="0" class="mt-1 w-full rounded-lg border border-slate-200 px-2 py-1.5 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20" /></label>
                </div>
                <div class="rounded-lg border border-slate-200 p-3 dark:border-white/10">
                  <label class="flex items-center gap-2 text-sm font-black"><input v-model="tempConfig.optimization.zscore.enabled" type="checkbox" class="h-4 w-4 accent-sky-600" /> Z分数异常检测</label>
                  <label class="mt-2 block text-xs font-bold text-slate-500">触发阈值(z): <input v-model.number="tempConfig.optimization.zscore.threshold" type="number" step="0.5" min="0" class="mt-1 w-full rounded-lg border border-slate-200 px-2 py-1.5 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20" /></label>
                  <label class="mt-2 block text-xs font-bold text-slate-500">权重: <input v-model.number="tempConfig.optimization.zscore.weight" type="number" step="0.05" min="0" class="mt-1 w-full rounded-lg border border-slate-200 px-2 py-1.5 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20" /></label>
                </div>
                <div class="rounded-lg border border-slate-200 p-3 dark:border-white/10">
                  <label class="flex items-center gap-2 text-sm font-black"><input v-model="tempConfig.optimization.new_player.enabled" type="checkbox" class="h-4 w-4 accent-sky-600" /> 新玩家保护</label>
                  <label class="mt-2 block text-xs font-bold text-slate-500">最少对局数: <input v-model.number="tempConfig.optimization.new_player.min_games" type="number" min="0" class="mt-1 w-full rounded-lg border border-slate-200 px-2 py-1.5 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20" /></label>
                  <label class="mt-2 block text-xs font-bold text-slate-500">放宽系数(&lt;1): <input v-model.number="tempConfig.optimization.new_player.relaxation_factor" type="number" step="0.05" min="0" max="1" class="mt-1 w-full rounded-lg border border-slate-200 px-2 py-1.5 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20" /></label>
                </div>
                <div class="rounded-lg border border-slate-200 p-3 dark:border-white/10">
                  <label class="flex items-center gap-2 text-sm font-black"><input v-model="tempConfig.optimization.risk_decay.enabled" type="checkbox" class="h-4 w-4 accent-sky-600" /> 风险时间衰减</label>
                  <label class="mt-2 block text-xs font-bold text-slate-500">衰减因子(0-1): <input v-model.number="tempConfig.optimization.risk_decay.decay_factor" type="number" step="0.05" min="0" max="1" class="mt-1 w-full rounded-lg border border-slate-200 px-2 py-1.5 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20" /></label>
                  <label class="mt-2 block text-xs font-bold text-slate-500">时间下限(小时): <input v-model.number="tempConfig.optimization.risk_decay.min_floor_hours" type="number" min="0" class="mt-1 w-full rounded-lg border border-slate-200 px-2 py-1.5 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20" /></label>
                </div>
              </div>
            </div>
          </div>
          <div class="mt-6 flex gap-3 justify-end">
            <button class="inline-flex items-center gap-2 rounded-lg bg-sky-600 px-4 py-2 text-sm font-black text-white hover:bg-sky-500" @click="saveConfig">
              <Save class="h-4 w-4" /> 保存配置
            </button>
            <button class="rounded-lg border border-slate-200 px-4 py-2 text-sm font-bold hover:bg-slate-50 dark:border-white/10 dark:hover:bg-white/5" @click="cancelEditConfig">取消</button>
          </div>
        </div>

        <!-- ======================== 检测规则离线测试 ======================== -->
        <div class="rounded-xl border border-slate-200 bg-white p-5 shadow-sm dark:border-white/10 dark:bg-[#111318]">
          <div class="flex items-center gap-2 mb-3">
            <ListChecks class="h-4 w-4 text-slate-400" />
            <span class="text-xs font-black text-slate-500 uppercase tracking-widest">规则离线测试（沙盒）</span>
          </div>
          <p class="text-xs text-slate-500 mb-3">用当前{{ editingConfig && tempConfig ? '草拟' : '线上' }}配置对历史检测样本重跑评分，预估命中分布与等级变化。该测试在隔离环境运行，不影响线上玩家、风险或封禁状态。</p>
          <div class="flex flex-wrap items-center gap-3">
            <label class="text-sm font-bold text-slate-500">样本数: <input v-model.number="ruleTestSampleLimit" type="number" min="1" max="500" class="ml-1 w-24 rounded-lg border border-slate-200 px-2 py-1.5 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20" /></label>
            <button
              :disabled="ruleTestLoading"
              class="inline-flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-black text-white hover:bg-indigo-500 disabled:opacity-50"
              @click="runRuleTest"
            >
              <RefreshCw class="h-4 w-4" /> {{ ruleTestLoading ? '测试中…' : '运行规则测试' }}
            </button>
          </div>

          <div v-if="ruleTestResult" class="mt-4 space-y-3">
            <div class="flex flex-wrap gap-3 text-sm">
              <span class="rounded-lg bg-slate-100 px-3 py-1.5 font-bold dark:bg-white/5">样本数: {{ ruleTestResult.sample_count }}</span>
              <span class="rounded-lg bg-rose-50 px-3 py-1.5 font-bold text-rose-600 dark:bg-rose-500/10 dark:text-rose-300">升级: {{ ruleTestResult.escalations }}</span>
              <span class="rounded-lg bg-emerald-50 px-3 py-1.5 font-bold text-emerald-600 dark:bg-emerald-500/10 dark:text-emerald-300">降级: {{ ruleTestResult.deescalations }}</span>
            </div>
            <div>
              <span class="block text-xs font-black text-slate-500 uppercase tracking-widest mb-1">命中分布（草拟配置）</span>
              <div class="flex flex-wrap gap-2 text-xs">
                <span v-for="(count, tier) in ruleTestResult.hit_distribution" :key="tier" class="rounded-full bg-slate-100 px-2.5 py-1 font-bold dark:bg-white/5">{{ tier }}: {{ count }}</span>
              </div>
            </div>
            <div v-if="ruleTestResult.tier_change_counts && Object.keys(ruleTestResult.tier_change_counts).length">
              <span class="block text-xs font-black text-slate-500 uppercase tracking-widest mb-1">等级变化（线上→草拟）</span>
              <div class="flex flex-wrap gap-2 text-xs">
                <span v-for="(count, change) in ruleTestResult.tier_change_counts" :key="change" class="rounded-full bg-amber-50 px-2.5 py-1 font-bold text-amber-700 dark:bg-amber-500/10 dark:text-amber-300">{{ change }}: {{ count }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ======================== 审计日志 ======================== -->
      <div v-if="activeTab === 'audit'" class="space-y-5 animate-in fade-in slide-in-from-bottom-4 duration-500">
        <div class="flex flex-wrap items-center gap-2">
          <div class="flex-1 flex items-center gap-2 h-10 min-w-0 rounded-lg border border-slate-200 bg-white px-3 focus-within:border-sky-400 dark:border-white/10 dark:bg-[#111318]">
            <SearchIcon class="h-4 w-4 flex-shrink-0 text-slate-400" />
            <input v-model="auditSearchTerm" type="text" placeholder="搜索玩家ID" class="w-full bg-transparent text-sm text-slate-900 outline-none dark:text-white placeholder:text-slate-400" />
          </div>
          <input v-model="auditStartDate" type="date" class="h-10 flex-1 min-w-[130px] rounded-lg border border-slate-200 bg-white px-3 text-sm text-slate-900 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-[#111318] dark:text-white" />
          <input v-model="auditEndDate" type="date" class="h-10 flex-1 min-w-[130px] rounded-lg border border-slate-200 bg-white px-3 text-sm text-slate-900 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-[#111318] dark:text-white" />
          <button class="inline-flex items-center gap-2 h-10 rounded-lg bg-sky-600 px-4 text-sm font-black text-white hover:bg-sky-500" @click="loadAuditLog">
            <Filter class="h-4 w-4" /> 查询
          </button>
          <button class="inline-flex items-center gap-2 h-10 rounded-lg border border-slate-200 bg-white px-4 text-sm font-bold hover:bg-slate-50 dark:border-white/10 dark:bg-[#111318] dark:hover:bg-white/5" @click="exportAuditLog">
            <Download class="h-4 w-4" /> 导出
          </button>
        </div>

        <div class="rounded-xl border border-slate-200 bg-white shadow-sm dark:border-white/10 dark:bg-[#111318]">
          <table class="w-full text-left text-sm">
            <thead>
              <tr class="border-b border-slate-200 bg-slate-50 dark:border-white/10 dark:bg-black/20">
                <th class="p-2.5 font-black text-[10px] uppercase tracking-widest text-slate-500">玩家</th>
                <th class="p-2.5 font-black text-[10px] uppercase tracking-widest text-slate-500">操作</th>
                <th class="p-2.5 font-black text-[10px] uppercase tracking-widest text-slate-500">详情</th>
                <th class="p-2.5 font-black text-[10px] uppercase tracking-widest text-slate-500">补偿</th>
                <th class="p-2.5 font-black text-[10px] uppercase tracking-widest text-slate-500">时间</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-200 dark:divide-white/10">
              <tr v-if="auditLogs.length === 0">
                <td colspan="5" class="p-8 text-center text-slate-400 font-bold">暂无审计日志</td>
              </tr>
              <tr v-for="log in auditLogs" :key="log.id" class="hover:bg-slate-50 dark:hover:bg-white/5 transition-colors">
                <td class="p-2.5 font-bold whitespace-nowrap">{{ log.player_id }}</td>
                <td class="p-2.5">
                  <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-black bg-slate-100 text-slate-700 dark:bg-white/10 dark:text-slate-200 whitespace-nowrap">
                    {{ log.action_type }}
                  </span>
                </td>
                <td class="p-2.5 max-w-[200px]">
                  <div class="truncate text-sm" :title="log.details">{{ log.details }}</div>
                </td>
                <td class="p-2.5">
                  <div v-if="log.compensation_status" class="flex flex-wrap items-center gap-1">
                    <span :class="['inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-black whitespace-nowrap', getCompensationBadge(log.compensation_status).color]">
                      {{ getCompensationBadge(log.compensation_status).label }}
                    </span>
                    <span v-if="log.compensation_amount" class="text-[10px] text-emerald-600 dark:text-emerald-400 font-bold whitespace-nowrap">{{ log.compensation_amount }}燃素</span>
                  </div>
                  <span v-else class="text-slate-400">-</span>
                </td>
                <td class="p-2.5 text-[11px] text-slate-500 whitespace-nowrap">{{ formatTime(log.created_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 分页 -->
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

    <!-- ======================== 模态框 ======================== -->

    <!-- 快速封禁 -->
    <Teleport to="body">
      <div v-if="showQuickBanModal" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 p-4 backdrop-blur-sm" @click.self="showQuickBanModal = false">
        <div class="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl dark:bg-[#111318] border border-slate-200 dark:border-white/10 animate-in zoom-in-95 duration-200">
          <div class="flex items-center gap-3 mb-6">
            <div class="w-10 h-10 rounded-xl bg-rose-500/10 flex items-center justify-center">
              <Ban class="w-5 h-5 text-rose-500" />
            </div>
            <div>
              <h2 class="text-lg font-black">快速封禁</h2>
              <p class="text-sm text-slate-500">玩家 UID: {{ quickBanTarget.uid || '待输入' }}</p>
            </div>
          </div>

          <div class="space-y-4">
            <div>
              <label class="block text-sm font-bold text-slate-500 mb-1">玩家UID</label>
              <input v-model.number="quickBanTarget.uid" type="number" placeholder="输入玩家UID" class="w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-rose-400 dark:border-white/10 dark:bg-black/20 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-bold text-slate-500 mb-1">封禁截止时间</label>
              <input v-model="quickBanTarget.until" type="datetime-local" class="w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-rose-400 dark:border-white/10 dark:bg-black/20 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-bold text-slate-500 mb-1">封禁原因</label>
              <textarea v-model="quickBanTarget.reason" rows="2" class="w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-rose-400 dark:border-white/10 dark:bg-black/20 resize-none text-sm" placeholder="输入封禁原因"></textarea>
            </div>
          </div>

          <div class="mt-6 flex gap-3 justify-end border-t border-slate-100 pt-4 dark:border-white/10">
            <button class="rounded-lg bg-rose-600 px-4 py-2 text-sm font-black text-white hover:bg-rose-500 disabled:opacity-50" :disabled="operating" @click="executeQuickBan">
              确认封禁
            </button>
            <button class="rounded-lg border border-slate-200 px-4 py-2 text-sm font-bold hover:bg-slate-50 dark:border-white/10 dark:hover:bg-white/5" @click="showQuickBanModal = false">取消</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- 检测详情 -->
    <Teleport to="body">
      <div v-if="showDetailModal" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 p-4 backdrop-blur-sm" @click.self="showDetailModal = false">
        <div class="w-full max-w-2xl max-h-[90vh] overflow-y-auto overflow-x-hidden custom-scrollbar rounded-2xl bg-white p-6 shadow-2xl dark:bg-[#111318] border border-slate-200 dark:border-white/10 animate-in zoom-in-95 duration-200">
          <div class="flex items-center justify-between mb-6">
            <h2 class="text-xl font-black">检测详情</h2>
            <button class="rounded-lg border border-slate-200 px-3 py-1.5 text-sm font-bold hover:bg-slate-50 dark:border-white/10 dark:hover:bg-white/5" @click="showDetailModal = false">关闭</button>
          </div>

          <div v-if="selectedDetection" class="space-y-6">
            <!-- 基本信息 -->
            <div class="grid grid-cols-2 gap-4 rounded-xl border border-slate-100 bg-slate-50 p-4 dark:border-white/5 dark:bg-black/20">
              <div class="min-w-0"><span class="text-[10px] font-black uppercase tracking-widest text-slate-500 block mb-1">玩家ID</span><span class="font-bold truncate block">{{ selectedDetection.player_id }}</span></div>
              <div class="min-w-0"><span class="text-[10px] font-black uppercase tracking-widest text-slate-500 block mb-1">房间ID</span><span class="font-bold truncate block">{{ selectedDetection.room_id }}</span></div>
              <div class="min-w-0"><span class="text-[10px] font-black uppercase tracking-widest text-slate-500 block mb-1">风险分数</span><span :class="['font-black block', getRiskColor(selectedDetection.risk_score)]">{{ selectedDetection.risk_score.toFixed(1) }} · {{ getRiskLevel(selectedDetection.risk_score) }}</span></div>
              <div class="min-w-0"><span class="text-[10px] font-black uppercase tracking-widest text-slate-500 block mb-1">建议处置</span><span :class="['inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-black', getSanctionBadge(selectedDetection.suggested_action).color]">{{ getSanctionBadge(selectedDetection.suggested_action).label }}</span></div>
              <div class="min-w-0"><span class="text-[10px] font-black uppercase tracking-widest text-slate-500 block mb-1">回放ID</span><span class="font-bold truncate block">{{ selectedDetection.replay_id || '未记录' }}</span></div>
              <div class="min-w-0"><span class="text-[10px] font-black uppercase tracking-widest text-slate-500 block mb-1">检测时间</span><span class="font-bold truncate block">{{ formatTime(selectedDetection.operation_timestamp) }}</span></div>
              <div v-if="selectedDetection.threshold_source" class="min-w-0"><span class="text-[10px] font-black uppercase tracking-widest text-slate-500 block mb-1">阈值来源</span><span class="font-bold truncate block">{{ selectedDetection.threshold_source === 'personal' ? '个人基线' : selectedDetection.threshold_source === 'global' ? '全局基线' : selectedDetection.threshold_source }}</span></div>
              <div v-if="selectedDetection.decay_factor != null" class="min-w-0"><span class="text-[10px] font-black uppercase tracking-widest text-slate-500 block mb-1">历史衰减因子</span><span class="font-bold truncate block">{{ Number(selectedDetection.decay_factor).toFixed(3) }}</span></div>
              <div v-if="selectedDetection.new_player_observe" class="min-w-0 col-span-2"><span class="inline-flex items-center rounded-full bg-amber-100 px-2 py-0.5 text-[10px] font-black text-amber-700 dark:bg-amber-500/10 dark:text-amber-300">新手观察期（自动封禁已抑制，转人工复核）</span></div>
            </div>

            <!-- 回放证据 -->
            <div>
              <h3 class="mb-3 text-sm font-black border-b border-slate-100 pb-2 dark:border-white/10 flex items-center gap-2">
                <FileText class="w-4 h-4 text-sky-500" /> 回放证据
              </h3>
              <div v-if="relatedEvidenceAnchors.length" class="space-y-2 text-sm">
                <div v-for="anchor in relatedEvidenceAnchors" :key="anchorKey(anchor)" class="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 dark:border-white/10 dark:bg-black/20">
                  <div class="flex flex-wrap items-center justify-between gap-2">
                    <span class="font-bold text-slate-700 dark:text-slate-200">
                      {{ anchor.event_type || '房间' }} · {{ formatAnchorPosition(anchor) }}
                    </span>
                    <RouterLink v-if="replayRouteForAnchor(anchor)" :to="replayRouteForAnchor(anchor)" class="inline-flex items-center rounded-md border border-cyan-500/20 bg-cyan-500/10 px-2 py-1 text-[10px] font-black uppercase tracking-widest text-cyan-700 hover:bg-cyan-500/20 dark:text-cyan-300">
                      打开回放
                    </RouterLink>
                  </div>
                  <div class="mt-1 grid gap-1 text-xs text-slate-500 sm:grid-cols-2">
                    <span>玩家: {{ anchor.player_uid || '-' }}</span>
                    <span>时间: {{ formatAnchorTime(anchor) }}</span>
                    <span>精度: {{ precisionLabel(anchor.evidence_precision) }}</span>
                    <span class="truncate" :title="anchor.action_summary">摘要: {{ anchor.action_summary || '-' }}</span>
                  </div>
                </div>
              </div>
              <div v-else class="rounded-lg bg-slate-50 px-3 py-2 text-sm text-slate-400 dark:bg-black/20">无回放证据</div>
            </div>

            <!-- 判定指标 -->
            <div>
              <h3 class="mb-3 text-sm font-black border-b border-slate-100 pb-2 dark:border-white/10 flex items-center gap-2">
                <AlertTriangle class="w-4 h-4 text-amber-500" /> 作弊判定指标
              </h3>
              <div class="space-y-2 text-sm">
                <div v-for="indicator in selectedDetection.indicator_details" :key="indicator.name" class="rounded-lg bg-slate-50 px-3 py-2 dark:bg-black/20">
                  <div class="flex items-center justify-between gap-3">
                    <span class="font-bold text-slate-700 dark:text-slate-200">{{ translateIndicatorName(indicator.name) }}</span>
                    <span class="font-black">{{ Number(indicator.contribution || 0).toFixed(1) }}</span>
                  </div>
                  <div class="mt-1 text-xs text-slate-500">
                    原始值 {{ indicator.raw_value }} · 归一化 {{ Number(indicator.normalized_score || 0).toFixed(1) }} · 权重 {{ indicator.weight }}
                  </div>
                  <div v-if="indicator.explanation" class="mt-1 text-xs text-slate-400">{{ indicator.explanation }}</div>
                  <div v-if="anchorsFromIndicator(indicator).length" class="mt-2 space-y-1">
                    <div v-for="anchor in anchorsFromIndicator(indicator)" :key="`indicator-${indicator.name}-${anchorKey(anchor)}`" class="flex flex-wrap items-center justify-between gap-2 rounded-md border border-slate-200 bg-white px-2 py-1 text-xs dark:border-white/10 dark:bg-black/20">
                      <span class="truncate min-w-0">{{ translateEventType(anchor.event_type) }} · {{ formatAnchorPosition(anchor) }} · UID {{ anchor.player_uid || '-' }} · {{ precisionLabel(anchor.evidence_precision) }}</span>
                      <RouterLink v-if="replayRouteForAnchor(anchor)" :to="replayRouteForAnchor(anchor)" class="font-black text-cyan-700 hover:text-cyan-600 dark:text-cyan-300 flex-shrink-0">打开回放</RouterLink>
                    </div>
                  </div>
                </div>
                <div v-if="!selectedDetection.indicator_details?.length" class="rounded-lg bg-slate-50 px-3 py-2 text-slate-400 dark:bg-black/20">暂无指标明细</div>
              </div>
            </div>

            <!-- 举报贡献 -->
            <div v-if="selectedDetection.report_contribution">
              <h3 class="mb-3 text-sm font-black border-b border-slate-100 pb-2 dark:border-white/10 flex items-center gap-2">
                <UserCheck class="w-4 h-4 text-emerald-500" /> 举报贡献
              </h3>
              <div class="rounded-lg bg-slate-50 px-3 py-2 text-sm dark:bg-black/20">
                去重举报数：<b>{{ selectedDetection.report_contribution.deduplicated_count }}</b>
                · 贡献分：<b>{{ Number(selectedDetection.report_contribution.contribution || 0).toFixed(1) }}</b>
                <div class="mt-1 text-xs text-slate-500">{{ selectedDetection.report_contribution.source_summary }}</div>
                <div class="mt-2 text-xs text-slate-500">
                  权重: {{ selectedDetection.report_contribution.weight ?? '-' }}
                  · 精度: {{ anchorsFromReportContribution(selectedDetection.report_contribution).map(a => precisionLabel(a.evidence_precision)).join(', ') || '未知' }}
                </div>
                <div v-if="anchorsFromReportContribution(selectedDetection.report_contribution).length" class="mt-2 space-y-1">
                  <div v-for="anchor in anchorsFromReportContribution(selectedDetection.report_contribution)" :key="`report-${anchorKey(anchor)}`" class="flex flex-wrap items-center justify-between gap-2 rounded-md border border-slate-200 bg-white px-2 py-1 text-xs dark:border-white/10 dark:bg-black/20">
                    <span class="truncate min-w-0">{{ anchor.event_type || '举报' }} · {{ formatAnchorPosition(anchor) }} · {{ anchor.action_summary || '举报回放点' }}</span>
                    <RouterLink v-if="replayRouteForAnchor(anchor)" :to="replayRouteForAnchor(anchor)" class="font-black text-cyan-700 hover:text-cyan-600 dark:text-cyan-300">打开回放</RouterLink>
                  </div>
                </div>
              </div>
            </div>

            <!-- 处置与审核 -->
            <div class="rounded-xl border border-rose-200 bg-rose-50/50 p-4 dark:border-rose-500/20 dark:bg-rose-500/5">
              <h3 class="mb-4 text-sm font-black flex items-center gap-2 text-rose-600 dark:text-rose-400">
                <Gavel class="w-4 h-4" /> 处置与审核
              </h3>

              <div v-if="selectedDetection.review_status === 'processed'" class="mb-4 p-3 rounded-lg bg-white/60 dark:bg-black/20 border border-slate-200 dark:border-white/10">
                <label class="block text-sm font-bold text-slate-500 mb-2">更改处罚决定</label>
                <div class="flex gap-2">
                  <select v-model="punishmentDecision" class="flex-1 rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20 text-sm">
                    <option value="observe">观察</option>
                    <option value="warning">警告</option>
                    <option value="mute">禁言</option>
                    <option value="ban">封号</option>
                  </select>
                  <button class="rounded-lg bg-sky-600 px-4 py-2 text-sm font-black text-white hover:bg-sky-500 disabled:opacity-50" :disabled="punishmentChangeLoading" @click="changePunishmentDecision">
                    保存
                  </button>
                </div>
              </div>

              <div class="p-3 rounded-lg bg-white/60 dark:bg-black/20 border border-slate-200 dark:border-white/10 space-y-3">
                <div v-if="!selectedDetection.replay_navigation" class="flex items-center gap-3">
                  <label class="block text-sm font-bold text-slate-500 mb-0 whitespace-nowrap">封禁截止:</label>
                  <input v-model="enforcementUntil" type="datetime-local" class="flex-1 rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-rose-400 dark:border-white/10 dark:bg-black/20 text-sm" />
                </div>
                <div>
                  <label class="block text-sm font-bold text-slate-500 mb-1">处置原因</label>
                  <textarea v-model="enforcementReason" rows="2" class="w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-rose-400 dark:border-white/10 dark:bg-black/20 resize-none text-sm" placeholder="请输入封禁原因"></textarea>
                </div>
                <div class="flex gap-2 justify-end">
                  <button class="rounded-lg bg-rose-600 px-4 py-2 text-sm font-black text-white hover:bg-rose-500 disabled:opacity-50" :disabled="enforcementLoading" @click="handlePanelBan">
                    <Ban class="h-4 w-4 inline-block mr-1" />执行封禁
                  </button>
                </div>
              </div>

              <div class="mt-3 p-3 rounded-lg bg-white/60 dark:bg-black/20 border border-slate-200 dark:border-white/10 space-y-3">
                <label class="block text-sm font-bold text-slate-500">审核备注:
                  <textarea v-model="reviewNote" rows="2" class="mt-1 block w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20 resize-none text-sm" placeholder="输入审核备注..."></textarea>
                </label>
                <div class="flex gap-2 justify-end">
                  <button class="inline-flex items-center gap-2 rounded-lg bg-sky-600 px-4 py-2 text-sm font-black text-white hover:bg-sky-500" @click="submitReview">
                    <Save class="h-4 w-4" /> 提交审核
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- 批准申诉 -->
    <Teleport to="body">
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
                调整补偿配置 <ChevronRight class="h-4 w-4 transition-transform group-open:rotate-90" />
              </summary>
              <div class="border-t border-slate-200 p-4 space-y-4 dark:border-white/10">
                <label class="block text-sm font-bold text-slate-500">
                  补偿金额（燃素）:
                  <div class="flex items-center gap-2 mt-1">
                    <input v-model.number="compensationAmount" type="number" min="0" step="1" class="flex-1 rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/40 text-sm" />
                    <span class="text-xs text-slate-400">默认: 100</span>
                  </div>
                </label>
                <label class="block text-sm font-bold text-slate-500">
                  补偿文案:
                  <textarea v-model="compensationMessage" rows="3" class="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/40 resize-none text-sm" placeholder="输入自定义补偿文案..."></textarea>
                </label>
                <button class="text-xs font-bold text-sky-600 hover:text-sky-500 dark:text-sky-400" @click="compensationMessage = defaultCompensationMessage">恢复默认文案</button>
              </div>
            </details>

            <label class="block text-sm font-bold text-slate-500">
              审核备注（可选）:
              <textarea v-model="approvalNote" rows="2" class="mt-1 w-full rounded-lg border border-slate-200 px-3 py-2 outline-none focus:border-sky-400 dark:border-white/10 dark:bg-black/20 resize-none text-sm" placeholder="输入审核备注..."></textarea>
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
