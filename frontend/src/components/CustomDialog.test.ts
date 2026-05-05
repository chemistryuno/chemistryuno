import { mount } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'
import CustomDialog from './CustomDialog.vue'
import { useDialog } from '../utils/dialog'

describe('CustomDialog viewport placement', () => {
  it('uses the viewport-centered overlay contract after page scroll', async () => {
    Object.defineProperty(window, 'scrollY', {
      configurable: true,
      value: 600,
    })

    const dialog = useDialog()
    const wrapper = mount(CustomDialog, {
      attachTo: document.body,
      global: {
        stubs: {
          teleport: false,
        },
      },
    })
    void dialog.showAlert('Scrolled page alert', 'Viewport Check')
    await wrapper.vm.$nextTick()

    const overlay = document.body.querySelector('.viewport-modal-overlay')
    expect(overlay).toBeTruthy()
    expect(overlay?.parentElement).toBe(document.body)
    expect(overlay?.classList.contains('viewport-dialog-overlay')).toBe(true)
    expect(overlay?.classList.contains('absolute')).toBe(false)
    expect(document.body.querySelector('.viewport-modal-panel')).toBeTruthy()

    dialog.handleConfirm()
    wrapper.unmount()
  })

  it('keeps scrolling on the dialog panel without showing modal scrollbars', () => {
    const css = readFileSync(join(process.cwd(), 'src/index.css'), 'utf8')
    const overlayRule = css.match(/\.viewport-modal-overlay\s*\{[^}]*\}/)?.[0] ?? ''
    const dialogOverlayRule = css.match(/\.viewport-dialog-overlay\s*\{[^}]*\}/)?.[0] ?? ''
    const panelRule = css.match(/\.viewport-modal-panel\s*\{[^}]*\}/)?.[0] ?? ''
    const panelScrollbarRule = css.match(/\.viewport-modal-panel::-webkit-scrollbar\s*\{[^}]*\}/)?.[0] ?? ''

    expect(overlayRule).toContain('overflow: hidden;')
    expect(overlayRule).not.toContain('overflow-y: auto;')
    expect(overlayRule).toContain('z-index: var(--app-modal-z-index) !important;')
    expect(dialogOverlayRule).toContain('z-index: var(--app-dialog-z-index) !important;')
    expect(panelRule).toContain('overflow-y: auto;')
    expect(panelRule).toContain('scrollbar-width: none;')
    expect(panelScrollbarRule).toContain('display: none;')
  })
})
