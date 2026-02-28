<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { X, ChevronLeft, ChevronRight, Sparkles } from 'lucide-vue-next'

interface TutorialStep {
  id: string
  titlePlaceholder: string
  contentPlaceholder: string
  targetSelector?: string  // 高亮目标的CSS选择器
  position?: 'top' | 'bottom' | 'left' | 'right' | 'center'
  spotlightRadius?: number  // 聚光灯半径
}

const props = defineProps<{
  show: boolean
  steps?: TutorialStep[]
}>()

const emit = defineEmits<{
  close: []
  complete: []
}>()

// 默认步骤
const defaultSteps: TutorialStep[] = [
  {
    id: 'welcome',
    titlePlaceholder: '欢迎来到实验室！',
    contentPlaceholder: '欢迎来到化学UNO！这是一场融合化学知识与策略的卡牌对决。接下来将带你快速了解游戏的基本操作，准备好了吗？',
    position: 'center'
  },
  {
    id: 'hand-cards',
    titlePlaceholder: '你的元素库',
    contentPlaceholder: '这里是你的手牌区，存放着你持有的所有化学元素与化合物卡牌。点击卡牌可以选中并出牌，合理搭配元素，触发更强的化学反应！',
    targetSelector: '.hand-container-mobile',
    position: 'top',
    spotlightRadius: 180
  },
  {
    id: 'operation-area',
    titlePlaceholder: '实验操作台',
    contentPlaceholder: '这是你的操作中心。在输入框手动注入化学式精准出牌，或点击「摸牌」按钮补充手牌。若遇到「摸牌惩罚」，按钮会显示需要摸取的张数，务必承担！',
    targetSelector: '.operation-area',
    position: 'bottom',
    spotlightRadius: 200
  },
  {
    id: 'center-play',
    titlePlaceholder: '反应堆——出牌区',
    contentPlaceholder: '这里显示场上的当前牌，也是化学反应的发生地。你需要打出与当前牌匹配的物质，或触发特定化学反应改变局势。最先打完手牌的玩家获胜！',
    targetSelector: '.center-play-area',
    position: 'bottom',
    spotlightRadius: 220
  },
  {
    id: 'complete',
    titlePlaceholder: '准备好了吗，化学家？',
    contentPlaceholder: '太棒了！你已掌握基本操作。记住：善用化学反应是制胜关键。进入教学关卡，跟随系统指引完成第一局对战，祝实验顺利！',
    position: 'center'
  }
]

const tutorialSteps = computed(() => props.steps || defaultSteps)
const currentStep = ref(0)
const isVisible = ref(false)
const spotlightStyle = ref<any>({})
const tooltipStyle = ref<any>({})
const clamp = (value: number, min: number, max: number) => Math.min(Math.max(value, min), max)
const clampRange = (value: number, min: number, max: number) => clamp(value, Math.min(min, max), Math.max(min, max))

// 计算聚光灯位置
const calculateSpotlight = async () => {
  await nextTick()
  const step = tutorialSteps.value[currentStep.value]

  if (!step.targetSelector) {
    spotlightStyle.value = {}
    tooltipStyle.value = { position: 'center' }
    return
  }

  const targetCandidates = Array.from(document.querySelectorAll(step.targetSelector))
  const target = targetCandidates.find((el) => {
    const rect = (el as HTMLElement).getBoundingClientRect()
    return rect.width > 0 && rect.height > 0
  }) || targetCandidates[0]

  if (!target) {
    spotlightStyle.value = {}
    tooltipStyle.value = { position: 'center' }
    return
  }

  const rect = target.getBoundingClientRect()
  if (rect.width === 0 || rect.height === 0) {
    spotlightStyle.value = {}
    tooltipStyle.value = { position: 'center' }
    return
  }

  // 精确匹配元素尺寸，添加适度的padding（16px）
  const padding = 16
  spotlightStyle.value = {
    left: `${rect.left + rect.width / 2}px`,
    top: `${rect.top + rect.height / 2}px`,
    width: `${rect.width + padding * 2}px`,
    height: `${rect.height + padding * 2}px`,
  }

  // 计算提示框位置（带边界检测）
  const position = step.position || 'bottom'
  const tooltipOffset = 20
  const viewportWidth = window.innerWidth
  const viewportHeight = window.innerHeight
  const edgePadding = 16 // 距离屏幕边缘的最小距离
  const tooltipEl = document.querySelector('.tutorial-tooltip') as HTMLElement | null
  const tooltipMeasuredWidth = tooltipEl?.offsetWidth || 0
  const tooltipMeasuredHeight = tooltipEl?.offsetHeight || 0
  const tooltipEstimatedWidth = tooltipMeasuredWidth || Math.min(320, viewportWidth - edgePadding * 2)
  const tooltipEstimatedHeight = tooltipMeasuredHeight || 200

  let finalPosition = position
  let tooltipLeft = 0
  let tooltipTop = 0
  let transform = ''
  let arrowPosition = 'top'

  // 根据初始位置计算坐标
  switch (position) {
    case 'top':
      tooltipLeft = rect.left + rect.width / 2
      tooltipTop = rect.top - tooltipOffset
      transform = 'translate(-50%, -100%)'
      arrowPosition = 'bottom'
      // 检查是否超出顶部
      if (tooltipTop - tooltipEstimatedHeight < edgePadding) {
        finalPosition = 'bottom'
      }
      break
    case 'bottom':
      tooltipLeft = rect.left + rect.width / 2
      tooltipTop = rect.bottom + tooltipOffset
      transform = 'translate(-50%, 0)'
      arrowPosition = 'top'
      // 检查是否超出底部
      if (tooltipTop + tooltipEstimatedHeight > viewportHeight - edgePadding) {
        finalPosition = 'top'
      }
      break
    case 'left':
      tooltipLeft = rect.left - tooltipOffset
      tooltipTop = rect.top + rect.height / 2
      transform = 'translate(-100%, -50%)'
      arrowPosition = 'right'
      // 检查是否超出左侧
      if (tooltipLeft - tooltipEstimatedWidth < edgePadding) {
        finalPosition = 'right'
      }
      break
    case 'right':
      tooltipLeft = rect.right + tooltipOffset
      tooltipTop = rect.top + rect.height / 2
      transform = 'translate(0, -50%)'
      arrowPosition = 'left'
      // 检查是否超出右侧
      if (tooltipLeft + tooltipEstimatedWidth > viewportWidth - edgePadding) {
        finalPosition = 'left'
      }
      break
  }

  // 如果位置需要调整，重新计算
  if (finalPosition !== position) {
    switch (finalPosition) {
      case 'top':
        tooltipLeft = rect.left + rect.width / 2
        tooltipTop = rect.top - tooltipOffset
        transform = 'translate(-50%, -100%)'
        arrowPosition = 'bottom'
        break
      case 'bottom':
        tooltipLeft = rect.left + rect.width / 2
        tooltipTop = rect.bottom + tooltipOffset
        transform = 'translate(-50%, 0)'
        arrowPosition = 'top'
        break
      case 'left':
        tooltipLeft = rect.left - tooltipOffset
        tooltipTop = rect.top + rect.height / 2
        transform = 'translate(-100%, -50%)'
        arrowPosition = 'right'
        break
      case 'right':
        tooltipLeft = rect.right + tooltipOffset
        tooltipTop = rect.top + rect.height / 2
        transform = 'translate(0, -50%)'
        arrowPosition = 'left'
        break
    }
  }

  let arrowOffsetX = 0
  let arrowOffsetY = 0

  // 最终边界修正：确保提示框在屏幕内，并保持箭头尽量指向目标
  const halfWidth = tooltipEstimatedWidth / 2
  const halfHeight = tooltipEstimatedHeight / 2
  const maxArrowOffsetX = Math.max(0, halfWidth - 14)
  const maxArrowOffsetY = Math.max(0, halfHeight - 14)

  switch (finalPosition) {
    case 'top': {
      const anchoredLeft = tooltipLeft
      tooltipLeft = clampRange(tooltipLeft, edgePadding + halfWidth, viewportWidth - edgePadding - halfWidth)
      arrowOffsetX = clamp(anchoredLeft - tooltipLeft, -maxArrowOffsetX, maxArrowOffsetX)
      tooltipTop = clampRange(tooltipTop, edgePadding + tooltipEstimatedHeight, viewportHeight - edgePadding)
      break
    }
    case 'bottom': {
      const anchoredLeft = tooltipLeft
      tooltipLeft = clampRange(tooltipLeft, edgePadding + halfWidth, viewportWidth - edgePadding - halfWidth)
      arrowOffsetX = clamp(anchoredLeft - tooltipLeft, -maxArrowOffsetX, maxArrowOffsetX)
      tooltipTop = clampRange(tooltipTop, edgePadding, viewportHeight - edgePadding - tooltipEstimatedHeight)
      break
    }
    case 'left': {
      tooltipLeft = clampRange(tooltipLeft, edgePadding + tooltipEstimatedWidth, viewportWidth - edgePadding)
      const anchoredTop = tooltipTop
      tooltipTop = clampRange(tooltipTop, edgePadding + halfHeight, viewportHeight - edgePadding - halfHeight)
      arrowOffsetY = clamp(anchoredTop - tooltipTop, -maxArrowOffsetY, maxArrowOffsetY)
      break
    }
    case 'right': {
      tooltipLeft = clampRange(tooltipLeft, edgePadding, viewportWidth - edgePadding - tooltipEstimatedWidth)
      const anchoredTop = tooltipTop
      tooltipTop = clampRange(tooltipTop, edgePadding + halfHeight, viewportHeight - edgePadding - halfHeight)
      arrowOffsetY = clamp(anchoredTop - tooltipTop, -maxArrowOffsetY, maxArrowOffsetY)
      break
    }
  }

  tooltipStyle.value = {
    left: `${tooltipLeft}px`,
    top: `${tooltipTop}px`,
    transform,
    arrowPosition,
    arrowOffsetX: `${arrowOffsetX}px`,
    arrowOffsetY: `${arrowOffsetY}px`
  }
}

// 下一步
const nextStep = async () => {
  if (currentStep.value < tutorialSteps.value.length - 1) {
    currentStep.value++
    await calculateSpotlight()
  } else {
    completeTutorial()
  }
}

// 上一步
const prevStep = async () => {
  if (currentStep.value > 0) {
    currentStep.value--
    await calculateSpotlight()
  }
}

// 跳过教程
const skipTutorial = () => {
  emit('close')
  isVisible.value = false
}

// 完成教程
const completeTutorial = () => {
  emit('complete')
  isVisible.value = false
}

// 监听显示状态
watch(() => props.show, async (newVal) => {
  if (newVal) {
    isVisible.value = true
    currentStep.value = 0
    await calculateSpotlight()
  } else {
    isVisible.value = false
  }
}, { immediate: true })

// 监听步骤变化
watch(currentStep, () => {
  calculateSpotlight()
})

// 监听窗口大小变化
onMounted(() => {
  window.addEventListener('resize', calculateSpotlight)
})

onUnmounted(() => {
  window.removeEventListener('resize', calculateSpotlight)
})
</script>

<template>
  <Teleport to="body">
    <Transition name="tutorial-fade">
      <div
        v-if="isVisible"
        class="tutorial-overlay fixed inset-0 z-[9999]"
        @click.self="skipTutorial"
      >
        <!-- 聚光灯高亮区域 -->
        <Transition name="spotlight">
          <div
            v-if="spotlightStyle.left"
            class="tutorial-spotlight absolute pointer-events-none"
            :style="{
              left: spotlightStyle.left,
              top: spotlightStyle.top,
              width: spotlightStyle.width,
              height: spotlightStyle.height,
              transform: 'translate(-50%, -50%)'
            }"
          >
            <!-- 发光边框动画 -->
            <div class="absolute inset-0 rounded-3xl border-4 border-cyan-400/80 animate-pulse-slow"></div>
            <div class="absolute inset-0 rounded-3xl shadow-[0_0_60px_20px_rgba(34,211,238,0.6)]"></div>

            <!-- 扫描线效果 -->
            <div class="absolute inset-0 rounded-3xl overflow-hidden">
              <div class="scan-line absolute inset-x-0 h-1 bg-gradient-to-r from-transparent via-cyan-400 to-transparent"></div>
            </div>

            <!-- 角落装饰 -->
            <div class="corner-tl absolute top-0 left-0 w-8 h-8 border-t-4 border-l-4 border-cyan-400"></div>
            <div class="corner-tr absolute top-0 right-0 w-8 h-8 border-t-4 border-r-4 border-cyan-400"></div>
            <div class="corner-bl absolute bottom-0 left-0 w-8 h-8 border-b-4 border-l-4 border-cyan-400"></div>
            <div class="corner-br absolute bottom-0 right-0 w-8 h-8 border-b-4 border-r-4 border-cyan-400"></div>
          </div>
        </Transition>

        <!-- 提示卡片 -->
        <Transition :name="tooltipStyle.position === 'center' ? 'tooltip-scale' : 'tooltip-slide'">
          <div
            v-if="tutorialSteps[currentStep]"
            class="tutorial-tooltip absolute"
            :class="tooltipStyle.position === 'center' ? 'tooltip-center' : ''"
            :style="tooltipStyle.position === 'center' ? {} : {
              left: tooltipStyle.left,
              top: tooltipStyle.top,
              transform: tooltipStyle.transform,
              '--arrow-offset-x': tooltipStyle.arrowOffsetX || '0px',
              '--arrow-offset-y': tooltipStyle.arrowOffsetY || '0px'
            }"
          >
            <!-- 箭头指示器 -->
            <div
              v-if="tooltipStyle.arrowPosition"
              class="tooltip-arrow absolute w-0 h-0"
              :class="{
                'arrow-top': tooltipStyle.arrowPosition === 'top',
                'arrow-bottom': tooltipStyle.arrowPosition === 'bottom',
                'arrow-left': tooltipStyle.arrowPosition === 'left',
                'arrow-right': tooltipStyle.arrowPosition === 'right'
              }"
            ></div>

            <!-- 卡片内容 -->
            <div class="tutorial-card relative w-[90vw] max-w-md sm:max-w-lg bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 rounded-3xl shadow-[0_20px_60px_-15px_rgba(0,0,0,0.8)] border-2 border-cyan-400/40 overflow-hidden ring-4 ring-black/20">
              <!-- 背景装饰 -->
              <div class="absolute inset-0 bg-[radial-gradient(circle_at_50%_120%,rgba(34,211,238,0.1),transparent_50%)]"></div>
              <div class="absolute top-0 right-0 w-64 h-64 bg-cyan-500/5 rounded-full blur-3xl"></div>

              <!-- 顶部装饰条 -->
              <div class="absolute top-0 inset-x-0 h-1 bg-gradient-to-r from-transparent via-cyan-400 to-transparent"></div>

              <!-- 内容区 -->
              <div class="relative p-6 sm:p-8">
                <!-- 头部 -->
                <div class="flex items-start justify-between mb-6">
                  <div class="flex items-center gap-3">
                    <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-cyan-400 to-blue-500 flex items-center justify-center shadow-lg shadow-cyan-500/30">
                      <Sparkles class="w-5 h-5 text-white" />
                    </div>
                    <div>
                      <h3 class="text-xl sm:text-2xl font-black text-white tracking-tight">
                        {{ tutorialSteps[currentStep].titlePlaceholder }}
                      </h3>
                      <div class="flex items-center gap-2 mt-1">
                        <span class="text-xs font-bold text-cyan-400">Step {{ currentStep + 1 }}/{{ tutorialSteps.length }}</span>
                        <div class="flex gap-1">
                          <div
                            v-for="(_, index) in tutorialSteps"
                            :key="index"
                            class="w-6 h-1 rounded-full transition-all duration-300"
                            :class="index === currentStep ? 'bg-cyan-400' : 'bg-slate-700'"
                          ></div>
                        </div>
                      </div>
                    </div>
                  </div>
                  <button
                    @click="skipTutorial"
                    class="w-8 h-8 rounded-lg bg-slate-800/50 hover:bg-slate-700 flex items-center justify-center transition-colors border border-slate-700 hover:border-slate-600"
                  >
                    <X class="w-4 h-4 text-slate-400" />
                  </button>
                </div>

                <!-- 内容 -->
                <p class="text-slate-300 text-sm sm:text-base leading-relaxed mb-8">
                  {{ tutorialSteps[currentStep].contentPlaceholder }}
                </p>

                <!-- 底部按钮 -->
                <div class="flex items-center justify-between gap-4">
                  <button
                    v-if="currentStep > 0"
                    @click="prevStep"
                    class="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-slate-800/50 hover:bg-slate-700 text-slate-300 font-bold text-sm transition-all border border-slate-700 hover:border-slate-600"
                  >
                    <ChevronLeft class="w-4 h-4" />
                    <span>Previous</span>
                  </button>
                  <button
                    v-else
                    @click="skipTutorial"
                    class="px-4 py-2.5 rounded-xl bg-slate-800/50 hover:bg-slate-700 text-slate-400 font-bold text-sm transition-all border border-slate-700 hover:border-slate-600"
                  >
                    Skip Tutorial
                  </button>

                  <button
                    @click="nextStep"
                    class="flex-1 flex items-center justify-center gap-2 px-6 py-3 rounded-xl bg-gradient-to-r from-cyan-500 to-blue-500 hover:from-cyan-400 hover:to-blue-400 text-white font-black text-sm transition-all shadow-lg shadow-cyan-500/30 hover:shadow-cyan-500/50 hover:scale-105"
                  >
                    <span>{{ currentStep === tutorialSteps.length - 1 ? 'Get Started' : 'Next' }}</span>
                    <ChevronRight v-if="currentStep < tutorialSteps.length - 1" class="w-4 h-4" />
                  </button>
                </div>
              </div>

              <!-- 粒子装饰 -->
              <div class="absolute top-4 left-4 w-2 h-2 rounded-full bg-cyan-400 animate-ping opacity-30"></div>
              <div class="absolute bottom-8 right-8 w-1.5 h-1.5 rounded-full bg-blue-400 animate-ping opacity-20" style="animation-delay: 0.5s;"></div>
              <div class="absolute top-1/2 right-4 w-1 h-1 rounded-full bg-cyan-300 animate-ping opacity-25" style="animation-delay: 1s;"></div>
            </div>
          </div>
        </Transition>

        <!-- 粒子背景效果 -->
        <div class="particles-container absolute inset-0 pointer-events-none overflow-hidden">
          <div v-for="i in 20" :key="i" class="particle" :style="{ animationDelay: `${i * 0.3}s` }"></div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
/* 遮罩动画 */
.tutorial-fade-enter-active,
.tutorial-fade-leave-active {
  transition: opacity 0.4s ease;
}

.tutorial-fade-enter-from,
.tutorial-fade-leave-to {
  opacity: 0;
}

/* 聚光灯动画 */
.spotlight-enter-active,
.spotlight-leave-active {
  transition: all 0.5s cubic-bezier(0.4, 0, 0.2, 1);
}

.spotlight-enter-from,
.spotlight-leave-to {
  opacity: 0;
  transform: translate(-50%, -50%) scale(0.8);
}

/* 扫描线动画 */
@keyframes scan {
  0% {
    top: 0;
    opacity: 0;
  }
  10% {
    opacity: 1;
  }
  90% {
    opacity: 1;
  }
  100% {
    top: 100%;
    opacity: 0;
  }
}

.scan-line {
  animation: scan 3s ease-in-out infinite;
}

/* 角落装饰动画 */
.corner-tl,
.corner-tr,
.corner-bl,
.corner-br {
  animation: corner-pulse 2s ease-in-out infinite;
}

@keyframes corner-pulse {
  0%, 100% {
    opacity: 0.4;
  }
  50% {
    opacity: 1;
  }
}

/* 提示卡片动画 - 居中缩放 */
.tooltip-scale-enter-active,
.tooltip-scale-leave-active {
  transition: all 0.5s cubic-bezier(0.68, -0.55, 0.265, 1.55);
}

.tooltip-scale-enter-from,
.tooltip-scale-leave-to {
  opacity: 0;
  transform: scale(0.7) translateY(-20px);
}

.tooltip-center {
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

/* 提示卡片动画 - 滑入 */
.tooltip-slide-enter-active,
.tooltip-slide-leave-active {
  transition: all 0.4s cubic-bezier(0.68, -0.55, 0.265, 1.55);
}

.tooltip-slide-enter-from,
.tooltip-slide-leave-to {
  opacity: 0;
  transform: var(--slide-transform, translate(-50%, -100%)) scale(0.9);
}

/* 箭头样式 */
.arrow-top {
  left: calc(50% + var(--arrow-offset-x, 0px));
  top: -8px;
  transform: translateX(-50%);
  border-left: 12px solid transparent;
  border-right: 12px solid transparent;
  border-bottom: 12px solid rgba(15, 23, 42, 0.95);
}

.arrow-bottom {
  left: calc(50% + var(--arrow-offset-x, 0px));
  bottom: -8px;
  transform: translateX(-50%);
  border-left: 12px solid transparent;
  border-right: 12px solid transparent;
  border-top: 12px solid rgba(15, 23, 42, 0.95);
}

.arrow-left {
  left: -8px;
  top: calc(50% + var(--arrow-offset-y, 0px));
  transform: translateY(-50%);
  border-top: 12px solid transparent;
  border-bottom: 12px solid transparent;
  border-right: 12px solid rgba(15, 23, 42, 0.95);
}

.arrow-right {
  right: -8px;
  top: calc(50% + var(--arrow-offset-y, 0px));
  transform: translateY(-50%);
  border-top: 12px solid transparent;
  border-bottom: 12px solid transparent;
  border-left: 12px solid rgba(15, 23, 42, 0.95);
}

/* 粒子效果 */
.particle {
  position: absolute;
  width: 2px;
  height: 2px;
  background: rgba(34, 211, 238, 0.6);
  border-radius: 50%;
  animation: particle-float 8s ease-in-out infinite;
}

@keyframes particle-float {
  0% {
    transform: translate(0, 100vh) scale(0);
    opacity: 0;
  }
  10% {
    opacity: 1;
  }
  90% {
    opacity: 1;
  }
  100% {
    transform: translate(calc(100vw * (var(--random, 0.5) - 0.5)), -100px) scale(1);
    opacity: 0;
  }
}

.particle:nth-child(odd) {
  --random: 0.3;
}

.particle:nth-child(even) {
  --random: 0.7;
}

/* 脉冲动画 */
@keyframes pulse-slow {
  0%, 100% {
    opacity: 0.4;
    transform: scale(1);
  }
  50% {
    opacity: 0.8;
    transform: scale(1.02);
  }
}

.animate-pulse-slow {
  animation: pulse-slow 3s ease-in-out infinite;
}
</style>
