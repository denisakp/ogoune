import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import IncidentTimeline from './IncidentTimeline.vue'
import type { IncidentEventStep } from '@/types'

const stubs = { UIcon: { template: '<span />' } }

function mkEvent(step: string, secondsAgo: number, message?: string): IncidentEventStep {
  return {
    id: `e-${step}-${secondsAgo}`,
    incident_id: 'i1',
    step,
    message,
    created_at: new Date(Date.now() - secondsAgo * 1000).toISOString(),
    updated_at: new Date(Date.now() - secondsAgo * 1000).toISOString(),
  } as unknown as IncidentEventStep
}

describe('IncidentTimeline', () => {
  it.each(['detected', 'resource_down_alert', 'alert_sent', 'resource_up_alert', 'resolved'])(
    'renders %s step with the right data-step attribute',
    (step) => {
      const w = mount(IncidentTimeline, {
        global: { stubs },
        props: { events: [mkEvent(step, 60)] },
      })
      expect(w.find(`[data-step="${step}"]`).exists()).toBe(true)
    },
  )

  it('renders events oldest first regardless of input order', () => {
    const events = [mkEvent('resolved', 0), mkEvent('detected', 600), mkEvent('alert_sent', 300)]
    const w = mount(IncidentTimeline, { global: { stubs }, props: { events } })
    const steps = w.findAll('[data-step]').map((el) => el.attributes('data-step'))
    expect(steps).toEqual(['detected', 'alert_sent', 'resolved'])
  })

  it('compact mode hides messages', () => {
    const w = mount(IncidentTimeline, {
      global: { stubs },
      props: { events: [mkEvent('detected', 60, 'Down at 02:00')], compact: true },
    })
    expect(w.text()).not.toContain('Down at 02:00')
  })

  it('unknown step falls back to neutral icon + slate color (no crash)', () => {
    const w = mount(IncidentTimeline, {
      global: { stubs },
      props: { events: [mkEvent('comment', 60, 'Note added')] },
    })
    expect(w.find('[data-step="comment"]').exists()).toBe(true)
  })

  it('empty events shows "No events yet"', () => {
    const w = mount(IncidentTimeline, { global: { stubs }, props: { events: [] } })
    expect(w.text()).toContain('No events yet')
  })
})
