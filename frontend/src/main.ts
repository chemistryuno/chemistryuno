import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './index.css'
import LiquidGlass from '@wxperia/liquid-glass-vue'

const app = createApp(App)
app.use(router)
app.use(LiquidGlass)
app.mount('#app')
