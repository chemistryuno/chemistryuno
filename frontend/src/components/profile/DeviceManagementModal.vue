<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../../utils/api'
import { useDialog } from '../../utils/dialog'
import {
  X,
  Monitor,
  Smartphone,
  Globe,
  LogOut,
  Clock,
  ShieldAlert,
  Snowflake,
  AlertTriangle,
  Loader2
} from 'lucide-vue-next'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const router = useRouter()
const { showAlert, showConfirm, showPrompt } = useDialog()
const sessions = ref<any[]>([])
const loading = ref(false)
const freezeLoading = ref(false)

const fetchSessions = async () => {
  loading.value = true
  try {
    const res = await authAPI.getSessions()
    sessions.value = res.data
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (props.show) fetchSessions()
})

watch(() => props.show, (newVal) => {
  if (newVal) fetchSessions()
})

const handleLogoutSession = async (session: any) => {
  const confirmed = await showConfirm(
    session.is_current 
      ? '确定要登出当前设备吗？此操作将立即中断此连接。' 
      : '确定要撤销该设备的访问权限吗？如果该设备正在线，其后续操作将需要重新登录。',
    '设备访问撤销'
  )
  if (!confirmed) return

  try {
    await authAPI.logoutSession(session.id)
    if (session.is_current) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      router.push('/login')
    } else {
      await fetchSessions()
      showAlert('已成功撤销该设备的访问权限', '操作完成')
    }
  } catch (err: any) {
    showAlert(err.response?.data?.error || '操作失败', '错误')
  }
}

const handleFreezeAccount = async () => {
  const hoursStr = await showPrompt(
    '请输入冻结时长（1-24小时）。冻结期间您将无法登录。冻结成功后将立即强制登出所有当前活跃设备。',
    '24',
    '账号自助冻结'
  )
  
  if (!hoursStr) return
  const hours = parseInt(hoursStr)
  if (isNaN(hours) || hours < 1 || hours > 24) {
    showAlert('冻结时长必须在 1 到 24 之间', '参数无效')
    return
  }

  const confirmFreeze = await showConfirm(
    `确定要冻结账号 ${hours} 小时吗？在 ${new Date(Date.now() + hours * 3600000).toLocaleString()} 之前，您将无法访问本实验室。`,
    '⚠️ 最终确认'
  )
  if (!confirmFreeze) return

  freezeLoading.value = true
  try {
    await authAPI.freezeAccount(hours)
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    router.push('/login')
  } catch (err: any) {
    showAlert(err.response?.data?.error || '冻结失败', '错误')
  } finally {
    freezeLoading.value = false
  }
}

const getIcon = (ua: string) => {
  const lowerUA = ua.toLowerCase()
  if (lowerUA.includes('windows') || lowerUA.includes('macintosh') || lowerUA.includes('linux')) return Monitor
  if (lowerUA.includes('android') || lowerUA.includes('iphone')) return Smartphone
  return Globe
}

const formatUA = (ua: string) => {
  if (!ua) return '未知终端'
  const lowerUA = ua.toLowerCase()
  if (lowerUA.includes('windows')) return 'Windows PC'
  if (lowerUA.includes('macintosh') || lowerUA.includes('mac os')) return 'MacBook / iMac'
  if (lowerUA.includes('iphone')) return 'Apple iPhone'
  if (lowerUA.includes('android')) return 'Android 设备'
  if (lowerUA.includes('linux')) return 'Linux 终端'
  
  // 提取简单的浏览器信息
  if (lowerUA.includes('chrome')) return 'Google Chrome'
  if (lowerUA.includes('firefox')) return 'Firefox'
  if (lowerUA.includes('safari') && !lowerUA.includes('chrome')) return 'Safari'
  
  return ua.split(')')[0].split('(')[1] || 'Web 浏览器'
}
</script>

<template>
  <div v-if="show" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
    <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="emit('close')" />
    
    <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 w-full max-w-2xl rounded-[2.5rem] shadow-2xl relative overflow-hidden animate-in zoom-in-95 duration-200">
      <!-- Header -->
      <div class="p-6 border-b border-slate-100 dark:border-white/5 flex items-center justify-between">
        <div class="flex items-center gap-4">
          <div class="p-2.5 bg-amber-500/10 rounded-xl">
            <Smartphone class="w-5 h-5 text-amber-500" />
          </div>
          <div>
            <h3 class="text-lg font-black uppercase italic tracking-tighter">终端设备管理</h3>
            <p class="text-[9px] text-slate-500 font-mono">ACTIVE_SESSIONS_CONTROL_PROTOCOL</p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button @click="fetchSessions" :disabled="loading" class="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl transition-colors text-slate-400">
             <Loader2 :class="['w-4 h-4', { 'animate-spin': loading }]" />
          </button>
          <button @click="emit('close')" class="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl transition-colors text-slate-400">
            <X class="w-5 h-5" />
          </button>
        </div>
      </div>

      <div class="p-6 max-h-[70vh] overflow-y-auto space-y-6 custom-scrollbar">
        <!-- Freeze Account Section -->
        <div class="bg-blue-500/5 border border-blue-500/20 rounded-2xl p-5">
          <div class="flex items-start gap-4 mb-4">
            <div class="p-2 bg-blue-500/20 rounded-lg">
              <Snowflake class="w-4 h-4 text-blue-500" />
            </div>
            <div class="flex-1">
              <h4 class="text-xs font-bold text-slate-900 dark:text-white uppercase tracking-wider">紧急冷冻协议 / Account Freeze</h4>
              <p class="text-[10px] text-slate-500 mt-1">怀疑账户异常活动？可暂时冻结账号（限24小时内）。冻结期间无法登录，且所有活跃会话将立即撤销。</p>
            </div>
          </div>
          <button 
            @click="handleFreezeAccount"
            :disabled="freezeLoading"
            class="w-full flex items-center justify-center gap-2 py-2.5 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white rounded-lg font-bold text-xs transition-all"
          >
            <Loader2 v-if="freezeLoading" class="w-3 h-3 animate-spin" />
            <Snowflake v-else class="w-3 h-3" />
            激活冷冻序列 (1-24h)
          </button>
        </div>

        <!-- Session List -->
        <div class="space-y-3">
          <h4 class="text-[10px] font-black text-slate-400 uppercase tracking-[0.2em] px-2 mb-1">活跃终端列表</h4>
          
          <div v-if="loading" class="py-10 flex flex-col items-center justify-center gap-4">
             <Loader2 class="w-6 h-6 text-blue-500 animate-spin" />
             <span class="text-[10px] font-mono text-slate-500">正在检索活跃节点...</span>
          </div>

          <div v-else-if="sessions.length === 0" class="py-10 text-center bg-slate-50 dark:bg-white/5 rounded-2xl border border-dashed border-slate-200 dark:border-white/5">
             <Globe class="w-10 h-10 text-slate-300 dark:text-slate-700 mx-auto mb-3 opacity-20" />
             <p class="text-xs text-slate-500">未发现活跃会话</p>
             <p class="text-[9px] text-slate-400 mt-2 px-6">提示：如果您刚更新系统，可能需要重新登录以同步当前设备的会话状态。</p>
          </div>

          <div 
            v-for="session in sessions" 
            :key="session.id"
            class="group p-4 bg-white dark:bg-white/5 border border-slate-200 dark:border-white/5 rounded-2xl hover:border-blue-500/30 transition-all flex items-center justify-between"
          >
            <div class="flex items-center gap-4">
              <div :class="[
                'p-3 rounded-xl transition-colors',
                session.is_current ? 'bg-blue-500/10 text-blue-500' : 'bg-slate-100 dark:bg-white/5 text-slate-400'
              ]">
                <component :is="getIcon(session.user_agent)" class="w-5 h-5" />
              </div>
              <div>
                <div class="flex items-center gap-2">
                  <span class="text-sm font-bold text-slate-900 dark:text-white">{{ formatUA(session.user_agent) }}</span>
                  <span v-if="session.is_current" class="text-[8px] font-black uppercase px-1.5 py-0.5 bg-blue-500/20 text-blue-500 rounded-md tracking-widest">CURRENT</span>
                </div>
                <div class="flex items-center gap-3 mt-0.5">
                  <span class="flex items-center gap-1 text-[9px] text-slate-500 font-mono">
                    <Globe class="w-2.5 h-2.5" /> {{ session.ip }}
                  </span>
                  <span class="flex items-center gap-1 text-[9px] text-slate-500 font-mono">
                    <Clock class="w-2.5 h-2.5" /> {{ new Date(session.last_active).toLocaleString() }}
                  </span>
                </div>
              </div>
            </div>

            <button 
              @click="handleLogoutSession(session)"
              :class="[
                'p-2.5 rounded-lg transition-all',
                session.is_current ? 'text-red-500 hover:bg-red-500/10' : 'text-slate-400 hover:text-red-500 hover:bg-red-500/10'
              ]"
              title="撤回访问权限"
            >
              <LogOut class="w-4 h-4" />
            </button>
          </div>

          <!-- Current Session Not Found Notice -->
          <div v-if="sessions.length > 0 && !sessions.some(s => s.is_current)" class="p-4 bg-blue-500/5 border border-blue-500/10 rounded-2xl mt-4">
             <div class="flex gap-3">
                <ShieldAlert class="w-4 h-4 text-blue-500 shrink-0" />
                <p class="text-[9px] text-blue-600/80 leading-relaxed font-black uppercase tracking-widest">
                  SESSION_NOT_BOUND // 当前终端未绑定会话 ID。为了获得最佳安全性与控制体验，建议您重新登录此设备。
                </p>
             </div>
          </div>
        </div>

        <!-- Warning Info -->
        <div class="flex gap-3 p-4 bg-amber-500/5 border border-amber-500/20 rounded-2xl">
          <AlertTriangle class="w-5 h-5 text-amber-500 shrink-0" />
          <p class="text-[10px] text-amber-600/80 leading-relaxed font-medium">
            提示：撤回权限将立即废弃对应的访问令牌。如果该设备属于公共终端且您忘记登出，请立即通过此界面撤销访问权限以保护您的研究数据。
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style src="./DeviceManagementModal.css" scoped></style>
