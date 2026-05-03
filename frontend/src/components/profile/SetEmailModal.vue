<script setup lang="ts">
import { ref, watch } from 'vue'
import { Mail, Shield, Loader2, HelpCircle } from 'lucide-vue-next'
import { authAPI } from '../../utils/api'
import { useDialog } from '../../utils/dialog'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'success', newEmail: string): void
}>()

const dialog = useDialog()
const newEmail = ref('')
const newCode = ref('')
const securityAnswer = ref('')
const loading = ref(false)
const codeLoading = ref(false)
const countdown = ref(0)
const error = ref('')

// 安全状态（在modal显示时加载）
const hasSecurityQuestion = ref(false)
const securityQuestion = ref('')

const loadSecurityInfo = async () => {
  try {
    const res = await authAPI.getMySecurityQuestion()
    hasSecurityQuestion.value = res.data.has_security_question
    securityQuestion.value = res.data.security_question || ''
  } catch {
    // ignore
  }
}

// 当modal显示时加载
watch(() => props.show, (val) => {
  if (val) {
    error.value = ''
    newEmail.value = ''
    newCode.value = ''
    securityAnswer.value = ''
    loadSecurityInfo()
    // 禁用背景滚动
    document.documentElement.style.overflow = 'hidden'
    document.body.style.overflow = 'hidden'
  } else {
    // 启用背景滚动
    document.documentElement.style.overflow = ''
    document.body.style.overflow = ''
  }
})

const handleSendCode = async () => {
  if (!newEmail.value || !newEmail.value.includes('@')) {
    error.value = '请输入有效的邮箱地址'
    return
  }
  codeLoading.value = true
  try {
    await authAPI.sendCode(newEmail.value, 'change_email_new')
    dialog.showAlert('验证码已发送至新邮箱，请查收。', '发送成功')
    countdown.value = 60
    const timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) clearInterval(timer)
    }, 1000)
  } catch (err: any) {
    error.value = err.response?.data?.error || '验证码发送失败'
  } finally {
    codeLoading.value = false
  }
}

const handleSubmit = async () => {
  error.value = ''
  if (!newEmail.value || !newEmail.value.includes('@')) {
    error.value = '请输入有效的邮箱地址'
    return
  }

  loading.value = true
  try {
    await authAPI.setEmail({
      new_email: newEmail.value,
      new_code: newCode.value,
      security_answer: hasSecurityQuestion.value ? securityAnswer.value : undefined,
    })
    emit('success', newEmail.value)
    emit('close')
    dialog.showAlert('邮箱绑定成功！', '操作成功')
  } catch (err: any) {
    const errMsg = err.response?.data?.error || ''
    if (errMsg.includes('security answer required') || errMsg.includes('incorrect security answer')) {
      error.value = '密保答案错误'
    } else if (errMsg.includes('already registered')) {
      error.value = '该邮箱已被其他账号占用'
    } else if (errMsg.includes('verification code')) {
      error.value = '邮箱验证码错误或已过期'
    } else {
      error.value = errMsg || '绑定失败，请重试'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div v-if="show" class="viewport-modal-overlay z-[100] p-4 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md">
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-white/10 rounded-2xl p-6 max-w-md w-full shadow-2xl animate-in fade-in zoom-in duration-300">
      <div class="flex items-center gap-3 mb-5">
        <div class="w-10 h-10 bg-cyan-500/10 rounded-xl flex items-center justify-center">
          <Mail class="w-5 h-5 text-cyan-600 dark:text-cyan-400" />
        </div>
        <div>
          <h3 class="text-base font-black text-slate-900 dark:text-white">绑定邮箱</h3>
          <p class="text-[10px] text-slate-400 uppercase tracking-widest font-mono">BIND_EMAIL</p>
        </div>
      </div>

      <div v-if="error" class="mb-4 bg-red-50 dark:bg-red-500/10 border border-red-100 dark:border-red-500/20 text-red-500 px-3 py-2 rounded-xl text-xs font-bold">
        {{ error }}
      </div>

      <form @submit.prevent="handleSubmit" class="space-y-3">
        <!-- 新邮箱 + 验证码 -->
        <div class="flex gap-2">
          <div class="relative flex-1 group">
            <Mail class="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-400 group-focus-within:text-cyan-500 transition-colors" />
            <input
              v-model="newEmail"
              type="email"
              required
              placeholder="新邮箱地址"
              class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-9 pr-3 py-2.5 rounded-xl focus:ring-2 focus:ring-cyan-500/20 focus:border-cyan-500 outline-none transition-all text-xs sm:text-sm font-bold placeholder:text-slate-500/50"
            />
          </div>
        </div>

        <div class="flex gap-2">
          <div class="relative flex-1 group">
            <Shield class="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-400 group-focus-within:text-cyan-500 transition-colors" />
            <input
              v-model="newCode"
              type="text"
              required
              maxlength="6"
              placeholder="邮箱验证码"
              class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-slate-100 pl-9 pr-3 py-2.5 rounded-xl focus:ring-2 focus:ring-cyan-500/20 focus:border-cyan-500 outline-none transition-all text-xs sm:text-sm font-bold placeholder:text-slate-500/50"
            />
          </div>
          <button
            type="button"
            @click="handleSendCode"
            :disabled="codeLoading || countdown > 0"
            class="px-3 rounded-xl font-black text-[10px] uppercase tracking-widest transition-all bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 hover:bg-cyan-500/20 disabled:opacity-50 border border-cyan-500/20 min-w-[60px]"
          >
            {{ countdown > 0 ? `${countdown}S` : (codeLoading ? '...' : '发送') }}
          </button>
        </div>

        <!-- 密保验证（若有密保问题） -->
        <div v-if="hasSecurityQuestion" class="space-y-2 p-3 bg-amber-50 dark:bg-amber-500/10 border border-amber-200 dark:border-amber-500/20 rounded-xl">
          <p class="text-[10px] font-black text-amber-600 dark:text-amber-400 uppercase tracking-widest flex items-center gap-1">
            <HelpCircle class="w-3 h-3" />
            密保验证
          </p>
          <p class="text-xs font-bold text-slate-700 dark:text-slate-300">{{ securityQuestion }}</p>
          <input
            v-model="securityAnswer"
            type="text"
            required
            placeholder="密保答案"
            class="w-full bg-white dark:bg-black/20 border border-amber-200 dark:border-amber-500/30 text-slate-900 dark:text-slate-100 px-3 py-2 rounded-lg focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500 outline-none transition-all text-xs sm:text-sm font-bold placeholder:text-slate-500/50"
          />
        </div>

        <div class="flex gap-2 pt-1">
          <button
            type="button"
            @click="$emit('close')"
            class="flex-1 py-2.5 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 border border-slate-200 dark:border-white/5 rounded-xl font-bold text-xs text-slate-500 dark:text-slate-400 transition-all uppercase tracking-widest"
          >
            取消
          </button>
          <button
            type="submit"
            :disabled="loading"
            class="flex-[2] py-2.5 bg-cyan-600 hover:bg-cyan-700 disabled:bg-slate-300 dark:disabled:bg-slate-700 text-white rounded-xl font-black text-xs transition-all shadow-lg shadow-cyan-500/20 flex items-center justify-center gap-1.5 uppercase tracking-widest"
          >
            <Loader2 v-if="loading" class="w-3 h-3 animate-spin" />
            {{ loading ? '绑定中...' : '确认绑定' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
