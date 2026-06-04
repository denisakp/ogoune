import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import NotificationsPanel from './NotificationsPanel.vue'
import type { IncidentEventStep } from '@/types'

const stubs = { UIcon: { template: '<span />' } }

function mkEvent(step: string, message?: string): IncidentEventStep {
  return {
    id: `e-${step}-${Math.random()}`,
    incident_id: 'i1',
    step,
    message,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  } as unknown as IncidentEventStep
}

describe('NotificationsPanel', () => {
  it('renders alert_sent events with Delivered badge', () => {
    const w = mount(NotificationsPanel, {
      global: { stubs },
      props: {
        events: [mkEvent('alert_sent', 'Slack channel #ops'), mkEvent('alert_sent', 'Email')],
      },
    })
    expect(w.text()).toContain('Slack channel #ops')
    expect(w.text()).toContain('Delivered')
  })

  it('filters out non-alert events', () => {
    const w = mount(NotificationsPanel, {
      global: { stubs },
      props: {
        events: [
          mkEvent('detected', 'Down'),
          mkEvent('alert_sent', 'Notif #1'),
          mkEvent('resolved', 'OK'),
        ],
      },
    })
    expect(w.text()).toContain('Notif #1')
    expect(w.text()).not.toContain('Down')
    expect(w.text()).not.toContain('OK')
  })

  it('empty state when no alert_sent events', () => {
    const w = mount(NotificationsPanel, {
      global: { stubs },
      props: { events: [mkEvent('detected')] },
    })
    expect(w.text()).toContain('No notifications were dispatched')
  })
})
