import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import Appeals from '@/pages/Appeals.vue'
import { adminAPI, authAPI } from '@/utils/api'

const routerPush = vi.hoisted(() => vi.fn())
const showConfirm = vi.hoisted(() => vi.fn())

vi.mock('@/utils/api', () => ({
  authAPI: {
    getUserInfo: vi.fn(),
    refreshUserInfo: vi.fn(),
    getPlayerAppeals: vi.fn(),
    getAppealEntryStatus: vi.fn(),
    getPlayerSanctions: vi.fn(),
    submitAppeal: vi.fn(),
    claimAppealCompensation: vi.fn(),
  },
  adminAPI: {
    unbanUser: vi.fn(),
    unbanFromAnticheatPanel: vi.fn(),
    approveAppeal: vi.fn(),
  },
}))

vi.mock('@/utils/dialog', () => ({
  useDialog: () => ({
    showConfirm,
  }),
}))

vi.mock('vue-router', () => ({
  RouterLink: {
    props: ['to'],
    template: '<a><slot /></a>',
  },
  useRouter: () => ({
    push: routerPush,
  }),
}))

const apiResponse = (data: any) => ({
  data,
  status: 200,
  statusText: 'OK',
  headers: {},
  config: {} as any,
})

const futureIso = () => new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString()

const mountAppeals = async () => {
  const wrapper = mount(Appeals)
  await flushPromises()
  return wrapper
}

describe('Appeals', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    routerPush.mockClear()
    showConfirm.mockResolvedValue(true)
    vi.mocked(authAPI.getUserInfo).mockResolvedValue(apiResponse({ uid: 1001, banned_until: null, ban_reason: '' }))
    vi.mocked(authAPI.refreshUserInfo).mockResolvedValue(apiResponse({ uid: 1001, banned_until: null, ban_reason: '', points: 1100, fuel: 100 }))
    vi.mocked(authAPI.getAppealEntryStatus).mockResolvedValue(apiResponse({
      is_banned: true,
      can_submit: true,
      latest_risk_score_id: 3,
      first_room_id: 'room-7',
      room_ids: ['room-7'],
    }))
  })

  it('submits a valid appeal and refreshes history', async () => {
    vi.mocked(authAPI.getPlayerAppeals)
      .mockResolvedValueOnce(apiResponse({ appeals: [], total: 0 }))
      .mockResolvedValueOnce(apiResponse({
        appeals: [{ id: 1, room_id: 'room-7', reason: 'appeal reason', status: 'pending', submitted_at: '2026-05-05T00:00:00Z' }],
        total: 1,
      }))
    vi.mocked(authAPI.getPlayerSanctions).mockResolvedValue(apiResponse({ sanctions: [{ id: 9, room_id: 'room-7', risk_score_id: 3 }] }))
    vi.mocked(authAPI.submitAppeal).mockResolvedValue(apiResponse({ appeal: { id: 1 } }))

    const wrapper = await mountAppeals()
    const textareas = wrapper.findAll('textarea')
    await textareas[0].setValue('false positive appeal')
    await textareas[1].setValue('network jitter at replay point')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(authAPI.submitAppeal).toHaveBeenCalledWith('room-7', {
      risk_score_id: 3,
      sanction_id: 9,
      reason: 'false positive appeal',
      evidence: 'network jitter at replay point',
    })
    expect(wrapper.text()).toContain('appeal reason')
  })

  it('prevents duplicate pending appeals', async () => {
    vi.mocked(authAPI.getPlayerAppeals).mockResolvedValue(apiResponse({
      appeals: [{ id: 1, room_id: 'room-1', reason: 'already pending', status: 'pending' }],
      total: 1,
    }))
    vi.mocked(authAPI.getPlayerSanctions).mockResolvedValue(apiResponse({ sanctions: [] }))

    const wrapper = await mountAppeals()
    await wrapper.findAll('textarea')[0].setValue('new appeal')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(authAPI.submitAppeal).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('already pending')
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
        reason: 'accepted appeal',
        status: 'approved',
        review_remark: 'accepted',
        compensation_amount: 120,
        compensation_status: 'ok',
        compensation_note: 'compensation completed',
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

    expect(wrapper.text()).toContain('accepted appeal')
    expect(wrapper.text()).toContain('120')
    expect(wrapper.text()).not.toContain('legacy ban')
    expect(wrapper.text()).toContain('compensation completed')
  })

  it('does not auto-claim compensation or call admin unban APIs on load', async () => {
    vi.mocked(authAPI.getPlayerAppeals).mockResolvedValue(apiResponse({
      appeals: [{
        id: 8,
        reason: 'accepted appeal',
        status: 'approved',
        review_remark: 'accepted',
        compensation_amount: 120,
        compensation_status: 'pending',
      }],
      total: 1,
    }))
    vi.mocked(authAPI.getPlayerSanctions).mockResolvedValue(apiResponse({ sanctions: [] }))

    const wrapper = await mountAppeals()

    expect(wrapper.text()).toContain('accepted appeal')
    expect(authAPI.claimAppealCompensation).not.toHaveBeenCalled()
    expect(adminAPI.unbanUser).not.toHaveBeenCalled()
    expect(adminAPI.unbanFromAnticheatPanel).not.toHaveBeenCalled()
    expect(adminAPI.approveAppeal).not.toHaveBeenCalled()
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
  })

  it('submits a banned-player appeal with the locked room list context', async () => {
    vi.mocked(authAPI.getAppealEntryStatus).mockResolvedValue(apiResponse({
      is_banned: true,
      can_submit: true,
      latest_risk_score_id: 33,
      first_room_id: 'locked-room-a',
      room_ids: ['locked-room-a', 'locked-room-b'],
    }))
    vi.mocked(authAPI.getPlayerAppeals)
      .mockResolvedValueOnce(apiResponse({ appeals: [], total: 0 }))
      .mockResolvedValueOnce(apiResponse({
        appeals: [{ id: 5, room_id: 'locked-room-a', reason: 'false positive', status: 'pending', submitted_at: '2026-05-05T00:00:00Z' }],
        total: 1,
      }))
    vi.mocked(authAPI.getPlayerSanctions).mockResolvedValue(apiResponse({ sanctions: [{ id: 99, room_id: 'locked-room-a', risk_score_id: 33 }] }))
    vi.mocked(authAPI.submitAppeal).mockResolvedValue(apiResponse({ appeal: { id: 5 } }))

    const wrapper = await mountAppeals()

    expect(wrapper.text()).toContain('locked-room-a')
    expect(wrapper.text()).toContain('locked-room-b')
    expect(wrapper.text()).toContain('33')

    const textareas = wrapper.findAll('textarea')
    await textareas[0].setValue('false positive')
    await textareas[1].setValue('network jitter at replay point')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(authAPI.submitAppeal).toHaveBeenCalledWith('locked-room-a', {
      risk_score_id: 33,
      sanction_id: 99,
      reason: 'false positive',
      evidence: 'network jitter at replay point',
    })
  })

  it('submits with only appeal reason when optional context is unavailable', async () => {
    vi.mocked(authAPI.getAppealEntryStatus).mockResolvedValue(apiResponse({
      is_banned: true,
      can_submit: true,
      room_ids: [],
    }))
    vi.mocked(authAPI.getPlayerAppeals)
      .mockResolvedValueOnce(apiResponse({ appeals: [], total: 0 }))
      .mockResolvedValueOnce(apiResponse({
        appeals: [{ id: 12, room_id: 'account', reason: 'reason only', status: 'pending', submitted_at: '2026-05-05T00:00:00Z' }],
        total: 1,
      }))
    vi.mocked(authAPI.getPlayerSanctions).mockResolvedValue(apiResponse({ sanctions: [] }))
    vi.mocked(authAPI.submitAppeal).mockResolvedValue(apiResponse({ appeal: { id: 12 } }))

    const wrapper = await mountAppeals()
    await wrapper.findAll('textarea')[0].setValue('reason only')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(authAPI.submitAppeal).toHaveBeenCalledWith('account', {
      reason: 'reason only',
    })
  })

  it('uses appeal entry ban state when refreshed user info has no ban timestamp', async () => {
    vi.mocked(authAPI.getAppealEntryStatus).mockResolvedValue(apiResponse({
      is_banned: true,
      can_submit: true,
      banned_until: futureIso(),
      ban_reason: 'active anticheat sanction',
      latest_risk_score_id: 33,
      first_room_id: 'locked-room-a',
      room_ids: ['locked-room-a'],
    }))
    vi.mocked(authAPI.refreshUserInfo).mockResolvedValue(apiResponse({
      uid: 1001,
      banned_until: null,
      ban_reason: '',
      points: 1100,
      fuel: 100,
    }))
    vi.mocked(authAPI.getPlayerAppeals).mockResolvedValue(apiResponse({ appeals: [], total: 0 }))
    vi.mocked(authAPI.getPlayerSanctions).mockResolvedValue(apiResponse({ sanctions: [{ id: 99, room_id: 'locked-room-a', risk_score_id: 33 }] }))

    const wrapper = await mountAppeals()

    expect(wrapper.text()).toContain('账号处于封禁中')
    expect(wrapper.text()).toContain('active anticheat sanction')
    expect(wrapper.text()).not.toContain('账号未处于封禁状态')
  })

  it('falls back to active ban sanctions when appeal entry reports unbanned', async () => {
    vi.mocked(authAPI.getAppealEntryStatus).mockResolvedValue(apiResponse({
      is_banned: false,
      can_submit: false,
      room_ids: [],
    }))
    vi.mocked(authAPI.refreshUserInfo).mockResolvedValue(apiResponse({
      uid: 1001,
      banned_until: null,
      ban_reason: '',
      points: 1100,
      fuel: 100,
    }))
    vi.mocked(authAPI.getPlayerAppeals).mockResolvedValue(apiResponse({ appeals: [], total: 0 }))
    vi.mocked(authAPI.getPlayerSanctions).mockResolvedValue(apiResponse({
      sanctions: [{
        id: 99,
        room_id: 'locked-room-a',
        risk_score_id: 33,
        sanction_type: 'ban',
        status: 'active',
        effective_until: futureIso(),
        reason: 'active anticheat sanction',
      }],
    }))

    const wrapper = await mountAppeals()

    expect(wrapper.text()).toContain('账号处于封禁中')
    expect(wrapper.text()).toContain('active anticheat sanction')
    expect(wrapper.text()).not.toContain('账号未处于封禁状态')
  })

  it('shows a query failure state when ban status endpoints fail', async () => {
    vi.mocked(authAPI.getAppealEntryStatus).mockRejectedValue(new Error('entry unavailable'))
    vi.mocked(authAPI.getPlayerAppeals).mockResolvedValue(apiResponse({ appeals: [], total: 0 }))
    vi.mocked(authAPI.getPlayerSanctions).mockRejectedValue(new Error('sanctions unavailable'))
    vi.mocked(authAPI.refreshUserInfo).mockResolvedValue(apiResponse({
      uid: 1001,
      banned_until: null,
      ban_reason: '',
      points: 1100,
      fuel: 100,
    }))

    const wrapper = await mountAppeals()

    expect(wrapper.text()).toContain('账号封禁状态查询失败')
    expect(wrapper.text()).not.toContain('账号未处于封禁状态')
  })

  it('redirects unbanned players to feedback instead of submitting appeal', async () => {
    vi.mocked(authAPI.getAppealEntryStatus).mockResolvedValue(apiResponse({
      is_banned: false,
      can_submit: false,
      room_ids: [],
    }))
    vi.mocked(authAPI.getPlayerAppeals).mockResolvedValue(apiResponse({ appeals: [], total: 0 }))
    vi.mocked(authAPI.getPlayerSanctions).mockResolvedValue(apiResponse({ sanctions: [] }))

    const wrapper = await mountAppeals()
    await wrapper.findAll('textarea')[0].setValue('I need help with a non-ban issue')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(authAPI.submitAppeal).not.toHaveBeenCalled()
    expect(showConfirm).toHaveBeenCalledWith(
      '当前账号未处于封禁状态，不能提交申诉。你可以前往反馈页面撰写反馈。',
      '申诉受限',
      '前往反馈',
      '留在本页'
    )
    expect(routerPush).toHaveBeenCalledWith({ path: '/feedbacks', query: { compose: '1' } })
  })

  it('stays on appeals when unbanned player cancels the feedback redirect dialog', async () => {
    showConfirm.mockResolvedValue(false)
    vi.mocked(authAPI.getAppealEntryStatus).mockResolvedValue(apiResponse({
      is_banned: false,
      can_submit: false,
      room_ids: [],
    }))
    vi.mocked(authAPI.getPlayerAppeals).mockResolvedValue(apiResponse({ appeals: [], total: 0 }))
    vi.mocked(authAPI.getPlayerSanctions).mockResolvedValue(apiResponse({ sanctions: [] }))

    const wrapper = await mountAppeals()
    await wrapper.findAll('textarea')[0].setValue('I need help with a non-ban issue')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(authAPI.submitAppeal).not.toHaveBeenCalled()
    expect(showConfirm).toHaveBeenCalled()
    expect(routerPush).not.toHaveBeenCalled()
  })
})
