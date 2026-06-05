import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import UUptimeCalendar from './UUptimeCalendar.vue'

function mkEntry(day: string, ratio: number | null) {
  return { day, ratio }
}

describe('UUptimeCalendar — spec 060', () => {
  it('renders leading blank cells for the first weekday offset', () => {
    // June 2026: 1st is a Monday → 0 leading blanks (ISO week start).
    // May 2026: 1st is a Friday → 4 leading blanks.
    const w = mount(UUptimeCalendar, {
      props: { year: 2026, month: 5, entries: [] },
    })
    const blanks = w.findAll('[data-blank="1"]')
    expect(blanks.length).toBe(4)
  })

  it('future-month cells render with unknown band when no entries provided', () => {
    const w = mount(UUptimeCalendar, {
      props: { year: 2099, month: 1, entries: [] },
    })
    const dayCells = w.findAll('[data-day-num]')
    expect(dayCells.length).toBe(31)
    for (const c of dayCells) {
      expect(c.attributes('data-band')).toBe('unknown')
    }
  })

  it('computes monthly % as the mean of known daily ratios', () => {
    const entries = [
      mkEntry('2026-05-01', 1),
      mkEntry('2026-05-02', 0.98),
      mkEntry('2026-05-03', null), // ignored
    ]
    const w = mount(UUptimeCalendar, {
      props: { year: 2026, month: 5, entries },
    })
    // (1 + 0.98) / 2 = 0.99 → 99.00%
    expect(w.text()).toContain('99.00%')
  })

  it('maps each band to the right cell per FR-004 thresholds', () => {
    const entries = [
      mkEntry('2026-05-01', 1), // operational
      mkEntry('2026-05-02', 0.99), // minor
      mkEntry('2026-05-03', 0.96), // major
      mkEntry('2026-05-04', 0.5), // outage
    ]
    const w = mount(UUptimeCalendar, {
      props: { year: 2026, month: 5, entries },
    })
    const cells = w.findAll('[data-day-num]')
    const byDay = (n: number) => cells.find((c) => c.attributes('data-day-num') === String(n))
    expect(byDay(1)?.attributes('data-band')).toBe('operational')
    expect(byDay(2)?.attributes('data-band')).toBe('minor')
    expect(byDay(3)?.attributes('data-band')).toBe('major')
    expect(byDay(4)?.attributes('data-band')).toBe('outage')
  })
})
