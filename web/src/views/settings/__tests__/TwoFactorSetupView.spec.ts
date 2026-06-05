import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'

const setupMock = vi.fn()
const verifyMock = vi.fn()
const disableMock = vi.fn()
vi.mock('@/services/twoFactorService', () => ({
  default: {
    setup: (...a: unknown[]) => setupMock(...a),
    verify: (...a: unknown[]) => verifyMock(...a),
    disable: (...a: unknown[]) => disableMock(...a),
  },
}))

vi.mock('@/components/settings/twofactor/QrStep.vue', () => ({
  default: {
    name: 'QrStep',
    template: '<div data-testid="qr" />',
    props: ['secret', 'otpauthUrl'],
  },
}))
vi.mock('@/components/settings/twofactor/VerifyStep.vue', () => ({
  default: { name: 'VerifyStep', template: '<div data-testid="verify-step" />' },
}))
vi.mock('@/components/settings/twofactor/BackupCodesStep.vue', () => ({
  default: { name: 'BackupCodesStep', template: '<div data-testid="codes" />', props: ['codes'] },
}))

const confirmMock = vi.fn()
vi.mock('@/composables/useConfirm', () => ({
  useConfirm: (opts: unknown) => confirmMock(opts),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn(), resolve: () => ({ href: '#' }) }),
  useRoute: () => ({
    path: '/settings/security/2fa',
    params: {},
    query: {},
    name: 'SettingsSecurity2FA',
  }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

const verifyAuthMock = vi.fn()
const userRef: { value: { two_factor_enabled?: boolean } | null } = { value: null }
vi.mock('@/stores/authStore', () => ({
  useAuthStore: () => ({
    get user() {
      return userRef.value
    },
    verify: () => verifyAuthMock(),
  }),
}))

import TwoFactorSetupView from '../TwoFactorSetupView.vue'

type Vm = {
  step: string
  setup: { secret: string; otpauth_url: string } | null
  codes: string[]
  enabled: boolean
  start: () => Promise<void>
  toVerify: () => void
  onVerifySubmit: (code: string) => Promise<void>
  onDisable: () => Promise<void>
}

beforeEach(() => {
  setActivePinia(createPinia())
  setupMock.mockReset()
  verifyMock.mockReset()
  disableMock.mockReset()
  confirmMock.mockReset()
  verifyAuthMock.mockReset()
  userRef.value = null
})

describe('TwoFactorSetupView', () => {
  it('idle state when disabled shows enabled=false', () => {
    userRef.value = { two_factor_enabled: false }
    const w = mount(TwoFactorSetupView)
    const vm = w.vm as unknown as Vm
    expect(vm.step).toBe('idle')
    expect(vm.enabled).toBe(false)
  })

  it('start() calls service.setup and advances to scan', async () => {
    setupMock.mockResolvedValue({ secret: 'ABCDEF', otpauth_url: 'otpauth://test' })
    const w = mount(TwoFactorSetupView)
    const vm = w.vm as unknown as Vm
    await vm.start()
    expect(setupMock).toHaveBeenCalled()
    expect(vm.step).toBe('scan')
    expect(vm.setup?.secret).toBe('ABCDEF')
  })

  it('onVerifySubmit happy path → codes step + stores codes', async () => {
    verifyMock.mockResolvedValue({ backup_codes: ['aaa', 'bbb', 'ccc'] })
    const w = mount(TwoFactorSetupView)
    const vm = w.vm as unknown as Vm
    await vm.onVerifySubmit('123456')
    await flushPromises()
    expect(verifyMock).toHaveBeenCalledWith('123456')
    expect(vm.step).toBe('codes')
    expect(vm.codes.length).toBe(3)
  })

  it('onDisable: confirmed + code → service.disable called', async () => {
    confirmMock.mockResolvedValue(true)
    disableMock.mockResolvedValue(undefined)
    vi.spyOn(window, 'prompt').mockReturnValue('654321')
    const w = mount(TwoFactorSetupView)
    const vm = w.vm as unknown as Vm
    await vm.onDisable()
    await flushPromises()
    expect(confirmMock).toHaveBeenCalled()
    expect(disableMock).toHaveBeenCalledWith('654321')
  })

  it('onDisable: dismissed → service not called', async () => {
    confirmMock.mockResolvedValue(false)
    const w = mount(TwoFactorSetupView)
    const vm = w.vm as unknown as Vm
    await vm.onDisable()
    expect(disableMock).not.toHaveBeenCalled()
  })

  it('Lost access link renders when 2FA is enabled', async () => {
    userRef.value = { two_factor_enabled: true }
    const w = mount(TwoFactorSetupView)
    await flushPromises()
    expect(w.html()).toContain('/auth/2fa-recover')
  })
})
