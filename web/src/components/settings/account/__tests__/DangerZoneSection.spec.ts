import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'

const deleteAccountMock = vi.fn()
vi.mock('@/services/accountService', () => ({
  default: { deleteAccount: (...a: unknown[]) => deleteAccountMock(...a) },
}))

const confirmMock = vi.fn()
vi.mock('@/composables/useConfirm', () => ({
  useConfirm: (opts: unknown) => confirmMock(opts),
}))

const replaceMock = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({ replace: replaceMock, push: vi.fn(), resolve: () => ({ href: '#' }) }),
  useRoute: () => ({ path: '/settings/account', params: {}, query: {}, name: 'SettingsAccount' }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

const logoutMock = vi.fn()
vi.mock('@/stores/authStore', () => ({
  useAuthStore: () => ({ user: { email: 'me@x.test' }, email: 'me@x.test', logout: logoutMock }),
}))

import DangerZoneSection from '../DangerZoneSection.vue'

type Vm = {
  open: boolean
  typed: string
  matches: boolean
  onConfirm: () => Promise<void>
}

beforeEach(() => {
  setActivePinia(createPinia())
  deleteAccountMock.mockReset()
  confirmMock.mockReset()
  replaceMock.mockReset()
  logoutMock.mockReset()
})

describe('DangerZoneSection', () => {
  it('matches is false until typed email equals user email', async () => {
    const w = mount(DangerZoneSection)
    const vm = w.vm as unknown as Vm
    vm.open = true
    expect(vm.matches).toBe(false)
    vm.typed = 'wrong@x.test'
    await flushPromises()
    expect(vm.matches).toBe(false)
    vm.typed = 'me@x.test'
    await flushPromises()
    expect(vm.matches).toBe(true)
  })

  it('does not call deleteAccount when typed email mismatches', async () => {
    const w = mount(DangerZoneSection)
    const vm = w.vm as unknown as Vm
    vm.typed = 'nope@x.test'
    await vm.onConfirm()
    expect(confirmMock).not.toHaveBeenCalled()
    expect(deleteAccountMock).not.toHaveBeenCalled()
  })

  it('aborts when useConfirm resolves false', async () => {
    confirmMock.mockResolvedValue(false)
    const w = mount(DangerZoneSection)
    const vm = w.vm as unknown as Vm
    vm.typed = 'me@x.test'
    await vm.onConfirm()
    expect(confirmMock).toHaveBeenCalled()
    expect(deleteAccountMock).not.toHaveBeenCalled()
  })

  it('deletes account, logs out, redirects to /login on confirm', async () => {
    confirmMock.mockResolvedValue(true)
    deleteAccountMock.mockResolvedValue({ message: 'ok' })
    const w = mount(DangerZoneSection)
    const vm = w.vm as unknown as Vm
    vm.typed = 'me@x.test'
    await vm.onConfirm()
    await flushPromises()
    expect(deleteAccountMock).toHaveBeenCalledWith('me@x.test')
    expect(logoutMock).toHaveBeenCalled()
    expect(replaceMock).toHaveBeenCalledWith('/login')
  })
})
