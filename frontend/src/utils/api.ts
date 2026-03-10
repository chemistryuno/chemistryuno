import axios, { AxiosInstance, InternalAxiosRequestConfig, AxiosResponse } from 'axios'
import router from '../router'

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
        // 使用路由跳转而非页面刷新
        const currentPath = window.location.pathname + window.location.search
        router.push({
          path: '/login',
          query: { redirect: currentPath }
        })
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
  unbindOAuth: (provider: string) =>
    api.post(`/auth/oauth/unbind?provider=${provider}`),
  sendCode: (email: string, type: string = 'register', recaptcha_token?: string) =>
    api.post('/auth/send-code', { email, type, recaptcha_token }),
  resetPasswordByEmail: (data: any) =>
    api.post('/auth/reset-password', data),
  resetPasswordBy2FA: (data: any) =>
    api.post('/auth/2fa/reset-password', data),
  beginWebAuthnLogin: (identifier: string = '') =>
    api.get(`/auth/webauthn/login/begin${identifier ? `?identifier=${encodeURIComponent(identifier)}` : ''}`),
  finishWebAuthnLogin: (credential: any, identifier: string = '') =>
    api.post(`/auth/webauthn/login/finish${identifier ? `?identifier=${encodeURIComponent(identifier)}` : ''}`, credential),
  beginResetPasswordWebAuthn: (identifier: string) =>
    api.post('/auth/webauthn/reset-password/begin', { identifier }),
  finishResetPasswordWebAuthn: (identifier: string, newPassword: string, credential: any) =>
    api.post(`/auth/webauthn/reset-password/finish?identifier=${encodeURIComponent(identifier)}&new_password=${encodeURIComponent(newPassword)}`, credential),
  getUserInfo: () =>
    api.get('/user/info'),
  changePassword: (oldPassword: string, newPassword: string, code: string = '', useEmail: boolean = false) =>
    api.put('/user/password', { old_password: oldPassword, new_password: newPassword, code, use_email: useEmail }),
  beginChangePasswordWebAuthn: () =>
    api.post('/user/webauthn/change-password/begin'),
  finishChangePasswordWebAuthn: (newPassword: string, credential: any) =>
    api.post(`/user/webauthn/change-password/finish?newPassword=${newPassword}`, credential),
  updateAvatar: (avatar: string) =>
    api.put('/user/avatar', { avatar }),
  updateProfile: (data: {
    nickname: string,
    bio?: string,
    wechat?: string,
    qq?: string,
    show_email?: boolean,
    birthday?: string | null,
    sound_volume?: number,
    vibration_enabled?: boolean,
    enable_element_input?: boolean,
    custom_contact?: string
  }) =>
    api.put('/user/profile', data),
  getUserPublicProfile: (uid: number) =>
    api.get(`/user/profile/${uid}`),
  changeEmail: (data: { old_code: string, new_email: string, new_code: string }) =>
    api.post('/user/change-email', data),
  setEmail: (data: { new_email: string, new_code: string, security_answer?: string }) =>
    api.post('/user/set-email', data),
  getMySecurityQuestion: () =>
    api.get('/user/security-question'),
  updateSecurityQuestion: (data: { security_question: string, security_answer: string, current_password: string }) =>
    api.put('/user/security-question', data),
  getSecurityQuestion: (username: string) =>
    api.get(`/auth/security-question?username=${encodeURIComponent(username)}`),
  resetPasswordBySecurityQuestion: (data: { username: string, security_answer: string, new_password: string }) =>
    api.post('/auth/security-question/reset-password', data),
  deleteAccount: (code: string) =>
    api.delete('/user/account', { data: { code } }),
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
  getActiveSurveys: () =>
    api.get('/surveys/active'),
  getAllActiveSurveys: () =>
    api.get('/surveys/all'),
  dismissSurvey: (id: number) =>
    api.post(`/surveys/${id}/dismiss`),
  getSurveyDetail: (id: number) =>
    api.get(`/surveys/${id}`),
  submitSurveyAnswers: (id: number, answers: any[]) =>
    api.post(`/surveys/${id}/submit`, { answers }),
  getGlobalChatHistory: (limit: number = 50) =>
    api.get(`/chat/global/history?limit=${limit}`),
  getPrivateChatHistory: (friendUID: number, limit: number = 50) =>
    api.get(`/chat/private/history/${friendUID}?limit=${limit}`),

  // 2FA相关
  setup2FA: () => api.post('/user/2fa/setup'),
  enable2FA: (code: string, password: string) => api.post('/user/2fa/enable', { code, password }),
  disable2FA: (code: string) => api.post('/user/2fa/disable', { code }),

  // 版本信息
  getVersion: () => api.get('/version'),
  verify2FALogin: (uid: number, code: string) => api.post('/auth/2fa/verify', { uid, code }),

  // 会话管理
  getSessions: () => api.get('/user/sessions'),
  logoutSession: (id: string) => api.post('/user/sessions/logout', { id }),
  freezeAccount: (hours: number) => api.post('/user/account/freeze', { hours }),

  // WebAuthn 凭证管理 (由 HardwareKeyModal.vue 使用)
  getWebAuthnCredentials: () => api.get('/user/webauthn/credentials'),
  beginWebAuthnRegistration: () => api.get('/user/webauthn/register/begin'),
  finishWebAuthnRegistration: (credential: any) => api.post('/user/webauthn/register/finish', credential),
  removeWebAuthnCredential: (id: string) => api.delete(`/user/webauthn/credentials/${id}`),
}

// 游戏API
export const gameAPI = {
  getRooms: () =>
    api.get('/rooms'),
  createRoom: (name: string, maxPlayers: number, deckID: number, isPointsMode: boolean = false, isPrivate: boolean = false, accessKey?: string, isPvE: boolean = false, pveDifficulty: number = 0, aiCount: number = 0, enableAIBackfill: boolean = false, aiBackfillDifficulty: number = 50, isRanked: boolean = false, levelRange: number = 5, tutorialScript: boolean = false) =>
    api.post('/rooms', {
      name,
      max_players: maxPlayers,
      deck_id: deckID,
      is_points_mode: isPointsMode,
      is_private: isPrivate,
      access_key: accessKey,
      is_pve: isPvE,
      pve_difficulty: pveDifficulty,
      ai_count: aiCount,
      enable_ai_backfill: enableAIBackfill,
      ai_backfill_difficulty: aiBackfillDifficulty,
      is_ranked: isRanked,
      level_range: levelRange,
      tutorial_script: tutorialScript
    }),
  getRoomState: (roomId: string) =>
    api.get(`/rooms/${roomId}`),
  checkRoomStatus: (roomId: string) =>
    api.get(`/rooms/${roomId}/status`),
  joinRoom: (roomId: string, accessKey?: string) =>
    api.post(`/rooms/${roomId}/join${accessKey ? `?key=${accessKey}` : ''}`),
  leaveRoom: (roomId: string) =>
    api.post(`/rooms/${roomId}/leave`),
  ready: (roomId: string) =>
    api.post(`/rooms/${roomId}/ready`),
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
  getReactionHints: (roomId: string) =>
    api.get(`/rooms/${roomId}/reaction-hints`),
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
  getStats: () =>
    api.get('/admin/stats'),
  getAllUsers: () =>
    api.get('/admin/users'),
  createUser: (username: string, password: string) =>
    api.post('/admin/users', { username, password }),
  deleteUser: (uid: string) =>
    api.delete(`/admin/users/${uid}`),
  changeUserPassword: (uid: string, newPassword: string) =>
    api.put(`/admin/users/${uid}/password`, { new_password: newPassword }),
  promoteUser: (uid: string, role: string) =>
    api.put(`/admin/users/${uid}/role`, { role }),
  banUser: (targetUID: number, bannedUntil: string, reason: string) =>
    api.post('/admin/users/ban', { target_uid: targetUID, banned_until: bannedUntil, reason }),
  kickPlayer: (targetUID: number, reason: string) =>
    api.post('/admin/users/kick', { target_uid: targetUID, reason }),
  getGlobalDeckConfig: () =>
    api.get('/admin/deck-config'),
  updateGlobalDeckConfig: (name: string, cards: Record<string, number>, initialCards?: number) =>
    api.put('/admin/deck-config', { name, cards, initial_cards: initialCards }),
  resetGlobalDeckConfig: () =>
    api.post('/admin/deck-config/reset'),
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
  getGameTimeConfigs: () =>
    api.get('/admin/game-time-configs'),
  updateGameTimeConfig: (data: any) =>
    api.put('/admin/game-time-configs', data),

  // Excel导出
  exportSubstances: () =>
    api.get('/admin/export/substances', { responseType: 'blob' }),
  exportReactions: () =>
    api.get('/admin/export/reactions', { responseType: 'blob' }),
  exportAllData: () =>
    api.get('/admin/export/all', { responseType: 'blob' }),

  // 批量批准/拒绝物质
  batchApproveSubstances: (groupIDs: number[]) =>
    api.post('/admin/substances/batch-approve', { group_ids: groupIDs }),
  batchRejectSubstances: (groupIDs: number[]) =>
    api.post('/admin/substances/batch-reject', { group_ids: groupIDs }),

  // 批量批准/拒绝反应
  batchApproveReactions: (groupIDs: number[]) =>
    api.post('/admin/reactions/batch-approve', { group_ids: groupIDs }),
  batchRejectReactions: (groupIDs: number[]) =>
    api.post('/admin/reactions/batch-reject', { group_ids: groupIDs }),

  // 广播系统
  broadcast: (data: {
    scope: 'global' | 'room' | 'user'
    target?: string
    msg_type: 'info' | 'warning' | 'success' | 'error'
    title?: string
    content: string
  }) =>
    api.post('/admin/broadcast', data),
  getActiveRooms: () =>
    api.get('/admin/rooms/active'),

  // 公告管理
  getAnnouncements: () =>
    api.get('/admin/announcements'),
  createAnnouncement: (title: string, content: string, type: string, is_ticker: boolean, expires_in?: string, on_join: boolean = false, cron_interval: number = 0, close_delay: number = 0, is_persistent: boolean = false) =>
    api.post('/admin/announcements', { title, content, type, is_ticker, expires_in, on_join, cron_interval, close_delay, is_persistent }),
  updateAnnouncement: (id: number, title: string, content: string, type: string, is_ticker: boolean, expires_in?: string, on_join: boolean = false, cron_interval: number = 0, close_delay: number = 0, is_persistent: boolean = false) =>
    api.put(`/admin/announcements/${id}`, { title, content, type, is_ticker, expires_in, on_join, cron_interval, close_delay, is_persistent }),
  updateAnnouncementStatus: (id: number, active: boolean) =>
    api.put(`/admin/announcements/${id}/status`, { active }),
  deleteAnnouncement: (id: number) =>
    api.delete(`/admin/announcements/${id}`),

  // 问卷管理
  getSurveys: () =>
    api.get('/admin/surveys'),
  createSurvey: (data: any) =>
    api.post('/admin/surveys', data),
  updateSurvey: (id: number, data: any) =>
    api.put(`/admin/surveys/${id}`, data),
  updateSurveyStatus: (id: number, isActive: boolean) =>
    api.put(`/admin/surveys/${id}/status`, { is_active: isActive }),
  getSurveyResponses: (id: number, sortBy: string = 'created_at', order: string = 'desc') =>
    api.get(`/admin/surveys/${id}/responses`, { params: { sort_by: sortBy, order } }),
  repairSurvey: (id: number) =>
    api.post(`/admin/surveys/${id}/repair`),
  exportSurvey: (id: number) =>
    api.get(`/admin/surveys/${id}/export`, { responseType: 'blob' }),
  getSurveyConfig: (id: number) =>
    api.get(`/admin/surveys/${id}/config`),
  importSurveyConfig: (data: any) =>
    api.post('/admin/surveys/import', data),
}

export const commonAPI = {
  getAnnouncements: () =>
    api.get('/announcements'),
  getHints: () =>
    api.get('/hints'),
}

// 等级系统API
export const levelAPI = {
  getLevelInfo: () =>
    api.get('/level/info'),
  getUserLevelInfo: (uid: number) =>
    api.get(`/level/user/${uid}`),
  getLevelLeaderboard: (limit: number = 100) =>
    api.get(`/level/leaderboard?limit=${limit}`),
  getLevelConfigs: () =>
    api.get('/level/configs'),
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
  // 获取全量物质名称映射（不带版本，用于自动推导）
  getSubstanceNames: () =>
    api.get('/substances/names'),
  // 获取所有物质（分组）
  getSubstances: () =>
    api.get('/data/substances'),
  // 获取我的物质
  getMySubstances: () =>
    api.get('/data/substances/my'),
  // 获取物质组内所有版本
  getSubstanceGroup: (id: number) =>
    api.get(`/data/substances/${id}/group`),
  // 提交新物质建议
  submitNewSubstance: (formula: string, name: string, elements?: string) =>
    api.post('/data/substances/new', { formula, name, elements }),
  // 提交物质更新建议
  submitSubstanceUpdate: (id: number, formula: string, name: string, elements?: string) =>
    api.post(`/data/substances/${id}/update`, { formula, name, elements }),
  // 管理员直接更新物质
  updateSubstance: (id: number, formula: string, name: string, elements?: string) =>
    api.put(`/data/substances/${id}`, { formula, name, elements }),
  // 管理员批准物质更新
  approveSubstance: (id: number) =>
    api.post(`/data/substances/${id}/approve`),
  // 管理员拒绝物质更新
  rejectSubstance: (id: number) =>
    api.delete(`/data/substances/${id}/reject`),
}

// 好友系统API
export const friendAPI = {
  sendRequest: (friendUID: number, message: string = '') =>
    api.post('/friends/request', { friend_uid: friendUID, message }),
  getPendingRequests: () =>
    api.get('/friends/pending'),
  handleRequest: (requestId: number, action: 'accept' | 'decline') =>
    api.post('/friends/handle', { request_id: requestId, action }),
  getFriends: () =>
    api.get('/friends'),
  deleteFriend: (friendUID: number) =>
    api.delete(`/friends/${friendUID}`),
  setRemark: (friendUID: number, remark: string) =>
    api.post('/friends/remark', { friend_uid: friendUID, remark }),
}

// 插件系统API
export const pluginAPI = {
  // 公开接口（需要登录）
  getPluginCards: () =>
    api.get('/plugin-cards'),
  // 用户只读 - 插件浏览
  getPluginsWithCards: () =>
    api.get('/plugins'),
  getPluginScript: (pluginId: number) =>
    api.get(`/plugins/${pluginId}/script`, { responseType: 'text' }),
  getPluginSettings: (pluginId: number) =>
    api.get(`/plugins/${pluginId}/settings`),

  // 管理员接口 - 插件管理
  getPlugins: () =>
    api.get('/admin/plugins'),
  createPlugin: (data: { name: string; description?: string }) =>
    api.post('/admin/plugins', data),
  updatePlugin: (id: number, data: { name?: string; description?: string; is_active?: boolean }) =>
    api.put(`/admin/plugins/${id}`, data),
  updatePluginSettings: (pluginId: number, settings: Record<string, string>) =>
    api.put(`/admin/plugins/${pluginId}/settings`, { settings }),
  getPluginSettingsHistory: (pluginId: number) =>
    api.get(`/admin/plugins/${pluginId}/settings/history`),
  rollbackPluginSettings: (pluginId: number, snapshotId: string) =>
    api.post(`/admin/plugins/${pluginId}/settings/rollback`, { snapshot_id: snapshotId }),
  deletePlugin: (id: number) =>
    api.delete(`/admin/plugins/${id}`),

  // .cumod 文件安装
  installPlugin: (file: File) => {
    const form = new FormData()
    form.append('file', file)
    return api.post('/admin/plugins/install', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },

  // 管理员接口 - 卡牌管理
  getPluginCardsByPlugin: (pluginId: number) =>
    api.get(`/admin/plugins/${pluginId}/cards`),
  createCard: (pluginId: number, data: {
    symbol: string
    display_name?: string
    effect_type: string
    effect_config: object
    default_count?: number
    color?: string
  }) =>
    api.post(`/admin/plugins/${pluginId}/cards`, data),
  updateCard: (cardId: number, data: {
    symbol?: string
    display_name?: string
    effect_type?: string
    effect_config?: object
    default_count?: number
    color?: string
  }) =>
    api.put(`/admin/plugin-cards/${cardId}`, data),
  deleteCard: (cardId: number) =>
    api.delete(`/admin/plugin-cards/${cardId}`),

  // 热重载
  reloadPlugins: () =>
    api.post('/admin/plugins/reload'),

  // 服务器重启管理
  scheduleRestart: (delaySeconds: number, reason?: string) =>
    api.post('/admin/server/restart', { delay_seconds: delaySeconds, reason }),
  cancelRestart: () =>
    api.post('/admin/server/restart/cancel'),
}

export default api
