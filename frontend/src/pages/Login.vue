<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api, { authAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import { Beaker, Lock, User, Loader2, Fingerprint, Shield, Cpu, Mail, Github, Globe, Chrome, Apple, Eye, EyeOff } from 'lucide-vue-next'
import ResetPassword2FAModal from '../components/ResetPassword2FAModal.vue'
import websocket from '../utils/websocket'
import { get } from '@github/webauthn-json'

const identifier = ref(localStorage.getItem('last_username') || '')
const password = ref('')
const showPassword = ref(false)

const twoFactorCode = ref('')
const show2FA = ref(false)
const showResetModal = ref(false)
const resetLoading = ref(false)
const tempUID = ref<number | null>(null)
const error = ref('')
const loading = ref(false)
const smtpEnabled = ref(false)
const router = useRouter()
const dialog = useDialog()

onMounted(async () => {
  try {
    const res = await authAPI.getAuthConfig()
    smtpEnabled.value = res.data.smtp_enabled
    
  } catch (err) {
    console.error('获取配置失败', err)
  }
})

const handleForgotPassword = () => {
  showResetModal.value = true
}

const handleResetSubmit = async (username: string, code: string, newPw: string) => {
  resetLoading.value = true
  try {
    await authAPI.resetPasswordBy2FA({
      username,
      code,
      new_password: newPw
    })
    showResetModal.value = false
    dialog.showAlert('实验凭证已成功找回并更新，请尝试重新授权登录。', '协议更新成功')
  } catch (err: any) {
    dialog.showAlert(err.response?.data?.error || '凭证验证失败，请核对用户名及动态验证码。', '协议冲突')
  } finally {
    resetLoading.value = false
  }
}

const handleSubmit = async () => {
  error.value = ''

  loading.value = true
  
  // 保存最后一次输入的用户名
  localStorage.setItem('last_username', identifier.value)

  try {
    const response = await authAPI.login({
      username: identifier.value,
      password: password.value,
    })
    
    if (response.data.two_factor_required) {
      show2FA.value = true
      tempUID.value = response.data.uid
      loading.value = false
      return
    }

    const { token, user, announcements } = response.data
    handleLoginSuccess(token, user, announcements)
  } catch (err: any) {
    error.value = err.response?.data?.error || '身份验证失败，请核对凭证'
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
    const { token, user, announcements } = response.data
    handleLoginSuccess(token, user, announcements)
  } catch (err: any) {
    error.value = err.response?.data?.error || '2FA验证失败'
  } finally {
    loading.value = false
  }
}

const handleLoginSuccess = (token: string, user: any, announcements: any[] = []) => {
  localStorage.setItem('token', token)
  localStorage.setItem('user', JSON.stringify(user))
  websocket.connect()

  // 处理登录时的公告 - 已禁用开屏提示
  // if (announcements && announcements.length > 0) {
  //   announcements.forEach((ann: any) => {
  //     // 只处理模态框类型的，跑马灯交给 AnnouncementTicker 自动获取
  //     if (!ann.is_ticker) {
  //       let title = ann.title || '系统公告'
  //       if (ann.type === 'emergency' && !ann.title) title = '紧急通知'
  //       if (ann.type === 'maintenance' && !ann.title) title = '维护通知'
  //       dialog.showAlert(ann.content, title, '确定', ann.close_delay || 0)
  //     }
  //   })
  // }

  router.push('/')
}

const handleWebAuthnLogin = async () => {
  error.value = ''
  loading.value = true
  try {
    // 1. 开始 WebAuthn 登录
    const beginRes = await authAPI.beginWebAuthnLogin(identifier.value)
    
    // 2. 调用浏览器 WebAuthn API
    const credential = await get(beginRes.data)
    
    // 3. 完成 WebAuthn 登录
    const finishRes = await authAPI.finishWebAuthnLogin(credential, identifier.value)
    
    const { token, user, announcements } = finishRes.data
    handleLoginSuccess(token, user, announcements)
    dialog.showAlert('已通过物理研究密钥验证身份，准许进入。', '授权成功')
  } catch (err: any) {
    console.error('WebAuthn 登录失败', err)
    if (err.name === 'NotAllowedError') {
      error.value = '硬件验证被取消'
    } else {
      error.value = err.response?.data?.error || '硬件验证失败，请确保您已绑定密钥'
    }
  } finally {
    loading.value = false
  }
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
    dialog.showAlert('弹出窗口被拦截，请允许弹出窗口后重试。', '拦截提示')
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

  // 轮询检查窗口是否关闭
  const timer = setInterval(() => {
    if (popup.closed) {
      clearInterval(timer)
      loading.value = false
      window.removeEventListener('message', messageHandler)
    }
  }, 1000)
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4 bg-slate-50 dark:bg-[#1a1a1e] relative overflow-hidden font-sans">
    <div class="absolute top-[-10%] right-[-10%] w-[40%] h-[40%] bg-blue-500/5 rounded-full blur-[120px]"></div>
    <div class="absolute bottom-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-500/5 rounded-full blur-[120px]"></div>

    <div class="w-full max-w-md relative z-10 animate-in fade-in zoom-in duration-500">
      <div class="glass-panel-light rounded-[32px] sm:rounded-[40px] shadow-[0_20px_60px_rgba(0,0,0,0.1)] dark:shadow-[0_20px_60px_rgba(0,0,0,0.3)] overflow-hidden">
        <div class="p-6 sm:p-8 md:p-10">
          <div class="flex flex-col items-center mb-6 sm:mb-8">
            <div class="w-14 h-14 sm:w-16 sm:h-16 bg-blue-600 rounded-2xl flex items-center justify-center mb-3 sm:mb-4 shadow-lg transform rotate-3 hover:rotate-0 transition-transform duration-500">
              <Beaker class="w-7 h-7 sm:w-8 sm:h-8 text-white" />
            </div>
            <h1 class="text-2xl sm:text-3xl font-black text-slate-900 dark:text-slate-100 tracking-tighter">
              化学<span class="text-blue-600">UNO</span>
            </h1>
            <p class="text-slate-400 dark:text-slate-500 text-xs-mobile font-black uppercase tracking-[0.2em] mt-2 font-mono">LABORATORY ACCESS</p>
          </div>

          <div v-if="error" class="bg-red-50 dark:bg-red-500/10 border border-red-100 dark:border-red-500/20 text-red-500 px-4 py-3 rounded-2xl mb-5 sm:mb-6 text-center text-xs font-bold animate-shake">
            {{ error }}
          </div>

          <div v-if="!show2FA" class="space-y-5 sm:space-y-6">
            <form @submit.prevent="handleSubmit" class="space-y-4 sm:space-y-5">
              <div class="space-y-1.5">
                  <label class="text-xs-mobile font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest ml-1">
                    {{ smtpEnabled ? '电子邮箱' : '账号' }}
                  </label>
                  <div class="relative group">
                    <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                      <component :is="smtpEnabled ? Mail : User" class="w-4 h-4" />
                    </div>
                    <input
                      v-model="identifier"
                      type="text"
                      required
                      class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-11 pr-4 py-4 rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-500/50 text-base font-bold"
                      :placeholder="smtpEnabled ? '注册时的邮箱' : '请输入用户名'"
                    />
                  </div>
              </div>

              <div class="space-y-1.5">
                <div class="flex justify-between items-center px-1">
                  <label class="text-xs-mobile font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">访问秘钥</label>
                  <button
                    type="button"
                    @click="handleForgotPassword"
                    class="text-xs-mobile font-black text-blue-500 hover:text-blue-600 uppercase tracking-widest transition-colors cursor-pointer touch-feedback"
                  >
                    找回凭证?
                  </button>
                </div>
                <div class="relative group">
                  <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                    <Lock class="w-4 h-4" />
                  </div>
                  <input
                    v-model="password"
                    :type="showPassword ? 'text' : 'password'"
                    required
                    class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-11 pr-12 py-4 rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-500/50 text-base font-bold font-mono"
                    placeholder="请输入访问凭证"
                  />
                  <button
                    type="button"
                    @click="showPassword = !showPassword"
                    class="absolute inset-y-0 right-0 pr-4 flex items-center text-slate-400 hover:text-blue-500 transition-colors touch-feedback"
                  >
                    <component :is="showPassword ? EyeOff : Eye" class="w-4 h-4" />
                  </button>
                </div>
              </div>

              <button
                type="submit"
                :disabled="loading"
                class="w-full h-12 sm:h-14 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-400 text-white rounded-2xl font-black transition-all shadow-lg shadow-blue-500/25 touch-feedback flex items-center justify-center gap-2"
              >
                <template v-if="loading">
                  <Loader2 class="w-5 h-5 animate-spin" />
                  核验中...
                </template>
                <template v-else>
                  授权并进入
                </template>
              </button>

              <div class="relative flex items-center py-2">
                <div class="flex-grow border-t border-slate-100 dark:border-white/5"></div>
                <span class="flex-shrink mx-4 text-xs-mobile font-black text-slate-400 dark:text-slate-600 uppercase tracking-widest">OR</span>
                <div class="flex-grow border-t border-slate-100 dark:border-white/5"></div>
              </div>

              <button
                type="button"
                @click="handleWebAuthnLogin"
                :disabled="loading"
                class="w-full h-12 sm:h-14 bg-blue-600/5 dark:bg-blue-600/10 hover:bg-blue-600/10 dark:hover:bg-blue-600/20 text-blue-700 dark:text-blue-400 font-black rounded-2xl touch-feedback transition-all text-xs-mobile uppercase tracking-[0.2em] flex items-center justify-center gap-3 group border border-blue-600/20 shadow-sm"
              >
                <Cpu class="w-4 h-4 text-blue-600 animate-pulse" />
                使用物理研究密钥登录
              </button>

              <div class="grid grid-cols-2 gap-3">
                <button
                  type="button"
                  @click="handleOAuthLogin('github')"
                  :disabled="loading"
                  class="h-11 sm:h-12 bg-slate-50 dark:bg-black/40 hover:bg-white dark:hover:bg-black/60 text-slate-600 dark:text-slate-400 font-bold rounded-xl touch-feedback transition-all text-xs-mobile uppercase tracking-widest flex items-center justify-center gap-2 border border-slate-200 dark:border-white/5 hover:border-blue-500/50 hover:text-blue-600 shadow-sm"
                >
                  <Github class="w-4 h-4" />
                  GitHub 授权
                </button>
                <button
                  type="button"
                  @click="handleOAuthLogin('ms')"
                  :disabled="loading"
                  class="h-11 sm:h-12 bg-slate-50 dark:bg-black/40 hover:bg-white dark:hover:bg-black/60 text-slate-600 dark:text-slate-400 font-bold rounded-xl touch-feedback transition-all text-xs-mobile uppercase tracking-widest flex items-center justify-center gap-2 border border-slate-200 dark:border-white/5 hover:border-blue-500/50 hover:text-blue-600 shadow-sm"
                >
                  <Globe class="w-4 h-4 text-sky-500" />
                  Microsoft
                </button>
                <button
                  type="button"
                  @click="handleOAuthLogin('google')"
                  :disabled="loading"
                  class="h-11 sm:h-12 bg-slate-50 dark:bg-black/40 hover:bg-white dark:hover:bg-black/60 text-slate-600 dark:text-slate-400 font-bold rounded-xl touch-feedback transition-all text-xs-mobile uppercase tracking-widest flex items-center justify-center gap-2 border border-slate-200 dark:border-white/5 hover:border-blue-500/50 hover:text-blue-600 shadow-sm"
                >
                  <Chrome class="w-4 h-4 text-rose-500" />
                  Google
                </button>
                <button
                  type="button"
                  @click="handleOAuthLogin('apple')"
                  :disabled="loading"
                  class="h-11 sm:h-12 bg-slate-50 dark:bg-black/40 hover:bg-white dark:hover:bg-black/60 text-slate-600 dark:text-slate-400 font-bold rounded-xl touch-feedback transition-all text-xs-mobile uppercase tracking-widest flex items-center justify-center gap-2 border border-slate-200 dark:border-white/5 hover:border-blue-500/50 hover:text-blue-600 shadow-sm"
                >
                  <Apple class="w-4 h-4" />
                  Apple ID
                </button>
              </div>
            </form>
          </div>

          <div v-else class="space-y-5 sm:space-y-6 animate-in slide-in-from-bottom duration-500">
            <div class="text-center mb-5 sm:mb-6">
              <div class="w-14 h-14 sm:w-16 sm:h-16 bg-blue-600/10 flex items-center justify-center rounded-3xl mx-auto mb-3 sm:mb-4 border border-blue-500/20 shadow-inner">
                <Shield class="w-7 h-7 sm:w-8 sm:h-8 text-blue-600 animate-pulse" />
              </div>
              <h2 class="text-lg sm:text-xl font-black text-slate-900 dark:text-slate-100 tracking-tight">二次身份核验</h2>
              <p class="text-xs-mobile text-slate-500 font-bold uppercase tracking-widest mt-1">AUTHORIZED PERSONNEL ONLY</p>
            </div>

            <div class="space-y-1.5">
              <label class="text-xs-mobile font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest ml-1">动态安全令牌</label>
              <div class="relative group">
                <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 group-focus-within:text-blue-500">
                  <Fingerprint class="w-4 h-4" />
                </div>
                <input
                  v-model="twoFactorCode"
                  type="text"
                  maxlength="6"
                  required
                  class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-11 pr-4 py-5 rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all text-center tracking-[0.6em] font-black text-2xl font-mono shadow-inner"
                  placeholder="------"
                  @keyup.enter="handle2FAVerify"
                />
              </div>
            </div>

            <button
              @click="handle2FAVerify"
              :disabled="loading"
              class="w-full h-12 sm:h-14 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-400 text-white rounded-2xl font-black transition-all shadow-lg shadow-blue-500/25 flex items-center justify-center gap-2 touch-feedback"
            >
              <template v-if="loading">
                <Loader2 class="w-5 h-5 animate-spin" />
                验证中...
              </template>
              <template v-else>
                完成核验并进入
              </template>
            </button>

            <button
              @click="show2FA = false"
              class="w-full text-xs-mobile font-black text-slate-400 hover:text-slate-600 uppercase tracking-widest transition-colors touch-feedback"
            >
              ← 返回基础授权
            </button>
          </div>

          <div class="mt-6 sm:mt-8 pt-6 sm:pt-8 border-t border-slate-100 dark:border-white/5 text-center">
            <p class="text-xs-mobile font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">
              还不是正式研究员？
              <router-link to="/register" class="text-blue-600 hover:text-blue-500">提交申请</router-link>
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- 2FA 重置模态框 -->
    <ResetPassword2FAModal 
      :show="showResetModal"
      :loading="resetLoading"
      @close="showResetModal = false"
      @submit="handleResetSubmit"
    />
  </div>
</template>
