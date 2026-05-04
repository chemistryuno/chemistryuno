export interface BanState {
  isBanned: boolean
  bannedUntil?: string
  banReason?: string
}

export const getStoredUser = (): any => {
  try {
    return JSON.parse(localStorage.getItem('user') || '{}')
  } catch {
    return {}
  }
}

export const getBanState = (user: any = getStoredUser(), now: Date = new Date()): BanState => {
  const bannedUntil = user?.banned_until || user?.bannedUntil
  if (!bannedUntil) {
    return { isBanned: false }
  }

  const until = new Date(bannedUntil)
  if (Number.isNaN(until.getTime()) || until <= now) {
    return { isBanned: false }
  }

  return {
    isBanned: true,
    bannedUntil,
    banReason: user?.ban_reason || user?.banReason || ''
  }
}

export const banStateFromBlockedMessage = (message: any): BanState => {
  const data = message?.data || {}
  return getBanState({
    banned_until: data.banned_until,
    ban_reason: data.ban_reason
  })
}

export const formatBanUntil = (bannedUntil?: string): string => {
  if (!bannedUntil) return '未提供'
  const until = new Date(bannedUntil)
  if (Number.isNaN(until.getTime())) return '未提供'
  return until.toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' })
}
