import { fileURLToPath, URL } from 'node:url'
import { existsSync } from 'node:fs'

import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import ui from '@nuxt/ui/vite'

// Local resolver: `U*` components in `src/components/ui/` win over NuxtUI built-ins.
const LocalUiResolver = () => ({
  type: 'component' as const,
  resolve(name: string) {
    if (!/^U[A-Z]/.test(name)) return
    const path = fileURLToPath(new URL(`./src/components/ui/${name}.vue`, import.meta.url))
    if (!existsSync(path)) return
    return { name: 'default', from: path }
  },
})

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
    ui({
      ui: {
        colors: {
          primary: 'indigo',
          neutral: 'slate',
        },
      },
      components: {
        resolvers: [LocalUiResolver()],
        dts: false,
      },
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
    // Force a single resolved instance for libraries that NuxtUI v4 imports
    // (e.g. createRef was added in @vueuse/core 14; vaul-vue still pins 10).
    // Without this, vite optimizeDeps may pre-bundle the older copy and
    // NuxtUI's Table.vue crashes silently — vue-router swallows the
    // SyntaxError, the page renders blank, no error reaches the user.
    dedupe: ['@vueuse/core', 'vue', 'vue-router', '@nuxt/ui'],
  },
  server: {
    proxy: {
      // Same-origin proxy: browser hits `/api/*` on the Vite dev server,
      // which forwards to the backend. Avoids CORS entirely in dev.
      // Set `VITE_API_BASE_URL=/api` in `.env.local`.
      '/api': {
        target: 'http://localhost:9596',
        changeOrigin: true,
      },
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['src/test/setup.ts'],
    include: ['src/**/*.spec.ts'],
    // Stable base URL for tests so Ky's `prefix` resolves to a known origin
    // that MSW handler patterns (`*/path`) can match.
    env: {
      VITE_API_BASE_URL: 'http://test.local/api/',
    },
  },
  build: {
    rollupOptions: {
      input: {
        main: fileURLToPath(new URL('./index.html', import.meta.url)),
        status: fileURLToPath(new URL('./status.html', import.meta.url)),
      },
    },
  },
})
