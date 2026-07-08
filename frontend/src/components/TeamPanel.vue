<template>
  <div class="bg-white dark:bg-gray-900 rounded-xl border p-4 space-y-4 text-sm">
    <h3 class="font-bold text-base">队伍</h3>

    <!-- Not in a team -->
    <template v-if="!myTeam">
      <div class="space-y-2">
        <div>
          <label class="block text-xs text-gray-500 mb-1">创建队伍</label>
          <div class="flex gap-2">
            <input v-model="createName" placeholder="队伍名称" class="flex-1 border rounded px-3 py-1.5 text-sm" />
            <button @click="handleCreate" class="px-3 py-1.5 bg-blue-600 text-white rounded text-sm">创建</button>
          </div>
        </div>
        <div>
          <label class="block text-xs text-gray-500 mb-1">加入队伍</label>
          <div class="flex gap-2">
            <input v-model="joinCode" placeholder="邀请码" class="flex-1 border rounded px-3 py-1.5 text-sm uppercase" maxlength="6" />
            <button @click="handleJoin" class="px-3 py-1.5 bg-green-600 text-white rounded text-sm">加入</button>
          </div>
        </div>
        <p v-if="error" class="text-red-500 text-xs">{{ error }}</p>
      </div>
    </template>

    <!-- In a team -->
    <template v-else>
      <div class="flex items-center justify-between">
        <div>
          <p class="font-semibold">{{ myTeam.team.name }}</p>
          <p class="text-xs text-gray-400">邀请码: <span class="font-mono font-bold tracking-widest">{{ myTeam.team.invite_code }}</span></p>
        </div>
        <button
          @click="isLeader ? handleDisband() : handleLeave()"
          class="text-xs text-red-500 hover:underline"
        >
          {{ isLeader ? '解散队伍' : '退出队伍' }}
        </button>
      </div>
      <div>
        <p class="text-xs text-gray-400 mb-1">成员 ({{ myTeam.members?.length || 0 }}/4)</p>
        <div class="flex flex-wrap gap-2">
          <div
            v-for="uid in myTeam.members"
            :key="uid"
            class="flex items-center gap-1 bg-gray-100 dark:bg-gray-800 rounded-full px-2 py-0.5 text-xs"
          >
            <span>{{ uid === myTeam.team.leader_uid ? '👑' : '👤' }}</span>
            <span>UID {{ uid }}</span>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { teamAPI } from '../utils/api'

const emit = defineEmits(['teamChanged'])
const myTeam = ref<any>(null)
const createName = ref('')
const joinCode = ref('')
const error = ref('')

const uid = computed(() => {
  try { return JSON.parse(localStorage.getItem('user') || '{}').uid } catch { return 0 }
})
const isLeader = computed(() => myTeam.value?.team?.leader_uid === uid.value)

onMounted(loadTeam)

async function loadTeam() {
  try {
    const res = await teamAPI.getMyTeam()
    myTeam.value = res.data?.team ? res.data : null
  } catch { myTeam.value = null }
}

async function handleCreate() {
  error.value = ''
  if (!createName.value.trim()) { error.value = '请输入队伍名称'; return }
  try {
    await teamAPI.createTeam(createName.value.trim())
    createName.value = ''
    await loadTeam()
    emit('teamChanged')
  } catch (e: any) {
    error.value = e.response?.data?.error || '创建失败'
  }
}

async function handleJoin() {
  error.value = ''
  if (!joinCode.value.trim()) { error.value = '请输入邀请码'; return }
  try {
    await teamAPI.joinTeam(joinCode.value.trim().toUpperCase())
    joinCode.value = ''
    await loadTeam()
    emit('teamChanged')
  } catch (e: any) {
    error.value = e.response?.data?.error || '加入失败'
  }
}

async function handleLeave() {
  try { await teamAPI.leaveTeam(); await loadTeam(); emit('teamChanged') } catch (e: any) {
    error.value = e.response?.data?.error || '退出失败'
  }
}

async function handleDisband() {
  try { await teamAPI.disbandTeam(); await loadTeam(); emit('teamChanged') } catch (e: any) {
    error.value = e.response?.data?.error || '解散失败'
  }
}
</script>
