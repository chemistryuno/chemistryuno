<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { Sun, Moon, Monitor, Palette } from 'lucide-vue-next'

const themes = [
  { id: 'light', name: '明亮模式 / Light', icon: Sun },
  { id: 'dark', name: '深邃模式 / Dark', icon: Moon },
  { id: 'system', name: '系统同步 / System', icon: Monitor }
]

const currentTheme = ref(localStorage.getItem('theme') || 'system')

const applyTheme = (theme: string) => {
  const oldTheme = localStorage.getItem('theme')
  localStorage.setItem('theme', theme)
  // 只有当主题真正改变时才触发事件
  if (oldTheme !== theme) {
    window.dispatchEvent(new Event('theme-changed'))
  }
}

onMounted(() => {
  applyTheme(currentTheme.value)
})

watch(currentTheme, (newTheme) => {
  applyTheme(newTheme)
})
</script>

<template>
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
</template>
