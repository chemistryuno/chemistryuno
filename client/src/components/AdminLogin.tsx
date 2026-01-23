import React, { useState } from 'react';
import axios from 'axios';
import API_ENDPOINTS from '../config/api';
import './AdminLogin.css';

interface AdminLoginProps {
  onSuccess: () => void;
  expectedPassword?: string; // 已弃用，改用服务器端验证
}

const AdminLogin: React.FC<AdminLoginProps> = ({ onSuccess }) => {
  const [password, setPassword] = useState<string>('');
  const [error, setError] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(false);

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>): Promise<void> => {
    e.preventDefault();
    setError('');
    setLoading(true);
    
    try {
      // 使用服务器端验证替代本地验证
      // 解决因构建时环境变量导致的其他设备无法登录的问题
      const response = await axios.post(API_ENDPOINTS.verifyPassword, { password });
      
      if (response.data.success) {
        onSuccess();
      } else {
        setError(response.data.message || '密码错误');
      }
    } catch (err: any) {
      console.error('Login verify error:', err);
      // 如果后端没有此接口（旧版本兼容），尝试本地验证
      if (err.response && err.response.status === 404) {
         const storedPassword = localStorage.getItem('adminPassword') || '';
         if (password === storedPassword) {
            onSuccess();
            return;
         }
      }
      setError('登录验证失败：请检查网络连接');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="admin-login-page">
      <div className="admin-login-card">
        <h1>管理后台登录</h1>
        <p className="login-subtitle">请输入管理员密码以进入配置后台</p>
        <form onSubmit={handleSubmit}>
          <input
            type="password"
            placeholder="管理员密码"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            disabled={loading}
            autoFocus
          />
          {error && <div className="login-error">{error}</div>}
          <button type="submit" disabled={loading}>
            {loading ? '验证中...' : '进入后台'}
          </button>
        </form>
      </div>
    </div>
  );
};

export default AdminLogin;
