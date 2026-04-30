<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api, { authAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import feedback from '../utils/feedback'
import { Beaker, Lock, Loader2, Fingerprint, Shield, Cpu, Mail, Eye, EyeOff } from 'lucide-vue-next'
import ResetPassword2FAModal from '../components/ResetPassword2FAModal.vue'
import FeedbackButton from '../components/FeedbackButton.vue'
import OAuthLogos from '../components/icons/OAuthLogos.vue'
import websocket from '../utils/websocket'
import { API_BASE_URL } from '../utils/runtimeConfig'
import { get } from '@github/webauthn-json'

const identifier = ref(localStorage.getItem('last_email') || '')
const password = ref('')
const showPassword = ref(false)

const twoFactorCode = ref('')
const show2FA = ref(false)
const showResetModal = ref(false)
const resetLoading = ref(false)
const tempUID = ref<number | null>(null)
const error = ref('')
const loading = ref(false)
const githubEnabled = ref(false)
const msEnabled = ref(false)
const googleEnabled = ref(false)
const appleEnabled = ref(false)
const router = useRouter()
const dialog = useDialog()

onMounted(async () => {
  try {
    const res = await authAPI.getAuthConfig()
    githubEnabled.value = res.data.github_enabled
    msEnabled.value = res.data.ms_enabled
    googleEnabled.value = res.data.google_enabled
    appleEnabled.value = res.data.apple_enabled

  } catch (err) {
    console.error('获取配置失败', err)
  }
})

const handleForgotPassword = () => {
  showResetModal.value = true
}

const handleResetSubmit = async (email: string, code: string, newPw: string) => {
  resetLoading.value = true
  try {
    await authAPI.resetPasswordBy2FA({
      email,
      code,
      new_password: newPw
    })
    showResetModal.value = false
    dialog.showAlert('实验凭证已成功找回并更新，请尝试重新授权登录。', '协议更新成功')
  } catch (err: any) {
    dialog.showAlert(err.response?.data?.error || '凭证验证失败，请核对邮箱及动态验证码。', '协议冲突')
  } finally {
    resetLoading.value = false
  }
}

const handleSubmit = async () => {
  error.value = ''

  loading.value = true
  
  // 保存最后一次输入的标识符
  localStorage.setItem('last_email', identifier.value)

  try {
    const response = await authAPI.login({
      identifier: identifier.value,
      password: password.value,
    })
    
    if (response.data.two_factor_required) {
      show2FA.value = true
      tempUID.value = response.data.uid
      loading.value = false
      return
    }

    const { token, access_token, refresh_token, user, announcements, is_returning_player, days_since_last_login } = response.data
    handleLoginSuccess(token, user, announcements, is_returning_player, days_since_last_login, access_token, refresh_token)
  } catch (err: any) {
    const data = err.response?.data
    if (data?.banned_until) {
      const until = new Date(data.banned_until).toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' })
      error.value = `${data.error}，封禁至 ${until}（UTC+8:00）`
    } else if (data?.frozen_until) {
      const until = new Date(data.frozen_until).toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' })
      error.value = `${data.error}，直到 ${until}（UTC+8:00）`
    } else {
      error.value = data?.error || '身份验证失败，请核对凭证'
    }
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
    const { token, access_token, refresh_token, user, announcements, is_returning_player, days_since_last_login } = response.data
    handleLoginSuccess(token, user, announcements, is_returning_player, days_since_last_login, access_token, refresh_token)
  } catch (err: any) {
    error.value = err.response?.data?.error || '2FA验证失败'
  } finally {
    loading.value = false
  }
}

const handleLoginSuccess = (token: string | null, user: any, announcements: any[] = [], isReturningPlayer: boolean = false, daysSinceLastLogin: number = 0, accessToken?: string, refreshToken?: string) => {
  // Token已由后端通过HttpOnly Cookie设置，前端不需要存储
  localStorage.setItem('user', JSON.stringify(user))
  websocket.connect()
  window.dispatchEvent(new Event('auth-changed'))

  // 处理回归玩家
  if (isReturningPlayer) {
    const hasSkippedLobbyTutorial = localStorage.getItem('chemistry-uno-lobby-tutorial-skipped') === 'true'
    if (!hasSkippedLobbyTutorial) {
      // 清除已完成的大厅教程标记，重新触发新手指引
      localStorage.removeItem('chemistry-uno-lobby-tutorial-completed')
      console.log(`[老玩家回归] 用户 ${user.nickname} 距离上次登录 ${daysSinceLastLogin} 天，自动触发新手指引`)

      // 显示欢迎回归消息
      dialog.showAlert(
        `欢迎老玩家${user.nickname}回归！您已经 ${daysSinceLastLogin} 天没有登录了，让我们重新熟悉一下游戏界面吧！`,
        '🎉 欢迎回归',
        '开始指引'
      )
    } else {
      console.log(`[老玩家回归] 用户 ${user.nickname} 曾跳过大厅教程，保持不自动弹出`)
      dialog.showAlert(
        `欢迎老玩家${user.nickname}回归！您已经 ${daysSinceLastLogin} 天没有登录了，祝您游戏愉快！`,
        '🎉 欢迎回归',
        '继续游戏'
      )
    }
  }

  //处理登录时的公告
  if (announcements && announcements.length > 0) {
    announcements.forEach((ann: any) => {
      // 只处理模态框类型的，跑马灯交给 AnnouncementTicker 自动获取
      if (!ann.is_ticker) {
        let title = ann.title || '系统公告'
        if (ann.type === 'emergency' && !ann.title) title = '紧急通知'
        if (ann.type === 'maintenance' && !ann.title) title = '维护通知'
        dialog.showAlert(ann.content, title, '确定', ann.close_delay || 0)
      }
    })
  }

  // 检查是否有重定向参数
  const redirect = router.currentRoute.value.query.redirect as string
  if (redirect) {
    // 跳转到原始访问的页面（包括查询参数）
    router.push(redirect)
  } else {
    router.push('/')
  }
}

const handleWebAuthnLogin = async () => {
  // 验证是否输入了用户名/邮箱
  if (!identifier.value || identifier.value.trim() === '') {
    error.value = '请先输入您的用户名或邮箱'
    return
  }

  error.value = ''
  loading.value = true

  try {
    console.log('开始 WebAuthn 登录流程，用户:', identifier.value)

    // 1. 开始 WebAuthn 登录
    const beginRes = await authAPI.beginWebAuthnLogin(identifier.value)
    console.log('WebAuthn 选项获取成功')

    // 2. 调用浏览器 WebAuthn API
    const credential = await get(beginRes.data)
    console.log('WebAuthn 凭证获取成功')

    // 3. 完成 WebAuthn 登录
    const finishRes = await authAPI.finishWebAuthnLogin(credential, identifier.value)
    console.log('WebAuthn 验证成功')

    const { token, access_token, refresh_token, user, announcements } = finishRes.data
    handleLoginSuccess(token, user, announcements, false, 0, access_token, refresh_token)
    dialog.showAlert('已通过物理研究密钥验证身份，准许进入。', '授权成功')
  } catch (err: any) {
    console.error('WebAuthn 登录失败:', err)

    // 检查是否是用户取消操作
    if (err.name === 'NotAllowedError') {
      error.value = '硬件验证已取消'
    }
    // 检查是否是浏览器不支持
    else if (err.name === 'NotSupportedError') {
      error.value = '您的浏览器不支持 WebAuthn，请使用 Chrome、Edge 或 Safari'
    }
    // 检查是否是超时
    else if (err.name === 'TimeoutError') {
      error.value = '硬件验证超时，请重试'
    }
    // 检查后端返回的错误
    else if (err.response?.data?.error) {
      error.value = err.response.data.error
    }
    // 其他错误
    else {
      error.value = '硬件验证失败，请确保您已绑定物理密钥'
      console.error('详细错误:', err.message || err)
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
  
  const url = `${API_BASE_URL}/auth/${provider}/login`
  const popup = window.open(url, 'OAuth Login', `width=${width},height=${height},left=${left},top=${top}`)
  
  if (!popup) {
    loading.value = false
    dialog.showAlert('弹出窗口被拦截，请允许弹出窗口后重试。', '拦截提示')
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
      const { token, access_token, refresh_token, user } = event.data
      handleLoginSuccess(token, user, [], false, 0, access_token, refresh_token)
      loading.value = false
    } else if (event.data.type === 'oauth-error') {
      oauthFinished = true
      cleanup()
      error.value = event.data.error || '授权失败'
      loading.value = false
    }
  }

  window.addEventListener('message', messageHandler)

  // 轮询检查窗口是否关闭
  const timer = setInterval(() => {
    if (popup.closed) {
      clearInterval(timer)
      // 留一点时间等待 postMessage 到达，避免关闭与消息发送竞态导致登录丢失
      setTimeout(() => {
        if (!oauthFinished) {
          cleanup()
          loading.value = false
        }
      }, 600)
    }
  }, 1000)
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4 bg-slate-50 dark:bg-[#1a1a1e] relative overflow-hidden font-sans">
    <div class="absolute top-[-10%] right-[-10%] w-[40%] h-[40%] bg-blue-500/5 rounded-full blur-[120px]"></div>
    <div class="absolute bottom-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-500/5 rounded-full blur-[120px]"></div>

    <div class="w-full max-w-sm relative z-10 animate-in fade-in zoom-in duration-500">
      <div class="glass-panel-light rounded-2xl sm:rounded-3xl shadow-[0_20px_60px_rgba(0,0,0,0.1)] dark:shadow-[0_20px_60px_rgba(0,0,0,0.3)] overflow-hidden">
        <div class="p-3 sm:p-4 md:p-6">
          <div class="flex flex-col items-center mb-3 sm:mb-4">
            <div class="w-10 h-10 sm:w-12 sm:h-12 bg-blue-600 rounded-xl sm:rounded-2xl flex items-center justify-center mb-1.5 sm:mb-2 shadow-lg transform rotate-3 hover:rotate-0 transition-transform duration-500">
              <Beaker class="w-5 h-5 sm:w-6 sm:h-6 text-white" />
            </div>
            <h1 class="text-lg sm:text-xl font-black text-slate-900 dark:text-slate-100 tracking-tighter">
              化学<span class="text-blue-600">UNO</span>
            </h1>
            <p class="text-slate-400 dark:text-slate-500 text-[10px] sm:text-xs font-black uppercase tracking-[0.2em] mt-0.5 font-mono">LABORATORY ACCESS</p>
          </div>

          <div v-if="error" class="bg-red-50 dark:bg-red-500/10 border border-red-100 dark:border-red-500/20 text-red-500 px-2.5 py-2 rounded-xl mb-2.5 sm:mb-3 text-center text-xs font-bold animate-shake relative">
            {{ error }}
            <div v-if="error.includes('禁止操作') || error.includes('封禁') || error.includes('冷冻')" class="mt-2 flex justify-center">
              <FeedbackButton />
            </div>
          </div>

          <div v-if="!show2FA" class="space-y-2.5 sm:space-y-3">
            <form @submit.prevent="handleSubmit" class="space-y-2 sm:space-y-2.5">
              <div class="space-y-0.5">
                  <label class="text-[10px] sm:text-xs font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest ml-1">
                    用户名 / 邮箱
                  </label>
                  <div class="relative group">
                    <div class="absolute inset-y-0 left-0 pl-2.5 flex items-center pointer-events-none text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                      <Mail class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
                    </div>
                    <input
                      v-model="identifier"
                      type="text"
                      required
                      autocomplete="username"
                      class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-9 pr-2.5 py-2.5 sm:py-3 rounded-xl sm:rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold"
                      placeholder="请输入用户名或邮箱"
                    />
                  </div>
              </div>

              <div class="space-y-0.5">
                <div class="flex justify-between items-center px-1">
                  <label class="text-[10px] sm:text-xs font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">访问秘钥</label>
                  <button
                    type="button"
                    @click="handleForgotPassword"
                    class="text-[10px] sm:text-xs font-black text-blue-500 hover:text-blue-600 uppercase tracking-widest transition-colors cursor-pointer touch-feedback"
                  >
                    找回凭证?
                  </button>
                </div>
                <div class="relative group">
                  <div class="absolute inset-y-0 left-0 pl-2.5 flex items-center pointer-events-none text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                    <Lock class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
                  </div>
                  <input
                    v-model="password"
                    :type="showPassword ? 'text' : 'password'"
                    required
                    class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-9 pr-10 py-2.5 sm:py-3 rounded-xl sm:rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-500/50 text-xs sm:text-sm font-bold font-mono"
                    placeholder="请输入访问凭证"
                  />
                  <button
                    type="button"
                    @click="showPassword = !showPassword"
                    class="absolute inset-y-0 right-0 pr-2.5 flex items-center text-slate-400 hover:text-blue-500 transition-colors touch-feedback"
                  >
                    <component :is="showPassword ? EyeOff : Eye" class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
                  </button>
                </div>
              </div>

              <button
                type="submit"
                :disabled="loading"
                class="w-full h-9 sm:h-10 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-400 text-white rounded-xl sm:rounded-2xl font-black transition-all shadow-lg shadow-blue-500/25 touch-feedback flex items-center justify-center gap-2 text-xs sm:text-sm"
              >
                <template v-if="loading">
                  <Loader2 class="w-4 h-4 animate-spin" />
                  核验中...
                </template>
                <template v-else>
                  授权并进入
                </template>
              </button>

              <div class="relative flex items-center py-1">
                <div class="flex-grow border-t border-slate-100 dark:border-white/5"></div>
                <span class="flex-shrink mx-2.5 text-[10px] sm:text-xs font-black text-slate-400 dark:text-slate-600 uppercase tracking-widest">OR</span>
                <div class="flex-grow border-t border-slate-100 dark:border-white/5"></div>
              </div>

              <button
                type="button"
                @click="handleWebAuthnLogin"
                :disabled="loading"
                class="w-full h-8 sm:h-9 bg-blue-600/5 dark:bg-blue-600/10 hover:bg-blue-600/10 dark:hover:bg-blue-600/20 text-blue-700 dark:text-blue-400 font-black rounded-lg sm:rounded-xl touch-feedback transition-all text-[10px] sm:text-xs uppercase tracking-[0.2em] flex items-center justify-center gap-1.5 sm:gap-2 group border border-blue-600/20 shadow-sm"
              >
                <Cpu class="w-3.5 h-3.5 text-blue-600 animate-pulse" />
                使用物理研究密钥登录
              </button>

              <div class="grid grid-cols-2 gap-2">
                <button
                  v-if="githubEnabled"
                  type="button"
                  @click="handleOAuthLogin('github')"
                  :disabled="loading"
                  class="h-8 sm:h-9 bg-slate-50 dark:bg-black/40 hover:bg-white dark:hover:bg-black/60 text-slate-600 dark:text-slate-400 font-bold rounded-lg sm:rounded-xl touch-feedback transition-all text-[10px] sm:text-xs uppercase tracking-widest flex items-center justify-center gap-1 sm:gap-1.5 border border-slate-200 dark:border-white/5 hover:border-blue-500/50 hover:text-blue-600 shadow-sm"
                >
                  <OAuthLogos provider="github" :size="14" class="text-slate-800 dark:text-white" />
                  GitHub 授权
                </button>
                <button
                  v-if="msEnabled"
                  type="button"
                  @click="handleOAuthLogin('ms')"
                  :disabled="loading"
                  class="h-8 sm:h-9 bg-slate-50 dark:bg-black/40 hover:bg-white dark:hover:bg-black/60 text-slate-600 dark:text-slate-400 font-bold rounded-lg sm:rounded-xl touch-feedback transition-all text-[10px] sm:text-xs uppercase tracking-widest flex items-center justify-center gap-1 sm:gap-1.5 border border-slate-200 dark:border-white/5 hover:border-blue-500/50 hover:text-blue-600 shadow-sm"
                >
                  <OAuthLogos provider="microsoft" :size="14" />
                  Microsoft
                </button>
                <button
                  v-if="googleEnabled"
                  type="button"
                  @click="handleOAuthLogin('google')"
                  :disabled="loading"
                  class="h-8 sm:h-9 bg-slate-50 dark:bg-black/40 hover:bg-white dark:hover:bg-black/60 text-slate-600 dark:text-slate-400 font-bold rounded-lg sm:rounded-xl touch-feedback transition-all text-[10px] sm:text-xs uppercase tracking-widest flex items-center justify-center gap-1 sm:gap-1.5 border border-slate-200 dark:border-white/5 hover:border-blue-500/50 hover:text-blue-600 shadow-sm"
                >
                  <OAuthLogos provider="google" :size="14" />
                  Google
                </button>
                <button
                  v-if="appleEnabled"
                  type="button"
                  @click="handleOAuthLogin('apple')"
                  :disabled="loading"
                  class="h-8 sm:h-9 bg-slate-50 dark:bg-black/40 hover:bg-white dark:hover:bg-black/60 text-slate-600 dark:text-slate-400 font-bold rounded-lg sm:rounded-xl touch-feedback transition-all text-[10px] sm:text-xs uppercase tracking-widest flex items-center justify-center gap-1 sm:gap-1.5 border border-slate-200 dark:border-white/5 hover:border-blue-500/50 hover:text-blue-600 shadow-sm"
                >
                  <OAuthLogos provider="apple" :size="14" class="text-slate-800 dark:text-white" />
                  Apple ID
                </button>
              </div>
            </form>
          </div>

          <div v-else class="space-y-2.5 sm:space-y-3 animate-in slide-in-from-bottom duration-500">
            <div class="text-center mb-3 sm:mb-4">
              <div class="w-10 h-10 sm:w-12 sm:h-12 bg-blue-600/10 flex items-center justify-center rounded-xl sm:rounded-2xl mx-auto mb-1.5 sm:mb-2 border border-blue-500/20 shadow-inner">
                <Shield class="w-5 h-5 sm:w-6 sm:h-6 text-blue-600 animate-pulse" />
              </div>
              <h2 class="text-base sm:text-lg font-black text-slate-900 dark:text-slate-100 tracking-tight">二次身份核验</h2>
              <p class="text-[10px] sm:text-xs text-slate-500 font-bold uppercase tracking-widest mt-0.5">AUTHORIZED PERSONNEL ONLY</p>
            </div>

            <div class="space-y-0.5">
              <label class="text-[10px] sm:text-xs font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest ml-1">动态安全令牌</label>
              <div class="relative group">
                <div class="absolute inset-y-0 left-0 pl-2.5 flex items-center pointer-events-none text-slate-400 group-focus-within:text-blue-500">
                  <Fingerprint class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
                </div>
                <input
                  v-model="twoFactorCode"
                  type="text"
                  maxlength="6"
                  required
                  class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-9 pr-2.5 py-3 sm:py-4 rounded-xl sm:rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all text-center tracking-[0.6em] font-black text-lg sm:text-xl font-mono shadow-inner"
                  placeholder="------"
                  @keyup.enter="handle2FAVerify"
                />
              </div>
            </div>

            <button
              @click="handle2FAVerify"
              :disabled="loading"
              class="w-full h-9 sm:h-10 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-400 text-white rounded-xl sm:rounded-2xl font-black transition-all shadow-lg shadow-blue-500/25 flex items-center justify-center gap-2 touch-feedback text-xs sm:text-sm"
            >
              <template v-if="loading">
                <Loader2 class="w-4 h-4 animate-spin" />
                验证中...
              </template>
              <template v-else>
                完成核验并进入
              </template>
            </button>

            <button
              @click="show2FA = false"
              class="w-full text-[10px] sm:text-xs font-black text-slate-400 hover:text-slate-600 uppercase tracking-widest transition-colors touch-feedback"
            >
              ← 返回基础授权
            </button>
          </div>

          <div class="mt-4 sm:mt-5 pt-4 sm:pt-5 border-t border-slate-100 dark:border-white/5 text-center">
            <p class="text-[10px] sm:text-xs font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest mb-2">
              还不是正式研究员？
              <router-link to="/register" class="text-blue-600 hover:text-blue-500">提交申请</router-link>
            </p>
            <div class="flex justify-center mt-2">
              <!-- 提供账号申诉/通用反馈入口 -->
              <FeedbackButton />
            </div>
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
