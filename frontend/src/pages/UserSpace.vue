<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { authAPI } from '../utils/api'
import { 
  ArrowLeft, 
  MessageCircle, 
  AtSign, 
  Hash, 
  Info, 
  Trophy, 
  Shield,
  Mail,
  Send
} from 'lucide-vue-next'
import LevelBadge from '../components/LevelBadge.vue'

const route = useRoute()
const router = useRouter()
const user = ref<any>(null)
const loading = ref(true)
const error = ref<string | null>(null)

const fetchUserProfile = async () => {
  const uid = parseInt(route.params.uid as string)
  if (isNaN(uid)) {
    error.value = '无效的用户ID'
    loading.value = false
    return
  }

  try {
    const response = await authAPI.getUserPublicProfile(uid)
    user.value = response.data
  } catch (err: any) {
    error.value = err.response?.data?.error || '无法获取研究员档案'
  } finally {
    loading.value = false
  }
}

const handleStartChat = () => {
  if (user.value) {
    router.push({
      path: '/chat',
      query: { uid: user.value.uid, nickname: user.value.nickname }
    })
  }
}

onMounted(fetchUserProfile)
</script>

<template>
  <div class="min-h-screen bg-slate-50 dark:bg-[#0a0a0c] selection:bg-blue-500/30">
    <!-- Background Effects -->
    <div class="fixed inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-[-10%] right-[-10%] w-[50%] h-[50%] bg-blue-500/5 rounded-full blur-[120px]" />
      <div class="absolute bottom-[-10%] left-[-10%] w-[50%] h-[50%] bg-purple-500/5 rounded-full blur-[120px]" />
    </div>

    <div class="max-w-4xl mx-auto relative z-10 px-4 py-8 md:py-16">
      <!-- Back Button -->
      <button 
        @click="router.back()" 
        class="mb-8 flex items-center gap-2 text-slate-400 hover:text-slate-900 dark:hover:text-white transition-colors group"
      >
        <ArrowLeft class="w-4 h-4 group-hover:-translate-x-1 transition-transform" />
        <span class="text-xs font-black uppercase tracking-widest">返回 / BACK</span>
      </button>

      <div v-if="loading" class="flex flex-col items-center justify-center py-20">
        <div class="w-12 h-12 border-4 border-blue-500/20 border-t-blue-500 rounded-full animate-spin mb-4"></div>
        <p class="text-xs font-black text-slate-400 uppercase tracking-widest animate-pulse">Archiving_Researcher_Data...</p>
      </div>

      <div v-else-if="error" class="bg-white dark:bg-[#111114] border border-red-500/20 rounded-3xl p-12 text-center">
        <div class="w-16 h-16 bg-red-500/10 rounded-2xl flex items-center justify-center text-red-500 mx-auto mb-6">
          <Info class="w-8 h-8" />
        </div>
        <h2 class="text-xl font-black text-slate-800 dark:text-white mb-2">访问权限受限</h2>
        <p class="text-slate-500 text-sm mb-6">{{ error }}</p>
        <button @click="router.push('/')" class="px-6 py-2.5 bg-slate-100 dark:bg-white/5 rounded-xl text-xs font-black uppercase tracking-widest hover:bg-slate-200 dark:hover:bg-white/10 transition-all">
          返回主界面
        </button>
      </div>

      <div v-else class="space-y-6">
        <!-- Profile Header Card -->
        <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2.5rem] p-8 md:p-12 shadow-xl shadow-blue-500/5 relative overflow-hidden">
          <div class="absolute top-0 right-0 p-8 opacity-5">
            <Shield class="w-64 h-64 -rotate-12" />
          </div>

          <div class="flex flex-col md:flex-row items-center md:items-start gap-8 md:gap-12 relative z-10">
            <!-- Avatar Section -->
            <div class="shrink-0">
              <div class="w-32 h-32 md:w-40 md:h-40 bg-gradient-to-tr from-blue-500/20 to-purple-500/20 rounded-[2rem] p-1.5 shadow-2xl">
                <div class="w-full h-full bg-white dark:bg-[#0d0d10] rounded-[1.8rem] flex items-center justify-center text-6xl border border-slate-100 dark:border-white/5 overflow-hidden">
                  <template v-if="(user.avatar || '').length > 50">
                    <img :src="user.avatar" class="w-full h-full object-cover" />
                  </template>
                  <template v-else>
                    <span>{{ user.avatar || '🧪' }}</span>
                  </template>
                </div>
              </div>
            </div>

            <!-- Basic Info Section -->
            <div class="flex-1 text-center md:text-left space-y-4">
              <div>
                <div class="flex items-center justify-center md:justify-start gap-2 mb-2">
                  <span class="px-2 py-0.5 bg-blue-500/10 text-blue-500 text-[10px] font-black rounded-lg border border-blue-500/20 uppercase tracking-widest">
                    Researcher_Space
                  </span>
                  <LevelBadge :level="user.level" size="sm" />
                </div>
                <h1 class="text-4xl md:text-5xl font-black text-slate-900 dark:text-white tracking-tighter italic uppercase">
                  {{ user.nickname }}
                </h1>
                <p class="text-[10px] font-mono text-slate-400 dark:text-slate-500 mt-1">UID: {{ user.uid }} | ROLE: {{ user.role.toUpperCase() }}</p>
              </div>

              <div class="flex flex-wrap items-center justify-center md:justify-start gap-6 pt-2">
                <div class="flex flex-col">
                  <span class="text-[9px] font-black text-slate-400 uppercase tracking-widest">积分 / Points</span>
                  <span class="text-lg font-black text-slate-800 dark:text-white font-mono">{{ Math.floor(user.points) }}</span>
                </div>
                <div class="w-px h-8 bg-slate-100 dark:bg-white/5" />
                <div class="flex flex-col">
                  <span class="text-[9px] font-black text-slate-400 uppercase tracking-widest">胜场 / Wins</span>
                  <span class="text-lg font-black text-slate-800 dark:text-white font-mono">{{ user.win_count }}</span>
                </div>
                <div class="w-px h-8 bg-slate-100 dark:bg-white/5" />
                <div class="flex flex-col">
                  <span class="text-[9px] font-black text-slate-400 uppercase tracking-widest">总对局 / Total</span>
                  <span class="text-lg font-black text-slate-800 dark:text-white font-mono">{{ user.total_games }}</span>
                </div>
              </div>

              <div class="pt-4 flex flex-wrap justify-center md:justify-start gap-3">
                <button 
                  @click="handleStartChat"
                  class="flex items-center gap-2 px-6 py-2.5 bg-blue-600 hover:bg-blue-500 text-white rounded-xl font-black text-[10px] uppercase tracking-widest transition-all shadow-lg shadow-blue-500/20 active:scale-95"
                >
                  <Send class="w-3.5 h-3.5" />
                  发起私聊 / Message
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <!-- Bio Section -->
          <div class="md:col-span-2 space-y-6">
            <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2rem] p-8 shadow-sm h-full">
              <h3 class="text-xs font-black uppercase tracking-widest text-slate-400 mb-6 flex items-center gap-2">
                <Info class="w-3.5 h-3.5" />
                个人简介 / Biography
              </h3>
              <div class="relative">
                <div class="absolute -left-4 top-0 bottom-0 w-1 bg-blue-500/20 rounded-full" />
                <p class="text-slate-600 dark:text-slate-300 leading-relaxed italic whitespace-pre-wrap pl-2">
                  {{ user.bio || '该研究员比较神秘，还没有留下任何介绍信息。' }}
                </p>
              </div>
            </div>
          </div>

          <!-- Contact Section -->
          <div class="space-y-6">
            <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2rem] p-8 shadow-sm">
              <h3 class="text-xs font-black uppercase tracking-widest text-slate-400 mb-6 flex items-center gap-2">
                <Mail class="w-3.5 h-3.5" />
                联系方式 / Contacts
              </h3>
              
              <div class="space-y-4">
                <div v-if="user.wechat" class="flex items-center gap-3 p-3 bg-slate-50 dark:bg-white/[0.02] rounded-xl border border-slate-100 dark:border-white/5">
                  <MessageCircle class="w-4 h-4 text-emerald-500" />
                  <div class="min-w-0">
                    <p class="text-[8px] font-black text-slate-400 uppercase tracking-tighter">WECHAT</p>
                    <p class="text-sm font-bold truncate dark:text-white">{{ user.wechat }}</p>
                  </div>
                </div>

                <div v-if="user.qq" class="flex items-center gap-3 p-3 bg-slate-50 dark:bg-white/[0.02] rounded-xl border border-slate-100 dark:border-white/5">
                  <Hash class="w-4 h-4 text-blue-500" />
                  <div class="min-w-0">
                    <p class="text-[8px] font-black text-slate-400 uppercase tracking-tighter">QQ</p>
                    <p class="text-sm font-bold truncate dark:text-white">{{ user.qq }}</p>
                  </div>
                </div>

                <div v-if="user.show_email" class="flex items-center gap-3 p-3 bg-slate-50 dark:bg-white/[0.02] rounded-xl border border-slate-100 dark:border-white/5">
                  <AtSign class="w-4 h-4 text-purple-500" />
                  <div class="min-w-0">
                    <p class="text-[8px] font-black text-slate-400 uppercase tracking-tighter">EMAIL</p>
                    <p class="text-sm font-bold truncate dark:text-white">{{ user.email }}</p>
                  </div>
                </div>

                <div v-if="user.custom_contact" class="flex items-center gap-3 p-3 bg-slate-50 dark:bg-white/[0.02] rounded-xl border border-slate-100 dark:border-white/5">
                  <Info class="w-4 h-4 text-amber-500" />
                  <div class="min-w-0">
                    <p class="text-[8px] font-black text-slate-400 uppercase tracking-tighter">OTHER</p>
                    <p class="text-sm font-bold truncate dark:text-white">{{ user.custom_contact }}</p>
                  </div>
                </div>

                <div v-if="!user.wechat && !user.qq && !user.show_email && !user.custom_contact" class="text-center py-4">
                  <p class="text-[10px] text-slate-400 font-bold uppercase tracking-widest italic opacity-50">Empty_Buffer</p>
                </div>
              </div>
            </div>

            <!-- Membership Card -->
            <div class="bg-gradient-to-br from-slate-900 to-slate-800 dark:from-slate-800 dark:to-slate-900 rounded-[2rem] p-8 shadow-xl relative overflow-hidden group">
              <div class="absolute top-0 right-0 p-4 opacity-10 group-hover:scale-110 transition-transform">
                <Trophy class="w-20 h-20 text-blue-500" />
              </div>
              <h3 class="text-[10px] font-black uppercase tracking-[0.2em] text-blue-400 mb-6">Researcher_ID</h3>
              <div class="space-y-4 relative z-10">
                <div class="flex items-center justify-between">
                  <span class="text-[10px] text-slate-500 font-bold uppercase">Member Since</span>
                  <span class="text-xs font-mono text-white">{{ new Date(user.created_at).toLocaleDateString() }}</span>
                </div>
                <div class="h-px bg-white/5" />
                <div class="flex items-center justify-between">
                  <span class="text-[10px] text-slate-500 font-bold uppercase">Data Status</span>
                  <span class="text-[10px] text-emerald-400 font-black uppercase">Encrypted_OK</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
