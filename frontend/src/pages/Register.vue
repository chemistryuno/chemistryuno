<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import { Lock, User, FlaskConical, ShieldCheck, Zap, Loader2, Key, Mail, Send } from 'lucide-vue-next'
import OAuthLogos from '../components/icons/OAuthLogos.vue'
import websocket from '../utils/websocket'

const username = ref('')
const email = ref('')
const nickname = ref('')
const password = ref('')
const confirmPassword = ref('')
const code = ref('')
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
      error.value = event.data.error || '授权失败'
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
    
    // 倒计时
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

  if (password.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return
  }

  if (smtpEnabled.value && !code.value) {
    error.value = '请输入邮箱验证码'
    return
  }

  loading.value = true

  try {
    await authAPI.register({
      username: smtpEnabled.value ? undefined : username.value,
      email: smtpEnabled.value ? email.value : undefined,
      code: smtpEnabled.value ? code.value : undefined,
      nickname: nickname.value,
      password: password.value,
    })
    await showAlert('注册成功，请使用新凭据登录。', '研究员注册成功')
    router.push('/login')
  } catch (err: any) {
    error.value = err.response?.data?.error || '注册失败'
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

            <div v-if="!smtpEnabled" class="relative group">
              <div class="absolute left-0 pl-2.5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 group-focus-within:text-cyan-500 transition-colors pointer-events-none">
                <User class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
              </div>
              <input
                v-model="username"
                type="text"
                required
                class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-9 pr-2.5 py-2.5 sm:py-3 rounded-xl sm:rounded-2xl focus:ring-2 focus:ring-cyan-500/20 focus:border-cyan-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold"
                placeholder="设定登录名 (Username)"
              />
            </div>

            <div v-else class="space-y-2.5 sm:space-y-3">
              <div class="relative group">
                <div class="absolute left-0 pl-2.5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 group-focus-within:text-cyan-500 transition-colors pointer-events-none">
                  <Mail class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
                </div>
                <input
                  v-model="email"
                  type="email"
                  required
                  class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-9 pr-2.5 py-2.5 sm:py-3 rounded-xl sm:rounded-2xl focus:ring-2 focus:ring-cyan-500/20 focus:border-cyan-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold"
                  placeholder="实验室联系邮箱 (Email)"
                />
              </div>

              <div class="relative group flex gap-2">
                <div class="relative flex-1 group">
                  <div class="absolute left-0 pl-2.5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 group-focus-within:text-cyan-500 transition-colors pointer-events-none">
                    <ShieldCheck class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
                  </div>
                  <input
                    v-model="code"
                    type="text"
                    required
                    class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-9 pr-2.5 py-2.5 sm:py-3 rounded-xl sm:rounded-2xl focus:ring-2 focus:ring-cyan-500/20 focus:border-cyan-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold"
                    placeholder="通讯校验码"
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

            <div class="relative group">
              <div class="absolute left-0 pl-2.5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 group-focus-within:text-cyan-500 transition-colors pointer-events-none">
                <FlaskConical class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
              </div>
              <input
                v-model="nickname"
                type="text"
                required
                class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-9 pr-2.5 py-2.5 sm:py-3 rounded-xl sm:rounded-2xl focus:ring-2 focus:ring-cyan-500/20 focus:border-cyan-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold"
                placeholder="显示昵称 (Researcher Name)"
              />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-2.5 sm:gap-3">
              <div class="relative group">
                <div class="absolute left-0 pl-2.5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors pointer-events-none">
                  <Lock class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
                </div>
                <input
                  v-model="password"
                  type="password"
                  required
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
                  class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-9 pr-2.5 py-2.5 sm:py-3 rounded-xl sm:rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold"
                  placeholder="确认密钥"
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
