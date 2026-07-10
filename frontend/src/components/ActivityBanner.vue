<template>
  <div v-if="visibleActivities.length > 0" class="activity-banner-row space-y-2">
    <div
      v-for="act in visibleActivities"
      :key="act.id"
      :class="['activity-banner-item', isExhausted(act) ? 'activity-banner-item--exhausted' : '']"
    >
      <!-- Icon -->
      <div :class="['activity-banner-item__icon', typeIconClass(act.type)]">
        <Zap v-if="act.type === 'double_points'" class="w-4 h-4" />
        <Grid3x3 v-else-if="act.type === 'bingo'" class="w-4 h-4" />
        <Star v-else class="w-4 h-4" />
      </div>

      <!-- Info -->
      <div class="flex-1 min-w-0">
        <p class="activity-banner-item__name">
          {{ act.name }}
          <span v-if="act.type === 'double_points' && !isExhausted(act)" class="activity-banner-item__quota">
            · 今日剩余 {{ doublePointsRemaining }} 次
          </span>
          <span v-if="act.type === 'double_points' && isExhausted(act)" class="activity-banner-item__exhausted-label">
            · 今日已用完
          </span>
        </p>
        <p class="activity-banner-item__countdown">{{ countdowns[act.id] }}</p>
      </div>

      <!-- CTA -->
      <button
        v-if="!isExhausted(act)"
        @click="$emit('go', act)"
        class="activity-banner-item__btn"
      >
        前往
        <ArrowRight class="w-3 h-3" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Zap, Grid3x3, Star, ArrowRight } from 'lucide-vue-next'

const props = defineProps<{
  activities: any[]
  doublePointsRemaining?: number
}>()

defineEmits<{ go: [act: any] }>()

const now = ref(Date.now())
let timer: ReturnType<typeof setInterval>

onMounted(() => {
  timer = setInterval(() => { now.value = Date.now() }, 1000)
})
onUnmounted(() => clearInterval(timer))

const visibleActivities = computed(() =>
  (props.activities || []).filter(a => {
    if (!a.is_active) return false
    const start = new Date(a.start_time).getTime()
    const end = new Date(a.end_time).getTime()
    return now.value >= start && now.value <= end
  }).slice(0, 2)
)

const countdowns = computed(() => {
  const result: Record<number, string> = {}
  for (const act of visibleActivities.value) {
    const end = new Date(act.end_time).getTime()
    const diff = Math.max(0, Math.floor((end - now.value) / 1000))
    const h = Math.floor(diff / 3600)
    const m = Math.floor((diff % 3600) / 60)
    const s = diff % 60
    result[act.id] = h > 0
      ? `${h}小时${m}分后结束`
      : m > 0
        ? `${m}分${s}秒后结束`
        : `${s}秒后结束`
  }
  return result
})

function isExhausted(act: any) {
  return act.type === 'double_points' && (props.doublePointsRemaining ?? 1) <= 0
}

function typeIconClass(type: string) {
  if (type === 'double_points') return 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20'
  if (type === 'bingo') return 'bg-purple-500/10 text-purple-500 border-purple-500/20'
  return 'bg-blue-500/10 text-blue-500 border-blue-500/20'
}
</script>
