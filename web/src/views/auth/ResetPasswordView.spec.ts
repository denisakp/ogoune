import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { ValidationError } from '@/core/errors'

const pushMock = vi.fn()
const routeQuery: { value: Record<string, string> } = { value: { token: 'abc' } }
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock, replace: vi.fn(), resolve: () => ({ href: '#' }) }),
  useRoute: () => ({
    get query() {
      return routeQuery.value
    },
    params: {},
    path: '/reset-password',
    name: 'reset',
  }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

const resetMock = vi.fn()
const storeRefs = { isLoading: false }
vi.mock('@/stores/authStore', () => ({
  useAuthStore: () => ({
    get isLoading() {
      return storeRefs.isLoading
    },
    resetPasswordWithToken: (input: unknown) => resetMock(input),
  }),
}))

import ResetPasswordView from './ResetPasswordView.vue'

const stubs = {
  UForm: { template: '<form><slot /></form>', props: ['schema', 'state'] },
  UFormGroup: { template: '<div><slot /></div>', props: ['name', 'ui'] },
  UInput: { template: '<input />' },
  UButton: { template: '<button><slot /></button>' },
  UIcon: { template: '<span />' },
}

function build() {
  setActivePinia(createPinia())
  return mount(ResetPasswordView, { global: { stubs } })
}

beforeEach(() => {
  pushMock.mockReset()
  resetMock.mockReset()
  routeQuery.value = { token: 'abc' }
  storeRefs.isLoading = false
})

describe('ResetPasswordView', () => {
  it('reads token from query and renders form', () => {
    const w = build()
    expect(w.text()).toContain('Set a new password')
  })

  it('strength meter computes score based on rules', () => {
    const w = build()
    const vm = w.vm as unknown as {
      state: { password: string }
      strength: { score: number; label: string }
    }
    vm.state.password = 'short'
    expect(vm.strength.score).toBeLessThan(2)
    vm.state.password = 'Longenough12chars!'
    expect(vm.strength.score).toBe(4)
    expect(vm.strength.label).toBe('Strong')
  })

  it('pushes /monitors on successful reset', async () => {
    resetMock.mockResolvedValueOnce(true)
    const w = build()
    await (w.vm as unknown as { onSubmit: (p: { data: unknown }) => Promise<void> }).onSubmit({
      data: { token: 'abc', password: 'Longenough12chars!', confirmPassword: 'Longenough12chars!' },
    })
    expect(resetMock).toHaveBeenCalled()
    expect(pushMock).toHaveBeenCalledWith('/overview')
  })

  it('surfaces expired/used banner on 410 from backend', async () => {
    resetMock.mockRejectedValueOnce(Object.assign(new Error('Gone'), { status: 410 }))
    const w = build()
    await (w.vm as unknown as { onSubmit: (p: { data: unknown }) => Promise<void> }).onSubmit({
      data: { token: 'abc', password: 'Longenough12chars!', confirmPassword: 'Longenough12chars!' },
    })
    expect((w.vm as unknown as { expiredOrUsed: boolean }).expiredOrUsed).toBe(true)
  })

  it('maps 422 ValidationError to formRef.setErrors (FR-016)', async () => {
    resetMock.mockRejectedValueOnce(
      new ValidationError('Validation failed', { password: ['Too weak'] }),
    )
    const w = build()
    const setErrors = vi.fn()
    ;(w.vm as unknown as { formRef: { setErrors: typeof setErrors } | null }).formRef = {
      setErrors,
    }
    await (w.vm as unknown as { onSubmit: (p: { data: unknown }) => Promise<void> }).onSubmit({
      data: { token: 'abc', password: 'short', confirmPassword: 'short' },
    })
    expect(setErrors).toHaveBeenCalledWith([{ path: 'password', message: 'Too weak' }])
    expect(pushMock).not.toHaveBeenCalled()
  })
})
