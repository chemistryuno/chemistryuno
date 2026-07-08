<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { pageClassNames } from '@lib'
import PhlogistonIcon from '../components/icons/PhlogistonIcon.vue'
import { gameAPI, adminAPI, friendAPI, authAPI, commonAPI, substanceAPI } from '../utils/api'
import { useDialog, setToastRef } from '../utils/dialog'
import websocket from '../utils/websocket'
import feedback from '../utils/feedback'
import { ArrowLeft, Play, RefreshCw, Zap, Activity, FlaskConical, Trophy, ChevronRight, Loader2, Users, Timer, Plus, QrCode, Copy, Sparkles, ShieldAlert, Ban, UserMinus, X, MessageCircle, UserPlus, Flag, Send, Binary } from 'lucide-vue-next'
import { cn } from '../utils/cn'
import ChatBox from '../components/ChatBox.vue'
import LevelUpAnimation from '../components/LevelUpAnimation.vue'
import GameToast from '../components/GameToast.vue'
import ChemicalKeyboard from '../components/ChemicalKeyboard.vue'
import FeedbackSettings from '../components/FeedbackSettings.vue'
import PingDisplay from '../components/PingDisplay.vue'
import UserAvatar from '../components/UserAvatar.vue'
import ConnectionOverlay from '../components/ConnectionOverlay.vue'
import { getTutorialStep, TUTORIAL_TOTAL_STEPS } from '../utils/tutorialScript'
import '../styles/mobile-game.css'

const route = useRoute()
const router = useRouter()
const { showAlert, showConfirm, showPrompt, showToast } = useDialog()
const gameToastRef = ref()
const id = route.params.id as string
const replayHistoryQueryID = computed(() => Number(route.query.replay_history_id || 0))
const replayStartIndexQuery = computed(() => {
  const raw = Number(route.query.replay_start_index || 0)
  if (!Number.isFinite(raw) || raw < 0) {
    return 0
  }
  return Math.floor(raw)
})
const replayAnchorQuery = computed(() => ({
  eventIndex: Number(route.query.event_index || 0),
  eventID: String(route.query.event_id || '').trim(),
  timestampMs: Number(route.query.timestamp_ms || 0),
  uid: Number(route.query.uid || 0),
}))
const isReplayBridgeMode = computed(() => Number.isFinite(replayHistoryQueryID.value) && replayHistoryQueryID.value > 0)
const isReplayRoomPath = computed(() => String(id) === 'replay')
const replayScopeAdmin = computed(() => String(route.query.scope || '') === 'admin')
const replayReturnPath = computed(() => {
  const raw = String(route.query.from || '').trim()
  if (/^\/admin(?:\/[a-zA-Z0-9_-]+)?$/.test(raw)) {
    return raw
  }
  if (/^\/profile(?:\/[a-zA-Z0-9_-]+)?$/.test(raw)) {
    return raw
  }
  return replayScopeAdmin.value ? '/admin/users' : '/profile/history'
})

let pageScrollLockSnapshot: {
  htmlOverflow: string
  bodyOverflow: string
  bodyWidth: string
  appOverflow: string
} | null = null

const lockPageScroll = () => {
  if (pageScrollLockSnapshot) return

  const appRoot = document.getElementById('app')
  pageScrollLockSnapshot = {
    htmlOverflow: document.documentElement.style.overflow,
    bodyOverflow: document.body.style.overflow,
    bodyWidth: document.body.style.width,
    appOverflow: appRoot?.style.overflow || ''
  }

  document.documentElement.style.overflow = 'hidden'
  document.body.style.overflow = 'hidden'
  document.body.style.width = '100%'
  if (appRoot) {
    appRoot.style.overflow = 'hidden'
  }
}

const unlockPageScroll = () => {
  if (!pageScrollLockSnapshot) return

  const appRoot = document.getElementById('app')
  document.documentElement.style.overflow = pageScrollLockSnapshot.htmlOverflow
  document.body.style.overflow = pageScrollLockSnapshot.bodyOverflow
  document.body.style.width = pageScrollLockSnapshot.bodyWidth
  if (appRoot) {
    appRoot.style.overflow = pageScrollLockSnapshot.appOverflow
  }

  pageScrollLockSnapshot = null
}

const buildReplayTimelineURL = () => {
  if (!(Number.isFinite(replayHistoryQueryID.value) && replayHistoryQueryID.value > 0)) {
    return replayReturnPath.value
  }
  const query = new URLSearchParams()
  if (replayScopeAdmin.value) {
    query.set('scope', 'admin')
  }
  query.set('from', replayReturnPath.value)
  const queryStr = query.toString()
  return queryStr
    ? `/replay/${replayHistoryQueryID.value}?${queryStr}`
    : `/replay/${replayHistoryQueryID.value}`
}

const navigateBackFromReplay = () => {
  router.push(buildReplayTimelineURL())
}

const user = ref<any>({})
try {
  const userData = JSON.parse(localStorage.getItem('user') || '{}')
  // 兼容旧版本的 id 字段
  if (userData.id && !userData.uid) {
    userData.uid = userData.id
  }
  user.value = userData
} catch (e) {
  console.error('Failed to parse user in GameRoom:', e)
}

const gameState = ref<any>(null)
const roomInfo = ref<any>(null)
const playersInfo = ref<any[]>([])
const friendsList = ref<any[]>([])
const availableSubstances = ref<string[]>([])

// 教学模式检测
const isTutorialMode = ref(false)
const tutorialHintText = ref('')
const tutorialCurrentStep = ref(1) // 当前脚本步骤
const tutorialScriptMode = ref(false) // 是否启用脚本化教学

const loading = ref(true)
const loadError = ref<string | null>(null)
const isRedirecting = ref(false)
const timeRemaining = ref(0)
const timePercent = ref(100)
let timerRaf: any = null
let lastTimeUpdate = 0
// 回合总时长（毫秒），在每次新回合开始时按剩余时间推算，避免硬编码 30s
let turnDurationMs = 30000
// 本回合"轮到我"的起始时刻，用于上报真实思考耗时（反作弊 think_time 指标）
let myTurnStartMs = 0
const selectedCard = ref<any>(null)
const selectedSubstance = ref<string | null>(null)
const turnReadySubstances = ref<string[]>([])
const doubleMode = ref(false)
const firstDoubleSubstance = ref<string | null>(null)
const secondDoubleSubstance = ref<string | null>(null)
const substanceInput = ref('')
const randomHints = ref<any[]>([])
const reactionHints = ref<any[]>([])

// 回放模拟播放状态
const replayEvents = ref<any[]>([])
const replayPlaybackIndex = ref(0)
const replayIsPlaying = ref(false)
const replaySpeed = ref(1)
const replayGameOver = ref(false)
const replayEndType = ref('')
const replayPerspectiveUID = ref<number | null>(null)
const replayInitialHands = ref<Record<number, any[]>>({})
const replaySpeedOptions = [0.5, 1, 1.5, 2, 3, 4]
let replayTimer: number | null = null
const replayActionEventTypes = new Set(['play_card', 'double_play', 'draw_card', 'timeout_auto_draw'])
let forceAutoDrawEnabled = false
let forceAutoDrawIntervalMs = 260
let forceAutoDrawSilent = true
let forceAutoDrawInFlight = false
let forceAutoDrawSuccessCount = 0
let forceAutoDrawTimer: number | null = null

// UI State
const isMobile = ref(false)
const showHints = ref(!isReplayBridgeMode.value)
const showDeckDetailModal = ref(false)
const showChemicalKeyboard = ref(false)
const handContainer = ref<HTMLElement | null>(null)
const substancesContainer = ref<HTMLElement | null>(null)
const playersContainer = ref<HTMLElement | null>(null)
const levelUpAnimationRef = ref<InstanceType<typeof LevelUpAnimation> | null>(null)

// 移动端自动全屏
const requestFullscreen = () => {
  if (!isMobile.value) return
  const el = document.documentElement as any
  const rfs = el.requestFullscreen || el.webkitRequestFullscreen || el.mozRequestFullScreen || el.msRequestFullscreen
  if (rfs) {
    rfs.call(el).catch(() => {})
  }
}

// 移动端退出全屏
const exitFullscreen = () => {
  if (!isMobile.value) return
  const doc = document as any
  if (doc.fullscreenElement || doc.webkitFullscreenElement || doc.mozFullScreenElement || doc.msFullscreenElement) {
    const efs = doc.exitFullscreen || doc.webkitExitFullscreen || doc.mozCancelFullScreen || doc.msExitFullscreen
    if (efs) {
      efs.call(doc).catch(() => {})
    }
  }
}

// 输入框焦点管理：移动端打开化学键盘，桌面端退出全屏
const handleInputFocus = () => {
  if (isMobile.value || user.value.enable_element_input) {
    // 移动端或启用了元素输入法：打开化学键盘
    showChemicalKeyboard.value = true
  }
  
  if (!isMobile.value) {
    // 桌面端：退出全屏（除非在教学模式）
    if (!isTutorialMode.value) {
      exitFullscreen()
    }
  }
}
const handleInputBlur = () => {
  // 桌面端：恢复全屏（除非在教学模式）
  if (!isMobile.value && !isTutorialMode.value) {
    requestFullscreen()
  }
}

// 自动滚动到当前行动玩家
const scrollToActivePlayer = () => {
  if (!playersContainer.value || gameState.value?.current_player == null) return
  const container = playersContainer.value
  const playerCards = container.querySelectorAll('[data-player-card]')
  const activeIndex = gameState.value.current_player
  const activeCard = playerCards[activeIndex] as HTMLElement
  if (!activeCard) return
  const containerRect = container.getBoundingClientRect()
  const cardRect = activeCard.getBoundingClientRect()
  const scrollLeft = activeCard.offsetLeft - containerRect.width / 2 + cardRect.width / 2
  container.scrollTo({ left: scrollLeft, behavior: 'smooth' })
}

const fetchRandomHints = async () => {
  try {
    const res = await commonAPI.getHints(id)
    randomHints.value = res.data || []
  } catch (error) {
    console.error('Failed to fetch hints from labs:', error)
  }
}

const fetchReactionHints = async () => {
  try {
    const res = await gameAPI.getReactionHints(id)
    reactionHints.value = res.data || []
  } catch (error) {
    console.error('Failed to fetch reaction hints:', error)
    reactionHints.value = []
  }
}

const viewCurrentDeckConfig = () => {
  if (roomInfo.value?.deck_config) {
    showDeckDetailModal.value = true
  }
}

const isReady = computed(() => {
  return roomInfo.value?.ready_uids?.includes(Number(user.value.uid))
})

const handleToggleReady = async () => {
  if (!roomInfo.value || !user.value.uid) {
    if (!roomInfo.value) {
      showToast('房间信息未加载，请刷新页面', '错误', 'error')
    } else if (!user.value.uid) {
      showToast('用户信息异常，请重新登录', '错误', 'error')
    }
    return
  }

  // 乐观更新：立即切换状态
  const uidNum = Number(user.value.uid)
  const isCurrentlyReady = roomInfo.value.ready_uids.includes(uidNum)

  if (isCurrentlyReady) {
    roomInfo.value.ready_uids = roomInfo.value.ready_uids.filter((id: number) => id !== uidNum)
  } else {
    roomInfo.value.ready_uids = [...roomInfo.value.ready_uids, uidNum]
  }

  try {
    await gameAPI.ready(id)
    // 状态也会通过 WebSocket 更新，但手动标记一下提高体验
    await loadGameState(true)
    feedback.success()
  } catch (error: any) {
    console.error('Ready API error:', error)
    // 恢复状态
    if (isCurrentlyReady) {
      if (!roomInfo.value.ready_uids.includes(uidNum)) {
        roomInfo.value.ready_uids.push(uidNum)
      }
    } else {
      roomInfo.value.ready_uids = roomInfo.value.ready_uids.filter((id: number) => id !== uidNum)
    }
    showToast(error.response?.data?.error || '操作失败', '错误', 'error')
    feedback.error()
  }
}

const isFriend = (uid: number) => {
  return friendsList.value?.some(f => Number(f.uid) === Number(uid)) ?? false
}

// 获取玩家显示名称（优先显示备注）
const getPlayerDisplayName = (player: any) => {
  if (!player) return '研究员'

  // 如果是 AI 玩家，显示其分配到的科学家姓名
  if (player.uid < 0 || player.is_ai) {
    return player.nickname || 'AI'
  }

  // 查找好友备注
  const friend = friendsList.value?.find(f => Number(f.uid) === Number(player.uid))
  if (friend?.remark) {
    return friend.remark
  }

  // 否则返回昵称或用户名
  return player.nickname || player.username || '研究员'
}

// 检查是否应该用蓝色显示（有备注的好友）
const shouldShowInBlue = (player: any) => {
  if (!player) return false
  const friend = friendsList.value?.find(f => Number(f.uid) === Number(player.uid))
  return !!(friend?.remark)
}

const handleAddFriend = async (player: any) => {
  try {
    const displayName = player.nickname || player.username
    await friendAPI.sendRequest(player.uid)
  showToast(`已向研究员 ${displayName} 发送同步请求，等待量子握手。`, '请求已发送', 'success')
  } catch (error: any) {
    showToast(error.response?.data?.error || '请求发送失败', '链路故障', 'error')
  }
}

// Chat system
const showPlayers = ref(false)
const showChat = ref(false)
const hasNewMessage = ref(false)
const showQrModal = ref(false)
const showInviteFriendsModal = ref(false)

// PvE Toasts
const pveToasts = ref<{ id: number, text: string }[]>([])
let toastIdCounter = 0

// 动画状态管理
const drawAnimatingUIDs = ref<Set<number>>(new Set())
const playerCardCounts = ref<Record<number, number>>({})

// 计算属性：只监听players的关键变化（uid + card_count）
const playersCardState = computed(() => {
  return gameState.value?.players?.map((p: any) => `${p.uid}:${p.card_count}`).join(',') || ''
})

// 改为监听简化版本而不是deep watch
watch(playersCardState, (_newState) => {
  if (!gameState.value?.players) return
  gameState.value.players.forEach((p: any) => {
    const oldVal = playerCardCounts.value[p.uid]
    // 只有当牌数增加且不是初始发牌（处于游戏中）时触发
    if (gameState.value?.status === 'playing' && typeof oldVal !== 'undefined' && p.card_count > oldVal) {
      drawAnimatingUIDs.value.add(p.uid)
      setTimeout(() => {
        drawAnimatingUIDs.value.delete(p.uid)
      }, 1000)
    }
    playerCardCounts.value[p.uid] = p.card_count
  })
})

// 监听对手掉线状态的变化，掉线时提示（避免快节奏对局中被忽略）
const playerOfflineState = ref<Record<number, boolean>>({})
const playersOfflineSignature = computed(() => {
  return gameState.value?.players?.map((p: any) => `${p.uid}:${p.is_offline ? 1 : 0}`).join(',') || ''
})
watch(playersOfflineSignature, () => {
  if (!gameState.value?.players || gameState.value.status !== 'playing') return
  const myUid = Number(user.value?.uid)
  gameState.value.players.forEach((p: any) => {
    const uid = Number(p.uid)
    const wasOffline = playerOfflineState.value[uid]
    const nowOffline = !!p.is_offline
    // 仅对“他人从在线变为掉线”提示，且需已有先前状态记录避免首帧误报
    if (typeof wasOffline !== 'undefined' && !wasOffline && nowOffline && uid !== myUid) {
      const name = p.nickname || p.username || `研究员 ${uid}`
      showToast(`${name} 已掉线，其回合可能被自动跳过`, '对手掉线', 'warning')
    }
    playerOfflineState.value[uid] = nowOffline
  })
})

const addPvEToast = (text: string) => {
  const id = ++toastIdCounter
  pveToasts.value.push({ id, text })
  // 保持最多3个显示
  if (pveToasts.value.length > 3) {
    pveToasts.value.shift()
  }
  setTimeout(() => {
    pveToasts.value = pveToasts.value.filter(t => t.id !== id)
  }, 4000)
}

// startPrivateChat 已被弃用，实验室内禁止私聊


const sendGameInvite = async (friend: any) => {
  const inviteData = {
    type: 'game_invite',
    room_id: id,
    room_name: roomInfo.value?.name || '实验室',
    player_count: allPlayers.value.length,
    max_players: roomInfo.value?.max_players || 0,
    is_points_mode: roomInfo.value?.is_points_mode || false,
    is_private: roomInfo.value?.is_private || false,
    access_key: roomInfo.value?.access_key || ''
  }

  websocket.send({
    type: 'private_chat',
    target_uid: friend.uid,
    message: JSON.stringify(inviteData),
    is_game_invite: true
  })

  feedback.success()
  showToast(`游戏邀请已发送给 ${friend.remark || friend.nickname || friend.username}`, '邀请已发送', 'success')
  showInviteFriendsModal.value = false
}

// Admin management state
const showAdminModal = ref(false)
const adminTargetUser = ref<any>(null)
const adminActionType = ref<'kick' | 'ban'>('kick')
const banUntil = ref('')
const banReason = ref('你由于违规游戏而被踢出')
const selectedBanPreset = ref<number | null>(24)

const formatDatetimeLocal = (d: Date) => {
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0') + 'T' + String(d.getHours()).padStart(2, '0') + ':' + String(d.getMinutes()).padStart(2, '0')
}

const getDefaultBanUntil = () => {
  return formatDatetimeLocal(new Date(Date.now() + 24 * 60 * 60 * 1000))
}

const banPresets = [
  { label: '1小时', hours: 1 },
  { label: '6小时', hours: 6 },
  { label: '24小时', hours: 24 },
  { label: '3天', hours: 72 },
  { label: '7天', hours: 168 },
  { label: '30天', hours: 720 },
  { label: '永久', hours: 87600 },
]

const setBanDuration = (hours: number) => {
  selectedBanPreset.value = hours
  banUntil.value = formatDatetimeLocal(new Date(Date.now() + hours * 3600 * 1000))
}

watch(showChat, (val) => {
  if (val) {
    hasNewMessage.value = false
    if (isMobile.value) {
      showHints.value = false
      showPlayers.value = false
    }
  }
})

watch(showPlayers, (val) => {
  if (val && isMobile.value) {
    showHints.value = false
    showChat.value = false
  }
})

watch(showHints, (val) => {
  if (val) {
    if (isMobile.value) {
      showChat.value = false
      showPlayers.value = false
    }
    if (randomHints.value.length === 0) {
      fetchRandomHints()
    }
    // 如果是玩家回合且提示为空，尝试获取提示（延迟检查，避免 computed 未初始化）
    nextTick(() => {
      if (isMyTurn.value && turnReadySubstances.value.length === 0) {
        fetchTurnSubstances()
      }
    })
  }
}, { immediate: true })

// 移动端自动关闭提示面板
watch(isMobile, (val) => {
  if (val) {
    showHints.value = false
  }
})

// 教学模式：监听选择物质和双元素模式变化
watch([selectedSubstance, doubleMode], () => {
  if (isTutorialMode.value && isMyTurn.value) {
    generateTutorialHint()
  }
})

// 教学模式：监听游戏状态变化，游戏开始时提示
watch(() => gameState.value?.status, (newStatus, oldStatus) => {
  if (isTutorialMode.value && newStatus === 'playing' && oldStatus === 'waiting') {
    setTimeout(() => {
      showToast(
        '游戏已开始！注意查看底部的橙色提示卡片，它会在你的回合时告诉你该做什么。',
        '🎮 开始游戏',
        'info',
        6000
      )
    }, 1000)
  }
  if (roomInfo.value?.is_points_mode && newStatus === 'playing' && oldStatus !== 'playing') {
    setTimeout(() => {
      showToast(
        '关闭提示面板即可在结算时获得额外积分加成。',
        '💡 无提示加成',
        'info',
        5000
      )
    }, 1500)
  }
})

const openAdminAction = (player: any) => {
  if (user.value.role !== 'admin' || player.uid === user.value.uid) return
  adminTargetUser.value = player
  adminActionType.value = 'kick'
  banReason.value = '你由于违规游戏而被踢出'
  selectedBanPreset.value = 24
  banUntil.value = getDefaultBanUntil()
  showAdminModal.value = true
}

const handleReportPlayer = async (player: any) => {
  const displayName = player.nickname || player.username
  const reason = await showPrompt(`举报研究员 ${displayName} (UID: ${player.uid})`, '请输入举报原因', '违规行为举报')
  if (!reason) return
  
  try {
    const replayAnchor = currentReplayEvidenceAnchor(Number(player.uid))
    await authAPI.submitFeedback(`举报用户: ${displayName} (UID: ${player.uid})\n原因: ${reason}`, 'report', replayAnchor ? {
      room_id: replayAnchor.room_id,
      reported_uid: Number(player.uid),
      replay_anchor: replayAnchor,
    } : undefined)
    showToast('举报已提交，系统正在量子分析中。', '已收到报告', 'success')
  } catch (err: any) {
    showToast(err.response?.data?.error || '无法建立举报链路', '网络干扰', 'error')
  }
}

const handleAdminAction = async () => {
  if (!adminTargetUser.value) return
  try {
    if (adminActionType.value === 'kick') {
      await adminAPI.kickPlayer(adminTargetUser.value.uid, banReason.value)
      feedback.success()
      showToast('该玩家已被强制下线并清除登录状态', '成功', 'success')
    } else {
      if (!banUntil.value) {
        showToast('请选择封禁截止时间', '参数缺失', 'warning')
        feedback.error()
        return
      }
      const until = new Date(banUntil.value)
      if (until <= new Date()) {
        showToast('封禁截止时间必须晚于当前时间', '时间无效', 'warning')
        feedback.error()
        return
      }
      await adminAPI.banUser(adminTargetUser.value.uid, until.toISOString(), banReason.value)
      feedback.success()
      showToast('该玩家已被封禁', '成功', 'success')
    }
    showAdminModal.value = false
  } catch (e: any) {
    feedback.error()
    showToast(e.response?.data?.error || '操作失败', '错误', 'error')
  }
}

const allPlayers = computed(() => {
  if (gameState.value?.players) {
    return gameState.value.players.map((p: any) => {
      const baseInfo = playersInfo.value.find(b => Number(b.uid) === Number(p.uid))
      return {
        ...p,
        avatar: p.avatar || baseInfo?.avatar || '🧪',
        username: p.username || baseInfo?.username,
        nickname: p.nickname || baseInfo?.nickname,
        is_ready: roomInfo.value?.ready_uids?.includes(Number(p.uid)),
        is_offline: baseInfo?.is_offline
      }
    })
  }
  return playersInfo.value.map(p => ({
    ...p,
    avatar: p.avatar || '🧪',
    username: p.username,
    nickname: p.nickname,
    is_ready: roomInfo.value?.ready_uids?.includes(Number(p.uid)),
    is_offline: p.is_offline
  }))
})

const currentPlayerObj = computed(() => {
  if (!gameState.value) return null
  return gameState.value.players?.[gameState.value.current_player]
})
const isSpectator = computed(() => {
  if (!user.value?.uid) return false
  
  // 优先检查根级别的 spectators 字段（用于房间等待状态）
  if (roomInfo.value?.spectators) {
    const inRoomSpectators = (roomInfo.value.spectators as number[]).includes(Number(user.value.uid))
    if (inRoomSpectators) return true
  }
  
  // 游戏运行时检查游戏状态中的观战者信息
  if (gameState.value) {
    // 已完成比赛的玩家
    const isFinished = gameState.value.finished_players?.includes(user.value.uid)
    // 直接加入的观战者
    const inSpectatorsList = gameState.value.spectators?.includes(user.value.uid)
    // 不在选手列表中的玩家也是观战者
    const isNotPlayer = !gameState.value.players?.some((p: any) => Number(p.uid) === Number(user.value?.uid))
    
    if (isFinished || inSpectatorsList || isNotPlayer) return true
  }
  
  return false
})

const isMyTurn = computed(() => {
  if (isReplayBridgeMode.value) return false
  if (!currentPlayerObj.value || !user.value || isSpectator.value) return false
  return Number(currentPlayerObj.value.uid) === Number(user.value.uid)
})
const hasTurnLimit = computed(() => {
  if (tutorialScriptMode.value) return false
  return !!(gameState.value?.turn_end_time && gameState.value.turn_end_time > 0)
})
const tutorialStepDisplay = computed(() => Math.min(tutorialCurrentStep.value, TUTORIAL_TOTAL_STEPS))
const tutorialProgressPercent = computed(() => Math.round((tutorialStepDisplay.value / TUTORIAL_TOTAL_STEPS) * 100))
const replayPerspectivePlayer = computed(() => {
  if (!isReplayBridgeMode.value || replayPerspectiveUID.value == null || !gameState.value?.players) {
    return null
  }
  return (gameState.value.players || []).find((p: any) => Number(p.uid) === Number(replayPerspectiveUID.value)) || null
})
const replayPerspectiveName = computed(() => {
  const player = replayPerspectivePlayer.value
  if (player) {
    return getPlayerDisplayName(player)
  }
  return '系统视角'
})
const myData = computed(() => {
  if (!gameState.value) return null
  if (isReplayBridgeMode.value) {
    if (replayPerspectiveUID.value == null) return null
    return (gameState.value.players || []).find((p: any) => Number(p.uid) === Number(replayPerspectiveUID.value)) || null
  }
  if (!user.value) return null
  return (gameState.value.players || []).find((p: any) => Number(p.uid) === Number(user.value.uid))
})
const myIndex = computed(() => {
  if (!gameState.value) return -1
  if (isReplayBridgeMode.value) {
    if (replayPerspectiveUID.value == null) return -1
    return (gameState.value.players || []).findIndex((p: any) => Number(p.uid) === Number(replayPerspectiveUID.value))
  }
  if (!user.value) return -1
  return (gameState.value.players || []).findIndex((p: any) => Number(p.uid) === Number(user.value.uid))
})
const allowedAny = computed(() => {
  if (!gameState.value) return false
  return typeof gameState.value?.allowed_any_player !== 'undefined' && gameState.value?.allowed_any_player === myIndex.value
})
const winner = computed(() => {
  if (!gameState.value) return null
  const finishers = gameState.value.finished_players || []
  const winnerUid = finishers.length > 0 ? finishers[0] : null
  if (winnerUid !== null) {
     return gameState.value.players.find((p: any) => p.uid === winnerUid)
  }
  return gameState.value.players?.find((p: any) => p.card_count === 0)
})

const showSettlementPanel = computed(() => {
  if (isReplayBridgeMode.value) return false
  const status = String(gameState.value?.status || '')
  return status === 'finished' || status === 'terminated'
})

const isTerminatedSettlement = computed(() => {
  return String(gameState.value?.status || '') === 'terminated'
})

const isManualSettlement = ref(false)

const sortedPointsChanges = computed(() => {
  if (!gameState.value?.points_changes) return []
  return Object.entries(gameState.value.points_changes)
    .map(([uid, points]) => ({
      uid: Number(uid),
      points: Number(points),
      xp: gameState.value.xp_changes?.[uid] || 0,
      player: gameState.value.players.find((p: any) => String(p.uid) === String(uid))
    }))
    .sort((a, b) => {
      // 优先按照 finished_players 中的顺序（排名）排序
      const rankA = gameState.value.finished_players?.indexOf(a.uid) ?? 999
      const rankB = gameState.value.finished_players?.indexOf(b.uid) ?? 999
      
      if (rankA !== rankB) {
        return rankA - rankB
      }
      
      // 如果都没在 finished_players 中（比如中途退出的），按积分降序
      return b.points - a.points
    })
})

const fastReactionUIDs = computed(() => {
  return gameState.value?.fast_reaction_users || {}
})

const hasCheatDetection = computed(() => {
  return Object.keys(fastReactionUIDs.value).length > 0
})

const ELEMENTS_DATA: Record<string, { name: string, class: string }> = {
  'H': { name: '氢', class: 'element-H' },
  'O': { name: '氧', class: 'element-O' },
  'C': { name: '碳', class: 'element-C' },
  'N': { name: '氮', class: 'element-N' },
  'S': { name: '硫', class: 'element-S' },
  'F': { name: '氟', class: 'element-F' },
  'P': { name: '磷', class: 'element-P' },
  'Cl': { name: '氯', class: 'element-Cl' },
  'Br': { name: '溴', class: 'element-Br' },
  'I': { name: '碘', class: 'element-I' },
  'Na': { name: '钠', class: 'element-Na' },
  'K': { name: '钾', class: 'element-K' },
  'Mg': { name: '镁', class: 'element-Mg' },
  'Ca': { name: '钙', class: 'element-Ca' },
  'Ba': { name: '钡', class: 'element-Ba' },
  'Al': { name: '铝', class: 'element-Al' },
  'Fe': { name: '铁', class: 'element-Fe' },
  'Zn': { name: '锌', class: 'element-Zn' },
  'Ag': { name: '银', class: 'element-Ag' },
  'Hg': { name: '汞', class: 'element-Hg' },
  'Cu': { name: '铜', class: 'element-Cu' },
}

// 物质名称映射（从 API 加载）
const substanceNames = ref<Record<string, string>>({})

// 加载物质名称映射
const loadSubstanceNames = async () => {
  try {
    const response = await substanceAPI.getSubstanceNames()
    substanceNames.value = response.data || {}
  } catch (error) {
    console.error('[GameRoom] Failed to load substance names:', error)
    // 使用默认映射作为后备
    substanceNames.value = {
      'H2O': '水', 'H2': '氢气', 'O2': '氧气', 'HCl': '盐酸', 'H2SO4': '硫酸',
      'NaOH': '氢氧化钠', 'NaCl': '氯化钠', 'CO2': '二氧化碳', 'CaO': '氧化钙',
      'CuO': '氧化铜', 'Fe2O3': '氧化铁', 'Fe': '铁', 'Cu': '铜', 'Zn': '锌',
      'Mg': '镁', 'Al': '铝', 'C': '碳', 'S': '硫', 'Cl2': '氯气', 'AgNO3': '硝酸银'
    }
  }
}

const formatFormula = (formula: string) => {
  if (!formula) return ''
  return formula.replace(/(\d+)/g, '<sub>$1</sub>')
}

const getSubstanceName = (formula: string) => {
  if (substanceNames.value[formula]) return substanceNames.value[formula]
  // 回退到硬编码的元素数据
  if (ELEMENTS_DATA[formula]) return ELEMENTS_DATA[formula].name
  return formula
}

const getDynamicWidthClass = (formula: string, type: 'single' | 'double' = 'single') => {
  if (!formula) return type === 'single' ? 'w-40 sm:w-48' : 'w-28 sm:w-32'
  const len = formula.length
  if (type === 'single') {
    if (len <= 3) return 'w-40 sm:w-48'
    if (len <= 6) return 'w-48 sm:w-56'
    if (len <= 9) return 'w-56 sm:w-64'
    return 'w-64 sm:w-80'
  } else {
    if (len <= 3) return 'w-28 sm:w-32'
    if (len <= 6) return 'w-32 sm:w-40'
    return 'w-40 sm:w-48'
  }
}

const getFormulaFontSize = (formula: string, type: 'single' | 'double' = 'single') => {
  if (!formula) return ''
  const len = formula.length
  if (type === 'single') {
    if (len > 12) return 'text-[20px] sm:text-[28px]'
    if (len > 8) return 'text-[24px] sm:text-[32px]'
    if (len > 5) return 'text-[28px] sm:text-[38px]'
    return 'text-[32px] sm:text-[44px]'
  } else {
    if (len > 8) return 'text-[18px] sm:text-[24px]'
    if (len > 5) return 'text-[22px] sm:text-[30px]'
    return 'text-[28px] sm:text-[36px]'
  }
}

// 解析化学式，返回元素及其数量（与后端 parseSubstance 逻辑一致）
const parseSubstanceElements = (substance: string): Record<string, number> => {
  const result: Record<string, number> = {}
  const stack: Record<string, number>[] = [result]
  let i = 0
  while (i < substance.length) {
    const c = substance[i]
    if (c === '(') {
      stack.push({})
      i++
    } else if (c === ')') {
      i++
      let count = 0
      while (i < substance.length && substance[i] >= '0' && substance[i] <= '9') {
        count = count * 10 + (substance.charCodeAt(i) - 48)
        i++
      }
      if (count === 0) count = 1
      const top = stack.pop()!
      const parent = stack[stack.length - 1]
      for (const [k, v] of Object.entries(top)) {
        parent[k] = (parent[k] || 0) + v * count
      }
    } else if (c >= 'A' && c <= 'Z') {
      const start = i
      i++
      while (i < substance.length && substance[i] >= 'a' && substance[i] <= 'z') i++
      const element = substance.slice(start, i)
      let count = 0
      while (i < substance.length && substance[i] >= '0' && substance[i] <= '9') {
        count = count * 10 + (substance.charCodeAt(i) - 48)
        i++
      }
      if (count === 0) count = 1
      const current = stack[stack.length - 1]
      current[element] = (current[element] || 0) + count
    } else {
      i++
    }
  }
  return result
}

// 检查玩家手牌是否包含合成该物质所需的所有元素
const canPlayerMakeSubstance = (substance: string): boolean => {
  if (!myData.value?.hand_cards) return false
  const required = parseSubstanceElements(substance)
  // 统计手牌中各元素数量
  const handElements: Record<string, number> = {}
  for (const card of myData.value.hand_cards) {
    handElements[card.type] = (handElements[card.type] || 0) + 1
  }
  for (const [elem, count] of Object.entries(required)) {
    if ((handElements[elem] || 0) < count) return false
  }
  return true
}

// 过滤并随机取最多3个可接续反应物提示
const filteredReactionHints = computed(() => {
  if (!reactionHints.value.length || !isMyTurn.value) return []
  const eligible = reactionHints.value.filter((hint: any) => canPlayerMakeSubstance(hint.substance))
  if (eligible.length <= 3) return eligible
  // 随机打乱后取前3个
  const shuffled = [...eligible]
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]]
  }
  return shuffled.slice(0, 3)
})

const exp = ref(Number(localStorage.getItem('chem_exp') || '0'))
const achievements = ref<string[]>(JSON.parse(localStorage.getItem('chem_achievements') || '[]'))

const checkAchievements = (substance: string) => {
  if (!substance) return
  if (substance.includes('Au') && !achievements.value.includes('炼金术士')) {
    achievements.value.push('炼金术士')
    showToast('获得成就：炼金术士 (合成单质金)', '成就达成！', 'success')
  }
  localStorage.setItem('chem_achievements', JSON.stringify(achievements.value))
}

const addExp = (amount: number) => {
  exp.value += amount
  localStorage.setItem('chem_exp', exp.value.toString())
}

// --- 移植结束 ---

const startTimer = () => {
  if (timerRaf) cancelAnimationFrame(timerRaf)
  lastTimeUpdate = 0
  let hasTimeoutFired = false

  // 在回合开始时按剩余时间推算本回合总时长，作为进度条分母；
  // 兼容后台可配置的回合时长，不再硬编码 30 秒
  if (gameState.value?.turn_end_time) {
    const remaining = gameState.value.turn_end_time - Date.now()
    if (remaining > 0) {
      turnDurationMs = remaining
    }
  }

  const animate = () => {
    if (!gameState.value || !gameState.value.turn_end_time || gameState.value.status !== 'playing') {
      timeRemaining.value = 0
      timePercent.value = 0
      return
    }

    const now = Date.now()
    const diffMs = gameState.value.turn_end_time - now

    // 进度条平滑更新（每帧），分母使用本回合实际总时长
    timePercent.value = Math.max(0, Math.min(100, (diffMs / turnDurationMs) * 100))

    // 秒数显示每秒更新一次
    if (now - lastTimeUpdate >= 1000) {
      const diff = Math.max(0, Math.floor(diffMs / 1000))
      timeRemaining.value = diff
      lastTimeUpdate = now
    }

    if (diffMs > 0) {
      timerRaf = requestAnimationFrame(animate)
    } else {
      timeRemaining.value = 0
      timePercent.value = 0

      // 倒计时结束后自动摸牌（仅在我的回合且未触发过）
      if (isMyTurn.value && !hasTimeoutFired && !tutorialScriptMode.value) {
        hasTimeoutFired = true
        console.log('⏰ 回合倒计时已到期，自动摸牌')
        showToast('回合超时，已自动为你摸牌', '时间到', 'warning')
        handleDrawCard()
      }
    }
  }

  timerRaf = requestAnimationFrame(animate)
}

watch(() => gameState.value?.turn_end_time, () => {
  startTimer()
})

// 记录"轮到我"的起始时刻，用于上报真实思考耗时（反作弊 think_time 指标）
watch(isMyTurn, (mine) => {
  if (mine) {
    myTurnStartMs = Date.now()
  }
}, { immediate: true })

// 计算本次出牌的思考耗时（毫秒）；无有效起点则返回 0（后端按未上报处理）
const computeThinkMs = (): number => {
  if (myTurnStartMs <= 0) return 0
  const elapsed = Date.now() - myTurnStartMs
  return elapsed > 0 ? elapsed : 0
}

// 场上物质变化时，刷新反应提示
watch(() => gameState.value?.last_card?.substance, () => {
  if (gameState.value?.status === 'playing') {
    fetchReactionHints()
  }
}, { immediate: true })

const fetchTurnSubstances = async () => {
  if (isReplayBridgeMode.value) {
    turnReadySubstances.value = []
    return
  }
  if (!isMyTurn.value) {
    turnReadySubstances.value = []
    return
  }
  try {
    const response = await gameAPI.getAvailableSubstances(id)
    turnReadySubstances.value = response.data || []
  } catch (error) {
    console.error('获取回合可用物质失败:', error)
  }
}

watch(() => isMyTurn.value, (val) => {
  if (isReplayBridgeMode.value) {
    turnReadySubstances.value = []
    return
  }
  if (val) {
    fetchTurnSubstances()
    // 回合开始反馈
    feedback.turnStart()
  } else {
    turnReadySubstances.value = []
  }
}, { immediate: true })

// 监听化学反应 - 播放反应音效
watch(() => gameState.value?.current_reaction, (newReaction, oldReaction) => {
  if (newReaction && newReaction !== oldReaction) {
    feedback.reaction()
  }
})

// 监听游戏结束 - 播放胜利/失败音效并标记教学完成
watch(() => gameState.value?.status, (newStatus) => {
  if (newStatus === 'finished' && gameState.value?.finished_players) {
    // 教学模式：游戏结束时立即标记已完成
    if (isTutorialMode.value) {
      localStorage.setItem('chemistry-uno-tutorial-completed', 'true')
    }
    const finishedPlayers = gameState.value.finished_players
    const myUID = user.value.uid
    if (finishedPlayers.length > 0 && finishedPlayers[0] === myUID) {
      // 我是第一名 - 胜利音效
      feedback.win()
    } else if (finishedPlayers.includes(myUID)) {
      // 我完成了但不是第一名 - 成功音效
      feedback.success()
    }
  }
})

const handleGameUpdate = (message: any) => {
  // 只有在手动结算且游戏尚未结束时才跳过更新
  // 如果游戏已经结束（有points_changes），即使isManualSettlement也要更新以显示排名
  if (isManualSettlement.value && !message.data?.points_changes) return

  // 如果收到的是完整的游戏状态对象
  if (message.data && typeof message.data === 'object') {
    gameState.value = message.data
    syncTutorialStateFromGameState()
    if (isMyTurn.value) {
      fetchTurnSubstances()
      if (isTutorialMode.value) {
        generateTutorialHint()
      }
    }
  } else {
    // 如果收到的是房间ID字符串，则重新拉取完整状态
    loadGameState(true).then(() => {
      if (isMyTurn.value) {
        fetchTurnSubstances()
      }
    })
  }
}

// 处理升级事件
const handleLevelUp = (data: any) => {
  const levelData = data.data || data
  if (levelUpAnimationRef.value) {
    levelUpAnimationRef.value.show(levelData)
  }
}

const handleActionToast = (msg: any) => {
  const content = msg.data || msg.message
  if (content) {
    if (roomInfo.value?.is_pve) {
      addPvEToast(content)
    } else {
      showToast(content, '实验动态', 'info')
    }
  }
}

const handleRoomTerminated = async (msg: any) => {
  const reason = msg.message || '由于连接中断，实验室已关闭'

  const hasSettlementData = Boolean(
    (gameState.value?.finished_players && gameState.value.finished_players.length > 0) ||
    (gameState.value?.points_changes && Object.keys(gameState.value.points_changes).length > 0) ||
    (gameState.value?.xp_changes && Object.keys(gameState.value.xp_changes).length > 0)
  )

  if (hasSettlementData && !isReplayBridgeMode.value) {
    if (gameState.value?.status !== 'finished') {
      gameState.value.status = 'terminated'
    }
    showToast(reason, '实验结束', 'warning')
    return
  }

  isRedirecting.value = true
  await showAlert(reason, '实验结束')
  router.push('/')
}

const handlePlayerKicked = async (msg: any) => {
  isRedirecting.value = true
  await showAlert(msg.message || '由于消极游戏，您已被踢出', '权限移除')
  router.push('/')
}

const handleChatNotify = () => {
  if (!showChat.value) {
    hasNewMessage.value = true
  }
}

// 为 WebSocket 事件创建包装函数，确保类型匹配
const handlePlayerJoined = () => {
  loadGameState(true)
}

const handlePlayerLeft = () => {
  loadGameState(true)
}

const syncTutorialStateFromGameState = () => {
  if (!gameState.value) return

  if (typeof gameState.value.tutorial_script_mode === 'boolean') {
    tutorialScriptMode.value = gameState.value.tutorial_script_mode
    isTutorialMode.value = gameState.value.tutorial_script_mode
  }

  if (typeof gameState.value.tutorial_current_step === 'number' && gameState.value.tutorial_current_step > 0) {
    tutorialCurrentStep.value = gameState.value.tutorial_current_step
  }
}

// 教学模式智能提示
const generateTutorialHint = () => {
  if (!gameState.value || !isMyTurn.value) {
    tutorialHintText.value = ''
    return
  }

  // 脚本化教学模式：显示脚本中的提示
  if (tutorialScriptMode.value) {
    const currentStep = getTutorialStep(tutorialCurrentStep.value)
    if (currentStep && currentStep.player === 'human') {
      tutorialHintText.value = currentStep.hint
    } else {
      tutorialHintText.value = ''
    }
    return
  }

  // 原有的通用教学提示逻辑
  const myPlayer = gameState.value.players?.[gameState.value.current_player]
  if (!myPlayer) {
    tutorialHintText.value = ''
    return
  }

  const handSize = myPlayer.hand?.length || 0
  const topCard = gameState.value.discard_top

  // 根据游戏状态生成提示
  if (!topCard) {
    tutorialHintText.value = '💡 回合开始：你可以先在「化学库」中选择一个物质，然后点击「打出卡牌」开始游戏！'
  } else if (handSize === 0) {
    tutorialHintText.value = '💡 手牌用完了！点击「摸牌」按钮抽一张新牌（你将失去这个回合）'
  } else if (doubleMode.value) {
    tutorialHintText.value = '💡 双元素模式：选择第二个物质，两个物质将一起打出到战场中'
  } else if (selectedSubstance.value) {
    tutorialHintText.value = '💡 已选择物质！点击「打出卡牌」按钮将它放到战场上，或点击「双元素」同时打出两张'
  } else {
    tutorialHintText.value = '💡 轮到你了：在「化学库」中选择一个物质，让它与战场中央的卡牌发生化学反应！'
  }
}

const parseReplayTimestampMs = (event: any): number | null => {
  const unixMs = Number(event?.unix_ms || 0)
  if (Number.isFinite(unixMs) && unixMs > 0) return unixMs
  const timeStr = event?.at || event?.timestamp
  if (!timeStr) return null
  const normalized = String(timeStr).includes('T') ? String(timeStr) : String(timeStr).replace(' ', 'T')
  const ms = new Date(normalized).getTime()
  return Number.isFinite(ms) ? ms : null
}

const normalizeReplayEvents = (events: any[]) => {
  return events
    .map((event: any, index: number) => ({
      ...event,
      __index: index,
      __type: event?.type || event?.event || '',
      __eventIndex: Number(event?.event_index || 0) || index + 1,
      __eventID: String(event?.event_id || ''),
      __actorUID: Number(event?.actor_uid ?? event?.uid ?? 0),
      __payload: event?.payload || {},
      __timeMs: parseReplayTimestampMs(event)
    }))
    .sort((a: any, b: any) => {
      if (a.__timeMs == null && b.__timeMs == null) return a.__index - b.__index
      if (a.__timeMs == null) return 1
      if (b.__timeMs == null) return -1
      if (a.__timeMs !== b.__timeMs) return a.__timeMs - b.__timeMs
      return a.__index - b.__index
    })
    .map((event: any, sortedIndex: number) => ({
      ...event,
      __sortedIndex: sortedIndex,
    }))
}

const resolveReplayCardKeyFromObject = (card: any) => {
  return String(card?.card_symbol || card?.card_type || card?.type || '?')
}

const resolveReplayCardKeyForPlay = (payload: any) => {
  if (Array.isArray(payload?.cards) && payload.cards.length) {
    return resolveReplayCardKeyFromObject(payload.cards[0])
  }
  return payload?.card_symbol || payload?.card_type || '未知卡'
}

const resolveReplayCardKeysForDouble = (payload: any) => {
  if (Array.isArray(payload?.cards) && payload.cards.length) {
    return payload.cards.map((card: any) => resolveReplayCardKeyFromObject(card))
  }
  const symbol = payload?.card_symbol || payload?.card_type
  if (symbol) return [symbol, symbol]
  return [payload?.sub1 || payload?.substance_1 || '未知卡', payload?.sub2 || payload?.substance_2 || '未知卡']
}

const replayEffectCardTypes = new Set(['+2', '+4', 'Au', 'He', 'Ne', 'Ar', 'Kr'])

const makeReplayHandCard = (cardKey: string) => {
  const normalized = String(cardKey || '?')
  return {
    type: normalized,
    count: 1,
    effect: replayEffectCardTypes.has(normalized) ? normalized : ''
  }
}

const cloneReplayHandCards = (cards: any[]) => {
  if (!Array.isArray(cards)) return []
  return cards.map((card: any) => ({ ...card }))
}

const resolveReplayDrawCardKeys = (payload: any, drawCount: number) => {
  if (drawCount <= 0) return []

  if (Array.isArray(payload?.cards) && payload.cards.length) {
    const keys = payload.cards
      .map((card: any) => card?.card_symbol || card?.card_type || card?.type || '?')
      .slice(0, drawCount)
    if (keys.length === drawCount) {
      return keys
    }
    return [...keys, ...Array.from({ length: drawCount - keys.length }, () => '?')]
  }

  const key = payload?.card_symbol || payload?.card_type || '?'
  return Array.from({ length: drawCount }, () => String(key || '?'))
}

const removeReplayHandCard = (player: any, preferredKey: string) => {
  const hand = Array.isArray(player?.hand_cards) ? player.hand_cards : []
  let removeIndex = hand.findIndex((card: any) => String(card?.type || '') === String(preferredKey || ''))

  if (removeIndex < 0) {
    removeIndex = hand.findIndex((card: any) => String(card?.type || '') === '?')
  }
  if (removeIndex < 0 && hand.length > 0) {
    removeIndex = 0
  }

  if (removeIndex >= 0) {
    hand.splice(removeIndex, 1)
  }

  player.hand_cards = hand
  player.card_count = hand.length
}

const addReplayHandCards = (player: any, cardKeys: string[]) => {
  if (!Array.isArray(player?.hand_cards)) {
    player.hand_cards = []
  }
  cardKeys.forEach((key: string) => {
    player.hand_cards.push(makeReplayHandCard(key))
  })
  player.card_count = player.hand_cards.length
}

const parseReplayInitialHandsFromEvents = (events: any[]) => {
  const gameStartEvent = events.find((event: any) => event?.__type === 'game_start' && event?.__payload?.initial_hands)
  const initialHandsPayload = gameStartEvent?.__payload?.initial_hands
  if (!initialHandsPayload || typeof initialHandsPayload !== 'object') {
    return null
  }

  const result: Record<number, string[]> = {}
  Object.entries(initialHandsPayload).forEach(([uidKey, cards]) => {
    const uid = Number(uidKey)
    if (!Number.isFinite(uid)) return
    if (!Array.isArray(cards)) {
      result[uid] = []
      return
    }
    result[uid] = cards.map((card: any) => resolveReplayCardKeyFromObject(card))
  })
  return result
}

const buildReplayInitialHands = (players: any[], events: any[]) => {
  const explicitInitialHands = parseReplayInitialHandsFromEvents(events)
  if (explicitInitialHands) {
    players.forEach((player: any) => {
      const uid = Number(player.uid)
      if (!Array.isArray(explicitInitialHands[uid])) {
        explicitInitialHands[uid] = []
      }
    })
    return explicitInitialHands
  }

  const handsByUID: Record<number, string[]> = {}
  const cardBalanceByUID: Record<number, number> = {}

  players.forEach((player: any) => {
    const uid = Number(player.uid)
    handsByUID[uid] = []
    cardBalanceByUID[uid] = 0
  })

  const ensurePlayableCard = (uid: number, cardKey: string) => {
    if (cardBalanceByUID[uid] <= 0) {
      handsByUID[uid].push(cardKey || '?')
      cardBalanceByUID[uid] += 1
    }
    cardBalanceByUID[uid] -= 1
  }

  events.forEach((event: any) => {
    const uid = Number(event.__actorUID || 0)
    if (!(uid in handsByUID)) {
      return
    }

    const eventType = event.__type
    const payload = event.__payload || {}

    if (eventType === 'play_card') {
      ensurePlayableCard(uid, resolveReplayCardKeyForPlay(payload))
      return
    }

    if (eventType === 'double_play') {
      const keys = resolveReplayCardKeysForDouble(payload)
      keys.forEach((key: string) => {
        ensurePlayableCard(uid, key)
      })
      return
    }

    if (eventType === 'draw_card' || eventType === 'timeout_auto_draw') {
      const drawCount = Number(payload.actual_count ?? payload.draw_count ?? payload.requested_count ?? 1)
      if (Number.isFinite(drawCount) && drawCount > 0) {
        cardBalanceByUID[uid] += drawCount
      }
    }
  })

  return handsByUID
}

const buildReplayLastCardFromEvent = (event: any) => {
  const eventType = event?.__type || event?.type || event?.event
  const payload = event?.__payload || event?.payload || {}

  if (eventType === 'double_play') {
    const doubleKeys = resolveReplayCardKeysForDouble(payload)
    const reactants = [payload.sub1 || payload.substance_1, payload.sub2 || payload.substance_2].filter(Boolean)
    const result = payload.substance || payload.result_substance || reactants.join(' + ')
    return {
      substance: result,
      reactants,
      card: {
        type: doubleKeys[0] || reactants[0] || 'R',
        effect: payload.card_effect || ''
      }
    }
  }

  if (eventType === 'play_card') {
    const playedKey = resolveReplayCardKeyForPlay(payload)
    const substance = payload.substance || payload.result_substance || payload.card_symbol || payload.card_type || 'R'
    return {
      substance,
      card: {
        type: playedKey || substance,
        effect: payload.card_effect || ''
      }
    }
  }

  return null
}

const clearReplayTimer = () => {
  if (replayTimer != null) {
    window.clearTimeout(replayTimer)
    replayTimer = null
  }
}

const normalizeReplaySpeedValue = (speed: number) => {
  if (!Number.isFinite(speed) || speed <= 0) {
    return 1
  }
  return speed
}

const replayStatusText = computed(() => {
  if (!isReplayBridgeMode.value) return ''
  if (replayGameOver.value) return '游戏结束'
  return replayIsPlaying.value ? '播放中' : '已暂停'
})

const replayProgressText = computed(() => {
  const total = replayEvents.value.length
  const current = Math.min(replayPlaybackIndex.value, total)
  return `${current}/${total}`
})

const normalizeReplaySeekIndex = (targetIndex: number) => {
  if (!Number.isFinite(targetIndex)) {
    return 0
  }
  return Math.max(0, Math.min(Math.floor(targetIndex), replayEvents.value.length))
}

const findReplayStartIndexFromAnchor = (events: any[]) => {
  const query = replayAnchorQuery.value
  if (!(query.eventIndex > 0) && !query.eventID && !(query.timestampMs > 0)) {
    return replayStartIndexQuery.value
  }
  const matched = events.find((event: any) => {
    if (query.eventID && event.__eventID === query.eventID) return true
    if (query.eventIndex > 0 && Number(event.__eventIndex || event.__index + 1) === query.eventIndex) return true
    if (query.timestampMs > 0 && Number(event.__timeMs || 0) === query.timestampMs) return true
    return false
  })
  if (!matched) return replayStartIndexQuery.value
  const sortedIndex = Number(matched.__sortedIndex)
  return Number.isFinite(sortedIndex) && sortedIndex >= 0 ? sortedIndex : replayStartIndexQuery.value
}

const seekReplayToIndex = (targetIndex: number) => {
  if (!isReplayBridgeMode.value || !gameState.value) return

  clearReplayTimer()
  replayIsPlaying.value = false
  replayGameOver.value = false
  replayEndType.value = ''

  const clampedIndex = normalizeReplaySeekIndex(targetIndex)
  replayPlaybackIndex.value = 0
  resetReplaySimulationBoard()

  for (let index = 0; index < clampedIndex; index += 1) {
    const event = replayEvents.value[index]
    applyReplayEventToBoard(event)
    replayPlaybackIndex.value = index + 1

    const eventType = event?.__type || ''
    if (eventType === 'game_finished' || eventType === 'game_terminated_invalid') {
      replayGameOver.value = true
      replayEndType.value = eventType
      return
    }
  }

  if (clampedIndex >= replayEvents.value.length && replayEvents.value.length > 0) {
    replayGameOver.value = true
    replayEndType.value = replayEvents.value[replayEvents.value.length - 1]?.__type || 'game_finished'
  }
}

const jumpReplayByStep = (delta: number) => {
  seekReplayToIndex(replayPlaybackIndex.value + delta)
}

const handleReplaySeekInput = (event: Event) => {
  const nextIndex = Number((event.target as HTMLInputElement | null)?.value ?? 0)
  seekReplayToIndex(nextIndex)
}

const replaySummary = computed(() => {
  const cardCounts: Record<string, number> = {}
  let roundCount = 0
  let totalPlays = 0

  for (const event of replayEvents.value) {
    const eventType = event.__type
    const payload = event.__payload || {}

    if (replayActionEventTypes.has(eventType)) {
      roundCount += 1
    }

    if (eventType === 'play_card') {
      totalPlays += 1
      const key = resolveReplayCardKeyForPlay(payload)
      cardCounts[key] = (cardCounts[key] || 0) + 1
    }

    if (eventType === 'double_play') {
      totalPlays += 2
      const keys = resolveReplayCardKeysForDouble(payload)
      keys.forEach((key: string) => {
        cardCounts[key] = (cardCounts[key] || 0) + 1
      })
    }
  }

  return {
    roundCount,
    totalPlays,
    cardCounts
  }
})

const replayCardCountEntries = computed(() => {
  return Object.entries(replaySummary.value.cardCounts).sort((a, b) => b[1] - a[1])
})

const handleReplayPerspectiveSwitch = (player: any) => {
  if (!isReplayBridgeMode.value) return
  const uid = Number(player?.uid)
  if (!Number.isFinite(uid)) return
  replayPerspectiveUID.value = uid
}

const resetReplaySimulationBoard = () => {
  if (!gameState.value?.players) return

  gameState.value.status = 'playing'
  gameState.value.current_reaction = ''
  gameState.value.last_card = null
  gameState.value.discard_pile = []
  gameState.value.pending_draw_count = 0
  gameState.value.pending_forced_plays = 0
  gameState.value.finished_players = []

  gameState.value.players = gameState.value.players.map((player: any) => ({
    ...player,
    card_count: Array.isArray(replayInitialHands.value[Number(player.uid)])
      ? replayInitialHands.value[Number(player.uid)].length
      : 0,
    hand_cards: cloneReplayHandCards(replayInitialHands.value[Number(player.uid)] || [])
  }))
}

const applyReplayEventToBoard = (event: any) => {
  if (!gameState.value?.players) return

  const eventType = event.__type
  const payload = event.__payload || {}
  const actorUID = Number(event.__actorUID || 0)
  const actorIndex = gameState.value.players.findIndex((player: any) => Number(player.uid) === actorUID)

  if (actorIndex >= 0) {
    gameState.value.current_player = actorIndex
  }

  if (eventType === 'play_card') {
    gameState.value.last_card = buildReplayLastCardFromEvent(event)
    gameState.value.current_reaction = payload.substance || payload.result_substance || payload.card_symbol || payload.card_type || ''
    gameState.value.discard_pile = [...(gameState.value.discard_pile || []), { replay: true, event: eventType }]
    if (actorIndex >= 0) {
      const actor = gameState.value.players[actorIndex]
      removeReplayHandCard(actor, resolveReplayCardKeyForPlay(payload))
    }
    return
  }

  if (eventType === 'double_play') {
    gameState.value.last_card = buildReplayLastCardFromEvent(event)
    const substance = payload.substance || payload.result_substance || payload.sub1 || payload.substance_1 || ''
    gameState.value.current_reaction = substance
    gameState.value.discard_pile = [...(gameState.value.discard_pile || []), { replay: true, event: eventType }]
    if (actorIndex >= 0) {
      const actor = gameState.value.players[actorIndex]
      const keys = resolveReplayCardKeysForDouble(payload)
      keys.forEach((key: string) => {
        removeReplayHandCard(actor, key)
      })
    }
    return
  }

  if (eventType === 'draw_card' || eventType === 'timeout_auto_draw') {
    const drawCount = Number(payload.actual_count ?? payload.draw_count ?? payload.requested_count ?? 1)
    if (actorIndex >= 0 && drawCount > 0) {
      const actor = gameState.value.players[actorIndex]
      addReplayHandCards(actor, resolveReplayDrawCardKeys(payload, drawCount))
    }
    return
  }

  if (eventType === 'game_finished' || eventType === 'game_terminated_invalid') {
    const winnerUID = Number(payload.winner_uid || payload.uid || payload.winner || 0)
    if (winnerUID) {
      gameState.value.finished_players = [winnerUID]
    }
  }
}

const computeReplayDelayMs = (currentEvent: any, nextEvent: any) => {
  const currentMs = currentEvent?.__timeMs
  const nextMs = nextEvent?.__timeMs
  if (Number.isFinite(currentMs) && Number.isFinite(nextMs)) {
    const delta = Math.max(200, Math.min(5000, Number(nextMs) - Number(currentMs)))
    return Math.max(120, Math.round(delta / replaySpeed.value))
  }
  return Math.max(150, Math.round(850 / replaySpeed.value))
}

const finishReplayPlayback = (endType: string) => {
  replayIsPlaying.value = false
  replayGameOver.value = true
  replayEndType.value = endType
  clearReplayTimer()
}

const scheduleReplayPlaybackStep = () => {
  clearReplayTimer()
  if (!replayIsPlaying.value || !isReplayBridgeMode.value) return

  if (replayPlaybackIndex.value >= replayEvents.value.length) {
    finishReplayPlayback('game_finished')
    return
  }

  const currentEvent = replayEvents.value[replayPlaybackIndex.value]
  applyReplayEventToBoard(currentEvent)
  replayPlaybackIndex.value += 1

  const currentType = currentEvent?.__type || ''
  if (currentType === 'game_finished' || currentType === 'game_terminated_invalid') {
    finishReplayPlayback(currentType)
    return
  }

  if (replayPlaybackIndex.value >= replayEvents.value.length) {
    finishReplayPlayback('game_finished')
    return
  }

  const nextEvent = replayEvents.value[replayPlaybackIndex.value]
  const delay = computeReplayDelayMs(currentEvent, nextEvent)
  replayTimer = window.setTimeout(() => {
    scheduleReplayPlaybackStep()
  }, delay)
}

const startReplayPlayback = (restart = false, restartIndex = 0) => {
  if (!isReplayBridgeMode.value || !gameState.value) return

  if (restart) {
    seekReplayToIndex(restartIndex)
  }

  if (!replayEvents.value.length) {
    replayGameOver.value = true
    replayEndType.value = 'game_finished'
    replayIsPlaying.value = false
    return
  }

  if (replayGameOver.value || replayPlaybackIndex.value >= replayEvents.value.length) {
    replayIsPlaying.value = false
    return
  }

  replayIsPlaying.value = true
  clearReplayTimer()
  replayTimer = window.setTimeout(() => {
    scheduleReplayPlaybackStep()
  }, Math.max(120, Math.round(240 / normalizeReplaySpeedValue(replaySpeed.value))))
}

const toggleReplayPlayback = () => {
  if (!isReplayBridgeMode.value) return
  if (replayGameOver.value) return

  if (replayIsPlaying.value) {
    replayIsPlaying.value = false
    clearReplayTimer()
    return
  }

  startReplayPlayback(false)
}

const setReplaySpeed = (speed: number) => {
  const normalized = normalizeReplaySpeedValue(Number(speed))
  replaySpeed.value = normalized

  if (replayIsPlaying.value) {
    clearReplayTimer()
    replayTimer = window.setTimeout(() => {
      scheduleReplayPlaybackStep()
    }, Math.max(120, Math.round(220 / normalizeReplaySpeedValue(replaySpeed.value))))
  }
}

const restartReplayPlayback = (startIndex = 0) => {
  startReplayPlayback(true, startIndex)
}

const currentReplayEvidenceAnchor = (reportedUID?: number) => {
  if (!isReplayBridgeMode.value) return null
  const currentIndex = Math.max(0, Math.min(replayPlaybackIndex.value - 1, replayEvents.value.length - 1))
  const event = replayEvents.value[currentIndex]
  if (!event) {
    return {
      room_id: undefined,
      game_history_id: replayHistoryQueryID.value,
      replay_id: String(replayHistoryQueryID.value),
      event_index: replayAnchorQuery.value.eventIndex || undefined,
      event_id: replayAnchorQuery.value.eventID || undefined,
      event_timestamp_ms: replayAnchorQuery.value.timestampMs || undefined,
      player_uid: reportedUID || replayAnchorQuery.value.uid || undefined,
      evidence_precision: replayAnchorQuery.value.eventIndex || replayAnchorQuery.value.eventID || replayAnchorQuery.value.timestampMs ? 'operation' : 'room',
      action_summary: 'replay report',
    }
  }
  return {
    room_id: roomInfo.value?.id === `replay-${replayHistoryQueryID.value}` ? undefined : roomInfo.value?.id,
    game_history_id: replayHistoryQueryID.value,
    replay_id: String(replayHistoryQueryID.value),
    event_index: Number(event.__eventIndex || event.__index + 1),
    event_id: event.__eventID || undefined,
    event_type: event.__type || undefined,
    player_uid: reportedUID || Number(event.__actorUID || 0) || undefined,
    event_timestamp_ms: Number(event.__timeMs || 0) || undefined,
    action_summary: event.action_summary || event.__type || 'replay report',
    evidence_precision: 'operation',
  }
}

const loadReplaySimulationState = async () => {
  loading.value = true
  loadError.value = null
  showChat.value = false
  showHints.value = false
  showPlayers.value = false
  showChemicalKeyboard.value = false
  clearReplayTimer()
  replayEvents.value = []
  replayPlaybackIndex.value = 0
  replayIsPlaying.value = false
  replayGameOver.value = false
  replayEndType.value = ''
  replayPerspectiveUID.value = null
  replayInitialHands.value = {}

  try {
    await loadSubstanceNames()

    if (!replayHistoryQueryID.value) {
      throw new Error('无效的回放编号')
    }

    let response: any
    try {
      if (replayScopeAdmin.value && user.value?.role === 'admin') {
        response = await adminAPI.getGameReplay(replayHistoryQueryID.value)
      } else {
        response = await gameAPI.getMyGameReplay(replayHistoryQueryID.value)
      }
    } catch (firstError) {
      if (user.value?.role === 'admin') {
        response = await adminAPI.getGameReplay(replayHistoryQueryID.value)
      } else {
        throw firstError
      }
    }

    const replayPayload = response?.data || {}
    const rawReplayEvents = Array.isArray(replayPayload?.replay?.events) ? replayPayload.replay.events : []
    const normalizedReplayEvents = normalizeReplayEvents(rawReplayEvents)

    const profiles = Array.isArray(replayPayload?.player_profiles)
      ? replayPayload.player_profiles
      : (Array.isArray(replayPayload?.players)
        ? replayPayload.players.map((uid: number) => ({ uid, nickname: '', username: `UID ${uid}` }))
        : [])

    const players = profiles.map((p: any, index: number) => ({
      uid: Number(p.uid),
      username: p.username || `UID ${p.uid}`,
      nickname: p.nickname || '',
      avatar: p.avatar || '🧪',
      card_count: 0,
      hand_cards: [],
      index,
      is_ai: typeof p.is_ai === 'boolean' ? p.is_ai : Number(p.uid) < 0,
      is_offline: false,
      points: 0,
      exp: 0
    }))

    const initialHandsByUID = buildReplayInitialHands(players, normalizedReplayEvents)
    players.forEach((player: any) => {
      const initialHand = (initialHandsByUID[Number(player.uid)] || []).map((key: string) => makeReplayHandCard(key))
      player.hand_cards = initialHand
      player.card_count = initialHand.length
    })

    playersInfo.value = profiles.map((p: any) => ({
      uid: Number(p.uid),
      username: p.username || p.nickname || `UID ${p.uid}`,
      nickname: p.nickname || '',
      avatar: p.avatar || '🧪',
      is_ai: typeof p.is_ai === 'boolean' ? p.is_ai : Number(p.uid) < 0,
      is_offline: false
    }))

    replayInitialHands.value = {}
    players.forEach((player: any) => {
      replayInitialHands.value[Number(player.uid)] = cloneReplayHandCards(player.hand_cards)
    })

    const preferredPerspective =
      players.find((player: any) => Number(player.uid) === Number(user.value?.uid)) ||
      players.find((player: any) => !player.is_ai) ||
      players[0]
    replayPerspectiveUID.value = preferredPerspective ? Number(preferredPerspective.uid) : null

    roomInfo.value = {
      id: `replay-${replayHistoryQueryID.value}`,
      name: `回放模拟 #${String(replayHistoryQueryID.value).padStart(4, '0')}`,
      players: players.map((p: any) => p.uid),
      spectators: [Number(user.value?.uid)],
      ready_uids: [],
      countdown: 0,
      max_players: Math.max(players.length, 2),
      status: 'playing',
      is_points_mode: false,
      deck_config: {
        name: 'Replay Simulation',
        cards: {}
      },
      is_private: false,
      access_key: '',
      is_pve: false
    }

    gameState.value = {
      status: 'playing',
      players,
      current_player: 0,
      direction: 1,
      turn_end_time: 0,
      pending_draw_count: 0,
      pending_forced_plays: 0,
      allowed_any_player: -1,
      current_reaction: '',
      discard_pile: [],
      last_card: null,
      finished_players: [],
      spectators: [Number(user.value?.uid)],
      points_changes: {},
      xp_changes: {}
    }

    replayEvents.value = normalizedReplayEvents
    restartReplayPlayback(findReplayStartIndexFromAnchor(normalizedReplayEvents))
  } catch (error: any) {
    console.error('[GameRoom] 加载回放模拟失败:', error)
    loadError.value = error?.response?.data?.error || error?.message || '回放模拟加载失败'
  } finally {
    loading.value = false
  }
}

const loadGameState = async (silent = false) => {
  if (isRedirecting.value) {
    loading.value = false
    return
  }
  try {
    if (!silent && !roomInfo.value) {
      loading.value = true
      // 首次加载时加载物质名称映射
      await loadSubstanceNames()
    }
    // 只在首次加载时尝试加入房间
    if (!silent) {
      try {
        // 从 URL 查询参数中获取访问密钥和观战模式
        const accessKey = route.query.key as string | undefined
        const asSpectator = route.query.spectator === 'true' || route.query.spectator === '1'
        await gameAPI.joinRoom(id, accessKey, asSpectator)
      } catch (joinError: any) {
        // 如果加入失败（例如房间已满、被封禁等），显示错误并返回
        console.error('[GameRoom] Failed to join room:', joinError)
        const errorMsg = joinError.response?.data?.error || '无法加入该房间'
        loadError.value = errorMsg
        showToast(errorMsg, '加入失败', 'error')
        loading.value = false
        router.push('/')
        return
      }
    }

    const response = await gameAPI.getRoomState(id)
    const data = response.data

    roomInfo.value = {
      id: data.id,
      name: data.name,
      players: data.players,
      ready_uids: data.ready_uids || [],
      countdown: data.countdown || 0,
      max_players: data.max_players,
      status: data.status,
      is_points_mode: data.is_points_mode,
      deck_config: data.deck_config,
      is_private: data.is_private,
      access_key: data.access_key,
      is_pve: data.is_pve
    }

    playersInfo.value = data.players_info || []

    if (data.game_state) {
      gameState.value = data.game_state
      syncTutorialStateFromGameState()

      // 教学模式提示生成
      if (isTutorialMode.value && isMyTurn.value) {
        generateTutorialHint()
      }
    } else {
      // no game_state yet, room is in waiting status
    }
    
    loading.value = false
  } catch (error: any) {
    console.error('加载游戏状态失败:', error)
    loading.value = false

    if (error.response?.status === 404) {
      loadError.value = '房间不存在或已被关闭'
      isRedirecting.value = true
      showToast('房间不存在或已被关闭', '未知实验室', 'error')
      router.push('/')
    } else if (error.response?.status === 401) {
      loadError.value = '身份验证失败，请重新登录'
      isRedirecting.value = true
      showToast('身份验证失败，请重新登录', '准入失败', 'error')
      router.push('/login')
    } else if (error.response?.status === 403) {
      loadError.value = '您不在该房间中'
      isRedirecting.value = true
      showToast('您不在该房间中', '准入失败', 'error')
      router.push('/')
    } else {
      loadError.value = '实验环境加载异常，请重试'
      // 非致命错误不自动跳转，允许用户重试
      if (!silent) {
        isRedirecting.value = true
        router.push('/')
      }
    }
  }
}

// 监控弹窗状态以禁用/启用背景滚动
watch(
  [showDeckDetailModal, showInviteFriendsModal, showAdminModal, showSettlementPanel, replayGameOver],
  ([deck, invite, admin, settlement, replay]) => {
    const hasModal = deck || invite || admin || settlement || replay
    if (hasModal) {
      document.documentElement.style.overflow = 'hidden'
      document.body.style.overflow = 'hidden'
    } else {
      document.documentElement.style.overflow = ''
      document.body.style.overflow = ''
    }
  }
)

onMounted(() => {
  lockPageScroll()
  exposeForceAutoDrawConsoleAPI()

  // 检测教学模式
  const tutorialMode = localStorage.getItem('chemistry-uno-tutorial-mode')
  if (tutorialMode === 'true') {
    isTutorialMode.value = true
    tutorialScriptMode.value = true  // 启用脚本化教学
  }

  // 设置浮窗提示组件引用
  setToastRef(gameToastRef)

  // 重置状态，防止之前的错误状态影响
  isRedirecting.value = false

  if (isReplayBridgeMode.value) {
    loadReplaySimulationState()
    return
  }

  // 设置一个安全超时，如果15秒后还在loading状态，强制重置
  const safetyTimeout = setTimeout(() => {
    if (loading.value) {
      console.error('Loading timeout - forcing reset')
      loading.value = false
      loadError.value = '实验室初始化超时，请检查网络连接后重试'
      showToast('实验室初始化超时，请检查网络连接后重试', '连接超时', 'error')
      router.push('/')
    }
  }, 15000)

  // 加载好友列表，添加错误处理
  friendAPI.getFriends()
    .then(res => {
      friendsList.value = res.data || []
    })
    .catch(err => {
      console.error('Failed to load friends list:', err)
      friendsList.value = [] // 确保失败时也初始化为空数组
      // 继续加载游戏状态，即使好友列表加载失败
    })

  loadGameState()
    .then(() => {
      clearTimeout(safetyTimeout) // 成功加载后清除超时

      // 游戏状态加载成功后，获取提示信息（避免 setup 阶段的 API 调用失败导致提示为空）
      if (showHints.value && randomHints.value.length === 0) {
        fetchRandomHints()
      }

      // 确保WebSocket已连接
      if (!websocket.isConnected()) {
        websocket.connect()
      }

      websocket.joinRoom(id)
      websocket.on('game_update', handleGameUpdate)
      websocket.on('player_joined', handlePlayerJoined)
      websocket.on('player_left', handlePlayerLeft)
      websocket.on('action_toast', handleActionToast)
      websocket.on('room_terminated', handleRoomTerminated)
      websocket.on('player_kicked', handlePlayerKicked)
      websocket.on('chat', handleChatNotify)
      websocket.on('private_chat', handleChatNotify)
      websocket.on('level_up', handleLevelUp)

      // 教学模式欢迎提示
      if (isTutorialMode.value) {
        const tutorialWelcomeShown = localStorage.getItem('chemistry-uno-tutorial-welcome-shown')
        if (!tutorialWelcomeShown) {
          setTimeout(() => {
            if (tutorialScriptMode.value) {
              showToast(
                '🎓 欢迎来到脚本化教学关卡！你将跟随系统指引，按照固定步骤学习游戏机制。请严格按照提示的顺序出牌。',
                '📖 教学脚本已加载',
                'success',
                9000
              )
            } else {
              showToast(
                '💡 欢迎来到教学关卡！这是一场低难度的AI对战，在你的回合时会出现实时提示帮助你学习游戏。祝你玩得开心！',
                '🎯 教学模式已开启',
                'success',
                8000
              )
            }
            localStorage.setItem('chemistry-uno-tutorial-welcome-shown', 'true')
          }, 1500)
        }
      }
    })
    .catch(err => {
      clearTimeout(safetyTimeout) // 捕获错误后也清除超时
      // loadGameState 内部已经处理了错误，这里只是确保不会有未处理的promise rejection
      console.error('Failed to initialize game room:', err)
      loading.value = false
    })
})

onUnmounted(() => {
  unlockPageScroll()
  cleanupForceAutoDrawConsoleAPI()
  clearReplayTimer()

  // 清除教学模式标记，并记录已完成
  if (isTutorialMode.value) {
    localStorage.removeItem('chemistry-uno-tutorial-mode')
    localStorage.removeItem('chemistry-uno-tutorial-welcome-shown')
    if (gameState.value?.status === 'finished') {
      localStorage.setItem('chemistry-uno-tutorial-completed', 'true')
    }
  }

  if (timerRaf) cancelAnimationFrame(timerRaf)
  websocket.leaveRoom()
  websocket.off('game_update', handleGameUpdate)
  websocket.off('player_joined', handlePlayerJoined)
  websocket.off('player_left', handlePlayerLeft)
  websocket.off('action_toast', handleActionToast)
  websocket.off('room_terminated', handleRoomTerminated)
  websocket.off('player_kicked', handlePlayerKicked)
  websocket.off('chat', handleChatNotify)
  websocket.off('private_chat', handleChatNotify)
})

const canRunTutorialAction = (action: 'play' | 'draw' | 'double') => {
  if (!tutorialScriptMode.value) return true

  const currentStep = getTutorialStep(tutorialCurrentStep.value)
  if (!currentStep) return true

  if (currentStep.player !== 'human') {
    showToast('当前是 AI 演示步骤，请等待 AI 操作', '⚠️ 教学模式', 'warning', 2500)
    return false
  }

  if (currentStep.action !== action) {
    showToast('当前步骤不支持该操作，请按提示执行', '⚠️ 教学模式', 'warning', 2500)
    return false
  }

  return true
}

const handleCardClick = async (card: any) => {
  if (!isMyTurn.value) return
  if (!canRunTutorialAction('play')) return

  feedback.click()

  // 功能牌直接打出
  const specialTypes = ['+2', '+4', 'Au', 'He', 'Ne', 'Ar', 'Kr']
  if (specialTypes.includes(card.type) || card.effect) {
    try {
      await gameAPI.playCard(id, card, card.type, computeThinkMs())
      feedback.playCard()
      selectedCard.value = null
      selectedSubstance.value = null
      availableSubstances.value = []
      return
    } catch (error: any) {
      showToast(error.response?.data?.error || '出牌失败', '反应中断', 'error')
      feedback.error()
      return
    }
  }

  // 元素牌：直接出牌该元素符号（单质）
  // 例如：点击 H 手牌 → 出牌物质为 H（不管单质是 H 还是 H₂）
  // 后端会进行物质合法性检测（substances表）和反应检查（reactions表）
  try {
    await gameAPI.playCard(id, card, card.type, computeThinkMs())
    feedback.playCard()
    selectedCard.value = null
    selectedSubstance.value = null
    availableSubstances.value = []
    // 增加经验值
    addExp(10)
    checkAchievements(card.type)
  } catch (error: any) {
    showToast(error.response?.data?.error || '出牌失败', '反应中断', 'error')
    feedback.error()
  }
}

const handlePlayCard = async () => {
  if (!canRunTutorialAction('play')) return
  if (!selectedSubstance.value) {
    showToast('请选择要合成或放置的化学物质', '未选择目标', 'warning')
    return
  }

  // 脚本化教学模式：验证是否是正确的牌
  if (tutorialScriptMode.value) {
    const currentStep = getTutorialStep(tutorialCurrentStep.value)
    if (currentStep && currentStep.player === 'human') {
      if (selectedSubstance.value !== currentStep.substance) {
        showToast(
          `请按照教学提示打出 <strong>${currentStep.substance}</strong>`,
          '⚠️ 教学模式',
          'warning',
          3000
        )
        return
      }
    }
  }

  if (doubleMode.value) {
    if (!firstDoubleSubstance.value) {
      firstDoubleSubstance.value = selectedSubstance.value
    } else if (!secondDoubleSubstance.value) {
      secondDoubleSubstance.value = selectedSubstance.value
    }
    selectedCard.value = null
    selectedSubstance.value = null
    availableSubstances.value = []
    return
  }

  try {
    // 如果没有选中的卡片，则传递一个带类型的占位符，后端会根据物质消耗手牌
    const cardToPlay = selectedCard.value || { type: selectedSubstance.value, count: 1, effect: '' }
    await gameAPI.playCard(id, cardToPlay, selectedSubstance.value, computeThinkMs())

    // 播放打牌反馈
    feedback.playCard()

    // 增加经验值并检查成就
    addExp(10)
    checkAchievements(selectedSubstance.value)

    selectedCard.value = null
    selectedSubstance.value = null
    availableSubstances.value = []
  } catch (error: any) {
    showToast(error.response?.data?.error || '出牌失败', '反应中断', 'error')
  }
}

const handleDoublePlay = async () => {
  if (!canRunTutorialAction('double')) return
  if (!firstDoubleSubstance.value || !secondDoubleSubstance.value) {
    showToast('请选择参与双联反应的两种物质', '未就绪', 'warning')
    feedback.error()
    return
  }

  try {
    await gameAPI.playDouble(id, firstDoubleSubstance.value, secondDoubleSubstance.value)

    feedback.playCard()

    // 增加经验值
    addExp(25)
    checkAchievements(firstDoubleSubstance.value)
    checkAchievements(secondDoubleSubstance.value)

    firstDoubleSubstance.value = null
    secondDoubleSubstance.value = null
    doubleMode.value = false
    selectedCard.value = null
    selectedSubstance.value = null
    availableSubstances.value = []
  } catch (error: any) {
    showToast(error.response?.data?.error || '双联行动失败', '反应中断', 'error')
    feedback.error()
  }
}

const toggleDoubleMode = () => {
  if (!myData.value?.double_action_available) {
    showToast('双联反应尚未就绪，请先进行普通实验（行动）', '无法发动', 'warning')
    feedback.error()
    return
  }
  doubleMode.value = !doubleMode.value
  feedback.doubleMode()
  firstDoubleSubstance.value = null
  secondDoubleSubstance.value = null
  selectedSubstance.value = null
}

const removeSubstance = (pos: number) => {
  feedback.click()
  if (pos === 1) {
    firstDoubleSubstance.value = secondDoubleSubstance.value
    secondDoubleSubstance.value = null
  } else {
    secondDoubleSubstance.value = null
  }
}

const handleInputPlay = async () => {
  if (!canRunTutorialAction('play')) return
  if (!substanceInput.value) return

  if (doubleMode.value) {
    const sub = substanceInput.value
    if (!firstDoubleSubstance.value) {
      firstDoubleSubstance.value = sub
    } else if (!secondDoubleSubstance.value) {
      secondDoubleSubstance.value = sub
    }
    substanceInput.value = ''
    return
  }

  try {
    // 为兼容原API，传一个空Card对象
    await gameAPI.playCard(id, { type: '', count: 0, effect: '' }, substanceInput.value, computeThinkMs())

    feedback.playCard()

    // 增加经验值并检查成就
    addExp(10)
    checkAchievements(substanceInput.value)

    substanceInput.value = ''
    selectedCard.value = null
    selectedSubstance.value = null
    availableSubstances.value = []
  } catch (error: any) {
    showToast(error.response?.data?.error || '出牌失败', '反应中断', 'error')
    feedback.error()
  }
}

const handleDrawCard = async () => {
  if (!canRunTutorialAction('draw')) return
  try {
    await gameAPI.drawCard(id)
    feedback.drawCard()
  } catch (error: any) {
    showToast(error.response?.data?.error || '摸牌失败', '系统异常', 'error')
    feedback.error()
  }
}

const stopForceAutoDrawLoop = () => {
  if (forceAutoDrawTimer != null) {
    window.clearInterval(forceAutoDrawTimer)
    forceAutoDrawTimer = null
  }
}

const getForceAutoDrawStatus = () => {
  return {
    enabled: forceAutoDrawEnabled,
    interval_ms: forceAutoDrawIntervalMs,
    silent: forceAutoDrawSilent,
    room_id: id,
    in_flight: forceAutoDrawInFlight,
    success_count: forceAutoDrawSuccessCount,
    replay_mode: isReplayBridgeMode.value || isReplayRoomPath.value
  }
}

const tryForceAutoDraw = async () => {
  if (forceAutoDrawInFlight || !id) return
  forceAutoDrawInFlight = true
  try {
    await gameAPI.drawCard(id)
    forceAutoDrawSuccessCount += 1
  } catch (error: any) {
    if (!forceAutoDrawSilent) {
      const message = error?.response?.data?.error || error?.message || 'unknown_error'
      console.debug('[ForceAutoDraw] draw failed:', message)
    }
  } finally {
    forceAutoDrawInFlight = false
  }
}

const scheduleForceAutoDrawLoop = () => {
  stopForceAutoDrawLoop()
  if (!forceAutoDrawEnabled) return

  forceAutoDrawTimer = window.setInterval(() => {
    void tryForceAutoDraw()
  }, forceAutoDrawIntervalMs)
}

const startForceAutoDrawFromConsole = (options?: { intervalMs?: number, silent?: boolean }) => {
  if (typeof options?.intervalMs === 'number' && Number.isFinite(options.intervalMs)) {
    forceAutoDrawIntervalMs = Math.max(80, Math.floor(options.intervalMs))
  }
  if (typeof options?.silent === 'boolean') {
    forceAutoDrawSilent = options.silent
  }

  forceAutoDrawEnabled = true
  scheduleForceAutoDrawLoop()
  void tryForceAutoDraw()
  console.info('[ForceAutoDraw] started', getForceAutoDrawStatus())
  return getForceAutoDrawStatus()
}

const stopForceAutoDrawFromConsole = () => {
  forceAutoDrawEnabled = false
  stopForceAutoDrawLoop()
  console.info('[ForceAutoDraw] stopped', getForceAutoDrawStatus())
  return getForceAutoDrawStatus()
}

const exposeForceAutoDrawConsoleAPI = () => {
  const win = window as Window & { __chemForceAutoDraw?: any }
  win.__chemForceAutoDraw = {
    start: (options?: { intervalMs?: number, silent?: boolean }) => startForceAutoDrawFromConsole(options),
    stop: () => stopForceAutoDrawFromConsole(),
    once: async () => {
      await tryForceAutoDraw()
      return getForceAutoDrawStatus()
    },
    status: () => getForceAutoDrawStatus()
  }

  console.info('[ForceAutoDraw] console API ready: window.__chemForceAutoDraw.start({ intervalMs: 200, silent: true })')
}

const cleanupForceAutoDrawConsoleAPI = () => {
  stopForceAutoDrawFromConsole()
  const win = window as Window & { __chemForceAutoDraw?: any }
  if (win.__chemForceAutoDraw) {
    delete win.__chemForceAutoDraw
  }
}

// 化学键盘确认处理
const handleKeyboardConfirm = async (formula: string) => {
  substanceInput.value = formula
  showChemicalKeyboard.value = false
  await handleInputPlay()
}

const handleLeaveRoom = async () => {
  if (isReplayBridgeMode.value || isReplayRoomPath.value) {
    navigateBackFromReplay()
    return
  }

  // 人机对战模式下，如果玩家已经完成（进入观战状态），点击退出改为“结算”
  if (roomInfo.value?.is_pve && isSpectator.value) {
    try {
      // 调用API通知服务器玩家离开房间
      await gameAPI.leaveRoom(id)
    } catch (error) {
      console.error('离开房间API调用失败:', error)
    }
    // 断开房间连接并停止监听（直接“关闭”房间逻辑）
    websocket.leaveRoom()
    websocket.off('game_update', handleGameUpdate)
    websocket.off('player_joined', handlePlayerJoined)
    websocket.off('player_left', handlePlayerLeft)
    websocket.off('action_toast', handleActionToast)
    websocket.off('room_terminated', handleRoomTerminated)
    websocket.off('player_kicked', handlePlayerKicked)
    websocket.off('chat', handleChatNotify)
    websocket.off('private_chat', handleChatNotify)

    if (gameState.value) {
      isManualSettlement.value = true
      gameState.value.status = 'finished'
      // 构造临时积分数据用于展示（如果尚未结算）
      if (!gameState.value.points_changes) {
        const changes: Record<string, number> = {}
        const finishers = gameState.value.finished_players || []
        
        // 确保所有玩家都在列表中，即便有些还没打完
        const allUIDs = gameState.value.players.map((p: any) => p.uid)
        const remainingUIDs = allUIDs.filter((id: number) => !finishers.includes(id))
        const fullFinishers = [...finishers, ...remainingUIDs]
        
        const quittedCount = gameState.value.quitted_count || 0
        const originalCount = gameState.value.original_player_count || gameState.value.players.length
        const multiplier = originalCount > 0 ? (1.0 - (quittedCount / originalCount)) : 1.0
        const difficulty = roomInfo.value?.pve_difficulty || 100

        fullFinishers.forEach((uid: number, index: number) => {
          const rank = index + 1
          let points = 0
          
          // 后端结算逻辑：
          // 如果是最后一名且总人数大于1，给予固定参与分 5
          if (index === fullFinishers.length - 1 && fullFinishers.length > 1) {
            points = 5
          } else {
            points = Math.floor(100 / rank)
          }

          // 应用倍率 (受中途退出人数影响)
          points = Math.floor(points * multiplier)

          // PvE 模式积分修正
          if (roomInfo.value?.is_pve) {
            // 难度 < 50，无法获得积分
            if (difficulty < 50) {
              points = 0
            } else {
              // 积分 = 原始积分 * (难度 / 100)
              points = Math.floor(points * (difficulty / 100.0))
            }
          }

          if (points < 1 && (!roomInfo.value?.is_pve || difficulty >= 50)) {
            points = 1
          }

          changes[String(uid)] = points
        })
        gameState.value.points_changes = changes
      }
    }
    router.push('/')
    return
  }

  try {
    // 对局进行中离开会被判为逃跑并影响积分，需如实告知（与大厅退出提示一致）
    let message = gameState.value?.status === 'playing'
      ? '对局仍在进行中，现在离开将被系统判定为逃跑，可能扣除积分。确定要离开吗？'
      : '确定要离开实验室吗？'
    let title = '离开实验'

    // 如果玩家已完成实验且是积分模式，提示已获得分值并可直接安全离开
    if (isSpectator.value) {
      const rank = (gameState.value?.finished_players || []).indexOf(user.value?.uid) + 1
      const points = rank === 1 ? 100 : (rank === 2 ? 50 : (rank === 3 ? 33 : 25))
      message = `实验已完成！你获得了第 ${rank} 名，系统已为你发放约 ${points} 燃素。确定现在结束这次实验吗？`
      title = '实验结算'
    }

    const confirmed = await showConfirm(message, title)
    if (confirmed) {
      feedback.click()
      // 调用API通知服务器玩家彻底离开房间
      // 注意：即便是正在游戏中，用户点击“退出”也应该执行 leaveRoom 逻辑
      // 以释放该用户的“同时只能进行一次游戏”锁定
      try {
        await gameAPI.leaveRoom(id)
      } catch (error) {
        console.error('离开房间API调用失败:', error)
      }
      
      // 断开房间连接并停止监听
      websocket.leaveRoom()
      router.push('/')
    }
  } catch (error) {
    console.error('离开房间失败:', error)
    router.push('/')
  }
}

// 生成分享链接（私密房间自动带密钥）
const shareLink = computed(() => {
  const currentUrl = window.location.href
  // 如果是私密房间且有访问密钥，自动添加key参数
  if (roomInfo.value?.is_private && roomInfo.value?.access_key) {
    // 检查URL中是否已经包含key参数
    if (!currentUrl.includes('?key=')) {
      const separator = currentUrl.includes('?') ? '&' : '?'
      return `${currentUrl}${separator}key=${roomInfo.value.access_key}`
    }
  }
  return currentUrl
})

const handleCopyLink = async () => {
  try {
    await navigator.clipboard.writeText(shareLink.value)
    feedback.success()
    if (roomInfo.value?.is_private) {
      showToast('私密房间邀请链接已复制（含访问密钥），快发送给你的科研伙伴吧！', '任务下达', 'success')
    } else {
      showToast('实验邀请链接已复制到剪贴板，快发送给你的科研伙伴吧！', '任务下达', 'success')
    }
  } catch (err) {
    feedback.error()
    showToast('链接复制失败，请手动复制浏览器地址栏', '设备故障', 'error')
  }
}

const getDynamicCardClass = (card: any, formula?: string) => {
  if (!card) {
    if (formula) {
      const elements = formula.match(/[A-Z][a-z]?/g) || []
      if (elements.length > 1) return 'card-reaction'
      if (elements.length === 1 && ELEMENTS_DATA[elements[0]]) return ELEMENTS_DATA[elements[0]].class
    }
    return ''
  }

  // 特殊性质卡牌优先
  const nobleGases = ['He', 'Ne', 'Ar', 'Kr']
  if (nobleGases.includes(card.type)) return 'card-noble'
  if (card.effect || card.type === 'Au') return 'card-func'

  // 如果提供了分子式（通常是反应结果）
  if (formula) {
    const elements = formula.match(/[A-Z][a-z]?/g) || []
    // 判读是否为化合物（包含多种元素）
    if (elements.length > 1) return 'card-reaction'
    // 单质则使用该元素的颜色
    if (elements.length === 1 && ELEMENTS_DATA[elements[0]]) return ELEMENTS_DATA[elements[0]].class
  }

  // 基础元素颜色
  if (ELEMENTS_DATA[card.type]) return ELEMENTS_DATA[card.type].class
  
  return ''
}

const setupDraggable = (el: HTMLElement | null) => {
  if (!el) return
  let isDown = false
  let startX: number
  let scrollLeft: number

  const onMouseDown = (e: MouseEvent) => {
    isDown = true
    el.style.cursor = 'grabbing'
    startX = e.pageX - el.offsetLeft
    scrollLeft = el.scrollLeft
  }

  const onMouseLeave = () => {
    isDown = false
    el.style.cursor = 'grab'
  }

  const onMouseUp = () => {
    isDown = false
    el.style.cursor = 'grab'
  }

  const onMouseMove = (e: MouseEvent) => {
    if (!isDown) return
    e.preventDefault()
    const x = e.pageX - el.offsetLeft
    const walk = (x - startX) * 2
    el.scrollLeft = scrollLeft - walk
  }

  el.addEventListener('mousedown', onMouseDown)
  el.addEventListener('mouseleave', onMouseLeave)
  el.addEventListener('mouseup', onMouseUp)
  el.addEventListener('mousemove', onMouseMove)

  return () => {
    el.removeEventListener('mousedown', onMouseDown)
    el.removeEventListener('mouseleave', onMouseLeave)
    el.removeEventListener('mouseup', onMouseUp)
    el.removeEventListener('mousemove', onMouseMove)
  }
}

onMounted(() => {
  isMobile.value = window.innerWidth < 640

  // 移动端自动关闭提示面板
  if (isMobile.value) {
    showHints.value = false
  }

  const handleResize = () => {
    isMobile.value = window.innerWidth < 640
  }
  window.addEventListener('resize', handleResize)

  // 移动端自动全屏
  if (isMobile.value) {
    // 用户首次交互后请求全屏
    const onFirstInteraction = () => {
      requestFullscreen()
      document.removeEventListener('touchstart', onFirstInteraction)
      document.removeEventListener('click', onFirstInteraction)
    }
    document.addEventListener('touchstart', onFirstInteraction, { once: true })
    document.addEventListener('click', onFirstInteraction, { once: true })
  }

  // 初始化拖拽滑动
  let cleanupHand: (() => void) | undefined
  let cleanupSubstances: (() => void) | undefined
  setTimeout(() => {
    cleanupHand = setupDraggable(handContainer.value)
    cleanupSubstances = setupDraggable(substancesContainer.value)
  }, 500)

  onUnmounted(() => {
    window.removeEventListener('resize', handleResize)
    cleanupHand?.()
    cleanupSubstances?.()
  })
})

// 监听当前玩家变化，自动滚动到行动玩家
watch(() => gameState.value?.current_player, () => {
  nextTick(() => scrollToActivePlayer())
})
</script>

<template>
  <div :class="pageClassNames.gameRoom">
    <!-- 连接中断 / 重连反馈遮罩 -->
    <ConnectionOverlay />
    <!-- Loading State -->
    <div v-if="loading" class="h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 flex flex-col items-center justify-center p-4 relative overflow-hidden">
      <!-- Background Elements -->
      <div class="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] bg-blue-600/20 dark:bg-blue-500/30 rounded-full blur-[120px] animate-pulse"></div>
      <div class="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-purple-600/20 dark:bg-purple-500/30 rounded-full blur-[120px]"></div>
      <div class="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/carbon-fibre.png')] opacity-20"></div>

      <div class="relative z-10 flex flex-col items-center gap-6 animate-in fade-in zoom-in duration-700">
        <div class="relative group">
          <div class="w-24 h-24 bg-blue-500/20 dark:bg-blue-500/30 border-2 border-blue-500/50 dark:border-blue-400/50 rounded-[32px] flex items-center justify-center transform rotate-12 group-hover:rotate-0 transition-all duration-700 shadow-lg shadow-blue-500/20">
            <FlaskConical class="w-12 h-12 text-blue-600 dark:text-blue-400 group-hover:scale-110 transition-transform drop-shadow-lg" />
          </div>
          <div class="absolute -top-2 -right-2 w-8 h-8 bg-blue-500 dark:bg-blue-400 rounded-xl flex items-center justify-center animate-bounce shadow-[0_0_20px_rgba(59,130,246,0.5)]">
             <Zap class="w-4 h-4 text-white fill-current" />
          </div>
        </div>
        <div class="text-center space-y-3">
          <h2 class="text-2xl font-black text-slate-800 dark:text-white tracking-widest uppercase drop-shadow-lg">Initializing Lab</h2>
          <p class="text-sm text-slate-600 dark:text-slate-300 font-medium">正在连接实验室...</p>
          <div class="flex items-center gap-1 justify-center">
             <span class="w-2 h-2 bg-blue-500 dark:bg-blue-400 rounded-full animate-bounce [animation-delay:-0.3s] shadow-lg shadow-blue-500/50"></span>
             <span class="w-2 h-2 bg-blue-500 dark:bg-blue-400 rounded-full animate-bounce [animation-delay:-0.15s] shadow-lg shadow-blue-500/50"></span>
             <span class="w-2 h-2 bg-blue-500 dark:bg-blue-400 rounded-full animate-bounce shadow-lg shadow-blue-500/50"></span>
          </div>
        </div>
      </div>
    </div>

    <!-- Error / No Data State - 防止黑屏 -->
    <div v-else-if="!roomInfo" class="h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 flex flex-col items-center justify-center p-4 relative overflow-hidden">
      <div class="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] bg-red-600/10 dark:bg-red-500/20 rounded-full blur-[120px]"></div>
      <div class="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-purple-600/10 dark:bg-purple-500/20 rounded-full blur-[120px]"></div>

      <div class="relative z-10 flex flex-col items-center gap-6">
        <div class="w-24 h-24 bg-red-500/10 dark:bg-red-500/20 border-2 border-red-500/30 rounded-[32px] flex items-center justify-center shadow-lg">
          <Activity class="w-12 h-12 text-red-500 dark:text-red-400" />
        </div>
        <div class="text-center space-y-3">
          <h2 class="text-2xl font-black text-slate-800 dark:text-white tracking-widest uppercase">Connection Lost</h2>
          <p class="text-sm text-slate-600 dark:text-slate-300 font-medium">{{ loadError || '实验室连接异常' }}</p>
        </div>
        <div class="flex items-center gap-3 mt-4">
          <button
            @click="loadError = null; loading = true; isReplayBridgeMode ? loadReplaySimulationState() : loadGameState()"
            class="px-6 py-3 bg-blue-600 hover:bg-blue-500 text-white font-black rounded-xl transition-all shadow-lg active:scale-95 uppercase tracking-widest text-xs flex items-center gap-2"
          >
            <RefreshCw class="w-4 h-4" />
            重新连接
          </button>
          <button
            @click="router.push('/')"
            class="px-6 py-3 bg-slate-200 dark:bg-white/10 hover:bg-slate-300 dark:hover:bg-white/20 text-slate-700 dark:text-white font-black rounded-xl transition-all shadow-lg active:scale-95 uppercase tracking-widest text-xs flex items-center gap-2"
          >
            <ArrowLeft class="w-4 h-4" />
            返回大厅
          </button>
        </div>
      </div>
    </div>

    <template v-else>
      <!-- Dynamic Background -->
      <div class="fixed inset-0 pointer-events-none">
        <div class="absolute top-1/4 left-1/4 w-[50%] h-[50%] bg-blue-600/5 rounded-full blur-[150px] animate-pulse"></div>
        <div class="absolute bottom-1/4 right-1/4 w-[50%] h-[50%] bg-purple-600/5 rounded-full blur-[150px] animate-pulse delay-1000"></div>
        <div class="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/carbon-fibre.png')] opacity-20"></div>
        <!-- Scanning Line -->
        <div class="absolute top-0 left-0 w-full h-px bg-blue-500/20 shadow-[0_0_15px_rgba(59,130,246,0.5)] animate-scan"></div>
      </div>

      <!-- Compressed Header - 移动端优化 -->
      <header class="h-11 sm:h-16 bg-white/70 dark:bg-black/60 backdrop-blur-3xl border-b border-slate-200 dark:border-white/5 px-2 sm:px-6 flex items-center gap-2 sm:gap-3 z-50 sticky top-0 overflow-x-auto custom-scrollbar-hidden">
        <div class="flex items-center gap-2 sm:gap-4 shrink-0">
          <button
            @click="handleLeaveRoom"
            class="btn-touch flex items-center gap-1.5 px-3 py-1.5 hover:bg-slate-100 dark:hover:bg-white/10 rounded-xl text-slate-500 hover:text-blue-500 transition-all touch-feedback border border-transparent hover:border-blue-500/20"
          >
            <ArrowLeft v-if="!(roomInfo?.is_pve && isSpectator)" class="icon-touch" />
            <Trophy v-else class="w-4 h-4 text-amber-500" />
            <span class="text-[10px] font-black uppercase tracking-widest">{{ isReplayBridgeMode ? '返回时间线' : ((roomInfo?.is_pve && isSpectator) ? '结算实验' : '') }}</span>
          </button>
          <div class="hidden xs:block">
            <h2 class="text-xs-mobile font-black tracking-widest uppercase font-mono text-slate-400">Node: {{ roomInfo?.name || id.substring(0, 6) }}</h2>
            <div class="flex items-center gap-1.5">
              <PingDisplay />
            </div>
          </div>
        </div>

        <!-- Center Area: Reaction Display & Turn Indicator -->
        <div class="flex-1 flex justify-center items-center gap-1.5 sm:gap-4 overflow-hidden">
            <!-- Reaction Widget - 反应记录监控 -->
            <transition name="reaction-slide">
              <div v-if="gameState?.current_reaction" class="flex items-center gap-1.5 sm:gap-3 px-2 sm:px-4 py-1 sm:py-1.5 rounded-lg sm:rounded-xl bg-emerald-500/5 dark:bg-emerald-500/10 border border-emerald-500/20 dark:border-emerald-400/20 shadow-sm backdrop-blur-md shrink-0 max-w-[120px] xs:max-w-[180px] sm:max-w-none group hover:border-emerald-500/40 transition-colors">
                  <div class="flex flex-col items-start leading-none gap-0.5 min-w-0">
                    <span class="text-[5px] sm:text-[7px] font-black uppercase tracking-[0.2em] text-emerald-600 dark:text-emerald-400 opacity-80 flex items-center gap-1">
                      <Binary class="w-1.5 h-1.5 sm:w-2 sm:h-2" />
                      REACTION_SYNC
                    </span>
                    <span class="text-[8px] sm:text-[11px] font-mono font-black text-slate-700 dark:text-emerald-300 drop-shadow-sm truncate">
                      {{ gameState.current_reaction }}
                    </span>
                  </div>
                  <div class="w-1 h-1 sm:w-1.5 sm:h-1.5 rounded-full bg-emerald-500 animate-pulse shrink-0 shadow-[0_0_8px_rgba(16,185,129,0.5)]"></div>
              </div>
            </transition>

            <!-- Top Turn Indicator Inline -->
            <div v-if="gameState?.status === 'playing'" class="animate-in fade-in zoom-in duration-500">
              <div :class="cn(
                'flex items-center gap-2 sm:gap-3 px-3 sm:px-4 py-1 sm:py-1.5 rounded-xl border shadow-sm backdrop-blur-xl transition-all duration-500 relative overflow-hidden',
                isMyTurn 
                  ? 'bg-gradient-to-r from-blue-600 to-indigo-600 border-blue-400 text-white ring-4 ring-blue-500/10' 
                  : 'bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/10 text-slate-700 dark:text-slate-200'
              )">
                <!-- Time Progress Bar -->
                <div 
                  v-if="hasTurnLimit"
                  class="absolute bottom-0 left-0 h-[3px] transition-all duration-100 ease-linear"
                  :class="[
                    isMyTurn ? 'bg-white/30' : 'bg-blue-500/30',
                    timeRemaining <= 5 && 'bg-rose-500/50'
                  ]"
                  :style="{ width: `${timePercent}%` }"
                ></div>

                <div class="relative flex items-center justify-center shrink-0">
                  <div v-if="isMyTurn" class="absolute inset-0 bg-white rounded-full blur-sm animate-pulse"></div>
                  <Zap v-if="isMyTurn" class="w-3 sm:w-3.5 h-3 sm:h-3.5 fill-current relative z-10" />
                  <Timer v-else class="w-2.5 sm:w-3 h-2.5 sm:h-3 relative z-10" :class="hasTurnLimit && timeRemaining <= 10 && 'text-rose-500 animate-spin-slow'" />
                </div>
                <div class="flex flex-col items-start leading-none gap-0.5 min-w-0">
                  <span class="text-[7px] sm:text-[9px] font-black uppercase tracking-widest opacity-70">{{ isMyTurn ? 'Your Operation' : 'Active' }}</span>
                  <span class="text-[9px] sm:text-xs font-black uppercase tracking-tight truncate max-w-[80px] sm:max-w-[120px]" :class="!isMyTurn && shouldShowInBlue(currentPlayerObj) ? 'text-blue-600 dark:text-blue-400' : ''">
                    {{ isMyTurn ? '轮到你了' : getPlayerDisplayName(currentPlayerObj) }}
                  </span>
                </div>
                <div v-if="isMyTurn" class="pl-2 border-l border-white/20">
                  <span v-if="hasTurnLimit" class="text-[10px] font-mono font-black">{{ timeRemaining }}S</span>
                  <span v-else class="text-[10px] font-mono font-black">∞</span>
                </div>
                <!-- Mini Direction indicator -->
                <div v-if="!isMyTurn" class="pl-2 ml-1 border-l border-slate-200 dark:border-white/10 shrink-0">
                   <div class="w-1.5 h-1.5 rounded-full" :class="gameState?.direction === 1 ? 'bg-blue-500' : 'bg-amber-500'"></div>
                </div>
              </div>
            </div>
        </div>

        <!-- Global Status -->
        <div class="flex items-center gap-2 sm:gap-1.5 pl-3 border-l border-slate-200 dark:border-white/10 shrink-0">
          <button @click="feedback.click(); showPlayers = !showPlayers" class="btn-touch relative flex items-center justify-center gap-1 bg-slate-100 dark:bg-white/5 rounded-lg border border-slate-200 dark:border-white/10 text-slate-500 hover:text-blue-500 touch-feedback">
             <Users class="icon-touch" :class="showPlayers && 'fill-current text-blue-500'" />
             <span class="text-[10px] sm:text-xs-mobile font-black text-slate-400">{{ allPlayers.length }}</span>
          </button>

           <button v-if="!isReplayBridgeMode" @click="feedback.click(); showHints = !showHints" class="btn-touch flex items-center justify-center bg-slate-100 dark:bg-white/5 rounded-lg border border-slate-200 dark:border-white/10 text-slate-500 hover:text-blue-500 touch-feedback">
             <Sparkles class="icon-touch" :class="showHints && 'fill-current text-blue-500'" />
          </button>

           <button v-if="!isReplayBridgeMode" @click="feedback.click(); showChat = !showChat; hasNewMessage = false" class="btn-touch relative flex items-center justify-center bg-slate-100 dark:bg-white/5 rounded-lg border border-slate-200 dark:border-white/10 text-slate-500 hover:text-blue-500 touch-feedback">
             <MessageCircle class="icon-touch" :class="showChat && 'fill-current text-blue-500'" />
             <div v-if="hasNewMessage" class="absolute -top-1 -right-1 w-3 h-3 sm:w-2.5 sm:h-2.5 bg-rose-500 border-2 border-white dark:border-[#0d0d10] rounded-full animate-pulse"></div>
          </button>
        </div>
      </header>

      <div v-if="isReplayBridgeMode" class="relative z-[80] px-3 py-2 bg-amber-500/10 border-b border-amber-500/20 pointer-events-auto">
        <div class="max-w-[1400px] mx-auto flex flex-wrap items-center justify-between gap-2">
          <div class="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-amber-700 dark:text-amber-300">
            <span>Replay Simulation</span>
            <span class="px-2 py-0.5 rounded-md bg-amber-500/20">{{ replayStatusText }}</span>
            <span class="px-2 py-0.5 rounded-md bg-cyan-500/20 text-cyan-700 dark:text-cyan-300">视角 {{ replayPerspectiveName }}</span>
            <span class="text-amber-600/90 dark:text-amber-200">输入选项已禁用</span>
          </div>

          <div class="flex items-center gap-2">
            <span class="text-[10px] font-black uppercase tracking-widest text-amber-700 dark:text-amber-300">倍速</span>
            <div class="inline-flex items-center gap-1 p-1 rounded-lg border border-amber-500/20 bg-white/60 dark:bg-white/5">
              <button
                v-for="speed in replaySpeedOptions"
                :key="`speed-${speed}`"
                type="button"
                @click.stop.prevent="setReplaySpeed(Number(speed))"
                :class="cn('px-2 py-1 rounded text-[10px] font-black', replaySpeed === speed ? 'bg-amber-500 text-white' : 'text-amber-700 dark:text-amber-300 hover:bg-amber-500/15')"
              >
                {{ speed }}x
              </button>
            </div>

            <div class="inline-flex items-center gap-1 p-1 rounded-lg border border-amber-500/20 bg-white/60 dark:bg-white/5">
              <button
                type="button"
                @click.stop.prevent="jumpReplayByStep(-1)"
                :disabled="replayPlaybackIndex === 0"
                :class="cn('px-2 py-1 rounded text-[10px] font-black uppercase tracking-widest', replayPlaybackIndex === 0 ? 'text-slate-400 cursor-not-allowed' : 'text-amber-700 dark:text-amber-300 hover:bg-amber-500/15')"
              >
                上一步
              </button>
              <button
                type="button"
                @click.stop.prevent="jumpReplayByStep(1)"
                :disabled="replayPlaybackIndex >= replayEvents.length"
                :class="cn('px-2 py-1 rounded text-[10px] font-black uppercase tracking-widest', replayPlaybackIndex >= replayEvents.length ? 'text-slate-400 cursor-not-allowed' : 'text-amber-700 dark:text-amber-300 hover:bg-amber-500/15')"
              >
                下一步
              </button>
            </div>

            <div class="flex items-center gap-2 px-2 py-1 rounded-lg border border-amber-500/20 bg-white/60 dark:bg-white/5">
              <span class="text-[10px] font-black uppercase tracking-widest text-amber-700 dark:text-amber-300">定位</span>
              <input
                type="range"
                min="0"
                :max="replayEvents.length"
                :value="replayPlaybackIndex"
                @input="handleReplaySeekInput"
                class="w-28 accent-amber-500"
              />
              <span class="text-[10px] font-black text-amber-700 dark:text-amber-300">{{ replayPlaybackIndex }}</span>
            </div>

            <button
              type="button"
              @click.stop.prevent="toggleReplayPlayback"
              :disabled="replayGameOver"
              :class="cn('px-3 py-1.5 rounded-lg text-[10px] font-black uppercase tracking-widest border', replayGameOver ? 'border-slate-300 text-slate-400 cursor-not-allowed' : 'border-amber-500/30 text-amber-700 dark:text-amber-300 hover:bg-amber-500/15')"
            >
              {{ replayIsPlaying ? '暂停' : '继续' }}
            </button>

            <button
              type="button"
              @click.stop.prevent="restartReplayPlayback()"
              class="px-3 py-1.5 rounded-lg text-[10px] font-black uppercase tracking-widest border border-emerald-500/30 text-emerald-700 dark:text-emerald-300 hover:bg-emerald-500/15"
            >
              重新播放
            </button>

            <span class="text-[10px] font-black uppercase tracking-widest text-amber-700 dark:text-amber-300">进度 {{ replayProgressText }}</span>
          </div>
        </div>
      </div>

      <!-- PvE Experimental Dynamics Floating Window -->
      <div v-if="roomInfo?.is_pve && pveToasts.length > 0" class="fixed top-14 sm:top-20 left-1/2 -translate-x-1/2 z-[60] flex flex-col items-center gap-2 pointer-events-none">
        <div 
          v-for="toast in pveToasts" 
          :key="toast.id"
          class="bg-blue-600/90 backdrop-blur-md text-white px-4 py-2 rounded-xl shadow-lg border border-white/20 text-[10px] sm:text-xs font-bold animate-in slide-in-from-top-2 fade-in duration-300 flex items-center gap-2"
        >
          <Activity class="w-3.5 h-3.5 text-blue-200" />
          {{ toast.text }}
        </div>
      </div>

      <!-- Input box and Timer - 顶栏下方7px -->
      <div v-if="isMyTurn" class="relative w-full flex justify-center px-4 z-[60]" style="margin-top: 7px;">
        <div class="flex flex-col items-center gap-2 sm:gap-2 animate-in slide-in-from-top-4">
          <div class="flex items-center bg-white/90 dark:bg-black/80 backdrop-blur-xl border border-slate-200 dark:border-white/10 rounded-xl sm:rounded-lg p-1 sm:p-0.5 shadow-xl">
            <input
              v-model="substanceInput"
              @keyup.enter="handleInputPlay"
              @focus="handleInputFocus"
              @blur="handleInputBlur"
              placeholder="手动注入化学式"
              :inputmode="isMobile || user.enable_element_input ? 'none' : 'text'"
              autocomplete="off"
              class="bg-transparent border-none outline-none text-sm sm:text-xs-mobile px-3 sm:px-2 py-1.5 sm:py-0.5 w-32 sm:w-40 font-black tracking-widest placeholder:text-slate-400 text-slate-900 dark:text-white"
            />

            <div class="flex items-center gap-1">
               <button
                  @click="handleInputPlay"
                  class="btn-touch bg-blue-600 hover:bg-blue-500 rounded-lg sm:rounded-md flex items-center justify-center transition-all touch-feedback shadow-md group"
                  title="执行反应"
               >
                  <ChevronRight class="w-4 h-4 sm:w-3.5 sm:h-3.5 text-white group-hover:translate-x-0.5 transition-transform" />
               </button>

               <div class="w-px h-5 sm:h-4 bg-slate-200 dark:bg-white/10 mx-1 sm:mx-0.5"></div>

               <button
                  v-if="!tutorialScriptMode"
                  @click="handleDrawCard"
                  :disabled="!isMyTurn"
                  :class="cn(
                    'px-3 sm:px-2 btn-touch rounded-lg sm:rounded-md flex items-center justify-center gap-1.5 sm:gap-1 transition-all touch-feedback shadow-md group relative overflow-hidden',
                    isMyTurn ? (gameState?.pending_draw_count > 0 ? 'bg-red-600 hover:bg-red-500 text-white' : 'bg-slate-800 dark:bg-white/10 hover:bg-slate-700 dark:hover:bg-white/20 text-white') : 'bg-slate-200 dark:bg-slate-800 text-slate-400 cursor-not-allowed grayscale'
                  )"
               >
                  <Plus v-if="!(gameState?.pending_draw_count > 0)" class="w-3 h-3 sm:w-2.5 sm:h-2.5" />
                  <RefreshCw v-else class="w-3 h-3 sm:w-2.5 sm:h-2.5 animate-spin-slow" />
                  <span class="text-xs-mobile font-black uppercase tracking-widest whitespace-nowrap">
                    摸牌{{ gameState?.pending_draw_count > 0 ? gameState.pending_draw_count : '2' }}张
                  </span>
               </button>
            </div>
          </div>

          <div class="flex items-center gap-2 sm:gap-1.5">
            <div class="bg-blue-600/90 backdrop-blur-md px-4 sm:px-3 py-2 sm:py-1 rounded-full border border-white/20 shadow-md flex items-center gap-2.5 sm:gap-2">
              <Zap class="w-3 h-3 sm:w-2.5 sm:h-2.5 fill-current animate-pulse text-white" />
              <span class="text-xs-mobile font-black uppercase tracking-widest text-white">
                操作 ({{ hasTurnLimit ? (timeRemaining + 's') : '无时限' }})
              </span>

              <!-- 双联行动按钮 -->
              <button
                v-if="myData?.double_action_available"
                @click.stop="toggleDoubleMode"
                :class="cn(
                  'px-3 sm:px-2.5 py-1.5 sm:py-1 rounded-2xl border border-white/20 transition-all flex items-center gap-2 relative overflow-hidden touch-feedback',
                  doubleMode 
                    ? 'bg-blue-600 text-white border-blue-400 shadow-[0_0_15px_rgba(37,99,235,0.4)]' 
                    : 'bg-white/10 text-white/70 hover:text-white hover:bg-white/20'
                )"
              >
                 <div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/10 to-transparent -translate-x-full group-hover/btn:animate-shimmer"></div>
                 <Activity :class="cn('w-4 h-4 sm:w-3.5 sm:h-3.5', doubleMode && 'animate-spin')" />
                 <span class="text-[10px] font-black uppercase tracking-widest">{{ doubleMode ? '解除超限操作' : '发动超限双联' }}</span>
              </button>

              <!-- 强制出牌提示 -->
              <div
                v-if="isMyTurn && gameState?.pending_forced_plays > 0"
                class="flex items-center gap-1.5 px-3 py-1.5 rounded-2xl bg-blue-500/80 border border-blue-400/50 shadow-[0_0_12px_rgba(59,130,246,0.3)] animate-pulse pointer-events-none"
              >
                <Zap class="w-3.5 h-3.5 fill-current text-white" />
                <span class="text-[10px] font-black uppercase tracking-widest text-white whitespace-nowrap">强制出牌 ×{{ gameState.pending_forced_plays }}</span>
              </div>
            </div>
          </div>

          <!-- 双联模式提示状态 -->
          <div v-if="doubleMode" class="mt-1 flex flex-wrap items-center justify-center gap-3 animate-in slide-in-from-top-4 duration-500">
            <div class="flex items-center gap-2">
              <div
                @click="firstDoubleSubstance && removeSubstance(1)"
                :class="cn(
                  'w-8 h-8 rounded-lg flex items-center justify-center border-2 transition-all duration-300 relative group/sub',
                  firstDoubleSubstance ? 'bg-blue-500/20 border-blue-500 shadow-md cursor-pointer hover:border-red-500/50' : 'bg-slate-800/50 border-white/10 opacity-50'
                )"
              >
                <span v-if="firstDoubleSubstance" class="text-[9px] font-black transition-opacity" v-html="formatFormula(firstDoubleSubstance)"></span>
                <X v-if="firstDoubleSubstance" class="w-2.5 h-2.5 text-red-500 absolute top-0.5 right-0.5 transition-opacity" />
                <FlaskConical v-else class="w-3 h-3 text-slate-500" />
              </div>
              <div class="w-3 h-0.5 bg-blue-500/30"></div>
              <div
                @click="secondDoubleSubstance && removeSubstance(2)"
                :class="cn(
                  'w-8 h-8 rounded-lg flex items-center justify-center border-2 transition-all duration-300 relative group/sub',
                  secondDoubleSubstance ? 'bg-blue-500/20 border-blue-500 shadow-md cursor-pointer hover:border-red-500/50' : 'bg-slate-800/50 border-white/10 opacity-50'
                )"
              >
                <span v-if="secondDoubleSubstance" class="text-[9px] font-black transition-opacity" v-html="formatFormula(secondDoubleSubstance)"></span>
                <X v-if="secondDoubleSubstance" class="w-2.5 h-2.5 text-red-500 absolute top-0.5 right-0.5 transition-opacity" />
                <FlaskConical v-else class="w-3 h-3 text-slate-500" />
              </div>
            </div>

            <div class="flex items-center gap-1.5">
              <button
                v-if="firstDoubleSubstance && secondDoubleSubstance"
                @click="handleDoublePlay"
                class="bg-emerald-600 hover:bg-emerald-500 text-white px-3 py-1.5 rounded-xl flex items-center gap-1.5 shadow-md animate-in zoom-in duration-300 group"
              >
                <span class="text-[9px] font-black uppercase tracking-widest">启动反应</span>
                <Play class="w-3 h-3 fill-current group-hover:translate-x-0.5 transition-transform" />
              </button>

              <button
                @click="toggleDoubleMode"
                class="bg-slate-800/80 hover:bg-slate-700 text-white/80 px-3 py-1.5 rounded-xl flex items-center gap-1.5 border border-white/10 shadow-md transition-all"
              >
                <span class="text-[9px] font-black uppercase tracking-widest">取消</span>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Main Action Focus Area -->
      <div class="absolute left-0 right-0 flex flex-col items-center overflow-hidden px-2 sm:px-4"
           :style="{
             top: isMobile ? '44px' : '64px',
             bottom: showChemicalKeyboard ? '144px' : '0',
             paddingBottom: showChemicalKeyboard ? '0' : (isMobile ? '4rem' : '5rem'),
             justifyContent: 'center',
             transform: 'translateY(-100px)'
           }">
          <!-- Left Sidebar: Hint & Status -->
          <Teleport to="body" :disabled="isReplayBridgeMode">
             <div v-if="showHints" class="fixed inset-0 bg-white/10 dark:bg-black/20 backdrop-blur-[2px] z-[95] lg:hidden clickable" @click="showHints = false"></div>
             <div :class="cn(
               'fixed left-0 top-0 bottom-0 w-full lg:w-80 z-[100] bg-white/95 dark:bg-slate-900/60 backdrop-blur-3xl border-r lg:border border-slate-200 dark:border-white/10 lg:rounded-[40px] lg:top-6 lg:bottom-52 lg:left-6 shadow-3xl transition-all duration-500 flex flex-col overflow-hidden',
               showHints ? 'translate-x-0 opacity-100' : '-translate-x-full opacity-0 pointer-events-none'
             )">
             <div class="p-4 py-3 border-b border-slate-200 dark:border-white/10 flex items-center justify-between bg-slate-50/50 dark:bg-white/[0.02]">
                <div class="flex items-center gap-2">
                   <div class="w-6 h-6 rounded-lg bg-blue-500/10 flex items-center justify-center">
                      <Trophy class="w-3.5 h-3.5 text-blue-500" />
                   </div>
                   <div>
                      <h3 class="text-[10px] font-black uppercase tracking-widest text-slate-800 dark:text-white">实验辅助情报</h3>
                      <p class="text-[8px] font-mono text-slate-400 uppercase tracking-tighter">Intelligence_Protocol</p>
                   </div>
                </div>
                <button @click="feedback.click(); showHints = false" class="p-1 hover:bg-slate-200 dark:hover:bg-white/10 rounded-lg transition-colors text-slate-400 hover:text-slate-600 dark:hover:text-white">
                   <ArrowLeft class="w-4 h-4" />
                </button>
             </div>
             
             <div class="flex-1 overflow-y-auto p-3 custom-scrollbar space-y-4">
                <!-- Status Banners -->
                <div class="space-y-2">
                   <div v-if="allowedAny" class="bg-amber-500/10 border border-amber-500/20 p-2.5 rounded-xl animate-pulse">
                      <div class="flex items-center gap-1.5 text-amber-500 mb-0.5">
                         <Zap class="w-3 h-3 fill-current" />
                         <span class="text-[9px] font-black uppercase tracking-wider">AU 特权激活</span>
                      </div>
                      <p class="text-[8px] font-bold text-slate-500">已跳过所有反应规则限制</p>
                   </div>

                   <div v-if="gameState?.pending_draw_count > 0" class="bg-red-500/10 border border-red-500/20 p-2.5 rounded-xl animate-bounce">
                      <div class="flex items-center gap-1.5 text-red-500 mb-0.5">
                         <RefreshCw class="w-3 h-3 animate-spin-slow" />
                         <span class="text-[9px] font-black uppercase tracking-wider">正在加牌</span>
                      </div>
                      <p class="text-[8px] font-bold text-slate-500">需结算或叠加累计: {{ gameState.pending_draw_count }}</p>
                   </div>
                </div>

                <div v-if="roomInfo?.status === 'waiting'" class="space-y-3">
                   <!-- 燃素模式提示 -->
                   <div v-if="roomInfo?.is_points_mode" class="p-3 bg-amber-500/10 border border-amber-500/20 rounded-xl flex items-center gap-2">
                      <PhlogistonIcon :size="16" color="#f59e0b" class="shrink-0" />
                      <div class="text-left">
                         <p class="text-[9px] font-black uppercase tracking-widest text-amber-600 dark:text-amber-500">Competitive Mode</p>
                         <p class="text-[8px] font-bold text-slate-500 mt-0.5">燃素竞技模式：胜者获得燃素，败者扣除燃素。</p>
                      </div>
                   </div>

                   <div class="p-3 bg-blue-500/5 border border-blue-500/10 rounded-xl flex flex-col items-center text-center">
                      <Users class="w-5 h-5 text-blue-500 mb-1.5" />
                      <span class="text-[9px] font-black uppercase tracking-widest text-blue-500">准备就绪?</span>
                      <p class="text-[8px] font-bold text-slate-500 mt-0.5">当前连接数 {{ allPlayers.length }}/{{ roomInfo?.max_players }}，等待就绪后自动开启。</p>
                   </div>
                   <div class="p-3 bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl">
                      <div class="flex items-center gap-1.5 mb-1.5">
                         <QrCode class="w-3 h-3 text-blue-500" />
                         <span class="text-[9px] font-black uppercase tracking-widest text-slate-500">快速邀请</span>
                      </div>
                      <p class="text-[7px] font-bold text-slate-400 leading-relaxed uppercase">
                         点击中间区域的"招募伙伴"按钮可快速复制链接，或点击二维码图标让好友扫码加入反应室。
                      </p>
                   </div>
                </div>
                
                <div v-else class="py-8 flex flex-col items-center justify-center opacity-20 text-center">
                   <Timer class="w-6 h-6 mb-2" />
                   <p class="text-[9px] font-black uppercase tracking-widest">等待其他研究员行动</p>
                </div>

                <!-- Reaction-based Hints (场上物质反应提示) -->
                <div v-if="filteredReactionHints.length > 0 && gameState?.status === 'playing' && isMyTurn" class="pt-3 border-t border-slate-200 dark:border-white/10">
                   <div class="flex items-center justify-between mb-3">
                      <div class="flex items-center gap-1.5">
                         <Activity class="w-3 h-3 text-emerald-500" />
                         <span class="text-[9px] font-black uppercase tracking-widest text-slate-500">可接续反应物</span>
                      </div>
                      <button @click="feedback.click(); fetchReactionHints()" class="p-1 hover:bg-slate-100 dark:hover:bg-white/5 rounded-full transition-colors text-slate-400 hover:text-emerald-500">
                         <RefreshCw class="w-2.5 h-2.5" />
                      </button>
                   </div>
                   <div class="space-y-1.5">
                      <button
                         v-for="(hint, idx) in filteredReactionHints"
                         :key="idx"
                         @click="selectedSubstance = hint.substance; handlePlayCard()"
                         class="w-full text-left px-3 py-2 bg-white/50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl hover:border-emerald-500 hover:bg-emerald-500/5 transition-all group cursor-pointer"
                      >
                         <div class="flex items-center justify-between">
                            <span class="text-[10px] font-black dark:text-white" v-html="formatFormula(hint.substance)"></span>
                            <div class="flex items-center gap-1.5">
                              <span class="text-[8px] font-bold text-slate-400 tracking-tighter">{{ getSubstanceName(hint.substance) }}</span>
                              <div class="w-1.5 h-1.5 rounded-full bg-emerald-500 group-hover:scale-125 transition-transform shadow-[0_0_8px_rgba(16,185,129,0.5)]"></div>
                            </div>
                         </div>
                      </button>
                   </div>
                </div>

                <!-- Database Trivia Hints -->
                <div v-if="randomHints.length > 0" class="pt-3 border-t border-slate-200 dark:border-white/10">
                   <div class="flex items-center justify-between mb-3">
                      <div class="flex items-center gap-1.5">
                         <Sparkles class="w-3 h-3 text-blue-500" />
                         <span class="text-[9px] font-black uppercase tracking-widest text-slate-500">实验小贴士</span>
                      </div>
                      <button @click="feedback.click(); fetchRandomHints()" class="p-1 hover:bg-slate-100 dark:hover:bg-white/5 rounded-full transition-colors text-slate-400 hover:text-blue-500">
                         <RefreshCw class="w-2.5 h-2.5" />
                      </button>
                   </div>
                   <div class="space-y-3">
                      <div v-for="hint in randomHints" :key="hint.id" class="relative pl-3">
                         <div class="absolute left-0 top-1 bottom-1 w-0.5 bg-blue-500/30 rounded-full"></div>
                         <h4 v-if="hint.title" class="text-[9px] font-black text-slate-600 dark:text-slate-300 mb-0.5">{{ hint.title }}</h4>
                         <p class="text-[8px] font-bold text-slate-400 leading-relaxed">{{ hint.content }}</p>
                      </div>
                   </div>
                </div>
             </div>
           </div>
          </Teleport>

          <!-- Latest Reaction Display -->
          <div v-if="gameState?.last_card"
               :key="gameState?.last_card?.substance + (gameState?.discard_pile?.length || 0)"
               class="relative group scale-75 sm:scale-85 flex flex-col items-center justify-center animate-stamp">
             <div class="absolute -inset-16 bg-blue-600/10 rounded-full blur-[100px] opacity-50 group-hover:opacity-80 transition-opacity animate-pulse"></div>
             
             <!-- Double Play Display (Side by Side) -->
             <div v-if="gameState?.last_card?.reactants?.length > 0" class="flex items-center gap-6 sm:gap-10 relative z-10">
                <div v-for="(sub, idx) in gameState.last_card.reactants" :key="idx" class="relative group/card">
                   <div :class="cn(
                      'uno-card h-32 sm:h-40 rounded-[32px] flex flex-col items-center justify-center gap-4 hover:scale-105 transition-all duration-500',
                      getDynamicWidthClass(sub, 'double'),
                      getDynamicCardClass(gameState?.last_card?.card, sub)
                   )">
                      <span :class="cn('font-black font-mono italic drop-shadow-lg', getFormulaFontSize(sub, 'double'))" v-html="formatFormula(sub)"></span>
                      <div class="px-3 py-1 bg-white/10 backdrop-blur-md rounded-lg border border-white/20 max-w-[85%]">
                         <span class="text-[8px] font-black tracking-widest truncate block text-center">{{ getSubstanceName(sub) }}</span>
                      </div>
                   </div>
                </div>
                <!-- Plus Operator -->
                <div class="w-10 h-10 rounded-xl bg-white/10 backdrop-blur-md border border-white/20 flex items-center justify-center text-white shadow-lg">
                   <Plus class="w-4 h-4 stroke-[4px]" />
                </div>
             </div>

             <!-- Single Play Display -->
             <div v-else :class="cn(
                'uno-card h-44 sm:h-52 rounded-[32px] flex flex-col items-center justify-center gap-4 sm:gap-6 hover:scale-105 transition-all duration-500',
                getDynamicWidthClass(gameState?.last_card?.substance, 'single'),
                getDynamicCardClass(gameState?.last_card?.card, gameState?.last_card?.substance)
              )">
                <div class="absolute top-4 left-4 opacity-20 text-[8px] uppercase font-black tracking-widest leading-none">Result</div>
                <span :class="cn('font-black font-mono italic drop-shadow-lg leading-none', getFormulaFontSize(gameState?.last_card?.substance, 'single'))" v-html="formatFormula(gameState?.last_card?.substance)"></span>
                <div class="px-4 py-1.5 bg-white/10 backdrop-blur-md rounded-xl border border-white/20 max-w-[85%]">
                   <span class="text-[9px] sm:text-[10px] font-black tracking-widest text-center block leading-tight">{{ getSubstanceName(gameState?.last_card?.substance) }}</span>
                </div>
                <div class="absolute bottom-4 right-4 opacity-30">
                   <FlaskConical class="w-4 h-4 fill-current" />
                </div>
             </div>

             <!-- Direction Ring -->
             <div class="absolute -inset-12 pointer-events-none">
                <div :class="cn(
                   'absolute -inset-12 pointer-events-none border-2 border-blue-500/10 rounded-full',
                   gameState?.direction === 1 ? 'animate-spin-slow' : 'animate-reverse-spin-slow'
                )"></div>
             </div>
          </div>

          <!-- Waiting for play state (Au triggered or Initial) -->
          <div v-else-if="gameState?.status === 'playing' && !gameState?.last_card" class="flex flex-col items-center gap-4 sm:gap-6 animate-in fade-in zoom-in duration-700">
             <div class="relative group">
                <div class="absolute -inset-8 bg-emerald-500/10 rounded-full blur-[60px] group-hover:bg-emerald-500/20 transition-all animate-pulse"></div>
                <div class="w-24 h-24 sm:w-32 sm:h-32 rounded-[32px] sm:rounded-[40px] border-2 border-emerald-500/30 flex items-center justify-center relative z-10 backdrop-blur-md bg-emerald-500/5">
                   <Zap class="w-10 h-10 sm:w-14 sm:h-14 text-emerald-500/40" />
                </div>
             </div>
             <div class="text-center relative z-10">
                <h3 class="text-lg sm:text-xl font-black uppercase tracking-[0.2em]" :class="shouldShowInBlue(allPlayers[gameState?.current_player]) ? 'text-blue-600 dark:text-blue-400' : 'text-slate-800 dark:text-white'">
                   等待 <span>{{ getPlayerDisplayName(allPlayers[gameState?.current_player]) }}</span> 出牌
                </h3>
                <p class="text-[8px] font-bold text-slate-500 mt-1 uppercase italic tracking-tighter">
                   Reaction Reactor Reseted _ New Deployment Window Open
                </p>
             </div>
          </div>
          
          <div v-else-if="roomInfo?.status === 'waiting'" class="flex flex-col items-center gap-3 sm:gap-5 animate-in fade-in zoom-in duration-1000">
             <div class="relative">
                <div class="absolute inset-0 bg-blue-500/10 rounded-full blur-[40px] animate-pulse"></div>
                <!-- Spectator Banner -->
                <div v-if="isSpectator && gameState?.status === 'playing'" class="absolute -top-32 left-1/2 -translate-x-1/2 z-[60] animate-in slide-in-from-top-4 pointer-events-none whitespace-nowrap">
                  <div class="bg-indigo-600/90 backdrop-blur-md text-white px-4 py-1.5 rounded-full shadow-lg border border-white/20 flex items-center gap-2">
                      <Trophy class="w-3.5 h-3.5 text-yellow-300" />
                      <span class="text-[10px] font-bold uppercase tracking-widest">已完成比赛 - 观战模式</span>
                  </div>
                </div>
                <div class="w-20 h-20 sm:w-24 sm:h-24 rounded-[24px] sm:rounded-[32px] border-2 border-dashed border-blue-500/30 flex items-center justify-center rotate-45 group hover:rotate-0 transition-all duration-700 backdrop-blur-md bg-blue-500/5">
                   <FlaskConical class="w-8 h-8 sm:w-10 sm:h-10 text-blue-500/40 -rotate-45 group-hover:rotate-0 transition-all" />
                </div>
                <div v-if="roomInfo?.countdown > 0" class="absolute inset-0 flex items-center justify-center z-50 pointer-events-none">
                   <div class="scale-[2.5] sm:scale-[4] opacity-20 font-black italic select-none animate-ping text-blue-500">
                      {{ roomInfo.countdown }}
                   </div>
                </div>

                <div v-if="roomInfo?.countdown > 0" class="absolute -top-2 -right-2 bg-red-500 text-white px-3 py-1 rounded-lg text-base font-black shadow-lg animate-bounce z-10">
                   {{ roomInfo.countdown }}
                </div>
                <div v-else class="absolute -top-2 -right-2 bg-amber-500 text-white px-2 py-0.5 rounded-md text-[7px] font-black uppercase tracking-widest shadow-lg animate-pulse">
                   Ready Check
                </div>
             </div>

             <div class="flex flex-col items-center gap-3">
                <div class="flex flex-col items-center gap-2">
                  <h3 class="text-base sm:text-lg font-black text-slate-800 dark:text-white uppercase tracking-[0.1em] text-center">{{ roomInfo?.name || '实验室准备中' }}</h3>

                  <!-- Compact Ready Button -->
                  <button
                    @click="handleToggleReady"
                    :class="cn(
                      'px-6 sm:px-8 py-2 sm:py-3 rounded-xl text-xs sm:text-sm font-black uppercase tracking-[0.2em] transition-all duration-500 shadow-lg relative overflow-hidden active:scale-95 text-white',
                      isReady ? 'bg-emerald-500 shadow-emerald-500/40' : 'bg-blue-600 shadow-blue-500/40'
                    )"
                  >
                    <div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 to-transparent -translate-x-full group-hover:animate-shimmer"></div>
                    <div class="flex items-center gap-2">
                      <Zap :class="cn('w-3.5 h-3.5 sm:w-4 sm:h-4', isReady ? 'fill-current' : 'animate-pulse')" />
                      <span>{{ isReady ? '已就绪' : '手动准备' }}</span>
                    </div>
                  </button>

                  <!-- Countdown Tip -->
                  <div v-if="roomInfo?.countdown > 0" class="flex flex-col items-center gap-0.5 mt-1">
                    <p class="text-[9px] font-black uppercase tracking-[0.2em] text-blue-500 animate-pulse">
                      实验即将开始: <span class="text-base">{{ roomInfo.countdown }}</span>S
                    </p>
                    <p class="text-[6px] font-bold text-slate-400 dark:text-slate-600 uppercase tracking-tighter italic">
                      实验室压力充盈中，即将开启研究循环...
                    </p>
                  </div>
                </div>

                <div class="flex flex-col items-center gap-2 bg-white/50 dark:bg-white/5 backdrop-blur-xl p-3 sm:p-4 rounded-2xl border border-slate-200 dark:border-white/10 shadow-sm w-full max-w-sm">
                  <div class="flex flex-wrap justify-center gap-1.5 sm:gap-2">
                    <div class="flex items-center gap-1.5 px-2.5 py-1 bg-slate-100 dark:bg-white/5 rounded-lg border border-slate-200 dark:border-white/10">
                      <Users class="w-2.5 h-2.5 text-blue-500" />
                      <span class="text-[7px] font-black uppercase tracking-widest text-slate-600 dark:text-slate-400">
                        研究员: {{ allPlayers.length }} / {{ roomInfo?.max_players }}
                      </span>
                    </div>
                    <div
                      @click="viewCurrentDeckConfig"
                      class="flex items-center gap-1.5 px-2.5 py-1 bg-slate-100 dark:bg-white/5 rounded-lg border border-slate-200 dark:border-white/10 cursor-pointer hover:bg-slate-200 dark:hover:bg-white/10 transition-colors"
                      title="点击查看牌组详情"
                    >
                      <FlaskConical class="w-2.5 h-2.5 text-emerald-500" />
                      <span class="text-[7px] font-black uppercase tracking-widest text-slate-600 dark:text-slate-400">
                        方案: {{ roomInfo?.deck_config?.name || '基础协议' }}
                      </span>
                    </div>
                  </div>

                  <div class="flex items-center gap-1.5 w-full">
                    <button
                        @click="handleCopyLink"
                        class="flex-1 flex items-center justify-center gap-1.5 py-2 bg-slate-800 dark:bg-white/10 hover:bg-slate-700 text-white rounded-lg transition-all active:scale-95 group shadow-md"
                    >
                        <Copy class="w-2.5 h-2.5 group-hover:rotate-12 transition-transform" />
                        <span class="text-[8px] font-black uppercase tracking-widest">招募成员</span>
                    </button>
                    <button
                        @click="feedback.click(); showInviteFriendsModal = true"
                        class="flex-1 flex items-center justify-center gap-1.5 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg transition-all active:scale-95 group shadow-md"
                    >
                        <UserPlus class="w-2.5 h-2.5 group-hover:scale-110 transition-transform" />
                        <span class="text-[8px] font-black uppercase tracking-widest">邀请好友</span>
                    </button>
                    <button
                        @click="feedback.click(); showQrModal = !showQrModal"
                        class="w-8 h-8 flex items-center justify-center bg-white dark:bg-white/10 border border-slate-200 dark:border-white/10 rounded-lg text-slate-500 hover:text-blue-500 transition-all active:scale-90 shadow-md"
                    >
                        <QrCode class="w-4 h-4" />
                    </button>
                  </div>

                  <!-- QR Code 浮窗 -->
                  <div v-if="showQrModal" class="mt-1 p-2 bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-xl shadow-2xl animate-in zoom-in duration-300 flex flex-col items-center gap-2">
                     <div class="p-1.5 bg-white rounded-lg border-2 border-blue-500/20">
                        <img
                          :src="`https://api.qrserver.com/v1/create-qr-code/?size=100x100&data=${encodeURIComponent(shareLink)}`"
                          alt="Join QR Code"
                          class="w-24 h-24"
                        />
                     </div>
                     <div class="text-center pb-0.5">
                        <p class="text-[8px] font-black uppercase tracking-widest text-blue-500">实验室快传</p>
                     </div>
                  </div>
                </div>
             </div>
          </div>
        <!-- Table Console Background Removed or Simplified -->
          <div class="absolute inset-0 pointer-events-none overflow-hidden">
            <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[65%] h-[70%] bg-slate-300/[0.14] dark:bg-slate-500/[0.08] rounded-full blur-[100px]"></div>
          </div>
      </div>

      <!-- Hand / Deck Area -->
      <div class="fixed bottom-0 left-0 right-0 z-[70] bg-white/70 dark:bg-black/60 backdrop-blur-2xl border-t border-slate-200 dark:border-white/5 flex flex-col items-center">
        <!-- 教学模式提示 - 手牌栏上方，紧凑样式 -->
        <div v-if="isTutorialMode && tutorialHintText && isMyTurn" class="absolute bottom-full mb-4 left-0 right-0 flex justify-center px-4 animate-in slide-in-from-bottom-4 z-50 pointer-events-none">
          <div class="bg-gradient-to-br from-amber-500 to-orange-500 backdrop-blur-xl border border-amber-300 rounded-xl shadow-lg px-4 py-2 max-w-md mx-auto relative overflow-hidden pointer-events-none">
            <!-- 背景装饰 -->
            <div class="absolute inset-0 bg-[radial-gradient(circle_at_30%_50%,rgba(255,255,255,0.15),transparent_50%)]"></div>

            <!-- 脚本进度指示器 -->
            <div v-if="tutorialScriptMode" class="relative mb-2 flex items-center justify-between">
              <div class="text-[9px] font-black text-white/80 uppercase tracking-wider">
                Step {{ tutorialStepDisplay }}/{{ TUTORIAL_TOTAL_STEPS }}
              </div>
              <div class="flex-1 mx-2 h-1 bg-white/20 rounded-full overflow-hidden">
                <div
                  class="h-full bg-white/60 rounded-full transition-all duration-500"
                  :style="{ width: `${tutorialProgressPercent}%` }"
                ></div>
              </div>
              <div class="text-[9px] font-bold text-white/60">
                {{ tutorialProgressPercent }}%
              </div>
            </div>

            <!-- 内容 -->
            <div class="relative flex items-center gap-2">
              <div class="w-6 h-6 rounded-full bg-white/20 flex items-center justify-center flex-shrink-0">
                <Sparkles class="w-4 h-4 text-white" />
              </div>
              <div class="flex-1">
                <p class="text-white text-xs sm:text-sm font-bold leading-snug" v-html="tutorialHintText"></p>
              </div>
            </div>

            <!-- 装饰条 -->
            <div class="absolute bottom-0 left-0 right-0 h-0.5 bg-gradient-to-r from-transparent via-white/30 to-transparent"></div>
          </div>
        </div>

        <div class="w-full max-w-7xl mx-auto flex justify-center items-end py-2 sm:py-1">
           <div ref="handContainer" class="hand-container-mobile w-full custom-scrollbar-hidden">
            <div v-if="isReplayBridgeMode && replayPerspectivePlayer" class="mb-1 flex items-center justify-center">
              <div class="inline-flex items-center gap-2 px-3 py-1 rounded-xl border border-cyan-500/20 bg-cyan-500/10 text-[10px] font-black uppercase tracking-widest text-cyan-700 dark:text-cyan-300">
                <span>当前视角</span>
                <span>{{ replayPerspectiveName }}</span>
              </div>
            </div>
            <div v-if="roomInfo?.status === 'waiting'" class="flex flex-col items-center justify-center opacity-30 pb-1 min-w-full">
              <Loader2 class="w-8 h-8 sm:w-6 sm:h-6 mb-1 animate-spin text-blue-500" />
              <p class="font-black uppercase tracking-widest text-xs-mobile text-slate-500 text-center">正在同步量子状态并等待开场就绪...</p>
            </div>
            <template v-else-if="myData?.hand_cards?.length > 0">
              <div
                v-for="(card, index) in myData.hand_cards"
                :key="index"
                role="button"
                :tabindex="isMyTurn ? 0 : -1"
                :aria-disabled="!isMyTurn"
                :aria-pressed="selectedCard === card"
                :aria-label="`${card.type}${ELEMENTS_DATA[card.type] ? ' ' + getSubstanceName(card.type) : ''}${card.effect ? ' 功能牌' : ''}${isMyTurn ? '' : '（当前非你的回合）'}`"
                @click="isMyTurn && handleCardClick(card)"
                @keydown.enter.prevent="isMyTurn && handleCardClick(card)"
                @keydown.space.prevent="isMyTurn && handleCardClick(card)"
                :class="cn(
                  'uno-card card-mobile flex flex-col items-center justify-center cursor-pointer shrink-0 touch-feedback focus:outline-none focus-visible:ring-2 focus-visible:ring-cyan-400',
                  getDynamicCardClass(card),
                  selectedCard === card && 'ring-2 ring-blue-500 scale-105 z-10',
                  !isMyTurn && 'opacity-60 grayscale cursor-not-allowed'
                )"
                :style="{
                  transform: selectedCard === card ? (isMobile ? 'translateY(-10px)' : 'translateY(-12px)') : 'none'
                }"
              >
                <div class="absolute top-1 left-1 text-xs-mobile sm:text-[6px] font-black opacity-30 uppercase tracking-tighter">{{ ELEMENTS_DATA[card.type] ? 'Elem' : 'Spec' }}</div>
                <div class="flex flex-col items-center justify-center">
                  <div class="text-base sm:text-base font-black font-mono italic tracking-tighter leading-none">{{ card.type }}</div>
                  <div v-if="card.effect || ['He','Ne','Ar','Kr'].includes(card.type)" class="mt-1 px-1.5 sm:px-1 py-0.5 bg-black/10 rounded-md text-xs-mobile sm:text-[8px] font-black uppercase tracking-tighter">
                    {{ ['He','Ne','Ar','Kr'].includes(card.type) ? '转向' : card.effect === 'Au' ? '跳过' : card.effect === '+2' ? '+2' : card.effect === '+4' ? '+4' : card.effect }}
                  </div>
                  <div v-else-if="ELEMENTS_DATA[card.type]" class="text-xs-mobile sm:text-[8px] font-bold opacity-80 mt-0.5 uppercase tracking-tighter font-serif italic text-black/40">
                    {{ getSubstanceName(card.type) }}
                  </div>
                </div>
                <div class="absolute bottom-1 right-1 text-xs-mobile sm:text-[6px] font-mono opacity-40 uppercase tracking-tighter">
                  {{ card.effect ? 'Func' : 'Pass' }}
                </div>
              </div>
            </template>
            <div v-else class="flex flex-col items-center justify-center opacity-10 pb-3 sm:pb-4">
              <FlaskConical class="w-10 h-10 sm:w-12 sm:h-12 mb-1" />
              <p class="font-black uppercase tracking-widest text-xs-mobile">Inventory_Empty</p>
            </div>
           </div>
        </div>
      </div>

      <Teleport to="body">
      <!-- Experimental Victory / Failure Protocol -->
      <div v-if="showSettlementPanel" class="viewport-modal-overlay z-[100] p-4 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md transition-all duration-500">
        <!-- Cool Background Effects (Minimized for focus) -->
        <div class="absolute inset-0 pointer-events-none overflow-hidden opacity-30">
           <div v-for="i in 8" :key="i" 
                class="absolute w-px h-[200%] bg-gradient-to-t from-transparent via-blue-500/20 to-transparent animate-beam"
                :style="{ left: (i * 12) + '%', animationDelay: (i * 0.3) + 's', animationDuration: (3 + Math.random() * 2) + 's' }">
           </div>
        </div>

        <div class="relative w-full max-w-sm sm:max-w-md bg-white/95 dark:bg-[#0d0d10]/95 border border-slate-200 dark:border-white/10 rounded-[32px] shadow-3xl flex flex-col items-center text-center overflow-hidden animate-zoom-in p-6 sm:p-8 backdrop-blur-md max-h-[85vh]">
           <div class="absolute top-0 left-0 w-full h-1.5 bg-gradient-to-r from-blue-600 via-purple-600 to-emerald-600 animate-shimmer"></div>

           <div class="relative mb-4 group shrink-0">
              <div class="absolute -inset-8 bg-blue-500/10 rounded-full blur-[40px] animate-pulse"></div>
              
              <div class="relative w-16 h-16 sm:w-20 h-20 bg-gradient-to-br from-slate-800 to-black dark:from-blue-600 dark:to-blue-900 rounded-2xl sm:rounded-3xl flex items-center justify-center shadow-xl border border-white/20 transform hover:rotate-0 transition-transform duration-500">
                 <Trophy class="w-8 h-8 sm:w-10 h-10 text-yellow-400 drop-shadow-[0_0_10px_rgba(250,204,21,0.5)]" />
              </div>
           </div>

           <div class="flex-1 w-full overflow-y-auto custom-scrollbar px-1 mb-4">
             <div class="space-y-3 mb-4 w-full">
                <div class="inline-flex items-center gap-2 px-3 py-1 bg-blue-500/10 border border-blue-500/20 rounded-full">
                   <span class="text-[8px] font-black text-blue-600 dark:text-blue-400 uppercase tracking-widest font-mono">Mission_Complete</span>
                </div>

                <div class="px-2">
                  <template v-if="isTerminatedSettlement">
                    <h2 class="text-2xl sm:text-3xl font-black text-slate-900 dark:text-white tracking-tighter leading-tight mb-1">
                      实验异常中止
                    </h2>
                    <p class="text-[11px] text-slate-500 dark:text-slate-400 font-medium">
                      对局已结束并进入终止结算，以下为当前可用实验数据。
                    </p>
                  </template>
                  <template v-else-if="winner?.uid === user.uid">
                    <h2 class="text-3xl sm:text-4xl font-black text-transparent bg-clip-text bg-gradient-to-b from-slate-900 to-slate-600 dark:from-white dark:to-blue-200 tracking-tighter leading-tight mb-1">
                      实验大成功
                    </h2>
                    <p class="text-[11px] text-slate-500 dark:text-blue-400/60 font-medium">
                      恭喜首席研究员！量子反应核心已稳定。
                    </p>
                  </template>
                  <template v-else>
                    <h2 class="text-2xl sm:text-3xl font-black text-slate-900 dark:text-white tracking-tighter leading-tight mb-1">
                      反应已终止
                    </h2>
                    <p class="text-[11px] text-slate-500 dark:text-slate-400 font-medium">
                      本次实验由 <span class="font-black text-blue-600 dark:text-blue-400">{{ getPlayerDisplayName(winner) }}</span> 成功收官。
                    </p>
                  </template>
                </div>

                <div class="w-full mt-2 bg-slate-50/50 dark:bg-white/[0.02] border border-slate-200 dark:border-white/5 rounded-2xl p-3 shadow-inner relative overflow-hidden group/board">
                   <div class="flex items-center justify-between mb-2 px-1 relative z-10">
                      <span class="text-[9px] font-black uppercase tracking-widest text-slate-500">实验数据摘要</span>
                      <div class="flex items-center gap-1 px-1.5 py-0.5 bg-blue-500/10 rounded-md border border-blue-500/10">
                         <span class="text-[7px] font-black text-blue-500 font-mono">SYNC_OK</span>
                      </div>
                   </div>

                   <div class="space-y-1 relative z-10">
                      <div 
                        v-for="(item, index) in sortedPointsChanges" 
                        :key="item.uid"
                        class="flex items-center justify-between group/row p-2 rounded-xl transition-all hover:bg-white dark:hover:bg-white/5 animate-slide-in-bottom"
                        :style="{ animationDelay: (index * 0.05) + 's' }"
                        :class="item.uid === user.uid ? 'bg-blue-600/5 ring-1 ring-blue-500/10' : ''"
                      >
                         <div class="flex items-center gap-3">
                            <div class="relative">
                              <div class="w-6 h-6 rounded-lg flex items-center justify-center font-black text-[10px]" :class="[
                                index === 0 ? 'bg-yellow-400 text-yellow-900 shadow-lg' :
                                'bg-slate-100 dark:bg-black/40 text-slate-500'
                              ]">
                                {{ index + 1 }}
                              </div>
                            </div>
                            
                            <div class="flex flex-col items-start leading-tight">
                              <span class="text-xs font-black flex items-center gap-1.5" :class="shouldShowInBlue(item.player) ? 'text-blue-600 dark:text-blue-400' : 'text-slate-700 dark:text-slate-100'">
                                <span class="truncate max-w-[100px]">{{ getPlayerDisplayName(item.player) }}</span>
                                <span v-if="item.uid === user.uid" class="px-1 py-0.5 bg-blue-500 text-[6px] text-white rounded-md uppercase font-mono shrink-0">ME</span>
                              </span>
                            </div>
                         </div>
                         
                         <div class="flex items-center gap-1.5">
                            <div :class="cn(
                               'px-2 py-0.5 rounded-lg font-black font-mono text-[10px]',
                               item.points >= 0 ? 'text-emerald-500' : 'text-rose-500'
                            )">
                              {{ item.points >= 0 ? '+' : '' }}{{ item.points }}
                            </div>
                            <div v-if="gameState?.hint_bonus_applied?.[item.uid]" class="px-1.5 py-0.5 rounded-md font-black font-mono text-[8px] text-amber-500 bg-amber-500/10 border border-amber-500/20">
                              无提示
                            </div>
                            <div v-if="item.xp > 0" class="px-2 py-0.5 rounded-lg font-black font-mono text-[10px] text-blue-500">
                              +{{ item.xp }}XP
                            </div>
                         </div>
                      </div>
                   </div>
                </div>

                <!-- 反作弊检测结果（游戏结束后展示） -->
                <div v-if="hasCheatDetection" class="w-full mt-2 bg-rose-500/5 border border-rose-500/20 rounded-2xl p-3">
                  <div class="flex items-center justify-between mb-2 px-1">
                    <span class="text-[9px] font-black uppercase tracking-widest text-rose-500">反作弊检测报告</span>
                    <span class="px-1.5 py-0.5 bg-rose-500/10 rounded-md border border-rose-500/10 text-[7px] font-black text-rose-500 font-mono">FLAGGED</span>
                  </div>
                  <div class="space-y-1">
                    <div
                      v-for="(count, uid) in fastReactionUIDs"
                      :key="uid"
                      class="flex items-center justify-between p-2 rounded-xl bg-rose-500/5"
                    >
                      <div class="flex items-center gap-2">
                        <span class="text-[10px] font-black text-rose-600 dark:text-rose-400">
                          {{ getPlayerDisplayName(gameState?.players?.find((p: any) => String(p.uid) === String(uid))) || `研究员 ${uid}` }}
                        </span>
                      </div>
                      <div class="flex items-center gap-1">
                        <span class="px-2 py-0.5 rounded-lg bg-rose-500/10 text-rose-500 font-black font-mono text-[10px]">
                          快速出牌 x{{ count }}
                        </span>
                      </div>
                    </div>
                  </div>
                  <p class="text-[9px] text-rose-400/60 mt-2 px-1 font-medium">
                    系统已记录相关回放数据，将用于进一步审核。
                  </p>
                </div>

             </div>
           </div>

           <div class="w-full shrink-0 flex gap-3">
              <button
                @click="feedback.click(); router.push('/')"
                class="flex-1 h-12 bg-blue-600 hover:bg-blue-500 text-white font-black rounded-xl transition-all shadow-lg active:scale-95 flex items-center justify-center gap-2 group"
              >
                 <span class="uppercase tracking-widest text-xs">回到主页</span>
                 <ChevronRight class="w-4 h-4 group-hover:translate-x-1 transition-transform" />
              </button>
           </div>
        </div>
      </div>

      <div v-if="isReplayBridgeMode && replayGameOver" class="viewport-modal-overlay z-[110] p-4 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md">
        <div class="w-full max-w-xl bg-white dark:bg-[#121216] border border-slate-200 dark:border-white/10 rounded-[28px] shadow-2xl overflow-hidden">
          <div class="px-6 py-5 border-b border-slate-200 dark:border-white/10 bg-slate-50/70 dark:bg-white/[0.03]">
            <p class="text-[10px] font-black uppercase tracking-widest text-blue-500">Replay Result</p>
            <h3 class="text-2xl font-black text-slate-900 dark:text-white mt-1">游戏结束</h3>
            <p class="text-xs text-slate-500 mt-1">{{ replayEndType === 'game_terminated_invalid' ? '本次对局判定为无效结算' : '本次对局已完成回放' }}</p>
          </div>

          <div class="p-6 space-y-4">
            <div class="grid grid-cols-3 gap-2">
              <div class="rounded-xl border border-slate-200 dark:border-white/10 bg-slate-50 dark:bg-white/5 px-3 py-3 text-center">
                <p class="text-[10px] font-black uppercase tracking-widest text-slate-400">回合数</p>
                <p class="text-lg font-black text-slate-900 dark:text-white">{{ replaySummary.roundCount }}</p>
              </div>
              <div class="rounded-xl border border-slate-200 dark:border-white/10 bg-slate-50 dark:bg-white/5 px-3 py-3 text-center">
                <p class="text-[10px] font-black uppercase tracking-widest text-slate-400">出牌总和</p>
                <p class="text-lg font-black text-slate-900 dark:text-white">{{ replaySummary.totalPlays }}</p>
              </div>
              <div class="rounded-xl border border-slate-200 dark:border-white/10 bg-slate-50 dark:bg-white/5 px-3 py-3 text-center">
                <p class="text-[10px] font-black uppercase tracking-widest text-slate-400">事件总数</p>
                <p class="text-lg font-black text-slate-900 dark:text-white">{{ replayEvents.length }}</p>
              </div>
            </div>

            <div class="rounded-xl border border-slate-200 dark:border-white/10 bg-slate-50 dark:bg-white/5 p-3">
              <p class="text-[10px] font-black uppercase tracking-widest text-slate-400 mb-2">各卡牌数目</p>
              <div v-if="replayCardCountEntries.length > 0" class="max-h-40 overflow-y-auto space-y-1 pr-1">
                <div
                  v-for="entry in replayCardCountEntries"
                  :key="`replay-card-stat-${entry[0]}`"
                  class="flex items-center justify-between rounded-lg border border-slate-200 dark:border-white/10 bg-white dark:bg-black/20 px-3 py-2"
                >
                  <span class="text-xs font-semibold text-slate-700 dark:text-slate-200">{{ entry[0] }}</span>
                  <span class="text-xs font-black text-blue-600 dark:text-blue-300">{{ entry[1] }}</span>
                </div>
              </div>
              <div v-else class="text-xs text-slate-400">无可统计的出牌记录</div>
            </div>

            <div class="flex items-center gap-3 pt-1">
              <button
                @click="restartReplayPlayback()"
                class="flex-1 h-11 rounded-xl bg-blue-600 hover:bg-blue-500 text-white font-black uppercase tracking-widest text-xs transition-all"
              >
                重新播放
              </button>
              <button
                @click="navigateBackFromReplay"
                class="flex-1 h-11 rounded-xl border border-slate-300 dark:border-white/10 text-slate-700 dark:text-slate-200 font-black uppercase tracking-widest text-xs hover:bg-slate-100 dark:hover:bg-white/10 transition-all"
              >
                返回时间线
              </button>
            </div>
          </div>
        </div>
      </div>
      </Teleport>
    </template>

    <Teleport to="body">
    <!-- Admin Management Modal -->
    <div v-if="showAdminModal" class="viewport-modal-overlay z-[100] p-4 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md">
      <div class="absolute inset-0 bg-black/80 backdrop-blur-md" @click="feedback.click(); showAdminModal = false"></div>
      <div class="viewport-modal-panel modal-mobile relative w-full max-w-lg bg-white dark:bg-[#121216] border border-slate-200 dark:border-white/10 rounded-[40px] shadow-2xl overflow-y-auto overflow-x-hidden animate-in zoom-in duration-300">
        <div class="p-8 border-b border-slate-200 dark:border-white/5 bg-slate-50/50 dark:bg-white/[0.02]">
          <div class="flex items-center justify-between mb-2">
            <h3 class="text-2xl font-black text-slate-900 dark:text-white tracking-tighter flex items-center gap-3">
              <ShieldAlert class="w-6 h-6 text-red-500" />
              权限执行控制
            </h3>
            <button @click="feedback.click(); showAdminModal = false" class="p-2 hover:bg-slate-200 dark:hover:bg-white/5 rounded-full transition-colors">
              <X class="w-6 h-6 text-slate-400" />
            </button>
          </div>
          <p class="text-[10px] text-slate-500 font-mono uppercase tracking-[0.2em]">Target: {{ adminTargetUser?.nickname || adminTargetUser?.username }} (UID: {{ adminTargetUser?.uid }})</p>
        </div>

        <div class="p-8 space-y-8">
          <div class="grid grid-cols-2 gap-4">
            <button 
              @click="adminActionType = 'kick'; banReason = '你由于违规游戏而被踢出'"
              :class="cn(
                'flex flex-col items-center gap-3 p-6 rounded-3xl border transition-all group',
                adminActionType === 'kick' ? 'bg-amber-500/10 border-amber-500/50 text-amber-500' : 'bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/10 text-slate-500'
              )"
            >
              <UserMinus class="w-8 h-8 group-hover:scale-110 transition-transform" />
              <span class="text-xs font-black uppercase tracking-widest">驱逐出场</span>
            </button>
            <button 
              @click="adminActionType = 'ban'; banReason = '你由于违规游戏而被封禁'"
              :class="cn(
                'flex flex-col items-center gap-3 p-6 rounded-3xl border transition-all group',
                adminActionType === 'ban' ? 'bg-red-500/10 border-red-500/50 text-red-500' : 'bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/10 text-slate-500'
              )"
            >
              <Ban class="w-8 h-8 group-hover:scale-110 transition-transform" />
              <span class="text-xs font-black uppercase tracking-widest">限制访问</span>
            </button>
          </div>

          <div v-if="adminActionType === 'ban'" class="space-y-4 animate-in slide-in-from-top-4 duration-300">
            <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest block">封禁时长</label>
            <div class="grid grid-cols-4 gap-2">
              <button
                v-for="preset in banPresets"
                :key="preset.hours"
                @click="setBanDuration(preset.hours)"
                :class="cn(
                  'px-2 py-2 rounded-xl text-[10px] font-black uppercase tracking-wider transition-all border active:scale-95',
                  selectedBanPreset === preset.hours
                    ? 'bg-red-500/10 border-red-500/30 text-red-500 shadow-sm'
                    : 'bg-slate-50 dark:bg-black/20 border-slate-200 dark:border-white/10 text-slate-500 hover:border-red-500/20 hover:text-red-400'
                )"
              >
                {{ preset.label }}
              </button>
              <button
                @click="selectedBanPreset = null"
                :class="cn(
                  'px-2 py-2 rounded-xl text-[10px] font-black uppercase tracking-wider transition-all border active:scale-95',
                  selectedBanPreset === null
                    ? 'bg-red-500/10 border-red-500/30 text-red-500 shadow-sm'
                    : 'bg-slate-50 dark:bg-black/20 border-slate-200 dark:border-white/10 text-slate-500 hover:border-red-500/20 hover:text-red-400'
                )"
              >
                自定义
              </button>
            </div>
            <div v-if="selectedBanPreset === null" class="animate-in slide-in-from-top-2 duration-200">
              <input
                v-model="banUntil"
                type="datetime-local"
                :min="formatDatetimeLocal(new Date())"
                @focus="handleInputFocus"
                @blur="handleInputBlur"
                class="w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 rounded-2xl px-5 py-3 text-sm font-bold text-slate-700 dark:text-white focus:outline-none focus:border-red-500/50 transition-all"
              />
            </div>
            <div class="flex items-center gap-2 ml-1 mt-1">
              <div class="w-1.5 h-1.5 rounded-full" :class="banUntil ? 'bg-red-500 animate-pulse' : 'bg-slate-300 dark:bg-slate-600'"></div>
              <span class="text-[9px] font-bold text-slate-400 dark:text-slate-600 uppercase tracking-wider">
                截止: {{ banUntil ? new Date(banUntil).toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' }) + '（UTC+8）' : '未设置' }}
              </span>
            </div>
          </div>

          <div class="space-y-4">
            <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest block">操作事由</label>
            <div class="relative group">
              <div class="absolute inset-0 bg-red-500/5 rounded-2xl blur-lg group-focus-within:bg-red-500/10 transition-all"></div>
              <textarea
                v-model="banReason"
                placeholder="请输入详细的违规事由..."
                @focus="handleInputFocus"
                @blur="handleInputBlur"
                class="relative w-full bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 rounded-2xl px-6 py-4 text-sm font-medium text-slate-700 dark:text-white focus:outline-none focus:border-red-500/50 min-h-[100px] transition-all"
              ></textarea>
            </div>
          </div>

          <button 
            @click="handleAdminAction"
            :class="cn(
              'w-full h-16 rounded-[24px] font-black uppercase tracking-[0.2em] text-xs transition-all shadow-xl active:scale-95',
              adminActionType === 'kick' ? 'bg-amber-500 hover:bg-amber-400 text-white shadow-amber-500/20' : 'bg-red-600 hover:bg-red-500 text-white shadow-red-500/20'
            )"
          >
            确认执行操作
          </button>
        </div>
      </div>
    </div>
    </Teleport>

    <!-- Invite Friends Modal -->
    <Teleport to="body">
    <div v-if="showInviteFriendsModal" class="viewport-modal-overlay z-[100] p-4 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md">
      <div class="absolute inset-0 bg-black/80 backdrop-blur-md clickable" @click="showInviteFriendsModal = false"></div>
      <div class="relative w-full max-w-lg bg-white dark:bg-[#121216] border border-slate-200 dark:border-white/10 rounded-[40px] shadow-2xl overflow-hidden animate-in zoom-in duration-300">
        <div class="p-8 border-b border-slate-200 dark:border-white/5 bg-slate-50/50 dark:bg-white/[0.02]">
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-2xl font-black text-slate-900 dark:text-white tracking-tighter flex items-center gap-3">
                <UserPlus class="w-6 h-6 text-blue-500" />
                邀请好友加入
              </h3>
              <p class="text-[10px] text-slate-500 font-mono uppercase tracking-[0.2em] mt-2">选择一位好友发送游戏邀请</p>
            </div>
            <button @click="showInviteFriendsModal = false" class="p-2 hover:bg-slate-200 dark:hover:bg-white/5 rounded-full transition-colors">
              <X class="w-6 h-6 text-slate-400" />
            </button>
          </div>
        </div>

        <div class="p-8 max-h-[500px] overflow-y-auto custom-scrollbar">
          <div v-if="friendsList.length === 0" class="flex flex-col items-center justify-center py-16 opacity-20 grayscale">
            <Users class="w-16 h-16 mb-4" />
            <p class="text-sm font-black uppercase tracking-[0.2em]">暂无好友</p>
            <p class="text-[10px] mt-2 italic font-medium uppercase">请先添加好友后再邀请</p>
          </div>
          <div v-else class="space-y-3">
            <button
              v-for="friend in friendsList"
              :key="friend.uid"
              @click="sendGameInvite(friend)"
              class="w-full p-4 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-[24px] flex items-center justify-between hover:border-blue-500/30 hover:bg-blue-500/5 transition-all group"
            >
              <div class="flex items-center gap-4">
                <div class="relative">
                  <div class="w-12 h-12 rounded-xl bg-white dark:bg-white/10 flex items-center justify-center text-2xl border border-slate-200 dark:border-white/10 shadow-sm overflow-hidden">
                    <UserAvatar :avatar="friend.avatar" />
                  </div>
                  <div v-if="friend.is_online" class="absolute -bottom-1 -right-1 w-3.5 h-3.5 bg-emerald-500 border-3 border-white dark:border-[#121216] rounded-full shadow-lg shadow-emerald-500/20"></div>
                </div>
                <div class="text-left">
                  <div class="text-base font-bold flex items-center gap-2" :class="friend.remark ? 'text-blue-600 dark:text-blue-400' : 'text-slate-700 dark:text-white'">
                    {{ friend.remark || friend.nickname || friend.username }}
                    <span v-if="friend.is_online" class="px-2 py-0.5 bg-emerald-500/10 text-emerald-500 text-[8px] font-black rounded uppercase tracking-widest">Online</span>
                  </div>
                  <div class="text-[9px] text-slate-400 font-mono mt-1">UID: {{ friend.uid }}</div>
                </div>
              </div>
              <div class="w-8 h-8 rounded-lg bg-blue-500/10 flex items-center justify-center group-hover:bg-blue-500 transition-all">
                <Send class="w-4 h-4 text-blue-500 group-hover:text-white group-hover:translate-x-0.5 transition-all" />
              </div>
            </button>
          </div>
        </div>
      </div>
    </div>
    </Teleport>

    <!-- Players Floating Panel -->
    <div
      v-if="showPlayers"
      class="fixed right-0 top-0 bottom-0 w-[85%] sm:w-80 z-[110] bg-white dark:bg-slate-900 border-l lg:border border-slate-200 dark:border-white/10 lg:rounded-[40px] lg:top-6 lg:bottom-52 lg:right-6 shadow-2xl flex flex-col overflow-hidden"
    >
      <div class="px-5 py-4 border-b border-slate-200 dark:border-white/10 flex items-center justify-between sticky top-0 z-20 bg-inherit pb-6 lg:pb-4">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-2xl bg-blue-500/10 flex items-center justify-center border border-blue-500/20">
            <Users class="w-5 h-5 text-blue-500" />
          </div>
          <div>
            <h3 class="text-sm font-black uppercase tracking-widest text-slate-800 dark:text-white leading-none">研究员列表</h3>
            <div class="flex items-center gap-2 mt-2">
               <span class="text-[9px] font-mono font-bold text-slate-400 bg-slate-100 dark:bg-white/5 px-1.5 py-0.5 rounded">
                  POS: {{ allPlayers.length }}/{{ roomInfo?.max_players }}
               </span>
               <div class="w-1 h-1 rounded-full bg-emerald-500 animate-pulse"></div>
            </div>
          </div>
        </div>
        <button
          @click="showPlayers = false"
          class="p-2.5 bg-slate-50 dark:bg-white/5 hover:bg-slate-100 dark:hover:bg-white/10 rounded-2xl transition-all text-slate-400 hover:text-slate-900 dark:hover:text-white active:scale-90"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <div ref="playersContainer" class="flex-1 overflow-y-auto p-4 sm:p-3 custom-scrollbar space-y-2.5 sm:space-y-2 pb-24 lg:pb-4">
        <template v-if="allPlayers.length > 0">
          <div
            v-for="(player, index) in allPlayers"
            :key="player.uid || index"
            data-player-card
            @click="isReplayBridgeMode && handleReplayPerspectiveSwitch(player)"
            :class="cn(
              'flex items-center gap-2.5 sm:gap-2 p-2.5 sm:p-2 rounded-2xl border transition-all duration-300 relative overflow-hidden',
              gameState?.current_player === index
                ? 'bg-blue-600 border-blue-400 shadow-xl shadow-blue-500/20 active-player-card'
                : (gameState ? 'bg-slate-50 dark:bg-white/[0.03] border-slate-200 dark:border-white/5' : 'bg-slate-50 dark:bg-white/[0.03] border-slate-200 dark:border-white/10'),
              isReplayBridgeMode && 'cursor-pointer hover:border-cyan-400/50',
              isReplayBridgeMode && replayPerspectiveUID === Number(player.uid) && 'ring-2 ring-cyan-400/70 border-cyan-400/70 bg-cyan-500/10 dark:bg-cyan-500/15',
              drawAnimatingUIDs.has(player.uid) && 'ring-2 ring-rose-500 border-rose-500',
              gameState?.finished_players?.includes(player.uid) && 'opacity-60 grayscale-[0.8] bg-slate-200/50 dark:bg-black/40 border-slate-300/30'
            )"
          >
            <!-- Background Decoration for active player -->
            <div v-if="gameState?.current_player === index" class="absolute -right-4 -top-4 w-16 h-16 bg-white/10 rounded-full blur-xl"></div>

            <!-- Finished Badge -->
            <div v-if="gameState?.finished_players?.includes(player.uid)" class="absolute top-0 right-0 z-20">
               <div class="bg-emerald-500 text-white text-[6px] font-black px-1.5 py-0.5 rounded-bl-lg shadow-sm animate-in fade-in zoom-in duration-500 uppercase tracking-tighter">
                  Finished
               </div>
            </div>

            <div class="relative w-9 h-9 sm:w-8 sm:h-8 shrink-0">
               <div :class="cn(
                 'w-full h-full rounded-xl flex items-center justify-center border transition-all duration-300 shadow-sm overflow-hidden relative',
                 gameState?.current_player === index ? 'bg-white border-white scale-105 shadow-blue-500/20 shadow-lg' : 'bg-white dark:bg-slate-800 border-slate-100 dark:border-white/10'
               )">
                 <div class="w-full h-full flex items-center justify-center pointer-events-none">
                    <UserAvatar :avatar="player.avatar" />
                 </div>

                 <!-- Offline Overlay -->
                 <div v-if="player.is_offline" class="absolute inset-0 bg-rose-500/60 flex items-center justify-center backdrop-blur-sm z-10">
                   <Activity class="w-4 h-4 text-white animate-pulse" />
                 </div>
               </div>
               
               <!-- Action Progress -->
               <div v-if="gameState" class="absolute -bottom-1 -right-1 flex gap-0.5 z-20">
                 <div 
                   v-for="i in 2" 
                   :key="i" 
                   :class="cn(
                     'w-1.5 h-1.5 rounded-full border border-black/10 shadow-sm transition-all duration-300', 
                     i <= (player.action_progress || 0) 
                       ? (gameState?.current_player === index ? 'bg-blue-400' : 'bg-blue-500') 
                       : 'bg-slate-200 dark:bg-slate-700'
                   )"
                 ></div>
               </div>
            </div>

            <div class="flex-1 min-w-0 flex flex-col justify-center">
               <div class="flex items-center justify-between h-4">
                  <span class="text-xs sm:text-[11px] font-black truncate tracking-tight uppercase" :class="[
                     gameState?.current_player === index ? 'text-white' : (shouldShowInBlue(player) ? 'text-blue-600 dark:text-blue-400' : 'text-slate-800 dark:text-slate-200')
                  ]">{{ getPlayerDisplayName(player) }}</span>
                  <div class="flex items-center gap-1">
                     <Zap v-if="player.double_action_available" :class="cn('w-3 h-3 fill-current', gameState?.current_player === index ? 'text-amber-300' : 'text-amber-500')" />
                     <template v-if="gameState">
                        <div class="flex items-center gap-1 px-1 py-0.5 rounded-lg" :class="gameState?.current_player === index ? 'bg-white/20 text-white' : 'bg-slate-900/5 dark:bg-white/5 text-slate-500'">
                           <span class="text-[9px] font-black font-mono leading-none">{{ player.card_count || 0 }}</span>
                        </div>
                     </template>
                  </div>
               </div>
               <div class="flex items-center justify-between mt-0.5 h-3.5">
                  <div class="flex items-center gap-1">
                     <span class="text-[7px] font-mono opacity-40 shrink-0 uppercase tracking-tighter" :class="gameState?.current_player === index ? 'text-white/80' : 'text-slate-500'">UID: {{ player.uid }}</span>
                    <span v-if="isReplayBridgeMode && replayPerspectiveUID === Number(player.uid)" class="text-[6px] font-black uppercase text-cyan-700 dark:text-cyan-300 px-1 py-0.5 bg-cyan-500/20 rounded-md border border-cyan-500/30 leading-none">POV</span>
                     <span v-if="player.is_offline" class="text-[6px] font-black uppercase text-rose-100 px-1 py-0.5 bg-rose-500/80 rounded-md border border-rose-500/20 leading-none">OFF</span>
                     <template v-if="!gameState">
                        <div :class="cn('px-1 py-0.5 rounded-md border text-[6px] font-black uppercase tracking-widest transition-all leading-none', 
                           player.is_ready 
                           ? 'text-emerald-500 bg-emerald-500/10 border-emerald-500/20' 
                           : 'text-slate-400 bg-slate-100 dark:bg-white/5 border-slate-200 dark:border-white/10')">
                           {{ player.is_ready ? 'RDY' : 'WT' }}
                        </div>
                     </template>
                  </div>
                  
                  <div class="flex items-center gap-0.5">
                     <button v-if="Number(player.uid) !== Number(user.uid) && !isFriend(player.uid) && !isReplayBridgeMode"
                             @click.stop="handleAddFriend(player)"
                             :class="cn('p-0.5 rounded-md transition-all active:scale-90', gameState?.current_player === index ? 'bg-white/20 text-white' : 'bg-slate-200/50 dark:bg-white/5 text-slate-500')"
                     >
                       <UserPlus class="w-2.5 h-2.5" />
                     </button>
                     <button v-if="Number(player.uid) !== Number(user.uid) && (!isReplayBridgeMode || Number(player.uid) > 0)"
                             @click.stop="handleReportPlayer(player)"
                             :class="cn('p-0.5 rounded-md transition-all active:scale-90', gameState?.current_player === index ? 'bg-white/20 text-white' : 'bg-slate-200/50 dark:bg-white/5 text-slate-500')"
                     >
                       <Flag class="w-2.5 h-2.5" />
                     </button>
                     <button v-if="user.role === 'admin' && Number(player.uid) !== Number(user.uid) && !isReplayBridgeMode"
                             @click.stop="openAdminAction(player)"
                             :class="cn('p-0.5 rounded-md transition-all active:scale-90', gameState?.current_player === index ? 'bg-white/20 text-white' : 'bg-rose-500/20 text-rose-500')"
                     >
                       <ShieldAlert class="w-2.5 h-2.5" />
                     </button>
                  </div>
               </div>
            </div>
          </div>

          <!-- Empty Slots -->
          <div
            v-for="i in (roomInfo?.max_players || 0) - allPlayers.length"
            :key="'empty-slot-' + i"
            class="flex items-center gap-2 px-3 py-2.5 rounded-xl border border-dashed border-slate-200 dark:border-white/5 opacity-30"
          >
            <div class="w-8 h-8 rounded-lg border border-dashed border-slate-300 dark:border-white/10 flex items-center justify-center">
              <Plus class="w-3.5 h-3.5 text-slate-400" />
            </div>
            <span class="text-[10px] font-black uppercase tracking-tighter text-slate-400">EMPTY_SLOT</span>
          </div>
        </template>
        <div v-else class="flex flex-col items-center justify-center py-8 opacity-30">
          <Loader2 class="w-5 h-5 animate-spin mb-2" />
          <span class="text-[10px] font-black uppercase tracking-widest italic">Awaiting Peers...</span>
        </div>
      </div>
    </div>

    <!-- Mobile Overlay for Players Panel -->
    <div
      v-if="showPlayers"
      class="fixed inset-0 bg-white/10 dark:bg-black/20 backdrop-blur-[2px] z-[95] lg:hidden clickable"
      @click="showPlayers = false"
    ></div>

    <!-- Chat Floating Sidebar -->
    <div
      v-if="showChat && !isReplayBridgeMode"
      class="fixed right-0 top-0 bottom-0 w-full lg:w-80 z-[100] lg:top-6 lg:bottom-52 lg:right-6 flex flex-col"
    >
      <ChatBox
        :roomId="id"
        title="实验内通信线程"
        maxHeight="100%"
        class="h-full !bg-white/95 dark:!bg-slate-900/60 backdrop-blur-3xl shadow-3xl lg:rounded-[40px] border-l lg:border border-slate-200 dark:border-white/10"
        @close="showChat = false"
        @input-focus="handleInputFocus"
        @input-blur="handleInputBlur"
      />
    </div>

    <!-- Mobile Overlay for Chat -->
    <div
      v-if="showChat && !isReplayBridgeMode"
      class="fixed inset-0 bg-white/10 dark:bg-black/20 backdrop-blur-[2px] z-[95] lg:hidden clickable"
      @click="showChat = false"
    ></div>

    <!-- 化学键盘 -->
    <ChemicalKeyboard
      v-if="showChemicalKeyboard && !isReplayBridgeMode"
      v-model="substanceInput"
      :deckCards="roomInfo?.deck_config?.cards || {}"
      :myHand="myData?.hand_cards || []"
      @confirm="handleKeyboardConfirm"
      @close="showChemicalKeyboard = false"
    />

    <!-- 牌组详情查看模态框 -->
    <Teleport to="body">
    <div v-if="showDeckDetailModal && roomInfo?.deck_config" class="viewport-modal-overlay z-[200] p-4 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md">
      <div class="absolute inset-0 bg-slate-900/40 dark:bg-black/80 backdrop-blur-md clickable" @click="showDeckDetailModal = false" />
      <div class="relative w-full max-w-2xl bg-white dark:bg-[#121216] border border-slate-200 dark:border-white/10 rounded-[32px] shadow-2xl overflow-hidden">
         <div class="px-5 py-4 border-b border-slate-100 dark:border-white/5 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 bg-emerald-500/10 border border-emerald-500/20 rounded-xl flex items-center justify-center text-emerald-500">
                <FlaskConical class="w-4 h-4" />
              </div>
              <div>
                <h2 class="text-base font-black text-slate-800 dark:text-white tracking-tight leading-none">{{ roomInfo.deck_config.name }}</h2>
                <p class="text-[8px] text-slate-400 dark:text-slate-500 font-mono uppercase tracking-widest mt-1">Deck_Configuration</p>
              </div>
            </div>
            <button @click="showDeckDetailModal = false" class="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl transition-colors text-slate-400 hover:text-slate-900 dark:hover:text-white">
              <X class="w-4 h-4" />
            </button>
         </div>
         <div class="p-5 max-h-[60vh] overflow-y-auto custom-scrollbar space-y-4">
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-2">
              <div class="p-3 bg-slate-50 dark:bg-white/5 rounded-xl border border-slate-200 dark:border-white/10">
                <p class="text-[8px] text-slate-400 mb-1 uppercase tracking-wider">牌组名称</p>
                <p class="text-[11px] font-black text-slate-900 dark:text-white">{{ roomInfo.deck_config.name }}</p>
              </div>
              <div class="p-3 bg-slate-50 dark:bg-white/5 rounded-xl border border-slate-200 dark:border-white/10">
                <p class="text-[8px] text-slate-400 mb-1 uppercase tracking-wider">元素种类</p>
                <p class="text-[11px] font-black text-blue-600 dark:text-blue-400">{{ Object.keys(roomInfo.deck_config.cards || {}).length }} 种</p>
              </div>
              <div class="p-3 bg-slate-50 dark:bg-white/5 rounded-xl border border-slate-200 dark:border-white/10">
                <p class="text-[8px] text-slate-400 mb-1 uppercase tracking-wider">总卡牌数</p>
                <p class="text-[11px] font-black text-slate-900 dark:text-white">{{ (Object.values(roomInfo.deck_config.cards || {}) as number[]).reduce((a, b) => a + b, 0) }} 张</p>
              </div>
              <div class="p-3 bg-slate-50 dark:bg-white/5 rounded-xl border border-slate-200 dark:border-white/10">
                <p class="text-[8px] text-slate-400 mb-1 uppercase tracking-wider">起始手牌</p>
                <p class="text-[11px] font-black text-slate-900 dark:text-white">{{ roomInfo.deck_config.initial_cards || 10 }} 张</p>
              </div>
            </div>
            <div class="p-4 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl">
              <div class="flex items-center justify-between mb-3">
                <span class="text-[9px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">卡牌配置</span>
                <span class="text-[8px] text-blue-500/40 font-mono">CARD_LIST</span>
              </div>
              <div class="grid grid-cols-2 sm:grid-cols-3 gap-2 max-h-64 overflow-y-auto custom-scrollbar pr-1">
                <div
                  v-for="(count, formula) in roomInfo.deck_config.cards"
                  :key="formula"
                  class="p-2.5 bg-white dark:bg-black/20 rounded-lg border border-slate-200 dark:border-white/10"
                >
                  <div class="flex items-center justify-between">
                    <span class="text-[10px] font-black text-slate-900 dark:text-white font-mono" v-html="String(formula).replace(/(\d+)/g, '<sub>$1</sub>')"></span>
                    <span class="text-[9px] font-black text-blue-600 dark:text-blue-400">×{{ count }}</span>
                  </div>
                </div>
              </div>
            </div>
         </div>
         <div class="px-5 py-3 border-t border-slate-100 dark:border-white/5 flex justify-end">
            <button @click="showDeckDetailModal = false" class="px-4 py-2 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-700 dark:text-slate-300 font-bold rounded-xl transition-all uppercase tracking-widest text-[10px] border border-slate-200 dark:border-white/5">
              关闭
            </button>
         </div>
      </div>
    </div>
    </Teleport>
  </div>

  <!-- Level Up Animation -->
  <LevelUpAnimation ref="levelUpAnimationRef" />

  <!-- Game Toast -->
  <GameToast ref="gameToastRef" />

  <!-- Feedback Settings - 仅在准备阶段显示 -->
  <FeedbackSettings v-if="roomInfo?.status === 'waiting'" />
</template>

<style scoped src="./GameRoom.css"></style>
