import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AdminAnticheat from './AdminAnticheat.vue'
import { adminAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'

vi.mock('../utils/api', () => ({
  adminAPI: {
    getDetectionList: vi.fn(),
    getDetectionDetail: vi.fn(),
    getAppealsList: vi.fn(),
    approveAppeal: vi.fn(),
    rejectAppeal: vi.fn(),
    banFromAnticheatPanel: vi.fn(),
    unbanFromAnticheatPanel: vi.fn(),
    changeDetectionPunishment: vi.fn(),
    getAnticheatConfig: vi.fn(),
    updateAnticheatConfig: vi.fn(),
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

const mountAdminAnticheat = async (detections: any[] = []) => {
  vi.mocked(adminAPI.getDetectionList).mockResolvedValue(apiResponse({ detections, total: detections.length }))
  vi.mocked(adminAPI.getDetectionDetail).mockResolvedValue(apiResponse({
    risk_score: {
      id: 7,
      player_uid: 42,
      room_id: 'room-1',
      risk_score: 86,
      response_time_score: 92,
      review_status: 'processed',
      punishment_decision: 'ban',
      replay_id: '77',
      game_history_id: 77,
      primary_evidence: {
        room_id: 'room-1',
        game_history_id: 77,
        replay_id: '77',
        event_index: 3,
        event_id: 'evt-3',
        event_type: 'play_card',
        player_uid: 42,
        event_timestamp_ms: 1710000000000,
        evidence_precision: 'operation',
        action_summary: 'played H2O',
      },
      indicator_details: [{
        name: 'response_time',
        raw_value: 12,
        normalized_score: 90,
        weight: 0.25,
        contribution: 22.5,
        explanation: 'fast operation',
        evidence_anchors: [{
          game_history_id: 77,
          event_index: 3,
          event_id: 'evt-3',
          event_type: 'play_card',
          player_uid: 42,
          evidence_precision: 'operation',
        }],
      }],
      report_contribution: {
        deduplicated_count: 1,
        weight: 0.1,
        contribution: 6,
        source_summary: 'player report',
        evidence_anchors: [{
          game_history_id: 77,
          event_index: 3,
          event_id: 'evt-3',
          event_type: 'report',
          player_uid: 42,
          evidence_precision: 'operation',
        }],
      },
    },
    sanctions: [{ id: 88, sanction_type: 'ban' }],
  }))
  vi.mocked(adminAPI.getAppealsList).mockResolvedValue(apiResponse({
      appeals: [
        {
          id: 'appeal-1',
          player_uid: 42,
          room_id: 'room-1',
          reason: '误封申诉',
          status: 'pending',
          primary_evidence: {
            game_history_id: 77,
            event_index: 3,
            event_id: 'evt-3',
            evidence_precision: 'operation',
          },
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
          primary_evidence: {
            game_history_id: 77,
            event_index: 3,
            event_id: 'evt-3',
            evidence_precision: 'operation',
          },
          compensation_status: 'ok',
          compensation_amount: 100,
          created_at: '2026-05-03T00:05:00Z',
        },
      ],
      total: 1,
    }))
  vi.mocked(adminAPI.approveAppeal).mockResolvedValue(apiResponse({ compensation_status: 'ok' }))
  vi.mocked(adminAPI.banFromAnticheatPanel).mockResolvedValue(apiResponse({ message: 'player banned' }))
  vi.mocked(adminAPI.unbanFromAnticheatPanel).mockResolvedValue(apiResponse({ message: 'player unbanned' }))
  vi.mocked(adminAPI.changeDetectionPunishment).mockResolvedValue(apiResponse({ message: 'updated' }))

  const wrapper = mount(AdminAnticheat, {
    global: {
      stubs: {
        RouterLink: {
          props: ['to'],
          template: '<a :href="String(to)"><slot /></a>',
        },
      },
    },
  })
  await flushPromises()
  return wrapper
}

describe('AdminAnticheat', () => {
  it('walks through appeal approval and shows audit visibility', async () => {
    const wrapper = await mountAdminAnticheat()

    await wrapper.findAll('button').find(button => button.text().includes('申诉管理'))!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('误封申诉')
    expect(wrapper.text()).toContain('42')

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

  it('validates and submits manual ban from detection detail', async () => {
    const wrapper = await mountAdminAnticheat([{
      id: 7,
      player_id: 42,
      room_id: 'room-1',
      risk_score: 86,
      sanction_type: 'ban',
      created_at: '2026-05-03T00:00:00Z',
    }])
    await wrapper.findAll('button').find(button => button.text().includes('查看'))!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('封禁处置')

    const untilInput = wrapper.find('input[type="datetime-local"]')
    await untilInput.setValue('2026-05-07T10:00')
    const textareas = wrapper.findAll('textarea')
    await textareas[textareas.length - 1].setValue('manual evidence review')

    await wrapper.findAll('button').find(button => button.text().includes('执行封禁'))!.trigger('click')
    await flushPromises()

    expect(adminAPI.banFromAnticheatPanel).toHaveBeenCalledWith({
      player_uid: 42,
      banned_until: new Date('2026-05-07T10:00').toISOString(),
      reason: 'manual evidence review',
      room_id: 'room-1',
      risk_score_id: 7,
    })
  })

  it('shows suspicious point replay evidence and submits punishment changes', async () => {
    const wrapper = await mountAdminAnticheat([{
      id: 7,
      player_id: 42,
      room_id: 'room-1',
      risk_score: 86,
      sanction_type: 'ban',
      review_status: 'processed',
      created_at: '2026-05-03T00:00:00Z',
    }])

    await wrapper.findAll('button').find(button => button.text().includes('查看'))!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Replay Evidence')
    expect(wrapper.text()).toContain('Open replay point')
    expect(wrapper.text()).toContain('response_time')
    expect(wrapper.text()).toContain('player report')
    expect(wrapper.html()).toContain('/replay/77?scope=admin')
    expect(wrapper.html()).toContain('event_index=3')
    expect(wrapper.html()).toContain('event_id=evt-3')

    const selects = wrapper.findAll('select')
    await selects[selects.length - 1].setValue('mute')
    await wrapper.findAll('button').find(button => button.text().includes('保存处罚决定'))!.trigger('click')
    await flushPromises()

    expect(adminAPI.changeDetectionPunishment).toHaveBeenCalledWith(7, {
      punishment_decision: 'mute',
      sanction_id: 88,
      reason: 'Anticheat panel manual enforcement',
    })
  })

  it('hides replay links when evidence has no corresponding replay', async () => {
    const wrapper = await mountAdminAnticheat([{
      id: 7,
      player_id: 42,
      room_id: 'room-1',
      risk_score: 86,
      sanction_type: 'ban',
      review_status: 'processed',
      created_at: '2026-05-03T00:00:00Z',
    }])
    vi.mocked(adminAPI.getDetectionDetail).mockResolvedValueOnce(apiResponse({
      risk_score: {
        id: 7,
        player_uid: 42,
        room_id: 'room-1',
        risk_score: 86,
        review_status: 'processed',
        punishment_decision: 'ban',
        replay_id: '',
        game_history_id: 0,
        has_replay: false,
        primary_evidence: {
          room_id: 'room-1',
          has_replay: false,
          replay_available: false,
          evidence_precision: 'room',
          action_summary: 'room-level evidence without stored replay',
        },
        related_evidence: [{
          room_id: 'room-1',
          has_replay: false,
          replay_available: false,
          evidence_precision: 'room',
        }],
        indicator_details: [{
          name: 'response_time',
          raw_value: 12,
          normalized_score: 90,
          weight: 0.25,
          contribution: 22.5,
          explanation: 'fast operation',
          evidence_anchors: [{
            room_id: 'room-1',
            has_replay: false,
            replay_available: false,
            evidence_precision: 'room',
          }],
        }],
        report_contribution: {
          deduplicated_count: 1,
          weight: 0.1,
          contribution: 6,
          source_summary: 'player report',
          evidence_anchors: [{
            room_id: 'room-1',
            has_replay: false,
            replay_available: false,
            evidence_precision: 'room',
          }],
        },
      },
      sanctions: [{ id: 88, sanction_type: 'ban' }],
    }))

    await wrapper.find('tbody button').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Replay Evidence')
    expect(wrapper.text()).not.toContain('Open replay point')
    expect(wrapper.html()).not.toContain('/replay/')
  })

  it('shows backend rejection when a processed punishment cancellation is attempted', async () => {
    vi.mocked(adminAPI.changeDetectionPunishment).mockRejectedValueOnce({
      response: { data: { error: 'processed punishment cannot be cancelled' } },
    })
    const wrapper = await mountAdminAnticheat([{
      id: 7,
      player_id: 42,
      room_id: 'room-1',
      risk_score: 86,
      sanction_type: 'ban',
      review_status: 'processed',
      created_at: '2026-05-03T00:00:00Z',
    }])

    await wrapper.find('tbody button').trigger('click')
    await flushPromises()

    const selects = wrapper.findAll('select')
    await selects[selects.length - 1].setValue('observe')
    await wrapper.findAll('button').find(button => button.text().includes('保存') || button.text().includes('澶勭綒'))!.trigger('click')
    await flushPromises()

    expect(adminAPI.changeDetectionPunishment).toHaveBeenCalledWith(7, {
      punishment_decision: 'observe',
      sanction_id: 88,
      reason: 'Anticheat panel manual enforcement',
    })
    expect(showAlert).toHaveBeenCalledWith('processed punishment cannot be cancelled', '操作失败')
  })
})
