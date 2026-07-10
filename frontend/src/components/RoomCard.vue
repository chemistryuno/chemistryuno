<template>
  <div class="room-card">
    <!-- Header row -->
    <div class="room-card-header">
      <div class="status-indicator">
        <div :class="['status-dot', statusDotClass]"></div>
        <span class="status-label">{{ statusText }}</span>
      </div>
      <div class="room-type-badges">
        <div v-if="isFull && room.status === 'waiting'" class="mode-badge" style="background:rgba(239,68,68,0.08);border:1px solid rgba(239,68,68,0.2);color:#ef4444;">满员</div>
        <div v-if="room.is_points_mode" class="mode-badge mode-ranked">排位</div>
      </div>
    </div>

    <!-- Body -->
    <div class="room-card-body">
      <div class="room-main-info">
        <h3 class="room-display-name">{{ room.name }}</h3>
        <span class="room-sub-id">房间号: {{ room.id }}</span>
      </div>
      <div class="room-meta-container">
        <div v-if="room.deck_config" class="meta-item hidden sm:flex">
          <span class="meta-label">牌组</span>
          <button @click.stop="$emit('view-deck', room.deck_config)" class="deck-trigger">
            <Beaker class="w-3 h-3" />
            {{ room.deck_config.name }}
          </button>
        </div>
        <div class="occupancy-section">
          <div class="occupancy-header">
            <span class="meta-label">玩家人数</span>
            <div class="occupancy-count">{{ room.players?.length || 0 }}<span class="occupancy-max">/{{ room.max_players }}</span></div>
          </div>
          <div class="progress-track">
            <div class="progress-bar-fill" :style="{ width: `${((room.players?.length || 0) / room.max_players) * 100}%`, background: isFull ? 'linear-gradient(to right, #ef4444, #dc2626)' : undefined }"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Footer actions -->
    <div class="room-card-footer">
      <template v-if="adminView">
        <button @click="$emit('terminate', room.id)" class="btn-room-action btn-terminate"><X class="w-3.5 h-3.5" />结束</button>
        <button v-if="room.status !== 'playing' && !isFull" @click="$emit('join', room.id, false)" class="btn-room-action btn-enter"><Play class="w-3.5 h-3.5 fill-current" />加入</button>
        <button @click="$emit('join', room.id, true)" class="btn-room-action btn-spectate"><Shield class="w-3.5 h-3.5" />管理员旁观</button>
      </template>
      <template v-else-if="!room.is_private">
        <button v-if="room.status !== 'playing' && !isFull" @click="$emit('join', room.id, false)" class="btn-room-action btn-enter"><Play class="w-3.5 h-3.5 fill-current" />加入</button>
        <button @click="$emit('join', room.id, true)" class="btn-room-action btn-spectate"><Shield class="w-3.5 h-3.5" />旁观</button>
      </template>
      <button v-else @click="$emit('join', room.id, room.status === 'playing' || isFull)" :class="['btn-room-action', room.status === 'playing' || isFull ? 'btn-spectate' : 'btn-enter']">
        <component :is="room.status === 'playing' || isFull ? Shield : Play" class="w-3.5 h-3.5" :class="room.status !== 'playing' && !isFull ? 'fill-current' : ''" />
        {{ room.status === 'playing' || isFull ? '旁观' : '加入' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Beaker, Play, Shield, X } from 'lucide-vue-next'

const props = defineProps<{
  room: any
  adminView?: boolean
  countdown?: number
}>()

defineEmits<{
  join: [id: string, spectator: boolean]
  terminate: [id: string]
  'view-deck': [config: any]
}>()

const isFull = computed(() => (props.room.players?.length || 0) >= props.room.max_players)

const statusDotClass = computed(() => {
  if (props.room.status === 'waiting') return (props.countdown ?? 0) > 0 ? 'starting' : 'waiting'
  if (props.room.status === 'playing') return 'playing'
  return ''
})

const statusText = computed(() => {
  if (props.room.status === 'waiting') return (props.countdown ?? 0) > 0 ? `${props.countdown}秒` : '就绪'
  if (props.room.status === 'playing') return '进行中'
  return '已关闭'
})
</script>
