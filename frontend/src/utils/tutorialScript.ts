/**
 * 教学关卡脚本配置
 * 定义固定的出牌顺序和提示
 */

export interface TutorialStep {
  stepNumber: number
  player: 'human' | 'ai'
  action: 'play' | 'double'
  substance?: string
  substances?: string[] // 用于双元素
  hint: string
  aiMessage?: string
}

export const TUTORIAL_SCRIPT: TutorialStep[] = [
  {
    stepNumber: 1,
    player: 'human',
    action: 'play',
    substance: 'Mg',
    hint: '💡 第一步：从手牌中选择 <strong>Mg</strong>（镁）打出，它可以和场上的 Cl₂ 反应生成 MgCl₂'
  },
  {
    stepNumber: 2,
    player: 'ai',
    action: 'play',
    substance: 'HCl',
    hint: '⚗️ AI 的回合',
    aiMessage: 'AI 打出了 HCl（盐酸）'
  },
  {
    stepNumber: 3,
    player: 'human',
    action: 'play',
    substance: 'NaOH',
    hint: '💡 第二步：使用 <strong>Na</strong>、<strong>O</strong>、<strong>H</strong> 组合打出 <strong>NaOH</strong>（氢氧化钠），中和 AI 的 HCl'
  },
  {
    stepNumber: 4,
    player: 'ai',
    action: 'play',
    substance: 'Br2',
    hint: '⚗️ AI 的回合',
    aiMessage: 'AI 打出了 Br₂（溴单质）'
  },
  {
    stepNumber: 5,
    player: 'human',
    action: 'play',
    substance: 'Ar',
    hint: '💡 第三步：打出 <strong>Ar</strong>（氩气），触发稳定性效果跳过 AI'
  },
  {
    stepNumber: 6,
    player: 'ai',
    action: 'play',
    substance: '',
    hint: '⚗️ AI 无法响应，正在摸牌...',
    aiMessage: 'AI 选择摸牌'
  },
  {
    stepNumber: 7,
    player: 'human',
    action: 'play',
    substance: 'Au',
    hint: '💡 第四步：打出 <strong>Au</strong>（金），金属性稳定，可接在任何牌之后'
  },
  {
    stepNumber: 8,
    player: 'human',
    action: 'play',
    substance: '+2',
    hint: '💡 最后一步：打出 <strong>+2</strong>，完成本次教学关卡！'
  }
]

export const TUTORIAL_TOTAL_STEPS = TUTORIAL_SCRIPT.length

export interface TutorialInitialState {
  humanHand: string[]
  aiHand: string[]
  discardTop: string
}

export const TUTORIAL_INITIAL_STATE: TutorialInitialState = {
  humanHand: ['Na', 'Mg', 'O', 'H', 'Au', 'Ar', 'K'],
  aiHand: ['H', 'Cl', 'Br', 'Mn', 'Fe', 'Zn', 'Ca'],
  discardTop: 'Cl2'
}

/**
 * 获取当前步骤
 */
export const getTutorialStep = (stepNumber: number): TutorialStep | undefined => {
  return TUTORIAL_SCRIPT.find(step => step.stepNumber === stepNumber)
}

/**
 * 检查玩家是否可以打出指定物质
 */
export const canPlaySubstance = (substance: string, currentStep: number): boolean => {
  const step = getTutorialStep(currentStep)
  if (!step || step.player !== 'human') return false
  return step.substance === substance
}

/**
 * 获取教学进度描述
 */
export const getTutorialProgress = (currentStep: number): string => {
  const totalSteps = TUTORIAL_SCRIPT.filter(s => s.player === 'human').length
  const completedSteps = TUTORIAL_SCRIPT.filter(s =>
    s.player === 'human' && s.stepNumber < currentStep
  ).length
  return `${completedSteps}/${totalSteps}`
}
