<template>
  <div class="flex flex-col h-full">
    <div class="flex-1 overflow-y-auto space-y-2 p-2">
      <div
        v-for="msg in messages"
        :key="msg.id"
        class="text-sm"
        :class="msg.uid === myUID ? 'text-right' : 'text-left'"
      >
        <span class="inline-block bg-gray-100 dark:bg-gray-800 rounded-xl px-3 py-1.5 max-w-[85%]">
          <span class="block text-xs text-gray-400 mb-0.5">UID {{ msg.uid }}</span>
          {{ msg.content }}
        </span>
      </div>
    </div>
    <div class="flex gap-2 p-2 border-t">
      <input
        v-model="draft"
        @keydown.enter="sendMessage"
        placeholder="发送消息..."
        class="flex-1 border rounded-lg px-3 py-1.5 text-sm"
      />
      <button @click="sendMessage" class="px-3 py-1.5 bg-blue-600 text-white rounded-lg text-sm">发送</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { teamAPI } from '../utils/api'
import { buildApiURL } from '../utils/runtimeConfig'

const messages = ref<any[]>([])
const draft = ref('')
let ws: WebSocket | null = null

const myUID = computed(() => {
  try { return JSON.parse(localStorage.getItem('user') || '{}').uid } catch { return 0 }
})

import { computed } from 'vue'

onMounted(async () => {
  // Load history.
  try {
    const res = await teamAPI.getChatHistory()
    messages.value = res.data || []
  } catch { /* ignore */ }

  // Open WebSocket.
  const token = document.cookie.match(/access_token=([^;]+)/)?.[1] || ''
  const wsUrl = buildApiURL('/teams/chat/ws').replace(/^http/, 'ws') + (token ? `?token=${token}` : '')
  ws = new WebSocket(wsUrl)

  ws.onmessage = (e) => {
    try {
      const data = JSON.parse(e.data)
      if (data.type === 'team_chat_history') {
        messages.value = data.messages || []
      } else if (data.type === 'team_chat_message') {
        messages.value.push(data)
      }
    } catch { /* ignore */ }
  }
})

onUnmounted(() => {
  ws?.close()
})

async function sendMessage() {
  if (!draft.value.trim() || !ws || ws.readyState !== WebSocket.OPEN) return
  ws.send(JSON.stringify({ content: draft.value.trim() }))
  draft.value = ''
}
</script>
