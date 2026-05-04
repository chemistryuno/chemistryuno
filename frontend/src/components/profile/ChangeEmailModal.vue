<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { Mail, Shield, Loader2, CheckCircle2, AlertTriangle } from 'lucide-vue-next'
import { authAPI } from '../../utils/api'
import { useDialog } from '../../utils/dialog'

const props = defineProps<{
  show: boolean
  currentEmail: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'success', newEmail: string): void
}>()

const dialog = useDialog()

const oldCode = ref('')
const newEmail = ref('')
const newCode = ref('')
const loading = ref(false)
const oldCountdown = ref(0)
const newCountdown = ref(0)

watch(() => props.show, (val) => {
  if (!val) {
    oldCode.value = ''
    newEmail.value = ''
    newCode.value = ''
    oldCountdown.value = 0
    newCountdown.value = 0
  }
  // 监控 show 状态以禁用/启用背景滚动
  if (val) {
    document.documentElement.style.overflow = 'hidden'
    document.body.style.overflow = 'hidden'
  } else {
    document.documentElement.style.overflow = ''
    document.body.style.overflow = ''
  }
})

const sendOldCode = async () => {
  try {
    await authAPI.sendCode(props.currentEmail, 'change_email_old')
    dialog.showAlert('验证码已发送至您的当前邮箱。', '发送成功')
    oldCountdown.value = 60
    const timer = setInterval(() => {
      oldCountdown.value--
      if (oldCountdown.value <= 0) clearInterval(timer)
    }, 1000)
  } catch (err: any) {
    dialog.showAlert(err.response?.data?.error || '发送失败', '探测失败')
  }
}

const sendNewCode = async () => {
  if (!newEmail.value) {
    dialog.showAlert('请输入新的电子邮箱地址', '警告')
    return
  }
  // 简单的邮箱格式校验
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(newEmail.value)) {
    dialog.showAlert('邮箱格式不正确', '警告')
    return
  }
  try {
    await authAPI.sendCode(newEmail.value, 'change_email_new')
    dialog.showAlert('验证码已发送至您的新邮箱，请查收。', '发送成功')
    newCountdown.value = 60
    const timer = setInterval(() => {
      newCountdown.value--
      if (newCountdown.value <= 0) clearInterval(timer)
    }, 1000)
  } catch (err: any) {
    dialog.showAlert(err.response?.data?.error || '发送失败', '探测失败')
  }
}

const handleSubmit = async () => {
  if (!oldCode.value || !newEmail.value || !newCode.value) {
    dialog.showAlert('请填写完整的验证信息', '提示')
    return
  }
  loading.value = true
  try {
    await authAPI.changeEmail({
      old_code: oldCode.value,
      new_email: newEmail.value,
      new_code: newCode.value
    })
    dialog.showAlert('研究员通讯地址已更新。', '变更成功')
    emit('success', newEmail.value)
    emit('close')
  } catch (err: any) {
    dialog.showAlert(err.response?.data?.error || '更新失败', '协议冲突')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <Teleport to="body">
  <div v-if="show" class="viewport-modal-overlay z-[100] p-4 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md">
    <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[3rem] p-10 max-w-md w-full shadow-2xl relative animate-in fade-in zoom-in duration-300">
      <div class="flex flex-col items-center mb-8">
        <div class="w-16 h-16 bg-blue-600/10 rounded-2xl flex items-center justify-center mb-4">
          <Mail class="w-8 h-8 text-blue-600 dark:text-blue-500" />
        </div>
        <h3 class="text-2xl font-black italic uppercase text-slate-900 dark:text-white tracking-tight text-center">
          变更通讯地址 / Email Change
        </h3>
        <p class="text-slate-500 text-[10px] font-black mt-2 uppercase tracking-[0.2em] font-mono">SECURE PROTOCOL REDIRECTION</p>
      </div>

      <!-- Security Notice -->
      <div class="mb-8 p-4 bg-orange-500/10 border border-orange-500/20 rounded-2xl flex gap-3 items-start">
        <AlertTriangle class="w-5 h-5 text-orange-600 shrink-0 mt-0.5" />
        <div class="text-[10px] text-orange-700 dark:text-orange-500 font-medium leading-relaxed">
          <p class="font-black mb-1 uppercase tracking-widest text-orange-800 dark:text-orange-400">凭证安全警告 / Safety Warning</p>
          邮箱是您找回账号的最高权限凭证。更换邮箱意味着转移了账号的控制权，请确保新邮箱为您本人私有，且已开启双重验证。严禁将账号转让或租借给他人。
        </div>
      </div>

      <div class="space-y-6">
        <!-- Step 1: Current Email Verification -->
        <div class="space-y-3">
          <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest block px-1">STEP 01: 验证当前地址</label>
          <div class="relative group">
            <Mail class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-300 dark:text-slate-600" />
            <input
              :value="currentEmail"
              disabled
              type="text"
              class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl py-4 pl-12 pr-4 outline-none text-slate-400 dark:text-slate-500 font-bold text-sm"
            />
          </div>
          <div class="relative group flex gap-2">
            <div class="relative flex-1">
              <Shield class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-600 dark:group-focus-within:text-blue-500 transition-colors" />
              <input
                v-model="oldCode"
                type="text"
                placeholder="原邮箱验证码"
                maxlength="6"
                class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-4 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 font-bold text-sm"
              />
            </div>
            <button 
              @click="sendOldCode"
              :disabled="oldCountdown > 0"
              class="px-4 rounded-2xl bg-blue-600/10 hover:bg-blue-600/20 text-blue-600 dark:text-blue-400 text-[10px] font-black uppercase tracking-widest transition-all disabled:opacity-50 min-w-[80px]"
            >
              {{ oldCountdown > 0 ? `${oldCountdown}s` : '获取' }}
            </button>
          </div>
        </div>

        <div class="h-px bg-slate-100 dark:bg-white/5 mx-4" />

        <!-- Step 2: New Email Verification -->
        <div class="space-y-3">
          <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest block px-1">STEP 02: 绑定新地址</label>
          <div class="relative group">
            <Mail class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-600 dark:group-focus-within:text-blue-500 transition-colors" />
            <input
              v-model="newEmail"
              type="email"
              placeholder="输入新的电子邮箱地址"
              class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-4 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 font-bold text-sm"
            />
          </div>
          <div class="relative group flex gap-2">
            <div class="relative flex-1">
              <Shield class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-600 dark:group-focus-within:text-blue-500 transition-colors" />
              <input
                v-model="newCode"
                type="text"
                placeholder="新邮箱验证码"
                maxlength="6"
                class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-4 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 font-bold text-sm"
              />
            </div>
            <button 
              @click="sendNewCode"
              :disabled="newCountdown > 0"
              class="px-4 rounded-2xl bg-blue-600/10 hover:bg-blue-600/20 text-blue-600 dark:text-blue-400 text-[10px] font-black uppercase tracking-widest transition-all disabled:opacity-50 min-w-[80px]"
            >
              {{ newCountdown > 0 ? `${newCountdown}s` : '获取' }}
            </button>
          </div>
        </div>

        <div class="flex gap-3 pt-4">
          <button 
            @click="emit('close')"
            class="flex-1 py-4 rounded-2xl bg-slate-100 dark:bg-white/5 text-slate-600 dark:text-slate-400 font-black uppercase text-[10px] tracking-widest hover:bg-slate-200 dark:hover:bg-white/10 transition-all"
          >
            取消 / Cancel
          </button>
          <button 
            @click="handleSubmit"
            :disabled="loading"
            class="flex-1 py-4 rounded-2xl bg-blue-600 hover:bg-blue-700 text-white font-black uppercase text-[10px] tracking-widest shadow-lg shadow-blue-500/25 transition-all flex items-center justify-center gap-2"
          >
            <Loader2 v-if="loading" class="w-3 h-3 animate-spin" />
            <CheckCircle2 v-else class="w-3 h-3" />
            执行同步 / Sync
          </button>
        </div>
      </div>
    </div>
  </div>
  </Teleport>
</template>
