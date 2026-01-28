<script setup lang="ts">
import { useDialog } from '../utils/dialog'
import { AlertCircle, HelpCircle, MessageSquare } from 'lucide-vue-next'

const { state, handleConfirm, handleCancel } = useDialog()
</script>

<template>
  <Transition name="fade">
    <div v-if="state.show" class="fixed inset-0 bg-black/70 backdrop-blur-sm z-[9999] flex items-center justify-center p-4">
      <div class="bg-[#111114] border border-white/10 rounded-[2rem] p-8 max-w-md w-full shadow-2xl overflow-hidden relative">
        <!-- 装饰背景 -->
        <div class="absolute -top-24 -right-24 w-48 h-48 bg-blue-500/5 rounded-full blur-3xl pointer-events-none" />
        <div class="absolute -bottom-24 -left-24 w-48 h-48 bg-purple-500/5 rounded-full blur-3xl pointer-events-none" />
        
        <div class="flex items-center gap-4 mb-6 relative z-10">
          <div v-if="state.type === 'alert'" class="w-12 h-12 rounded-2xl bg-blue-500/10 flex items-center justify-center border border-blue-500/20">
            <AlertCircle class="w-6 h-6 text-blue-400" />
          </div>
          <div v-else-if="state.type === 'confirm'" class="w-12 h-12 rounded-2xl bg-orange-500/10 flex items-center justify-center border border-orange-500/20">
            <HelpCircle class="w-6 h-6 text-orange-400" />
          </div>
          <div v-else-if="state.type === 'prompt'" class="w-12 h-12 rounded-2xl bg-purple-500/10 flex items-center justify-center border border-purple-500/20">
            <MessageSquare class="w-6 h-6 text-purple-400" />
          </div>
          <h3 class="text-xl font-black text-white italic tracking-tight">{{ state.title }}</h3>
        </div>
        
        <div class="mb-8 relative z-10">
          <p class="text-slate-400 leading-relaxed whitespace-pre-line font-medium">{{ state.message }}</p>
          
          <!-- 输入框（仅prompt类型显示） -->
          <div v-if="state.type === 'prompt'" class="mt-6 relative">
            <input 
              v-model="state.inputValue"
              type="text" 
              :placeholder="state.inputPlaceholder"
              class="w-full bg-white/5 border border-white/10 rounded-2xl px-5 py-4 text-white focus:outline-none focus:border-purple-500/50 focus:ring-4 focus:ring-purple-500/5 transition-all placeholder:text-slate-600 font-medium"
              @keydown.enter="handleConfirm"
              autofocus
            />
          </div>
        </div>
        
        <div class="flex gap-4 relative z-10">
          <button 
            v-if="state.type === 'confirm' || state.type === 'prompt'"
            @click="handleCancel"
            class="flex-1 px-6 py-4 bg-white/5 hover:bg-white/10 text-slate-400 hover:text-white rounded-2xl font-black text-xs uppercase tracking-widest transition-all border border-white/5"
          >
            {{ state.cancelText }}
          </button>
          <button 
            @click="handleConfirm"
            :class="[
              'flex-1 px-6 py-4 rounded-2xl font-black text-xs uppercase tracking-widest transition-all text-white shadow-xl active:scale-95',
              state.type === 'alert' ? 'bg-blue-600 hover:bg-blue-500 shadow-blue-900/20' :
              state.type === 'confirm' ? 'bg-orange-600 hover:bg-orange-500 shadow-orange-900/20' :
              'bg-purple-600 hover:bg-purple-500 shadow-purple-900/20'
            ]"
          >
            {{ state.confirmText }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.fade-enter-active .bg-\[\#111114\] {
  animation: zoomIn 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes zoomIn {
  from {
    opacity: 0;
    transform: scale(0.9);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}
</style>
