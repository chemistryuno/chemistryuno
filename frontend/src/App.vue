<script setup lang="ts">
import { ref, onMounted } from 'vue'
import websocket from './utils/websocket'

const loading = ref(true)

onMounted(() => {
  // 检查本地存储的token
  const token = localStorage.getItem('token')
  const userData = localStorage.getItem('user')
  
  if (token && userData) {
    // 用户已登录，建立 WebSocket 连接
    websocket.connect()
  }
  loading.value = false
})
</script>

<template>
  <div v-if="loading" class="min-h-screen bg-[#0a0a0c] flex flex-col items-center justify-center gap-4">
    <div class="relative w-16 h-16">
      <div class="absolute inset-0 rounded-full border-4 border-emerald-500/20"></div>
      <div class="absolute inset-0 rounded-full border-4 border-emerald-500 border-t-transparent animate-spin"></div>
    </div>
    <p class="text-emerald-500/70 font-mono tracking-widest text-sm animate-pulse">
      INITIALIZING LABORATORY...
    </p>
  </div>
  <router-view v-else></router-view>
</template>

<style>
/* Any global styles can go here if not in index.css */
</style>
