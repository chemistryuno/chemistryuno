<template>
  <div class="reaction-toast-demo">
    <div class="demo-header">
      <h1 class="demo-title">⚗️ 化学反应通知系统</h1>
      <p class="demo-subtitle">Chemistry Reaction Toast Notifications</p>
    </div>

    <div class="demo-content">
      <div class="demo-section">
        <h2 class="section-title">反应类型演示</h2>
        <div class="demo-grid">
          <!-- 合成反应 -->
          <button
            @click="showSynthesis"
            class="demo-button synthesis"
          >
            <div class="button-icon">🔬</div>
            <div class="button-info">
              <div class="button-name">合成反应</div>
              <div class="button-desc">Synthesis</div>
            </div>
          </button>

          <!-- 分解反应 -->
          <button
            @click="showDecomposition"
            class="demo-button decomposition"
          >
            <div class="button-icon">💥</div>
            <div class="button-info">
              <div class="button-name">分解反应</div>
              <div class="button-desc">Decomposition</div>
            </div>
          </button>

          <!-- 置换反应 -->
          <button
            @click="showDisplacement"
            class="demo-button displacement"
          >
            <div class="button-icon">🔄</div>
            <div class="button-info">
              <div class="button-name">置换反应</div>
              <div class="button-desc">Displacement</div>
            </div>
          </button>

          <!-- 燃烧反应 -->
          <button
            @click="showCombustion"
            class="demo-button combustion"
          >
            <div class="button-icon">🔥</div>
            <div class="button-info">
              <div class="button-name">燃烧反应</div>
              <div class="button-desc">Combustion</div>
            </div>
          </button>

          <!-- 中和反应 -->
          <button
            @click="showNeutralization"
            class="demo-button neutralization"
          >
            <div class="button-icon">⚖️</div>
            <div class="button-info">
              <div class="button-name">中和反应</div>
              <div class="button-desc">Neutralization</div>
            </div>
          </button>

          <!-- 随机反应 -->
          <button
            @click="showRandom"
            class="demo-button random"
          >
            <div class="button-icon">🎲</div>
            <div class="button-info">
              <div class="button-name">随机反应</div>
              <div class="button-desc">Random</div>
            </div>
          </button>
        </div>
      </div>

      <div class="demo-section">
        <h2 class="section-title">批量测试</h2>
        <div class="demo-actions">
          <button @click="showSequence" class="action-button">
            <span class="action-icon">▶️</span>
            播放序列
          </button>
          <button @click="showBurst" class="action-button">
            <span class="action-icon">💫</span>
            连续爆发
          </button>
          <button @click="clearAll" class="action-button">
            <span class="action-icon">🗑️</span>
            清除所有
          </button>
        </div>
      </div>

      <div class="demo-section">
        <h2 class="section-title">配色方案</h2>
        <div class="color-palette">
          <div class="color-item">
            <div class="color-swatch synthesis-swatch"></div>
            <div class="color-label">
              <div class="color-name">Synthesis</div>
              <div class="color-code">#06b6d4 → #3b82f6</div>
            </div>
          </div>
          <div class="color-item">
            <div class="color-swatch decomposition-swatch"></div>
            <div class="color-label">
              <div class="color-name">Decomposition</div>
              <div class="color-code">#f97316 → #dc2626</div>
            </div>
          </div>
          <div class="color-item">
            <div class="color-swatch displacement-swatch"></div>
            <div class="color-label">
              <div class="color-name">Displacement</div>
              <div class="color-code">#a855f7 → #7e22ce</div>
            </div>
          </div>
          <div class="color-item">
            <div class="color-swatch combustion-swatch"></div>
            <div class="color-label">
              <div class="color-name">Combustion</div>
              <div class="color-code">#fb923c → #f59e0b</div>
            </div>
          </div>
          <div class="color-item">
            <div class="color-swatch neutralization-swatch"></div>
            <div class="color-label">
              <div class="color-name">Neutralization</div>
              <div class="color-code">#10b981 → #047857</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Toast 组件 -->
    <ReactionToast ref="toastRef" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import ReactionToast from './ReactionToast.vue'

const toastRef = ref<InstanceType<typeof ReactionToast> | null>(null)

// 反应示例数据
const reactions = {
  synthesis: [
    { equation: '2H₂ + O₂ → 2H₂O', name: 'Water Formation', energy: 85 },
    { equation: 'N₂ + 3H₂ → 2NH₃', name: 'Ammonia Synthesis', energy: 78 },
    { equation: 'C + O₂ → CO₂', name: 'Carbon Dioxide', energy: 90 }
  ],
  decomposition: [
    { equation: '2H₂O → 2H₂ + O₂', name: 'Water Electrolysis', energy: 75 },
    { equation: '2HgO → 2Hg + O₂', name: 'Mercury Oxide', energy: 82 },
    { equation: 'CaCO₃ → CaO + CO₂', name: 'Limestone', energy: 88 }
  ],
  displacement: [
    { equation: 'Zn + CuSO₄ → ZnSO₄ + Cu', name: 'Zinc Copper', energy: 72 },
    { equation: 'Fe + CuSO₄ → FeSO₄ + Cu', name: 'Iron Copper', energy: 68 },
    { equation: 'Mg + 2HCl → MgCl₂ + H₂', name: 'Magnesium Acid', energy: 80 }
  ],
  combustion: [
    { equation: 'CH₄ + 2O₂ → CO₂ + 2H₂O', name: 'Methane Burn', energy: 95 },
    { equation: 'C₃H₈ + 5O₂ → 3CO₂ + 4H₂O', name: 'Propane Burn', energy: 98 },
    { equation: '2C₈H₁₈ + 25O₂ → 16CO₂ + 18H₂O', name: 'Octane Burn', energy: 100 }
  ],
  neutralization: [
    { equation: 'HCl + NaOH → NaCl + H₂O', name: 'Acid Base', energy: 65 },
    { equation: 'H₂SO₄ + 2NaOH → Na₂SO₄ + 2H₂O', name: 'Sulfuric Acid', energy: 70 },
    { equation: 'CH₃COOH + NaOH → CH₃COONa + H₂O', name: 'Acetic Acid', energy: 60 }
  ]
}

// 显示合成反应
const showSynthesis = () => {
  const r = reactions.synthesis[Math.floor(Math.random() * reactions.synthesis.length)]
  toastRef.value?.showToast('synthesis', r.equation, r.name, r.energy)
}

// 显示分解反应
const showDecomposition = () => {
  const r = reactions.decomposition[Math.floor(Math.random() * reactions.decomposition.length)]
  toastRef.value?.showToast('decomposition', r.equation, r.name, r.energy)
}

// 显示置换反应
const showDisplacement = () => {
  const r = reactions.displacement[Math.floor(Math.random() * reactions.displacement.length)]
  toastRef.value?.showToast('displacement', r.equation, r.name, r.energy)
}

// 显示燃烧反应
const showCombustion = () => {
  const r = reactions.combustion[Math.floor(Math.random() * reactions.combustion.length)]
  toastRef.value?.showToast('combustion', r.equation, r.name, r.energy)
}

// 显示中和反应
const showNeutralization = () => {
  const r = reactions.neutralization[Math.floor(Math.random() * reactions.neutralization.length)]
  toastRef.value?.showToast('neutralization', r.equation, r.name, r.energy)
}

// 随机显示
const showRandom = () => {
  const types = ['synthesis', 'decomposition', 'displacement', 'combustion', 'neutralization'] as const
  const type = types[Math.floor(Math.random() * types.length)]
  const r = reactions[type][Math.floor(Math.random() * reactions[type].length)]
  toastRef.value?.showToast(type, r.equation, r.name, r.energy)
}

// 播放序列
const showSequence = () => {
  const sequence = [showSynthesis, showDecomposition, showDisplacement, showCombustion, showNeutralization]
  sequence.forEach((fn, index) => {
    setTimeout(fn, index * 800)
  })
}

// 连续爆发
const showBurst = () => {
  for (let i = 0; i < 5; i++) {
    setTimeout(showRandom, i * 400)
  }
}

// 清除所有（刷新页面实现）
const clearAll = () => {
  window.location.reload()
}
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Orbitron:wght@700;900&family=JetBrains+Mono:wght@400;500;700&display=swap');

.reaction-toast-demo {
  min-height: 100vh;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #0f172a 100%);
  padding: 40px 20px;
  font-family: 'JetBrains Mono', monospace;
}

.demo-header {
  text-align: center;
  margin-bottom: 60px;
}

.demo-title {
  font-family: 'Orbitron', monospace;
  font-size: 48px;
  font-weight: 900;
  background: linear-gradient(135deg, #06b6d4, #3b82f6, #a855f7);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0 0 16px 0;
  text-shadow: 0 0 40px rgba(6, 182, 212, 0.5);
}

.demo-subtitle {
  font-size: 14px;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 4px;
  margin: 0;
}

.demo-content {
  max-width: 1200px;
  margin: 0 auto;
}

.demo-section {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 24px;
  padding: 32px;
  margin-bottom: 32px;
}

.section-title {
  font-family: 'Orbitron', monospace;
  font-size: 24px;
  font-weight: 700;
  color: #ffffff;
  margin: 0 0 24px 0;
  text-transform: uppercase;
  letter-spacing: 2px;
}

/* 演示按钮网格 */
.demo-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.demo-button {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: rgba(255, 255, 255, 0.05);
  border: 2px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.3s ease;
  font-family: inherit;
}

.demo-button:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
}

.demo-button:active {
  transform: translateY(-2px);
}

.demo-button.synthesis:hover {
  border-color: #06b6d4;
  box-shadow: 0 8px 24px rgba(6, 182, 212, 0.4);
}

.demo-button.decomposition:hover {
  border-color: #f97316;
  box-shadow: 0 8px 24px rgba(249, 115, 22, 0.4);
}

.demo-button.displacement:hover {
  border-color: #a855f7;
  box-shadow: 0 8px 24px rgba(168, 85, 247, 0.4);
}

.demo-button.combustion:hover {
  border-color: #fb923c;
  box-shadow: 0 8px 24px rgba(251, 146, 60, 0.4);
}

.demo-button.neutralization:hover {
  border-color: #10b981;
  box-shadow: 0 8px 24px rgba(16, 185, 129, 0.4);
}

.demo-button.random:hover {
  border-color: #ffffff;
  box-shadow: 0 8px 24px rgba(255, 255, 255, 0.2);
}

.button-icon {
  font-size: 32px;
  line-height: 1;
}

.button-info {
  flex: 1;
  text-align: left;
}

.button-name {
  font-size: 14px;
  font-weight: 700;
  color: #ffffff;
  margin-bottom: 4px;
}

.button-desc {
  font-size: 10px;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 1px;
}

/* 操作按钮 */
.demo-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.action-button {
  flex: 1;
  min-width: 160px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 16px 24px;
  background: rgba(255, 255, 255, 0.1);
  border: 2px solid rgba(255, 255, 255, 0.2);
  border-radius: 12px;
  color: #ffffff;
  font-family: inherit;
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.3s ease;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.action-button:hover {
  background: rgba(255, 255, 255, 0.15);
  transform: translateY(-2px);
}

.action-icon {
  font-size: 18px;
}

/* 配色方案 */
.color-palette {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 16px;
}

.color-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
}

.color-swatch {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.synthesis-swatch {
  background: linear-gradient(135deg, #06b6d4, #3b82f6);
}

.decomposition-swatch {
  background: linear-gradient(135deg, #f97316, #dc2626);
}

.displacement-swatch {
  background: linear-gradient(135deg, #a855f7, #7e22ce);
}

.combustion-swatch {
  background: linear-gradient(135deg, #fb923c, #f59e0b);
}

.neutralization-swatch {
  background: linear-gradient(135deg, #10b981, #047857);
}

.color-label {
  flex: 1;
}

.color-name {
  font-size: 12px;
  font-weight: 700;
  color: #ffffff;
  margin-bottom: 4px;
}

.color-code {
  font-size: 10px;
  color: #94a3b8;
  font-family: 'JetBrains Mono', monospace;
}

/* 响应式 */
@media (max-width: 768px) {
  .demo-title {
    font-size: 32px;
  }

  .demo-subtitle {
    font-size: 11px;
  }

  .demo-section {
    padding: 20px;
  }

  .section-title {
    font-size: 18px;
  }

  .demo-grid {
    grid-template-columns: 1fr;
  }

  .color-palette {
    grid-template-columns: 1fr;
  }
}
</style>
