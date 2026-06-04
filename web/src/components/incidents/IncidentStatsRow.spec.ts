import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import IncidentStatsRow from './IncidentStatsRow.vue'

const now = Date.now()
const day = 24 * 3_600_000

function makeIncident(opts: { startedAt: number; resolvedAt?: number | null }) {
  return {
    id: `i${Math.random()}`,
    resource_id: 'r1',
    reason: 'x',
    cause: 'x',
    started_at: new Date(opts.startedAt).toISOString(),
    resolved_at: opts.resolvedAt ? new Date(opts.resolvedAt).toISOString() : null,
    created_at: new Date(opts.startedAt).toISOString(),
    updated_at: new Date(opts.startedAt).toISOString(),
  }
}

describe('IncidentStatsRow', () => {
  it('Active count = incidents with resolved_at == null', () => {
    const incidents = [
      makeIncident({ startedAt: now - day, resolvedAt: null }),
      makeIncident({ startedAt: now - day, resolvedAt: null }),
      makeIncident({ startedAt: now - day, resolvedAt: now - day / 2 }),
    ]
    const w = mount(IncidentStatsRow, { props: { incidents } })
    expect(w.text()).toContain('Active Incidents')
    expect(w.text()).toContain('2')
  })

  it('Resolved 30d filters by resolved_at within last 30 days', () => {
    const incidents = [
      makeIncident({ startedAt: now - 60 * day, resolvedAt: now - 31 * day }), // outside window
      makeIncident({ startedAt: now - 10 * day, resolvedAt: now - 9 * day }), // inside
      makeIncident({ startedAt: now - 5 * day, resolvedAt: now - 4 * day }), // inside
    ]
    const w = mount(IncidentStatsRow, { props: { incidents } })
    const vm = w.vm as unknown as { resolved30d: unknown[] }
    expect(vm.resolved30d.length).toBe(2)
  })

  it('MTTR averages duration of resolved-within-30d', () => {
    const incidents = [
      makeIncident({ startedAt: now - day, resolvedAt: now - day + 60_000 }), // 60s
      makeIncident({ startedAt: now - day, resolvedAt: now - day + 120_000 }), // 120s
    ]
    const w = mount(IncidentStatsRow, { props: { incidents } })
    const vm = w.vm as unknown as { mttrSeconds: number | null }
    expect(vm.mttrSeconds).toBe(90)
  })
})
