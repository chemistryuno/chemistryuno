/**
 * 教程管理 Composable
 * 负责处理游戏教程逻辑
 */

import { ref, computed, watch, type Ref } from 'vue'
import { getTutorialStep, canPlaySubstance, getTutorialProgress } from '../utils/tutorialScript'

export function useTutorial(
  roomInfo: Ref<any>,
  gameState: Ref<any>,
  isMyTurn: Ref<boolean>
) {
  // 状态
  const isTutorialMode = ref(false)
  const tutorialHintText = ref('')
  const tutorialCurrentStep = ref(1)
  const tutorialScriptMode = ref(false)

  // 计算属性
  const tutorialProgress = computed(() => {
    if (!tutorialScriptMode.value) return null
    return getTutorialProgress(tutorialCurrentStep.value)
  })

  // 方法
  const checkTutorialMode = () => {
    if (roomInfo.value?.is_pve && roomInfo.value?.pve_difficulty === 0) {
      isTutorialMode.value = true
      tutorialScriptMode.value = true
      console.log('[教程] 启用教程模式')
    }
  }

  const generateTutorialHint = () => {
    if (!tutorialScriptMode.value || !isMyTurn.value) {
      tutorialHintText.value = ''
      return
    }

    const currentStep = getTutorialStep(tutorialCurrentStep.value)
    if (!currentStep) {
      tutorialHintText.value = '教程已完成！'
      return
    }

    let hint = `步骤 ${tutorialCurrentStep.value}: ${currentStep.hint}`
    if (currentStep.substance) {
      hint += ` → 打出 ${currentStep.substance}`
    }
    tutorialHintText.value = hint
  }

  const validateTutorialAction = (substance: string): boolean => {
    if (!tutorialScriptMode.value) return true

    const currentStep = getTutorialStep(tutorialCurrentStep.value)
    if (!currentStep) return true

    return canPlaySubstance(substance, tutorialCurrentStep.value)
  }

  const advanceTutorialStep = () => {
    if (tutorialScriptMode.value) {
      tutorialCurrentStep.value++
      generateTutorialHint()
    }
  }

  const resetTutorial = () => {
    tutorialCurrentStep.value = 1
    tutorialHintText.value = ''
    generateTutorialHint()
  }

  // 监听游戏状态变化
  watch([() => roomInfo.value, () => gameState.value?.status], () => {
    checkTutorialMode()
  }, { immediate: true })

  watch(isMyTurn, (val) => {
    if (val && isTutorialMode.value) {
      generateTutorialHint()
    }
  })

  return {
    // 状态
    isTutorialMode,
    tutorialHintText,
    tutorialCurrentStep,
    tutorialScriptMode,

    // 计算属性
    tutorialProgress,

    // 方法
    checkTutorialMode,
    generateTutorialHint,
    validateTutorialAction,
    advanceTutorialStep,
    resetTutorial,
  }
}
