import { beforeEach, describe, expect, it, vi } from 'vitest'

const routerPush = vi.fn(() => Promise.resolve())
const axiosPost = vi.fn()
const requestUse = vi.fn()
const responseUse = vi.fn()

const apiInstance: any = vi.fn()
Object.assign(apiInstance, {
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
  interceptors: {
    request: { use: requestUse },
    response: { use: responseUse },
  },
})

vi.mock('../router', () => ({
  default: {
    push: routerPush,
  },
}))

vi.mock('axios', () => ({
  default: {
    create: vi.fn(() => apiInstance),
    post: axiosPost,
  },
}))

describe('API adapters', () => {
  beforeEach(() => {
    vi.resetModules()
    vi.clearAllMocks()
    localStorage.clear()
  })

  it('sends login payloads through the auth adapter', async () => {
    const { authAPI } = await import('./api')
    vi.mocked(apiInstance.post).mockResolvedValueOnce({ data: { user: { uid: 1 } } })

    await authAPI.login({ identifier: 'test', password: 'test123' })

    expect(apiInstance.post).toHaveBeenCalledWith('/auth/login', {
      identifier: 'test',
      password: 'test123',
    })
  })

  it('sends room creation payloads through the game adapter', async () => {
    const { gameAPI } = await import('./api')
    vi.mocked(apiInstance.post).mockResolvedValueOnce({ data: { id: 'room-1' } })

    await gameAPI.createRoom('QA Room', 4, 1, false, true, 'secret', false, 0, 0, true, 50, true, 5)

    expect(apiInstance.post).toHaveBeenCalledWith('/rooms', {
      name: 'QA Room',
      max_players: 4,
      deck_id: 1,
      is_points_mode: false,
      is_private: true,
      access_key: 'secret',
      is_pve: false,
      pve_difficulty: 0,
      ai_count: 0,
      enable_ai_backfill: true,
      ai_backfill_difficulty: 50,
      is_ranked: true,
      level_range: 5,
      tutorial_script: false,
    })
  })

  it('sends replay evidence metadata with feedback reports', async () => {
    const { authAPI } = await import('./api')
    vi.mocked(apiInstance.post).mockResolvedValueOnce({ data: { message: 'ok' } })

    await authAPI.submitFeedback('reported suspicious event', 'report', {
      room_id: 'room-1',
      reported_uid: 42,
      replay_anchor: {
        game_history_id: 77,
        event_index: 3,
        event_id: 'evt-3',
      },
    })

    expect(apiInstance.post).toHaveBeenCalledWith('/feedback', {
      content: 'reported suspicious event',
      type: 'report',
      room_id: 'room-1',
      reported_uid: 42,
      replay_anchor: {
        game_history_id: 77,
        event_index: 3,
        event_id: 'evt-3',
      },
    })
  })

  it('surfaces validation errors from admin API calls', async () => {
    const { adminAPI } = await import('./api')
    const validationError = { response: { status: 400, data: { error: 'invalid compensation' } } }
    vi.mocked(apiInstance.post).mockRejectedValueOnce(validationError)

    await expect(adminAPI.approveAppeal('appeal-1', { compensation_amount: 0 })).rejects.toEqual(validationError)
  })

  it('sends admin log attribution filters as query parameters', async () => {
    const { adminAPI } = await import('./api')
    vi.mocked(apiInstance.get).mockResolvedValueOnce({ data: { logs: [] } })

    await adminAPI.getLogs({
      count: 25,
      level: 'WARNING',
      uid: 100000101,
      source_ip: '203.0.113',
      category: 'request',
      status_class: '4xx',
      q: 'rooms',
    })

    expect(apiInstance.get).toHaveBeenCalledWith('/admin/logs?count=25&level=WARNING&uid=100000101&source_ip=203.0.113&category=request&status_class=4xx&q=rooms')
  })

  it('keeps network failures rejected for callers', async () => {
    const { authAPI } = await import('./api')
    const networkError = new Error('network down')
    vi.mocked(apiInstance.get).mockRejectedValueOnce(networkError)

    await expect(authAPI.getSessions()).rejects.toThrow('network down')
  })

  it('refreshes and replays an authenticated request after a 401 response', async () => {
    await import('./api')
    const rejected = responseUse.mock.calls[0][1]
    const originalRequest = { url: '/user/info' }
    const authError = { response: { status: 401 }, config: originalRequest }

    axiosPost.mockResolvedValueOnce({ data: {} })
    vi.mocked(apiInstance).mockResolvedValueOnce({ data: { ok: true } })

    const response = await rejected(authError)

    expect(axiosPost).toHaveBeenCalledWith('/api/auth/refresh', {}, { withCredentials: true })
    expect(apiInstance).toHaveBeenCalledWith({ ...originalRequest, _retry: true })
    expect(response).toEqual({ data: { ok: true } })
  })
})
