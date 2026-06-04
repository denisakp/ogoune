import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const changePasswordMock = vi.fn()
vi.mock('@/services/accountService', () => ({
  default: { changePassword: (...a: unknown[]) => changePasswordMock(...a) },
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn(), resolve: () => ({ href: '#' }) }),
  useRoute: () => ({ path: '/', params: {}, query: {}, name: 'x' }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

import ChangePasswordSection from '../ChangePasswordSection.vue'

type Vm = {
  state: { current: string; new: string; confirm: string }
  lastResult: string
  submit: (data: { current: string; new: string; confirm: string }) => Promise<void>
}

beforeEach(() => changePasswordMock.mockReset())

describe('ChangePasswordSection', () => {
  it('submits current + new password and resets state on success', async () => {
    changePasswordMock.mockResolvedValue({ message: 'ok' })
    const w = mount(ChangePasswordSection)
    const vm = w.vm as unknown as Vm
    await vm.submit({
      current: 'oldpass',
      new: 'newverylongpassword',
      confirm: 'newverylongpassword',
    })
    await flushPromises()
    expect(changePasswordMock).toHaveBeenCalledWith('oldpass', 'newverylongpassword')
    expect(vm.state.current).toBe('')
    expect(vm.lastResult).toBe('success')
  })
})
