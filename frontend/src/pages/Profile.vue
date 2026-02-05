<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import { 
  ArrowLeft, 
  Shield, 
  Award,
  MessageSquare,
  Trophy
} from 'lucide-vue-next'

// Components
import ProfileHeader from '../components/profile/ProfileHeader.vue'
import StatsGrid from '../components/profile/StatsGrid.vue'
import SecurityPanel from '../components/profile/SecurityPanel.vue'
import SettingsPanel from '../components/profile/SettingsPanel.vue'
import CustomDecks from '../components/profile/CustomDecks.vue'
import MatchHistory from '../components/profile/MatchHistory.vue'
import ChangeAvatarModal from '../components/profile/ChangeAvatarModal.vue'
import ChangePasswordModal from '../components/profile/ChangePasswordModal.vue'
import TwoFactorSetupModal from '../components/profile/TwoFactorSetupModal.vue'
import HardwareKeyModal from '../components/profile/HardwareKeyModal.vue'
import DeviceManagementModal from '../components/profile/DeviceManagementModal.vue'
import { LayoutDashboard, ShieldCheck, FlaskConical, History, Sliders, Menu, X as CloseIcon } from 'lucide-vue-next'

const router = useRouter()
const { showAlert, showConfirm, showPrompt } = useDialog()

let initialUser = {}
try {
  initialUser = JSON.parse(localStorage.getItem('user') || '{}')
} catch (e) {
  console.error('Failed to parse user in Profile:', e)
}
const user = ref<any>(initialUser)

const currentCategory = ref('overview')
const isSidebarOpen = ref(false)

const categories = [
  { id: 'overview', name: '个人主页', icon: LayoutDashboard, eng: 'Dashboard' },
  { id: 'security', name: '安全中心', icon: ShieldCheck, eng: 'Security' },
  { id: 'research', name: '实验资产', icon: FlaskConical, eng: 'Research' },
  { id: 'history', name: '反应记录', icon: History, eng: 'Records' },
  { id: 'settings', name: '参数偏好', icon: Sliders, eng: 'Preferences' }
]

const userStats = computed(() => {
  const total = user.value.total_games || 0
  const wins = user.value.win_count || 0
  return {
    totalGames: total,
    winRate: total > 0 ? Math.round((wins / total) * 100) : 0
  }
})

const showChangePassword = ref(false)
const showChangeAvatar = ref(false)
const loading = ref(false)
const twoFactorLoading = ref(false)
const show2FASetup = ref(false)
const qrCode = ref('')
const showHardwareKeys = ref(false)
const showDeviceManagement = ref(false)

const fetchLatestUserInfo = async () => {
  try {
    const response = await authAPI.getUserInfo()
    user.value = response.data
    localStorage.setItem('user', JSON.stringify(response.data))
  } catch (error) {
    console.error('获取用户信息失败:', error)
  }
}

onMounted(fetchLatestUserInfo)

const handleUpdateAvatar = async (avatar: string) => {
  try {
    await authAPI.updateAvatar(avatar)
    user.value.avatar = avatar
    localStorage.setItem('user', JSON.stringify(user.value))
    showChangeAvatar.value = false
    showAlert('研究员标识已更新。', '变更成功')
  } catch (error: any) {
    showAlert(error.response?.data?.error || '更新标识失败', '错误')
  }
}

const handleUpdateNickname = async () => {
  const newNickname = await showPrompt('请输入新的研究员昵称:', user.value.nickname, '修改昵称')
  if (newNickname === null || newNickname === user.value.nickname) return

  try {
    await authAPI.updateNickname(newNickname)
    user.value.nickname = newNickname
    localStorage.setItem('user', JSON.stringify(user.value))
    showAlert('研究员昵称已成功同步。', '变更成功')
  } catch (error: any) {
    showAlert(error.response?.data?.error || '更新昵称失败', '错误')
  }
}

const handleLogout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  router.push('/login')
}

const handleSetup2FA = async () => {
  twoFactorLoading.value = true
  try {
    const response = await authAPI.setup2FA()
    qrCode.value = response.data.qr_code
    show2FASetup.value = true
  } catch (error: any) {
    showAlert(error.response?.data?.error || '获取2FA设置失败', '错误')
  } finally {
    twoFactorLoading.value = false
  }
}

const handleEnable2FA = async (code: string, password: string) => {
  if (!code || !password) return
  twoFactorLoading.value = true
  try {
    await authAPI.enable2FA(code, password)
    await showAlert('双重验证已成功开启', '成功')
    show2FASetup.value = false
    fetchLatestUserInfo()
  } catch (error: any) {
    showAlert(error.response?.data?.error || '开启2FA失败', '错误')
  } finally {
    twoFactorLoading.value = false
  }
}

const handleDisable2FA = async () => {
  const code = await showPrompt('为了安全起见，请输入 6 位验证码以停用双重验证：', '请输入验证码', '停用双重验证')
  if (!code) return
  
  twoFactorLoading.value = true
  try {
    await authAPI.disable2FA(code)
    await showAlert('双重验证已关闭', '系统提示')
    fetchLatestUserInfo()
  } catch (error: any) {
    showAlert(error.response?.data?.error || '关闭2FA失败', '错误')
  } finally {
    twoFactorLoading.value = false
  }
}

const handleChangePassword = async (oldPassword: string, newPassword: string, code: string) => {
  loading.value = true
  try {
    await authAPI.changePassword(oldPassword, newPassword, code)
    await showAlert('密码修改成功，请重新登录', '重置成功')
    handleLogout()
  } catch (error: any) {
    showAlert(error.response?.data?.error || '修改密码失败', '错误')
  } finally {
    loading.value = false
  }
}

const handleChangeAvatar = async (newAvatar: string) => {
  loading.value = true
  try {
    await authAPI.updateAvatar(newAvatar)
    const updatedUser = { ...user.value, avatar: newAvatar }
    localStorage.setItem('user', JSON.stringify(updatedUser))
    user.value = updatedUser
    await showAlert('头像更新成功！', '同步完成')
    showChangeAvatar.value = false
  } catch (error: any) {
    showAlert(error.response?.data?.error || '更新头像失败', '错误')
  } finally {
    loading.value = false
  }
}

const handleDeleteAccount = async () => {
  const confirm1 = await showConfirm('确定要注销账号吗？此操作无法恢复！', '⚠️ 警告')
  if (!confirm1) return
  const confirm2 = await showConfirm('再次确认：确定要删除账号吗？', '⚠️ 再次确认')
  if (!confirm2) return

  try {
    await authAPI.deleteAccount()
    await showAlert('账号已注销', '注销成功')
    handleLogout()
  } catch (error: any) {
    showAlert(error.response?.data?.error || '注销账号失败', '错误')
  }
}
</script>

<template>
  <div class="min-h-screen bg-slate-50 dark:bg-[#0a0a0c] text-slate-900 dark:text-white selection:bg-blue-500/30">
    <!-- Background Effects -->
    <div class="fixed inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-[-10%] right-[-10%] w-[50%] h-[50%] bg-blue-500/5 rounded-full blur-[120px]" />
      <div class="absolute bottom-[-10%] left-[-10%] w-[50%] h-[50%] bg-purple-500/5 rounded-full blur-[120px]" />
    </div>

    <!-- Mobile Sidebar overlay -->
    <div v-if="isSidebarOpen" @click="isSidebarOpen = false" class="fixed inset-0 bg-black/60 backdrop-blur-sm z-[60] lg:hidden" />

    <!-- Mobile Sidebar -->
    <aside 
      :class="[
        'fixed top-0 left-0 bottom-0 w-64 bg-white dark:bg-[#0d0d10] border-r border-slate-200 dark:border-white/5 z-[70] transition-transform duration-300 lg:hidden',
        isSidebarOpen ? 'translate-x-0' : '-translate-x-full'
      ]"
    >
      <div class="p-6 border-b border-slate-100 dark:border-white/5 flex items-center justify-between">
        <span class="font-black text-xs tracking-[0.2em] text-slate-400">RESEARCH_NAVIGATION</span>
        <button @click="isSidebarOpen = false" class="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-lg">
          <CloseIcon class="w-4 h-4" />
        </button>
      </div>
      <nav class="p-3 space-y-1">
        <button 
          v-for="cat in categories" 
          :key="cat.id" 
          @click="currentCategory = cat.id; isSidebarOpen = false"
          class="w-full flex items-center gap-3 px-4 py-3 rounded-2xl transition-all font-bold text-sm"
          :class="[
            currentCategory === cat.id 
              ? 'bg-blue-600/10 text-blue-600 dark:text-blue-400' 
              : 'text-slate-500 hover:bg-slate-100 dark:hover:bg-white/5'
          ]"
        >
          <component :is="cat.icon" class="w-4 h-4" />
          <span class="text-sm">{{ cat.name }}</span>
        </button>
      </nav>
    </aside>

    <div class="max-w-[1400px] mx-auto relative z-10 px-4 pt-10 pb-20 md:px-8">
      <!-- Desktop Header & Mobile Control -->
      <div class="mb-8 flex flex-col md:flex-row md:items-center justify-between gap-6">
        <div class="flex items-center gap-4">
          <button @click="router.push('/')" class="p-3 bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl hover:scale-105 transition-all text-slate-400 hover:text-slate-900 dark:hover:text-white">
            <ArrowLeft class="w-5 h-5" />
          </button>
          <div class="lg:hidden">
            <button @click="isSidebarOpen = true" class="flex items-center gap-2 px-4 py-3 bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl font-black text-xs uppercase tracking-widest">
              <Menu class="w-4 h-4" /> 导航
            </button>
          </div>
          <div class="hidden md:block">
            <h1 class="text-3xl font-black tracking-tighter uppercase italic">实验室档案 <span class="text-blue-500 font-mono text-sm not-italic ml-2">/ RESH_PROFILE_V2</span></h1>
          </div>
        </div>

        <!-- PC Top Navigation -->
        <nav class="hidden lg:flex items-center gap-1 p-1 bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-[1.5rem] backdrop-blur-xl shrink-0 overflow-hidden">
          <button 
            v-for="cat in categories" 
            :key="cat.id" 
            @click="currentCategory = cat.id"
            class="flex flex-col items-center justify-center min-w-[110px] py-3 px-6 rounded-2xl transition-all"
            :class="[
              currentCategory === cat.id 
                ? 'bg-blue-600/10 text-blue-600 dark:text-blue-400 font-bold' 
                : 'text-slate-400 hover:text-slate-600 dark:hover:text-white'
            ]"
          >
            <component :is="cat.icon" class="w-4 h-4" />
            <span class="text-[11px] font-black uppercase tracking-tight">{{ cat.name }}</span>
          </button>
        </nav>

        <router-link 
          to="/ranking" 
          class="flex items-center gap-2 px-6 py-3 bg-amber-500/10 border border-amber-500/20 rounded-2xl text-amber-500 hover:bg-amber-500/20 transition-all font-black text-xs uppercase tracking-widest"
        >
          <Trophy class="w-4 h-4" />
          全球排名
        </router-link>
      </div>

      <div class="flex flex-col lg:flex-row gap-8 items-start">
        <!-- Persistent Profile Sidebar (Stats & Header) -->
        <div class="w-full lg:w-[360px] space-y-6 shrink-0 lg:sticky lg:top-8">
          <ProfileHeader 
            :user="user" 
            @change-avatar="showChangeAvatar = true" 
            @change-nickname="handleUpdateNickname"
          />
          <StatsGrid :stats="userStats" />
        </div>

        <!-- Dynamic Content Area -->
        <div class="flex-1 w-full space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
          
          <div v-if="currentCategory === 'overview'" class="space-y-8">
            <!-- Achievements Section -->
            <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2.5rem] p-10 shadow-sm">
              <h3 class="text-xl font-bold uppercase tracking-widest mb-6 flex items-center gap-3 text-slate-400">
                <Award class="w-5 h-5 transition-colors group-hover:text-blue-500" />
                实验室成就 / Achievements
              </h3>
              <div class="flex flex-col items-center justify-center py-20 border-2 border-dashed border-slate-200 dark:border-white/5 rounded-[2rem] bg-slate-50 dark:bg-white/[0.02]">
                <Shield class="w-12 h-12 text-slate-300 dark:text-slate-700 mb-4 opacity-30" />
                <p class="text-slate-400 font-medium italic text-sm">尚未获得勋章记录。去开启一场化学反应吧！</p>
              </div>
            </div>
          </div>

          <div v-if="currentCategory === 'security'" class="space-y-6">
            <SecurityPanel 
              :two-factor-enabled="user.two_factor_enabled"
              :two-factor-loading="twoFactorLoading"
              @change-password="showChangePassword = true"
              @setup2fa="handleSetup2FA"
              @disable2fa="handleDisable2FA"
              @manage-hardware-keys="showHardwareKeys = true"
              @manage-devices="showDeviceManagement = true"
              @delete-account="handleDeleteAccount"
            />
          </div>

          <div v-if="currentCategory === 'research'">
            <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2.5rem] p-10 shadow-sm">
              <CustomDecks />
            </div>
          </div>

          <div v-if="currentCategory === 'history'" class="space-y-6">
            <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2.5rem] p-10 shadow-sm">
              <MatchHistory />
            </div>
          </div>

          <div v-if="currentCategory === 'settings'" class="space-y-6">
            <!-- Visual Settings Section -->
            <SettingsPanel />

            <!-- Feedback Section -->
            <router-link 
              to="/feedbacks/my"
              class="group flex items-center justify-between p-8 bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2.5rem] shadow-sm hover:shadow-lg transition-all"
            >
              <div class="flex items-center gap-6">
                <div class="w-14 h-14 bg-blue-500/10 rounded-2xl flex items-center justify-center text-blue-500 group-hover:bg-blue-500 group-hover:text-white transition-all duration-300 group-hover:rotate-6">
                  <MessageSquare class="w-6 h-6" />
                </div>
                <div>
                  <h3 class="text-lg font-bold uppercase tracking-widest text-slate-800 dark:text-white">反馈与消息 / Feedback</h3>
                  <p class="text-slate-500 dark:text-slate-400 mt-1 font-medium text-xs">查看提交的建议、错误报告及管理员回复</p>
                </div>
              </div>
              <div class="w-10 h-10 rounded-full border border-slate-200 dark:border-white/5 flex items-center justify-center text-slate-400 group-hover:bg-blue-500 group-hover:text-white group-hover:translate-x-1 transition-all">
                <ArrowLeft class="w-4 h-4 rotate-180" />
              </div>
            </router-link>
          </div>

        </div>
      </div>
    </div>

    <!-- Modals -->
    <ChangeAvatarModal 
      :show="showChangeAvatar"
      :current-avatar="user.avatar"
      :loading="loading"
      @close="showChangeAvatar = false"
      @save="handleChangeAvatar"
    />

    <ChangePasswordModal 
      :show="showChangePassword"
      :loading="loading"
      :is2fa-enabled="user.two_factor_enabled"
      @close="showChangePassword = false"
      @save="handleChangePassword"
      @success="showAlert('凭证已通过硬件加密协议更新，请重新登录。', '同步完成'); handleLogout()"
    />

    <TwoFactorSetupModal 
      :show="show2FASetup"
      :qr-code="qrCode"
      :loading="twoFactorLoading"
      @close="show2FASetup = false"
      @enable="handleEnable2FA"
    />

    <HardwareKeyModal 
      :show="showHardwareKeys"
      @close="showHardwareKeys = false"
    />

    <DeviceManagementModal
      :show="showDeviceManagement"
      @close="showDeviceManagement = false"
    />
  </div>
</template>
