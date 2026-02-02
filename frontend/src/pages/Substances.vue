<template>
  <div class="min-h-screen bg-slate-50 dark:bg-[#0a0a0c] text-slate-900 dark:text-slate-200 p-4 lg:p-10 font-sans selection:bg-emerald-500/30">
    <div class="fixed inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] bg-emerald-500/5 rounded-full blur-[120px]" />
      <div class="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-blue-500/5 rounded-full blur-[120px]" />
      <div class="absolute inset-0 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-20 brightness-50 contrast-150" />
    </div>

    <div class="max-w-7xl mx-auto relative z-10">
      <header class="flex flex-col lg:flex-row items-center justify-between gap-8 mb-12">
        <div class="flex items-center gap-6">
          <div class="relative group">
            <div class="absolute inset-x-0 inset-y-0 bg-emerald-500 blur-2xl opacity-20 group-hover:opacity-40 transition-opacity" />
            <div class="w-16 h-16 rounded-2xl bg-[#111114] border border-emerald-500/40 flex items-center justify-center relative z-10 shadow-2xl">
              <FlaskConical class="w-8 h-8 text-emerald-400 group-hover:scale-110 transition-transform" />
            </div>
          </div>
          <div>
            <h1 class="text-3xl font-black text-white italic tracking-tighter uppercase flex items-center gap-3">
              Substance Wiki <span class="text-xs font-mono bg-emerald-500/20 text-emerald-400 px-2 py-1 rounded border border-emerald-500/30 not-italic uppercase">{{ user?.role || 'USER' }}</span>
            </h1>
            <p class="text-slate-500 text-sm font-bold tracking-widest uppercase mt-1">
              物质百科全书管理 / Substance encyclopedia
            </p>
          </div>
        </div>

        <div class="flex items-center gap-4">
          <div class="px-6 py-3 bg-[#111114] border border-white/5 rounded-2xl flex items-center gap-4 shadow-xl">
            <div class="flex flex-col items-end">
              <span class="text-[10px] font-black text-slate-500 uppercase">Registry Status</span>
              <span class="text-xs font-bold text-emerald-400 flex items-center gap-1.5">
                <span class="w-2 h-2 bg-emerald-500 rounded-full animate-pulse" />
                ONLINE
              </span>
            </div>
            <div class="w-px h-8 bg-white/5" />
            <router-link 
              to="/data"
              class="flex items-center gap-2 text-slate-400 hover:text-white transition-colors group"
            >
              <ArrowLeft class="w-4 h-4 group-hover:-translate-x-1 transition-transform" />
              <span class="text-xs font-black uppercase tracking-widest">Back</span>
            </router-link>
          </div>
        </div>
      </header>

      <div class="grid grid-cols-1 lg:grid-cols-4 gap-8">
        <!-- 添加/编辑面板 -->
        <div class="lg:col-span-1 bg-[#111114] border border-white/10 p-8 rounded-[2rem] shadow-xl h-fit sticky top-10">
          <div class="flex items-center gap-3 mb-8">
            <div class="w-10 h-10 rounded-xl bg-emerald-500/10 flex items-center justify-center border border-emerald-500/20">
              <Plus class="w-5 h-5 text-emerald-400" />
            </div>
            <h3 class="text-xl font-black text-white italic uppercase tracking-tight">
              {{ editingId ? 'Edit Entry' : 'New Entry' }}
            </h3>
          </div>
          
          <form @submit.prevent="saveSub" class="space-y-6">
            <div class="space-y-2">
              <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest pl-1">Chemical Formula</label>
              <input 
                v-model="form.formula" 
                type="text" 
                class="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm font-bold text-white focus:outline-none focus:border-emerald-500/50 transition-all placeholder:text-slate-700 uppercase italic tracking-tighter" 
                placeholder="E.G. H2O"
              />
            </div>
            <div class="space-y-2">
              <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest pl-1">Chinese Name</label>
              <input 
                v-model="form.name" 
                type="text" 
                class="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm font-bold text-white focus:outline-none focus:border-emerald-500/50 transition-all placeholder:text-slate-700" 
                placeholder="物质名称"
              />
            </div>
            
            <div class="pt-4 flex flex-col gap-3">
              <button 
                type="submit"
                :disabled="loading"
                class="w-full bg-emerald-600 hover:bg-emerald-500 px-6 py-4 rounded-2xl font-black text-white uppercase tracking-widest transition-all shadow-[0_10px_20px_rgba(16,185,129,0.2)] hover:scale-[1.02] active:scale-95 disabled:opacity-50"
              >
                {{ editingId ? 'Update Substance' : 'Add Substance' }}
              </button>
              <button 
                v-if="editingId"
                type="button"
                @click="closeModal"
                class="w-full bg-white/5 hover:bg-white/10 px-6 py-3 rounded-2xl font-black text-slate-400 uppercase tracking-widest transition-all"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>

        <!-- 列表面板 -->
        <div class="lg:col-span-3 bg-[#111114] border border-white/10 p-8 rounded-[2rem] shadow-xl">
          <div class="flex flex-col md:flex-row md:items-center justify-between gap-6 mb-8">
            <h3 class="text-xl font-black text-white flex items-center gap-3 italic">
              <Database class="w-6 h-6 text-blue-400" />
              DATABASE_ENTRIES <span class="text-slate-600 text-[10px] font-mono not-italic uppercase tracking-widest">/ Substances_Registry</span>
            </h3>
            <div class="flex items-center gap-4">
              <div class="relative group">
                <SearchIcon class="w-4 h-4 absolute left-4 top-1/2 -translate-y-1/2 text-slate-600 group-focus-within:text-emerald-400 transition-colors" />
                <input 
                  v-model="searchTerm"
                  type="text" 
                  placeholder="Search formula/name..."
                  class="bg-white/5 border border-white/10 rounded-xl pl-12 pr-6 py-2.5 text-xs focus:outline-none focus:border-emerald-500/50 w-full md:w-64 transition-all"
                />
              </div>
            </div>
          </div>
          
          <div class="overflow-x-auto custom-scrollbar">
            <table class="w-full text-left">
              <thead>
                <tr class="text-slate-600 text-[10px] font-black uppercase tracking-[0.2em] border-b border-white/5">
                  <th class="px-6 py-4">Formula</th>
                  <th class="px-6 py-4">Descriptor</th>
                  <th class="px-6 py-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-white/5 font-mono">
                <tr v-for="sub in filteredSubstances" :key="sub.id" class="hover:bg-white/5 transition-colors group">
                  <td class="px-6 py-5">
                    <span class="text-lg font-black italic text-white group-hover:text-emerald-400 transition-colors tracking-tighter uppercase">{{ sub.formula }}</span>
                  </td>
                  <td class="px-6 py-5">
                    <span class="text-xs font-bold text-slate-400 group-hover:text-white transition-colors">{{ sub.name }}</span>
                  </td>
                  <td class="px-6 py-5 text-right">
                    <div class="flex justify-end gap-2">
                      <button @click="editSub(sub)" class="p-2 hover:bg-emerald-500/10 text-slate-500 hover:text-emerald-400 rounded-xl transition-all">
                        <Edit class="w-4 h-4" />
                      </button>
                      <button @click="deleteSub(sub.id)" class="p-2 hover:bg-red-500/10 text-slate-500 hover:text-red-500 rounded-xl transition-all">
                        <Trash2 class="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
                <tr v-if="filteredSubstances.length === 0 && !loading">
                  <td colspan="4" class="py-20 text-center">
                    <FlaskConical class="w-12 h-12 text-slate-700 mx-auto mb-4" />
                    <p class="text-slate-500 font-medium italic">未检索到相关物质数据</p>
                  </td>
                </tr>
              </tbody>
            </table>
            
            <div v-if="loading" class="text-center py-12">
              <div class="w-8 h-8 border-2 border-emerald-500/30 border-t-emerald-500 rounded-full animate-spin mx-auto mb-4"></div>
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
import { substanceAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import { 
  FlaskConical, 
  Database, 
  Plus, 
  ArrowLeft, 
  Search as SearchIcon, 
  Trash2, 
  Edit 
} from 'lucide-vue-next'

interface Substance {
  id: number
  formula: string
  name: string
  elements: string
}

const { showAlert, showConfirm } = useDialog()
const user = ref<any>(JSON.parse(localStorage.getItem('user') || 'null'))
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
  const term = searchTerm.value.toLowerCase()
  return substances.value.filter(s => 
    s.formula.toLowerCase().includes(term) || 
    s.name.toLowerCase().includes(term)
  )
})

onMounted(fetchSubstances)
</script>