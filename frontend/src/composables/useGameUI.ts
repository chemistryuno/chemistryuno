/**
 * UI 状态管理 Composable
 * 负责管理游戏界面的 UI 状态（模态框、面板、显示/隐藏等）
 */

import { ref, computed } from 'vue'

export function useGameUI() {
  // UI 状态
  const showHints = ref(true)
  const showPlayers = ref(false)
  const showChat = ref(false)
  const showAdminModal = ref(false)
  const showInviteFriendsModal = ref(false)
  const showDeckDetailModal = ref(false)
  const showChemicalKeyboard = ref(false)
  const showQrModal = ref(false)
  const hasNewMessage = ref(false)
  const isMobile = ref(false)

  // 计算属性
  const isAnyModalOpen = computed(() => {
    return showAdminModal.value ||
           showInviteFriendsModal.value ||
           showDeckDetailModal.value
  })

  // 方法
  const closeAllModals = () => {
    showAdminModal.value = false
    showInviteFriendsModal.value = false
    showDeckDetailModal.value = false
    showQrModal.value = false
  }

  const closeAllPanels = () => {
    showHints.value = false
    showPlayers.value = false
    showChat.value = false
    showChemicalKeyboard.value = false
  }

  const togglePanel = (panel: 'hints' | 'players' | 'chat') => {
    switch (panel) {
      case 'hints':
        showHints.value = !showHints.value
        break
      case 'players':
        showPlayers.value = !showPlayers.value
        break
      case 'chat':
        showChat.value = !showChat.value
        hasNewMessage.value = false
        break
    }
  }

  const openModal = (modal: 'admin' | 'inviteFriends' | 'deckDetail') => {
    switch (modal) {
      case 'admin':
        showAdminModal.value = true
        break
      case 'inviteFriends':
        showInviteFriendsModal.value = true
        break
      case 'deckDetail':
        showDeckDetailModal.value = true
        break
    }
  }

  const detectMobile = () => {
    isMobile.value = window.innerWidth < 1024
  }

  // 全屏管理
  const requestFullscreen = () => {
    if (!isMobile.value) return
    const el = document.documentElement as any
    const rfs = el.requestFullscreen || el.webkitRequestFullscreen || el.mozRequestFullScreen || el.msRequestFullscreen
    if (rfs) {
      rfs.call(el).catch(() => {})
    }
  }

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

  return {
    // 状态
    showHints,
    showPlayers,
    showChat,
    showAdminModal,
    showInviteFriendsModal,
    showDeckDetailModal,
    showChemicalKeyboard,
    showQrModal,
    hasNewMessage,
    isMobile,

    // 计算属性
    isAnyModalOpen,

    // 方法
    closeAllModals,
    closeAllPanels,
    togglePanel,
    openModal,
    detectMobile,
    requestFullscreen,
    exitFullscreen,
  }
}
