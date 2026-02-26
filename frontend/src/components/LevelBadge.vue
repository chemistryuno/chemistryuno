<script setup lang="ts">
import { computed } from 'vue'
import { 
  Award, 
  FlaskConical, 
  Star, 
  Atom, 
  Trophy, 
  Crown 
} from 'lucide-vue-next'

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

// 段位配置映射
const tierMap: Record<string, { icon: any, gradient: string, color: string }> = {
  bronze: { 
    icon: Award, 
    gradient: 'from-orange-400 to-orange-700',
    color: 'text-orange-100'
  },
  silver: { 
    icon: FlaskConical, 
    gradient: 'from-slate-300 to-slate-500',
    color: 'text-slate-100'
  },
  gold: { 
    icon: Star, 
    gradient: 'from-yellow-300 to-amber-600',
    color: 'text-yellow-50'
  },
  platinum: { 
    icon: Atom, 
    gradient: 'from-cyan-300 to-blue-600',
    color: 'text-cyan-50'
  },
  diamond: { 
    icon: Trophy, 
    gradient: 'from-blue-400 to-indigo-700',
    color: 'text-blue-50'
  },
  master: { 
    icon: Crown, 
    gradient: 'from-purple-500 to-pink-700',
    color: 'text-purple-50'
  }
}

// 段位配置
const tierConfig = computed(() => {
  const tier = props.tier || getTierFromLevel(props.level)
  const config = tierMap[tier] || tierMap.bronze
  return {
    ...config,
    name: props.tierName || getTierName(tier),
    tier: tier
  }
})

// 根据等级获取段位
function getTierFromLevel(level: number): string {
  if (level === undefined || level === null || isNaN(level)) return 'bronze'
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

// 尺寸配置
const sizeClasses = computed(() => {
  const sizes = {
    xs: {
      badge: 'px-1.5 py-0.5 gap-1 rounded-md text-[9px]',
      icon: 'w-2.5 h-2.5',
      level: 'text-[9px]'
    },
    sm: {
      badge: 'px-2 py-0.5 gap-1.5 rounded-lg text-[10px]',
      icon: 'w-3 h-3',
      level: 'text-[10px]'
    },
    md: {
      badge: 'px-2.5 py-1 gap-1.5 rounded-xl text-xs',
      icon: 'w-3.5 h-3.5',
      level: 'text-xs'
    },
    lg: {
      badge: 'px-3 py-1.5 gap-2 rounded-xl text-sm',
      icon: 'w-4 h-4',
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
      'bg-gradient-to-br shadow-lg border border-white/20',
      tierConfig.gradient,
      tierConfig.color,
      sizeClasses.badge,
      'transition-all hover:scale-105 active:scale-95 cursor-default group'
    ]"
    :title="`${tierConfig.name} ${level} 级`"
  >
    <component :is="tierConfig.icon" :class="[sizeClasses.icon, 'filter drop-shadow-sm group-hover:rotate-12 transition-transform']" />
    <span v-if="showLevel" :class="['font-mono leading-none drop-shadow-md', sizeClasses.level]">
      LV.{{ level }}
    </span>
  </div>
</template>

<style scoped>
/* 增强立体感 */
div {
  box-shadow: 0 4px 12px -2px rgba(0, 0, 0, 0.2), inset 0 1px 1px rgba(255, 255, 255, 0.3);
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
}
</style>
