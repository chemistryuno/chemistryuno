<template>
  <div
    class="chemical-keyboard fixed inset-x-0 bg-slate-900/95 dark:bg-black/95 backdrop-blur-2xl border-t border-white/10 z-50 pb-safe shadow-[0_-8px_30px_rgb(0,0,0,0.5)] flex flex-col transition-transform duration-300"
    :style="{
      transform: isDragging ? `translateY(${dragOffset}px)` : 'translateY(0)',
      bottom: '144px'
    }"
  >
    <!-- 顶部拖拽手柄 & 功能栏 -->
    <div
      class="relative pt-2 pb-1 cursor-grab active:cursor-grabbing"
      @touchstart="handleTouchStart"
      @touchmove="handleTouchMove"
      @touchend="handleTouchEnd"
    >
      <!-- 拖拽手柄指示器 -->
      <div class="mx-auto w-12 h-1.5 bg-white/20 rounded-full mb-2 group-hover:bg-white/30 transition-colors"></div>

      <div class="px-4 flex items-center justify-between gap-3">
        <div class="flex items-center gap-2">
          <div class="w-2 h-2 rounded-full bg-blue-500 animate-pulse"></div>
          <span class="text-[10px] font-black uppercase tracking-[0.2em] text-blue-400 font-mono">Input_Matrix</span>
        </div>

        <div class="flex items-center gap-2">
          <!-- 退格按钮 -->
          <button
            @click="handleBackspace"
            class="w-10 h-10 flex items-center justify-center bg-white/5 hover:bg-amber-500/20 rounded-[14px] border border-white/10 hover:border-amber-500/30 transition-all touch-feedback group"
          >
            <ArrowLeft class="w-4 h-4 text-slate-400 group-hover:text-amber-400" />
          </button>

          <!-- 清空按钮 -->
          <button
            @click="handleClear"
            class="w-10 h-10 flex items-center justify-center bg-white/5 hover:bg-red-500/20 rounded-[14px] border border-white/10 hover:border-red-500/30 transition-all touch-feedback group"
          >
            <Delete class="w-4 h-4 text-slate-400 group-hover:text-red-400" />
          </button>

          <!-- 关闭按钮 -->
          <button
            @click="feedback.click(); $emit('close')"
            class="w-10 h-10 flex items-center justify-center bg-white/10 hover:bg-white/20 rounded-[14px] border border-white/10 transition-all touch-feedback"
          >
            <X class="w-4 h-4 text-white" />
          </button>
        </div>
      </div>
    </div>

    <!-- 可滚动键盘区 -->
    <div class="keyboard-scroll px-3 py-3 overflow-y-auto overscroll-contain flex flex-col gap-4 max-h-[calc(45vh-120px)]">
      <!-- 元素分类区块（如果需要可以加标题，这里保持紧凑） -->
      <div class="space-y-2">
        <div class="flex items-center gap-2 px-1">
          <span class="text-[9px] font-black text-slate-500 uppercase tracking-widest">Chemical Elements</span>
          <div class="h-px flex-1 bg-white/5"></div>
        </div>

        <div class="overflow-x-auto flex gap-3 pb-2 custom-scrollbar-hidden snap-x px-1">
          <div class="grid grid-rows-3 grid-flow-col gap-2">
            <button
              v-for="element in availableElements"
              :key="element"
              @click="handleInput(element)"
              :class="cn(
                'group relative h-12 w-14 xs:w-16 flex flex-col items-center justify-center rounded-xl border transition-all touch-feedback active:scale-90 snap-start',
                element.length === 2 ? 'bg-blue-600/10 border-blue-500/20' : 'bg-indigo-600/10 border-indigo-500/20'
              )"
            >
               <div class="absolute inset-x-0 bottom-0 h-0.5 bg-blue-500/20 scale-x-0 group-hover:scale-x-100 transition-transform"></div>
               <span class="text-sm font-black text-blue-100 group-hover:text-white">{{ element }}</span>
               <span class="text-[7px] font-bold text-blue-500/60 transition-colors uppercase font-serif italic">{{ getElementName(element) }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- 数字区块 -->
      <div class="space-y-2">
        <div class="flex items-center gap-2 px-1">
          <span class="text-[9px] font-black text-slate-500 uppercase tracking-widest">Scientific Symbols</span>
          <div class="h-px flex-1 bg-white/5"></div>
        </div>

        <div class="grid grid-cols-6 gap-2">
          <!-- 数字 1-9, 0 -->
          <button
            v-for="num in [1, 2, 3, 4, 5, 6, 7, 8, 9, 0]"
            :key="num"
            @click="handleInput(String(num))"
            class="h-12 flex items-center justify-center bg-emerald-600/10 hover:bg-emerald-600/20 rounded-xl border border-emerald-500/20 hover:border-emerald-500/40 transition-all touch-feedback text-emerald-100 font-mono font-black"
          >
            {{ num }}
          </button>

          <!-- 括号和中括号 -->
          <button
            v-for="sym in ['(', ')', '[', ']']"
            :key="sym"
            @click="handleInput(sym)"
            class="h-12 flex items-center justify-center bg-amber-600/10 hover:bg-amber-600/20 rounded-xl border border-amber-500/20 hover:border-amber-500/40 transition-all touch-feedback text-amber-100 font-mono font-black text-lg"
          >
            {{ sym }}
          </button>

          <!-- 确认按钮 (占最后两格) -->
          <button
            @click="handleConfirm"
            class="col-span-4 h-12 flex items-center justify-center gap-2 bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 rounded-xl shadow-lg shadow-blue-500/20 transition-all touch-feedback active:scale-95 text-white"
          >
            <Check class="w-4 h-4 stroke-[3px]" />
            <span class="text-xs font-black uppercase tracking-widest">确认注入</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Delete, ArrowLeft, Check, X } from 'lucide-vue-next'
import { cn } from '../utils/cn'
import feedback from '../utils/feedback'

interface Props {
  modelValue: string
  deckCards?: Record<string, number>
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  deckCards: () => ({})
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'confirm': [value: string]
  'close': []
}>()

// 拖拽相关状态
const isDragging = ref(false)
const dragOffset = ref(0)
let startY = 0

const handleTouchStart = (e: TouchEvent) => {
  startY = e.touches[0].clientY
  isDragging.value = true
}

const handleTouchMove = (e: TouchEvent) => {
  if (!isDragging.value) return
  const currentY = e.touches[0].clientY
  const delta = currentY - startY
  if (delta > 0) {
    dragOffset.value = delta
  }
}

const handleTouchEnd = () => {
  if (dragOffset.value > 150) {
    emit('close')
  }
  isDragging.value = false
  dragOffset.value = 0
}

const ELEMENT_NAMES: Record<string, string> = {
  'H': '氢', 'He': '氦', 'Li': '锂', 'Be': '铍', 'B': '硼', 'C': '碳', 'N': '氮', 'O': '氧', 'F': '氟', 'Ne': '氖',
  'Na': '钠', 'Mg': '镁', 'Al': '铝', 'Si': '硅', 'P': '磷', 'S': '硫', 'Cl': '氯', 'Ar': '氩', 'K': '钾', 'Ca': '钙',
  'Fe': '铁', 'Cu': '铜', 'Zn': '锌', 'Ag': '银', 'Au': '金', 'Br': '溴', 'I': '碘'
}

const getElementName = (symbol: string) => ELEMENT_NAMES[symbol] || ''

const availableElements = computed(() => {
  const elements = new Set<string>()
  if (props.deckCards && Object.keys(props.deckCards).length > 0) {
    Object.keys(props.deckCards).forEach(formula => {
      const matches = formula.match(/[A-Z][a-z]?/g)
      if (matches) matches.forEach(el => elements.add(el))
    })
  } else {
    ['H', 'O', 'C', 'N', 'S', 'Cl', 'Na', 'K', 'Ca', 'Mg', 'Fe', 'Cu', 'Zn', 'Al', 'Ag', 'Br', 'I', 'F', 'P'].forEach(el => elements.add(el))
  }
  return Array.from(elements).sort((a, b) => {
    if (a.length !== b.length) return a.length - b.length
    return a.localeCompare(b)
  })
})

const handleInput = (char: string) => {
  feedback.click()
  emit('update:modelValue', props.modelValue + char)
}

const handleBackspace = () => {
  if (props.modelValue.length > 0) {
    feedback.click()
    const lastTwo = props.modelValue.slice(-2)
    if (lastTwo.length === 2 && lastTwo[0] === lastTwo[0].toUpperCase() && lastTwo[1] === lastTwo[1].toLowerCase()) {
      emit('update:modelValue', props.modelValue.slice(0, -2))
    } else {
      emit('update:modelValue', props.modelValue.slice(0, -1))
    }
  }
}

const handleClear = () => {
  feedback.click()
  emit('update:modelValue', '')
}

const handleConfirm = () => {
  if (props.modelValue.trim()) {
    feedback.success()
    emit('confirm', props.modelValue)
  }
}
</script>

<style scoped>
.chemical-keyboard {
  animation: slideUp 0.4s cubic-bezier(0.16, 1, 0.3, 1);
  max-height: calc(45vh - 60px);
  will-change: transform;
}

.keyboard-scroll {
  flex: 1;
  /* 自定义滚动条 */
  scrollbar-width: thin;
  scrollbar-color: rgba(255, 255, 255, 0.1) transparent;
}

.keyboard-scroll::-webkit-scrollbar {
  width: 3px;
}

.keyboard-scroll::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 10px;
}

@keyframes slideUp {
  from { transform: translateY(100%); }
  to { transform: translateY(0); }
}

.touch-feedback {
  transition: transform 0.1s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.touch-feedback:active {
  transform: scale(0.9);
}

/* 安全区域适配 */
.pb-safe {
  padding-bottom: max(env(safe-area-inset-bottom), 12px);
}
</style>
