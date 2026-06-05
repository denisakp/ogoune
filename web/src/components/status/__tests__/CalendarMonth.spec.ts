import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import CalendarMonth from '../CalendarMonth.vue'

function mkDay(day: string, ratio: number, samples = 100) {
  return { day, uptime_ratio: ratio, samples, incidents: 0 }
}

describe('CalendarMonth — FR-008 threshold colors', () => {
  it('1.0 → operational', () => {
    const w = mount(CalendarMonth, {
      props: { year: 2026, month: 5, days: [mkDay('2026-05-01', 1)] },
    })
    expect(w.find('[data-day-num="1"]').attributes('data-band')).toBe('operational')
  })

  it('≥ 0.99 → minor', () => {
    const w = mount(CalendarMonth, {
      props: { year: 2026, month: 5, days: [mkDay('2026-05-02', 0.995)] },
    })
    expect(w.find('[data-day-num="2"]').attributes('data-band')).toBe('minor')
  })

  it('≥ 0.95 → major', () => {
    const w = mount(CalendarMonth, {
      props: { year: 2026, month: 5, days: [mkDay('2026-05-03', 0.97)] },
    })
    expect(w.find('[data-day-num="3"]').attributes('data-band')).toBe('major')
  })

  it('< 0.95 → outage', () => {
    const w = mount(CalendarMonth, {
      props: { year: 2026, month: 5, days: [mkDay('2026-05-04', 0.5)] },
    })
    expect(w.find('[data-day-num="4"]').attributes('data-band')).toBe('outage')
  })
})
