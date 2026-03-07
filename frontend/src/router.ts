import { createRouter, createWebHistory, type RouteLocationNormalized } from 'vue-router'
import Login from './pages/Login.vue'
import Register from './pages/Register.vue'
import Lobby from './pages/Lobby.vue'
import GameRoom from './pages/GameRoom.vue'
import Profile from './pages/Profile.vue'
import Admin from './pages/Admin.vue'
import AdminPlugins from './pages/AdminPlugins.vue'
import AdminSurveyResponses from './pages/AdminSurveyResponses.vue'
import Plugins from './pages/Plugins.vue'
import Reactions from './pages/Reactions.vue'
import Feedbacks from './pages/Feedbacks.vue'
import Survey from './pages/Survey.vue'
import Ranking from './pages/Ranking.vue'
import DataConfig from './pages/DataConfig.vue'
import Substances from './pages/Substances.vue'
import Chat from './pages/Chat.vue'
import UserSpace from './pages/UserSpace.vue'
import OAuthCallback from './pages/OAuthCallback.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { guestOnly: true }
  },
  {
    path: '/register',
    name: 'Register',
    component: Register,
    meta: { guestOnly: true }
  },
  {
    path: '/',
    name: 'Lobby',
    component: Lobby,
    meta: { requiresAuth: true }
  },
  {
    path: '/chat',
    name: 'Chat',
    component: Chat,
    meta: { requiresAuth: true }
  },
  {
    path: '/room/:id',
    name: 'GameRoom',
    component: GameRoom,
    meta: { requiresAuth: true }
  },
  {
    path: '/profile',
    name: 'Profile',
    component: Profile,
    meta: { requiresAuth: true }
  },
  {
    path: '/feedbacks',
    name: 'Feedbacks',
    component: Feedbacks,
    meta: { requiresAuth: true }
  },
  {
    path: '/surveys/:id',
    name: 'Survey',
    component: Survey,
    meta: { requiresAuth: true }
  },
  {
    path: '/feedbacks/my',
    redirect: '/feedbacks'
  },
  {
    path: '/admin',
    name: 'Admin',
    component: Admin,
    meta: { requiresAuth: true, coWorkerOnly: true }
  },
  {
    path: '/admin/plugins',
    name: 'AdminPlugins',
    component: AdminPlugins,
    meta: { requiresAuth: true, adminOnly: true }
  },
  {
    path: '/admin/surveys/:id/responses',
    name: 'AdminSurveyResponses',
    component: AdminSurveyResponses,
    meta: { requiresAuth: true, coWorkerOnly: true }
  },
  {
    path: '/plugins',
    name: 'Plugins',
    component: Plugins,
    meta: { requiresAuth: true }
  },
  {
    path: '/data',
    name: 'DataConfig',
    component: DataConfig,
    meta: { requiresAuth: true }
  },
  {
    path: '/data/reactions',
    name: 'Reactions',
    component: Reactions,
    meta: { requiresAuth: true }
  },
  {
    path: '/data/substances',
    name: 'Substances',
    component: Substances,
    meta: { requiresAuth: true }
  },
  {
    path: '/ranking',
    name: 'Ranking',
    component: Ranking,
    meta: { requiresAuth: true }
  },
  {
    path: '/user/:uid',
    name: 'UserSpace',
    component: UserSpace,
    meta: { requiresAuth: true }
  },
  {
    // OAuth 降级回调页：当 window.opener 不可用时，后端将 token 通过 URL hash 重定向至此
    path: '/oauth-callback',
    name: 'OAuthCallback',
    component: OAuthCallback
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

const isDataRoute = (path: string): boolean => path === '/data' || path.startsWith('/data/')

const findActiveRoomId = async (token: string, uid: number): Promise<string | null> => {
  try {
    const res = await fetch('/api/rooms', {
      headers: {
        Authorization: `Bearer ${token}`
      }
    })
    if (!res.ok) return null

    const payload = await res.json()
    const rooms = Array.isArray(payload) ? payload : (Array.isArray(payload?.data) ? payload.data : [])
    const activeRoom = rooms.find((room: any) => {
      const roomStatus = room?.status
      const players = Array.isArray(room?.players) ? room.players : []
      return (roomStatus === 'waiting' || roomStatus === 'playing') &&
        players.some((playerUID: any) => Number(playerUID) === uid)
    })

    return activeRoom?.id ? String(activeRoom.id) : null
  } catch (error) {
    console.error('Failed to check active room before entering data routes', error)
    return null
  }
}

router.beforeEach(async (to: RouteLocationNormalized, _from: RouteLocationNormalized) => {
  const token = localStorage.getItem('token')
  let user = null
  try {
    user = JSON.parse(localStorage.getItem('user') || 'null')
  } catch (e) {
    console.error('Failed to parse user from localStorage', e)
    localStorage.removeItem('user')
  }

  if (to.meta.requiresAuth && !token) {
    return {
      path: '/login',
      query: { redirect: to.fullPath }
    }
  } else if (to.meta.guestOnly && token) {
    const redirect = to.query.redirect as string
    if (redirect && redirect.startsWith('/') && !redirect.startsWith('//')) {
      return redirect
    }
    return '/'
  } else if (to.meta.adminOnly && (!user || !user.is_admin)) {
    return '/'
  } else if (to.meta.coWorkerOnly && (!user || (user.role !== 'admin' && user.role !== 'co-worker'))) {
    return '/'
  }

  if (token && user?.uid && isDataRoute(to.path)) {
    const activeRoomID = await findActiveRoomId(token, Number(user.uid))
    if (activeRoomID) {
      return `/room/${activeRoomID}`
    }
  }

  return true
})

export default router
