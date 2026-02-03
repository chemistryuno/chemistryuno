<script setup lang="ts">
import { Lock, Shield, UserX, Loader2, Cpu, Smartphone } from 'lucide-vue-next'

defineProps<{
  twoFactorEnabled: boolean
  twoFactorLoading: boolean
}>()

defineEmits<{
  (e: 'changePassword'): void
  (e: 'setup2fa'): void
  (e: 'disable2fa'): void
  (e: 'manageHardwareKeys'): void
  (e: 'manageDevices'): void
  (e: 'deleteAccount'): void
}>()
</script>

<template>
  <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2.5rem] p-8 relative overflow-hidden shadow-sm dark:shadow-none">
    <div class="flex items-center justify-between mb-8">
      <div>
        <h3 class="text-xl font-black uppercase italic tracking-tighter flex items-center gap-3 text-slate-900 dark:text-white">
          <span class="w-1.5 h-6 bg-blue-600 rounded-full" />
          安全中心 / Security
        </h3>
        <p class="text-slate-500 text-xs mt-0.5">管理您的账户凭证与安全设置</p>
      </div>
    </div>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <button 
        @click="$emit('changePassword')"
        class="group relative flex items-center gap-4 p-4 bg-slate-50 dark:bg-white/5 hover:bg-blue-50 dark:hover:bg-blue-500/10 border border-slate-200 dark:border-white/5 hover:border-blue-300 dark:hover:border-blue-500/30 rounded-2xl transition-all text-left"
      >
        <div class="bg-blue-500/10 dark:bg-blue-500/20 p-2.5 rounded-xl group-hover:rotate-12 transition-transform shrink-0">
          <Lock class="w-5 h-5 text-blue-600 dark:text-blue-400" />
        </div>
        <div>
          <span class="text-sm font-bold text-slate-900 dark:text-white block">修改密码</span>
          <span class="text-slate-500 text-[10px]">定期更新以确保安全</span>
        </div>
      </button>

      <button 
        v-if="!twoFactorEnabled"
        @click="$emit('setup2fa')"
        :disabled="twoFactorLoading"
        class="group relative flex items-center gap-4 p-4 bg-slate-50 dark:bg-white/5 hover:bg-emerald-50 dark:hover:bg-emerald-500/10 border border-slate-200 dark:border-white/5 hover:border-emerald-300 dark:hover:border-emerald-500/30 rounded-2xl transition-all text-left"
      >
        <div class="bg-emerald-500/10 dark:bg-emerald-500/20 p-2.5 rounded-xl group-hover:rotate-12 transition-transform shrink-0">
          <Shield class="w-5 h-5 text-emerald-600 dark:text-emerald-400" />
        </div>
        <div>
          <span class="text-sm font-bold text-slate-900 dark:text-white block">双重验证 (2FA)</span>
          <span class="text-slate-500 text-[10px]">添加额外的安全保护层</span>
        </div>
        <Loader2 v-if="twoFactorLoading" class="absolute top-2 right-2 w-3 h-3 animate-spin text-emerald-500" />
      </button>

      <button 
        v-else
        @click="$emit('disable2fa')"
        :disabled="twoFactorLoading"
        class="group relative flex items-center gap-4 p-4 bg-emerald-50 dark:bg-emerald-500/5 hover:bg-red-50 dark:hover:bg-red-500/10 border border-emerald-200 dark:border-emerald-500/20 hover:border-red-300 dark:hover:border-red-500/30 rounded-2xl transition-all text-left"
      >
        <div class="bg-red-500/10 dark:bg-red-500/20 p-2.5 rounded-xl group-hover:rotate-12 transition-transform shrink-0">
          <Shield class="w-5 h-5 text-red-600" />
        </div>
        <div>
          <span class="text-sm font-bold text-emerald-600 group-hover:text-red-600 block">管理 2FA</span>
          <span class="text-slate-500 text-[10px]">保护模式已开启</span>
        </div>
        <div class="absolute top-2 right-2 w-2 h-2 bg-emerald-500 rounded-full animate-pulse" />
      </button>

      <button 
        @click="$emit('manageHardwareKeys')"
        class="group relative flex items-center gap-4 p-4 bg-slate-50 dark:bg-white/5 hover:bg-indigo-50 dark:hover:bg-indigo-500/10 border border-slate-200 dark:border-white/5 hover:border-indigo-300 dark:hover:border-indigo-500/30 rounded-2xl transition-all text-left"
      >
        <div class="bg-indigo-500/10 dark:bg-indigo-500/20 p-2.5 rounded-xl group-hover:rotate-12 transition-transform shrink-0">
          <Cpu class="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
        </div>
        <div>
          <span class="text-sm font-bold text-slate-900 dark:text-white block">硬件密钥</span>
          <span class="text-slate-500 text-[10px]">FIDO2 / 生物识别</span>
        </div>
      </button>

      <button 
        @click="$emit('manageDevices')"
        class="group relative flex items-center gap-4 p-4 bg-slate-50 dark:bg-white/5 hover:bg-amber-50 dark:hover:bg-amber-500/10 border border-slate-200 dark:border-white/5 hover:border-amber-300 dark:hover:border-amber-500/30 rounded-2xl transition-all text-left"
      >
        <div class="bg-amber-500/10 dark:bg-amber-500/20 p-2.5 rounded-xl group-hover:rotate-12 transition-transform shrink-0">
          <Smartphone class="w-5 h-5 text-amber-600 dark:text-amber-400" />
        </div>
        <div>
          <span class="text-sm font-bold text-slate-900 dark:text-white block">登录设备管理</span>
          <span class="text-slate-500 text-[10px]">活跃会话与设备管理</span>
        </div>
      </button>

      <button 
        @click="$emit('deleteAccount')"
        class="group relative flex items-center gap-4 p-4 bg-slate-50 dark:bg-white/5 hover:bg-red-50 dark:hover:bg-red-500/10 border border-slate-200 dark:border-white/5 hover:border-red-300 dark:hover:border-red-500/30 rounded-2xl transition-all text-left"
      >
        <div class="bg-red-500/10 dark:bg-red-500/20 p-2.5 rounded-xl group-hover:rotate-12 transition-transform shrink-0">
          <UserX class="w-5 h-5 text-red-600 dark:text-red-500" />
        </div>
        <div>
          <span class="text-sm font-bold text-red-600 dark:text-red-400 block">注销账号</span>
          <span class="text-slate-500 text-[10px]">删除所有实验室数据</span>
        </div>
      </button>
    </div>
  </div>
</template>
