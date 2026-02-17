<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { api } from '@/utils/api'
import LevelBadge from './LevelBadge.vue'

interface LevelInfo {
  level: number
  xp: number
  total_xp: number
  tier: string
  tier_name: string
  next_level_xp: number
  progress_percent: number
}

const levelInfo = ref<LevelInfo | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)

// 获取等级信息
async function fetchLevelInfo() {
  try {
    loading.value = true
    const response = await api.get('/level/info')
    levelInfo.value = response.data
    error.value = null
  } catch (err: any) {
    console.error('获取等级信息失败:', err)
    error.value = err.response?.data?.error || '获取等级信息失败'
  } finally {
    loading.value = false
  }
}

// 格式化数字（添加千分位）
function formatNumber(num: number): string {
  return num.toLocaleString()
}

// 计算需要的XP
const requiredXP = computed(() => {
  if (!levelInfo.value) return 0
  return levelInfo.value.next_level_xp - levelInfo.value.xp
})

onMounted(() => {
  fetchLevelInfo()
})

// 暴露刷新方法给父组件
defineExpose({
  refresh: fetchLevelInfo
})
</script>

<template>
  <div class="w-full bg-white dark:bg-[#0d0d10] border border-slate-200 dark:border-white/10 rounded-2xl p-4 sm:p-6 shadow-lg">
    <!-- 加载状态 -->
    <div v-if="loading" class="flex items-center justify-center py-8">
      <div class="w-8 h-8 border-4 border-blue-500/30 border-t-blue-500 rounded-full animate-spin"></div>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="text-center py-4">
      <p class="text-red-500 text-sm mb-2">{{ error }}</p>
      <button
        @click="fetchLevelInfo"
        class="px-4 py-2 bg-blue-500 text-white rounded-lg text-sm hover:bg-blue-600 transition-colors"
      >
        重试
      </button>
    </div>

    <!-- 等级信息 -->
    <div v-else-if="levelInfo" class="space-y-4">
      <!-- 头部：段位徽章和等级 -->
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <LevelBadge :level="levelInfo.level" :tier="levelInfo.tier" :tier-name="levelInfo.tier_name" size="lg" />
          <div class="flex flex-col">
            <span class="text-lg font-black text-slate-900 dark:text-white">
              {{ levelInfo.tier_name }} {{ levelInfo.level }} 级
            </span>
            <span class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">
              研究员等级
            </span>
          </div>
        </div>
        <div class="text-right">
          <div class="text-xs font-bold text-slate-500 dark:text-slate-400 mb-0.5">当前经验</div>
          <div class="text-lg font-black text-blue-500 font-mono">
            {{ levelInfo.xp }} / {{ levelInfo.next_level_xp }}
          </div>
        </div>
      </div>

      <!-- 进度条 -->
      <div class="relative">
        <div class="h-3 bg-slate-100 dark:bg-white/5 rounded-full overflow-hidden shadow-inner">
          <div
            class="h-full bg-gradient-to-r from-blue-500 to-cyan-400 rounded-full transition-all duration-1000 ease-out relative overflow-hidden"
            :style="{ width: levelInfo.progress_percent + '%' }"
          >
            <!-- 动画光效 -->
            <div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/30 to-transparent animate-shimmer"></div>
          </div>
        </div>
        <!-- 百分比标签 -->
        <div
          class="absolute -top-6 px-2 py-0.5 bg-blue-500 text-white text-[10px] font-black rounded-md shadow-lg transition-all duration-1000"
          :style="{ left: Math.max(0, Math.min(levelInfo.progress_percent - 5, 95)) + '%' }"
        >
          {{ levelInfo.progress_percent }}%
        </div>
      </div>

      <!-- 统计信息 -->
      <div class="grid grid-cols-2 gap-3 pt-2">
        <div class="bg-slate-50 dark:bg-white/[0.02] border border-slate-200 dark:border-white/5 rounded-xl p-3">
          <div class="text-[10px] font-black uppercase tracking-wider text-slate-500 mb-1">
            总经验值
          </div>
          <div class="text-xl font-black text-slate-900 dark:text-white font-mono">
            {{ formatNumber(levelInfo.total_xp) }}
          </div>
        </div>
        <div class="bg-slate-50 dark:bg-white/[0.02] border border-slate-200 dark:border-white/5 rounded-xl p-3">
          <div class="text-[10px] font-black uppercase tracking-wider text-slate-500 mb-1">
            距离升级
          </div>
          <div class="text-xl font-black text-blue-500 font-mono">
            {{ formatNumber(requiredXP) }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@keyframes shimmer {
  0% {
    transform: translateX(-100%);
  }
  100% {
    transform: translateX(100%);
  }
}

.animate-shimmer {
  animation: shimmer 2s infinite;
}
</style>
