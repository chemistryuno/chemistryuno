<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../utils/api'
import { 
  ArrowLeft, 
  Key, 
  UserX, 
  Shield, 
  Loader2, 
  Award, 
  RefreshCw,
  Fingerprint,
  Activity,
  Zap,
  Lock,
  Eye,
  EyeOff,
  Upload,
  Calendar,
  User as UserIcon
} from 'lucide-vue-next'
import { cn } from '../utils/cn'

const router = useRouter()
const user = ref<any>(JSON.parse(localStorage.getItem('user') || '{}'))

const showChangePassword = ref(false)
const showChangeAvatar = ref(false)
const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const showPasswords = ref(false)
const selectedAvatar = ref(user.value.avatar)
const loading = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

const avatarOptions = ["🧪", "🧬", "⚗️", "🔬", "🛰️", "🚀", "🪐", "⚛️", "📡", "🧠", "🦾", "👾"]

const fetchLatestUserInfo = async () => {
  try {
    const response = await authAPI.getUserInfo()
    user.value = response.data
    localStorage.setItem('user', JSON.stringify(response.data))
    selectedAvatar.value = response.data.avatar
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

const handleChangePassword = async () => {
  if (newPassword.value !== confirmPassword.value) {
    alert('两次输入的密码不一致')
    return
  }
  loading.value = true
  try {
    await authAPI.changePassword(oldPassword.value, newPassword.value)
    alert('密码修改成功，请重新登录')
    handleLogout()
  } catch (error: any) {
    alert(error.response?.data?.error || '修改密码失败')
  } finally {
    loading.value = false
  }
}

const handleFileUpload = (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  if (file.size > 2 * 1024 * 1024) {
    alert('头像文件不能超过 2MB')
    return
  }

  const reader = new FileReader()
  reader.onload = (e) => {
    selectedAvatar.value = e.target?.result as string
  }
  reader.readAsDataURL(file)
}

const handleChangeAvatar = async () => {
  loading.value = true
  try {
    await authAPI.updateAvatar(selectedAvatar.value)
    const updatedUser = { ...user.value, avatar: selectedAvatar.value }
    localStorage.setItem('user', JSON.stringify(updatedUser))
    user.value = updatedUser
    alert('头像更新成功！')
    showChangeAvatar.value = false
  } catch (error: any) {
    alert(error.response?.data?.error || '更新头像失败')
  } finally {
    loading.value = false
  }
}

const handleDeleteAccount = async () => {
  if (!window.confirm('确定要注销账号吗？此操作无法恢复！')) return
  if (!window.confirm('再次确认：确定要删除账号吗？')) return

  try {
    await authAPI.deleteAccount()
    alert('账号已注销')
    handleLogout()
  } catch (error: any) {
    alert(error.response?.data?.error || '注销账号失败')
  }
}
</script>

<template>
  <div class="min-h-screen bg-[#0a0a0c] text-white p-4 md:p-8 selection:bg-blue-500/30">
    <div class="fixed inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-[-10%] right-[-10%] w-[50%] h-[50%] bg-blue-500/5 rounded-full blur-[120px]" />
      <div class="absolute bottom-[-10%] left-[-10%] w-[50%] h-[50%] bg-purple-500/5 rounded-full blur-[120px]" />
      <div class="absolute inset-0 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-20 brightness-50 contrast-150" />
    </div>

    <div class="max-w-5xl mx-auto relative z-10">
      <button 
        @click="router.push('/')" 
        class="group flex items-center gap-3 text-slate-400 hover:text-white mb-10 transition-all px-4 py-2 rounded-full hover:bg-white/5 border border-transparent hover:border-white/10"
      >
        <ArrowLeft class="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
        <span class="font-bold tracking-wider uppercase text-xs">返回指挥大厅 / Back to Hub</span>
      </button>

      <div class="grid grid-cols-1 lg:grid-cols-12 gap-8">
        <div class="lg:col-span-4 space-y-6">
          <div class="bg-[#111114] border border-white/10 rounded-[2.5rem] p-8 relative overflow-hidden group shadow-2xl">
            <div class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-transparent via-blue-500/50 to-transparent" />
            
            <div class="flex flex-col items-center">
              <div class="relative group/avatar mb-8">
                <div class="w-40 h-40 bg-gradient-to-tr from-[#1a1c1e] to-[#2d3035] rounded-[3rem] p-1 shadow-2xl transition-transform duration-500 group-hover/avatar:scale-105">
                  <div class="w-full h-full bg-[#111114] rounded-[2.8rem] flex items-center justify-center text-7xl relative overflow-hidden group/inner transition-all border border-white/5">
                    <div class="absolute inset-0 bg-blue-500/5 opacity-0 group-hover/inner:opacity-100 transition-opacity" />
                    <template v-if="user.avatar && user.avatar.startsWith('data:')">
                       <img :src="user.avatar" class="w-full h-full object-cover relative z-10" />
                    </template>
                    <template v-else>
                       <span class="relative z-10 scale-110 drop-shadow-[0_0_15px_rgba(255,255,255,0.3)]">{{ user.avatar || '🧪' }}</span>
                    </template>
                  </div>
                </div>
                
                <button 
                  @click="showChangeAvatar = true"
                  class="absolute -bottom-2 -right-2 bg-blue-600 hover:bg-blue-500 p-3 rounded-2xl shadow-[0_0_20px_rgba(37,99,235,0.4)] z-20 group-hover:rotate-12 transition-all active:scale-95"
                  title="更改研究员原型"
                >
                  <RefreshCw class="w-5 h-5 text-white" />
                </button>
              </div>

              <div class="text-center space-y-2 w-full">
                <div class="flex items-center justify-center gap-2 mb-1">
                  <UserIcon class="w-4 h-4 text-blue-500 opacity-50" />
                  <span class="text-[10px] font-mono text-slate-500 uppercase tracking-widest">Researcher ID</span>
                </div>
                <h2 class="text-3xl font-black tracking-tight text-white group-hover:text-blue-400 transition-colors uppercase italic truncate px-4">
                  {{ user.username }}
                </h2>
                <div class="flex items-center justify-center gap-2 pt-2">
                  <span v-if="user.is_admin" class="bg-blue-500/10 text-blue-400 text-[10px] font-black px-4 py-1.5 rounded-full border border-blue-500/20 flex items-center gap-2 uppercase tracking-[0.2em]">
                    <Shield class="w-3 h-3" /> 首席研究员 / CORE ADM
                  </span>
                  <span v-else class="bg-slate-500/10 text-slate-400 text-[10px] font-black px-4 py-1.5 rounded-full border border-slate-500/20 flex items-center gap-2 uppercase tracking-[0.2em]">
                    <Fingerprint class="w-3 h-3" /> 各级研究员 / RESEARCHER
                  </span>
                </div>
              </div>

              <div class="w-full mt-10 pt-10 border-t border-white/5 space-y-4">
                <div class="flex justify-between items-center text-xs">
                  <span class="text-slate-500 font-bold uppercase tracking-widest flex items-center gap-2"><Fingerprint class="w-3 h-3" /> System UID</span>
                  <span class="font-mono text-blue-400/80">{{ user.uid }}</span>
                </div>
                <div v-if="user.created_at" class="flex justify-between items-center text-xs">
                  <span class="text-slate-500 font-bold uppercase tracking-widest flex items-center gap-2"><Calendar class="w-3 h-3" /> Joined Date</span>
                  <span class="font-mono text-slate-400">{{ new Date(user.created_at).toLocaleDateString() }}</span>
                </div>
                <div class="flex justify-between items-center text-xs">
                  <span class="text-slate-500 font-bold uppercase tracking-widest flex items-center gap-2"><Award class="w-3 h-3" /> Exp Level</span>
                  <div class="flex items-center gap-2">
                    <div class="w-24 h-1.5 bg-white/5 rounded-full overflow-hidden">
                      <div class="w-1/3 h-full bg-blue-500 shadow-[0_0_10px_rgba(59,130,246,0.5)]" />
                    </div>
                    <span class="text-blue-500 font-black">LV.01</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="bg-white/5 border border-white/5 rounded-3xl p-5 hover:bg-white/[0.08] transition-colors">
              <div class="flex items-center gap-3 mb-2">
                <Zap class="w-4 h-4 text-yellow-400" />
                <span class="text-[10px] font-black uppercase text-slate-500">总场次</span>
              </div>
              <div class="text-2xl font-black">--</div>
            </div>
            <div class="bg-white/5 border border-white/5 rounded-3xl p-5 hover:bg-white/[0.08] transition-colors">
              <div class="flex items-center gap-3 mb-2">
                <Activity class="w-4 h-4 text-emerald-400" />
                <span class="text-[10px] font-black uppercase text-slate-500">胜率</span>
              </div>
              <div class="text-2xl font-black">--%</div>
            </div>
          </div>
        </div>

        <div class="lg:col-span-8 space-y-8">
          <div class="bg-[#111114] border border-white/10 rounded-[2.5rem] p-10 relative overflow-hidden">
            <div class="flex items-center justify-between mb-10">
              <div>
                <h3 class="text-2xl font-black uppercase italic tracking-tighter flex items-center gap-3">
                  <span class="w-2 h-8 bg-blue-500 rounded-full" />
                  账户安全管理 / Security
                </h3>
                <p class="text-slate-500 text-sm mt-1">维护您的研究员凭证与实验室访问权限</p>
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <button 
                @click="showChangePassword = true"
                class="group relative flex flex-col items-start p-6 bg-white/5 hover:bg-blue-500/10 border border-white/5 hover:border-blue-500/30 rounded-3xl transition-all text-left"
              >
                <div class="bg-blue-500/20 p-3 rounded-2xl mb-4 group-hover:scale-110 transition-transform">
                  <Lock class="w-6 h-6 text-blue-400" />
                </div>
                <span class="text-lg font-bold">修改研究密码</span>
                <span class="text-slate-500 text-xs mt-1">更新安全凭证以确保实验室数据安全</span>
              </button>

              <button 
                @click="handleDeleteAccount"
                class="group relative flex flex-col items-start p-6 bg-white/5 hover:bg-red-500/10 border border-white/5 hover:border-red-500/30 rounded-3xl transition-all text-left"
              >
                <div class="bg-red-500/20 p-3 rounded-2xl mb-4 group-hover:scale-110 transition-transform">
                  <UserX class="w-6 h-6 text-red-500" />
                </div>
                <span class="text-lg font-bold text-red-400">注销席位</span>
                <span class="text-slate-500 text-xs mt-1">永久注销账户并清除所有研究数据</span>
              </button>
            </div>
          </div>

          <div class="bg-[#111114] border border-white/10 rounded-[2.5rem] p-10">
            <h3 class="text-xl font-bold uppercase tracking-widest mb-6 flex items-center gap-3 text-slate-400">
              <Award class="w-5 h-5" />
              实验室成就 / Achievements
            </h3>
            <div class="flex flex-col items-center justify-center py-20 border-2 border-dashed border-white/5 rounded-[2rem] bg-white/[0.02]">
              <Shield class="w-12 h-12 text-slate-700 mb-4" />
              <p class="text-slate-500 font-medium italic">尚未获得勋章记录。去开启一场化学反应吧！</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Change Avatar Modal -->
    <div v-if="showChangeAvatar" class="fixed inset-0 z-[100] flex items-center justify-center p-4 backdrop-blur-xl bg-black/60">
      <div class="bg-[#111114] border border-white/10 rounded-[3rem] p-10 max-w-xl w-full shadow-2xl relative animate-in fade-in zoom-in duration-300">
        <h3 class="text-2xl font-black mb-8 italic uppercase text-center">选择新的身份标识 / Select Avatar</h3>
        
        <div class="flex flex-col items-center gap-8 mb-10">
          <div class="relative group/preview">
            <div class="w-32 h-32 bg-[#1a1c1e] rounded-[2.5rem] border-2 border-dashed border-white/10 flex items-center justify-center overflow-hidden transition-all group-hover/preview:border-blue-500/50">
               <template v-if="selectedAvatar && selectedAvatar.startsWith('data:')">
                  <img :src="selectedAvatar" class="w-full h-full object-cover" />
               </template>
               <template v-else>
                  <span class="text-6xl">{{ selectedAvatar || '🧪' }}</span>
               </template>
            </div>
            <button 
              @click="fileInput?.click()"
              class="absolute -bottom-2 -right-2 bg-blue-600 p-2 rounded-xl text-white shadow-lg shadow-blue-500/20 hover:scale-110 transition-transform"
            >
              <Upload class="w-4 h-4" />
            </button>
            <input 
              type="file" 
              ref="fileInput" 
              class="hidden" 
              accept="image/*"
              @change="handleFileUpload"
            />
          </div>

          <div class="w-full">
            <p class="text-[10px] font-black uppercase text-slate-500 tracking-[0.2em] mb-4 text-center">快捷原型选择器 / Quick Lab Presets</p>
            <div class="grid grid-cols-4 sm:grid-cols-6 gap-4">
              <button
                v-for="emoji in avatarOptions"
                :key="emoji"
                @click="selectedAvatar = emoji"
                :class="cn(
                  'w-16 h-16 text-3xl flex items-center justify-center rounded-[1.5rem] transition-all duration-300 border-2',
                  selectedAvatar === emoji 
                    ? 'bg-blue-600 border-blue-400 scale-110 shadow-[0_0_20px_rgba(59,130,246,0.5)]' 
                    : 'bg-white/5 border-transparent hover:border-white/20 hover:scale-105'
                )"
              >
                {{ emoji }}
              </button>
            </div>
          </div>
          
          <div class="bg-blue-500/5 border border-blue-500/10 rounded-2xl p-4 w-full flex items-center gap-4">
            <div class="p-2 bg-blue-500/10 rounded-xl text-blue-400">
              <Upload class="w-4 h-4" />
            </div>
            <div class="flex flex-col">
              <span class="text-xs font-bold text-white">本地图像上传协议 (MAX 2MB)</span>
              <span class="text-[10px] text-slate-500">支持 JPG, PNG, WEBP 等格式</span>
            </div>
          </div>
        </div>

        <div class="flex gap-4">
          <button 
            @click="showChangeAvatar = false"
            class="flex-1 py-4 bg-white/5 hover:bg-white/10 rounded-2xl font-bold transition-all text-slate-400"
          >
            取消
          </button>
          <button 
            @click="handleChangeAvatar"
            :disabled="loading"
            class="flex-1 py-4 bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 rounded-2xl font-black text-white shadow-xl shadow-blue-500/20 disabled:opacity-50 flex items-center justify-center gap-2"
          >
            <Loader2 v-if="loading" class="w-5 h-5 animate-spin" />
            同步身份更改
          </button>
        </div>
      </div>
    </div>

    <!-- Change Password Modal -->
    <div v-if="showChangePassword" class="fixed inset-0 z-[100] flex items-center justify-center p-4 backdrop-blur-xl bg-black/60">
      <div class="bg-[#111114] border border-white/10 rounded-[3rem] p-10 max-w-md w-full shadow-2xl relative animate-in fade-in zoom-in duration-300">
        <h3 class="text-2xl font-black mb-8 italic uppercase text-center">重置实验凭证 / Reset Key</h3>
        <form @submit.prevent="handleChangePassword" class="space-y-5">
          <div class="space-y-4">
            <div class="relative group">
              <Key class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 group-focus-within:text-blue-500 transition-colors" />
              <input
                v-model="oldPassword"
                :type="showPasswords ? 'text' : 'password'"
                placeholder="当前密码 / Current Secret"
                class="w-full bg-white/5 border border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-4 outline-none transition-all placeholder:text-slate-600 font-mono"
                required
              />
            </div>
            <div class="relative group">
              <Lock class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 group-focus-within:text-blue-500 transition-colors" />
              <input
                v-model="newPassword"
                :type="showPasswords ? 'text' : 'password'"
                placeholder="核准新密码 / New Authorized Key"
                class="w-full bg-white/5 border border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-12 outline-none transition-all placeholder:text-slate-600 font-mono"
                required
              />
              <button 
                type="button"
                @click="showPasswords = !showPasswords"
                class="absolute right-4 top-1/2 -translate-y-1/2 text-slate-500 hover:text-white"
              >
                <EyeOff v-if="showPasswords" class="w-5 h-5" />
                <Eye v-else class="w-5 h-5" />
              </button>
            </div>
            <div class="relative group">
              <Lock class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 group-focus-within:text-blue-500 transition-colors" />
              <input
                v-model="confirmPassword"
                :type="showPasswords ? 'text' : 'password'"
                placeholder="再次输入新密码"
                class="w-full bg-white/5 border border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-4 outline-none transition-all placeholder:text-slate-600 font-mono"
                required
              />
            </div>
          </div>
          <div class="flex gap-4 pt-4">
            <button 
              type="button"
              @click="showChangePassword = false; oldPassword = ''; newPassword = ''; confirmPassword = '';" 
              class="flex-1 py-4 bg-white/5 hover:bg-white/10 rounded-2xl font-bold transition-all text-slate-400"
            >
              取消
            </button>
            <button 
              type="submit"
              :disabled="loading"
              class="flex-1 py-4 bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 rounded-2xl font-black text-white shadow-xl shadow-blue-500/20 disabled:opacity-50 flex items-center justify-center gap-2"
            >
              <Loader2 v-if="loading" class="w-5 h-5 animate-spin" />
              执行重置
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
