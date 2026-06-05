import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import CalendarDayTooltip from '../CalendarDayTooltip.vue'
import type { PublicUptimeDay } from '@/types'

function mkDay(overrides: Partial<PublicUptimeDay> = {}): PublicUptimeDay {
  return {
    day: '2026-05-06',
    uptime_ratio: 0.97,
    samples: 1440,
    incidents: 2,
    downtime_seconds: 9000, // 2h 30m
    related_incidents: [
      {
        id: 'inc-1',
        title: 'Elevated errors across multiple models',
        started_at: '2026-05-06T09:00:00Z',
        resolved_at: '2026-05-06T11:30:00Z',
        severity: 'major',
      },
    ],
    ...overrides,
  }
}

describe('CalendarDayTooltip', () => {
  it('renders the day heading in human format', () => {
    const w = mount(CalendarDayTooltip, { props: { day: mkDay() } })
    expect(w.text()).toContain('6 May 2026')
  })

  it('shows "Partial outage" and downtime "2 hrs  30 mins" for major bucket', () => {
    const w = mount(CalendarDayTooltip, { props: { day: mkDay() } })
    expect(w.text()).toContain('Partial outage')
    expect(w.text()).toContain('2 hrs')
    expect(w.text()).toContain('30 mins')
  })

  it('renders related incident titles', () => {
    const w = mount(CalendarDayTooltip, { props: { day: mkDay() } })
    expect(w.text()).toContain('Elevated errors across multiple models')
  })

  it('hides the Related block when no related incidents', () => {
    const w = mount(CalendarDayTooltip, {
      props: { day: mkDay({ related_incidents: [] }) },
    })
    expect(w.text()).not.toContain('Related')
  })

  it('reads "Fully operational" with no downtime label at ratio 1', () => {
    const w = mount(CalendarDayTooltip, {
      props: { day: mkDay({ uptime_ratio: 1, downtime_seconds: 0 }) },
    })
    expect(w.text()).toContain('Fully operational')
    expect(w.text()).not.toMatch(/\d+ (hr|min)/)
  })
})
