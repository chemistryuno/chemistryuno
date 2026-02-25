/**
 * 管理员操作 Composable
 * 负责处理管理员的踢人、封禁等操作
 */

import { ref } from 'vue'
import { adminAPI } from '../utils/api'
import feedback from '../utils/feedback'

export function useAdminActions() {
  // 状态
  const adminTargetUser = ref<any>(null)
  const adminActionType = ref<'kick' | 'ban'>('kick')
  const banReason = ref('你由于违规游戏而被踢出')
  const banUntil = ref('')
  const selectedBanPreset = ref<number | null>(24)

  // 封禁时长预设（小时）
  const banDurationPresets = [
    { label: '1小时', hours: 1 },
    { label: '6小时', hours: 6 },
    { label: '24小时', hours: 24 },
    { label: '3天', hours: 72 },
    { label: '7天', hours: 168 },
    { label: '30天', hours: 720 },
    { label: '永久', hours: 8760 }, // 1年
  ]

  // 方法
  const setBanDuration = (hours: number) => {
    selectedBanPreset.value = hours
    const now = new Date()
    now.setHours(now.getHours() + hours)
    banUntil.value = now.toISOString().slice(0, 16) // YYYY-MM-DDTHH:mm
  }

  const getDefaultBanUntil = () => {
    const now = new Date()
    now.setHours(now.getHours() + 24)
    return now.toISOString().slice(0, 16)
  }

  const openAdminModal = (player: any, actionType: 'kick' | 'ban' = 'kick') => {
    adminTargetUser.value = player
    adminActionType.value = actionType
    banReason.value = actionType === 'kick'
      ? '你由于违规游戏而被踢出'
      : '你由于违规游戏而被封禁'
    selectedBanPreset.value = 24
    banUntil.value = getDefaultBanUntil()
  }

  const executeAdminAction = async () => {
    if (!adminTargetUser.value) {
      throw new Error('未选择目标用户')
    }

    if (adminActionType.value === 'kick') {
      await adminAPI.kickPlayer(adminTargetUser.value.uid, banReason.value)
      feedback.success()
      return { success: true, message: '该玩家已被强制下线并清除登录状态' }
    } else {
      if (!banUntil.value) {
        feedback.error()
        throw new Error('请选择封禁截止时间')
      }

      const until = new Date(banUntil.value)
      if (until <= new Date()) {
        feedback.error()
        throw new Error('封禁截止时间必须晚于当前时间')
      }

      await adminAPI.banUser(adminTargetUser.value.uid, until.toISOString(), banReason.value)
      feedback.success()
      return { success: true, message: '该玩家已被封禁' }
    }
  }

  const resetAdminState = () => {
    adminTargetUser.value = null
    adminActionType.value = 'kick'
    banReason.value = '你由于违规游戏而被踢出'
    selectedBanPreset.value = 24
    banUntil.value = getDefaultBanUntil()
  }

  return {
    // 状态
    adminTargetUser,
    adminActionType,
    banReason,
    banUntil,
    selectedBanPreset,
    banDurationPresets,

    // 方法
    setBanDuration,
    getDefaultBanUntil,
    openAdminModal,
    executeAdminAction,
    resetAdminState,
  }
}
