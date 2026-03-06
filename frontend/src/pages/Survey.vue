<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { authAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import {
  ArrowLeft,
  FileText,
  Send,
  CheckCircle2
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const { showAlert, showConfirm } = useDialog()

interface Answer {
  question_id: number
  answer: string | string[]
}

const surveyID = parseInt(route.params.id as string)
const survey = ref<any>(null)
const answers = ref<Answer[]>([])
const loading = ref(true)
const submitting = ref(false)
const completed = ref(false)

const loadSurvey = async () => {
  try {
    const res = await authAPI.getSurveyDetail(surveyID)
    survey.value = res.data
    // 初始化答案数组
    answers.value = survey.value.questions.map((q: any) => ({
      question_id: q.ID,
      answer: q.type === 'checkbox' ? [] : ''
    }))
  } catch (error: any) {
    showAlert(error.response?.data?.error || '无法加载问卷详情', '错误')
    router.push('/feedbacks')
  } finally {
    loading.value = false
  }
}

const handleSubmit = async () => {
  // 验证必填 (目前逻辑上简单认为题目都需要回答，可以根据需求调整)
  for (let i = 0; i < answers.value.length; i++) {
    const a = answers.value[i]
    if (Array.isArray(a.answer)) {
      if (a.answer.length === 0) {
        showAlert(`请完成第 ${i + 1} 题`, '验证失败')
        return
      }
    } else if (!a.answer) {
      showAlert(`请完成第 ${i + 1} 题`, '验证失败')
      return
    }
  }

  const confirmed = await showConfirm('确认提交这些回答吗？提交后无法更改。', '提交确认')
  if (!confirmed) return

  submitting.value = true
  try {
    const formattedAnswers = answers.value.map(a => ({
      question_id: a.question_id,
      answer: Array.isArray(a.answer) ? JSON.stringify(a.answer) : a.answer
    }))
    await authAPI.submitSurveyAnswers(surveyID, formattedAnswers)
    completed.value = true
    await showAlert('感谢您的反馈！您的回答已安全记录。', '提交成功')
  } catch (error: any) {
    showAlert(error.response?.data?.error || '提交失败', '错误')
  } finally {
    submitting.value = false
  }
}

onMounted(loadSurvey)

const parseOptions = (optionsStr: string) => {
  try {
    return JSON.parse(optionsStr)
  } catch (e) {
    return []
  }
}
</script>

<template>
  <div class="min-h-screen bg-slate-50 dark:bg-[#0a0a0c] text-slate-900 dark:text-white p-4 md:p-8 selection:bg-indigo-500/30">
    <!-- Background Effects -->
    <div class="fixed inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-[-10%] right-[-10%] w-[50%] h-[50%] bg-indigo-500/5 rounded-full blur-[120px]" />
      <div class="absolute bottom-[-10%] left-[-10%] w-[50%] h-[50%] bg-blue-500/5 rounded-full blur-[120px]" />
      <div class="absolute inset-0 bg-[url('/noise.svg')] opacity-20 brightness-50 contrast-150" />
    </div>

    <div class="max-w-2xl mx-auto relative z-10">
      <div class="mb-6 flex items-center justify-between">
        <button 
          @click="router.push('/feedbacks')" 
          class="group flex items-center gap-2 text-slate-400 hover:text-slate-900 dark:hover:text-white transition-all px-3 py-1.5 rounded-lg hover:bg-white dark:hover:bg-white/5 border border-transparent hover:border-slate-200 dark:hover:border-white/10"
        >
          <ArrowLeft class="w-4 h-4 group-hover:-translate-x-0.5 transition-transform" />
          <span class="font-bold tracking-wider uppercase text-[10px]">返回列表</span>
        </button>
      </div>

      <div v-if="loading" class="flex flex-col items-center justify-center py-20">
        <div class="w-10 h-10 border-4 border-indigo-500/20 border-t-indigo-500 rounded-full animate-spin mb-4"></div>
        <p class="text-slate-400 font-black uppercase tracking-[0.2em] text-[10px]">Initializing...</p>
      </div>

      <div v-else-if="completed" class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 p-8 rounded-[2rem] text-center shadow-xl">
        <div class="w-16 h-16 bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 rounded-2xl flex items-center justify-center mx-auto mb-6">
          <CheckCircle2 class="w-8 h-8" />
        </div>
        <h2 class="text-2xl font-black text-slate-900 dark:text-white mb-2 uppercase italic">Success</h2>
        <p class="text-xs text-slate-500 dark:text-slate-400 mb-8 max-w-xs mx-auto font-medium">感谢您的配合。您的反馈已被安全记录。</p>
        <button @click="router.push('/feedbacks')" class="px-6 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl font-black uppercase tracking-widest text-xs transition-all shadow-lg shadow-indigo-500/20 active:scale-95">
          返回主界面
        </button>
      </div>

      <div v-else-if="survey" class="space-y-6 mb-16 animate-in fade-in slide-in-from-bottom-2 duration-300">
        <!-- Header -->
        <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 p-6 md:p-8 rounded-[2.5rem] shadow-sm relative overflow-hidden group">
          <div class="absolute top-0 right-0 w-48 h-48 bg-indigo-500/5 blur-[60px] -mr-24 -mt-24 group-hover:bg-indigo-500/10 transition-all"></div>
          <div class="relative z-10">
            <div class="flex items-center gap-3 mb-4">
              <div class="w-10 h-10 bg-indigo-500/10 border border-indigo-500/20 rounded-xl flex items-center justify-center text-indigo-500">
                <FileText class="w-5 h-5" />
              </div>
              <div>
                <span class="text-[9px] font-black text-slate-400 uppercase tracking-[0.3em] block mb-0.5">Research Protocol</span>
                <h1 class="text-xl font-black text-slate-900 dark:text-white uppercase italic tracking-tighter">{{ survey.title }}</h1>
                <div v-if="survey.reward_points > 0 || survey.reward_exp > 0" class="flex items-center gap-2 mt-2">
                  <div v-if="survey.reward_points > 0" class="px-2 py-0.5 bg-amber-500/10 border border-amber-500/30 rounded-md flex items-center gap-1">
                    <div class="w-1.5 h-1.5 rounded-full bg-amber-500 shadow-[0_0_5px_rgba(245,158,11,1)]"></div>
                    <span class="text-[9px] font-black text-amber-500 uppercase tracking-tighter">{{ survey.reward_points }} PTS</span>
                  </div>
                  <div v-if="survey.reward_exp > 0" class="px-2 py-0.5 bg-indigo-500/10 border border-indigo-500/30 rounded-md flex items-center gap-1">
                    <div class="w-1.5 h-1.5 rounded-full bg-indigo-500 shadow-[0_0_5px_rgba(99,102,241,1)]"></div>
                    <span class="text-[9px] font-black text-indigo-500 uppercase tracking-tighter">{{ survey.reward_exp }} EXP</span>
                  </div>
                </div>
              </div>
            </div>
            <p class="text-[11px] text-slate-500 dark:text-slate-400 font-bold leading-relaxed max-w-xl italic">
              {{ survey.description || '研究员：请根据您的实际体验回答以下问题。您的诚实反馈对实验室环境的优化至关重要。' }}
            </p>
          </div>
        </div>

        <!-- Questions -->
        <div class="space-y-4">
          <div v-for="(q, idx) in (survey.questions as any[])" :key="q.ID" class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 p-6 rounded-[1.5rem] shadow-sm hover:border-indigo-500/20 transition-all group">
            <div class="flex items-start gap-4">
              <div class="w-8 h-8 rounded-lg bg-slate-100 dark:bg-white/5 flex items-center justify-center text-slate-400 dark:text-slate-500 font-mono font-black shrink-0 border border-slate-200 dark:border-white/5 text-[10px]">
                {{ String(idx + 1).padStart(2, '0') }}
              </div>
              <div class="flex-1 space-y-4">
                <h4 class="text-sm font-black text-slate-900 dark:text-white leading-tight uppercase italic tracking-tight">{{ q.title }}</h4>
                
                <!-- Radio -->
                <div v-if="q.type === 'radio'" class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                  <label 
                    v-for="opt in parseOptions(q.options)" 
                    :key="opt"
                    class="flex items-center gap-3 p-3 rounded-xl border border-slate-100 dark:border-white/5 hover:bg-slate-50 dark:hover:bg-white/[0.02] cursor-pointer transition-all active:scale-[0.98]"
                    :class="answers[idx].answer === opt ? 'bg-indigo-500/5 border-indigo-500/20 ring-1 ring-indigo-500/10' : ''"
                  >
                    <input 
                      type="radio" 
                      :name="'q' + q.ID" 
                      :value="opt" 
                      v-model="answers[idx].answer"
                      class="w-4 h-4 border-2 border-slate-300 dark:border-white/10 bg-transparent text-indigo-600 focus:ring-indigo-500"
                    />
                    <span class="font-bold text-slate-700 dark:text-slate-300 text-[11px] italic">{{ opt }}</span>
                  </label>
                </div>

                <!-- Checkbox -->
                <div v-if="q.type === 'checkbox'" class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                  <label 
                    v-for="opt in parseOptions(q.options)" 
                    :key="opt"
                    class="flex items-center gap-3 p-3 rounded-xl border border-slate-100 dark:border-white/5 hover:bg-slate-50 dark:hover:bg-white/[0.02] cursor-pointer transition-all active:scale-[0.98]"
                    :class="(answers[idx].answer as string[]).includes(opt) ? 'bg-indigo-500/5 border-indigo-500/20 ring-1 ring-indigo-500/10' : ''"
                  >
                    <input 
                      type="checkbox" 
                      :value="opt" 
                      v-model="answers[idx].answer"
                      class="w-4 h-4 rounded-md border-2 border-slate-300 dark:border-white/10 bg-transparent text-indigo-600 focus:ring-indigo-500"
                    />
                    <span class="font-bold text-slate-700 dark:text-slate-300 text-[11px] italic">{{ opt }}</span>
                  </label>
                </div>

                <!-- Text -->
                <div v-if="q.type === 'text'">
                  <input 
                    v-model="answers[idx].answer"
                    type="text"
                    placeholder="Enter answer..."
                    class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/5 rounded-xl px-4 py-2.5 text-[11px] font-bold text-slate-900 dark:text-white focus:outline-none focus:border-indigo-500/30 transition-all placeholder:text-slate-400 dark:placeholder:text-slate-700"
                  />
                </div>

                <!-- Textarea -->
                <div v-if="q.type === 'textarea'">
                  <textarea 
                    v-model="answers[idx].answer"
                    rows="3"
                    placeholder="Enter detailed feedback..."
                    class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/5 rounded-xl px-4 py-3 text-[11px] font-bold text-slate-900 dark:text-white focus:outline-none focus:border-indigo-500/30 transition-all resize-none placeholder:text-slate-400 dark:placeholder:text-slate-700"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Submit Button -->
        <div class="flex items-center justify-center pt-6">
          <button 
            @click="handleSubmit"
            :disabled="submitting"
            class="px-8 py-3 bg-gradient-to-r from-indigo-600 to-blue-600 hover:from-indigo-500 hover:to-blue-500 text-white rounded-xl font-black uppercase tracking-[0.2em] text-[10px] transition-all shadow-xl shadow-indigo-500/20 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 active:scale-95 group"
          >
            <Send v-if="!submitting" class="w-3.5 h-3.5 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 transition-transform" />
            <div v-else class="w-3.5 h-3.5 border-2 border-white/20 border-t-white rounded-full animate-spin"></div>
            {{ submitting ? 'Submitting...' : 'Submit Report / 提交' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
