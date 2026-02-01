import { createRouter, createWebHistory, type RouteLocationNormalized, type NavigationGuardNext } from 'vue-router'
import Login from './pages/Login.vue'
import Register from './pages/Register.vue'
import ForgotPassword from './pages/ForgotPassword.vue'
import Lobby from './pages/Lobby.vue'
import GameRoom from './pages/GameRoom.vue'
import Profile from './pages/Profile.vue'
import Admin from './pages/Admin.vue'
import Reactions from './pages/Reactions.vue'
import Feedbacks from './pages/Feedbacks.vue'

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
    path: '/forgot-password',
    name: 'ForgotPassword',
    component: ForgotPassword,
    meta: { guestOnly: true }
  },
  {
    path: '/',
    name: 'Lobby',
    component: Lobby,
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
    path: '/admin',
    name: 'Admin',
    component: Admin,
    meta: { requiresAuth: true, adminOnly: true }
  },
  {
    path: '/reactions',
    name: 'Reactions',
    component: Reactions,
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to: RouteLocationNormalized, _from: RouteLocationNormalized, next: NavigationGuardNext) => {
  const token = localStorage.getItem('token')
  const user = JSON.parse(localStorage.getItem('user') || 'null')

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
