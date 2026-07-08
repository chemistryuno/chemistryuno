<template>
  <div class="p-4 max-w-7xl mx-auto">
    <h1 class="text-2xl font-bold mb-6">活动管理</h1>

    <!-- Tab switch -->
    <div class="flex gap-2 mb-6">
      <button
        v-for="tab in ['gantt', 'table']"
        :key="tab"
        @click="activeTab = tab"
        :class="['px-4 py-2 rounded text-sm font-medium transition-colors',
          activeTab === tab ? 'bg-blue-600 text-white' : 'bg-gray-100 hover:bg-gray-200']"
      >
        {{ tab === 'gantt' ? '甘特图' : '表格' }}
      </button>
      <button @click="showVersionForm = true" class="ml-auto px-4 py-2 bg-green-600 text-white rounded text-sm">+ 新建版本</button>
      <button @click="openCreateActivity" class="px-4 py-2 bg-blue-600 text-white rounded text-sm">+ 新建活动</button>
    </div>

    <!-- Gantt view -->
    <div v-if="activeTab === 'gantt'" class="overflow-x-auto">
      <ActivityGantt :versions="versions" :activities="activities" @edit="openEditActivity" />
    </div>

    <!-- Table view -->
    <div v-else>
      <table class="w-full text-sm border-collapse">
        <thead>
          <tr class="bg-gray-50 text-left">
            <th class="px-3 py-2 border">ID</th>
            <th class="px-3 py-2 border">名称</th>
            <th class="px-3 py-2 border">类型</th>
            <th class="px-3 py-2 border">开始时间</th>
            <th class="px-3 py-2 border">结束时间</th>
            <th class="px-3 py-2 border">状态</th>
            <th class="px-3 py-2 border">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="act in activities" :key="act.id" class="border-b hover:bg-gray-50">
            <td class="px-3 py-2 border">{{ act.id }}</td>
            <td class="px-3 py-2 border">{{ act.name }}</td>
            <td class="px-3 py-2 border">
              <span :class="typeColor(act.type)" class="px-2 py-0.5 rounded text-xs">{{ act.type }}</span>
            </td>
            <td class="px-3 py-2 border">{{ formatDate(act.start_time) }}</td>
            <td class="px-3 py-2 border">{{ formatDate(act.end_time) }}</td>
            <td class="px-3 py-2 border">
              <span :class="statusClass(act)" class="px-2 py-0.5 rounded text-xs font-medium">
                {{ statusLabel(act) }}
              </span>
            </td>
            <td class="px-3 py-2 border">
              <button @click="openEditActivity(act)" class="text-blue-600 hover:underline mr-2">编辑</button>
              <button @click="toggleActivity(act)" class="hover:underline" :class="act.is_active ? 'text-red-500' : 'text-green-600'">
                {{ act.is_active ? '禁用' : '启用' }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Activity form modal -->
    <div v-if="showActivityForm" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg p-6 w-full max-w-lg shadow-xl">
        <h2 class="text-lg font-bold mb-4">{{ editingActivity ? '编辑活动' : '新建活动' }}</h2>
        <div class="space-y-3">
          <div>
            <label class="block text-sm font-medium mb-1">名称</label>
            <input v-model="form.name" class="w-full border rounded px-3 py-2 text-sm" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">类型</label>
            <select v-model="form.type" class="w-full border rounded px-3 py-2 text-sm">
              <option value="double_points">双倍积分</option>
              <option value="bingo">BINGO对战</option>
              <option value="other">其他</option>
            </select>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-sm font-medium mb-1">开始时间</label>
              <input type="datetime-local" v-model="form.start_time" class="w-full border rounded px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">结束时间</label>
              <input type="datetime-local" v-model="form.end_time" class="w-full border rounded px-3 py-2 text-sm" />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">参数 (JSON)</label>
            <textarea v-model="form.params" rows="3" class="w-full border rounded px-3 py-2 text-sm font-mono text-xs" placeholder='{"daily_limit": 3}'></textarea>
          </div>
          <div v-if="formError" class="text-red-500 text-sm">{{ formError }}</div>
        </div>
        <div class="flex gap-2 mt-4 justify-end">
          <button @click="showActivityForm = false" class="px-4 py-2 border rounded text-sm">取消</button>
          <button @click="saveActivity" class="px-4 py-2 bg-blue-600 text-white rounded text-sm">保存</button>
        </div>
      </div>
    </div>

    <!-- Version form modal -->
    <div v-if="showVersionForm" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg p-6 w-full max-w-md shadow-xl">
        <h2 class="text-lg font-bold mb-4">新建版本</h2>
        <div class="space-y-3">
          <div>
            <label class="block text-sm font-medium mb-1">版本名称</label>
            <input v-model="versionForm.name" class="w-full border rounded px-3 py-2 text-sm" />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-sm font-medium mb-1">开始日期</label>
              <input type="date" v-model="versionForm.start_date" class="w-full border rounded px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">结束日期</label>
              <input type="date" v-model="versionForm.end_date" class="w-full border rounded px-3 py-2 text-sm" />
            </div>
          </div>
        </div>
        <div class="flex gap-2 mt-4 justify-end">
          <button @click="showVersionForm = false" class="px-4 py-2 border rounded text-sm">取消</button>
          <button @click="saveVersion" class="px-4 py-2 bg-green-600 text-white rounded text-sm">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { activityAPI } from '../utils/api'
import ActivityGantt from '../components/ActivityGantt.vue'

const activeTab = ref('table')
const activities = ref<any[]>([])
const versions = ref<any[]>([])
const showActivityForm = ref(false)
const showVersionForm = ref(false)
const editingActivity = ref<any>(null)
const formError = ref('')

const form = ref({
  name: '', type: 'double_points', start_time: '', end_time: '', params: ''
})
const versionForm = ref({ name: '', start_date: '', end_date: '' })

onMounted(async () => {
  await Promise.all([loadActivities(), loadVersions()])
})

async function loadActivities() {
  const res = await activityAPI.listActivities()
  activities.value = res.data || []
}

async function loadVersions() {
  const res = await activityAPI.listVersions()
  versions.value = res.data || []
}

function openCreateActivity() {
  editingActivity.value = null
  form.value = { name: '', type: 'double_points', start_time: '', end_time: '', params: '{}' }
  formError.value = ''
  showActivityForm.value = true
}

function openEditActivity(act: any) {
  editingActivity.value = act
  form.value = {
    name: act.name,
    type: act.type,
    start_time: toLocalDatetime(act.start_time),
    end_time: toLocalDatetime(act.end_time),
    params: JSON.stringify(act.params || {}, null, 2),
  }
  formError.value = ''
  showActivityForm.value = true
}

async function saveActivity() {
  formError.value = ''
  let params: any = {}
  try { params = JSON.parse(form.value.params || '{}') } catch { formError.value = '参数JSON格式错误'; return }

  const payload = {
    name: form.value.name,
    type: form.value.type,
    start_time: new Date(form.value.start_time).toISOString(),
    end_time: new Date(form.value.end_time).toISOString(),
    params,
  }

  try {
    if (editingActivity.value) {
      await activityAPI.updateActivity(editingActivity.value.id, payload)
    } else {
      await activityAPI.createActivity(payload)
    }
    showActivityForm.value = false
    await loadActivities()
  } catch (e: any) {
    formError.value = e.response?.data?.error || '保存失败'
  }
}

async function toggleActivity(act: any) {
  await activityAPI.toggleActivity(act.id, !act.is_active)
  await loadActivities()
}

async function saveVersion() {
  await activityAPI.createVersion({
    name: versionForm.value.name,
    start_date: new Date(versionForm.value.start_date).toISOString(),
    end_date: new Date(versionForm.value.end_date).toISOString(),
  })
  showVersionForm.value = false
  await loadVersions()
}

function formatDate(s: string) {
  if (!s) return ''
  return new Date(s).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

function toLocalDatetime(s: string) {
  if (!s) return ''
  const d = new Date(s)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function statusLabel(act: any) {
  if (!act.is_active) return '已禁用'
  const now = new Date()
  const s = new Date(act.start_time)
  const e = new Date(act.end_time)
  if (now < s) return '未开始'
  if (now > e) return '已结束'
  return '进行中'
}

function statusClass(act: any) {
  const s = statusLabel(act)
  if (s === '进行中') return 'bg-green-100 text-green-700'
  if (s === '未开始') return 'bg-blue-100 text-blue-700'
  if (s === '已结束') return 'bg-gray-100 text-gray-500'
  return 'bg-red-100 text-red-500'
}

function typeColor(type: string) {
  const map: Record<string, string> = {
    double_points: 'bg-yellow-100 text-yellow-700',
    bingo: 'bg-purple-100 text-purple-700',
  }
  return map[type] || 'bg-gray-100 text-gray-600'
}
</script>
