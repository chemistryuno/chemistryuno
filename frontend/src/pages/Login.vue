<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../utils/api'
import { Beaker, Lock, User, Loader2, Fingerprint } from 'lucide-vue-next'
import { cn } from '../utils/cn'
import websocket from '../utils/websocket'

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const router = useRouter()

const handleSubmit = async () => {
  error.value = ''
  loading.value = true

  try {
    const response = await authAPI.login(username.value, password.value)
    const { token, user } = response.data
    
    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(user))
    
    // 登录成功后建立 WebSocket 连接
    websocket.connect()
    
    router.push('/')
  } catch (err: any) {
    error.value = err.response?.data?.error || '身份验证失败，请重试'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4 bg-[#1a1a1e] relative overflow-hidden font-sans">
    <!-- Subtle Background Elements -->
    <div class="absolute top-[-10%] right-[-10%] w-[40%] h-[40%] bg-blue-500/5 rounded-full blur-[120px]"></div>
    <div class="absolute bottom-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-500/5 rounded-full blur-[120px]"></div>

    <div class="w-full max-w-md relative z-10 animate-in fade-in zoom-in duration-500">
      <div class="glass-panel-light rounded-[40px] shadow-[0_20px_60px_rgba(0,0,0,0.3)] overflow-hidden">
        <div class="p-10 md:p-12">
          <!-- Header Section -->
          <div class="flex flex-col items-center mb-10">
            <div class="w-20 h-20 bg-blue-600 rounded-3xl flex items-center justify-center mb-4 shadow-lg transform rotate-3 hover:rotate-0 transition-transform duration-500">
              <Beaker class="w-10 h-10 text-white" />
            </div>
            <h1 class="text-3xl font-black text-slate-800 tracking-tighter">
              化学<span class="text-blue-600">UNO</span>
            </h1>
            <p class="text-slate-400 text-xs font-bold uppercase tracking-[0.2em] mt-2">Laboratory System Access</p>
          </div>

          <div v-if="error" class="bg-red-50 border border-red-100 text-red-500 px-4 py-3 rounded-2xl mb-6 text-center text-xs font-bold">
            {{ error }}
          </div>

          <form @submit.prevent="handleSubmit" class="space-y-5">
            <div class="space-y-1.5">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">识别码 / Username</label>
              <div class="relative">
                <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400">
                  <User class="w-4 h-4" />
                </div>
                <input
                  v-model="username"
                  type="text"
                  required
                  class="w-full bg-slate-100 border border-slate-200 text-slate-800 pl-11 pr-4 py-3.5 rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-400 text-sm font-medium"
                  placeholder="Researcher ID"
                />
              </div>
            </div>

            <div class="space-y-1.5">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">访问秘钥 / Password</label>
              <div class="relative">
                <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400">
                  <Lock class="w-4 h-4" />
                </div>
                <input
                  v-model="password"
                  type="password"
                  required
                  class="w-full bg-slate-100 border border-slate-200 text-slate-800 pl-11 pr-4 py-3.5 rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-400 text-sm font-medium"
                  placeholder="Auth Token"
                />
              </div>
            </div>

            <button
              type="submit"
              :disabled="loading"
              :class="cn(
                'w-full h-14 rounded-2xl font-black text-white transition-all shadow-lg active:scale-95 flex items-center justify-center gap-2',
                loading 
                  ? 'bg-slate-400 cursor-not-allowed' 
                  : 'bg-blue-700 hover:bg-blue-600 shadow-blue-500/20'
              )"
            >
              <template v-if="loading">
                <Loader2 class="w-5 h-5 animate-spin" />
              </template>
              <template v-else>
                <span class="uppercase tracking-widest text-sm">初始化访问</span>
                <Fingerprint class="w-4 h-4" />
              </template>
            </button>
          </form>

          <div class="mt-8 text-center">
            <p class="text-slate-400 text-xs font-bold">
              初次参与实验？
              <router-link to="/register" class="text-blue-600 hover:text-blue-700 transition-colors">
                注册研究员账号
              </router-link>
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
