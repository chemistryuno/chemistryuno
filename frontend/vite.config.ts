import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss()
  ],
  server: {
    host: true,
    allowedHosts: true, // true = 允许所有主机名
    port: 5000,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
        ws: true,
        xfwd: true
      }
    }
  },
  build: {
    target: 'ES2020',
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true
      }
    },
    sourcemap: false,
    assetsInlineLimit: 4096,
    cssCodeSplit: true,
    rollupOptions: {
      output: {
        manualChunks: (id) => {
          // 第三方库分离
          if (id.includes('node_modules')) {
            if (id.includes('vue')) return 'framework'
            if (id.includes('axios')) return 'http'
            if (id.includes('lucide-vue-next')) return 'icons'
            return 'vendor'
          }
          
          // 页面按功能分割
          if (id.includes('pages')) {
            if (id.includes('Admin')) return 'admin'
            if (id.includes('Game') || id.includes('Lobby')) return 'game'
            if (id.includes('Auth') || id.includes('Login')) return 'auth'
            if (id.includes('Profile') || id.includes('UserSpace')) return 'profile'
            return 'pages'
          }
          
          if (id.includes('composables')) return 'composables'
        },
        entryFileNames: 'js/[name]-[hash].js',
        chunkFileNames: 'js/[name]-[hash].js',
        assetFileNames: (assetInfo) => {
          const ext = assetInfo.name.split('.').pop()
          if (/png|jpe?g|gif|svg|webp|ico|woff|woff2/.test(ext)) {
            return `assets/[name]-[hash][extname]`
          } else if (ext === 'css') {
            return `css/[name]-[hash][extname]`
          }
          return `assets/[name]-[hash][extname]`
        }
      }
    },
    chunkSizeWarningLimit: 500
  }
})
