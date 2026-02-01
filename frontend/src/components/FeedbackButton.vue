<script setup lang="ts">
import { ref, watch } from 'vue'
import { MessageCircle, Send, X, ChevronDown } from 'lucide-vue-next'
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
  } catch (error: any) {
    showAlert(error.response?.data?.error || '提交失败', '错误')
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div class="fixed bottom-6 right-6 z-[999]">
    <!-- Floating Button -->
    <button 
      @click="isOpen = !isOpen"
      class="w-14 h-14 bg-blue-600 hover:bg-blue-500 text-white rounded-full shadow-2xl shadow-blue-500/40 flex items-center justify-center transition-all active:scale-90 hover:rotate-12"
    >
      <MessageCircle v-if="!isOpen" class="w-6 h-6" />
      <X v-else class="w-6 h-6" />
    </button>

    <!-- Feedback Panel -->
    <div 
      v-if="isOpen"
      class="absolute bottom-20 right-0 w-80 max-h-[80vh] overflow-y-auto custom-scrollbar bg-white dark:bg-slate-900 border border-slate-200 dark:border-white/10 rounded-3xl shadow-2xl p-6 animate-in slide-in-from-bottom-10 fade-in duration-300"
    >
      <h3 class="text-lg font-black italic tracking-tighter mb-4 uppercase">Send Feedback</h3>
      
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

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 2px;
}
</style>

