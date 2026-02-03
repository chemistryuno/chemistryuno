<script setup lang="ts">
import { ref } from 'vue'
import { User, Lock, Fingerprint, Loader2, Eye, EyeOff, Cpu } from 'lucide-vue-next'
import { authAPI } from '../utils/api'
import { get } from '@github/webauthn-json'
import { useDialog } from '../utils/dialog'

defineProps<{
  show: boolean
  loading: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'submit', username: string, code: string, newPassword: string): void
}>()

const dialog = useDialog()
const username = ref('')
const code = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const showPassword = ref(false)
const recoveryMode = ref<'2fa' | 'webauthn'>('2fa')
const webauthnLoading = ref(false)

const handleReset = () => {
  if (newPassword.value !== confirmPassword.value) {
    alert('两次输入的密码不一致')
    return
  }
  emit('submit', username.value, code.value, newPassword.value)
}

const handleWebAuthnRecovery = async () => {
  if (!username.value) {
    alert('请输入用户名')
    return
  }
  if (!newPassword.value || newPassword.value !== confirmPassword.value) {
    alert('请正确设置新密码')
    return
  }

  webauthnLoading.value = true
  try {
    const res = await authAPI.beginResetPasswordWebAuthn(username.value)
    const credential = await get(res.data)
    await authAPI.finishResetPasswordWebAuthn(username.value, newPassword.value, credential)
    
    emit('close')
    dialog.showAlert('已通过硬件密钥验证身份，凭证重置成功。', '回收协议完成')
  } catch (err: any) {
    console.error('WebAuthn recovery error:', err)
    alert(err.response?.data?.error || '硬件密钥验证失败')
  } finally {
    webauthnLoading.value = false
  }
}
</script>

<template>
  <div v-if="show" class="fixed inset-0 z-[100] flex items-center justify-center p-4 backdrop-blur-xl bg-slate-900/40 dark:bg-black/80">
    <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[3rem] p-10 max-w-md w-full shadow-2xl relative animate-in fade-in zoom-in duration-300">
      <div class="flex flex-col items-center mb-6">
        <div class="w-16 h-16 bg-blue-600/10 rounded-2xl flex items-center justify-center mb-4">
          <Cpu v-if="recoveryMode === 'webauthn'" class="w-8 h-8 text-blue-600 dark:text-blue-500" />
          <Fingerprint v-else class="w-8 h-8 text-blue-600 dark:text-blue-500" />
        </div>
        <h3 class="text-2xl font-black italic uppercase text-slate-900 dark:text-white tracking-tight text-center">
          {{ recoveryMode === 'webauthn' ? '硬件凭证回放' : '2FA 凭证授权' }}
        </h3>
        <p class="text-slate-500 text-[10px] font-black mt-2 uppercase tracking-[0.2em] font-mono">AUTHORIZED RECOVERY PROTOCOL</p>
      </div>

      <!-- Mode Selector -->
      <div class="flex bg-slate-100 dark:bg-white/5 p-1 rounded-xl mb-6">
        <button 
          @click="recoveryMode = '2fa'"
          :class="[
            'flex-1 py-2 text-[10px] font-black uppercase tracking-widest rounded-lg transition-all',
            recoveryMode === '2fa' ? 'bg-blue-600 text-white shadow-lg' : 'text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'
          ]"
        >
          TOTP 验证
        </button>
        <button 
          @click="recoveryMode = 'webauthn'"
          :class="[
            'flex-1 py-2 text-[10px] font-black uppercase tracking-widest rounded-lg transition-all',
            recoveryMode === 'webauthn' ? 'bg-blue-600 text-white shadow-lg' : 'text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'
          ]"
        >
          硬件密钥
        </button>
      </div>

      <form @submit.prevent="handleReset" class="space-y-4">
        <div class="space-y-4">
          <div class="relative group">
            <User class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-600 dark:group-focus-within:text-blue-500 transition-colors" />
            <input
              v-model="username"
              type="text"
              placeholder="确认用户名 / Entry Username"
              class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-4 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 font-bold text-sm"
              required
            />
          </div>

          <div class="relative group">
            <Lock class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-600 dark:group-focus-within:text-blue-500 transition-colors" />
            <input
              v-model="newPassword"
              :type="showPassword ? 'text' : 'password'"
              placeholder="审核新密码 / New Credentials"
              class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-12 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 text-sm font-mono"
              required
            />
            <button 
              type="button"
              @click="showPassword = !showPassword"
              class="absolute right-4 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-900 dark:text-slate-500 dark:hover:text-white transition-colors"
            >
              <EyeOff v-if="showPassword" class="w-4 h-4" />
              <Eye v-else class="w-4 h-4" />
            </button>
          </div>

          <div class="relative group">
            <Lock class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-600 dark:group-focus-within:text-blue-500 transition-colors" />
            <input
              v-model="confirmPassword"
              :type="showPassword ? 'text' : 'password'"
              placeholder="重复确认 / Re-Verify"
              class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-4 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 text-sm font-mono"
              required
            />
          </div>
          
          <div v-if="recoveryMode === '2fa'" class="relative animate-in slide-in-from-top-2 duration-300">
            <div class="relative group">
              <Fingerprint class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-600 dark:group-focus-within:text-blue-500 transition-colors" />
              <input
                v-model="code"
                type="text"
                maxlength="6"
                placeholder="6 位动态验证码"
                class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-4 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 font-mono tracking-[0.3em] text-center text-lg italic"
                :required="recoveryMode === '2fa'"
              />
            </div>
          </div>
        </div>

        <div class="flex gap-4 pt-4">
          <button 
            type="button"
            @click="$emit('close')" 
            class="flex-1 py-4 bg-slate-50 dark:bg-white/5 hover:bg-slate-100 dark:hover:bg-white/10 border border-slate-200 dark:border-white/5 rounded-2xl font-bold transition-all text-slate-500 dark:text-slate-400 uppercase text-[10px] tracking-widest"
          >
            中止实验
          </button>

          <!-- Submit Button based on Mode -->
          <button 
            v-if="recoveryMode === '2fa'"
            type="submit"
            :disabled="loading"
            class="flex-1 py-4 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-300 dark:disabled:bg-slate-700 rounded-2xl font-black transition-all text-white shadow-lg shadow-blue-500/20 active:scale-95 flex items-center justify-center gap-2 uppercase text-[10px] tracking-widest"
          >
            <Loader2 v-if="loading" class="w-3 h-3 animate-spin" />
            提交重置
          </button>
          <button 
            v-else
            type="button"
            @click="handleWebAuthnRecovery"
            :disabled="webauthnLoading"
            class="flex-1 py-4 bg-emerald-600 hover:bg-emerald-700 disabled:bg-slate-300 dark:disabled:bg-slate-700 rounded-2xl font-black transition-all text-white shadow-lg shadow-emerald-500/20 active:scale-95 flex items-center justify-center gap-2 uppercase text-[10px] tracking-widest"
          >
            <Loader2 v-if="webauthnLoading" class="w-3 h-3 animate-spin" />
            调起硬件密钥
          </button>
        </div>
      </form>
    </div>
  </div>
</template>