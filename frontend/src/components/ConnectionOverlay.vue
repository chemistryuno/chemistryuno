<script setup lang="ts">
import { useConnectionState } from '../composables/useConnectionState'

const { showConnectionOverlay, isFailed, statusText, retry } = useConnectionState()
</script>

<template>
  <Transition name="conn-fade">
    <div
      v-if="showConnectionOverlay"
      class="fixed inset-0 z-[9999] flex items-center justify-center bg-slate-900/70 backdrop-blur-sm"
      role="alertdialog"
      aria-live="assertive"
      :aria-label="statusText"
    >
      <div class="flex flex-col items-center gap-4 rounded-2xl bg-slate-800/90 px-8 py-6 text-center shadow-2xl ring-1 ring-white/10">
        <!-- 重连中：旋转指示 -->
        <svg
          v-if="!isFailed"
          class="h-10 w-10 animate-spin text-cyan-400"
          fill="none"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
        <!-- 失败：警告图标 -->
        <svg
          v-else
          class="h-10 w-10 text-rose-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01M5.07 19h13.86a2 2 0 001.74-2.99l-6.93-12a2 2 0 00-3.48 0l-6.93 12A2 2 0 005.07 19z" />
        </svg>

        <p class="text-base font-medium text-slate-100">{{ statusText }}</p>
        <p v-if="!isFailed" class="text-sm text-slate-400">请保持页面打开，正在自动恢复连接</p>
        <p v-else class="text-sm text-slate-400">网络连接已断开，请检查网络后重试</p>

        <button
          v-if="isFailed"
          type="button"
          class="mt-1 rounded-lg bg-cyan-500 px-5 py-2 text-sm font-semibold text-white transition hover:bg-cyan-400 focus:outline-none focus:ring-2 focus:ring-cyan-300"
          @click="retry"
        >
          重新连接
        </button>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.conn-fade-enter-active,
.conn-fade-leave-active {
  transition: opacity 0.2s ease;
}
.conn-fade-enter-from,
.conn-fade-leave-to {
  opacity: 0;
}
@media (prefers-reduced-motion: reduce) {
  .animate-spin {
    animation: none;
  }
  .conn-fade-enter-active,
  .conn-fade-leave-active {
    transition: none;
  }
}
</style>
