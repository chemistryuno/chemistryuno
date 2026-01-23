import React, { useState } from 'react';
import axios from 'axios';
import './Setup.css';
import API_ENDPOINTS from '../config/api';

interface SetupProps {
  onComplete: () => void;
}

const Setup: React.FC<SetupProps> = ({ onComplete }) => {
  const [adminPassword, setAdminPassword] = useState<string>('');
  const [confirmPassword, setConfirmPassword] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>('');
  const [step, setStep] = useState<number>(1);

  const validatePassword = (): boolean => {
    if (!adminPassword) {
      setError('请输入管理员密码');
      return false;
    }
    if (adminPassword.length < 6) {
      setError('密码长度至少6位');
      return false;
    }
    if (adminPassword !== confirmPassword) {
      setError('两次输入的密码不一致');
      return false;
    }
    return true;
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>): Promise<void> => {
    e.preventDefault();
    setError('');

    if (!validatePassword()) {
      return;
    }

    setLoading(true);

    try {
      const response = await axios.post(API_ENDPOINTS.setup, {
        adminPassword: adminPassword
      });

      if (response.data && response.data.success) {
        // 重构：只在localStorage保存登录状态，不保存密码明文
        // 密码安全由服务器端验证保障
        localStorage.removeItem('adminPassword'); // 清理旧逻辑残留
        setStep(2);
        // 提示用户需要重启服务
        setTimeout(() => {
          alert('设置已保存！\n\n请重启服务以使密码生效：\n1. 停止当前服务（Ctrl+C）\n2. 运行: ./start-integrated.sh (Linux) 或 start-integrated.bat (Windows)');
          window.location.href = '/';
        }, 2000);
      } else {
        // 如果返回了 200 但不是预期的 JSON 格式（例如返回了 index.html），说明 API 请求没有正确到达后端
        console.error('Unexpected response:', response);
        const isHtml = typeof response.data === 'string' && response.data.includes('<!DOCTYPE html>');
        
        if (isHtml) {
          setError('配置错误：API 请求被拦截。请确保 Nginx 已配置及 /api 反向代理规则生效。');
        } else {
          setError('服务器响应异常，请检查后端日志');
        }
        setLoading(false);
      }
    } catch (err: any) {
      console.error('Setup error:', err);
      const errorMessage = err.response?.data?.error || err.response?.data?.details || err.message || '保存失败，请检查网络或服务器日志';
      setError(`保存失败: ${errorMessage}`);
      setLoading(false);
    }
  };

  if (step === 2) {
    return (
      <div className="setup-container">
        <div className="setup-card success">
          <div className="success-icon">✓</div>
          <h1>设置完成！</h1>
          <p>管理员密码已保存到配置文件</p>
          <p className="redirect-hint" style={{ color: '#ff6b6b', fontWeight: 'bold' }}>
            ⚠️ 请重启服务以使密码生效
          </p>
          <div className="loading-spinner"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="setup-container">
      <div className="setup-card">
        <div className="setup-header">
          <h1>⚗️ 化学UNO 初始化设置</h1>
          <p className="setup-subtitle">欢迎！请设置管理员密码以继续</p>
        </div>

        <div className="setup-info">
          <div className="info-item">
            <span className="info-icon">🔐</span>
            <div>
              <strong>管理员密码</strong>
              <p>用于访问 /admin 管理面板，可以修改游戏配置和化学反应规则</p>
            </div>
          </div>
          <div className="info-item">
            <span className="info-icon">⚙️</span>
            <div>
              <strong>安全提示</strong>
              <p>请设置一个强密码，建议至少8位，包含字母和数字</p>
            </div>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="setup-form">
          <div className="form-group">
            <label htmlFor="adminPassword">管理员密码</label>
            <input
              type="password"
              id="adminPassword"
              value={adminPassword}
              onChange={(e) => setAdminPassword(e.target.value)}
              placeholder="请输入管理员密码（至少6位）"
              disabled={loading}
              autoFocus
            />
          </div>

          <div className="form-group">
            <label htmlFor="confirmPassword">确认密码</label>
            <input
              type="password"
              id="confirmPassword"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder="请再次输入密码"
              disabled={loading}
            />
          </div>

          {error && <div className="error-message">{error}</div>}

          <button type="submit" className="setup-button" disabled={loading}>
            {loading ? '正在保存...' : '完成设置'}
          </button>
        </form>

        <div className="setup-footer">
          <p>💡 提示：设置完成后可以在管理面板中修改密码</p>
        </div>
      </div>
    </div>
  );
};

export default Setup;
