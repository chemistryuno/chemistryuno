<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { authAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import ws from '../utils/websocket'

const { showAlert } = useDialog()
const feedbacks = ref<any[]>([])
const loading = ref(false)

const load = async () => {
  loading.value = true
  try {
    const res = await authAPI.getMyFeedbacks()
    feedbacks.value = res.data
  } catch (e: any) {
    showAlert(e.response?.data?.error || '获取反馈失败', '错误')
  } finally {
    loading.value = false
  }
}

onMounted(load)

onMounted(() => {
  ws.connect()
  ws.on('feedback_update', (msg: any) => {
    // 如果是当前用户的反馈被更新，刷新列表
    if (msg && msg.feedback_id) {
      load()
    }
  })
})

onBeforeUnmount(() => {
  ws.off('feedback_update', () => {})
})

const canUrge = (f: any) => {
  if (!f.last_urged_at) return true
  const t = new Date(f.last_urged_at)
  const next = new Date(t.getTime() + 4 * 3600 * 1000)
  return Date.now() >= next.getTime()
}

const urge = async (id: number, idx: number) => {
  try {
    await authAPI.urgeFeedback(id)
    showAlert('催促已发送', '已发送')
    // update local item last_urged_at approximately to now
    feedbacks.value[idx].last_urged_at = new Date().toISOString().slice(0, 19).replace('T', ' ')
    feedbacks.value[idx].urge_count = (feedbacks.value[idx].urge_count || 0) + 1
  } catch (e: any) {
    showAlert(e.response?.data?.error || '催促失败', '错误')
  }
}
</script>

<template>
  <div class="min-h-screen p-6">
    <h2 class="text-2xl font-bold mb-4">我的反馈 / Messages</h2>

    <div v-if="loading" class="text-slate-500">加载中...</div>

    <div v-else>
      <div v-if="feedbacks.length === 0" class="text-slate-400">暂时没有提交的反馈。</div>

      <div class="space-y-4">
        <div v-for="(f, idx) in feedbacks" :key="f.id" class="bg-white dark:bg-[#0f0f10] p-4 rounded-lg border border-slate-200 dark:border-white/5">
          <div class="flex justify-between items-start">
            <div>
              <div class="text-sm text-slate-500">类型: {{ f.type }} · 提交于: {{ f.created_at }}</div>
              <div class="mt-2 text-base">{{ f.content }}</div>
            </div>
            <div class="text-right">
              <div class="text-sm">状态: <span class="font-medium">{{ f.status }}</span></div>
              <div v-if="f.processed_at" class="text-sm text-slate-400">处理于: {{ f.processed_at }}</div>
              <div v-if="f.processed_by" class="text-sm text-slate-400">处理人 ID: {{ f.processed_by }}</div>
            </div>
          </div>

          <div class="mt-3 flex items-center justify-between">
            <div class="text-sm text-slate-400">催促次数: {{ f.urge_count || 0 }}</div>
            <div>
              <button
                class="px-3 py-1 text-sm rounded bg-blue-500 text-white disabled:opacity-50"
                :disabled="!canUrge(f)"
                @click="urge(f.id, idx)"
              >催促管理员</button>
            </div>
          </div>
          <div v-if="f.resolution_note" class="mt-3 text-sm text-slate-300">处理说明：{{ f.resolution_note }}</div>
        </div>
      </div>
    </div>
  </div>
</template>
