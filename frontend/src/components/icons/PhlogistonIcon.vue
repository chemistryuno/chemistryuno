<script setup lang="ts">
import { computed } from "vue"

const props = withDefaults(defineProps<{
  size?: number | string
  className?: string
  color?: string
}>(), {
  size: 24,
  className: "",
  color: "#f59e0b"
})

const iconSize = computed(() => typeof props.size === "number" ? `${props.size}px` : props.size)
</script>

<template>
  <svg 
    xmlns="http://www.w3.org/2000/svg" 
    :width="iconSize" 
    :height="iconSize" 
    viewBox="0 0 24 24" 
    fill="none" 
    class="phlogiston-icon"
    :style="{ color: props.color }"
  >
    <g class="animate-spin-slow-custom">
       <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5" stroke-dasharray="2 4" opacity="0.4" />
       <circle cx="22" cy="12" r="1.5" fill="currentColor" />
    </g>

    <circle cx="12" cy="12" r="6" :fill="color" opacity="0.15" class="animate-pulse-subtle-custom" />
    
    <path 
      d="M12 3C12 3 15 7.5 15 11C15 14 13.5 16 12 17C10.5 16 9 14 9 11C9 7.5 12 3 12 3Z" 
      :fill="color" 
      class="animate-float-custom"
    />
    
    <path 
      d="M7 18C7 18 8 21 12 21C16 21 17 18 17 18" 
      stroke="currentColor" 
      stroke-width="2" 
      stroke-linecap="round" 
    />
  </svg>
</template>

<style scoped>
.phlogiston-icon {
  overflow: visible;
}
.animate-spin-slow-custom {
  transform-origin: center;
  animation: spin-custom 10s linear infinite;
}
.animate-pulse-subtle-custom {
  transform-origin: center;
  animation: pulse-custom 3s ease-in-out infinite;
}
.animate-float-custom {
  transform-origin: bottom;
  animation: float-custom 2.5s ease-in-out infinite;
}
@keyframes spin-custom {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
@keyframes pulse-custom {
  0%, 100% { transform: scale(0.9); opacity: 0.1; }
  50% { transform: scale(1.1); opacity: 0.25; }
}
@keyframes float-custom {
  0%, 100% { transform: translateY(0) scaleY(1); }
  50% { transform: translateY(-1.5px) scaleY(1.05); }
}
</style>