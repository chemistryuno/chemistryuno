<script setup lang="ts">
import { Shield, Fingerprint, Calendar, Award, User as UserIcon, RefreshCw, Zap } from 'lucide-vue-next'

defineProps<{
  user: any
}>()

defineEmits<{
  (e: 'changeAvatar'): void
}>()
</script>

<template>
  <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2.5rem] p-6 relative overflow-hidden group shadow-2xl transition-all hover:shadow-[0_20px_50px_rgba(0,0,0,0.1)] dark:hover:shadow-[0_20px_50px_rgba(0,0,0,0.4)]">
    <div class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-transparent via-blue-500/50 to-transparent" />
    
    <div class="flex flex-col items-center">
      <div class="relative group/avatar mb-6">
        <div class="w-32 h-32 bg-gradient-to-tr from-slate-200 to-slate-100 dark:from-[#1a1c1e] dark:to-[#2d3035] rounded-[2.5rem] p-1 shadow-2xl transition-transform duration-500 group-hover/avatar:scale-105">
          <div class="w-full h-full bg-white dark:bg-[#111114] rounded-[2.3rem] flex items-center justify-center text-6xl relative overflow-hidden group/inner transition-all border border-slate-200 dark:border-white/5">
            <div class="absolute inset-0 bg-blue-500/5 opacity-0 group-hover/inner:opacity-100 transition-opacity" />
            <template v-if="user.avatar && user.avatar.startsWith('data:')">
               <img :src="user.avatar" class="w-full h-full object-cover relative z-10" />
            </template>
            <template v-else>
               <span class="relative z-10 scale-110 drop-shadow-[0_0_15px_rgba(0,0,0,0.1)] dark:drop-shadow-[0_0_15px_rgba(255,255,255,0.3)]">{{ user.avatar || '🧪' }}</span>
            </template>
          </div>
        </div>
        <button 
          @click="$emit('changeAvatar')"
          class="absolute -bottom-1 -right-1 bg-blue-600 hover:bg-blue-500 p-2.5 rounded-2xl shadow-[0_0_20px_rgba(37,99,235,0.2)] dark:shadow-[0_0_20px_rgba(37,99,235,0.4)] z-20 group-hover:rotate-12 transition-all active:scale-95"
          title="更改研究员原型"
        >
          <RefreshCw class="w-4 h-4 text-white" />
        </button>
      </div>

      <div class="text-center space-y-1 w-full">
        <div class="flex items-center justify-center gap-1.5 mb-0.5">
          <UserIcon class="w-3.5 h-3.5 text-blue-500 opacity-50" />
          <span class="text-[9px] font-mono text-slate-400 dark:text-slate-500 uppercase tracking-widest">ID: {{ user.uid }}</span>
        </div>
        <h2 class="text-2xl font-black tracking-tight text-slate-900 dark:text-white group-hover:text-blue-500 transition-colors uppercase italic truncate px-2">
          {{ user.username }}
        </h2>
        <div class="flex items-center justify-center gap-2 pt-1">
          <span v-if="user.is_admin" class="bg-blue-500/10 text-blue-600 dark:text-blue-400 text-[9px] font-black px-3 py-1 rounded-full border border-blue-500/20 flex items-center gap-1.5 uppercase tracking-[0.15em]">
            <Shield class="w-2.5 h-2.5" /> CORE ADMIN 
          </span>
          <span v-else class="bg-slate-500/10 text-slate-600 dark:text-slate-400 text-[9px] font-black px-3 py-1 rounded-full border border-slate-500/20 flex items-center gap-1.5 uppercase tracking-[0.15em]">
            <Fingerprint class="w-2.5 h-2.5" /> RESEARCHER
          </span>
        </div>
      </div>

      <div class="w-full mt-6 pt-6 border-t border-slate-200 dark:border-white/5 space-y-3">
        <div class="flex justify-between items-center text-[11px]">
          <span class="text-slate-500 font-bold uppercase tracking-widest flex items-center gap-1.5"><Award class="w-2.5 h-2.5" /> Exp Level</span>
          <div class="flex items-center gap-2">
            <div class="w-20 h-1.5 bg-slate-200 dark:bg-white/5 rounded-full overflow-hidden">
              <div class="w-1/3 h-full bg-blue-500" />
            </div>
            <span class="text-blue-600 dark:text-blue-500 font-black">LV.01</span>
          </div>
        </div>
        <div class="flex justify-between items-center text-[11px]">
          <span class="text-slate-500 font-bold uppercase tracking-widest flex items-center gap-1.5"><Zap class="w-2.5 h-2.5 text-yellow-500" /> Points</span>
          <span class="font-black text-slate-900 dark:text-white uppercase font-mono">{{ user.points || 0 }}</span>
        </div>
        <div v-if="user.created_at" class="flex justify-between items-center text-[11px]">
          <span class="text-slate-500 font-bold uppercase tracking-widest flex items-center gap-1.5"><Calendar class="w-2.5 h-2.5" /> Since</span>
          <span class="font-mono text-slate-500 dark:text-slate-400">{{ new Date(user.created_at).toLocaleDateString() }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
