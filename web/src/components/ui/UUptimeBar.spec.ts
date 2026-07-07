import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import UUptimeBar from './UUptimeBar.vue'

function mkEntry(day: string, ratio: number | null) {
  return { day, ratio }
}

describe('UUptimeBar — spec 060 thresholds', () => {
  it('1.0 → operational', () => {
    const w = mount(UUptimeBar, { props: { entries: [mkEntry('2026-06-01', 1)] } })
    expect(w.find('[data-band]').attributes('data-band')).toBe('operational')
  })

  it('≥ 0.99 → minor', () => {
    const w = mount(UUptimeBar, { props: { entries: [mkEntry('2026-06-02', 0.995)] } })
    expect(w.find('[data-band]').attributes('data-band')).toBe('minor')
  })

  it('≥ 0.95 → major', () => {
    const w = mount(UUptimeBar, { props: { entries: [mkEntry('2026-06-03', 0.97)] } })
    expect(w.find('[data-band]').attributes('data-band')).toBe('major')
  })

  it('< 0.95 → outage', () => {
    const w = mount(UUptimeBar, { props: { entries: [mkEntry('2026-06-04', 0.8)] } })
    expect(w.find('[data-band]').attributes('data-band')).toBe('outage')
  })

  it('null ratio → unknown (empty-day)', () => {
    const w = mount(UUptimeBar, { props: { entries: [mkEntry('2026-06-05', null)] } })
    expect(w.find('[data-band]').attributes('data-band')).toBe('unknown')
  })

  it('renders one cell per entry and supports compact mode', () => {
    const entries = Array.from({ length: 90 }, (_, i) => mkEntry(`2026-d${i}`, 1))
    const w = mount(UUptimeBar, { props: { entries, compact: true } })
    expect(w.findAll('[data-band]')).toHaveLength(90)
    expect(w.classes().some((c) => c.includes('h-1.5'))).toBe(true)
  })
})
