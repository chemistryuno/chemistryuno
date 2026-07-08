<template>
  <div class="inline-block">
    <!-- Grid -->
    <div :style="{ display: 'grid', gridTemplateColumns: `repeat(${size}, 1fr)`, gap: '4px' }">
      <div
        v-for="cell in flatCells"
        :key="`${cell.row}-${cell.col}`"
        @click="$emit('cell-click', cell.row, cell.col)"
        :class="[
          'w-16 h-16 rounded-lg border-2 flex flex-col items-center justify-center text-center cursor-pointer transition-all select-none text-xs p-1 font-mono',
          cellBorderClass(cell),
          cellBgClass(cell),
          isSwapSource(cell) ? 'ring-4 ring-orange-400 scale-105' : '',
          isSelected(cell) ? 'ring-4 ring-purple-400 scale-105' : '',
        ]"
      >
        <span class="font-bold text-[10px] leading-tight">{{ cell.formula }}</span>
        <span class="text-[8px] text-gray-400 leading-tight truncate w-full text-center">{{ cell.name }}</span>
      </div>
    </div>
    <!-- Legend -->
    <div class="flex gap-3 mt-3 text-xs">
      <div class="flex items-center gap-1"><div class="w-3 h-3 rounded bg-gray-100 border"></div> 未占领</div>
      <div class="flex items-center gap-1"><div class="w-3 h-3 rounded bg-blue-200 border-blue-400"></div> 己方</div>
      <div class="flex items-center gap-1"><div class="w-3 h-3 rounded bg-red-200 border-red-400"></div> 对方</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  cells: any[][]
  myTeamId: number
  selectedCell?: { row: number; col: number } | null
  swapSource?: { row: number; col: number } | null
  mode?: string | null
}>()
defineEmits(['cell-click'])

const size = computed(() => props.cells?.length || 5)

const flatCells = computed(() => {
  const result: any[] = []
  for (const row of props.cells || [])
    for (const cell of row)
      result.push(cell)
  return result
})

function cellBgClass(cell: any) {
  if (!cell.owner_team_id) return 'bg-gray-50 dark:bg-gray-800 hover:bg-gray-100'
  // owner_team_id: 1 = team A, 2 = team B; myTeamId: 0 = A, 1 = B
  return cell.owner_team_id === props.myTeamId + 1
    ? 'bg-blue-100 dark:bg-blue-900/40'
    : 'bg-red-100 dark:bg-red-900/40'
}

function cellBorderClass(cell: any) {
  if (!cell.owner_team_id) return 'border-gray-200 dark:border-gray-700'
  return cell.owner_team_id === props.myTeamId + 1
    ? 'border-blue-400'
    : 'border-red-400'
}

function isSwapSource(cell: any) {
  return props.swapSource?.row === cell.row && props.swapSource?.col === cell.col
}

function isSelected(cell: any) {
  return props.selectedCell?.row === cell.row && props.selectedCell?.col === cell.col
}
</script>
