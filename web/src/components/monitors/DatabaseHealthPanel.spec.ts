import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import DatabaseHealthPanel from './DatabaseHealthPanel.vue'
import type { DatabaseHealth } from '@/types'

const health = (overrides: Partial<DatabaseHealth> = {}): DatabaseHealth => ({
  connections_active: 195,
  connections_max: 200,
  longest_query_seconds: 412.5,
  replication_lag_seconds: 0.75,
  privilege_limited: false,
  unsupported_version: false,
  collected_at: '2026-09-09T14:03:12Z',
  ...overrides,
})

function build(h: DatabaseHealth) {
  return mount(DatabaseHealthPanel, { props: { health: h }, global: { stubs: { UIcon: { template: '<span />' } } } })
}

describe('DatabaseHealthPanel', () => {
  // T022 — the populated case.
  it('renders saturation, query age and replication lag', () => {
    const w = build(health())
    expect(w.get('[data-test="db-health-connections"]').text()).toContain('195 / 200')
    expect(w.get('[data-test="db-health-connections"]').text()).toContain('98%')
    expect(w.get('[data-test="db-health-longest-query"]').text()).toContain('7 min')
    expect(w.get('[data-test="db-health-replication"]').text()).toContain('750 ms')
  })

  // T022 — collected_at is always shown, so a paused monitor's figures look their age.
  it('always shows when the figures were collected', () => {
    expect(build(health()).find('[data-test="db-health-collected"]').exists()).toBe(true)
  })

  // The saturation row needs both halves. A bare active count with no maximum is
  // not a ratio and must not be shown as one.
  it('hides saturation when only half the pair is present', () => {
    expect(build(health({ connections_max: null })).find('[data-test="db-health-connections"]').exists()).toBe(false)
    expect(build(health({ connections_active: null })).find('[data-test="db-health-connections"]').exists()).toBe(false)
  })

  // T045 — a null privileged field hides its own row and nothing else.
  it('hides only the rows whose values are absent', () => {
    const w = build(health({ longest_query_seconds: null, replication_lag_seconds: null }))
    expect(w.find('[data-test="db-health-longest-query"]').exists()).toBe(false)
    expect(w.find('[data-test="db-health-replication"]').exists()).toBe(false)
    expect(w.get('[data-test="db-health-connections"]').text()).toContain('195 / 200')
  })

  // T045 / FR-022 — a missing grant is an opportunity, never a failure.
  it('names the optional grant when the credential is limited', () => {
    const w = build(health({ privilege_limited: true, longest_query_seconds: null }))
    const note = w.get('[data-test="db-health-privilege"]')
    expect(note.text()).toContain('pg_read_all_stats')
    expect(note.text()).toContain('optional')
    expect(note.text().toLowerCase()).not.toContain('error')
    expect(note.text().toLowerCase()).not.toContain('failed')
  })

  // T043 / FR-014 — a privilege gap and an old server ask for different actions,
  // so they must not read the same.
  it('distinguishes an unsupported version from a missing grant', () => {
    const w = build(health({ unsupported_version: true, privilege_limited: true }))
    expect(w.find('[data-test="db-health-unsupported"]').exists()).toBe(true)
    expect(w.find('[data-test="db-health-privilege"]').exists()).toBe(false)
    expect(w.get('[data-test="db-health-unsupported"]').text()).toContain('PostgreSQL 12+')
  })

  it('shows no limitation note when nothing is limited', () => {
    const w = build(health())
    expect(w.find('[data-test="db-health-privilege"]').exists()).toBe(false)
    expect(w.find('[data-test="db-health-unsupported"]').exists()).toBe(false)
  })

  // Saturation is coloured as a hint of pressure. It never alerts — health is
  // context, not a signal (FR-021).
  it('tints saturation as it climbs without ever claiming a failure', () => {
    expect(build(health({ connections_active: 10 })).get('[data-test="db-health-connections"]').html()).toContain('text-highlighted')
    expect(build(health({ connections_active: 160 })).get('[data-test="db-health-connections"]').html()).toContain('text-warning')
    expect(build(health({ connections_active: 195 })).get('[data-test="db-health-connections"]').html()).toContain('text-error')
  })
})
