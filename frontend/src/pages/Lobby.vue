<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { pageClassNames } from '@lib'
import PhlogistonIcon from '../components/icons/PhlogistonIcon.vue'
import { gameAPI, authAPI, commonAPI, friendAPI, adminAPI, clearAuthState } from '../utils/api'
import { useDialog } from '../utils/dialog'
import UserAvatar from '../components/UserAvatar.vue'
import websocket from '../utils/websocket'
import { Beaker, Plus, Shield, ShieldAlert, LogOut, Settings, Play, X, Loader2, Database, MessageCircle, Megaphone, Menu, Puzzle, FileText, ChevronRight } from 'lucide-vue-next'
import { cn } from '../utils/cn'
import ChatBox from '../components/ChatBox.vue'
import TutorialGuide from '../components/TutorialGuide.vue'
import PingDisplay from '../components/PingDisplay.vue'
import TeamPanel from '../components/TeamPanel.vue'
import '../styles/lobby.css'

const props = defineProps<{
  // user props can be added if we pass from App.vue
}>()

const router = useRouter()
const { showAlert, showConfirm } = useDialog()
const user = ref<any>({})
const loadUserInfo = async () => {
  try {
    const res = await authAPI.getUserInfo()
    user.value = res.data
    localStorage.setItem('user', JSON.stringify(res.data))
    if (res.data?.role !== 'admin') {
      lobbyViewMode.value = 'player'
    }
  } catch (e) {
    console.error('Failed to load user info:', e)
  }
}

const isAdminUser = computed(() => user.value?.role === 'admin')
const isAdminView = computed(() => isAdminUser.value && lobbyViewMode.value === 'admin')
const anticheatPanelPath = computed(() =>
  user.value?.role === 'admin'
    ? '/admin/anticheat'
    : '/appeals'
)
const anticheatPanelTitle = computed(() =>
  user.value?.role === 'admin'
    ? '反作弊管理'
    : '申诉中心'
)

try {
  const userData = JSON.parse(localStorage.getItem('user') || '{}')
  // 兼容旧版本的 id 字段
  if (userData.id && !userData.uid) {
    userData.uid = userData.id
  }
  user.value = userData
} catch (e) {
  console.error('Failed to parse user in Lobby:', e)
}

const friendsList = ref<any[]>([])

const loadFriends = async () => {
  try {
    const res = await friendAPI.getFriends()
    friendsList.value = res.data
  } catch (err) {
    console.error(err)
  }
}

const rooms = ref<any[]>([])
const decks = ref<any[]>([])
const pendingFeedbacks = ref<any[]>([])
const unreadChatCount = ref(0)
const persistentAnnouncements = ref<any[]>([])
const showCreateModal = ref(false)
const showAIArenaModal = ref(false)
const showDeckDetailModal = ref(false)
const showAccessKeyModal = ref(false)
const createdRoomInfo = ref<any>(null)
const selectedDeckConfig = ref<any>(null)
const roomName = ref('')
const maxPlayers = ref(4)
const deckID = ref(0)
const isPointsMode = ref(true)
const isPrivate = ref(false)
const lobbyViewMode = ref<'player' | 'admin'>('player')

// 等级匹配与排位
const isRanked = ref(false)
const levelRange = ref(5)

const pveDifficulty = ref(50)
const aiCount = ref(1)
const customAccessKey = ref('') // 自定义访问密钥

// AI补位功能配置
const enableAIBackfill = ref(false)
const aiBackfillDifficulty = ref(50)

const loading = ref(false)
const activeSurveys = ref<any[]>([])
const showSurveyModal = ref(false)
const currentSurvey = ref<any>(null)
const showLegalModal = ref(false)
const legalModalTitle = ref('')
const legalModalContent = ref('')
const pendingCompensationAppeal = ref<any>(null)
const showCompensationReminder = ref(false)

const normalizeApiList = (payload: any, key: string) => {
  if (Array.isArray(payload?.[key])) return payload[key]
  if (Array.isArray(payload?.data)) return payload.data
  return []
}

const isPendingCompensationAppeal = (appeal: any) => {
  const status = String(appeal?.status || '').toLowerCase()
  const compensationStatus = String(appeal?.compensation_status || '').toLowerCase()
  return status === 'approved' && Number(appeal?.compensation_amount || 0) > 0 && compensationStatus !== 'ok'
}

const compensationReminderMutedKey = (appeal: any) => {
  const uid = user.value?.uid || user.value?.id || 'anonymous'
  const appealId = appeal?.id || appeal?.ID || 'latest'
  return `chemistryuno-compensation-reminder-muted-${uid}-${appealId}`
}

const loadCompensationReminder = async () => {
  try {
    const res = await authAPI.getPlayerAppeals()
    const pending = normalizeApiList(res.data, 'appeals').find(isPendingCompensationAppeal)
    pendingCompensationAppeal.value = pending || null
    if (pending && localStorage.getItem(compensationReminderMutedKey(pending)) !== 'true') {
      showCompensationReminder.value = true
    }
  } catch (err) {
    console.error('加载补偿领取提醒失败:', err)
  }
}

const goToCompensationClaim = () => {
  showCompensationReminder.value = false
  router.push('/appeals')
}

const remindCompensationLater = () => {
  showCompensationReminder.value = false
}

const muteCompensationReminder = () => {
  if (pendingCompensationAppeal.value) {
    localStorage.setItem(compensationReminderMutedKey(pendingCompensationAppeal.value), 'true')
  }
  showCompensationReminder.value = false
}

const openUserAgreement = async () => {
  try {
    const res = await fetch('/USER_AGREEMENT.md')
    legalModalContent.value = await res.text()
  } catch (e) {
    legalModalContent.value = '无法加载协议内容，请检查实验室连接。'
  }
  legalModalTitle.value = '用户协议 / User Agreement'
  showLegalModal.value = true
}

const openPrivacyPolicy = async () => {
  try {
    const res = await fetch('/PRIVACY_POLICY.md')
    legalModalContent.value = await res.text()
  } catch (e) {
    legalModalContent.value = '无法加载政策内容，请检查实验室连接。'
  }
  legalModalTitle.value = '隐私政策 / Privacy Policy'
  showLegalModal.value = true
}

const loadSurveys = async () => {
  try {
    const res = await authAPI.getActiveSurveys()
    const surveys = res.data || []
    activeSurveys.value = surveys
    
    if (surveys.length > 0) {
      const survey = surveys[0]
      // 检查是否已经弹出过此问卷（防止大厅定时刷新导致弹窗反复跳出）
      const dismissedKey = `survey_dismissed_${survey.id || survey.ID}`
      const isDismissedInSession = sessionStorage.getItem(dismissedKey)
      
      if (!isDismissedInSession) {
        currentSurvey.value = {
          ...survey,
          id: survey.id || survey.ID
        }
        showSurveyModal.value = true
      }
    }
  } catch (err) {
    console.error('加载调查问卷失败:', err)
  }
}

const handleCloseSurveyModal = () => {
  if (currentSurvey.value) {
    const sId = currentSurvey.value.id || currentSurvey.value.ID
    sessionStorage.setItem(`survey_dismissed_${sId}`, 'true')
  }
  showSurveyModal.value = false
}

const handleDismissSurvey = async () => {
  if (!currentSurvey.value) return
  try {
    const sId = currentSurvey.value.id || currentSurvey.value.ID
    await authAPI.dismissSurvey(sId)
    sessionStorage.setItem(`survey_dismissed_${sId}`, 'true')
    showSurveyModal.value = false
    // 检查是否还有其他问卷
    const nextSurveys = activeSurveys.value.filter(s => (s.id || s.ID) !== sId)
    if (nextSurveys.length > 0) {
      setTimeout(() => {
        activeSurveys.value = nextSurveys
        const nextS = nextSurveys[0]
        currentSurvey.value = {
          ...nextS,
          id: nextS.id || nextS.ID
        }
        showSurveyModal.value = true
      }, 500)
    }
  } catch (err) {
    console.error('忽略问卷失败:', err)
  }
}
const currentTime = ref(new Date())
const roomsFetchedAt = ref(Date.now())
const onlineCount = ref(0)
const lastTimeUpdate = ref(Date.now())

const getRoomLiveCountdown = (room: any): number => {
  if (!room.countdown || room.countdown <= 0) return 0
  // 检查是否需要更新 currentTime 以保持倒计时活跃
  if (Date.now() - lastTimeUpdate.value > 1000) {
    currentTime.value = new Date()
    lastTimeUpdate.value = Date.now()
  }
  const elapsed = Math.floor((currentTime.value.getTime() - roomsFetchedAt.value) / 1000)
  return Math.max(0, room.countdown - elapsed)
}

const isRoomFull = (room: any): boolean => {
  return (room.players?.length || 0) >= room.max_players
}

const annTypeMap: Record<string, string> = {
  emergency: '紧急',
  maintenance: '维护',
  info: '公告'
}
const isMobileMenuOpen = ref(false)
const appVersion = ref('V1.2.1 Mendeleef')

// 大厅新手指引
const showTutorial = ref(false)
const LOBBY_TUTORIAL_COMPLETED_KEY = 'chemistry-uno-lobby-tutorial-completed'
const LOBBY_TUTORIAL_SKIPPED_KEY = 'chemistry-uno-lobby-tutorial-skipped'

// 导航步骤同时匹配移动端/桌面端，组件内部会优先选择可见元素
const navigationSelector = '[data-tutorial="mobile-menu"], [data-tutorial="desktop-nav"]'

const lobbyTutorialSteps = computed(() => [
  { id: 'welcome-lobby', titlePlaceholder: '欢迎来到化学UNO大厅', contentPlaceholder: '这里是所有玩家的集结地！你可以创建房间邀请朋友，或加入其他玩家的对局，开启一场融合化学知识与策略的卡牌对决。', position: 'center' as const },
  { id: 'create-room', titlePlaceholder: '创建你的实验室', contentPlaceholder: '点击这里可以新建游戏房间，设置房间名称、对战人数和游戏模式。还可以开启「AI补位」，让AI填满空位，随时开始对战，不用等人！', targetSelector: '[data-tutorial="create-room"]', position: 'bottom' as const },
  { id: 'room-list', titlePlaceholder: '寻找对战房间', contentPlaceholder: '这里列出了当前所有开放的游戏房间。找到感兴趣的房间直接点击加入，与来自各地的化学玩家展开激烈对决！', targetSelector: '[data-tutorial="room-list"]', position: 'top' as const },
  { id: 'navigation', titlePlaceholder: '快捷导航', contentPlaceholder: '通过导航栏可以快速切换到排行榜、好友列表等页面。随时掌握自己的全球排名，查看对战战绩，探索更多功能！', targetSelector: navigationSelector, position: 'bottom' as const },
  { id: 'user-profile', titlePlaceholder: '你的化学家档案', contentPlaceholder: '点击这里进入个人主页，查看你的等级、燃素、历史战绩和成就勋章。努力提升实力，向化学大师的称号冲击！', targetSelector: '[data-tutorial="user-chip"]', position: 'bottom' as const },
  { id: 'ai-arena', titlePlaceholder: 'AI竞技场', contentPlaceholder: '想练习牌技或享受单人挑战？在AI竞技场中与不同难度的AI对战，从初级到专家逐步进阶，随时磨练你的化学牌技！', targetSelector: '[data-tutorial="ai-arena"]', position: 'bottom' as const },
  { id: 'complete-lobby', titlePlaceholder: '开始你的化学之旅！', contentPlaceholder: '你已经了解了大厅的全部功能！现在就创建或加入一个房间，与其他玩家一较高下吧。愿化学元素的力量与你同在，旗开得胜！', position: 'center' as const }
])

const checkFirstTimeLobby = () => {
  const hasSeenLobbyTutorial = localStorage.getItem(LOBBY_TUTORIAL_COMPLETED_KEY) === 'true'
  const hasSkippedLobbyTutorial = localStorage.getItem(LOBBY_TUTORIAL_SKIPPED_KEY) === 'true'
  if (!hasSeenLobbyTutorial && !hasSkippedLobbyTutorial) {
    setTimeout(() => showTutorial.value = true, 1500)
  }
}

const handleTutorialComplete = async () => {
  localStorage.setItem(LOBBY_TUTORIAL_COMPLETED_KEY, 'true')
  localStorage.removeItem(LOBBY_TUTORIAL_SKIPPED_KEY)
  showTutorial.value = false

  // 自动创建教学关卡
  await createTutorialMatch()
}

const handleTutorialClose = () => {
  // 玩家选择跳过后，后续不再自动弹出教程
  localStorage.setItem(LOBBY_TUTORIAL_COMPLETED_KEY, 'true')
  localStorage.setItem(LOBBY_TUTORIAL_SKIPPED_KEY, 'true')
  showTutorial.value = false
}

// 创建教学关卡
const createTutorialMatch = async () => {
  loading.value = true
  try {
    const response = await gameAPI.createRoom(
      '教学: 首战AI',
      2, // 1 (Human) + 1 (AI)
      deckID.value,
      false, // 非积分模式
      true, // 私密房间
      undefined,
      true, // PvE模式
      20, // 低难度AI (20/100)
      1, // 1个AI
      false, // 不启用AI补位
      0,
      false, // 非排位模式
      0,
      true // ⭐ 启用脚本化教学
    )
    const room = response.data

    // 标记为教学模式
    localStorage.setItem('chemistry-uno-tutorial-mode', 'true')

    router.push(`/room/${room.id}`)
  } catch (error: any) {
    showAlert(error.response?.data?.error || '创建教学关卡失败', '系统异常')
  } finally {
    loading.value = false
  }
}

// 获取当前域名用于生成分享链接
const currentOrigin = window.location.origin

const activeRoom = computed(() => {
  return rooms.value.find(r => 
    (r.status === 'playing' || r.status === 'waiting') && 
    r.players && 
    r.players.includes(Number(user.value.uid))
  )
})

const totalUnreadCount = computed(() => {
  return unreadChatCount.value + pendingFeedbacks.value.length
})

const loadDecks = async () => {
  try {
    const res = await gameAPI.getMyDecks()
    const allDecks = res.data || []
    // 排序：全局优先
    allDecks.sort((a: any, b: any) => {
      if (a.is_global && !b.is_global) return -1
      if (!a.is_global && b.is_global) return 1
      return 0
    })
    decks.value = allDecks
    if (decks.value.length > 0) {
      // 默认选择全局牌组，如果没有则选择第一个
      const globalDeck = decks.value.find((d: any) => d.is_global)
      deckID.value = globalDeck ? globalDeck.id : decks.value[0].id
    }
  } catch (e) {
    console.error(e)
  }
}

let roomInterval: any
let timeInterval: any

const handleOnlineCountUpdate = (msg: any) => {
  onlineCount.value = msg.data || 0
}

const handleSystemAnnouncement = (msg: any) => {
  const ann = msg.data
  if (ann && ann.is_persistent) {
    const exists = persistentAnnouncements.value.some(a => a.id === ann.id)
    if (!exists) {
      persistentAnnouncements.value.unshift(ann)
    } else {
      const idx = persistentAnnouncements.value.findIndex(a => a.id === ann.id)
      persistentAnnouncements.value[idx] = ann
    }
  }
}

// WebSocket房间更新事件处理（后续支持）
const handleRoomsUpdate = (msg: any) => {
  if (isAdminView.value) {
    loadRooms()
    return
  }
  if (msg.data && Array.isArray(msg.data)) {
    rooms.value = msg.data
    roomsFetchedAt.value = Date.now()
  }
}

const handleChatUnreadUpdate = (msg: any) => {
  if (typeof msg.count === 'number') {
    unreadChatCount.value = msg.count
  }
}

// 监控所有弹窗状态以禁用/启用背景滚动
watch(
  [showCreateModal, showAIArenaModal, showDeckDetailModal, showAccessKeyModal, showSurveyModal, showLegalModal, showCompensationReminder],
  ([create, ai, deck, access, survey, legal, compensationReminder]) => {
    const hasModal = create || ai || deck || access || survey || legal || compensationReminder
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
  loadUserInfo()
  loadRooms()
  loadDecks()
  loadPendingFeedbacks()
  loadPersistentAnnouncements()
  loadFriends()
  loadVersion()
  loadSurveys() // 加载问卷调查
  loadCompensationReminder()
  websocket.connect()
  websocket.on('online_count', handleOnlineCountUpdate)
  websocket.on('system_announcement', handleSystemAnnouncement)
  websocket.on('rooms_update', handleRoomsUpdate)
  websocket.on('chat_unread_count', handleChatUnreadUpdate)

  // 使用WebSocket实时推送房间列表，不再使用轮询
  // roomInterval = setInterval(loadRooms, 10000)  // 已移除轮询机制
  
  // 移除不必要的1秒时间更新，改为按需更新
  // timeInterval = setInterval(() => {
  //   currentTime.value = new Date()
  // }, 1000)

  // 检查大厅新手指引
  checkFirstTimeLobby()

  // 大厅控制台指令
  if (typeof window !== 'undefined') {
    const win = window as any
    win.showLobbyTutorial = () => { showTutorial.value = true; console.log('✨ 大厅新手指引已启动') }
    win.resetLobbyTutorial = () => {
      localStorage.removeItem(LOBBY_TUTORIAL_COMPLETED_KEY)
      localStorage.removeItem(LOBBY_TUTORIAL_SKIPPED_KEY)
      console.log('🔄 大厅教程已重置')
    }
    win.checkLobbyTutorial = () => {
      const completed = localStorage.getItem(LOBBY_TUTORIAL_COMPLETED_KEY) === 'true'
      const skipped = localStorage.getItem(LOBBY_TUTORIAL_SKIPPED_KEY) === 'true'
      console.log('📊 大厅教程:', completed ? '已完成' : '未完成', '| 跳过:', skipped ? '是' : '否', '| 显示:', showTutorial.value)
      return { completed, skipped, showing: showTutorial.value }
    }
    console.log('%c🏠 Lobby Console Commands', 'color: #06b6d4; font-weight: bold')
    console.log('  showLobbyTutorial() | resetLobbyTutorial() | checkLobbyTutorial()')
  }
})

onUnmounted(() => {
  if (roomInterval) clearInterval(roomInterval)
  if (timeInterval) clearInterval(timeInterval)
  websocket.off('online_count', handleOnlineCountUpdate)
  websocket.off('system_announcement', handleSystemAnnouncement)
  websocket.off('rooms_update', handleRoomsUpdate)
  websocket.off('chat_unread_count', handleChatUnreadUpdate)
})

const loadRooms = async () => {
  try {
    const response = isAdminView.value ? await adminAPI.getActiveRooms() : await gameAPI.getRooms()
    rooms.value = response.data || []
    roomsFetchedAt.value = Date.now()
  } catch (error) {
    console.error('加载房间列表失败:', error)
  }
}

watch(isAdminView, () => {
  loadRooms()
})

const loadPendingFeedbacks = async () => {
  try {
    const res = await authAPI.getMyFeedbacks()
    pendingFeedbacks.value = (res.data || []).filter((f: any) => f.status === 'unread')
  } catch (e) {
    console.error(e)
  }
}

const loadVersion = async () => {
  try {
    const res = await authAPI.getVersion()
    if (res.data && res.data.fullVersion) {
      appVersion.value = res.data.fullVersion
    }
  } catch (e) {
    console.error('获取版本信息失败:', e)
  }
}

const loadPersistentAnnouncements = async () => {
  try {
    const res = await commonAPI.getAnnouncements()
    persistentAnnouncements.value = (res.data || []).filter((a: any) => a.is_persistent)
  } catch (e) {
    console.error(e)
  }
}

// 查看牌组配置详情
const handleViewDeckConfig = (deckConfig: any) => {
  if (!deckConfig) return
  selectedDeckConfig.value = deckConfig
  showDeckDetailModal.value = true
}

const handleCreateRoom = async () => {
  loading.value = true

  try {
    const response = await gameAPI.createRoom(
      roomName.value,
      maxPlayers.value,
      deckID.value,
      isPointsMode.value,
      isPrivate.value,
      customAccessKey.value || undefined,
      false, // isPvE
      0, // pveDifficulty
      0, // aiCount
      enableAIBackfill.value, // 启用AI补位
      aiBackfillDifficulty.value, // AI补位难度
      isRanked.value,
      levelRange.value,
      false // tutorialScript（普通房间非教学模式）
    )
    const room = response.data
    // 重置状态
    showCreateModal.value = false
    roomName.value = ''
    customAccessKey.value = ''
    enableAIBackfill.value = false
    aiBackfillDifficulty.value = 50
    isRanked.value = false
    levelRange.value = 5

    // 如果是私密房间且有访问密钥，显示密钥模态框
    if (isPrivate.value && room.access_key) {
      createdRoomInfo.value = {
        id: room.id,
        access_key: room.access_key,
        name: room.name
      }
      showAccessKeyModal.value = true
      isPrivate.value = false
    } else {
      isPrivate.value = false
      router.push(`/room/${room.id}`)
    }
  } catch (error: any) {
    showAlert(error.response?.data?.error || '创建房间失败', '系统异常')
  } finally {
    loading.value = false
  }
}

const handleCreateAIRoom = async () => {
  loading.value = true
  try {
    const response = await gameAPI.createRoom(
      roomName.value || `AI挑战 · ${pveDifficulty.value}`,
      1 + aiCount.value, // MaxPlayers = 1 (Human) + AI Count
      deckID.value,
      isPointsMode.value,
      true, // AI 房间强制私密/不公开
      undefined,
      true, // IsPvE
      pveDifficulty.value,
      aiCount.value,
      false, // PvE模式不需要补位
      0, // AI补位难度（PvE模式忽略）
      false, // PvE 不设排位
      0,
      false // tutorialScript（AI竞技场非教学模式）
    )
    const room = response.data
    router.push(`/room/${room.id}`)
  } catch (error: any) {
    showAlert(error.response?.data?.error || '创建AI对战失败', '系统异常')
  } finally {
    loading.value = false
  }
}

const handleJoinRoom = async (roomId: string, asSpectator: boolean = false) => {
  try {
    await gameAPI.joinRoom(roomId, undefined, asSpectator)
    const url = asSpectator ? `/room/${roomId}?spectator=true` : `/room/${roomId}`
    router.push(url)
  } catch (error: any) {
    if (error.response?.data?.error === '房间不存在') {
      const index = rooms.value.findIndex(r => r.id === roomId)
      if (index !== -1) {
        rooms.value.splice(index, 1)
      }
      showAlert('该房间已失效并自动清理', '提示')
    } else {
      showAlert(error.response?.data?.error || '加入房间失败', '连接错误')
    }
  }
}

const handleLeaveRoom = async (roomId: string) => {
  const ok = await showConfirm('确定要中止并退出当前的实验吗？如果你正在游戏中，系统将按逃跑处理。')
  if (!ok) return

  try {
    await gameAPI.leaveRoom(roomId)
    loadRooms()
    showAlert('已成功从核心节点撤离。', '实验中止')
  } catch (error: any) {
    if (error.response?.data?.error === '房间不存在') {
      loadRooms()
      showAlert('房间已不存在，正在刷新状态。', '实验中止')
    } else {
      showAlert(error.response?.data?.error || '退出失败', '链路故障')
    }
  }
}

const handleTerminateRoom = async (roomId: string) => {
  const ok = await showConfirm('确定要强制结束这个房间吗？房间内所有玩家都会被移除。')
  if (!ok) return

  try {
    await adminAPI.terminateRoom(roomId)
    loadRooms()
    showAlert('已强制结束该房间。', '操作成功')
  } catch (error: any) {
    if (error.response?.data?.error === '房间不存在') {
      loadRooms()
      showAlert('房间已不存在，正在刷新状态。', '操作成功')
    } else {
      showAlert(error.response?.data?.error || '结束房间失败', '链路故障')
    }
  }
}

const handleLogout = () => {
  // Token已存储在HttpOnly Cookie中，浏览器会自动处理
  clearAuthState()
  websocket.disconnect()
  router.push('/login')
}

const activeNodesCount = computed(() => rooms.value.filter(r => r.status === 'playing').length)

const copyToClipboard = (text: string) => {
  if (window.navigator && window.navigator.clipboard) {
    window.navigator.clipboard.writeText(text)
    showAlert('复制成功', '成功')
  } else {
    showAlert('当前环境不支持剪贴板操作', '错误')
  }
}
</script>

<template>
  <div class="lobby-page">
    <!-- Background Decor -->
    <div class="lobby-bg-decor">
      <div class="lobby-bg-decor-blob-1"></div>
      <div class="lobby-bg-decor-blob-2"></div>
      <div class="lobby-bg-decor-pattern"></div>
    </div>

    <!-- Main Layout Layer -->
    <div :class="pageClassNames.lobbyContent">
      
      <!-- Top Command Bar - 移动端优化 -->
      <header class="lobby-header">
        <div class="lobby-header-container">
          <div class="flex items-center gap-3 sm:gap-4">
            <div class="lobby-logo-bundle">
              <Beaker class="w-5 h-5 sm:w-6 sm:h-6 text-blue-500 group-hover:rotate-12 transition-transform" />
              <div>
                 <h1 class="lobby-logo-title">CHEMISTRY <span class="text-blue-500">UNO</span></h1>
                 <p class="lobby-logo-subtitle">{{ appVersion }}</p>
              </div>
            </div>

            <!-- Ping显示（移动端） -->
            <PingDisplay class="lg:hidden" />

            <!-- Status Indicators (Desktop) -->
            <div class="hidden lg:flex items-center gap-4 text-xs-mobile font-mono tracking-widest text-slate-500 border-l border-slate-200 dark:border-white/10 pl-6 uppercase">
              <div class="flex items-center gap-2">
                <div class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></div>
                稳定运行
              </div>
              <div class="flex items-center gap-2">
                 运行时间: {{ currentTime.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) }}
              </div>
              <!-- Ping显示 -->
              <PingDisplay />
            </div>
          </div>

          <div class="flex items-center gap-2 sm:gap-3">
            <!-- User Identity Chip -->
            <div data-tutorial="user-chip" @click="router.push('/profile')" class="user-identity-chip">
               <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-slate-200 to-slate-300 dark:from-slate-700 dark:to-slate-800 flex items-center justify-center text-base shadow-inner group-hover:scale-105 transition-transform overflow-hidden border border-slate-200 dark:border-white/5">
                 <UserAvatar :avatar="user.avatar" />
               </div>
               <div class="hidden sm:flex flex-col">
                 <span class="text-xs-mobile font-black text-slate-900 dark:text-white">{{ user.nickname || user.username }}</span>
                 <span class="text-[9px] sm:text-[8px] text-slate-500 font-mono uppercase">
                   {{ user.role === 'admin' ? '管理员' : '研究员' }}
                 </span>
               </div>
            </div>

            <!-- Desktop Navigation -->
            <div data-tutorial="desktop-nav" class="hidden lg:flex items-center gap-1.5">
              <router-link
                to="/ranking"
                class="flex items-center gap-1.5 px-3 py-1.5 bg-amber-500/5 hover:bg-amber-500/10 border border-amber-500/20 hover:border-amber-500/40 rounded-xl transition-all group/phlogiston"
                title="燃素排行榜"
              >
                <PhlogistonIcon :size="16" class="!text-amber-500 group-hover/phlogiston:scale-110 transition-transform shrink-0" />
                <span class="text-[10px] font-black uppercase tracking-widest hidden md:block text-amber-600/80 group-hover/phlogiston:text-amber-500 transition-colors house-style">排位</span>
              </router-link>
              <div class="w-px h-4 bg-slate-200 dark:bg-white/10 mx-0.5"></div>
              <router-link to="/feedbacks" class="lobby-nav-link lobby-nav-link-amber relative" title="反馈与公告">
                <div v-if="pendingFeedbacks.length > 0" class="absolute -top-1 -right-1 min-w-[14px] h-[14px] px-0.5 bg-amber-500 text-white text-[8px] font-black rounded-full flex items-center justify-center border border-white dark:border-slate-900 z-10">
                  {{ pendingFeedbacks.length }}
                </div>
                <Megaphone class="w-4 h-4" />
              </router-link>
              <router-link to="/profile" class="lobby-nav-link" title="个人主页">
                <Settings class="w-4 h-4" />
              </router-link>
              <router-link to="/data" class="lobby-nav-link lobby-nav-link-blue" title="数据库">
                <Database class="w-4 h-4" />
              </router-link>
              <router-link to="/chat" class="lobby-nav-link lobby-nav-link-indigo relative" title="公共频道">
                <div v-if="unreadChatCount > 0" class="absolute -top-1 -right-1 min-w-[14px] h-[14px] px-0.5 bg-red-500 text-white text-[8px] font-black rounded-full flex items-center justify-center border border-white dark:border-slate-900 z-10 animate-pulse">
                  {{ unreadChatCount > 99 ? '99+' : unreadChatCount }}
                </div>
                <MessageCircle class="w-4 h-4" />
              </router-link>
              <router-link to="/plugins" class="lobby-nav-link lobby-nav-link-purple" title="插件市场">
                <Puzzle class="w-4 h-4" />
              </router-link>
              <router-link v-if="user.role === 'admin' || user.role === 'co-worker'" to="/admin" class="lobby-nav-link lobby-nav-link-amber" title="管理面板">
                <Shield class="w-4 h-4 text-yellow-500" />
              </router-link>
              <router-link :to="anticheatPanelPath" class="lobby-nav-link lobby-nav-link-red" :title="anticheatPanelTitle">
                <ShieldAlert class="w-4 h-4" />
              </router-link>
              <div class="w-px h-5 bg-white/10 mx-1"></div>
              <button @click="handleLogout" class="lobby-nav-link lobby-nav-link-red" title="退出登录">
                <LogOut class="w-4 h-4" />
              </button>
            </div>

            <!-- Mobile Menu Toggle -->
            <button
              @click="isMobileMenuOpen = !isMobileMenuOpen"
              data-tutorial="mobile-menu"
              class="lg:hidden p-2.5 bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl text-slate-500 touch-feedback transition-all relative"
            >
              <div v-if="totalUnreadCount > 0" class="absolute -top-1.5 -right-1.5 min-w-[16px] h-[16px] px-1 bg-red-500 text-white text-[9px] font-black rounded-full flex items-center justify-center border-2 border-white dark:border-slate-900 z-10 shadow-sm animate-pulse-subtle">
                {{ totalUnreadCount > 99 ? '99+' : totalUnreadCount }}
              </div>
              <Menu v-if="!isMobileMenuOpen" class="w-5 h-5" />
              <X v-else class="w-5 h-5" />
            </button>
          </div>
        </div>
      </header>

      <!-- Mobile Menu Overlay -->
      <transition
        enter-active-class="transition duration-200 ease-out"
        enter-from-class="opacity-0 -translate-y-4"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition duration-150 ease-in"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 -translate-y-4"
      >
        <div v-if="isMobileMenuOpen" class="lg:hidden fixed inset-0 z-[45] pt-20 bg-white/95 dark:bg-slate-900/95 backdrop-blur-lg">
          <div class="px-4 py-3 space-y-4 overflow-y-auto" style="max-height: calc(var(--app-height) - 80px);">
            <!-- 顶部装饰与统计 -->
            <div class="flex items-center justify-between p-3 bg-slate-50 dark:bg-white/5 rounded-2xl border border-slate-200 dark:border-white/10">
               <div class="flex items-center gap-2">
                  <div class="w-10 h-10 rounded-lg bg-slate-200 dark:bg-slate-800 shadow-sm flex items-center justify-center border border-slate-100 dark:border-white/5 overflow-hidden">
                     <UserAvatar :avatar="user.avatar" />
                  </div>
                  <div>
                     <h4 class="text-xs font-black text-slate-800 dark:text-white leading-tight uppercase">{{ user.nickname || user.username }}</h4>
                     <p class="text-[8px] font-mono text-slate-400 mt-0.5 uppercase tracking-widest">ID: {{ user.uid }}</p>
                  </div>
               </div>
               <div class="text-right">
                  <div class="flex items-center justify-end gap-1 text-amber-500">
                     <PhlogistonIcon :size="12" color="#f59e0b" />
                     <span class="text-[10px] font-black font-mono">{{ Math.floor(user.points || 0) }}</span>
                  </div>
                  <p class="text-[7px] font-bold text-slate-400 uppercase mt-0.5">燃素</p>
               </div>
            </div>

            <div class="grid grid-cols-2 gap-2">
              <router-link @click="isMobileMenuOpen = false" to="/ranking" class="flex flex-col items-center justify-center p-3 bg-slate-50 dark:bg-white/5 rounded-2xl border border-slate-200 dark:border-white/10 shadow-sm active:scale-95 transition-all">
                <div class="w-10 h-10 bg-amber-500/10 rounded-xl flex items-center justify-center mb-2 text-amber-500">
                  <PhlogistonIcon :size="20" color="#f59e0b" />
                </div>
                <span class="text-[9px] font-black uppercase tracking-widest text-slate-600 dark:text-slate-300 text-center leading-tight">排位榜单</span>
              </router-link>

              <router-link @click="isMobileMenuOpen = false" to="/profile" class="flex flex-col items-center justify-center p-3 bg-slate-50 dark:bg-white/5 rounded-2xl border border-slate-200 dark:border-white/10 shadow-sm active:scale-95 transition-all">
                <div class="w-10 h-10 bg-blue-500/10 rounded-xl flex items-center justify-center mb-2 text-blue-500">
                  <Settings class="w-5 h-5" />
                </div>
                <span class="text-[9px] font-black uppercase tracking-widest text-slate-600 dark:text-slate-300 text-center leading-tight">个人中心</span>
              </router-link>

              <router-link @click="isMobileMenuOpen = false" to="/data" class="flex flex-col items-center justify-center p-3 bg-slate-50 dark:bg-white/5 rounded-2xl border border-slate-200 dark:border-white/10 shadow-sm active:scale-95 transition-all">
                <div class="w-10 h-10 bg-cyan-500/10 rounded-xl flex items-center justify-center mb-2 text-cyan-500">
                  <Database class="w-5 h-5" />
                </div>
                <span class="text-[9px] font-black uppercase tracking-widest text-slate-600 dark:text-slate-300 text-center leading-tight">物质百科</span>
              </router-link>

              <router-link @click="isMobileMenuOpen = false" to="/chat" class="flex flex-col items-center justify-center p-3 bg-slate-50 dark:bg-white/5 rounded-2xl border border-slate-200 dark:border-white/10 shadow-sm active:scale-95 transition-all relative">
                <div v-if="unreadChatCount > 0" class="absolute top-2 right-2 min-w-[18px] h-[18px] px-1 bg-red-500 text-white text-[10px] font-black rounded-full flex items-center justify-center border-2 border-white dark:border-slate-900 z-10 animate-pulse">
                  {{ unreadChatCount > 99 ? '99+' : unreadChatCount }}
                </div>
                <div class="w-10 h-10 bg-indigo-500/10 rounded-xl flex items-center justify-center mb-2 text-indigo-500">
                  <MessageCircle class="w-5 h-5" />
                </div>
                <span class="text-[9px] font-black uppercase tracking-widest text-slate-600 dark:text-slate-300 text-center leading-tight">公共频道</span>
              </router-link>

              <router-link @click="isMobileMenuOpen = false" to="/feedbacks" class="flex flex-col items-center justify-center p-3 bg-slate-50 dark:bg-white/5 rounded-2xl border border-slate-200 dark:border-white/10 shadow-sm active:scale-95 transition-all relative">
                <div v-if="pendingFeedbacks.length > 0" class="absolute top-2 right-2 min-w-[18px] h-[18px] px-1 bg-amber-500 text-white text-[10px] font-black rounded-full flex items-center justify-center border-2 border-white dark:border-slate-900 z-10">
                  {{ pendingFeedbacks.length }}
                </div>
                <div class="w-10 h-10 bg-emerald-500/10 rounded-xl flex items-center justify-center mb-2 text-emerald-500">
                  <Megaphone class="w-5 h-5" />
                </div>
                <span class="text-[9px] font-black uppercase tracking-widest text-slate-600 dark:text-slate-300 text-center leading-tight">反馈公告</span>
              </router-link>

              <router-link @click="isMobileMenuOpen = false" to="/plugins" class="flex flex-col items-center justify-center p-3 bg-slate-50 dark:bg-white/5 rounded-2xl border border-slate-200 dark:border-white/10 shadow-sm active:scale-95 transition-all group">
                <div class="w-10 h-10 bg-blue-600/10 rounded-xl flex items-center justify-center mb-2 text-blue-600 group-hover:bg-blue-600 group-hover:text-white transition-all">
                  <Puzzle class="w-5 h-5" />
                </div>
                <span class="text-[9px] font-black uppercase tracking-widest text-slate-600 dark:text-slate-300 text-center leading-tight">插件扩展市场</span>
              </router-link>

              <router-link @click="isMobileMenuOpen = false" :to="anticheatPanelPath" class="flex flex-col items-center justify-center p-3 bg-slate-50 dark:bg-white/5 rounded-2xl border border-slate-200 dark:border-white/10 shadow-sm active:scale-95 transition-all">
                <div class="w-10 h-10 bg-rose-500/10 rounded-xl flex items-center justify-center mb-2 text-rose-500">
                  <ShieldAlert class="w-5 h-5" />
                </div>
                <span class="text-[9px] font-black uppercase tracking-widest text-slate-600 dark:text-slate-300 text-center leading-tight">{{ anticheatPanelTitle }}</span>
              </router-link>

              <router-link v-if="user.role === 'admin' || user.role === 'co-worker'" @click="isMobileMenuOpen = false" to="/admin" class="flex flex-col items-center justify-center p-3 bg-slate-50 dark:bg-white/5 rounded-2xl border border-slate-200 dark:border-white/10 shadow-sm active:scale-95 transition-all col-span-2">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 bg-rose-500/10 rounded-lg flex items-center justify-center text-rose-500 shrink-0">
                    <Shield class="w-4 h-4" />
                  </div>
                  <span class="text-[9px] font-black uppercase tracking-[0.2em] text-slate-600 dark:text-slate-300">系统核心管理面板</span>
                </div>
              </router-link>
            </div>
            
            <div class="pt-2">
               <button 
                 @click="handleLogout" 
                 class="w-full flex items-center justify-between px-4 py-3 bg-slate-50 dark:bg-white/5 hover:bg-red-500/10 text-slate-400 hover:text-red-500 rounded-2xl border border-slate-200 dark:border-white/10 transition-all font-black uppercase tracking-widest text-[9px] group"
               >
                 <div class="flex items-center gap-3">
                    <LogOut class="w-4 h-4" />
                    停止实验并离线
                 </div>
                 <ChevronRight class="w-4 h-4 group-hover:translate-x-1 transition-transform" />
               </button>
            </div>
          </div>
        </div>
      </transition>

      <main class="flex-1 max-w-[1400px] mx-auto w-full px-4 sm:px-5 py-4 flex flex-col min-h-0">
        <!-- Welcome & Global Actions -->
        <div class="hub-header-section">
          <div class="hub-title-group">
            <div class="hub-status-badge">
              <span class="w-1 h-1 bg-blue-500 rounded-full animate-ping"></span>
              <span class="text-[8px] font-black text-blue-500 uppercase tracking-widest">试验场大厅</span>
            </div>
            <h2 class="hub-title">试验场枢纽</h2>
            <p class="text-[10px] text-slate-500 font-medium max-w-md leading-none">当前有 <span class="text-blue-500 font-black">{{ onlineCount }}</span> 名研究员在线进行博弈。</p>
          </div>

          <div class="flex items-center gap-3">
              <div class="hidden xl:flex items-center gap-3 px-3 py-2 bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/5 rounded-xl">
               <div class="text-center min-w-[60px]">
                 <p class="text-[7px] text-slate-400 uppercase font-black tracking-widest mb-0.5">在线人数</p>
                 <p class="text-sm font-black text-slate-900 dark:text-white font-mono leading-none">{{ onlineCount }}</p>
               </div>
               <div class="w-px h-4 bg-slate-200 dark:bg-white/5"></div>
               <div class="text-center min-w-[60px]">
                 <p class="text-[7px] text-slate-400 uppercase font-black tracking-widest mb-0.5">活跃房间</p>
                 <p class="text-sm font-black text-blue-600 dark:text-blue-400 font-mono leading-none">{{ activeNodesCount }}</p>
               </div>
             </div>

            <button
              @click="showCreateModal = true"
              data-tutorial="create-room"
              class="btn-action-primary"
            >
              <Plus class="w-3.5 h-3.5 group-hover:rotate-90 transition-transform duration-500" />
              <span class="uppercase tracking-widest text-[9px]">开启实验</span>
            </button>

            <button
              @click="showAIArenaModal = true; isPointsMode = true; isPrivate = true"
              data-tutorial="ai-arena"
              class="btn-action-secondary ml-2"
            >
              <div class="relative">
                <Beaker class="w-3.5 h-3.5" />
                <div class="absolute -top-1 -right-1 w-2 h-2 bg-yellow-400 rounded-full animate-pulse border-2 border-purple-600"></div>
              </div>
              <span class="uppercase tracking-widest text-[9px]">AI 竞技场</span>
            </button>
          </div>
        </div>

        <!-- Main Layout Grid -->
        <div class="lobby-main-grid">
          <!-- Left Column: Notifications & Room List -->
          <div class="lobby-content-pane">
            <!-- Persistent Announcements -->
            <div v-if="persistentAnnouncements.length > 0" class="space-y-3 animate-in fade-in duration-700">
               <div v-for="ann in persistentAnnouncements" :key="ann.id" 
                    :class="cn(
                      'relative overflow-hidden p-5 rounded-2xl border transition-all hover:shadow-md',
                      ann.type === 'emergency' ? 'bg-red-500/5 border-red-500/20 shadow-red-500/5' : 
                      ann.type === 'maintenance' ? 'bg-amber-500/5 border-amber-500/20 shadow-amber-500/5' : 
                      'bg-blue-500/5 border-blue-500/20 shadow-blue-500/5'
                    )">
                  <div class="relative z-10 flex flex-col md:flex-row md:items-center gap-4">
                    <div :class="cn(
                       'w-10 h-10 rounded-xl flex items-center justify-center shrink-0 border',
                       ann.type === 'emergency' ? 'bg-red-500/10 text-red-500 border-red-500/20' : 
                       ann.type === 'maintenance' ? 'bg-amber-500/10 text-amber-500 border-amber-500/20' : 
                       'bg-blue-500/10 text-blue-500 border-blue-500/20'
                    )">
                      <Megaphone class="w-5 h-5" />
                    </div>
                    <div class="flex-1">
                       <div class="flex items-center gap-2 mb-0.5">
                          <span :class="cn(
                            'text-[9px] font-black uppercase tracking-widest',
                            ann.type === 'emergency' ? 'text-red-500' : 
                            ann.type === 'maintenance' ? 'text-amber-500' : 
                            'text-blue-500'
                          )">
                          {{ annTypeMap[ann.type] || ann.type }}
                          </span>
                       </div>
                       <h3 class="text-base font-black text-slate-900 dark:text-white mb-1" v-if="ann.title">{{ ann.title }}</h3>
                       <p class="text-xs font-medium text-slate-600 dark:text-slate-400 leading-relaxed">{{ ann.content }}</p>
                    </div>
                  </div>
               </div>
            </div>

            <!-- Rejoin Banner -->
            <div v-if="activeRoom" class="rejoin-banner">
               <div class="flex items-center gap-5">
                  <div class="w-12 h-12 bg-white/20 rounded-xl flex items-center justify-center animate-[pulse_2s_infinite]">
                     <Beaker class="w-6 h-6 text-white" />
                  </div>
                  <div class="flex flex-col">
                     <span class="text-[9px] font-black uppercase text-blue-200 tracking-widest leading-none mb-1">
                        {{ activeRoom.status === 'waiting' && getRoomLiveCountdown(activeRoom) > 0 ? `实验启动中: ${getRoomLiveCountdown(activeRoom)}秒` : '活跃实验中' }}
                     </span>
                     <h3 class="text-lg font-black text-white uppercase tracking-wider">{{ activeRoom.name }}</h3>
                  </div>
               </div>
               <div class="flex items-center gap-3">
                  <button 
                    @click="handleLeaveRoom(activeRoom.id)"
                    class="px-5 py-3 bg-red-500 text-white rounded-xl text-[11px] font-black uppercase tracking-widest border border-red-500/20 transition-all hover:scale-105 active:scale-95 flex items-center gap-2"
                  >
                    <X class="w-4 h-4" />
                    结束
                  </button>
                  <button 
                    @click="handleJoinRoom(activeRoom.id)"
                    class="px-8 py-3 bg-white text-blue-600 rounded-xl text-[11px] font-black uppercase tracking-[0.2em] shadow-lg transition-all hover:scale-105 active:scale-95 flex items-center gap-2"
                  >
                    <Play class="w-4 h-4" />
                    重连
                  </button>
               </div>
            </div>

            <div v-if="isAdminUser" class="mb-3 flex items-center justify-end">
              <div class="inline-flex items-center gap-1 p-1 rounded-xl border border-slate-200 dark:border-white/10 bg-white/70 dark:bg-white/[0.03]">
                <button
                  @click="lobbyViewMode = 'player'"
                  :class="cn('px-3 py-1.5 rounded-lg text-[10px] font-black uppercase tracking-widest transition-all', lobbyViewMode === 'player' ? 'bg-blue-600 text-white' : 'text-slate-500 hover:text-slate-900 dark:hover:text-white')"
                >
                  玩家视图
                </button>
                <button
                  @click="lobbyViewMode = 'admin'"
                  :class="cn('px-3 py-1.5 rounded-lg text-[10px] font-black uppercase tracking-widest transition-all', lobbyViewMode === 'admin' ? 'bg-amber-500 text-white' : 'text-slate-500 hover:text-slate-900 dark:hover:text-white')"
                >
                  管理员视图
                </button>
              </div>
            </div>

            <!-- Experimental Nodes (Room List Cards) -->
            <div data-tutorial="room-list" class="room-grid">
              <!-- Empty State -->
              <div v-if="rooms.length === 0" class="lobby-empty-state bg-white/60 dark:bg-white/[0.02] border border-dashed border-slate-300 dark:border-white/10 rounded-3xl">
                <div class="empty-icon-wrap">
                  <Database class="w-8 h-8" />
                </div>
                <p class="empty-text-primary">暂无开放的房间</p>
                <p class="empty-text-secondary">等待玩家创建房间开始游戏...</p>
              </div>

              <!-- Room Cards -->
              <div 
                v-for="room in rooms" 
                :key="room.id"
                class="room-card"
              >
                <!-- Card Header -->
                <div class="room-card-header">
                  <div class="status-indicator">
                    <div :class="cn(
                      'status-dot',
                      room.status === 'waiting' ? (getRoomLiveCountdown(room) > 0 ? 'starting' : 'waiting') :
                      room.status === 'playing' ? 'playing' : ''
                    )"></div>
                    <span class="status-label">
                      {{ room.status === 'waiting' ? (getRoomLiveCountdown(room) > 0 ? getRoomLiveCountdown(room) + '秒' : '就绪') : room.status === 'playing' ? '进行中' : '已关闭' }}
                    </span>
                  </div>

                  <div class="room-type-badges">
                    <div v-if="isRoomFull(room) && room.status === 'waiting'" class="mode-badge" style="background:rgba(239,68,68,0.08);border:1px solid rgba(239,68,68,0.2);color:#ef4444;">
                      满员
                    </div>
                    <div v-if="room.is_points_mode" class="mode-badge mode-ranked">
                      排位
                    </div>
                  </div>
                </div>

                <!-- Card Body -->
                <div class="room-card-body">
                  <div class="room-main-info">
                    <h3 class="room-display-name">{{ room.name }}</h3>
                    <span class="room-sub-id">房间号: {{ room.id }}</span>
                  </div>

                  <div class="room-meta-container">
                    <!-- Config / Deck（在小屏三列布局下隐藏以保持紧凑） -->
                    <div v-if="room.deck_config" class="meta-item hidden sm:flex">
                      <span class="meta-label">牌组</span>
                      <button 
                        @click.stop="handleViewDeckConfig(room.deck_config)"
                        class="deck-trigger"
                      >
                        <Beaker class="w-3 h-3" />
                        {{ room.deck_config.name }}
                      </button>
                    </div>

                    <!-- Players Occupancy -->
                    <div class="occupancy-section">
                      <div class="occupancy-header">
                        <span class="meta-label">玩家人数</span>
                        <div class="occupancy-count">
                          {{ room.players?.length || 0 }}<span class="occupancy-max">/{{ room.max_players }}</span>
                        </div>
                      </div>
                      <div class="progress-track">
                        <div
                          class="progress-bar-fill"
                          :style="{ width: `${((room.players?.length || 0) / room.max_players) * 100}%`, background: isRoomFull(room) ? 'linear-gradient(to right, #ef4444, #dc2626)' : undefined }"
                        ></div>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Card Footer Action -->
                <div class="room-card-footer">
                  <template v-if="isAdminView">
                    <button
                      @click="handleTerminateRoom(room.id)"
                      class="btn-room-action btn-terminate"
                    >
                      <X class="w-3.5 h-3.5" />
                      结束
                    </button>
                    <button
                      v-if="room.status !== 'playing' && !isRoomFull(room)"
                      @click="handleJoinRoom(room.id, false)"
                      class="btn-room-action btn-enter"
                    >
                      <Play class="w-3.5 h-3.5 fill-current" />
                      加入
                    </button>
                    <button
                      @click="handleJoinRoom(room.id, true)"
                      class="btn-room-action btn-spectate"
                    >
                      <Shield class="w-3.5 h-3.5" />
                      管理员旁观
                    </button>
                  </template>

                  <!-- 非私密房间：显示加入和旁观两个选项 -->
                  <template v-else-if="!room.is_private">
                    <button
                      v-if="room.status !== 'playing' && !isRoomFull(room)"
                      @click="handleJoinRoom(room.id, false)"
                      class="btn-room-action btn-enter"
                    >
                      <Play class="w-3.5 h-3.5 fill-current" />
                      加入
                    </button>
                    <button
                      @click="handleJoinRoom(room.id, true)"
                      class="btn-room-action btn-spectate"
                    >
                      <Shield class="w-3.5 h-3.5" />
                      旁观
                    </button>
                  </template>
                  
                  <!-- 私密房间：仅显示加入或旁观 -->
                  <button
                    v-else
                    @click="handleJoinRoom(room.id, room.status === 'playing' || isRoomFull(room))"
                    :class="cn(
                      'btn-room-action',
                      room.status === 'playing' || isRoomFull(room) ? 'btn-spectate' : 'btn-enter'
                    )"
                  >
                    <component :is="room.status === 'playing' || isRoomFull(room) ? Shield : Play" class="w-3.5 h-3.5" :class="room.status !== 'playing' && !isRoomFull(room) ? 'fill-current' : ''" />
                    {{ room.status === 'playing' || isRoomFull(room) ? '旁观' : '加入' }}
                  </button>
                </div>
              </div>
                </div> <!-- room-grid end -->
            </div> <!-- lobby-content-pane end -->
            
            <!-- Right Column: World Chat & Team -->
            <div class="lobby-sidebar-pane bg-white/40 dark:bg-black/20 rounded-2xl overflow-hidden border border-slate-200 dark:border-white/5 flex flex-col gap-2">
                <TeamPanel />
                <ChatBox title="全域通信频率" placeholder="发送消息..." maxHeight="100%" class="h-full flex-1" />
            </div>
        </div> <!-- lobby-main-grid end -->
      </main>

      <!-- Global Footer Terminal -->
      <footer class="lobby-footer bg-black/40 backdrop-blur-md p-4 shrink-0">
        <div class="max-w-[1400px] mx-auto flex flex-col md:flex-row justify-between items-center text-[10px] font-mono text-slate-500 uppercase tracking-[0.15em] gap-4">
          <div class="flex items-center gap-4 order-2 md:order-1">
             <span class="hidden lg:inline text-emerald-500/60 uppercase">System_Healthy</span>
             <span class="hidden lg:inline h-3 w-px bg-white/10"></span>
             <button @click="openUserAgreement" class="hover:text-blue-400 transition-colors uppercase cursor-pointer">服务协议</button>
             <span class="h-3 w-px bg-white/10"></span>
             <button @click="openPrivacyPolicy" class="hover:text-blue-400 transition-colors uppercase cursor-pointer">隐私政策</button>
             <span class="h-3 w-px bg-white/10"></span>
             <span class="text-blue-500/50">v{{ appVersion }}</span>
          </div>
          <div class="text-center md:text-right order-1 md:order-2 opacity-40 hover:opacity-100 transition-opacity">
            &copy; 2026 MENDELEEF PROTCOL. ALL RIGHTS RESERVED.
          </div>
        </div>
      </footer>
    </div>


    <!-- 解封补偿领取提醒 -->
    <div v-if="showCompensationReminder && pendingCompensationAppeal" class="viewport-modal-overlay z-[210] p-4 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md">
      <div class="absolute inset-0 bg-slate-900/40 dark:bg-black/80 backdrop-blur-md animate-in fade-in" @click="remindCompensationLater" />
      <div class="relative w-full max-w-md overflow-hidden rounded-2xl border border-emerald-200 bg-white shadow-2xl animate-in fade-in zoom-in slide-in-from-bottom-6 duration-300 dark:border-emerald-500/20 dark:bg-[#121216]">
        <div class="flex items-start justify-between gap-4 border-b border-slate-100 px-5 py-4 dark:border-white/5">
          <div class="flex items-center gap-3">
            <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border border-emerald-500/20 bg-emerald-500/10 text-emerald-600 dark:text-emerald-300">
              <PhlogistonIcon :size="20" color="#10b981" />
            </div>
            <div>
              <h3 class="text-sm font-black text-slate-900 dark:text-white">补偿待领取</h3>
              <p class="mt-1 text-[10px] font-bold uppercase tracking-widest text-emerald-600/80 dark:text-emerald-300/80">Appeal Compensation</p>
            </div>
          </div>
          <button class="rounded-xl p-2 text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-700 dark:hover:bg-white/5 dark:hover:text-white" @click="remindCompensationLater">
            <X class="h-5 w-5" />
          </button>
        </div>

        <div class="space-y-4 px-5 py-5">
          <div class="rounded-xl border border-emerald-200 bg-emerald-50 p-4 text-sm text-emerald-800 dark:border-emerald-500/20 dark:bg-emerald-500/10 dark:text-emerald-100">
            <div class="font-black">你的申诉已通过，{{ pendingCompensationAppeal.compensation_amount || 0 }} 燃素补偿正在等待领取。</div>
            <div class="mt-2 text-xs leading-relaxed text-emerald-700 dark:text-emerald-200/80">请前往申诉中心领取。补偿不会过期，稍后领取也会一直保留。</div>
          </div>

          <div class="grid gap-2 sm:grid-cols-[1fr_auto]">
            <button
              class="inline-flex h-10 items-center justify-center gap-2 rounded-xl bg-emerald-600 px-4 text-xs font-black text-white transition-colors hover:bg-emerald-500"
              @click="goToCompensationClaim"
            >
              <FileText class="h-4 w-4" />
              前往申诉中心
            </button>
            <button
              class="inline-flex h-10 items-center justify-center rounded-xl border border-slate-200 bg-white px-4 text-xs font-black text-slate-600 transition-colors hover:bg-slate-100 dark:border-white/10 dark:bg-white/5 dark:text-slate-200 dark:hover:bg-white/10"
              @click="remindCompensationLater"
            >
              稍后提醒
            </button>
          </div>

          <button
            class="w-full text-center text-xs font-bold text-slate-400 transition-colors hover:text-slate-700 dark:hover:text-slate-200"
            @click="muteCompensationReminder"
          >
            不再提醒
          </button>
        </div>
      </div>
    </div>

    <!-- Modern Create Modal -->
    <div v-if="showCreateModal" class="viewport-modal-overlay z-[100] p-3 sm:p-4 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md">
      <div class="absolute inset-0 bg-slate-900/40 dark:bg-black/80 backdrop-blur-md animate-in fade-in" @click="showCreateModal = false" />
      <div class="relative w-full max-w-lg bg-white dark:bg-[#121216] border border-slate-200 dark:border-white/10 rounded-3xl sm:rounded-[40px] shadow-2xl overflow-hidden flex flex-col modal-mobile animate-in fade-in zoom-in slide-in-from-bottom-10 duration-500">
         <!-- Modal Header -->
         <div class="px-5 py-4 sm:px-6 sm:py-5 border-b border-slate-100 dark:border-white/5 flex items-center justify-between">
            <div class="flex items-center gap-2.5 sm:gap-3">
              <div class="w-9 h-9 sm:w-10 sm:h-10 bg-blue-500/10 border border-blue-500/20 rounded-xl flex items-center justify-center text-blue-500 dark:text-blue-400">
                <Plus class="w-5 h-5" />
              </div>
              <div>
                <h2 class="text-base sm:text-lg font-black text-slate-800 dark:text-white tracking-tight leading-none">开启新实验</h2>
                <p class="text-label-mobile text-slate-400 dark:text-slate-500 font-mono uppercase tracking-widest mt-0.5 sm:mt-1">Setup_Experiment</p>
              </div>
            </div>
            <button
              @click="showCreateModal = false"
              class="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl transition-colors text-slate-400 hover:text-slate-900 dark:hover:text-white touch-feedback"
            >
              <X class="w-5 h-5" />
            </button>
         </div>

        <form @submit.prevent="handleCreateRoom" class="flex flex-col min-h-0">
          <div class="modal-content-mobile space-y-4 sm:space-y-5 overflow-y-auto custom-scrollbar flex-1">
            <div class="space-y-2">
            <div class="flex justify-between items-center px-1">
               <label class="text-label-mobile font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">实验空间命名</label>
               <span class="text-caption-mobile text-blue-500/40 font-mono">IDENTIFIER</span>
            </div>
            <input
              v-model="roomName"
              type="text"
              autofocus
              placeholder="默认随机分配名称..."
              class="w-full input-mobile bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/5 text-slate-900 dark:text-white px-4 sm:px-4 py-3 rounded-xl focus:ring-1 focus:ring-blue-500/50 focus:border-blue-500/50 outline-none transition-all placeholder:text-slate-300 dark:placeholder:text-slate-700 font-mono"
            />
          </div>

          <div class="space-y-2">
            <div class="flex justify-between items-center px-1">
               <label class="text-label-mobile font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">研究员容量</label>
               <span class="text-caption-mobile text-blue-500/40 font-mono">CAPACITY</span>
            </div>
            <div class="grid grid-cols-4 gap-2.5 sm:gap-3">
              <button
                v-for="num in [2, 3, 4, 8]"
                :key="num"
                type="button"
                @click="maxPlayers = num"
                :class="cn(
                  'h-11 sm:h-10 rounded-xl text-xs sm:text-[11px] font-black border transition-all flex items-center justify-center relative group/opt overflow-hidden touch-feedback',
                  maxPlayers === num
                    ? 'bg-blue-500/10 border-blue-500/50 text-blue-600 dark:text-blue-400 ring-1 ring-blue-500/20 shadow-[0_4px_12px_rgba(59,130,246,0.1)]'
                    : 'bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/5 text-slate-400 dark:text-slate-600 hover:bg-slate-100 dark:hover:bg-white/10'
                )"
              >
                <span class="relative z-10">{{ num }}P</span>
                <div v-if="maxPlayers === num" class="absolute inset-0 bg-blue-500/5 animate-pulse"></div>
              </button>
            </div>
          </div>

          <div class="space-y-2">
            <div class="flex justify-between items-center px-1">
               <label class="text-label-mobile font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">核心配置</label>
               <span class="text-caption-mobile text-blue-500/40 font-mono">PROTOCOL</span>
            </div>
            <div class="flex items-center gap-3 p-3.5 sm:p-3 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/5 rounded-xl group/toggle cursor-pointer transition-all hover:bg-slate-100 dark:hover:bg-white/10 touch-feedback" @click="isPrivate = !isPrivate">
              <div :class="cn(
                'w-10 h-5 sm:w-8 sm:h-4.5 rounded-full relative transition-colors duration-300',
                isPrivate ? 'bg-amber-600' : 'bg-slate-300 dark:bg-slate-700'
              )">
                <div :class="cn(
                  'absolute top-1 left-1 sm:top-0.75 sm:left-0.75 w-3 h-3 bg-white rounded-full transition-transform duration-300',
                  isPrivate ? 'translate-x-5 sm:translate-x-3.5' : 'translate-x-0'
                )"></div>
              </div>
              <div class="flex flex-col">
                <span :class="cn('text-xs sm:text-[9px] font-black uppercase tracking-wider', isPrivate ? 'text-amber-600 dark:text-amber-400' : 'text-slate-400 dark:text-slate-400')">
                  私密隐藏频道
                </span>
                <span class="text-caption-mobile text-slate-400 dark:text-slate-500 mt-0.5 leading-tight">
                  不显示在大厅，仅能通过二维码/链接加入
                </span>
              </div>
            </div>

            <!-- AI补位功能 -->
            <div class="flex items-center gap-3 p-3.5 sm:p-3 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/5 rounded-xl group/toggle cursor-pointer transition-all hover:bg-slate-100 dark:hover:bg-white/10 touch-feedback" @click="enableAIBackfill = !enableAIBackfill">
              <div :class="cn(
                'w-10 h-5 sm:w-8 sm:h-4.5 rounded-full relative transition-colors duration-300',
                enableAIBackfill ? 'bg-purple-600' : 'bg-slate-300 dark:bg-slate-700'
              )">
                <div :class="cn(
                  'absolute top-1 left-1 sm:top-0.75 sm:left-0.75 w-3 h-3 bg-white rounded-full transition-transform duration-300',
                  enableAIBackfill ? 'translate-x-5 sm:translate-x-3.5' : 'translate-x-0'
                )"></div>
              </div>
              <div class="flex flex-col">
                <span :class="cn('text-xs sm:text-[9px] font-black uppercase tracking-wider', enableAIBackfill ? 'text-purple-600 dark:text-purple-400' : 'text-slate-400 dark:text-slate-400')">
                  AI智能补位
                </span>
                <span class="text-caption-mobile text-slate-400 dark:text-slate-500 mt-0.5 leading-tight">
                  开始游戏时自动用AI填补空缺位置
                </span>
              </div>
            </div>

            <!-- 等级匹配模式 -->
            <div class="flex items-center gap-3 p-3.5 sm:p-3 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/5 rounded-xl group/toggle cursor-pointer transition-all hover:bg-slate-100 dark:hover:bg-white/10 touch-feedback" @click="isRanked = !isRanked">
              <div :class="cn(
                'w-10 h-5 sm:w-8 sm:h-4.5 rounded-full relative transition-colors duration-300',
                isRanked ? 'bg-indigo-600' : 'bg-slate-300 dark:bg-slate-700'
              )">
                <div :class="cn(
                  'absolute top-1 left-1 sm:top-0.75 sm:left-0.75 w-3 h-3 bg-white rounded-full transition-transform duration-300',
                  isRanked ? 'translate-x-5 sm:translate-x-3.5' : 'translate-x-0'
                )"></div>
              </div>
              <div class="flex flex-col">
                <span :class="cn('text-xs sm:text-[9px] font-black uppercase tracking-wider', isRanked ? 'text-indigo-600 dark:text-indigo-400' : 'text-slate-400 dark:text-slate-400')">
                  等级均衡匹配
                </span>
                <span class="text-caption-mobile text-slate-400 dark:text-slate-500 mt-0.5 leading-tight">
                  自动过滤等级差距过大的研究员
                </span>
              </div>
            </div>
          </div>

          <!-- 等级范围设置 -->
          <div v-if="isRanked" class="space-y-2 animate-in slide-in-from-top-2 fade-in duration-300">
            <div class="flex justify-between items-center px-1">
               <label class="text-label-mobile font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">匹配跨度（等级差距）</label>
               <span class="text-caption-mobile text-indigo-500/40 font-mono">RANGE: ±{{ levelRange }}</span>
            </div>
            <div class="p-4 bg-indigo-50 dark:bg-indigo-500/5 border border-indigo-200 dark:border-indigo-500/20 rounded-xl">
              <input
                v-model.number="levelRange"
                type="range"
                min="3"
                max="10"
                class="w-full h-2 bg-indigo-200 dark:bg-indigo-500/20 rounded-lg appearance-none cursor-pointer accent-indigo-600"
              />
              <div class="flex justify-between mt-2.5 sm:mt-2">
                <span class="text-caption-mobile text-indigo-600 dark:text-indigo-400 font-mono">严苛(3)</span>
                <span class="text-caption-mobile text-indigo-600 dark:text-indigo-400 font-mono">标准(5)</span>
                <span class="text-caption-mobile text-indigo-600 dark:text-indigo-400 font-mono">宽松(10)</span>
              </div>
            </div>
          </div>

          <!-- AI补位难度设置 -->
          <div v-if="enableAIBackfill" class="space-y-2 animate-in slide-in-from-top-2 fade-in duration-300">
            <div class="flex justify-between items-center px-1">
               <label class="text-label-mobile font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">AI难度等级</label>
               <span class="text-caption-mobile text-purple-500/40 font-mono">DIFFICULTY: {{ aiBackfillDifficulty }}</span>
            </div>
            <div class="p-4 bg-purple-50 dark:bg-purple-500/5 border border-purple-200 dark:border-purple-500/20 rounded-xl">
              <input
                v-model.number="aiBackfillDifficulty"
                type="range"
                min="1"
                max="100"
                class="w-full h-2 bg-purple-200 dark:bg-purple-500/20 rounded-lg appearance-none cursor-pointer accent-purple-600"
              />
              <div class="flex justify-between mt-2.5 sm:mt-2">
                <span class="text-caption-mobile text-purple-600 dark:text-purple-400 font-mono">简单</span>
                <span class="text-caption-mobile text-purple-600 dark:text-purple-400 font-mono">中等</span>
                <span class="text-caption-mobile text-purple-600 dark:text-purple-400 font-mono">困难</span>
              </div>
            </div>
            <p class="text-caption-mobile text-purple-600 dark:text-purple-500 px-1 leading-relaxed">
              🤖 <span class="font-bold">提示：</span>AI难度越高，决策越优秀。建议燃素模式下使用60+难度
            </p>
          </div>

          <!-- 私密房间密钥输入框 -->
          <div v-if="isPrivate" class="space-y-2 animate-in slide-in-from-top-2 fade-in duration-300">
            <div class="flex justify-between items-center px-1">
               <label class="text-label-mobile font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">访问密钥</label>
               <span class="text-caption-mobile text-amber-500/40 font-mono">ACCESS_KEY</span>
            </div>
            <input
              v-model="customAccessKey"
              type="text"
              placeholder="留空自动生成8位密钥..."
              maxlength="20"
              class="w-full input-mobile bg-amber-50 dark:bg-amber-500/5 border border-amber-200 dark:border-amber-500/20 text-slate-900 dark:text-white px-4 rounded-xl focus:ring-1 focus:ring-amber-500/50 focus:border-amber-500/50 outline-none transition-all placeholder:text-slate-300 dark:placeholder:text-slate-700 font-mono"
            />
            <p class="text-caption-mobile text-amber-600 dark:text-amber-500 px-1 leading-relaxed">
              💡 <span class="font-bold">提示：</span>可自定义4-20位密钥，或留空由系统自动生成
            </p>
          </div>

          <div v-if="!isPointsMode" class="space-y-2">
            <div class="flex justify-between items-center px-1">
               <label class="text-label-mobile font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">选择牌组</label>
               <span class="text-caption-mobile text-blue-500/40 font-mono">DECK</span>
            </div>
            <div v-if="decks.length === 0" class="p-4 bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl text-center">
              <p class="text-xs-mobile text-slate-500 dark:text-slate-400 font-mono">暂无可用牌组，将使用默认全局牌组</p>
            </div>
            <div v-else class="space-y-2">
              <button
                v-for="deck in decks"
                :key="deck.id"
                type="button"
                @click="deckID = deck.id"
                :class="cn(
                  'w-full flex items-center gap-3 p-3.5 sm:p-3 rounded-xl border transition-all text-left group/deck touch-feedback',
                  deckID === deck.id
                    ? 'bg-blue-600/5 dark:bg-blue-600/10 border-blue-500/50 shadow-[0_4px_12px_rgba(59,130,246,0.05)]'
                    : 'bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/5 hover:border-slate-300 dark:hover:border-white/10'
                )"
              >
                <div :class="cn(
                  'w-10 h-10 sm:w-9 sm:h-9 rounded-lg flex items-center justify-center transition-colors',
                  deckID === deck.id ? 'bg-blue-500 text-white' : 'bg-slate-200 dark:bg-white/5 text-slate-400 dark:text-slate-500'
                )">
                   <Beaker class="w-5 h-5 sm:w-4 sm:h-4" />
                </div>
                <div class="flex-1">
                  <p :class="cn('text-xs sm:text-[11px] font-black uppercase tracking-wider', deckID === deck.id ? 'text-blue-600 dark:text-blue-400' : 'text-slate-700 dark:text-white')">
                    {{ deck.name }}
                  </p>
                  <p class="text-caption-mobile text-slate-400 dark:text-slate-500 mt-0.5 font-mono uppercase tracking-tighter">
                    {{ Object.keys(deck.cards || {}).length }} Elements
                  </p>
                </div>
                <div v-if="deckID === deck.id" class="w-2 h-2 sm:w-1.5 sm:h-1.5 rounded-full bg-blue-500 animate-pulse mr-1"></div>
              </button>
            </div>
          </div>

          </div>
          <div class="p-5 sm:p-6 border-t border-slate-100 dark:border-white/5 bg-slate-50/50 dark:bg-white/[0.02] flex gap-3 shrink-0">
            <button
              type="button"
              @click="showCreateModal = false"
              class="flex-1 button-mobile bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-500 dark:text-slate-400 font-bold rounded-xl transition-all uppercase tracking-widest text-xs sm:text-[10px] border border-slate-200 dark:border-white/5 touch-feedback"
            >
              放弃
            </button>
            <button
              type="submit"
              :disabled="loading"
              class="flex-1 button-mobile bg-blue-600 hover:bg-blue-500 text-white font-black rounded-xl transition-all shadow-[0_8px_20px_rgba(37,99,235,0.2)] touch-feedback disabled:grayscale flex items-center justify-center gap-2 group/sub relative overflow-hidden"
            >
              <template v-if="loading">
                <Loader2 class="w-5 h-5 animate-spin" />
              </template>
              <template v-else>
                <Play class="w-3.5 h-3.5 fill-current" />
                <span class="uppercase tracking-[0.2em] text-xs">执行初始化</span>
              </template>
              <div class="absolute inset-0 w-full h-full bg-gradient-to-r from-transparent via-white/10 to-transparent -translate-x-full group-hover/sub:animate-shimmer"></div>
            </button>
          </div>
        </form>
      </div>

      </div>
    </div>

    <!-- AI Arena Modal -->
    <div v-if="showAIArenaModal" class="viewport-modal-overlay z-[100] p-4 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md">
      <div class="absolute inset-0 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md animate-in fade-in" @click="showAIArenaModal = false" />
      <div class="relative w-full max-w-lg bg-white dark:bg-[#121216] border border-purple-500/30 rounded-[40px] shadow-2xl shadow-purple-500/10 overflow-hidden flex flex-col max-h-[90vh] animate-in fade-in zoom-in slide-in-from-bottom-10 duration-500">
         <!-- Modal Header -->
         <div class="px-6 py-5 border-b border-slate-100 dark:border-white/5 flex items-center justify-between bg-purple-500/5">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 bg-purple-500/10 border border-purple-500/20 rounded-xl flex items-center justify-center text-purple-600 dark:text-purple-400">
                <Beaker class="w-5 h-5" />
              </div>
              <div>
                <h2 class="text-lg font-black text-slate-800 dark:text-white tracking-tight leading-none">人机竞技场</h2>
                <p class="text-[9px] text-purple-500/60 font-mono uppercase tracking-widest mt-1">AI_Challenge_Mode</p>
              </div>
            </div>
            <button 
              @click="showAIArenaModal = false"
              class="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl transition-colors text-slate-400 hover:text-slate-900 dark:hover:text-white"
            >
              <X class="w-5 h-5" />
            </button>
         </div>

        <form @submit.prevent="handleCreateAIRoom" class="flex flex-col min-h-0">
          <div class="p-6 space-y-6 overflow-y-auto custom-scrollbar flex-1">
            
            <!-- Difficulty Slider -->
            <div class="space-y-4">
              <div class="flex justify-between items-end px-1">
                 <div>
                   <label class="text-[9px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest block mb-1">AI 智能等级</label>
                   <h3 class="text-2xl font-black text-slate-800 dark:text-white">{{ pveDifficulty }}<span class="text-sm text-slate-400 ml-1">%</span></h3>
                 </div>
                 <span class="text-[8px] text-purple-500/40 font-mono">DIFFICULTY</span>
              </div>
              <div class="relative h-6 flex items-center">
                <input 
                  type="range" 
                  v-model.number="pveDifficulty" 
                  min="1" 
                  max="100" 
                  class="w-full h-2 bg-slate-200 dark:bg-white/10 rounded-lg appearance-none cursor-pointer accent-purple-600"
                />
              </div>
              <div class="flex justify-between text-[10px] text-slate-400 font-mono">
                <span>Rookie (1%)</span>
                <span>Grandmaster (100%)</span>
              </div>
            </div>

            <!-- AI Count -->
            <div class="space-y-3">
              <div class="flex justify-between items-center px-1">
                 <label class="text-[9px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">AI 对手数量</label>
                 <span class="text-[8px] text-purple-500/40 font-mono">OPPONENTS</span>
              </div>
              <div class="grid grid-cols-4 gap-3">
                <button
                  v-for="num in [1, 2, 3, 7]"
                  :key="num"
                  type="button"
                  @click="aiCount = num"
                  :class="cn(
                    'h-12 rounded-xl text-sm font-black border transition-all flex flex-col items-center justify-center relative group/opt overflow-hidden',
                    aiCount === num
                      ? 'bg-purple-500/10 border-purple-500/50 text-purple-600 dark:text-purple-400 ring-1 ring-purple-500/20 shadow-[0_4px_12px_rgba(168,85,247,0.1)]'
                      : 'bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/5 text-slate-400 dark:text-slate-600 hover:bg-slate-100 dark:hover:bg-white/10'
                  )"
                >
                  <span class="relative z-10">{{ num }} AI</span>
                  <div v-if="aiCount === num" class="absolute inset-0 bg-purple-500/5 animate-pulse"></div>
                </button>
              </div>
            </div>

            <!-- Deck Selection -->
            <div class="space-y-2">
              <div class="flex justify-between items-center px-1">
                <label class="text-[9px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">选择牌组</label>
                <span class="text-[8px] text-purple-500/40 font-mono">DECK</span>
              </div>
              <div v-if="decks.length === 0" class="p-4 bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl text-center">
                <p class="text-[10px] text-slate-500 dark:text-slate-400 font-mono">暂无可用牌组，将使用默认全局牌组</p>
              </div>
              <div v-else class="space-y-2">
                <button
                  v-for="deck in decks"
                  :key="deck.id"
                  type="button"
                  @click="deckID = deck.id"
                  :class="cn(
                    'w-full flex items-center gap-3 p-3.5 rounded-xl border transition-all text-left',
                    deckID === deck.id
                      ? 'bg-purple-600/5 dark:bg-purple-600/10 border-purple-500/50 shadow-[0_4px_12px_rgba(168,85,247,0.05)]'
                      : 'bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/5 hover:border-slate-300 dark:hover:border-white/10'
                  )"
                >
                  <div :class="cn(
                    'w-9 h-9 rounded-lg flex items-center justify-center transition-colors',
                    deckID === deck.id ? 'bg-purple-500 text-white' : 'bg-slate-200 dark:bg-white/5 text-slate-400 dark:text-slate-500'
                  )">
                    <Beaker class="w-4 h-4" />
                  </div>
                  <div class="flex-1">
                    <p :class="cn('text-[11px] font-black uppercase tracking-wider', deckID === deck.id ? 'text-purple-600 dark:text-purple-400' : 'text-slate-700 dark:text-white')">
                      {{ deck.name }}
                    </p>
                    <p class="text-[10px] text-slate-400 dark:text-slate-500 mt-0.5 font-mono uppercase tracking-tighter">
                      {{ Object.keys(deck.cards || {}).length }} Elements
                    </p>
                  </div>
                  <div v-if="deckID === deck.id" class="w-1.5 h-1.5 rounded-full bg-purple-500 animate-pulse mr-1"></div>
                </button>
              </div>
            </div>

          </div>
          <div class="p-6 border-t border-slate-100 dark:border-white/5 bg-slate-50/50 dark:bg-white/[0.02] flex gap-3 shrink-0">
            <button 
              type="button"
              @click="showAIArenaModal = false"
              class="flex-1 px-4 py-3 border border-slate-200 dark:border-white/10 rounded-xl text-[11px] font-black uppercase tracking-widest hover:bg-slate-100 dark:hover:bg-white/5 transition-colors text-slate-500"
            >
              取消
            </button>
            <button 
              type="submit" 
              :disabled="loading"
              class="flex-[2] px-6 py-3 bg-purple-600 hover:bg-purple-500 text-white rounded-xl text-[11px] font-black uppercase tracking-widest shadow-lg shadow-purple-500/20 hover:shadow-purple-500/40 hover:-translate-y-0.5 active:translate-y-0 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              <Loader2 v-if="loading" class="w-4 h-4 animate-spin" />
              <span>开始挑战</span>
            </button>
          </div>
        </form>
      </div>
    </div>


    <!-- 牌组详情查看模态框 -->
    <div v-if="showDeckDetailModal && selectedDeckConfig" class="viewport-modal-overlay z-[100] p-4 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md">
      <div class="absolute inset-0 bg-slate-900/40 dark:bg-black/80 backdrop-blur-md animate-in fade-in" @click="showDeckDetailModal = false" />
      <div class="relative w-full max-w-2xl bg-white dark:bg-[#121216] border border-slate-200 dark:border-white/10 rounded-[40px] shadow-2xl overflow-hidden animate-in fade-in zoom-in slide-in-from-bottom-10 duration-500">
         <!-- Modal Header -->
         <div class="px-6 py-5 border-b border-slate-100 dark:border-white/5 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 bg-blue-500/10 border border-blue-500/20 rounded-xl flex items-center justify-center text-blue-500 dark:text-blue-400">
                <Database class="w-5 h-5" />
              </div>
              <div>
                <h2 class="text-lg font-black text-slate-800 dark:text-white tracking-tight leading-none">{{ selectedDeckConfig.name }}</h2>
                <p class="text-[9px] text-slate-400 dark:text-slate-500 font-mono uppercase tracking-widest mt-1">Deck_Configuration_Details</p>
              </div>
            </div>
            <button
              @click="showDeckDetailModal = false"
              class="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl transition-colors text-slate-400 hover:text-slate-900 dark:hover:text-white"
            >
              <X class="w-5 h-5" />
            </button>
         </div>

         <!-- Modal Content -->
         <div class="p-6 max-h-[60vh] overflow-y-auto custom-scrollbar">
            <div class="space-y-4">
              <!-- 基础信息 -->
              <div class="p-4 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl">
                <div class="flex items-center justify-between mb-3">
                  <span class="text-[9px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">基础信息</span>
                  <span class="text-[8px] text-blue-500/40 font-mono">BASIC_INFO</span>
                </div>
                <div class="grid grid-cols-2 gap-3">
                  <div class="p-3 bg-white dark:bg-black/20 rounded-lg">
                    <p class="text-[8px] text-slate-400 mb-1 uppercase tracking-wider">牌组名称</p>
                    <p class="text-[11px] font-black text-slate-900 dark:text-white">{{ selectedDeckConfig.name }}</p>
                  </div>
                  <div class="p-3 bg-white dark:bg-black/20 rounded-lg">
                    <p class="text-[8px] text-slate-400 mb-1 uppercase tracking-wider">元素种类</p>
                    <p class="text-[11px] font-black text-blue-600 dark:text-blue-400">{{ Object.keys(selectedDeckConfig.cards || {}).length }} 种</p>
                  </div>
                  <div class="p-3 bg-white dark:bg-black/20 rounded-lg">
                    <p class="text-[8px] text-slate-400 mb-1 uppercase tracking-wider">总卡牌数</p>
                    <p class="text-[11px] font-black text-slate-900 dark:text-white">{{ (Object.values(selectedDeckConfig.cards || {}) as number[]).reduce((a, b) => a + b, 0) }} 张</p>
                  </div>
                  <div class="p-3 bg-white dark:bg-black/20 rounded-lg">
                    <p class="text-[8px] text-slate-400 mb-1 uppercase tracking-wider">起始手牌</p>
                    <p class="text-[11px] font-black text-slate-900 dark:text-white">{{ selectedDeckConfig.initial_cards || 10 }} 张</p>
                  </div>
                </div>
              </div>

              <!-- 卡牌列表 -->
              <div class="p-4 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl">
                <div class="flex items-center justify-between mb-3">
                  <span class="text-[9px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">卡牌配置</span>
                  <span class="text-[8px] text-blue-500/40 font-mono">CARD_LIST</span>
                </div>
                <div class="grid grid-cols-2 sm:grid-cols-3 gap-2 max-h-64 overflow-y-auto custom-scrollbar pr-2">
                  <div
                    v-for="(count, formula) in selectedDeckConfig.cards"
                    :key="formula"
                    class="p-2.5 bg-white dark:bg-black/20 rounded-lg border border-slate-200 dark:border-white/10 hover:border-blue-500/50 transition-colors"
                  >
                    <div class="flex items-center justify-between">
                      <span class="text-[10px] font-black text-slate-900 dark:text-white font-mono" v-html="String(formula).replace(/(\d+)/g, '<sub>$1</sub>')"></span>
                      <span class="text-[9px] font-black text-blue-600 dark:text-blue-400">×{{ count }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
         </div>

         <!-- Modal Footer -->
         <div class="px-6 py-4 border-t border-slate-100 dark:border-white/5 flex justify-end">
            <button
              @click="showDeckDetailModal = false"
              class="px-5 py-2.5 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-700 dark:text-slate-300 font-bold rounded-xl transition-all uppercase tracking-widest text-[10px] border border-slate-200 dark:border-white/5"
            >
              关闭
            </button>
         </div>
      </div>
    </div>

    <!-- 私密房间密钥模态框 -->
    <div v-if="showAccessKeyModal && createdRoomInfo" class="viewport-modal-overlay z-[100] p-4 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md">
      <div class="absolute inset-0 bg-slate-900/40 dark:bg-black/80 backdrop-blur-md animate-in fade-in" @click="showAccessKeyModal = false" />
      <div class="relative w-full max-w-md bg-white dark:bg-[#121216] border border-slate-200 dark:border-white/10 rounded-[40px] shadow-2xl overflow-hidden animate-in fade-in zoom-in slide-in-from-bottom-10 duration-500">
         <!-- Modal Header -->
         <div class="px-6 py-5 border-b border-slate-100 dark:border-white/5 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 bg-green-500/10 border border-green-500/20 rounded-xl flex items-center justify-center text-green-500 dark:text-green-400">
                <Shield class="w-5 h-5" />
              </div>
              <div>
                <h2 class="text-lg font-black text-slate-800 dark:text-white tracking-tight leading-none">私密房间已创建</h2>
                <p class="text-[9px] text-slate-400 dark:text-slate-500 font-mono uppercase tracking-widest mt-1">Private_Room_Created</p>
              </div>
            </div>
            <button
              @click="showAccessKeyModal = false; router.push(`/room/${createdRoomInfo.id}`)"
              class="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl transition-colors text-slate-400 hover:text-slate-900 dark:hover:text-white"
            >
              <X class="w-5 h-5" />
            </button>
         </div>

         <!-- Modal Body -->
         <div class="p-6 space-y-6">
            <div class="bg-green-500/5 border border-green-500/20 rounded-2xl p-5">
              <p class="text-xs text-slate-600 dark:text-slate-400 mb-4 leading-relaxed">
                ✅ 房间已成功创建！以下是访问密钥和分享链接，请妥善保管。
              </p>

              <!-- 访问密钥 -->
              <div class="mb-4">
                <label class="text-[9px] font-black text-slate-500 dark:text-slate-500 uppercase tracking-[0.2em] block mb-2">Access Key / 访问密钥</label>
                <div class="flex gap-2">
                  <div class="flex-1 bg-white dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-3 font-mono text-lg font-black text-green-600 dark:text-green-400 tracking-widest text-center select-all">
                    {{ createdRoomInfo.access_key }}
                  </div>
                  <button
                    @click="copyToClipboard(createdRoomInfo.access_key)"
                    class="px-4 py-3 bg-green-500/10 hover:bg-green-500/20 border border-green-500/20 text-green-600 dark:text-green-400 rounded-xl transition-all font-black text-[10px] uppercase tracking-widest"
                    title="复制密钥"
                  >
                    复制
                  </button>
                </div>
              </div>

              <!-- 分享链接 -->
              <div>
                <label class="text-[9px] font-black text-slate-500 dark:text-slate-500 uppercase tracking-[0.2em] block mb-2">Share Link / 分享链接</label>
                <div class="flex gap-2">
                  <div class="flex-1 bg-white dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-3 font-mono text-[10px] text-slate-600 dark:text-slate-400 truncate select-all">
                    {{ currentOrigin }}/room/{{ createdRoomInfo.id }}?key={{ createdRoomInfo.access_key }}
                  </div>
                  <button
                    @click="copyToClipboard(`${currentOrigin}/room/${createdRoomInfo.id}?key=${createdRoomInfo.access_key}`)"
                    class="px-4 py-3 bg-blue-500/10 hover:bg-blue-500/20 border border-blue-500/20 text-blue-600 dark:text-blue-400 rounded-xl transition-all font-black text-[10px] uppercase tracking-widest"
                    title="复制链接"
                  >
                    复制
                  </button>
                </div>
              </div>
            </div>

            <div class="bg-amber-500/5 border border-amber-500/20 rounded-xl p-4">
              <p class="text-[10px] text-amber-700 dark:text-amber-400 leading-relaxed">
                ⚠️ <span class="font-black">提示：</span>其他玩家需要使用访问密钥或完整链接才能加入此私密房间。
              </p>
            </div>
         </div>

         <!-- Modal Footer -->
         <div class="px-6 py-4 border-t border-slate-100 dark:border-white/5 flex gap-3">
            <button
              @click="showAccessKeyModal = false"
              class="flex-1 px-5 py-2.5 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-700 dark:text-slate-300 font-bold rounded-xl transition-all uppercase tracking-widest text-[10px] border border-slate-200 dark:border-white/5"
            >
              稍后进入
            </button>
            <button
              @click="showAccessKeyModal = false; router.push(`/room/${createdRoomInfo.id}`)"
              class="flex-1 px-5 py-2.5 bg-green-600 hover:bg-green-500 text-white font-black rounded-xl transition-all shadow-lg shadow-green-500/20 active:scale-95 uppercase tracking-widest text-[10px]"
            >
              立即进入房间
            </button>
         </div>
      </div>
    </div>

    <!-- Tutorial Guide for Lobby -->
    <TutorialGuide :show="showTutorial" :steps="lobbyTutorialSteps" @close="handleTutorialClose" @complete="handleTutorialComplete" />

    <!-- 调查问卷弹窗 (内部) -->
    <div v-if="showSurveyModal && currentSurvey" class="viewport-modal-overlay bg-slate-900/40 dark:bg-black/80 backdrop-blur-md z-[100] p-3">
      <!-- ... existing survey modal code ... -->
    </div>

    <!-- Legal Document Modal -->
    <div v-if="showLegalModal" class="viewport-modal-overlay bg-slate-900/60 dark:bg-black/80 backdrop-blur-md z-[200] p-5">
      <div class="relative w-full max-w-2xl bg-white dark:bg-[#0c0c0e] border border-slate-200 dark:border-white/10 rounded-[2.5rem] shadow-4xl flex flex-col max-h-[85vh] animate-in zoom-in duration-300">
        <div class="px-8 py-6 border-b border-slate-100 dark:border-white/5 flex items-center justify-between sticky top-0 bg-inherit z-10">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 bg-blue-500/10 border border-blue-500/20 rounded-2xl flex items-center justify-center text-blue-500">
              <FileText class="w-5 h-5" />
            </div>
            <div>
              <h3 class="text-sm font-black text-slate-800 dark:text-white uppercase tracking-widest">{{ legalModalTitle }}</h3>
              <p class="text-[9px] font-mono text-slate-400 mt-1 uppercase tracking-widest">Protocol_Legal_Archive</p>
            </div>
          </div>
          <button @click="showLegalModal = false" class="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-full transition-colors text-slate-400">
            <X class="w-6 h-6" />
          </button>
        </div>
        
        <div class="p-8 overflow-y-auto custom-scrollbar flex-1 whitespace-pre-wrap">
          <div class="text-xs sm:text-sm font-medium text-slate-600 dark:text-slate-300 leading-relaxed font-sans">
            {{ legalModalContent }}
            
            <div class="mt-8 pt-8 border-t border-slate-100 dark:border-white/5">
              <p class="text-[10px] text-slate-400 italic">这是一份旨在通过实验精神而非法律博弈来维护社区的通用协议。如果您有任何疑问，请联系基地管理员。</p>
            </div>
          </div>
        </div>

        <div class="px-8 py-6 border-t border-slate-100 dark:border-white/5 bg-slate-50/50 dark:bg-white/[0.02]">
          <button 
            @click="showLegalModal = false"
            class="w-full py-4 bg-blue-600 hover:bg-blue-500 text-white font-black rounded-2xl transition-all shadow-lg shadow-blue-500/20 uppercase tracking-widest text-xs active:scale-95"
          >
            我已阅读并知悉协议内容 (Protocol_ACK)
          </button>
        </div>
      </div>
    </div>

    <!-- 调查问卷弹窗 (仅在此处保留一个逻辑完整的版本) -->
    <div v-if="showSurveyModal && currentSurvey" class="viewport-modal-overlay bg-slate-900/60 dark:bg-black/80 backdrop-blur-md z-[100] p-3">
      <div class="bg-white dark:bg-[#0c0c0e] border border-slate-200 dark:border-white/10 rounded-[2rem] p-5 max-w-sm w-full shadow-[0_50px_100px_-20px_rgba(79,70,229,0.3)] animate-in zoom-in relative overflow-hidden group">
        <div class="absolute top-0 right-0 w-32 h-32 bg-indigo-500/5 blur-[50px] -mr-16 -mt-16 group-hover:bg-indigo-500/10 transition-all opacity-50" />
        
        <div class="relative z-10">
          <div class="flex items-center gap-3 mb-4">
            <div class="w-10 h-10 bg-indigo-500/10 text-indigo-500 border border-indigo-500/20 rounded-xl flex items-center justify-center shadow-indigo-500/5 transition-all group-hover:scale-110">
              <FileText class="w-5 h-5" />
            </div>
            <div class="flex-1">
              <h3 class="text-xs font-black text-slate-400 dark:text-slate-500 uppercase tracking-[0.2em] mb-0.5">研究调查 / Survey</h3>
              <h4 class="text-sm font-black text-slate-900 dark:text-white uppercase italic tracking-tighter line-clamp-1">{{ currentSurvey.title }}</h4>
            </div>
          </div>

          <div class="bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/5 rounded-2xl p-4 mb-4 transition-all">
             <p class="text-[11px] font-bold text-slate-500 dark:text-slate-400 leading-relaxed italic line-clamp-2 mb-3">“{{ currentSurvey.description || '您的意见对我们非常重要，请根据实际体验完成本次问卷。' }}”</p>
             
             <!-- Rewards Badge -->
             <div v-if="currentSurvey.reward_points > 0 || currentSurvey.reward_exp > 0" class="flex items-center gap-2 pt-3 border-t border-slate-200 dark:border-white/5">
               <div v-if="currentSurvey.reward_points > 0" class="flex items-center gap-1 px-2 py-0.5 bg-amber-500/10 border border-amber-500/20 rounded-lg">
                 <div class="w-1 h-1 rounded-full bg-amber-500 animate-pulse"></div>
                 <span class="text-[8px] font-black text-amber-600 dark:text-amber-500 uppercase">+{{ currentSurvey.reward_points }} PTS</span>
               </div>
               <div v-if="currentSurvey.reward_exp > 0" class="flex items-center gap-1 px-2 py-0.5 bg-blue-500/10 border border-blue-500/20 rounded-lg">
                 <div class="w-1 h-1 rounded-full bg-blue-500 animate-pulse"></div>
                 <span class="text-[8px] font-black text-blue-600 dark:text-blue-400 uppercase">+{{ currentSurvey.reward_exp }} EXP</span>
               </div>
             </div>
          </div>

          <div class="grid grid-cols-1 gap-2">
            <button
              @click="handleCloseSurveyModal(); router.push('/surveys/' + currentSurvey.id)"
              class="w-full px-5 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl font-black uppercase tracking-[0.15em] text-[10px] transition-all shadow-lg shadow-indigo-500/20 flex items-center justify-center gap-2 border border-white/10 active:scale-95"
            >
              <FileText class="w-3.5 h-3.5" />
              立即填写 / RESPOND_NOW
            </button>
            
            <div class="grid grid-cols-2 gap-2">
              <button 
                @click="handleCloseSurveyModal"
                class="px-4 py-2 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-500 dark:text-slate-400 rounded-xl font-black uppercase tracking-widest text-[9px] transition-all border border-slate-200 dark:border-white/5"
              >
                稍后再说
              </button>
              <button 
                @click="handleDismissSurvey"
                class="px-4 py-2 bg-slate-50 dark:bg-white/5 hover:bg-red-500/10 text-slate-400 hover:text-red-500 rounded-xl font-black uppercase tracking-widest text-[9px] transition-all border border-slate-200 dark:border-white/5"
              >
                不再提醒
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
</template>
