import { flushPromises, mount } from '@vue/test-utils'
import type { AxiosResponse } from 'axios'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import Admin from './Admin.vue'
import { adminAPI } from '../utils/api'

const push = vi.fn()
const replace = vi.fn()
const showConfirm = vi.fn()
const showAlert = vi.fn()
const showPrompt = vi.fn()

let eventSources: MockEventSource[] = []

class MockEventSource {
  url: string
  onopen: (() => void) | null = null
  onmessage: ((event: MessageEvent) => void) | null = null
  onerror: (() => void) | null = null
  close = vi.fn()

  constructor(url: string) {
    this.url = url
    eventSources.push(this)
  }

  emit(data: any) {
    this.onmessage?.({ data: JSON.stringify(data) } as MessageEvent)
  }
}

vi.mock('vue-router', () => ({
  useRouter: () => ({ push, replace }),
  useRoute: () => ({ path: '/admin/logs', params: { tab: 'logs' } }),
}))

vi.mock('../utils/dialog', () => ({
  useDialog: () => ({
    showAlert,
    showConfirm,
    showPrompt,
  }),
}))

vi.mock('../utils/api', () => ({
  adminAPI: {
    getStats: vi.fn(),
    getLogs: vi.fn(),
    getLogsStreamURL: vi.fn(),
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
    eventSources = []
    vi.clearAllMocks()
    vi.useRealTimers()
    vi.stubGlobal('EventSource', MockEventSource)
    localStorage.setItem('user', JSON.stringify({ role: 'admin' }))
    vi.mocked(adminAPI.getStats).mockResolvedValue(apiResponse({}))
    vi.mocked(adminAPI.getLogsStreamURL).mockImplementation((filters: any) => `/api/admin/logs/stream?${new URLSearchParams(filters).toString()}`)
    showAlert.mockResolvedValue(undefined)
    showConfirm.mockResolvedValue(true)
    showPrompt.mockResolvedValue('')
    vi.mocked(adminAPI.getLogs).mockResolvedValue(apiResponse({
      logs: [
        {
          sequence: 1,
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
          sequence: 2,
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

  it('opens a live stream and merges new log entries without duplicates', async () => {
    const wrapper = mount(Admin, {
      global: { stubs: { UserAvatar: true, RouterLink: true } },
    })
    await flushPromises()

    expect(eventSources).toHaveLength(1)
    expect(eventSources[0].url).toContain('/api/admin/logs/stream?')
    eventSources[0].onopen?.()
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('LIVE')

    eventSources[0].emit({
      sequence: 3,
      timestamp: '2026-05-06 16:31:17',
      level: 'ERROR',
      category: 'websocket',
      message: 'WS disconnected',
      uid: 100000102,
      websocket: { event: 'disconnect', type: 'close' },
    })
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('WS disconnected')

    eventSources[0].emit({
      sequence: 3,
      timestamp: '2026-05-06 16:31:17',
      level: 'ERROR',
      category: 'websocket',
      message: 'WS disconnected',
      uid: 100000102,
      websocket: { event: 'disconnect', type: 'close' },
    })
    await wrapper.vm.$nextTick()
    expect(wrapper.text().match(/WS disconnected/g)).toHaveLength(1)

    for (let i = 0; i < 101; i += 1) {
      eventSources[0].emit({
        sequence: 10 + i,
        timestamp: `2026-05-06 16:32:${String(i % 60).padStart(2, '0')}`,
        level: 'INFO',
        category: 'request',
        message: `bulk-log-${i}`,
        request: { method: 'GET', path: `/api/bulk/${i}`, status: 200 },
      })
    }
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).not.toContain('Vitest Agent')
    expect(wrapper.text()).not.toContain('bulk-log-0')
    expect(wrapper.text()).toContain('bulk-log-100')

    wrapper.unmount()
    expect(eventSources[0].close).toHaveBeenCalled()
  })

  it('keeps expanded log details open when streamed logs arrive', async () => {
    const wrapper = mount(Admin, {
      global: { stubs: { UserAvatar: true, RouterLink: true } },
    })
    await flushPromises()

    await wrapper.find('button.grid').trigger('click')
    expect(wrapper.text()).toContain('"client_ip": "127.0.0.1"')

    eventSources[0].emit({
      sequence: 4,
      timestamp: '2026-05-06 16:31:18',
      level: 'INFO',
      category: 'request',
      message: 'GET /api/new -> 200',
      request: { method: 'GET', path: '/api/new', status: 200 },
    })
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('"client_ip": "127.0.0.1"')
    expect(wrapper.text()).toContain('GET /api/new -> 200')
    wrapper.unmount()
  })

  it('reconnects the stream after debounced filter changes without clearing input', async () => {
    vi.useFakeTimers()
    const wrapper = mount(Admin, {
      global: { stubs: { UserAvatar: true, RouterLink: true } },
    })
    await flushPromises()

    const uidInput = wrapper.find('input[placeholder="UID"]')
    await uidInput.setValue('100000101')
    vi.advanceTimersByTime(350)
    await flushPromises()

    expect(adminAPI.getLogs).toHaveBeenLastCalledWith(expect.objectContaining({ uid: '100000101' }))
    expect(eventSources.length).toBeGreaterThanOrEqual(2)
    expect(eventSources[eventSources.length - 1]?.url).toContain('uid=100000101')
    expect((uidInput.element as HTMLInputElement).value).toBe('100000101')
    wrapper.unmount()
  })

  it('manual refresh reloads the current snapshot and keeps live updates active', async () => {
    const wrapper = mount(Admin, {
      global: { stubs: { UserAvatar: true, RouterLink: true } },
    })
    await flushPromises()

    const refreshButton = wrapper.findAll('button').find(button => button.text().includes('REFRESH'))
    expect(refreshButton).toBeTruthy()
    await refreshButton!.trigger('click')
    await flushPromises()

    expect(adminAPI.getLogs).toHaveBeenCalledTimes(2)
    expect(eventSources.length).toBeGreaterThanOrEqual(2)
    expect(eventSources[eventSources.length - 1]?.url).toContain('/api/admin/logs/stream')
    wrapper.unmount()
  })

  it('clears local logs and resumes live updates after successful clear', async () => {
    vi.mocked(adminAPI.clearLogs).mockResolvedValue(apiResponse({ message: 'ok' }))
    vi.mocked(adminAPI.getLogs).mockResolvedValueOnce(apiResponse({ logs: [] }))
    const wrapper = mount(Admin, {
      global: { stubs: { UserAvatar: true, RouterLink: true } },
    })
    await flushPromises()

    const clearButton = wrapper.findAll('button').find(button => button.text().includes('CLEAR'))
    expect(clearButton).toBeTruthy()
    await clearButton!.trigger('click')
    await flushPromises()

    expect(adminAPI.clearLogs).toHaveBeenCalled()
    expect(wrapper.text()).toContain('/ NO_LOGS_LOADED')
    expect(eventSources[eventSources.length - 1]?.url).toContain('/api/admin/logs/stream')
    wrapper.unmount()
  })
})
