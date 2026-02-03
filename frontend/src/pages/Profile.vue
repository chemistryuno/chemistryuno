<script setup lang="ts">
import { ref, onMounted } from 'vue'
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

const router = useRouter()
const { showAlert, showConfirm, showPrompt } = useDialog()
const user = ref<any>(JSON.parse(localStorage.getItem('user') || '{}'))

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
  <div class="min-h-screen bg-slate-50 dark:bg-[#0a0a0c] text-slate-900 dark:text-white p-4 md:p-8 selection:bg-blue-500/30">
    <!-- Background Effects -->
    <div class="fixed inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-[-10%] right-[-10%] w-[50%] h-[50%] bg-blue-500/5 rounded-full blur-[120px]" />
      <div class="absolute bottom-[-10%] left-[-10%] w-[50%] h-[50%] bg-purple-500/5 rounded-full blur-[120px]" />
      <div class="absolute inset-0 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-20 brightness-50 contrast-150" />
    </div>

    <div class="max-w-6xl mx-auto relative z-10">
      <!-- Back Button -->
      <div class="mb-10 flex items-center justify-between">
        <button 
          @click="router.push('/')" 
          class="group flex items-center gap-3 text-slate-400 hover:text-slate-900 dark:hover:text-white transition-all px-4 py-2 rounded-xl hover:bg-white dark:hover:bg-white/5 border border-transparent hover:border-slate-200 dark:hover:border-white/10"
        >
          <ArrowLeft class="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
          <span class="font-bold tracking-wider uppercase text-xs">返回指挥大厅</span>
        </button>

        <div class="flex items-center gap-3">
          <router-link 
            to="/ranking" 
            class="flex items-center gap-2 px-4 py-2 bg-amber-500/10 border border-amber-500/20 rounded-xl text-amber-500 hover:bg-amber-500/20 transition-all group"
          >
            <Trophy class="w-4 h-4 group-hover:scale-110 transition-transform" />
            <span class="text-[10px] font-black uppercase tracking-widest">全球排名</span>
          </router-link>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-12 gap-8">
        <!-- Sidebar: Header and Stats -->
        <div class="lg:col-span-4 space-y-6">
          <ProfileHeader 
            :user="user" 
            @change-avatar="showChangeAvatar = true" 
          />
          
          <StatsGrid />
        </div>

        <!-- Main Content: Security and Achievements -->
        <div class="lg:col-span-8 space-y-8">
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

          <!-- Visual Settings Section -->
          <SettingsPanel />

          <!-- Feedback Section -->
          <router-link 
            to="/feedbacks/my"
            class="group flex items-center justify-between p-8 bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2.5rem] shadow-sm dark:shadow-none transition-all hover:shadow-lg"
          >
            <div class="flex items-center gap-6">
              <div class="w-16 h-16 bg-blue-500/10 rounded-2xl flex items-center justify-center text-blue-500 group-hover:bg-blue-500 group-hover:text-white transition-all duration-500 group-hover:rotate-6">
                <MessageSquare class="w-8 h-8" />
              </div>
              <div>
                <h3 class="text-xl font-bold uppercase tracking-widest text-slate-800 dark:text-white">反馈与消息 / Feedback</h3>
                <p class="text-slate-500 dark:text-slate-400 mt-1 font-medium">查看提交的建议、错误报告及管理员回复</p>
              </div>
            </div>
            <div class="w-12 h-12 rounded-full border border-slate-200 dark:border-white/5 flex items-center justify-center text-slate-400 group-hover:bg-white dark:group-hover:bg-white/5 group-hover:translate-x-2 transition-all">
              <ArrowLeft class="w-5 h-5 rotate-180" />
            </div>
          </router-link>

          <!-- Custom Decks Section -->
          <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2.5rem] p-10 shadow-sm dark:shadow-none transition-all hover:shadow-lg">
            <CustomDecks />
          </div>

          <!-- Match History Section -->
          <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2.5rem] p-10 shadow-sm dark:shadow-none transition-all hover:shadow-lg">
            <MatchHistory />
          </div>

          <!-- Achievements Section -->
          <div class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2.5rem] p-10 shadow-sm dark:shadow-none transition-all hover:shadow-lg">
            <h3 class="text-xl font-bold uppercase tracking-widest mb-6 flex items-center gap-3 text-slate-400">
              <Award class="w-5 h-5" />
              实验室成就 / Achievements
            </h3>
            <div class="flex flex-col items-center justify-center py-20 border-2 border-dashed border-slate-200 dark:border-white/5 rounded-[2rem] bg-slate-50 dark:bg-white/[0.02]">
              <Shield class="w-12 h-12 text-slate-300 dark:text-slate-700 mb-4 opacity-50" />
              <p class="text-slate-400 font-medium italic">尚未获得勋章记录。去开启一场化学反应吧！</p>
            </div>
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
