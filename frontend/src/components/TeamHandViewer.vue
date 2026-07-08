<template>
  <div class="p-3 space-y-2">
    <p class="text-xs text-gray-400 font-medium">队友手牌</p>
    <div v-if="hand?.length" class="flex flex-wrap gap-1.5">
      <div
        v-for="card in hand"
        :key="card.substance_id"
        class="px-2 py-1 bg-indigo-100 dark:bg-indigo-900/40 rounded-lg text-xs font-mono border border-indigo-200 dark:border-indigo-700"
      >
        <span class="font-bold">{{ card.formula }}</span>
        <span class="text-gray-400 ml-1 text-[10px]">{{ card.name }}</span>
      </div>
    </div>
    <p v-else class="text-xs text-gray-400">暂无手牌信息</p>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { teamAPI } from '../utils/api'

const props = defineProps<{ targetUID: number }>()
const hand = ref<any[]>([])

watch(() => props.targetUID, async (uid) => {
  if (!uid) { hand.value = []; return }
  try {
    const res = await teamAPI.getTeammateHand(uid)
    hand.value = res.data?.hand || []
  } catch { hand.value = [] }
}, { immediate: true })
</script>
