import { useState, useEffect, FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { gameAPI } from '../utils/api'
import websocket from '../utils/websocket'
import { Beaker, Plus, Users, Shield, LogOut, Settings, Play, Info, X, Loader2 } from 'lucide-react'
import { cn } from '../utils/cn'

interface LobbyProps {
  user: any
  onLogout: () => void
}

export default function Lobby({ user, onLogout }: LobbyProps) {
  const [rooms, setRooms] = useState<any[]>([])
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [roomName, setRoomName] = useState('')
  const [maxPlayers, setMaxPlayers] = useState(4)
  const [deckID, setDeckID] = useState(0)
  const [loading, setLoading] = useState(false)
  const [currentTime, setCurrentTime] = useState(new Date())
  const navigate = useNavigate()

  const decks = [
    { id: 0, name: '标准元素包', desc: '包含基础 H, O, C 及常用金属元素', icon: <Beaker className="w-5 h-5" /> },
    { id: 1, name: '有机化学包', desc: '侧重碳链增长与官能团反应 (未解锁)', disabled: true },
    { id: 2, name: '重金属实验室', desc: '增加放射性元素与特殊衰变机制 (未解锁)', disabled: true },
  ]

  useEffect(() => {
    loadRooms()
    websocket.connect()

    const interval = setInterval(loadRooms, 3000)
    const timeInterval = setInterval(() => setCurrentTime(new Date()), 1000)
    
    return () => {
      clearInterval(interval)
      clearInterval(timeInterval)
      websocket.disconnect()
    }
  }, [])

  const loadRooms = async () => {
    try {
      const response = await gameAPI.getRooms()
      setRooms(response.data || [])
    } catch (error) {
      console.error('加载房间列表失败:', error)
    }
  }

  const handleCreateRoom = async (e: FormEvent) => {
    e.preventDefault()
    setLoading(true)

    try {
      const response = await gameAPI.createRoom(roomName, maxPlayers, deckID)
      const room = response.data
      navigate(`/room/${room.id}`)
    } catch (error: any) {
      alert(error.response?.data?.error || '创建房间失败')
    } finally {
      setLoading(false)
    }
  }

  const handleJoinRoom = async (roomId: string) => {
    try {
      await gameAPI.joinRoom(roomId)
      navigate(`/room/${roomId}`)
    } catch (error: any) {
      alert(error.response?.data?.error || '加入房间失败')
    }
  }

  return (
    <div className="min-h-screen bg-[#0a0a0c] text-slate-200 font-sans selection:bg-blue-500/30 overflow-x-hidden">
      {/* Background Decor */}
      <div className="fixed inset-0 pointer-events-none overflow-hidden">
        <div className="absolute top-[-20%] left-[-10%] w-[60%] h-[60%] bg-blue-600/5 rounded-full blur-[150px]"></div>
        <div className="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-purple-600/5 rounded-full blur-[150px]"></div>
        <div className="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/carbon-fibre.png')] opacity-20"></div>
      </div>

      {/* Main Layout Layer */}
      <div className="relative z-10 flex flex-col min-h-screen">
        
        {/* Top Command Bar */}
        <header className="h-20 border-b border-white/5 bg-black/40 backdrop-blur-xl sticky top-0 z-50">
          <div className="max-w-[1400px] mx-auto h-full px-6 flex items-center justify-between">
            <div className="flex items-center gap-6">
              <div className="flex items-center gap-3 group px-4 py-2 bg-gradient-to-br from-blue-500/10 to-blue-600/5 border border-blue-500/20 rounded-2xl">
                <Beaker className="w-8 h-8 text-blue-400 group-hover:rotate-12 transition-transform" />
                <div>
                   <h1 className="text-lg font-black tracking-tighter text-white leading-none">CHEMISTRY <span className="text-blue-500">UNO</span></h1>
                   <p className="text-[10px] text-blue-500/50 font-mono tracking-widest leading-none mt-1 uppercase">Lab_Control_v4</p>
                </div>
              </div>

              {/* Status Indicators (Desktop) */}
              <div className="hidden lg:flex items-center gap-6 text-[10px] font-mono tracking-[0.2em] text-slate-500 border-l border-white/10 pl-6 uppercase">
                <div className="flex items-center gap-2">
                  <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></div>
                  SERVER: STABLE
                </div>
                <div className="flex items-center gap-2">
                   <div className="w-1.5 h-1.5 rounded-full bg-blue-500"></div>
                   UP_TIME: {currentTime.toLocaleTimeString()}
                </div>
              </div>
            </div>

            <div className="flex items-center gap-4">
              {/* User Identity Chip */}
              <div className="hidden sm:flex items-center gap-3 pl-2 pr-4 py-1.5 bg-white/5 border border-white/10 rounded-2xl hover:bg-white/10 transition-all cursor-pointer group">
                 <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-slate-700 to-slate-800 flex items-center justify-center text-xl shadow-inner group-hover:scale-105 transition-transform">
                   {user.avatar}
                 </div>
                 <div className="flex flex-col">
                   <span className="text-xs font-black text-white">{user.username}</span>
                   <span className="text-[9px] text-slate-500 font-mono flex items-center gap-1 uppercase">
                     {user.is_admin ? (
                       <><Shield className="w-2.5 h-2.5 text-yellow-500" /> Research_Lead</>
                     ) : (
                       <>Researcher_Alpha</>
                     )}
                   </span>
                 </div>
              </div>

              <div className="flex items-center gap-1.5">
                <Link to="/profile" className="p-3 hover:bg-white/5 rounded-2xl transition-all text-slate-400 hover:text-white" title="实验室档案">
                  <Settings className="w-5 h-5" />
                </Link>
                {user.is_admin && (
                  <Link to="/admin" className="p-3 hover:bg-yellow-500/10 rounded-2xl transition-all text-yellow-500/70 hover:text-yellow-400" title="科研管理">
                    <Shield className="w-5 h-5" />
                  </Link>
                )}
                <div className="w-px h-6 bg-white/10 mx-1"></div>
                <button onClick={onLogout} className="p-3 hover:bg-red-500/10 rounded-2xl transition-all text-red-500/70 hover:text-red-400" title="切断连接">
                  <LogOut className="w-5 h-5" />
                </button>
              </div>
            </div>
          </div>
        </header>

        <main className="flex-1 max-w-[1400px] mx-auto w-full px-6 py-8">
          {/* Welcome & Global Actions */}
          <div className="flex flex-col lg:flex-row lg:items-end justify-between mb-12 gap-8">
            <div className="space-y-4">
              <div className="inline-flex items-center gap-2 px-3 py-1 bg-blue-500/10 border border-blue-500/20 rounded-full">
                <span className="w-1.5 h-1.5 bg-blue-500 rounded-full animate-ping"></span>
                <span className="text-[10px] font-bold text-blue-400 uppercase tracking-widest">Live Research Hall</span>
              </div>
              <h2 className="text-5xl font-black text-white tracking-tighter leading-none">
                实验大厅
              </h2>
              <p className="text-slate-400 max-w-lg font-medium leading-relaxed">
                欢迎回到元素实验室。目前有 <span className="text-white font-bold">{rooms.length}</span> 个活跃实验，请加入现有队列或开启全新化学反应序列。
              </p>
            </div>

            <div className="flex items-center gap-6">
               <div className="hidden xl:flex items-center gap-8 px-8 py-5 bg-white/5 border border-white/5 rounded-[32px]">
                 <div className="text-center">
                   <p className="text-[10px] text-slate-500 uppercase font-bold tracking-widest mb-1">Total_Players</p>
                   <p className="text-2xl font-black text-white font-mono">1,248</p>
                 </div>
                 <div className="w-px h-8 bg-white/5 font-mono"></div>
                 <div className="text-center font-mono">
                   <p className="text-[10px] text-slate-500 uppercase font-bold tracking-widest mb-1">Active_Nodes</p>
                   <p className="text-2xl font-black text-blue-400">{rooms.filter(r => r.status === 'playing').length}</p>
                 </div>
               </div>

               <button 
                onClick={() => setShowCreateModal(true)} 
                className="group relative flex items-center gap-3 bg-blue-600 hover:bg-blue-500 px-8 py-5 rounded-[24px] font-black text-white shadow-[0_20px_40px_rgba(37,99,235,0.2)] transition-all hover:scale-[1.02] hover:-translate-y-1 active:scale-95 overflow-hidden"
              >
                <Plus className="w-5 h-5 group-hover:rotate-90 transition-transform duration-500" />
                <span className="uppercase tracking-widest text-sm">启动新实验</span>
                <div className="absolute inset-0 w-full h-full bg-gradient-to-r from-transparent via-white/10 to-transparent -translate-x-full group-hover:animate-shimmer"></div>
              </button>
            </div>
          </div>

          {/* Experimental Nodes (Room List) */}
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
            {rooms.length === 0 ? (
              <div className="col-span-full py-32 flex flex-col items-center justify-center bg-white/[0.02] border-2 border-dashed border-white/5 rounded-[40px] text-slate-600 transition-colors hover:bg-white/[0.03] hover:border-white/10 group">
                <div className="w-24 h-24 bg-white/5 rounded-full flex items-center justify-center mb-6 group-hover:scale-110 transition-transform">
                  <Info className="w-10 h-10 opacity-30" />
                </div>
                <p className="text-2xl font-black text-slate-500 tracking-tight">NO_ACTIVE_EXPERIMENTS</p>
                <p className="text-sm mt-3 font-mono opacity-50 uppercase tracking-widest">请等待节点激活或手动创建</p>
              </div>
            ) : (
              rooms.map(room => (
                <div 
                  key={room.id} 
                  className="group relative bg-[#121216]/60 backdrop-blur-xl border border-white/10 rounded-[32px] p-1 transition-all hover:bg-[#16161c] hover:border-blue-500/30 hover:shadow-[0_20px_50px_rgba(0,0,0,0.5)] flex flex-col h-[320px]"
                >
                  <div className="flex-1 p-6 flex flex-col">
                    <div className="flex justify-between items-start mb-6">
                      <div className="flex flex-col">
                        <span className="text-[10px] font-mono text-blue-500/60 uppercase tracking-widest mb-1">Experiment_ID_{room.id.substring(0, 4)}</span>
                        <h3 className="text-xl font-black text-white group-hover:text-blue-400 transition-colors truncate max-w-[180px] leading-tight">
                          {room.name}
                        </h3>
                      </div>
                      <div className={cn(
                        "px-2.5 py-1 rounded-lg text-[9px] font-black uppercase tracking-widest border",
                        room.status === 'waiting' ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20" : 
                        room.status === 'playing' ? "bg-amber-500/10 text-amber-400 border-amber-500/20" : 
                        "bg-slate-500/10 text-slate-400 border-slate-500/20"
                      )}>
                        {room.status === 'waiting' ? '● Ready' : room.status === 'playing' ? '○ Active' : 'End'}
                      </div>
                    </div>

                    <div className="space-y-4 mb-auto">
                      <div className="flex items-center gap-3 p-3 bg-white/5 rounded-2xl border border-white/5 group-hover:border-white/10 transition-colors">
                        <div className="w-8 h-8 rounded-lg bg-blue-500/10 flex items-center justify-center">
                          <Users className="w-4 h-4 text-blue-400" />
                        </div>
                        <div className="flex flex-col">
                          <span className="text-[9px] text-slate-500 uppercase tracking-widest font-bold leading-none mb-1">Participants</span>
                          <span className="text-sm font-black text-white leading-none">
                            {room.players?.length || 0} <span className="text-slate-600 font-normal">/ {room.max_players}</span>
                          </span>
                        </div>
                      </div>

                      <div className="flex items-center gap-3 p-3 bg-white/5 rounded-2xl border border-white/5 group-hover:border-white/10 transition-colors">
                        <div className="w-8 h-8 rounded-lg bg-purple-500/10 flex items-center justify-center">
                          <Shield className="w-4 h-4 text-purple-400" />
                        </div>
                        <div className="flex flex-col">
                          <span className="text-[9px] text-slate-500 uppercase tracking-widest font-bold leading-none mb-1">Safety_Level</span>
                          <span className="text-sm font-black text-white leading-none uppercase">Standard_Alpha</span>
                        </div>
                      </div>
                    </div>

                    <div className="mt-6">
                      {room.status === 'waiting' && (room.players?.length || 0) < room.max_players ? (
                        <button 
                          onClick={() => handleJoinRoom(room.id)} 
                          className="w-full h-14 bg-white/5 hover:bg-blue-600 text-white border border-white/10 hover:border-blue-500 rounded-[20px] font-black transition-all flex items-center justify-center gap-2 group/btn relative overflow-hidden active:scale-95"
                        >
                          <Play className="w-4 h-4 fill-current group-hover/btn:translate-x-1 transition-transform" />
                          <span className="uppercase tracking-widest text-xs">执行初始化</span>
                        </button>
                      ) : (
                        <div className="w-full h-14 bg-slate-800/20 border border-white/5 rounded-[20px] flex items-center justify-center gap-2 grayscale opacity-50 cursor-not-allowed">
                          <Loader2 className="w-4 h-4 animate-spin text-slate-500" />
                          <span className="uppercase tracking-widest text-xs font-bold text-slate-500">正在进行中</span>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </main>

        {/* Global Footer Terminal */}
        <footer className="mt-auto border-t border-white/5 bg-black/40 backdrop-blur-md p-4">
          <div className="max-w-[1400px] mx-auto flex flex-col md:flex-row justify-between items-center text-[10px] font-mono text-slate-500 uppercase tracking-[0.2em] gap-4">
            <div className="flex items-center gap-4">
              <span>System_Core_Ready</span>
              <span className="h-3 w-px bg-white/10"></span>
              <span className="text-blue-500">Secure_WebSocket_Active</span>
              <span className="h-3 w-px bg-white/10"></span>
              <span className="hidden sm:inline">AES_ENCRYPTION_ENABLED</span>
            </div>
            <div>
              &copy; 2024 LAB_V4-ALPHA PROTCOL. ALL RIGHTS RESERVED.
            </div>
          </div>
        </footer>
      </div>

      {/* Modern Create Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-[#000]/80 backdrop-blur-md animate-in fade-in" onClick={() => setShowCreateModal(false)} />
          <div className="relative w-full max-w-lg bg-[#121216] border border-white/10 rounded-[40px] shadow-2xl overflow-hidden animate-in fade-in zoom-in slide-in-from-bottom-10 duration-500">
             {/* Modal Header */}
             <div className="px-8 py-8 border-b border-white/5 flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <div className="w-12 h-12 bg-blue-500/10 border border-blue-500/20 rounded-2xl flex items-center justify-center text-blue-400">
                    <Plus className="w-6 h-6" />
                  </div>
                  <div>
                    <h2 className="text-2xl font-black text-white tracking-tight">开启新实验</h2>
                    <p className="text-[10px] text-slate-500 font-mono uppercase tracking-widest">Setup_Experiment_Parameters</p>
                  </div>
                </div>
                <button 
                  onClick={() => setShowCreateModal(false)}
                  className="p-3 hover:bg-white/5 rounded-2xl transition-colors text-slate-500 hover:text-white"
                >
                  <X className="w-6 h-6" />
                </button>
             </div>

            <form onSubmit={handleCreateRoom} className="p-10 space-y-8">
              <div className="space-y-3">
                <div className="flex justify-between items-center px-1">
                   <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest">实验空间命名</label>
                   <span className="text-[9px] text-blue-500/40">IDENTIFIER_ALPHA</span>
                </div>
                <input
                  type="text"
                  value={roomName}
                  onChange={(e) => setRoomName(e.target.value)}
                  required
                  autoFocus
                  placeholder="请输入实验代号..."
                  className="w-full bg-black/40 border border-white/5 text-white px-6 py-5 rounded-3xl focus:ring-1 focus:ring-blue-500/50 focus:border-blue-500/50 outline-none transition-all placeholder:text-slate-800 font-mono text-sm"
                />
              </div>

              <div className="space-y-4">
                <div className="flex justify-between items-center px-1">
                   <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest">参与研究员人数</label>
                   <span className="text-[9px] text-blue-500/40">CAPACITY_CONFIG</span>
                </div>
                <div className="grid grid-cols-4 gap-4">
                  {[2, 3, 4, 8].map(num => (
                    <button
                      key={num}
                      type="button"
                      onClick={() => setMaxPlayers(num)}
                      className={cn(
                        "h-16 rounded-2xl text-sm font-black border transition-all flex items-center justify-center relative group/opt overflow-hidden",
                        maxPlayers === num 
                          ? "bg-blue-500/10 border-blue-500/50 text-blue-400 ring-1 ring-blue-500/20 shadow-[0_0_20px_rgba(59,130,246,0.1)]" 
                          : "bg-white/5 border-white/5 text-slate-600 hover:bg-white/10 hover:border-white/10"
                      )}
                    >
                      <span className="relative z-10">{num}P</span>
                      {maxPlayers === num && <div className="absolute inset-0 bg-blue-500/5 animate-pulse"></div>}
                    </button>
                  ))}
                </div>
              </div>

              <div className="space-y-4">
                <div className="flex justify-between items-center px-1">
                   <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest">选择实验牌组</label>
                   <span className="text-[9px] text-blue-500/40">DECK_PROTOCOL</span>
                </div>
                <div className="space-y-3">
                  {decks.map(deck => (
                    <button
                      key={deck.id}
                      type="button"
                      disabled={deck.disabled}
                      onClick={() => setDeckID(deck.id)}
                      className={cn(
                        "w-full flex items-center gap-4 p-4 rounded-3xl border transition-all text-left group/deck",
                        deck.disabled ? "opacity-40 cursor-not-allowed border-white/5 grayscale" :
                        deckID === deck.id 
                          ? "bg-blue-600/10 border-blue-500/50 shadow-[0_10px_30px_rgba(59,130,246,0.1)]" 
                          : "bg-white/5 border-white/5 hover:border-white/10 hover:bg-white/10"
                      )}
                    >
                      <div className={cn(
                        "w-12 h-12 rounded-2xl flex items-center justify-center transition-colors",
                        deckID === deck.id ? "bg-blue-500 text-white" : "bg-white/5 text-slate-500"
                      )}>
                        {deck.icon}
                      </div>
                      <div className="flex-1">
                        <p className={cn("text-xs font-black uppercase tracking-wider", deckID === deck.id ? "text-blue-400" : "text-white")}>
                          {deck.name}
                        </p>
                        <p className="text-[10px] text-slate-500 mt-1 font-medium">{deck.desc}</p>
                      </div>
                      {!deck.disabled && deckID === deck.id && (
                        <div className="w-2 h-2 rounded-full bg-blue-500 animate-pulse mr-2"></div>
                      )}
                    </button>
                  ))}
                </div>
              </div>

              <div className="flex gap-4 pt-6">
                <button 
                  type="button" 
                  onClick={() => setShowCreateModal(false)} 
                  className="flex-1 h-14 bg-white/5 hover:bg-white/10 text-slate-400 font-bold rounded-2xl transition-all uppercase tracking-widest text-xs border border-white/5"
                >
                  放弃设置
                </button>
                <button 
                  type="submit" 
                  disabled={loading}
                  className="flex-1 h-14 bg-blue-600 hover:bg-blue-500 text-white font-black rounded-2xl transition-all shadow-[0_15px_30px_rgba(37,99,235,0.3)] active:scale-95 disabled:grayscale flex items-center justify-center gap-2 group/sub relative overflow-hidden"
                >
                  {loading ? (
                    <Loader2 className="w-5 h-5 animate-spin" />
                  ) : (
                    <>
                      <Play className="w-3.5 h-3.5 fill-current" />
                      <span className="uppercase tracking-[0.2em] text-xs">执行初始化</span>
                    </>
                  )}
                  <div className="absolute inset-0 w-full h-full bg-gradient-to-r from-transparent via-white/10 to-transparent -translate-x-full group-hover/sub:animate-shimmer"></div>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
