<script setup lang="ts">
import { computed } from 'vue'
import { AVATAR_PRESETS, isPresetAvatar } from '../utils/avatarPresets'

const props = defineProps<{ avatar?: string | null }>()

const isImage  = computed(() => (props.avatar?.length ?? 0) > 50)
const isPreset = computed(() => isPresetAvatar(props.avatar))
const presetIcon = computed(() => props.avatar ? AVATAR_PRESETS[props.avatar] : null)
</script>

<template>
  <img v-if="isImage" :src="avatar!" class="w-full h-full object-cover" />
  <component v-else-if="isPreset" :is="presetIcon" class="w-[60%] h-[60%]" />
  <span v-else class="scale-110">{{ avatar || '🧪' }}</span>
</template>
