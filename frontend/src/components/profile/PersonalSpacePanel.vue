<script setup lang="ts">
import { ref, watch } from 'vue'
import { User, MessageCircle, AtSign, Info, Eye, EyeOff, Hash } from 'lucide-vue-next'
import { useDialog } from '../../utils/dialog'
import { authAPI } from '../../utils/api'

const props = defineProps<{
  user: any
}>()

const emit = defineEmits<{
  (e: 'update'): void
}>()

const { showAlert } = useDialog()
const loading = ref(false)

const form = ref({
  nickname: props.user.nickname || '',
  bio: props.user.bio || '',
  wechat: props.user.wechat || '',
  qq: props.user.qq || '',
  show_email: props.user.show_email || false,
  custom_contact: props.user.custom_contact || ''
})

// 监听 props 变化，确保数据同步
watch(() => props.user, (newUser) => {
  form.value = {
    nickname: newUser.nickname || '',
    bio: newUser.bio || '',
    wechat: newUser.wechat || '',
    qq: newUser.qq || '',
    show_email: newUser.show_email || false,
    custom_contact: newUser.custom_contact || ''
  }
}, { deep: true })

const handleSave = async () => {
  if (!form.value.nickname) {
    showAlert('昵称不能为空', '校验失败')
    return
  }
  
  loading.value = true
  try {
    await authAPI.updateProfile(form.value)
    showAlert('个人空间资料已更新。', '变更成功')
    emit('update')
  } catch (error: any) {
    showAlert(error.response?.data?.error || '更新失败', '错误')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-2xl p-6 shadow-sm">
    <h3 class="text-base font-black uppercase tracking-widest mb-6 flex items-center gap-2.5 text-slate-800 dark:text-white">
      <User class="w-4 h-4 text-blue-500" />
      个人空间设置 <span class="text-[10px] font-mono opacity-30">/ PERSONAL_SPACE</span>
    </h3>

    <div class="space-y-5">
      <!-- 昵称 -->
      <div class="space-y-2">
        <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">昵称 / Nickname</label>
        <div class="relative group">
          <div class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
            <User class="h-4 w-4 text-slate-400 group-focus-within:text-blue-500 transition-colors" />
          </div>
          <input 
            v-model="form.nickname"
            type="text" 
            placeholder="你的研究员昵称"
            class="block w-full pl-10 pr-4 py-2.5 bg-slate-50 dark:bg-white/[0.02] border border-slate-200 dark:border-white/5 rounded-xl text-sm transition-all focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none"
          />
        </div>
      </div>

      <!-- 自我介绍 -->
      <div class="space-y-2">
        <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">自我介绍 / Bio</label>
        <textarea 
          v-model="form.bio"
          rows="3"
          placeholder="介绍一下你自己..."
          class="block w-full px-4 py-2.5 bg-slate-50 dark:bg-white/[0.02] border border-slate-200 dark:border-white/5 rounded-xl text-sm transition-all focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none resize-none"
        ></textarea>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <!-- 微信 -->
        <div class="space-y-2">
          <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">微信 / WeChat</label>
          <div class="relative group">
            <div class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
              <MessageCircle class="h-4 w-4 text-slate-400 group-focus-within:text-emerald-500 transition-colors" />
            </div>
            <input 
              v-model="form.wechat"
              type="text" 
              placeholder="WeChat ID"
              class="block w-full pl-10 pr-4 py-2.5 bg-slate-50 dark:bg-white/[0.02] border border-slate-200 dark:border-white/5 rounded-xl text-sm transition-all focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500 outline-none"
            />
          </div>
        </div>

        <!-- QQ -->
        <div class="space-y-2">
          <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">QQ</label>
          <div class="relative group">
            <div class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
              <Hash class="h-4 w-4 text-slate-400 group-focus-within:text-blue-500 transition-colors" />
            </div>
            <input 
              v-model="form.qq"
              type="text" 
              placeholder="QQ 号码"
              class="block w-full pl-10 pr-4 py-2.5 bg-slate-50 dark:bg-white/[0.02] border border-slate-200 dark:border-white/5 rounded-xl text-sm transition-all focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none"
            />
          </div>
        </div>
      </div>

      <!-- 邮箱展示开关 -->
      <div class="p-4 bg-slate-50 dark:bg-white/[0.02] border border-slate-200 dark:border-white/5 rounded-2xl">
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 bg-blue-500/10 rounded-lg flex items-center justify-center text-blue-500">
              <AtSign class="w-4 h-4" />
            </div>
            <div>
              <p class="text-[10px] font-black uppercase tracking-wider text-slate-800 dark:text-white leading-none">展示注册邮箱</p>
              <p class="text-[9px] text-slate-400 mt-1 uppercase">Show registered email</p>
            </div>
          </div>
          <button 
            @click="form.show_email = !form.show_email"
            :class="[
              'relative inline-flex h-5 w-10 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2',
              form.show_email ? 'bg-blue-600' : 'bg-slate-200 dark:bg-white/10'
            ]"
          >
            <span 
              :class="[
                'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                form.show_email ? 'translate-x-5' : 'translate-x-0'
              ]"
            />
          </button>
        </div>
        <div class="flex items-center gap-2 px-3 py-2 bg-white dark:bg-[#1a1a1e] border border-slate-200 dark:border-white/5 rounded-xl">
          <component :is="form.show_email ? Eye : EyeOff" class="w-3.5 h-3.5 text-slate-400" />
          <span class="text-xs font-mono" :class="form.show_email ? 'text-slate-600 dark:text-slate-300' : 'text-slate-400'">
            {{ form.show_email ? user.email : '••••••••••••••••' }}
          </span>
        </div>
      </div>

      <!-- 随意填写的栏目 -->
      <div class="space-y-2">
        <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">其他联系方式 / Other</label>
        <div class="relative group">
          <div class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
            <Info class="h-4 w-4 text-slate-400 group-focus-within:text-purple-500 transition-colors" />
          </div>
          <input 
            v-model="form.custom_contact"
            type="text" 
            placeholder="任意你想展示的信息 (如 Telegram, Twitter 等)"
            class="block w-full pl-10 pr-4 py-2.5 bg-slate-50 dark:bg-white/[0.02] border border-slate-200 dark:border-white/5 rounded-xl text-sm transition-all focus:ring-2 focus:ring-purple-500/20 focus:border-purple-500 outline-none"
          />
        </div>
      </div>

      <div class="pt-2">
        <button 
          @click="handleSave"
          :disabled="loading"
          class="w-full flex items-center justify-center gap-2 px-6 py-3 bg-blue-600 hover:bg-blue-500 disabled:bg-blue-600/50 text-white rounded-xl font-black text-[10px] uppercase tracking-widest transition-all shadow-lg shadow-blue-500/20 active:scale-[0.98]"
        >
          <span v-if="loading">同步中...</span>
          <span v-else>保存资料设置 / Save Profile</span>
        </button>
      </div>
    </div>
  </div>
</template>
