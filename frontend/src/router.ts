import { createRouter, createWebHistory, type RouteLocationNormalized } from 'vue-router'
import { buildApiURL } from './utils/runtimeConfig'

const Login = () => import('./pages/Login.vue')
const Register = () => import('./pages/Register.vue')
const Lobby = () => import('./pages/Lobby.vue')
const GameRoom = () => import('./pages/GameRoom.vue')
const Profile = () => import('./pages/Profile.vue')
const Appeals = () => import('./pages/Appeals.vue')
const Admin = () => import('./pages/Admin.vue')
const AdminPlugins = () => import('./pages/AdminPlugins.vue')
const AdminSurveyResponses = () => import('./pages/AdminSurveyResponses.vue')
const AdminAnticheat = () => import('./pages/AdminAnticheat.vue')
const ActivityAdmin = () => import('./pages/ActivityAdmin.vue')
const BingoRoom = () => import('./pages/BingoRoom.vue')
const Plugins = () => import('./pages/Plugins.vue')
const Reactions = () => import('./pages/Reactions.vue')
const Feedbacks = () => import('./pages/Feedbacks.vue')
const Survey = () => import('./pages/Survey.vue')
const Ranking = () => import('./pages/Ranking.vue')
const DataConfig = () => import('./pages/DataConfig.vue')
const Substances = () => import('./pages/Substances.vue')
const Chat = () => import('./pages/Chat.vue')
const UserSpace = () => import('./pages/UserSpace.vue')
const ReplayRoom = () => import('./pages/ReplayRoom.vue')
const OAuthCallback = () => import('./pages/OAuthCallback.vue')

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
    path: '/replay/:historyId',
    name: 'ReplayRoom',
    component: ReplayRoom,
    meta: { requiresAuth: true }
  },
  {
    path: '/profile',
    name: 'Profile',
    component: Profile,
    meta: { requiresAuth: true }
  },
  {
    path: '/profile/:tab',
    name: 'ProfileTab',
    component: Profile,
    meta: { requiresAuth: true }
  },
  {
    path: '/appeals',
    name: 'Appeals',
    component: Appeals,
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
    path: '/admin/activities',
    name: 'ActivityAdmin',
    component: ActivityAdmin,
    meta: { requiresAuth: true, coWorkerOnly: true }
  },
  {
    path: '/bingo/rooms/:id',
    name: 'BingoRoom',
    component: BingoRoom,
    meta: { requiresAuth: true }
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
    meta: { requiresAuth: true, coWorkerOnly: true }
  },
  {
    path: '/admin/surveys/:id/responses',
    name: 'AdminSurveyResponses',
    component: AdminSurveyResponses,
    meta: { requiresAuth: true, coWorkerOnly: true }
  },
  {
    path: '/admin/anticheat',
    name: 'AdminAnticheat',
    component: AdminAnticheat,
    meta: { requiresAuth: true, adminOnly: true }
  },
  {
    path: '/admin/:tab',
    name: 'AdminTab',
    component: Admin,
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
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'InvalidUriRedirect',
    redirect: '/'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

const isDataRoute = (path: string): boolean => path === '/data' || path.startsWith('/data/')

const findActiveRoomId = async (_token: string | null, uid: number): Promise<string | null> => {
  try {
    const res = await fetch(buildApiURL('/rooms'), {
      credentials: 'include', // 自动发送cookie中的token
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
  // Token存储在HttpOnly Cookie中，检查user信息判断是否已登录
  let user = null
  try {
    user = JSON.parse(localStorage.getItem('user') || 'null')
  } catch (e) {
    console.error('Failed to parse user from localStorage', e)
    localStorage.removeItem('user')
  }

  if (to.meta.requiresAuth && !user) {
    return {
      path: '/login',
      query: { redirect: to.fullPath }
    }
  } else if (to.meta.guestOnly && user) {
    const redirect = to.query.redirect as string
    if (redirect && redirect.startsWith('/') && !redirect.startsWith('//')) {
      return redirect
    }
    return '/'
  } else if (to.meta.adminOnly && (!user || user.role !== 'admin')) {
    return '/'
  } else if (to.meta.coWorkerOnly && (!user || (user.role !== 'admin' && user.role !== 'co-worker'))) {
    return '/'
  }

  if (user?.uid && isDataRoute(to.path)) {
    const activeRoomID = await findActiveRoomId(null, Number(user.uid))
    if (activeRoomID) {
      return `/room/${activeRoomID}`
    }
  }

  return true
})

export default router
