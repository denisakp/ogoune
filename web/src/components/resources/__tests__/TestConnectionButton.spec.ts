import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import TestConnectionButton from '@/components/resources/TestConnectionButton.vue'

const { testCredentialMock, retryAfterSecondsMock } = vi.hoisted(() => ({
  testCredentialMock: vi.fn(),
  retryAfterSecondsMock: vi.fn(),
}))

vi.mock('@/services/credentialService', () => ({
  testCredential: testCredentialMock,
  retryAfterSeconds: retryAfterSecondsMock,
}))

beforeEach(() => {
  testCredentialMock.mockReset()
  retryAfterSecondsMock.mockReset().mockReturnValue(null)
})

afterEach(() => {
  vi.restoreAllMocks()
})

describe('TestConnectionButton', () => {
  it('is disabled until a resourceId and password are provided', () => {
    const wrapper = mount(TestConnectionButton, {
      props: { resourceId: undefined, payload: null },
      global: { stubs: { UTooltip: { template: '<div><slot /></div>' } } },
    })
    const btn = wrapper.find('[data-testid="test-connection-button"]')
    expect(btn.attributes('disabled')).toBeDefined()
    wrapper.unmount()
  })

  it('shows a success alert when the test succeeds', async () => {
    testCredentialMock.mockResolvedValue({ status: 'ok', latency_ms: 42 })
    const wrapper = mount(TestConnectionButton, {
      props: { resourceId: 'r1', payload: { password: 's3cret' } },
      global: { stubs: { UTooltip: { template: '<div><slot /></div>' } } },
    })
    await wrapper.find('[data-testid="test-connection-button"]').trigger('click')
    await flushPromises()
    const alert = wrapper.find('[data-testid="test-connection-result"]')
    expect(alert.exists()).toBe(true)
    expect(alert.text()).toContain('Connection successful')
    expect(alert.text()).toContain('42 ms')
    wrapper.unmount()
  })

  it('shows a failure alert with the cause when the test fails', async () => {
    testCredentialMock.mockResolvedValue({ status: 'failed', cause: 'auth_failed', latency_ms: 12 })
    const wrapper = mount(TestConnectionButton, {
      props: { resourceId: 'r1', payload: { password: 'wrong' } },
      global: { stubs: { UTooltip: { template: '<div><slot /></div>' } } },
    })
    await wrapper.find('[data-testid="test-connection-button"]').trigger('click')
    await flushPromises()
    const alert = wrapper.find('[data-testid="test-connection-result"]')
    expect(alert.exists()).toBe(true)
    expect(alert.text()).toContain('auth_failed')
    wrapper.unmount()
  })

  it('surfaces the rate-limit error with retry-after seconds', async () => {
    testCredentialMock.mockRejectedValue(new Error('429'))
    retryAfterSecondsMock.mockReturnValue(45)
    const wrapper = mount(TestConnectionButton, {
      props: { resourceId: 'r1', payload: { password: 's3cret' } },
      global: { stubs: { UTooltip: { template: '<div><slot /></div>' } } },
    })
    await wrapper.find('[data-testid="test-connection-button"]').trigger('click')
    await flushPromises()
    const err = wrapper.find('[data-testid="test-connection-error"]')
    expect(err.exists()).toBe(true)
    expect(err.text()).toContain('45 seconds')
    wrapper.unmount()
  })
})
