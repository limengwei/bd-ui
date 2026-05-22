import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/ws': {
        target: 'http://127.0.0.1:3000',
        ws: true,
      },
      '/healthz': 'http://127.0.0.1:3000',
      '/api': 'http://127.0.0.1:3000',
    },
  },
  build: {
    outDir: '../web-dist',
    emptyOutDir: true,
  },
})
