<template>
  <div class="overflow-x-auto rounded border">
    <div class="min-w-[800px]">
      <!-- Header: month labels -->
      <div class="flex bg-gray-50 border-b text-xs text-gray-500">
        <div class="w-48 shrink-0 px-3 py-2 font-medium">名称</div>
        <div class="flex-1 relative h-8">
          <div
            v-for="m in monthLabels"
            :key="m.label"
            class="absolute top-0 h-full flex items-center border-l text-xs px-1"
            :style="{ left: m.left + 'px' }"
          >
            {{ m.label }}
          </div>
        </div>
      </div>

      <!-- Version rows -->
      <template v-for="ver in versions" :key="'v' + ver.id">
        <!-- Version header row -->
        <div class="flex bg-blue-50 border-b text-sm font-medium">
          <div class="w-48 shrink-0 px-3 py-1.5 text-blue-800">📅 {{ ver.name }}</div>
          <div class="flex-1 relative h-7">
            <div
              class="absolute top-1 h-5 bg-blue-200 rounded border border-blue-300"
              :style="barStyle(ver.start_date, ver.end_date)"
            ></div>
          </div>
        </div>

        <!-- Activity rows under this version -->
        <div
          v-for="act in activitiesForVersion(ver.id)"
          :key="act.id"
          class="flex border-b hover:bg-gray-50 cursor-pointer text-sm"
          @click="$emit('edit', act)"
        >
          <div class="w-48 shrink-0 px-3 py-1.5 truncate text-gray-700 flex items-center gap-1">
            <span :class="typeColor(act.type)" class="w-2 h-2 rounded-full inline-block shrink-0"></span>
            {{ act.name }}
          </div>
          <div class="flex-1 relative h-7">
            <div
              class="absolute top-1 h-5 rounded border text-xs text-white flex items-center px-1 truncate overflow-hidden transition-opacity"
              :class="[typeBarColor(act.type), !act.is_active ? 'opacity-40' : '', isRunning(act) ? 'ring-2 ring-yellow-400' : '']"
              :style="barStyle(act.start_time, act.end_time)"
            >
              {{ act.name }}
            </div>
          </div>
        </div>
      </template>

      <!-- Activities without a version -->
      <div
        v-for="act in activitiesWithoutVersion"
        :key="act.id"
        class="flex border-b hover:bg-gray-50 cursor-pointer text-sm"
        @click="$emit('edit', act)"
      >
        <div class="w-48 shrink-0 px-3 py-1.5 truncate text-gray-700 flex items-center gap-1">
          <span :class="typeColor(act.type)" class="w-2 h-2 rounded-full inline-block shrink-0"></span>
          {{ act.name }}
        </div>
        <div class="flex-1 relative h-7">
          <div
            class="absolute top-1 h-5 rounded border text-xs text-white flex items-center px-1 truncate"
            :class="[typeBarColor(act.type), !act.is_active ? 'opacity-40' : '']"
            :style="barStyle(act.start_time, act.end_time)"
          >
            {{ act.name }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  versions: any[]
  activities: any[]
}>()
defineEmits(['edit'])

// Determine the date range for the chart (min start - max end across all items)
const allDates = computed(() => {
  const dates: number[] = []
  for (const v of props.versions) {
    dates.push(new Date(v.start_date).getTime(), new Date(v.end_date).getTime())
  }
  for (const a of props.activities) {
    dates.push(new Date(a.start_time).getTime(), new Date(a.end_time).getTime())
  }
  if (dates.length === 0) {
    const now = Date.now()
    return { min: now - 86400000 * 30, max: now + 86400000 * 30 }
  }
  return { min: Math.min(...dates), max: Math.max(...dates) }
})

const totalDays = computed(() => (allDates.value.max - allDates.value.min) / 86400000 || 1)
const chartWidth = 600 // px width of the bar area

function dayOffset(dateStr: string) {
  const t = new Date(dateStr).getTime()
  return ((t - allDates.value.min) / 86400000 / totalDays.value) * chartWidth
}

function barStyle(start: string, end: string) {
  const left = Math.max(0, dayOffset(start))
  const right = Math.min(chartWidth, dayOffset(end))
  const width = Math.max(4, right - left)
  return { left: left + 'px', width: width + 'px' }
}

const monthLabels = computed(() => {
  const labels: { label: string; left: number }[] = []
  const start = new Date(allDates.value.min)
  start.setDate(1)
  const end = new Date(allDates.value.max)
  let cur = new Date(start)
  while (cur <= end) {
    const label = `${cur.getMonth() + 1}月`
    const left = dayOffset(cur.toISOString())
    labels.push({ label, left })
    cur = new Date(cur.getFullYear(), cur.getMonth() + 1, 1)
  }
  return labels
})

function activitiesForVersion(versionId: number) {
  return props.activities.filter(a => a.version_id === versionId)
}

const activitiesWithoutVersion = computed(() =>
  props.activities.filter(a => !a.version_id)
)

function isRunning(act: any) {
  if (!act.is_active) return false
  const now = Date.now()
  return new Date(act.start_time).getTime() <= now && new Date(act.end_time).getTime() >= now
}

function typeColor(type: string) {
  const m: Record<string, string> = { double_points: 'bg-yellow-400', bingo: 'bg-purple-500' }
  return m[type] || 'bg-gray-400'
}

function typeBarColor(type: string) {
  const m: Record<string, string> = {
    double_points: 'bg-yellow-500 border-yellow-600',
    bingo: 'bg-purple-500 border-purple-600',
  }
  return m[type] || 'bg-gray-400 border-gray-500'
}
</script>
