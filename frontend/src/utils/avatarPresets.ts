import {
  FlaskConical,
  Dna,
  TestTube2,
  Microscope,
  Satellite,
  Rocket,
  Orbit,
  Atom,
  Radio,
  Brain,
  Bot,
  Ghost,
} from 'lucide-vue-next'
import type { Component } from 'vue'

export const AVATAR_PRESETS: Record<string, Component> = {
  flask:  FlaskConical,
  dna:    Dna,
  tube:   TestTube2,
  micro:  Microscope,
  sat:    Satellite,
  rocket: Rocket,
  orbit:  Orbit,
  atom:   Atom,
  radio:  Radio,
  brain:  Brain,
  bot:    Bot,
  ghost:  Ghost,
}

export const isPresetAvatar = (avatar: string | null | undefined): boolean =>
  !!avatar && Object.prototype.hasOwnProperty.call(AVATAR_PRESETS, avatar)
