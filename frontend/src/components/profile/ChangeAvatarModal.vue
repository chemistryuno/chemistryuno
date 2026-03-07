<script setup lang="ts">
import { ref, watch, onUnmounted, nextTick } from 'vue'
import { 
  Upload, 
  Loader2, 
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
  Crop
} from 'lucide-vue-next'
import { cn } from '../../utils/cn'
import Cropper from 'cropperjs'

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
const imageToCrop = ref<HTMLImageElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
let cropper: Cropper | null = null

const avatarOptions = [
  { id: 'flask', icon: FlaskConical },
  { id: 'dna', icon: Dna },
  { id: 'tube', icon: TestTube2 },
  { id: 'micro', icon: Microscope },
  { id: 'sat', icon: Satellite },
  { id: 'rocket', icon: Rocket },
  { id: 'orbit', icon: Orbit },
  { id: 'atom', icon: Atom },
  { id: 'radio', icon: Radio },
  { id: 'brain', icon: Brain },
  { id: 'bot', icon: Bot },
  { id: 'ghost', icon: Ghost }
]

const initCropper = () => {
  if (cropper) {
    cropper.destroy()
  }
  if (imageToCrop.value) {
    cropper = new Cropper(imageToCrop.value, {
      aspectRatio: 1,
      viewMode: 1,
      dragMode: 'move',
      autoCropArea: 1,
      restore: false,
      guides: true,
      center: true,
      highlight: false,
      cropBoxMovable: true,
      cropBoxResizable: true,
      toggleDragModeOnDblclick: false,
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
    await nextTick()
    initCropper()
  }
  reader.readAsDataURL(file)
}

const handleSave = () => {
  if (cropper && previewImage.value) {
    const canvas = cropper.getCroppedCanvas({
      width: 400,
      height: 400,
    })
    const croppedDataUrl = canvas.toDataURL('image/webp')
    emit('save', croppedDataUrl)
  } else {
    emit('save', selectedAvatar.value)
  }
}

const clearCrop = () => {
  previewImage.value = null
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
const isPreset = (val: string) => {
  return avatarOptions.some(opt => opt.id === val)
}

// 获取内置图标组件
const getPresetIcon = (id: string) => {
  return avatarOptions.find(opt => opt.id === id)?.icon || FlaskConical
}

</script>

<template>
  <div v-if="show" class="fixed inset-0 z-[100] flex items-center justify-center p-4 backdrop-blur-xl bg-slate-900/40 dark:bg-black/80">
    <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[3rem] p-10 max-w-2xl w-full shadow-2xl relative animate-in fade-in zoom-in duration-300">
      <h3 class="text-2xl font-black mb-8 italic uppercase text-center text-slate-900 dark:text-white">选择新的身份标识 / Select Avatar</h3>
      
      <div class="flex flex-col items-center gap-8 mb-10">
        <!-- 预览与裁剪区 -->
        <div class="relative group/preview w-full flex justify-center">
          <div :class="cn(
            'w-48 h-48 bg-slate-50 dark:bg-[#1a1c1e] rounded-[2.5rem] border-2 border-dashed border-slate-200 dark:border-white/10 flex items-center justify-center overflow-hidden transition-all',
            previewImage ? 'w-full h-64' : 'group-hover/preview:border-blue-500/50'
          )">
             <template v-if="previewImage">
                <img ref="imageToCrop" :src="previewImage" class="max-w-full block" />
             </template>
             <template v-else-if="selectedAvatar && selectedAvatar.startsWith('data:')">
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
          
          <div class="absolute -bottom-2 right-1/2 translate-x-24 flex gap-2">
            <button 
              v-if="previewImage"
              @click="clearCrop"
              class="bg-red-500 p-2 rounded-xl text-white shadow-lg shadow-red-500/20 hover:scale-110 transition-transform"
              title="取消裁剪"
            >
              <Ghost class="w-4 h-4" />
            </button>
            <button 
              @click="fileInput?.click()"
              class="bg-blue-600 p-2 rounded-xl text-white shadow-lg shadow-blue-500/20 hover:scale-110 transition-transform"
              title="上传图片"
            >
              <Upload class="w-4 h-4" />
            </button>
          </div>

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
                selectedAvatar === option.id && !previewImage
                  ? 'bg-blue-600 border-blue-400 scale-110 shadow-[0_0_20px_rgba(59,130,246,0.5)] text-white' 
                  : 'bg-slate-50 dark:bg-white/5 border-slate-100 dark:border-transparent hover:border-blue-300 dark:hover:border-white/20 hover:scale-105 text-slate-600 dark:text-slate-400'
              )"
            >
              <component :is="option.icon" class="w-8 h-8" />
            </button>
          </div>
        </div>
        
        <!-- 上传协议/按钮 -->
        <button 
          @click="fileInput?.click()"
          class="bg-blue-500/5 hover:bg-blue-500/10 border border-blue-500/10 rounded-2xl p-4 w-full flex items-center gap-4 transition-colors group/upload"
        >
          <div class="p-2 bg-blue-500/10 rounded-xl text-blue-600 dark:text-blue-400 group-hover/upload:scale-110 transition-transform">
            <Upload class="w-4 h-4" />
          </div>
          <div class="flex flex-col text-left">
            <span class="text-xs font-bold text-slate-900 dark:text-white">本地图像上传协议 (MAX 10MB)</span>
            <span class="text-[10px] text-slate-500">自动进入裁剪模式，支持 JPG, PNG, WEBP</span>
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
          {{ previewImage ? '裁剪并同步' : '同步身份更改' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 确保裁剪区域容器样式正确 */
img {
  max-width: 100%;
}
</style>
