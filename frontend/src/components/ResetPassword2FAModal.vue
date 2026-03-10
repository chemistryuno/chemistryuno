<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { Lock, Fingerprint, Loader2, Eye, EyeOff, Mail, Shield, AlertTriangle, HelpCircle } from 'lucide-vue-next'
import { authAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'

defineProps<{
  show: boolean
  loading: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'submit', email: string, code: string, newPassword: string): void
}>()

const dialog = useDialog()
const smtpEnabled = ref(false)
const identifier = ref('')

const code = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const showPassword = ref(false)
const recoveryMode = ref<'2fa' | 'email' | 'security_question'>('email')
const emailLoading = ref(false)
const countdown = ref(0)

// 密保问题相关
const securityQuestion = ref('')
const securityAnswer = ref('')
const loadingQuestion = ref(false)

onMounted(async () => {
  try {
    const res = await authAPI.getAuthConfig()
    smtpEnabled.value = res.data.smtp_enabled
    if (smtpEnabled.value) {
      recoveryMode.value = 'email'
    } else {
      recoveryMode.value = 'security_question'
    }
  } catch (err) {
    console.error('获取配置失败', err)
  }
})

// 当标识符变化时，若在密保问题模式，尝试获取密保问题
watch([identifier, recoveryMode], async () => {
  if (recoveryMode.value === 'security_question' && identifier.value.length >= 3) {
    loadingQuestion.value = true
    securityQuestion.value = ''
    try {
      const res = await authAPI.getSecurityQuestion(identifier.value)
      if (res.data.has_security_question) {
        securityQuestion.value = res.data.security_question
      }
    } catch {
      // ignore
    } finally {
      loadingQuestion.value = false
    }
  }
})

const handleSendCode = async () => {
  if (!identifier.value) {
    alert('请输入您的注册邮箱')
    return
  }

  emailLoading.value = true
  try {
    await authAPI.sendCode(identifier.value, 'reset')
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
        email: identifier.value,
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

  if (recoveryMode.value === 'security_question') {
    if (!securityAnswer.value) {
      alert('请输入密保答案')
      return
    }
    emailLoading.value = true
    try {
      await authAPI.resetPasswordBySecurityQuestion({
        username: identifier.value,
        security_answer: securityAnswer.value,
        new_password: newPassword.value
      })
      emit('close')
      dialog.showAlert('访问凭证已重置，请使用新密码重新登录。', '协议同步成功')
    } catch (err: any) {
      alert(err.response?.data?.error || '重置失败，请检查密保答案')
    } finally {
      emailLoading.value = false
    }
    return
  }

  emit('submit', identifier.value, code.value, newPassword.value)
}
</script>

<template>
  <div v-if="show" class="fixed inset-0 z-[100] flex items-center justify-center p-4 backdrop-blur-2xl bg-slate-900/60 dark:bg-black/90">
    <div class="bg-white/90 dark:bg-slate-900/90 border border-white dark:border-white/10 rounded-[3rem] p-6 md:p-8 max-w-md w-full shadow-[0_0_50px_rgba(0,0,0,0.5)] relative animate-in fade-in zoom-in duration-300 backdrop-blur-3xl">
      <div class="flex flex-col items-center mb-6">
        <div class="relative group">
          <div class="absolute -inset-2 bg-cyan-500/20 rounded-full blur animate-pulse"></div>
          <div class="relative w-12 h-12 bg-gradient-to-tr from-cyan-600 to-blue-700 rounded-xl flex items-center justify-center mb-3 shadow-xl">
            <Mail v-if="recoveryMode === 'email'" class="w-6 h-6 text-white" />
            <HelpCircle v-else-if="recoveryMode === 'security_question'" class="w-6 h-6 text-white" />
            <Fingerprint v-else class="w-6 h-6 text-white" />
          </div>
        </div>
        <h3 class="text-2xl font-black text-slate-900 dark:text-white tracking-tight text-center">
          {{ recoveryMode === 'email' ? '凭证回收' : recoveryMode === 'security_question' ? '密保验证' : '协议重授' }}
        </h3>
        <p class="text-[9px] font-black mt-1.5 uppercase tracking-[0.3em] font-mono text-cyan-600 dark:text-cyan-400">AUTHORIZED RECOVERY PROTOCOL</p>
      </div>

      <!-- Security Notice -->
      <div class="mb-6 p-3 bg-amber-500/10 border border-amber-500/20 rounded-xl flex gap-2.5 items-start backdrop-blur-sm">
        <AlertTriangle class="w-4 h-4 text-amber-600 shrink-0 mt-0.5" />
        <div class="text-[10px] text-amber-700 dark:text-amber-500 font-bold leading-relaxed">
          <p class="font-black mb-1 uppercase tracking-widest text-amber-600">协议警示 / WARNING</p>
          重置密码是一个高风险协议变更。请确保您的新密钥复杂度符合实验室安全标准。
        </div>
      </div>

      <!-- Mode Selector -->
      <div class="flex bg-slate-100 dark:bg-white/5 p-1 rounded-xl mb-6 border border-slate-200 dark:border-white/10 gap-1">
        <button
          v-if="smtpEnabled"
          @click="recoveryMode = 'email'"
          :class="[
            'flex-1 py-2.5 text-[10px] font-black uppercase tracking-widest rounded-lg transition-all',
            recoveryMode === 'email' ? 'bg-gradient-to-r from-cyan-600 to-blue-700 text-white shadow-lg' : 'text-slate-500 hover:text-slate-300'
          ]"
        >
          邮箱
        </button>
        <button
          @click="recoveryMode = 'security_question'"
          :class="[
            'flex-1 py-2.5 text-[10px] font-black uppercase tracking-widest rounded-lg transition-all',
            recoveryMode === 'security_question' ? 'bg-gradient-to-r from-amber-500 to-orange-600 text-white shadow-lg' : 'text-slate-500 hover:text-slate-300'
          ]"
        >
          密保
        </button>
        <button
          @click="recoveryMode = '2fa'"
          :class="[
            'flex-1 py-2.5 text-[10px] font-black uppercase tracking-widest rounded-lg transition-all',
            recoveryMode === '2fa' ? 'bg-gradient-to-r from-cyan-600 to-blue-700 text-white shadow-lg' : 'text-slate-500 hover:text-slate-300'
          ]"
        >
          2FA
        </button>
      </div>

      <form @submit.prevent="handleReset" class="space-y-4">
        <div class="space-y-4">
          <div class="relative group">
            <Mail class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 dark:text-slate-500 group-focus-within:text-cyan-500 transition-colors" />
            <input
              v-model="identifier"
              type="text"
              :placeholder="recoveryMode === 'email' ? '注册邮箱' : '用户名'"
              class="w-full bg-slate-50/50 dark:bg-black/40 border border-slate-200 dark:border-white/10 focus:border-cyan-500/50 rounded-xl py-4 pl-12 pr-5 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 font-bold text-sm"
              required
            />
          </div>

          <!-- 邮箱模式：发送验证码 -->
          <div v-if="recoveryMode === 'email'" class="relative group flex gap-2.5">
            <div class="relative flex-1">
              <Shield class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 dark:text-slate-500 group-focus-within:text-cyan-500 transition-colors" />
              <input
                v-model="code"
                type="text"
                placeholder="指令校验码"
                maxlength="6"
                class="w-full bg-slate-50/50 dark:bg-black/40 border border-slate-200 dark:border-white/10 focus:border-cyan-500/50 rounded-xl py-4 pl-12 pr-5 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 font-bold text-sm"
                required
              />
            </div>
            <button
              type="button"
              @click="handleSendCode"
              :disabled="countdown > 0 || emailLoading"
              class="px-4 rounded-xl font-black text-[10px] uppercase tracking-widest transition-all bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 hover:bg-cyan-500/20 disabled:opacity-50 border border-cyan-500/20 min-w-[70px]"
            >
              {{ countdown > 0 ? `${countdown}S` : (emailLoading ? '...' : '发送') }}
            </button>
          </div>

          <!-- 密保模式 -->
          <div v-if="recoveryMode === 'security_question'" class="space-y-3">
            <div v-if="loadingQuestion" class="text-center text-xs text-slate-400 py-2">
              <Loader2 class="w-4 h-4 animate-spin inline mr-1" />查询密保问题中...
            </div>
            <div v-else-if="securityQuestion" class="p-3 bg-amber-50 dark:bg-amber-500/10 border border-amber-200 dark:border-amber-500/20 rounded-xl">
              <p class="text-[10px] font-black text-amber-600 dark:text-amber-400 uppercase tracking-widest mb-1">密保问题</p>
              <p class="text-sm font-bold text-slate-800 dark:text-slate-200">{{ securityQuestion }}</p>
            </div>
            <div v-else-if="identifier.length >= 3" class="text-center text-xs text-slate-400 py-1">
              该用户未设置密保问题，请使用其他方式
            </div>
            <div v-if="securityQuestion" class="relative group">
              <HelpCircle class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 dark:text-slate-500 group-focus-within:text-amber-500 transition-colors" />
              <input
                v-model="securityAnswer"
                type="text"
                placeholder="密保答案"
                class="w-full bg-slate-50/50 dark:bg-black/40 border border-slate-200 dark:border-white/10 focus:border-amber-500/50 rounded-xl py-4 pl-12 pr-5 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 font-bold text-sm"
                required
              />
            </div>
          </div>

          <div class="relative group">
            <Lock class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors" />
            <input
              v-model="newPassword"
              :type="showPassword ? 'text' : 'password'"
              placeholder="新访问密钥"
              class="w-full bg-slate-50/50 dark:bg-black/40 border border-slate-200 dark:border-white/10 focus:border-blue-500/50 rounded-xl py-4 pl-12 pr-11 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 font-bold text-sm"
              required
            />
            <button
              type="button"
              @click="showPassword = !showPassword"
              class="absolute right-4 top-1/2 -translate-y-1/2 text-slate-400 hover:text-blue-500 transition-colors"
            >
              <component :is="showPassword ? EyeOff : Eye" class="w-4 h-4" />
            </button>
          </div>

          <div class="relative group">
            <Shield class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors" />
            <input
              v-model="confirmPassword"
              type="password"
              placeholder="重复确认新密钥"
              class="w-full bg-slate-50/50 dark:bg-black/40 border border-slate-200 dark:border-white/10 focus:border-blue-500/50 rounded-xl py-4 pl-12 pr-5 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 font-bold text-sm"
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
                class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 focus:border-blue-500/50 rounded-xl py-3.5 pl-11 pr-4 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 font-mono tracking-[0.3em] text-center text-base italic"
                :required="recoveryMode === '2fa'"
              />
            </div>
          </div>
        </div>

        <div class="flex gap-3 pt-3">
          <button
            type="button"
            @click="$emit('close')"
            class="flex-1 py-3.5 bg-slate-50 dark:bg-white/5 hover:bg-slate-100 dark:hover:bg-white/10 border border-slate-200 dark:border-white/5 rounded-xl font-bold transition-all text-slate-500 dark:text-slate-400 uppercase text-[10px] tracking-widest"
          >
            中止
          </button>

          <button
            type="submit"
            :disabled="loading || emailLoading"
            class="flex-[2] py-3.5 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-300 dark:disabled:bg-slate-700 rounded-xl font-black transition-all text-white shadow-lg shadow-blue-500/20 active:scale-95 flex items-center justify-center gap-2 uppercase text-[10px] tracking-widest"
          >
            <Loader2 v-if="loading || emailLoading" class="w-3 h-3 animate-spin" />
            {{ recoveryMode === 'email' ? '确认同步新密码' : recoveryMode === 'security_question' ? '密保验证并重置' : '验证 2FA 并重置' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
