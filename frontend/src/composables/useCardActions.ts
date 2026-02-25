/**
 * 卡牌操作 Composable
 * 负责处理卡牌点击、打牌、摸牌等操作
 */

import { ref, type Ref } from 'vue'
import { gameAPI } from '../utils/api'
import feedback from '../utils/feedback'

export function useCardActions(roomId: string, gameState: Ref<any>) {
  // 状态
  const selectedCard = ref<any>(null)
  const selectedSubstance = ref<string | null>(null)
  const availableSubstances = ref<string[]>([])
  const doubleMode = ref(false)
  const firstDoubleSubstance = ref<string | null>(null)
  const secondDoubleSubstance = ref<string | null>(null)
  const substanceInput = ref('')

  // 方法
  const handleCardClick = async (card: any) => {
    const specialTypes = ['+2', '+4', 'Au', 'He', 'Ne', 'Ar', 'Kr']

    feedback.click()

    if (specialTypes.includes(card.type) || card.effect) {
      try {
        await gameAPI.playCard(roomId, card, card.type)
        feedback.playCard()
        resetSelection()
        return
      } catch (error: any) {
        feedback.error()
        throw error
      }
    }

    // 元素牌：直接出牌该元素符号
    try {
      await gameAPI.playCard(roomId, card, card.type)
      feedback.playCard()
      resetSelection()
    } catch (error: any) {
      feedback.error()
      throw error
    }
  }

  const handlePlayCard = async (substance: string) => {
    if (doubleMode.value) {
      if (!firstDoubleSubstance.value) {
        firstDoubleSubstance.value = substance
      } else if (!secondDoubleSubstance.value) {
        secondDoubleSubstance.value = substance
      }
      selectedCard.value = null
      selectedSubstance.value = null
      availableSubstances.value = []
      return
    }

    try {
      const cardToPlay = selectedCard.value || { type: substance, count: 1, effect: '' }
      await gameAPI.playCard(roomId, cardToPlay, substance)
      feedback.playCard()
      resetSelection()
    } catch (error: any) {
      feedback.error()
      throw error
    }
  }

  const handleDoublePlay = async () => {
    if (!firstDoubleSubstance.value || !secondDoubleSubstance.value) {
      feedback.error()
      throw new Error('请选择参与双联反应的两种物质')
    }

    try {
      await gameAPI.playDouble(roomId, firstDoubleSubstance.value, secondDoubleSubstance.value)
      feedback.playCard()
      resetDoubleMode()
    } catch (error: any) {
      feedback.error()
      throw error
    }
  }

  const handleDrawCard = async () => {
    try {
      await gameAPI.drawCard(roomId)
      feedback.drawCard()
    } catch (error: any) {
      feedback.error()
      throw error
    }
  }

  const toggleDoubleMode = () => {
    const myData = gameState.value?.players.find((p: any) => {
      const user = JSON.parse(localStorage.getItem('user') || '{}')
      return p.uid === user.uid
    })

    if (!myData?.double_action_available) {
      feedback.error()
      throw new Error('双联反应尚未就绪，请先进行普通实验')
    }

    doubleMode.value = !doubleMode.value
    feedback.doubleMode()
    firstDoubleSubstance.value = null
    secondDoubleSubstance.value = null
    selectedSubstance.value = null
  }

  const removeSubstance = (pos: number) => {
    feedback.click()
    if (pos === 1) {
      firstDoubleSubstance.value = secondDoubleSubstance.value
      secondDoubleSubstance.value = null
    } else {
      secondDoubleSubstance.value = null
    }
  }

  const handleInputPlay = async () => {
    if (!substanceInput.value) return

    if (doubleMode.value) {
      const sub = substanceInput.value
      if (!firstDoubleSubstance.value) {
        firstDoubleSubstance.value = sub
      } else if (!secondDoubleSubstance.value) {
        secondDoubleSubstance.value = sub
      }
      substanceInput.value = ''
      return
    }

    try {
      await gameAPI.playCard(roomId, { type: '', count: 0, effect: '' }, substanceInput.value)
      feedback.playCard()
      substanceInput.value = ''
      resetSelection()
    } catch (error: any) {
      feedback.error()
      throw error
    }
  }

  const resetSelection = () => {
    selectedCard.value = null
    selectedSubstance.value = null
    availableSubstances.value = []
  }

  const resetDoubleMode = () => {
    firstDoubleSubstance.value = null
    secondDoubleSubstance.value = null
    doubleMode.value = false
    resetSelection()
  }

  return {
    // 状态
    selectedCard,
    selectedSubstance,
    availableSubstances,
    doubleMode,
    firstDoubleSubstance,
    secondDoubleSubstance,
    substanceInput,

    // 方法
    handleCardClick,
    handlePlayCard,
    handleDoublePlay,
    handleDrawCard,
    toggleDoubleMode,
    removeSubstance,
    handleInputPlay,
    resetSelection,
    resetDoubleMode,
  }
}
