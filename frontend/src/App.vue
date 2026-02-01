<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import websocket from './utils/websocket'
import CustomDialog from './components/CustomDialog.vue'
import FeedbackButton from './components/FeedbackButton.vue'
import { useDialog } from './utils/dialog'

const loading = ref(true)
const { showAlert } = useDialog()
const route = useRoute()
const feedbackBtnRef = ref<any>(null)

watch(() => route.query.report, (val) => {
  if (val) {
    feedbackBtnRef.value?.prefill(String(val), 'equation')
  }
})

const updateTheme = () => {
  if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
    document.documentElement.classList.add('dark')
    document.documentElement.classList.remove('light')
  } else {
    document.documentElement.classList.add('light')
    document.documentElement.classList.remove('dark')
  }
}

onMounted(() => {
  // 监听系统主题变化
  updateTheme()
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  mediaQuery.addEventListener('change', updateTheme)

  // 检查本地存储的token
  const token = localStorage.getItem('token')
  const userData = localStorage.getItem('user')
  
  if (token && userData) {
    // 用户已登录，建立 WebSocket 连接
    websocket.connect()
  }

  // 监听全局反馈更新
  websocket.on('feedback_update', (msg: any) => {
    if (msg && msg.status) {
      const statusLabel = msg.status === 'accepted' ? '已受理' : '不予受理'
      showAlert(`您的反馈有新进展：状态更新为 [${statusLabel}]。\n回复：${msg.resolution_note || '无'}`, '反馈通知')
    }
  })

  loading.value = false
})

onUnmounted(() => {
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  mediaQuery.removeEventListener('change', updateTheme)
  websocket.off('feedback_update', () => {})
})
</script>

<template>
  <div v-if="loading" class="min-h-screen bg-slate-50 dark:bg-[#0a0a0c] flex flex-col items-center justify-center gap-4 transition-colors duration-300">
    <div class="relative w-16 h-16">
      <div class="absolute inset-0 rounded-full border-4 border-emerald-500/20"></div>
      <div class="absolute inset-0 rounded-full border-4 border-emerald-500 border-t-transparent animate-spin"></div>
    </div>
    <p class="text-emerald-600 dark:text-emerald-500/70 font-mono tracking-widest text-sm animate-pulse">
      INITIALIZING LABORATORY...
    </p>
  </div>
  <template v-else>
    <div class="transition-colors duration-300 min-h-screen bg-slate-50 dark:bg-[#0a0a0c] text-slate-900 dark:text-slate-200">
      <router-view></router-view>
      <CustomDialog />
      <FeedbackButton ref="feedbackBtnRef" />
    </div>
  </template>
</template>

<style>
/* Any global styles can go here if not in index.css */
</style>
