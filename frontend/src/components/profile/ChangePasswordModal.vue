<script setup lang="ts">
import { ref } from 'vue'
import { Key, Lock, Eye, EyeOff, Loader2 } from 'lucide-vue-next'

defineProps<{
  show: boolean
  loading: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'save', oldPw: string, newPw: string): void
}>()

const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const showPasswords = ref(false)

const handleSave = () => {
  if (newPassword.value !== confirmPassword.value) {
    alert('两次输入的密码不一致')
    return
  }
  emit('save', oldPassword.value, newPassword.value)
}
</script>

<template>
  <div v-if="show" class="fixed inset-0 z-[100] flex items-center justify-center p-4 backdrop-blur-xl bg-black/60">
    <div class="bg-[#111114] border border-white/10 rounded-[3rem] p-10 max-w-md w-full shadow-2xl relative animate-in fade-in zoom-in duration-300">
      <h3 class="text-2xl font-black mb-8 italic uppercase text-center text-white">重置实验凭证 / Reset Key</h3>
      <form @submit.prevent="handleSave" class="space-y-5">
        <div class="space-y-4">
          <div class="relative group">
            <Key class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 group-focus-within:text-blue-500 transition-colors" />
            <input
              v-model="oldPassword"
              :type="showPasswords ? 'text' : 'password'"
              placeholder="当前密码 / Current Secret"
              class="w-full bg-white/5 border border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-4 outline-none transition-all text-white placeholder:text-slate-600 font-mono"
              required
            />
          </div>
          <div class="relative group">
            <Lock class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 group-focus-within:text-blue-500 transition-colors" />
            <input
              v-model="newPassword"
              :type="showPasswords ? 'text' : 'password'"
              placeholder="核准新密码 / New Authorized Key"
              class="w-full bg-white/5 border border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-12 outline-none transition-all text-white placeholder:text-slate-600 font-mono"
              required
            />
            <button 
              type="button"
              @click="showPasswords = !showPasswords"
              class="absolute right-4 top-1/2 -translate-y-1/2 text-slate-500 hover:text-white"
            >
              <EyeOff v-if="showPasswords" class="w-5 h-5" />
              <Eye v-else class="w-5 h-5" />
            </button>
          </div>
          <div class="relative group">
            <Lock class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 group-focus-within:text-blue-500 transition-colors" />
            <input
              v-model="confirmPassword"
              :type="showPasswords ? 'text' : 'password'"
              placeholder="再次输入新密码"
              class="w-full bg-white/5 border border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-4 outline-none transition-all text-white placeholder:text-slate-600 font-mono"
              required
            />
          </div>
        </div>
        <div class="flex gap-4 pt-4">
          <button 
            type="button"
            @click="$emit('close')" 
            class="flex-1 py-4 bg-white/5 hover:bg-white/10 rounded-2xl font-bold transition-all text-slate-400"
          >
            取消
          </button>
          <button 
            type="submit"
            :disabled="loading"
            class="flex-1 py-4 bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 rounded-2xl font-black text-white shadow-xl shadow-blue-500/20 disabled:opacity-50 flex items-center justify-center gap-2"
          >
            <Loader2 v-if="loading" class="w-5 h-5 animate-spin" />
            执行重置
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
