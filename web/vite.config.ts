import { fileURLToPath, URL } from 'node:url'
import { existsSync } from 'node:fs'

import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import ui from '@nuxt/ui/vite'
import { AntDesignVueResolver } from 'unplugin-vue-components/resolvers'

// Local resolver: `U*` components in `src/components/ui/` win over NuxtUI built-ins.
// Only resolves when the file actually exists — otherwise NuxtUI's built-in
// auto-import takes over for unknown `U*` names.
// Contract: specs/053-slice-nuxtui-foundation/contracts/component-resolver.md
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
    // NuxtUI v3 bundles `unplugin-vue-components` internally — we pass extra
    // resolvers (local `U*` shadow + AntDV cohabitation) via its `components`
    // option rather than mounting a second `Components(...)` plugin.
    ui({
      components: {
        resolvers: [
          LocalUiResolver(),
          AntDesignVueResolver({
            importStyle: 'less',
          }),
        ],
        dts: false,
      },
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      'ant-design-vue/es/time-picker/style': fileURLToPath(
        new URL('./src/antdv-timepicker-style-shim.ts', import.meta.url),
      ),
    },
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
