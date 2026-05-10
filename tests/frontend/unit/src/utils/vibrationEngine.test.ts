import { beforeEach, describe, expect, it, vi } from 'vitest'
import { VibrationEngine } from '@/utils/vibrationEngine'

describe('VibrationEngine', () => {
  beforeEach(() => {
    vi.spyOn(console, 'log').mockImplementation(() => {})
    vi.spyOn(console, 'warn').mockImplementation(() => {})
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  it('returns false when disabled', () => {
    const engine = new VibrationEngine()
    engine.setEnabled(false)

    expect(engine.vibrate('light')).toBe(false)
  })

  it('returns false when the vibration API is unavailable', () => {
    const engine = new VibrationEngine()
    expect(engine.isSupported()).toBe(false)
    expect(engine.vibrate('light')).toBe(false)
  })

  it('normalizes preset and readonly array patterns for navigator.vibrate', () => {
    const vibrate = vi.fn(() => true)
    Object.defineProperty(navigator, 'vibrate', {
      configurable: true,
      value: vibrate,
    })
    const engine = new VibrationEngine()

    expect(engine.vibrate('double')).toBe(true)
    expect(vibrate).toHaveBeenCalledWith([20, 50, 20])

    expect(engine.vibrate([1, 2, 3] as const)).toBe(true)
    expect(vibrate).toHaveBeenLastCalledWith([1, 2, 3])
  })

  it('stops vibration and reports diagnostics', () => {
    const vibrate = vi.fn(() => true)
    Object.defineProperty(navigator, 'vibrate', {
      configurable: true,
      value: vibrate,
    })
    const engine = new VibrationEngine()

    engine.stop()
    expect(vibrate).toHaveBeenCalledWith(0)

    expect(engine.diagnose()).toMatchObject({
      enabled: true,
      apiAvailable: true,
    })
  })
})
