<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { X, Cpu, Plus, Trash2, Calendar, HardDrive, Loader2, Key } from 'lucide-vue-next'
import { create } from '@github/webauthn-json'
import api from '../../utils/api'
import { useDialog } from '../../utils/dialog'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const { showAlert } = useDialog()
const keys = ref<any[]>([])
const loading = ref(false)
const registering = ref(false)

const fetchKeys = async () => {
  loading.value = true
  try {
    const res = await api.get('/user/webauthn/credentials')
    keys.value = res.data
  } catch (error) {
    console.error('获取密钥列表失败', error)
  } finally {
    loading.value = false
  }
}

const addKey = async () => {
  registering.value = true
  try {
    // 1. 获取注册选项
    const res = await api.get('/user/webauthn/register/begin')
    
    // 2. 调用浏览器 API
    const credential = await create(res.data)
    
    // 3. 发送结果到服务器
    await api.post('/user/webauthn/register/finish', credential)
    
    // 4. 刷新列表
    await fetchKeys()
    await showAlert('硬件密钥注册成功！', '同步完成')
  } catch (error: any) {
    console.error('注册密钥失败', error)
    showAlert('注册失败: ' + (error.response?.data?.error || error.message), '错误')
  } finally {
    registering.value = false
  }
}

const removeKey = async (id: string) => {
  if (!confirm('确定要移除此硬件密钥吗？此操作不可撤销。')) return
  
  try {
    await api.delete(`/user/webauthn/credentials/${id}`)
    await fetchKeys()
  } catch (error) {
    console.error('删除密钥失败', error)
  }
}

onMounted(() => {
  if (props.show) {
    fetchKeys()
  }
})
</script>

<template>
  <Transition name="modal">
    <div v-if="show" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-md" @click="$emit('close')" />
      
      <div class="relative bg-white dark:bg-[#111114] w-full max-w-2xl rounded-[2.5rem] shadow-2xl overflow-hidden border border-slate-200 dark:border-white/10">
        <!-- 头部 -->
        <div class="p-8 border-b border-slate-100 dark:border-white/5 flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div class="bg-blue-500/10 p-3 rounded-2xl">
              <Cpu class="w-6 h-6 text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <h3 class="text-xl font-black uppercase italic italic tracking-tighter text-slate-900 dark:text-white">
                硬件安全密钥 / Hardware Keys
              </h3>
              <p class="text-slate-500 text-xs">使用 FIDO2/WebAuthn 设备（如 YubiKey、指纹、面容 ID）增强安全性</p>
            </div>
          </div>
          <button @click="$emit('close')" class="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl transition-colors">
            <X class="w-6 h-6 text-slate-400" />
          </button>
        </div>

        <!-- 内容 -->
        <div class="p-8 max-h-[60vh] overflow-y-auto">
          <div v-if="loading" class="flex flex-col items-center justify-center py-12">
            <Loader2 class="w-10 h-10 animate-spin text-blue-500 mb-4" />
            <p class="text-slate-500 animate-pulse">正在检索已授权设备...</p>
          </div>

          <div v-else-if="keys.length === 0" class="flex flex-col items-center justify-center py-12 bg-slate-50 dark:bg-white/5 rounded-[2rem] border-2 border-dashed border-slate-200 dark:border-white/10">
            <Key class="w-12 h-12 text-slate-300 dark:text-white/20 mb-4" />
            <p class="text-slate-500 font-medium">尚未配置硬件密钥</p>
            <p class="text-slate-400 text-xs mt-1">添加物理密钥或生物识别设备以实现无密码登录</p>
          </div>

          <div v-else class="space-y-4">
            <div v-for="key in keys" :key="key.id" class="group flex items-center justify-between p-6 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/5 rounded-3xl hover:border-blue-300 dark:hover:border-blue-500/30 transition-all">
              <div class="flex items-center gap-4">
                <div class="bg-slate-200 dark:bg-white/10 p-3 rounded-2xl">
                  <HardDrive class="w-5 h-5 text-slate-600 dark:text-white/60" />
                </div>
                <div>
                  <div class="text-slate-900 dark:text-white font-bold flex items-center gap-2">
                    {{ key.type || 'FIDO2 Device' }}
                    <span class="px-2 py-0.5 bg-blue-500/10 text-blue-600 dark:text-blue-400 text-[10px] rounded-full uppercase tracking-widest font-black">Active</span>
                  </div>
                  <div class="flex items-center gap-3 mt-1">
                    <span class="flex items-center gap-1 text-[10px] text-slate-400 uppercase font-bold tracking-tighter">
                      <Calendar class="w-3 h-3" />
                      ADD: {{ new Date(key.date).toLocaleDateString() }}
                    </span>
                    <span class="text-[10px] text-slate-300 dark:text-white/10 hidden lg:block">ID: {{ key.id.substring(0, 16) }}...</span>
                  </div>
                </div>
              </div>
              <button 
                @click="removeKey(key.id)"
                class="p-2 text-slate-400 hover:text-red-500 hover:bg-red-500/10 rounded-xl transition-all"
                title="删除此密钥"
              >
                <Trash2 class="w-5 h-5" />
              </button>
            </div>
          </div>
        </div>

        <!-- 底部 -->
        <div class="p-8 bg-slate-50 dark:bg-white/5 border-t border-slate-100 dark:border-white/5">
          <button 
            @click="addKey"
            :disabled="registering"
            class="w-full flex items-center justify-center gap-3 py-4 bg-slate-900 dark:bg-blue-600 hover:bg-black dark:hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed text-white font-black uppercase italic tracking-widest transition-all rounded-2xl shadow-xl shadow-blue-500/20"
          >
            <Loader2 v-if="registering" class="w-5 h-5 animate-spin" />
            <Plus v-else class="w-5 h-5" />
            {{ registering ? '正在等待硬件响应...' : '添加新硬件密钥' }}
          </button>
          <p class="text-[10px] text-slate-400 text-center mt-4 uppercase tracking-widest font-bold">
            Chemistry Uno Security Framework V1.0 - WebAuthn Protocol
          </p>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.modal-enter-active, .modal-leave-active { transition: all 0.3s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; transform: scale(0.95); }
</style>
