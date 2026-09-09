import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import type { DatabaseHealth } from '@/types'

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn(), resolve: () => ({ href: '#' }) }),
  useRoute: () => ({
    query: {},
    params: { id: 'r1' },
    path: '/resources/r1',
    name: 'ResourceDetail',
  }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

vi.mock('@/composables/useConfirm', () => ({ useConfirm: () => vi.fn() }))

const baseResource = {
  id: 'r1',
  name: 'primary-db',
  type: 'protocol',
  status: 'up',
  target: 'postgres://db.acme.com:5432/app',
  interval: 60,
  last_checked: new Date().toISOString(),
}

const loadResourceMock = vi.fn()

vi.mock('@/stores/resourceStore', () => ({
  useResourceStore: () => ({
    loadResource: loadResourceMock,
    loadResourceWithResponseTimes: loadResourceMock,
    removeResource: vi.fn().mockResolvedValue(true),
    pauseMonitoring: vi.fn().mockResolvedValue(true),
    resumeMonitoring: vi.fn().mockResolvedValue(true),
  }),
}))

vi.mock('@/services/activityService', () => ({ fetchActivities: () => Promise.resolve([]) }))
vi.mock('@/services/resourceService', () => ({
  fetchUptimeStats: () => Promise.resolve({ resource_id: 'r1', stats: [] }),
}))
vi.mock('@/components/incidents/IncidentsListBody.vue', () => ({
  default: { name: 'IncidentsListBody', template: '<div />', props: ['filter'] },
}))
vi.mock('@/components/resources/ResourceModal.vue', () => ({ default: { template: '<div />' } }))
vi.mock('@/components/ResponseTimeChart.vue', () => ({ default: { template: '<div />' } }))
vi.mock('@/components/hosts/MonitorHostPanel.vue', () => ({
  default: { name: 'MonitorHostPanel', template: '<div />', props: ['hostId', 'monitorId'] },
}))

import ResourceDetailView from '../ResourceDetailView.vue'

const stubs = {
  UButton: { template: '<button><slot /></button>' },
  UIcon: { template: '<span />' },
}

const health: DatabaseHealth = {
  connections_active: 195,
  connections_max: 200,
  longest_query_seconds: 412.5,
  replication_lag_seconds: null,
  privilege_limited: false,
  unsupported_version: false,
  collected_at: '2026-09-09T14:03:12Z',
}

async function buildWith(dbHealth: DatabaseHealth | null | undefined) {
  loadResourceMock.mockResolvedValue(
    dbHealth === undefined ? { ...baseResource } : { ...baseResource, database_health: dbHealth },
  )
  setActivePinia(createPinia())
  const w = mount(ResourceDetailView, { global: { stubs } })
  await flushPromises()
  return w
}

beforeEach(() => {
  loadResourceMock.mockReset()
})

describe('ResourceDetailView database health', () => {
  // T023 / FR-019 — the point of this spec. Absent means absent: no placeholder,
  // no empty-state text, no reserved container. The component spec only ever
  // exercises the populated panel, so without this nothing covers the null case.
  it('renders nothing at all when database_health is null', async () => {
    const w = await buildWith(null)
    expect(w.findComponent({ name: 'DatabaseHealthPanel' }).exists()).toBe(false)
    expect(w.find('[data-test="db-health-connections"]').exists()).toBe(false)
    expect(w.text()).not.toContain('Database health')
  })

  // A non-database monitor never has the field at all.
  it('renders nothing when the field is missing entirely', async () => {
    const w = await buildWith(undefined)
    expect(w.findComponent({ name: 'DatabaseHealthPanel' }).exists()).toBe(false)
    expect(w.text()).not.toContain('Database health')
  })

  it('renders the panel when health is present', async () => {
    const w = await buildWith(health)
    expect(w.findComponent({ name: 'DatabaseHealthPanel' }).exists()).toBe(true)
    expect(w.get('[data-test="db-health-connections"]').text()).toContain('195 / 200')
  })

  // The addition is additive on the layout as well as on the contract: the page
  // renders its own content identically whether or not health is present.
  it('leaves the rest of the page unchanged either way', async () => {
    const without = await buildWith(null)
    const with_ = await buildWith(health)

    for (const w of [without, with_]) {
      expect(w.text()).toContain('primary-db')
      expect(w.findComponent({ name: 'MonitorHostPanel' }).exists()).toBe(false)
    }
    // The only difference between the two renders is the health block itself.
    expect(without.findComponent({ name: 'DatabaseHealthPanel' }).exists()).toBe(false)
    expect(with_.findComponent({ name: 'DatabaseHealthPanel' }).exists()).toBe(true)
  })
})
