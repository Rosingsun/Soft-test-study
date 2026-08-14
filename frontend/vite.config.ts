import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://1.12.248.91:9000',
        changeOrigin: true,
      },
      '/uploads': {
        target: 'http://1.12.248.91:9000',
        changeOrigin: true,
      },
    },
  },
})
