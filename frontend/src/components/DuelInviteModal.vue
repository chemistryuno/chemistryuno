<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { Swords, Check, Timer, FlaskConical } from 'lucide-vue-next'
import { buttonClasses, modalClasses } from '@lib'
import { gameAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'

const props = defineProps<{
  invite: {
    challenger_name: string
    challenger_uid: number
  }
}>()

const emit = defineEmits(['close'])

const { showAlert } = useDialog()
const timeLeft = ref(20)
let timer: any = null
const duelModalClasses = modalClasses({
  width: 'sm',
  zIndex: 'z-[1000]',
  panelClassName: 'relative dark:bg-[#121216] rounded-[40px] overflow-hidden duration-300 pointer-events-auto',
})
const rejectButtonClass = buttonClasses({
  tone: 'secondary',
  variant: 'outline',
  size: 'xl',
  radius: 'xl',
  block: true,
  className: 'h-14 bg-white dark:bg-white/5 hover:bg-slate-50 shadow-none font-bold',
})
const acceptButtonClass = buttonClasses({
  tone: 'primary',
  size: 'xl',
  radius: 'xl',
  block: true,
  className: 'h-14 shadow-xl shadow-blue-500/20',
})

const handleResponse = async (accept: boolean) => {
  try {
    await gameAPI.respondToDuel(props.invite.challenger_uid, accept)
  } catch (err: any) {
    showAlert(err.response?.data?.error || '响应失败', '错误')
  } finally {
    emit('close')
  }
}

watch(() => props.invite, (newVal) => {
  if (newVal) {
    // 禁用背景滚动
    document.documentElement.style.overflow = 'hidden'
    document.body.style.overflow = 'hidden'
  } else {
    // 启用背景滚动
    document.documentElement.style.overflow = ''
    document.body.style.overflow = ''
  }
}, { immediate: true })

onMounted(() => {
  timer = setInterval(() => {
    timeLeft.value--
    if (timeLeft.value <= 0) {
      handleResponse(false)
    }
  }, 1000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
  // 确保清理时恢复背景滚动
  document.documentElement.style.overflow = ''
  document.body.style.overflow = ''
})
</script>

<template>
  <Teleport to="body">
    <div :class="duelModalClasses.overlay" @click="handleResponse(false)">

    <div :class="duelModalClasses.panel">
      <!-- Decoration -->
      <div class="absolute top-0 right-0 p-8 opacity-5">
        <Swords class="w-24 h-24 -mr-8 -mt-8" />
      </div>
      
      <div class="p-8 text-center">
        <div class="w-16 h-16 bg-rose-500/10 border border-rose-500/20 rounded-[24px] flex items-center justify-center mx-auto mb-6">
          <FlaskConical class="w-8 h-8 text-rose-500 animate-pulse" />
        </div>
        
        <h3 class="text-xl font-black text-slate-900 dark:text-white italic tracking-tighter uppercase mb-2">
          发现不稳定反应场
        </h3>
        <p class="text-[10px] font-mono text-slate-400 uppercase tracking-widest mb-6">Anomaly_Detected: High_Risk_Duel</p>
        
        <div class="py-6 px-4 bg-slate-50 dark:bg-white/5 rounded-3xl border border-slate-100 dark:border-white/5 mb-8">
           <p class="text-sm font-bold text-slate-600 dark:text-slate-300 italic leading-relaxed">
             研究员 <span class="text-rose-600 dark:text-rose-500 font-black not-italic">{{ invite.challenger_name }}</span> 
             向你发送了单挑对局同步请求。
           </p>
        </div>

        <div class="flex items-center justify-between mb-8 px-2">
          <div class="flex items-center gap-2">
            <Timer class="w-4 h-4 text-slate-400" />
            <span class="text-[10px] font-black font-mono text-slate-500">{{ timeLeft }}S 后自动失效</span>
          </div>
          <div class="h-1 flex-1 mx-4 bg-slate-100 dark:bg-white/5 rounded-full overflow-hidden">
             <div 
              class="h-full bg-rose-500 transition-all duration-1000 ease-linear"
              :style="{ width: (timeLeft / 20) * 100 + '%' }"
             ></div>
          </div>
        </div>

        <div class="flex gap-4">
          <button 
            @click="handleResponse(false)"
            :class="rejectButtonClass"
          >
            拒绝同步
          </button>
          <button 
            @click="handleResponse(true)"
            :class="acceptButtonClass"
          >
            <Check class="w-4 h-4" />
            建立连接
          </button>
        </div>
      </div>
    </div>
    </div>
  </Teleport>
</template>
