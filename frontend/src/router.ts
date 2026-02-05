import { createRouter, createWebHistory, type RouteLocationNormalized, type NavigationGuardNext } from 'vue-router'
import Login from './pages/Login.vue'
import Register from './pages/Register.vue'
import Lobby from './pages/Lobby.vue'
import GameRoom from './pages/GameRoom.vue'
import Profile from './pages/Profile.vue'
import Admin from './pages/Admin.vue'
import Reactions from './pages/Reactions.vue'
import Feedbacks from './pages/Feedbacks.vue'
import Ranking from './pages/Ranking.vue'
import DataConfig from './pages/DataConfig.vue'
import Substances from './pages/Substances.vue'
import Chat from './pages/Chat.vue'

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
    next('/login')
  } else if (to.meta.guestOnly && token) {
    next('/')
  } else if (to.meta.adminOnly && (!user || !user.is_admin)) {
    next('/')
  } else if (to.meta.coWorkerOnly && (!user || (user.role !== 'admin' && user.role !== 'co-worker'))) {
    next('/')
  } else {
    next()
  }
})

export default router
