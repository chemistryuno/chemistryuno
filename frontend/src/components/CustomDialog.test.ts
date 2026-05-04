import { mount } from '@vue/test-utils'
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
    expect(overlay?.classList.contains('absolute')).toBe(false)
    expect(document.body.querySelector('.viewport-modal-panel')).toBeTruthy()

    dialog.handleConfirm()
    wrapper.unmount()
  })
})
