<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { Sun, Moon, Monitor, Palette, Volume2, Smartphone, Keyboard, Save } from 'lucide-vue-next'
import { authAPI } from '../../utils/api'
import { useDialog } from '../../utils/dialog'
import { feedback } from '../../utils/feedback'

const props = defineProps<{
  user: any
}>()

const emit = defineEmits<{
  (e: 'update'): void
}>()

const { showAlert } = useDialog()
const loading = ref(false)

const themes = [
  { id: 'light', name: '明亮模式 / Light', icon: Sun },
  { id: 'dark', name: '深邃模式 / Dark', icon: Moon },
  { id: 'system', name: '系统同步 / System', icon: Monitor }
]

const currentTheme = ref(localStorage.getItem('theme') || 'system')
const soundVolume = ref(props.user.sound_volume ?? 0.3)
const vibrationEnabled = ref(props.user.vibration_enabled ?? true)
const enableElementInput = ref(props.user.enable_element_input ?? true)

const applyTheme = (theme: string) => {
  const oldTheme = localStorage.getItem('theme')
  localStorage.setItem('theme', theme)
  // 只有当主题真正改变时才触发事件
  if (oldTheme !== theme) {
    window.dispatchEvent(new Event('theme-changed'))
  }
}

const handleSaveSettings = async () => {
  loading.value = true
  try {
    await authAPI.updateProfile({
      nickname: props.user.nickname,
      sound_volume: soundVolume.value,
      vibration_enabled: vibrationEnabled.value,
      enable_element_input: enableElementInput.value
    })
    
    // 同步到本地反馈管理器
    feedback.setVolume(soundVolume.value)
    feedback.setVibrationEnabled(vibrationEnabled.value)
    
    showAlert('系统设置已云端同步。', '变更成功')
    emit('update')
  } catch (error: any) {
    showAlert(error.response?.data?.error || '同步失败', '错误')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  applyTheme(currentTheme.value)
})

watch(currentTheme, (newTheme) => {
  applyTheme(newTheme)
})

watch(() => props.user, (newUser) => {
  if (newUser) {
    soundVolume.value = newUser.sound_volume ?? 0.3
    vibrationEnabled.value = newUser.vibration_enabled ?? true
    enableElementInput.value = newUser.enable_element_input ?? true
  }
}, { deep: true })
</script>

<template>
  <div class="space-y-6">
    <!-- 主题设置 -->
    <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-2xl p-6 shadow-sm dark:shadow-none transition-all">
      <h3 class="text-base font-black uppercase tracking-widest mb-5 flex items-center gap-2.5 text-slate-800 dark:text-white">
        <Palette class="w-4 h-4 text-blue-500" />
        外观控制 <span class="text-[10px] font-mono opacity-30">/ THEME</span>
      </h3>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-2.5">
        <button
          v-for="theme in themes"
          :key="theme.id"
          @click="currentTheme = theme.id"
          :class="[
            'flex items-center gap-2.5 p-3 rounded-xl border transition-all group backdrop-blur-md',
            currentTheme === theme.id 
              ? 'border-blue-500/50 bg-blue-500/10 dark:bg-blue-500/20 text-blue-600 dark:text-blue-400 shadow-[0_4px_12px_rgba(59,130,246,0.1)]' 
              : 'border-slate-100 dark:border-white/5 bg-slate-50 dark:bg-white/[0.02] text-slate-400 hover:border-slate-200 dark:hover:border-white/10'
          ]"
        >
          <component 
            :is="theme.icon" 
            :class="[
              'w-4 h-4 transition-transform duration-500',
              currentTheme === theme.id ? 'scale-110 shadow-[0_0_8px_currentColor]' : 'group-hover:rotate-12'
            ]" 
          />
          <span class="text-[10px] font-black uppercase tracking-tight text-left leading-none">{{ theme.name.split('/')[0].trim() }}</span>
        </button>
      </div>

      <div class="mt-4 p-3 bg-slate-50 dark:bg-white/[0.02] rounded-xl border border-slate-100 dark:border-white/5">
        <p class="text-[9px] text-slate-400 font-black uppercase tracking-widest text-center leading-relaxed opacity-50">
          Adaptive protocol: system_sync_active
        </p>
      </div>
    </div>

    <!-- 系统偏好设置 -->
    <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-2xl p-6 shadow-sm dark:shadow-none transition-all">
      <h3 class="text-base font-black uppercase tracking-widest mb-6 flex items-center gap-2.5 text-slate-800 dark:text-white">
        <Monitor class="w-4 h-4 text-purple-500" />
        系统交互 <span class="text-[10px] font-mono opacity-30">/ SYSTEM_INTERACTION</span>
      </h3>

      <div class="space-y-6">
        <!-- 音效大小 -->
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <Volume2 class="w-4 h-4 text-slate-400" />
              <span class="text-[10px] font-black uppercase tracking-widest text-slate-500">音效大小 / volume</span>
            </div>
            <span class="text-[10px] font-mono text-blue-500">{{ Math.round(soundVolume * 100) }}%</span>
          </div>
          <input 
            v-model.number="soundVolume" 
            type="range" 
            min="0" 
            max="1" 
            step="0.01"
            class="w-full h-1.5 bg-slate-100 dark:bg-white/5 rounded-lg appearance-none cursor-pointer accent-blue-500"
          />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <!-- 震动反馈 -->
          <div class="p-4 bg-slate-50 dark:bg-white/[0.02] border border-slate-100 dark:border-white/5 rounded-2xl flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-lg flex items-center justify-center" :class="vibrationEnabled ? 'bg-orange-500/10 text-orange-500' : 'bg-slate-200 dark:bg-white/5 text-slate-400'">
                <Smartphone class="w-4 h-4" />
              </div>
              <div>
                <p class="text-[10px] font-black uppercase tracking-wider text-slate-800 dark:text-white leading-none">触感反馈</p>
                <p class="text-[9px] text-slate-400 mt-1 uppercase">Haptic Feedback</p>
              </div>
            </div>
            <button 
              @click="vibrationEnabled = !vibrationEnabled"
              :class="[
                'relative inline-flex h-5 w-10 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2',
                vibrationEnabled ? 'bg-orange-500' : 'bg-slate-200 dark:bg-white/10'
              ]"
            >
              <span 
                :class="[
                  'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                  vibrationEnabled ? 'translate-x-5' : 'translate-x-0'
                ]"
              />
            </button>
          </div>

          <!-- 元素输入法 (PC) -->
          <div class="p-4 bg-slate-50 dark:bg-white/[0.02] border border-slate-100 dark:border-white/5 rounded-2xl flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-lg flex items-center justify-center" :class="enableElementInput ? 'bg-indigo-500/10 text-indigo-500' : 'bg-slate-200 dark:bg-white/5 text-slate-400'">
                <Keyboard class="w-4 h-4" />
              </div>
              <div>
                <p class="text-[10px] font-black uppercase tracking-wider text-slate-800 dark:text-white leading-none">元素输入法</p>
                <p class="text-[9px] text-slate-400 mt-1 uppercase">Element Input (PC)</p>
              </div>
            </div>
            <button 
              @click="enableElementInput = !enableElementInput"
              :class="[
                'relative inline-flex h-5 w-10 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2',
                enableElementInput ? 'bg-indigo-500' : 'bg-slate-200 dark:bg-white/10'
              ]"
            >
              <span 
                :class="[
                  'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                  enableElementInput ? 'translate-x-5' : 'translate-x-0'
                ]"
              />
            </button>
          </div>
        </div>

        <button 
          @click="handleSaveSettings"
          :disabled="loading"
          class="w-full flex items-center justify-center gap-2.5 py-3.5 bg-slate-900 dark:bg-white text-white dark:text-black rounded-xl font-black text-[10px] uppercase tracking-[0.2em] transition-all hover:opacity-90 active:scale-[0.98] shadow-lg shadow-black/10 dark:shadow-white/5"
        >
          <Save v-if="!loading" class="w-3.5 h-3.5" />
          <div v-else class="w-3.5 h-3.5 border-2 border-current border-t-transparent rounded-full animate-spin"></div>
          同步系统偏好 / Sync Preferences
        </button>
      </div>
    </div>
  </div>
</template>

