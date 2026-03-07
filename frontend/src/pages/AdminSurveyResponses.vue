<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { adminAPI } from '../utils/api'
import {
  ArrowLeft,
  FileText,
  Download,
  ChevronUp,
  ChevronDown,
  ChevronsUpDown,
  Users
} from 'lucide-vue-next'
import { cn } from '../utils/cn'

const route = useRoute()
const router = useRouter()

const surveyID = parseInt(route.params.id as string)
const survey = ref<any>(null)
const responses = ref<any[]>([])
const loading = ref(true)
const sortBy = ref('created_at')
const sortOrder = ref<'asc' | 'desc'>('desc')

const loadData = async () => {
  loading.value = true
  try {
    const [surveyRes, responsesRes] = await Promise.all([
      adminAPI.getSurveys(),
      adminAPI.getSurveyResponses(surveyID, sortBy.value, sortOrder.value)
    ])
    const all: any[] = surveyRes.data || []
    survey.value = all.find((s: any) => s.id === surveyID) || null
    responses.value = responsesRes.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const toggleSort = (field: string) => {
  if (sortBy.value === field) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = field
    sortOrder.value = 'desc'
  }
  loadData()
}

const getAnswer = (res: any, questionID: number): string => {
  if (!res.answers) return '—'
  const ans = res.answers.find((a: any) => a.question_id === questionID)
  if (!ans || !ans.answer) return '—'
  const raw = ans.answer
  // Try to parse JSON arrays (checkbox)
  if (raw.startsWith('[')) {
    try {
      const arr = JSON.parse(raw)
      return Array.isArray(arr) ? arr.join('、') : raw
    } catch {
      return raw
    }
  }
  return raw
}

const handleExport = async () => {
  try {
    const response = await adminAPI.exportSurvey(surveyID)
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `survey_${surveyID}_export.xlsx`)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  } catch (e) {
    console.error(e)
  }
}

const sortIcon = (field: string) => {
  if (sortBy.value !== field) return null
  return sortOrder.value === 'asc' ? 'up' : 'down'
}

onMounted(loadData)
</script>

<template>
  <div class="min-h-screen bg-slate-50 dark:bg-[#070709] text-slate-900 dark:text-white">
    <!-- 顶栏 -->
    <header class="sticky top-0 z-20 bg-white/80 dark:bg-black/60 backdrop-blur-xl border-b border-slate-100 dark:border-white/5 px-4 lg:px-8 py-3 flex items-center gap-4">
      <button
        @click="router.back()"
        class="p-2 rounded-xl hover:bg-slate-100 dark:hover:bg-white/10 transition-all text-slate-500 dark:text-slate-400"
      >
        <ArrowLeft class="w-4 h-4" />
      </button>

      <div class="flex items-center gap-3 flex-1 min-w-0">
        <FileText class="w-4 h-4 text-indigo-500 shrink-0" />
        <div class="min-w-0">
          <p class="text-[9px] font-black uppercase tracking-widest text-slate-400">SURVEY_{{ surveyID }} / RESPONSES</p>
          <h1 class="text-sm font-black truncate text-slate-900 dark:text-white italic uppercase">
            {{ survey?.title || '...' }}
          </h1>
        </div>
      </div>

      <div class="flex items-center gap-2 shrink-0">
        <span class="hidden sm:flex items-center gap-1.5 px-3 py-1.5 bg-slate-100 dark:bg-white/5 rounded-xl text-[9px] font-black uppercase tracking-widest text-slate-400 border border-slate-200 dark:border-white/5">
          <Users class="w-3 h-3" />
          {{ responses.length }} 份
        </span>
        <button
          @click="handleExport"
          class="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl text-[10px] font-black uppercase tracking-widest transition-all shadow-lg shadow-indigo-500/20 active:scale-95"
        >
          <Download class="w-3.5 h-3.5" />
          <span class="hidden sm:inline">下载 Excel</span>
        </button>
      </div>
    </header>

    <!-- 加载中 -->
    <div v-if="loading" class="flex items-center justify-center py-32 text-slate-400 text-[10px] font-black uppercase tracking-[0.3em]">
      LOADING...
    </div>

    <!-- 无数据 -->
    <div v-else-if="responses.length === 0" class="flex flex-col items-center justify-center py-32 gap-4 text-slate-400">
      <FileText class="w-10 h-10 opacity-30" />
      <p class="text-[10px] font-black uppercase tracking-[0.3em] italic">/ NO_SUBMISSIONS_FOUND</p>
    </div>

    <!-- 表格 -->
    <div v-else-if="survey" class="p-4 lg:p-8">
      <div class="overflow-x-auto rounded-[1.5rem] border border-slate-200 dark:border-white/5 shadow-sm">
        <table class="w-full text-[11px] border-collapse">
          <thead>
            <tr class="bg-slate-100 dark:bg-white/5 border-b border-slate-200 dark:border-white/5">
              <!-- UID列 -->
              <th
                @click="toggleSort('user_uid')"
                class="px-4 py-3 text-left font-black uppercase tracking-widest text-slate-500 dark:text-slate-400 cursor-pointer select-none whitespace-nowrap hover:text-indigo-500 transition-colors group w-32"
              >
                <div class="flex items-center gap-1.5">
                  UID
                  <ChevronUp v-if="sortIcon('user_uid') === 'up'" class="w-3 h-3 text-indigo-500" />
                  <ChevronDown v-else-if="sortIcon('user_uid') === 'down'" class="w-3 h-3 text-indigo-500" />
                  <ChevronsUpDown v-else class="w-3 h-3 opacity-30 group-hover:opacity-60" />
                </div>
              </th>
              <!-- 时间列 -->
              <th
                @click="toggleSort('created_at')"
                class="px-4 py-3 text-left font-black uppercase tracking-widest text-slate-500 dark:text-slate-400 cursor-pointer select-none whitespace-nowrap hover:text-indigo-500 transition-colors group"
              >
                <div class="flex items-center gap-1.5">
                  提交时间
                  <ChevronUp v-if="sortIcon('created_at') === 'up'" class="w-3 h-3 text-indigo-500" />
                  <ChevronDown v-else-if="sortIcon('created_at') === 'down'" class="w-3 h-3 text-indigo-500" />
                  <ChevronsUpDown v-else class="w-3 h-3 opacity-30 group-hover:opacity-60" />
                </div>
              </th>
              <!-- 每道题一列 -->
              <th
                v-for="q in survey.questions"
                :key="q.id"
                class="px-4 py-3 text-left font-black uppercase tracking-widest text-slate-500 dark:text-slate-400 whitespace-nowrap max-w-[200px]"
              >
                <div class="flex flex-col gap-0.5">
                  <span class="truncate max-w-[180px]" :title="q.title">{{ q.title }}</span>
                  <span class="text-[8px] font-mono text-slate-400 dark:text-slate-600 normal-case tracking-normal">
                    {{ q.type }}{{ q.is_required ? ' · 必填' : '' }}
                  </span>
                </div>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(res, i) in responses"
              :key="res.id"
              :class="cn(
                'border-b border-slate-100 dark:border-white/5 transition-colors',
                i % 2 === 0 ? 'bg-white dark:bg-transparent' : 'bg-slate-50/50 dark:bg-white/[0.02]',
                'hover:bg-indigo-50/50 dark:hover:bg-indigo-500/5'
              )"
            >
              <!-- UID -->
              <td class="px-4 py-3 whitespace-nowrap w-32">
                <div class="inline-flex items-center justify-center min-w-[72px] px-3 h-8 rounded-xl bg-indigo-500/10 text-indigo-500 dark:text-indigo-400 text-[11px] font-black border border-indigo-500/20">
                  {{ res.user_uid }}
                </div>
              </td>
              <!-- 时间 -->
              <td class="px-4 py-3 whitespace-nowrap font-mono text-slate-500 dark:text-slate-400">
                {{ new Date(res.created_at).toLocaleString() }}
              </td>
              <!-- 答案 -->
              <td
                v-for="q in survey.questions"
                :key="q.id"
                class="px-4 py-3 text-slate-700 dark:text-slate-300 max-w-[220px]"
              >
                <span
                  v-if="getAnswer(res, q.id) !== '—'"
                  class="block truncate"
                  :title="getAnswer(res, q.id)"
                >{{ getAnswer(res, q.id) }}</span>
                <span v-else class="text-slate-300 dark:text-slate-600">—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <p class="mt-4 text-[9px] font-black uppercase tracking-widest text-slate-400 text-center">
        共 {{ responses.length }} 份答卷 · 点击列标题排序
      </p>
    </div>
  </div>
</template>
