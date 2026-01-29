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
  submitFeedback: (content: string, type: string) =>
    api.post('/feedback', { content, type }),
  
  // 2FA相关
  setup2FA: () => api.post('/user/2fa/setup'),
  enable2FA: (code: string) => api.post('/user/2fa/enable', { code }),
  disable2FA: (code: string) => api.post('/user/2fa/disable', { code }),
  verify2FALogin: (uid: number, code: string) => api.post('/auth/2fa/verify', { uid, code }),
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
  checkReaction: (r1: string, r2: string) =>
    api.post('/game/check-reaction', { r1, r2 }),
  getMyDecks: () => 
    api.get('/my-decks'),
  createMyDeck: (name: string, cards: Record<string, number>) =>
    api.post('/my-decks', { name, cards }),
  updateMyDeck: (id: number, name: string, cards: Record<string, number>) =>
    api.put(`/my-decks/${id}`, { name, cards }),
  deleteMyDeck: (id: number) =>
    api.delete(`/my-decks/${id}`),
}

// 管理员API
export const adminAPI = {
  getAllUsers: () => 
    api.get('/admin/users'),
  createUser: (username: string, password: string) =>
    api.post('/admin/users', { username, password }),
  deleteUser: (userId: string) => 
    api.delete(`/admin/users/${userId}`),
  changeUserPassword: (userId: string, newPassword: string) => 
    api.put(`/admin/users/${userId}/password`, { new_password: newPassword }),
  promoteUser: (userId: string, role: string) => 
    api.put(`/admin/users/${userId}/role`, { role }),
  getGlobalDeckConfig: () => 
    api.get('/admin/deck-config'),
  updateGlobalDeckConfig: (name: string, cards: Record<string, number>) => 
    api.put('/admin/deck-config', { name, cards }),
  getGameHistory: () => 
    api.get('/admin/game-history'),
  getFeedbacks: () =>
    api.get('/admin/feedbacks'),
  updateFeedbackStatus: (id: number, status: string) =>
    api.put(`/admin/feedbacks/${id}/status`, { status }),
}

// 反应管理API
export const reactionAPI = {
  getReactions: () => 
    api.get('/reactions'),
  getAllReactions: () => 
    api.get('/reactions/all'),
  getMyReactions: () => 
    api.get('/reactions/my'),
  addReaction: (display: string) => 
    api.post('/reactions', { display }),
  batchAddReactions: (reactions: { display: string }[]) =>
    api.post('/reactions/batch', reactions),
  approveReaction: (groupId: string, display?: string, reject?: boolean) => 
    api.put(`/reactions/approve/${groupId}`, { display, reject }),
  deleteReaction: (reactionId: string) => 
    api.delete(`/reactions/${reactionId}`),
}

export default api
