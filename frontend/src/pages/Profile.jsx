import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { authAPI } from '../utils/api'

export default function Profile({ user, onLogout }) {
  const navigate = useNavigate()
  const [showChangePassword, setShowChangePassword] = useState(false)
  const [showChangeAvatar, setShowChangeAvatar] = useState(false)
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [selectedAvatar, setSelectedAvatar] = useState(user.avatar)
  const [loading, setLoading] = useState(false)

  const avatarOptions = ['🧪', '⚗️', '🔬', '🧬', '⚛️', '🎓', '👨‍🔬', '👩‍🔬', '🦠', '💊']

  const handleChangePassword = async (e) => {
    e.preventDefault()

    if (newPassword !== confirmPassword) {
      alert('两次输入的密码不一致')
      return
    }

    setLoading(true)
    try {
      await authAPI.changePassword(oldPassword, newPassword)
      alert('密码修改成功！请重新登录')
      onLogout()
      navigate('/login')
    } catch (error) {
      alert(error.response?.data?.error || '修改密码失败')
    } finally {
      setLoading(false)
    }
  }

  const handleChangeAvatar = async () => {
    setLoading(true)
    try {
      await authAPI.updateAvatar(selectedAvatar)
      alert('头像更新成功！')
      
      // 更新本地用户信息
      const updatedUser = { ...user, avatar: selectedAvatar }
      localStorage.setItem('user', JSON.stringify(updatedUser))
      window.location.reload()
    } catch (error) {
      alert(error.response?.data?.error || '更新头像失败')
    } finally {
      setLoading(false)
      setShowChangeAvatar(false)
    }
  }

  const handleDeleteAccount = async () => {
    if (!window.confirm('确定要注销账号吗？此操作不可恢复！')) {
      return
    }

    if (!window.confirm('再次确认：真的要删除账号吗？')) {
      return
    }

    try {
      await authAPI.deleteAccount()
      alert('账号已注销')
      onLogout()
      navigate('/login')
    } catch (error) {
      alert(error.response?.data?.error || '注销账号失败')
    }
  }

  return (
    <div style={styles.container}>
      <div className="card" style={styles.card}>
        <button onClick={() => navigate('/')} className="btn btn-secondary" style={styles.backBtn}>
          ← 返回大厅
        </button>

        <div style={styles.profile}>
          <div style={styles.avatarSection}>
            <div style={styles.avatar}>{user.avatar}</div>
            <button 
              onClick={() => setShowChangeAvatar(true)} 
              className="btn btn-primary"
            >
              更换头像
            </button>
          </div>

          <div style={styles.userDetails}>
            <h2>{user.username}</h2>
            {user.is_admin && <span style={styles.adminBadge}>👑 管理员</span>}
            <p style={{ color: '#666', marginTop: '10px' }}>
              ID: {user.id}
            </p>
          </div>
        </div>

        <div style={styles.actions}>
          <button 
            onClick={() => setShowChangePassword(true)} 
            className="btn btn-primary"
          >
            修改密码
          </button>
          <button 
            onClick={handleDeleteAccount} 
            className="btn btn-danger"
          >
            注销账号
          </button>
        </div>
      </div>

      {/* 修改密码模态框 */}
      {showChangePassword && (
        <div style={styles.modal}>
          <div className="card" style={styles.modalContent}>
            <h2>修改密码</h2>
            <form onSubmit={handleChangePassword}>
              <div className="input-group">
                <label>旧密码</label>
                <input
                  type="password"
                  value={oldPassword}
                  onChange={(e) => setOldPassword(e.target.value)}
                  required
                />
              </div>
              <div className="input-group">
                <label>新密码</label>
                <input
                  type="password"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  required
                  minLength={6}
                />
              </div>
              <div className="input-group">
                <label>确认新密码</label>
                <input
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  required
                  minLength={6}
                />
              </div>
              <div style={styles.modalButtons}>
                <button 
                  type="button" 
                  onClick={() => setShowChangePassword(false)} 
                  className="btn btn-secondary"
                >
                  取消
                </button>
                <button type="submit" className="btn btn-primary" disabled={loading}>
                  {loading ? <span className="loading"></span> : '确认修改'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* 更换头像模态框 */}
      {showChangeAvatar && (
        <div style={styles.modal}>
          <div className="card" style={styles.modalContent}>
            <h2>选择头像</h2>
            <div style={styles.avatarGrid}>
              {avatarOptions.map(avatar => (
                <div
                  key={avatar}
                  style={{
                    ...styles.avatarOption,
                    border: selectedAvatar === avatar ? '3px solid #667eea' : '2px solid #e0e0e0',
                  }}
                  onClick={() => setSelectedAvatar(avatar)}
                >
                  {avatar}
                </div>
              ))}
            </div>
            <div style={styles.modalButtons}>
              <button 
                onClick={() => setShowChangeAvatar(false)} 
                className="btn btn-secondary"
              >
                取消
              </button>
              <button 
                onClick={handleChangeAvatar} 
                className="btn btn-primary"
                disabled={loading}
              >
                {loading ? <span className="loading"></span> : '确认更换'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

const styles = {
  container: {
    minHeight: '100vh',
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    padding: '20px',
  },
  card: {
    maxWidth: '600px',
    width: '100%',
  },
  backBtn: {
    marginBottom: '20px',
  },
  profile: {
    display: 'flex',
    alignItems: 'center',
    gap: '30px',
    padding: '20px 0',
    borderBottom: '2px solid #e0e0e0',
  },
  avatarSection: {
    textAlign: 'center',
  },
  avatar: {
    fontSize: '80px',
    marginBottom: '15px',
  },
  userDetails: {
    flex: 1,
  },
  adminBadge: {
    background: '#ffd700',
    padding: '5px 15px',
    borderRadius: '20px',
    fontSize: '14px',
    fontWeight: 'bold',
    marginLeft: '10px',
  },
  actions: {
    display: 'flex',
    gap: '15px',
    marginTop: '30px',
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
  modalButtons: {
    display: 'flex',
    gap: '10px',
    marginTop: '20px',
  },
  avatarGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(5, 1fr)',
    gap: '10px',
    margin: '20px 0',
  },
  avatarOption: {
    fontSize: '40px',
    padding: '15px',
    textAlign: 'center',
    borderRadius: '8px',
    cursor: 'pointer',
    transition: 'all 0.3s',
  },
}
