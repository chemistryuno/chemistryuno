<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { X, Smartphone, Monitor, Globe, Trash2, Calendar, Loader2, ShieldCheck, Clock } from 'lucide-vue-next'
import { authAPI } from '../../utils/api'
import { useDialog } from '../../utils/dialog'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const { showAlert, showConfirm } = useDialog()
const sessions = ref<any[]>([])
const loading = ref(false)

const getDeviceIcon = (ua: string) => {
  const lowercaseUA = ua.toLowerCase()
  if (lowercaseUA.includes('mobile') || lowercaseUA.includes('android') || lowercaseUA.includes('iphone')) {
    return Smartphone
  }
  return Monitor
}

const getBrowserName = (ua: string) => {
  if (ua.includes('Firefox')) return 'Firefox'
  if (ua.includes('Edg')) return 'Edge'
  if (ua.includes('Chrome')) return 'Chrome'
  if (ua.includes('Safari')) return 'Safari'
  return '未知浏览器'
}

const getOSName = (ua: string) => {
  if (ua.includes('Windows')) return 'Windows'
  if (ua.includes('Mac OS')) return 'macOS'
  if (ua.includes('Android')) return 'Android'
  if (ua.includes('iPhone') || ua.includes('iPad')) return 'iOS'
  if (ua.includes('Linux')) return 'Linux'
  return '未知系统'
}

const fetchSessions = async () => {
  loading.value = true
  try {
    const res = await authAPI.getSessions()
    sessions.value = res.data || []
  } catch (error) {
    console.error('获取登录设备失败', error)
  } finally {
    loading.value = false
  }
}

const revokeSession = async (id: string, isCurrent: boolean) => {
  if (isCurrent) {
    showAlert('您不能在设备管理中终止当前的会话。如需退出登录，请使用主页的退出功能。', '操作受限')
    return
  }

  const confirmed = await showConfirm(
    '确认要终止该设备的访问权限吗？该设备上的用户将被强制下线。',
    '身份令牌失效警告',
    '强制下线',
    '保持连接'
  )
  
  if (!confirmed) return
  
  try {
    await authAPI.revokeSession(id)
    await fetchSessions()
    showAlert('该设备的访问权限已被永久撤回。', '操作成功')
  } catch (error: any) {
    showAlert('撤回失败: ' + (error.response?.data?.error || '未知错误'), '系统异常')
  }
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleString('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const currentSessionID = ref('')
const parseCurrentSID = () => {
    const token = localStorage.getItem('token')
    if (token) {
        try {
            const payload = JSON.parse(atob(token.split('.')[1]))
            currentSessionID.value = payload.sid
        } catch (e) {}
    }
}

onMounted(() => {
  if (props.show) {
    parseCurrentSID()
    fetchSessions()
  }
})
</script>

<template>
  <Transition name="modal">
    <div v-if="show" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-md" @click="$emit('close')" />
      
      <div class="relative bg-white dark:bg-[#111114] w-full max-w-2xl rounded-[2.5rem] shadow-2xl overflow-hidden border border-slate-200 dark:border-white/10">
        <!-- 头部 -->
        <div class="p-8 border-b border-slate-100 dark:border-white/5 flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div class="bg-blue-500/10 p-3 rounded-2xl">
              <Monitor class="w-6 h-6 text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <h3 class="text-xl font-black uppercase italic tracking-tighter text-slate-900 dark:text-white">
                登录设备管理 / Login Devices
              </h3>
              <p class="text-slate-500 text-xs">监控并管理所有已授权访问实验室的终端设备</p>
            </div>
          </div>
          <button @click="$emit('close')" class="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl transition-colors">
            <X class="w-6 h-6 text-slate-400" />
          </button>
        </div>

        <!-- 内容 -->
        <div class="p-8 max-h-[60vh] overflow-y-auto">
          <div v-if="loading" class="py-20 flex flex-col items-center justify-center text-slate-400">
            <Loader2 class="w-10 h-10 animate-spin mb-4 text-blue-500" />
            <p class="text-sm font-medium animate-pulse">正在检索活跃会话...</p>
          </div>

          <div v-else-if="sessions.length === 0" class="py-20 flex flex-col items-center justify-center text-slate-400 text-center">
            <Globe class="w-16 h-16 opacity-10 mb-4" />
            <p class="text-lg font-bold text-slate-600 dark:text-slate-400">未检索到活跃会话</p>
            <p class="text-xs mt-1">这可能是由于系统回溯延迟或所有连接已重置</p>
          </div>

          <div v-else class="space-y-4">
            <div 
              v-for="session in sessions" 
              :key="session.id"
              :class="[
                'group relative p-6 rounded-[2rem] border transition-all flex items-center justify-between',
                session.id === currentSessionID 
                  ? 'bg-blue-500/[0.03] border-blue-500/30 dark:border-blue-500/20' 
                  : 'bg-slate-50 dark:bg-white/[0.02] border-slate-100 dark:border-white/5 hover:border-slate-300 dark:hover:border-white/10'
              ]"
            >
              <div class="flex items-center gap-5">
                <div :class="[
                  'p-4 rounded-2xl shadow-inner transition-colors',
                  session.id === currentSessionID 
                    ? 'bg-blue-500/20 text-blue-600 dark:text-blue-400' 
                    : 'bg-slate-200/50 dark:bg-white/5 text-slate-500 group-hover:bg-slate-200 dark:group-hover:bg-white/10'
                ]">
                  <component :is="getDeviceIcon(session.user_agent)" class="w-6 h-6" />
                </div>

                <div class="space-y-1">
                  <div class="flex items-center gap-2">
                    <span class="font-bold text-slate-900 dark:text-white">
                      {{ getOSName(session.user_agent) }} · {{ getBrowserName(session.user_agent) }}
                    </span>
                    <span v-if="session.id === currentSessionID" class="px-2 py-0.5 bg-blue-500/10 text-blue-600 dark:text-blue-400 text-[10px] font-black uppercase rounded-full border border-blue-500/20">
                      当前活跃
                    </span>
                  </div>
                  
                  <div class="flex flex-wrap items-center gap-x-4 gap-y-1 text-[11px] text-slate-500 font-medium">
                    <div class="flex items-center gap-1.5">
                      <Globe class="w-3 h-3" />
                      {{ session.ip || 'Unknown IP' }}
                    </div>
                    <div class="flex items-center gap-1.5">
                      <Clock class="w-3 h-3" />
                      最后活跃: {{ formatDate(session.last_active) }}
                    </div>
                    <div class="flex items-center gap-1.5">
                      <Calendar class="w-3 h-3" />
                      创建时间: {{ formatDate(session.created_at) }}
                    </div>
                  </div>
                </div>
              </div>

              <button 
                v-if="session.id !== currentSessionID"
                @click="revokeSession(session.id, false)"
                class="p-3 text-slate-400 hover:text-red-500 hover:bg-red-500/10 rounded-xl transition-all opacity-0 group-hover:opacity-100"
                title="终止会话"
              >
                <Trash2 class="w-5 h-5" />
              </button>
            </div>
          </div>
        </div>

        <!-- 底部提示 -->
        <div class="p-8 bg-slate-50 dark:bg-white/[0.02] border-t border-slate-100 dark:border-white/5">
          <div class="flex items-start gap-4">
            <div class="bg-amber-500/10 p-2 rounded-lg mt-1">
              <ShieldCheck class="w-4 h-4 text-amber-500" />
            </div>
            <div>
              <p class="text-xs font-bold text-slate-900 dark:text-slate-300 uppercase tracking-wider">安全提示 / Protocol</p>
              <p class="text-[10px] text-slate-500 mt-1 leading-relaxed">
                如果您在列表中发现不属于您的可疑设备，请立即终止其访问权限并更改系统密码。恶意劫持身份会导致实验数据严重流失。
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.modal-enter-active, .modal-leave-active { transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1); }
.modal-enter-from, .modal-leave-to { opacity: 0; transform: scale(0.95); }
</style>
