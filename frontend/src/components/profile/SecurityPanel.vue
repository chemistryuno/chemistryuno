<script setup lang="ts">
import { Lock, Shield, UserX, Loader2, Cpu, Smartphone, Mail } from 'lucide-vue-next'

defineProps<{
  twoFactorEnabled: boolean
  twoFactorLoading: boolean
  smtpEnabled: boolean
}>()

defineEmits<{
  (e: 'changePassword'): void
  (e: 'changeEmail'): void
  (e: 'setup2fa'): void
  (e: 'disable2fa'): void
  (e: 'manageHardwareKeys'): void
  (e: 'manageDevices'): void
  (e: 'deleteAccount'): void
}>()
</script>

<template>
  <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-2xl p-6 relative overflow-hidden shadow-sm dark:shadow-none">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h3 class="text-base font-black uppercase italic tracking-tighter flex items-center gap-2.5 text-slate-800 dark:text-white leading-none">
          <span class="w-1 h-5 bg-blue-600 rounded-full" />
          安全中心 <span class="text-[10px] font-mono opacity-30">/ SECURITY</span>
        </h3>
        <p class="text-slate-500 text-[11px] mt-1">管理您的账户凭证与安全设置</p>
      </div>
    </div>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
      <button 
        @click="$emit('changePassword')"
        class="group relative flex items-center gap-3.5 p-3.5 bg-slate-50 dark:bg-white/5 hover:bg-blue-50 dark:hover:bg-blue-500/10 border border-slate-200 dark:border-white/5 hover:border-blue-300 dark:hover:border-blue-500/30 rounded-xl transition-all text-left"
      >
        <div class="bg-blue-500/10 dark:bg-blue-500/20 p-2 rounded-lg group-hover:rotate-12 transition-transform shrink-0 outline outline-1 outline-blue-500/10">
          <Lock class="w-4 h-4 text-blue-600 dark:text-blue-400" />
        </div>
        <div>
          <span class="text-xs font-black text-slate-800 dark:text-white block uppercase tracking-tight">修改密码</span>
          <span class="text-slate-400 text-[9px] uppercase font-mono">ENCRYPT_UPDATE</span>
        </div>
      </button>

      <button 
        v-if="smtpEnabled"
        @click="$emit('changeEmail')"
        class="group relative flex items-center gap-3.5 p-3.5 bg-slate-50 dark:bg-white/5 hover:bg-orange-50 dark:hover:bg-orange-500/10 border border-slate-200 dark:border-white/5 hover:border-orange-300 dark:hover:border-orange-500/30 rounded-xl transition-all text-left"
      >
        <div class="bg-orange-500/10 dark:bg-orange-500/20 p-2 rounded-lg group-hover:rotate-12 transition-transform shrink-0 outline outline-1 outline-orange-500/10">
          <Mail class="w-4 h-4 text-orange-600 dark:text-orange-400" />
        </div>
        <div>
          <span class="text-xs font-black text-slate-800 dark:text-white block uppercase tracking-tight">重置通讯邮箱</span>
          <span class="text-slate-400 text-[9px] uppercase font-mono">SMTP_RESET</span>
        </div>
      </button>

      <button 
        v-if="!twoFactorEnabled"
        @click="$emit('setup2fa')"
        :disabled="twoFactorLoading"
        class="group relative flex items-center gap-3.5 p-3.5 bg-slate-50 dark:bg-white/5 hover:bg-emerald-50 dark:hover:bg-emerald-500/10 border border-slate-200 dark:border-white/5 hover:border-emerald-300 dark:hover:border-emerald-500/30 rounded-xl transition-all text-left"
      >
        <div class="bg-emerald-500/10 dark:bg-emerald-500/20 p-2 rounded-lg group-hover:rotate-12 transition-transform shrink-0 outline outline-1 outline-emerald-500/10">
          <Shield class="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
        </div>
        <div>
          <span class="text-xs font-black text-slate-800 dark:text-white block uppercase tracking-tight">开启双重验证</span>
          <span class="text-slate-400 text-[9px] uppercase font-mono">2FA_PROTOCOL</span>
        </div>
        <Loader2 v-if="twoFactorLoading" class="absolute top-2 right-2 w-3 h-3 animate-spin text-emerald-500" />
      </button>

      <button 
        v-else
        @click="$emit('disable2fa')"
        :disabled="twoFactorLoading"
        class="group relative flex items-center gap-3.5 p-3.5 bg-emerald-500/5 dark:bg-emerald-500/5 hover:bg-red-50 dark:hover:bg-red-500/10 border border-emerald-500/20 hover:border-red-500/30 rounded-xl transition-all text-left border-dashed"
      >
        <div class="bg-red-500/10 dark:bg-red-500/20 p-2 rounded-lg group-hover:rotate-12 transition-transform shrink-0 outline outline-1 outline-red-500/10">
          <Shield class="w-4 h-4 text-red-600 dark:text-red-500" />
        </div>
        <div>
          <span class="text-xs font-black text-emerald-600 dark:text-emerald-400 group-hover:text-red-600 block uppercase tracking-tight leading-none mb-1">2FA_ACTIVE</span>
          <span class="text-[9px] text-slate-400 font-mono tracking-tighter">保护模式已就绪</span>
        </div>
        <div class="absolute top-2 right-2 w-2 h-2 bg-emerald-500 rounded-full animate-pulse" />
      </button>

      <button 
        @click="$emit('manageHardwareKeys')"
        class="group relative flex items-center gap-3.5 p-3.5 bg-slate-50 dark:bg-white/5 hover:bg-indigo-50 dark:hover:bg-indigo-500/10 border border-slate-200 dark:border-white/5 hover:border-indigo-300 dark:hover:border-indigo-500/30 rounded-xl transition-all text-left"
      >
        <div class="bg-indigo-500/10 dark:bg-indigo-500/20 p-2 rounded-lg group-hover:rotate-12 transition-transform shrink-0 outline outline-1 outline-indigo-500/10">
          <Cpu class="w-4 h-4 text-indigo-600 dark:text-indigo-400" />
        </div>
        <div>
          <span class="text-xs font-black text-slate-800 dark:text-white block uppercase tracking-tight">硬件密钥</span>
          <span class="text-slate-400 text-[9px] uppercase font-mono">FIDO2_WEB_AUTHN</span>
        </div>
      </button>

      <button 
        @click="$emit('manageDevices')"
        class="group relative flex items-center gap-3.5 p-3.5 bg-slate-50 dark:bg-white/5 hover:bg-amber-50 dark:hover:bg-amber-500/10 border border-slate-200 dark:border-white/5 hover:border-amber-300 dark:hover:border-amber-500/30 rounded-xl transition-all text-left"
      >
        <div class="bg-amber-500/10 dark:bg-amber-500/20 p-2 rounded-lg group-hover:rotate-12 transition-transform shrink-0 outline outline-1 outline-amber-500/10">
          <Smartphone class="w-4 h-4 text-amber-600 dark:text-amber-400" />
        </div>
        <div>
          <span class="text-xs font-black text-slate-800 dark:text-white block uppercase tracking-tight">会话管理</span>
          <span class="text-slate-400 text-[9px] uppercase font-mono">DEVICE_LINKAGE</span>
        </div>
      </button>

      <button 
        @click="$emit('deleteAccount')"
        class="group relative flex items-center gap-3.5 p-3.5 bg-slate-50 dark:bg-white/5 hover:bg-red-50 dark:hover:bg-red-500/10 border border-slate-200 dark:border-white/5 hover:border-red-300 dark:hover:border-red-500/30 rounded-xl transition-all text-left"
      >
        <div class="bg-red-500/10 dark:bg-red-500/20 p-2 rounded-lg group-hover:rotate-12 transition-transform shrink-0 outline outline-1 outline-red-500/10">
          <UserX class="w-4 h-4 text-red-600 dark:text-red-500" />
        </div>
        <div>
          <span class="text-xs font-black text-red-600 dark:text-red-400 block uppercase tracking-tight leading-none mb-1">注销账户</span>
          <span class="text-[9px] text-slate-400 font-mono tracking-tighter">RESET_ALL_DATA</span>
        </div>
      </button>
    </div>
  </div>
</template>
