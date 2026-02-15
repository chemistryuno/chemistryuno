<template>
  <Teleport to="body">
    <div class="reaction-toast-container">
      <TransitionGroup name="toast" tag="div">
        <div
          v-for="toast in toasts"
          :key="toast.id"
          :class="['reaction-toast', `reaction-${toast.type}`]"
          :style="{ '--toast-index': toast.index }"
        >
          <!-- 背景粒子效果 -->
          <div class="particles-bg">
            <div v-for="i in 12" :key="i" class="particle" :style="{ '--i': i }"></div>
          </div>

          <!-- 发光边框 -->
          <div class="glow-border"></div>

          <!-- 内容区域 -->
          <div class="toast-content">
            <!-- 反应图标 -->
            <div class="reaction-icon">
              <div class="atom-nucleus"></div>
              <div class="electron-orbit orbit-1"></div>
              <div class="electron-orbit orbit-2"></div>
              <div class="electron"></div>
            </div>

            <!-- 反应信息 -->
            <div class="reaction-info">
              <div class="reaction-equation">{{ toast.equation }}</div>
              <div class="reaction-name">{{ toast.name }}</div>
            </div>

            <!-- 能量指示器 -->
            <div class="energy-indicator">
              <div class="energy-bar" :style="{ width: `${toast.energy}%` }"></div>
            </div>
          </div>

          <!-- 波纹效果 -->
          <div class="ripple"></div>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface ReactionToast {
  id: number
  type: 'synthesis' | 'decomposition' | 'displacement' | 'combustion' | 'neutralization'
  equation: string
  name: string
  energy: number
  index: number
}

const toasts = ref<ReactionToast[]>([])
let toastId = 0

const showToast = (
  type: ReactionToast['type'],
  equation: string,
  name: string,
  energy: number = 75
) => {
  const id = toastId++
  const index = toasts.value.length

  toasts.value.push({
    id,
    type,
    equation,
    name,
    energy,
    index
  })

  setTimeout(() => {
    const idx = toasts.value.findIndex(t => t.id === id)
    if (idx !== -1) {
      toasts.value.splice(idx, 1)
    }
  }, 4000)
}

// 暴露方法供父组件调用
defineExpose({
  showToast
})
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Orbitron:wght@700;900&family=JetBrains+Mono:wght@500;700&display=swap');

.reaction-toast-container {
  position: fixed;
  top: 80px;
  right: 20px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 16px;
  pointer-events: none;
}

@media (max-width: 640px) {
  .reaction-toast-container {
    top: 60px;
    right: 12px;
    left: 12px;
  }
}

/* ==================== Toast 基础样式 ==================== */
.reaction-toast {
  position: relative;
  width: 380px;
  max-width: calc(100vw - 24px);
  padding: 20px;
  border-radius: 16px;
  backdrop-filter: blur(20px) saturate(180%);
  border: 2px solid rgba(255, 255, 255, 0.2);
  overflow: hidden;
  pointer-events: auto;
  animation: toast-pulse 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}

/* ==================== 反应类型配色 ==================== */
/* 合成反应 - 蓝绿色科技感 */
.reaction-synthesis {
  background: linear-gradient(135deg,
    rgba(6, 182, 212, 0.15) 0%,
    rgba(14, 165, 233, 0.25) 50%,
    rgba(59, 130, 246, 0.15) 100%);
  box-shadow:
    0 8px 32px rgba(6, 182, 212, 0.3),
    0 0 0 1px rgba(6, 182, 212, 0.2),
    inset 0 0 60px rgba(6, 182, 212, 0.1);
}

.reaction-synthesis .glow-border {
  background: linear-gradient(45deg,
    #06b6d4 0%,
    #0ea5e9 25%,
    #3b82f6 50%,
    #0ea5e9 75%,
    #06b6d4 100%);
}

.reaction-synthesis .energy-bar {
  background: linear-gradient(90deg, #06b6d4, #3b82f6);
}

/* 分解反应 - 橙红色能量爆发 */
.reaction-decomposition {
  background: linear-gradient(135deg,
    rgba(249, 115, 22, 0.15) 0%,
    rgba(239, 68, 68, 0.25) 50%,
    rgba(220, 38, 38, 0.15) 100%);
  box-shadow:
    0 8px 32px rgba(249, 115, 22, 0.3),
    0 0 0 1px rgba(249, 115, 22, 0.2),
    inset 0 0 60px rgba(249, 115, 22, 0.1);
}

.reaction-decomposition .glow-border {
  background: linear-gradient(45deg,
    #f97316 0%,
    #ef4444 25%,
    #dc2626 50%,
    #ef4444 75%,
    #f97316 100%);
}

.reaction-decomposition .energy-bar {
  background: linear-gradient(90deg, #f97316, #dc2626);
}

/* 置换反应 - 紫色魔幻感 */
.reaction-displacement {
  background: linear-gradient(135deg,
    rgba(168, 85, 247, 0.15) 0%,
    rgba(147, 51, 234, 0.25) 50%,
    rgba(126, 34, 206, 0.15) 100%);
  box-shadow:
    0 8px 32px rgba(168, 85, 247, 0.3),
    0 0 0 1px rgba(168, 85, 247, 0.2),
    inset 0 0 60px rgba(168, 85, 247, 0.1);
}

.reaction-displacement .glow-border {
  background: linear-gradient(45deg,
    #a855f7 0%,
    #9333ea 25%,
    #7e22ce 50%,
    #9333ea 75%,
    #a855f7 100%);
}

.reaction-displacement .energy-bar {
  background: linear-gradient(90deg, #a855f7, #7e22ce);
}

/* 燃烧反应 - 火焰橙黄色 */
.reaction-combustion {
  background: linear-gradient(135deg,
    rgba(251, 146, 60, 0.15) 0%,
    rgba(251, 191, 36, 0.25) 50%,
    rgba(245, 158, 11, 0.15) 100%);
  box-shadow:
    0 8px 32px rgba(251, 146, 60, 0.4),
    0 0 0 1px rgba(251, 146, 60, 0.3),
    inset 0 0 60px rgba(251, 146, 60, 0.15);
}

.reaction-combustion .glow-border {
  background: linear-gradient(45deg,
    #fb923c 0%,
    #fbbf24 25%,
    #f59e0b 50%,
    #fbbf24 75%,
    #fb923c 100%);
}

.reaction-combustion .energy-bar {
  background: linear-gradient(90deg, #fb923c, #f59e0b);
}

/* 中和反应 - 青绿色平衡感 */
.reaction-neutralization {
  background: linear-gradient(135deg,
    rgba(16, 185, 129, 0.15) 0%,
    rgba(5, 150, 105, 0.25) 50%,
    rgba(4, 120, 87, 0.15) 100%);
  box-shadow:
    0 8px 32px rgba(16, 185, 129, 0.3),
    0 0 0 1px rgba(16, 185, 129, 0.2),
    inset 0 0 60px rgba(16, 185, 129, 0.1);
}

.reaction-neutralization .glow-border {
  background: linear-gradient(45deg,
    #10b981 0%,
    #059669 25%,
    #047857 50%,
    #059669 75%,
    #10b981 100%);
}

.reaction-neutralization .energy-bar {
  background: linear-gradient(90deg, #10b981, #047857);
}

/* ==================== 发光边框动画 ==================== */
.glow-border {
  position: absolute;
  inset: -2px;
  border-radius: 16px;
  padding: 2px;
  -webkit-mask:
    linear-gradient(#fff 0 0) content-box,
    linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  animation: glow-rotate 4s linear infinite;
  opacity: 0.8;
}

@keyframes glow-rotate {
  0%, 100% { opacity: 0.6; }
  50% { opacity: 1; }
}

/* ==================== 粒子背景 ==================== */
.particles-bg {
  position: absolute;
  inset: 0;
  overflow: hidden;
  border-radius: 16px;
}

.particle {
  position: absolute;
  width: 4px;
  height: 4px;
  background: currentColor;
  border-radius: 50%;
  opacity: 0;
  animation: particle-float 3s ease-in-out infinite;
  animation-delay: calc(var(--i) * 0.2s);
}

.particle:nth-child(odd) {
  left: calc(var(--i) * 8%);
  animation-duration: 2.5s;
}

.particle:nth-child(even) {
  right: calc(var(--i) * 8%);
  animation-duration: 3.5s;
}

@keyframes particle-float {
  0%, 100% {
    transform: translateY(0) scale(0);
    opacity: 0;
  }
  25% {
    opacity: 1;
    transform: translateY(-20px) scale(1);
  }
  50% {
    opacity: 0.8;
    transform: translateY(-40px) scale(0.8);
  }
  75% {
    opacity: 0.4;
    transform: translateY(-60px) scale(0.6);
  }
  100% {
    opacity: 0;
    transform: translateY(-80px) scale(0);
  }
}

/* ==================== 内容区域 ==================== */
.toast-content {
  position: relative;
  display: flex;
  align-items: center;
  gap: 16px;
  z-index: 1;
}

/* ==================== 反应图标（原子模型） ==================== */
.reaction-icon {
  position: relative;
  width: 56px;
  height: 56px;
  flex-shrink: 0;
  animation: atom-spin 8s linear infinite;
}

.atom-nucleus {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 16px;
  height: 16px;
  background: currentColor;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  box-shadow:
    0 0 10px currentColor,
    0 0 20px currentColor,
    0 0 30px currentColor;
  animation: nucleus-pulse 2s ease-in-out infinite;
}

@keyframes nucleus-pulse {
  0%, 100% {
    transform: translate(-50%, -50%) scale(1);
    opacity: 1;
  }
  50% {
    transform: translate(-50%, -50%) scale(1.2);
    opacity: 0.8;
  }
}

.electron-orbit {
  position: absolute;
  top: 50%;
  left: 50%;
  border: 2px solid currentColor;
  border-radius: 50%;
  opacity: 0.4;
  animation: orbit-pulse 3s ease-in-out infinite;
}

.orbit-1 {
  width: 40px;
  height: 40px;
  margin-left: -20px;
  margin-top: -20px;
  transform: rotate(30deg);
}

.orbit-2 {
  width: 52px;
  height: 52px;
  margin-left: -26px;
  margin-top: -26px;
  transform: rotate(-30deg);
  animation-delay: 0.5s;
}

@keyframes orbit-pulse {
  0%, 100% { opacity: 0.3; }
  50% { opacity: 0.6; }
}

.electron {
  position: absolute;
  width: 6px;
  height: 6px;
  background: currentColor;
  border-radius: 50%;
  top: 6px;
  left: 50%;
  margin-left: -3px;
  box-shadow: 0 0 8px currentColor;
  animation: electron-orbit 2s linear infinite;
}

@keyframes electron-orbit {
  from {
    transform: rotate(0deg) translateX(20px) rotate(0deg);
  }
  to {
    transform: rotate(360deg) translateX(20px) rotate(-360deg);
  }
}

@keyframes atom-spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* ==================== 反应信息 ==================== */
.reaction-info {
  flex: 1;
  min-width: 0;
}

.reaction-equation {
  font-family: 'Orbitron', monospace;
  font-size: 16px;
  font-weight: 900;
  color: #ffffff;
  text-shadow:
    0 0 10px currentColor,
    0 2px 4px rgba(0, 0, 0, 0.3);
  letter-spacing: 1px;
  margin-bottom: 6px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  animation: text-glow 2s ease-in-out infinite;
}

@keyframes text-glow {
  0%, 100% { text-shadow: 0 0 10px currentColor, 0 2px 4px rgba(0, 0, 0, 0.3); }
  50% { text-shadow: 0 0 20px currentColor, 0 0 30px currentColor, 0 2px 4px rgba(0, 0, 0, 0.3); }
}

.reaction-name {
  font-family: 'JetBrains Mono', monospace;
  font-size: 11px;
  font-weight: 700;
  color: rgba(255, 255, 255, 0.7);
  text-transform: uppercase;
  letter-spacing: 2px;
}

/* ==================== 能量指示器 ==================== */
.energy-indicator {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: rgba(255, 255, 255, 0.1);
  overflow: hidden;
}

.energy-bar {
  height: 100%;
  width: 0;
  box-shadow: 0 0 10px currentColor;
  animation: energy-fill 4s ease-out forwards;
  position: relative;
}

.energy-bar::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg,
    transparent 0%,
    rgba(255, 255, 255, 0.5) 50%,
    transparent 100%);
  animation: energy-shimmer 1.5s ease-in-out infinite;
}

@keyframes energy-fill {
  from { width: 0%; }
  to { width: var(--energy, 75%); }
}

@keyframes energy-shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(200%); }
}

/* ==================== 波纹效果 ==================== */
.ripple {
  position: absolute;
  inset: 0;
  border-radius: 16px;
  background: radial-gradient(circle at center,
    rgba(255, 255, 255, 0.3) 0%,
    transparent 70%);
  opacity: 0;
  animation: ripple-expand 0.8s cubic-bezier(0.4, 0, 0.2, 1);
}

@keyframes ripple-expand {
  0% {
    opacity: 1;
    transform: scale(0.8);
  }
  100% {
    opacity: 0;
    transform: scale(1.5);
  }
}

/* ==================== Toast 动画 ==================== */
@keyframes toast-pulse {
  0% {
    transform: translateX(400px) scale(0.8) rotate(5deg);
    opacity: 0;
  }
  60% {
    transform: translateX(-10px) scale(1.05) rotate(-2deg);
    opacity: 1;
  }
  80% {
    transform: translateX(5px) scale(0.98) rotate(1deg);
  }
  100% {
    transform: translateX(0) scale(1) rotate(0);
    opacity: 1;
  }
}

/* Vue Transition */
.toast-enter-active {
  animation: toast-pulse 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.toast-leave-active {
  animation: toast-exit 0.4s ease-in forwards;
}

@keyframes toast-exit {
  0% {
    transform: translateX(0) scale(1);
    opacity: 1;
  }
  100% {
    transform: translateX(400px) scale(0.8) rotate(10deg);
    opacity: 0;
  }
}

.toast-move {
  transition: transform 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}

/* 响应式优化 */
@media (max-width: 640px) {
  .reaction-toast {
    padding: 16px;
  }

  .reaction-icon {
    width: 48px;
    height: 48px;
  }

  .reaction-equation {
    font-size: 14px;
  }

  .reaction-name {
    font-size: 10px;
  }
}

/* 暗色模式优化 */
@media (prefers-color-scheme: dark) {
  .reaction-toast {
    border-color: rgba(255, 255, 255, 0.15);
  }

  .energy-indicator {
    background: rgba(255, 255, 255, 0.05);
  }
}
</style>
