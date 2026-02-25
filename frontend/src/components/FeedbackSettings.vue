<template>
  <div class="fixed bottom-4 right-4 z-[110]">
    <!-- 浮动按钮 -->
    <button
      v-if="!showPanel"
      @click="showPanel = true"
      class="w-14 h-14 bg-gradient-to-br from-blue-600 to-purple-600 hover:from-blue-500 hover:to-purple-500 rounded-full shadow-lg flex items-center justify-center text-white transition-all hover:scale-110 active:scale-95"
    >
      <Settings class="w-6 h-6" />
    </button>

    <!-- 设置面板 -->
    <transition name="panel">
      <div
        v-if="showPanel"
        class="bg-white/95 dark:bg-slate-900/95 backdrop-blur-xl border border-slate-200 dark:border-white/10 rounded-2xl shadow-2xl p-4 w-80 animate-in slide-in-from-bottom-4"
      >
        <!-- 标题栏 -->
        <div class="flex items-center justify-between mb-4">
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-600 to-purple-600 flex items-center justify-center">
              <Settings class="w-4 h-4 text-white" />
            </div>
            <h3 class="text-base font-black text-slate-800 dark:text-white uppercase tracking-wider">
              反馈设置
            </h3>
          </div>
          <button
            @click="showPanel = false"
            class="p-2 hover:bg-slate-100 dark:hover:bg-white/10 rounded-lg transition-colors text-slate-500 hover:text-slate-700 dark:hover:text-white"
          >
            <X class="w-4 h-4" />
          </button>
        </div>

        <!-- 音效开关 -->
        <div class="space-y-3">
          <div class="flex items-center justify-between p-3 bg-slate-50 dark:bg-white/5 rounded-xl">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-lg bg-blue-500/10 flex items-center justify-center">
                <Volume2 v-if="settings.soundEnabled" class="w-5 h-5 text-blue-500" />
                <VolumeX v-else class="w-5 h-5 text-slate-400" />
              </div>
              <div>
                <p class="text-sm font-bold text-slate-700 dark:text-white">音效</p>
                <p class="text-xs text-slate-500">操作提示音</p>
              </div>
            </div>
            <button
              @click="toggleSound"
              :class="[
                'relative w-12 h-6 rounded-full transition-colors',
                settings.soundEnabled ? 'bg-blue-500' : 'bg-slate-300 dark:bg-slate-600'
              ]"
            >
              <div
                :class="[
                  'absolute top-0.5 w-5 h-5 rounded-full bg-white shadow-md transition-transform',
                  settings.soundEnabled ? 'translate-x-6' : 'translate-x-0.5'
                ]"
              ></div>
            </button>
          </div>

          <!-- 音量滑块 -->
          <div
            v-if="settings.soundEnabled"
            class="p-3 bg-slate-50 dark:bg-white/5 rounded-xl animate-in slide-in-from-top-2"
          >
            <div class="flex items-center justify-between mb-2">
              <span class="text-xs font-bold text-slate-600 dark:text-slate-400">音量</span>
              <span class="text-xs font-mono font-bold text-blue-500">{{ Math.round(settings.volume * 100) }}%</span>
            </div>
            <input
              type="range"
              min="0"
              max="100"
              :value="settings.volume * 100"
              @input="updateVolume"
              class="w-full h-2 bg-slate-200 dark:bg-slate-700 rounded-full appearance-none cursor-pointer
                     [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-4 [&::-webkit-slider-thumb]:h-4
                     [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-blue-500 [&::-webkit-slider-thumb]:cursor-pointer
                     [&::-webkit-slider-thumb]:shadow-lg [&::-webkit-slider-thumb]:hover:scale-110 [&::-webkit-slider-thumb]:transition-transform
                     [&::-moz-range-thumb]:w-4 [&::-moz-range-thumb]:h-4 [&::-moz-range-thumb]:rounded-full
                     [&::-moz-range-thumb]:bg-blue-500 [&::-moz-range-thumb]:border-0 [&::-moz-range-thumb]:cursor-pointer"
            />
          </div>

          <!-- 震动开关 -->
          <div class="flex items-center justify-between p-3 bg-slate-50 dark:bg-white/5 rounded-xl">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-lg bg-purple-500/10 flex items-center justify-center">
                <Smartphone v-if="settings.vibrationEnabled" class="w-5 h-5 text-purple-500" />
                <SmartphoneNfc v-else class="w-5 h-5 text-slate-400" />
              </div>
              <div>
                <p class="text-sm font-bold text-slate-700 dark:text-white">震动</p>
                <p class="text-xs text-slate-500">触觉反馈</p>
              </div>
            </div>
            <button
              @click="toggleVibration"
              :class="[
                'relative w-12 h-6 rounded-full transition-colors',
                settings.vibrationEnabled ? 'bg-purple-500' : 'bg-slate-300 dark:bg-slate-600'
              ]"
            >
              <div
                :class="[
                  'absolute top-0.5 w-5 h-5 rounded-full bg-white shadow-md transition-transform',
                  settings.vibrationEnabled ? 'translate-x-6' : 'translate-x-0.5'
                ]"
              ></div>
            </button>
          </div>

          <!-- 测试按钮 -->
          <button
            @click="testFeedback"
            class="w-full py-3 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-500 hover:to-purple-500 text-white font-bold rounded-xl transition-all active:scale-95 flex items-center justify-center gap-2"
          >
            <Play class="w-4 h-4" />
            测试反馈效果
          </button>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Settings, X, Volume2, VolumeX, Smartphone, SmartphoneNfc, Play } from 'lucide-vue-next'
import feedback from '../utils/feedback'

const showPanel = ref(false)
const settings = ref({
  soundEnabled: true,
  vibrationEnabled: true,
  volume: 0.3
})

onMounted(() => {
  settings.value = feedback.getSettings()
})

const toggleSound = () => {
  settings.value.soundEnabled = !settings.value.soundEnabled
  feedback.setSoundEnabled(settings.value.soundEnabled)
  if (settings.value.soundEnabled) {
    feedback.playSound('success')
  }
}

const toggleVibration = () => {
  settings.value.vibrationEnabled = !settings.value.vibrationEnabled
  feedback.setVibrationEnabled(settings.value.vibrationEnabled)
  if (settings.value.vibrationEnabled) {
    feedback.vibrate('medium')
  }
}

const updateVolume = (event: Event) => {
  const target = event.target as HTMLInputElement
  const volume = parseInt(target.value) / 100
  settings.value.volume = volume
  feedback.setVolume(volume)
}

const testFeedback = () => {
  console.log('[测试] 触发反馈测试')
  feedback.diagnoseVibration() // 运行诊断
  feedback.feedback({
    sound: 'success',
    vibration: 'success'
  })
}
</script>

<style scoped>
.panel-enter-active,
.panel-leave-active {
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.panel-enter-from {
  opacity: 0;
  transform: translateY(20px) scale(0.9);
}

.panel-leave-to {
  opacity: 0;
  transform: translateY(10px) scale(0.95);
}
</style>
