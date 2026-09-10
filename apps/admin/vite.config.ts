import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// 后台构建为纯静态产物，由 Nginx 托管
// 后端地址：默认本机 8888；联调远程服务器时用 API_TARGET 覆盖，例如
//   API_TARGET=http://1.2.3.4:8888 pnpm dev:admin
// 这样浏览器只与 localhost:5173 通信，由 Vite 代理转发，不产生跨域
const API_TARGET = process.env.API_TARGET || 'http://54.176.222.252:8888'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: API_TARGET,
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
  },
})
