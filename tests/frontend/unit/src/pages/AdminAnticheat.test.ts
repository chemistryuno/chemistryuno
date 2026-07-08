import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import AdminAnticheat from '@/pages/AdminAnticheat.vue'
import { adminAPI } from '@/utils/api'
import { useDialog } from '@/utils/dialog'

vi.mock('@/utils/api', () => ({
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

vi.mock('@/utils/dialog', () => ({
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
    attachTo: document.body,
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

// 检测详情弹窗通过 Teleport 挂到 body，wrapper.text()/find 无法捕获，
// 需直接查询 document.body。
const detailModalText = () => document.body.textContent || ''
const detailModalHtml = () => document.body.innerHTML || ''
const findModalButton = (keyword: string): HTMLButtonElement | undefined =>
  Array.from(document.body.querySelectorAll('button')).find(b => (b.textContent || '').includes(keyword)) as HTMLButtonElement | undefined
const modalSelects = (): HTMLSelectElement[] =>
  Array.from(document.body.querySelectorAll('select')) as HTMLSelectElement[]

// 检测列表现为「房间 → 玩家 → 检测」可折叠结构，需依次展开后点击「详情」打开检测详情。
// 展开层级由最内层的可点击表头触发（room header / player header 各带一个 @click 切换），
// 因此用「最后一个包含对应文案的 cursor-pointer 元素」定位，避免命中外层容器。
const clickDeepest = async (wrapper: any, keyword: string) => {
  const candidates = wrapper.findAll('[class*="cursor-pointer"]').filter((el: any) => el.text().includes(keyword))
  const target = candidates[candidates.length - 1]
  if (target) {
    await target.trigger('click')
    await flushPromises()
  }
  return target
}

const openFirstDetectionDetail = async (wrapper: any) => {
  await clickDeepest(wrapper, '房间')
  await clickDeepest(wrapper, '玩家')

  const detailButton = wrapper.findAll('button').find((button: any) => button.text().includes('详情'))
  if (!detailButton) {
    throw new Error('未找到检测详情按钮，检测列表结构可能已变化')
  }
  await detailButton.trigger('click')
  await flushPromises()
}

afterEach(() => {
  // 清理 attachTo/Teleport 遗留在 body 上的 DOM，避免跨用例污染
  document.body.innerHTML = ''
})

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
    // 无回放导航证据时，详情弹窗才会展示「封禁截止时间」输入
    vi.mocked(adminAPI.getDetectionDetail).mockResolvedValueOnce(apiResponse({
      risk_score: {
        id: 7,
        player_uid: 42,
        room_id: 'room-1',
        risk_score: 86,
        review_status: 'pending',
        primary_evidence: null,
        replay_navigation: null,
        related_evidence: [],
        indicator_details: [],
      },
      sanctions: [{ id: 88, sanction_type: 'ban' }],
    }))
    const wrapper = await mountAdminAnticheat([{
      id: 7,
      player_id: 42,
      room_id: 'room-1',
      risk_score: 86,
      sanction_type: 'ban',
      created_at: '2026-05-03T00:00:00Z',
    }])
    await openFirstDetectionDetail(wrapper)

    expect(detailModalText()).toContain('执行封禁')

    // 封禁截止时间必须晚于当前时间，故取相对于「现在」的未来时间，避免用例随日期失效
    const future = new Date(Date.now() + 4 * 24 * 60 * 60 * 1000)
    const pad = (n: number) => String(n).padStart(2, '0')
    const futureLocal = `${future.getFullYear()}-${pad(future.getMonth() + 1)}-${pad(future.getDate())}T${pad(future.getHours())}:${pad(future.getMinutes())}`

    const untilInput = document.body.querySelector('input[type="datetime-local"]') as HTMLInputElement
    untilInput.value = futureLocal
    untilInput.dispatchEvent(new Event('input', { bubbles: true }))
    const textareas = Array.from(document.body.querySelectorAll('textarea')) as HTMLTextAreaElement[]
    const reasonArea = textareas[0]
    reasonArea.value = 'manual evidence review'
    reasonArea.dispatchEvent(new Event('input', { bubbles: true }))
    await flushPromises()

    findModalButton('执行封禁')!.click()
    await flushPromises()

    expect(adminAPI.banFromAnticheatPanel).toHaveBeenCalledWith({
      player_uid: 42,
      banned_until: new Date(futureLocal).toISOString(),
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

    await openFirstDetectionDetail(wrapper)

    expect(detailModalText()).toContain('回放证据')
    expect(detailModalText()).toContain('打开回放')
    expect(detailModalText()).toContain('响应时间') // response_time 指标的中文名
    expect(detailModalHtml()).toContain('/replay/77?scope=admin')
    expect(detailModalHtml()).toContain('event_index=3')
    expect(detailModalHtml()).toContain('event_id=evt-3')

    const selects = modalSelects()
    const decisionSelect = selects[selects.length - 1]
    decisionSelect.value = 'mute'
    decisionSelect.dispatchEvent(new Event('change', { bubbles: true }))
    await flushPromises()

    findModalButton('保存')!.click()
    await flushPromises()

    expect(adminAPI.changeDetectionPunishment).toHaveBeenCalledWith(7, {
      punishment_decision: 'mute',
      sanction_id: 88,
      reason: '反作弊面板人工处置',
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

    await openFirstDetectionDetail(wrapper)

    expect(detailModalText()).toContain('回放证据')
    // 无对应回放时不应出现「打开回放」链接或 /replay/ 路由
    expect(detailModalText()).not.toContain('打开回放')
    expect(detailModalHtml()).not.toContain('/replay/')
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

    await openFirstDetectionDetail(wrapper)

    const selects = modalSelects()
    const decisionSelect = selects[selects.length - 1]
    decisionSelect.value = 'observe'
    decisionSelect.dispatchEvent(new Event('change', { bubbles: true }))
    await flushPromises()

    findModalButton('保存')!.click()
    await flushPromises()

    expect(adminAPI.changeDetectionPunishment).toHaveBeenCalledWith(7, {
      punishment_decision: 'observe',
      sanction_id: 88,
      reason: '反作弊面板人工处置',
    })
    expect(showAlert).toHaveBeenCalledWith('processed punishment cannot be cancelled', '操作失败')
  })
})
