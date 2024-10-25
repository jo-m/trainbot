import { fileURLToPath, URL } from 'node:url'

import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import vuetify, { transformAssetUrls } from 'vite-plugin-vuetify'

export default ({ mode }) => {
  process.env = { ...process.env, ...loadEnv(mode, process.cwd()) }

  // https://vitejs.dev/config/
  return defineConfig({
    plugins: [
      vue({
        template: { transformAssetUrls }
      }),
      vuetify({
        autoImport: true,
        styles: {
          configFile: 'src/styles/settings.scss'
        }
      })
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url))
      }
    },
    server: {
      proxy: {
        '/data': {
          target: process.env.VITE_FRONTEND_BASE_URL,
          changeOrigin: true,
          secure: false
        }
      }
    }
  })
}
