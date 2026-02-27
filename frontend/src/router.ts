import { createRouter, createWebHistory, type RouteLocationNormalized, type NavigationGuardNext } from 'vue-router'
import Login from './pages/Login.vue'
import Register from './pages/Register.vue'
import Lobby from './pages/Lobby.vue'
import GameRoom from './pages/GameRoom.vue'
import Profile from './pages/Profile.vue'
import Admin from './pages/Admin.vue'
import AdminPlugins from './pages/AdminPlugins.vue'
import Plugins from './pages/Plugins.vue'
import Reactions from './pages/Reactions.vue'
import Feedbacks from './pages/Feedbacks.vue'
import Ranking from './pages/Ranking.vue'
import DataConfig from './pages/DataConfig.vue'
import Substances from './pages/Substances.vue'
import Chat from './pages/Chat.vue'
import UserSpace from './pages/UserSpace.vue'

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
    path: '/feedbacks/my',
    redirect: '/feedbacks'
  },
  {
    path: '/admin',
    name: 'Admin',
    component: Admin,
    meta: { requiresAuth: true, adminOnly: true }
  },
  {
    path: '/admin/plugins',
    name: 'AdminPlugins',
    component: AdminPlugins,
    meta: { requiresAuth: true, adminOnly: true }
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
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to: RouteLocationNormalized, _from: RouteLocationNormalized, next: NavigationGuardNext) => {
  const token = localStorage.getItem('token')
  let user = null
  try {
    user = JSON.parse(localStorage.getItem('user') || 'null')
  } catch (e) {
    console.error('Failed to parse user from localStorage', e)
    localStorage.removeItem('user')
  }

  if (to.meta.requiresAuth && !token) {
    // 保存当前完整路径（包括查询参数）作为登录后的重定向目标
    const redirectPath = to.fullPath
    next({
      path: '/login',
      query: { redirect: redirectPath }
    })
  } else if (to.meta.guestOnly && token) {
    // 如果已登录用户访问登录页，检查是否有redirect参数
    // 只允许同源路径（必须以 / 开头），防止开放重定向攻击
    const redirect = to.query.redirect as string
    if (redirect && redirect.startsWith('/') && !redirect.startsWith('//')) {
      next(redirect)
    } else {
      next('/')
    }
  } else if (to.meta.adminOnly && (!user || !user.is_admin)) {
    next('/')
  } else if (to.meta.coWorkerOnly && (!user || (user.role !== 'admin' && user.role !== 'co-worker'))) {
    next('/')
  } else {
    next()
  }
})

export default router
