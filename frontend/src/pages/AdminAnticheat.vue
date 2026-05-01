<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useDialog } from '../utils/dialog'
import { adminAPI } from '../utils/api'
import UserAvatar from '../components/UserAvatar.vue'
import {
  Shield,
  Search as SearchIcon,
  ChevronLeft,
  ChevronRight,
  CheckCircle,
  XCircle,
  Clock,
  Eye,
  MessageSquare,
  Save,
  AlertTriangle,
  Settings,
  BarChart3,
  Download,
  Filter,
} from 'lucide-vue-next'
import { cn } from '../utils/cn'

const { showAlert, showConfirm, showPrompt } = useDialog()
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
    detectionList.value = response.data?.detections || []
    detectionTotal.value = response.data?.total || 0
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

const openDetectionDetail = async (detection: any) => {
  try {
    const response = await adminAPI.getDetectionDetail(detection.id)
    selectedDetection.value = response.data
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
      a.player_id?.toString().includes(term) || 
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
    appealsList.value = response.data?.appeals || []
    appealsTotal.value = response.data?.total || 0
  } catch (error: any) {
    showAlert(error.response?.data?.error || '加载申诉列表失败', '错误')
  } finally {
    loading.value = false
  }
}

const handleApproveAppeal = async (appealId: string) => {
  if (!await showConfirm('确认批准此申诉吗？')) return
  
  try {
    await adminAPI.approveAppeal(appealId, { note: '通过审核' })
    showAlert('申诉已批准', '成功')
    loadAppeals()
  } catch (error: any) {
    showAlert(error.response?.data?.error || '批准申诉失败', '错误')
  }
}

const handleRejectAppeal = async (appealId: string) => {
  const note = await showPrompt('请输入拒绝理由（可选）：', '拒绝申诉')
  if (note === null) return
  
  try {
    await adminAPI.rejectAppeal(appealId, { note })
    showAlert('申诉已拒绝', '成功')
    loadAppeals()
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
    const url = window.URL.createObjectURL(new Blob([response]))
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
</script>

<template>
  <div class="admin-anticheat">
    <div class="header">
      <div class="title-bar">
        <Shield class="icon" />
        <h1>反作弊系统管理</h1>
      </div>
    </div>

    <!-- Tabs -->
    <div class="tabs">
      <button
        v-for="tab in ['detection', 'appeals', 'config', 'audit']"
        :key="tab"
        :class="['tab', { active: activeTab === tab }]"
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
    <div v-if="activeTab === 'detection'" class="tab-content">
      <div class="controls">
        <div class="search-box">
          <SearchIcon class="icon" />
          <input 
            v-model="detectionSearchTerm"
            type="text"
            placeholder="搜索玩家ID或房间ID"
            @input="detectionPage = 1"
          />
        </div>

        <select v-model="detectionStatusFilter" @change="detectionPage = 1">
          <option value="all">所有状态</option>
          <option value="observe">观察</option>
          <option value="warning">警告</option>
          <option value="mute">禁言</option>
          <option value="ban">封号</option>
        </select>
      </div>

      <div class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th>玩家ID</th>
              <th>房间ID</th>
              <th>风险分数</th>
              <th>处罚类型</th>
              <th>检测时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="filteredDetections.length === 0">
              <td colspan="6" class="empty">暂无检测记录</td>
            </tr>
            <tr v-for="detection in filteredDetections" :key="detection.id">
              <td>{{ detection.player_id }}</td>
              <td>{{ detection.room_id }}</td>
              <td>
                <span :class="['risk-score', getRiskColor(detection.risk_score)]">
                  {{ detection.risk_score.toFixed(1) }}
                </span>
              </td>
              <td>
                <span :class="['badge', getSanctionBadge(detection.sanction_type).color]">
                  {{ getSanctionBadge(detection.sanction_type).label }}
                </span>
              </td>
              <td>{{ new Date(detection.created_at).toLocaleString('zh-CN') }}</td>
              <td>
                <button class="btn-small btn-primary" @click="openDetectionDetail(detection)">
                  <Eye class="icon" />
                  查看
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div class="pagination">
        <button :disabled="detectionPage === 1" @click="detectionPage--">
          <ChevronLeft class="icon" /> 上一页
        </button>
        <span class="page-info">第 {{ detectionPage }} 页，共 {{ Math.ceil(detectionTotal / detectionLimit) }} 页</span>
        <button :disabled="detectionPage * detectionLimit >= detectionTotal" @click="detectionPage++">
          下一页 <ChevronRight class="icon" />
        </button>
      </div>
    </div>

    <!-- Detection Detail Modal -->
    <div v-if="showDetailModal" class="modal-overlay" @click.self="showDetailModal = false">
      <div class="modal-content">
        <h2>检测详情</h2>
        <div v-if="selectedDetection" class="detail-content">
          <div class="detail-grid">
            <div class="detail-item">
              <label>玩家ID:</label>
              <span>{{ selectedDetection.player_id }}</span>
            </div>
            <div class="detail-item">
              <label>房间ID:</label>
              <span>{{ selectedDetection.room_id }}</span>
            </div>
            <div class="detail-item">
              <label>风险分数:</label>
              <span :class="getRiskColor(selectedDetection.risk_score)">
                {{ selectedDetection.risk_score.toFixed(1) }}
              </span>
            </div>
            <div class="detail-item">
              <label>处罚类型:</label>
              <span :class="['badge', getSanctionBadge(selectedDetection.sanction_type).color]">
                {{ getSanctionBadge(selectedDetection.sanction_type).label }}
              </span>
            </div>
          </div>

          <div class="detail-section">
            <h3>维度分数</h3>
            <div class="dimension-scores">
              <div class="score-item">
                <span>响应时间:</span>
                <span>{{ selectedDetection.response_time_score?.toFixed(1) || 'N/A' }}</span>
              </div>
              <div class="score-item">
                <span>操作频率:</span>
                <span>{{ selectedDetection.frequency_score?.toFixed(1) || 'N/A' }}</span>
              </div>
              <div class="score-item">
                <span>胜率异常:</span>
                <span>{{ selectedDetection.win_rate_score?.toFixed(1) || 'N/A' }}</span>
              </div>
              <div class="score-item">
                <span>操作模式:</span>
                <span>{{ selectedDetection.pattern_score?.toFixed(1) || 'N/A' }}</span>
              </div>
              <div class="score-item">
                <span>账号年龄:</span>
                <span>{{ selectedDetection.account_age_score?.toFixed(1) || 'N/A' }}</span>
              </div>
            </div>
          </div>

          <div class="detail-section">
            <h3>人工审核</h3>
            <div class="review-form">
              <div class="form-group">
                <label>审核决策:</label>
                <select v-model="reviewDecision">
                  <option value="confirm">确认处罚</option>
                  <option value="override">推翻处罚</option>
                </select>
              </div>
              <div class="form-group">
                <label>审核备注:</label>
                <textarea v-model="reviewNote" placeholder="输入审核备注..."></textarea>
              </div>
            </div>
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn btn-primary" @click="submitReview">
            <Save class="icon" /> 提交审核
          </button>
          <button class="btn btn-secondary" @click="showDetailModal = false">关闭</button>
        </div>
      </div>
    </div>

    <!-- Appeals Tab -->
    <div v-if="activeTab === 'appeals'" class="tab-content">
      <div class="controls">
        <div class="search-box">
          <SearchIcon class="icon" />
          <input 
            v-model="appealsSearchTerm"
            type="text"
            placeholder="搜索玩家ID或房间ID"
            @input="appealsPage = 1"
          />
        </div>

        <select v-model="appealsStatusFilter" @change="appealsPage = 1">
          <option value="all">所有状态</option>
          <option value="pending">待审核</option>
          <option value="approved">已批准</option>
          <option value="rejected">已拒绝</option>
        </select>
      </div>

      <div class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th>玩家ID</th>
              <th>申诉理由</th>
              <th>状态</th>
              <th>申诉时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="filteredAppeals.length === 0">
              <td colspan="5" class="empty">暂无申诉记录</td>
            </tr>
            <tr v-for="appeal in filteredAppeals" :key="appeal.id">
              <td>{{ appeal.player_id }}</td>
              <td class="appeal-reason">{{ appeal.reason }}</td>
              <td>
                <span :class="['badge', getStatusBadge(appeal.status).color]">
                  {{ getStatusBadge(appeal.status).label }}
                </span>
              </td>
              <td>{{ new Date(appeal.created_at).toLocaleString('zh-CN') }}</td>
              <td v-if="appeal.status === 'pending'" class="actions">
                <button class="btn-small btn-success" @click="handleApproveAppeal(appeal.id)">
                  <CheckCircle class="icon" /> 批准
                </button>
                <button class="btn-small btn-danger" @click="handleRejectAppeal(appeal.id)">
                  <XCircle class="icon" /> 拒绝
                </button>
              </td>
              <td v-else>-</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div class="pagination">
        <button :disabled="appealsPage === 1" @click="appealsPage--">
          <ChevronLeft class="icon" /> 上一页
        </button>
        <span class="page-info">第 {{ appealsPage }} 页，共 {{ Math.ceil(appealsTotal / appealsLimit) }} 页</span>
        <button :disabled="appealsPage * appealsLimit >= appealsTotal" @click="appealsPage++">
          下一页 <ChevronRight class="icon" />
        </button>
      </div>
    </div>

    <!-- Config Tab -->
    <div v-if="activeTab === 'config'" class="tab-content">
      <div v-if="!editingConfig && configData" class="config-view">
        <button class="btn btn-primary" @click="startEditConfig">
          <Settings class="icon" /> 编辑配置
        </button>

        <div class="config-grid">
          <div class="config-section">
            <h3>检测权重</h3>
            <div class="config-item">
              <label>响应时间权重:</label>
              <span>{{ configData.dimensions?.response_time?.weight || 0.25 }}</span>
            </div>
            <div class="config-item">
              <label>操作频率权重:</label>
              <span>{{ configData.dimensions?.frequency?.weight || 0.25 }}</span>
            </div>
            <div class="config-item">
              <label>胜率异常权重:</label>
              <span>{{ configData.dimensions?.win_rate?.weight || 0.20 }}</span>
            </div>
            <div class="config-item">
              <label>操作模式权重:</label>
              <span>{{ configData.dimensions?.pattern?.weight || 0.15 }}</span>
            </div>
            <div class="config-item">
              <label>账号年龄权重:</label>
              <span>{{ configData.dimensions?.account_age?.weight || 0.15 }}</span>
            </div>
          </div>

          <div class="config-section">
            <h3>处罚阈值</h3>
            <div class="config-item">
              <label>观察阈值:</label>
              <span>{{ configData.sanctions?.observe || 20 }}-40</span>
            </div>
            <div class="config-item">
              <label>警告阈值:</label>
              <span>{{ configData.sanctions?.warning || 40 }}-60</span>
            </div>
            <div class="config-item">
              <label>禁言阈值:</label>
              <span>{{ configData.sanctions?.mute || 60 }}-80</span>
            </div>
            <div class="config-item">
              <label>封号阈值:</label>
              <span>{{ configData.sanctions?.ban || 80 }}-100</span>
            </div>
          </div>
        </div>
      </div>

      <div v-if="editingConfig && tempConfig" class="config-edit">
        <h3>编辑配置</h3>
        <!-- Response Time -->
        <div class="edit-section">
          <h4>响应时间检测</h4>
          <div class="form-group">
            <label>权重:</label>
            <input v-model.number="tempConfig.dimensions.response_time.weight" type="number" step="0.01" min="0" max="1" />
          </div>
          <div class="form-group">
            <label>阈值 (ms):</label>
            <input v-model.number="tempConfig.dimensions.response_time.threshold" type="number" />
          </div>
        </div>

        <!-- Frequency -->
        <div class="edit-section">
          <h4>操作频率检测</h4>
          <div class="form-group">
            <label>权重:</label>
            <input v-model.number="tempConfig.dimensions.frequency.weight" type="number" step="0.01" min="0" max="1" />
          </div>
          <div class="form-group">
            <label>阈值 (每10秒):</label>
            <input v-model.number="tempConfig.dimensions.frequency.threshold" type="number" />
          </div>
        </div>

        <!-- Win Rate -->
        <div class="edit-section">
          <h4>胜率异常检测</h4>
          <div class="form-group">
            <label>权重:</label>
            <input v-model.number="tempConfig.dimensions.win_rate.weight" type="number" step="0.01" min="0" max="1" />
          </div>
          <div class="form-group">
            <label>胜率阈值:</label>
            <input v-model.number="tempConfig.dimensions.win_rate.threshold" type="number" step="0.01" min="0" max="1" />
          </div>
        </div>

        <!-- Sanctions -->
        <div class="edit-section">
          <h4>处罚阈值</h4>
          <div class="form-group">
            <label>观察下界:</label>
            <input v-model.number="tempConfig.sanctions.observe" type="number" min="0" max="100" />
          </div>
          <div class="form-group">
            <label>警告下界:</label>
            <input v-model.number="tempConfig.sanctions.warning" type="number" min="0" max="100" />
          </div>
          <div class="form-group">
            <label>禁言下界:</label>
            <input v-model.number="tempConfig.sanctions.mute" type="number" min="0" max="100" />
          </div>
          <div class="form-group">
            <label>封号下界:</label>
            <input v-model.number="tempConfig.sanctions.ban" type="number" min="0" max="100" />
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn btn-primary" @click="saveConfig">
            <Save class="icon" /> 保存配置
          </button>
          <button class="btn btn-secondary" @click="cancelEditConfig">取消</button>
        </div>
      </div>
    </div>

    <!-- Audit Log Tab -->
    <div v-if="activeTab === 'audit'" class="tab-content">
      <div class="controls">
        <div class="search-box">
          <SearchIcon class="icon" />
          <input 
            v-model="auditSearchTerm"
            type="text"
            placeholder="搜索玩家ID"
          />
        </div>

        <input 
          v-model="auditStartDate"
          type="date"
          placeholder="开始日期"
        />

        <input 
          v-model="auditEndDate"
          type="date"
          placeholder="结束日期"
        />

        <button class="btn btn-primary" @click="loadAuditLog">
          <Filter class="icon" /> 查询
        </button>

        <button class="btn btn-secondary" @click="exportAuditLog">
          <Download class="icon" /> 导出
        </button>
      </div>

      <div class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th>玩家ID</th>
              <th>操作类型</th>
              <th>详情</th>
              <th>时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="auditLogs.length === 0">
              <td colspan="4" class="empty">暂无审计日志</td>
            </tr>
            <tr v-for="log in auditLogs" :key="log.id">
              <td>{{ log.player_id }}</td>
              <td>{{ log.action_type }}</td>
              <td class="log-detail">{{ log.details }}</td>
              <td>{{ new Date(log.created_at).toLocaleString('zh-CN') }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div class="pagination">
        <button :disabled="auditPage === 1" @click="auditPage--">
          <ChevronLeft class="icon" /> 上一页
        </button>
        <span class="page-info">第 {{ auditPage }} 页，共 {{ Math.ceil(auditTotal / auditLimit) }} 页</span>
        <button :disabled="auditPage * auditLimit >= auditTotal" @click="auditPage++">
          下一页 <ChevronRight class="icon" />
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.admin-anticheat {
  padding: 20px;
  background: #f5f5f5;
  min-height: 100vh;
}

.header {
  margin-bottom: 30px;
}

.title-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
}

.title-bar h1 {
  font-size: 28px;
  font-weight: 600;
  color: #333;
  margin: 0;
}

.title-bar .icon {
  width: 32px;
  height: 32px;
  color: #0066cc;
}

.tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
  border-bottom: 2px solid #e0e0e0;
}

.tab {
  padding: 12px 20px;
  border: none;
  background: none;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  color: #666;
  border-bottom: 3px solid transparent;
  transition: all 0.3s ease;
}

.tab.active {
  color: #0066cc;
  border-bottom-color: #0066cc;
}

.tab:hover:not(.active) {
  color: #333;
}

.tab-content {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.controls {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  flex-wrap: wrap;
  align-items: center;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 250px;
  background: #f5f5f5;
  border-radius: 6px;
  padding: 8px 12px;
}

.search-box .icon {
  width: 18px;
  height: 18px;
  color: #999;
  flex-shrink: 0;
}

.search-box input {
  border: none;
  background: none;
  outline: none;
  font-size: 14px;
  flex: 1;
}

.controls select,
.controls input[type="date"],
.controls button {
  padding: 8px 12px;
  border-radius: 6px;
  border: 1px solid #ddd;
  font-size: 14px;
  cursor: pointer;
  background: white;
}

.controls select:focus,
.controls input[type="date"]:focus {
  outline: none;
  border-color: #0066cc;
  box-shadow: 0 0 0 3px rgba(0, 102, 204, 0.1);
}

.table-container {
  overflow-x: auto;
  margin-bottom: 20px;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.data-table thead {
  background: #f9f9f9;
  border-top: 1px solid #e0e0e0;
  border-bottom: 1px solid #e0e0e0;
}

.data-table th {
  padding: 12px;
  text-align: left;
  font-weight: 600;
  color: #333;
}

.data-table td {
  padding: 12px;
  border-bottom: 1px solid #e0e0e0;
  color: #666;
}

.data-table tbody tr:hover {
  background: #f9f9f9;
}

.data-table .empty {
  text-align: center;
  color: #999;
  padding: 40px 12px;
}

.risk-score {
  font-weight: 600;
}

.badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.appeal-reason {
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.actions {
  display: flex;
  gap: 8px;
}

.btn-small {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: all 0.2s ease;
}

.btn-small .icon {
  width: 14px;
  height: 14px;
}

.btn-primary {
  background: #0066cc;
  color: white;
}

.btn-primary:hover {
  background: #0052a3;
}

.btn-success {
  background: #28a745;
  color: white;
}

.btn-success:hover {
  background: #218838;
}

.btn-danger {
  background: #dc3545;
  color: white;
}

.btn-danger:hover {
  background: #c82333;
}

.btn-secondary {
  background: #6c757d;
  color: white;
}

.btn-secondary:hover {
  background: #5a6268;
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-top: 20px;
}

.pagination button {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  transition: all 0.2s ease;
}

.pagination button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.pagination button:not(:disabled):hover {
  background: #f5f5f5;
  border-color: #0066cc;
  color: #0066cc;
}

.pagination .icon {
  width: 16px;
  height: 16px;
}

.page-info {
  color: #666;
  font-size: 14px;
  min-width: 150px;
  text-align: center;
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 8px;
  padding: 24px;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.modal-content h2 {
  margin: 0 0 20px 0;
  font-size: 20px;
  color: #333;
}

.modal-content h3 {
  margin: 20px 0 12px 0;
  font-size: 16px;
  color: #333;
}

.detail-content {
  margin-bottom: 20px;
}

.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 20px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-item label {
  font-weight: 600;
  color: #666;
  font-size: 12px;
  text-transform: uppercase;
}

.detail-item span {
  color: #333;
  font-size: 14px;
}

.detail-section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #e0e0e0;
}

.dimension-scores {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.score-item {
  display: flex;
  justify-content: space-between;
  padding: 8px 12px;
  background: #f5f5f5;
  border-radius: 4px;
}

.score-item span:first-child {
  color: #666;
  font-weight: 500;
}

.score-item span:last-child {
  color: #333;
  font-weight: 600;
}

.review-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.form-group label {
  font-weight: 600;
  color: #666;
  font-size: 14px;
}

.form-group select,
.form-group textarea {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  font-family: inherit;
}

.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #0066cc;
  box-shadow: 0 0 0 3px rgba(0, 102, 204, 0.1);
}

.form-group textarea {
  min-height: 80px;
  resize: vertical;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 20px;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s ease;
}

.btn .icon {
  width: 16px;
  height: 16px;
}

/* Config View/Edit */
.config-view {
  padding: 20px;
}

.config-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 30px;
  margin-top: 20px;
}

.config-section {
  background: #f9f9f9;
  padding: 20px;
  border-radius: 8px;
}

.config-section h3 {
  margin: 0 0 16px 0;
  font-size: 16px;
  color: #333;
}

.config-item {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #e0e0e0;
  font-size: 14px;
}

.config-item:last-child {
  border-bottom: none;
}

.config-item label {
  color: #666;
  font-weight: 500;
}

.config-item span {
  color: #333;
  font-weight: 600;
}

.config-edit {
  padding: 20px;
}

.config-edit h3 {
  margin: 0 0 20px 0;
  font-size: 18px;
  color: #333;
}

.edit-section {
  margin-bottom: 24px;
  padding-bottom: 24px;
  border-bottom: 1px solid #e0e0e0;
}

.edit-section:last-child {
  border-bottom: none;
}

.edit-section h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #333;
  font-weight: 600;
}

.edit-section .form-group {
  margin-bottom: 12px;
}

.form-group input[type="number"] {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.form-group input[type="number"]:focus {
  outline: none;
  border-color: #0066cc;
  box-shadow: 0 0 0 3px rgba(0, 102, 204, 0.1);
}

.log-detail {
  max-width: 400px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 768px) {
  .config-grid {
    grid-template-columns: 1fr;
  }

  .detail-grid {
    grid-template-columns: 1fr;
  }

  .dimension-scores {
    grid-template-columns: 1fr;
  }

  .controls {
    flex-direction: column;
  }

  .search-box {
    min-width: auto;
    flex: 1;
    width: 100%;
  }

  .pagination {
    flex-direction: column;
  }

  .modal-content {
    max-width: 90vw;
  }
}
</style>
