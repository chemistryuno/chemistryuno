<template>
  <div class="appeal-form-wrapper">
    <!-- 申诉表单 -->
    <div v-if="!showHistory" class="appeal-form">
      <div class="form-header">
        <h2 class="form-title">提交申诉</h2>
        <p class="form-subtitle">如果您认为处罚有误，可以提交申诉进行复审</p>
      </div>

      <form @submit.prevent="submitAppeal" class="form-content">
        <!-- 申诉原因 -->
        <div class="form-group">
          <label class="form-label">申诉原因 *</label>
          <textarea
            v-model="formData.reason"
            class="form-textarea"
            placeholder="请详细说明您认为处罚有误的原因..."
            rows="5"
            required
          ></textarea>
          <div class="char-count">{{ formData.reason.length }}/500</div>
        </div>

        <!-- 证据说明 -->
        <div class="form-group">
          <label class="form-label">证据说明</label>
          <textarea
            v-model="formData.evidence"
            class="form-textarea"
            placeholder="例如：对局回放、截图、日志等证据的说明..."
            rows="4"
          ></textarea>
          <div class="char-count">{{ formData.evidence.length }}/300</div>
        </div>

        <!-- 联系方式 -->
        <div class="form-group">
          <label class="form-label">联系方式</label>
          <input
            v-model="formData.contact"
            type="email"
            class="form-input"
            placeholder="邮箱地址（可选）"
          />
        </div>

        <!-- 提交按钮 -->
        <div class="form-actions">
          <button
            type="submit"
            class="btn btn-primary"
            :disabled="isSubmitting || !formData.reason.trim()"
          >
            <span v-if="!isSubmitting">提交申诉</span>
            <span v-else>提交中...</span>
          </button>
          <button
            type="button"
            class="btn btn-secondary"
            @click="showHistory = true"
          >
            查看申诉历史
          </button>
        </div>
      </form>

      <!-- 提示信息 -->
      <div class="form-tips">
        <div class="tip-item">
          <span class="tip-icon">ℹ</span>
          <span class="tip-text">申诉将由管理员进行审核，通常在 24-48 小时内处理</span>
        </div>
        <div class="tip-item">
          <span class="tip-icon">✓</span>
          <span class="tip-text">提供详细的证据和说明可以提高申诉成功率</span>
        </div>
        <div class="tip-item">
          <span class="tip-icon">⚠</span>
          <span class="tip-text">虚假申诉可能导致账号进一步处罚</span>
        </div>
      </div>
    </div>

    <!-- 申诉历史 -->
    <div v-else class="appeal-history">
      <div class="history-header">
        <h2 class="history-title">申诉历史</h2>
        <button class="btn-back" @click="showHistory = false">← 返回</button>
      </div>

      <div v-if="appeals.length === 0" class="empty-state">
        <div class="empty-icon">📋</div>
        <p class="empty-text">暂无申诉记录</p>
      </div>

      <div v-else class="appeals-list">
        <div
          v-for="appeal in appeals"
          :key="appeal.id"
          :class="['appeal-item', `status-${appeal.status}`]"
        >
          <div class="appeal-header">
            <div class="appeal-info">
              <span class="appeal-id">#{{ appeal.id }}</span>
              <span class="appeal-date">{{ formatDate(appeal.createdAt) }}</span>
            </div>
            <span :class="['appeal-status', `status-${appeal.status}`]">
              {{ getStatusLabel(appeal.status) }}
            </span>
          </div>

          <div class="appeal-content">
            <div class="appeal-section">
              <span class="section-label">申诉原因：</span>
              <p class="section-text">{{ appeal.reason }}</p>
            </div>

            <div v-if="appeal.evidence" class="appeal-section">
              <span class="section-label">证据说明：</span>
              <p class="section-text">{{ appeal.evidence }}</p>
            </div>

            <div v-if="appeal.remark" class="appeal-section">
              <span class="section-label">审核意见：</span>
              <p class="section-text remark">{{ appeal.remark }}</p>
            </div>

            <div class="appeal-meta">
              <span v-if="appeal.reviewedAt" class="meta-item">
                审核时间：{{ formatDate(appeal.reviewedAt) }}
              </span>
              <span v-if="appeal.reviewedBy" class="meta-item">
                审核人：{{ appeal.reviewedBy }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface Appeal {
  id: number
  reason: string
  evidence?: string
  contact?: string
  status: 'pending' | 'approved' | 'rejected'
  remark?: string
  reviewedAt?: string
  reviewedBy?: string
  createdAt: string
}

interface FormData {
  reason: string
  evidence: string
  contact: string
}

const showHistory = ref(false)
const isSubmitting = ref(false)
const appeals = ref<Appeal[]>([])

const formData = ref<FormData>({
  reason: '',
  evidence: '',
  contact: ''
})

const submitAppeal = async () => {
  if (!formData.value.reason.trim()) {
    return
  }

  isSubmitting.value = true

  try {
    // 调用 API 提交申诉
    const response = await fetch('/api/game/room_001/appeal', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        reason: formData.value.reason,
        evidence: formData.value.evidence,
        contact: formData.value.contact
      })
    })

    if (response.ok) {
      // 重置表单
      formData.value = {
        reason: '',
        evidence: '',
        contact: ''
      }

      // 显示成功提示
      emit('appeal-submitted', {
        success: true,
        message: '申诉已提交，请等待管理员审核'
      })

      // 刷新申诉历史
      await loadAppeals()
    } else {
      emit('appeal-submitted', {
        success: false,
        message: '申诉提交失败，请稍后重试'
      })
    }
  } catch (error) {
    console.error('提交申诉失败:', error)
    emit('appeal-submitted', {
      success: false,
      message: '网络错误，请检查连接'
    })
  } finally {
    isSubmitting.value = false
  }
}

const loadAppeals = async () => {
  try {
    const response = await fetch('/api/player/appeals')
    if (response.ok) {
      appeals.value = await response.json()
    }
  } catch (error) {
    console.error('加载申诉历史失败:', error)
  }
}

const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const getStatusLabel = (status: string) => {
  const labels: Record<string, string> = {
    pending: '待审核',
    approved: '已批准',
    rejected: '已拒绝'
  }
  return labels[status] || status
}

const emit = defineEmits<{
  'appeal-submitted': [data: { success: boolean; message: string }]
}>()

onMounted(() => {
  loadAppeals()
})
</script>

<style scoped>
.appeal-form-wrapper {
  width: 100%;
  max-width: 600px;
  margin: 0 auto;
  padding: 20px;
}

/* ==================== 表单样式 ==================== */
.appeal-form {
  background: linear-gradient(135deg,
    rgba(59, 130, 246, 0.08) 0%,
    rgba(37, 99, 235, 0.12) 100%);
  border: 1px solid rgba(59, 130, 246, 0.2);
  border-radius: 12px;
  padding: 24px;
  backdrop-filter: blur(10px);
}

.form-header {
  margin-bottom: 24px;
}

.form-title {
  font-size: 20px;
  font-weight: 700;
  color: #ffffff;
  margin: 0 0 8px 0;
}

.form-subtitle {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.7);
  margin: 0;
}

.form-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* ==================== 表单组 ==================== */
.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-label {
  font-size: 14px;
  font-weight: 600;
  color: #ffffff;
}

.form-input,
.form-textarea {
  padding: 12px 14px;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(255, 255, 255, 0.05);
  color: #ffffff;
  font-family: inherit;
  font-size: 14px;
  transition: all 0.2s ease;
}

.form-input::placeholder,
.form-textarea::placeholder {
  color: rgba(255, 255, 255, 0.5);
}

.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: rgba(59, 130, 246, 0.5);
  background: rgba(59, 130, 246, 0.1);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-textarea {
  resize: vertical;
  min-height: 100px;
}

.char-count {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
  text-align: right;
}

/* ==================== 按钮 ==================== */
.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 8px;
}

.btn {
  padding: 12px 20px;
  border-radius: 8px;
  border: none;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  flex: 1;
}

.btn-primary {
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: #ffffff;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(59, 130, 246, 0.4);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.1);
  color: #ffffff;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.15);
  border-color: rgba(255, 255, 255, 0.3);
}

/* ==================== 提示信息 ==================== */
.form-tips {
  margin-top: 24px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.tip-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  border-left: 3px solid rgba(59, 130, 246, 0.5);
}

.tip-icon {
  flex-shrink: 0;
  font-size: 16px;
  line-height: 1.4;
}

.tip-text {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.7);
  line-height: 1.4;
}

/* ==================== 申诉历史 ==================== */
.appeal-history {
  background: linear-gradient(135deg,
    rgba(59, 130, 246, 0.08) 0%,
    rgba(37, 99, 235, 0.12) 100%);
  border: 1px solid rgba(59, 130, 246, 0.2);
  border-radius: 12px;
  padding: 24px;
  backdrop-filter: blur(10px);
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.history-title {
  font-size: 20px;
  font-weight: 700;
  color: #ffffff;
  margin: 0;
}

.btn-back {
  padding: 8px 16px;
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(255, 255, 255, 0.05);
  color: #ffffff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-back:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.3);
}

/* ==================== 空状态 ==================== */
.empty-state {
  text-align: center;
  padding: 40px 20px;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 12px;
}

.empty-text {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.6);
  margin: 0;
}

/* ==================== 申诉列表 ==================== */
.appeals-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.appeal-item {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 16px;
  transition: all 0.2s ease;
}

.appeal-item:hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.15);
}

.appeal-item.status-pending {
  border-left: 3px solid #f97316;
}

.appeal-item.status-approved {
  border-left: 3px solid #22c55e;
}

.appeal-item.status-rejected {
  border-left: 3px solid #ef4444;
}

.appeal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.appeal-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.appeal-id {
  font-size: 13px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.8);
}

.appeal-date {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
}

.appeal-status {
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.appeal-status.status-pending {
  background: rgba(249, 115, 22, 0.2);
  color: #f97316;
}

.appeal-status.status-approved {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
}

.appeal-status.status-rejected {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

.appeal-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.appeal-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.section-label {
  font-size: 12px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.7);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.section-text {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.8);
  line-height: 1.5;
  margin: 0;
  word-break: break-word;
}

.section-text.remark {
  padding: 10px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 4px;
  border-left: 2px solid rgba(59, 130, 246, 0.5);
}

.appeal-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  padding-top: 8px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
}

.meta-item {
  display: flex;
  align-items: center;
}

/* ==================== 响应式 ==================== */
@media (max-width: 640px) {
  .appeal-form-wrapper {
    padding: 12px;
  }

  .appeal-form,
  .appeal-history {
    padding: 16px;
  }

  .form-title,
  .history-title {
    font-size: 18px;
  }

  .form-actions {
    flex-direction: column;
  }

  .btn {
    flex: 1;
  }

  .appeal-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .appeal-status {
    align-self: flex-start;
  }
}
</style>
