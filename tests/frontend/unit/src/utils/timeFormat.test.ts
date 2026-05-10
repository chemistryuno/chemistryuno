import { beforeEach, describe, expect, it, vi } from 'vitest'
import { formatLastOfflineForRanking, formatLastOnline, getOnlineStatus } from '@/utils/timeFormat'

describe('time format utilities', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-05-03T12:00:00Z'))
  })

  it('formats missing and invalid last-online values defensively', () => {
    expect(formatLastOnline(null)).toBe('从未上线')
    expect(formatLastOnline('not-a-date')).toBe('时间未知')
  })

  it('formats recent, hourly, daily, weekly, monthly, and yearly distances', () => {
    expect(formatLastOnline('2026-05-03T11:59:31Z')).toBe('刚刚在线')
    expect(formatLastOnline('2026-05-03T11:30:00Z')).toBe('30分钟前')
    expect(formatLastOnline('2026-05-03T09:00:00Z')).toBe('3小时前')
    expect(formatLastOnline('2026-05-01T12:00:00Z')).toBe('2天前')
    expect(formatLastOnline('2026-04-12T12:00:00Z')).toBe('3周前')
    expect(formatLastOnline('2026-02-02T12:00:00Z')).toBe('3月前')
    expect(formatLastOnline('2024-05-03T12:00:00Z')).toBe('2年前')
  })

  it('normalizes numeric seconds, milliseconds, and compact timezone strings', () => {
    expect(formatLastOnline(1777806000)).toBe('1小时前')
    expect(formatLastOnline(1777806000000)).toBe('1小时前')
    expect(formatLastOnline('2026-05-03 19:00:00+0800')).toBe('1小时前')
  })

  it('returns online and offline badge metadata', () => {
    expect(getOnlineStatus(true, null)).toMatchObject({
      text: '在线',
      color: 'text-emerald-500',
    })
    expect(getOnlineStatus(false, '2026-05-03T10:00:00Z')).toMatchObject({
      text: '2小时前',
      color: 'text-slate-400',
    })
  })

  it('formats ranking offline values by day buckets', () => {
    expect(formatLastOfflineForRanking(null)).toBe('从未上线')
    expect(formatLastOfflineForRanking('bad')).toBe('时间未知')
    expect(formatLastOfflineForRanking('2026-05-03T01:00:00Z')).toBe('今日内')
    expect(formatLastOfflineForRanking('2026-04-30T23:59:59Z')).toBe('2天前')
    expect(formatLastOfflineForRanking('2026-03-01T00:00:00Z')).toBe('30天前最后登录')
  })
})
