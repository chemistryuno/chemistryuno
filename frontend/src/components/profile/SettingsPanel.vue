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
  <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2.5rem] p-10 shadow-sm dark:shadow-none transition-all hover:shadow-lg">
    <h3 class="text-xl font-bold uppercase tracking-widest mb-8 flex items-center gap-3 text-slate-800 dark:text-white">
      <Palette class="w-6 h-6 text-blue-500" />
      视觉风格设置 / Theme Settings
    </h3>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <button
        v-for="theme in themes"
        :key="theme.id"
        @click="currentTheme = theme.id"
        :class="[
          'flex flex-col items-center justify-center p-6 rounded-[2rem] border-2 transition-all group',
          currentTheme === theme.id 
            ? 'border-blue-500 bg-blue-500/5 text-blue-600 dark:text-blue-400' 
            : 'border-slate-100 dark:border-white/5 bg-slate-50 dark:bg-white/[0.02] text-slate-400 hover:border-slate-200 dark:hover:border-white/10'
        ]"
      >
        <component 
          :is="theme.icon" 
          :class="[
            'w-8 h-8 mb-3 transition-transform duration-500',
            currentTheme === theme.id ? 'scale-110' : 'group-hover:scale-110'
          ]" 
        />
        <span class="text-[10px] font-black uppercase tracking-widest">{{ theme.name }}</span>
      </button>
    </div>

    <div class="mt-8 p-4 bg-slate-50 dark:bg-white/[0.02] rounded-2xl border border-slate-100 dark:border-white/5">
      <p class="text-[10px] text-slate-400 font-bold uppercase tracking-widest text-center leading-relaxed">
        调整显示模式以适应您的实验环境。系统同步将根据您的操作系统设置自动切换明亮与实验室深色风格。
      </p>
    </div>
  </div>
</template>
