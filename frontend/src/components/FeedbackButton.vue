<script setup lang="ts">
import { ref } from 'vue'
import { MessageCircle, Send, X } from 'lucide-vue-next'
import { authAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'

const isOpen = ref(false)
const content = ref('')
const type = ref('general')
const isSubmitting = ref(false)
const { showAlert } = useDialog()

const submitFeedback = async () => {
  if (!content.value.trim()) return
  
  isSubmitting.value = true
  try {
    await authAPI.submitFeedback(content.value, type.value)
    showAlert('您的反馈已提交，我们会尽快处理！', '提交成功')
    content.value = ''
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
      class="absolute bottom-20 right-0 w-80 bg-white dark:bg-slate-900 border border-slate-200 dark:border-white/10 rounded-3xl shadow-2xl p-6 animate-in slide-in-from-bottom-10 fade-in duration-300"
    >
      <h3 class="text-lg font-black italic tracking-tighter mb-4 uppercase">Send Feedback</h3>
      
      <div class="space-y-4">
        <div>
          <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-1.5 block">Feedback Type</label>
          <select 
            v-model="type"
            class="w-full bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2 text-sm outline-none focus:border-blue-500/50"
          >
            <option value="general">一般建议</option>
            <option value="bug">反馈Bug</option>
            <option value="feature">功能请求</option>
            <option value="equation">方程式纠错</option>
          </select>
        </div>

        <div>
           <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-1.5 block">Description</label>
           <textarea 
            v-model="content"
            rows="4"
            placeholder="Tell us what you think..."
            class="w-full bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-3 text-sm outline-none focus:border-blue-500/50 resize-none h-32"
           ></textarea>
        </div>

        <button 
          @click="submitFeedback"
          :disabled="isSubmitting || !content.trim()"
          class="w-full bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white py-3 rounded-xl font-black uppercase tracking-widest text-xs transition-all flex items-center justify-center gap-2"
        >
          <Send class="w-4 h-4" />
          {{ isSubmitting ? 'Sending...' : 'Transmit Feedback' }}
        </button>
      </div>
    </div>
  </div>
</template>
