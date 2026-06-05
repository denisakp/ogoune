import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import UUptimeBar from './UUptimeBar.vue'

describe('UUptimeBar', () => {
  it('renders one cell per day', () => {
    const days = Array.from({ length: 90 }, () => 'up' as const)
    const wrapper = mount(UUptimeBar, { props: { days } })
    expect(wrapper.findAll('[data-day]')).toHaveLength(90)
  })

  it('marks each cell with its status', () => {
    const wrapper = mount(UUptimeBar, {
      props: { days: ['up', 'warning', 'down', 'nodata'] },
    })
    const cells = wrapper.findAll('[data-day]')
    expect(cells.map((c) => c.attributes('data-day'))).toEqual(['up', 'warning', 'down', 'nodata'])
  })

  it('accepts compact mode', () => {
    const wrapper = mount(UUptimeBar, { props: { days: ['up'], compact: true } })
    expect(wrapper.classes().some((c) => c.includes('h-1.5'))).toBe(true)
  })
})
