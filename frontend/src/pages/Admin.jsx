import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { adminAPI } from '../utils/api'

export default function Admin({ user }) {
  const navigate = useNavigate()
  const [users, setUsers] = useState([])
  const [gameHistory, setGameHistory] = useState([])
  const [deckConfig, setDeckConfig] = useState(null)
  const [editingDeck, setEditingDeck] = useState(false)
  const [activeTab, setActiveTab] = useState('users')

  useEffect(() => {
    loadData()
  }, [activeTab])

  const loadData = async () => {
    try {
      if (activeTab === 'users') {
        const response = await adminAPI.getAllUsers()
        setUsers(response.data || [])
      } else if (activeTab === 'history') {
        const response = await adminAPI.getGameHistory()
        setGameHistory(response.data || [])
      } else if (activeTab === 'deck') {
        const response = await adminAPI.getGlobalDeckConfig()
        setDeckConfig(response.data)
      }
    } catch (error) {
      console.error('加载数据失败:', error)
    }
  }

  const handleDeleteUser = async (userId) => {
    if (!window.confirm('确定要删除此用户吗？')) return

    try {
      await adminAPI.deleteUser(userId)
      alert('用户已删除')
      loadData()
    } catch (error) {
      alert(error.response?.data?.error || '删除用户失败')
    }
  }

  const handleUpdateDeck = async () => {
    try {
      await adminAPI.updateGlobalDeckConfig(deckConfig.name, deckConfig.cards)
      alert('全局牌组配置已更新')
      setEditingDeck(false)
    } catch (error) {
      alert(error.response?.data?.error || '更新失败')
    }
  }

  const handleCardCountChange = (cardType, value) => {
    setDeckConfig({
      ...deckConfig,
      cards: {
        ...deckConfig.cards,
        [cardType]: parseInt(value) || 0,
      },
    })
  }

  return (
    <div style={styles.container}>
      <div className="card" style={styles.card}>
        <div style={styles.header}>
          <h1>👑 管理面板</h1>
          <button onClick={() => navigate('/')} className="btn btn-secondary">
            返回大厅
          </button>
        </div>

        <div style={styles.tabs}>
          <button
            className={`btn ${activeTab === 'users' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('users')}
          >
            用户管理
          </button>
          <button
            className={`btn ${activeTab === 'deck' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('deck')}
          >
            牌组配置
          </button>
          <button
            className={`btn ${activeTab === 'history' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('history')}
          >
            游戏历史
          </button>
        </div>

        {/* 用户管理 */}
        {activeTab === 'users' && (
          <div style={styles.content}>
            <h2>用户列表</h2>
            <table style={styles.table}>
              <thead>
                <tr>
                  <th>ID</th>
                  <th>头像</th>
                  <th>用户名</th>
                  <th>管理员</th>
                  <th>注册时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {users.map(u => (
                  <tr key={u.id}>
                    <td>{u.id}</td>
                    <td style={{ fontSize: '24px' }}>{u.avatar}</td>
                    <td>{u.username}</td>
                    <td>{u.is_admin ? '✅' : '❌'}</td>
                    <td>{new Date(u.created_at).toLocaleDateString()}</td>
                    <td>
                      {!u.is_admin && (
                        <button
                          onClick={() => handleDeleteUser(u.id)}
                          className="btn btn-danger"
                          style={{ padding: '5px 10px', fontSize: '14px' }}
                        >
                          删除
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* 牌组配置 */}
        {activeTab === 'deck' && deckConfig && (
          <div style={styles.content}>
            <div style={styles.deckHeader}>
              <h2>全局牌组配置</h2>
              <button
                onClick={() => setEditingDeck(!editingDeck)}
                className="btn btn-primary"
              >
                {editingDeck ? '取消编辑' : '编辑配置'}
              </button>
            </div>

            <div style={styles.deckGrid}>
              {Object.entries(deckConfig.cards).map(([cardType, count]) => (
                <div key={cardType} style={styles.deckItem}>
                  <strong>{cardType}</strong>
                  {editingDeck ? (
                    <input
                      type="number"
                      value={count}
                      onChange={(e) => handleCardCountChange(cardType, e.target.value)}
                      min="0"
                      max="20"
                      style={styles.deckInput}
                    />
                  ) : (
                    <span>×{count}</span>
                  )}
                </div>
              ))}
            </div>

            {editingDeck && (
              <button
                onClick={handleUpdateDeck}
                className="btn btn-success"
                style={{ marginTop: '20px' }}
              >
                保存配置
              </button>
            )}
          </div>
        )}

        {/* 游戏历史 */}
        {activeTab === 'history' && (
          <div style={styles.content}>
            <h2>游戏历史记录</h2>
            <table style={styles.table}>
              <thead>
                <tr>
                  <th>房间ID</th>
                  <th>获胜者</th>
                  <th>开始时间</th>
                  <th>结束时间</th>
                </tr>
              </thead>
              <tbody>
                {gameHistory.map(game => (
                  <tr key={game.id}>
                    <td>{game.room_id}</td>
                    <td>{game.winner_name || game.winner_id}</td>
                    <td>{game.started_at ? new Date(game.started_at).toLocaleString() : '-'}</td>
                    <td>{game.finished_at ? new Date(game.finished_at).toLocaleString() : '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}

const styles = {
  container: {
    minHeight: '100vh',
    padding: '20px',
  },
  card: {
    maxWidth: '1200px',
    margin: '0 auto',
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '30px',
  },
  tabs: {
    display: 'flex',
    gap: '10px',
    marginBottom: '30px',
  },
  content: {
    marginTop: '20px',
  },
  table: {
    width: '100%',
    borderCollapse: 'collapse',
    marginTop: '20px',
  },
  deckHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '20px',
  },
  deckGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fill, minmax(120px, 1fr))',
    gap: '15px',
  },
  deckItem: {
    background: '#f8f9fa',
    padding: '15px',
    borderRadius: '8px',
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    gap: '10px',
  },
  deckInput: {
    width: '60px',
    padding: '5px',
    textAlign: 'center',
    border: '2px solid #667eea',
    borderRadius: '4px',
  },
}

// CSS for table
const tableStyles = `
  table th, table td {
    padding: 12px;
    text-align: left;
    border-bottom: 1px solid #e0e0e0;
  }
  table th {
    background: #f8f9fa;
    font-weight: bold;
  }
  table tr:hover {
    background: #f8f9fa;
  }
`

// 注入样式
if (typeof document !== 'undefined') {
  const style = document.createElement('style')
  style.textContent = tableStyles
  document.head.appendChild(style)
}
