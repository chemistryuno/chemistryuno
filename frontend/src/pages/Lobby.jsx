import { useState, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { gameAPI } from '../utils/api'
import websocket from '../utils/websocket'

export default function Lobby({ user, onLogout }) {
  const [rooms, setRooms] = useState([])
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [roomName, setRoomName] = useState('')
  const [maxPlayers, setMaxPlayers] = useState(4)
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    loadRooms()
    websocket.connect()

    const interval = setInterval(loadRooms, 3000)
    return () => {
      clearInterval(interval)
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

  const handleCreateRoom = async (e) => {
    e.preventDefault()
    setLoading(true)

    try {
      const response = await gameAPI.createRoom(roomName, maxPlayers, 0)
      const room = response.data
      navigate(`/room/${room.id}`)
    } catch (error) {
      alert(error.response?.data?.error || '创建房间失败')
    } finally {
      setLoading(false)
    }
  }

  const handleJoinRoom = async (roomId) => {
    try {
      await gameAPI.joinRoom(roomId)
      navigate(`/room/${roomId}`)
    } catch (error) {
      alert(error.response?.data?.error || '加入房间失败')
    }
  }

  return (
    <div style={styles.container}>
      <header style={styles.header}>
        <h1 style={styles.title}>🧪 化学UNO 大厅</h1>
        <div style={styles.userInfo}>
          <span style={styles.avatar}>{user.avatar}</span>
          <span style={styles.username}>{user.username}</span>
          {user.is_admin && <span style={styles.adminBadge}>👑管理员</span>}
          <Link to="/profile" className="btn btn-secondary" style={styles.btn}>个人中心</Link>
          {user.is_admin && (
            <Link to="/admin" className="btn btn-primary" style={styles.btn}>管理面板</Link>
          )}
          <button onClick={onLogout} className="btn btn-danger" style={styles.btn}>退出登录</button>
        </div>
      </header>

      <div style={styles.content}>
        <div style={styles.topBar}>
          <h2>游戏房间列表</h2>
          <button 
            onClick={() => setShowCreateModal(true)} 
            className="btn btn-success"
          >
            ➕ 创建房间
          </button>
        </div>

        <div style={styles.roomList}>
          {rooms.length === 0 ? (
            <div style={styles.emptyState}>
              <p>暂无游戏房间，创建一个开始游戏吧！</p>
            </div>
          ) : (
            rooms.map(room => (
              <div key={room.id} className="card" style={styles.roomCard}>
                <div style={styles.roomHeader}>
                  <h3>{room.name}</h3>
                  <span style={styles.roomStatus(room.status)}>
                    {room.status === 'waiting' ? '等待中' : room.status === 'playing' ? '游戏中' : '已结束'}
                  </span>
                </div>
                <div style={styles.roomInfo}>
                  <p>👥 玩家: {room.players?.length || 0} / {room.max_players}</p>
                  <p>🎮 房主ID: {room.host_id}</p>
                </div>
                {room.status === 'waiting' && (
                  <button 
                    onClick={() => handleJoinRoom(room.id)} 
                    className="btn btn-primary"
                    style={{ width: '100%', marginTop: '10px' }}
                  >
                    加入房间
                  </button>
                )}
              </div>
            ))
          )}
        </div>
      </div>

      {showCreateModal && (
        <div style={styles.modal}>
          <div className="card" style={styles.modalContent}>
            <h2>创建新房间</h2>
            <form onSubmit={handleCreateRoom}>
              <div className="input-group">
                <label>房间名称</label>
                <input
                  type="text"
                  value={roomName}
                  onChange={(e) => setRoomName(e.target.value)}
                  required
                  placeholder="输入房间名称"
                />
              </div>

              <div className="input-group">
                <label>最大玩家数</label>
                <select 
                  value={maxPlayers} 
                  onChange={(e) => setMaxPlayers(Number(e.target.value))}
                  style={styles.select}
                >
                  {[2, 3, 4, 5, 6, 7, 8].map(num => (
                    <option key={num} value={num}>{num}人</option>
                  ))}
                </select>
              </div>

              <div style={styles.modalButtons}>
                <button 
                  type="button" 
                  onClick={() => setShowCreateModal(false)} 
                  className="btn btn-secondary"
                >
                  取消
                </button>
                <button 
                  type="submit" 
                  className="btn btn-success"
                  disabled={loading}
                >
                  {loading ? <span className="loading"></span> : '创建'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}

const styles = {
  container: {
    minHeight: '100vh',
  },
  header: {
    background: 'white',
    padding: '20px',
    boxShadow: '0 2px 10px rgba(0,0,0,0.1)',
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    flexWrap: 'wrap',
  },
  title: {
    fontSize: '32px',
    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    WebkitBackgroundClip: 'text',
    WebkitTextFillColor: 'transparent',
  },
  userInfo: {
    display: 'flex',
    alignItems: 'center',
    gap: '15px',
  },
  avatar: {
    fontSize: '32px',
  },
  username: {
    fontWeight: 'bold',
    fontSize: '18px',
  },
  adminBadge: {
    background: '#ffd700',
    padding: '5px 10px',
    borderRadius: '20px',
    fontSize: '14px',
    fontWeight: 'bold',
  },
  btn: {
    marginLeft: '5px',
  },
  content: {
    maxWidth: '1200px',
    margin: '0 auto',
    padding: '20px',
  },
  topBar: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '20px',
  },
  roomList: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))',
    gap: '20px',
  },
  roomCard: {
    cursor: 'pointer',
    transition: 'transform 0.3s',
  },
  roomHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '15px',
  },
  roomStatus: (status) => ({
    padding: '5px 15px',
    borderRadius: '20px',
    fontSize: '14px',
    fontWeight: 'bold',
    background: status === 'waiting' ? '#28a745' : status === 'playing' ? '#ffc107' : '#6c757d',
    color: 'white',
  }),
  roomInfo: {
    color: '#666',
  },
  emptyState: {
    gridColumn: '1 / -1',
    textAlign: 'center',
    padding: '50px',
    color: '#666',
  },
  modal: {
    position: 'fixed',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    background: 'rgba(0,0,0,0.5)',
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    zIndex: 1000,
  },
  modalContent: {
    maxWidth: '500px',
    width: '90%',
  },
  select: {
    width: '100%',
    padding: '12px',
    border: '2px solid #e0e0e0',
    borderRadius: '8px',
    fontSize: '16px',
  },
  modalButtons: {
    display: 'flex',
    gap: '10px',
    marginTop: '20px',
  },
}
