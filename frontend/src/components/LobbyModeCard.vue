<template>
  <button
    @click="$emit('select')"
    :class="[
      'mode-card group',
      selected ? 'mode-card--selected' : ''
    ]"
    :aria-pressed="selected"
  >
    <!-- Activity badge -->
    <div v-if="badge" class="mode-card__badge">{{ badge }}</div>

    <!-- Icon -->
    <div :class="['mode-card__icon', iconClass]">
      <component :is="icon" class="w-6 h-6" />
    </div>

    <!-- Text -->
    <div class="flex-1 text-left">
      <p class="mode-card__title">{{ title }}</p>
      <p class="mode-card__desc">{{ description }}</p>
    </div>

    <!-- Arrow indicator -->
    <ChevronRight class="w-4 h-4 text-slate-300 dark:text-white/20 group-hover:text-blue-400 transition-colors shrink-0" />
  </button>
</template>

<script setup lang="ts">
import { ChevronRight } from 'lucide-vue-next'

defineProps<{
  icon: any
  iconClass?: string
  title: string
  description: string
  badge?: string
  selected?: boolean
}>()

defineEmits<{ select: [] }>()
</script>
