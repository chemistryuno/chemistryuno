<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import { Mail, Key, Lock, ShieldCheck, ArrowLeft, Loader2, FlaskConical, Zap } from 'lucide-vue-next'

const email = ref('')
const code = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const step = ref<'email' | 'reset'>('email')
const loading = ref(false)
const codeLoading = ref(false)
const countdown = ref(0)
const error = ref('')
const router = useRouter()
const { showAlert } = useDialog()

const startCountdown = () => {
  countdown.value = 60
  const timer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) clearInterval(timer)
  }, 1000)
}

const sendCode = async () => {
  if (!email.value || !email.value.includes('@')) {
    error.value = '请输入有效的电子邮箱地址'
    return
  }
  
  codeLoading.value = true
  try {
    await authAPI.sendCode(email.value, 'reset')
    startCountdown()
    step.value = 'reset'
    await showAlert('验证码已发送，请查收您的邮箱。', '验证码已发送')
  } catch (err: any) {
    error.value = err.response?.data?.error || '发送验证码失败'
  } finally {
    codeLoading.value = false
  }
}

const handleReset = async () => {
  if (newPassword.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return
  }
  
  loading.value = true
  try {
    await authAPI.resetPassword({ 
      email: email.value, 
      code: code.value, 
      password: newPassword.value 
    })
    await showAlert('密码重置成功，请使用新密码登录。', '成功')
    router.push('/login')
  } catch (err: any) {
    error.value = err.response?.data?.error || '重置失败，请检查验证码'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4 bg-slate-50 dark:bg-[#1a1a1e] relative overflow-hidden font-sans">
    <div class="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-500/5 rounded-full blur-[120px]"></div>
    <div class="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-blue-500/5 rounded-full blur-[120px]"></div>

    <div class="w-full max-w-md relative z-10 animate-in fade-in zoom-in duration-500">
      <div class="glass-panel-light rounded-[40px] shadow-[0_20px_60px_rgba(0,0,0,0.1)] dark:shadow-[0_20px_60px_rgba(0,0,0,0.3)] overflow-hidden">
        <div class="p-8 md:p-10">
          <div class="flex flex-col items-center mb-8">
            <div class="w-16 h-16 bg-blue-600 rounded-2xl flex items-center justify-center mb-4 shadow-lg transform rotate-3">
              <FlaskConical class="w-8 h-8 text-white" />
            </div>
            <h1 class="text-3xl font-black text-slate-900 dark:text-white tracking-tighter">找回<span class="text-blue-600">凭证</span></h1>
            <p class="text-slate-500 dark:text-slate-400 text-sm mt-1 font-medium italic">Recover Laboratory Access</p>
          </div>

          <div v-if="error" class="bg-red-50 dark:bg-red-500/10 border border-red-100 dark:border-red-500/20 text-red-600 p-4 rounded-2xl mb-6 text-sm font-bold flex items-center gap-2">
            <div class="w-1.5 h-1.5 rounded-full bg-red-500 shrink-0"></div>
            {{ error }}
          </div>

          <div v-if="step === 'email'" class="space-y-6">
            <div class="space-y-1.5">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">注册邮箱</label>
              <div class="relative group">
                <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 group-focus-within:text-blue-500 transition-colors">
                  <Mail class="w-4 h-4" />
                </div>
                <input
                  v-model="email"
                  type="email"
                  required
                  class="w-full bg-slate-100/50 dark:bg-black/20 border-2 border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-black/40 rounded-2xl pl-11 pr-4 py-4 text-sm font-bold outline-none transition-all"
                  placeholder="your@email.com"
                />
              </div>
            </div>

            <button
              @click="sendCode"
              :disabled="codeLoading"
              class="w-full py-4 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-400 text-white rounded-2xl font-black shadow-lg shadow-blue-500/20 transition-all flex items-center justify-center gap-2"
            >
              <template v-if="codeLoading">
                <Loader2 class="w-5 h-5 animate-spin" />
                正在核实身份...
              </template>
              <template v-else>
                发送验证码到邮箱
                <Zap class="w-4 h-4 ml-1" />
              </template>
            </button>
          </div>

          <div v-else class="space-y-6">
            <div class="space-y-1.5">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">重置验证码</label>
              <div class="relative group">
                <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 group-focus-within:text-blue-500 transition-colors">
                  <Key class="w-4 h-4" />
                </div>
                <input
                  v-model="code"
                  type="text"
                  required
                  class="w-full bg-slate-100/50 dark:bg-black/20 border-2 border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-black/40 rounded-2xl pl-11 pr-4 py-4 text-sm font-bold outline-none transition-all tracking-widest"
                  placeholder="请输入核验码"
                />
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div class="space-y-1.5">
                <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">新通行密钥</label>
                <div class="relative group">
                  <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 group-focus-within:text-blue-500 transition-colors">
                    <Lock class="w-4 h-4" />
                  </div>
                  <input
                    v-model="newPassword"
                    type="password"
                    required
                    class="w-full bg-slate-100/50 dark:bg-black/20 border-2 border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-black/40 rounded-2xl pl-11 pr-4 py-4 text-sm font-bold outline-none transition-all"
                    placeholder=""
                  />
                </div>
              </div>

              <div class="space-y-1.5">
                <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">确认新密钥</label>
                <div class="relative group">
                  <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 group-focus-within:text-blue-500 transition-colors">
                    <ShieldCheck class="w-4 h-4" />
                  </div>
                  <input
                    v-model="confirmPassword"
                    type="password"
                    required
                    class="w-full bg-slate-100/50 dark:bg-black/20 border-2 border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-black/40 rounded-2xl pl-11 pr-4 py-4 text-sm font-bold outline-none transition-all"
                    placeholder=""
                  />
                </div>
              </div>
            </div>

            <button
              @click="handleReset"
              :disabled="loading"
              class="w-full py-4 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-400 text-white rounded-2xl font-black shadow-lg shadow-blue-500/20 transition-all flex items-center justify-center gap-2"
            >
              <template v-if="loading">
                <Loader2 class="w-5 h-5 animate-spin" />
                正在更新权限...
              </template>
              <template v-else>
                立即重置密钥
              </template>
            </button>

            <button
              @click="step = 'email'"
              class="w-full flex items-center justify-center gap-2 text-xs font-black text-slate-400 hover:text-blue-600 transition-colors border-2 border-dashed border-slate-200 dark:border-white/10 rounded-2xl py-3"
            >
              <ArrowLeft class="w-3 h-3" />
              回退到邮箱输入
            </button>
          </div>

          <div class="mt-8 text-center pt-8 border-t border-slate-100 dark:border-white/5">
            <p class="text-slate-400 dark:text-slate-500 text-[10px] font-black uppercase tracking-widest">
              回想起通行密钥了？
              <router-link to="/login" class="text-blue-600 hover:text-blue-500 transition-colors">
                返回实验室入口
              </router-link>
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
