import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import path from 'node:path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
      '@lib': path.resolve(__dirname, 'lib/index.ts'),
      '@vue/test-utils': path.resolve(__dirname, 'node_modules/@vue/test-utils/dist/vue-test-utils.esm-bundler.mjs'),
      vue: path.resolve(__dirname, 'node_modules/vue/dist/vue.runtime.esm-bundler.js'),
      'vue-router': path.resolve(__dirname, 'node_modules/vue-router/dist/vue-router.mjs'),
      axios: path.resolve(__dirname, 'node_modules/axios/index.js'),
    },
  },
  server: {
    fs: {
      allow: [path.resolve(__dirname, '..')],
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    include: ['../tests/frontend/unit/src/**/*.test.ts'],
    setupFiles: ['../tests/frontend/setup.ts'],
    coverage: {
      include: ['src/**/*.{ts,vue}', 'lib/**/*.ts'],
      exclude: [
        '../tests/frontend/**',
        'src/types/**',
        'src/vite-env.d.ts',
      ],
    },
  },
})
