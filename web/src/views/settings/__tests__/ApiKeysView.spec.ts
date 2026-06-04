/* eslint-disable @typescript-eslint/ban-ts-comment */
// @ts-nocheck — spec 059 polish debt: index-access narrowing
import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'

const listMock = vi.fn()
const createMock = vi.fn()
const revokeMock = vi.fn()
vi.mock('@/services/accountService', () => ({
  default: {
    listAPIKeys: (...a: unknown[]) => listMock(...a),
    createAPIKey: (...a: unknown[]) => createMock(...a),
    revokeAPIKey: (...a: unknown[]) => revokeMock(...a),
  },
}))

const confirmMock = vi.fn()
vi.mock('@/composables/useConfirm', () => ({
  useConfirm: (opts: unknown) => confirmMock(opts),
}))

vi.mock('@/components/settings/apikeys/CreateKeyModal.vue', () => ({
  default: { name: 'CreateKeyModal', template: '<div data-testid="modal" />', props: ['open'] },
}))
vi.mock('@/components/settings/apikeys/OneShotRevealBanner.vue', () => ({
  default: {
    name: 'OneShotRevealBanner',
    template: '<div data-testid="banner">{{ payload.key }}</div>',
    props: ['payload'],
  },
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn(), resolve: () => ({ href: '#' }) }),
  useRoute: () => ({ path: '/settings/api-keys', params: {}, query: {}, name: 'SettingsApiKeys' }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

import ApiKeysView from '../ApiKeysView.vue'
import { useApiKeyStore } from '@/stores/useApiKeyStore'

type Vm = {
  keys: { id: string; name: string; key_prefix: string; scope: string; is_active: boolean }[]
  stats: { key: string; label: string; value: string; meta: string }[]
  store: ReturnType<typeof useApiKeyStore>
  onSubmit: (p: { name: string; scope: 'read' | 'read_write'; expiry: string }) => Promise<void>
  onRevoke: (k: { id: string; name: string }) => Promise<void>
  dismissBanner: () => void
}

beforeEach(() => {
  setActivePinia(createPinia())
  listMock.mockReset()
  createMock.mockReset()
  revokeMock.mockReset()
  confirmMock.mockReset()
})

describe('ApiKeysView', () => {
  it('Create flow stores lastCreated and banner renders with full secret', async () => {
    listMock.mockResolvedValue([])
    createMock.mockResolvedValue({
      id: 'k1',
      name: 'CI',
      key: 'pk_live_FULLSECRET',
      key_prefix: 'pk_live_FULL',
      scope: 'read',
      expires_at: null,
      created_at: '2026-06-04T10:00:00Z',
    })
    const w = mount(ApiKeysView)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    await vm.onSubmit({ name: 'CI', scope: 'read', expiry: 'never' })
    await flushPromises()
    expect(vm.store.lastCreated?.key).toBe('pk_live_FULLSECRET')
    expect(w.find('[data-testid="banner"]').text()).toContain('pk_live_FULLSECRET')
  })

  it('Reload sim (remount) → banner gone, masked prefix listed', async () => {
    listMock.mockResolvedValue([
      {
        id: 'k1',
        name: 'CI',
        key_prefix: 'pk_live_ABCD',
        scope: 'read',
        expires_at: null,
        last_used_at: null,
        last_used_ip: '',
        is_active: true,
        created_at: '2026-06-04T10:00:00Z',
      },
    ])
    const w = mount(ApiKeysView)
    await flushPromises()
    expect(w.find('[data-testid="banner"]').exists()).toBe(false)
    expect(w.text()).toContain('pk_live_ABCD')
  })

  it('Tab switch (visibilitychange) → banner still present', async () => {
    listMock.mockResolvedValue([])
    createMock.mockResolvedValue({
      id: 'k1',
      name: 'CI',
      key: 'pk_live_KEEP',
      key_prefix: 'pk_live_KEEP',
      scope: 'read',
      expires_at: null,
      created_at: '',
    })
    const w = mount(ApiKeysView)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    await vm.onSubmit({ name: 'CI', scope: 'read', expiry: 'never' })
    await flushPromises()
    document.dispatchEvent(new Event('visibilitychange'))
    await flushPromises()
    expect(vm.store.lastCreated?.key).toBe('pk_live_KEEP')
    expect(w.find('[data-testid="banner"]').exists()).toBe(true)
  })

  it('Revoke confirm body mentions "~60 seconds"', async () => {
    listMock.mockResolvedValue([
      {
        id: 'k1',
        name: 'CI',
        key_prefix: 'pk',
        scope: 'read',
        expires_at: null,
        last_used_at: null,
        last_used_ip: '',
        is_active: true,
        created_at: '',
      },
    ])
    confirmMock.mockResolvedValue(true)
    revokeMock.mockResolvedValue({ message: 'ok' })
    const w = mount(ApiKeysView)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    await vm.onRevoke({ id: 'k1', name: 'CI' })
    expect(confirmMock).toHaveBeenCalled()
    const opts = confirmMock.mock.calls[0]?.[0] as { body: string }
    expect(opts.body).toContain('~60 seconds')
  })

  it('Revoke → service called → row removed from list', async () => {
    const k = {
      id: 'k1',
      name: 'CI',
      key_prefix: 'pk',
      scope: 'read' as const,
      expires_at: null,
      last_used_at: null,
      last_used_ip: '',
      is_active: true,
      created_at: '',
    }
    listMock.mockResolvedValue([k])
    confirmMock.mockResolvedValue(true)
    revokeMock.mockResolvedValue({ message: 'ok' })
    const w = mount(ApiKeysView)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    await vm.onRevoke(k)
    expect(revokeMock).toHaveBeenCalledWith('k1')
    expect(vm.keys.find((x) => x.id === 'k1')).toBeUndefined()
  })

  it('stats renders 4 KPIs with placeholder for backend-pending metric', async () => {
    listMock.mockResolvedValue([
      {
        id: 'a',
        name: 'A',
        key_prefix: 'p',
        scope: 'read_write',
        expires_at: null,
        last_used_at: null,
        last_used_ip: '',
        is_active: true,
        created_at: '',
      },
      {
        id: 'b',
        name: 'B',
        key_prefix: 'p',
        scope: 'read',
        expires_at: null,
        last_used_at: null,
        last_used_ip: '',
        is_active: true,
        created_at: '',
      },
    ])
    const w = mount(ApiKeysView)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    expect(vm.stats[0].label).toBe('TOTAL KEYS')
    expect(vm.stats[0].value).toBe('2')
    expect(vm.stats[1].label).toBe('READ_WRITE')
    expect(vm.stats[1].value).toBe('1')
    expect(vm.stats[2].label).toBe('READ')
    expect(vm.stats[2].value).toBe('1')
    expect(vm.stats[3].value).toBe('—')
    expect(vm.stats[3].meta).toContain('Backend metric pending')
  })
})
