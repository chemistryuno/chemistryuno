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
    const wrapper = mount(CustomDialog)
    void dialog.showAlert('Scrolled page alert', 'Viewport Check')
    await wrapper.vm.$nextTick()

    const overlay = wrapper.find('.viewport-modal-overlay')
    expect(overlay.exists()).toBe(true)
    expect(overlay.classes()).toContain('viewport-modal-overlay')
    expect(overlay.classes()).not.toContain('absolute')
    expect(wrapper.find('.viewport-modal-panel').exists()).toBe(true)

    dialog.handleConfirm()
  })
})
