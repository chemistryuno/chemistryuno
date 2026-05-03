import { expect, test } from '@playwright/test'
import { loginAs, seededUsers } from './helpers'

test('seeded player can log in and load the lobby from the real backend', async ({ page }) => {
  await loginAs(page, seededUsers.player)

  await expect(page.getByText('试验场大厅')).toBeVisible()
  await expect(page.getByText('暂无开放的房间').or(page.getByText('房间号:').first())).toBeVisible()
})

test('seeded admin can log in and reach admin navigation', async ({ page }) => {
  await loginAs(page, seededUsers.admin)

  await expect(page.getByText('管理员').first()).toBeVisible()
  await expect(page.getByTitle('管理面板')).toBeVisible()
})
