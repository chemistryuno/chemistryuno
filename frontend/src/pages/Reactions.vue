<template>
  <div class="min-h-screen bg-[#0a0a0c] text-slate-200 p-4 lg:p-10 font-sans selection:bg-blue-500/30">
    <div class="fixed inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] bg-blue-500/5 rounded-full blur-[120px]" />
      <div class="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-green-500/5 rounded-full blur-[120px]" />
      <div class="absolute inset-0 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-20 brightness-50 contrast-150" />
    </div>

    <div class="max-w-7xl mx-auto relative z-10">
      <header class="flex flex-col lg:flex-row items-center justify-between gap-8 mb-12">
        <div class="flex items-center gap-6">
          <div class="relative group">
            <div class="absolute inset-x-0 inset-y-0 bg-blue-500 blur-2xl opacity-20 group-hover:opacity-40 transition-opacity" />
            <div class="w-16 h-16 rounded-2xl bg-[#111114] border border-blue-500/40 flex items-center justify-center relative z-10 shadow-2xl">
              <Beaker class="w-8 h-8 text-blue-400 group-hover:scale-110 transition-transform" />
            </div>
          </div>
          <div>
            <h1 class="text-3xl font-black text-white italic tracking-tighter uppercase flex items-center gap-3">
              Chemical Database <span class="text-xs font-mono bg-blue-500/20 text-blue-400 px-2 py-1 rounded border border-blue-500/30 not-italic">Co-Worker</span>
            </h1>
            <p class="text-slate-500 text-sm font-bold tracking-widest uppercase mt-1">化学反应库管理 / Reaction Database Manager</p>
          </div>
        </div>

        <div class="flex items-center gap-4">
          <div class="px-6 py-3 bg-[#111114] border border-white/5 rounded-2xl flex items-center gap-4 shadow-xl">
            <div class="flex flex-col items-end">
              <span class="text-[10px] font-black text-slate-500 uppercase">User Role</span>
              <span class="text-xs font-bold text-blue-400 flex items-center gap-1.5">
                <span class="w-2 h-2 bg-blue-500 rounded-full animate-pulse" />
                {{ user?.role?.toUpperCase() || 'CO-WORKER' }}
              </span>
            </div>
            <div class="w-px h-8 bg-white/5" />
            <button 
              @click="router.push('/')"
              class="flex items-center gap-2 text-slate-400 hover:text-white transition-colors group"
            >
              <ArrowLeft class="w-4 h-4 group-hover:-translate-x-1 transition-transform" />
              <span class="text-xs font-black uppercase tracking-widest">Exit</span>
            </button>
          </div>
        </div>
      </header>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <!-- 添加反应面板 -->
        <div class="bg-[#111114] border border-white/10 p-8 rounded-[2rem] shadow-xl">
          <h3 class="text-xl font-black text-white mb-6 flex items-center gap-3">
            <Plus class="w-6 h-6 text-blue-400" />
            添加新反应
          </h3>
          
          <form @submit.prevent="handleAddReaction" class="space-y-6">
            <div>
              <label class="block text-xs font-black text-slate-500 uppercase tracking-widest mb-2">反应物</label>
              <input 
                v-model="newReaction.reactant"
                type="text" 
                placeholder="如: H2SO4 + NaOH"
                class="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-2xl text-white placeholder-slate-500 focus:outline-none focus:border-blue-500/50 transition-colors"
                required
              />
            </div>
            
            <div>
              <label class="block text-xs font-black text-slate-500 uppercase tracking-widest mb-2">生成物</label>
              <input 
                v-model="newReaction.product"
                type="text" 
                placeholder="如: Na2SO4 + H2O"
                class="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-2xl text-white placeholder-slate-500 focus:outline-none focus:border-blue-500/50 transition-colors"
                required
              />
            </div>
            
            <div>
              <label class="block text-xs font-black text-slate-500 uppercase tracking-widest mb-2">反应类型</label>
              <select 
                v-model="newReaction.type"
                class="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-2xl text-white focus:outline-none focus:border-blue-500/50 transition-colors"
                required
              >
                <option value="">选择反应类型</option>
                <option value="acid-base">酸碱中和</option>
                <option value="redox">氧化还原</option>
                <option value="displacement">置换反应</option>
                <option value="precipitation">沉淀反应</option>
                <option value="combustion">燃烧反应</option>
                <option value="synthesis">化合反应</option>
                <option value="decomposition">分解反应</option>
              </select>
            </div>
            
            <button 
              type="submit"
              :disabled="loading"
              class="w-full bg-blue-600 hover:bg-blue-500 px-6 py-4 rounded-2xl font-black text-white uppercase tracking-widest transition-all shadow-[0_10px_20px_rgba(37,99,235,0.2)] hover:scale-[1.02] active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <span v-if="!loading">添加反应</span>
              <span v-else class="flex items-center justify-center gap-2">
                <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                处理中...
              </span>
            </button>
          </form>
        </div>

        <!-- 反应列表 -->
        <div class="lg:col-span-2 bg-[#111114] border border-white/10 p-8 rounded-[2rem] shadow-xl">
          <div class="flex flex-col md:flex-row md:items-center justify-between gap-6 mb-8">
            <h3 class="text-xl font-black text-white flex items-center gap-3">
              <Database class="w-6 h-6 text-green-400" />
              反应数据库
            </h3>
            <div class="flex items-center gap-4">
              <div class="relative group">
                <SearchIcon class="w-4 h-4 absolute left-4 top-1/2 -translate-y-1/2 text-slate-600 group-focus-within:text-blue-400 transition-colors" />
                <input 
                  v-model="searchTerm"
                  type="text" 
                  placeholder="搜索反应物/生成物..."
                  class="bg-white/5 border border-white/10 rounded-xl pl-12 pr-6 py-2.5 text-xs focus:outline-none focus:border-blue-500/50 w-full md:w-64 transition-all"
                />
              </div>
              <div class="text-[10px] font-black text-white bg-blue-600/20 px-3 py-2 rounded-xl border border-blue-600/30 whitespace-nowrap">
                MATCHED: {{ filteredReactions.length }}
              </div>
            </div>
          </div>
          
          <div class="overflow-x-auto custom-scrollbar">
            <table class="w-full text-left">
              <thead>
                <tr class="text-slate-600 text-[10px] font-black uppercase tracking-[0.2em] border-b border-white/5">
                  <th class="px-6 py-4">Reaction Formula</th>
                  <th class="px-6 py-4">Type</th>
                  <th class="px-6 py-4">Creator</th>
                  <th class="px-6 py-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-white/5 font-mono">
                <tr v-for="reaction in filteredReactions" :key="reaction.id" class="hover:bg-white/5 transition-colors group">
                  <td class="px-4 py-2 font-bold text-white text-xs">
                    {{ reaction.reactant }} <span class="text-blue-500 mx-1">→</span> {{ reaction.product }}
                  </td>
                  <td class="px-4 py-2">
                    <span class="text-[8px] px-1.5 py-0.5 bg-blue-500/10 text-blue-400 rounded border border-blue-500/20 font-black tracking-widest uppercase">
                      {{ getReactionTypeLabel(reaction.type) }}
                    </span>
                  </td>
                  <td class="px-4 py-2 text-[10px] text-slate-500 font-mono">
                    {{ reaction.creator_name || 'System' }}
                  </td>
                  <td class="px-4 py-2 text-right">
                    <button 
                      @click="handleDeleteReaction(reaction.id)"
                      class="p-1.5 hover:bg-red-500/20 text-slate-600 hover:text-red-400 rounded-lg transition-all"
                      title="删除反应"
                    >
                      <Trash2 class="w-3.5 h-3.5" />
                    </button>
                  </td>
                </tr>
                <tr v-if="filteredReactions.length === 0 && !loading">
                  <td colspan="4" class="py-20 text-center">
                    <Database class="w-12 h-12 text-slate-700 mx-auto mb-4" />
                    <p class="text-slate-500 font-medium italic">未检索到相关化学反应数据</p>
                  </td>
                </tr>
              </tbody>
            </table>
            
            <div v-if="loading" class="text-center py-12">
              <div class="w-8 h-8 border-2 border-blue-500/30 border-t-blue-500 rounded-full animate-spin mx-auto mb-4"></div>
              <p class="text-slate-500 font-medium">加载中...</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { reactionAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import { 
  ArrowLeft, 
  Beaker, 
  Plus, 
  Database, 
  Trash2,
  Search as SearchIcon
} from 'lucide-vue-next'

const router = useRouter()
const { showAlert, showConfirm } = useDialog()

const user = ref<any>(JSON.parse(localStorage.getItem('user') || '{}'))
const reactions = ref<any[]>([])
const loading = ref(false)
const searchTerm = ref('')
const newReaction = ref({
  reactant: '',
  product: '',
  type: ''
})

const filteredReactions = computed(() => {
  return reactions.value.filter(r => 
    r.reactant.includes(searchTerm.value) ||
    r.product.includes(searchTerm.value) ||
    getReactionTypeLabel(r.type).includes(searchTerm.value)
  )
})

// 检查权限
if (!user.value || (user.value.role !== 'admin' && user.value.role !== 'co-worker')) {
  router.push('/')
}

const loadReactions = async () => {
  loading.value = true
  try {
    const response = await reactionAPI.getReactions()
    reactions.value = response.data || []
  } catch (error) {
    console.error('加载反应数据失败:', error)
    await showAlert('加载反应数据失败', '错误')
  } finally {
    loading.value = false
  }
}

const handleAddReaction = async () => {
  loading.value = true
  try {
    await reactionAPI.addReaction(
      newReaction.value.reactant,
      newReaction.value.product,
      newReaction.value.type
    )
    
    // 重置表单
    newReaction.value = {
      reactant: '',
      product: '',
      type: ''
    }
    
    // 重新加载数据
    await loadReactions()
    await showAlert('反应添加成功！', '成功')
  } catch (error: any) {
    console.error('添加反应失败:', error)
    await showAlert(error.response?.data?.error || '添加反应失败', '错误')
  } finally {
    loading.value = false
  }
}

const handleDeleteReaction = async (id: number) => {
  const confirmed = await showConfirm('确定要删除这个反应吗？', '确认删除')
  if (!confirmed) {
    return
  }
  
  try {
    await reactionAPI.deleteReaction(id.toString())
    await loadReactions()
    await showAlert('反应删除成功！', '成功')
  } catch (error: any) {
    console.error('删除反应失败:', error)
    await showAlert(error.response?.data?.error || '删除反应失败', '错误')
  }
}

const getReactionTypeLabel = (type: string) => {
  const typeMap: Record<string, string> = {
    'acid-base': '酸碱中和',
    'redox': '氧化还原',
    'displacement': '置换反应',
    'precipitation': '沉淀反应',
    'combustion': '燃烧反应',
    'synthesis': '化合反应',
    'decomposition': '分解反应'
  }
  return typeMap[type] || type
}

onMounted(() => {
  loadReactions()
})
</script>