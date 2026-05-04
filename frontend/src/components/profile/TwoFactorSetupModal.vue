<script setup lang="ts">
import { ref, watch } from 'vue'
import { Fingerprint, Loader2, Lock } from 'lucide-vue-next'

defineProps<{
  show: boolean
  qrCode: string
  loading: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'enable', code: string, password: string): void
}>()

const verificationCode = ref('')
const currentPassword = ref('')

// 监控 show 状态以禁用/启用背景滚动
watch((props: any) => props.show, (show) => {
  if (show) {
    document.documentElement.style.overflow = 'hidden'
    document.body.style.overflow = 'hidden'
  } else {
    document.documentElement.style.overflow = ''
    document.body.style.overflow = ''
  }
}, { immediate: true })
</script>

<template>
  <Teleport to="body">
  <div v-if="show" class="viewport-modal-overlay z-[100] p-4 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md">
    <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[3rem] p-10 max-w-md w-full shadow-2xl relative animate-in fade-in zoom-in duration-300 overflow-y-auto max-h-[90vh]">
      <h3 class="text-2xl font-black mb-4 italic uppercase text-center text-slate-900 dark:text-white">配置双重验证 / 2FA Config</h3>
      <p class="text-slate-500 text-xs text-center mb-8">请使用手机验证器应用扫描下方二维码，并在下方输入当前账户密码以确认身份</p>
      
      <div class="flex flex-col items-center gap-6">
        <div class="bg-white p-4 rounded-[2rem] shadow-[0_0_30px_rgba(0,0,0,0.05)] dark:shadow-[0_0_30px_rgba(255,255,255,0.1)]">
          <img :src="qrCode" alt="2FA QR Code" class="w-48 h-48" />
        </div>

        <div class="w-full space-y-4">
          <div class="relative group">
            <Lock class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors" />
            <input
              v-model="currentPassword"
              type="password"
              placeholder="请输入当前登录密码"
              class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-4 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 text-sm font-bold"
              required
            />
          </div>

          <div class="relative group">
            <Fingerprint class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors" />
            <input
              v-model="verificationCode"
              type="text"
              maxlength="6"
              placeholder="000000"
              class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-4 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 font-mono tracking-[0.5em] text-center text-xl"
              required
            />
          </div>
        </div>

        <div class="flex gap-4 w-full">
          <button 
            @click="$emit('close')" 
            class="flex-1 py-4 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 rounded-2xl font-bold transition-all text-slate-500 dark:text-slate-400"
          >
            取消
          </button>
          <button 
            @click="$emit('enable', verificationCode, currentPassword)"
            :disabled="loading || verificationCode.length !== 6 || !currentPassword"
            class="flex-[2] py-4 bg-gradient-to-r from-emerald-600 to-emerald-500 hover:from-emerald-500 hover:to-emerald-400 rounded-2xl font-black text-white shadow-xl shadow-emerald-500/20 disabled:scale-100 active:scale-95 flex items-center justify-center gap-2"
          >
            <Loader2 v-if="loading" class="w-5 h-5 animate-spin" />
            激活保护
          </button>
        </div>
      </div>
    </div>
  </div>
  </Teleport>
</template>
