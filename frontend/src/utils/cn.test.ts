import { describe, expect, it } from 'vitest'
import { cn } from './cn'

describe('cn', () => {
  it('combines conditional class names and resolves tailwind conflicts', () => {
    expect(cn('px-2', false && 'hidden', ['text-sm', 'px-4'])).toBe('text-sm px-4')
  })
})
