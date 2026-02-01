<script setup lang="ts">
import { Lock, Shield, UserX, Loader2 } from 'lucide-vue-next'

defineProps<{
  twoFactorEnabled: boolean
  twoFactorLoading: boolean
}>()

defineEmits<{
  (e: 'changePassword'): void
  (e: 'setup2fa'): void
  (e: 'disable2fa'): void
  (e: 'deleteAccount'): void
}>()
</script>

<template>
  <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2.5rem] p-10 relative overflow-hidden shadow-sm dark:shadow-none">
    <div class="flex items-center justify-between mb-10">
      <div>
        <h3 class="text-2xl font-black uppercase italic tracking-tighter flex items-center gap-3 text-slate-900 dark:text-white">
          <span class="w-2 h-8 bg-blue-600 rounded-full" />
          账户安全管理 / Security
        </h3>
        <p class="text-slate-500 text-sm mt-1">维护您的研究员凭证与实验室访问权限</p>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <button 
        @click="$emit('changePassword')"
        class="group relative flex flex-col items-start p-6 bg-slate-50 dark:bg-white/5 hover:bg-blue-50 dark:hover:bg-blue-500/10 border border-slate-200 dark:border-white/5 hover:border-blue-300 dark:hover:border-blue-500/30 rounded-3xl transition-all text-left"
      >
        <div class="bg-blue-500/10 dark:bg-blue-500/20 p-3 rounded-2xl mb-4 group-hover:scale-110 transition-transform">
          <Lock class="w-6 h-6 text-blue-600 dark:text-blue-400" />
        </div>
        <span class="text-lg font-bold text-slate-900 dark:text-white">修改研究密码</span>
        <span class="text-slate-500 text-xs mt-1">更新安全凭证以确保实验室数据安全</span>
      </button>

      <button 
        v-if="!twoFactorEnabled"
        @click="$emit('setup2fa')"
        :disabled="twoFactorLoading"
        class="group relative flex flex-col items-start p-6 bg-slate-50 dark:bg-white/5 hover:bg-emerald-50 dark:hover:bg-emerald-500/10 border border-slate-200 dark:border-white/5 hover:border-emerald-300 dark:hover:border-emerald-500/30 rounded-3xl transition-all text-left"
      >
        <div class="bg-emerald-500/10 dark:bg-emerald-500/20 p-3 rounded-2xl mb-4 group-hover:scale-110 transition-transform">
          <Shield class="w-6 h-6 text-emerald-600 dark:text-emerald-400" />
        </div>
        <span class="text-lg font-bold text-slate-900 dark:text-white">开启双重验证</span>
        <span class="text-slate-500 text-xs mt-1">通过 TOTP 协议为您的账户增加第二层保护</span>
        <Loader2 v-if="twoFactorLoading" class="absolute top-6 right-6 w-5 h-5 animate-spin text-emerald-500" />
      </button>

      <button 
        v-else
        @click="$emit('disable2fa')"
        :disabled="twoFactorLoading"
        class="group relative flex flex-col items-start p-6 bg-emerald-50 dark:bg-emerald-500/5 hover:bg-red-50 dark:hover:bg-red-500/10 border border-emerald-200 dark:border-emerald-500/20 hover:border-red-300 dark:hover:border-red-500/30 rounded-3xl transition-all text-left"
      >
        <div class="bg-red-500/10 dark:bg-red-500/20 p-3 rounded-2xl mb-4 group-hover:scale-110 transition-transform">
          <Shield class="w-6 h-6 text-red-600" />
        </div>
        <span class="text-lg font-bold text-emerald-600 group-hover:text-red-600">管理双重验证</span>
        <span class="text-slate-500 text-xs mt-1">2FA 已激活。点击可停用验证保护。</span>
        <div class="absolute top-6 right-6 w-3 h-3 bg-emerald-500 rounded-full animate-pulse shadow-[0_0_10px_rgba(16,185,129,0.5)]" />
      </button>

      <button 
        @click="$emit('deleteAccount')"
        class="group relative flex flex-col items-start p-6 bg-slate-50 dark:bg-white/5 hover:bg-red-50 dark:hover:bg-red-500/10 border border-slate-200 dark:border-white/5 hover:border-red-300 dark:hover:border-red-500/30 rounded-3xl transition-all text-left"
      >
        <div class="bg-red-500/10 dark:bg-red-500/20 p-3 rounded-2xl mb-4 group-hover:scale-110 transition-transform">
          <UserX class="w-6 h-6 text-red-600 dark:text-red-500" />
        </div>
        <span class="text-lg font-bold text-red-600 dark:text-red-400">注销席位</span>
        <span class="text-slate-500 text-xs mt-1">永久注销账户并清除所有研究 data</span>
      </button>
    </div>
  </div>
</template>
