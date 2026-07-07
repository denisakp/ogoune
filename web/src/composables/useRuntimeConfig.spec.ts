import { beforeEach, describe, expect, it, vi } from 'vitest'
import runtimeConfigService from '@/services/runtimeConfigService'
import {
  _resetRuntimeConfigForTest,
  loadRuntimeConfig,
  useRuntimeConfig,
} from '@/composables/useRuntimeConfig'

describe('useRuntimeConfig', () => {
  beforeEach(() => {
    _resetRuntimeConfigForTest()
    vi.restoreAllMocks()
  })

  it('first call invokes the service and caches the result', async () => {
    const spy = vi.spyOn(runtimeConfigService, 'get').mockResolvedValue({
      ssl_provider: 'letsencrypt',
      edition: 'enterprise',
      version: '1.0.0',
    })
    const cfg = await loadRuntimeConfig()
    expect(spy).toHaveBeenCalledTimes(1)
    expect(cfg.ssl_provider).toBe('letsencrypt')
    expect(useRuntimeConfig().edition).toBe('enterprise')
  })

  it('second call returns the cached value without a second fetch', async () => {
    const spy = vi.spyOn(runtimeConfigService, 'get').mockResolvedValue({
      ssl_provider: 'external',
      edition: 'community',
      version: '0.1',
    })
    await loadRuntimeConfig()
    await loadRuntimeConfig()
    expect(spy).toHaveBeenCalledTimes(1)
  })

  it('on service error falls back to safe defaults', async () => {
    vi.spyOn(runtimeConfigService, 'get').mockRejectedValue(new Error('boom'))
    const cfg = await loadRuntimeConfig()
    expect(cfg.ssl_provider).toBe('external')
    expect(cfg.edition).toBe('community')
  })
})
