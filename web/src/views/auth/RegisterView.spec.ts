import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { ValidationError } from '@/core/errors'

const pushMock = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock, replace: vi.fn(), resolve: () => ({ href: '#' }) }),
  useRoute: () => ({ query: {}, params: {}, path: '/register', name: 'register' }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

const signUpMock = vi.fn()
const storeRefs = { isLoading: false }
vi.mock('@/stores/authStore', () => ({
  useAuthStore: () => ({
    get isLoading() {
      return storeRefs.isLoading
    },
    signUp: signUpMock,
  }),
}))

const hasAccountsMock = vi.fn()
vi.mock('@/services/systemService', () => ({
  default: { hasAccounts: () => hasAccountsMock() },
}))

import RegisterView from './RegisterView.vue'

const stubs = {
  UForm: { template: '<form><slot /></form>', props: ['schema', 'state'] },
  UFormField: { template: '<div><slot /></div>', props: ['name', 'ui'] },
  UInput: { template: '<input />' },
  UButton: { template: '<button><slot /></button>' },
  UCheckbox: { template: '<input type="checkbox" />' },
  UIcon: { template: '<span />' },
}

function build() {
  setActivePinia(createPinia())
  return mount(RegisterView, { global: { stubs } })
}

beforeEach(() => {
  pushMock.mockReset()
  signUpMock.mockReset()
  hasAccountsMock.mockReset()
  storeRefs.isLoading = false
})

describe('RegisterView', () => {
  it('shows admin note when has_accounts === false', async () => {
    hasAccountsMock.mockResolvedValueOnce(false)
    const w = build()
    await flushPromises()
    expect(w.text()).toContain('First account becomes the admin')
  })

  it('hides admin note when has_accounts === true', async () => {
    hasAccountsMock.mockResolvedValueOnce(true)
    const w = build()
    await flushPromises()
    expect(w.text()).not.toContain('First account becomes the admin')
  })

  it('falls back to showing the note when probe throws (safe fallback)', async () => {
    hasAccountsMock.mockRejectedValueOnce(new Error('boom'))
    const w = build()
    await flushPromises()
    expect(w.text()).toContain('First account becomes the admin')
  })

  it('pushes /monitors on successful signUp', async () => {
    hasAccountsMock.mockResolvedValueOnce(false)
    signUpMock.mockResolvedValueOnce(true)
    const w = build()
    await flushPromises()
    await (w.vm as unknown as { onSubmit: (p: { data: unknown }) => Promise<void> }).onSubmit({
      data: {
        email: 'a@b.co',
        password: 'longenough12chars',
        confirmPassword: 'longenough12chars',
        newsletter: true,
      },
    })
    expect(signUpMock).toHaveBeenCalled()
    expect(pushMock).toHaveBeenCalledWith('/overview')
  })

  it('maps server-side ValidationError to formRef.setErrors (FR-016)', async () => {
    hasAccountsMock.mockResolvedValueOnce(false)
    signUpMock.mockRejectedValueOnce(
      new ValidationError('Validation failed', { email: ['Already taken'] }),
    )
    const w = build()
    await flushPromises()
    const setErrors = vi.fn()
    ;(w.vm as unknown as { formRef: { setErrors: typeof setErrors } | null }).formRef = {
      setErrors,
    }
    await (w.vm as unknown as { onSubmit: (p: { data: unknown }) => Promise<void> }).onSubmit({
      data: {
        email: 'a@b.co',
        password: 'longenough12chars',
        confirmPassword: 'longenough12chars',
        newsletter: false,
      },
    })
    expect(setErrors).toHaveBeenCalledWith([{ path: 'email', message: 'Already taken' }])
    expect(pushMock).not.toHaveBeenCalled()
  })
})
