<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import { Lock, FlaskConical, ShieldCheck, Loader2, Mail, User, HelpCircle } from 'lucide-vue-next'
import OAuthLogos from '../components/icons/OAuthLogos.vue'
import websocket from '../utils/websocket'

const username = ref('')
const email = ref('')
const nickname = ref('')
const password = ref('')
const confirmPassword = ref('')
const code = ref('')
const securityQuestion = ref('')
const securityAnswer = ref('')
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

// 用户名只允许英文字母、数字、下划线
const usernameRegex = /^[a-zA-Z0-9_]+$/

onMounted(async () => {
  try {
    const res = await authAPI.getAuthConfig()
    smtpEnabled.value = res.data.smtp_enabled
    githubEnabled.value = res.data.github_enabled
    msEnabled.value = res.data.ms_enabled
    googleEnabled.value = res.data.google_enabled
    appleEnabled.value = res.data.apple_enabled
  } catch (err) {
    console.error('获取配置失败', err)
  }
})

const handleLoginSuccess = (token: string, user: any) => {
  localStorage.setItem('token', token)
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

  const url = `${import.meta.env.VITE_API_BASE_URL || '/api'}/auth/${provider}/login`
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

  // 验证用户名
  if (!username.value || username.value.length < 3) {
    error.value = '用户名至少需要3个字符'
    return
  }
  if (!usernameRegex.test(username.value)) {
    error.value = '用户名只能包含英文字母、数字和下划线'
    return
  }

  if (password.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return
  }

  if (password.value.length < 6) {
    error.value = '密码长度至少需要6位'
    return
  }

  if (!securityQuestion.value.trim()) {
    error.value = '请设置密保问题'
    return
  }

  if (!securityAnswer.value.trim()) {
    error.value = '请设置密保答案'
    return
  }

  // 如果填写了邮箱且SMTP开启，需要验证码
  if (email.value && smtpEnabled.value && !code.value) {
    error.value = '请输入邮箱验证码（或清空邮箱以跳过）'
    return
  }

  loading.value = true

  try {
    await authAPI.register({
      username: username.value,
      email: email.value || undefined,
      code: (email.value && smtpEnabled.value) ? code.value : undefined,
      nickname: nickname.value || username.value,
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
    } else if (backendError.includes('nickname already taken')) {
      error.value = '该昵称已被占用，请换一个'
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
  <div class="min-h-screen flex items-center justify-center p-4 bg-slate-50 dark:bg-[#1a1a1e] relative overflow-hidden font-sans">
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

            <!-- 邮箱（可选） -->
            <div class="space-y-2.5 sm:space-y-3">
              <div class="relative group">
                <div class="absolute left-0 pl-2.5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 group-focus-within:text-cyan-500 transition-colors pointer-events-none">
                  <Mail class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
                </div>
                <input
                  v-model="email"
                  type="email"
                  autocomplete="email"
                  class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-9 pr-2.5 py-2.5 sm:py-3 rounded-xl sm:rounded-2xl focus:ring-2 focus:ring-cyan-500/20 focus:border-cyan-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold"
                  placeholder="联系邮箱（可选，可在注册后绑定）"
                />
              </div>

              <div v-if="email && smtpEnabled" class="relative group flex gap-2">
                <div class="relative flex-1 group">
                  <div class="absolute left-0 pl-2.5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 group-focus-within:text-cyan-500 transition-colors pointer-events-none">
                    <ShieldCheck class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
                  </div>
                  <input
                    v-model="code"
                    type="text"
                    class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-9 pr-2.5 py-2.5 sm:py-3 rounded-xl sm:rounded-2xl focus:ring-2 focus:ring-cyan-500/20 focus:border-cyan-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold"
                    placeholder="邮箱验证码"
                  />
                </div>
                <button
                  type="button"
                  @click="handleSendCode"
                  :disabled="codeLoading || countdown > 0"
                  class="px-3 sm:px-4 rounded-xl font-black text-[10px] uppercase tracking-widest transition-all bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 hover:bg-cyan-500/20 disabled:opacity-50 border border-cyan-500/20 min-w-[70px]"
                >
                  {{ countdown > 0 ? `${countdown}S` : '发送' }}
                </button>
              </div>
            </div>

            <!-- 密码 -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-2.5 sm:gap-3">
              <div class="relative group">
                <div class="absolute left-0 pl-2.5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors pointer-events-none">
                  <Lock class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
                </div>
                <input
                  v-model="password"
                  type="password"
                  required
                  autocomplete="new-password"
                  class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-9 pr-2.5 py-2.5 sm:py-3 rounded-xl sm:rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold"
                  placeholder="入职密钥"
                />
              </div>

              <div class="relative group">
                <div class="absolute left-0 pl-2.5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors pointer-events-none">
                  <ShieldCheck class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
                </div>
                <input
                  v-model="confirmPassword"
                  type="password"
                  required
                  autocomplete="new-password"
                  class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-9 pr-2.5 py-2.5 sm:py-3 rounded-xl sm:rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold"
                  placeholder="确认密钥"
                />
              </div>
            </div>

            <!-- 密保设置 -->
            <div class="border border-amber-200/50 dark:border-amber-500/20 rounded-xl sm:rounded-2xl p-2.5 sm:p-3 space-y-2 bg-amber-50/50 dark:bg-amber-500/5">
              <p class="text-[10px] sm:text-xs font-black text-amber-600 dark:text-amber-400 uppercase tracking-widest flex items-center gap-1.5">
                <HelpCircle class="w-3.5 h-3.5" />
                密保问题（用于无邮箱时验证身份）
              </p>
              <div class="relative group">
                <input
                  v-model="securityQuestion"
                  type="text"
                  required
                  maxlength="200"
                  class="w-full bg-white dark:bg-black/20 border border-amber-200 dark:border-amber-500/30 text-slate-900 dark:text-slate-100 px-2.5 py-2 sm:py-2.5 rounded-xl focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold"
                  placeholder="自定义密保问题（如：我的第一只宠物叫什么？）"
                />
              </div>
              <div class="relative group">
                <input
                  v-model="securityAnswer"
                  type="text"
                  required
                  maxlength="100"
                  class="w-full bg-white dark:bg-black/20 border border-amber-200 dark:border-amber-500/30 text-slate-900 dark:text-slate-100 px-2.5 py-2 sm:py-2.5 rounded-xl focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold"
                  placeholder="密保答案"
                />
              </div>
            </div>

            <button
              type="submit"
              :disabled="loading"
              class="w-full h-9 sm:h-10 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-400 text-white rounded-xl sm:rounded-2xl font-black transition-all shadow-lg shadow-blue-500/25 touch-feedback flex items-center justify-center gap-2 text-xs sm:text-sm"
            >
              <template v-if="loading">
                <Loader2 class="w-4 h-4 animate-spin" />
                协议签署中...
              </template>
              <template v-else>
                签署入职协议
              </template>
            </button>

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
  </div>
</template>
