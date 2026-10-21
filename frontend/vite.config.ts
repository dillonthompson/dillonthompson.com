import path from 'path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// Inside docker-compose the API is reachable as `api`; for native dev (Go on the
// host) run Vite with API_PROXY_TARGET=http://localhost:8080.
const apiTarget = process.env.API_PROXY_TARGET ?? 'http://api:8080'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  // Cache-busts the blog stylesheets the admin preview loads from the Go API.
  define: { __BUILD_ID__: JSON.stringify(Date.now().toString(36)) },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      '/api': {
        target: apiTarget,
        changeOrigin: true,
      },
      // Server-rendered blog pages, feed and sitemap live in the Go API.
      '/blog': { target: apiTarget, changeOrigin: true },
      '/rss.xml': { target: apiTarget, changeOrigin: true },
      '/sitemap.xml': { target: apiTarget, changeOrigin: true },
    },
  },
})
