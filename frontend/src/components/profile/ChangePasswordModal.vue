<script setup lang="ts">
import { ref, computed } from 'vue'
import { Key, Lock, Eye, EyeOff, Loader2, Fingerprint, Cpu, AlertTriangle, Mail } from 'lucide-vue-next'
import api, { authAPI } from '../../utils/api'
import { get } from '@github/webauthn-json'
import { onMounted, watch } from 'vue'
import { useDialog } from '../../utils/dialog'

const props = defineProps<{
  show: boolean
  loading: boolean
  is2faEnabled: boolean
  userEmail?: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'save', oldPw: string, newPw: string, code: string, useEmail: boolean): void
  (e: 'success'): void
}>()

const oldPassword = ref('')
const code = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const showPasswords = ref(false)
const localLoading = ref(false)
const hasWebauthnKeys = ref(false)
const useEmailMode = ref(false)
const sendingCode = ref(false)
const countdown = ref(0)
const timer = ref<any>(null)
const { showAlert } = useDialog()

// 监控 show 状态以禁用/启用背景滚动
watch(
  () => props.show,
  (show) => {
    if (show) {
      document.documentElement.style.overflow = 'hidden'
      document.body.style.overflow = 'hidden'
    } else {
      document.documentElement.style.overflow = ''
      document.body.style.overflow = ''
    }
  },
  { immediate: true }
)

const startCountdown = () => {
  countdown.value = 60
  timer.value = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearInterval(timer.value)
    }
  }, 1000)
}

const handleSendCode = async () => {
  if (!props.userEmail) {
    showAlert('未获取到您的邮箱，无法发送验证码', '发送失败')
    return
  }
  
  sendingCode.value = true
  try {
    await authAPI.sendCode(props.userEmail, 'change_password')
    startCountdown()
  } catch (err: any) {
    showAlert(err.response?.data?.error || '发送失败', '发送失败')
  } finally {
    sendingCode.value = false
  }
}

// Check if user has hardware keys
const checkKeys = async () => {
  try {
    const res = await authAPI.getWebAuthnCredentials()
    hasWebauthnKeys.value = res.data && res.data.length > 0
  } catch(e) {}
}

onMounted(() => {
  if (props.show) checkKeys()
})

watch(() => props.show, (val) => {
  if (val) {
    checkKeys()
  } else {
    // Reset form when modal closes
    oldPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
    code.value = ''
    useEmailMode.value = false
    if (timer.value) clearInterval(timer.value)
    countdown.value = 0
  }
})

const mode = computed(() => {
  if (hasWebauthnKeys.value) return 'webauthn'
  if (useEmailMode.value) return 'email'
  if (props.is2faEnabled) return '2fa'
  return 'classic'
})

const handleSave = () => {
  if (newPassword.value !== confirmPassword.value) {
    showAlert('两次输入的密码不一致', '校验失败')
    return
  }
  if (newPassword.value.length < 6) {
    showAlert('新密码长度至少为 6 位', '校验失败')
    return
  }
  emit('save', oldPassword.value, newPassword.value, code.value, useEmailMode.value)
}

const handleWebauthnReset = async () => {
  if (newPassword.value !== confirmPassword.value) {
    showAlert('两次输入的密码不一致', '校验失败')
    return
  }
  if (!newPassword.value || newPassword.value.length < 6) {
    showAlert('新密码长度至少为 6 位', '校验失败')
    return
  }

  localLoading.value = true
  try {
    const res = await authAPI.beginChangePasswordWebAuthn()
    const credential = await get(res.data)
    await authAPI.finishChangePasswordWebAuthn(newPassword.value, credential)
    emit('success')
    emit('close')
  } catch (err: any) {
    console.error(err)
    showAlert(err.response?.data?.error || '硬件验证失败', '验证失败')
  } finally {
    localLoading.value = false
  }
}
</script>

<template>
  <Teleport to="body">
  <div v-if="show" class="viewport-modal-overlay z-[100] p-4 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md">
    <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[3rem] p-10 max-w-md w-full shadow-2xl relative animate-in fade-in zoom-in duration-300">
      <div class="flex flex-col items-center mb-6">
          <div class="w-16 h-16 bg-blue-600/10 rounded-2xl flex items-center justify-center mb-4">
            <Cpu v-if="mode === 'webauthn'" class="w-8 h-8 text-blue-600 dark:text-blue-500" />
            <Fingerprint v-else-if="mode === '2fa'" class="w-8 h-8 text-blue-600 dark:text-blue-500" />
            <Mail v-else-if="mode === 'email'" class="w-8 h-8 text-blue-600 dark:text-blue-500" />
            <Key v-else class="w-8 h-8 text-blue-600 dark:text-blue-500" />
          </div>
          <h3 class="text-2xl font-black italic uppercase text-slate-900 dark:text-white tracking-tight">重置实验凭证</h3>
          <p class="text-slate-500 text-[10px] font-black mt-2 uppercase tracking-[0.2em] font-mono">
            {{ mode === 'webauthn' ? 'BY HARDWARE TOKEN' : mode === '2fa' ? 'BY AUTHENTICATOR APP' : mode === 'email' ? 'BY EMAIL CODE' : 'BY CLASSIC SECRET' }}
          </p>
      </div>

      <!-- Security Notice -->
      <div class="mb-6 p-4 bg-blue-500/5 border border-blue-500/10 rounded-2xl flex gap-3 items-start">
        <AlertTriangle class="w-5 h-5 text-blue-500 shrink-0 mt-0.5" />
        <div class="text-[10px] text-slate-500 font-medium leading-relaxed">
          <p class="font-black mb-1 uppercase tracking-widest text-slate-700 dark:text-slate-300">密码安全准则 / Password Policy</p>
          严禁使用过于简单的密码（如 123456 或生日）。建议结合字母、数字及符号，长度不少于 8 位。系统会自动记录异常登录尝试并锁定账户。
        </div>
      </div>

      <form @submit.prevent="handleSave" class="space-y-5">
        <div class="space-y-4">
          <!-- 2FA Enabled view -->
          <div v-if="mode === '2fa'" class="relative group">
            <Fingerprint class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-600 dark:group-focus-within:text-blue-500 transition-colors" />
            <input
              v-model="code"
              type="text"
              maxlength="6"
              placeholder="请输入 6 位 2FA 验证码"
              class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-4 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 font-mono tracking-[0.3em] text-center"
              required
            />
          </div>

          <!-- Email Code view -->
          <div v-else-if="mode === 'email'" class="space-y-3">
             <div class="relative group">
                <Mail class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-600 dark:group-focus-within:text-blue-500 transition-colors" />
                <input
                  v-model="code"
                  type="text"
                  maxlength="6"
                  placeholder="请输入邮箱验证码"
                  class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-32 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 font-mono"
                  required
                />
                <button 
                  type="button"
                  @click="handleSendCode"
                  :disabled="sendingCode || countdown > 0"
                  class="absolute right-2 top-1/2 -translate-y-1/2 px-4 py-2 bg-blue-600 hover:bg-blue-500 disabled:bg-slate-300 dark:disabled:bg-white/10 text-white text-[10px] font-black rounded-xl transition-all uppercase tracking-widest disabled:text-slate-500"
                >
                  {{ countdown > 0 ? `${countdown}S` : sendingCode ? '发送中...' : '发送验证码' }}
                </button>
             </div>
             <p class="text-[10px] text-slate-400 px-2">验证码将发送至：{{ props.userEmail }}</p>
          </div>
          
          <!-- Webauthn Info -->
          <div v-else-if="mode === 'webauthn'" class="bg-blue-600/5 dark:bg-blue-500/5 border border-blue-600/10 dark:border-blue-500/20 rounded-2xl p-4 mb-2 text-center">
            <p class="text-[10px] font-bold text-blue-600 dark:text-blue-400 uppercase tracking-widest leading-relaxed">
              检测到已绑定的硬件密钥<br/>将通过生物识别或物理令牌验证身份
            </p>
          </div>

          <!-- Classic view if NO 2FA and NO Webauthn -->
          <div v-else class="space-y-3">
            <div class="relative group">
              <Key class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-600 dark:group-focus-within:text-blue-500 transition-colors" />
              <input
                v-model="oldPassword"
                :type="showPasswords ? 'text' : 'password'"
                placeholder="当前密码 / Current Secret"
                class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-4 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 font-mono"
                required
              />
            </div>
            <div v-if="props.userEmail" @click="useEmailMode = true" class="text-right px-2">
               <button type="button" class="text-[10px] font-black text-blue-600 dark:text-blue-400 hover:underline uppercase tracking-widest">忘记密码？使用邮箱验证码</button>
            </div>
          </div>

          <div class="relative group">
            <Lock class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-600 dark:group-focus-within:text-blue-500 transition-colors" />
            <input
              v-model="newPassword"
              :type="showPasswords ? 'text' : 'password'"
              placeholder="核准新密码 / New Authorized Key"
              class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-12 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 font-mono"
              required
            />
            <button 
              type="button"
              @click="showPasswords = !showPasswords"
              class="absolute right-4 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-900 dark:text-slate-500 dark:hover:text-white transition-colors"
            >
              <EyeOff v-if="showPasswords" class="w-5 h-5" />
              <Eye v-else class="w-5 h-5" />
            </button>
          </div>
          <div class="relative group">
            <Lock class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 dark:text-slate-500 group-focus-within:text-blue-600 dark:group-focus-within:text-blue-500 transition-colors" />
            <input
              v-model="confirmPassword"
              :type="showPasswords ? 'text' : 'password'"
              placeholder="再次输入新密码"
              class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-4 outline-none transition-all text-slate-900 dark:text-white placeholder:text-slate-400 dark:placeholder:text-slate-600 font-mono"
              required
            />
          </div>
        </div>

        <div class="flex gap-4 pt-4">
          <button 
            type="button"
            @click="$emit('close')" 
            class="flex-1 py-4 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 rounded-2xl font-bold transition-all text-slate-500 dark:text-slate-400"
          >
            取消
          </button>
          
          <button 
            v-if="mode === 'webauthn'"
            type="button"
            @click="handleWebauthnReset"
            :disabled="localLoading"
            class="flex-1 py-4 bg-gradient-to-r from-emerald-600 to-emerald-500 hover:from-emerald-500 hover:to-emerald-400 rounded-2xl font-black text-white shadow-xl shadow-emerald-500/20 disabled:opacity-50 flex items-center justify-center gap-2"
          >
            <Loader2 v-if="localLoading" class="w-5 h-5 animate-spin" />
            调起硬件验证
          </button>

          <button 
            v-else
            type="submit"
            :disabled="loading"
            class="flex-1 py-4 bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 rounded-2xl font-black text-white shadow-xl shadow-blue-500/20 disabled:opacity-50 flex items-center justify-center gap-2"
          >
            <Loader2 v-if="loading" class="w-5 h-5 animate-spin" />
            确认修改
          </button>
        </div>
      </form>
    </div>
  </div>
  </Teleport>
</template>

