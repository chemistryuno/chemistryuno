import { flushPromises, mount } from '@vue/test-utils'
import type { AxiosResponse } from 'axios'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import Admin from './Admin.vue'
import { adminAPI } from '../utils/api'

const push = vi.fn()
const replace = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push, replace }),
  useRoute: () => ({ path: '/admin/logs', params: { tab: 'logs' } }),
}))

vi.mock('../utils/dialog', () => ({
  useDialog: () => ({
    showAlert: vi.fn(),
    showConfirm: vi.fn(),
    showPrompt: vi.fn(),
  }),
}))

vi.mock('../utils/api', () => ({
  adminAPI: {
    getStats: vi.fn(),
    getLogs: vi.fn(),
    clearLogs: vi.fn(),
  },
}))

const apiResponse = (data: any): AxiosResponse => ({
  data,
  status: 200,
  statusText: 'OK',
  headers: {},
  config: {} as any,
})

describe('Admin logs view', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.setItem('user', JSON.stringify({ role: 'admin' }))
    vi.mocked(adminAPI.getStats).mockResolvedValue(apiResponse({}))
    vi.mocked(adminAPI.getLogs).mockResolvedValue(apiResponse({
      logs: [
        {
          timestamp: '2026-05-06 16:31:15',
          level: 'INFO',
          category: 'request',
          message: 'GET /api/rooms -> 200',
          uid: 100000101,
          auth_state: 'authenticated',
          source: { client_ip: '127.0.0.1', user_agent: 'Vitest Agent' },
          request: { method: 'GET', path: '/api/rooms', status: 200, status_class: '2xx', latency_ms: 8 },
        },
        {
          timestamp: '2026-05-06 16:31:16',
          level: 'WARNING',
          category: 'request',
          message: 'GET /api/private -> 401',
          auth_state: 'anonymous',
          source: { client_ip: '203.0.113.20' },
          request: { method: 'GET', path: '/api/private', status: 401, status_class: '4xx', latency_ms: 3 },
        },
      ],
    }))
  })

  it('renders attributed and anonymous log rows', async () => {
    const wrapper = mount(Admin, {
      global: {
        stubs: {
          UserAvatar: true,
          RouterLink: true,
        },
      },
    })
    await flushPromises()

    expect(adminAPI.getLogs).toHaveBeenCalledWith(expect.objectContaining({
      count: 100,
      level: '',
    }))
    expect(wrapper.text()).toContain('100000101')
    expect(wrapper.text()).toContain('匿名')
    expect(wrapper.text()).toContain('127.0.0.1')
    expect(wrapper.text()).toContain('Vitest Agent')
    expect(wrapper.text()).toContain('GET /api/rooms 200 8ms')

    wrapper.unmount()
  })
})
