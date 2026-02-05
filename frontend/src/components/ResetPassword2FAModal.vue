<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { User, Lock, Fingerprint, Loader2, Eye, EyeOff, Mail, Shield, AlertTriangle } from 'lucide-vue-next'
import { authAPI } from '../utils/api'
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
const smtpEnabled = ref(false)
const username = ref('')

const code = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const showPassword = ref(false)
const recoveryMode = ref<'2fa' | 'email'>('2fa')
const emailLoading = ref(false)
const countdown = ref(0)

onMounted(async () => {
  try {
    const res = await authAPI.getAuthConfig()
    smtpEnabled.value = res.data.smtp_enabled
    if (smtpEnabled.value) {
      recoveryMode.value = 'email'
    }
  } catch (err) {
    console.error('获取配置失败', err)
  }
})

const handleSendCode = async () => {
  if (!username.value) {
    alert('请输入您的注册邮箱')
    return
  }
  emailLoading.value = true
  try {
    await authAPI.sendCode(username.value, 'reset')
    dialog.showAlert('验证码已发送至您的电子邮箱，请在10分钟内完成重置。', '发送成功')
    countdown.value = 60
    const timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) clearInterval(timer)
    }, 1000)
  } catch (err: any) {
    alert(err.response?.data?.error || '发送失败')
  } finally {
    emailLoading.value = false
  }
}

const handleReset = async () => {
  if (newPassword.value !== confirmPassword.value) {
    alert('两次输入的密码不一致')
    return
  }

  if (recoveryMode.value === 'email') {
    emailLoading.value = true
    try {
      await authAPI.resetPasswordByEmail({
        email: username.value,
        code: code.value,
        new_password: newPassword.value
      })
      emit('close')
      dialog.showAlert('访问凭证已重置，请尝试使用新密码重新登录。', '协议同步成功')
    } catch (err: any) {
      alert(err.response?.data?.error || '重置失败')
    } finally {
      emailLoading.value = false
    }
    return
  }
  
  emit('submit', username.value, code.value, newPassword.value)
}
</script>

<template>
  <div v-if="show" class="fixed inset-0 z-[100] flex items-center justify-center p-4 backdrop-blur-xl bg-slate-900/40 dark:bg-black/80">
    <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[3rem] p-10 max-w-md w-full shadow-2xl relative animate-in fade-in zoom-in duration-300">
      <div class="flex flex-col items-center mb-6">
        <div class="w-16 h-16 bg-blue-600/10 rounded-2xl flex items-center justify-center mb-4">
          <Mail v-if="recoveryMode === 'email'" class="w-8 h-8 text-blue-600 dark:text-blue-500" />
          <Fingerprint v-else class="w-8 h-8 text-blue-600 dark:text-blue-500" />
        </div>
        <h3 class="text-2xl font-black italic uppercase text-slate-900 dark:text-white tracking-tight text-center">
          {{ recoveryMode === 'email' ? '电子邮箱凭证回收' : '2FA 凭证授权' }}
        </h3>
        <p class="text-slate-500 text-[10px] font-black mt-2 uppercase tracking-[0.2em] font-mono">AUTHORIZED RECOVERY PROTOCOL</p>
      </div>

      <!-- Security Notice -->
      <div class="mb-6 p-4 bg-amber-500/10 border border-amber-500/20 rounded-2xl flex gap-3 items-start">
        <AlertTriangle class="w-5 h-5 text-amber-600 shrink-0 mt-0.5" />
        <div class="text-[10px] text-amber-700 dark:text-amber-500 font-medium leading-relaxed">
          <p class="font-black mb-1 uppercase tracking-widest">安全建议 / Security Protocol</p>
          重置密码是一个高风险操作。请确保您的新密码包含大小写字母、数字及特殊符号，且不要在其他平台重复使用。请务必保护好您的邮箱与 2FA 凭证。
        </div>
      </div>

      <!-- Mode Selector -->
      <div class="flex bg-slate-100 dark:bg-white/5 p-1 rounded-xl mb-6">
        <button 
          v-if="smtpEnabled"
          @click="recoveryMode = 'email'"
          :class="[
            'flex-1 py-2 text-[10px] font-black uppercase tracking-widest rounded-lg transition-all',
            recoveryMode === 'email' ? 'bg-blue-600 text-white shadow-lg' : 'text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'
          ]"
        >
          邮箱验证
        </button>
        <button 
          @click="recoveryMode = '2fa'"
          :class="[
            'flex-1 py-2 text-[10px] font-black uppercase tracking-widest rounded-lg transition-all',
            recoveryMode === '2fa' ? 'bg-blue-600 text-white shadow-lg' : 'text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'
          ]"
        >
          2FA 验证
        </button>
      </div>

      <form @submit.prevent="handleReset" class="space-y-4">
        <div class="space-y-4">
          <div class="relative group">
            <component :is="recoveryMode === 'email' ? Mail : User" class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-600 dark:group-focus-within:text-blue-500 transition-colors" />
            <input
              v-model="username"
              type="text"
              :placeholder="recoveryMode === 'email' ? '确认注册邮箱 / Entry Email' : '确认用户名 / Entry Username'"
              class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-4 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 font-bold text-sm"
              required
            />
          </div>

          <div v-if="recoveryMode === 'email'" class="relative group flex gap-2">
            <div class="relative flex-1">
              <Shield class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-600 dark:group-focus-within:text-blue-500 transition-colors" />
              <input
                v-model="code"
                type="text"
                placeholder="验证码"
                maxlength="6"
                class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-4 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 font-bold text-sm"
                required
              />
            </div>
            <button 
              type="button"
              @click="handleSendCode"
              :disabled="countdown > 0 || emailLoading"
              class="px-4 rounded-2xl bg-slate-100 dark:bg-white/10 text-[10px] font-black uppercase text-slate-600 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-white/20 transition-all disabled:opacity-50 min-w-[80px]"
            >
              {{ countdown > 0 ? `${countdown}s` : (emailLoading ? '...' : '发送') }}
            </button>
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

          <button 
            type="submit"
            :disabled="loading || emailLoading"
            class="flex-[2] py-4 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-300 dark:disabled:bg-slate-700 rounded-2xl font-black transition-all text-white shadow-lg shadow-blue-500/20 active:scale-95 flex items-center justify-center gap-2 uppercase text-[10px] tracking-widest"
          >
            <Loader2 v-if="loading || emailLoading" class="w-3 h-3 animate-spin" />
            {{ recoveryMode === 'email' ? '确认同步新密码' : '验证 2FA 并重置' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>