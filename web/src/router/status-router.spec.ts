import { describe, expect, it } from 'vitest'
import { createMemoryHistory, createRouter } from 'vue-router'

import { statusRoutes } from './status-router'

describe('status-router', () => {
  const buildRouter = () => createRouter({ history: createMemoryHistory(), routes: statusRoutes })

  it('resolves /status route and keeps lazy-loaded status view', () => {
    const router = buildRouter()
    const resolved = router.resolve('/status')

    expect(resolved.name).toBe('StatusPage')
    expect(typeof resolved.matched[0]?.components?.default).toBe('function')
  })

  it('resolves /status/:id route and preserves params', () => {
    const router = buildRouter()
    const resolved = router.resolve('/status/resource-123')

    expect(resolved.name).toBe('StatusPageDetail')
    expect(resolved.params.id).toBe('resource-123')
    expect(typeof resolved.matched[0]?.components?.default).toBe('function')
  })

  it('redirects / to /status', () => {
    const rootRedirect = statusRoutes.find((route) => route.path === '/')

    expect(rootRedirect?.redirect).toBe('/status')
  })

  it('redirects unknown routes to /status', () => {
    const catchAllRedirect = statusRoutes.find((route) => route.path === '/:pathMatch(.*)*')

    expect(catchAllRedirect?.redirect).toBe('/status')
  })
})
