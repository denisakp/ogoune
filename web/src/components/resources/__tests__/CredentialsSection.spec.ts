import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import CredentialsSection from '@/components/resources/CredentialsSection.vue'

/**
 * Feature 028 — FR-012: credentials section is visible only for protocol types
 * that accept authentication. For all other types it renders nothing.
 */
describe('CredentialsSection', () => {
  it.each(['redis', 'mysql', 'postgres'] as const)(
    'renders for protocol type %s',
    (protocolType) => {
      const wrapper = mount(CredentialsSection, {
        props: { protocolType, modelValue: null },
      })
      expect(wrapper.find('[data-testid="credentials-section"]').exists()).toBe(true)
      expect(wrapper.find('[data-testid="credentials-username"]').exists()).toBe(true)
      expect(wrapper.find('[data-testid="credentials-password"]').exists()).toBe(true)
      wrapper.unmount()
    },
  )

  it.each(['mongodb', 'ftp', 'ssh', 'http', 'tcp', undefined] as const)(
    'does not render for protocol type %s',
    (protocolType) => {
      const wrapper = mount(CredentialsSection, {
        props: { protocolType, modelValue: null },
      })
      expect(wrapper.find('[data-testid="credentials-section"]').exists()).toBe(false)
      wrapper.unmount()
    },
  )

  it('emits update:modelValue when username changes', async () => {
    const wrapper = mount(CredentialsSection, {
      props: { protocolType: 'redis', modelValue: { password: '' } },
    })
    // a-input renders <input> directly under the [data-testid] wrapper.
    const usernameInput = wrapper.find<HTMLInputElement>(
      'input[data-testid="credentials-username"]',
    )
    if (usernameInput.exists()) {
      await usernameInput.setValue('monitor')
    } else {
      // Fall back to the inner input (Ant Design may wrap with a span).
      await wrapper
        .find<HTMLInputElement>('[data-testid="credentials-username"] input')
        .setValue('monitor')
    }
    const events = wrapper.emitted('update:modelValue')
    expect(events).toBeTruthy()
    const eventListA = events ?? []
    const lastA = eventListA[eventListA.length - 1]
    expect(lastA).toBeDefined()
    const payloadA = (lastA ?? [])[0] as { username?: string; password: string }
    expect(payloadA.username).toBe('monitor')
  })

  it('emits update:modelValue when password changes', async () => {
    const wrapper = mount(CredentialsSection, {
      props: { protocolType: 'redis', modelValue: { password: '' } },
    })
    const passwordInput = wrapper.find<HTMLInputElement>(
      'input[data-testid="credentials-password"]',
    )
    if (passwordInput.exists()) {
      await passwordInput.setValue('s3cret')
    } else {
      await wrapper
        .find<HTMLInputElement>('[data-testid="credentials-password"] input')
        .setValue('s3cret')
    }
    const events = wrapper.emitted('update:modelValue')
    expect(events).toBeTruthy()
    const eventListB = events ?? []
    const lastB = eventListB[eventListB.length - 1]
    expect(lastB).toBeDefined()
    const payloadB = (lastB ?? [])[0] as { password: string }
    expect(payloadB.password).toBe('s3cret')
  })

  it('renders the "Remove credentials" button only when one already exists', () => {
    const without = mount(CredentialsSection, {
      props: { protocolType: 'redis', modelValue: null, hasExistingCredential: false },
    })
    expect(without.find('[data-testid="credentials-clear"]').exists()).toBe(false)
    without.unmount()

    const withExisting = mount(CredentialsSection, {
      props: { protocolType: 'redis', modelValue: null, hasExistingCredential: true },
    })
    expect(withExisting.find('[data-testid="credentials-clear"]').exists()).toBe(true)
    withExisting.unmount()
  })

  it('emits clear when the Remove credentials button is clicked', async () => {
    const wrapper = mount(CredentialsSection, {
      props: { protocolType: 'redis', modelValue: null, hasExistingCredential: true },
    })
    await wrapper.find('[data-testid="credentials-clear"]').trigger('click')
    expect(wrapper.emitted('clear')).toBeTruthy()
    wrapper.unmount()
  })
})
