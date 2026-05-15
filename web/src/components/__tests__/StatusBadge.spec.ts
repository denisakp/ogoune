import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import StatusBadge from '@/components/StatusBadge.vue'

/**
 * SC-007: Exhaustive status badge component tests.
 *
 * Each test asserts the correct Ant Design color prop and label text for
 * all four named monitor states: waiting, up, down, paused.
 *
 * Also verifies that no named state produces an "ERROR" label or an
 * unintended badge color (FR-007 through FR-012).
 */
describe('StatusBadge — exhaustive state rendering (SC-007)', () => {
  // Ant Design renders color as CSS class: color="green" → ant-tag-green, color="default" → ant-tag-default

  it('renders waiting state with neutral color and WAITING label', () => {
    // FR-007: waiting → grey/neutral badge (ant-tag-default) + "WAITING" label
    const wrapper = mount(StatusBadge, { props: { status: 'waiting' } })
    expect(wrapper.text()).toBe('WAITING')
    expect(wrapper.html()).toContain('ant-tag-default')
    wrapper.unmount()
  })

  it('renders up state with green color and UP label', () => {
    // FR-009: green badge (ant-tag-green) exclusively for up state
    const wrapper = mount(StatusBadge, { props: { status: 'up' } })
    expect(wrapper.text()).toBe('UP')
    expect(wrapper.html()).toContain('ant-tag-green')
    wrapper.unmount()
  })

  it('renders down state with red color and DOWN label', () => {
    // FR-010: red badge (ant-tag-red) exclusively for down state
    const wrapper = mount(StatusBadge, { props: { status: 'down' } })
    expect(wrapper.text()).toBe('DOWN')
    expect(wrapper.html()).toContain('ant-tag-red')
    wrapper.unmount()
  })

  it('renders paused state with orange color and PAUSED label', () => {
    // FR-012: paused badge behavior unchanged (ant-tag-orange)
    const wrapper = mount(StatusBadge, { props: { status: 'paused' } })
    expect(wrapper.text()).toBe('PAUSED')
    expect(wrapper.html()).toContain('ant-tag-orange')
    wrapper.unmount()
  })

  it('waiting badge is NOT green (ant-tag-green must not appear)', () => {
    // FR-007: waiting must never display a green badge
    const wrapper = mount(StatusBadge, { props: { status: 'waiting' } })
    expect(wrapper.html()).not.toContain('ant-tag-green')
    wrapper.unmount()
  })

  it('waiting badge label is NOT ERROR', () => {
    // FR-008: waiting must never display "ERROR" as its label
    const wrapper = mount(StatusBadge, { props: { status: 'waiting' } })
    expect(wrapper.text()).not.toBe('ERROR')
    wrapper.unmount()
  })

  it('no named state produces an ERROR label', () => {
    // FR-008 / FR-011: exhaustive check — none of the four named states renders "ERROR"
    const namedStates = ['waiting', 'up', 'down', 'paused'] as const
    for (const status of namedStates) {
      const wrapper = mount(StatusBadge, { props: { status } })
      expect(wrapper.text()).not.toBe('ERROR')
      wrapper.unmount()
    }
  })

  it('green badge is exclusively for the up state', () => {
    // FR-009: ant-tag-green must not appear for waiting, down, or paused
    const nonUpStates = ['waiting', 'down', 'paused'] as const
    for (const status of nonUpStates) {
      const wrapper = mount(StatusBadge, { props: { status } })
      expect(wrapper.html()).not.toContain('ant-tag-green')
      wrapper.unmount()
    }
  })

  it('red badge is exclusively for the down state', () => {
    // FR-010: ant-tag-red must not appear for waiting, up, or paused
    const nonDownStates = ['waiting', 'up', 'paused'] as const
    for (const status of nonDownStates) {
      const wrapper = mount(StatusBadge, { props: { status } })
      expect(wrapper.html()).not.toContain('ant-tag-red')
      wrapper.unmount()
    }
  })
})
