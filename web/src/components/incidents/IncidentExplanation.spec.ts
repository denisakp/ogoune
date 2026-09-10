import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import IncidentExplanation from './IncidentExplanation.vue'
import type { IncidentExplanation as Explanation } from '@/types'

const stubs = { UIcon: { template: '<span />' } }

const expl = (over: Partial<Explanation> = {}): Explanation => ({
  text: 'HTTP check failed: 502 Bad Gateway at 2026-09-10 14:03:00 UTC. The kernel OOM-killed postgres (pid 4711) on web-01 at 2026-09-10 14:02:47 UTC, 13 seconds earlier.',
  host_id: '01JHOST',
  host_name: 'web-01',
  incident_at: '2026-09-10T14:03:00Z',
  cause: 'HTTP check failed: 502 Bad Gateway',
  event: {
    id: 'e1',
    kind: 'oom_kill',
    occurredAt: '2026-09-10T14:02:47Z',
    source: 'kmsg',
    occurrences: 1,
    detail: {
      process: 'postgres',
      pid: 4711,
      cgroup: null,
      distinctProcesses: [],
      distinctTruncated: false,
    },
  },
  precedes: true,
  other_events: 0,
  window_from: '2026-09-10T13:58:00Z',
  window_to: '2026-09-10T14:04:00Z',
  ...over,
})

function build(explanation: Explanation | null) {
  return mount(IncidentExplanation, { props: { explanation }, global: { stubs } })
}

describe('IncidentExplanation', () => {
  it('renders the sentence the API produced', () => {
    const w = build(expl())
    expect(w.find('[data-test="incident-explanation-text"]').text()).toContain(
      'OOM-killed postgres (pid 4711) on web-01',
    )
  })

  // Both timestamps are what let an operator overrule the sentence. Dropping
  // one turns an interpretation into an assertion.
  it('always shows both times', () => {
    const w = build(expl())
    expect(w.find('[data-test="incident-explanation-incident-at"]').exists()).toBe(true)
    expect(w.find('[data-test="incident-explanation-event-at"]').exists()).toBe(true)
  })

  it('says which side of the failure the event fell on', () => {
    expect(build(expl()).text()).toContain('before the failure')
    expect(build(expl({ precedes: false })).text()).toContain('after the failure')
  })

  // Absence is silence, not a panel announcing that nothing was found.
  it('renders nothing at all without an explanation', () => {
    for (const value of [null, undefined]) {
      const w = build(value as never)
      expect(w.find('[data-test="incident-explanation"]').exists()).toBe(false)
      expect(w.text()).toBe('')
    }
  })

  it('counts the other events only when there are some', () => {
    expect(build(expl()).find('[data-test="incident-explanation-others"]').exists()).toBe(false)

    const one = build(expl({ other_events: 1 }))
    expect(one.find('[data-test="incident-explanation-others"]').text()).toContain(
      '1 other kernel event',
    )

    const many = build(expl({ other_events: 4 }))
    expect(many.find('[data-test="incident-explanation-others"]').text()).toContain(
      '4 other kernel events',
    )
  })

  it('falls back to the host id when the name is missing', () => {
    expect(build(expl({ host_name: '' })).text()).toContain('01JHOST')
  })

  // An operator calibrates on what the tool claims. The panel must not dress a
  // co-occurrence up as a diagnosis.
  it('states co-occurrence rather than proven cause', () => {
    const text = build(expl()).text().toLowerCase()
    expect(text).toContain('not a proven cause')
    for (const overclaim of ['confidence', 'likely cause', 'root cause', '%']) {
      expect(text).not.toContain(overclaim)
    }
  })
})
