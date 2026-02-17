<script setup lang="ts">
import { ref } from 'vue'
import { Zap, TrendingUp, Award } from 'lucide-vue-next'

interface LevelUpData {
  level: number
  tier: string
  tier_name: string
  xp: number
  total_xp: number
}

const visible = ref(false)
const levelData = ref<LevelUpData | null>(null)

// 段位图标
const tierIcons: Record<string, string> = {
  bronze: '🥉',
  silver: '🥈',
  gold: '🥇',
  platinum: '💎',
  diamond: '💠',
  master: '⭐'
}

// 段位渐变色
const tierGradients: Record<string, string> = {
  bronze: 'from-orange-400 via-amber-500 to-orange-600',
  silver: 'from-slate-300 via-slate-400 to-slate-500',
  gold: 'from-yellow-300 via-yellow-400 to-yellow-500',
  platinum: 'from-cyan-300 via-blue-400 to-indigo-500',
  diamond: 'from-blue-400 via-indigo-500 to-purple-600',
  master: 'from-purple-500 via-pink-500 to-red-500'
}

// 显示升级动画
function show(data: LevelUpData) {
  levelData.value = data
  visible.value = true

  // 播放音效（如果有）
  playLevelUpSound()

  // 3秒后自动关闭
  setTimeout(() => {
    hide()
  }, 3000)
}

// 关闭动画
function hide() {
  visible.value = false
  setTimeout(() => {
    levelData.value = null
  }, 500)
}

// 播放升级音效
function playLevelUpSound() {
  // 可以在这里添加音效播放逻辑
  try {
    const audio = new Audio('/sounds/levelup.mp3')
    audio.volume = 0.5
    audio.play().catch(() => {
      // 忽略音频播放错误
    })
  } catch (error) {
    // 忽略错误
  }
}

// 暴露方法
defineExpose({
  show,
  hide
})
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-all duration-500 ease-out"
      enter-from-class="opacity-0 scale-50"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition-all duration-300 ease-in"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-50"
    >
      <div
        v-if="visible && levelData"
        class="fixed inset-0 z-[9999] flex items-center justify-center pointer-events-none"
      >
        <!-- 背景光效 -->
        <div class="absolute inset-0 bg-black/40 backdrop-blur-sm pointer-events-auto" @click="hide"></div>

        <!-- 主内容 -->
        <div class="relative z-10 pointer-events-none">
          <!-- 光环效果 -->
          <div
            class="absolute inset-0 -m-32 rounded-full blur-3xl opacity-60 animate-pulse"
            :class="`bg-gradient-to-r ${tierGradients[levelData.tier] || 'from-blue-400 to-cyan-500'}`"
          ></div>

          <!-- 卡片 -->
          <div class="relative bg-white dark:bg-[#0d0d10] rounded-3xl sm:rounded-[40px] p-8 sm:p-12 shadow-2xl border-2 overflow-hidden"
               :class="`border-${levelData.tier}`"
               style="min-width: 300px; max-width: 90vw">
            <!-- 顶部装饰条 -->
            <div
              class="absolute top-0 left-0 right-0 h-2 bg-gradient-to-r animate-shimmer"
              :class="tierGradients[levelData.tier] || 'from-blue-400 to-cyan-500'"
            ></div>

            <!-- 粒子效果 -->
            <div class="absolute inset-0 overflow-hidden pointer-events-none">
              <div
                v-for="i in 30"
                :key="i"
                class="absolute w-2 h-2 rounded-full animate-particle opacity-0"
                :class="`bg-gradient-to-br ${tierGradients[levelData.tier] || 'from-blue-400 to-cyan-500'}`"
                :style="{
                  left: Math.random() * 100 + '%',
                  animationDelay: Math.random() * 2 + 's',
                  animationDuration: (2 + Math.random() * 2) + 's'
                }"
              ></div>
            </div>

            <!-- 内容 -->
            <div class="relative z-10 text-center space-y-4 sm:space-y-6">
              <!-- 标题 -->
              <div class="space-y-2">
                <div class="inline-flex items-center gap-2 px-3 py-1.5 bg-blue-500/10 border border-blue-500/20 rounded-full">
                  <Zap class="w-4 h-4 text-blue-500 animate-pulse" />
                  <span class="text-xs font-black text-blue-500 uppercase tracking-widest">LEVEL UP!</span>
                </div>

                <h2 class="text-2xl sm:text-4xl font-black text-slate-900 dark:text-white tracking-tight">
                  恭喜升级！
                </h2>
              </div>

              <!-- 段位图标 -->
              <div class="flex justify-center">
                <div class="relative group">
                  <!-- 外圈光环 -->
                  <div
                    class="absolute -inset-8 rounded-full blur-2xl opacity-50 animate-pulse"
                    :class="`bg-gradient-to-r ${tierGradients[levelData.tier] || 'from-blue-400 to-cyan-500'}`"
                  ></div>

                  <!-- 主图标容器 -->
                  <div
                    class="relative w-28 h-28 sm:w-32 sm:h-32 rounded-3xl flex items-center justify-center shadow-2xl animate-bounce-slow"
                    :class="`bg-gradient-to-br ${tierGradients[levelData.tier] || 'from-blue-400 to-cyan-500'}`"
                  >
                    <div class="text-6xl sm:text-7xl animate-spin-slow">
                      {{ tierIcons[levelData.tier] || '⭐' }}
                    </div>
                  </div>

                  <!-- 装饰元素 -->
                  <div class="absolute -top-3 -right-3 w-12 h-12 bg-white dark:bg-slate-900 rounded-2xl flex items-center justify-center shadow-xl animate-float">
                    <TrendingUp class="w-6 h-6 text-green-500" />
                  </div>
                  <div class="absolute -bottom-3 -left-3 w-12 h-12 bg-white dark:bg-slate-900 rounded-2xl flex items-center justify-center shadow-xl animate-float" style="animation-delay: 0.5s;">
                    <Award class="w-6 h-6 text-yellow-500" />
                  </div>
                </div>
              </div>

              <!-- 等级信息 -->
              <div class="space-y-2">
                <div class="text-5xl sm:text-6xl font-black bg-gradient-to-r bg-clip-text text-transparent animate-gradient"
                     :class="tierGradients[levelData.tier] || 'from-blue-400 to-cyan-500'">
                  {{ levelData.tier_name }} {{ levelData.level }} 级
                </div>
                <p class="text-sm sm:text-base text-slate-600 dark:text-slate-400 font-medium">
                  你已成为 <span class="font-black text-blue-500">{{ levelData.tier_name }}</span> 段位研究员
                </p>
              </div>

              <!-- 提示文字 -->
              <div class="pt-4 text-xs text-slate-500 dark:text-slate-400">
                点击任意处继续
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
@keyframes particle {
  0% {
    transform: translateY(0) scale(0);
    opacity: 0;
  }
  50% {
    opacity: 1;
  }
  100% {
    transform: translateY(-200px) scale(1);
    opacity: 0;
  }
}

@keyframes bounce-slow {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-10px);
  }
}

@keyframes spin-slow {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@keyframes float {
  0%, 100% {
    transform: translateY(0px);
  }
  50% {
    transform: translateY(-10px);
  }
}

@keyframes shimmer {
  0% {
    background-position: -200% center;
  }
  100% {
    background-position: 200% center;
  }
}

@keyframes gradient {
  0%, 100% {
    background-size: 200% 200%;
    background-position: left center;
  }
  50% {
    background-size: 200% 200%;
    background-position: right center;
  }
}

.animate-particle {
  animation: particle 3s ease-out infinite;
}

.animate-bounce-slow {
  animation: bounce-slow 2s ease-in-out infinite;
}

.animate-spin-slow {
  animation: spin-slow 3s linear infinite;
}

.animate-float {
  animation: float 3s ease-in-out infinite;
}

.animate-shimmer {
  background-size: 200% 100%;
  animation: shimmer 2s linear infinite;
}

.animate-gradient {
  animation: gradient 3s ease infinite;
}
</style>
