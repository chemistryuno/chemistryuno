import { useState, FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { authAPI } from '../utils/api'
import { Beaker, Lock, User, Loader2, Fingerprint } from 'lucide-react'
import { cn } from '../utils/cn'

interface LoginProps {
  onLogin: (userData: any, token: string) => void
}

export default function Login({ onLogin }: LoginProps) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const response = await authAPI.login(username, password)
      const { token, user } = response.data
      onLogin(user, token)
      navigate('/')
    } catch (err: any) {
      setError(err.response?.data?.error || '身份验证失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-[#1a1a1e] relative overflow-hidden font-sans">
      {/* Subtle Background Elements */}
      <div className="absolute top-[-10%] right-[-10%] w-[40%] h-[40%] bg-blue-500/5 rounded-full blur-[120px]"></div>
      <div className="absolute bottom-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-500/5 rounded-full blur-[120px]"></div>

      <div className="w-full max-w-md relative z-10 animate-in fade-in zoom-in duration-500">
        <div className="glass-panel-light rounded-[40px] shadow-[0_20px_60px_rgba(0,0,0,0.3)] overflow-hidden">
          <div className="p-10 md:p-12">
            {/* Header Section */}
            <div className="flex flex-col items-center mb-10">
              <div className="w-20 h-20 bg-blue-600 rounded-3xl flex items-center justify-center mb-4 shadow-lg transform rotate-3 hover:rotate-0 transition-transform duration-500">
                <Beaker className="w-10 h-10 text-white" />
              </div>
              <h1 className="text-3xl font-black text-slate-800 tracking-tighter">
                化学<span className="text-blue-600">UNO</span>
              </h1>
              <p className="text-slate-400 text-xs font-bold uppercase tracking-[0.2em] mt-2">Laboratory System Access</p>
            </div>

            {error && (
              <div className="bg-red-50 border border-red-100 text-red-500 px-4 py-3 rounded-2xl mb-6 text-center text-xs font-bold">
                {error}
              </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-5">
              <div className="space-y-1.5">
                <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">识别码 / Username</label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400">
                    <User className="w-4 h-4" />
                  </div>
                  <input
                    type="text"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    required
                    className="w-full bg-slate-100 border border-slate-200 text-slate-800 pl-11 pr-4 py-3.5 rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-400 text-sm font-medium"
                    placeholder="Researcher ID"
                  />
                </div>
              </div>

              <div className="space-y-1.5">
                <label className="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">访问秘钥 / Password</label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400">
                    <Lock className="w-4 h-4" />
                  </div>
                  <input
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                    className="w-full bg-slate-100 border border-slate-200 text-slate-800 pl-11 pr-4 py-3.5 rounded-2xl focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none transition-all placeholder:text-slate-400 text-sm font-medium"
                    placeholder="Auth Token"
                  />
                </div>
              </div>

              <button
                type="submit"
                disabled={loading}
                className={cn(
                  "w-full h-14 rounded-2xl font-black text-white transition-all shadow-lg active:scale-95 flex items-center justify-center gap-2",
                  loading 
                    ? "bg-slate-400 cursor-not-allowed" 
                    : "bg-blue-700 hover:bg-blue-600 shadow-blue-500/20"
                )}
              >
                {loading ? (
                  <Loader2 className="w-5 h-5 animate-spin" />
                ) : (
                  <>
                    <span className="uppercase tracking-widest text-sm">初始化访问</span>
                    <Fingerprint className="w-4 h-4" />
                  </>
                )}
              </button>
            </form>

            <div className="mt-8 text-center">
              <p className="text-slate-400 text-xs font-bold">
                初次参与实验？{" "}
                <Link to="/register" className="text-blue-600 hover:text-blue-700 transition-colors">
                  注册研究员账号
                </Link>
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
