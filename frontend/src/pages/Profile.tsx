import { useState, FormEvent } from "react"
import { useNavigate } from "react-router-dom"
import { authAPI } from "../utils/api"
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
  EyeOff
} from "lucide-react"
import { cn } from "../utils/cn"

interface ProfileProps {
  user: any
  onLogout: () => void
}

export default function Profile({ user, onLogout }: ProfileProps) {
  const navigate = useNavigate()
  const [showChangePassword, setShowChangePassword] = useState(false)
  const [showChangeAvatar, setShowChangeAvatar] = useState(false)
  const [oldPassword, setOldPassword] = useState("")
  const [newPassword, setNewPassword] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
  const [showPasswords, setShowPasswords] = useState(false)
  const [selectedAvatar, setSelectedAvatar] = useState(user.avatar)
  const [loading, setLoading] = useState(false)

  const avatarOptions = ["", "", "", "", "", "", "", "", "", "", "", ""]

  const handleChangePassword = async (e: FormEvent) => {
    e.preventDefault()
    if (newPassword !== confirmPassword) {
      alert("������������벻һ��")
      return
    }
    setLoading(true)
    try {
      await authAPI.changePassword(oldPassword, newPassword)
      alert("�����޸ĳɹ��������µ�¼")
      onLogout()
      navigate("/login")
    } catch (error: any) {
      alert(error.response?.data?.error || "�޸�����ʧ��")
    } finally {
      setLoading(false)
    }
  }

  const handleChangeAvatar = async () => {
    setLoading(true)
    try {
      await authAPI.updateAvatar(selectedAvatar)
      const updatedUser = { ...user, avatar: selectedAvatar }
      localStorage.setItem("user", JSON.stringify(updatedUser))
      alert("ͷ����³ɹ���")
      window.location.reload()
    } catch (error: any) {
      alert(error.response?.data?.error || "����ͷ��ʧ��")
    } finally {
      setLoading(false)
      setShowChangeAvatar(false)
    }
  }

  const handleDeleteAccount = async () => {
    if (!window.confirm("ȷ��Ҫע���˺��𣿴˲������ɻָ���")) return
    if (!window.confirm("�ٴ�ȷ�ϣ����Ҫɾ���˺���")) return

    try {
      await authAPI.deleteAccount()
      alert("�˺���ע��")
      onLogout()
      navigate("/login")
    } catch (error: any) {
      alert(error.response?.data?.error || "ע���˺�ʧ��")
    }
  }

  return (
    <div className="min-h-screen bg-[#0a0a0c] text-white p-4 md:p-8 selection:bg-blue-500/30">
      <div className="fixed inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-[-10%] right-[-10%] w-[50%] h-[50%] bg-blue-500/5 rounded-full blur-[120px]" />
        <div className="absolute bottom-[-10%] left-[-10%] w-[50%] h-[50%] bg-purple-500/5 rounded-full blur-[120px]" />
        <div className="absolute inset-0 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-20 brightness-50 contrast-150" />
      </div>

      <div className="max-w-5xl mx-auto relative z-10">
        <button 
          onClick={() => navigate("/")} 
          className="group flex items-center gap-3 text-slate-400 hover:text-white mb-10 transition-all px-4 py-2 rounded-full hover:bg-white/5 border border-transparent hover:border-white/10"
        >
          <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
          <span className="font-bold tracking-wider uppercase text-xs">����������� / Back to Hub</span>
        </button>

        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
          <div className="lg:col-span-4 space-y-6">
            <div className="bg-[#111114] border border-white/10 rounded-[2.5rem] p-8 relative overflow-hidden group shadow-2xl">
              <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-transparent via-blue-500/50 to-transparent" />
              
              <div className="flex flex-col items-center">
                <div className="relative group/avatar mb-8">
                  <div className="w-40 h-40 bg-gradient-to-tr from-[#1a1c1e] to-[#2d3035] rounded-[3rem] p-1 shadow-2xl transition-transform duration-500 group-hover/avatar:scale-105">
                    <div className="w-full h-full bg-[#111114] rounded-[2.8rem] flex items-center justify-center text-7xl relative overflow-hidden group/inner transition-all border border-white/5">
                      <div className="absolute inset-0 bg-blue-500/5 opacity-0 group-hover/inner:opacity-100 transition-opacity" />
                      <span className="relative z-10 scale-110 drop-shadow-[0_0_15px_rgba(255,255,255,0.3)]">{user.avatar}</span>
                    </div>
                  </div>
                  
                  <button 
                    onClick={() => setShowChangeAvatar(true)}
                    className="absolute -bottom-2 -right-2 bg-blue-600 hover:bg-blue-500 p-3 rounded-2xl shadow-[0_0_20px_rgba(37,99,235,0.4)] z-20 group-hover:rotate-12 transition-all active:scale-95"
                    title="������������"
                  >
                    <RefreshCw className="w-5 h-5 text-white" />
                  </button>
                </div>

                <div className="text-center space-y-2">
                  <h2 className="text-3xl font-black tracking-tight text-white group-hover:text-blue-400 transition-colors uppercase italic">
                    {user.username}
                  </h2>
                  <div className="flex items-center justify-center gap-2">
                    {user.is_admin ? (
                      <span className="bg-blue-500/10 text-blue-400 text-[10px] font-black px-4 py-1.5 rounded-full border border-blue-500/20 flex items-center gap-2 uppercase tracking-[0.2em]">
                        <Shield className="w-3 h-3" /> ��ϯ�о�Ա / CORE ADM
                      </span>
                    ) : (
                      <span className="bg-slate-500/10 text-slate-400 text-[10px] font-black px-4 py-1.5 rounded-full border border-slate-500/20 flex items-center gap-2 uppercase tracking-[0.2em]">
                        <Fingerprint className="w-3 h-3" /> �����о�Ա / RESEARCHER
                      </span>
                    )}
                  </div>
                </div>

                <div className="w-full mt-10 pt-10 border-t border-white/5 space-y-4">
                  <div className="flex justify-between items-center text-xs">
                    <span className="text-slate-500 font-bold uppercase tracking-widest">Researcher UID</span>
                    <span className="font-mono text-blue-400/80">{user.id.substring(0, 12).toUpperCase()}</span>
                  </div>
                  <div className="flex justify-between items-center text-xs">
                    <span className="text-slate-500 font-bold uppercase tracking-widest">Exp Level</span>
                    <div className="flex items-center gap-2">
                      <div className="w-24 h-1.5 bg-white/5 rounded-full overflow-hidden">
                        <div className="w-1/3 h-full bg-blue-500 shadow-[0_0_10px_rgba(59,130,246,0.5)]" />
                      </div>
                      <span className="text-blue-500 font-black">LV.01</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="bg-white/5 border border-white/5 rounded-3xl p-5 hover:bg-white/[0.08] transition-colors">
                <div className="flex items-center gap-3 mb-2">
                  <Zap className="w-4 h-4 text-yellow-400" />
                  <span className="text-[10px] font-black uppercase text-slate-500">ʤ����</span>
                </div>
                <div className="text-2xl font-black">--</div>
              </div>
              <div className="bg-white/5 border border-white/5 rounded-3xl p-5 hover:bg-white/[0.08] transition-colors">
                <div className="flex items-center gap-3 mb-2">
                  <Activity className="w-4 h-4 text-emerald-400" />
                  <span className="text-[10px] font-black uppercase text-slate-500">ʤ��</span>
                </div>
                <div className="text-2xl font-black">--%</div>
              </div>
            </div>
          </div>

          <div className="lg:col-span-8 space-y-8">
            <div className="bg-[#111114] border border-white/10 rounded-[2.5rem] p-10 relative overflow-hidden">
              <div className="flex items-center justify-between mb-10">
                <div>
                  <h3 className="text-2xl font-black uppercase italic tracking-tighter flex items-center gap-3">
                    <span className="w-2 h-8 bg-blue-500 rounded-full" />
                    �˻���ȫ���� / Security
                  </h3>
                  <p className="text-slate-500 text-sm mt-1">���������о�Աƾ����ʵ���ҷ���Ȩ��</p>
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <button 
                  onClick={() => setShowChangePassword(true)}
                  className="group relative flex flex-col items-start p-6 bg-white/5 hover:bg-blue-500/10 border border-white/5 hover:border-blue-500/30 rounded-3xl transition-all text-left"
                >
                  <div className="bg-blue-500/20 p-3 rounded-2xl mb-4 group-hover:scale-110 transition-transform">
                    <Lock className="w-6 h-6 text-blue-400" />
                  </div>
                  <span className="text-lg font-bold">�޸��о�����</span>
                  <span className="text-slate-500 text-xs mt-1">���ڸ���ƾ����ȷ��ʵ�����ݰ�ȫ</span>
                </button>

                <button 
                  onClick={handleDeleteAccount}
                  className="group relative flex flex-col items-start p-6 bg-white/5 hover:bg-red-500/10 border border-white/5 hover:border-red-500/30 rounded-3xl transition-all text-left"
                >
                  <div className="bg-red-500/20 p-3 rounded-2xl mb-4 group-hover:scale-110 transition-transform">
                    <UserX className="w-6 h-6 text-red-500" />
                  </div>
                  <span className="text-lg font-bold text-red-400">�����о�ϯλ</span>
                  <span className="text-slate-500 text-xs mt-1">����ע���˻��������о����ݽ�������</span>
                </button>
              </div>
            </div>

            <div className="bg-[#111114] border border-white/10 rounded-[2.5rem] p-10">
              <h3 className="text-xl font-bold uppercase tracking-widest mb-6 flex items-center gap-3 text-slate-400">
                <Award className="w-5 h-5" />
                ʵ���ҳɾ� / Achievements
              </h3>
              <div className="flex flex-col items-center justify-center py-20 border-2 border-dashed border-white/5 rounded-[2rem] bg-white/[0.02]">
                <Shield className="w-12 h-12 text-slate-700 mb-4" />
                <p className="text-slate-500 font-medium italic">����ѫ�¼�¼����ȥ����һ����ѧ��Ӧ�ɣ�</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      {showChangeAvatar && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 backdrop-blur-xl bg-black/60">
          <div className="bg-[#111114] border border-white/10 rounded-[3rem] p-10 max-w-xl w-full shadow-2xl relative animate-in fade-in zoom-in duration-300">
            <h3 className="text-2xl font-black mb-8 italic uppercase text-center">ѡ���µ��������� / Select Avatar</h3>
            <div className="grid grid-cols-4 sm:grid-cols-6 gap-4 mb-10">
              {avatarOptions.map(emoji => (
                <button
                  key={emoji}
                  onClick={() => setSelectedAvatar(emoji)}
                  className={cn(
                    "w-16 h-16 text-3xl flex items-center justify-center rounded-[1.5rem] transition-all duration-300 border-2",
                    selectedAvatar === emoji 
                      ? "bg-blue-600 border-blue-400 scale-110 shadow-[0_0_20px_rgba(59,130,246,0.5)]" 
                      : "bg-white/5 border-transparent hover:border-white/20 hover:scale-105"
                  )}
                >
                  {emoji}
                </button>
              ))}
            </div>
            <div className="flex gap-4">
              <button 
                onClick={() => setShowChangeAvatar(false)}
                className="flex-1 py-4 bg-white/5 hover:bg-white/10 rounded-2xl font-bold transition-all text-slate-400"
              >
                ȡ��
              </button>
              <button 
                onClick={handleChangeAvatar}
                disabled={loading}
                className="flex-1 py-4 bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 rounded-2xl font-black text-white shadow-xl shadow-blue-500/20 disabled:opacity-50 flex items-center justify-center gap-2"
              >
                {loading && <Loader2 className="w-5 h-5 animate-spin" />}
                ͬ����������
              </button>
            </div>
          </div>
        </div>
      )}

      {showChangePassword && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 backdrop-blur-xl bg-black/60">
          <div className="bg-[#111114] border border-white/10 rounded-[3rem] p-10 max-md w-full shadow-2xl relative animate-in fade-in zoom-in duration-300">
            <h3 className="text-2xl font-black mb-8 italic uppercase text-center">�����о�ƾ�� / Reset Key</h3>
            <form onSubmit={handleChangePassword} className="space-y-5">
              <div className="space-y-4">
                <div className="relative group">
                  <Key className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 group-focus-within:text-blue-500 transition-colors" />
                  <input
                    type={showPasswords ? "text" : "password"}
                    placeholder="��ǰ����"
                    value={oldPassword}
                    onChange={(e) => setOldPassword(e.target.value)}
                    className="w-full bg-white/5 border border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-4 outline-none transition-all placeholder:text-slate-600 font-mono"
                    required
                  />
                </div>
                <div className="relative group">
                  <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 group-focus-within:text-blue-500 transition-colors" />
                  <input
                    type={showPasswords ? "text" : "password"}
                    placeholder="��׼������"
                    value={newPassword}
                    onChange={(e) => setNewPassword(e.target.value)}
                    className="w-full bg-white/5 border border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-12 outline-none transition-all placeholder:text-slate-600 font-mono"
                    required
                  />
                  <button 
                    type="button"
                    onClick={() => setShowPasswords(!showPasswords)}
                    className="absolute right-4 top-1/2 -translate-y-1/2 text-slate-500 hover:text-white"
                  >
                    {showPasswords ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                  </button>
                </div>
                <div className="relative group">
                  <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500 group-focus-within:text-blue-500 transition-colors" />
                  <input
                    type={showPasswords ? "text" : "password"}
                    placeholder="�ٴκ�׼������"
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    className="w-full bg-white/5 border border-white/10 focus:border-blue-500/50 rounded-2xl py-4 pl-12 pr-4 outline-none transition-all placeholder:text-slate-600 font-mono"
                    required
                  />
                </div>
              </div>
              
              <div className="flex gap-4 mt-10">
                <button 
                  type="button"
                  onClick={() => setShowChangePassword(false)}
                  className="flex-1 py-4 bg-white/5 hover:bg-white/10 rounded-2xl font-bold transition-all text-slate-400"
                >
                  ȡ��
                </button>
                <button 
                  type="submit"
                  disabled={loading}
                  className="flex-1 py-4 bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 rounded-2xl font-black text-white shadow-xl shadow-blue-500/20 disabled:opacity-50 flex items-center justify-center gap-2"
                >
                  {loading && <Loader2 className="w-5 h-5 animate-spin" />}
                  ִ�и���
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
