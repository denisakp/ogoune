import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import HostEventsList from './HostEventsList.vue'
import type { HostEvent } from '@/types'

const stubs = { UIcon: { template: '<span />' } }

const evt = (over: Partial<HostEvent> = {}): HostEvent => ({
  id: 'e1',
  kind: 'oom_kill',
  occurredAt: '2026-09-09T14:02:47Z',
  source: 'kmsg',
  occurrences: 1,
  detail: {
    process: 'postgres',
    pid: 4711,
    cgroup: null,
    distinctProcesses: [],
    distinctTruncated: false,
  },
  ...over,
})

function build(events: HostEvent[]) {
  return mount(HostEventsList, { props: { events }, global: { stubs } })
}

describe('HostEventsList', () => {
  // T021 — the populated case.
  it('renders the kind, the process and when it happened', () => {
    const w = build([evt()])
    const row = w.get('[data-test="host-event-oom_kill"]')
    expect(row.text()).toContain('Out-of-memory kill')
    expect(row.text()).toContain('postgres')
    expect(row.text()).toContain('4711')
  })

  // T021 — "37 kills" and "1 kill" are different findings and must not look
  // alike, so the count is shown rather than hidden behind a plural.
  it('shows the occurrence count when a report aggregates several', () => {
    expect(build([evt({ occurrences: 37 })]).get('[data-test="host-event-occurrences"]').text()).toContain('37')
  })

  it('does not clutter a single occurrence with a count', () => {
    expect(build([evt({ occurrences: 1 })]).find('[data-test="host-event-occurrences"]').exists()).toBe(false)
  })

  // T021 — the distinct list tells a storm killing forty copies of one process
  // from one killing forty different processes.
  it('lists the distinct processes a storm affected', () => {
    const w = build([
      evt({
        occurrences: 40,
        detail: {
          process: 'postgres',
          pid: 1,
          cgroup: null,
          distinctProcesses: ['postgres', 'node', 'python3'],
          distinctTruncated: false,
        },
      }),
    ])
    const distinct = w.get('[data-test="host-event-distinct"]')
    expect(distinct.text()).toContain('postgres')
    expect(distinct.text()).toContain('node')
    expect(w.find('[data-test="host-event-truncated"]').exists()).toBe(false)
  })

  // T021 — a truncated list must SAY it is truncated. A partial list read as
  // complete is worse than an obviously partial one.
  it('marks a truncated distinct list rather than passing it off as complete', () => {
    const w = build([
      evt({
        detail: {
          process: 'postgres',
          pid: 1,
          cgroup: null,
          distinctProcesses: ['a', 'b', 'c'],
          distinctTruncated: true,
        },
      }),
    ])
    expect(w.find('[data-test="host-event-truncated"]').exists()).toBe(true)
  })

  // T021 — the kind set is open. An older interface paired with a newer agent
  // must show what it does not fully understand rather than hide it.
  it('renders an unrecognised kind instead of dropping it', () => {
    const w = build([evt({ kind: 'something_new' })])
    expect(w.find('[data-test="host-event-something_new"]').exists()).toBe(true)
    expect(w.text()).toContain('something_new')
  })

  // An event whose source reported a count but no process — the cgroup counter
  // says how many, not which — still renders.
  it('renders an event with no detail at all', () => {
    const w = build([evt({ detail: null, occurrences: 12 })])
    expect(w.get('[data-test="host-event-oom_kill"]').text()).toContain('Out-of-memory kill')
    expect(w.find('[data-test="host-event-distinct"]').exists()).toBe(false)
  })

  it('orders newest first', () => {
    const w = build([
      evt({ id: 'old', occurredAt: '2026-09-09T10:00:00Z' }),
      evt({ id: 'new', occurredAt: '2026-09-09T18:00:00Z', kind: 'segfault' }),
    ])
    const rows = w.findAll('li')
    expect(rows).toHaveLength(2)
    expect(rows[0]?.text()).toContain('Segmentation fault')
  })
})
