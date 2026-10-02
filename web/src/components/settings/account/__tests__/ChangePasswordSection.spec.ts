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
import { ValidationError } from '@/core/errors'

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

  // The server answers a wrong current password with 422 + a field error --
  // not 401, which would sign the user out. The form shows it and stays put.
  it('a wrong current password is a field error, not a sign-out', async () => {
    changePasswordMock.mockRejectedValueOnce(
      new ValidationError('Invalid current password', { current: ['Invalid current password'] }),
    )
    const w = mount(ChangePasswordSection)
    const setErrors = vi.fn()
    ;(w.vm as unknown as { formRef: { setErrors: typeof setErrors } | null }).formRef = { setErrors }
    const vm = w.vm as unknown as Vm
    await vm.submit({ current: 'typo', new: 'newverylongpassword', confirm: 'newverylongpassword' })
    expect(setErrors).toHaveBeenCalledWith([{ path: 'current', message: 'Invalid current password' }])
    expect(vm.lastResult).toBe('server-error')
  })
})
