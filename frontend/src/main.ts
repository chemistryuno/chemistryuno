import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { buildApiURL } from './utils/runtimeConfig'
import './index.css'

const syncViewportHeight = () => {
  const viewportHeight = window.visualViewport?.height ?? window.innerHeight
  document.documentElement.style.setProperty('--app-height', `${viewportHeight}px`)
}

let viewportSyncRaf = 0

const scheduleViewportSync = () => {
  if (viewportSyncRaf) return

  viewportSyncRaf = window.requestAnimationFrame(() => {
    viewportSyncRaf = 0
    syncViewportHeight()
  })
}

const clearStaleAuthState = () => {
  localStorage.removeItem('user')
  localStorage.removeItem('token')
  localStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')
}

const ensureStartupSession = async () => {
  const rawUser = localStorage.getItem('user')
  if (!rawUser) {
    return
  }

  try {
    JSON.parse(rawUser)
  } catch {
    clearStaleAuthState()
    return
  }

  try {
    const userInfoResp = await fetch(buildApiURL('/user/info'), {
      method: 'GET',
      credentials: 'include'
    })

    if (userInfoResp.ok) {
      return
    }

    if (userInfoResp.status !== 401) {
      return
    }

    const refreshResp = await fetch(buildApiURL('/auth/refresh'), {
      method: 'POST',
      credentials: 'include'
    })

    if (!refreshResp.ok) {
      clearStaleAuthState()
      return
    }

    const verifyResp = await fetch(buildApiURL('/user/info'), {
      method: 'GET',
      credentials: 'include'
    })

    if (!verifyResp.ok) {
      clearStaleAuthState()
    }
  } catch (error) {
    console.warn('[Auth] startup session check failed', error)
  }
}

const markBootSplashReady = () => {
  const splash = document.getElementById('boot-splash')

  window.requestAnimationFrame(() => {
    document.body.classList.add('app-ready')

    if (!splash) return

    window.setTimeout(() => {
      splash.remove()
    }, 360)
  })
}

const scheduleNonCriticalTask = (task: () => void, timeout = 1000) => {
  const requestIdleCallback = (window as Window & typeof globalThis & {
    requestIdleCallback?: (callback: () => void, options?: { timeout: number }) => number
  }).requestIdleCallback

  if (typeof requestIdleCallback === 'function') {
    requestIdleCallback(task, { timeout })
    return
  }

  window.setTimeout(task, timeout)
}

async function bootstrap() {
  scheduleViewportSync()
  window.addEventListener('resize', scheduleViewportSync, { passive: true })
  window.addEventListener('orientationchange', scheduleViewportSync, { passive: true })
  window.visualViewport?.addEventListener('resize', scheduleViewportSync)

  await ensureStartupSession()

  const app = createApp(App)
  app.use(router)
  app.mount('#app')
  markBootSplashReady()

  scheduleNonCriticalTask(() => {
    void import('./utils/plugin-runtime')
      .then(({ initializePluginRuntime }) => initializePluginRuntime())
      .catch((error) => {
        console.warn('[Plugin] deferred runtime init failed', error)
      })
  })
}

void bootstrap()
