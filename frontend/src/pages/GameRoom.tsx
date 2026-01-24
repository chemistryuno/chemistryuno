import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { gameAPI } from '../utils/api'
import websocket from '../utils/websocket'
import { ArrowLeft, Play, RefreshCw, Zap, FlaskConical, Trophy, Info, ChevronRight, Loader2, Users } from 'lucide-react'
import { cn } from '../utils/cn'

interface GameRoomProps {
  user: any
}

export default function GameRoom({ user }: GameRoomProps) {
  const { id } = useParams()
  const navigate = useNavigate()
  const [gameState, setGameState] = useState<any>(null)
  const [availableSubstances, setAvailableSubstances] = useState([])
  const [selectedCard, setSelectedCard] = useState<any>(null)
  const [selectedSubstance, setSelectedSubstance] = useState<any>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadGameState()
    
    websocket.joinRoom(id || '')
    websocket.on('game_update', handleGameUpdate)
    websocket.on('player_joined', loadGameState)
    websocket.on('player_left', loadGameState)

    return () => {
      websocket.leaveRoom()
      websocket.off('game_update', handleGameUpdate)
      websocket.off('player_joined', loadGameState)
      websocket.off('player_left', loadGameState)
    }
  }, [id])

  const loadGameState = async () => {
    try {
      // In a real app, we'd fetch the current state here
      // setGameState(initialData)
      setLoading(false)
    } catch (error) {
      console.error('加载游戏状态失败:', error)
      setLoading(false)
    }
  }

  const handleGameUpdate = (message: any) => {
    if (message.data) {
      setGameState(message.data)
    }
  }

  const handleStartGame = async () => {
    try {
      await gameAPI.startGame(id || '')
      // 游戏开始后立即加载游戏状态
      await loadGameState()
    } catch (error: any) {
      alert(error.response?.data?.error || '开始游戏失败')
    }
  }

  const handleCardClick = async (card: any) => {
    if (selectedCard?.type === card.type) {
      setSelectedCard(null)
      setSelectedSubstance(null)
      setAvailableSubstances([])
      return
    }

    setSelectedCard(card)
    setSelectedSubstance(null)
    
    try {
      const response = await gameAPI.getAvailableSubstances(id || '')
      setAvailableSubstances(response.data || [])
    } catch (error) {
      console.error('获取可用物质失败:', error)
    }
  }

  const handlePlayCard = async () => {
    if (!selectedCard || !selectedSubstance) {
      alert('请选择物质')
      return
    }

    try {
      await gameAPI.playCard(id || '', selectedCard, selectedSubstance)
      setSelectedCard(null)
      setSelectedSubstance(null)
      setAvailableSubstances([])
    } catch (error: any) {
      alert(error.response?.data?.error || '出牌失败')
    }
  }

  const handleDrawCard = async () => {
    try {
      await gameAPI.drawCard(id || '')
    } catch (error: any) {
      alert(error.response?.data?.error || '摸牌失败')
    }
  }

  const handleLeaveRoom = async () => {
    try {
      if (window.confirm('确定要离开房间吗？')) {
        await gameAPI.leaveRoom(id || '')
        navigate('/')
      }
    } catch (error) {
      console.error('离开房间失败:', error)
      navigate('/')
    }
  }

  const getCardStyle = (card: any) => {
    if (card.effect === 'reverse' || card.effect === 'skip' || card.effect === 'draw2') return 'special'
    if (card.effect === 'wild' || card.effect === 'wild4') return 'noble'
    return 'element'
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-[#0a0a0c] flex flex-col items-center justify-center p-4 relative overflow-hidden">
        {/* Background Elements */}
        <div className="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] bg-blue-600/10 rounded-full blur-[120px] animate-pulse"></div>
        <div className="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-purple-600/10 rounded-full blur-[120px]"></div>
        <div className="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/carbon-fibre.png')] opacity-20"></div>

        <div className="relative z-10 flex flex-col items-center gap-6 animate-in fade-in zoom-in duration-700">
          <div className="relative group">
            <div className="w-24 h-24 bg-blue-500/10 border border-blue-500/30 rounded-[32px] flex items-center justify-center transform rotate-12 group-hover:rotate-0 transition-all duration-700">
              <FlaskConical className="w-12 h-12 text-blue-400 group-hover:scale-110 transition-transform" />
            </div>
            <div className="absolute -top-2 -right-2 w-8 h-8 bg-blue-500 rounded-xl flex items-center justify-center animate-bounce shadow-[0_0_20px_rgba(59,130,246,0.5)]">
               <Zap className="w-4 h-4 text-white fill-current" />
            </div>
          </div>
          <div className="text-center space-y-2">
            <h2 className="text-2xl font-black text-white tracking-widest uppercase">Initializing Lab</h2>
            <div className="flex items-center gap-1 justify-center">
               <span className="w-1.5 h-1.5 bg-blue-500 rounded-full animate-bounce [animation-delay:-0.3s]"></span>
               <span className="w-1.5 h-1.5 bg-blue-500 rounded-full animate-bounce [animation-delay:-0.15s]"></span>
               <span className="w-1.5 h-1.5 bg-blue-500 rounded-full animate-bounce"></span>
            </div>
          </div>
        </div>
      </div>
    )
  }

  const currentPlayerObj = gameState?.players?.[gameState.current_player]
  const isMyTurn = currentPlayerObj?.user_id === user.id
  const myData = gameState?.players?.find((p: any) => p.user_id === user.id)

  return (
    <div className="min-h-screen bg-[#0a0a0c] text-white overflow-hidden flex flex-col font-sans selection:bg-blue-500/30">
      {/* Dynamic Background */}
      <div className="fixed inset-0 pointer-events-none">
        <div className="absolute top-1/4 left-1/4 w-[50%] h-[50%] bg-blue-600/5 rounded-full blur-[150px] animate-pulse"></div>
        <div className="absolute bottom-1/4 right-1/4 w-[50%] h-[50%] bg-purple-600/5 rounded-full blur-[150px] animate-pulse delay-1000"></div>
        <div className="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/carbon-fibre.png')] opacity-20"></div>
        {/* Scanning Line */}
        <div className="absolute top-0 left-0 w-full h-px bg-blue-500/20 shadow-[0_0_15px_rgba(59,130,246,0.5)] animate-scan"></div>
      </div>

      {/* Experimental Header */}
      <header className="h-[72px] bg-black/40 backdrop-blur-2xl border-b border-white/5 px-6 flex justify-between items-center z-50 sticky top-0">
        <div className="flex items-center gap-6">
          <button 
            onClick={handleLeaveRoom} 
            className="w-10 h-10 flex items-center justify-center hover:bg-white/5 rounded-2xl text-slate-500 hover:text-white transition-all group"
          >
            <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
          </button>
          <div className="h-8 w-px bg-white/10 hidden sm:block"></div>
          <div>
            <div className="flex items-center gap-2">
               <span className="text-[10px] font-mono text-blue-500/50 uppercase tracking-widest">Node_Identifier:</span>
               <h2 className="text-sm font-black tracking-widest uppercase font-mono">{id?.substring(0, 8)}</h2>
            </div>
            <div className="flex items-center gap-2 mt-0.5">
               <div className={cn("w-1.5 h-1.5 rounded-full animate-pulse", gameState?.status === 'waiting' ? "bg-amber-500" : "bg-emerald-500")}></div>
               <p className="text-[10px] text-slate-400 font-bold uppercase tracking-widest">
                {gameState?.status === 'waiting' ? 'Waiting_for_Connection' : 'Reaction_In_Progress'}
               </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-4">
          {gameState?.status === 'waiting' && myData?.user_id === gameState?.host_id && (
            <button 
              onClick={handleStartGame} 
              className="bg-blue-600 hover:bg-blue-500 px-6 h-11 rounded-2xl font-black text-xs uppercase tracking-[0.2em] shadow-[0_10px_20px_rgba(37,99,235,0.2)] transition-all active:scale-95 flex items-center gap-3 group overflow-hidden relative"
            >
              <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/10 to-transparent -translate-x-full group-hover:animate-shimmer"></div>
              <Play className="w-3.5 h-3.5 fill-current" />
              <span>启动反应堆</span>
            </button>
          )}
          
          <div className="px-5 h-11 bg-white/5 rounded-2xl border border-white/10 flex items-center gap-4 font-mono">
             <div className="flex flex-col items-end">
               <span className="text-[9px] text-slate-500 font-bold uppercase tracking-tight">Personnel</span>
               <span className="text-xs font-black text-white leading-none">{gameState?.players?.length || 0} / {gameState?.max_players || 4}</span>
             </div>
             <Users className="w-4 h-4 text-blue-400 opacity-50" />
          </div>
        </div>
      </header>

      {/* Reaction Chamber (Main Table) */}
      <div className="flex-1 relative flex items-center justify-center p-12">
        {/* Table Console Background */}
        <div className="absolute w-full max-w-5xl aspect-[16/10] bg-[#121216]/20 rounded-[80px] border border-white/5 shadow-[inset_0_0_100px_rgba(0,0,0,0.5)] pointer-events-none">
           {/* Decorative UI elements */}
           <div className="absolute top-8 left-1/2 -translate-x-1/2 flex gap-12 opacity-30">
              <div className="w-32 h-1 bg-gradient-to-r from-transparent via-blue-500 to-transparent"></div>
           </div>
        </div>

        {/* Players Radial Layout (Abstracted logic for more players) */}
        <div className="w-full max-w-5xl aspect-[16/10] relative flex items-center justify-center">
          
          {/* Reaction Core (Center Pile) */}
          <div className="relative z-20 flex flex-col md:flex-row items-center gap-12 lg:gap-20 scale-90 lg:scale-100">
            {/* Draw Pile */}
            <div className="group relative">
               <div className="absolute -inset-4 bg-blue-500/20 rounded-[40px] blur-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>
               <div 
                  onClick={isMyTurn ? handleDrawCard : undefined}
                  className={cn(
                    "w-36 h-52 bg-gradient-to-br from-slate-800 to-slate-900 rounded-[32px] border-2 border-white/5 flex flex-col items-center justify-center gap-4 shadow-2xl transition-all relative overflow-hidden",
                    isMyTurn ? "cursor-pointer hover:border-blue-500/50 hover:-translate-y-2 active:scale-95 group" : "grayscale opacity-40 cursor-not-allowed"
                  )}
                >
                  {/* Subtle data readout inside card pile */}
                  <div className="absolute top-4 left-4 right-4 flex justify-between items-center opacity-20">
                     <span className="text-[8px] font-mono">SEQ_DRAW</span>
                     <span className="text-[8px] font-mono">0x4F</span>
                  </div>
                  <div className="w-16 h-16 bg-white/5 rounded-[24px] flex items-center justify-center border border-white/10 group-hover:scale-110 group-hover:bg-blue-500/10 transition-all duration-500">
                    <RefreshCw className="w-8 h-8 text-slate-500 group-hover:text-blue-400 group-hover:rotate-180 transition-all duration-700" />
                  </div>
                  <div className="text-center">
                    <span className="text-[10px] font-black uppercase tracking-[0.3em] text-slate-500 group-hover:text-blue-500/70 transition-colors">Extraction</span>
                    <p className="text-[9px] text-slate-700 font-mono mt-1">Ready for input</p>
                  </div>
                </div>
            </div>

            {/* Discard / Current reaction */}
            <div className="relative flex flex-col items-center gap-6">
              {gameState?.last_card ? (
                 <div className="relative">
                    {/* Pulsing reaction glow */}
                    <div className="absolute -inset-12 bg-blue-600/10 rounded-full blur-3xl animate-pulse"></div>
                    <div className={cn(
                      "game-card scale-[1.4] pointer-events-none shadow-[0_30px_60px_rgba(0,0,0,0.8)] z-10", 
                      getCardStyle(gameState.last_card.card)
                    )}>
                        <div className="absolute top-2 left-2 text-[10px] uppercase font-black opacity-30 tracking-widest leading-none">Element</div>
                        <div className="text-4xl tracking-tighter font-black">{gameState.last_card.card.type}</div>
                        <div className="absolute bottom-2 right-2 text-[8px] font-mono opacity-40 uppercase tracking-tighter bg-black/40 px-1.5 py-0.5 rounded">
                          {gameState.last_card.card.effect || 'Passive'}
                        </div>
                    </div>
                 </div>
               ) : (
                 <div className="w-36 h-52 rounded-[32px] border-2 border-dashed border-white/5 flex flex-col items-center justify-center gap-2 opacity-30">
                    <Loader2 className="w-6 h-6 animate-spin text-slate-600" />
                    <p className="text-[10px] text-slate-600 font-black uppercase tracking-widest">Awaiting_Initial</p>
                 </div>
               )}
               
               {gameState?.last_card && (
                 <div className="bg-blue-600/10 backdrop-blur-3xl px-6 py-2.5 rounded-2xl border border-blue-500/20 text-xs font-black text-blue-400 shadow-[0_10px_30px_rgba(0,0,0,0.3)] flex items-center gap-3 animate-in fade-in slide-in-from-top-4">
                   <div className="w-2 h-2 rounded-full bg-blue-400 animate-pulse"></div>
                   <span className="uppercase tracking-widest text-[9px] text-slate-500">Reactant:</span>
                   <span>{gameState.last_card.substance}</span>
                 </div>
               )}
            </div>
          </div>

          {/* Player badges positions (radial) */}
          {gameState?.players?.map((player: any, index: number) => {
            const isActive = gameState.current_player === index
            const isLocal = player.user_id === user.id
            
            // Positioning for 4 players
            const positions = [
              "bottom-0 translate-y-1/2 translate-x-1/2 right-1/2", // Player (Me usually)
              "left-0 -translate-x-1/2 top-1/2 -translate-y-1/2",   // Left
              "top-0 -translate-y-1/2 translate-x-1/2 right-1/2",  // Top
              "right-0 translate-x-1/2 top-1/2 -translate-y-1/2",  // Right
            ]

            return (
              <div 
                key={player.user_id}
                className={cn(
                  "absolute transition-all duration-700 z-30",
                  positions[index % positions.length]
                )}
              >
                <div className={cn(
                  "p-2 rounded-[32px] transition-all duration-500 group/player border backdrop-blur-xl shadow-2xl flex items-center gap-4 min-w-[180px]",
                  isActive 
                    ? "bg-blue-600/10 border-blue-500 shadow-[0_0_40px_rgba(59,130,246,0.3)] scale-110" 
                    : "bg-black/60 border-white/5 hover:border-white/20"
                )}>
                  <div className="relative">
                    <div className="w-14 h-14 bg-slate-800 rounded-2xl flex items-center justify-center text-2xl border border-white/10 group-hover/player:scale-105 transition-transform overflow-hidden shadow-inner">
                      {player.avatar}
                    </div>
                    {isActive && (
                      <div className="absolute -top-1.5 -right-1.5 bg-blue-500 p-1.5 rounded-lg shadow-lg animate-bounce">
                        <Zap className="w-3 h-3 text-white fill-current" />
                      </div>
                    )}
                  </div>
                  <div className="flex flex-col pr-2">
                    <div className="flex items-center gap-2">
                       <span className="text-[10px] font-bold text-white uppercase tracking-tight truncate max-w-[80px]">{player.username}</span>
                       {isLocal && <span className="text-[8px] bg-blue-500/20 text-blue-400 px-1 rounded font-mono">YOU</span>}
                    </div>
                    <div className="flex items-center gap-3 mt-1.5">
                       <div className="flex items-center gap-1">
                          <div className="grid grid-cols-2 gap-0.5">
                             {[...Array(4)].map((_, i) => (<div key={i} className={cn("w-1 h-1 rounded-full", i < player.card_count ? "bg-blue-400" : "bg-slate-700")}></div>))}
                          </div>
                          <span className="text-[10px] font-mono text-slate-400">{player.card_count}</span>
                       </div>
                    </div>
                  </div>
                  
                  {/* Active turn progress ring (visual) */}
                  {isActive && (
                    <div className="absolute inset-0 rounded-[32px] border-2 border-blue-500/50 animate-pulse"></div>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      </div>

      {/* Hand / Deck Area */}
      <div className="h-[240px] bg-gradient-to-t from-blue-900/10 to-transparent relative mt-auto px-8 flex flex-col items-center">
        {/* Turn Tip */}
        <div className="h-0 relative w-full flex justify-center">
           {isMyTurn && (
             <div className="absolute -top-6 translate-y-[-100%] flex flex-col items-center gap-2 animate-in fade-in slide-in-from-bottom-2">
                <div className="bg-blue-600 px-8 py-2.5 rounded-full shadow-[0_15px_30px_rgba(37,99,235,0.4)] flex items-center gap-3 active:scale-95 transition-transform">
                  <Zap className="w-4 h-4 fill-current animate-pulse text-white" />
                  <span className="text-xs font-black uppercase tracking-[0.3em] text-white">Your_Turn_Active</span>
                </div>
                <div className="w-px h-12 bg-gradient-to-b from-blue-500 to-transparent opacity-50"></div>
             </div>
           )}
        </div>

        <div className="flex-1 w-full max-w-6xl flex justify-center items-end pb-8">
           <div className="flex flex-nowrap justify-center gap-x-2 lg:gap-x-4 px-12 h-[180px] w-full overflow-x-auto custom-scrollbar-hidden py-4 translate-y-4 hover:translate-y-0 transition-transform duration-500">
            {myData?.hand_cards?.length > 0 ? (
              myData.hand_cards.map((card: any, index: number) => (
                <div
                  key={index}
                  onClick={() => isMyTurn && handleCardClick(card)}
                  className={cn(
                    "game-card flex-shrink-0 cursor-pointer transition-all duration-300 transform-gpu origin-bottom",
                    getCardStyle(card),
                    selectedCard === card ? "selected -translate-y-8 scale-110 shadow-[0_20px_40px_rgba(0,0,0,0.5)] z-50 ring-2 ring-blue-500/50" : "hover:-translate-y-6 hover:rotate-2 hover:z-40",
                    !isMyTurn && "opacity-40 grayscale-[0.8] cursor-not-allowed pointer-events-none translate-y-12"
                  )}
                  style={{
                    // Spread effect
                    transform: `rotate(${(index - (myData.hand_cards.length - 1) / 2) * 2}deg)`
                  }}
                >
                  <div className="absolute top-2 left-2 text-[8px] font-black uppercase opacity-30 tracking-widest">Base_Elem</div>
                  <div className="text-3xl font-black tracking-tighter">{card.type}</div>
                  <div className="absolute bottom-2 right-2 text-[8px] font-mono opacity-40 uppercase tracking-tighter bg-black/40 px-1 py-0.5 rounded">
                    {card.effect || 'Passive'}
                  </div>
                </div>
              ))
            ) : (
              <div className="flex flex-col items-center justify-center opacity-10 pb-10">
                <FlaskConical className="w-20 h-20 mb-2" />
                <p className="font-black uppercase tracking-widest text-xs">Inventory_Empty</p>
              </div>
            )}
           </div>
        </div>
      </div>

      {/* Modern Substance Recombinator (Selection Modal) */}
      {selectedCard && availableSubstances.length > 0 && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-6">
          <div className="absolute inset-0 bg-black/90 backdrop-blur-xl animate-in fade-in" onClick={() => setSelectedCard(null)} />
          <div className="relative w-full max-w-2xl bg-[#0d0d10] border border-white/10 rounded-[48px] shadow-[0_0_100px_rgba(0,0,0,0.8)] overflow-hidden animate-in fade-in zoom-in slide-in-from-bottom-12 duration-500">
             {/* Modal Header Decor */}
             <div className="absolute top-0 left-0 w-full h-1.5 bg-gradient-to-r from-blue-600 via-purple-600 to-blue-600 opacity-50"></div>
             
             <div className="p-12">
               <div className="flex flex-col md:flex-row justify-between items-start gap-10 mb-12">
                 <div className="space-y-4">
                    <div className="inline-flex items-center gap-2 px-3 py-1 bg-blue-500/10 border border-blue-500/20 rounded-full">
                       <Zap className="w-3 h-3 text-blue-400" />
                       <span className="text-[10px] font-bold text-blue-400 uppercase tracking-widest">Reaction_Protocol_Active</span>
                    </div>
                    <h3 className="text-4xl font-black text-white tracking-tighter leading-none">
                      化学物质重组
                    </h3>
                    <p className="text-slate-500 max-w-sm font-medium leading-relaxed">
                      请从下方数据库中选择一个与 <span className="text-white font-black underline decoration-blue-500 underline-offset-4">{selectedCard.type}</span> 兼容的目标物质开始反应。
                    </p>
                 </div>
                 
                 <div className="relative group self-center md:self-auto">
                    <div className="absolute -inset-8 bg-blue-600/10 rounded-full blur-2xl group-hover:bg-blue-600/20 transition-all"></div>
                    <div className={cn("game-card scale-125 !cursor-default", getCardStyle(selectedCard))}>
                       <div className="text-3xl font-black tracking-tighter">{selectedCard.type}</div>
                    </div>
                 </div>
               </div>

               <div className="grid grid-cols-2 sm:grid-cols-3 gap-4 mb-12 max-h-[300px] overflow-y-auto pr-4 custom-scrollbar">
                  {availableSubstances.map((substance: string, index: number) => (
                    <button
                      key={index}
                      onClick={() => setSelectedSubstance(substance)}
                      className={cn(
                        "group relative p-6 rounded-3xl border transition-all flex flex-col items-center justify-center gap-3 overflow-hidden",
                        selectedSubstance === substance 
                          ? "bg-blue-600/10 border-blue-500 text-white shadow-[0_15px_35px_rgba(59,130,246,0.15)]" 
                          : "bg-white/[0.03] border-white/5 text-slate-500 hover:bg-white/[0.05] hover:border-white/10"
                      )}
                    >
                      <div className={cn(
                        "w-12 h-12 rounded-2xl flex items-center justify-center border transition-all duration-500",
                        selectedSubstance === substance ? "bg-blue-500/20 border-blue-500/30 rotate-12" : "bg-black/40 border-white/5 opacity-40 group-hover:rotate-12"
                      )}>
                        <FlaskConical className={cn("w-6 h-6", selectedSubstance === substance ? "text-blue-400" : "text-slate-600")} />
                      </div>
                      <span className="font-black tracking-widest text-[11px] uppercase truncate w-full text-center">{substance}</span>
                      {selectedSubstance === substance && <div className="absolute inset-0 bg-blue-500/5 animate-pulse"></div>}
                    </button>
                  ))}
               </div>

               <div className="flex gap-4">
                  <button 
                    onClick={() => {setSelectedCard(null); setSelectedSubstance(null);}} 
                    className="flex-1 h-16 bg-white/5 hover:bg-white/10 text-slate-500 hover:text-white font-black rounded-2xl transition-all uppercase tracking-widest text-[11px]"
                  >
                    终止操作
                  </button>
                  <button 
                    onClick={handlePlayCard}
                    disabled={!selectedSubstance}
                    className="flex-[2] h-16 bg-blue-600 hover:bg-blue-500 text-white font-black rounded-2xl transition-all shadow-[0_20px_40px_rgba(37,99,235,0.3)] disabled:opacity-50 disabled:grayscale flex items-center justify-center gap-3 group/confirm relative overflow-hidden"
                  >
                    <span className="uppercase tracking-[0.2em] text-xs">执行化合反应</span>
                    <ChevronRight className="w-5 h-5 group-hover/confirm:translate-x-1 transition-transform" />
                    <div className="absolute inset-0 w-full h-full bg-gradient-to-r from-transparent via-white/10 to-transparent -translate-x-full group-hover/confirm:animate-shimmer"></div>
                  </button>
               </div>
             </div>
          </div>
        </div>
      )
      }

      {/* Experimental Victory / Failure Protocol */}
      {gameState?.status === 'finished' && (
        <div className="fixed inset-0 z-[200] flex items-center justify-center p-6">
          <div className="absolute inset-0 bg-black/95 backdrop-blur-2xl animate-in fade-in duration-1000" />
          
          <div className="relative w-full max-w-xl bg-[#0a0a0c] border border-blue-500/20 rounded-[64px] p-16 flex flex-col items-center text-center overflow-hidden animate-in fade-in zoom-in spin-in-1 duration-1000">
             {/* Background Glow */}
             <div className="absolute -top-32 -left-32 w-64 h-64 bg-blue-500/20 rounded-full blur-[100px]"></div>
             <div className="absolute -bottom-32 -right-32 w-64 h-64 bg-purple-500/10 rounded-full blur-[100px]"></div>

             <div className="relative mb-12 transform-gpu">
                <div className="absolute inset-0 bg-blue-500/30 rounded-full blur-3xl animate-pulse"></div>
                <div className="w-32 h-32 bg-gradient-to-br from-blue-500 to-blue-700 rounded-[40px] flex items-center justify-center shadow-[0_20px_60px_rgba(59,130,246,0.5)] rotate-12">
                   <Trophy className="w-16 h-16 text-white" />
                </div>
                <div className="absolute -bottom-4 -right-4 w-12 h-12 bg-white rounded-2xl flex items-center justify-center shadow-2xl animate-bounce">
                   <Zap className="w-6 h-6 text-blue-600 fill-current" />
                </div>
             </div>

             <div className="space-y-4 mb-16 px-4">
                <div className="inline-flex items-center gap-2 px-4 py-1.5 bg-blue-500/10 border border-blue-500/20 rounded-full">
                   <span className="w-2 h-2 bg-blue-500 rounded-full animate-ping"></span>
                   <span className="text-[10px] font-black text-blue-400 uppercase tracking-widest font-mono">Mission_Success_Protocol</span>
                </div>
                {gameState.players?.find((p: any) => p.card_count === 0)?.user_id === user.id ? (
                  <>
                    <h2 className="text-6xl font-black text-white tracking-tighter leading-none">
                      实验大获成功
                    </h2>
                    <p className="text-slate-400 font-medium leading-relaxed max-w-sm mx-auto">
                      恭喜研究员！你已成功稳定了反应核心。所有元素均已合成为稳定态，此项成果将被载入实验室历史。
                    </p>
                  </>
                ) : (
                  <>
                    <h2 className="text-6xl font-black text-white tracking-tighter leading-none">
                      反应链终止
                    </h2>
                    <p className="text-slate-400 font-medium leading-relaxed max-w-sm mx-auto">
                      实验由 <span className="text-white font-black">{gameState.players?.find((p: any) => p.card_count === 0)?.username}</span> 成功收官。请在下一次化学反应中尝试更好的战术。
                    </p>
                  </>
                )}
             </div>

             <div className="w-full space-y-4">
                <button 
                  onClick={() => navigate('/')}
                  className="w-full h-18 bg-blue-600 hover:bg-blue-500 text-white font-black rounded-3xl transition-all shadow-[0_20px_40px_rgba(37,99,235,0.3)] hover:scale-105 active:scale-95 flex items-center justify-center gap-3 group relative overflow-hidden"
                >
                   <span className="uppercase tracking-[0.3em] text-sm">返回指挥大厅</span>
                   <ChevronRight className="w-6 h-6 group-hover:translate-x-1 transition-transform" />
                   <div className="absolute inset-0 w-full h-full bg-gradient-to-r from-transparent via-white/10 to-transparent -translate-x-full group-hover:animate-shimmer"></div>
                </button>
                <div className="flex items-center justify-center gap-2 text-[10px] font-mono text-slate-600 uppercase tracking-widest mt-6">
                   <Info className="w-3 h-3" />
                   <span>Experimental_Data_Logged_To_Archive</span>
                </div>
             </div>
          </div>
        </div>
      )}
    </div>
  )
}
