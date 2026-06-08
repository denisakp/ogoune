import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import StatusBadge from '@/components/StatusBadge.vue'

/**
 * SC-007: Exhaustive status badge component tests.
 *
 * UBadge is auto-imported from NuxtUI which renders to a <span> with utility
 * classes such as `text-success` / `bg-error/10`. We assert label text and
 * presence of a color-named class to verify the semantic mapping without
 * depending on the exact CSS layout.
 */
const includesColorClass = (wrapper: ReturnType<typeof mount>, color: string): boolean =>
  wrapper.html().includes(`text-${color}`)

describe('StatusBadge — exhaustive state rendering (SC-007)', () => {
  it('renders waiting state with neutral color and WAITING label', () => {
    const wrapper = mount(StatusBadge, { props: { status: 'waiting' } })
    expect(wrapper.text()).toBe('WAITING')
    // neutral renders without a semantic color class; assert none of the
    // semantic colors leak into a neutral state
    expect(includesColorClass(wrapper, 'success')).toBe(false)
    expect(includesColorClass(wrapper, 'warning')).toBe(false)
    expect(includesColorClass(wrapper, 'error')).toBe(false)
    wrapper.unmount()
  })

  it('renders up state with success color and UP label', () => {
    const wrapper = mount(StatusBadge, { props: { status: 'up' } })
    expect(wrapper.text()).toBe('UP')
    expect(includesColorClass(wrapper, 'success')).toBe(true)
    wrapper.unmount()
  })

  it('renders down state with error color and DOWN label', () => {
    const wrapper = mount(StatusBadge, { props: { status: 'down' } })
    expect(wrapper.text()).toBe('DOWN')
    expect(includesColorClass(wrapper, 'error')).toBe(true)
    wrapper.unmount()
  })

  it('renders paused state with warning color and PAUSED label', () => {
    const wrapper = mount(StatusBadge, { props: { status: 'paused' } })
    expect(wrapper.text()).toBe('PAUSED')
    expect(includesColorClass(wrapper, 'warning')).toBe(true)
    wrapper.unmount()
  })

  it('waiting badge does NOT use the success color', () => {
    const wrapper = mount(StatusBadge, { props: { status: 'waiting' } })
    expect(includesColorClass(wrapper, 'success')).toBe(false)
    wrapper.unmount()
  })

  it('waiting badge label is NOT ERROR', () => {
    const wrapper = mount(StatusBadge, { props: { status: 'waiting' } })
    expect(wrapper.text()).not.toBe('ERROR')
    wrapper.unmount()
  })

  it('no named state produces an ERROR label', () => {
    const namedStates = ['waiting', 'up', 'down', 'paused'] as const
    for (const status of namedStates) {
      const wrapper = mount(StatusBadge, { props: { status } })
      expect(wrapper.text()).not.toBe('ERROR')
      wrapper.unmount()
    }
  })

  it('success color is exclusively for the up state', () => {
    const nonUpStates = ['waiting', 'down', 'paused'] as const
    for (const status of nonUpStates) {
      const wrapper = mount(StatusBadge, { props: { status } })
      expect(includesColorClass(wrapper, 'success')).toBe(false)
      wrapper.unmount()
    }
  })

  it('error color is exclusively for the down state', () => {
    const nonDownStates = ['waiting', 'up', 'paused'] as const
    for (const status of nonDownStates) {
      const wrapper = mount(StatusBadge, { props: { status } })
      expect(includesColorClass(wrapper, 'error')).toBe(false)
      wrapper.unmount()
    }
  })
})
