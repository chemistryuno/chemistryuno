/**
 * 格式化时间为"X小时前"或"X天前"的形式
 * @param time - ISO时间字符串或Date对象
 * @returns 格式化后的时间字符串
 */
export function formatLastOnline(time: string | Date | null | undefined): string {
  if (!time) return '从未上线'

  try {
    const lastOnlineTime = typeof time === 'string' ? new Date(time) : time
    const now = new Date()
    const diffMs = now.getTime() - lastOnlineTime.getTime()

    // 转换为秒
    const diffSeconds = Math.floor(diffMs / 1000)

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
export function getOnlineStatus(isOnline: boolean, lastOfflineAt: string | Date | null | undefined) {
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
