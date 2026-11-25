import { fileURLToPath, URL } from 'node:url'

import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import nightwatchPlugin from 'vite-plugin-nightwatch'

// https://vitejs.dev/config/

export default defineConfig(({ command, mode }) => {
  // Load env file based on `mode` in the current working directory.
  // Set the third parameter to '' to load all env regardless of the `VITE_` prefix.
  const env = loadEnv(mode, process.cwd(), '')
  let port = 8000
  if (env.VITE_GINFRA_PORT) {
    port = parseInt(env.VITE_GINFRA_PORT)
  }
  let host = "0.0.0.0"
  //if (env.SERVICE_HOST) {
  //  host = env.SERVICE_HOST
  //}
  return {
    base: './',
    plugins: [
      vue(),
      nightwatchPlugin(),
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url))
      }
    },
    server: {
      port: port,
      host: "localhost"
    }
  }
})

