import { useState, FormEvent } from "react"
import { Link, useNavigate } from "react-router-dom"
import { authAPI } from "../utils/api"
import { Lock, User, FlaskConical, ShieldCheck, Zap, Loader2 } from "lucide-react"
import { cn } from "../utils/cn"

export default function Register() {
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
  const [error, setError] = useState("")
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError("")

    if (password !== confirmPassword) {
      setError("两次输入的密码不一致")
      return
    }

    setLoading(true)

    try {
      await authAPI.register(username, password)
      alert("注册成功，请登录")
      navigate("/login")
    } catch (err: any) {
      setError(err.response?.data?.error || "注册失败，用户名可能已存在")
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-[#1a1a1e] relative overflow-hidden font-sans">
      <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-500/5 rounded-full blur-[120px]"></div>
      <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-blue-500/5 rounded-full blur-[120px]"></div>

      <div className="w-full max-w-md relative z-10 animate-in fade-in zoom-in duration-500">
        <div className="glass-panel-light rounded-[40px] shadow-[0_20px_60px_rgba(0,0,0,0.3)] overflow-hidden">
          <div className="p-10 md:p-12">
            <div className="flex flex-col items-center mb-10">
              <div className="w-20 h-20 bg-blue-600 rounded-3xl flex items-center justify-center mb-4 shadow-lg transform -rotate-3 hover:rotate-0 transition-transform duration-500">
                <FlaskConical className="w-10 h-10 text-white" />
              </div>
              <h1 className="text-3xl font-black text-slate-800 tracking-tighter">
                加入<span className="text-blue-600">实验室</span>
              </h1>
              <p className="text-slate-500 text-sm mt-2 font-medium">创建您的研究员账户</p>
            </div>

            <form onSubmit={handleSubmit} className="space-y-6">
              {error && (
                <div className="flex items-center gap-2 p-4 bg-red-50 border border-red-100 text-red-600 text-sm rounded-2xl animate-shake">
                  <div className="w-2 h-2 rounded-full bg-red-400"></div>
                  {error}
                </div>
              )}

              <div className="space-y-4">
                <div className="relative group">
                  <div className="absolute left-5 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-blue-500 transition-colors">
                    <User size={20} strokeWidth={2.5} />
                  </div>
                  <input
                    type="text"
                    required
                    className="w-full pl-14 pr-6 py-5 bg-slate-100/50 border-2 border-transparent focus:border-blue-500 focus:bg-white rounded-2xl text-slate-800 placeholder:text-slate-400 font-bold outline-none transition-all"
                    placeholder="用户名"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                  />
                </div>

                <div className="relative group">
                  <div className="absolute left-5 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-blue-500 transition-colors">
                    <Lock size={20} strokeWidth={2.5} />
                  </div>
                  <input
                    type="password"
                    required
                    className="w-full pl-14 pr-6 py-5 bg-slate-100/50 border-2 border-transparent focus:border-blue-500 focus:bg-white rounded-2xl text-slate-800 placeholder:text-slate-400 font-bold outline-none transition-all"
                    placeholder="密 码"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                  />
                </div>

                <div className="relative group">
                  <div className="absolute left-5 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-blue-500 transition-colors">
                    <ShieldCheck size={20} strokeWidth={2.5} />
                  </div>
                  <input
                    type="password"
                    required
                    className="w-full pl-14 pr-6 py-5 bg-slate-100/50 border-2 border-transparent focus:border-blue-500 focus:bg-white rounded-2xl text-slate-800 placeholder:text-slate-400 font-bold outline-none transition-all"
                    placeholder="确认密码"
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                  />
                </div>
              </div>

              <button
                type="submit"
                disabled={loading}
                className="w-full py-5 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-400 text-white rounded-2xl font-black text-lg shadow-[0_15px_30px_rgba(37,99,235,0.3)] hover:shadow-[0_20px_40px_rgba(37,99,235,0.4)] transition-all flex items-center justify-center gap-3 transform active:scale-[0.98]"
              >
                {loading ? (
                  <Loader2 className="w-6 h-6 animate-spin" />
                ) : (
                  <>
                    <Zap className="w-5 h-5 fill-current" />
                    立即注册
                  </>
                )}
              </button>
            </form>

            <div className="mt-10 text-center">
              <p className="text-slate-500 font-medium">
                已有账户？{" "}
                <Link to="/login" className="text-blue-600 font-black hover:underline cursor-pointer">
                  登录系统
                </Link>
              </p>
            </div>
          </div>
          
          <div className="bg-slate-50 p-6 flex justify-around border-t border-slate-100">
            <div className="flex flex-col items-center">
              <span className="text-[10px] font-black text-slate-400 uppercase tracking-widest">Protocol</span>
              <span className="text-xs font-bold text-slate-600">Secure SHA-256</span>
            </div>
            <div className="w-px h-8 bg-slate-200"></div>
            <div className="flex flex-col items-center">
              <span className="text-[10px] font-black text-slate-400 uppercase tracking-widest">Database</span>
              <span className="text-xs font-bold text-slate-600">Chemistry DB</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
