import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const serverOrigin =
    env.VITE_SERVER_ORIGIN ||
    env.CHEM_SERVER_ORIGIN ||
    env.VITE_API_ORIGIN ||
    'http://127.0.0.1:8080'

  return {
    plugins: [
      vue(),
      tailwindcss()
    ],
    server: {
      host: true,
      allowedHosts: true, // true = ???????????
      port: 5000,
      proxy: {
        '/api': {
          target: serverOrigin,
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
            if (id.includes('node_modules')) {
              if (id.includes('lucide-vue-next')) return 'icons'
              if (id.includes('axios')) return 'http'
              if (
                id.includes('/vue/') ||
                id.includes('\\\\vue\\\\') ||
                id.includes('vue-router') ||
                id.includes('@vue')
              ) {
                return 'framework'
              }
              return 'vendor'
            }

            return undefined
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
  }
})
