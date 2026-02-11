<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { Send, MessageSquare, User, X, FlaskConical, Users, Trophy } from 'lucide-vue-next'
import websocket from '../utils/websocket'
import { authAPI, gameAPI } from '../utils/api'
import { cn } from '../utils/cn'

const router = useRouter()

const props = defineProps<{
  roomId?: string
  title?: string
  placeholder?: string
  maxHeight?: string
  hideHeader?: boolean
}>()

const emit = defineEmits(['close', 'input-focus', 'input-blur'])

const messages = ref<any[]>([])
const newMessage = ref('')
const currentUID = ref(JSON.parse(localStorage.getItem('user') || '{}').uid)
const scrollContainer = ref<HTMLElement | null>(null)

// 聊天模式切换
const chatMode = ref<'normal' | 'private'>('normal')
const privateTarget = ref<{uid: number, username: string} | null>(null)

// 房间状态缓存
const roomStatusCache = ref<Record<string, { status: string, checkedAt: number }>>({})

const scrollToBottom = () => {
  if (scrollContainer.value) {
    scrollContainer.value.scrollTop = scrollContainer.value.scrollHeight
  }
}

// 检查房间状态
const checkRoomStatus = async (roomId: string) => {
  // 检查缓存（缓存5秒）
  const cached = roomStatusCache.value[roomId]
  if (cached && Date.now() - cached.checkedAt < 5000) {
    return cached.status
  }

  try {
    const res = await gameAPI.checkRoomStatus(roomId)
    const status = res.data.exists ? res.data.status : 'closed'
    roomStatusCache.value[roomId] = {
      status,
      checkedAt: Date.now()
    }
    return status
  } catch (err) {
    // 如果出错，假设房间已关闭
    return 'closed'
  }
}

const loadHistory = async () => {
  try {
    // 加载大厅聊天历史
    const chatRes = await authAPI.getGlobalChatHistory(50)
    const chatMessages = (chatRes.data || []).map((m: any) => ({
      uid: m.user_uid,
      username: m.username,
      avatar: m.avatar,
      text: m.message,
      time: new Date(m.created_at),
      type: 'normal'
    }))

    messages.value = chatMessages

    nextTick(scrollToBottom)
  } catch (err) {
    console.error('加载聊天历史失败', err)
  }
}

onMounted(() => {
  // 只有非房间内聊天才加载全服历史并加入大厅频道
  if (!props.roomId) {
    loadHistory()
    // 确保连接建立后立即尝试加入
    websocket.send({ type: 'join_room', room_id: 'lobby' })
  }

  const handleChatMessage = (msg: any) => {
    // 检查是否已存在（避免重复显示历史记录中的消息）
    const isDuplicate = messages.value.some(m =>
      m.uid === msg.uid &&
      m.text === msg.message &&
      Math.abs(new Date(m.time).getTime() - new Date().getTime()) < 2000
    )
    if (isDuplicate) return

    messages.value.push({
      uid: msg.uid,
      username: msg.data?.nickname || msg.data?.username || '研究员',
      avatar: msg.data?.avatar,
      text: msg.message,
      time: new Date(),
      type: 'normal'
    })
    nextTick(scrollToBottom)
  }

  const handlePrivateMessage = async (msg: any) => {
    // 尝试解析游戏邀请消息
    let isGameInvite = false
    let gameInviteData = null

    try {
      const parsed = JSON.parse(msg.message)
      if (parsed.type === 'game_invite') {
        isGameInvite = true
        gameInviteData = parsed

        // 检查房间状态
        if (gameInviteData.room_id) {
          gameInviteData.room_status = await checkRoomStatus(gameInviteData.room_id)
        }
      }
    } catch (e) {
      // 不是JSON或不是游戏邀请，按普通消息处理
    }

    messages.value.push({
      uid: msg.uid,
      target_uid: msg.target_uid,
      username: msg.data?.nickname || msg.data?.username || '研究员',
      avatar: msg.data?.avatar,
      text: isGameInvite ? '' : msg.message,
      time: new Date(),
      type: isGameInvite ? 'game_invite' : 'private',
      gameInviteData: gameInviteData
    })
    nextTick(scrollToBottom)
  }

  const handlePrivateMessage = async (msg: any) => {
    // 尝试解析游戏邀请消息
    let isGameInvite = false
    let gameInviteData = null

    try {
      const parsed = JSON.parse(msg.message)
      if (parsed.type === 'game_invite') {
        isGameInvite = true
        gameInviteData = parsed

        // 检查房间状态
        if (gameInviteData.room_id) {
          gameInviteData.room_status = await checkRoomStatus(gameInviteData.room_id)
        }
      }
    } catch (e) {
      // 不是JSON或不是游戏邀请，按普通消息处理
    }

    messages.value.push({
      uid: msg.uid,
      target_uid: msg.target_uid,
      username: msg.data?.nickname || msg.data?.username || '研究员',
      avatar: msg.data?.avatar,
      text: isGameInvite ? '' : msg.message,
      time: new Date(),
      type: isGameInvite ? 'game_invite' : 'private',
      gameInviteData: gameInviteData
    })
    nextTick(scrollToBottom)
  }

  websocket.on('chat', handleChatMessage)
  websocket.on('private_chat', handlePrivateMessage)

  const handleStartPrivateChat = (e: CustomEvent) => {
    privateTarget.value = (e as any).detail
    chatMode.value = 'private'
  }

  // 监听外部私聊请求
  window.addEventListener('start-private-chat', handleStartPrivateChat as any)

  onUnmounted(() => {
    websocket.off('chat', handleChatMessage)
    websocket.off('private_chat', handlePrivateMessage)
    window.removeEventListener('start-private-chat', handleStartPrivateChat as any)
  })
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

const handleJoinGame = (roomId: string) => {
  router.push(`/room/${roomId}`)
}

const formatTime = (date: Date) => {
  return date.getHours().toString().padStart(2, '0') + ':' +
         date.getMinutes().toString().padStart(2, '0')
}
</script>

<template>
  <div
    :class="cn('flex flex-col bg-white/95 dark:bg-slate-900/60 backdrop-blur-xl border border-slate-200 dark:border-white/10 rounded-[28px] overflow-hidden shadow-2xl', $attrs.class)"
    :style="maxHeight ? { height: maxHeight } : {}"
  >
    <!-- Header - 移动端优化 -->
    <div v-if="!hideHeader" class="px-4 py-3 border-b border-slate-100 dark:border-white/5 flex items-center justify-between bg-slate-50/50 dark:bg-white/[0.02] shrink-0">
      <div class="flex items-center gap-2">
        <div class="w-7 h-7 sm:w-6 sm:h-6 rounded-lg bg-blue-500/10 flex items-center justify-center">
          <MessageSquare class="w-4 h-4 sm:w-3.5 sm:h-3.5 text-blue-500" />
        </div>
        <div>
          <h3 class="text-xs-mobile font-black uppercase tracking-widest text-slate-800 dark:text-white">{{ title || '实验通信频道' }}</h3>
          <p class="text-[9px] sm:text-[8px] font-mono text-slate-400 uppercase tracking-tighter">Messaging_Protocol</p>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <div class="flex items-center gap-1 px-2 py-0.5 bg-blue-500/10 rounded-full border border-blue-500/20">
          <span class="w-1 h-1 bg-blue-500 rounded-full animate-pulse"></span>
          <span class="text-[9px] sm:text-[8px] font-black text-blue-500 uppercase">Live</span>
        </div>
        <button
          v-if="roomId"
          @click="emit('close')"
          class="p-1.5 sm:p-1 hover:bg-slate-200 dark:hover:bg-white/10 rounded-lg transition-colors text-slate-400 hover:text-slate-600 dark:hover:text-white touch-feedback"
        >
          <X class="w-4 h-4" />
        </button>
      </div>
    </div>

    <!-- Messages - 移动端优化 -->
    <div
      ref="scrollContainer"
      class="flex-1 overflow-y-auto p-3 sm:p-2 space-y-2.5 sm:space-y-2 custom-scrollbar bg-transparent"
    >
      <div v-if="messages.length === 0" class="flex flex-col items-center justify-center h-full py-10 opacity-20">
        <div class="w-14 h-14 sm:w-12 sm:h-12 rounded-2xl bg-slate-200 dark:bg-white/10 flex items-center justify-center mb-3">
          <MessageSquare class="w-7 h-7 sm:w-6 sm:h-6" />
        </div>
        <p class="text-xs-mobile font-black uppercase tracking-widest text-slate-500">连接已建立 | 等待数据包...</p>
      </div>

      <div
        v-for="(msg, idx) in messages"
        :key="idx"
        :class="cn(
          'flex gap-2.5 sm:gap-2 max-w-[95%]',
          msg.uid === currentUID ? 'ml-auto flex-row-reverse' : 'mr-auto flex-row'
        )"
      >
        <!-- Avatar - 移动端增大 -->
        <div class="shrink-0 mt-1">
          <div class="w-8 h-8 sm:w-7 sm:h-7 rounded-lg bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 flex items-center justify-center text-xs sm:text-[10px] overflow-hidden shadow-sm">
            <template v-if="msg.avatar && msg.avatar.startsWith('data:')">
              <img :src="msg.avatar" class="w-full h-full object-cover" />
            </template>
            <template v-else>
              {{ msg.avatar || '🧪' }}
            </template>
          </div>
        </div>

        <div :class="cn(
          'flex flex-col gap-0.5',
          msg.uid === currentUID ? 'items-end' : 'items-start'
        )">
          <div class="flex items-center gap-1 px-0.5">
            <span v-if="msg.uid !== currentUID" class="text-[9px] sm:text-[8px] font-black text-slate-400 uppercase tracking-tighter">
              {{ msg.username }}
              <span v-if="msg.type === 'private'" class="text-rose-500 ml-1">(私语)</span>
              <span v-if="msg.type === 'game_invite'" class="text-blue-500 ml-1">(游戏邀请)</span>
            </span>
            <span v-else-if="msg.type === 'private'" class="text-[9px] sm:text-[8px] font-black text-rose-500 uppercase tracking-tighter">
              对 {{ msg.target_uid === currentUID ? '自己' : '研究员' }} 说道
            </span>
            <span v-else-if="msg.type === 'game_invite'" class="text-[9px] sm:text-[8px] font-black text-blue-500 uppercase tracking-tighter">
              发送了游戏邀请
            </span>
            <span class="text-[8px] sm:text-[7px] font-mono text-slate-300 dark:text-slate-600">{{ formatTime(msg.time) }}</span>
          </div>

          <!-- Game Invite Card - 移动端优化 -->
          <div v-if="msg.type === 'game_invite' && msg.gameInviteData"
            :class="cn(
              'w-full max-w-xs p-3.5 sm:p-3 rounded-2xl bg-gradient-to-br border-2 shadow-lg backdrop-blur-sm transition-all',
              msg.gameInviteData.room_status === 'finished' || msg.gameInviteData.room_status === 'closed'
                ? 'from-slate-300/30 to-slate-400/30 border-slate-400/30 grayscale opacity-60'
                : 'from-blue-500/10 to-purple-500/10 border-blue-500/20'
            )"
          >
            <div class="flex items-center gap-2 mb-2">
              <div :class="cn(
                'w-9 h-9 sm:w-8 sm:h-8 rounded-lg flex items-center justify-center',
                msg.gameInviteData.room_status === 'finished' || msg.gameInviteData.room_status === 'closed'
                  ? 'bg-slate-500/20'
                  : 'bg-blue-500/20'
              )">
                <FlaskConical :class="cn(
                  'w-5 h-5 sm:w-4 sm:h-4',
                  msg.gameInviteData.room_status === 'finished' || msg.gameInviteData.room_status === 'closed'
                    ? 'text-slate-500'
                    : 'text-blue-500'
                )" />
              </div>
              <div class="flex-1 min-w-0">
                <div class="text-xs-mobile font-black text-slate-800 dark:text-white truncate">
                  {{ msg.gameInviteData.room_name }}
                </div>
                <div :class="cn(
                  'text-[9px] sm:text-[8px] font-mono uppercase',
                  msg.gameInviteData.room_status === 'finished' || msg.gameInviteData.room_status === 'closed'
                    ? 'text-slate-500'
                    : 'text-slate-500'
                )">
                  {{ msg.gameInviteData.room_status === 'finished' || msg.gameInviteData.room_status === 'closed' ? '房间已关闭' : '实验室邀请' }}
                </div>
              </div>
            </div>

            <div v-if="msg.gameInviteData.room_status !== 'finished' && msg.gameInviteData.room_status !== 'closed'" class="flex items-center gap-3 mb-3 text-xs-mobile">
              <div class="flex items-center gap-1 text-slate-600 dark:text-slate-400">
                <Users class="w-3.5 h-3.5 sm:w-3 sm:h-3" />
                <span class="font-bold">{{ msg.gameInviteData.player_count }}/{{ msg.gameInviteData.max_players }}</span>
              </div>
              <div v-if="msg.gameInviteData.is_points_mode" class="flex items-center gap-1 px-2 py-0.5 bg-amber-500/10 text-amber-600 dark:text-amber-500 rounded-lg">
                <Trophy class="w-3.5 h-3.5 sm:w-3 sm:h-3" />
                <span class="font-black uppercase tracking-widest text-[9px] sm:text-[8px]">积分模式</span>
              </div>
            </div>

            <button
              @click="msg.gameInviteData.room_status !== 'finished' && msg.gameInviteData.room_status !== 'closed' && handleJoinGame(msg.gameInviteData.room_id)"
              :disabled="msg.gameInviteData.room_status === 'finished' || msg.gameInviteData.room_status === 'closed'"
              :class="cn(
                'w-full h-10 sm:h-9 rounded-xl font-black text-xs-mobile uppercase tracking-widest transition-all shadow-md touch-feedback flex items-center justify-center gap-2',
                msg.gameInviteData.room_status === 'finished' || msg.gameInviteData.room_status === 'closed'
                  ? 'bg-slate-400 text-slate-600 cursor-not-allowed'
                  : 'bg-blue-600 hover:bg-blue-500 text-white'
              )"
            >
              <FlaskConical class="w-4 h-4 sm:w-3.5 sm:h-3.5" />
              {{ msg.gameInviteData.room_status === 'finished' || msg.gameInviteData.room_status === 'closed' ? '房间已关闭' : '立即加入实验室' }}
            </button>
          </div>

          <!-- Normal Message - 移动端优化 -->
          <div v-else :class="cn(
            'px-3 sm:px-2 py-2 sm:py-1 rounded-xl text-xs-mobile font-medium leading-relaxed break-words shadow-sm',
            msg.type === 'private' ? 'border-2 border-rose-500/10' : '',
            msg.uid === currentUID
              ? 'bg-blue-600 text-white rounded-tr-none'
              : 'bg-slate-100 dark:bg-white/5 text-slate-700 dark:text-slate-200 border border-slate-200/50 dark:border-white/5 rounded-tl-none'
          )">
            {{ msg.text }}
          </div>
        </div>
      </div>
    </div>

    <!-- Input - 移动端优化 -->
    <div class="p-3 sm:p-2.5 pb-4 sm:pb-3 border-t border-slate-100 dark:border-white/10 bg-white/90 dark:bg-slate-800/60 space-y-2 shrink-0">
      <!-- Mode Selector -->
      <div v-if="chatMode === 'private'" class="flex items-center gap-2 animate-in slide-in-from-bottom-1">
        <div class="flex items-center gap-2 px-3 py-1.5 sm:py-1 rounded-lg text-xs-mobile font-black uppercase tracking-widest transition-all bg-rose-500 text-white shadow-lg shadow-rose-500/20">
          <User class="w-3 h-3 sm:w-2.5 sm:h-2.5" />
          {{ `私密传输: ${privateTarget?.username}` }}
        </div>
        <button @click="chatMode = 'normal'; privateTarget = null" class="p-1.5 sm:p-1 px-2.5 sm:px-2 rounded-lg bg-slate-200 dark:bg-white/10 text-slate-500 hover:text-slate-700 dark:hover:text-white transition-all uppercase text-[9px] sm:text-[8px] font-black touch-feedback">
          取消
        </button>
      </div>

      <div class="flex gap-2">
        <div class="flex-1 relative group">
          <div class="absolute -inset-0.5 bg-gradient-to-r from-blue-500 to-indigo-500 rounded-xl blur opacity-0 group-focus-within:opacity-30 transition duration-500"></div>
          <input
            v-model="newMessage"
            type="text"
            :placeholder="placeholder || '输入实验指令或群聊讯息...'"
            class="relative w-full h-11 sm:h-10 bg-white dark:bg-slate-700/80 border border-slate-200 dark:border-white/10 rounded-xl px-3 sm:px-2.5 text-sm sm:text-xs font-bold focus:outline-none focus:border-blue-500/50 transition-all dark:text-white dark:placeholder:text-slate-500"
            @keydown.enter="handleSend"
            @focus="emit('input-focus')"
            @blur="emit('input-blur')"
          />
        </div>
        <button
          @click="handleSend"
          :disabled="!newMessage.trim()"
          class="btn-touch bg-blue-600 hover:bg-blue-500 disabled:opacity-50 disabled:grayscale text-white rounded-xl flex items-center justify-center transition-all shadow-lg shadow-blue-500/20 touch-feedback shrink-0 group/send"
        >
          <Send class="w-4 h-4 sm:w-3.5 sm:h-3.5 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 transition-transform" />
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
