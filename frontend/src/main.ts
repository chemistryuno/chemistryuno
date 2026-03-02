import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { initializePluginRuntime } from './utils/plugin-runtime'
import './index.css'

async function bootstrap() {
  await initializePluginRuntime()
  const app = createApp(App)
  app.use(router)
  app.mount('#app')
}

bootstrap()
