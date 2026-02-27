/**
 * 格式化时间为"X小时前"或"X天前"的形式
 * @param time - ISO时间字符串或Date对象
 * @returns 格式化后的时间字符串
 */
function normalizeToDate(time: string | number | Date | null | undefined): Date | null {
  if (!time) return null
  if (time instanceof Date) return time

  if (typeof time === 'number') {
    const ms = time < 1e12 ? time * 1000 : time
    return new Date(ms)
  }

  if (typeof time === 'string') {
    const trimmed = time.trim()
    if (!trimmed) return null
    if (/^\d+$/.test(trimmed)) {
      const numeric = Number(trimmed)
      const ms = numeric < 1e12 ? numeric * 1000 : numeric
      return new Date(ms)
    }
    const withT = trimmed.replace(' ', 'T')
    const normalized = withT.replace(/([+-]\d{2})(\d{2})$/, '$1:$2')
    const parsed = new Date(normalized)
    if (!Number.isNaN(parsed.getTime())) return parsed
    const fallback = new Date(trimmed)
    if (!Number.isNaN(fallback.getTime())) return fallback
  }

  return null
}

export function formatLastOnline(time: string | number | Date | null | undefined): string {
  if (!time) return '从未上线'

  try {
    const lastOnlineTime = normalizeToDate(time)
    if (!lastOnlineTime || Number.isNaN(lastOnlineTime.getTime())) return '时间未知'
    const now = new Date()
    const diffMs = now.getTime() - lastOnlineTime.getTime()

    // 转换为秒
    const diffSeconds = Math.floor(diffMs / 1000)

    if (diffSeconds < 0) {
      return '刚刚在线'
    }

    // 小于1分钟
    if (diffSeconds < 60) {
      return '刚刚在线'
    }

    // 转换为分钟
    const diffMinutes = Math.floor(diffSeconds / 60)
    if (diffMinutes < 60) {
      return `${diffMinutes}分钟前`
    }

    // 转换为小时
    const diffHours = Math.floor(diffMinutes / 60)
    if (diffHours < 24) {
      return `${diffHours}小时前`
    }

    // 转换为天数
    const diffDays = Math.floor(diffHours / 24)
    if (diffDays < 7) {
      return `${diffDays}天前`
    }

    // 转换为周
    const diffWeeks = Math.floor(diffDays / 7)
    if (diffWeeks < 4) {
      return `${diffWeeks}周前`
    }

    // 转换为月
    const diffMonths = Math.floor(diffDays / 30)
    if (diffMonths < 12) {
      return `${diffMonths}月前`
    }

    // 超过一年
    const diffYears = Math.floor(diffDays / 365)
    return `${diffYears}年前`
  } catch (error) {
    console.error('时间格式化错误:', error)
    return '时间未知'
  }
}

/**
 * 获取在线状态的颜色和文字
 * @param isOnline - 是否在线
 * @param lastOfflineAt - 最后离线时间
 * @returns 状态信息
 */
export function getOnlineStatus(isOnline: boolean, lastOfflineAt: string | number | Date | null | undefined) {
  if (isOnline) {
    return {
      text: '在线',
      color: 'text-emerald-500',
      bgColor: 'bg-emerald-500/10',
      borderColor: 'border-emerald-500/20'
    }
  }

  const lastOnlineText = formatLastOnline(lastOfflineAt)

  return {
    text: lastOnlineText,
    color: 'text-slate-400',
    bgColor: 'bg-slate-100 dark:bg-white/5',
    borderColor: 'border-slate-200 dark:border-white/10'
  }
}

/**
 * 排行榜专用：显示距离上次下线的天数
 * 今日内显示“今日内”，<30天显示“X天前”，>=30天显示“30天前最后登录”
 */
export function formatLastOfflineForRanking(time: string | number | Date | null | undefined): string {
  if (!time) return '从未上线'

  try {
    const lastOfflineTime = normalizeToDate(time)
    if (!lastOfflineTime || Number.isNaN(lastOfflineTime.getTime())) return '时间未知'

    const now = new Date()
    const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate())
    const startOfLast = new Date(lastOfflineTime.getFullYear(), lastOfflineTime.getMonth(), lastOfflineTime.getDate())
    const diffDays = Math.floor((startOfToday.getTime() - startOfLast.getTime()) / (24 * 60 * 60 * 1000))

    if (diffDays <= 0) {
      return '今日内'
    }
    if (diffDays < 30) {
      return `${diffDays}天前`
    }
    return '30天前最后登录'
  } catch (error) {
    console.error('时间格式化错误:', error)
    return '时间未知'
  }
}
