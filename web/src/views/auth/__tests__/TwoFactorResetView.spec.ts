import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const confirmResetMock = vi.fn()
vi.mock('@/services/twoFactorService', () => ({
  default: { confirmReset: (...a: unknown[]) => confirmResetMock(...a) },
}))

const replaceMock = vi.fn()
const routeRef = { query: { token: 'abc' } as Record<string, unknown> }
vi.mock('vue-router', () => ({
  useRouter: () => ({ replace: replaceMock, push: vi.fn(), resolve: () => ({ href: '#' }) }),
  useRoute: () => routeRef,
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

import TwoFactorResetView from '../TwoFactorResetView.vue'

type Vm = { state: 'pending' | 'success' | 'error'; errorMessage: string | null }

beforeEach(() => {
  confirmResetMock.mockReset()
  replaceMock.mockReset()
  routeRef.query = { token: 'abc' }
})

describe('TwoFactorResetView', () => {
  it('valid token → confirmReset called and redirects to re-setup', async () => {
    confirmResetMock.mockResolvedValue({ token: 'jwt-token', session_id: 'u1' })
    const w = mount(TwoFactorResetView)
    await flushPromises()
    expect(confirmResetMock).toHaveBeenCalledWith('abc')
    expect(replaceMock).toHaveBeenCalledWith('/settings/security/2fa?action=re-setup')
    const vm = w.vm as unknown as Vm
    expect(vm.state).toBe('success')
  })

  it('410 error → renders error card', async () => {
    confirmResetMock.mockRejectedValue(new Error('410'))
    const w = mount(TwoFactorResetView)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    expect(vm.state).toBe('error')
    expect(w.html()).toContain('Reset link no longer valid')
  })
})
