import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { gameAPI } from '../utils/api'
import websocket from '../utils/websocket'

export default function GameRoom({ user }) {
  const { id } = useParams()
  const navigate = useNavigate()
  const [gameState, setGameState] = useState(null)
  const [availableSubstances, setAvailableSubstances] = useState([])
  const [selectedCard, setSelectedCard] = useState(null)
  const [selectedSubstance, setSelectedSubstance] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadGameState()
    
    websocket.joinRoom(id)
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
      // 这里应该有获取游戏状态的API
      setLoading(false)
    } catch (error) {
      console.error('加载游戏状态失败:', error)
      setLoading(false)
    }
  }

  const handleGameUpdate = (message) => {
    if (message.data) {
      setGameState(message.data)
    }
  }

  const handleStartGame = async () => {
    try {
      await gameAPI.startGame(id)
      alert('游戏开始！')
    } catch (error) {
      alert(error.response?.data?.error || '开始游戏失败')
    }
  }

  const handleCardClick = async (card) => {
    setSelectedCard(card)
    
    try {
      const response = await gameAPI.getAvailableSubstances(id)
      setAvailableSubstances(response.data || [])
    } catch (error) {
      console.error('获取可用物质失败:', error)
    }
  }

  const handlePlayCard = async () => {
    if (!selectedCard || !selectedSubstance) {
      alert('请选择卡牌和物质')
      return
    }

    try {
      await gameAPI.playCard(id, selectedCard, selectedSubstance)
      setSelectedCard(null)
      setSelectedSubstance(null)
      setAvailableSubstances([])
    } catch (error) {
      alert(error.response?.data?.error || '出牌失败')
    }
  }

  const handleDrawCard = async () => {
    try {
      await gameAPI.drawCard(id)
    } catch (error) {
      alert(error.response?.data?.error || '摸牌失败')
    }
  }

  const handleLeaveRoom = async () => {
    try {
      await gameAPI.leaveRoom(id)
      navigate('/')
    } catch (error) {
      console.error('离开房间失败:', error)
      navigate('/')
    }
  }

  const getCardStyle = (card) => {
    if (card.effect === 'reverse') return 'noble'
    if (card.effect) return 'special'
    return 'element'
  }

  if (loading) {
    return (
      <div style={styles.loading}>
        <div className="loading"></div>
        <p>加载中...</p>
      </div>
    )
  }

  return (
    <div style={styles.container}>
      <header style={styles.header}>
        <h2>游戏房间: {id}</h2>
        <div>
          <button onClick={handleStartGame} className="btn btn-success" style={{ marginRight: '10px' }}>
            开始游戏
          </button>
          <button onClick={handleLeaveRoom} className="btn btn-danger">
            离开房间
          </button>
        </div>
      </header>

      <div style={styles.gameArea}>
        {/* 弃牌堆 - 显示上一张出的牌 */}
        <div className="card" style={styles.discardPile}>
          <h3>弃牌堆</h3>
          {gameState?.last_card ? (
            <div>
              <div className={`game-card ${getCardStyle(gameState.last_card.card)}`}>
                {gameState.last_card.card.type}
              </div>
              <p>物质: {gameState.last_card.substance}</p>
            </div>
          ) : (
            <p>还没有人出牌</p>
          )}
        </div>

        {/* 玩家手牌 */}
        <div className="card" style={styles.handCards}>
          <h3>你的手牌</h3>
          <div style={styles.cardList}>
            {gameState?.players?.find(p => p.user_id === user.id)?.hand_cards?.map((card, index) => (
              <div
                key={index}
                className={`game-card ${getCardStyle(card)}`}
                onClick={() => handleCardClick(card)}
                style={{
                  border: selectedCard?.type === card.type ? '3px solid yellow' : 'none',
                }}
              >
                {card.type}
              </div>
            )) || <p>暂无手牌</p>}
          </div>

          {selectedCard && (
            <div style={styles.substanceSelector}>
              <h4>选择要组成的物质:</h4>
              <div style={styles.substanceList}>
                {availableSubstances.map((substance, index) => (
                  <button
                    key={index}
                    className={`btn ${selectedSubstance === substance ? 'btn-primary' : 'btn-secondary'}`}
                    onClick={() => setSelectedSubstance(substance)}
                    style={{ margin: '5px' }}
                  >
                    {substance}
                  </button>
                ))}
              </div>
              <div style={{ marginTop: '10px' }}>
                <button onClick={handlePlayCard} className="btn btn-success">
                  出牌
                </button>
                <button onClick={() => {
                  setSelectedCard(null)
                  setSelectedSubstance(null)
                  setAvailableSubstances([])
                }} className="btn btn-secondary" style={{ marginLeft: '10px' }}>
                  取消
                </button>
              </div>
            </div>
          )}

          <button onClick={handleDrawCard} className="btn btn-primary" style={{ marginTop: '20px' }}>
            摸2张牌并跳过
          </button>
        </div>

        {/* 玩家列表 */}
        <div className="card" style={styles.playerList}>
          <h3>玩家列表</h3>
          {gameState?.players?.map((player, index) => (
            <div key={player.user_id} style={styles.playerCard}>
              <span style={{ fontSize: '24px' }}>{player.avatar}</span>
              <span style={{ fontWeight: 'bold' }}>{player.username}</span>
              <span>手牌: {player.card_count}张</span>
              {gameState.current_player === index && <span style={styles.currentPlayer}>轮到ta了</span>}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

const styles = {
  container: {
    minHeight: '100vh',
    padding: '20px',
  },
  header: {
    background: 'white',
    padding: '20px',
    borderRadius: '12px',
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '20px',
  },
  gameArea: {
    display: 'grid',
    gridTemplateColumns: '1fr 2fr 1fr',
    gap: '20px',
  },
  discardPile: {
    textAlign: 'center',
  },
  handCards: {
    minHeight: '300px',
  },
  cardList: {
    display: 'flex',
    flexWrap: 'wrap',
    justifyContent: 'center',
    gap: '10px',
    marginTop: '20px',
  },
  playerList: {},
  playerCard: {
    display: 'flex',
    alignItems: 'center',
    gap: '10px',
    padding: '10px',
    background: '#f8f9fa',
    borderRadius: '8px',
    marginBottom: '10px',
  },
  currentPlayer: {
    background: '#28a745',
    color: 'white',
    padding: '5px 10px',
    borderRadius: '20px',
    fontSize: '12px',
    marginLeft: 'auto',
  },
  substanceSelector: {
    marginTop: '20px',
    padding: '15px',
    background: '#f8f9fa',
    borderRadius: '8px',
  },
  substanceList: {
    display: 'flex',
    flexWrap: 'wrap',
    marginTop: '10px',
  },
  loading: {
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'center',
    alignItems: 'center',
    height: '100vh',
    gap: '20px',
  },
}
