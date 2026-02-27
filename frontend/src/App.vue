<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import websocket from './utils/websocket'
import CustomDialog from './components/CustomDialog.vue'
import AnnouncementTicker from './components/AnnouncementTicker.vue'
import DuelInviteModal from './components/DuelInviteModal.vue'
import feedback from './utils/feedback'
import { useDialog } from './utils/dialog'
import { loadPluginScripts, dispatchPluginMessage } from './utils/plugin-runtime'

// --- 服务器重启横幅状态 ---
const restartBanner = ref<{ visible: boolean; countdown: number; reason: string }>({
  visible: false,
  countdown: 0,
  reason: ''
})
let restartInterval: ReturnType<typeof setInterval> | null = null

const loading = ref(true)
const { showAlert } = useDialog()
const router = useRouter()

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
    router.push(`/room/${msg.room_id}`)
  }
}

const handleDuelInvite = (msg: any) => {
  activeDuelInvite.value = {
    challenger_name: msg.data.challenger_nickname || msg.data.challenger_name,
    challenger_uid: msg.data.challenger_uid
  }
}

const handleDuelDeclined = (msg: any) => {
  const name = msg.data.nickname || msg.data.username || '研究员'
  showAlert(`研究员 ${name} 拒绝了你的挑战邀请。`, '挑战被拒绝')
}

const handleForceLogout = async (msg: any) => {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  websocket.disconnect()
  const reason = msg?.message || msg?.data || '您已被管理员强制下线'
  await showAlert(reason, '账号操作通知')
  router.push('/login')
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

// 管理员广播：区分 global / room / user 三种来源
const handleAdminBroadcast = (msg: any) => {
  const d = msg.data
  if (!d) return
  const titleMap: Record<string, string> = {
    info:    d.title || '📢 管理员通知',
    warning: d.title || '⚠️ 管理员警告',
    success: d.title || '✅ 系统提示',
    error:   d.title || '🚨 紧急通知',
  }
  const title = titleMap[d.msg_type] || d.title || '管理员广播'
  showAlert(d.content, title)
}

// 服务器重启 WebSocket 事件
const handleServerRestart = (msg: any) => {
  const d = msg.data || {}
  const seconds = typeof d === 'object' ? (d.seconds ?? 30) : 30
  const reason = typeof d === 'object' ? (d.reason ?? '') : ''
  restartBanner.value = { visible: true, countdown: seconds, reason }
  if (restartInterval) clearInterval(restartInterval)
  restartInterval = setInterval(() => {
    restartBanner.value.countdown--
    if (restartBanner.value.countdown <= 0) {
      clearInterval(restartInterval!)
      restartInterval = null
    }
  }, 1000)
}

const handleServerRestartNow = () => {
  restartBanner.value = { visible: true, countdown: 0, reason: '服务器正在重启，请稍候…' }
  if (restartInterval) { clearInterval(restartInterval); restartInterval = null }
}

const handleServerRestartCancelled = () => {
  if (restartInterval) { clearInterval(restartInterval); restartInterval = null }
  restartBanner.value = { visible: false, countdown: 0, reason: '' }
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

const handleGlobalClick = (e: MouseEvent) => {
  const target = e.target as HTMLElement
  // 检查点击目标或其祖先是否是可点击元素
  const clickable = target.closest('button, a, [role="button"], input[type="button"], input[type="submit"], select, summary, .clickable, .touch-feedback, .cursor-pointer')
  if (clickable) {
    // 只有当没有被明确标记为禁用音效时才播放
    if (!clickable.hasAttribute('data-no-click-sound')) {
      feedback.click()
    }
  }
}

const handlePluginMessage = (msg: any) => {
  dispatchPluginMessage(msg)
}

const handleAuthChanged = () => {
  const token = localStorage.getItem('token')
  if (token) {
    loadPluginScripts()
  }
}

onMounted(() => {
  try {
    // 监听主题变化
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    mediaQuery.addEventListener('change', updateTheme)
    window.addEventListener('storage', (e) => {
      if (e.key === 'theme') updateTheme()
    })
    window.addEventListener('theme-changed', updateTheme)
    window.addEventListener('auth-changed', handleAuthChanged)
    const token = localStorage.getItem('token')
    const userData = localStorage.getItem('user')
    
    if (token && userData) {
      // 用户已登录，建立 WebSocket 连接
      websocket.connect()
      loadPluginScripts()
    }

    websocket.on('feedback_update', handleFeedbackUpdate)
    websocket.on('duel_start', handleDuelStart)
    websocket.on('duel_invite', handleDuelInvite)
    websocket.on('duel_declined', handleDuelDeclined)
    websocket.on('system_announcement', handleSystemAnnouncement)
    websocket.on('force_logout', handleForceLogout)
    websocket.on('admin_broadcast', handleAdminBroadcast)
    websocket.on('server_restart', handleServerRestart)
    websocket.on('server_restart_now', handleServerRestartNow)
    websocket.on('server_restart_cancelled', handleServerRestartCancelled)
    websocket.on('plugin_message', handlePluginMessage)

    // 全局点击音效监听
    window.addEventListener('click', handleGlobalClick, true)
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
  websocket.off('force_logout', handleForceLogout)
  websocket.off('admin_broadcast', handleAdminBroadcast)
  websocket.off('server_restart', handleServerRestart)
  websocket.off('server_restart_now', handleServerRestartNow)
  websocket.off('server_restart_cancelled', handleServerRestartCancelled)
  websocket.off('plugin_message', handlePluginMessage)
  if (restartInterval) clearInterval(restartInterval)

  // 移除全局监听
  window.removeEventListener('click', handleGlobalClick, true)
  window.removeEventListener('auth-changed', handleAuthChanged)
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
      <!-- 服务器重启横幅 -->
      <Transition name="slide-down">
        <div
          v-if="restartBanner.visible"
          class="fixed top-0 left-0 right-0 z-[9999] bg-orange-500 text-white text-sm font-bold px-4 py-2 flex items-center justify-center gap-3 shadow-lg"
        >
          <span>⚠️</span>
          <span v-if="restartBanner.countdown > 0">
            服务器将在 {{ restartBanner.countdown }} 秒后重启{{ restartBanner.reason ? `：${restartBanner.reason}` : '' }}，请尽快保存进度
          </span>
          <span v-else>{{ restartBanner.reason || '服务器正在重启，请稍候…' }}</span>
        </div>
      </Transition>
      <router-view></router-view>
      <CustomDialog />
      <DuelInviteModal v-if="activeDuelInvite" :invite="activeDuelInvite" @close="activeDuelInvite = null" />
    </div>
  </template>
</template>
