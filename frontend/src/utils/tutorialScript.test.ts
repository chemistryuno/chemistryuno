import { describe, expect, it } from 'vitest'
import {
  TUTORIAL_INITIAL_STATE,
  TUTORIAL_SCRIPT,
  TUTORIAL_TOTAL_STEPS,
  canPlaySubstance,
  getTutorialProgress,
  getTutorialStep,
} from './tutorialScript'

describe('tutorial script helpers', () => {
  it('keeps total steps synchronized with the script', () => {
    expect(TUTORIAL_TOTAL_STEPS).toBe(TUTORIAL_SCRIPT.length)
    expect(TUTORIAL_INITIAL_STATE.discardTop).toBe('Cl2')
  })

  it('finds tutorial steps by step number', () => {
    expect(getTutorialStep(1)?.substance).toBe('Mg')
    expect(getTutorialStep(999)).toBeUndefined()
  })

  it('only allows the exact human step substance', () => {
    expect(canPlaySubstance('Mg', 1)).toBe(true)
    expect(canPlaySubstance('NaOH', 1)).toBe(false)
    expect(canPlaySubstance('HCl', 2)).toBe(false)
  })

  it('reports human tutorial progress', () => {
    expect(getTutorialProgress(1)).toBe('0/5')
    expect(getTutorialProgress(4)).toBe('2/5')
    expect(getTutorialProgress(9)).toBe('5/5')
  })
})
