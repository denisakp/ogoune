import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import UStatusBadge from './UStatusBadge.vue'

describe('UStatusBadge', () => {
  it.each(['up', 'down', 'warning', 'maintenance', 'unknown'] as const)(
    'renders the %s status with its semantic label',
    (status) => {
      const wrapper = mount(UStatusBadge, { props: { status } })
      expect(wrapper.attributes('data-status')).toBe(status)
      expect(wrapper.text().length).toBeGreaterThan(0)
    },
  )

  it('renders an optional dot when dot=true', () => {
    const wrapper = mount(UStatusBadge, { props: { status: 'up', dot: true } })
    expect(wrapper.find('span span.rounded-full').exists()).toBe(true)
  })

  it.each(['sm', 'md', 'lg'] as const)('accepts size=%s', (size) => {
    const wrapper = mount(UStatusBadge, { props: { status: 'up', size } })
    expect(wrapper.exists()).toBe(true)
  })
})
