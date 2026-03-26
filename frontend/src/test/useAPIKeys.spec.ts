import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useAPIKeys } from '@/composables/useAPIKeys'
import accountService from '@/services/accountService'

describe('useAPIKeys', () => {
  beforeEach(() => {
    vi.spyOn(accountService, 'listAPIKeys').mockResolvedValue([])
    vi.spyOn(accountService, 'createAPIKey').mockResolvedValue({
      id: 'key-1',
      name: 'CI',
      key: 'pk_live_example',
      key_prefix: 'pk_live_exam',
      scope: 'read_write',
      expires_at: null,
      created_at: new Date().toISOString(),
    })
    vi.spyOn(accountService, 'revokeAPIKey').mockResolvedValue({ message: 'API key revoked' })
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  const mountComposable = () => {
    let exposed: ReturnType<typeof useAPIKeys> | null = null

    const TestComponent = defineComponent({
      setup() {
        exposed = useAPIKeys()
        return () => null
      },
    })

    const wrapper = mount(TestComponent)
    return {
      wrapper,
      get composable() {
        if (!exposed) {
          throw new Error('Composable not initialized')
        }
        return exposed
      },
    }
  }

  it('loads keys and supports empty state', async () => {
    const { composable, wrapper } = mountComposable()
    await composable.loadKeys()

    expect(composable.keys.value).toEqual([])
    expect(accountService.listAPIKeys).toHaveBeenCalledOnce()
    wrapper.unmount()
  })

  it('creates key and exposes one-time revealed key', async () => {
    const { composable, wrapper } = mountComposable()

    await composable.createKey('CI', 'read_write')

    expect(accountService.createAPIKey).toHaveBeenCalledWith({
      name: 'CI',
      scope: 'read_write',
    })
    expect(composable.revealedKey.value).toBe('pk_live_example')
    expect(composable.revealKeyPrefix.value).toBe('pk_live_exam')

    composable.clearRevealedKey()
    expect(composable.revealedKey.value).toBeNull()
    wrapper.unmount()
  })

  it('serializes expiry date when provided', async () => {
    const { composable, wrapper } = mountComposable()
    const expiry = '2026-12-31T23:59:59.000Z'

    await composable.createKey('Temporary', 'read', expiry)

    expect(accountService.createAPIKey).toHaveBeenCalledWith({
      name: 'Temporary',
      scope: 'read',
      expires_at: expiry,
    })
    wrapper.unmount()
  })

  it('revokes key and refreshes list', async () => {
    const { composable, wrapper } = mountComposable()

    await composable.revokeKey('key-1')

    expect(accountService.revokeAPIKey).toHaveBeenCalledWith('key-1')
    expect(accountService.listAPIKeys).toHaveBeenCalled()
    wrapper.unmount()
  })
})
