import { describe, expect, it } from 'vitest'
import { AVATAR_PRESETS, isPresetAvatar } from './avatarPresets'

describe('avatar presets', () => {
  it('recognizes known preset avatars only', () => {
    expect(Object.keys(AVATAR_PRESETS)).toContain('flask')
    expect(isPresetAvatar('flask')).toBe(true)
    expect(isPresetAvatar('missing')).toBe(false)
    expect(isPresetAvatar('toString')).toBe(false)
    expect(isPresetAvatar(null)).toBe(false)
  })
})
