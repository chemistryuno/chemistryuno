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
import PersonalSpacePanel from '../components/profile/PersonalSpacePanel.vue'
import MatchHistory from '../components/profile/MatchHistory.vue'
import ChangeAvatarModal from '../components/profile/ChangeAvatarModal.vue'
import ChangePasswordModal from '../components/profile/ChangePasswordModal.vue'
import TwoFactorSetupModal from '../components/profile/TwoFactorSetupModal.vue'
import HardwareKeyModal from '../components/profile/HardwareKeyModal.vue'
import DeviceManagementModal from '../components/profile/DeviceManagementModal.vue'
import ChangeEmailModal from '../components/profile/ChangeEmailModal.vue'
import SetEmailModal from '../components/profile/SetEmailModal.vue'
import LevelProgress from '../components/LevelProgress.vue'
import { LayoutDashboard, ShieldCheck, FlaskConical, History, Sliders, Menu, X as CloseIcon, LogOut, User as UserIcon } from 'lucide-vue-next'

const router = useRouter()
const { showAlert, showConfirm, showPrompt } = useDialog()

let initialUser: any = {}
try {
  initialUser = JSON.parse(localStorage.getItem('user') || '{}')
  // 兼容旧版本的 id 字段
  if (initialUser.id && !initialUser.uid) {
    initialUser.uid = initialUser.id
  }
} catch (e) {
  console.error('Failed to parse user in Profile:', e)
}
const user = ref<any>(initialUser)

const currentCategory = ref('overview')
const isSidebarOpen = ref(false)

const categories = [
  { id: 'overview', name: '数据概览', icon: LayoutDashboard, eng: 'Dashboard' },
  { id: 'space', name: '个人空间', icon: UserIcon, eng: 'Space' },
  { id: 'security', name: '安全中心', icon: ShieldCheck, eng: 'Security' },
  { id: 'research', name: '实验资产', icon: FlaskConical, eng: 'Research' },
  { id: 'history', name: '反应记录', icon: History, eng: 'Records' },
  { id: 'settings', name: '外观偏好', icon: Sliders, eng: 'Preferences' }
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
const showChangeEmail = ref(false)
const showSetEmail = ref(false)
const smtpEnabled = ref(false)

const fetchLatestUserInfo = async () => {
  try {
    const response = await authAPI.getUserInfo()
    user.value = response.data
    localStorage.setItem('user', JSON.stringify(response.data))
  } catch (error) {
    console.error('获取用户信息失败:', error)
  }
}

onMounted(async () => {
  fetchLatestUserInfo()
  try {
    const res = await authAPI.getAuthConfig()
    smtpEnabled.value = res.data.smtp_enabled
  } catch (error) {
    console.error('获取配置失败:', error)
  }
})

const handleUpdateNickname = async () => {
  const newNickname = await showPrompt('请输入新的研究员昵称:', user.value.nickname, '修改昵称')
  if (newNickname === null || newNickname === user.value.nickname) return

  try {
    await authAPI.updateProfile({ nickname: newNickname })
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

const handleChangePassword = async (oldPassword: string, newPassword: string, code: string, useEmail: boolean = false) => {
  loading.value = true
  try {
    await authAPI.changePassword(oldPassword, newPassword, code, useEmail)
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
    // 发送注销验证码
    await authAPI.sendCode(user.value.email, 'delete_account')
    const code = await showPrompt('验证码已发送至您的通讯邮箱，请输入以授权注销操作', '档案注销授权', '安全验证')
    if (!code) return

    await authAPI.deleteAccount(code)
    await showAlert('您的研究员档案已被彻底移除。', '注销成功')
    handleLogout()
  } catch (error: any) {
    showAlert(error.response?.data?.error || '注销流程中断，请稍后重试', '操作受阻')
  }
}

// OAuth 绑定与解绑逻辑
const handleOAuthBind = (provider: 'github' | 'ms' | 'google' | 'apple') => {
  const width = 600
  const height = 700
  const left = window.screen.width / 2 - width / 2
  const top = window.screen.height / 2 - height / 2
  
  const token = localStorage.getItem('token') || ''
  const baseUrl = import.meta.env.VITE_API_BASE_URL || '/api'
  const url = `${baseUrl}/auth/${provider}/bind?token=${token}`
  
  const popup = window.open(url, 'OAuth Bind', `width=${width},height=${height},left=${left},top=${top}`)
  
  if (!popup) {
    showAlert('弹出窗口被拦截，请允许弹出窗口后重试。', '拦截提示')
    return
  }

  const messageHandler = (event: MessageEvent) => {
    if (event.data.type === 'oauth-bind-success') {
      window.removeEventListener('message', messageHandler)
      showAlert('同步账号绑定成功！', '绑定完成')
      fetchLatestUserInfo()
    } else if (event.data.type === 'oauth-error') {
      window.removeEventListener('message', messageHandler)
      showAlert(event.data.error || '绑定失败', '错误')
    }
  }
  
  window.addEventListener('message', messageHandler)
}

const handleOAuthUnbind = async (provider: 'github' | 'ms' | 'google' | 'apple') => {
  const providerNames = { github: 'GitHub', ms: 'Microsoft', google: 'Google', apple: 'Apple' }
  const confirmed = await showConfirm(`确定要解除 ${providerNames[provider]} 的账号绑定吗？`, '解绑确认')
  if (!confirmed) return

  try {
    await authAPI.unbindOAuth(provider)
    showAlert('账号解绑成功', '提示')
    fetchLatestUserInfo()
  } catch (error: any) {
    showAlert(error.response?.data?.error || '解绑失败', '错误')
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

        <div class="pt-4 mt-4 border-t border-slate-100 dark:border-white/5">
          <button 
            @click="handleLogout"
            class="w-full flex items-center gap-3 px-4 py-3 rounded-2xl transition-all font-bold text-sm text-red-500 hover:bg-red-500/10"
          >
            <LogOut class="w-4 h-4" />
            <span>退出登录 / Logout</span>
          </button>
        </div>
      </nav>
    </aside>

    <div class="max-w-[1400px] mx-auto relative z-10 px-4 pt-6 pb-12 md:px-6">
      <!-- Desktop Header & Mobile Control -->
      <div class="mb-6 flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div class="flex items-center gap-3">
          <button @click="router.push('/')" class="p-2.5 bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl hover:scale-105 transition-all text-slate-400 hover:text-slate-900 dark:hover:text-white">
            <ArrowLeft class="w-4 h-4" />
          </button>
          <div class="lg:hidden">
            <button @click="isSidebarOpen = true" class="flex items-center gap-2 px-3 py-2 bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl font-black text-[10px] uppercase tracking-widest text-slate-500">
              <Menu class="w-3.5 h-3.5" /> NAV
            </button>
          </div>
          <div class="hidden md:block">
            <h1 class="text-2xl font-black tracking-tighter uppercase italic text-slate-800 dark:text-white">实验室档案 <span class="text-blue-500 font-mono text-[10px] not-italic ml-2 opacity-50">/ RESH_PROFILE_V2</span></h1>
          </div>
        </div>

        <!-- PC Top Navigation -->
        <nav class="hidden lg:flex items-center gap-1 p-1 bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl backdrop-blur-xl shrink-0 overflow-hidden">
          <button 
            v-for="cat in categories" 
            :key="cat.id" 
            @click="currentCategory = cat.id"
            class="flex flex-col items-center justify-center min-w-[90px] py-2 px-4 rounded-xl transition-all"
            :class="[
              currentCategory === cat.id 
                ? 'bg-blue-600/10 text-blue-600 dark:text-blue-400 font-bold' 
                : 'text-slate-400 hover:text-slate-600 dark:hover:text-white'
            ]"
          >
            <component :is="cat.icon" class="w-3.5 h-3.5" />
            <span class="text-[9px] font-black uppercase tracking-tight mt-0.5">{{ cat.name }}</span>
          </button>
        </nav>

        <router-link 
          to="/ranking" 
          class="flex items-center gap-2 px-4 py-2 bg-amber-500/10 border border-amber-500/20 rounded-xl text-amber-500 hover:bg-amber-500/20 transition-all font-black text-[10px] uppercase tracking-widest"
        >
          <Trophy class="w-3.5 h-3.5" />
          排名中心
        </router-link>
      </div>

      <div class="flex flex-col lg:flex-row gap-6 items-start">
        <!-- Persistent Profile Sidebar (Stats & Header) -->
        <div class="w-full lg:w-[320px] space-y-5 shrink-0 lg:sticky lg:top-6">
          <ProfileHeader 
            :user="user" 
            @change-avatar="showChangeAvatar = true" 
            @change-nickname="handleUpdateNickname"
          />
          <StatsGrid :stats="userStats" />
        </div>

        <!-- Dynamic Content Area -->
        <div class="flex-1 w-full space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">

          <div v-if="currentCategory === 'overview'" class="space-y-6">
            <!-- Level Progress Section -->
            <LevelProgress />

            <!-- Achievements Section -->
            <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-2xl p-8 shadow-sm">
              <h3 class="text-base font-black uppercase tracking-widest mb-5 flex items-center gap-2 text-slate-400">
                <Award class="w-4 h-4 transition-colors group-hover:text-blue-500" />
                实验室成就 <span class="text-[10px] font-mono opacity-30">/ ACHIEVEMENTS</span>
              </h3>
              <div class="flex flex-col items-center justify-center py-16 border-2 border-dashed border-slate-200 dark:border-white/5 rounded-2xl bg-slate-50 dark:bg-white/[0.02]">
                <Shield class="w-10 h-10 text-slate-300 dark:text-slate-700 mb-3 opacity-30" />
                <p class="text-slate-400 font-medium italic text-xs uppercase tracking-widest">No_Archive_Found</p>
              </div>
            </div>
          </div>

          <div v-if="currentCategory === 'space'" class="space-y-6">
            <PersonalSpacePanel :user="user" @update="fetchLatestUserInfo" />
          </div>

          <div v-if="currentCategory === 'security'" class="space-y-6">
            <SecurityPanel
              :two-factor-enabled="user.two_factor_enabled"
              :two-factor-loading="twoFactorLoading"
              :smtp-enabled="smtpEnabled"
              :has-email="!!user.email"
              :github-id="user.github_id"
              :microsoft-id="user.microsoft_id"
              :google-id="user.google_id"
              :apple-id="user.apple_id"
              @change-password="showChangePassword = true"
              @change-email="showChangeEmail = true"
              @set-email="showSetEmail = true"
              @setup2fa="handleSetup2FA"
              @disable2fa="handleDisable2FA"
              @manage-hardware-keys="showHardwareKeys = true"
              @manage-devices="showDeviceManagement = true"
              @delete-account="handleDeleteAccount"
              @bind-github="handleOAuthBind('github')"
              @bind-microsoft="handleOAuthBind('ms')"
              @bind-google="handleOAuthBind('google')"
              @bind-apple="handleOAuthBind('apple')"
              @unbind-github="handleOAuthUnbind('github')"
              @unbind-microsoft="handleOAuthUnbind('ms')"
              @unbind-google="handleOAuthUnbind('google')"
              @unbind-apple="handleOAuthUnbind('apple')"
            />
          </div>

          <div v-if="currentCategory === 'research'">
            <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-2xl p-8 shadow-sm">
              <CustomDecks />
            </div>
          </div>

          <div v-if="currentCategory === 'history'" class="space-y-6">
            <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-2xl p-8 shadow-sm">
              <MatchHistory />
            </div>
          </div>

          <div v-if="currentCategory === 'settings'" class="space-y-6">
            <!-- Visual Settings Section -->
            <SettingsPanel :user="user" @update="fetchLatestUserInfo" />

            <!-- Feedback Section -->
            <router-link 
              to="/feedbacks/my"
              class="group flex items-center justify-between p-6 bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-2xl shadow-sm hover:shadow-lg transition-all"
            >
              <div class="flex items-center gap-5">
                <div class="w-12 h-12 bg-blue-500/10 rounded-xl flex items-center justify-center text-blue-500 group-hover:bg-blue-500 group-hover:text-white transition-all duration-300">
                  <MessageSquare class="w-5 h-5" />
                </div>
                <div>
                  <h3 class="text-base font-black uppercase tracking-widest text-slate-800 dark:text-white leading-none">反馈与消息 / Feedback</h3>
                  <p class="text-slate-500 dark:text-slate-400 mt-1.5 font-medium text-[11px]">查看提交的建议、错误报告及管理员回复</p>
                </div>
              </div>
              <div class="w-8 h-8 rounded-full border border-slate-200 dark:border-white/5 flex items-center justify-center text-slate-400 group-hover:bg-blue-500 group-hover:text-white group-hover:translate-x-1 transition-all">
                <ArrowLeft class="w-3.5 h-3.5 rotate-180" />
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
      :user-email="user.email"
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
    <ChangeEmailModal
      :show="showChangeEmail"
      :current-email="user.email"
      @close="showChangeEmail = false"
      @success="(newEmail) => { user.email = newEmail }"
    />
    <SetEmailModal
      :show="showSetEmail"
      @close="showSetEmail = false"
      @success="(newEmail) => { user.email = newEmail; fetchLatestUserInfo() }"
    />
  </div>
</template>
