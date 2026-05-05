import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import Appeals from './Appeals.vue'
import { authAPI } from '../utils/api'

vi.mock('../utils/api', () => ({
  authAPI: {
    getUserInfo: vi.fn(),
    refreshUserInfo: vi.fn(),
    getPlayerAppeals: vi.fn(),
    getPlayerSanctions: vi.fn(),
    submitAppeal: vi.fn(),
  },
}))

const apiResponse = (data: any) => ({
  data,
  status: 200,
  statusText: 'OK',
  headers: {},
  config: {} as any,
})

const mountAppeals = async () => {
  const wrapper = mount(Appeals, {
    global: {
      stubs: {
        RouterLink: {
          template: '<a><slot /></a>',
        },
      },
    },
  })
  await flushPromises()
  return wrapper
}

describe('Appeals', () => {
  beforeEach(() => {
    vi.mocked(authAPI.getUserInfo).mockResolvedValue(apiResponse({ uid: 1001, banned_until: null, ban_reason: '' }))
    vi.mocked(authAPI.refreshUserInfo).mockResolvedValue(apiResponse({ uid: 1001, banned_until: null, ban_reason: '', points: 1100, fuel: 100 }))
  })

  it('submits a valid appeal and refreshes history', async () => {
    vi.mocked(authAPI.getPlayerAppeals)
      .mockResolvedValueOnce(apiResponse({ appeals: [], total: 0 }))
      .mockResolvedValueOnce(apiResponse({
        appeals: [{ id: 1, room_id: 'room-7', reason: '误判说明', status: 'pending', submitted_at: '2026-05-05T00:00:00Z' }],
        total: 1,
      }))
    vi.mocked(authAPI.getPlayerSanctions).mockResolvedValue(apiResponse({ sanctions: [{ id: 9, room_id: 'room-7', risk_score_id: 3 }] }))
    vi.mocked(authAPI.submitAppeal).mockResolvedValue(apiResponse({ appeal: { id: 1 } }))

    const wrapper = await mountAppeals()
    const textareas = wrapper.findAll('textarea')
    await textareas[0].setValue('这次异常检测是误判')
    await textareas[1].setValue('网络延迟导致了异常时间点')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(authAPI.submitAppeal).toHaveBeenCalledWith('room-7', {
      risk_score_id: 3,
      sanction_id: 9,
      reason: '这次异常检测是误判',
      evidence: '网络延迟导致了异常时间点',
    })
    expect(wrapper.text()).toContain('待审核')
  })

  it('prevents duplicate pending appeals', async () => {
    vi.mocked(authAPI.getPlayerAppeals).mockResolvedValue(apiResponse({
      appeals: [{ id: 1, room_id: 'room-1', reason: '处理中', status: 'pending' }],
      total: 1,
    }))
    vi.mocked(authAPI.getPlayerSanctions).mockResolvedValue(apiResponse({ sanctions: [] }))

    const wrapper = await mountAppeals()
    await wrapper.findAll('textarea')[0].setValue('新的申诉')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(authAPI.submitAppeal).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('已有待处理申诉')
  })

  it('shows approved outcome and compensation details', async () => {
    localStorage.setItem('user', JSON.stringify({
      uid: 1001,
      banned_until: '2026-05-06T09:42:00+08:00',
      ban_reason: 'legacy ban',
    }))
    vi.mocked(authAPI.getPlayerAppeals).mockResolvedValue(apiResponse({
      appeals: [{
        id: 2,
        reason: '误封',
        status: 'approved',
        review_remark: '已核实',
        compensation_amount: 120,
        compensation_status: 'ok',
        compensation_note: '补偿成功',
      }],
      total: 1,
    }))
    vi.mocked(authAPI.getPlayerSanctions).mockResolvedValue(apiResponse({ sanctions: [] }))
    vi.mocked(authAPI.refreshUserInfo).mockResolvedValue(apiResponse({
      uid: 1001,
      banned_until: null,
      ban_reason: '',
      points: 1100,
      fuel: 100,
    }))

    const wrapper = await mountAppeals()

    expect(wrapper.text()).toContain('已通过')
    expect(wrapper.text()).toContain('120')
    expect(wrapper.text()).not.toContain('legacy ban')
    expect(wrapper.text()).toContain('补偿成功')
  })

  it('shows failed-load retry state', async () => {
    vi.mocked(authAPI.getPlayerAppeals)
      .mockRejectedValueOnce({ response: { data: { error: 'load failed' } } })
      .mockResolvedValueOnce(apiResponse({ appeals: [], total: 0 }))
    vi.mocked(authAPI.getPlayerSanctions).mockResolvedValue(apiResponse({ sanctions: [] }))

    const wrapper = await mountAppeals()
    expect(wrapper.text()).toContain('load failed')

    await wrapper.findAll('button').find(button => button.text().includes('重试'))!.trigger('click')
    await flushPromises()

    expect(authAPI.getPlayerAppeals).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('暂无申诉记录')
  })
})
