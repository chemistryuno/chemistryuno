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
              Experimental Wiki <span class="text-xs font-mono bg-blue-500/20 text-blue-400 px-2 py-1 rounded border border-blue-500/30 not-italic">{{ user?.role?.toUpperCase() || 'USER' }}</span>
            </h1>
            <p class="text-slate-500 text-sm font-bold tracking-widest uppercase mt-1">
              {{ (user.role === 'admin' || user.role === 'co-worker') ? '化学反应库管理 / Reaction Database Manager' : '实验室化学反应百科 / Chemical Reaction Encyclopedia' }}
            </p>
          </div>
        </div>

        <div class="flex items-center gap-4">
          <div class="px-6 py-3 bg-[#111114] border border-white/5 rounded-2xl flex items-center gap-4 shadow-xl">
            <div class="flex flex-col items-end">
              <span class="text-[10px] font-black text-slate-500 uppercase">User Role</span>
              <span class="text-xs font-bold text-blue-400 flex items-center gap-1.5">
                <span class="w-2 h-2 bg-blue-500 rounded-full animate-pulse" />
                {{ user?.role?.toUpperCase() || 'USER' }}
              </span>
            </div>
            <div class="w-px h-8 bg-white/5" />
            <router-link 
              to="/ranking"
              class="flex items-center gap-2 text-amber-500 hover:text-amber-400 transition-colors group"
            >
              <Trophy class="w-4 h-4 group-hover:scale-110 transition-transform" />
              <span class="text-xs font-black uppercase tracking-widest">Rank</span>
            </router-link>
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
          <div class="flex items-center justify-between mb-6">
            <h3 class="text-xl font-black text-white flex items-center gap-3">
              <Plus class="w-6 h-6 text-blue-400" />
              {{ editingId ? '编辑反应' : '发现新反应' }}
            </h3>
            <button 
              v-if="user.role === 'admin'"
              @click="triggerFileInput"
              class="p-2.5 bg-blue-500/10 hover:bg-blue-500/20 text-blue-400 border border-blue-500/20 rounded-xl transition-all group"
              title="批量导入JSON"
            >
              <Upload class="w-5 h-5 group-hover:scale-110 transition-transform" />
              <input 
                type="file" 
                ref="fileInput" 
                class="hidden" 
                accept=".json"
                @change="handleFileUpload"
              />
            </button>
          </div>
          
          <form @submit.prevent="handleAddReaction" class="space-y-6">
            <EquationEditor ref="editorRef" v-model="generatedDisplay" />
            
            <button 
              type="submit"
              :disabled="loading"
              class="w-full bg-blue-600 hover:bg-blue-500 px-6 py-4 rounded-2xl font-black text-white uppercase tracking-widest transition-all shadow-[0_10px_20px_rgba(37,99_235_0.2)] hover:scale-[1.02] active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <span v-if="!loading">添加反应</span>
              <span v-else class="flex items-center justify-center gap-2">
                <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                处理中...
              </span>
            </button>
          </form>
        </div>

        <!-- 反应列表 (全宽) -->
        <div class="lg:col-span-2 bg-[#111114] border border-white/10 p-8 rounded-[2rem] shadow-xl">
          <div class="flex flex-col md:flex-row md:items-center justify-between gap-6 mb-8">
            <h3 class="text-xl font-black text-white flex items-center gap-3">
              <Database class="w-6 h-6 text-green-400" />
              反应库索引 <span class="text-slate-600 text-[10px] font-mono not-italic uppercase tracking-widest">/ Global_Wiki</span>
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
                  <th class="px-6 py-4">Status</th>
                  <th class="px-6 py-4">Creator</th>
                  <th class="px-6 py-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-white/5 font-mono">
                <tr v-for="reaction in filteredReactions" :key="reaction.id" class="hover:bg-white/5 transition-colors group">
                  <td class="px-6 py-5 font-bold text-white text-xs">
                    <div class="flex items-center gap-4 border-l-2" 
                      :class="{
                        'border-amber-500/50 pl-4': reaction.status === 'pending_coworker',
                        'border-blue-500/50 pl-4': reaction.status === 'pending_admin',
                        'border-emerald-500/50 pl-4': reaction.status === 'approved',
                        'border-red-500/50 pl-4': reaction.status === 'rejected'
                      }">
                      <div class="flex flex-col gap-1.5 flex-1">
                        <div v-if="editingReactionId === reaction.id" class="flex flex-col gap-2 p-4 bg-slate-100 dark:bg-slate-900/50 rounded-2xl border border-blue-500/20 my-2">
                          <EquationEditor v-model="editForm.display" />
                          <div class="flex gap-2 justify-end mt-2">
                            <button @click="editingReactionId = null" class="px-3 py-1.5 text-xs font-bold text-slate-500 hover:text-slate-700">取消</button>
                            <button @click="saveEdit(reaction.id)" class="px-4 py-1.5 bg-blue-600 hover:bg-blue-500 text-white text-xs font-bold rounded-lg shadow-lg">保存修改</button>
                          </div>
                        </div>
                        <span v-else class="text-white text-sm tracking-tight leading-relaxed">{{ reaction.display }}</span>
                      </div>
                    </div>
                  </td>
                  <td class="px-6 py-5">
                    <div class="flex items-center gap-2">
                      <span v-if="reaction.status === 'pending_coworker'" class="flex items-center gap-1 text-[8px] font-black text-amber-500 uppercase tracking-widest bg-amber-500/10 px-2 py-0.5 rounded border border-amber-500/20">
                        <Clock class="w-2.5 h-2.5" /> PENDING CO-WORKER
                      </span>
                      <span v-else-if="reaction.status === 'pending_admin'" class="flex items-center gap-1 text-[8px] font-black text-blue-500 uppercase tracking-widest bg-blue-500/10 px-2 py-0.5 rounded border border-blue-500/20">
                        <Clock class="w-2.5 h-2.5" /> PENDING ADMIN
                      </span>
                      <span v-else-if="reaction.status === 'approved'" class="flex items-center gap-1 text-[8px] font-black text-emerald-500 uppercase tracking-widest bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/20">
                        <CheckCircle class="w-2.5 h-2.5" /> VERIFIED
                      </span>
                      <span v-else-if="reaction.status === 'rejected'" class="flex items-center gap-1 text-[8px] font-black text-red-500 uppercase tracking-widest bg-red-500/10 px-2 py-0.5 rounded border border-red-500/20">
                        <Trash2 class="w-2.5 h-2.5" /> REJECTED
                      </span>
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
                      <!-- Approve logic -->
                      <button 
                        v-if="canApprove(reaction)"
                        @click="handleApproveReaction(reaction, false)"
                        class="px-3 py-1.5 bg-emerald-600/10 hover:bg-emerald-600 dark:bg-emerald-600/20 text-emerald-600 dark:text-emerald-400 hover:text-white rounded-xl text-[9px] font-black uppercase tracking-widest transition-all shadow-lg shadow-emerald-900/10 whitespace-nowrap"
                      >
                        Approve
                      </button>
                      <button 
                        v-if="canApprove(reaction)"
                        @click="handleApproveReaction(reaction, true)"
                        class="px-3 py-1.5 bg-red-600/10 hover:bg-red-600 text-red-600 hover:text-white rounded-xl text-[9px] font-black uppercase tracking-widest transition-all shadow-lg whitespace-nowrap"
                      >
                        Reject
                      </button>

                      <!-- 直接修改 -->
                      <button 
                        v-if="user.role === 'admin' || user.role === 'co-worker'"
                        @click="handleEdit(reaction)"
                        class="p-2 hover:bg-blue-500/10 dark:hover:bg-blue-500/20 text-slate-400 hover:text-blue-500 rounded-xl transition-all"
                        title="直接修改"
                      >
                        <Edit class="w-4 h-4" />
                      </button>

                      <!-- 反馈纠错 -->
                      <button 
                        v-else
                        @click="handleReport(reaction)"
                        class="p-2 hover:bg-amber-500/10 text-slate-400 hover:text-amber-500 rounded-xl transition-all"
                        title="反馈纠错"
                      >
                        <MessageSquare class="w-4 h-4" />
                      </button>
                      
                      <button 
                        v-if="canDelete(reaction)"
                        @click="handleDeleteReaction(reaction.id)"
                        class="p-2 hover:bg-red-500/10 dark:hover:bg-red-500/20 text-slate-400 hover:text-red-600 dark:text-slate-600 dark:hover:text-red-400 rounded-xl transition-all"
                        :title="reaction.status !== 'approved' ? 'Reject' : 'Delete'"
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
import EquationEditor from '../components/EquationEditor.vue'
import { 
  ArrowLeft, 
  Beaker, 
  Plus, 
  Database, 
  Trash2,
  Search as SearchIcon,
  CheckCircle,
  Clock,
  FlaskConical,
  Upload,
  MessageSquare,
  Edit,
  Trophy
} from 'lucide-vue-next'

const router = useRouter()
const { showAlert, showConfirm } = useDialog()

const user = ref<any>({ uid: 0, role: 'user', username: 'Guest' })
const reactions = ref<any[]>([])
const loading = ref(false)
const searchTerm = ref('')
const fileInput = ref<HTMLInputElement | null>(null)
const editorRef = ref<any>(null)

// 编辑与反馈逻辑
const editingReactionId = ref<number | null>(null)
const editForm = ref({ display: '' })

const handleEdit = (reaction: any) => {
  editingReactionId.value = reaction.id
  editForm.value.display = reaction.display
}

const saveEdit = async (id: number) => {
  if (!editForm.value.display) return
  try {
    await reactionAPI.updateReaction(id, editForm.value.display)
    editingReactionId.value = null
    await loadReactions()
    showAlert('方程式已直接修改并同步至所有终端', '修改成功')
  } catch (e: any) {
    showAlert(e.response?.data?.error || '修改失败', '错误')
  }
}

const handleReport = (reaction: any) => {
  const content = `【方程式纠错】\n\n原方程式：${reaction.display}\n建议修改：`
  router.push({
    path: '/',
    query: { report: content }
  })
}

// 方程式由组件同步
const generatedDisplay = ref('')

// 初始化用户信息并增加容错
onMounted(() => {
  try {
    const savedUser = localStorage.getItem('user')
    if (savedUser) {
      user.value = JSON.parse(savedUser)
    }
  } catch (e) {
    console.error('解析用户信息失败', e)
  }
  loadReactions()
})

// 权限检查
const canApprove = (reaction: any) => {
  if (user.value.role === 'admin') {
    return reaction.status === 'pending_coworker' || reaction.status === 'pending_admin'
  }
  if (user.value.role === 'co-worker') {
    return reaction.status === 'pending_coworker'
  }
  return false
}

const canDelete = (reaction: any) => {
  if (user.value.role === 'admin') return true
  if (reaction.status === 'approved') return false // Once approved, only admin can delete
  return reaction.created_by === user.value.uid
}

const triggerFileInput = () => {
  fileInput.value?.click()
}

const handleFileUpload = async (event: Event) => {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return

  const reader = new FileReader()
  reader.onload = async (e) => {
    try {
      const content = e.target?.result as string
      const data = JSON.parse(content)
      
      if (!Array.isArray(data)) {
        throw new Error('JSON必须是一个数据数组')
      }

      loading.value = true
      const reactionsToImport = data.map((item: any) => {
        if (typeof item === 'string') return { display: item }
        if (item.display) return { display: item.display }
        return null
      }).filter(Boolean) as { display: string }[]

      if (reactionsToImport.length === 0) {
        throw new Error('未找到有效的反应方程式')
      }

      const response = await reactionAPI.batchAddReactions(reactionsToImport)
      await loadReactions()
      await showAlert(response.data.message, '批量导入成功')
    } catch (error: any) {
      console.error('导入失败:', error)
      await showAlert(error.message || '导入过程中发生错误', '导入失败')
    } finally {
      loading.value = false
      if (fileInput.value) fileInput.value.value = ''
    }
  }
  reader.readAsText(file)
}

const filteredReactions = computed(() => {
  return reactions.value.filter(r => 
    r.display.toLowerCase().includes(searchTerm.value.toLowerCase())
  )
})

const loadReactions = async () => {
  loading.value = true
  try {
    let response;
    if (user.value.role === 'admin' || user.value.role === 'co-worker') {
      // 管理员和协作者看到所有待审核和已审核的
      response = await reactionAPI.getReactions()
    } else {
      // 普通用户看到公共Wiki库 + 自己的提交
      const [wikiRes, myRes] = await Promise.all([
        reactionAPI.getAllReactions(),
        reactionAPI.getMyReactions()
      ])
      
      const wikiData = (wikiRes.data || []).map((r: any) => ({
        ...r,
        status: 'approved',
        creator_name: 'SYSTEM (Wiki)'
      }))
      
      response = { data: [...wikiData, ...(myRes.data || [])] }
    }
    reactions.value = response.data || []
  } catch (error) {
    console.error('加载反应数据失败:', error)
  } finally {
    loading.value = false
  }
}

const handleApproveReaction = async (reaction: any, reject: boolean) => {
  try {
    const action = reject ? '拒绝' : '通过'
    await reactionAPI.approveReaction(reaction.group_id, reaction.display, reject)
    await loadReactions()
    await showAlert(`该化学方程式已${action}`, '审核操作已完成')
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
    editorRef.value?.reset()
    // 重新加载数据
    await loadReactions()
    let msg = '反应添加成功！'
    if (user.value.role === 'user') {
      msg = '建议已提交，数据已直接写入数据库并等待协作者/管理员审核（pending）。'
    }
    await showAlert(msg, '成功')
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
</script>
