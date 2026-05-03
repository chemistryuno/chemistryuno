import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import PlayerAnticheatStatsWidget from './PlayerAnticheatStatsWidget.vue'
import { authAPI } from '../../utils/api'

vi.mock('../../utils/api', () => ({
  authAPI: {
    getPlayerAnticheatStats: vi.fn(),
  },
}))

const apiResponse = (data: any) => ({
  data,
  status: 200,
  statusText: 'OK',
  headers: {},
  config: {} as any,
})

describe('PlayerAnticheatStatsWidget', () => {
  it('loads player anticheat stats and refreshes every five minutes', async () => {
    vi.useFakeTimers()
    const getStats = vi.mocked(authAPI.getPlayerAnticheatStats)
    getStats
      .mockResolvedValueOnce(apiResponse({ bans_today: 3, system_uptime_days: 12 }))
      .mockResolvedValueOnce(apiResponse({ bans_today: 4, system_uptime_days: 12 }))

    const wrapper = mount(PlayerAnticheatStatsWidget)
    await flushPromises()

    expect(wrapper.text()).toContain('Bans Today')
    expect(wrapper.text()).toContain('System Running')
    expect(wrapper.text()).toContain('3')
    expect(wrapper.text()).toContain('12')

    await vi.advanceTimersByTimeAsync(5 * 60 * 1000)
    await flushPromises()

    expect(getStats).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('4')

    wrapper.unmount()
    vi.useRealTimers()
  })

  it('keeps the last successful stats when a later refresh is rate-limited', async () => {
    vi.useFakeTimers()
    const getStats = vi.mocked(authAPI.getPlayerAnticheatStats)
    getStats
      .mockResolvedValueOnce(apiResponse({ bans_today: 7, system_uptime_days: 21 }))
      .mockRejectedValueOnce({ response: { status: 429 } })

    const wrapper = mount(PlayerAnticheatStatsWidget)
    await flushPromises()

    await vi.advanceTimersByTimeAsync(5 * 60 * 1000)
    await flushPromises()

    expect(wrapper.text()).toContain('7')
    expect(wrapper.text()).toContain('21')
    expect(wrapper.text()).toContain('Cached')

    wrapper.unmount()
    vi.useRealTimers()
  })
})
