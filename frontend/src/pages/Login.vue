<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../utils/api'
import { Beaker, Lock, User, Loader2, Fingerprint, LogIn } from 'lucide-vue-next'
import { cn } from '../utils/cn'
import websocket from '../utils/websocket'

const username = ref('')
const password = ref('')
const twoFactorCode = ref('')
const twoFactorRequired = ref(false)
const error = ref('')
const loading = ref(false)
const router = useRouter()

onMounted(() => {
  const urlParams = new URLSearchParams(window.location.search)
  const token = urlParams.get('token')
  if (token) {
    localStorage.setItem('token', token)
    authAPI.getUserInfo().then(res => {
      localStorage.setItem('user', JSON.stringify(res.data))
      websocket.connect()
      router.push('/')
    }).catch(err => {
      error.value = '社交登录失效，请重试'
    })
  }
})

const handleSubmit = async () => {
  error.value = ''
  loading.value = true

  try {
    const response = await authAPI.login(username.value, password.value, twoFactorCode.value)
    
    if (response.data.two_factor_required) {
      twoFactorRequired.value = true
      loading.value = false
      return
    }

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

const handleSocialLogin = (provider: string) => {
  window.location.href = `/api/auth/${provider}/login`
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
            <template v-if="!twoFactorRequired">
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
            </template>

            <template v-else>
              <div class="space-y-1.5 animate-in slide-in-from-right duration-300">
                <label class="text-[10px] font-black text-blue-600 uppercase tracking-widest ml-1">双重认证 / 2FA Verification</label>
                <div class="relative">
                  <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-blue-600">
                    <Lock class="w-4 h-4" />
                  </div>
                  <input
                    v-model="twoFactorCode"
                    type="text"
                    required
                    maxlength="6"
                    class="w-full bg-blue-50 border border-blue-200 text-slate-800 pl-11 pr-4 py-3.5 rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-400 text-center tracking-[0.5em] text-lg font-black"
                    placeholder="000000"
                    autofocus
                  />
                </div>
                <p class="text-[10px] text-slate-400 font-bold text-center mt-2 px-4">请输入您移动端认证应用中的 6 位动态验证码</p>
                <button type="button" @click="twoFactorRequired = false" class="w-full text-[10px] font-black text-slate-400 uppercase hover:text-slate-600 transition-colors mt-2">返回登录</button>
              </div>
            </template>

            <button
              type="submit"
              :disabled="loading"
              :class="cn(
                'w-full h-14 rounded-2xl font-black text-white transition-all shadow-lg active:scale-95 flex items-center justify-center gap-2',
                loading 
                  ? 'bg-slate-400 cursor-not-allowed' 
                  : (twoFactorRequired ? 'bg-blue-600 hover:bg-blue-500 shadow-blue-500/20' : 'bg-blue-700 hover:bg-blue-600 shadow-blue-500/20')
              )"
            >
              <template v-if="loading">
                <Loader2 class="w-5 h-5 animate-spin" />
              </template>
              <template v-else>
                <span class="uppercase tracking-widest text-sm">{{ twoFactorRequired ? '提交验证' : '初始化访问' }}</span>
                <Fingerprint class="w-4 h-4" />
              </template>
            </button>
          </form>

          <!-- Social Login -->
          <div class="mt-8">
            <div class="relative flex items-center justify-center mb-6">
              <div class="w-full h-px bg-slate-200 px-10"></div>
              <span class="absolute bg-white px-4 text-[10px] font-black text-slate-400 uppercase tracking-widest leading-none">第三方实验室认证</span>
            </div>
            
            <div class="grid grid-cols-3 gap-3">
              <button 
                @click="handleSocialLogin('microsoft')"
                class="flex flex-col items-center justify-center p-3 rounded-2xl border border-slate-100 bg-slate-50 hover:bg-white hover:border-blue-500/30 hover:shadow-md transition-all group"
                title="Microsoft Login"
              >
                <div class="w-5 h-5 mb-1 flex items-center justify-center">
                  <svg viewBox="0 0 23 23" xmlns="http://www.w3.org/2000/svg" class="w-full h-full"><path fill="#f35325" d="M1 1h10v10H1z"/><path fill="#81bc06" d="M12 1h10v10H12z"/><path fill="#05a6f0" d="M1 12h10v10H1z"/><path fill="#ffba08" d="M12 12h10v10H12z"/></svg>
                </div>
                <span class="text-[8px] font-black text-slate-400 group-hover:text-blue-600 uppercase tracking-tighter">Microsoft</span>
              </button>
              <button 
                @click="handleSocialLogin('google')"
                class="flex flex-col items-center justify-center p-3 rounded-2xl border border-slate-100 bg-slate-50 hover:bg-white hover:border-blue-500/30 hover:shadow-md transition-all group"
                title="Google Login"
              >
                <div class="w-5 h-5 mb-1 flex items-center justify-center">
                  <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" class="w-full h-full"><path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/><path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/><path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z" fill="#FBBC05"/><path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/></svg>
                </div>
                <span class="text-[8px] font-black text-slate-400 group-hover:text-blue-600 uppercase tracking-tighter">Google</span>
              </button>
              <button 
                class="flex flex-col items-center justify-center p-3 rounded-2xl border border-slate-100 bg-slate-50 hover:bg-white hover:border-blue-500/30 hover:shadow-md transition-all group opacity-50 cursor-not-allowed"
                title="WeChat Login (Coming Soon)"
              >
                <div class="w-5 h-5 mb-1 flex items-center justify-center">
                   <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" class="w-full h-full fill-slate-400 group-hover:fill-blue-600 transition-colors"><path d="M8.2 13.97c-.37 0-.67-.3-.67-.67s.3-.67.67-.67.67.3.67.67-.3.67-.67.67zm4.3 0c-.37 0-.67-.3-.67-.67s.3-.67.67-.67.67.3.67.67-.3.67-.67.67zm6.75-2.43c0-2.88-2.67-4.88-5.83-4.88-3.16 0-5.83 2-5.83 4.88s2.67 4.88 5.83 4.88c.67 0 1.3-.1 1.93-.24l1.62.91-.37-1.55c1.62-.91 2.65-2.31 2.65-4zm-8.25-1.5c-.37 0-.67-.3-.67-.67s.3-.67.67-.67.67.3.67.67-.3.67-.67.67zm3.3 0c-.37 0-.67-.3-.67-.67s.3-.67.67-.67.67.3.67.67-.3.67-.67.67zM23 9.47c0-3.66-3.37-6.52-7.5-6.52s-7.5 2.86-7.5 6.52c0 .3.04.6.11.89C3.67 10.36 1 12.87 1 15.91c0 3.33 2.97 6.13 6.67 6.13a8.9 8.9 0 0 0 1.94-.21l1.71 1.05-.39-1.8c1.67-.98 2.73-2.5 2.73-4.22.45.09.91.13 1.34.13 4.13 0 7.5-2.86 7.5-6.52z"/></svg>
                </div>
                <span class="text-[8px] font-black text-slate-400 group-hover:text-blue-600 uppercase tracking-tighter">WeChat</span>
              </button>
            </div>
          </div>

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
