<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Swords, Check, Timer, FlaskConical } from 'lucide-vue-next'
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

const handleResponse = async (accept: boolean) => {
  try {
    await gameAPI.respondToDuel(props.invite.challenger_uid, accept)
  } catch (err: any) {
    showAlert(err.response?.data?.error || '响应失败', '错误')
  } finally {
    emit('close')
  }
}

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
})
</script>

<template>
  <div class="viewport-modal-overlay z-[1000] p-4 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md" @click="handleResponse(false)">

    <div class="viewport-modal-panel relative w-full max-w-sm bg-white dark:bg-[#121216] border border-slate-200 dark:border-white/10 rounded-[40px] shadow-2xl overflow-hidden animate-in zoom-in duration-300 pointer-events-auto">
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
            class="flex-1 h-14 rounded-2xl bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 text-slate-500 font-bold text-xs uppercase tracking-widest hover:bg-slate-50 transition-all"
          >
            拒绝同步
          </button>
          <button 
            @click="handleResponse(true)"
            class="flex-1 h-14 rounded-2xl bg-blue-600 hover:bg-blue-500 text-white font-black text-xs uppercase tracking-widest shadow-xl shadow-blue-500/20 active:scale-95 transition-all flex items-center justify-center gap-2"
          >
            <Check class="w-4 h-4" />
            建立连接
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
