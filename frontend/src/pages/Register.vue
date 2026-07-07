<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { pageClassNames } from '@lib'
import { authAPI, clearAccountScopedCache } from '../utils/api'
import { useDialog } from '../utils/dialog'
import { Lock, FlaskConical, ShieldCheck, Loader2, Mail, User, HelpCircle, AtSign, FileText, X } from 'lucide-vue-next'
import OAuthLogos from '../components/icons/OAuthLogos.vue'
import websocket from '../utils/websocket'
import { API_BASE_URL } from '../utils/runtimeConfig'

// 注册模式: 'username' | 'email'
const registerMode = ref<'username' | 'email'>('username')

const username = ref('')
const email = ref('')
const nickname = ref('')
const password = ref('')
const confirmPassword = ref('')
const code = ref('')
const securityQuestion = ref('')
const securityAnswer = ref('')
const agreementAccepted = ref(false) // 是否同意协议
const error = ref('')
const loading = ref(false)
const codeLoading = ref(false)
const smtpEnabled = ref(false)
const githubEnabled = ref(false)
const msEnabled = ref(false)
const googleEnabled = ref(false)
const appleEnabled = ref(false)
const countdown = ref(0)
const router = useRouter()
const dialog = useDialog()
const { showAlert } = dialog

// 协议查看模态框
const showAgreementModal = ref(false)
const agreementType = ref<'user' | 'privacy'>('user')

const openAgreement = (type: 'user' | 'privacy') => {
  agreementType.value = type
  showAgreementModal.value = true
}

// 用户名只允许英文字母、数字、下划线
const usernameRegex = /^[a-zA-Z0-9_]+$/
// 昵称只允许中英文字母、数字和下划线
const nicknameRegex = /^[a-zA-Z0-9_\u4e00-\u9fa5]+$/

const hasOAuth = computed(() => githubEnabled.value || msEnabled.value || googleEnabled.value || appleEnabled.value)

onMounted(async () => {
  try {
    const res = await authAPI.getAuthConfig()
    smtpEnabled.value = res.data.smtp_enabled
    githubEnabled.value = res.data.github_enabled
    msEnabled.value = res.data.ms_enabled
    googleEnabled.value = res.data.google_enabled
    appleEnabled.value = res.data.apple_enabled
    // 如果SMTP未开启，强制使用用户名注册
    if (!res.data.smtp_enabled) {
      registerMode.value = 'username'
    }
  } catch (err) {
    console.error('获取配置失败', err)
  }
})

const handleLoginSuccess = (token: string, user: any) => {
  clearAccountScopedCache()
  websocket.disconnect()
  // Token已由后端通过HttpOnly Cookie设置，前端不需要存储
  localStorage.setItem('user', JSON.stringify(user))
  websocket.connect()
  window.dispatchEvent(new Event('auth-changed'))
  router.push('/')
}

const handleOAuthLogin = (provider: 'github' | 'ms' | 'google' | 'apple') => {
  loading.value = true
  error.value = ''

  const width = 600
  const height = 700
  const left = window.screen.width / 2 - width / 2
  const top = window.screen.height / 2 - height / 2

  const url = `${API_BASE_URL}/auth/${provider}/login`
  const popup = window.open(url, 'OAuth Login', `width=${width},height=${height},left=${left},top=${top}`)

  if (!popup) {
    loading.value = false
    showAlert('弹出窗口被拦截，请允许弹出窗口后重试。', '拦截提示')
    return
  }

  let oauthFinished = false

  const cleanup = () => {
    window.removeEventListener('message', messageHandler)
    clearInterval(timer)
  }

  const messageHandler = (event: MessageEvent) => {
    if (!event.data || typeof event.data !== 'object') return

    if (event.data.type === 'oauth-success') {
      oauthFinished = true
      cleanup()
      const { token, user } = event.data
      handleLoginSuccess(token, user)
      loading.value = false
    } else if (event.data.type === 'oauth-error') {
      oauthFinished = true
      cleanup()
      if (event.data.error === 'NEED_EMAIL') {
        error.value = '第三方账号未公开邮箱，请先手动填写用户名、昵称和密码进行常规注册，稍后在个人设置中绑定。'
        showAlert('您的第三方账号未设置或未公开邮箱地址。请先完成常规注册。', '需要补充信息')
      } else {
        error.value = event.data.error || '授权失败'
      }
      loading.value = false
    }
  }

  window.addEventListener('message', messageHandler)

  const timer = setInterval(() => {
    if (popup.closed) {
      clearInterval(timer)
      setTimeout(() => {
        if (!oauthFinished) {
          cleanup()
          loading.value = false
        }
      }, 600)
    }
  }, 1000)
}

const handleSendCode = async () => {
  if (!email.value || !email.value.includes('@')) {
    error.value = '请输入有效的邮箱地址'
    return
  }

  codeLoading.value = true
  try {
    await authAPI.sendCode(email.value, 'register')
    showAlert('验证码已发送至您的邮箱，请查收。', '发送成功')
    countdown.value = 60
    const timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) clearInterval(timer)
    }, 1000)
  } catch (err: any) {
    error.value = err.response?.data?.error || '验证码发送失败'
  } finally {
    codeLoading.value = false
  }
}

const handleSubmit = async () => {
  error.value = ''

  if (!agreementAccepted.value) {
    error.value = '请先阅读并同意用户协议及隐私政策。'
    return
  }

  // 验证用户名
  if (!username.value || username.value.length < 3) {
    error.value = '用户名至少需要3个字符'
    return
  }
  if (!usernameRegex.test(username.value)) {
    error.value = '用户名只能包含英文字母、数字和下划线'
    return
  }

  // 验证昵称
  if (nickname.value) {
    if (nickname.value.length > 20) {
      error.value = '昵称最大长度为 20 个字符'
      return
    }
    if (!nicknameRegex.test(nickname.value)) {
      error.value = '昵称只能包含中英文字母、数字和下划线'
      return
    }
  }

  if (password.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return
  }

  if (password.value.length < 6) {
    error.value = '密码长度至少需要6位'
    return
  }

  // 只有用户名注册模式下必须设置密保
  if (registerMode.value === 'username') {
    if (!securityQuestion.value.trim()) {
      error.value = '请设置密保问题'
      return
    }

    if (!securityAnswer.value.trim()) {
      error.value = '请设置密保答案'
      return
    }
  }

  // 邮箱注册模式：邮箱必填，验证码必填
  if (registerMode.value === 'email') {
    if (!email.value || !email.value.includes('@')) {
      error.value = '请输入有效的邮箱地址'
      return
    }
    if (!code.value) {
      error.value = '请输入邮箱验证码'
      return
    }
  }

  loading.value = true

  try {
    await authAPI.register({
      username: username.value,
      email: registerMode.value === 'email' ? email.value : undefined,
      code: registerMode.value === 'email' ? code.value : undefined,
      nickname: nickname.value.trim(),
      password: password.value,
      security_question: securityQuestion.value,
      security_answer: securityAnswer.value,
    })
    await showAlert('注册成功，请使用用户名登录。', '研究员注册成功')
    router.push('/login')
  } catch (err: any) {
    const backendError = err.response?.data?.error || ''

    if (backendError.includes('username already taken')) {
      error.value = '该用户名已被占用，请换一个'
    } else if (backendError.includes('email already registered')) {
      error.value = '该邮箱已被注册'
    } else if (backendError.includes('verification code invalid')) {
      error.value = '验证码错误或已过期，请重新获取'
    } else if (backendError.includes('只能包含英文字母')) {
      error.value = '用户名只能包含英文字母、数字和下划线'
    } else if (backendError.includes('failed to create user')) {
      error.value = '档案创建失败，请联系管理员'
    } else if (backendError) {
      error.value = backendError
    } else {
      error.value = '注册请求异常，请稍后重试'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div :class="pageClassNames.auth">
    <div class="absolute top-[-10%] right-[-10%] w-[40%] h-[40%] bg-blue-500/5 rounded-full blur-[120px]"></div>
    <div class="absolute bottom-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-500/5 rounded-full blur-[120px]"></div>

    <div class="w-full max-w-lg relative z-10 animate-in fade-in zoom-in duration-500">
      <div class="glass-panel-light rounded-2xl sm:rounded-3xl shadow-[0_20px_60px_rgba(0,0,0,0.1)] dark:shadow-[0_20px_60px_rgba(0,0,0,0.3)] overflow-hidden">
        <div class="p-3 sm:p-4 md:p-6">
          <div class="flex flex-col items-center mb-3 sm:mb-4">
            <div class="w-10 h-10 sm:w-12 sm:h-12 bg-gradient-to-tr from-cyan-600 to-blue-700 rounded-xl sm:rounded-2xl flex items-center justify-center mb-1.5 sm:mb-2 shadow-lg transform rotate-3 hover:rotate-0 transition-transform duration-500">
              <FlaskConical class="w-5 h-5 sm:w-6 sm:h-6 text-white" />
            </div>
            <h1 class="text-lg sm:text-xl font-black text-slate-900 dark:text-slate-100 tracking-tighter">
              研究员<span class="text-blue-600">入职</span>
            </h1>
            <p class="text-slate-400 dark:text-slate-500 text-[10px] sm:text-xs font-black uppercase tracking-[0.2em] mt-0.5 font-mono">CREATE ACCESS CREDENTIAL</p>
          </div>

          <!-- 注册模式切换（仅 SMTP 开启时显示） -->
          <div v-if="smtpEnabled" class="flex gap-1 p-1 bg-slate-100 dark:bg-white/5 rounded-xl mb-3">
            <button
              type="button"
              @click="registerMode = 'username'"
              :class="[
                'flex-1 flex flex-col items-center justify-center gap-1 py-2.5 rounded-lg text-[10px] font-black tracking-widest transition-all duration-300',
                registerMode === 'username'
                  ? 'bg-white dark:bg-white/10 text-blue-600 shadow-sm'
                  : 'text-slate-400 hover:text-slate-600 dark:hover:text-white'
              ]"
            >
              <User class="w-4 h-4" />
              <span>仅用户名</span>
            </button>
            <button
              type="button"
              @click="registerMode = 'email'"
              :class="[
                'flex-1 flex flex-col items-center justify-center gap-1 py-2.5 rounded-lg text-[10px] font-black tracking-widest transition-all duration-300',
                registerMode === 'email'
                  ? 'bg-white dark:bg-white/10 text-cyan-600 shadow-sm'
                  : 'text-slate-400 hover:text-slate-600 dark:hover:text-white'
              ]"
            >
              <AtSign class="w-4 h-4" />
              <span>绑定邮箱</span>
            </button>
          </div>

          <!-- 模式提示 -->
          <div class="mb-5 px-3">
            <div 
              class="flex items-center gap-2 p-2.5 rounded-xl text-[10px] font-bold transition-all duration-300"
              :class="registerMode === 'username' ? 'bg-amber-500/5 text-amber-600 border border-amber-200/20' : 'bg-cyan-500/5 text-cyan-600 border border-cyan-200/20'"
            >
              <HelpCircle v-if="registerMode === 'username'" class="w-3.5 h-3.5" />
              <ShieldCheck v-else class="w-3.5 h-3.5" />
              <span>
                {{ registerMode === 'username' ? '仅通过用户名注册，建议后续在“安全中心”补全联系方式以防遗忘。' : '绑定邮箱可用于找回密码与接收科研通知。' }}
              </span>
            </div>
          </div>

          <form @submit.prevent="handleSubmit" class="space-y-2.5 sm:space-y-3">
            <div v-if="error" class="bg-red-50 dark:bg-red-500/10 border border-red-100 dark:border-red-500/20 text-red-500 px-2.5 py-2 rounded-xl mb-2.5 sm:mb-3 text-center text-xs font-bold animate-shake">
              {{ error }}
            </div>

            <!-- 用户名 -->
            <div class="relative group">
              <div class="absolute left-0 pl-2.5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 group-focus-within:text-cyan-500 transition-colors pointer-events-none">
                <User class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
              </div>
              <input
                v-model="username"
                type="text"
                required
                autocomplete="username"
                class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-9 pr-2.5 py-2.5 sm:py-3 rounded-xl sm:rounded-2xl focus:ring-2 focus:ring-cyan-500/20 focus:border-cyan-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold"
                placeholder="用户名 (字母/数字/下划线，登录用)"
              />
            </div>

            <!-- 昵称 -->
            <div class="relative group">
              <div class="absolute left-0 pl-2.5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 group-focus-within:text-cyan-500 transition-colors pointer-events-none">
                <FlaskConical class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
              </div>
              <input
                v-model="nickname"
                type="text"
                class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-9 pr-2.5 py-2.5 sm:py-3 rounded-xl sm:rounded-2xl focus:ring-2 focus:ring-cyan-500/20 focus:border-cyan-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold"
                :placeholder="`显示昵称（可选，默认使用用户名）`"
              />
            </div>

            <!-- 邮箱模式：邮箱+验证码（必填） -->
            <div v-if="registerMode === 'email'" class="space-y-3 p-3 bg-cyan-500/5 dark:bg-cyan-500/10 rounded-2xl border border-cyan-500/20 group-focus-within:border-cyan-500/50 transition-all duration-300">
              <div class="relative group">
                <div class="absolute left-0 pl-3 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-cyan-500 transition-colors pointer-events-none">
                  <Mail class="w-4 h-4" />
                </div>
                <input
                  v-model="email"
                  type="email"
                  required
                  autocomplete="email"
                  class="w-full bg-white/50 dark:bg-black/40 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-10 pr-3 py-2.5 sm:py-3 rounded-xl focus:ring-2 focus:ring-cyan-500/20 focus:border-cyan-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold"
                  placeholder="注册邮箱（必填）"
                />
              </div>

              <!-- 验证码（邮箱模式必填） -->
              <div class="relative group flex gap-2">
                <div class="relative flex-1 group">
                  <div class="absolute left-0 pl-3 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-cyan-500 transition-colors pointer-events-none">
                    <ShieldCheck class="w-4 h-4" />
                  </div>
                  <input
                    v-model="code"
                    type="text"
                    required
                    class="w-full bg-white/50 dark:bg-black/40 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-10 pr-3 py-2.5 sm:py-3 rounded-xl focus:ring-2 focus:ring-cyan-500/20 focus:border-cyan-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold"
                    placeholder="邮箱验证码"
                  />
                </div>
                <button
                  type="button"
                  @click="handleSendCode"
                  :disabled="codeLoading || countdown > 0"
                  class="px-5 rounded-xl font-black text-[10px] uppercase tracking-widest transition-all bg-cyan-600 text-white hover:bg-cyan-700 disabled:opacity-50 disabled:bg-slate-400 shadow-md shadow-cyan-500/20 min-w-[90px]"
                >
                  {{ countdown > 0 ? `${countdown}S` : (codeLoading ? '...' : '获取') }}
                </button>
              </div>
            </div>

            <!-- 密码 -->
            <div class="space-y-3">
              <div class="relative group">
                <div class="absolute left-0 pl-3 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-blue-500 transition-colors pointer-events-none">
                  <Lock class="w-4 h-4" />
                </div>
                <input
                  v-model="password"
                  type="password"
                  required
                  autocomplete="new-password"
                  aria-describedby="password-hint"
                  class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-10 pr-3 py-2.5 sm:py-3 rounded-xl sm:rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold"
                  placeholder="入职密钥（至少 6 位）"
                />
              </div>
              <p
                id="password-hint"
                class="px-1 text-[11px] font-medium transition-colors"
                :class="password.length === 0 ? 'text-slate-400' : password.length < 6 ? 'text-rose-400' : 'text-emerald-400'"
              >
                {{ password.length === 0 ? '密钥至少需要 6 位' : password.length < 6 ? `还需 ${6 - password.length} 位` : '密钥长度符合要求' }}
              </p>

              <div class="relative group">
                <div class="absolute left-0 pl-3 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-blue-500 transition-colors pointer-events-none">
                  <ShieldCheck class="w-4 h-4" />
                </div>
                <input
                  v-model="confirmPassword"
                  type="password"
                  required
                  autocomplete="new-password"
                  class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-10 pr-3 py-2.5 sm:py-3 rounded-xl sm:rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold"
                  placeholder="确认密钥"
                />
              </div>
            </div>

            <!-- 密保设置 -->
            <div v-if="registerMode === 'username'" class="border border-amber-200/50 dark:border-amber-500/20 rounded-xl sm:rounded-2xl p-2.5 sm:p-3 space-y-2 bg-amber-50/50 dark:bg-amber-500/5">
              <p class="text-[10px] sm:text-xs font-black text-amber-600 dark:text-amber-400 uppercase tracking-widest flex items-center gap-1.5">
                <HelpCircle class="w-3.5 h-3.5" />
                密保问题（用于无邮箱时验证身份）
              </p>
              <div class="relative group">
                <input
                  v-model="securityQuestion"
                  type="text"
                  :required="registerMode === 'username'"
                  maxlength="200"
                  class="w-full bg-white dark:bg-black/20 border border-amber-200 dark:border-amber-500/30 text-slate-900 dark:text-slate-100 px-2.5 py-2 sm:py-2.5 rounded-xl focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold"
                  placeholder="自定义密保问题（如：我的第一只宠物叫什么？）"
                />
              </div>
              <div class="relative group">
                <input
                  v-model="securityAnswer"
                  type="text"
                  :required="registerMode === 'username'"
                  maxlength="100"
                  class="w-full bg-white dark:bg-black/20 border border-amber-200 dark:border-amber-500/30 text-slate-900 dark:text-slate-100 px-2.5 py-2 sm:py-2.5 rounded-xl focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold"
                  placeholder="密保答案"
                />
              </div>
            </div>

            <!-- 同意协议 -->
            <div class="flex items-start gap-2.5 px-2 transition-all">
              <label class="relative flex items-center cursor-pointer group/check">
                <input
                  type="checkbox"
                  v-model="agreementAccepted"
                  class="sr-only peer"
                />
                <div class="w-4 h-4 rounded border-2 border-slate-200 dark:border-white/10 peer-checked:bg-blue-600 peer-checked:border-blue-600 transition-all flex items-center justify-center">
                  <div class="w-2.5 h-2.5 opacity-0 peer-checked:opacity-100 transition-opacity">
                    <ShieldCheck class="w-full h-full text-white" />
                  </div>
                </div>
              </label>
              <div class="text-[10px] sm:text-xs font-medium text-slate-500 dark:text-slate-400 leading-tight">
                我已阅读、理解并完全同意
                <button type="button" @click="openAgreement('user')" class="text-blue-600 hover:underline">《用户协议》</button>
                与
                <button type="button" @click="openAgreement('privacy')" class="text-blue-600 hover:underline">《隐私政策》</button>
              </div>
            </div>

            <button
              type="submit"
              :disabled="loading"
              class="w-full h-9 sm:h-10 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-400 text-white rounded-xl sm:rounded-2xl font-black transition-all shadow-lg shadow-blue-500/25 touch-feedback flex items-center justify-center gap-2 text-xs sm:text-sm"
            >
              <template v-if="loading">
                <Loader2 class="w-4 h-4 animate-spin" />
                正在签署...
              </template>
              <template v-else>
                同意协议并注册
              </template>
            </button>

            <div v-if="hasOAuth">
              <div class="relative flex items-center py-1">
                <div class="flex-grow border-t border-slate-100 dark:border-white/5"></div>
                <span class="flex-shrink mx-2.5 text-[10px] sm:text-xs font-black text-slate-400 dark:text-slate-600 uppercase tracking-widest">OR</span>
                <div class="flex-grow border-t border-slate-100 dark:border-white/5"></div>
              </div>

              <div class="grid grid-cols-2 gap-2">
                <button
                  v-if="githubEnabled"
                  type="button"
                  @click="handleOAuthLogin('github')"
                  :disabled="loading"
                  class="h-9 sm:h-10 bg-slate-50 dark:bg-black/40 hover:bg-white dark:hover:bg-black/60 text-slate-600 dark:text-slate-400 font-bold rounded-lg sm:rounded-xl touch-feedback transition-all text-[10px] sm:text-xs uppercase tracking-widest flex items-center justify-center gap-1.5 border border-slate-200 dark:border-white/5 hover:border-cyan-500/50"
                >
                  <OAuthLogos provider="github" :size="14" class="text-slate-800 dark:text-white" />
                  GitHub
                </button>
                <button
                  v-if="msEnabled"
                  type="button"
                  @click="handleOAuthLogin('ms')"
                  :disabled="loading"
                  class="h-9 sm:h-10 bg-slate-50 dark:bg-black/40 hover:bg-white dark:hover:bg-black/60 text-slate-600 dark:text-slate-400 font-bold rounded-lg sm:rounded-xl touch-feedback transition-all text-[10px] sm:text-xs uppercase tracking-widest flex items-center justify-center gap-1.5 border border-slate-200 dark:border-white/5 hover:border-blue-500/50"
                >
                  <OAuthLogos provider="microsoft" :size="14" />
                  Microsoft
                </button>
                <button
                  v-if="googleEnabled"
                  type="button"
                  @click="handleOAuthLogin('google')"
                  :disabled="loading"
                  class="h-9 sm:h-10 bg-slate-50 dark:bg-black/40 hover:bg-white dark:hover:bg-black/60 text-slate-600 dark:text-slate-400 font-bold rounded-lg sm:rounded-xl touch-feedback transition-all text-[10px] sm:text-xs uppercase tracking-widest flex items-center justify-center gap-1.5 border border-slate-200 dark:border-white/5 hover:border-red-500/50"
                >
                  <OAuthLogos provider="google" :size="14" />
                  Google
                </button>
                <button
                  v-if="appleEnabled"
                  type="button"
                  @click="handleOAuthLogin('apple')"
                  :disabled="loading"
                  class="h-9 sm:h-10 bg-slate-50 dark:bg-black/40 hover:bg-white dark:hover:bg-black/60 text-slate-600 dark:text-slate-400 font-bold rounded-lg sm:rounded-xl touch-feedback transition-all text-[10px] sm:text-xs uppercase tracking-widest flex items-center justify-center gap-1.5 border border-slate-200 dark:border-white/5 hover:border-slate-600/50"
                >
                  <OAuthLogos provider="apple" :size="14" class="text-slate-800 dark:text-white" />
                  Apple ID
                </button>
              </div>
            </div>
          </form>

          <div class="mt-4 sm:mt-5 pt-4 sm:pt-5 border-t border-slate-100 dark:border-white/5 text-center">
            <p class="text-[10px] sm:text-xs font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">
              已有合法的研究员凭证？
              <router-link to="/login" class="text-blue-600 hover:text-blue-500">返回登录</router-link>
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- 协议模态框 -->
    <Teleport to="body">
    <transition name="fade">
      <div v-if="showAgreementModal" class="viewport-modal-overlay z-[200] p-4 bg-slate-900/60 dark:bg-black/70">
        <div class="absolute inset-0 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md" @click="showAgreementModal = false"></div>
        <div class="relative w-full max-w-2xl bg-white dark:bg-[#121216] border border-slate-200 dark:border-white/10 rounded-[32px] shadow-2xl overflow-hidden flex flex-col max-h-[85vh] animate-in zoom-in duration-300">
           <div class="px-6 py-5 border-b border-slate-100 dark:border-white/5 flex items-center justify-between sticky top-0 bg-white dark:bg-[#121216] z-10">
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 bg-blue-500/10 border border-blue-500/20 rounded-2xl flex items-center justify-center text-blue-500">
                  <FileText class="w-5 h-5" />
                </div>
                <div>
                  <h2 class="text-lg font-black text-slate-900 dark:text-white tracking-tight leading-none">
                    {{ agreementType === 'user' ? '研究员合规协议' : '实验室隐私保护准则' }}
                  </h2>
                  <p class="text-[9px] text-slate-400 dark:text-slate-500 font-mono uppercase tracking-widest mt-1.5">Legal_Documentation_System</p>
                </div>
              </div>
              <button @click="showAgreementModal = false" class="p-2.5 hover:bg-slate-100 dark:hover:bg-white/5 rounded-2xl transition-all text-slate-400 hover:text-slate-900 dark:hover:text-white active:scale-90">
                <X class="w-5 h-5" />
              </button>
           </div>
           
           <div class="flex-1 overflow-y-auto p-6 sm:p-8 custom-scrollbar">
              <iframe 
                :src="agreementType === 'user' ? '/USER_AGREEMENT.md' : '/PRIVACY_POLICY.md'"
                class="w-full h-[60vh] border-none dark:invert dark:opacity-80"
                frameborder="0"
              ></iframe>
           </div>

           <div class="px-6 py-5 border-t border-slate-100 dark:border-white/5 flex justify-end gap-3 bg-slate-50/50 dark:bg-white/[0.02]">
              <button 
                @click="showAgreementModal = false" 
                class="px-6 py-2.5 bg-blue-600 hover:bg-blue-500 text-white font-black rounded-2xl transition-all uppercase tracking-widest text-[10px] shadow-lg shadow-blue-500/20 active:scale-95"
              >
                已充分理解并确认
              </button>
           </div>
        </div>
      </div>
    </transition>
    </Teleport>
  </div>
</template>
