import axios, { AxiosInstance, InternalAxiosRequestConfig, AxiosResponse } from 'axios'

const api: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

// 请求拦截器
api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response: AxiosResponse) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// 认证API
export const authAPI = {
  register: (username: string, password: string) => 
    api.post('/auth/register', { username, password }),
  login: (username: string, password: string) => 
    api.post('/auth/login', { username, password }),
  getUserInfo: () => 
    api.get('/user/info'),
  changePassword: (oldPassword: string, newPassword: string) => 
    api.put('/user/password', { old_password: oldPassword, new_password: newPassword }),
  updateAvatar: (avatar: string) => 
    api.put('/user/avatar', { avatar }),
  deleteAccount: () => 
    api.delete('/user/account'),
}

// 游戏API
export const gameAPI = {
  getRooms: () => 
    api.get('/rooms'),
  createRoom: (name: string, maxPlayers: number, deckID: number) => 
    api.post('/rooms', { name, max_players: maxPlayers, deck_id: deckID }),
  getRoomState: (roomId: string) => 
    api.get(`/rooms/${roomId}`),
  joinRoom: (roomId: string) => 
    api.post(`/rooms/${roomId}/join`),
  leaveRoom: (roomId: string) => 
    api.post(`/rooms/${roomId}/leave`),
  startGame: (roomId: string) => 
    api.post(`/rooms/${roomId}/start`),
  playCard: (roomId: string, card: any, substance: string) => 
    api.post(`/rooms/${roomId}/play`, { card, substance }),
  drawCard: (roomId: string) => 
    api.post(`/rooms/${roomId}/draw`),
  getAvailableSubstances: (roomId: string) => 
    api.get(`/rooms/${roomId}/substances`),
}

// 管理员API
export const adminAPI = {
  getAllUsers: () => 
    api.get('/admin/users'),
  deleteUser: (userId: string) => 
    api.delete(`/admin/users/${userId}`),
  getGlobalDeckConfig: () => 
    api.get('/admin/deck-config'),
  updateGlobalDeckConfig: (name: string, cards: any[]) => 
    api.put('/admin/deck-config', { name, cards }),
  getGameHistory: () => 
    api.get('/admin/game-history'),
}

export default api
