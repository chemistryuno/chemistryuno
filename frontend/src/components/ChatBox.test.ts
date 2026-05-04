import { mount, flushPromises } from '@vue/test-utils'
import { describe, expect, it, beforeEach, vi } from 'vitest'
import ChatBox from './ChatBox.vue'
import UserAvatar from './UserAvatar.vue'

// Mock the API and websocket
vi.mock('../utils/api', () => ({
  authAPI: {
    getUserInfo: vi.fn(),
    getGlobalChatHistory: vi.fn(() => Promise.resolve({ data: [] }))
  },
  gameAPI: {
    checkRoomStatus: vi.fn()
  }
}))

vi.mock('../utils/websocket', () => ({
  default: {
    send: vi.fn(),
    on: vi.fn(),
    off: vi.fn()
  }
}))

vi.mock('../utils/banState', () => ({
  getBanState: vi.fn((user?: any) => {
    if (!user) {
      // Try to get from localStorage if not provided
      try {
        const stored = JSON.parse(localStorage.getItem('user') || '{}')
        return {
          isBanned: !!stored.banned_until && new Date(stored.banned_until) > new Date(),
          bannedUntil: stored.banned_until,
          banReason: stored.ban_reason
        }
      } catch {
        return { isBanned: false }
      }
    }
    return {
      isBanned: !!user.banned_until && new Date(user.banned_until) > new Date(),
      bannedUntil: user.banned_until,
      banReason: user.ban_reason
    }
  }),
  banStateFromBlockedMessage: vi.fn((message: any) => {
    const data = message?.data || {}
    const until = data.banned_until
    return {
      isBanned: !!until && new Date(until) > new Date(),
      bannedUntil: until,
      banReason: data.ban_reason
    }
  }),
  formatBanUntil: vi.fn((bannedUntil?: string) => {
    if (!bannedUntil) return '未提供'
    return new Date(bannedUntil).toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' })
  }),
  getStoredUser: vi.fn(() => {
    try {
      return JSON.parse(localStorage.getItem('user') || '{}')
    } catch {
      return {}
    }
  })
}))

describe('ChatBox - Banned User Display', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
  })

  it('should render normal chat when user is not banned', async () => {
    // Set non-banned user in localStorage
    localStorage.setItem('user', JSON.stringify({
      uid: 1,
      username: 'testuser',
      banned_until: null
    }))

    const wrapper = mount(ChatBox, {
      props: {
        roomId: '',
        title: 'Public Chat'
      },
      global: {
        stubs: {
          UserAvatar: true,
          teleport: true
        }
      }
    })

    await flushPromises()

    // Should not show the appeal component
    const appealComponent = wrapper.find('[data-testid="public-chat-appeal"]')
    expect(appealComponent.exists()).toBe(false)

    // Should show the messages container instead
    const messagesContainer = wrapper.find('.overflow-y-auto')
    expect(messagesContainer.exists()).toBe(true)

    wrapper.unmount()
  })

  it('should render appeal component when user is banned', async () => {
    // Set banned user in localStorage
    const futureDate = new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString()
    localStorage.setItem('user', JSON.stringify({
      uid: 1,
      username: 'banneduser',
      banned_until: futureDate,
      ban_reason: 'Violating community guidelines'
    }))

    const wrapper = mount(ChatBox, {
      props: {
        roomId: '',
        title: 'Public Chat'
      },
      global: {
        stubs: {
          UserAvatar: true,
          teleport: true
        }
      }
    })

    await wrapper.vm.$nextTick()

    // Should show the appeal component
    const appealComponent = wrapper.find('[data-testid="public-chat-appeal"]')
    expect(appealComponent.exists()).toBe(true)

    // Should contain ban information
    const appealText = appealComponent.text()
    expect(appealText).toContain('公共聊天暂不可用')
    expect(appealText).toContain('账号封禁期间无法参与公共聊天')

    wrapper.unmount()
  })

  it('should display ban expiration time in appeal component', async () => {
    const futureDate = new Date(Date.now() + 48 * 60 * 60 * 1000).toISOString()
    localStorage.setItem('user', JSON.stringify({
      uid: 1,
      username: 'banneduser',
      banned_until: futureDate,
      ban_reason: 'Spam'
    }))

    const wrapper = mount(ChatBox, {
      props: {
        roomId: '',
        title: 'Public Chat'
      },
      global: {
        stubs: {
          UserAvatar: true,
          teleport: true
        }
      }
    })

    await wrapper.vm.$nextTick()

    const appealComponent = wrapper.find('[data-testid="public-chat-appeal"]')
    expect(appealComponent.exists()).toBe(true)

    // Should display the ban expiration
    const appealText = appealComponent.text()
    expect(appealText).toContain('封禁截止')

    wrapper.unmount()
  })

  it('should display ban reason in appeal component', async () => {
    const futureDate = new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString()
    const banReason = 'Inappropriate language in chat'
    
    localStorage.setItem('user', JSON.stringify({
      uid: 1,
      username: 'banneduser',
      banned_until: futureDate,
      ban_reason: banReason
    }))

    const wrapper = mount(ChatBox, {
      props: {
        roomId: '',
        title: 'Public Chat'
      },
      global: {
        stubs: {
          UserAvatar: true,
          teleport: true
        }
      }
    })

    await wrapper.vm.$nextTick()

    const appealComponent = wrapper.find('[data-testid="public-chat-appeal"]')
    expect(appealComponent.exists()).toBe(true)

    const appealText = appealComponent.text()
    expect(appealText).toContain('封禁原因')
    expect(appealText).toContain(banReason)

    wrapper.unmount()
  })

  it('should display appeal button in the appeal component', async () => {
    const futureDate = new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString()
    localStorage.setItem('user', JSON.stringify({
      uid: 1,
      username: 'banneduser',
      banned_until: futureDate,
      ban_reason: 'Test ban'
    }))

    const wrapper = mount(ChatBox, {
      props: {
        roomId: '',
        title: 'Public Chat'
      },
      global: {
        stubs: {
          UserAvatar: true,
          teleport: true
        }
      }
    })

    await wrapper.vm.$nextTick()

    const appealComponent = wrapper.find('[data-testid="public-chat-appeal"]')
    expect(appealComponent.exists()).toBe(true)

    // Should contain an appeal button with specific text
    const appealButton = appealComponent.find('button')
    expect(appealButton.exists()).toBe(true)
    expect(appealButton.text()).toContain('提交或查看申诉')

    wrapper.unmount()
  })

  it('should handle appeal button click', async () => {
    const futureDate = new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString()
    localStorage.setItem('user', JSON.stringify({
      uid: 1,
      username: 'banneduser',
      banned_until: futureDate,
      ban_reason: 'Test ban'
    }))

    const wrapper = mount(ChatBox, {
      props: {
        roomId: '',
        title: 'Public Chat'
      },
      global: {
        stubs: {
          UserAvatar: true,
          teleport: true
        }
      }
    })

    await wrapper.vm.$nextTick()

    const appealComponent = wrapper.find('[data-testid="public-chat-appeal"]')
    const appealButton = appealComponent.find('button')

    await appealButton.trigger('click')
    await wrapper.vm.$nextTick()

    // Check if appeal status message appears
    const appealStatus = appealComponent.text()
    expect(appealStatus).toContain('申诉入口已打开') || 
    expect(wrapper.vm.appealStatus).toBeTruthy()

    wrapper.unmount()
  })

  it('should not block chat when ban has expired', async () => {
    // Set expired ban date (in the past)
    const pastDate = new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString()
    localStorage.setItem('user', JSON.stringify({
      uid: 1,
      username: 'testuser',
      banned_until: pastDate,
      ban_reason: 'Old ban'
    }))

    const wrapper = mount(ChatBox, {
      props: {
        roomId: '',
        title: 'Public Chat'
      },
      global: {
        stubs: {
          UserAvatar: true,
          teleport: true
        }
      }
    })

    await wrapper.vm.$nextTick()

    // Should not show the appeal component for expired ban
    const appealComponent = wrapper.find('[data-testid="public-chat-appeal"]')
    expect(appealComponent.exists()).toBe(false)

    wrapper.unmount()
  })

  it('should not block chat in room context when banned', async () => {
    // When in a room (roomId is set), even if user is banned,
    // the chat should not be blocked (server handles room-level restrictions)
    const futureDate = new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString()
    localStorage.setItem('user', JSON.stringify({
      uid: 1,
      username: 'banneduser',
      banned_until: futureDate,
      ban_reason: 'Test ban'
    }))

    const wrapper = mount(ChatBox, {
      props: {
        roomId: 'game-123',
        title: 'Room Chat'
      },
      global: {
        stubs: {
          UserAvatar: true,
          teleport: true
        }
      }
    })

    await wrapper.vm.$nextTick()

    // Should NOT show the appeal component when in a room
    const appealComponent = wrapper.find('[data-testid="public-chat-appeal"]')
    expect(appealComponent.exists()).toBe(false)

    wrapper.unmount()
  })
})
