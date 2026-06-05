import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const getProfileMock = vi.fn()
const updateProfileMock = vi.fn()

vi.mock('@/services/accountService', () => ({
  default: {
    getProfile: (...a: unknown[]) => getProfileMock(...a),
    updateProfile: (...a: unknown[]) => updateProfileMock(...a),
  },
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn(), resolve: () => ({ href: '#' }) }),
  useRoute: () => ({ path: '/', params: {}, query: {}, name: 'x' }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

import ProfileSection from '../ProfileSection.vue'

type Vm = {
  state: { first_name: string; last_name: string; email: string; timezone: string }
  lastResult: string
  submit: (data: {
    first_name: string
    last_name: string
    email: string
    timezone: string
  }) => Promise<void>
}

beforeEach(() => {
  getProfileMock.mockReset()
  updateProfileMock.mockReset()
})

describe('ProfileSection', () => {
  it('loads profile on mount, splits name into first/last', async () => {
    getProfileMock.mockResolvedValue({
      email: 'ada@x.test',
      name: 'Ada Lovelace',
      user_id: 'u1',
      force_password_change: false,
      two_factor_enabled: false,
    })
    const w = mount(ProfileSection)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    expect(vm.state.first_name).toBe('Ada')
    expect(vm.state.last_name).toBe('Lovelace')
    expect(vm.state.email).toBe('ada@x.test')
  })

  it('calls updateProfile with concatenated name on submit', async () => {
    getProfileMock.mockResolvedValue({
      email: 'a@b.co',
      name: 'A B',
      user_id: 'u',
      force_password_change: false,
      two_factor_enabled: false,
    })
    updateProfileMock.mockResolvedValue({ email: 'a@b.co', name: 'A B' })
    const w = mount(ProfileSection)
    await flushPromises()
    const vm = w.vm as unknown as Vm
    await vm.submit({ first_name: 'A', last_name: 'B', email: 'a@b.co', timezone: 'UTC' })
    expect(updateProfileMock).toHaveBeenCalledWith('A B', 'a@b.co')
    expect(vm.lastResult).toBe('success')
  })
})
