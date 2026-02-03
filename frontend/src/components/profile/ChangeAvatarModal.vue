<script setup lang="ts">
import { ref } from 'vue'
import { Upload, Loader2 } from 'lucide-vue-next'
import { cn } from '../../utils/cn'

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
const fileInput = ref<HTMLInputElement | null>(null)
const avatarOptions = ["🧪", "🧬", "⚗️", "🔬", "🛰️", "🚀", "🪐", "⚛️", "📡", "🧠", "🦾", "👾"]

const handleFileUpload = (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  const reader = new FileReader()
  reader.onload = (e) => {
    selectedAvatar.value = e.target?.result as string
  }
  reader.readAsDataURL(file)
}
</script>

<template>
  <div v-if="show" class="fixed inset-0 z-[100] flex items-center justify-center p-4 backdrop-blur-xl bg-slate-900/40 dark:bg-black/80">
    <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[3rem] p-10 max-xl-md w-full shadow-2xl relative animate-in fade-in zoom-in duration-300">
      <h3 class="text-2xl font-black mb-8 italic uppercase text-center text-slate-900 dark:text-white">选择新的身份标识 / Select Avatar</h3>
      
      <div class="flex flex-col items-center gap-8 mb-10">
        <div class="relative group/preview">
          <div class="w-32 h-32 bg-slate-50 dark:bg-[#1a1c1e] rounded-[2.5rem] border-2 border-dashed border-slate-200 dark:border-white/10 flex items-center justify-center overflow-hidden transition-all group-hover/preview:border-blue-500/50">
             <template v-if="selectedAvatar && selectedAvatar.startsWith('data:')">
                <img :src="selectedAvatar" class="w-full h-full object-cover" />
             </template>
             <template v-else>
                <span class="text-6xl">{{ selectedAvatar || '🧪' }}</span>
             </template>
          </div>
          <button 
            @click="fileInput?.click()"
            class="absolute -bottom-2 -right-2 bg-blue-600 p-2 rounded-xl text-white shadow-lg shadow-blue-500/20 hover:scale-110 transition-transform"
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

        <div class="w-full">
          <p class="text-[10px] font-black uppercase text-slate-500 tracking-[0.2em] mb-4 text-center">快捷原型选择器 / Quick Lab Presets</p>
          <div class="grid grid-cols-4 sm:grid-cols-6 gap-4">
            <button
              v-for="emoji in avatarOptions"
              :key="emoji"
              @click="selectedAvatar = emoji"
              :class="cn(
                'w-16 h-16 text-3xl flex items-center justify-center rounded-[1.5rem] transition-all duration-300 border-2',
                selectedAvatar === emoji 
                  ? 'bg-blue-600 border-blue-400 scale-110 shadow-[0_0_20px_rgba(59,130,246,0.5)]' 
                  : 'bg-slate-50 dark:bg-white/5 border-slate-100 dark:border-transparent hover:border-blue-300 dark:hover:border-white/20 hover:scale-105'
              )"
            >
              {{ emoji }}
            </button>
          </div>
        </div>
        
        <div class="bg-blue-500/5 border border-blue-500/10 rounded-2xl p-4 w-full flex items-center gap-4">
          <div class="p-2 bg-blue-500/10 rounded-xl text-blue-600 dark:text-blue-400">
            <Upload class="w-4 h-4" />
          </div>
          <div class="flex flex-col">
            <span class="text-xs font-bold text-slate-900 dark:text-white">本地图像上传协议 (MAX 2MB)</span>
            <span class="text-[10px] text-slate-500">支持 JPG, PNG, WEBP 等格式</span>
          </div>
        </div>
      </div>

      <div class="flex gap-4">
        <button 
          @click="$emit('close')"
          class="flex-1 py-4 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 rounded-2xl font-bold transition-all text-slate-500 dark:text-slate-400"
        >
          取消
        </button>
        <button 
          @click="$emit('save', selectedAvatar)"
          :disabled="loading"
          class="flex-1 py-4 bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 rounded-2xl font-black text-white shadow-xl shadow-blue-500/20 disabled:opacity-50 flex items-center justify-center gap-2"
        >
          <Loader2 v-if="loading" class="w-5 h-5 animate-spin" />
          同步身份更改
        </button>
      </div>
    </div>
  </div>
</template>
