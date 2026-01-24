import { useState, useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { adminAPI } from "../utils/api"
import { 
  Shield, 
  ArrowLeft, 
  Users, 
  Layers, 
  History, 
  Trash2, 
  Edit2, 
  Save, 
  ChevronRight, 
  Terminal,
  Activity,
  Cpu,
  Database,
  Search as SearchIcon
} from "lucide-react"
import { cn } from "../utils/cn"

interface AdminProps {
  user: any
}

export default function Admin(_props: AdminProps) {
  const navigate = useNavigate()
  const [users, setUsers] = useState<any[]>([])
  const [gameHistory, setGameHistory] = useState<any[]>([])
  const [deckConfig, setDeckConfig] = useState<any>(null)
  const [editingDeck, setEditingDeck] = useState(false)
  const [activeTab, setActiveTab] = useState("users")
  const [loading, setLoading] = useState(false)
  const [searchTerm, setSearchTerm] = useState("")

  useEffect(() => {
    loadData()
  }, [activeTab])

  const loadData = async () => {
    setLoading(true)
    try {
      if (activeTab === "users") {
        const response = await adminAPI.getAllUsers()
        setUsers(response.data || [])
      } else if (activeTab === "history") {
        const response = await adminAPI.getGameHistory()
        setGameHistory(response.data || [])
      } else if (activeTab === "deck") {
        const response = await adminAPI.getGlobalDeckConfig()
        setDeckConfig(response.data)
      }
    } catch (error) {
      console.error("��������ʧ��:", error)
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteUser = async (userId: string) => {
    if (!window.confirm("ȷ��Ҫ����ɾ�����о�Ա�𣿴˲��������棡")) return
    try {
      await adminAPI.deleteUser(userId)
      alert("�û��ѴӺ������ݿ�Ĩ��")
      loadData()
    } catch (error: any) {
      alert(error.response?.data?.error || "����ʧ��")
    }
  }

  const handleUpdateDeck = async () => {
    try {
      await adminAPI.updateGlobalDeckConfig(deckConfig.name, deckConfig.cards)
      alert("�������������ͬ����ȫ��")
      setEditingDeck(false)
    } catch (error: any) {
      alert(error.response?.data?.error || "����ʧ��")
    }
  }

  const handleCardCountChange = (cardType: string, value: string) => {
    setDeckConfig({
      ...deckConfig,
      cards: {
        ...deckConfig.cards,
        [cardType]: parseInt(value) || 0,
      },
    })
  }

  return (
    <div className="min-h-screen bg-[#0a0a0c] text-slate-200 p-4 lg:p-10 font-sans selection:bg-orange-500/30">
      <div className="fixed inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] bg-orange-500/5 rounded-full blur-[120px]" />
        <div className="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-blue-500/5 rounded-full blur-[120px]" />
        <div className="absolute inset-0 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-20 brightness-50 contrast-150" />
      </div>

      <div className="max-w-7xl mx-auto relative z-10">
        <header className="flex flex-col lg:flex-row items-center justify-between gap-8 mb-12">
          <div className="flex items-center gap-6">
            <div className="relative group">
              <div className="absolute inset-x-0 inset-y-0 bg-orange-500 blur-2xl opacity-20 group-hover:opacity-40 transition-opacity" />
              <div className="w-16 h-16 rounded-2xl bg-[#111114] border border-orange-500/40 flex items-center justify-center relative z-10 shadow-2xl">
                <Shield className="w-8 h-8 text-orange-400 group-hover:scale-110 transition-transform" />
              </div>
            </div>
            <div>
              <h1 className="text-3xl font-black text-white italic tracking-tighter uppercase flex items-center gap-3">
                System Override <span className="text-xs font-mono bg-orange-500/20 text-orange-400 px-2 py-1 rounded border border-orange-500/30 not-italic">v4.0.2</span>
              </h1>
              <p className="text-slate-500 text-sm font-bold tracking-widest uppercase mt-1">����ʵ����Ȩ�޿������� / Core Admin Console</p>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <div className="px-6 py-3 bg-[#111114] border border-white/5 rounded-2xl flex items-center gap-4 shadow-xl">
              <div className="flex flex-col items-end">
                <span className="text-[10px] font-black text-slate-500 uppercase">Server Status</span>
                <span className="text-xs font-bold text-emerald-400 flex items-center gap-1.5">
                  <span className="w-2 h-2 bg-emerald-500 rounded-full animate-pulse" />
                  STABLE / OP-CON 1
                </span>
              </div>
              <div className="w-px h-8 bg-white/5" />
              <button 
                onClick={() => navigate("/")}
                className="flex items-center gap-2 text-slate-400 hover:text-white transition-colors group"
              >
                <ArrowLeft className="w-4 h-4 group-hover:-translate-x-1 transition-transform" />
                <span className="text-xs font-black uppercase tracking-widest">Exit</span>
              </button>
            </div>
          </div>
        </header>

        <section className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-12">
          {[
            { label: "��Ծ�о�Ա", val: users.length, icon: Users, color: "blue" },
            { label: "ȫ�������", val: deckConfig ? Object.keys(deckConfig.cards).length : 0, icon: Cpu, color: "purple" },
            { label: "ʵ���¼", val: gameHistory.length, icon: History, color: "orange" },
            { label: "���ݿ⸺��", val: "12%", icon: Database, color: "emerald" }
          ].map((stat, i) => (
            <div key={i} className="bg-[#111114] border border-white/10 p-6 rounded-[2rem] hover:border-white/20 transition-all shadow-xl group">
              <div className="flex items-center justify-between mb-4">
                <div className={cn(
                  "p-3 rounded-2xl",
                  stat.color === "blue" && "bg-blue-500/10 text-blue-400",
                  stat.color === "purple" && "bg-purple-500/10 text-purple-400",
                  stat.color === "orange" && "bg-orange-500/10 text-orange-400",
                  stat.color === "emerald" && "bg-emerald-500/10 text-emerald-400"
                )}>
                  <stat.icon className="w-6 h-6" />
                </div>
                <div className="text-[10px] font-black text-slate-600 uppercase tracking-[0.2em] group-hover:text-slate-400 transition-colors">ʵʱ��� / Live</div>
              </div>
              <div className="text-3xl font-black text-white italic">{stat.val}</div>
              <div className="text-xs font-bold text-slate-500 uppercase mt-1 tracking-wider">{stat.label}</div>
            </div>
          ))}
        </section>

        <main className="bg-[#111114] border border-white/10 rounded-[2.5rem] shadow-2xl overflow-hidden min-h-[600px] flex flex-col">
          <nav className="flex border-b border-white/5 bg-black/20 p-2">
            {[
              { id: "users", label: "��Ա���� / PERSONNEL", icon: Users },
              { id: "deck", label: "ȫ����� / REDUCTION", icon: Layers },
              { id: "history", label: "ʵ����Դ / TRACING", icon: History }
            ].map(tab => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={cn(
                  "flex items-center gap-3 px-8 py-5 text-xs font-black uppercase tracking-[0.1em] transition-all rounded-2xl relative",
                  activeTab === tab.id 
                    ? "text-orange-400 bg-white/5" 
                    : "text-slate-500 hover:text-slate-300 hover:bg-white/5"
                )}
              >
                <tab.icon className="w-4 h-4" />
                {tab.label}
                {activeTab === tab.id && (
                  <div className="absolute inset-x-0 bottom-2 px-8">
                    <div className="h-0.5 bg-orange-500 shadow-[0_0_10px_rgba(249,115,22,0.5)] rounded-full" />
                  </div>
                )}
              </button>
            ))}
          </nav>

          <div className="p-10 flex-1">
            {loading ? (
              <div className="h-full flex flex-col items-center justify-center text-slate-500 gap-6 py-20">
                <div className="relative">
                  <div className="w-20 h-20 border-4 border-orange-500/20 border-t-orange-500 rounded-full animate-spin" />
                  <Terminal className="w-8 h-8 text-orange-400 absolute inset-0 m-auto" />
                </div>
                <p className="font-mono text-sm uppercase tracking-widest animate-pulse">Synchronizing Database Layers...</p>
              </div>
            ) : (
              <div className="animate-in fade-in slide-in-from-bottom-4 duration-500">
                {activeTab === "users" && (
                  <div className="space-y-8">
                    <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
                      <h3 className="text-xl font-black italic uppercase text-white flex items-center gap-4">
                        <Terminal className="w-5 h-5 text-orange-400" />
                        �о�Աȫ����¼ <span className="text-slate-600 font-mono not-italic text-xs">/ ROOT@ADMIN:~# list --all</span>
                      </h3>
                      <div className="relative group">
                        <SearchIcon className="w-4 h-4 absolute left-4 top-1/2 -translate-y-1/2 text-slate-600 group-focus-within:text-orange-400 transition-colors" />
                        <input 
                          type="text" 
                          placeholder="SEARCH UID / USERNAME..."
                          value={searchTerm}
                          onChange={(e) => setSearchTerm(e.target.value)}
                          className="bg-black/40 border border-white/5 rounded-2xl pl-12 pr-6 py-3 text-xs font-mono focus:outline-none focus:border-orange-500/30 w-full md:w-80 transition-all placeholder:text-slate-700"
                        />
                      </div>
                    </div>
                    
                    <div className="overflow-x-auto custom-scrollbar">
                      <table className="w-full text-left">
                        <thead>
                          <tr className="text-slate-600 text-[10px] font-black uppercase tracking-[0.2em] border-b border-white/5">
                            <th className="px-6 py-4">Researcher Profile</th>
                            <th className="px-6 py-4">Recognition UID</th>
                            <th className="px-6 py-4">Auth Level</th>
                            <th className="px-6 py-4">Join Date</th>
                            <th className="px-6 py-4 text-right">Actions</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-white/5 font-mono">
                          {users
                            .filter(u => u.username.toLowerCase().includes(searchTerm.toLowerCase()))
                            .map(u => (
                            <tr key={u.id} className="hover:bg-white/5 transition-colors group">
                              <td className="px-6 py-6 text-sm font-bold text-white flex items-center gap-4">
                                <div className="w-10 h-10 bg-white/5 rounded-xl flex items-center justify-center text-xl group-hover:scale-110 transition-transform">
                                  {u.avatar}
                                </div>
                                {u.username}
                              </td>
                              <td className="px-6 py-6 text-xs text-slate-500">{u.id}</td>
                              <td className="px-6 py-6">
                                {u.is_admin ? (
                                  <span className="text-[10px] px-3 py-1 bg-orange-500/10 text-orange-400 rounded-full border border-orange-500/20 font-black tracking-widest">LV.99 CORE</span>
                                ) : (
                                  <span className="text-[10px] px-3 py-1 bg-white/5 text-slate-400 rounded-full border border-white/10 font-black tracking-widest">LV.01 STAFF</span>
                                )}
                              </td>
                              <td className="px-6 py-6 text-xs text-slate-500">{new Date(u.created_at).toLocaleDateString()}</td>
                              <td className="px-6 py-6 text-right">
                                {!u.is_admin && (
                                  <button 
                                    onClick={() => handleDeleteUser(u.id)}
                                    className="p-3 hover:bg-red-500/20 text-slate-600 hover:text-red-400 rounded-xl transition-all"
                                    title="Ĩ��Ȩ��"
                                  >
                                    <Trash2 className="w-5 h-5" />
                                  </button>
                                )}
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                )}

                {activeTab === "deck" && deckConfig && (
                  <div className="max-w-4xl mx-auto space-y-10">
                    <div className="flex items-center justify-between mb-8">
                       <h3 className="text-xl font-black italic uppercase text-white flex items-center gap-4">
                        <Cpu className="w-5 h-5 text-blue-400" />
                        ȫ��������ȿ��� <span className="text-slate-600 font-mono not-italic text-xs">/ CONFIG@REDUCTION --GLOBAL</span>
                      </h3>
                      <button 
                        onClick={() => editingDeck ? handleUpdateDeck() : setEditingDeck(true)}
                        className={cn(
                          "px-6 py-3 rounded-2xl font-black text-xs uppercase tracking-widest flex items-center gap-2 transition-all shadow-xl",
                          editingDeck 
                            ? "bg-emerald-600 hover:bg-emerald-500 text-white" 
                            : "bg-blue-600 hover:bg-blue-500 text-white"
                        )}
                      >
                        {editingDeck ? <Save className="w-4 h-4" /> : <Edit2 className="w-4 h-4" />}
                        {editingDeck ? "�ϴ�ȫ������" : "��������ģʽ"}
                      </button>
                    </div>

                    <div className="grid grid-cols-2 lg:grid-cols-3 gap-6">
                      {Object.entries(deckConfig.cards).map(([type, count]: [string, any]) => (
                        <div key={type} className="bg-[#1a1c1e] border border-white/5 p-6 rounded-3xl hover:border-blue-500/20 transition-all">
                          <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest block mb-4">{type}</label>
                          {editingDeck ? (
                            <input
                              type="number"
                              value={count}
                              onChange={(e) => handleCardCountChange(type, e.target.value)}
                              className="w-full bg-black/40 border border-blue-500/20 rounded-xl px-4 py-3 font-mono text-blue-400 focus:outline-none focus:border-blue-500"
                            />
                          ) : (
                            <div className="text-3xl font-black text-white italic">{count}</div>
                          )}
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {activeTab === "history" && (
                  <div className="space-y-8">
                    <h3 className="text-xl font-black italic uppercase text-white flex items-center gap-4">
                      <History className="w-5 h-5 text-purple-400" />
                      ȫ��ʵ����Դ��¼ <span className="text-slate-600 font-mono not-italic text-xs">/ SCAN@LOGS --ALL</span>
                    </h3>
                    
                    <div className="grid gap-4">
                      {gameHistory.length === 0 ? (
                        <div className="py-20 flex flex-col items-center justify-center border-2 border-dashed border-white/5 rounded-[2.5rem] text-slate-600">
                          <Activity className="w-12 h-12 mb-4 opacity-10" />
                          <p className="italic font-bold">Ŀǰδ���ص��κ���ʷʵ��������</p>
                        </div>
                      ) : (
                        gameHistory.map((game, i) => (
                          <div key={i} className="bg-[#1a1c1e] border border-white/5 p-6 rounded-3xl hover:bg-white/5 transition-all group flex items-center justify-between">
                            <div className="flex items-center gap-6">
                              <div className="w-12 h-12 bg-white/5 rounded-2xl flex items-center justify-center font-mono text-purple-400 font-black">
                                #{i+1}
                              </div>
                              <div>
                                <div className="text-sm font-bold text-white group-hover:text-purple-400 transition-colors">ʵ�鷴Ӧ��: {game.id.substring(0, 8).toUpperCase()}</div>
                                <div className="text-[10px] text-slate-500 uppercase font-black tracking-widest mt-1">
                                  {new Date(game.created_at).toLocaleString()} | ״̬: �ѷ��
                                </div>
                              </div>
                            </div>
                            <ChevronRight className="w-5 h-5 text-slate-700 group-hover:text-purple-400 transition-all group-hover:translate-x-1" />
                          </div>
                        ))
                      )}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </main>
      </div>
    </div>
  )
}
