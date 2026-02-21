<template>
  <div class="min-h-screen bg-slate-50 dark:bg-[#0a0a0c] text-slate-900 dark:text-slate-200 p-4 lg:p-6 font-sans selection:bg-emerald-500/30">
    <div class="fixed inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] bg-emerald-500/5 rounded-full blur-[120px]" />
      <div class="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-blue-500/5 rounded-full blur-[120px]" />
      <div class="absolute inset-0 opacity-20 brightness-50 contrast-150" />
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

            <!-- 批量操作工具栏 -->
            <div v-if="(user.role === 'admin' || user.role === 'co-worker') && selectedSubstances.size > 0"
                 class="flex items-center gap-2 px-4 py-2 bg-blue-500/10 rounded-xl border border-blue-500/20">
              <span class="text-xs font-bold text-blue-600 dark:text-blue-400">
                已选: {{ selectedSubstances.size }}
              </span>
              <button
                @click="handleBatchApprove"
                class="px-3 py-1.5 bg-emerald-600/10 hover:bg-emerald-600 text-emerald-600 hover:text-white rounded-lg text-xs font-black uppercase tracking-widest transition-all"
              >
                批准
              </button>
              <button
                @click="handleBatchReject"
                class="px-3 py-1.5 bg-red-600/10 hover:bg-red-600 text-red-600 hover:text-white rounded-lg text-xs font-black uppercase tracking-widest transition-all"
              >
                拒绝
              </button>
            </div>

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
              <!-- 导出按钮 -->
              <button
                v-if="user.role === 'admin' || user.role === 'co-worker'"
                @click="handleExportSubstances"
                class="flex items-center gap-2 px-3 py-2 bg-emerald-600/10 hover:bg-emerald-600 text-emerald-600 hover:text-white rounded-xl text-xs font-black uppercase tracking-widest transition-all border border-emerald-600/20"
                title="导出物质数据为Excel"
              >
                <Download class="w-3.5 h-3.5" />
                Export
              </button>
            </div>
          </div>

          <!-- 过滤器标签 -->
          <div class="flex flex-wrap items-center gap-3 mb-6 pb-6 border-b border-slate-100 dark:border-white/5">
            <span class="text-[10px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest mr-2">Quick Filters:</span>
            
            <!-- 状态过滤器 -->
            <div class="flex items-center bg-slate-50 dark:bg-white/5 p-1 rounded-xl border border-slate-200 dark:border-white/10 overflow-x-auto no-scrollbar">
              <button 
                v-for="status in ['all', 'pending_coworker', 'pending_admin', 'approved', 'rejected']" 
                :key="status"
                @click="filterStatus = status"
                :class="[
                  'px-3 py-1.5 rounded-lg text-[10px] font-black uppercase tracking-tight transition-all whitespace-nowrap',
                  filterStatus === status 
                    ? 'bg-white dark:bg-emerald-500 text-emerald-600 dark:text-white shadow-sm' 
                    : 'text-slate-400 hover:text-slate-600 dark:hover:text-slate-300'
                ]"
              >
                {{ status === 'all' ? 'All' : status.replace('pending_', 'P_').toUpperCase() }}
              </button>
            </div>

            <!-- 特殊状态过滤器 -->
            <button 
              @click="filterNeedsImprovement = filterNeedsImprovement === true ? null : true"
              :class="[
                'px-4 py-1.5 rounded-xl text-[10px] font-black uppercase tracking-tight transition-all border',
                filterNeedsImprovement === true
                  ? 'bg-amber-500/10 border-amber-500/50 text-amber-600 dark:text-amber-400' 
                  : 'bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/10 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300'
              ]"
            >
              需完善
            </button>

            <button 
              @click="filterInvalidElements = filterInvalidElements === true ? null : true"
              :class="[
                'px-4 py-1.5 rounded-xl text-[10px] font-black uppercase tracking-tight transition-all border',
                filterInvalidElements === true
                  ? 'bg-red-500/10 border-red-500/50 text-red-600 dark:text-red-400' 
                  : 'bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/10 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300'
              ]"
            >
              无效元素
            </button>

            <button 
              v-if="filterStatus !== 'all' || filterNeedsImprovement !== null || filterInvalidElements !== null"
              @click="() => { filterStatus = 'all'; filterNeedsImprovement = null; filterInvalidElements = null; searchTerm = '' }"
              class="text-[10px] font-black text-slate-400 hover:text-red-500 transition-colors uppercase ml-2 flex items-center gap-1"
            >
              <Plus class="w-3 h-3 rotate-45" /> Clear All
            </button>
          </div>
          
          <div class="overflow-x-auto custom-scrollbar">
            <table class="w-full text-left">
              <thead>
                <tr class="text-slate-400 dark:text-slate-600 text-[10px] font-black uppercase tracking-[0.2em] border-b border-slate-100 dark:border-white/5">
                  <!-- 复选框列 -->
                  <th v-if="user.role === 'admin' || user.role === 'co-worker'" class="px-4 py-2.5 w-12">
                    <input
                      type="checkbox"
                      :checked="selectAll"
                      @change="toggleSelectAll"
                      :disabled="filteredSubstances.length === 0"
                      class="w-4 h-4 rounded border-slate-300 dark:border-white/20 text-emerald-600 focus:ring-emerald-500"
                    />
                  </th>
                  <th class="px-4 py-2.5">Formula / Name</th>
                  <th class="px-4 py-2.5">Status</th>
                  <th class="px-4 py-2.5">Author</th>
                  <th class="px-4 py-2.5 text-right">Actions</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 dark:divide-white/5 font-mono">
                <tr v-for="sub in filteredSubstances" :key="sub.id" class="hover:bg-slate-50 dark:hover:bg-white/5 transition-colors group">
                  <!-- 复选框列 -->
                  <td v-if="user.role === 'admin' || user.role === 'co-worker'" class="px-4 py-3">
                    <input
                      type="checkbox"
                      :checked="selectedSubstances.has(sub.id)"
                      @change="toggleSelect(sub.id)"
                      class="w-4 h-4 rounded border-slate-300 dark:border-white/20 text-emerald-600 focus:ring-emerald-500"
                    />
                  </td>
                  <td class="px-4 py-3">
                    <div class="flex flex-col">
                      <span class="text-base font-black italic text-slate-900 dark:text-white group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors tracking-tighter">{{ sub.formula }}</span>
                      <span class="text-[10px] font-bold text-slate-400 group-hover:text-slate-600 dark:group-hover:text-slate-300 transition-colors">{{ sub.name }}</span>
                    </div>
                  </td>
                  <td class="px-4 py-3 flex flex-col gap-1 items-start">
                    <span :class="[
                      'px-1.5 py-0.5 rounded text-[8px] font-black uppercase letter-spacing-widest border',
                      sub.status === 'approved' ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' :
                      sub.status.startsWith('pending') ? 'bg-blue-500/10 text-blue-400 border-blue-500/20' :
                      'bg-red-500/10 text-red-400 border-red-500/20'
                    ]">
                      {{ sub.status }}
                    </span>
                    <span v-if="sub.needs_improvement" class="px-1.5 py-0.5 rounded text-[8px] font-black uppercase letter-spacing-widest bg-amber-500/10 text-amber-500 border border-amber-500/20 flex items-center gap-1">
                      <AlertTriangle class="w-2.5 h-2.5" />
                      需完善
                    </span>
                    <span v-if="sub.has_invalid_elements" class="px-1.5 py-0.5 rounded text-[8px] font-black uppercase letter-spacing-widest bg-red-500/10 text-red-500 border border-red-500/20 flex items-center gap-1">
                      <AlertTriangle class="w-2.5 h-2.5" />
                      无效元素
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
                        <button @click="approveSub(sub)" class="p-1.5 hover:bg-emerald-500/10 text-emerald-500/50 hover:text-emerald-400 rounded-lg transition-all" title="批准建议 / Approve">
                          <Check class="w-3.5 h-3.5" />
                        </button>
                      </template>
                      
                      <button @click="editSub(sub)" class="p-1.5 hover:bg-emerald-500/10 text-slate-500 hover:text-emerald-400 rounded-lg transition-all" title="编辑详情 / Edit">
                        <Edit class="w-3.5 h-3.5" />
                      </button>

                      <button v-if="canDeleteSub(sub)" 
                              @click="deleteSub(sub.id)" 
                              class="p-1.5 hover:bg-red-500/10 text-red-500/50 hover:text-red-500 rounded-lg transition-all" 
                              :title="sub.status === 'approved' ? '从百科中删除 / Delete from Wiki' : '拒绝该建议 / Reject Suggestion'">
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
import { substanceAPI, adminAPI } from '../utils/api'
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
  User as UserIcon,
  AlertTriangle,
  Download
} from 'lucide-vue-next'

interface Substance {
  id: number
  formula: string
  name: string
  elements: string
  status: string
  group_id: number | null
  needs_improvement: boolean
  has_invalid_elements: boolean
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

// 过滤状态
const filterStatus = ref<string>('all')
const filterNeedsImprovement = ref<boolean | null>(null)
const filterInvalidElements = ref<boolean | null>(null)

// 批量选择状态
const selectedSubstances = ref<Set<number>>(new Set())
const selectAll = ref(false)

const fetchSubstances = async () => {
  loading.value = true
  try {
    const res = await substanceAPI.getSubstances()
    substances.value = res.data || []
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
      // 如果是管理员，直接更新；否则提交更新建议
      if (user.value?.role === 'admin' || user.value?.role === 'co-worker') {
        await substanceAPI.updateSubstance(editingId.value, form.value.formula, form.value.name)
        showAlert('物质已更新', '成功')
      } else {
        await substanceAPI.submitSubstanceUpdate(editingId.value, form.value.formula, form.value.name)
        showAlert('更新建议已提交，等待管理员审核', '已提交')
      }
    } else {
      // 提交新物质建议（所有用户都使用此接口）
      await substanceAPI.submitNewSubstance(form.value.formula, form.value.name)
      if (user.value?.role === 'admin' || user.value?.role === 'co-worker') {
        showAlert('新物质已添加到数据库', '完成')
      } else {
        showAlert('物质建议已提交，等待管理员审核', '已提交')
      }
    }
    closeModal()
    fetchSubstances()
  } catch (e: any) {
    showAlert(e.response?.data?.error || '操作失败', '错误')
  } finally {
    loading.value = false
  }
}

const editSub = (sub: Substance) => {
  editingId.value = sub.id
  form.value = { formula: sub.formula, name: sub.name }
}

const canApprove = (status: string) => {
  if (user.value?.role === 'admin' || user.value?.role === 'co-worker') {
    return status === 'pending'
  }
  return false
}

const approveSub = async (sub: Substance) => {
  try {
    await substanceAPI.approveSubstance(sub.id)
    showAlert('物质已批准', '成功')
    fetchSubstances()
  } catch (e: any) {
    showAlert(e.response?.data?.error || '批准失败', '错误')
  }
}



const deleteSub = async (id: number) => {
  const confirmed = await showConfirm('确定要删除此记录吗？', '警告')
  if (!confirmed) return
  
  try {
    await substanceAPI.rejectSubstance(id)
    showAlert('物质已删除', '成功')
    fetchSubstances()
  } catch (e: any) {
    showAlert(e.response?.data?.error || '删除失败', '错误')
  }
}

const closeModal = () => {
  editingId.value = null
  form.value = { formula: '', name: '' }
}

const canDeleteSub = (sub: Substance) => {
  if (user.value?.role === 'admin') return true
  if (sub.status === 'approved') return false
  return user.value?.role === 'co-worker'
}

// 批量操作相关
const toggleSelectAll = () => {
  if (selectAll.value) {
    selectedSubstances.value.clear()
    selectAll.value = false
  } else {
    filteredSubstances.value.forEach(s => selectedSubstances.value.add(s.id))
    selectAll.value = true
  }
}

const toggleSelect = (id: number) => {
  if (selectedSubstances.value.has(id)) {
    selectedSubstances.value.delete(id)
  } else {
    selectedSubstances.value.add(id)
  }
  selectAll.value = selectedSubstances.value.size === filteredSubstances.value.length
}

// 导出物质数据
const handleExportSubstances = async () => {
  try {
    const response = await adminAPI.exportSubstances()
    const url = window.URL.createObjectURL(response.data)
    const link = document.createElement('a')
    link.href = url
    link.download = `substances_${new Date().toISOString().replace(/[:.]/g, '-')}.xlsx`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    showAlert('物质数据已导出', '成功')
  } catch (e: any) {
    showAlert(e.response?.data?.error || '导出失败', '错误')
  }
}

// 批量批准物质
const handleBatchApprove = async () => {
  if (selectedSubstances.value.size === 0) return
  if (selectedSubstances.value.size > 100) {
    showAlert('一次最多批准100条记录', '超出限制')
    return
  }

  const confirmed = await showConfirm(`确定批准选中的 ${selectedSubstances.value.size} 条物质吗？`, '批量批准')
  if (!confirmed) return

  try {
    const groupIDs = Array.from(selectedSubstances.value).map(id => {
      const sub = substances.value.find(s => s.id === id)
      return sub?.group_id || id
    })
    const response = await adminAPI.batchApproveSubstances(groupIDs)
    showAlert(response.data.message || '批量批准成功', '完成')
    selectedSubstances.value.clear()
    selectAll.value = false
    fetchSubstances()
  } catch (e: any) {
    showAlert(e.response?.data?.error || '批量批准失败', '错误')
  }
}

// 批量拒绝物质
const handleBatchReject = async () => {
  if (selectedSubstances.value.size === 0) return
  if (selectedSubstances.value.size > 100) {
    showAlert('一次最多拒绝100条记录', '超出限制')
    return
  }

  const confirmed = await showConfirm(`确定拒绝选中的 ${selectedSubstances.value.size} 条物质吗？此操作不可恢复！`, '批量拒绝')
  if (!confirmed) return

  try {
    const groupIDs = Array.from(selectedSubstances.value).map(id => {
      const sub = substances.value.find(s => s.id === id)
      return sub?.group_id || id
    })
    const response = await adminAPI.batchRejectSubstances(groupIDs)
    showAlert(response.data.message || '批量拒绝成功', '完成')
    selectedSubstances.value.clear()
    selectAll.value = false
    fetchSubstances()
  } catch (e: any) {
    showAlert(e.response?.data?.error || '批量拒绝失败', '错误')
  }
}

const filteredSubstances = computed(() => {
  let filtered = substances.value

  // 1. 搜索词筛选
  if (searchTerm.value) {
    const term = searchTerm.value.toLowerCase()
    filtered = filtered.filter(s => 
      s.formula.toLowerCase().includes(term) || 
      s.name.toLowerCase().includes(term)
    )
  }

  // 2. 状态筛选
  if (filterStatus.value !== 'all') {
    filtered = filtered.filter(s => s.status === filterStatus.value)
  }

  // 3. 需完善筛选
  if (filterNeedsImprovement.value !== null) {
    filtered = filtered.filter(s => s.needs_improvement === filterNeedsImprovement.value)
  }

  // 4. 无效元素筛选
  if (filterInvalidElements.value !== null) {
    filtered = filtered.filter(s => s.has_invalid_elements === filterInvalidElements.value)
  }

  return filtered
})

onMounted(fetchSubstances)
</script>