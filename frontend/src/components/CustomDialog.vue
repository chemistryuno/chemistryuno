<script setup lang="ts">
import { ref, watch } from 'vue'
import { useDialog } from '../utils/dialog'
import { buttonClasses, inputClasses, modalClasses } from '@lib'
import { AlertCircle, HelpCircle, MessageSquare } from 'lucide-vue-next'

const { state, handleConfirm, handleCancel } = useDialog()

const countdown = ref(0)
const timer = ref<ReturnType<typeof setInterval> | null>(null)
const isComposing = ref(false)
const dialogModalClasses = modalClasses({
  width: 'md',
  overlayClassName: 'viewport-dialog-overlay',
  panelClassName: 'p-8 relative pointer-events-auto',
})
const promptInputClass = inputClasses({
  tone: 'special',
  size: 'large',
  radius: 'xl',
  className: 'px-5 font-medium focus:ring-4 focus:ring-purple-500/5',
})
const cancelButtonClass = buttonClasses({
  tone: 'secondary',
  variant: 'soft',
  size: 'xl',
  radius: 'xl',
  block: true,
  className: 'border border-slate-200 dark:border-white/5 shadow-none hover:text-slate-900 dark:hover:text-white',
})
const confirmButtonClass = () => buttonClasses({
  tone: state.type === 'alert' ? 'primary' : state.type === 'confirm' ? 'warning' : 'special',
  size: 'xl',
  radius: 'xl',
  block: true,
})

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

  // 当弹窗显示时禁用背景滚动，关闭时启用
  if (newVal) {
    document.documentElement.style.overflow = 'hidden'
    document.body.style.overflow = 'hidden'
    document.documentElement.style.scrollBehavior = 'auto'
  } else {
    document.documentElement.style.overflow = ''
    document.body.style.overflow = ''
    document.documentElement.style.scrollBehavior = ''
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
  <Teleport to="body">
    <Transition name="fade">
      <div v-if="state.show" :class="dialogModalClasses.overlay">
      <div :class="dialogModalClasses.panel">
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
              :class="promptInputClass"
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
            :class="cancelButtonClass"
          >
            {{ state.cancelText }}
          </button>
          <button 
            @click="handleConfirm"
            :disabled="countdown > 0"
            :class="confirmButtonClass()"
          >
            {{ countdown > 0 ? `${state.confirmText} (${countdown}s)` : state.confirmText }}
          </button>
        </div>
      </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style src="./CustomDialog.css" scoped></style>
