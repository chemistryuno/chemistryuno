<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import websocket from '../utils/websocket'
import { commonAPI } from '../utils/api'
import { Bell, X } from 'lucide-vue-next'

const announcements = ref<any[]>([])
const currentIndex = ref(0)
const isVisible = ref(false)

const fetchAnnouncements = async () => {
  try {
    const res = await commonAPI.getAnnouncements()
    // 只显示跑马灯类型的
    announcements.value = (res.data || []).filter((a: any) => a.is_ticker)
    if (announcements.value.length > 0) {
      isVisible.value = true
    } else {
      isVisible.value = false
    }
  } catch (err) {
    console.error('获取公告失败:', err)
  }
}

const handleAnnouncement = (msg: any) => {
  const ann = msg.data
  if (ann && ann.is_ticker) {
    // 检查是否已存在
    const idx = announcements.value.findIndex(a => a.id === ann.id)
    if (idx !== -1) {
      announcements.value[idx] = ann
    } else {
      announcements.value.unshift(ann)
    }
    isVisible.value = true
  }
}

let tickerInterval: any

onMounted(() => {
  fetchAnnouncements()
  websocket.on('system_announcement', handleAnnouncement)
  
  tickerInterval = setInterval(() => {
    if (announcements.value.length > 1) {
      currentIndex.value = (currentIndex.value + 1) % announcements.value.length
    }
  }, 8000)
})

onUnmounted(() => {
  websocket.off('system_announcement', handleAnnouncement)
  if (tickerInterval) clearInterval(tickerInterval)
})

const closeTicker = () => {
  isVisible.value = false
}
</script>

<template>
  <div v-if="isVisible && announcements.length > 0"
       class="fixed top-0 left-0 right-0 z-[80] bg-blue-600/90 dark:bg-blue-900/40 backdrop-blur-md border-b border-blue-400/30 text-white overflow-hidden shadow-lg shadow-blue-950/20 transition-opacity duration-300 ease-out">
    <div class="container mx-auto px-4 py-1 flex items-center gap-2">
      <Bell class="w-3.5 h-3.5 animate-bounce flex-shrink-0" />
      
      <div class="flex-grow overflow-hidden relative h-5">
        <TransitionGroup 
          name="ticker" 
          tag="div"
          class="absolute inset-0"
        >
          <div 
            v-for="(ann, index) in announcements" 
            :key="ann.id"
            v-show="index === currentIndex"
            class="flex items-center gap-2 whitespace-nowrap"
          >
            <span v-if="ann.type === 'emergency'" class="bg-red-500 text-[8px] px-1 py-0.5 rounded font-bold animate-pulse uppercase">紧急</span>
            <span v-else-if="ann.type === 'maintenance'" class="bg-amber-500 text-[8px] px-1 py-0.5 rounded font-bold uppercase">维护</span>
            <span v-else class="bg-emerald-500 text-[8px] px-1 py-0.5 rounded font-bold uppercase">公告</span>
            <span class="text-xs font-black italic tracking-wide truncate max-w-[80vw]">{{ ann.content }}</span>
          </div>
        </TransitionGroup>
      </div>

      <button @click="closeTicker" class="hover:bg-white/10 p-1 rounded-full transition-colors">
        <X class="w-3.5 h-3.5" />
      </button>
    </div>
  </div>
</template>

<style src="./AnnouncementTicker.css" scoped></style>
