<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import websocket from '../utils/websocket'

const router = useRouter()

onMounted(async () => {
  // Token已由后端通过HttpOnly Cookie设置，直接调用API获取用户信息
  try {
    // 调用 /api/user/info 获取用户信息，cookie会自动被发送
    const res = await fetch('/api/user/info', {
      credentials: 'include', // 确保cookie被发送
    })
    
    if (!res.ok) {
      console.error('[OAuthCallback] 获取用户信息失败:', res.status)
      router.replace('/login')
      return
    }

    const data = await res.json()
    const user = data.user ?? data
    
    if (!user || !user.uid) {
      console.error('[OAuthCallback] 无效的用户信息')
      router.replace('/login')
      return
    }

    // 存储用户信息
    localStorage.setItem('user', JSON.stringify(user))
    websocket.connect()
    window.dispatchEvent(new Event('auth-changed'))

    console.log('[OAuthCallback] OAuth登录成功，跳转主页')
    router.replace('/')
  } catch (e) {
    console.error('[OAuthCallback] 处理OAuth回调失败:', e)
    router.replace('/login')
  }
})
</script>

<template>
  <div class="min-h-screen flex flex-col items-center justify-center bg-slate-50 dark:bg-[#1a1a1e] gap-4">
    <div class="w-10 h-10 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
    <p class="text-slate-500 dark:text-slate-400 text-sm font-bold">正在完成授权，请稍候...</p>
  </div>
</template>
