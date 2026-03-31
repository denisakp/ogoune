import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import type { Resource } from '@/types'

vi.mock('@/services/resourceService', () => ({
  fetchResources: vi.fn().mockResolvedValue([]),
  fetchResource: vi.fn(),
  fetchCapabilities: vi.fn(),
}))

function makeResource(overrides: Partial<Resource>): Resource {
  return {
    id: overrides.id ?? 'r1',
    name: overrides.name ?? 'Monitor',
    type: overrides.type ?? 'http',
    target: 'https://example.com',
    interval: 300,
    timeout: 10,
    status: overrides.status ?? 'up',
    is_active: true,
    failure_count: 0,
    confirmation_checks: 2,
    confirmation_interval: 30,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    tags: [],
    ...overrides,
  }
}

describe('resourceStore — heartbeat waiting exclusion from UP/DOWN totals', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('counts up/down for non-heartbeat monitors normally', async () => {
    const { useResourceStore } = await import('./resourceStore')
    const store = useResourceStore()

    store.resources = [
      makeResource({ id: 'r1', type: 'http', status: 'up' }),
      makeResource({ id: 'r2', type: 'tcp', status: 'down' }),
      makeResource({ id: 'r3', type: 'dns', status: 'up' }),
    ]

    expect(store.upCount).toBe(2)
    expect(store.downCount).toBe(1)
    expect(store.waitingCount).toBe(0)
  })

  it('excludes waiting heartbeat monitors from UP/DOWN counts', async () => {
    const { useResourceStore } = await import('./resourceStore')
    const store = useResourceStore()

    store.resources = [
      makeResource({ id: 'r1', type: 'http', status: 'up' }),
      makeResource({ id: 'r2', type: 'heartbeat', status: 'waiting', waiting: true }),
      makeResource({ id: 'r3', type: 'heartbeat', status: 'up', waiting: false }),
    ]

    // Waiting heartbeat should not count toward up
    expect(store.upCount).toBe(2) // r1 + r3 (active heartbeat)
    expect(store.downCount).toBe(0)
    expect(store.waitingCount).toBe(1) // r2
  })

  it('counts waiting heartbeats separately', async () => {
    const { useResourceStore } = await import('./resourceStore')
    const store = useResourceStore()

    store.resources = [
      makeResource({ id: 'hb1', type: 'heartbeat', status: 'waiting', waiting: true }),
      makeResource({ id: 'hb2', type: 'heartbeat', status: 'waiting', waiting: true }),
      makeResource({ id: 'hb3', type: 'heartbeat', status: 'up', waiting: false }),
    ]

    expect(store.upCount).toBe(1) // hb3 only
    expect(store.downCount).toBe(0)
    expect(store.waitingCount).toBe(2) // hb1 + hb2
  })

  it('includes down heartbeat monitors in down count (not waiting)', async () => {
    const { useResourceStore } = await import('./resourceStore')
    const store = useResourceStore()

    store.resources = [
      makeResource({ id: 'hb1', type: 'heartbeat', status: 'down', waiting: false }),
      makeResource({ id: 'hb2', type: 'heartbeat', status: 'waiting', waiting: true }),
    ]

    expect(store.upCount).toBe(0)
    expect(store.downCount).toBe(1) // hb1 — down is counted
    expect(store.waitingCount).toBe(1) // hb2
  })
})
