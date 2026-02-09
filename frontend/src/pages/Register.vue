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
const countdown = ref(0)
const router = useRouter()
const dialog = useDialog()
const { showAlert } = dialog

onMounted(async () => {
  try {
    const res = await authAPI.getAuthConfig()
    smtpEnabled.value = res.data.smtp_enabled

  } catch (err) {
    console.error('获取配置失败', err)
  }
})

const handleLoginSuccess = (token: string, user: any) => {
  localStorage.setItem('token', token)
  localStorage.setItem('user', JSON.stringify(user))
  websocket.connect()
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

  const messageHandler = (event: MessageEvent) => {
    if (event.data.type === 'oauth-success') {
      window.removeEventListener('message', messageHandler)
      const { token, user } = event.data
      handleLoginSuccess(token, user)
      loading.value = false
    } else if (event.data.type === 'oauth-error') {
      window.removeEventListener('message', messageHandler)
      error.value = event.data.error || '授权失败'
      loading.value = false
    }
  }
  
  window.addEventListener('message', messageHandler)
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
  <div class="min-h-screen flex items-center justify-center p-4 bg-slate-50 dark:bg-black relative overflow-hidden font-sans">
    <!-- 实验性动态背景 -->
    <div class="absolute inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] bg-cyan-500/10 rounded-full blur-[120px] animate-pulse"></div>
      <div class="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-blue-500/10 rounded-full blur-[120px] animate-pulse" style="animation-delay: 2s"></div>
      <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-full h-full opacity-[0.03] dark:opacity-[0.05]" 
           style="background-image: radial-gradient(#4f46e5 1px, transparent 1px); background-size: 40px 40px;"></div>
    </div>

    <div class="w-full max-w-lg relative z-10 animate-in fade-in zoom-in duration-500">
      <div class="bg-white/80 dark:bg-slate-900/80 backdrop-blur-2xl rounded-2xl sm:rounded-3xl shadow-2xl border border-white dark:border-white/10 overflow-hidden">
        <div class="p-3 sm:p-4 md:p-6">
          <div class="flex flex-col items-center mb-4 sm:mb-5">
            <div class="relative group">
              <div class="absolute -inset-4 bg-gradient-to-tr from-cyan-500 to-blue-600 rounded-full blur opacity-25 group-hover:opacity-50 transition duration-1000"></div>
              <div class="relative w-12 h-12 sm:w-14 sm:h-14 bg-gradient-to-tr from-cyan-600 to-blue-700 rounded-xl sm:rounded-2xl flex items-center justify-center mb-2 sm:mb-3 shadow-xl transform rotate-3 group-hover:rotate-0 transition-transform duration-500">
                <FlaskConical class="w-6 h-6 sm:w-7 sm:h-7 text-white animate-bounce-slow" />
              </div>
            </div>
            <h1 class="text-xl sm:text-2xl md:text-3xl font-black text-slate-900 dark:text-white tracking-tighter">
              研究员<span class="text-transparent bg-clip-text bg-gradient-to-r from-cyan-500 to-blue-600">入职</span>
            </h1>
            <p class="text-slate-500 dark:text-slate-400 text-[10px] sm:text-xs mt-1 sm:mt-1.5 font-bold uppercase tracking-[0.2em]">创建实验室访问凭证</p>
          </div>

          <form @submit.prevent="handleSubmit" class="space-y-2.5 sm:space-y-3">
            <div v-if="error" class="flex items-center gap-2 p-2.5 sm:p-3 bg-red-500/10 border border-red-500/20 text-red-500 text-xs sm:text-sm rounded-xl sm:rounded-2xl animate-shake font-bold">
              <div class="w-2 h-2 rounded-full bg-red-500 animate-ping"></div>
              {{ error }}
            </div>

            <div v-if="!smtpEnabled" class="relative group">
              <div class="absolute left-3 sm:left-4 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-cyan-500 transition-colors">
                <User :size="16" class="sm:w-4 sm:h-4" />
              </div>
              <input
                v-model="username"
                type="text"
                required
                class="w-full pl-10 sm:pl-11 pr-3 sm:pr-4 py-3 sm:py-3.5 bg-slate-100/50 dark:bg-black/40 border-2 border-transparent focus:border-cyan-500 focus:bg-white dark:focus:bg-black/60 rounded-xl sm:rounded-2xl text-slate-900 dark:text-slate-100 placeholder:text-slate-500/50 text-xs sm:text-sm font-bold outline-none transition-all"
                placeholder="设定登录名 (Username)"
              />
            </div>

            <div v-else class="space-y-2.5 sm:space-y-3">
              <div class="relative group">
                <div class="absolute left-3 sm:left-4 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-cyan-500 transition-colors">
                  <Mail :size="16" class="sm:w-4 sm:h-4" />
                </div>
                <input
                  v-model="email"
                  type="email"
                  required
                  class="w-full pl-10 sm:pl-11 pr-3 sm:pr-4 py-3 sm:py-3.5 bg-slate-100/50 dark:bg-black/40 border-2 border-transparent focus:border-cyan-500 focus:bg-white dark:focus:bg-black/60 rounded-xl sm:rounded-2xl text-slate-900 dark:text-slate-100 placeholder:text-slate-500/50 text-xs sm:text-sm font-bold outline-none transition-all"
                  placeholder="实验室联系邮箱 (Email)"
                />
              </div>

              <div class="relative group flex gap-2">
                <div class="relative flex-1">
                  <div class="absolute left-3 sm:left-4 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-cyan-500 transition-colors">
                    <ShieldCheck :size="16" class="sm:w-4 sm:h-4" />
                  </div>
                  <input
                    v-model="code"
                    type="text"
                    required
                    class="w-full pl-10 sm:pl-11 pr-3 sm:pr-4 py-3 sm:py-3.5 bg-slate-100/50 dark:bg-black/40 border-2 border-transparent focus:border-cyan-500 focus:bg-white dark:focus:bg-black/60 rounded-xl sm:rounded-2xl text-slate-900 dark:text-slate-100 placeholder:text-slate-500/50 text-xs sm:text-sm font-bold outline-none transition-all"
                    placeholder="通讯校验码"
                  />
                </div>
                <button
                  type="button"
                  @click="handleSendCode"
                  :disabled="codeLoading || countdown > 0"
                  class="px-3 sm:px-4 rounded-xl sm:rounded-2xl font-black text-[9px] sm:text-[10px] uppercase tracking-widest transition-all bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 hover:bg-cyan-500/20 disabled:opacity-50 border border-cyan-500/20 flex items-center gap-1.5"
                >
                  {{ countdown > 0 ? `${countdown}S` : '发送指令' }}
                </button>
              </div>
            </div>

            <div class="relative group">
              <div class="absolute left-3 sm:left-4 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-cyan-500 transition-colors">
                <FlaskConical :size="16" class="sm:w-4 sm:h-4" />
              </div>
              <input
                v-model="nickname"
                type="text"
                required
                class="w-full pl-10 sm:pl-11 pr-3 sm:pr-4 py-3 sm:py-3.5 bg-slate-100/50 dark:bg-black/40 border-2 border-transparent focus:border-cyan-500 focus:bg-white dark:focus:bg-black/60 rounded-xl sm:rounded-2xl text-slate-900 dark:text-slate-100 placeholder:text-slate-500/50 text-xs sm:text-sm font-bold outline-none transition-all"
                placeholder="显示昵称 (Researcher Name)"
              />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-2.5 sm:gap-3">
              <div class="relative group">
                <div class="absolute left-3 sm:left-4 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-blue-500 transition-colors">
                  <Lock :size="16" class="sm:w-4 sm:h-4" />
                </div>
                <input
                  v-model="password"
                  type="password"
                  required
                  class="w-full pl-10 sm:pl-11 pr-3 sm:pr-4 py-3 sm:py-3.5 bg-slate-100/50 dark:bg-black/40 border-2 border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-black/60 rounded-xl sm:rounded-2xl text-slate-900 dark:text-slate-100 placeholder:text-slate-500/50 text-xs sm:text-sm font-bold outline-none transition-all"
                  placeholder="入职密钥"
                />
              </div>

              <div class="relative group">
                <div class="absolute left-3 sm:left-4 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-blue-500 transition-colors">
                  <ShieldCheck :size="16" class="sm:w-4 sm:h-4" />
                </div>
                <input
                  v-model="confirmPassword"
                  type="password"
                  required
                  class="w-full pl-10 sm:pl-11 pr-3 sm:pr-4 py-3 sm:py-3.5 bg-slate-100/50 dark:bg-black/40 border-2 border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-black/60 rounded-xl sm:rounded-2xl text-slate-900 dark:text-slate-100 placeholder:text-slate-500/50 text-xs sm:text-sm font-bold outline-none transition-all"
                  placeholder="确认密钥"
                />
              </div>
            </div>

            <button
              type="submit"
              :disabled="loading"
              class="w-full h-10 sm:h-12 bg-gradient-to-r from-cyan-600 to-blue-700 hover:from-cyan-500 hover:to-blue-600 disabled:from-slate-600 disabled:to-slate-700 text-white rounded-xl sm:rounded-2xl font-black text-xs sm:text-sm shadow-xl shadow-cyan-500/20 transition-all flex items-center justify-center gap-2 touch-feedback mt-1.5"
            >
              <template v-if="loading">
                <Loader2 class="w-4 h-4 sm:w-5 sm:h-5 animate-spin" />
                协议签署中...
              </template>
              <template v-else>
                签署入职协议
              </template>
            </button>

            <div class="relative flex items-center py-1.5 sm:py-2">
              <div class="flex-grow border-t border-slate-100 dark:border-white/5"></div>
              <span class="flex-shrink mx-2.5 sm:mx-3 text-[10px] sm:text-xs font-black text-slate-400 dark:text-slate-600 uppercase tracking-[0.25em]">其他接入方式</span>
              <div class="flex-grow border-t border-slate-100 dark:border-white/5"></div>
            </div>

            <div class="grid grid-cols-2 gap-2">
              <button
                type="button"
                @click="handleOAuthLogin('github')"
                :disabled="loading"
                class="h-9 sm:h-10 bg-slate-50 dark:bg-black/40 hover:bg-white dark:hover:bg-black/60 text-slate-600 dark:text-slate-400 font-bold rounded-lg sm:rounded-xl touch-feedback transition-all text-[10px] sm:text-xs uppercase tracking-widest flex items-center justify-center gap-1.5 border border-slate-200 dark:border-white/5 hover:border-cyan-500/50"
              >
                <OAuthLogos provider="github" :size="14" class="text-slate-800 dark:text-white" />
                GitHub
              </button>
              <button
                type="button"
                @click="handleOAuthLogin('ms')"
                :disabled="loading"
                class="h-9 sm:h-10 bg-slate-50 dark:bg-black/40 hover:bg-white dark:hover:bg-black/60 text-slate-600 dark:text-slate-400 font-bold rounded-lg sm:rounded-xl touch-feedback transition-all text-[10px] sm:text-xs uppercase tracking-widest flex items-center justify-center gap-1.5 border border-slate-200 dark:border-white/5 hover:border-blue-500/50"
              >
                <OAuthLogos provider="microsoft" :size="14" />
                Microsoft
              </button>
            </div>
          </form>

          <div class="mt-4 sm:mt-5 text-center pt-4 sm:pt-5 border-t border-slate-100 dark:border-white/5">
            <p class="text-slate-500 dark:text-slate-500 text-[10px] sm:text-xs font-bold uppercase tracking-widest">
              已有合法的研究员凭证？
              <router-link to="/login" class="text-cyan-500 hover:text-cyan-400 transition-colors ml-2 border-b border-cyan-500/30">
                立即返回登录
              </router-link>
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
