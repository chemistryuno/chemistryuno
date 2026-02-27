/**
 * 操作反馈系统 - 震动 + 提示音
 * 协调音效引擎和振动引擎，提供统一的反馈接口
 */

import { audioEngine, SoundType } from './audioEngine'
import { vibrationEngine, VibrationPattern, VibrationPreset } from './vibrationEngine'

class FeedbackManager {
  private soundEnabled = true
  private vibrationEnabled = true
  private volume = 0.3

  /**
   * 初始化
   */
  constructor() {
    this.loadSettings()
    this.logSystemInfo()
  }

  /**
   * 记录系统信息
   */
  private logSystemInfo() {
    console.log('[Feedback] 反馈系统初始化')
    console.log('[Feedback] 音效启用:', this.getSoundEnabled())
    console.log('[Feedback] 振动启用:', this.getVibrationEnabled())
    console.log('[Feedback] 音量:', this.getVolume())
    console.log('[Feedback] 振动API支持:', vibrationEngine.isSupported())

    if (vibrationEngine.isSupported()) {
      console.log('[Feedback] ✅ 振动API可用')
    } else {
      console.warn('[Feedback] ❌ 振动API不可用 - 可能是桌面浏览器或不支持的设备')
    }
  }

  /**
   * 加载用户设置
   */
  private loadSettings() {
    const soundSetting = localStorage.getItem('chemistry-uno-sound-enabled')
    const vibrationSetting = localStorage.getItem('chemistry-uno-vibration-enabled')
    const volumeSetting = localStorage.getItem('chemistry-uno-volume')

    this.soundEnabled = soundSetting !== 'false' // 默认启用
    this.vibrationEnabled = vibrationSetting !== 'false' // 默认启用
    const parsedVolume = volumeSetting ? parseFloat(volumeSetting) : 0.3
    this.volume = Number.isFinite(parsedVolume) ? parsedVolume : 0.3

    audioEngine.setEnabled(this.soundEnabled)
    audioEngine.setVolume(this.volume)
    vibrationEngine.setEnabled(this.vibrationEnabled)
  }

  /**
   * 保存用户设置
   */
  private saveSettings() {
    localStorage.setItem('chemistry-uno-sound-enabled', String(this.soundEnabled))
    localStorage.setItem('chemistry-uno-vibration-enabled', String(this.vibrationEnabled))
    localStorage.setItem('chemistry-uno-volume', String(this.volume))
  }

  /**
   * 播放音效
   */
  playSound(type: SoundType) {
    audioEngine.play(type)
  }

  /**
   * 触发振动
   */
  vibrate(pattern: VibrationPattern | VibrationPreset) {
    vibrationEngine.vibrate(pattern)
  }

  /**
   * 组合反馈：震动 + 音效
   */
  feedback(options: {
    sound?: SoundType
    vibration?: VibrationPattern | VibrationPreset
  }) {
    if (options.sound) {
      this.playSound(options.sound)
    }
    if (options.vibration) {
      this.vibrate(options.vibration)
    }
  }

  /**
   * 快捷方法：点击反馈
   */
  click() {
    this.feedback({ sound: 'click', vibration: 'light' })
  }

  /**
   * 快捷方法：打牌反馈
   */
  playCard() {
    this.feedback({ sound: 'play-card', vibration: 'medium' })
  }

  /**
   * 快捷方法：摸牌反馈
   */
  drawCard() {
    this.feedback({ sound: 'draw-card', vibration: 'light' })
  }

  /**
   * 快捷方法：化学反应反馈
   */
  reaction() {
    this.feedback({ sound: 'reaction', vibration: 'reaction' })
  }

  /**
   * 快捷方法：回合开始反馈
   */
  turnStart() {
    this.feedback({ sound: 'turn-start', vibration: 'medium' })
  }

  /**
   * 快捷方法：成功反馈
   */
  success() {
    this.feedback({ sound: 'success', vibration: 'success' })
  }

  /**
   * 快捷方法：错误反馈
   */
  error() {
    this.feedback({ sound: 'error', vibration: 'error' })
  }

  /**
   * 快捷方法：胜利反馈
   */
  win() {
    this.feedback({ sound: 'win', vibration: 'success' })
  }

  /**
   * 快捷方法：失败反馈
   */
  lose() {
    this.feedback({ sound: 'lose', vibration: 'heavy' })
  }

  /**
   * 快捷方法：双联模式反馈
   */
  doubleMode() {
    this.feedback({ sound: 'double-mode', vibration: 'double' })
  }

  /**
   * 快捷方法：升级反馈
   */
  levelUp() {
    this.feedback({ sound: 'level-up', vibration: 'success' })
  }

  /**
   * 设置音效开关
   */
  setSoundEnabled(enabled: boolean) {
    audioEngine.setEnabled(enabled)
    this.soundEnabled = enabled
    this.saveSettings()
  }

  /**
   * 设置震动开关
   */
  setVibrationEnabled(enabled: boolean) {
    vibrationEngine.setEnabled(enabled)
    this.vibrationEnabled = enabled
    this.saveSettings()
  }

  /**
   * 设置音量 (0.0 - 1.0)
   */
  setVolume(volume: number) {
    audioEngine.setVolume(volume)
    this.volume = volume
    this.saveSettings()
  }

  /**
   * 获取音效启用状态
   */
  getSoundEnabled(): boolean {
    return this.soundEnabled
  }

  /**
   * 获取振动启用状态
   */
  getVibrationEnabled(): boolean {
    return this.vibrationEnabled
  }

  /**
   * 获取音量
   */
  getVolume(): number {
    return this.volume
  }

  /**
   * 获取当前设置
   */
  getSettings() {
    return {
      soundEnabled: this.getSoundEnabled(),
      vibrationEnabled: this.getVibrationEnabled(),
      volume: this.getVolume(),
    }
  }

  /**
   * 诊断振动功能
   */
  diagnoseVibration() {
    return vibrationEngine.diagnose()
  }

  /**
   * 停止所有反馈
   */
  stopAll() {
    audioEngine.stopAll()
    vibrationEngine.stop()
  }
}

// 导出单例实例
export const feedback = new FeedbackManager()

// 默认导出
export default feedback

// 导出类型
export type { SoundType } from './audioEngine'
export type { VibrationPattern, VibrationPreset } from './vibrationEngine'
