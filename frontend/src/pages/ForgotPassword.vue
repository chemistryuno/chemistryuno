<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../utils/api'
import { Beaker, Lock, User, Loader2, Fingerprint, ArrowLeft, ShieldCheck } from 'lucide-vue-next'

const username = ref('')
const code = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const error = ref('')
const message = ref('')
const loading = ref(false)
const router = useRouter()

const handleReset = async () => {
  error.value = ''
  message.value = ''
  
  if (newPassword.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return
  }

  loading.value = true
  try {
    const response = await authAPI.resetPasswordBy2FA({
      username: username.value,
      code: code.value,
      new_password: newPassword.value
    })
    message.value = response.data.message
    setTimeout(() => {
      router.push('/login')
    }, 2000)
  } catch (err: any) {
    error.value = err.response?.data?.error || '重置失败，请核对信息'
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
            <div class="w-16 h-16 bg-blue-600 rounded-2xl flex items-center justify-center mb-4 shadow-lg transform rotate-3">
              <ShieldCheck class="w-8 h-8 text-white" />
            </div>
            <h1 class="text-2xl font-black text-slate-900 dark:text-slate-100 tracking-tighter">
              凭证<span class="text-blue-600">重置协议</span>
            </h1>
            <p class="text-slate-400 dark:text-slate-500 text-[10px] font-black uppercase tracking-[0.2em] mt-2 font-mono text-center">AUTHENTICATION RECOVERY VIA 2FA</p>
          </div>

          <div v-if="error" class="bg-red-50 dark:bg-red-500/10 border border-red-100 dark:border-red-500/20 text-red-500 px-4 py-3 rounded-2xl mb-6 text-center text-xs font-bold animate-shake">
            {{ error }}
          </div>
          
          <div v-if="message" class="bg-emerald-50 dark:bg-emerald-500/10 border border-emerald-100 dark:border-emerald-500/20 text-emerald-500 px-4 py-3 rounded-2xl mb-6 text-center text-xs font-bold">
            {{ message }}
          </div>

          <form @submit.prevent="handleReset" class="space-y-5">
            <div class="space-y-1.5">
              <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest ml-1">用户名</label>
              <div class="relative group">
                <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                  <User class="w-4 h-4" />
                </div>
                <input
                  v-model="username"
                  type="text"
                  required
                  class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-11 pr-4 py-4 rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-500/50 text-sm font-bold"
                  placeholder="请输入您的研究员账号"
                />
              </div>
            </div>

            <div class="space-y-1.5">
              <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest ml-1">2FA 令牌</label>
              <div class="relative group">
                <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                  <Fingerprint class="w-4 h-4" />
                </div>
                <input
                  v-model="code"
                  type="text"
                  maxlength="6"
                  required
                  class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-11 pr-4 py-4 rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-500/50 text-sm font-bold tracking-[0.2em]"
                  placeholder="000000"
                />
              </div>
            </div>

            <div class="space-y-1.5">
              <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest ml-1">新访问秘钥</label>
              <div class="relative group">
                <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                  <Lock class="w-4 h-4" />
                </div>
                <input
                  v-model="newPassword"
                  type="password"
                  required
                  class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-11 pr-4 py-4 rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-500/50 text-sm font-bold"
                  placeholder="请输入新密码"
                />
              </div>
            </div>

            <div class="space-y-1.5">
              <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest ml-1">确认新秘钥</label>
              <div class="relative group">
                <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                  <Lock class="w-4 h-4" />
                </div>
                <input
                  v-model="confirmPassword"
                  type="password"
                  required
                  class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-11 pr-4 py-4 rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-500/50 text-sm font-bold"
                  placeholder="请再次输入新密码"
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
                同步至数据库...
              </template>
              <template v-else>
                执行凭证重置
              </template>
            </button>

            <button
              type="button"
              @click="router.push('/login')"
              class="w-full h-14 bg-slate-100 dark:bg-white/5 text-slate-600 dark:text-slate-400 rounded-2xl font-black transition-all flex items-center justify-center gap-2 hover:bg-slate-200 dark:hover:bg-white/10"
            >
              <ArrowLeft class="w-4 h-4" />
              放弃并返回
            </button>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>
