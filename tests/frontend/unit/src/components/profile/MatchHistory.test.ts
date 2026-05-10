import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import MatchHistory from '@/components/profile/MatchHistory.vue'
import { gameAPI } from '@/utils/api'

vi.mock('vue-router', () => ({
  useRoute: () => ({ path: '/profile/history' }),
  useRouter: () => ({ push: vi.fn() }),
}))

vi.mock('@/utils/api', () => ({
  gameAPI: {
    getMyGameHistory: vi.fn(),
  },
}))

const apiResponse = (data: any) => ({
  data,
  status: 200,
  statusText: 'OK',
  headers: {},
  config: {} as any,
})

describe('MatchHistory', () => {
  it('does not render player-facing cheat or permanent replay labels', async () => {
    vi.mocked(gameAPI.getMyGameHistory).mockResolvedValue(apiResponse([
      {
        id: 1,
        room_id: 'room_privacy',
        winner_name: 'Player',
        is_invalid: false,
        has_replay: true,
        replay_expires_at: '2026-05-10T00:00:00Z',
        replay_permanent: true,
        cheat_detected: true,
        cheat_uids: [1001],
        players: [1001, 1002],
        finished_at: '2026-05-03T00:00:00Z',
      },
    ]))

    const wrapper = mount(MatchHistory)
    await flushPromises()

    expect(wrapper.text()).not.toContain('CHEAT')
    expect(wrapper.text()).not.toContain('PERMANENT')
    expect(wrapper.html()).not.toContain('game.cheat_detected')
    expect(wrapper.html()).not.toContain('game.replay_permanent')
  })
})
