import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    plugins: [react()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      port: 3008,
      strictPort: false,
      proxy: {
        '/api/admin/public': {
          target: env.VITE_API_BASE_URL || 'http://127.0.0.1:3000',
          changeOrigin: true,
          secure: false,
        },
        '/ws': {
          target: env.VITE_API_BASE_URL || 'http://127.0.0.1:3000',
          changeOrigin: true,
          ws: true,
          secure: false,
        },
      },
    },
    build: {
      outDir: './dist',
      emptyOutDir: true,
    },
  }
})
