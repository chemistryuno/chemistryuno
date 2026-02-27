/**
 * 振动引擎 - 使用 Vibration API 提供触觉反馈
 * 独立的振动处理模块，与反馈系统解�?
 */

export type VibrationPattern = number | number[] | ReadonlyArray<number>

// 预设振动模式
export const VIBRATION_PATTERNS = {
  light: 10,                    // 轻触
  medium: 20,                   // 中等
  heavy: 30,                    // 重击
  double: [20, 50, 20],         // 双击
  success: [10, 50, 10, 50, 20], // 成功
  error: [30, 100, 30],         // 错误
  reaction: [15, 30, 15, 30, 25], // 反应
} as const

export type VibrationPreset = keyof typeof VIBRATION_PATTERNS

export class VibrationEngine {
  private enabled: boolean = true

  /**
   * 设置启用状�?
   */
  setEnabled(enabled: boolean) {
    this.enabled = enabled
    console.log('[VibrationEngine] 振动', enabled ? '启用' : '禁用')
  }

  /**
   * 检查振�?API 是否可用
   */
  isSupported(): boolean {
    return 'vibrate' in navigator
  }

  /**
   * 触发振动
   */
  vibrate(pattern: VibrationPattern | VibrationPreset) {
    if (!this.enabled) {
      console.log('[VibrationEngine] Vibration is disabled')
      return false
    }

    if (!this.isSupported()) {
      console.warn('[VibrationEngine] 当前浏览器不支持振动API')
      return false
    }

    // 解析振动模式
    const vibrationPattern = typeof pattern === 'string'
      ? VIBRATION_PATTERNS[pattern]
      : pattern

    try {
      const normalizedPattern: number | number[] = Array.isArray(vibrationPattern)
        ? Array.from(vibrationPattern)
        : (vibrationPattern as number)
      const success = navigator.vibrate(normalizedPattern)
      if (success) {
        console.log('[VibrationEngine] 振动触发成功:', normalizedPattern)
      } else {
        console.warn('[VibrationEngine] 振动触发失败 - 可能需要用户交互或权限')
      }
      return success
    } catch (error) {
      console.error('[VibrationEngine] 振动API调用失败:', error)
      return false
    }
  }

  /**
   * 停止所有振�?
   */
  stop() {
    if (this.isSupported()) {
      navigator.vibrate(0)
      console.log('[VibrationEngine] Stopped all vibrations')
    }
  }

  /**
   * 诊断振动功能
   */
  diagnose() {
    const report = {
      enabled: this.enabled,
      apiAvailable: this.isSupported(),
      userAgent: navigator.userAgent,
      platform: navigator.platform,
    }

    console.log('=== 振动功能诊断 ===')
    console.log('振动启用:', report.enabled)
    console.log('API可用:', report.apiAvailable)
    console.log('平台:', report.platform)
    console.log('用户代理:', report.userAgent)

    if (!report.apiAvailable) {
      console.warn('⚠️ 振动API不可�?- 可能原因:')
      console.warn('  1. 桌面浏览器不支持振动')
      console.warn('  2. iOS Safari 不支持振动API')
      console.warn('  3. �������ȫ����')
    } else if (!report.enabled) {
      console.warn('⚠️ 振动已被用户禁用')
    } else {
      console.log('�?振动功能应该正常工作')
      // 尝试触发测试振动
      try {
        const success = navigator.vibrate(200)
        console.log('测试振动结果:', success ? '成功' : '失败')
      } catch (e) {
        console.error('测试振动异常:', e)
      }
    }

    return report
  }
}

// 导出单例实例
export const vibrationEngine = new VibrationEngine()
export default vibrationEngine


