<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { pageClassNames } from '@lib'
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
  Trophy,
  ChevronRight,
  FileText
} from 'lucide-vue-next'
import FeedbackButton from '../components/FeedbackButton.vue'

const router = useRouter()
const route = useRoute()
const { showAlert, showConfirm } = useDialog()
const feedbackButtonRef = ref<InstanceType<typeof FeedbackButton> | null>(null)
const feedbacks = ref<any[]>([])
const activeSurveys = ref<any[]>([])
const loading = ref(false)
const handleSurveyCompleted = () => {
  // Always re-fetch from backend so visibility follows DB state.
  load()
}

const handleFeedbackUpdate = (msg: any) => {
  // Refresh list when a feedback status update arrives.
  if (msg && msg.feedback_id) {
    load()
  }
}

const load = async () => {
  loading.value = true
  try {
    const [fbRes, svRes] = await Promise.all([
      authAPI.getMyFeedbacks(),
      authAPI.getAllActiveSurveys()
    ])
    feedbacks.value = fbRes.data
    activeSurveys.value = svRes.data || []
  } catch (e: any) {
    showAlert(e.response?.data?.error || '加载数据失败', '错误')
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

onMounted(async () => {
  await load()
  if (route.query.compose === '1') {
    feedbackButtonRef.value?.prefill('', 'general')
  }
})

onMounted(() => {
  ws.connect()
  ws.on('feedback_update', handleFeedbackUpdate)
  window.addEventListener('survey-completed', handleSurveyCompleted)
})

onBeforeUnmount(() => {
  ws.off('feedback_update', handleFeedbackUpdate)
  window.removeEventListener('survey-completed', handleSurveyCompleted)
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

// 删除异常登陆消息
const handleDismissAlert = async (id: number) => {
  const confirmed = await showConfirm('确认要消除这条安全警告吗？', '消除警告')
  if (!confirmed) return

  try {
    await authAPI.dismissFeedback(id)
    await load()
    showAlert('警告已消除', '成功')
  } catch (e: any) {
    showAlert(e.response?.data?.error || '操作失败', '错误')
  }
}

// 按类型分组反馈
const groupedFeedbacks = () => {
  const groups: Record<string, any[]> = {
    'system_alert': [],
    'other': []
  }
  
  feedbacks.value.forEach(f => {
    if (f.type === 'system_alert') {
      groups['system_alert'].push(f)
    } else {
      groups['other'].push(f)
    }
  })
  
  return groups
}

const typeLabel = (type: string) => {
  const labels: Record<string, string> = {
    'system_alert': '🔐 安全警告',
    'general': '📝 一般反馈',
    'bug': '🐛 内容反馈',
    'feature': '✨ 功能建议'
  }
  return labels[type] || type
}
</script>

<template>
  <div :class="pageClassNames.appPadded">
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

      <div class="flex flex-col md:flex-row items-start md:items-center justify-between gap-4 md:gap-0 mb-8">
        <h2 class="text-3xl font-black uppercase tracking-tighter flex items-center gap-4">
          <div class="p-3 bg-blue-500/10 rounded-2xl">
            <Megaphone class="w-8 h-8 text-blue-500" />
          </div>
          反馈与消息 / Feedbacks
          <span v-if="feedbacks.length + activeSurveys.length > 0" class="text-sm font-black bg-blue-500/10 text-blue-500 px-3 py-1 rounded-full border border-blue-500/20 tabular-nums">
            {{ feedbacks.length + activeSurveys.length }}
          </span>
        </h2>
        
        <button 
          @click="feedbackButtonRef?.prefill('', 'general')"
          class="w-full md:w-auto flex items-center justify-center gap-2 px-5 py-2.5 bg-blue-600 hover:bg-blue-500 text-white rounded-xl font-black uppercase tracking-widest text-xs transition-all shadow-lg shadow-blue-500/20"
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
        <!-- 激活的问卷调查 -->
        <div v-if="activeSurveys.length > 0" class="mb-8 space-y-3">
          <h3 class="text-xs font-black text-slate-400 dark:text-slate-500 uppercase tracking-[0.3em] flex items-center gap-2 mb-3">
            <span class="w-2 h-2 bg-indigo-500 rounded-full animate-pulse"></span>
            ACTIVE_RESEARCH_SURVEYS / 正在进行的调研
          </h3>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <router-link 
              v-for="sv in activeSurveys" 
              :key="sv.id" 
              :to="'/surveys/' + sv.id"
              class="group block bg-white dark:bg-[#111114] border border-slate-200 dark:border-indigo-500/20 p-4 md:p-6 rounded-2xl md:rounded-3xl hover:border-indigo-500 hover:shadow-xl hover:shadow-indigo-500/10 transition-all relative overflow-hidden"
            >
              <div class="absolute top-0 right-0 w-32 h-32 bg-indigo-500/5 blur-[40px] -mr-16 -mt-16 group-hover:bg-indigo-500/10 transition-all"></div>
              <div class="flex items-center justify-between relative z-10">
                <div class="flex items-center gap-3 md:gap-4">
                  <div class="w-10 md:w-12 h-10 md:h-12 bg-indigo-500/10 rounded-lg md:rounded-2xl flex items-center justify-center text-indigo-500 border border-indigo-500/20 group-hover:scale-110 transition-transform">
                    <FileText class="w-5 md:w-6 h-5 md:h-6" />
                  </div>
                  <div class="flex flex-col">
                    <span class="text-[9px] md:text-[10px] font-black text-slate-400 uppercase tracking-widest leading-none mb-1">调查问卷</span>
                    <h4 class="text-sm md:text-base font-black text-slate-900 dark:text-white group-hover:text-indigo-500 transition-colors tracking-tight">{{ sv.title }}</h4>
                  </div>
                </div>
                <ChevronRight class="w-4 md:w-5 h-4 md:h-5 text-slate-300 group-hover:text-indigo-500 group-hover:translate-x-1 transition-all shrink-0" />
              </div>
              <p class="mt-2 md:mt-4 text-xs font-medium text-slate-500 dark:text-slate-400 leading-relaxed line-clamp-2 relative z-10">{{ sv.description || '点击参与本次调查，帮助我们改进实验环境。' }}</p>
            </router-link>
          </div>
        </div>

        <!-- 安全警告模块 -->
        <div v-if="groupedFeedbacks()['system_alert'].length > 0" class="mb-8">
          <h3 class="text-xs font-black text-slate-400 dark:text-slate-500 uppercase tracking-[0.3em] flex items-center gap-2 mb-3">
            <span class="w-2 h-2 bg-red-500 rounded-full animate-pulse"></span>
            🔐 SECURITY_ALERTS / 安全警告
          </h3>
          <div class="space-y-2 md:space-y-3">
            <div 
              v-for="(f, idx) in groupedFeedbacks()['system_alert']"
              :key="f.id"
              class="bg-red-500/5 dark:bg-red-500/5 border border-red-500/20 dark:border-red-500/20 p-3 md:p-4 rounded-xl md:rounded-2xl hover:border-red-500/40 transition-all"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="flex-1">
                  <div class="flex items-center gap-2 mb-2">
                    <AlertCircle class="w-4 h-4 text-red-500 shrink-0" />
                    <span class="text-xs md:text-[10px] font-black text-red-600 dark:text-red-400 uppercase tracking-wider">
                      {{ new Date(f.created_at).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }) }}
                    </span>
                  </div>
                  <p class="text-xs md:text-sm font-bold text-slate-700 dark:text-slate-200 leading-relaxed whitespace-pre-line">
                    {{ f.content }}
                  </p>
                </div>
                <button
                  @click="handleDismissAlert(f.id)"
                  class="shrink-0 p-2 bg-red-500/10 text-red-500 hover:bg-red-500 hover:text-white rounded-lg transition-all border border-red-500/20"
                  title="消除警告"
                >
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- 其他反馈模块 -->
        <h3 class="text-xs font-black text-slate-400 dark:text-slate-500 uppercase tracking-[0.3em] flex items-center gap-2 mb-3">
          <span class="w-2 h-2 bg-blue-500 rounded-full"></span>
          MY_FEEDBACK_HISTORY / 我的反馈历史
        </h3>

        <div v-if="groupedFeedbacks()['other'].length === 0 && groupedFeedbacks()['system_alert'].length === 0" class="bg-white dark:bg-[#111114] border-2 border-dashed border-slate-200 dark:border-white/5 rounded-2xl md:rounded-[2.5rem] p-10 md:p-20 flex flex-col items-center justify-center text-center">
          <Clock class="w-12 md:w-16 h-12 md:h-16 text-slate-200 dark:text-white/5 mb-4 md:mb-6" />
          <h3 class="text-lg md:text-xl font-bold text-slate-400">尚无反馈记录</h3>
          <p class="text-sm text-slate-500 mt-2">您的反馈对我们非常重要，请在游戏过程中随时提出建议。</p>
        </div>

        <div v-else-if="groupedFeedbacks()['other'].length > 0" class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-xl md:rounded-[2rem] overflow-hidden shadow-sm">
          <div class="overflow-x-auto custom-scrollbar">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr class="text-slate-400 dark:text-slate-500 text-[9px] md:text-[10px] font-black uppercase tracking-[0.2em] border-b border-slate-100 dark:border-white/5 bg-slate-50/50 dark:bg-white/[0.02]">
                  <th class="px-3 md:px-6 py-3 md:py-5">类型 / TYPE</th>
                  <th class="px-3 md:px-6 py-3 md:py-5 w-1/3">内容 / CONTENT</th>
                  <th class="px-3 md:px-6 py-3 md:py-5">状态 / STATUS</th>
                  <th class="hidden md:table-cell px-6 py-5">通讯时间 / TIME</th>
                  <th class="px-3 md:px-6 py-3 md:py-5 text-right">操作 / ACTIONS</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 dark:divide-white/5 font-sans">
                <template v-for="(f, idx) in groupedFeedbacks()['other']" :key="f.id">
                  <tr class="hover:bg-slate-50 dark:hover:bg-white/[0.02] transition-colors group">
                    <!-- Type -->
                    <td class="px-3 md:px-6 py-3 md:py-4">
                      <span class="px-2 py-1 md:px-2.5 md:py-1 bg-slate-100 dark:bg-white/5 rounded-lg text-[8px] md:text-[10px] font-black uppercase tracking-wider text-slate-500 border border-slate-200 dark:border-white/10">
                        {{ f.type }}
                      </span>
                    </td>
                    
                    <!-- Content -->
                    <td class="px-3 md:px-6 py-3 md:py-4">
                      <p class="text-xs md:text-sm font-bold text-slate-700 dark:text-slate-200 line-clamp-2 leading-relaxed">
                        {{ f.content }}
                      </p>
                    </td>

                    <!-- Status -->
                    <td class="px-3 md:px-6 py-3 md:py-4">
                      <div class="flex items-center gap-2">
                        <div class="flex items-center gap-1.5 px-2 md:px-3 py-1 rounded-full text-[8px] md:text-[10px] font-black tracking-widest uppercase border"
                          :class="{
                            'bg-amber-100 text-amber-600 border-amber-200 dark:bg-amber-500/10 dark:text-amber-500 dark:border-amber-500/20': f.status === 'pending',
                            'bg-green-100 text-green-600 border-green-200 dark:bg-green-500/10 dark:text-green-500 dark:border-green-500/20': f.status === 'accepted',
                            'bg-slate-100 text-slate-600 border-slate-200 dark:bg-white/10 dark:text-slate-400 dark:border-white/10': f.status === 'dismissed'
                          }"
                        >
                          <component :is="f.status === 'accepted' ? CheckCircle2 : (f.status === 'pending' ? Clock : AlertCircle)" class="w-3 h-3" />
                          <span class="hidden md:inline">{{ f.status === 'accepted' ? '已接受' : (f.status === 'dismissed' ? '不予受理' : '待处理') }}</span>
                          <span class="md:hidden">{{ f.status === 'accepted' ? '✓' : (f.status === 'dismissed' ? '✗' : '⏳') }}</span>
                        </div>
                      </div>
                    </td>

                    <!-- Time (Hidden on Mobile) -->
                    <td class="hidden md:table-cell px-6 py-4">
                      <span class="text-[10px] font-bold text-slate-400 font-mono uppercase tracking-tighter">
                        {{ new Date(f.created_at).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }) }}
                      </span>
                    </td>

                    <!-- Actions -->
                    <td class="px-3 md:px-6 py-3 md:py-4 text-right">
                      <div class="flex items-center justify-end gap-1 md:gap-2">
                        <!-- Urge Button -->
                        <button
                          v-if="f.status === 'pending'"
                          @click="urge(f.id, idx)"
                          :disabled="!canUrge(f)"
                          class="p-1.5 md:p-2 rounded-lg md:rounded-xl transition-all disabled:opacity-30 flex items-center gap-1 md:gap-2"
                          :class="canUrge(f) 
                            ? 'bg-blue-500/10 text-blue-500 hover:bg-blue-500 hover:text-white border border-blue-500/20' 
                            : 'bg-slate-100 dark:bg-white/5 text-slate-400 border border-transparent'"
                          :title="canUrge(f) ? '催促' : '请稍后'"
                        >
                          <BellRing class="w-3 md:w-3.5 h-3 md:h-3.5" />
                          <span v-if="f.urge_count > 0" class="text-[7px] md:text-[8px] font-black hidden md:inline">{{ f.urge_count }}</span>
                        </button>

                        <!-- Withdraw Button -->
                        <button
                          v-if="f.status === 'pending'"
                          @click="handleWithdraw(f.id)"
                          class="p-1.5 md:p-2 bg-rose-500/10 text-rose-500 hover:bg-rose-500 hover:text-white rounded-lg md:rounded-xl transition-all border border-rose-500/20"
                          title="撤回反馈"
                        >
                          <Trash2 class="w-3 md:w-3.5 h-3 md:h-3.5" />
                        </button>

                        <div v-else class="text-[8px] text-slate-400 font-bold uppercase tracking-widest italic pr-2 hidden md:block">
                           <User v-if="f.processed_at" class="w-3 h-3 inline mr-1" />
                           Archived
                        </div>
                      </div>
                    </td>
                  </tr>

                  <!-- Resolution Note Row (Conditional) -->
                  <tr v-if="f.resolution_note" class="bg-blue-500/[0.02] dark:bg-blue-500/[0.01] hover:bg-blue-500/[0.05] transition-colors">
                    <td colspan="5" class="px-3 md:px-6 py-2 md:py-3 border-t border-blue-500/5">
                      <div class="flex items-start gap-2 md:gap-4">
                        <div class="shrink-0 pt-1">
                          <div class="w-1.5 h-1.5 rounded-full bg-blue-500/40 shadow-[0_0_8px_rgba(59,130,246,0.5)]"></div>
                        </div>
                        <div class="flex-1 space-y-1 min-w-0">
                          <div class="text-[8px] md:text-[9px] font-black text-blue-500/60 uppercase tracking-[0.2em] italic">管理回复 / COMMAND_CENTRAL_UPLINK</div>
                          <p class="text-xs text-slate-600 dark:text-blue-300/80 leading-relaxed italic break-words">
                            "{{ f.resolution_note }}"
                          </p>
                        </div>
                        <div class="text-[7px] md:text-[8px] text-slate-400 font-mono self-end pb-1 pr-2 uppercase shrink-0">
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
