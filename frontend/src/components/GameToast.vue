<template>
  <Teleport to="body">
    <div class="game-toast-container">
      <TransitionGroup name="toast" tag="div">
        <div
          v-for="toast in toasts"
          :key="toast.id"
          :class="['game-toast', `toast-${toast.type}`]"
          @click="dismissToast(toast.id)"
        >
          <!-- 背景粒子效果 -->
          <div class="particles-bg">
            <div v-for="i in 8" :key="i" class="particle" :style="{ '--i': i }"></div>
          </div>

          <!-- 发光边框 -->
          <div class="glow-border"></div>

          <!-- 内容区域 -->
          <div class="toast-content">
            <!-- 图标 -->
            <div class="toast-icon">
              <div v-if="toast.type === 'success'" class="icon-svg">✓</div>
              <div v-else-if="toast.type === 'error'" class="icon-svg">✕</div>
              <div v-else-if="toast.type === 'warning'" class="icon-svg">!</div>
              <div v-else class="icon-svg">ℹ</div>
            </div>

            <!-- 信息 -->
            <div class="toast-text">
              <div v-if="toast.title" class="toast-title">{{ toast.title }}</div>
              <div class="toast-message">{{ toast.message }}</div>
            </div>
          </div>

          <!-- 进度条 -->
          <div class="progress-bar" :style="{ animationDuration: `${toast.duration}ms` }"></div>

          <!-- 波纹效果 -->
          <div class="ripple"></div>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface GameToast {
  id: number
  type: 'info' | 'success' | 'warning' | 'error'
  title?: string
  message: string
  duration: number
}

const toasts = ref<GameToast[]>([])
let toastId = 0

const showToast = (
  message: string,
  title?: string,
  type: GameToast['type'] = 'info',
  duration: number = 4000
) => {
  const id = toastId++

  toasts.value.push({
    id,
    type,
    title,
    message,
    duration
  })

  setTimeout(() => {
    dismissToast(id)
  }, duration)
}

const dismissToast = (id: number) => {
  const idx = toasts.value.findIndex(t => t.id === id)
  if (idx !== -1) {
    toasts.value.splice(idx, 1)
  }
}

// 暴露方法供父组件调用
defineExpose({
  showToast
})
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@500;600;700&display=swap');

.game-toast-container {
  position: fixed;
  top: 80px;
  right: 20px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 12px;
  pointer-events: none;
}

@media (max-width: 640px) {
  .game-toast-container {
    top: 60px;
    right: 12px;
    left: 12px;
  }
}

/* ==================== Toast 基础样式 ==================== */
.game-toast {
  position: relative;
  width: 380px;
  max-width: calc(100vw - 24px);
  padding: 16px 18px;
  border-radius: 12px;
  backdrop-filter: blur(24px) saturate(180%);
  border: 1.5px solid rgba(255, 255, 255, 0.2);
  overflow: hidden;
  pointer-events: auto;
  cursor: pointer;
  animation: toast-slide-in 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
  transition: transform 0.2s ease;
}

.game-toast:hover {
  transform: translateX(-4px);
}

/* ==================== 类型配色 ==================== */
/* 信息提示 - 蓝色 */
.toast-info {
  background: linear-gradient(135deg,
    rgba(59, 130, 246, 0.12) 0%,
    rgba(37, 99, 235, 0.18) 100%);
  box-shadow:
    0 4px 24px rgba(59, 130, 246, 0.25),
    0 0 0 1px rgba(59, 130, 246, 0.15),
    inset 0 0 40px rgba(59, 130, 246, 0.08);
}

.toast-info .glow-border {
  background: linear-gradient(45deg,
    #3b82f6 0%,
    #2563eb 50%,
    #3b82f6 100%);
}

.toast-info .progress-bar {
  background: linear-gradient(90deg, #3b82f6, #2563eb);
}

.toast-info .icon-svg {
  color: #3b82f6;
}

/* 成功提示 - 绿色 */
.toast-success {
  background: linear-gradient(135deg,
    rgba(34, 197, 94, 0.12) 0%,
    rgba(22, 163, 74, 0.18) 100%);
  box-shadow:
    0 4px 24px rgba(34, 197, 94, 0.25),
    0 0 0 1px rgba(34, 197, 94, 0.15),
    inset 0 0 40px rgba(34, 197, 94, 0.08);
}

.toast-success .glow-border {
  background: linear-gradient(45deg,
    #22c55e 0%,
    #16a34a 50%,
    #22c55e 100%);
}

.toast-success .progress-bar {
  background: linear-gradient(90deg, #22c55e, #16a34a);
}

.toast-success .icon-svg {
  color: #22c55e;
}

/* 警告提示 - 橙色 */
.toast-warning {
  background: linear-gradient(135deg,
    rgba(249, 115, 22, 0.12) 0%,
    rgba(234, 88, 12, 0.18) 100%);
  box-shadow:
    0 4px 24px rgba(249, 115, 22, 0.25),
    0 0 0 1px rgba(249, 115, 22, 0.15),
    inset 0 0 40px rgba(249, 115, 22, 0.08);
}

.toast-warning .glow-border {
  background: linear-gradient(45deg,
    #f97316 0%,
    #ea580c 50%,
    #f97316 100%);
}

.toast-warning .progress-bar {
  background: linear-gradient(90deg, #f97316, #ea580c);
}

.toast-warning .icon-svg {
  color: #f97316;
}

/* 错误提示 - 红色 */
.toast-error {
  background: linear-gradient(135deg,
    rgba(239, 68, 68, 0.12) 0%,
    rgba(220, 38, 38, 0.18) 100%);
  box-shadow:
    0 4px 24px rgba(239, 68, 68, 0.25),
    0 0 0 1px rgba(239, 68, 68, 0.15),
    inset 0 0 40px rgba(239, 68, 68, 0.08);
}

.toast-error .glow-border {
  background: linear-gradient(45deg,
    #ef4444 0%,
    #dc2626 50%,
    #ef4444 100%);
}

.toast-error .progress-bar {
  background: linear-gradient(90deg, #ef4444, #dc2626);
}

.toast-error .icon-svg {
  color: #ef4444;
}

/* ==================== 发光边框动画 ==================== */
.glow-border {
  position: absolute;
  inset: -1.5px;
  border-radius: 12px;
  padding: 1.5px;
  -webkit-mask:
    linear-gradient(#fff 0 0) content-box,
    linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  animation: glow-pulse 2.5s ease-in-out infinite;
  opacity: 0.6;
}

@keyframes glow-pulse {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 0.8; }
}

/* ==================== 粒子背景 ==================== */
.particles-bg {
  position: absolute;
  inset: 0;
  overflow: hidden;
  border-radius: 12px;
  opacity: 0.6;
}

.particle {
  position: absolute;
  width: 3px;
  height: 3px;
  background: currentColor;
  border-radius: 50%;
  opacity: 0;
  animation: particle-rise 2.5s ease-out infinite;
  animation-delay: calc(var(--i) * 0.25s);
}

.particle:nth-child(odd) {
  left: calc(var(--i) * 12%);
}

.particle:nth-child(even) {
  right: calc(var(--i) * 12%);
}

@keyframes particle-rise {
  0% {
    transform: translateY(0) scale(0);
    opacity: 0;
  }
  20% {
    opacity: 0.8;
    transform: translateY(-15px) scale(1);
  }
  80% {
    opacity: 0.2;
    transform: translateY(-50px) scale(0.5);
  }
  100% {
    opacity: 0;
    transform: translateY(-60px) scale(0);
  }
}

/* ==================== 内容区域 ==================== */
.toast-content {
  position: relative;
  display: flex;
  align-items: center;
  gap: 12px;
  z-index: 1;
}

/* ==================== 图标 ==================== */
.toast-icon {
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(8px);
}

.icon-svg {
  font-size: 24px;
  font-weight: 700;
  line-height: 1;
  animation: icon-bounce 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes icon-bounce {
  0% {
    transform: scale(0) rotate(-180deg);
    opacity: 0;
  }
  60% {
    transform: scale(1.2) rotate(10deg);
  }
  100% {
    transform: scale(1) rotate(0);
    opacity: 1;
  }
}

/* ==================== 信息文本 ==================== */
.toast-text {
  flex: 1;
  min-width: 0;
}

.toast-title {
  font-family: 'Inter', sans-serif;
  font-size: 14px;
  font-weight: 700;
  color: #ffffff;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
  margin-bottom: 4px;
  line-height: 1.3;
}

.toast-message {
  font-family: 'Inter', sans-serif;
  font-size: 13px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.85);
  line-height: 1.4;
  overflow-wrap: break-word;
}

/* ==================== 进度条 ==================== */
.progress-bar {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 2.5px;
  box-shadow: 0 0 8px currentColor;
  animation: progress-shrink linear forwards;
}

@keyframes progress-shrink {
  from { width: 100%; }
  to { width: 0%; }
}

/* ==================== 波纹效果 ==================== */
.ripple {
  position: absolute;
  inset: 0;
  border-radius: 12px;
  background: radial-gradient(circle at center,
    rgba(255, 255, 255, 0.2) 0%,
    transparent 70%);
  opacity: 0;
  animation: ripple-fade 0.6s ease-out;
}

@keyframes ripple-fade {
  0% {
    opacity: 1;
    transform: scale(0.9);
  }
  100% {
    opacity: 0;
    transform: scale(1.3);
  }
}

/* ==================== Toast 动画 ==================== */
@keyframes toast-slide-in {
  0% {
    transform: translateX(400px) scale(0.9);
    opacity: 0;
  }
  60% {
    transform: translateX(-8px) scale(1.02);
    opacity: 1;
  }
  100% {
    transform: translateX(0) scale(1);
    opacity: 1;
  }
}

/* Vue Transition */
.toast-enter-active {
  animation: toast-slide-in 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.toast-leave-active {
  animation: toast-slide-out 0.35s ease-in forwards;
}

@keyframes toast-slide-out {
  0% {
    transform: translateX(0) scale(1);
    opacity: 1;
  }
  100% {
    transform: translateX(400px) scale(0.85);
    opacity: 0;
  }
}

.toast-move {
  transition: transform 0.35s cubic-bezier(0.34, 1.56, 0.64, 1);
}

/* 响应式优化 */
@media (max-width: 640px) {
  .game-toast {
    padding: 14px 16px;
  }

  .toast-icon {
    width: 36px;
    height: 36px;
  }

  .icon-svg {
    font-size: 20px;
  }

  .toast-title {
    font-size: 13px;
  }

  .toast-message {
    font-size: 12px;
  }
}

/* 暗色模式优化 */
@media (prefers-color-scheme: dark) {
  .game-toast {
    border-color: rgba(255, 255, 255, 0.12);
  }
}
</style>
