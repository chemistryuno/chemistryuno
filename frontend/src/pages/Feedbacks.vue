<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import ws from '../utils/websocket'
import {
  ArrowLeft,
  Megaphone,
  Clock,
  User,
  CheckCircle2,
  AlertCircle,
  BellRing,
  Trash2,
  Trophy
} from 'lucide-vue-next'
import FeedbackButton from '../components/FeedbackButton.vue'

const router = useRouter()
const { showAlert, showConfirm } = useDialog()
const feedbackButtonRef = ref<InstanceType<typeof FeedbackButton> | null>(null)
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

const handleWithdraw = async (id: number) => {
  const confirmed = await showConfirm('确认要撤回这条反馈吗？', '撤回反馈')
  if (!confirmed) return

  try {
    await authAPI.withdrawFeedback(id)
    await load()
    showAlert('反馈已成功撤回', '撤回成功')
  } catch (e: any) {
    showAlert(e.response?.data?.error || '撤回失败', '错误')
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
  if (f.status !== 'pending') return false
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
  <div class="min-h-screen bg-slate-50 dark:bg-[#0a0a0c] text-slate-900 dark:text-white p-4 md:p-8 selection:bg-blue-500/30">
    <!-- Background Effects -->
    <div class="fixed inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-[-10%] right-[-10%] w-[50%] h-[50%] bg-blue-500/5 rounded-full blur-[120px]" />
      <div class="absolute bottom-[-10%] left-[-10%] w-[50%] h-[50%] bg-purple-500/5 rounded-full blur-[120px]" />
      <div class="absolute inset-0 bg-[url('/noise.svg')] opacity-20 brightness-50 contrast-150" />
    </div>

    <div class="max-w-4xl mx-auto relative z-10">
      <div class="mb-10 flex items-center justify-between">
        <!-- Back Button -->
        <button 
          @click="router.push('/')" 
          class="group flex items-center gap-3 text-slate-400 hover:text-slate-900 dark:hover:text-white transition-all px-4 py-2 rounded-xl hover:bg-white dark:hover:bg-white/5 border border-transparent hover:border-slate-200 dark:hover:border-white/10"
        >
          <ArrowLeft class="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
          <span class="font-bold tracking-wider uppercase text-xs">返回大厅</span>
        </button>

        <router-link 
          to="/ranking" 
          class="flex items-center gap-2 px-4 py-2 bg-amber-500/10 border border-amber-500/20 rounded-xl text-amber-500 hover:bg-amber-500/20 transition-all group shadow-sm"
        >
          <Trophy class="w-4 h-4 group-hover:scale-110 transition-transform" />
          <span class="text-[10px] font-black uppercase tracking-widest">全球排名</span>
        </router-link>
      </div>

      <div class="flex items-center justify-between mb-8">
        <h2 class="text-3xl font-black uppercase tracking-tighter flex items-center gap-4">
          <div class="p-3 bg-blue-500/10 rounded-2xl">
            <Megaphone class="w-8 h-8 text-blue-500" />
          </div>
          反馈与消息 / Feedbacks
          <span v-if="feedbacks.length > 0" class="text-sm font-black bg-blue-500/10 text-blue-500 px-3 py-1 rounded-full border border-blue-500/20 tabular-nums">
            {{ feedbacks.length }}
          </span>
        </h2>
        
        <button 
          @click="feedbackButtonRef?.prefill('', 'general')"
          class="flex items-center gap-2 px-5 py-2.5 bg-blue-600 hover:bg-blue-500 text-white rounded-xl font-black uppercase tracking-widest text-xs transition-all shadow-lg shadow-blue-500/20"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-message-circle"><path d="M7.9 20A9 9 0 1 0 4 16.1L2 22Z"/></svg>
          撰写反馈 / New
        </button>
      </div>

      <div v-if="loading" class="flex flex-col items-center justify-center py-20">
        <div class="w-10 h-10 border-4 border-blue-500/20 border-t-blue-500 rounded-full animate-spin mb-4"></div>
        <p class="text-slate-400 font-medium">实验室正在检索记录...</p>
      </div>

      <div v-else>
        <div v-if="feedbacks.length === 0" class="bg-white dark:bg-[#111114] border-2 border-dashed border-slate-200 dark:border-white/5 rounded-[2.5rem] p-20 flex flex-col items-center justify-center text-center">
          <Clock class="w-16 h-16 text-slate-200 dark:text-white/5 mb-6" />
          <h3 class="text-xl font-bold text-slate-400">尚无反馈记录</h3>
          <p class="text-slate-500 mt-2">您的反馈对我们非常重要，请在游戏过程中随时提出建议。</p>
        </div>

        <div v-else class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2rem] overflow-hidden shadow-sm">
          <div class="overflow-x-auto custom-scrollbar">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr class="text-slate-400 dark:text-slate-500 text-[10px] font-black uppercase tracking-[0.2em] border-b border-slate-100 dark:border-white/5 bg-slate-50/50 dark:bg-white/[0.02]">
                  <th class="px-6 py-5">类型 / TYPE</th>
                  <th class="px-6 py-5 w-1/3">内容 / CONTENT</th>
                  <th class="px-6 py-5">状态 / STATUS</th>
                  <th class="px-6 py-5">通讯时间 / TIME</th>
                  <th class="px-6 py-5 text-right">协同操作 / ACTIONS</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 dark:divide-white/5 font-sans">
                <template v-for="(f, idx) in feedbacks" :key="f.id">
                  <tr class="hover:bg-slate-50 dark:hover:bg-white/[0.02] transition-colors group">
                    <!-- Type -->
                    <td class="px-6 py-4">
                      <span class="px-2.5 py-1 bg-slate-100 dark:bg-white/5 rounded-lg text-[10px] font-black uppercase tracking-wider text-slate-500 border border-slate-200 dark:border-white/10">
                        {{ f.type }}
                      </span>
                    </td>
                    
                    <!-- Content -->
                    <td class="px-6 py-4">
                      <p class="text-sm font-bold text-slate-700 dark:text-slate-200 line-clamp-2 leading-relaxed">
                        {{ f.content }}
                      </p>
                    </td>

                    <!-- Status -->
                    <td class="px-6 py-4">
                      <div class="flex items-center gap-2">
                        <div class="flex items-center gap-1.5 px-3 py-1 rounded-full text-[10px] font-black tracking-widest uppercase border"
                          :class="{
                            'bg-amber-100 text-amber-600 border-amber-200 dark:bg-amber-500/10 dark:text-amber-500 dark:border-amber-500/20': f.status === 'pending',
                            'bg-green-100 text-green-600 border-green-200 dark:bg-green-500/10 dark:text-green-500 dark:border-green-500/20': f.status === 'accepted',
                            'bg-slate-100 text-slate-600 border-slate-200 dark:bg-white/10 dark:text-slate-400 dark:border-white/10': f.status === 'dismissed'
                          }"
                        >
                          <component :is="f.status === 'accepted' ? CheckCircle2 : (f.status === 'pending' ? Clock : AlertCircle)" class="w-3 h-3" />
                          {{ f.status === 'accepted' ? '已接受' : (f.status === 'dismissed' ? '不予受理' : '待处理') }}
                        </div>
                      </div>
                    </td>

                    <!-- Time -->
                    <td class="px-6 py-4">
                      <span class="text-[10px] font-bold text-slate-400 font-mono uppercase tracking-tighter">
                        {{ new Date(f.created_at).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }) }}
                      </span>
                    </td>

                    <!-- Actions -->
                    <td class="px-6 py-4 text-right">
                      <div class="flex items-center justify-end gap-2">
                        <!-- Urge Button -->
                        <button
                          v-if="f.status === 'pending'"
                          @click="urge(f.id, idx)"
                          :disabled="!canUrge(f)"
                          class="p-2 rounded-xl transition-all disabled:opacity-30 flex items-center gap-2"
                          :class="canUrge(f) 
                            ? 'bg-blue-500/10 text-blue-500 hover:bg-blue-500 hover:text-white border border-blue-500/20' 
                            : 'bg-slate-100 dark:bg-white/5 text-slate-400 border border-transparent'"
                        >
                          <BellRing class="w-3.5 h-3.5" />
                          <span v-if="f.urge_count > 0" class="text-[8px] font-black">{{ f.urge_count }}</span>
                        </button>

                        <!-- Withdraw Button -->
                        <button
                          v-if="f.status === 'pending'"
                          @click="handleWithdraw(f.id)"
                          class="p-2 bg-rose-500/10 text-rose-500 hover:bg-rose-500 hover:text-white rounded-xl transition-all border border-rose-500/20"
                          title="撤回反馈"
                        >
                          <Trash2 class="w-3.5 h-3.5" />
                        </button>

                        <div v-else class="text-[10px] text-slate-400 font-bold uppercase tracking-widest italic pr-2">
                           <User v-if="f.processed_at" class="w-3 h-3 inline mr-1" />
                           Archived
                        </div>
                      </div>
                    </td>
                  </tr>

                  <!-- Resolution Note Row (Conditional) -->
                  <tr v-if="f.resolution_note" class="bg-blue-500/[0.02] dark:bg-blue-500/[0.01]">
                    <td colspan="5" class="px-6 py-3 border-t border-blue-500/5">
                      <div class="flex items-start gap-4">
                        <div class="shrink-0 pt-1">
                          <div class="w-1.5 h-1.5 rounded-full bg-blue-500/40 shadow-[0_0_8px_rgba(59,130,246,0.5)]"></div>
                        </div>
                        <div class="flex-1 space-y-1">
                          <div class="text-[9px] font-black text-blue-500/60 uppercase tracking-[0.2em] italic">管理回复 / COMMAND_CENTRAL_UPLINK</div>
                          <p class="text-xs text-slate-600 dark:text-blue-300/80 leading-relaxed italic">
                            "{{ f.resolution_note }}"
                          </p>
                        </div>
                        <div class="text-[8px] text-slate-400 font-mono self-end pb-1 pr-2 uppercase">
                           PROCESSED_AT: {{ f.processed_at ? new Date(f.processed_at).toLocaleString() : 'PENDING' }}
                        </div>
                      </div>
                    </td>
                  </tr>
                </template>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <div class="mt-12 text-center">
        <p class="text-xs text-slate-400 uppercase tracking-[0.2em] font-medium opacity-50">
          CHEMISTRY UNO MENDELEEF · FEEDBACK SYSTEM
        </p>
      </div>
    </div>

    <FeedbackButton ref="feedbackButtonRef" @submitted="load" />
  </div>
</template>
