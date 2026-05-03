<script setup lang="ts">
import { ref, watch } from 'vue'
import { useDialog } from '../utils/dialog'
import { AlertCircle, HelpCircle, MessageSquare } from 'lucide-vue-next'

const { state, handleConfirm, handleCancel } = useDialog()

const countdown = ref(0)
const timer = ref<ReturnType<typeof setInterval> | null>(null)
const isComposing = ref(false)

watch(() => state.show, (newVal) => {
  if (newVal && state.closeDelay > 0) {
    countdown.value = state.closeDelay
    if (timer.value) clearInterval(timer.value)
    timer.value = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) {
        if (timer.value) clearInterval(timer.value)
        timer.value = null
      }
    }, 1000)
  } else {
    if (timer.value) clearInterval(timer.value)
    timer.value = null
    countdown.value = 0
  }
})

const handleInputKeyDown = (e: KeyboardEvent) => {
  // 只有在非 composition 状态且计时器完成后，按 Enter 才确认
  if (e.key === 'Enter' && !isComposing.value && countdown.value <= 0) {
    handleConfirm()
  }
}

const handleCompositionStart = () => {
  isComposing.value = true
}

const handleCompositionEnd = () => {
  isComposing.value = false
}
</script>

<template>
  <Transition name="fade">
    <div v-if="state.show" class="viewport-modal-overlay bg-slate-900/60 dark:bg-black/80 backdrop-blur-md z-[9999] p-4">
      <div class="viewport-modal-panel bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2rem] p-8 max-w-md w-full shadow-2xl relative pointer-events-auto">
        <!-- 装饰背景 -->
        <div class="absolute -top-24 -right-24 w-48 h-48 bg-blue-500/5 rounded-full blur-3xl pointer-events-none" />
        <div class="absolute -bottom-24 -left-24 w-48 h-48 bg-purple-500/5 rounded-full blur-3xl pointer-events-none" />
        
        <div class="flex items-center gap-4 mb-6 relative z-10">
          <div v-if="state.type === 'alert'" class="w-12 h-12 rounded-2xl bg-blue-500/10 flex items-center justify-center border border-blue-500/20">
            <AlertCircle class="w-6 h-6 text-blue-600 dark:text-blue-400" />
          </div>
          <div v-else-if="state.type === 'confirm'" class="w-12 h-12 rounded-2xl bg-orange-500/10 flex items-center justify-center border border-orange-500/20">
            <HelpCircle class="w-6 h-6 text-orange-600 dark:text-orange-400" />
          </div>
          <div v-else-if="state.type === 'prompt'" class="w-12 h-12 rounded-2xl bg-purple-500/10 flex items-center justify-center border border-purple-500/20">
            <MessageSquare class="w-6 h-6 text-purple-600 dark:text-purple-400" />
          </div>
          <h3 class="text-xl font-black text-slate-900 dark:text-white italic tracking-tight">{{ state.title }}</h3>
        </div>
        
        <div class="mb-8 relative z-10">
          <p class="text-slate-500 dark:text-slate-400 leading-relaxed whitespace-pre-line font-medium">{{ state.message }}</p>
          
          <!-- 输入框（仅prompt类型显示） -->
          <div v-if="state.type === 'prompt'" class="mt-6 relative">
            <input 
              v-model="state.inputValue"
              type="text" 
              :placeholder="state.inputPlaceholder"
              class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl px-5 py-4 text-slate-900 dark:text-white focus:outline-none focus:border-purple-500/50 focus:ring-4 focus:ring-purple-500/5 transition-all placeholder:text-slate-400 dark:placeholder:text-slate-600 font-medium"
              @keydown="handleInputKeyDown"
              @compositionstart="handleCompositionStart"
              @compositionend="handleCompositionEnd"
              autofocus
            />
          </div>
        </div>
        
        <div class="flex gap-4 relative z-10">
          <button 
            v-if="state.type === 'confirm' || state.type === 'prompt'"
            @click="handleCancel"
            class="flex-1 px-6 py-4 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white rounded-2xl font-black text-xs uppercase tracking-widest transition-all border border-slate-200 dark:border-white/5"
          >
            {{ state.cancelText }}
          </button>
          <button 
            @click="handleConfirm"
            :disabled="countdown > 0"
            :class="[
              'flex-1 px-6 py-4 rounded-2xl font-black text-xs uppercase tracking-widest transition-all text-white shadow-xl active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed disabled:grayscale',
              state.type === 'alert' ? 'bg-blue-600 hover:bg-blue-500 shadow-blue-500/20 dark:shadow-blue-900/20' :
              state.type === 'confirm' ? 'bg-orange-600 hover:bg-orange-500 shadow-orange-500/20 dark:shadow-orange-900/20' :
              'bg-purple-600 hover:bg-purple-500 shadow-purple-500/20 dark:shadow-purple-900/20'
            ]"
          >
            {{ countdown > 0 ? `${state.confirmText} (${countdown}s)` : state.confirmText }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style src="./CustomDialog.css" scoped></style>
