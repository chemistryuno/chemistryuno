<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import { Lock, User, FlaskConical, ShieldCheck, Zap, Loader2, Key, Mail, Send } from 'lucide-vue-next'

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
const { showAlert } = useDialog()

onMounted(async () => {
  try {
    const res = await authAPI.getAuthConfig()
    smtpEnabled.value = res.data.smtp_enabled
  } catch (err) {
    console.error('获取配置失败', err)
  }
})

const handleSendCode = async () => {
  if (!email.value || !email.value.includes('@')) {
    error.value = '请输入有效的邮箱地址'
    return
  }

  codeLoading.value = true
  try {
    await authAPI.sendCode(email.value)
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
      password: password.value
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
    <div class="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-500/5 rounded-full blur-[120px]"></div>
    <div class="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-blue-500/5 rounded-full blur-[120px]"></div>

    <div class="w-full max-w-lg relative z-10 animate-in fade-in zoom-in duration-500">
      <div class="glass-panel-light rounded-[40px] shadow-[0_20px_60px_rgba(0,0,0,0.1)] dark:shadow-[0_20px_60px_rgba(0,0,0,0.3)] overflow-hidden">
        <div class="p-8 md:p-10">
          <div class="flex flex-col items-center mb-8">
            <div class="w-16 h-16 bg-blue-600 rounded-3xl flex items-center justify-center mb-4 shadow-lg transform -rotate-3 hover:rotate-0 transition-transform duration-500">
              <FlaskConical class="w-8 h-8 text-white" />
            </div>
            <h1 class="text-3xl font-black text-slate-900 dark:text-white tracking-tighter">
              加入<span class="text-blue-600">实验室</span>
            </h1>
            <p class="text-slate-500 dark:text-slate-400 text-sm mt-2 font-medium">创建您的研究员账户</p>
          </div>

          <form @submit.prevent="handleSubmit" class="space-y-4">
            <div v-if="error" class="flex items-center gap-2 p-4 bg-red-50 dark:bg-red-500/10 border border-red-100 dark:border-red-500/20 text-red-600 text-sm rounded-2xl animate-shake">
              <div class="w-2 h-2 rounded-full bg-red-400"></div>
              {{ error }}
            </div>

            <div v-if="!smtpEnabled" class="relative group">
              <div class="absolute left-5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                <User :size="18" :stroke-width="2.5" />
              </div>
              <input
                v-model="username"
                type="text"
                required
                class="w-full pl-12 pr-4 py-4 bg-slate-100/50 dark:bg-black/40 border-2 border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-black/60 rounded-2xl text-slate-900 dark:text-slate-100 placeholder:text-slate-500/70 font-bold outline-none transition-all text-sm"
                placeholder="用户名 (登录账号)"
              />
            </div>

            <div v-else class="space-y-4">
              <div class="relative group">
                <div class="absolute left-5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                  <Mail :size="18" :stroke-width="2.5" />
                </div>
                <input
                  v-model="email"
                  type="email"
                  required
                  class="w-full pl-12 pr-4 py-4 bg-slate-100/50 dark:bg-black/40 border-2 border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-black/60 rounded-2xl text-slate-900 dark:text-slate-100 placeholder:text-slate-500/70 font-bold outline-none transition-all text-sm"
                  placeholder="电子邮箱 (登录凭据)"
                />
              </div>

              <div class="relative group flex gap-3">
                <div class="relative flex-1 group">
                  <div class="absolute left-5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                    <ShieldCheck :size="18" :stroke-width="2.5" />
                  </div>
                  <input
                    v-model="code"
                    type="text"
                    required
                    class="w-full pl-12 pr-4 py-4 bg-slate-100/50 dark:bg-black/40 border-2 border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-black/60 rounded-2xl text-slate-900 dark:text-slate-100 placeholder:text-slate-500/70 font-bold outline-none transition-all text-sm"
                    placeholder="验证码"
                  />
                </div>
                <button
                  type="button"
                  @click="handleSendCode"
                  :disabled="codeLoading || countdown > 0"
                  class="px-6 rounded-2xl font-black text-xs uppercase tracking-widest transition-all bg-slate-200 dark:bg-white/5 text-slate-600 dark:text-slate-300 hover:bg-slate-300 dark:hover:bg-white/10 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                >
                  <Send v-if="!codeLoading" class="w-4 h-4" />
                  <Loader2 v-else class="w-4 h-4 animate-spin" />
                  {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
                </button>
              </div>
            </div>

            <div class="relative group">
              <div class="absolute left-5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                <FlaskConical :size="18" :stroke-width="2.5" />
              </div>
              <input
                v-model="nickname"
                type="text"
                required
                class="w-full pl-12 pr-4 py-4 bg-slate-100/50 dark:bg-black/40 border-2 border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-black/60 rounded-2xl text-slate-900 dark:text-slate-100 placeholder:text-slate-500/70 font-bold outline-none transition-all text-sm"
                placeholder="研究员昵称 (公开展示)"
              />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div class="relative group">
                <div class="absolute left-5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                  <Lock :size="18" :stroke-width="2.5" />
                </div>
                <input
                  v-model="password"
                  type="password"
                  required
                  class="w-full pl-12 pr-4 py-4 bg-slate-100/50 dark:bg-black/40 border-2 border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-black/60 rounded-2xl text-slate-900 dark:text-slate-100 placeholder:text-slate-500/70 font-bold outline-none transition-all text-sm"
                  placeholder="设定密码 (至少6位)"
                />
              </div>

              <div class="relative group">
                <div class="absolute left-5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                  <ShieldCheck :size="18" :stroke-width="2.5" />
                </div>
                <input
                  v-model="confirmPassword"
                  type="password"
                  required
                  class="w-full pl-12 pr-4 py-4 bg-slate-100/50 dark:bg-black/40 border-2 border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-black/60 rounded-2xl text-slate-900 dark:text-slate-100 placeholder:text-slate-500/70 font-bold outline-none transition-all text-sm"
                  placeholder="确认密码"
                />
              </div>
            </div>

            <button
              type="submit"
              :disabled="loading"
              class="w-full py-4 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-400 text-white rounded-2xl font-black text-lg shadow-[0_15px_30px_rgba(37,99,235,0.2)] dark:shadow-[0_15px_30px_rgba(37,99,235,0.3)] transition-all flex items-center justify-center gap-3 transform active:scale-[0.98] mt-4"
            >
              <template v-if="loading">
                <Loader2 class="w-6 h-6 animate-spin" />
              </template>
              <template v-else>
                <Zap class="w-5 h-5 fill-current" />
                注册研究员凭证
              </template>
            </button>
          </form>

          <div class="mt-8 text-center pt-8 border-t border-slate-100 dark:border-white/5">
            <p class="text-slate-400 dark:text-slate-500 text-[10px] font-black uppercase tracking-widest">
              已有研究员账号？
              <router-link to="/login" class="text-blue-600 hover:text-blue-500 transition-colors">
                立即登录实验室
              </router-link>
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
