import { expect, test } from '@playwright/test'
import { seededUsers } from './helpers'

const json = (body: unknown) => ({
  status: 200,
  contentType: 'application/json',
  body: JSON.stringify(body),
})

const seedUser = async (page: import('@playwright/test').Page) => {
  await page.addInitScript(() => {
    window.localStorage.setItem('user', JSON.stringify({
      uid: 1,
      username: 'admin',
      nickname: '系统管理员',
      role: 'admin',
      is_admin: true,
    }))
  })
}

test.beforeEach(async ({ page }) => {
  await page.route('**/api/announcements**', route => route.fulfill(json([])))
  await page.route('**/api/plugins**', route => route.fulfill(json([])))
  await page.route('**/api/level/info**', route => route.fulfill(json({ level: 1, xp: 0, total_xp: 0 })))
  await page.route('**/api/user/info', route => route.fulfill(json({
    uid: 1,
    username: seededUsers.admin.username,
    nickname: seededUsers.admin.nickname,
    role: 'admin',
    is_admin: true,
    total_games: 0,
    win_count: 0,
  })))
  await page.route('**/api/auth/config', route => route.fulfill(json({ smtp_enabled: false })))
  await page.route('**/api/admin/anticheat/detection-list**', route => route.fulfill(json({ detections: [], total: 0 })))
  await page.route('**/api/admin/anticheat/config', route => route.fulfill(json({
    dimensions: {
      response_time: { weight: 0.25, threshold: 100 },
      frequency: { weight: 0.25, threshold: 5 },
      win_rate: { weight: 0.2, threshold: 0.9 },
      pattern: { weight: 0.15, threshold: 0.7 },
      account_age: { weight: 0.15, threshold: 7 },
    },
    sanctions: { observe: 20, warning: 40, mute: 60, ban: 80 },
    unban: {
      compensation_amount: 100,
      default_message: 'Default compensation message',
    },
  })))
})

test('admin workflow approves an appeal and sees compensation in audit', async ({ page }) => {
  await seedUser(page)

  await page.route('**/api/admin/anticheat/appeals**', route => route.fulfill(json({
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
  })))

  let approvalPayload: any
  await page.route('**/api/admin/anticheat/appeals/appeal-1/approve', async route => {
    approvalPayload = route.request().postDataJSON()
    await route.fulfill(json({ status: 'approved', compensation_status: 'ok' }))
  })

  await page.route('**/api/admin/anticheat/audit-log**', route => route.fulfill(json({
    logs: [
      {
        id: 'audit-1',
        player_id: 42,
        action_type: 'unban',
        details: 'approved',
        compensation_status: 'ok',
        compensation_amount: 120,
        created_at: '2026-05-03T00:05:00Z',
      },
    ],
    total: 1,
  })))

  await page.goto('/admin/anticheat')
  await page.getByRole('button', { name: '申诉管理' }).click()
  await expect(page.getByText('误封申诉')).toBeVisible()

  await page.getByRole('button', { name: /批准/ }).click()
  await expect(page.getByText('批准申诉并发放补偿')).toBeVisible()
  await page.getByText('调整补偿配置').click()
  await page.locator('input[type="number"]').fill('120')
  await page.locator('textarea').first().fill('Custom player compensation message')
  await page.locator('textarea').last().fill('Reviewed clean replay')
  await page.getByRole('button', { name: /确认批准/ }).click()

  expect(approvalPayload).toEqual({
    note: 'Reviewed clean replay',
    compensation_amount: 120,
    compensation_message: 'Custom player compensation message',
  })

  await page.getByRole('button', { name: '确定' }).click()
  await page.getByRole('button', { name: '审计日志' }).click()
  await expect(page.getByText('已发放')).toBeVisible()
  await expect(page.getByText('120燃素')).toBeVisible()
})

test('player dashboard loads anticheat stats and refreshes periodically', async ({ page }) => {
  await seedUser(page)
  await page.addInitScript(() => {
    const originalSetInterval = window.setInterval
    window.setInterval = ((handler: TimerHandler, timeout?: number, ...args: any[]) => {
      if (timeout === 5 * 60 * 1000) {
        ;(window as any).__runAnticheatStatsRefresh = () => {
          if (typeof handler === 'function') {
            handler(...args)
          } else {
            new Function(handler)()
          }
        }
        return 1
      }
      return originalSetInterval(handler, timeout, ...args)
    }) as typeof window.setInterval
  })
  let statsCalls = 0

  await page.route('**/api/player/anticheat/stats', route => {
    statsCalls += 1
    return route.fulfill(json({
      bans_today: statsCalls <= 2 ? 2 : 5,
      system_uptime_days: 9,
    }))
  })

  await page.goto('/profile')

  const anticheatWidget = page.locator('section').filter({ hasText: 'Anticheat' })
  await expect(anticheatWidget.getByText('Bans Today')).toBeVisible()
  await expect(anticheatWidget.getByText('System Running')).toBeVisible()
  await expect(anticheatWidget.getByText('2', { exact: true })).toBeVisible()
  await expect(anticheatWidget.getByText(/9\s*days/)).toBeVisible()

  await page.evaluate(() => (window as any).__runAnticheatStatsRefresh())
  await expect(anticheatWidget.getByText('5', { exact: true })).toBeVisible()
  expect(statsCalls).toBeGreaterThanOrEqual(2)
})
