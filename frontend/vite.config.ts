import { fileURLToPath, URL } from 'node:url'
import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  const isDevBuild = mode === 'devbuild'

  return {
    plugins: [vue()],
    build: {
      sourcemap: isDevBuild,
      minify: !isDevBuild
    },
    server: {
      allowedHosts: ['vue.local']
    },
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
        '#': fileURLToPath(new URL('./src/components', import.meta.url))
      }
    }
  }
})
