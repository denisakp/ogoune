import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import type { HostContext } from '@/types'

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn(), resolve: () => ({ href: '#' }) }),
  useRoute: () => ({
    query: {},
    params: { id: 'i1' },
    path: '/incidents/i1',
    name: 'IncidentDetail',
  }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

const baseIncident = {
  id: 'i1',
  resource_id: 'r1',
  resource: { id: 'r1', name: 'api', type: 'http', status: 'down' },
  cause: 'HTTP 500',
  started_at: new Date(Date.now() - 600_000).toISOString(),
  resolved_at: null,
  created_at: '',
  updated_at: '',
  event_steps: [],
}

const getIncidentMock = vi.fn()

vi.mock('@/stores/incidentStore', () => ({
  useIncidentStore: () => ({ getIncidentById: getIncidentMock }),
}))

// Everything except HostContextPanel is stubbed: this spec is about whether the
// host context block exists in the DOM at all, not about what it renders.
vi.mock('@/components/incidents/IncidentHeader.vue', () => ({
  default: { name: 'IncidentHeader', template: '<div />', props: ['incident'] },
}))
vi.mock('@/components/incidents/IncidentTimeline.vue', () => ({
  default: { name: 'IncidentTimeline', template: '<div />', props: ['events'] },
}))
vi.mock('@/components/incidents/DiagnosticsPanel.vue', () => ({
  default: { name: 'DiagnosticsPanel', template: '<div />', props: ['diagnostics'] },
}))
vi.mock('@/components/incidents/NotificationsPanel.vue', () => ({
  default: { name: 'NotificationsPanel', template: '<div />', props: ['events'] },
}))
vi.mock('@/components/incidents/IncidentStatusUpdates.vue', () => ({
  default: { name: 'IncidentStatusUpdates', template: '<div />', props: ['incidentId'] },
}))

import IncidentView from '../IncidentView.vue'

const stubs = {
  UEmpty: { template: '<div />' },
  UIcon: { template: '<span />' },
  RouterLink: { template: '<a><slot /></a>', props: ['to'] },
}

const hostContext: HostContext = {
  host_id: 'host-1',
  host_name: 'web-01',
  peak_cpu_pct: 97.4,
  peak_mem_pct: 99.1,
  worst_disk: { mount: '/var', used_pct: 96.2 },
  sample_count: 36,
  resolution: 'full',
  window_from: '2026-03-10T11:53:00Z',
  window_to: '2026-03-10T11:59:00Z',
}

async function buildWith(hc: HostContext | null) {
  getIncidentMock.mockResolvedValue({ ...baseIncident, host_context: hc })
  setActivePinia(createPinia())
  const w = mount(IncidentView, { global: { stubs } })
  await flushPromises()
  return w
}

beforeEach(() => {
  getIncidentMock.mockReset()
})

describe('IncidentView host context', () => {
  // T019 / FR-015 — the whole point: absent means absent. No placeholder, no
  // empty-state text, no reserved container.
  it('renders nothing at all when host_context is null', async () => {
    const w = await buildWith(null)
    expect(w.findComponent({ name: 'HostContextPanel' }).exists()).toBe(false)
    expect(w.find('[data-test="host-context-cpu"]').exists()).toBe(false)
    expect(w.text()).not.toContain('Host context')
  })

  it('renders nothing when the field is missing entirely', async () => {
    getIncidentMock.mockResolvedValue({ ...baseIncident })
    setActivePinia(createPinia())
    const w = mount(IncidentView, { global: { stubs } })
    await flushPromises()
    expect(w.findComponent({ name: 'HostContextPanel' }).exists()).toBe(false)
  })

  it('renders the block when a context is present', async () => {
    const w = await buildWith(hostContext)
    expect(w.findComponent({ name: 'HostContextPanel' }).exists()).toBe(true)
    expect(w.text()).toContain('Host context')
  })

  // The sibling panels must be unaffected either way — this addition is additive
  // on the layout as well as on the contract.
  it('leaves the existing panels in place whether or not the context exists', async () => {
    for (const hc of [null, hostContext]) {
      const w = await buildWith(hc)
      expect(w.findComponent({ name: 'DiagnosticsPanel' }).exists()).toBe(true)
      expect(w.findComponent({ name: 'NotificationsPanel' }).exists()).toBe(true)
    }
  })
})
