import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { ValidationError } from '@/core/errors'

const pushMock = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock, replace: vi.fn(), resolve: () => ({ href: '#' }) }),
  useRoute: () => ({ query: {}, params: {}, path: '/login', name: 'login' }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

const loginMock = vi.fn()
const storeRefs = {
  isLoading: false,
  requiresPasswordInit: false,
  requires2FA: false,
}
vi.mock('@/stores/authStore', () => ({
  useAuthStore: () => ({
    get isLoading() {
      return storeRefs.isLoading
    },
    get requiresPasswordInit() {
      return storeRefs.requiresPasswordInit
    },
    get requires2FA() {
      return storeRefs.requires2FA
    },
    login: loginMock,
  }),
}))

import LoginView from './LoginView.vue'

const stubs = {
  UCard: { template: '<div><slot /></div>' },
  UForm: { template: '<form><slot /></form>', props: ['schema', 'state'] },
  UFormField: { template: '<div><slot /></div>', props: ['label', 'name'] },
  UInput: { template: '<input />' },
  UButton: {
    template: '<button :disabled="loading"><slot /></button>',
    props: ['loading', 'type', 'color', 'block'],
  },
}

function build() {
  setActivePinia(createPinia())
  return mount(LoginView, { global: { stubs } })
}

beforeEach(() => {
  pushMock.mockReset()
  loginMock.mockReset()
  storeRefs.isLoading = false
  storeRefs.requiresPasswordInit = false
  storeRefs.requires2FA = false
})

describe('LoginView', () => {
  it('mounts and renders form fields', () => {
    const w = build()
    expect(w.findAll('input').length).toBeGreaterThanOrEqual(2)
    expect(w.find('button').exists()).toBe(true)
  })

  it('pushes /monitors on successful login', async () => {
    loginMock.mockResolvedValueOnce(true)
    const w = build()
    await (w.vm as unknown as { onSubmit: (p: { data: unknown }) => Promise<void> }).onSubmit({
      data: { email: 'a@b.co', password: 'pw' },
    })
    expect(loginMock).toHaveBeenCalledWith('a@b.co', 'pw')
    expect(pushMock).toHaveBeenCalledWith('/overview')
  })

  it('redirects to password init when store flags requiresPasswordInit', async () => {
    loginMock.mockResolvedValueOnce(false)
    storeRefs.requiresPasswordInit = true
    const w = build()
    await (w.vm as unknown as { onSubmit: (p: { data: unknown }) => Promise<void> }).onSubmit({
      data: { email: 'a@b.co', password: 'pw' },
    })
    expect(pushMock).toHaveBeenCalledWith('/auth/initialize-password')
  })

  it('redirects to 2FA when store flags requires2FA', async () => {
    loginMock.mockResolvedValueOnce(false)
    storeRefs.requires2FA = true
    const w = build()
    await (w.vm as unknown as { onSubmit: (p: { data: unknown }) => Promise<void> }).onSubmit({
      data: { email: 'a@b.co', password: 'pw' },
    })
    expect(pushMock).toHaveBeenCalledWith('/auth/verify-2fa')
  })

  it('does not push on failed credentials (login returns false, no redirect flags)', async () => {
    loginMock.mockResolvedValueOnce(false)
    const w = build()
    await (w.vm as unknown as { onSubmit: (p: { data: unknown }) => Promise<void> }).onSubmit({
      data: { email: 'a@b.co', password: 'wrong' },
    })
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('maps server-side ValidationError to formRef.setErrors (FR-016)', async () => {
    loginMock.mockRejectedValueOnce(
      new ValidationError('Validation failed', { email: ['Invalid'] }),
    )
    const w = build()
    const setErrors = vi.fn()
    ;(w.vm as unknown as { formRef: { setErrors: typeof setErrors } | null }).formRef = {
      setErrors,
    }
    await (w.vm as unknown as { onSubmit: (p: { data: unknown }) => Promise<void> }).onSubmit({
      data: { email: 'a@b.co', password: 'pw' },
    })
    expect(setErrors).toHaveBeenCalledWith([{ path: 'email', message: 'Invalid' }])
    expect(pushMock).not.toHaveBeenCalled()
  })
})
