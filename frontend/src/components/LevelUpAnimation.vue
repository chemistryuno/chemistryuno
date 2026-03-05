<script setup lang="ts">
import { ref, computed } from 'vue'
import feedback from '../utils/feedback'
import { 
  Zap, 
  Award, 
  Trophy, 
  Star, 
  Crown, 
  FlaskConical, 
  Atom, 
  Sparkles,
  ChevronRight
} from 'lucide-vue-next'

interface LevelUpData {
  level: number
  tier: string
  tier_name: string
  xp: number
  total_xp: number
}

const visible = ref(false)
const levelData = ref<LevelUpData | null>(null)
const animationStage = ref(0) // 0: hidden, 1: appearing, 2: content active, 3: particles

// 段位配置
const tierConfig: Record<string, { icon: any, color: string, gradient: string, glow: string }> = {
  bronze: { 
    icon: Award, 
    color: '#cd7f32', 
    gradient: 'from-orange-400 via-amber-600 to-orange-700',
    glow: 'shadow-orange-500/50'
  },
  silver: { 
    icon: FlaskConical, 
    color: '#c0c0c0', 
    gradient: 'from-slate-300 via-slate-400 to-slate-500',
    glow: 'shadow-slate-400/50'
  },
  gold: { 
    icon: Star, 
    color: '#ffd700', 
    gradient: 'from-yellow-300 via-yellow-500 to-amber-500',
    glow: 'shadow-yellow-500/50'
  },
  platinum: { 
    icon: Atom, 
    color: '#e5e4e2', 
    gradient: 'from-cyan-300 via-blue-400 to-indigo-500',
    glow: 'shadow-cyan-400/50'
  },
  diamond: { 
    icon: Trophy, 
    color: '#b9f2ff', 
    gradient: 'from-blue-400 via-indigo-500 to-purple-600',
    glow: 'shadow-indigo-500/50'
  },
  master: { 
    icon: Crown, 
    color: '#ff00ff', 
    gradient: 'from-purple-500 via-pink-600 to-red-600',
    glow: 'shadow-purple-500/50'
  }
}

// 显示升级动画
async function show(data: LevelUpData) {
  levelData.value = data
  visible.value = true
  animationStage.value = 1
  
  // 播放音效
  feedback.levelUp()

  // 动画阶段演进
  setTimeout(() => { animationStage.value = 2 }, 100)
  setTimeout(() => { animationStage.value = 3 }, 600)

  // 5秒后自动关闭
  setTimeout(() => {
    hide()
  }, 5000)
}

// 关闭动画
function hide() {
  animationStage.value = 0
  setTimeout(() => {
    visible.value = false
    levelData.value = null
  }, 500)
}

// 暴露方法
defineExpose({
  show,
  hide
})

const currentTier = computed(() => levelData.value ? tierConfig[levelData.value.tier] || tierConfig.bronze : tierConfig.bronze)
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-all duration-700 cubic-bezier(0.34, 1.56, 0.64, 1)"
      enter-from-class="opacity-0 scale-90 backdrop-blur-0"
      enter-to-class="opacity-100 scale-100 backdrop-blur-md"
      leave-active-class="transition-all duration-500 ease-in"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-75"
    >
      <div
        v-if="visible && levelData"
        class="fixed inset-0 z-[9999] flex items-center justify-center pointer-events-none"
      >
        <!-- 背景高斯模糊叠加层 -->
        <div class="absolute inset-0 bg-slate-950/40 backdrop-blur-sm pointer-events-auto" @click="hide"></div>

        <!-- 环境光效 (Ambient Light) -->
        <div 
          class="absolute inset-0 pointer-events-none opacity-40 mix-blend-screen"
          :class="`bg-gradient-to-tr ${currentTier.gradient}`"
          style="mask-image: radial-gradient(circle at center, transparent 30%, black 100%);"
        ></div>

        <!-- 主容器 -->
        <div class="relative z-10 w-full max-w-lg px-6 pointer-events-none">
          
          <!-- 大背景光晕 -->
          <div 
            class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[120%] h-[120%] blur-[120px] opacity-20 transition-all duration-1000"
            :class="animationStage >= 2 ? 'scale-110 opacity-30' : 'scale-50 opacity-0'"
            :style="{ background: `radial-gradient(circle, ${currentTier.color}, transparent)` }"
          ></div>

          <!-- 升级卡片 -->
          <div 
            class="relative overflow-hidden rounded-[48px] border border-white/20 dark:border-white/10 shadow-[0_32px_128px_-16px_rgba(0,0,0,0.5)] transition-all duration-700 ease-out"
            :class="[
              animationStage >= 1 ? 'translate-y-0 opacity-100' : 'translate-y-20 opacity-0',
              'bg-[#121216]/80 backdrop-blur-2xl'
            ]"
          >
            <!-- 能量脉冲线条 -->
            <div class="absolute inset-0 pointer-events-none overflow-hidden opacity-30">
               <div v-for="i in 2" :key="i" 
                    class="absolute w-full h-[1px] bg-gradient-to-r from-transparent via-white/20 to-transparent animate-pulse-line"
                    :style="{ top: (i * 33) + '%', animationDelay: (i * 0.5) + 's' }"></div>
            </div>

            <div class="relative p-8 pt-12 flex flex-col items-center text-center">
              
              <!-- 顶部状态标签 -->
              <div 
                class="absolute top-6 left-1/2 -translate-x-1/2 flex items-center gap-2 px-3 py-1 rounded-full border bg-white/5 transition-all duration-500"
                :class="[
                  animationStage >= 2 ? 'opacity-100 translate-y-0' : 'opacity-0 -translate-y-4',
                  `border-${levelData.tier}/20 text-${levelData.tier}`
                ]"
              >
                <Sparkles class="w-3 h-3 animate-spin-slow text-blue-400" />
                <span class="text-[9px] font-black uppercase tracking-[0.2em] text-blue-400">Level Upgraded</span>
              </div>

              <!-- 等级徽章/图标容器 -->
              <div class="relative mb-6">
                <!-- 旋转环 -->
                <div 
                  class="absolute -inset-6 rounded-full border border-dashed border-white/10 animate-spin-slow opacity-20"
                ></div>

                <!-- 核心图标 -->
                <div 
                  class="relative w-24 h-32 sm:w-28 sm:h-36 flex flex-col items-center justify-center transition-all duration-700"
                  :class="animationStage >= 2 ? 'scale-100 rotate-0' : 'scale-75 rotate-12'"
                >
                  <div 
                    class="absolute inset-0 blur-[30px] opacity-30 transition-all duration-1000"
                    :class="animationStage >= 3 ? 'scale-125' : 'scale-50'"
                    :style="{ background: `radial-gradient(circle, ${currentTier.color}, transparent)` }"
                  ></div>
                  
                  <component 
                    :is="currentTier.icon" 
                    class="w-16 h-16 sm:w-20 sm:h-20 text-white filter drop-shadow-[0_0_15px_rgba(255,255,255,0.4)] transition-all duration-500"
                    :class="animationStage >= 3 ? 'scale-110' : 'scale-90'"
                  />
                  
                  <div class="mt-2 flex flex-col items-center">
                    <div class="text-[8px] font-black uppercase tracking-[0.3em] opacity-40 text-white">Rank Status</div>
                    <div class="text-xl font-black text-white tracking-tight uppercase">{{ levelData.tier_name }}</div>
                  </div>
                </div>
              </div>

              <!-- 核心文字信息 -->
              <div class="space-y-4 w-full mb-6">
                <div class="space-y-1">
                   <h2 
                    class="text-4xl sm:text-5xl font-black tracking-tighter text-white"
                    :class="animationStage >= 2 ? 'opacity-100' : 'opacity-0'"
                  >
                    LEVEL <span class="bg-gradient-to-b from-white to-white/60 bg-clip-text text-transparent">{{ levelData.level }}</span>
                  </h2>
                </div>

                <div 
                  class="flex flex-col items-center gap-2 transition-all duration-500 delay-300"
                  :class="animationStage >= 2 ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'"
                >
                  <p class="text-xs text-white/40 font-bold uppercase tracking-[0.15em]">基因序列进化完成</p>
                </div>
              </div>

              <!-- 操作按钮 -->
              <button 
                @click="hide"
                class="pointer-events-auto group relative flex items-center gap-3 px-8 py-3 bg-white text-black font-black rounded-2xl transition-all hover:scale-105 active:scale-95 overflow-hidden"
              >
                <div class="absolute inset-x-0 bottom-0 h-1 bg-gradient-to-r from-transparent via-black/10 to-transparent"></div>
                <span class="uppercase tracking-widest text-xs">接受进化</span>
                <ChevronRight class="w-4 h-4 group-hover:translate-x-1 transition-transform" />
              </button>
            </div>
          </div>

              <!-- 统计数值展示 -->
              <div 
                class="grid grid-cols-2 gap-4 w-full mt-12 transition-all duration-500 delay-500"
                :class="animationStage >= 3 ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-8'"
              >
                <div class="bg-white/5 border border-white/10 rounded-[28px] p-5">
                   <div class="flex items-center justify-center gap-2 mb-1">
                     <Zap class="w-3.5 h-3.5 text-yellow-400" />
                     <span class="text-[10px] font-black uppercase text-white/40 tracking-widest">Total XP</span>
                   </div>
                   <div class="text-xl font-black text-white font-mono">{{ levelData.total_xp.toLocaleString() }}</div>
                </div>
                <div class="bg-white/5 border border-white/10 rounded-[28px] p-5">
                   <div class="flex items-center justify-center gap-2 mb-1">
                     <Trophy class="w-3.5 h-3.5 text-blue-400" />
                     <span class="text-[10px] font-black uppercase text-white/40 tracking-widest">New Rank</span>
                     <ChevronRight class="w-3 h-3 text-white/20" />
                   </div>
                   <div class="text-xl font-black text-white">{{ levelData.tier_name }}</div>
                </div>
              </div>

              <!-- 装饰性粒子容器 -->
              <div v-if="animationStage >= 3" class="absolute inset-0 pointer-events-none">
                 <div v-for="i in 20" :key="i" 
                      class="absolute w-1 h-1 rounded-full animate-float-particle"
                      :style="{ 
                        left: Math.random() * 100 + '%', 
                        top: Math.random() * 100 + '%',
                        background: currentTier.color,
                        animationDelay: (Math.random() * 2) + 's',
                        animationDuration: (3 + Math.random() * 4) + 's',
                        opacity: 0.3 + (Math.random() * 0.4)
                      }"></div>
              </div>

              <!-- 交互提示 -->
              <div 
                class="mt-12 text-[10px] font-black uppercase tracking-[0.5em] text-white/20 animate-pulse transition-opacity duration-1000"
                :class="animationStage >= 3 ? 'opacity-100' : 'opacity-0'"
              >
                TOUCH_TO_CONTINUED
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
@keyframes shine {
  0% { transform: translateX(-100%) skewX(-15deg); }
  100% { transform: translateX(200%) skewX(-15deg); }
}

@keyframes spin-reverse {
  from { transform: rotate(360deg); }
  to { transform: rotate(0deg); }
}

@keyframes pulse-line {
  0%, 100% { opacity: 0; transform: scaleX(0.8); }
  50% { opacity: 1; transform: scaleX(1); }
}

@keyframes float-particle {
  0%, 100% { transform: translate(0, 0); opacity: 0; }
  25% { opacity: 1; }
  50% { transform: translate(20px, -40px); }
  75% { opacity: 1; }
}

.animate-shine {
  animation: shine 3s cubic-bezier(0.4, 0, 0.2, 1) infinite;
}

.animate-spin-slow {
  animation: spin 8s linear infinite;
}

.animate-spin-reverse {
  animation: spin-reverse 12s linear infinite;
}

.animate-pulse-line {
  animation: pulse-line 4s ease-in-out infinite;
}

.animate-float-particle {
  animation: float-particle 6s ease-in-out infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
