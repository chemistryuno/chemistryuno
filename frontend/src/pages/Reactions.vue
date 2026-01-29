<template>
  <div class="min-h-screen bg-slate-50 dark:bg-[#0a0a0c] text-slate-900 dark:text-slate-200 p-4 lg:p-10 font-sans selection:bg-blue-500/30">
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
            <!-- 规范化编辑器 -->
            <div class="space-y-4">
              <label class="block text-xs font-black text-slate-500 uppercase tracking-widest mb-2">反应方程式编辑器</label>
              
              <!-- 反应物区域 -->
              <div class="space-y-3">
                <div v-for="(item, index) in editor.reactants" :key="'r-'+index" class="flex items-center gap-3 p-3 bg-blue-500/5 rounded-2xl border border-blue-500/10 group/item">
                  <input 
                    v-model="item.coefficient"
                    type="text" 
                    placeholder="1"
                    class="w-14 px-2 py-2 bg-black/40 border border-white/5 rounded-xl text-center text-blue-400 font-black placeholder:text-slate-800 focus:border-blue-500/30 outline-none transition-all"
                  />
                  <input 
                    v-model="item.formula"
                    type="text" 
                    placeholder="Substance (e.g. H2SO4)"
                    class="flex-1 px-3 py-2 bg-black/40 border border-white/5 rounded-xl text-white font-mono placeholder:text-slate-800 focus:border-blue-500/30 outline-none transition-all"
                  />
                  <button @click.prevent="removeSubstance('reactants', index)" class="p-2 text-slate-700 hover:text-red-400 opacity-0 group-hover/item:opacity-100 transition-all">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
                <button @click.prevent="addSubstance('reactants')" class="w-full py-3 bg-blue-500/5 hover:bg-blue-500/10 border border-dashed border-blue-500/20 rounded-2xl text-[10px] font-black text-blue-400 flex items-center justify-center gap-2 uppercase tracking-[0.2em] transition-all">
                  <Plus class="w-3.5 h-3.5" /> Append Reactant
                </button>
              </div>

              <!-- 箭头/等号选择 -->
              <div class="flex items-center justify-center gap-4 py-2">
                <div class="h-px flex-1 bg-gradient-to-r from-transparent to-white/5" />
                <div class="relative group">
                  <select v-model="editor.arrow" class="appearance-none bg-[#1a1a1e] border border-blue-500/30 rounded-full px-8 py-2 text-blue-400 font-black text-xs focus:border-blue-500 outline-none transition-all cursor-pointer shadow-[0_0_20px_rgba(59,130,246,0.1)]">
                    <option value="=">=</option>
                    <option value="→">→</option>
                    <option value="⇌">⇌</option>
                  </select>
                  <FlaskConical class="w-3 h-3 text-blue-500/40 absolute left-3 top-1/2 -translate-y-1/2" />
                </div>
                <div class="h-px flex-1 bg-gradient-to-l from-transparent to-white/5" />
              </div>

              <!-- 生成物区域 -->
              <div class="space-y-3">
                <div v-for="(item, index) in editor.products" :key="'p-'+index" class="flex items-center gap-3 p-3 bg-green-500/5 rounded-2xl border border-green-500/10 group/item">
                  <input 
                    v-model="item.coefficient"
                    type="text" 
                    placeholder="1"
                    class="w-14 px-2 py-2 bg-black/40 border border-white/5 rounded-xl text-center text-green-400 font-black placeholder:text-slate-800 focus:border-green-500/30 outline-none transition-all"
                  />
                  <input 
                    v-model="item.formula"
                    type="text" 
                    placeholder="Substance (e.g. Na2SO4)"
                    class="flex-1 px-3 py-2 bg-black/40 border border-white/5 rounded-xl text-white font-mono placeholder:text-slate-800 focus:border-green-500/30 outline-none transition-all"
                  />
                  <button @click.prevent="removeSubstance('products', index)" class="p-2 text-slate-700 hover:text-red-400 opacity-0 group-hover/item:opacity-100 transition-all">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
                <button @click.prevent="addSubstance('products')" class="w-full py-3 bg-green-500/5 hover:bg-green-500/10 border border-dashed border-green-500/20 rounded-2xl text-[10px] font-black text-green-400 flex items-center justify-center gap-2 uppercase tracking-[0.2em] transition-all">
                  <Plus class="w-3.5 h-3.5" /> Append Product
                </button>
              </div>

            <div class="space-y-4">
              <div v-if="generatedDisplay" class="p-5 bg-blue-500/10 border border-blue-500/20 rounded-3xl shadow-inner group">
                <span class="block text-[9px] font-black text-blue-400 uppercase tracking-[0.3em] mb-2 opacity-60">Synthesis Result Preview</span>
                <span class="text-sm font-mono font-bold text-white break-all leading-relaxed">{{ generatedDisplay }}</span>
              </div>
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
            </div>
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
                  <th class="px-6 py-4">Creator</th>
                  <th class="px-6 py-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-white/5 font-mono">
                <tr v-for="reaction in filteredReactions" :key="reaction.id" class="hover:bg-white/5 transition-colors group">
                  <td class="px-6 py-5 font-bold text-white text-xs">
                    <div class="flex items-center gap-4 border-l-2" :class="reaction.status === 'pending' ? 'border-amber-500/50 pl-4' : 'border-emerald-500/50 pl-4'">
                      <div class="flex flex-col gap-1.5 flex-1">
                        <div v-if="user.role === 'admin' && reaction.status === 'pending'" class="flex items-center gap-2">
                          <input 
                            v-model="reaction.display"
                            class="bg-white/5 border border-white/10 rounded-lg px-3 py-1.5 text-blue-400 text-sm tracking-tight w-full focus:outline-none focus:border-blue-500/50 transition-all font-bold"
                            placeholder="修改方程式..."
                          />
                        </div>
                        <span v-else class="text-white text-sm tracking-tight leading-relaxed">{{ reaction.display }}</span>
                        <div class="flex items-center gap-2">
                          <span v-if="reaction.status === 'pending'" class="flex items-center gap-1 text-[8px] font-black text-amber-500 uppercase tracking-widest bg-amber-500/10 px-2 py-0.5 rounded border border-amber-500/20">
                            <Clock class="w-2.5 h-2.5" /> Pending Review
                          </span>
                          <span v-else class="flex items-center gap-1 text-[8px] font-black text-emerald-500 uppercase tracking-widest bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/20">
                            <CheckCircle class="w-2.5 h-2.5" /> Verified
                          </span>
                        </div>
                      </div>
                    </div>
                  </td>
                  <td class="px-6 py-5 text-[10px] text-slate-600 dark:text-slate-400">
                    <div class="flex flex-col">
                      <span class="font-bold text-slate-700 dark:text-slate-400 opacity-80 uppercase tracking-tighter">{{ reaction.creator_name || 'SYSTEM' }}</span>
                      <span class="text-[8px] opacity-60">{{ new Date(reaction.created_at).toLocaleDateString() }}</span>
                    </div>
                  </td>
                  <td class="px-6 py-5 text-right">
                    <div class="flex items-center justify-end gap-2">
                      <button 
                        v-if="user.role === 'admin' && reaction.status === 'pending'"
                        @click="handleApproveReaction(reaction)"
                        class="px-3 py-1.5 bg-emerald-600/10 hover:bg-emerald-600 dark:bg-emerald-600/20 text-emerald-600 dark:text-emerald-400 hover:text-white rounded-xl text-[9px] font-black uppercase tracking-widest transition-all shadow-lg shadow-emerald-900/10"
                      >
                        Approve
                      </button>
                      <button 
                        @click="handleDeleteReaction(reaction.id)"
                        class="p-2 hover:bg-red-500/10 dark:hover:bg-red-500/20 text-slate-400 hover:text-red-600 dark:text-slate-600 dark:hover:text-red-400 rounded-xl transition-all"
                        :title="reaction.status === 'pending' ? 'Reject' : 'Delete'"
                      >
                        <Trash2 class="w-4 h-4" />
                      </button>
                    </div>
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
  Search as SearchIcon,
  CheckCircle,
  Clock,
  FlaskConical
} from 'lucide-vue-next'

const router = useRouter()
const { showAlert, showConfirm } = useDialog()

const user = ref<any>(JSON.parse(localStorage.getItem('user') || '{}'))
const reactions = ref<any[]>([])
const loading = ref(false)
const searchTerm = ref('')

// 方程式编辑器状态
const editor = ref({
  reactants: [{ coefficient: '', formula: '' }],
  products: [{ coefficient: '', formula: '' }],
  arrow: '='
})

// 自动生成预览方程式
const generatedDisplay = computed(() => {
  const formatList = (list: { coefficient: string, formula: string }[]) => {
    return list
      .filter(item => item.formula.trim())
      .map(item => {
        const coef = item.coefficient.trim()
        const formula = item.formula.trim()
        // 隐藏所有 1
        return (coef && coef !== '1') ? `${coef}${formula}` : formula
      })
      .join(' + ')
  }

  const left = formatList(editor.value.reactants)
  const right = formatList(editor.value.products)
  
  if (!left || !right) return ''
  return `${left} ${editor.value.arrow} ${right}`
})

const addSubstance = (type: 'reactants' | 'products') => {
  editor.value[type].push({ coefficient: '', formula: '' })
}

const removeSubstance = (type: 'reactants' | 'products', index: number) => {
  if (editor.value[type].length > 1) {
    editor.value[type].splice(index, 1)
  } else {
    editor.value[type][index] = { coefficient: '', formula: '' }
  }
}

const filteredReactions = computed(() => {
  return reactions.value.filter(r => 
    r.display.toLowerCase().includes(searchTerm.value.toLowerCase())
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

const handleApproveReaction = async (reaction: any) => {
  try {
    await reactionAPI.approveReaction(reaction.group_id, reaction.display)
    await loadReactions()
    await showAlert('该化学方程式已通过专家审核并写入核心库', '审核通过')
  } catch (error: any) {
    await showAlert(error.response?.data?.error || '审核失败', '错误')
  }
}

const handleAddReaction = async () => {
  const display = generatedDisplay.value
  if (!display) {
    await showAlert('请填写完整的化学方程式', '预览错误')
    return
  }
  
  loading.value = true
  try {
    await reactionAPI.addReaction(display)
    
    // 重置编辑器
    editor.value = {
      reactants: [{ coefficient: '', formula: '' }],
      products: [{ coefficient: '', formula: '' }],
      arrow: '='
    }
    
    // 重新加载数据
    await loadReactions()
    await showAlert('反应添加成功！(系统已自动映射反应物)', '成功')
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

onMounted(() => {
  loadReactions()
})
</script>