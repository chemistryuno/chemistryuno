<template>
  <div class="min-h-screen bg-slate-50 dark:bg-[#0a0a0c] text-slate-900 dark:text-slate-200 p-4 lg:p-6 font-sans selection:bg-emerald-500/30">
    <div class="fixed inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] bg-emerald-500/5 rounded-full blur-[120px]" />
      <div class="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-blue-500/5 rounded-full blur-[120px]" />
      <div class="absolute inset-0 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-20 brightness-50 contrast-150" />
    </div>

    <div class="max-w-7xl mx-auto relative z-10">
      <header class="flex flex-col lg:flex-row items-center justify-between gap-6 mb-8">
        <div class="flex items-center gap-6">
          <div class="relative group">
            <div class="absolute inset-x-0 inset-y-0 bg-emerald-500 blur-2xl opacity-20 group-hover:opacity-40 transition-opacity" />
            <div class="w-12 h-12 rounded-xl bg-white dark:bg-[#111114] border border-slate-200 dark:border-emerald-500/40 flex items-center justify-center relative z-10 shadow-2xl">
              <FlaskConical class="w-6 h-6 text-emerald-600 dark:text-emerald-400 group-hover:scale-110 transition-transform" />
            </div>
          </div>
          <div>
            <h1 class="text-xl font-black text-slate-900 dark:text-white italic tracking-tighter uppercase flex items-center gap-3">
              Substance Wiki <span class="text-[8px] font-mono bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 px-1.5 py-0.5 rounded border border-emerald-500/30 not-italic uppercase">{{ user?.role || 'USER' }}</span>
            </h1>
            <p class="text-slate-400 dark:text-slate-500 text-[10px] font-bold tracking-widest uppercase mt-0.5">
              物质百科全书管理 / Substance encyclopedia
            </p>
          </div>
        </div>

        <div class="flex items-center gap-4">
          <div class="px-4 py-2 bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/5 rounded-xl flex items-center gap-4 shadow-xl">
            <div class="flex flex-col items-end">
              <span class="text-[8px] font-black text-slate-400 dark:text-slate-500 uppercase">Registry Status</span>
              <span class="text-[10px] font-bold text-emerald-600 dark:text-emerald-400 flex items-center gap-1.5">
                <span class="w-1.5 h-1.5 bg-emerald-500 rounded-full animate-pulse" />
                ONLINE
              </span>
            </div>
            <div class="w-px h-6 bg-slate-200 dark:bg-white/5" />
            <router-link 
              to="/data"
              class="flex items-center gap-2 text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white transition-colors group"
            >
              <ArrowLeft class="w-4 h-4 group-hover:-translate-x-1 transition-transform" />
              <span class="text-[10px] font-black uppercase tracking-widest">Back</span>
            </router-link>
          </div>
        </div>
      </header>

      <div class="grid grid-cols-1 lg:grid-cols-4 gap-6">
        <!-- 添加/编辑面板 -->
        <div class="lg:col-span-1 bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 p-5 rounded-2xl shadow-xl h-fit sticky top-6">
          <div class="flex items-center gap-3 mb-6">
            <div class="w-8 h-8 rounded-lg bg-emerald-500/10 flex items-center justify-center border border-emerald-500/20">
              <Plus class="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
            </div>
            <h3 class="text-base font-black text-slate-800 dark:text-white italic uppercase tracking-tight">
              {{ editingId ? 'Edit Entry' : 'Discover Substance' }}
            </h3>
          </div>
          
          <form @submit.prevent="saveSub" class="space-y-4">
            <div class="space-y-1.5">
              <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest pl-1">Chemical Formula</label>
              <input 
                v-model="form.formula" 
                type="text" 
                class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2.5 text-xs font-bold text-slate-900 dark:text-white focus:outline-none focus:border-emerald-500/50 transition-all placeholder:text-slate-400 dark:placeholder:text-slate-700 italic tracking-tighter" 
                placeholder="E.G. H2O"
              />
            </div>
            <div class="space-y-1.5">
              <label class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest pl-1">Chinese Name</label>
              <input 
                v-model="form.name" 
                type="text" 
                class="w-full bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2.5 text-xs font-bold text-slate-900 dark:text-white focus:outline-none focus:border-emerald-500/50 transition-all placeholder:text-slate-400 dark:placeholder:text-slate-700" 
                placeholder="物质名称"
              />
            </div>
            
            <div class="pt-2 flex flex-col gap-2">
              <button 
                type="submit"
                :disabled="loading"
                class="w-full bg-emerald-600 hover:bg-emerald-500 px-6 py-3 rounded-xl font-black text-white uppercase tracking-widest transition-all shadow-[0_10px_20px_rgba(16,185,129,0.1)] hover:scale-[1.02] active:scale-95 disabled:opacity-50 text-xs"
              >
                {{ editingId ? 'Update Substance' : (user.role === 'admin' || user.role === 'co-worker' ? 'Add Substance' : 'Propose Substance') }}
              </button>
              <button 
                v-if="editingId"
                type="button"
                @click="closeModal"
                class="w-full bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 px-6 py-2 rounded-xl font-black text-slate-500 dark:text-slate-400 uppercase tracking-widest transition-all text-xs"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>

        <!-- 列表面板 -->
        <div class="lg:col-span-3 bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 p-5 rounded-2xl shadow-xl">
          <div class="flex flex-col md:flex-row md:items-center justify-between gap-5 mb-6">
            <h3 class="text-base font-black text-slate-800 dark:text-white flex items-center gap-3 italic">
              <Database class="w-5 h-5 text-blue-600 dark:text-blue-400" />
              DATABASE_ENTRIES <span class="text-slate-400 dark:text-slate-600 text-[10px] font-mono not-italic uppercase tracking-widest">/ Substances_Registry</span>
            </h3>
            <div class="flex items-center gap-4">
              <div class="relative group">
                <SearchIcon class="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-emerald-500 transition-colors" />
                <input 
                  v-model="searchTerm"
                  type="text" 
                  placeholder="Search formula/name..."
                  class="bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl pl-9 pr-4 py-2 text-xs text-slate-900 dark:text-white focus:outline-none focus:border-emerald-500/50 w-full md:w-64 transition-all placeholder:text-slate-400"
                />
              </div>
            </div>
          </div>
          
          <div class="overflow-x-auto custom-scrollbar">
            <table class="w-full text-left">
              <thead>
                <tr class="text-slate-400 dark:text-slate-600 text-[10px] font-black uppercase tracking-[0.2em] border-b border-slate-100 dark:border-white/5">
                  <th class="px-4 py-2.5">Formula / Name</th>
                  <th class="px-4 py-2.5">Status</th>
                  <th class="px-4 py-2.5">Author</th>
                  <th class="px-4 py-2.5 text-right">Actions</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 dark:divide-white/5 font-mono">
                <tr v-for="sub in filteredSubstances" :key="sub.id" class="hover:bg-slate-50 dark:hover:bg-white/5 transition-colors group">
                  <td class="px-4 py-3">
                    <div class="flex flex-col">
                      <span class="text-base font-black italic text-slate-900 dark:text-white group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors tracking-tighter">{{ sub.formula }}</span>
                      <span class="text-[10px] font-bold text-slate-400 group-hover:text-slate-600 dark:group-hover:text-slate-300 transition-colors">{{ sub.name }}</span>
                    </div>
                  </td>
                  <td class="px-4 py-3">
                    <span :class="[
                      'px-1.5 py-0.5 rounded text-[8px] font-black uppercase letter-spacing-widest border',
                      sub.status === 'approved' ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' :
                      sub.status === 'pending_admin' ? 'bg-blue-500/10 text-blue-400 border-blue-500/20' :
                      sub.status === 'pending_coworker' ? 'bg-amber-500/10 text-amber-400 border-amber-500/20' :
                      'bg-red-500/10 text-red-400 border-red-500/20'
                    ]">
                      {{ sub.status.replace('_', ' ') }}
                    </span>
                  </td>
                  <td class="px-4 py-3">
                    <div class="flex items-center gap-2 text-slate-400">
                      <UserIcon class="w-3 h-3" />
                      <span class="text-[10px] font-bold">{{ sub.creator_name }}</span>
                    </div>
                  </td>
                  <td class="px-4 py-3 text-right">
                    <div class="flex justify-end gap-1">
                      <!-- 审批按钮 -->
                      <template v-if="canApprove(sub.status)">
                        <button @click="approveSub(sub)" class="p-1.5 hover:bg-emerald-500/10 text-emerald-500/50 hover:text-emerald-400 rounded-lg transition-all" title="Approve">
                          <Check class="w-3.5 h-3.5" />
                        </button>
                        <button @click="rejectSub(sub)" class="p-1.5 hover:bg-red-500/10 text-red-500/50 hover:text-red-500 rounded-lg transition-all" title="Reject">
                          <X class="w-3.5 h-3.5" />
                        </button>
                      </template>
                      
                      <button @click="editSub(sub)" class="p-1.5 hover:bg-emerald-500/10 text-slate-500 hover:text-emerald-400 rounded-lg transition-all">
                        <Edit class="w-3.5 h-3.5" />
                      </button>
                      <button v-if="user?.role === 'admin'" @click="deleteSub(sub.id)" class="p-1.5 hover:bg-red-500/10 text-slate-500 hover:text-red-500 rounded-lg transition-all">
                        <Trash2 class="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </td>
                </tr>
                <tr v-if="filteredSubstances.length === 0 && !loading">
                  <td colspan="4" class="py-10 text-center">
                    <FlaskConical class="w-10 h-10 text-slate-200 dark:text-slate-700 mx-auto mb-4" />
                    <p class="text-slate-400 dark:text-slate-500 font-medium italic text-xs">未检索到相关物质数据</p>
                  </td>
                </tr>
              </tbody>
            </table>
            
            <div v-if="loading" class="text-center py-6">
              <div class="w-6 h-6 border-2 border-emerald-500/30 border-t-emerald-500 rounded-full animate-spin mx-auto mb-4"></div>
              <p class="text-slate-400 dark:text-slate-500 font-medium text-xs">加载中...</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { substanceAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import { 
  FlaskConical, 
  Database, 
  Plus, 
  ArrowLeft, 
  Search as SearchIcon, 
  Trash2, 
  Edit,
  Check,
  X,
  User as UserIcon
} from 'lucide-vue-next'

interface Substance {
  id: number
  formula: string
  name: string
  elements: string
  status: string
  creator_name: string
  created_at: string
}

const { showAlert, showConfirm } = useDialog()
const user = ref<any>({})
try {
  const userData = JSON.parse(localStorage.getItem('user') || '{}')
  // 兼容旧版本的 id 字段
  if (userData.id && !userData.uid) {
    userData.uid = userData.id
  }
  user.value = userData
} catch (e) {
  console.error('Failed to parse user in Substances:', e)
}

const substances = ref<Substance[]>([])
const loading = ref(false)
const searchTerm = ref('')
const editingId = ref<number | null>(null)
const form = ref({ formula: '', name: '' })

const fetchSubstances = async () => {
  loading.value = true
  try {
    const res = await substanceAPI.getSubstances()
    substances.value = res.data
  } catch (e) {
    console.error('Failed to fetch substances')
  } finally {
    loading.value = false
  }
}

const saveSub = async () => {
  if (!form.value.formula || !form.value.name) return
  loading.value = true
  try {
    if (editingId.value) {
      await substanceAPI.updateSubstance(editingId.value, form.value.formula, form.value.name)
      showAlert('Entry updated successfully', 'Success')
    } else {
      await substanceAPI.addSubstance(form.value.formula, form.value.name)
      showAlert('New substance added to registry', 'Done')
    }
    closeModal()
    fetchSubstances()
  } catch (e: any) {
    showAlert(e.response?.data?.error || 'Operation failed', 'Error')
  } finally {
    loading.value = false
  }
}

const editSub = (sub: Substance) => {
  editingId.value = sub.id
  form.value = { formula: sub.formula, name: sub.name }
}

const canApprove = (status: string) => {
  if (user.value?.role === 'admin') {
    return status === 'pending_admin' || status === 'pending_coworker'
  }
  if (user.value?.role === 'co-worker') {
    return status === 'pending_coworker'
  }
  return false
}

const approveSub = async (sub: Substance) => {
  try {
    const res = await substanceAPI.approveSubstance(sub.id, { 
      formula: sub.formula, 
      name: sub.name, 
      reject: false 
    })
    showAlert(`Status updated to ${res.data.status}`, 'Success')
    fetchSubstances()
  } catch (e: any) {
    showAlert(e.response?.data?.error || 'Approval failed', 'Error')
  }
}

const rejectSub = async (sub: Substance) => {
  const confirmed = await showConfirm('Are you sure you want to reject this entry?', 'Reject Request')
  if (!confirmed) return
  try {
    await substanceAPI.approveSubstance(sub.id, { reject: true })
    showAlert('Substance rejected', 'System')
    fetchSubstances()
  } catch (e: any) {
    showAlert(e.response?.data?.error || 'Operation failed', 'Error')
  }
}

const deleteSub = async (id: number) => {
  const confirmed = await showConfirm('Are you sure you want to purge this record?', 'Warning')
  if (!confirmed) return
  
  try {
    await substanceAPI.deleteSubstance(id)
    fetchSubstances()
  } catch (e) {
    showAlert('Deletion failed', 'Error')
  }
}

const closeModal = () => {
  editingId.value = null
  form.value = { formula: '', name: '' }
}

const filteredSubstances = computed(() => {
  if (!searchTerm.value) return substances.value
  const term = searchTerm.value
  return substances.value.filter(s => 
    s.formula.includes(term) || 
    s.name.toLowerCase().includes(term.toLowerCase())
  )
})

onMounted(fetchSubstances)
</script>