<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../utils/api'
import { Beaker, Lock, User, Loader2, Fingerprint, Mail, Key, Shield } from 'lucide-vue-next'
import { cn } from '../utils/cn'
import websocket from '../utils/websocket'

const identifier = ref('') // Can be username or email
const password = ref('')
const loginMode = ref<'password' | 'code'>('password')
const verificationCode = ref('')
const codeSent = ref(false)
const countdown = ref(0)
const codeLoading = ref(false)

const twoFactorCode = ref('')
const show2FA = ref(false)
const tempUID = ref<number | null>(null)
const error = ref('')
const loading = ref(false)
const router = useRouter()

const startCountdown = () => {
  countdown.value = 60
  const timer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) clearInterval(timer)
  }, 1000)
}

const sendCode = async () => {
  if (!identifier.value || !identifier.value.includes('@')) {
    error.value = '邮箱验证码登录需要输入电子邮箱地址'
    return
  }
  
  codeLoading.value = true
  try {
    await authAPI.sendCode(identifier.value, 'login')
    codeSent.value = true
    startCountdown()
  } catch (err: any) {
    error.value = err.response?.data?.error || '发送验证码失败'
  } finally {
    codeLoading.value = false
  }
}

const handleSubmit = async () => {
  error.value = ''
  loading.value = true

  try {
    let response;
    if (loginMode.value === 'password') {
      response = await authAPI.login({
        username: identifier.value,
        password: password.value,
        method: 'password'
      })
    } else {
      if (!verificationCode.value) {
        error.value = '请输入验证码'
        loading.value = false
        return
      }
      response = await authAPI.login({
        username: identifier.value,
        code: verificationCode.value,
        method: 'code'
      })
    }
    
    if (response.data.two_factor_required) {
      show2FA.value = true
      tempUID.value = response.data.uid
      loading.value = false
      return
    }

    const { token, user } = response.data
    handleLoginSuccess(token, user)
  } catch (err: any) {
    error.value = err.response?.data?.error || '身份验证失败，请核对凭据'
  } finally {
    loading.value = false
  }
}

const handle2FAVerify = async () => {
  if (!tempUID.value) return
  error.value = ''
  loading.value = true

  try {
    const response = await authAPI.verify2FALogin(tempUID.value, twoFactorCode.value)
    const { token, user } = response.data
    handleLoginSuccess(token, user)
  } catch (err: any) {
    error.value = err.response?.data?.error || '2FA验证失败'
  } finally {
    loading.value = false
  }
}

const handleLoginSuccess = (token: string, user: any) => {
  localStorage.setItem('token', token)
  localStorage.setItem('user', JSON.stringify(user))
  websocket.connect()
  router.push('/')
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4 bg-slate-50 dark:bg-[#1a1a1e] relative overflow-hidden font-sans">
    <div class="absolute top-[-10%] right-[-10%] w-[40%] h-[40%] bg-blue-500/5 rounded-full blur-[120px]"></div>
    <div class="absolute bottom-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-500/5 rounded-full blur-[120px]"></div>

    <div class="w-full max-w-md relative z-10 animate-in fade-in zoom-in duration-500">
      <div class="glass-panel-light rounded-[40px] shadow-[0_20px_60px_rgba(0,0,0,0.1)] dark:shadow-[0_20px_60px_rgba(0,0,0,0.3)] overflow-hidden">
        <div class="p-8 md:p-10">
          <div class="flex flex-col items-center mb-8">
            <div class="w-16 h-16 bg-blue-600 rounded-2xl flex items-center justify-center mb-4 shadow-lg transform rotate-3 hover:rotate-0 transition-transform duration-500">
              <Beaker class="w-8 h-8 text-white" />
            </div>
            <h1 class="text-3xl font-black text-slate-900 dark:text-slate-100 tracking-tighter">
              化学<span class="text-blue-600">UNO</span>
            </h1>
            <p class="text-slate-400 dark:text-slate-500 text-[10px] font-black uppercase tracking-[0.2em] mt-2 font-mono">LABORATORY ACCESS</p>
          </div>

          <div v-if="error" class="bg-red-50 dark:bg-red-500/10 border border-red-100 dark:border-red-500/20 text-red-500 px-4 py-3 rounded-2xl mb-6 text-center text-xs font-bold animate-shake">
            {{ error }}
          </div>

          <div v-if="!show2FA" class="space-y-6">
            <div class="flex p-1 bg-slate-100 dark:bg-black/40 rounded-xl mb-2">
              <button 
                @click="loginMode = 'password'"
                :class="cn('flex-1 py-2 text-xs font-black rounded-lg transition-all', loginMode === 'password' ? 'bg-white dark:bg-slate-800 shadow-sm text-blue-600' : 'text-slate-500')"
              >密码登录</button>
              <button 
                @click="loginMode = 'code'"
                :class="cn('flex-1 py-2 text-xs font-black rounded-lg transition-all', loginMode === 'code' ? 'bg-white dark:bg-slate-800 shadow-sm text-blue-600' : 'text-slate-500')"
              >验证码登录</button>
            </div>

            <form @submit.prevent="handleSubmit" class="space-y-5">
              <div class="space-y-1.5">
                <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest ml-1">
                  {{ loginMode === 'password' ? '账号 (用户名/邮箱)' : '电子邮箱' }}
                </label>
                <div class="relative group">
                  <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                    <User v-if="loginMode === 'password'" class="w-4 h-4" />
                    <Mail v-else class="w-4 h-4" />
                  </div>
                  <input
                    v-model="identifier"
                    type="text"
                    required
                    class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-11 pr-4 py-4 rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-500/50 text-sm font-bold"
                    :placeholder="loginMode === 'password' ? '用户名 或 邮箱' : 'your@email.com'"
                  />
                </div>
              </div>

              <div v-if="loginMode === 'password'" class="space-y-1.5">
                <div class="flex justify-between items-center px-1">
                  <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">访问秘钥</label>
                  <router-link to="/forgot-password" class="text-[10px] font-black text-blue-600 hover:text-blue-500 uppercase tracking-widest">找回密码</router-link>
                </div>
                <div class="relative group">
                  <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                    <Lock class="w-4 h-4" />
                  </div>
                  <input
                    v-model="password"
                    type="password"
                    required
                    class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-11 pr-4 py-4 rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-500/50 text-sm font-bold"
                    placeholder="请输入密码"
                  />
                </div>
              </div>

              <div v-else class="space-y-1.5">
                <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest ml-1">验证码</label>
                <div class="flex gap-2">
                  <div class="relative flex-1 group">
                    <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                      <Key class="w-4 h-4" />
                    </div>
                    <input
                      v-model="verificationCode"
                      type="text"
                      required
                      class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-11 pr-4 py-4 rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-500/50 text-sm font-bold"
                      placeholder="6位验证码"
                    />
                  </div>
                  <button
                    type="button"
                    @click="sendCode"
                    :disabled="codeLoading || countdown > 0"
                    class="px-4 bg-blue-600/10 hover:bg-blue-600/20 text-blue-600 dark:text-blue-400 rounded-2xl text-[10px] font-black transition-all whitespace-nowrap disabled:opacity-50"
                  >
                    {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
                  </button>
                </div>
              </div>

              <button
                type="submit"
                :disabled="loading"
                class="w-full h-14 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-400 text-white rounded-2xl font-black transition-all shadow-lg shadow-blue-500/25 active:scale-95 flex items-center justify-center gap-2"
              >
                <template v-if="loading">
                  <Loader2 class="w-5 h-5 animate-spin" />
                  核验中...
                </template>
                <template v-else>
                  授权并进入
                </template>
              </button>
            </form>
          </div>

          <div v-else class="space-y-6 animate-in slide-in-from-bottom duration-500">
            <div class="text-center mb-2">
              <div class="w-12 h-12 bg-blue-600/10 flex items-center justify-center rounded-2xl mx-auto mb-3">
                <Shield class="w-6 h-6 text-blue-600" />
              </div>
              <h2 class="text-lg font-black text-slate-900 dark:text-slate-100">二次身份核验</h2>
              <p class="text-xs text-slate-500">此账号已开启安全协议，请输入 2FA 动态令牌</p>
            </div>
            
            <div class="relative group">
              <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 group-focus-within:text-blue-500">
                <Fingerprint class="w-4 h-4" />
              </div>
              <input
                v-model="twoFactorCode"
                type="text"
                maxlength="6"
                class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-11 pr-4 py-4 rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all text-center tracking-[0.5em] font-black text-lg"
                placeholder="000000"
              />
            </div>

            <button
              @click="handle2FAVerify"
              :disabled="loading"
              class="w-full h-14 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-400 text-white rounded-2xl font-black transition-all shadow-lg shadow-blue-500/25 flex items-center justify-center gap-2"
            >
              <template v-if="loading">
                <Loader2 class="w-5 h-5 animate-spin" />
              </template>
              完成核验
            </button>
          </div>

          <div class="mt-8 pt-8 border-t border-slate-100 dark:border-white/5 text-center">
            <p class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">
              还不是正式研究员？
              <router-link to="/register" class="text-blue-600 hover:text-blue-500">提交申请</router-link>
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
