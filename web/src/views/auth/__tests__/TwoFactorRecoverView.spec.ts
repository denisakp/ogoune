import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const requestResetMock = vi.fn()
vi.mock('@/services/twoFactorService', () => ({
  default: { requestReset: (...a: unknown[]) => requestResetMock(...a) },
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn(), resolve: () => ({ href: '#' }) }),
  useRoute: () => ({ path: '/auth/2fa-recover', params: {}, query: {}, name: 'TwoFactorRecover' }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

import TwoFactorRecoverView from '../TwoFactorRecoverView.vue'

type Vm = { email: string; submitted: boolean; onSubmit: () => Promise<void> }

beforeEach(() => requestResetMock.mockReset())

describe('TwoFactorRecoverView', () => {
  it('calls service.requestReset with trimmed email and renders anti-enumeration copy', async () => {
    requestResetMock.mockResolvedValue(undefined)
    const w = mount(TwoFactorRecoverView)
    const vm = w.vm as unknown as Vm
    vm.email = '  Me@x.test  '
    await vm.onSubmit()
    await flushPromises()
    expect(requestResetMock).toHaveBeenCalledWith('Me@x.test')
    expect(vm.submitted).toBe(true)
    expect(w.html()).toContain('If this email is registered')
  })

  it('skips service call when email is empty', async () => {
    const w = mount(TwoFactorRecoverView)
    const vm = w.vm as unknown as Vm
    await vm.onSubmit()
    expect(requestResetMock).not.toHaveBeenCalled()
    expect(vm.submitted).toBe(false)
  })
})
