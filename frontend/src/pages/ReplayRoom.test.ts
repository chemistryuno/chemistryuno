import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ReplayRoom from './ReplayRoom.vue'
import { adminAPI, gameAPI } from '../utils/api'

const mocks = vi.hoisted(() => ({
  route: {
    params: { historyId: '77' },
    query: {
      scope: 'admin',
      from: '/admin/anticheat',
      event_index: '3',
      event_id: 'evt-3',
      timestamp_ms: '1710000002000',
      uid: '42',
    },
    fullPath: '/replay/77?scope=admin&from=/admin/anticheat&event_index=3&event_id=evt-3&timestamp_ms=1710000002000&uid=42',
  },
  routerPush: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRoute: () => mocks.route,
  useRouter: () => ({
    push: mocks.routerPush,
  }),
}))

vi.mock('../utils/api', () => ({
  adminAPI: {
    getGameReplay: vi.fn(),
  },
  gameAPI: {
    getMyGameReplay: vi.fn(),
  },
}))

vi.mock('../utils/dialog', () => ({
  useDialog: () => ({
    showAlert: vi.fn(),
    showPrompt: vi.fn(),
  }),
}))

const apiResponse = (data: any) => ({
  data,
  status: 200,
  statusText: 'OK',
  headers: {},
  config: {} as any,
})

const replayPayload = {
  room_id: 'room-1',
  replay: {
    events: [
      {
        type: 'game_start',
        event_index: 1,
        event_id: 'evt-1',
        at: '2026-05-05T00:00:00Z',
        unix_ms: 1710000000000,
      },
      {
        type: 'play_card',
        event_index: 2,
        event_id: 'evt-2',
        actor_uid: 42,
        at: '2026-05-05T00:00:01Z',
        unix_ms: 1710000001000,
        payload: {
          card_symbol: 'Na',
          substance: 'Sodium',
        },
      },
      {
        type: 'play_card',
        event_index: 3,
        event_id: 'evt-3',
        actor_uid: 42,
        at: '2026-05-05T00:00:02Z',
        unix_ms: 1710000002000,
        payload: {
          card_symbol: 'H2O',
          substance: 'Water',
        },
      },
    ],
  },
  player_profiles: [
    { uid: 42, nickname: 'Target Player' },
    { uid: 1001, nickname: 'Admin' },
  ],
}

describe('ReplayRoom', () => {
  beforeEach(() => {
    localStorage.clear()
    localStorage.setItem('user', JSON.stringify({ uid: 1001, role: 'admin' }))
    mocks.routerPush.mockClear()
    vi.mocked(adminAPI.getGameReplay).mockResolvedValue(apiResponse(replayPayload))
    vi.mocked(gameAPI.getMyGameReplay).mockResolvedValue(apiResponse(replayPayload))
    Element.prototype.scrollIntoView = vi.fn()
    vi.stubGlobal('requestAnimationFrame', (callback: FrameRequestCallback) => {
      callback(0)
      return 1
    })
  })

  it('highlights the requested replay anchor and starts game view at that event', async () => {
    const wrapper = mount(ReplayRoom)
    await flushPromises()

    expect(adminAPI.getGameReplay).toHaveBeenCalledWith(77)

    const focusedEvent = wrapper.findAll('div').find((element) => {
      return element.text().includes('H2O') && element.classes().includes('ring-2')
    })
    expect(focusedEvent?.exists()).toBe(true)

    await wrapper.findAll('button').find((button) => button.text().includes('游戏视角'))!.trigger('click')
    await flushPromises()

    const pushCalls = mocks.routerPush.mock.calls
    const pushed = String(pushCalls[pushCalls.length - 1]?.[0] || '')
    expect(pushed).toContain('/room/replay?')
    expect(pushed).toContain('replay_history_id=77')
    expect(pushed).toContain('replay_start_index=2')
    expect(pushed).toContain('event_index=3')
    expect(pushed).toContain('event_id=evt-3')
    expect(pushed).toContain('timestamp_ms=1710000002000')
    expect(pushed).toContain('uid=42')
    expect(pushed).toContain('scope=admin')
  })
})
