import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ValidationError } from '@/core/errors'

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn(), resolve: () => ({ href: '#' }) }),
  useRoute: () => ({ query: {}, params: {}, path: '/forgot-password', name: 'forgot' }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

const forgotMock = vi.fn()
vi.mock('@/services/authService', () => ({
  default: { forgotPassword: (e: string) => forgotMock(e) },
}))

import ForgotPasswordView from './ForgotPasswordView.vue'

const stubs = {
  UForm: { template: '<form><slot /></form>', props: ['schema', 'state'] },
  UFormGroup: { template: '<div><slot /></div>', props: ['name', 'ui'] },
  UInput: { template: '<input />' },
  UButton: { template: '<button><slot /></button>' },
  UIcon: { template: '<span />' },
}

function build() {
  return mount(ForgotPasswordView, { global: { stubs } })
}

beforeEach(() => {
  forgotMock.mockReset()
})

describe('ForgotPasswordView', () => {
  it('renders email form by default', () => {
    const w = build()
    expect(w.text()).toContain('Forgot your password?')
  })

  it('shows success state on valid submit', async () => {
    forgotMock.mockResolvedValueOnce(undefined)
    const w = build()
    await (w.vm as unknown as { onSubmit: (p: { data: unknown }) => Promise<void> }).onSubmit({
      data: { email: 'a@b.co' },
    })
    await flushPromises()
    expect((w.vm as unknown as { submitted: boolean }).submitted).toBe(true)
    expect(w.text()).toContain('Check your inbox')
  })

  it('shows same success state even on unknown email / backend rejection (FR-004)', async () => {
    forgotMock.mockRejectedValueOnce(new Error('404'))
    const w = build()
    await (w.vm as unknown as { onSubmit: (p: { data: unknown }) => Promise<void> }).onSubmit({
      data: { email: 'unknown@nope.test' },
    })
    await flushPromises()
    expect((w.vm as unknown as { submitted: boolean }).submitted).toBe(true)
  })

  it('maps 422 ValidationError to formRef.setErrors (FR-016)', async () => {
    forgotMock.mockRejectedValueOnce(
      new ValidationError('Validation failed', { email: ['Invalid format'] }),
    )
    const w = build()
    const setErrors = vi.fn()
    ;(w.vm as unknown as { formRef: { setErrors: typeof setErrors } | null }).formRef = {
      setErrors,
    }
    await (w.vm as unknown as { onSubmit: (p: { data: unknown }) => Promise<void> }).onSubmit({
      data: { email: 'a@b.co' },
    })
    expect(setErrors).toHaveBeenCalledWith([{ path: 'email', message: 'Invalid format' }])
    expect((w.vm as unknown as { submitted: boolean }).submitted).toBe(false)
  })
})
