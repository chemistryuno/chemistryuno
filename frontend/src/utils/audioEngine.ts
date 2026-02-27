/**
 * 音效引擎 - 实验室主题音效系�?
 * 使用 Web Audio API 生成实验室特有的声音效果
 * 包含玻璃器皿、纸张摩擦、化学反应等真实音效
 * 独立的音效处理模块，与反馈系统解�?
 */

export type SoundType =
  | 'click'          // 普通点�?
  | 'play-card'      // 打牌
  | 'draw-card'      // 摸牌
  | 'reaction'       // 化学反应
  | 'turn-start'     // 回合开�?
  | 'success'        // 成功操作
  | 'error'          // 错误操作
  | 'win'            // 胜利
  | 'lose'           // 失败
  | 'double-mode'    // 双联模式
  | 'special'        // 特殊效果
  | 'level-up'       // 升级

export class AudioEngine {
  private context: AudioContext | null = null
  private volume: number = 0.3
  private enabled: boolean = true

  /**
   * 设置音量 (0.0 - 1.0)
   */
  setVolume(volume: number) {
    this.volume = Math.max(0, Math.min(1, volume))
  }

  /**
   * 设置启用状�?
   */
  setEnabled(enabled: boolean) {
    this.enabled = enabled
  }

  /**
   * 获取或创�?AudioContext
   */
  private getContext(): AudioContext {
    if (!this.context) {
      this.context = new (window.AudioContext || (window as any).webkitAudioContext)()
    }
    return this.context
  }

  /**
   * 生成白噪声缓冲区
   */
  private createNoiseBuffer(duration: number): AudioBuffer {
    const context = this.getContext()
    const bufferSize = context.sampleRate * duration
    const buffer = context.createBuffer(1, bufferSize, context.sampleRate)
    const output = buffer.getChannelData(0)

    for (let i = 0; i < bufferSize; i++) {
      output[i] = Math.random() * 2 - 1
    }

    return buffer
  }

  /**
   * 播放玻璃破碎音效（用于化学反应等�?
   */
  private playGlassBreak(intensity: number = 1) {
    if (!this.enabled) return

    try {
      const context = this.getContext()
      const now = context.currentTime

      // 噪声�?
      const noise = context.createBufferSource()
      noise.buffer = this.createNoiseBuffer(0.3)

      // 高通滤波器 - 模拟玻璃的高频破碎声
      const highpass = context.createBiquadFilter()
      highpass.type = 'highpass'
      highpass.frequency.value = 2000 + intensity * 1000

      // 增益控制 - 快速衰�?
      const gainNode = context.createGain()
      gainNode.gain.setValueAtTime(this.volume * intensity, now)
      gainNode.gain.exponentialRampToValueAtTime(0.01, now + 0.3)

      // 连接节点
      noise.connect(highpass)
      highpass.connect(gainNode)
      gainNode.connect(context.destination)

      noise.start(now)
      noise.stop(now + 0.3)
    } catch (error) {
      console.warn('[AudioEngine] 玻璃破碎音效播放失败:', error)
    }
  }

  /**
   * 播放纸张翻页/卡牌摩擦音效
   */
  private playPaperShuffle(duration: number = 0.15, pitch: number = 1) {
    if (!this.enabled) return

    try {
      const context = this.getContext()
      const now = context.currentTime

      // 噪声�?
      const noise = context.createBufferSource()
      noise.buffer = this.createNoiseBuffer(duration)

      // 带通滤波器 - 模拟纸张摩擦的中频声
      const bandpass = context.createBiquadFilter()
      bandpass.type = 'bandpass'
      bandpass.frequency.value = 800 * pitch
      bandpass.Q.value = 3

      // 增益包络
      const gainNode = context.createGain()
      gainNode.gain.setValueAtTime(0, now)
      gainNode.gain.linearRampToValueAtTime(this.volume * 0.4, now + 0.01)
      gainNode.gain.exponentialRampToValueAtTime(0.01, now + duration)

      // 连接节点
      noise.connect(bandpass)
      bandpass.connect(gainNode)
      gainNode.connect(context.destination)

      noise.start(now)
      noise.stop(now + duration)
    } catch (error) {
      console.warn('[AudioEngine] 纸张音效播放失败:', error)
    }
  }

  /**
   * 播放烧杯碰撞/玻璃器皿音效
   */
  private playBeakerClink(frequency: number = 2000) {
    if (!this.enabled) return

    try {
      const context = this.getContext()
      const now = context.currentTime

      // 主音�?- 玻璃的共振频�?
      const oscillator = context.createOscillator()
      oscillator.type = 'sine'
      oscillator.frequency.value = frequency

      // 添加谐波
      const oscillator2 = context.createOscillator()
      oscillator2.type = 'sine'
      oscillator2.frequency.value = frequency * 2.5

      // 增益包络 - 快速衰减模拟玻璃敲�?
      const gainNode = context.createGain()
      gainNode.gain.setValueAtTime(this.volume * 0.5, now)
      gainNode.gain.exponentialRampToValueAtTime(0.01, now + 0.15)

      const gainNode2 = context.createGain()
      gainNode2.gain.setValueAtTime(this.volume * 0.3, now)
      gainNode2.gain.exponentialRampToValueAtTime(0.01, now + 0.1)

      // 连接节点
      oscillator.connect(gainNode)
      oscillator2.connect(gainNode2)
      gainNode.connect(context.destination)
      gainNode2.connect(context.destination)

      oscillator.start(now)
      oscillator2.start(now)
      oscillator.stop(now + 0.15)
      oscillator2.stop(now + 0.1)
    } catch (error) {
      console.warn('[AudioEngine] 烧杯音效播放失败:', error)
    }
  }

  /**
   * 播放气泡/化学反应音效
   */
  private playBubble(frequency: number = 200, count: number = 1) {
    if (!this.enabled) return

    try {
      const context = this.getContext()

      for (let i = 0; i < count; i++) {
        setTimeout(() => {
          const now = context.currentTime

          // 低频振荡�?- 模拟气泡破裂
          const oscillator = context.createOscillator()
          oscillator.type = 'sine'
          oscillator.frequency.setValueAtTime(frequency + Math.random() * 100, now)
          oscillator.frequency.exponentialRampToValueAtTime(frequency * 0.5, now + 0.1)

          // 增益包络
          const gainNode = context.createGain()
          gainNode.gain.setValueAtTime(this.volume * 0.4, now)
          gainNode.gain.exponentialRampToValueAtTime(0.01, now + 0.1)

          oscillator.connect(gainNode)
          gainNode.connect(context.destination)

          oscillator.start(now)
          oscillator.stop(now + 0.1)
        }, i * 50)
      }
    } catch (error) {
      console.warn('[AudioEngine] 气泡音效播放失败:', error)
    }
  }

  /**
   * 播放实验室按�?开关音�?
   */
  private playLabClick(frequency: number = 1500) {
    if (!this.enabled) return

    try {
      const context = this.getContext()
      const now = context.currentTime

      // 短促的机械声
      const oscillator = context.createOscillator()
      oscillator.type = 'square'
      oscillator.frequency.value = frequency

      // 低通滤波器 - 柔化机械�?
      const lowpass = context.createBiquadFilter()
      lowpass.type = 'lowpass'
      lowpass.frequency.value = 2000

      // 快速衰�?
      const gainNode = context.createGain()
      gainNode.gain.setValueAtTime(this.volume * 0.3, now)
      gainNode.gain.exponentialRampToValueAtTime(0.01, now + 0.04)

      oscillator.connect(lowpass)
      lowpass.connect(gainNode)
      gainNode.connect(context.destination)

      oscillator.start(now)
      oscillator.stop(now + 0.04)
    } catch (error) {
      console.warn('[AudioEngine] 点击音效播放失败:', error)
    }
  }

  /**
   * 播放预设音效
   */
  play(type: SoundType) {
    if (!this.enabled) {
      console.log('[AudioEngine] Sound is disabled')
      return
    }

    console.log('[AudioEngine] 播放音效:', type)

    switch (type) {
      case 'click':
        // 实验室按钮点�?
        this.playLabClick(1400)
        break

      case 'play-card':
        // 卡牌翻出 - 纸张摩擦�?
        this.playPaperShuffle(0.12, 1.2)
        setTimeout(() => this.playPaperShuffle(0.08, 1.4), 60)
        break

      case 'draw-card':
        // 抽卡 - 快速翻页声
        this.playPaperShuffle(0.15, 0.9)
        break

      case 'reaction':
        // 化学反应 - 玻璃破碎 + 气泡�?
        this.playGlassBreak(0.8)
        setTimeout(() => this.playBubble(250, 3), 100)
        break

      case 'turn-start':
        // 回合开�?- 清脆的烧杯碰�?
        this.playBeakerClink(2200)
        break

      case 'success':
        // 成功 - 三声烧杯碰撞
        this.playBeakerClink(1800)
        setTimeout(() => this.playBeakerClink(2200), 70)
        setTimeout(() => this.playBeakerClink(2600), 140)
        break

      case 'error':
        // 错误 - 低沉的玻璃撞�?
        this.playBeakerClink(800)
        setTimeout(() => this.playBeakerClink(700), 100)
        break

      case 'win':
        // 胜利 - 上升的烧杯音 + 玻璃闪光
        this.playBeakerClink(1500)
        setTimeout(() => this.playBeakerClink(1800), 80)
        setTimeout(() => this.playBeakerClink(2200), 160)
        setTimeout(() => this.playBeakerClink(2800), 240)
        setTimeout(() => {
          this.playGlassBreak(0.5)
          this.playBubble(300, 5)
        }, 320)
        break

      case 'lose':
        // 失败 - 下降的玻璃音 + 沉闷气泡
        this.playBeakerClink(2000)
        setTimeout(() => this.playBeakerClink(1600), 120)
        setTimeout(() => {
          this.playBeakerClink(1200)
          this.playBubble(150, 2)
        }, 240)
        break

      case 'double-mode':
        // 双联模式 - 快速双击纸�?
        this.playPaperShuffle(0.08, 1.5)
        setTimeout(() => this.playPaperShuffle(0.08, 1.7), 50)
        break

      case 'special':
        // 特殊效果 - 玻璃碎裂 + 多重气泡
        this.playGlassBreak(1.2)
        setTimeout(() => this.playBubble(280, 4), 80)
        setTimeout(() => this.playBeakerClink(3000), 150)
        break

      case 'level-up':
        // 升级 - 连续上升的烧杯音�?+ 玻璃闪光
        const frequencies = [1200, 1400, 1600, 1900, 2300, 2800]
        frequencies.forEach((freq, index) => {
          setTimeout(() => this.playBeakerClink(freq), index * 80)
        })
        setTimeout(() => {
          this.playGlassBreak(0.6)
          this.playBubble(350, 6)
        }, 480)
        break
    }
  }

  /**
   * 停止所有音�?
   */
  stopAll() {
    if (this.context) {
      try {
        this.context.close()
        this.context = null
      } catch (error) {
        console.warn('[AudioEngine] 关闭音频上下文失�?', error)
      }
    }
  }
}

// 导出单例实例
export const audioEngine = new AudioEngine()
export default audioEngine
