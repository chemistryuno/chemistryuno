import { expect, type Page } from '@playwright/test'

export const seededUsers = {
  admin: {
    username: 'admin',
    password: 'admin123',
    nickname: '系统管理员',
  },
  player: {
    username: 'test',
    password: 'test123',
    nickname: '测试用户',
  },
}

export const loginAs = async (
  page: Page,
  user: { username: string; password: string },
) => {
  await page.goto('/login')
  await page.getByPlaceholder('请输入用户名或邮箱').fill(user.username)
  await page.getByPlaceholder('请输入访问凭证').fill(user.password)
  await page.getByRole('button', { name: '授权并进入' }).click()
  await expect(page).toHaveURL(/\/$/)
}

export const silenceOptionalRealtime = async (page: Page) => {
  await page.addInitScript(() => {
    const originalWebSocket = window.WebSocket
    window.WebSocket = class extends originalWebSocket {
      constructor(url: string | URL, protocols?: string | string[]) {
        super(url, protocols)
      }
    }
  })
}
