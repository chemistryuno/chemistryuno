import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AdminAnticheat from './AdminAnticheat.vue'
import { adminAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'

vi.mock('../utils/api', () => ({
  adminAPI: {
    getDetectionList: vi.fn(),
    getAppealsList: vi.fn(),
    approveAppeal: vi.fn(),
    getAnticheatConfig: vi.fn(),
    getAuditLog: vi.fn(),
    exportAuditLog: vi.fn(),
  },
}))

const showAlert = vi.fn()

vi.mock('../utils/dialog', () => ({
  useDialog: vi.fn(() => ({
    showAlert,
    showPrompt: vi.fn(),
  })),
}))

const defaultConfig = {
  dimensions: {
    response_time: { weight: 0.25, threshold: 100 },
    frequency: { weight: 0.25, threshold: 5 },
    win_rate: { weight: 0.2, threshold: 0.9 },
    pattern: { weight: 0.15, threshold: 0.7 },
    account_age: { weight: 0.15, threshold: 7 },
  },
  sanctions: {
    observe: 20,
    warning: 40,
    mute: 60,
    ban: 80,
  },
  unban: {
    compensation_amount: 100,
    default_message: 'Default compensation message',
  },
}

const apiResponse = (data: any) => ({
  data,
  status: 200,
  statusText: 'OK',
  headers: {},
  config: {} as any,
})

const mountAdminAnticheat = async () => {
  vi.mocked(adminAPI.getDetectionList).mockResolvedValue(apiResponse({ detections: [], total: 0 }))
  vi.mocked(adminAPI.getAppealsList).mockResolvedValue(apiResponse({
      appeals: [
        {
          id: 'appeal-1',
          player_id: 42,
          room_id: 'room-1',
          reason: '误封申诉',
          status: 'pending',
          created_at: '2026-05-03T00:00:00Z',
        },
      ],
      total: 1,
    }))
  vi.mocked(adminAPI.getAnticheatConfig).mockResolvedValue(apiResponse(defaultConfig))
  vi.mocked(adminAPI.getAuditLog).mockResolvedValue(apiResponse({
      logs: [
        {
          id: 'audit-1',
          player_id: 42,
          action_type: 'unban',
          details: 'approved',
          compensation_status: 'ok',
          compensation_amount: 100,
          created_at: '2026-05-03T00:05:00Z',
        },
      ],
      total: 1,
    }))
  vi.mocked(adminAPI.approveAppeal).mockResolvedValue(apiResponse({ compensation_status: 'ok' }))

  const wrapper = mount(AdminAnticheat)
  await flushPromises()
  return wrapper
}

describe('AdminAnticheat', () => {
  it('walks through appeal approval and shows audit visibility', async () => {
    const wrapper = await mountAdminAnticheat()

    await wrapper.findAll('button').find(button => button.text().includes('申诉管理'))!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('误封申诉')

    await wrapper.findAll('button').find(button => button.text().includes('批准'))!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('批准申诉并发放补偿')
    expect(wrapper.text()).toContain('100')

    const amountInput = wrapper.find('input[type="number"]')
    await amountInput.setValue(120)
    const textareas = wrapper.findAll('textarea')
    await textareas[0].setValue('Custom player compensation message')
    await textareas[textareas.length - 1].setValue('Reviewed clean replay')

    await wrapper.findAll('button').find(button => button.text().includes('确认批准'))!.trigger('click')
    await flushPromises()

    expect(adminAPI.approveAppeal).toHaveBeenCalledWith('appeal-1', {
      note: 'Reviewed clean replay',
      compensation_amount: 120,
      compensation_message: 'Custom player compensation message',
    })
    expect(showAlert).toHaveBeenCalledWith('申诉已批准，补偿已发放', '成功')

    await wrapper.findAll('button').find(button => button.text().includes('审计日志'))!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('已发放')
    expect(wrapper.text()).toContain('100燃素')
    expect(adminAPI.getAuditLog).toHaveBeenCalled()
    expect(useDialog).toHaveBeenCalled()
  })

  it('shows configurable compensation defaults in the config tab', async () => {
    const wrapper = await mountAdminAnticheat()

    await wrapper.findAll('button').find(button => button.text().includes('配置管理'))!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('解封补偿')
    expect(wrapper.text()).toContain('100 燃素')
    expect(wrapper.text()).toContain('Default compensation message')
  })
})
