import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  
  return {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': resolve(__dirname, 'src')
      }
    },
    server: {
      port: 3007,
      proxy: {
        '/api': {
          target: env.VITE_API_BASE_URL || 'http://127.0.0.1:3008',
          changeOrigin: true
        },
        '/ws': {
          target: 'http://localhost:3008',
          changeOrigin: true,
          ws: true
        }
      }
    },
    build: {
      outDir: '../public',
      emptyOutDir: true
    }
  }
})

