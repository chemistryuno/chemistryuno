<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import websocket from '../utils/websocket'

const router = useRouter()

onMounted(async () => {
  // 从 URL hash 中读取参数（后端降级重定向时写入）
  // 新方案：#access_token=...&refresh_token=...
  // 旧方案（兼容）：#token=...
  const hash = window.location.hash.slice(1) // 移除 #
  if (!hash) {
    console.warn('[OAuthCallback] 无效的回调参数，跳转到登录页')
    router.replace('/login')
    return
  }

  const params = new URLSearchParams(hash)
  const accessToken = params.get('access_token')
  const refreshToken = params.get('refresh_token')
  const legacyToken = params.get('token')
  
  // 优先使用新的双token方案，回退到旧的单token
  const token = accessToken || legacyToken
  if (!token) {
    console.warn('[OAuthCallback] token/access_token 为空，跳转到登录页')
    router.replace('/login')
    return
  }

  try {
    // 先用 JWT payload 解码出基础用户信息供临时存储
    const parts = token.split('.')
    if (parts.length !== 3) throw new Error('invalid jwt format')
    const base64Payload = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const padded = base64Payload + '='.repeat((4 - base64Payload.length % 4) % 4)
    const decoded = JSON.parse(atob(padded))

    // 写入 token（支持新旧方案）
    if (accessToken && refreshToken) {
      // 新的双token方案
      localStorage.setItem('access_token', accessToken)
      localStorage.setItem('refresh_token', refreshToken)
    } else if (legacyToken) {
      // 旧的单token方案
      localStorage.setItem('token', legacyToken)
    }

    // 调用 /api/user/info 获取完整用户信息（含 nickname、avatar 等）
    let user: any = { uid: decoded.uid, email: decoded.email, is_admin: decoded.is_admin, role: decoded.role }
    try {
      const res = await fetch('/api/user/info', {
        headers: { Authorization: `Bearer ${token}` }
      })
      if (res.ok) {
        const data = await res.json()
        user = data.user ?? data
      }
    } catch (fetchErr) {
      console.warn('[OAuthCallback] 获取完整用户信息失败，使用 JWT 解码的基础信息:', fetchErr)
    }

    localStorage.setItem('user', JSON.stringify(user))
    websocket.connect()
    window.dispatchEvent(new Event('auth-changed'))

    console.log('[OAuthCallback] 登录成功（降级模式），跳转主页')
    router.replace('/')
  } catch (e) {
    console.error('[OAuthCallback] 处理 token 失败:', e)
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('token')
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
