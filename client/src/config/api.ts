// API配置 - 支持移动设备访问
const getApiBaseUrl = (): string => {
  // 如果是开发环境且指定了API_URL
  if (process.env.REACT_APP_API_URL) {
    return process.env.REACT_APP_API_URL;
  }
  
  // 关键修复：如果是 HTTPS 环境，且没有指定 API URL，默认使用空字符串
  // 这意味着所有 API 请求将使用相对路径 (例如 /api/...)
  // 这样可以避免 Mixed Content 错误 (HTTPS 页面请求 HTTP 资源)
  // 注意：这要求服务器配置了反向代理 (Nginx/Apache) 将 /api 转发到后端端口
  if (window.location.protocol === 'https:') {
    return '';
  }

  // 在生产环境或移动设备访问时，使用当前主机
  // 如果前端和后端在同一台服务器上，使用相对路径
  if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    return 'http://localhost:4001';
  }
  
  // 移动设备访问时，使用实际的服务器地址
  // 假设后端运行在4001端口
  return `http://${window.location.hostname}:4001`;
};

export const API_BASE_URL = getApiBaseUrl();
export const SOCKET_URL = API_BASE_URL;

// API端点
export const API_ENDPOINTS = {
  // 游戏相关
  createGame: `${API_BASE_URL}/api/game/create`,
  joinGame: `${API_BASE_URL}/api/game/join`,
  startGame: (roomCode: string) => `${API_BASE_URL}/api/game/${roomCode}/start`,
  gameInfo: (roomCode: string) => `${API_BASE_URL}/api/game/${roomCode}/info`,
  gameQrcode: (roomCode: string) => `${API_BASE_URL}/api/game/${roomCode}/qrcode`,
  
  // 房间相关
  rooms: `${API_BASE_URL}/api/rooms`,
  
  // 玩家相关
  playerSession: (name: string) => `${API_BASE_URL}/api/player/${encodeURIComponent(name)}/session`,
  
  // 配置相关
  config: `${API_BASE_URL}/api/config`,
  configRefresh: `${API_BASE_URL}/api/config/refresh`,
  
  // 初始化设置
  checkSetup: `${API_BASE_URL}/api/setup/check`,
  setup: `${API_BASE_URL}/api/setup`,
  verifyPassword: `${API_BASE_URL}/api/verify-password`,
  
  // 化合物相关
  compounds: `${API_BASE_URL}/api/compounds`,
};

export default API_ENDPOINTS;
