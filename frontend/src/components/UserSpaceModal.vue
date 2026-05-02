<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../utils/api'
import UserAvatar from './UserAvatar.vue'
import {
  X,
  MessageCircle,
  AtSign,
  Hash,
  Info,
  Shield,
  Mail,
  Send
} from 'lucide-vue-next'
import LevelBadge from './LevelBadge.vue'

const props = defineProps<{
  show: boolean
  uid: number | string | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const router = useRouter()
const user = ref<any>(null)
const loading = ref(false)
const error = ref<string | null>(null)
let activeRequestId = 0

const displayNickname = computed(() => user.value?.nickname || 'Researcher')
const displayRole = computed(() => String(user.value?.role || 'user').toUpperCase())

const resetState = () => {
  user.value = null
  error.value = null
  loading.value = false
}

const normalizeUID = (uid: number | string | null | undefined): number | null => {
  const parsed = Number(uid)
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return null
  }
  return Math.trunc(parsed)
}

const fetchUserProfile = async (uid: number) => {
  const requestId = ++activeRequestId
  loading.value = true
  error.value = null
  user.value = null

  try {
    const response = await authAPI.getUserPublicProfile(uid)
    if (requestId !== activeRequestId) return
    user.value = response.data
  } catch (err: any) {
    if (requestId !== activeRequestId) return
    error.value = err.response?.data?.error || '无法获取研究员资料'
  } finally {
    if (requestId === activeRequestId) {
      loading.value = false
    }
  }
}

watch(
  () => [props.show, props.uid] as const,
  ([show, uid]) => {
    if (show) {
      const normalizedUID = normalizeUID(uid)
      if (normalizedUID !== null) {
        fetchUserProfile(normalizedUID)
        return
      }

      activeRequestId++
      user.value = null
      loading.value = false
      error.value = '无效的用户ID'
      return
    }

    activeRequestId++
    resetState()
  },
  { immediate: true }
)

const handleStartChat = () => {
  if (!user.value) return

  emit('close')
  router.push({
    path: '/chat',
    query: { uid: user.value.uid, nickname: user.value.nickname || 'Researcher' }
  })
}
</script>

<template>
  <Teleport to="body">
    <div v-if="show" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md" @click="emit('close')">

      <div class="relative w-full max-w-2xl bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2.5rem] shadow-2xl overflow-hidden animate-in zoom-in duration-300 flex flex-col max-h-[90vh] pointer-events-auto">
        <button
          @click="emit('close')"
          class="absolute top-6 right-6 p-2 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 rounded-xl transition-all text-slate-400 hover:text-slate-900 dark:hover:text-white z-50"
        >
          <X class="w-5 h-5" />
        </button>

        <div class="overflow-y-auto p-6 md:p-10 custom-scrollbar">
          <div v-if="loading" class="flex flex-col items-center justify-center py-20">
            <div class="w-12 h-12 border-4 border-blue-500/20 border-t-blue-500 rounded-full animate-spin mb-4"></div>
            <p class="text-xs font-black text-slate-400 uppercase tracking-widest animate-pulse">Archiving_Researcher_Data...</p>
          </div>

          <div v-else-if="error" class="py-12 text-center">
            <div class="w-16 h-16 bg-red-500/10 rounded-2xl flex items-center justify-center text-red-500 mx-auto mb-6">
              <Info class="w-8 h-8" />
            </div>
            <h2 class="text-xl font-black text-slate-800 dark:text-white mb-2">资料加载失败</h2>
            <p class="text-slate-500 text-sm mb-6">{{ error }}</p>
          </div>

          <div v-else-if="user" class="space-y-8">
            <div class="flex flex-col md:flex-row items-center md:items-start gap-6 md:gap-8 relative">
              <div class="absolute -top-10 -right-10 opacity-5 pointer-events-none">
                <Shield class="w-64 h-64 -rotate-12" />
              </div>

              <div class="shrink-0 relative z-10">
                <div class="w-28 h-28 md:w-32 md:h-32 bg-gradient-to-tr from-blue-500/20 to-purple-500/20 rounded-[1.8rem] p-1 shadow-xl">
                  <div class="w-full h-full bg-white dark:bg-[#0d0d10] rounded-[1.6rem] flex items-center justify-center text-5xl border border-slate-100 dark:border-white/5 overflow-hidden shadow-inner">
                    <UserAvatar :avatar="user.avatar" />
                  </div>
                </div>
              </div>

              <div class="flex-1 text-center md:text-left space-y-3 relative z-10">
                <div>
                  <div class="flex items-center justify-center md:justify-start gap-2 mb-1.5">
                    <span class="px-2 py-0.5 bg-blue-500/10 text-blue-500 text-[9px] font-black rounded-lg border border-blue-500/20 uppercase tracking-widest">
                      Researcher_Space
                    </span>
                    <LevelBadge :level="user.level" size="xs" />
                  </div>
                  <h1 class="text-3xl font-black text-slate-900 dark:text-white tracking-tighter italic uppercase">
                    {{ displayNickname }}
                  </h1>
                  <p class="text-[9px] font-mono text-slate-400 dark:text-slate-500">UID: {{ user.uid }} | ROLE: {{ displayRole }}</p>
                </div>

                <div class="flex flex-wrap items-center justify-center md:justify-start gap-5 py-1">
                  <div class="flex flex-col">
                    <span class="text-[8px] font-black text-slate-400 uppercase tracking-widest">Phlogiston</span>
                    <span class="text-sm font-black text-slate-800 dark:text-white font-mono">{{ Math.floor(user.points) }}</span>
                  </div>
                  <div class="w-px h-6 bg-slate-100 dark:bg-white/5" />
                  <div class="flex flex-col">
                    <span class="text-[8px] font-black text-slate-400 uppercase tracking-widest">Wins</span>
                    <span class="text-sm font-black text-slate-800 dark:text-white font-mono">{{ user.win_count }}</span>
                  </div>
                  <div class="w-px h-6 bg-slate-100 dark:bg-white/5" />
                  <div class="flex flex-col">
                    <span class="text-[8px] font-black text-slate-400 uppercase tracking-widest">Total Games</span>
                    <span class="text-sm font-black text-slate-800 dark:text-white font-mono">{{ user.total_games }}</span>
                  </div>
                </div>

                <div class="pt-2 flex justify-center md:justify-start">
                  <button
                    @click="handleStartChat"
                    class="flex items-center gap-2 px-5 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-xl font-black text-[9px] uppercase tracking-widest transition-all shadow-lg shadow-blue-500/20 active:scale-95"
                  >
                    <Send class="w-3 h-3" />
                    发起私聊 / Message
                  </button>
                </div>
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-5 gap-6">
              <div class="md:col-span-3 space-y-4">
                <div class="bg-slate-50 dark:bg-white/[0.02] border border-slate-100 dark:border-white/5 rounded-2xl p-6 h-full">
                  <h3 class="text-[10px] font-black uppercase tracking-widest text-slate-400 mb-4 flex items-center gap-2">
                    <Info class="w-3 h-3" />
                    个人简介 / Bio
                  </h3>
                  <p class="text-slate-600 dark:text-slate-300 text-xs leading-relaxed italic whitespace-pre-wrap">
                    {{ user.bio || '这位研究员还没有留下简介。' }}
                  </p>
                </div>
              </div>

              <div class="md:col-span-2 space-y-4">
                <div class="bg-slate-50 dark:bg-white/[0.02] border border-slate-100 dark:border-white/5 rounded-2xl p-6">
                  <h3 class="text-[10px] font-black uppercase tracking-widest text-slate-400 mb-4 flex items-center gap-2">
                    <Mail class="w-3 h-3" />
                    联系方式 / Contacts
                  </h3>

                  <div class="space-y-3">
                    <div v-if="user.wechat" class="flex items-center gap-2.5 p-2 bg-white dark:bg-black/20 rounded-lg border border-slate-100 dark:border-white/5">
                      <MessageCircle class="w-3.5 h-3.5 text-emerald-500 shrink-0" />
                      <span class="text-[10px] font-bold truncate dark:text-white">{{ user.wechat }}</span>
                    </div>

                    <div v-if="user.qq" class="flex items-center gap-2.5 p-2 bg-white dark:bg-black/20 rounded-lg border border-slate-100 dark:border-white/5">
                      <Hash class="w-3.5 h-3.5 text-blue-500 shrink-0" />
                      <span class="text-[10px] font-bold truncate dark:text-white">{{ user.qq }}</span>
                    </div>

                    <div v-if="user.show_email" class="flex items-center gap-2.5 p-2 bg-white dark:bg-black/20 rounded-lg border border-slate-100 dark:border-white/5">
                      <AtSign class="w-3.5 h-3.5 text-purple-500 shrink-0" />
                      <span class="text-[10px] font-bold truncate dark:text-white">{{ user.email }}</span>
                    </div>

                    <div v-if="user.custom_contact" class="flex items-center gap-2.5 p-2 bg-white dark:bg-black/20 rounded-lg border border-slate-100 dark:border-white/5">
                      <Info class="w-3.5 h-3.5 text-amber-500 shrink-0" />
                      <span class="text-[10px] font-bold truncate dark:text-white">{{ user.custom_contact }}</span>
                    </div>

                    <div v-if="!user.wechat && !user.qq && !user.show_email && !user.custom_contact" class="text-center py-2">
                      <p class="text-[9px] text-slate-400 font-bold uppercase tracking-widest italic opacity-40">No_Contacts</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div class="pt-2 text-center">
              <p class="text-[8px] text-slate-400 font-black uppercase tracking-[0.2em] opacity-30">Researcher_Protocol_Alpha_Verified_v2.0</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style src="./UserSpaceModal.css" scoped></style>
