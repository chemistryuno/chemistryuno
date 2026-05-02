<template>
  <Teleport to="body">
    <div class="anticheat-notification-container">
      <TransitionGroup name="notification" tag="div">
        <div
          v-for="notification in notifications"
          :key="notification.id"
          :class="['anticheat-notification', `severity-${notification.severity}`]"
          @click="dismissNotification(notification.id)"
        >
          <!-- 背景粒子效果 -->
          <div class="particles-bg">
            <div v-for="i in 6" :key="i" class="particle" :style="{ '--i': i }"></div>
          </div>

          <!-- 发光边框 -->
          <div class="glow-border"></div>

          <!-- 内容区域 -->
          <div class="notification-content">
            <!-- 图标 -->
            <div class="notification-icon">
              <div v-if="notification.severity === 'warning'" class="icon-svg">⚠</div>
              <div v-else-if="notification.severity === 'danger'" class="icon-svg">🚫</div>
              <div v-else-if="notification.severity === 'mute'" class="icon-svg">🔇</div>
              <div v-else class="icon-svg">ℹ</div>
            </div>

            <!-- 信息 -->
            <div class="notification-text">
              <div class="notification-title">{{ notification.title }}</div>
              <div class="notification-message">{{ notification.message }}</div>
              <div v-if="notification.details" class="notification-details">
                {{ notification.details }}
              </div>
            </div>

            <!-- 操作按钮 -->
            <div v-if="notification.actionLabel" class="notification-action">
              <button
                class="action-button"
                @click.stop="handleAction(notification)"
              >
                {{ notification.actionLabel }}
              </button>
            </div>
          </div>

          <!-- 进度条 -->
          <div class="progress-bar" :style="{ animationDuration: `${notification.duration}ms` }"></div>

          <!-- 波纹效果 -->
          <div class="ripple"></div>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref } from 'vue'

type SeverityLevel = 'info' | 'warning' | 'mute' | 'danger'

interface AnticheatNotification {
  id: number
  severity: SeverityLevel
  title: string
  message: string
  details?: string
  duration: number
  actionLabel?: string
  onAction?: () => void
}

const notifications = ref<AnticheatNotification[]>([])
let notificationId = 0

const showNotification = (
  title: string,
  message: string,
  severity: SeverityLevel = 'info',
  options?: {
    details?: string
    duration?: number
    actionLabel?: string
    onAction?: () => void
  }
) => {
  const id = notificationId++

  notifications.value.push({
    id,
    severity,
    title,
    message,
    details: options?.details,
    duration: options?.duration ?? 6000,
    actionLabel: options?.actionLabel,
    onAction: options?.onAction
  })

  setTimeout(() => {
    dismissNotification(id)
  }, options?.duration ?? 6000)
}

const dismissNotification = (id: number) => {
  const idx = notifications.value.findIndex(n => n.id === id)
  if (idx !== -1) {
    notifications.value.splice(idx, 1)
  }
}

const handleAction = (notification: AnticheatNotification) => {
  if (notification.onAction) {
    notification.onAction()
  }
  dismissNotification(notification.id)
}

// 便捷方法
const showWarning = (title: string, message: string, details?: string) => {
  showNotification(title, message, 'warning', { details, duration: 8000 })
}

const showMute = (title: string, message: string, duration?: number) => {
  showNotification(title, message, 'mute', { duration: duration ?? 10000 })
}

const showBan = (title: string, message: string, details?: string) => {
  showNotification(title, message, 'danger', { details, duration: 0 })
}

const showAppealPrompt = (onAppeal: () => void) => {
  showNotification(
    '检测到异常行为',
    '系统检测到您的账号存在异常行为。您可以提交申诉来解释这些行为。',
    'warning',
    {
      actionLabel: '提交申诉',
      onAction: onAppeal,
      duration: 15000
    }
  )
}

defineExpose({
  showNotification,
  showWarning,
  showMute,
  showBan,
  showAppealPrompt,
  dismissNotification
})
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@500;600;700&display=swap');

.anticheat-notification-container {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  z-index: 9998;
  display: flex;
  flex-direction: column;
  gap: 12px;
  pointer-events: none;
  max-width: 420px;
}

@media (max-width: 640px) {
  .anticheat-notification-container {
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    right: auto;
    max-width: none;
    width: 90vw;
  }
}

/* ==================== 通知基础样式 ==================== */
.anticheat-notification {
  position: relative;
  width: 100%;
  padding: 16px 18px;
  border-radius: 12px;
  backdrop-filter: blur(24px) saturate(180%);
  border: 1.5px solid rgba(255, 255, 255, 0.2);
  overflow: hidden;
  pointer-events: auto;
  animation: notification-scale-in 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
  transition: transform 0.2s ease;
  cursor: pointer;
}

.anticheat-notification:hover {
  transform: scale(1.02);
}

@keyframes notification-scale-in {
  from {
    opacity: 0;
    transform: scale(0.9);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

/* ==================== 严重级别配色 ==================== */
/* 信息 - 蓝色 */
.severity-info {
  background: linear-gradient(135deg,
    rgba(59, 130, 246, 0.12) 0%,
    rgba(37, 99, 235, 0.18) 100%);
  box-shadow:
    0 4px 24px rgba(59, 130, 246, 0.25),
    0 0 0 1px rgba(59, 130, 246, 0.15),
    inset 0 0 40px rgba(59, 130, 246, 0.08);
}

.severity-info .glow-border {
  background: linear-gradient(45deg,
    #3b82f6 0%,
    #2563eb 50%,
    #3b82f6 100%);
}

.severity-info .progress-bar {
  background: linear-gradient(90deg, #3b82f6, #2563eb);
}

.severity-info .notification-icon {
  color: #3b82f6;
}

/* 警告 - 橙色 */
.severity-warning {
  background: linear-gradient(135deg,
    rgba(249, 115, 22, 0.12) 0%,
    rgba(234, 88, 12, 0.18) 100%);
  box-shadow:
    0 4px 24px rgba(249, 115, 22, 0.25),
    0 0 0 1px rgba(249, 115, 22, 0.15),
    inset 0 0 40px rgba(249, 115, 22, 0.08);
}

.severity-warning .glow-border {
  background: linear-gradient(45deg,
    #f97316 0%,
    #ea580c 50%,
    #f97316 100%);
}

.severity-warning .progress-bar {
  background: linear-gradient(90deg, #f97316, #ea580c);
}

.severity-warning .notification-icon {
  color: #f97316;
}

/* 禁言 - 紫色 */
.severity-mute {
  background: linear-gradient(135deg,
    rgba(168, 85, 247, 0.12) 0%,
    rgba(147, 51, 234, 0.18) 100%);
  box-shadow:
    0 4px 24px rgba(168, 85, 247, 0.25),
    0 0 0 1px rgba(168, 85, 247, 0.15),
    inset 0 0 40px rgba(168, 85, 247, 0.08);
}

.severity-mute .glow-border {
  background: linear-gradient(45deg,
    #a855f7 0%,
    #9333ea 50%,
    #a855f7 100%);
}

.severity-mute .progress-bar {
  background: linear-gradient(90deg, #a855f7, #9333ea);
}

.severity-mute .notification-icon {
  color: #a855f7;
}

/* 危险 - 红色 */
.severity-danger {
  background: linear-gradient(135deg,
    rgba(239, 68, 68, 0.12) 0%,
    rgba(220, 38, 38, 0.18) 100%);
  box-shadow:
    0 4px 24px rgba(239, 68, 68, 0.25),
    0 0 0 1px rgba(239, 68, 68, 0.15),
    inset 0 0 40px rgba(239, 68, 68, 0.08);
}

.severity-danger .glow-border {
  background: linear-gradient(45deg,
    #ef4444 0%,
    #dc2626 50%,
    #ef4444 100%);
}

.severity-danger .progress-bar {
  background: linear-gradient(90deg, #ef4444, #dc2626);
}

.severity-danger .notification-icon {
  color: #ef4444;
}

/* ==================== 发光边框 ==================== */
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
  opacity: 0.4;
}

.particle {
  position: absolute;
  width: 2px;
  height: 2px;
  background: currentColor;
  border-radius: 50%;
  opacity: 0;
  animation: particle-rise 2.5s ease-out infinite;
  animation-delay: calc(var(--i) * 0.3s);
}

.particle:nth-child(odd) {
  left: calc(var(--i) * 15%);
}

.particle:nth-child(even) {
  right: calc(var(--i) * 15%);
}

@keyframes particle-rise {
  0% {
    transform: translateY(0) scale(0);
    opacity: 0;
  }
  20% {
    opacity: 0.6;
    transform: translateY(-12px) scale(1);
  }
  80% {
    opacity: 0.1;
    transform: translateY(-40px) scale(0.5);
  }
  100% {
    opacity: 0;
    transform: translateY(-50px) scale(0);
  }
}

/* ==================== 内容区域 ==================== */
.notification-content {
  position: relative;
  display: flex;
  align-items: flex-start;
  gap: 12px;
  z-index: 1;
}

/* ==================== 图标 ==================== */
.notification-icon {
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(8px);
  font-size: 24px;
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

/* ==================== 文本内容 ==================== */
.notification-text {
  flex: 1;
  min-width: 0;
}

.notification-title {
  font-family: 'Inter', sans-serif;
  font-size: 14px;
  font-weight: 700;
  color: #ffffff;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
  margin-bottom: 4px;
  line-height: 1.3;
}

.notification-message {
  font-family: 'Inter', sans-serif;
  font-size: 13px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.85);
  line-height: 1.4;
  overflow-wrap: break-word;
}

.notification-details {
  font-family: 'Inter', sans-serif;
  font-size: 12px;
  font-weight: 400;
  color: rgba(255, 255, 255, 0.7);
  margin-top: 6px;
  padding-top: 6px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  line-height: 1.3;
}

/* ==================== 操作按钮 ==================== */
.notification-action {
  flex-shrink: 0;
  margin-left: 8px;
}

.action-button {
  padding: 6px 12px;
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  background: rgba(255, 255, 255, 0.1);
  color: #ffffff;
  font-family: 'Inter', sans-serif;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.action-button:hover {
  background: rgba(255, 255, 255, 0.2);
  border-color: rgba(255, 255, 255, 0.5);
  transform: scale(1.05);
}

.action-button:active {
  transform: scale(0.95);
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

/* ==================== 通知动画 ==================== */
@keyframes notification-slide-in {
  0% {
    transform: translateX(-400px) scale(0.9);
    opacity: 0;
  }
  60% {
    transform: translateX(8px) scale(1.02);
    opacity: 1;
  }
  100% {
    transform: translateX(0) scale(1);
    opacity: 1;
  }
}

/* Vue Transition */
.notification-enter-active {
  animation: notification-slide-in 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.notification-leave-active {
  animation: notification-slide-out 0.35s ease-in forwards;
}

@keyframes notification-slide-out {
  0% {
    transform: translateX(0) scale(1);
    opacity: 1;
  }
  100% {
    transform: translateX(-400px) scale(0.85);
    opacity: 0;
  }
}

.notification-move {
  transition: transform 0.35s cubic-bezier(0.34, 1.56, 0.64, 1);
}

/* 响应式优化 */
@media (max-width: 640px) {
  .anticheat-notification {
    padding: 14px 16px;
  }

  .notification-icon {
    width: 36px;
    height: 36px;
    font-size: 20px;
  }

  .notification-title {
    font-size: 13px;
  }

  .notification-message {
    font-size: 12px;
  }

  .notification-details {
    font-size: 11px;
  }

  .action-button {
    padding: 5px 10px;
    font-size: 11px;
  }

  .notification-content {
    gap: 10px;
  }
}

/* 暗色模式优化 */
@media (prefers-color-scheme: dark) {
  .anticheat-notification {
    border-color: rgba(255, 255, 255, 0.12);
  }
}
</style>
