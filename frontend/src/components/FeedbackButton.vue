<script setup lang="ts">
import { ref, watch } from 'vue'
import { Send, X, ChevronDown } from 'lucide-vue-next'
import { buttonClasses, iconButtonClasses, inputClasses, modalClasses } from '@lib'
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
const feedbackModalClasses = modalClasses({
  width: 'sm',
  zIndex: 'z-[999]',
  panelRadius: '3xl',
  panelClassName: 'relative custom-scrollbar dark:bg-slate-900 p-6 animate-in zoom-in-95 fade-in duration-200',
})
const closeButtonClass = iconButtonClasses({
  size: 'sm',
  radius: 'full',
  className: 'text-slate-400 hover:text-slate-600 dark:hover:text-white',
})
const fieldClass = inputClasses({
  tone: 'primary',
  radius: 'lg',
  className: 'bg-slate-100 dark:bg-white/5',
})
const selectClass = inputClasses({
  tone: 'primary',
  radius: 'lg',
  className: 'appearance-none bg-slate-100 dark:bg-white/5 py-2.5 text-sm cursor-pointer hover:bg-slate-200 dark:hover:bg-white/10',
})
const submitButtonClass = buttonClasses({
  tone: 'primary',
  size: 'lg',
  block: true,
  loading: isSubmitting.value,
})

// 当反馈类型改变时，如果是方程式纠错，可能需要特殊初始化
watch(type, (newType) => {
  if (newType !== 'equation') {
    equationContent.value = ''
  }
})

// 监控弹窗状态以禁用/启用背景滚动
watch(
  isOpen,
  (open) => {
    if (open) {
      document.documentElement.style.overflow = 'hidden'
      document.body.style.overflow = 'hidden'
    } else {
      document.documentElement.style.overflow = ''
      document.body.style.overflow = ''
    }
  }
)

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
  <Teleport to="body">
    <div v-if="isOpen" :class="feedbackModalClasses.overlay" @click.self="isOpen = false">
    <!-- Feedback Panel -->
    <div 
      :class="feedbackModalClasses.panel"
    >
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-lg font-black italic tracking-tighter uppercase">Send Feedback</h3>
        <button @click="isOpen = false" :class="closeButtonClass">
          <X class="w-4 h-4" />
        </button>
      </div>
      
      <div class="space-y-4">
        <div>
          <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-1.5 block">Feedback Type</label>
          <div class="relative group">
            <select 
              v-model="type"
              :class="selectClass"
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
            :class="[fieldClass, 'resize-none h-32']"
           ></textarea>
        </div>

        <button 
          @click="submitFeedback"
          :disabled="isSubmitting || (!content.trim() && !equationContent.trim())"
          :class="submitButtonClass"
        >
          <Send class="w-4 h-4" />
          {{ isSubmitting ? 'Sending...' : 'Transmit Feedback' }}
        </button>
      </div>
    </div>
    </div>
  </Teleport>
</template>

<style src="./FeedbackButton.css" scoped></style>

