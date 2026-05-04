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
            <div class="tutorial-spotlight-frame absolute inset-0"></div>
            <div class="tutorial-spotlight-label absolute -top-8 left-0">当前区域</div>

            <div class="corner-tl absolute top-0 left-0 w-7 h-7"></div>
            <div class="corner-tr absolute top-0 right-0 w-7 h-7"></div>
            <div class="corner-bl absolute bottom-0 left-0 w-7 h-7"></div>
            <div class="corner-br absolute bottom-0 right-0 w-7 h-7"></div>
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
            <div class="tutorial-card relative w-[90vw] max-w-md sm:max-w-lg bg-white dark:bg-slate-950 rounded-lg shadow-[0_18px_48px_-24px_rgba(15,23,42,0.7)] dark:shadow-[0_18px_48px_-24px_rgba(0,0,0,0.95)] border border-slate-200 dark:border-slate-700 overflow-hidden">
              <!-- 顶部装饰条 -->
              <div class="absolute top-0 inset-x-0 h-1 bg-blue-600 dark:bg-cyan-400"></div>

              <!-- 内容区 -->
              <div class="relative p-5 sm:p-6">
                <!-- 头部 -->
                <div class="flex items-start justify-between gap-4 mb-5">
                  <div class="flex items-center gap-3">
                    <div class="w-10 h-10 rounded-lg bg-blue-600 dark:bg-cyan-500 flex items-center justify-center shrink-0">
                      <Sparkles class="w-5 h-5 text-white" />
                    </div>
                    <div class="min-w-0">
                      <h3 class="text-lg sm:text-xl font-black text-slate-900 dark:text-white tracking-tight leading-tight">
                        {{ tutorialSteps[currentStep].titlePlaceholder }}
                      </h3>
                      <div class="flex items-center gap-2 mt-2">
                        <span class="text-[11px] font-black text-blue-700 dark:text-cyan-300 uppercase">Step {{ currentStep + 1 }}/{{ tutorialSteps.length }}</span>
                        <div class="flex gap-1 min-w-0">
                          <div
                            v-for="(_, index) in tutorialSteps"
                            :key="index"
                            class="w-5 h-1 rounded-full transition-all duration-300"
                            :class="index === currentStep ? 'bg-blue-600 dark:bg-cyan-400' : 'bg-slate-200 dark:bg-slate-700'"
                          ></div>
                        </div>
                      </div>
                    </div>
                  </div>
                  <button
                    @click="skipTutorial"
                    class="w-8 h-8 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-900 dark:hover:bg-slate-800 flex items-center justify-center transition-colors border border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600 shrink-0"
                    aria-label="关闭向导"
                  >
                    <X class="w-4 h-4 text-slate-500 dark:text-slate-400" />
                  </button>
                </div>

                <!-- 内容 -->
                <p class="text-slate-600 dark:text-slate-300 text-sm sm:text-base leading-relaxed mb-6">
                  {{ tutorialSteps[currentStep].contentPlaceholder }}
                </p>

                <!-- 底部按钮 -->
                <div class="flex items-center justify-between gap-3">
                  <button
                    v-if="currentStep > 0"
                    @click="prevStep"
                    class="flex items-center gap-2 px-4 py-2.5 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-900 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-300 font-bold text-sm transition-all border border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600"
                  >
                    <ChevronLeft class="w-4 h-4" />
                    <span>上一步</span>
                  </button>
                  <button
                    v-else
                    @click="skipTutorial"
                    class="px-4 py-2.5 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-900 dark:hover:bg-slate-800 text-slate-500 dark:text-slate-400 font-bold text-sm transition-all border border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600"
                  >
                    跳过
                  </button>

                  <button
                    @click="nextStep"
                    class="flex-1 flex items-center justify-center gap-2 px-6 py-3 rounded-lg bg-blue-600 hover:bg-blue-500 dark:bg-cyan-500 dark:hover:bg-cyan-400 text-white dark:text-slate-950 font-black text-sm transition-all shadow-lg shadow-blue-500/20 dark:shadow-cyan-500/20"
                  >
                    <span>{{ currentStep === tutorialSteps.length - 1 ? '开始' : '下一步' }}</span>
                    <ChevronRight v-if="currentStep < tutorialSteps.length - 1" class="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped src="./TutorialGuide.css"></style>
