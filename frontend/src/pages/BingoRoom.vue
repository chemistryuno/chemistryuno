<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950 p-4">
    <div class="max-w-6xl mx-auto">
      <!-- Header -->
      <div class="flex items-center justify-between mb-4">
        <div>
          <h1 class="text-xl font-black">BINGO 对战</h1>
          <p class="text-xs text-gray-400">房间 #{{ roomId }}</p>
        </div>
        <div class="flex items-center gap-3">
          <div v-if="room?.status === 'playing'" class="text-sm font-mono">
            ⏱ <span :class="timeLeft < 60 ? 'text-red-500' : ''">{{ formatTime(timeLeft) }}</span>
          </div>
          <div class="px-3 py-1 rounded-full text-xs font-bold"
            :class="statusClass">{{ statusLabel }}</div>
        </div>
      </div>

      <div class="flex gap-4">
        <!-- Main board area -->
        <div class="flex-1">
          <!-- Vote refresh -->
          <div v-if="room?.status === 'waiting'" class="mb-4 p-4 bg-blue-50 dark:bg-blue-900/20 rounded-xl border border-blue-200 dark:border-blue-800">
            <p v-if="hasAI" class="text-sm font-medium mb-2">🤖 人机对战：可刷新棋盘后开始（仅可刷新一次）</p>
            <p v-else class="text-sm font-medium mb-2">双方可投票刷新棋盘（仅可刷新一次）</p>
            <div class="flex gap-2">
              <button v-if="room.vote_a === null || room.vote_b === null"
                @click="voteRefresh(true)"
                class="px-4 py-2 bg-green-600 text-white rounded-lg text-sm font-medium">
                {{ hasAI ? '刷新棋盘' : '同意刷新' }}
              </button>
              <button v-if="!hasAI && (room.vote_a === null || room.vote_b === null)"
                @click="voteRefresh(false)"
                class="px-4 py-2 bg-gray-400 text-white rounded-lg text-sm font-medium">
                拒绝刷新
              </button>
              <button v-if="canStart"
                @click="startGame"
                class="ml-auto px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium">
                开始游戏
              </button>
            </div>
          </div>

          <!-- Turn indicator -->
          <div v-if="room?.status === 'playing'" class="mb-3 text-sm font-medium text-center py-2 rounded-lg"
            :class="isMyTurn ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'">
            {{ isMyTurn ? '✅ 轮到你行动' : (hasAI ? '🤖 AI 思考中...' : '⏳ 等待对方...') }}
          </div>

          <!-- Board -->
          <div v-if="room?.board" class="inline-block">
            <BingoBoard
              :cells="room.board.cells"
              :my-team-id="myTeamIdx"
              :selected-cell="selectedCell"
              :swap-source="swapSource"
              :mode="currentMode"
              @cell-click="handleCellClick"
            />
          </div>

          <!-- Swap mode controls -->
          <div v-if="room?.status === 'playing' && isMyTurn" class="mt-4 flex gap-2 flex-wrap">
            <button
              @click="setMode('swap')"
              :class="['px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                currentMode === 'swap' ? 'bg-orange-500 text-white' : 'bg-gray-200 dark:bg-gray-700 hover:bg-orange-100']"
            >
              🔄 格子交换{{ swapSource ? ` (已选 ${swapSource.row},${swapSource.col})` : '' }}
            </button>
            <button
              @click="setMode('occupy')"
              :class="['px-4 py-2 rounded-lg text-sm font-medium transition-colors',
                currentMode === 'occupy' ? 'bg-purple-500 text-white' : 'bg-gray-200 dark:bg-gray-700 hover:bg-purple-100']"
            >
              📌 占领格子
            </button>
          </div>

          <!-- Card selection for occupy mode -->
          <div v-if="currentMode === 'occupy' && selectedCell && myHand.length" class="mt-3 p-3 bg-white dark:bg-gray-900 rounded-xl border">
            <p class="text-xs text-gray-400 mb-2">选择打出的手牌（目标格子: {{ room?.board?.cells?.[selectedCell.row]?.[selectedCell.col]?.formula }}）</p>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="card in myHand"
                :key="card.substance_id"
                @click="occupyWithCard(card)"
                class="px-3 py-1.5 bg-indigo-100 dark:bg-indigo-900/40 rounded-lg text-xs font-mono border border-indigo-200 hover:bg-indigo-200 transition-colors"
              >
                {{ card.formula }}
              </button>
            </div>
          </div>
        </div>

        <!-- Sidebar: Team chat & hand viewer -->
        <div class="w-72 space-y-3">
          <TeamHandViewer v-if="teammateUID" :target-u-i-d="teammateUID" />
          <div class="rounded-xl border overflow-hidden h-64">
            <TeamChat />
          </div>
        </div>
      </div>

      <!-- Result modal -->
      <div v-if="room?.status === 'finished'" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
        <div class="bg-white dark:bg-gray-900 rounded-2xl p-6 max-w-sm w-full text-center shadow-xl">
          <div class="text-4xl mb-3">{{ isWinner ? '🏆' : '😔' }}</div>
          <h2 class="text-xl font-black mb-1">{{ isWinner ? '胜利！' : '失败' }}</h2>
          <p class="text-sm text-gray-500 mb-4">
            {{ room.winner_team_idx === 0 ? 'A 队' : 'B 队' }}获胜
          </p>
          <!-- Board stats -->
          <div class="flex gap-4 justify-center text-sm mb-4">
            <div>
              <p class="font-bold text-blue-600">{{ teamACells }}</p>
              <p class="text-gray-400 text-xs">A队格子</p>
            </div>
            <div>
              <p class="font-bold text-red-600">{{ teamBCells }}</p>
              <p class="text-gray-400 text-xs">B队格子</p>
            </div>
          </div>
          <button @click="$router.push('/')" class="w-full py-2 bg-blue-600 text-white rounded-xl font-bold">
            返回大厅
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { bingoAPI } from '../utils/api'
import TeamChat from '../components/TeamChat.vue'
import TeamHandViewer from '../components/TeamHandViewer.vue'
import BingoBoard from '../components/BingoBoard.vue'

const route = useRoute()
const router = useRouter()
const roomId = computed(() => Number(route.params.id))

const room = ref<any>(null)
const myHand = ref<any[]>([])
const currentMode = ref<'swap' | 'occupy' | null>(null)
const swapSource = ref<{ row: number; col: number } | null>(null)
const selectedCell = ref<{ row: number; col: number } | null>(null)
const timeLeft = ref(600) // seconds
let pollInterval: ReturnType<typeof setInterval>
let countdownInterval: ReturnType<typeof setInterval>

const myUID = computed(() => {
  try { return Number(JSON.parse(localStorage.getItem('user') || '{}').uid) } catch { return 0 }
})

const AI_UID_BASE = 1000000
function isAIUid(uid: number) { return uid >= AI_UID_BASE }

const aiMembers = computed<number[]>(() => room.value?.ai_members || [])
const hasAI = computed(() => aiMembers.value.length > 0)

// myTeamIdx: 0 = team A, 1 = team B, -1 = not a participant
const myTeamIdx = computed<number>(() => {
  if (!room.value) return -1
  const a: number[] = room.value.team_a_members || []
  const b: number[] = room.value.team_b_members || []
  if (a.includes(myUID.value)) return 0
  if (b.includes(myUID.value)) return 1
  return -1
})

const isMyTurn = computed(() => room.value?.current_turn === myTeamIdx.value)

// Who may press "开始游戏": in PvE any human participant; in PvP the team-A side.
const canStart = computed(() => {
  if (room.value?.status !== 'waiting' || myTeamIdx.value < 0) return false
  return hasAI.value ? true : myTeamIdx.value === 0
})

const statusLabel = computed(() => {
  const s = room.value?.status || ''
  return s === 'waiting' ? '等待中' : s === 'playing' ? '进行中' : '已结束'
})

const statusClass = computed(() => {
  const s = room.value?.status || ''
  return s === 'playing' ? 'bg-green-100 text-green-700' : s === 'finished' ? 'bg-gray-100 text-gray-500' : 'bg-blue-100 text-blue-600'
})

// winner_team_idx: 0 = A, 1 = B
const isWinner = computed(() => room.value?.winner_team_idx === myTeamIdx.value)

// Cells owned by each team: owner_team_id 1 = A, 2 = B
const teamACells = computed(() => {
  if (!room.value?.board?.cells) return 0
  let c = 0
  for (const row of room.value.board.cells)
    for (const cell of row)
      if (cell.owner_team_id === 1) c++
  return c
})

const teamBCells = computed(() => {
  if (!room.value?.board?.cells) return 0
  let c = 0
  for (const row of room.value.board.cells)
    for (const cell of row)
      if (cell.owner_team_id === 2) c++
  return c
})

// Pick the first human teammate UID for the hand viewer (AI teammates have no viewable hand).
const teammateUID = computed(() => {
  if (!room.value || myTeamIdx.value < 0) return 0
  const members: number[] = myTeamIdx.value === 0
    ? (room.value.team_a_members || [])
    : (room.value.team_b_members || [])
  return members.find((uid: number) => uid !== myUID.value && !isAIUid(uid)) || 0
})

onMounted(async () => {
  await loadRoom()
  pollInterval = setInterval(loadRoom, 3000)
})

onUnmounted(() => {
  clearInterval(pollInterval)
  clearInterval(countdownInterval)
})

async function loadRoom() {
  try {
    const res = await bingoAPI.getRoom(roomId.value)
    room.value = res.data
    // Start countdown when game begins.
    if (res.data?.status === 'playing' && !countdownInterval) {
      timeLeft.value = (res.data.timeout_minutes || 10) * 60
      countdownInterval = setInterval(() => {
        timeLeft.value = Math.max(0, timeLeft.value - 1)
      }, 1000)
    }
  } catch { /* ignore */ }
}

function formatTime(s: number) {
  const m = Math.floor(s / 60)
  const sec = s % 60
  return `${m}:${String(sec).padStart(2, '0')}`
}

async function voteRefresh(agree: boolean) {
  await bingoAPI.voteRefresh(roomId.value, agree)
  await loadRoom()
}

async function startGame() {
  await bingoAPI.startGame(roomId.value)
  await loadRoom()
}

function setMode(mode: 'swap' | 'occupy') {
  currentMode.value = currentMode.value === mode ? null : mode
  swapSource.value = null
  selectedCell.value = null
}

async function handleCellClick(row: number, col: number) {
  if (!isMyTurn.value || room.value?.status !== 'playing') return

  if (currentMode.value === 'swap') {
    if (!swapSource.value) {
      swapSource.value = { row, col }
    } else {
      await bingoAPI.swapCells(roomId.value, swapSource.value.row, swapSource.value.col, row, col)
      swapSource.value = null
      currentMode.value = null
      await loadRoom()
    }
  } else if (currentMode.value === 'occupy') {
    selectedCell.value = { row, col }
  }
}

async function occupyWithCard(card: any) {
  if (!selectedCell.value) return
  const res = await bingoAPI.occupyCell(roomId.value, selectedCell.value.row, selectedCell.value.col, card.substance_id)
  selectedCell.value = null
  currentMode.value = null
  await loadRoom()
  if (res.data?.win) {
    // Game will show the result modal via room.status === 'finished'
  }
}
</script>
