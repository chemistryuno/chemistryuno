<script setup lang="ts">
import { ref, watch } from 'vue'
import { Send, X, ChevronDown } from 'lucide-vue-next'
import { authAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import EquationEditor from './EquationEditor.vue'

const isOpen = ref(false)
const content = ref('')
const equationContent = ref('')
const type = ref('general')

const prefill = (newContent: string, newType: string = 'general') => {
  isOpen.value = true
  content.value = newContent
  type.value = newType
}

defineExpose({ prefill })
const emit = defineEmits<{ submitted: [] }>()
const isSubmitting = ref(false)
const { showAlert } = useDialog()
const editorRef = ref<any>(null)

// 当反馈类型改变时，如果是方程式纠错，可能需要特殊初始化
watch(type, (newType) => {
  if (newType !== 'equation') {
    equationContent.value = ''
  }
})

const submitFeedback = async () => {
  const finalContent = type.value === 'equation' 
    ? `【方程式纠错】${equationContent.value}\n\n详细说明：${content.value}`
    : content.value

  if (!content.value.trim() && !equationContent.value.trim()) return
  
  isSubmitting.value = true
  try {
    await authAPI.submitFeedback(finalContent, type.value)
    showAlert('您的反馈已提交，我们会尽快处理！', '提交成功')
    content.value = ''
    equationContent.value = ''
    type.value = 'general'
    editorRef.value?.reset()
    isOpen.value = false
    emit('submitted')
  } catch (error: any) {
    showAlert(error.response?.data?.error || '提交失败', '错误')
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div v-if="isOpen" class="viewport-modal-overlay z-[999] bg-slate-900/60 dark:bg-black/80 backdrop-blur-md p-4" @click.self="isOpen = false">
    <!-- Feedback Panel -->
    <div 
      class="viewport-modal-panel relative w-full max-w-sm custom-scrollbar bg-white dark:bg-slate-900 border border-slate-200 dark:border-white/10 rounded-3xl shadow-2xl p-6 animate-in zoom-in-95 fade-in duration-200"
    >
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-lg font-black italic tracking-tighter uppercase">Send Feedback</h3>
        <button @click="isOpen = false" class="p-1.5 hover:bg-slate-100 dark:hover:bg-white/10 rounded-full transition-colors text-slate-400 hover:text-slate-600 dark:hover:text-white">
          <X class="w-4 h-4" />
        </button>
      </div>
      
      <div class="space-y-4">
        <div>
          <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-1.5 block">Feedback Type</label>
          <div class="relative group">
            <select 
              v-model="type"
              class="appearance-none w-full bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2.5 text-sm outline-none focus:border-blue-500/50 transition-all cursor-pointer hover:bg-slate-200 dark:hover:bg-white/10"
            >
              <option value="general" class="dark:bg-slate-900">一般建议</option>
              <option value="bug" class="dark:bg-slate-900">反馈Bug</option>
              <option value="feature" class="dark:bg-slate-900">功能请求</option>
              <option value="equation" class="dark:bg-slate-900">方程式纠错</option>
            </select>
            <ChevronDown class="w-4 h-4 text-slate-400 absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none transition-colors group-hover:text-blue-500" />
          </div>
        </div>

        <!-- Equation Editor (Only for equation type) -->
        <div v-if="type === 'equation'" class="animate-in fade-in slide-in-from-top-2 duration-300">
          <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-1.5 block">Equation Builder</label>
          <EquationEditor ref="editorRef" v-model="equationContent" />
        </div>

        <div>
           <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-1.5 block">
             {{ type === 'equation' ? 'Correction Details' : 'Description' }}
           </label>
           <textarea 
            v-model="content"
            rows="4"
            :placeholder="type === 'equation' ? '请说明该方程式在哪里有误...' : 'Tell us what you think...'"
            class="w-full bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-3 text-sm outline-none focus:border-blue-500/50 resize-none h-32"
           ></textarea>
        </div>

        <button 
          @click="submitFeedback"
          :disabled="isSubmitting || (!content.trim() && !equationContent.trim())"
          class="w-full bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white py-3 rounded-xl font-black uppercase tracking-widest text-xs transition-all flex items-center justify-center gap-2"
        >
          <Send class="w-4 h-4" />
          {{ isSubmitting ? 'Sending...' : 'Transmit Feedback' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style src="./FeedbackButton.css" scoped></style>

