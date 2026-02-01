<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import { Lock, User, FlaskConical, ShieldCheck, Zap, Loader2 } from 'lucide-vue-next'

const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const error = ref('')
const loading = ref(false)
const router = useRouter()
const { showAlert } = useDialog()

const handleSubmit = async () => {
  error.value = ''

  if (password.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return
  }

  loading.value = true

  try {
    await authAPI.register(username.value, password.value)
    await showAlert('注册成功，请使用新凭据登录。', '研究员注册成功')
    router.push('/login')
  } catch (err: any) {
    error.value = err.response?.data?.error || '注册失败，用户名可能已存在'
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
        <div class="p-10 md:p-12">
          <div class="flex flex-col items-center mb-10">
            <div class="w-20 h-20 bg-blue-600 rounded-3xl flex items-center justify-center mb-4 shadow-lg transform -rotate-3 hover:rotate-0 transition-transform duration-500">
              <FlaskConical class="w-10 h-10 text-white" />
            </div>
            <h1 class="text-3xl font-black text-slate-900 dark:text-white tracking-tighter">
              加入<span class="text-blue-600">实验室</span>
            </h1>
            <p class="text-slate-500 dark:text-slate-400 text-sm mt-2 font-medium">创建您的研究员账户</p>
          </div>

          <form @submit.prevent="handleSubmit" class="space-y-6">
            <div v-if="error" class="flex items-center gap-2 p-4 bg-red-50 dark:bg-red-500/10 border border-red-100 dark:border-red-500/20 text-red-600 text-sm rounded-2xl animate-shake">
              <div class="w-2 h-2 rounded-full bg-red-400"></div>
              {{ error }}
            </div>

            <div class="space-y-4">
              <div class="relative group">
                <div class="absolute left-5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                  <User :size="20" :stroke-width="2.5" />
                </div>
                <input
                  v-model="username"
                  type="text"
                  required
                  class="w-full pl-14 pr-6 py-5 bg-slate-100/50 dark:bg-black/40 border-2 border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-black/60 rounded-2xl text-slate-900 dark:text-slate-100 placeholder:text-slate-500/70 font-bold outline-none transition-all"
                  placeholder="用户名"
                />
              </div>

              <div class="relative group">
                <div class="absolute left-5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                  <Lock :size="20" :stroke-width="2.5" />
                </div>
                <input
                  v-model="password"
                  type="password"
                  required
                  class="w-full pl-14 pr-6 py-5 bg-slate-100/50 dark:bg-black/40 border-2 border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-black/60 rounded-2xl text-slate-900 dark:text-slate-100 placeholder:text-slate-500/70 font-bold outline-none transition-all"
                  placeholder="密 码"
                />
              </div>

              <div class="relative group">
                <div class="absolute left-5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-500 transition-colors">
                  <ShieldCheck :size="20" :stroke-width="2.5" />
                </div>
                <input
                  v-model="confirmPassword"
                  type="password"
                  required
                  class="w-full pl-14 pr-6 py-5 bg-slate-100/50 dark:bg-black/40 border-2 border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-black/60 rounded-2xl text-slate-900 dark:text-slate-100 placeholder:text-slate-500/70 font-bold outline-none transition-all"
                  placeholder="确认密码"
                />
              </div>
            </div>

            <button
              type="submit"
              :disabled="loading"
              class="w-full py-5 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-400 text-white rounded-2xl font-black text-lg shadow-[0_15px_30px_rgba(37,99,235,0.2)] dark:shadow-[0_15px_30px_rgba(37,99,235,0.3)] hover:shadow-[0_20px_40px_rgba(37,99,235,0.4)] transition-all flex items-center justify-center gap-3 transform active:scale-[0.98]"
            >
              <template v-if="loading">
                <Loader2 class="w-6 h-6 animate-spin" />
              </template>
              <template v-else>
                <Zap class="w-5 h-5 fill-current" />
                立即注册
              </template>
            </button>
          </form>

          <div class="mt-10 text-center">
            <p class="text-slate-500 dark:text-slate-400 font-medium">
              已有账户？
              <router-link to="/login" class="text-blue-600 font-black hover:underline cursor-pointer">
                登录系统
              </router-link>
            </p>
          </div>
        </div>
        
        <div class="bg-slate-50 dark:bg-black/20 p-6 flex justify-around border-t border-slate-100 dark:border-white/5">
          <div class="flex flex-col items-center">
            <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Protocol</span>
            <span class="text-xs font-bold text-slate-600 dark:text-slate-400">Secure SHA-256</span>
          </div>
          <div class="w-px h-8 bg-slate-200 dark:bg-white/10"></div>
          <div class="flex flex-col items-center">
            <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Database</span>
            <span class="text-xs font-bold text-slate-600 dark:text-slate-400">Chemistry DB</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
