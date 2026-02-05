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
      
      // 避免在登录页面或执行登录 API 时强制刷新/跳转
      const isLoginPage = window.location.pathname === '/login'
      const isAuthRequest = error.config.url.includes('/auth/login') || error.config.url.includes('/auth/webauthn/login')
      
      if (!isLoginPage && !isAuthRequest) {
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  }
)

// 认证API
export const authAPI = {
  register: (data: any) => 
    api.post('/auth/register', data),
  login: (data: any) => 
    api.post('/auth/login', data),
  getAuthConfig: () =>
    api.get('/auth/config'),
  sendCode: (email: string, type: string = 'register') =>
    api.post('/auth/send-code', { email, type }),
  resetPasswordByEmail: (data: any) =>
    api.post('/auth/reset-password', data),
  resetPasswordBy2FA: (data: any) =>
    api.post('/auth/2fa/reset-password', data),
  beginResetPasswordWebAuthn: (username: string) =>
    api.post('/auth/webauthn/reset-password/begin', { username }),
  finishResetPasswordWebAuthn: (username: string, newPassword: string, credential: any) =>
    api.post(`/auth/webauthn/reset-password/finish?username=${username}&new_password=${newPassword}`, credential),
  getUserInfo: () => 
    api.get('/user/info'),
  changePassword: (oldPassword: string, newPassword: string, code: string = '') => 
    api.put('/user/password', { old_password: oldPassword, new_password: newPassword, code }),
  beginChangePasswordWebAuthn: () =>
    api.post('/user/webauthn/change-password/begin'),
  finishChangePasswordWebAuthn: (newPassword: string, credential: any) =>
    api.post(`/user/webauthn/change-password/finish?newPassword=${newPassword}`, credential),
  updateAvatar: (avatar: string) => 
    api.put('/user/avatar', { avatar }),
  updateNickname: (nickname: string) =>
    api.put('/user/nickname', { nickname }),
  deleteAccount: () => 
    api.delete('/user/account'),
  searchUsers: (query: string) =>
    api.get(`/users/search?q=${encodeURIComponent(query)}`),
  submitFeedback: (content: string, type: string) =>
    api.post('/feedback', { content, type }),
  getMyFeedbacks: () =>
    api.get('/feedbacks/my'),
  urgeFeedback: (id: number) =>
    api.post(`/feedbacks/${id}/urge`),
  withdrawFeedback: (id: number) =>
    api.post('/feedback/withdraw', { id }),
  getGlobalChatHistory: (limit: number = 50) =>
    api.get(`/chat/global/history?limit=${limit}`),
  
  // 2FA相关
  setup2FA: () => api.post('/user/2fa/setup'),
  enable2FA: (code: string, password: string) => api.post('/user/2fa/enable', { code, password }),
  disable2FA: (code: string) => api.post('/user/2fa/disable', { code }),
  verify2FALogin: (uid: number, code: string) => api.post('/auth/2fa/verify', { uid, code }),

  // 会话管理
  getSessions: () => api.get('/user/sessions'),
  logoutSession: (id: string) => api.post('/user/sessions/logout', { id }),
  freezeAccount: (hours: number) => api.post('/user/account/freeze', { hours }),
}

// 游戏API
export const gameAPI = {
  getRooms: () => 
    api.get('/rooms'),
  createRoom: (name: string, maxPlayers: number, deckID: number, isPointsMode: boolean = false) => 
    api.post('/rooms', { name, max_players: maxPlayers, deck_id: deckID, is_points_mode: isPointsMode }),
  getRoomState: (roomId: string) => 
    api.get(`/rooms/${roomId}`),
  joinRoom: (roomId: string) => 
    api.post(`/rooms/${roomId}/join`),
  leaveRoom: (roomId: string) => 
    api.post(`/rooms/${roomId}/leave`),
  startGame: (roomId: string) => 
    api.post(`/rooms/${roomId}/start`),
  initiateDuel: (target_uid: number) =>
    api.post('/game/duel', { target_uid }),
  respondToDuel: (target_uid: number, accept: boolean) =>
    api.post('/game/duel/respond', { target_uid, accept }),
  getMyGameHistory: () =>
    api.get('/user/game-history'),
  playCard: (roomId: string, card: any, substance: string) => 
    api.post(`/rooms/${roomId}/play`, { card, substance }),
  playDouble: (roomId: string, sub1: string, sub2: string) =>
    api.post(`/rooms/${roomId}/play-double`, { sub1, sub2 }),
  drawCard: (roomId: string) => 
    api.post(`/rooms/${roomId}/draw`),
  getAvailableSubstances: (roomId: string) => 
    api.get(`/rooms/${roomId}/substances`),
  checkReaction: (r1: string, r2: string) =>
    api.post('/game/check-reaction', { r1, r2 }),
  getMyDecks: () => 
    api.get('/my-decks'),
  createMyDeck: (name: string, cards: Record<string, number>, initialCards?: number) =>
    api.post('/my-decks', { name, cards, initial_cards: initialCards }),
  updateMyDeck: (id: number, name: string, cards: Record<string, number>, initialCards?: number) =>
    api.put(`/my-decks/${id}`, { name, cards, initial_cards: initialCards }),
  deleteMyDeck: (id: number) =>
    api.delete(`/my-decks/${id}`),
}

export const pointsAPI = {
  getLeaderboard: (mode: string = 'total') => api.get(`/points/leaderboard?mode=${mode}`),
  createBounty: (target_uid: number, amount: number) => 
    api.post('/points/bounty', { target_uid, amount }),
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
  banUser: (targetUID: number, hours: number, reason: string) =>
    api.post('/admin/users/ban', { target_uid: targetUID, hours, reason }),
  kickPlayer: (roomID: string, targetUID: number, reason: string) =>
    api.post('/admin/rooms/kick', { room_id: roomID, target_uid: targetUID, reason }),
  getGlobalDeckConfig: () => 
    api.get('/admin/deck-config'),
  updateGlobalDeckConfig: (name: string, cards: Record<string, number>, initialCards?: number) => 
    api.put('/admin/deck-config', { name, cards, initial_cards: initialCards }),
  getGameHistory: () => 
    api.get('/admin/game-history'),
  getFeedbacks: () =>
    api.get('/admin/feedbacks'),
  updateFeedbackStatus: (id: number, status: string, note?: string) =>
    api.put(`/admin/feedbacks/${id}/status`, { status, note }),
  getConfigs: () =>
    api.get('/admin/configs'),
  updateConfig: (key: string, value: string) =>
    api.put('/admin/configs', { key, value }),
  
  // 公告管理
  getAnnouncements: () =>
    api.get('/admin/announcements'),
  createAnnouncement: (title: string, content: string, type: string, is_ticker: boolean, expires_in?: string, on_join: boolean = false, cron_interval: number = 0, close_delay: number = 0, is_persistent: boolean = false) =>
    api.post('/admin/announcements', { title, content, type, is_ticker, expires_in, on_join, cron_interval, close_delay, is_persistent }),
  updateAnnouncementStatus: (id: number, active: boolean) =>
    api.put(`/admin/announcements/${id}/status`, { active }),
  deleteAnnouncement: (id: number) =>
    api.delete(`/admin/announcements/${id}`),
}

export const commonAPI = {
  getAnnouncements: () =>
    api.get('/announcements'),
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
  updateReaction: (id: number, display: string) => 
    api.put(`/reactions/${id}`, { display }),
  approveReaction: (groupId: string, display: string, reject: boolean = false) =>
    api.put(`/reactions/approve/${groupId}`, { display, reject }),
  deleteReaction: (reactionId: number) => 
    api.delete(`/reactions/${reactionId}`),
}

// 物质管理API
export const substanceAPI = {
  getSubstances: () =>
    api.get('/substances'),
  addSubstance: (formula: string, name: string) =>
    api.post('/substances', { formula, name }),
  updateSubstance: (id: number, formula: string, name: string) =>
    api.put(`/substances/${id}`, { formula, name }),
  approveSubstance: (id: number, data: { formula?: string, name?: string, reject?: boolean }) =>
    api.put(`/substances/approve/${id}`, data),
  deleteSubstance: (id: number) =>
    api.delete(`/substances/${id}`),
}

// 好友系统API
export const friendAPI = {
  sendRequest: (friendId: number, message: string = '') =>
    api.post('/friends/request', { friend_id: friendId, message }),
  getPendingRequests: () =>
    api.get('/friends/pending'),
  handleRequest: (requestId: number, action: 'accept' | 'decline') =>
    api.post('/friends/handle', { request_id: requestId, action }),
  getFriends: () =>
    api.get('/friends'),
  deleteFriend: (friendId: number) =>
    api.delete(`/friends/${friendId}`),
}

export default api
