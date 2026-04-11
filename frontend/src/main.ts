import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './index.css'

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

function bootstrap() {
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

bootstrap()
