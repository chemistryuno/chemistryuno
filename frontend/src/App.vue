<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import websocket from './utils/websocket'
import { gameAPI } from './utils/api'
import CustomDialog from './components/CustomDialog.vue'
import AnnouncementTicker from './components/AnnouncementTicker.vue'
import DuelInviteModal from './components/DuelInviteModal.vue'
import { useDialog } from './utils/dialog'

const loading = ref(true)
const { showAlert, showConfirm, closeDialog } = useDialog()
const route = useRoute()

const activeDuelInvite = ref<any>(null)

const handleFeedbackUpdate = (msg: any) => {
  if (msg && msg.status) {
    const statusLabel = msg.status === 'accepted' ? '已受理' : '不予受理'
    showAlert(`您的反馈有新进展：状态更新为 [${statusLabel}]。\n回复：${msg.resolution_note || '无'}`, '反馈通知')
  }
}

const handleDuelStart = (msg: any) => {
  if (msg.room_id) {
    showAlert('量子隧道已建立，正在进入单挑战场...', '单挑协议启动')
    window.location.href = `/room/${msg.room_id}`
  }
}

const handleDuelInvite = (msg: any) => {
  activeDuelInvite.value = {
    challenger_name: msg.data.challenger_name,
    challenger_uid: msg.data.challenger_uid
  }
}

const handleDuelDeclined = (msg: any) => {
  showAlert(`研究员 ${msg.data.username} 拒绝了你的挑战邀请。`, '挑战被拒绝')
}

const handleSystemAnnouncement = (msg: any) => {
  const ann = msg.data
  // 如果不是跑马灯，则视为弹窗公告
  if (ann && !ann.is_ticker) {
    let title = ann.title || '系统公告'
    if (ann.type === 'emergency' && !ann.title) title = '紧急通知'
    if (ann.type === 'maintenance' && !ann.title) title = '维护通知'
    showAlert(ann.content, title, '确定', ann.close_delay || 0)
  }
}

const updateTheme = () => {
  const storedTheme = localStorage.getItem('theme') || 'system'
  const root = document.documentElement
  root.classList.remove('light', 'dark')

  if (storedTheme === 'system') {
    if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
      root.classList.add('dark')
    } else {
      root.classList.add('light')
    }
  } else {
    root.classList.add(storedTheme)
  }
}

// 立即运行一次以防止 FOUC
updateTheme()

onMounted(() => {
  try {
    // 监听主题变化
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    mediaQuery.addEventListener('change', updateTheme)
    window.addEventListener('storage', (e) => {
      if (e.key === 'theme') updateTheme()
    })
    window.addEventListener('theme-changed', updateTheme)
    const token = localStorage.getItem('token')
    const userData = localStorage.getItem('user')
    
    if (token && userData) {
      // 用户已登录，建立 WebSocket 连接
      websocket.connect()
    }

    websocket.on('feedback_update', handleFeedbackUpdate)
    websocket.on('duel_start', handleDuelStart)
    websocket.on('duel_invite', handleDuelInvite)
    websocket.on('duel_declined', handleDuelDeclined)
    websocket.on('system_announcement', handleSystemAnnouncement)
  } catch (err) {
    console.error('App initialization failed:', err)
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  const mq = window.matchMedia('(prefers-color-scheme: dark)')
  mq.removeEventListener('change', updateTheme)
  websocket.off('feedback_update', handleFeedbackUpdate)
  websocket.off('duel_start', handleDuelStart)
  websocket.off('duel_invite', handleDuelInvite)
  websocket.off('duel_declined', handleDuelDeclined)
  websocket.off('system_announcement', handleSystemAnnouncement)
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
      <AnnouncementTicker />
      <router-view></router-view>
      <CustomDialog />
      <DuelInviteModal v-if="activeDuelInvite" :invite="activeDuelInvite" @close="activeDuelInvite = null" />
    </div>
  </template>
</template>
