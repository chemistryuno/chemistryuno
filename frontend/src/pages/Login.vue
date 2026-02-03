<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import api, { authAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import { Beaker, Lock, User, Loader2, Fingerprint, Shield, Cpu } from 'lucide-vue-next'
import ResetPassword2FAModal from '../components/ResetPassword2FAModal.vue'
import websocket from '../utils/websocket'
import { get } from '@github/webauthn-json'

const identifier = ref(localStorage.getItem('last_username') || '')
const password = ref('')

const twoFactorCode = ref('')
const show2FA = ref(false)
const showResetModal = ref(false)
const resetLoading = ref(false)
const tempUID = ref<number | null>(null)
const error = ref('')
const loading = ref(false)
const router = useRouter()
const dialog = useDialog()

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

  // 处理登录时的公告
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

  router.push('/')
}

const handleWebAuthnLogin = async () => {
  if (!identifier.value) {
    error.value = '请输入用户名以调起硬件密钥'
    return
  }
  error.value = ''
  loading.value = true
  
  // 保存用户名
  localStorage.setItem('last_username', identifier.value)

  try {
    const res = await api.get(`/auth/webauthn/login/begin?username=${identifier.value}`)
    const credential = await get(res.data)
    const resFinish = await api.post(`/auth/webauthn/login/finish?username=${identifier.value}`, credential)
    
    // WebAuthn 登录返回的数据结构已统一
    const { token, user, announcements } = resFinish.data
    handleLoginSuccess(token, user, announcements)
  } catch (err: any) {
    console.error('WebAuthn login error:', err)
    error.value = err.response?.data?.error || '硬件密钥验证取消或失败'
  } finally {
    loading.value = false
  }
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
            <form @submit.prevent="handleSubmit" class="space-y-5">
              <div class="space-y-1.5">
                <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest ml-1">
                  账号
                </label>
                <div class="relative group">
                  <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                    <User class="w-4 h-4" />
                  </div>
                  <input
                    v-model="identifier"
                    type="text"
                    required
                    class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-11 pr-4 py-4 rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-500/50 text-sm font-bold"
                    placeholder="请输入用户名"
                  />
                </div>
              </div>

              <div class="space-y-1.5">
                <div class="flex justify-between items-center px-1">
                  <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">访问秘钥</label>
                  <button 
                    type="button"
                    @click="handleForgotPassword"
                    class="text-[10px] font-black text-blue-500 hover:text-blue-600 uppercase tracking-widest transition-colors cursor-pointer"
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
                    type="password"
                    required
                    class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-11 pr-4 py-4 rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-500/50 text-sm font-bold"
                    placeholder="请输入密码"
                  />
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

              <div class="relative flex items-center py-2">
                <div class="flex-grow border-t border-slate-100 dark:border-white/5"></div>
                <span class="flex-shrink mx-4 text-[10px] font-black text-slate-400 dark:text-slate-600 uppercase tracking-widest">OR</span>
                <div class="flex-grow border-t border-slate-100 dark:border-white/5"></div>
              </div>

              <button
                type="button"
                @click="handleWebAuthnLogin"
                :disabled="loading"
                class="w-full h-14 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-600 dark:text-slate-300 font-bold rounded-2xl active:scale-95 transition-all text-xs uppercase tracking-widest flex items-center justify-center gap-2 group border border-slate-200 dark:border-white/5"
              >
                <Cpu class="w-4 h-4 text-blue-500" />
                使用物理研究密钥登录
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

    <!-- 2FA 重置模态框 -->
    <ResetPassword2FAModal 
      :show="showResetModal"
      :loading="resetLoading"
      @close="showResetModal = false"
      @submit="handleResetSubmit"
    />
  </div>
</template>
