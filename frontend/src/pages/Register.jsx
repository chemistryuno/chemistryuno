import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { authAPI } from '../utils/api'

export default function Register() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')

    if (password !== confirmPassword) {
      setError('两次输入的密码不一致')
      return
    }

    setLoading(true)

    try {
      await authAPI.register(username, password)
      alert('注册成功！请登录')
      navigate('/login')
    } catch (err) {
      setError(err.response?.data?.error || '注册失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={styles.container}>
      <div className="card" style={styles.card}>
        <h1 style={styles.title}>🧪 化学UNO</h1>
        <h2 style={styles.subtitle}>注册新账号</h2>
        
        {error && <div style={styles.error}>{error}</div>}
        
        <form onSubmit={handleSubmit}>
          <div className="input-group">
            <label>用户名</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              minLength={3}
              maxLength={20}
              placeholder="3-20个字符"
            />
          </div>

          <div className="input-group">
            <label>密码</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={6}
              placeholder="至少6个字符"
            />
          </div>

          <div className="input-group">
            <label>确认密码</label>
            <input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
              minLength={6}
              placeholder="再次输入密码"
            />
          </div>

          <button 
            type="submit" 
            className="btn btn-primary" 
            style={styles.button}
            disabled={loading}
          >
            {loading ? <span className="loading"></span> : '注册'}
          </button>
        </form>

        <div style={styles.footer}>
          已有账号？ <Link to="/login" style={styles.link}>立即登录</Link>
        </div>
      </div>
    </div>
  )
}

const styles = {
  container: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    minHeight: '100vh',
    padding: '20px',
  },
  card: {
    maxWidth: '400px',
    width: '100%',
  },
  title: {
    textAlign: 'center',
    fontSize: '48px',
    marginBottom: '10px',
    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    WebkitBackgroundClip: 'text',
    WebkitTextFillColor: 'transparent',
  },
  subtitle: {
    textAlign: 'center',
    marginBottom: '30px',
    color: '#666',
  },
  button: {
    width: '100%',
    marginTop: '10px',
  },
  footer: {
    textAlign: 'center',
    marginTop: '20px',
    color: '#666',
  },
  link: {
    color: '#667eea',
    textDecoration: 'none',
    fontWeight: 'bold',
  },
  error: {
    background: '#ffe0e0',
    color: '#c00',
    padding: '10px',
    borderRadius: '8px',
    marginBottom: '15px',
    textAlign: 'center',
  },
}
