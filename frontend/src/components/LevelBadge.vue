<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  level: number
  tier?: string
  tierName?: string
  size?: 'xs' | 'sm' | 'md' | 'lg'
  showLevel?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
  showLevel: true
})

// 段位图标映射
const tierIcons: Record<string, string> = {
  bronze: '🥉',
  silver: '🥈',
  gold: '🥇',
  platinum: '💎',
  diamond: '💠',
  master: '⭐'
}

// 段位配置
const tierConfig = computed(() => {
  const tier = props.tier || getTierFromLevel(props.level)
  return {
    icon: tierIcons[tier] || '🎮',
    name: props.tierName || getTierName(tier),
    tier: tier,
    gradient: getTierGradient(tier)
  }
})

// 根据等级获取段位
function getTierFromLevel(level: number): string {
  if (level <= 10) return 'bronze'
  if (level <= 25) return 'silver'
  if (level <= 45) return 'gold'
  if (level <= 70) return 'platinum'
  if (level <= 90) return 'diamond'
  return 'master'
}

// 获取段位名称
function getTierName(tier: string): string {
  const names: Record<string, string> = {
    bronze: '青铜',
    silver: '白银',
    gold: '黄金',
    platinum: '铂金',
    diamond: '钻石',
    master: '大师'
  }
  return names[tier] || '未知'
}

// 段位渐变色
function getTierGradient(tier: string): string {
  const gradients: Record<string, string> = {
    bronze: 'from-orange-400 to-amber-600',
    silver: 'from-slate-300 to-slate-500',
    gold: 'from-yellow-300 to-yellow-500',
    platinum: 'from-cyan-200 to-blue-400',
    diamond: 'from-blue-300 to-indigo-500',
    master: 'from-purple-400 to-pink-500'
  }
  return gradients[tier] || 'from-gray-300 to-gray-500'
}

// 尺寸配置
const sizeClasses = computed(() => {
  const sizes = {
    xs: {
      badge: 'px-1.5 py-0.5 gap-1 rounded-md text-[9px]',
      icon: 'text-[10px]',
      level: 'text-[9px]'
    },
    sm: {
      badge: 'px-2 py-1 gap-1.5 rounded-lg text-[10px]',
      icon: 'text-xs',
      level: 'text-[10px]'
    },
    md: {
      badge: 'px-2.5 py-1 gap-1.5 rounded-xl text-xs',
      icon: 'text-sm',
      level: 'text-xs'
    },
    lg: {
      badge: 'px-3 py-1.5 gap-2 rounded-xl text-sm',
      icon: 'text-base',
      level: 'text-sm'
    }
  }
  return sizes[props.size]
})
</script>

<template>
  <div
    :class="[
      'inline-flex items-center font-black',
      'bg-gradient-to-r shadow-sm',
      tierConfig.gradient,
      sizeClasses.badge,
      'transition-all hover:scale-105 cursor-default'
    ]"
    :title="`${tierConfig.name} ${level} 级`"
  >
    <span :class="['leading-none', sizeClasses.icon]">{{ tierConfig.icon }}</span>
    <span v-if="showLevel" :class="['font-mono leading-none text-white drop-shadow-sm', sizeClasses.level]">
      Lv.{{ level }}
    </span>
  </div>
</template>

<style scoped>
/* 添加文字阴影以提高可读性 */
div {
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}
</style>
