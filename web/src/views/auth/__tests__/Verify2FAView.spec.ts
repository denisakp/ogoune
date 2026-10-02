import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const { pushMock, replaceMock, verifyMock } = vi.hoisted(() => ({
  pushMock: vi.fn(),
  replaceMock: vi.fn(),
  verifyMock: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock, replace: replaceMock, resolve: () => ({ href: '#' }) }),
  useRoute: () => ({ path: '/auth/verify-2fa', params: {}, query: {}, name: 'Verify2FA' }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

vi.mock('@/stores/authStore.ts', () => ({
  useAuthStore: () => ({
    isLoading: false,
    pending2FAEmail: 'jane@example.com',
    requires2FA: true,
    verifyTwoFactor: verifyMock,
  }),
}))

import Verify2FAView from '../Verify2FAView.vue'

type Vm = {
  otpDigits: string[]
  useBackupCode: boolean
  backupCode: string
  canSubmit: boolean
  handleVerify: () => Promise<void>
  toggleMode: () => void
}

// NuxtUI components render for real (auto-imported); only the layout is stubbed.
const stubs = {
  AuthLayout: { template: '<div><slot name="title" /><slot name="subtitle" /><slot /></div>' },
}
const PIN = 'input[aria-label^="pin input"]'

beforeEach(() => {
  pushMock.mockReset()
  verifyMock.mockReset()
})

describe('Verify2FAView', () => {
  it('submits a 6-digit authenticator code', async () => {
    verifyMock.mockResolvedValue(true)
    const w = mount(Verify2FAView, { global: { stubs } })
    const vm = w.vm as unknown as Vm
    vm.otpDigits = ['1', '2', '3', '4', '5', '6']
    await vm.handleVerify()
    await flushPromises()
    expect(verifyMock).toHaveBeenCalledWith('123456')
    expect(pushMock).toHaveBeenCalledWith('/monitors')
  })

  it('offers a backup code instead and submits it as typed', async () => {
    verifyMock.mockResolvedValue(true)
    const w = mount(Verify2FAView, { global: { stubs } })
    expect(w.find(PIN).exists()).toBe(true)

    await w.find('[data-test="toggle-backup-code"]').trigger('click')
    const vm = w.vm as unknown as Vm
    expect(vm.useBackupCode).toBe(true)
    expect(w.find('[data-test="backup-code-input"]').exists()).toBe(true)
    expect(w.find(PIN).exists()).toBe(false)

    vm.backupCode = 'abcd-efgh'
    expect(vm.canSubmit).toBe(false)
    vm.backupCode = ' ABCD-EFGH-JKMN '
    expect(vm.canSubmit).toBe(true)
    await vm.handleVerify()
    await flushPromises()
    expect(verifyMock).toHaveBeenCalledWith('ABCD-EFGH-JKMN')
    expect(pushMock).toHaveBeenCalledWith('/monitors')
  })

  it('does not submit an incomplete code', async () => {
    const w = mount(Verify2FAView, { global: { stubs } })
    const vm = w.vm as unknown as Vm
    vm.otpDigits = ['1', '2']
    await vm.handleVerify()
    expect(verifyMock).not.toHaveBeenCalled()
  })
})
