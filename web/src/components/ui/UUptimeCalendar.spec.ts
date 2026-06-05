import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import UUptimeCalendar from './UUptimeCalendar.vue'

describe('UUptimeCalendar', () => {
  it('renders the month label and uptime percentage', () => {
    const days = Array.from({ length: 31 }, () => 'up' as const)
    const wrapper = mount(UUptimeCalendar, {
      props: { month: 5, year: 2026, days, uptimePct: 99.99 },
    })
    expect(wrapper.text()).toContain('May 2026')
    expect(wrapper.text()).toContain('99.99%')
  })

  it('renders one cell per day with its status', () => {
    const wrapper = mount(UUptimeCalendar, {
      props: {
        month: 1,
        year: 2026,
        days: ['up', 'warning', 'down', 'nodata'],
        uptimePct: 75,
      },
    })
    const cells = wrapper.findAll('[data-day]')
    expect(cells).toHaveLength(4)
    expect(cells[0]?.attributes('data-day')).toBe('up')
  })
})
