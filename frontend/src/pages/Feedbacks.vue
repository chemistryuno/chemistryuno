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

const router = useRouter()
const { showAlert, showConfirm } = useDialog()
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
  if (f.status !== 'unread') return false
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
      <div class="absolute inset-0 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-20 brightness-50 contrast-150" />
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
        </h2>
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

        <div class="grid gap-6">
          <div v-for="(f, idx) in feedbacks" :key="f.id" 
            class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2rem] p-8 shadow-sm transition-all hover:shadow-xl hover:scale-[1.01] group"
          >
            <div class="flex flex-col md:flex-row justify-between gap-6">
              <div class="flex-1">
                <div class="flex items-center gap-3 mb-4">
                  <span class="px-3 py-1 bg-slate-100 dark:bg-white/5 rounded-full text-xs font-bold uppercase tracking-wider text-slate-500">
                    {{ f.type }}
                  </span>
                  <span class="flex items-center gap-1.5 text-xs text-slate-400">
                    <Clock class="w-3.5 h-3.5" />
                    {{ f.created_at }}
                  </span>
                </div>
                
                <p class="text-lg font-medium leading-relaxed text-slate-700 dark:text-slate-200">
                  {{ f.content }}
                </p>

                <div v-if="f.resolution_note" class="mt-6 p-5 bg-blue-500/5 dark:bg-blue-500/[0.03] border border-blue-500/10 rounded-2xl relative">
                  <div class="absolute -top-3 left-4 px-2 bg-slate-50 dark:bg-[#0a0a0c] text-[10px] font-black uppercase tracking-tighter text-blue-500">
                    处理回复 / Resolution Note
                  </div>
                  <p class="text-sm text-slate-600 dark:text-blue-200/80 leading-loose italic">
                    "{{ f.resolution_note }}"
                  </p>
                </div>
              </div>

              <div class="md:w-64 flex flex-col justify-between items-end gap-6 border-l border-slate-100 dark:border-white/5 md:pl-8">
                <div class="text-right flex flex-col items-end gap-2">
                  <div class="flex items-center gap-2 px-4 py-1.5 rounded-full text-sm font-bold tracking-tight"
                    :class="{
                      'bg-amber-100 text-amber-600 dark:bg-amber-500/10 dark:text-amber-500': f.status === 'unread',
                      'bg-green-100 text-green-600 dark:bg-green-500/10 dark:text-green-500': f.status === 'accepted',
                      'bg-slate-100 text-slate-600 dark:bg-white/10 dark:text-slate-400': f.status === 'dismissed'
                    }"
                  >
                    <component :is="f.status === 'accepted' ? CheckCircle2 : (f.status === 'unread' ? Clock : AlertCircle)" class="w-4 h-4" />
                    {{ f.status === 'accepted' ? '已接受' : (f.status === 'dismissed' ? '不予受理' : '待处理') }}
                  </div>
                  
                  <div v-if="f.processed_at" class="flex items-center gap-1.5 text-xs text-slate-400 mt-1">
                    <User class="w-3.5 h-3.5" />
                    管理员已检阅
                  </div>
                </div>

                <div class="w-full">
                  <button
                    class="w-full flex items-center justify-center gap-2 py-3 px-6 rounded-xl font-bold transition-all disabled:opacity-30 disabled:cursor-not-allowed group/btn"
                    :class="canUrge(f) 
                      ? 'bg-blue-500 text-white shadow-[0_4px_20px_rgba(59,130,246,0.3)] hover:shadow-[0_8px_30px_rgba(59,130,246,0.4)] hover:-translate-y-0.5 active:translate-y-0' 
                      : 'bg-slate-200 dark:bg-white/5 text-slate-400'"
                    :disabled="!canUrge(f)"
                    @click="urge(f.id, idx)"
                  >
                    <BellRing class="w-4 h-4 transition-transform group-hover/btn:rotate-12" />
                    {{ canUrge(f) ? '催促管理员' : '已发送催促' }}
                  </button>
                  <p v-if="f.urge_count > 0" class="text-[10px] text-center mt-2 text-slate-400 font-bold uppercase tracking-widest">
                    已催促 {{ f.urge_count }} 次
                  </p>
                  
                  <button
                    v-if="f.status === 'unread'"
                    @click="handleWithdraw(f.id)"
                    class="w-full mt-4 flex items-center justify-center gap-2 py-2 px-6 rounded-xl font-bold transition-all text-red-500 hover:bg-red-500/10 text-xs uppercase tracking-widest"
                  >
                    <Trash2 class="w-3.5 h-3.5" />
                    撤回反馈
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="mt-12 text-center">
        <p class="text-xs text-slate-400 uppercase tracking-[0.2em] font-medium opacity-50">
          CHEMISTRY UNO MENDELEEF · FEEDBACK SYSTEM
        </p>
      </div>
    </div>
  </div>
</template>
