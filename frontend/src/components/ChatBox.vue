<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { Send, MessageSquare, User, X } from 'lucide-vue-next'
import websocket from '../utils/websocket'
import { cn } from '../utils/cn'

const props = defineProps<{
  roomId?: string
  title?: string
  placeholder?: string
  maxHeight?: string
}>()

const messages = ref<any[]>([])
const newMessage = ref('')
const currentUID = ref(JSON.parse(localStorage.getItem('user') || '{}').uid)
const scrollContainer = ref<HTMLElement | null>(null)

// 聊天模式切换
const chatMode = ref<'normal' | 'private'>('normal')
const privateTarget = ref<{uid: number, username: string} | null>(null)

const scrollToBottom = () => {
  if (scrollContainer.value) {
    scrollContainer.value.scrollTop = scrollContainer.value.scrollHeight
  }
}

onMounted(() => {
  const handleChatMessage = (msg: any) => {
    messages.value.push({
      uid: msg.uid,
      username: msg.data?.username || '研究员',
      text: msg.message,
      time: new Date(),
      type: 'normal'
    })
    nextTick(scrollToBottom)
  }

  const handlePrivateMessage = (msg: any) => {
    messages.value.push({
      uid: msg.uid,
      target_uid: msg.target_uid,
      username: msg.data?.username || '研究员',
      text: msg.message,
      time: new Date(),
      type: 'private'
    })
    nextTick(scrollToBottom)
  }

  websocket.on('chat', handleChatMessage)
  websocket.on('private_chat', handlePrivateMessage)

  // 监听外部私聊请求
  window.addEventListener('start-private-chat', ((e: CustomEvent) => {
    privateTarget.value = e.detail
    chatMode.value = 'private'
  }) as any)
})

const handleSend = () => {
  if (!newMessage.value.trim()) return

  if (chatMode.value === 'private' && privateTarget.value) {
    websocket.send({
      type: 'private_chat',
      target_uid: privateTarget.value.uid,
      message: newMessage.value
    })
    // 服务器会给发送者也发一个 private_chat，所以这里不用手动 push
  } else {
    websocket.send({
      type: 'chat',
      message: newMessage.value
    })
  }
  
  newMessage.value = ''
}

const formatTime = (date: Date) => {
  return date.getHours().toString().padStart(2, '0') + ':' + 
         date.getMinutes().toString().padStart(2, '0')
}
</script>

<template>
  <div class="flex flex-col bg-white dark:bg-[#121216]/80 backdrop-blur-xl border border-slate-200 dark:border-white/10 rounded-[28px] overflow-hidden shadow-2xl">
    <!-- Header -->
    <div class="px-6 py-4 border-b border-slate-100 dark:border-white/5 flex items-center justify-between bg-slate-50/50 dark:bg-white/[0.02]">
      <div class="flex items-center gap-2">
        <div class="w-8 h-8 rounded-xl bg-blue-500/10 flex items-center justify-center">
          <MessageSquare class="w-4 h-4 text-blue-500" />
        </div>
        <div>
          <h3 class="text-xs font-black uppercase tracking-widest text-slate-800 dark:text-white">{{ title || '实验通信频道' }}</h3>
          <p class="text-[9px] font-mono text-slate-400 uppercase tracking-tighter">Secure_Messaging_Protocol</p>
        </div>
      </div>
      <div class="flex items-center gap-1.5 px-2.5 py-1 bg-blue-500/10 rounded-full border border-blue-500/20">
        <span class="w-1.5 h-1.5 bg-blue-500 rounded-full animate-pulse"></span>
        <span class="text-[9px] font-black text-blue-500 uppercase">Live</span>
      </div>
    </div>

    <!-- Messages -->
    <div 
      ref="scrollContainer"
      class="flex-1 overflow-y-auto p-6 space-y-4 custom-scrollbar"
      :style="{ maxHeight: maxHeight || '300px', minHeight: '200px' }"
    >
      <div v-if="messages.length === 0" class="flex flex-col items-center justify-center h-full py-10 opacity-20">
        <MessageSquare class="w-12 h-12 mb-2" />
        <p class="text-[10px] font-black uppercase tracking-widest">等待信号传输...</p>
      </div>
      
      <div 
        v-for="(msg, idx) in messages" 
        :key="idx"
        :class="cn(
          'flex flex-col gap-1 max-w-[85%]',
          msg.uid === currentUID ? 'ml-auto items-end' : 'mr-auto items-start'
        )"
      >
        <div class="flex items-center gap-2 px-1">
          <span v-if="msg.uid !== currentUID" class="text-[9px] font-black text-slate-400 uppercase tracking-tighter">
            {{ msg.username }}
            <span v-if="msg.type === 'private'" class="text-rose-500 ml-1">(私语)</span>
          </span>
          <span v-else-if="msg.type === 'private'" class="text-[9px] font-black text-rose-500 uppercase tracking-tighter">
            对 {{ msg.target_uid === currentUID ? '自己' : '研究员' }} 说道
          </span>
          <span class="text-[8px] font-mono text-slate-300 dark:text-slate-600">{{ formatTime(msg.time) }}</span>
        </div>
        <div :class="cn(
          'px-4 py-2.5 rounded-2xl text-sm font-medium leading-relaxed break-words shadow-sm',
          msg.type === 'private' ? 'border-2 border-rose-500/20' : '',
          msg.uid === currentUID 
            ? 'bg-blue-600 text-white rounded-tr-none' 
            : 'bg-slate-100 dark:bg-white/5 text-slate-700 dark:text-slate-200 border border-slate-200/50 dark:border-white/5 rounded-tl-none'
        )">
          {{ msg.text }}
        </div>
      </div>
    </div>

    <!-- Input -->
    <div class="p-6 border-t border-slate-100 dark:border-white/5 bg-slate-50/50 dark:bg-white/[0.01] space-y-4">
      <!-- Mode Selector -->
      <div v-if="chatMode === 'private'" class="flex items-center gap-2 animate-in slide-in-from-bottom-2">
        <div class="flex items-center gap-2 px-3 py-1.5 rounded-xl text-[10px] font-black uppercase tracking-widest transition-all bg-rose-500 text-white">
          <User class="w-3 h-3" />
          {{ `私聊: ${privateTarget?.username}` }}
        </div>
        <button @click="chatMode = 'normal'; privateTarget = null" class="p-1.5 rounded-lg bg-slate-200 dark:bg-white/10 text-slate-500 hover:bg-slate-300 transition-all">
          <X class="w-3 h-3" />
        </button>
      </div>

      <div class="flex gap-3">
        <div class="flex-1 relative group">
          <div class="absolute -inset-0.5 bg-gradient-to-r from-blue-500 to-cyan-500 rounded-2xl blur opacity-0 group-focus-within:opacity-20 transition duration-500"></div>
          <input 
            v-model="newMessage"
            type="text" 
            :placeholder="placeholder || '向各研究员发送讯息...'"
            class="relative w-full h-12 bg-white dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-2xl px-5 text-xs font-medium focus:outline-none focus:border-blue-500/50 transition-all dark:text-white"
            @keydown.enter="handleSend"
          />
        </div>
        <button 
          @click="handleSend"
          :disabled="!newMessage.trim()"
          class="w-12 h-12 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 disabled:grayscale text-white rounded-2xl flex items-center justify-center transition-all shadow-lg shadow-blue-500/20 active:scale-95 shrink-0"
        >
          <Send class="w-5 h-5" />
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.1);
  border-radius: 10px;
}
.dark .custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.05);
}
</style>
