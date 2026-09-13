import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

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

const hostContext = {
  host_id: 'host-1',
  host_name: 'web-01',
  peak_cpu_pct: 50,
  peak_mem_pct: 60,
  worst_disk: null,
  sample_count: 10,
  resolution: 'full',
  window_from: '2026-09-12T11:53:00Z',
  window_to: '2026-09-12T11:59:00Z',
}
const containerDeclaration = {
  state: 'declared',
  kmsg: { available: false, reason: 'unreadable' },
  cgroup_oom: { available: true },
  segfault: { available: false, reason: 'unreadable' },
  oom_detail: 'without_process',
}
const oomEvent = {
  id: 'ev1',
  kind: 'oom_kill',
  occurred_at: '2026-09-12T11:57:47Z',
  source: 'cgroup',
  occurrences: 1,
  host_id: 'host-1',
  host_name: 'web-01',
  detail: null,
}

function incident(over: Record<string, unknown>) {
  return {
    id: 'i1',
    resource_id: 'r1',
    resource: { id: 'r1', name: 'api', type: 'http', status: 'down' },
    cause: 'HTTP 500',
    started_at: '2026-09-12T11:58:00Z',
    resolved_at: null,
    created_at: '',
    updated_at: '',
    event_steps: [],
    ...over,
  }
}
const getIncidentMock = vi.fn()

vi.mock('@/stores/incidentStore', () => ({
  useIncidentStore: () => ({ getIncidentById: getIncidentMock }),
}))
for (const name of ['IncidentHeader', 'IncidentTimeline', 'DiagnosticsPanel', 'NotificationsPanel']) {
  vi.doMock(`@/components/incidents/${name}.vue`, () => ({
    default: { name, template: '<div />', props: ['incident', 'events', 'diagnostics'] },
  }))
}

import IncidentView from '../IncidentView.vue'

const stubs = { UEmpty: { template: '<div />' }, UIcon: { template: '<span />' } }

async function build(over: Record<string, unknown>) {
  setActivePinia(createPinia())
  getIncidentMock.mockResolvedValue(incident(over))
  const w = mount(IncidentView, { global: { stubs } })
  await flushPromises()
  return w
}
const notice = (w: Awaited<ReturnType<typeof build>>) =>
  w.find('[data-test="capability-notice-incident"]')

beforeEach(() => getIncidentMock.mockReset())

// Spec 093 US3: the incident page stops guessing. The sentence is spoken only
// when there is host context and no kernel events -- it explains an absence.
describe('IncidentView — capability notice (spec 093)', () => {
  it('shown for a declared container host with no events, from the FROZEN declaration', async () => {
    const w = await build({
      host_context: hostContext,
      host_link: { host_id: 'host-1', source: 'recorded', exists: true },
      host_capabilities: containerDeclaration,
      host_events: [],
    })
    expect(notice(w).exists()).toBe(true)
    expect(notice(w).attributes('data-state')).toBe('declared')
    expect(notice(w).text()).toContain('could not read the kernel log')
  })

  it('hidden when kernel events exist -- the declaration never annotates a presence', async () => {
    const w = await build({
      host_context: hostContext,
      host_capabilities: containerDeclaration,
      host_events: [oomEvent],
    })
    expect(notice(w).exists()).toBe(false)
  })

  it('hidden without host context', async () => {
    const w = await build({ host_capabilities: containerDeclaration, host_events: [] })
    expect(notice(w).exists()).toBe(false)
  })

  it('says "not known" for a pre-feature incident -- never the host\'s current declaration', async () => {
    const w = await build({
      host_context: hostContext,
      host_link: { host_id: 'host-1', source: 'inferred', exists: true },
      host_capabilities: { state: 'not_known' },
      host_events: [],
    })
    expect(notice(w).text()).toContain('is not known')
  })
})
