import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: {
      '@': resolve(import.meta.dirname, 'src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
      '/uploads': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
    },
  },
  // OPT-17: Vite 打包体积优化
  //  - 把 vue/vue-router/pinia 拆为独立长缓存 chunk，业务代码变更不重传
  //  - 调高 chunkSizeWarningLimit 抑制 Tailwind 产物噪声告警
  //  - 使用函数形式以兼容 Vite 8（rolldown）类型
  build: {
    chunkSizeWarningLimit: 1000,
    rollupOptions: {
      output: {
        manualChunks(id: string) {
          if (id.includes('node_modules')) {
            if (id.includes('/vue-router/') || id.includes('/pinia/') || /[\\/]vue[\\/]dist[\\/]/.test(id)) {
              return 'vue'
            }
          }
          return undefined
        },
      },
    },
  },
})
