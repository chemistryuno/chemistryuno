<script setup lang="ts">
import { computed } from 'vue'
import { usePing } from '../composables/usePing'

const { pingStatus } = usePing()
const shouldShowPingPrompt = computed(() => pingStatus.value.status !== 'disconnected' && pingStatus.value.latency >= 1000)
</script>

<template>
  <!-- 调整位置：右上角但在公告栏下方，z-index降低避免遮挡重要内容 -->
  <div v-if="shouldShowPingPrompt" class="fixed top-8 right-2 z-50 pointer-events-none animate-fade-in sm:top-10 sm:right-1">
    <div
      class="flex items-center gap-1.5 px-3 py-1.5 bg-white/90 dark:bg-slate-900/90 backdrop-blur-md border border-slate-200/50 dark:border-white/10 rounded-xl shadow-lg pointer-events-auto transition-all duration-300 cursor-pointer hover:bg-white dark:hover:bg-slate-900 hover:shadow-xl hover:scale-105 group sm:px-2 sm:py-1 sm:gap-1"
      :class="[pingStatus.statusColor]"
    >
      <!-- 信号图标 -->
      <div class="flex items-center justify-center transition-transform duration-300 group-hover:scale-110">
        <svg
          v-if="pingStatus.status === 'disconnected'"
          class="w-4 h-4 sm:w-3 sm:h-3"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 5.636a9 9 0 010 12.728m0 0l-2.829-2.829m2.829 2.829L21 21M15.536 8.464a5 5 0 010 7.072m0 0l-2.829-2.829m-4.243 2.829a4.978 4.978 0 01-1.414-2.83m-1.414 5.658a9 9 0 01-2.167-9.238m7.824 2.167a1 1 0 111.414 1.414m-1.414-1.414L3 3m8.293 8.293l1.414 1.414" />
        </svg>
        <svg
          v-else
          class="w-4 h-4 sm:w-3 sm:h-3"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0" />
        </svg>
      </div>

      <!-- 延迟显示 -->
      <div class="flex items-baseline gap-0.5 font-mono text-sm sm:text-xs">
        <span class="font-bold">{{ pingStatus.latency }}</span>
        <span class="text-xs opacity-70 sm:text-[10px]">ms</span>
      </div>

      <!-- 状态指示器 -->
      <div
        class="w-2 h-2 rounded-full transition-all duration-300 sm:w-1.5 sm:h-1.5"
        :class="[
          {
            'bg-emerald-500': pingStatus.status === 'excellent',
            'bg-green-500': pingStatus.status === 'good',
            'bg-yellow-500': pingStatus.status === 'fair',
            'bg-red-500': pingStatus.status === 'poor',
            'bg-slate-400': pingStatus.status === 'disconnected',
          },
          pingStatus.status !== 'disconnected' ? 'animate-pulse' : ''
        ]"
      ></div>

      <!-- 悬停提示（桌面端） -->
      <div class="absolute top-full right-0 mt-2 opacity-0 invisible transition-all duration-200 pointer-events-none group-hover:opacity-100 group-hover:visible sm:hidden">
        <div class="bg-slate-900 dark:bg-slate-800 text-white px-3 py-2 rounded-lg shadow-xl border border-white/10 min-w-[120px] whitespace-nowrap">
          <div class="text-xs font-bold mb-1">网络状态</div>
          <div class="text-sm">
            延迟：<span class="font-mono font-bold">{{ pingStatus.latency }}ms</span>
          </div>
          <div class="text-sm">
            状态：<span class="font-semibold">{{ pingStatus.statusText }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 动画定义 */
@keyframes fade-in {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-fade-in {
  animation: fade-in 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}
</style>
