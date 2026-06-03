import { afterAll, afterEach, beforeAll } from 'vitest'
import { config } from '@vue/test-utils'
import { server } from './msw/server'

// MSW server lifecycle — Spec 054 / T009. Contract: specs/054-.../contracts/mock-server.md
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

// Polyfill localStorage — @vue/devtools-kit calls localStorage.getItem at import
// time and some jsdom setups expose a non-functional stub that throws.
const store: Record<string, string> = {}
Object.defineProperty(globalThis, 'localStorage', {
  writable: true,
  value: {
    getItem: (key: string) => store[key] ?? null,
    setItem: (key: string, value: string) => { store[key] = value },
    removeItem: (key: string) => { delete store[key] },
    clear: () => { Object.keys(store).forEach(k => delete store[k]) },
    get length() { return Object.keys(store).length },
    key: (i: number) => Object.keys(store)[i] ?? null,
  },
})

if (!window.matchMedia) {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: (query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: () => undefined,
      removeListener: () => undefined,
      addEventListener: () => undefined,
      removeEventListener: () => undefined,
      dispatchEvent: () => false,
    }),
  })
}

config.global.config = config.global.config || {}
config.global.config.compilerOptions = config.global.config.compilerOptions || {}
config.global.config.compilerOptions.isCustomElement = (tag) => tag.startsWith('a-icon-')
