<script setup lang="ts">
import { ref, watch, onUnmounted, nextTick } from 'vue'
import { Upload, Loader2, Crop, Check } from 'lucide-vue-next'
import { cn } from '../../utils/cn'
import Cropper from 'cropperjs'
import 'cropperjs/dist/cropper.css'
import { AVATAR_PRESETS, isPresetAvatar } from '../../utils/avatarPresets'

const props = defineProps<{
  show: boolean
  currentAvatar: string
  loading: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'save', avatar: string): void
}>()

const selectedAvatar = ref(props.currentAvatar)
const previewImage = ref<string | null>(null)
const showCropperModal = ref(false)
const imageToCrop = ref<HTMLImageElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
let cropper: Cropper | null = null

const avatarOptions = Object.entries(AVATAR_PRESETS).map(([id, icon]) => ({ id, icon }))

const initCropper = () => {
  if (cropper) {
    cropper.destroy()
  }
  if (imageToCrop.value) {
    cropper = new Cropper(imageToCrop.value, {
      aspectRatio: 1,      // 锁定 1:1 正方形
      viewMode: 1,         // 限制裁剪框不能移出图片范围
      dragMode: 'move',    // 允许拖动图片
      autoCropArea: 1,     // 初始裁剪区域 100% (铺满)
      restore: false,
      guides: false,       
      center: true,
      highlight: false,
      responsive: true,
      modal: true, 
      movable: true,       
      zoomable: true,      
      scalable: true,
      rotatable: true,     // 允许旋转以适应某些手机拍摄
      cropBoxMovable: false, // 禁止移动裁剪框，让玩家通过移动图片来调整
      cropBoxResizable: false, // 禁止缩放裁剪框，固定为正方形
      toggleDragModeOnDblclick: false,
      checkOrientation: true,
    })
  }
}

const handleFileUpload = (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  // 10MB limit (10 * 1024 * 1024)
  if (file.size > 10 * 1024 * 1024) {
    alert('文件大小超过 10MB 限制 / File too large (Max 10MB)')
    return
  }
  
  if (file.size === 0) {
    alert('文件内容不能为空 / File is empty')
    return
  }

  const reader = new FileReader()
  reader.onload = async (e) => {
    previewImage.value = e.target?.result as string
    showCropperModal.value = true
    await nextTick()
    initCropper()
  }
  reader.readAsDataURL(file)
}

const confirmCrop = () => {
  if (cropper && previewImage.value) {
    const canvas = cropper.getCroppedCanvas({
      width: 400,
      height: 400,
      imageSmoothingEnabled: true,
      imageSmoothingQuality: 'high',
    })
    
    if (canvas) {
      selectedAvatar.value = canvas.toDataURL('image/webp', 0.5)
      showCropperModal.value = false
      previewImage.value = null
      if (cropper) {
        cropper.destroy()
        cropper = null
      }
    }
  }
}

const handleSave = () => {
  emit('save', selectedAvatar.value)
}

const clearCrop = () => {
  previewImage.value = null
  showCropperModal.value = false
  if (cropper) {
    cropper.destroy()
    cropper = null
  }
}

const selectPreset = (id: string) => {
  clearCrop()
  selectedAvatar.value = id
}

onUnmounted(() => {
  if (cropper) {
    cropper.destroy()
  }
})

watch(() => props.show, (newVal) => {
  if (!newVal) {
    clearCrop()
  }
})

// 判断是否为内置图标 ID
const isPreset = isPresetAvatar

// 获取内置图标组件
const getPresetIcon = (id: string) => AVATAR_PRESETS[id] ?? avatarOptions[0].icon

</script>

<template>
  <div v-if="show" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md" @click="$emit('close')">
    <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[3rem] p-10 max-w-2xl w-full shadow-2xl relative animate-in fade-in zoom-in duration-300 pointer-events-auto" @click.stop>
      <h3 class="text-2xl font-black mb-8 italic uppercase text-center text-slate-900 dark:text-white">选择新的身份标识 / Select Avatar</h3>
      
      <div class="flex flex-col items-center gap-8 mb-10">
        <!-- 预览区 -->
        <div class="relative group/preview w-full flex justify-center">
          <div :class="cn(
            'w-48 h-48 bg-slate-50 dark:bg-[#1a1c1e] rounded-[2.5rem] border-2 border-dashed border-slate-200 dark:border-white/10 flex items-center justify-center overflow-hidden transition-all group-hover/preview:border-blue-500/50'
          )">
             <template v-if="selectedAvatar && selectedAvatar.startsWith('data:')">
                <img :src="selectedAvatar" class="w-full h-full object-cover" />
             </template>
             <template v-else-if="isPreset(selectedAvatar)">
                <component :is="getPresetIcon(selectedAvatar)" class="w-24 h-24 text-blue-600 dark:text-blue-400" />
             </template>
             <template v-else>
                <!-- 兼容旧的 Emoji 头像 -->
                <span class="text-6xl">{{ selectedAvatar || '🧪' }}</span>
             </template>
          </div>
          
          <button 
            @click="fileInput?.click()"
            class="absolute -bottom-2 right-1/2 translate-x-24 bg-blue-600 p-2 rounded-xl text-white shadow-lg shadow-blue-500/20 hover:scale-110 transition-transform"
            title="上传图片"
          >
            <Upload class="w-4 h-4" />
          </button>

          <input 
            type="file" 
            ref="fileInput" 
            class="hidden" 
            accept="image/*"
            @change="handleFileUpload"
          />
        </div>

        <!-- 预设选择器 -->
        <div class="w-full">
          <p class="text-[10px] font-black uppercase text-slate-500 tracking-[0.2em] mb-4 text-center">快捷原型选择器 / Quick Lab Presets</p>
          <div class="grid grid-cols-4 sm:grid-cols-6 gap-4">
            <button
              v-for="option in avatarOptions"
              :key="option.id"
              @click="selectPreset(option.id)"
              :class="cn(
                'w-16 h-16 flex items-center justify-center rounded-[1.5rem] transition-all duration-300 border-2',
                selectedAvatar === option.id
                  ? 'bg-blue-600 border-blue-400 scale-110 shadow-[0_0_20px_rgba(59,130,246,0.5)] text-white' 
                  : 'bg-slate-50 dark:bg-white/5 border-slate-100 dark:border-transparent hover:border-blue-300 dark:hover:border-white/20 hover:scale-105 text-slate-600 dark:text-slate-400'
              )"
            >
              <component :is="option.icon" class="w-8 h-8" />
            </button>
          </div>
        </div>
        
        <!-- 上传按钮 -->
        <button 
          @click="fileInput?.click()"
          class="bg-blue-500/5 hover:bg-blue-500/10 border border-blue-500/10 rounded-2xl p-4 w-full flex items-center gap-4 transition-colors group/upload"
        >
          <div class="p-2 bg-blue-500/10 rounded-xl text-blue-600 dark:text-blue-400 group-hover/upload:scale-110 transition-transform">
            <Upload class="w-4 h-4" />
          </div>
          <div class="flex flex-col text-left">
            <span class="text-xs font-bold text-slate-900 dark:text-white">本地图像上传协议 (MAX 10MB)</span>
            <span class="text-[10px] text-slate-500">上传后进入独立裁剪窗口，支持 JPG, PNG, WEBP</span>
          </div>
        </button>
      </div>

      <div class="flex gap-4">
        <button 
          @click="$emit('close')"
          class="flex-1 py-4 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 rounded-2xl font-bold transition-all text-slate-500 dark:text-slate-400"
        >
          取消
        </button>
        <button 
          @click="handleSave"
          :disabled="loading"
          class="flex-1 py-4 bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 rounded-2xl font-black text-white shadow-xl shadow-blue-500/20 disabled:opacity-50 flex items-center justify-center gap-2"
        >
          <Loader2 v-if="loading" class="w-5 h-5 animate-spin" />
          同步身份更改
        </button>
      </div>
    </div>

    <!-- 独立裁剪弹窗 -->
    <div v-if="showCropperModal" class="fixed inset-0 z-[110] flex items-center justify-center p-2 sm:p-4 bg-black/80 dark:bg-black/95 backdrop-blur-2xl animate-in fade-in duration-300">
      <div class="bg-white dark:bg-[#111114] border border-white/10 rounded-[2rem] sm:rounded-[3rem] p-4 sm:p-8 max-w-xl w-full shadow-2xl relative flex flex-col h-full max-h-[85vh] sm:h-auto pointer-events-auto">
        <h4 class="text-lg sm:text-xl font-black mb-4 sm:mb-6 text-center text-slate-900 dark:text-white uppercase italic flex items-center justify-center gap-2">
          <Crop class="w-5 h-5 text-blue-500" />
          裁切头像 / CROP
        </h4>
        
        <div class="flex-1 min-h-0 bg-black/40 rounded-[1.5rem] sm:rounded-[2rem] overflow-hidden mb-4 sm:mb-8 border border-white/5 relative shadow-inner">
          <div class="absolute inset-0 flex items-center justify-center pointer-events-none z-10 opacity-30">
            <div class="border-2 border-dashed border-white/20 w-[80%] aspect-square rounded-full"></div>
          </div>
          <img ref="imageToCrop" :src="previewImage!" class="max-w-full block" />
        </div>

        <div class="text-[10px] text-center text-slate-500 dark:text-slate-400 mb-4 italic opacity-60">
          支持手势旋转缩放 · 自动锁定正方形区域
        </div>

        <div class="flex gap-3 sm:gap-4 mt-auto">
          <button 
            @click="clearCrop"
            class="flex-1 py-3 sm:py-4 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 rounded-xl sm:rounded-2xl font-bold transition-all text-slate-500 dark:text-slate-400 active:scale-95"
          >
            取消
          </button>
          <button 
            @click="confirmCrop"
            class="flex-1 py-3 sm:py-4 bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 rounded-xl sm:rounded-2xl font-black text-white shadow-xl shadow-blue-500/20 transition-all active:scale-95 flex items-center justify-center gap-2"
          >
            <Check class="w-5 h-5" />
            裁切并保存
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 确保裁剪区域容器样式正确 */
img {
  max-width: 100%;
}

/* 深度选择器修改 Cropper.js 样式 */
:deep(.cropper-view-box) {
  outline: 4px solid #3b82f6;
  outline-color: rgba(59, 130, 246, 0.8);
  border-radius: 4px; /* 稍微圆角配合整体 UI */
}

:deep(.cropper-line) {
  background-color: #3b82f6;
  height: 2px;
  opacity: 0.5;
}

:deep(.cropper-point) {
  background-color: #3b82f6;
  width: 12px !important;
  height: 12px !important;
  opacity: 1;
  border: 2px solid white;
  box-shadow: 0 0 10px rgba(59, 130, 246, 0.5);
}

/* 针对四个角 (Corners) 进行特别加粗处理 */
:deep(.cropper-point.point-se),
:deep(.cropper-point.point-sw),
:deep(.cropper-point.point-ne),
:deep(.cropper-point.point-nw) {
  width: 24px !important;
  height: 24px !important;
  border-radius: 6px;
  background-color: #2563eb;
}

/* 隐藏不必要的中间控制点，让四个角最显眼 */
:deep(.cropper-point.point-n),
:deep(.cropper-point.point-w),
:deep(.cropper-point.point-s),
:deep(.cropper-point.point-e) {
  display: none !important;
}

:deep(.cropper-face) {
  background-color: transparent;
  opacity: 0;
}
</style>
