import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import StatusIncidentView from '../StatusIncidentView.vue'
import type { PublicIncidentDetail, PublicStatusSummary } from '@/types'

vi.mock('@/services/statusPublicService', () => ({
  fetchPublicStatusSummary: vi.fn(),
  fetchPublicStatusIncidents: vi.fn(),
  fetchPublicStatusUptime: vi.fn(),
  fetchPublicStatusResourceWindows: vi.fn(),
  fetchPublicIncidentDetail: vi.fn(),
}))
vi.mock('@/composables/useRuntimeConfig', () => ({
  loadRuntimeConfig: vi.fn().mockResolvedValue({
    ssl_provider: 'external',
    edition: 'community',
    version: 'test',
    powered_by_required: true,
  }),
}))

import * as svc from '@/services/statusPublicService'

const summary: PublicStatusSummary = {
  generated_at: '2026-06-04T12:00:00Z',
  branding: { name: 'Acme' },
  uptime_window: { earliest_day: '2020-01-01', latest_day: '2026-06-04' },
  verdict: { status: 'operational', label: 'OK', color: 'green' },
  components: [],
  standalone_resources: [],
  current_month_incidents: [],
}

function mkDetail(overrides: Partial<PublicIncidentDetail> = {}): PublicIncidentDetail {
  return {
    id: 'inc-1',
    title: 'Elevated errors on Opus 4.7',
    severity: 'major',
    started_at: '2026-06-03T07:10:00Z',
    resolved_at: '2026-06-03T07:38:00Z',
    resource_id: 'res-api',
    updates: [
      { id: 'u3', status: 'resolved', message: 'This incident has been resolved.', posted_at: '2026-06-03T07:38:00Z' },
      { id: 'u2', status: 'monitoring', message: 'A fix has been implemented and we are monitoring the results.', posted_at: '2026-06-03T07:28:00Z' },
      { id: 'u1', status: 'investigating', message: 'We are currently investigating this issue.', posted_at: '2026-06-03T07:10:00Z' },
    ],
    ...overrides,
  }
}

function buildRouter(id = 'inc-1') {
  return createRouter({
    history: createMemoryHistory(`/incidents/${id}`),
    routes: [{ path: '/incidents/:id', name: 'PublicStatusIncident', component: StatusIncidentView }],
  })
}

async function render(detail: PublicIncidentDetail, id = 'inc-1') {
  vi.mocked(svc.fetchPublicStatusSummary).mockResolvedValue(summary)
  vi.mocked(svc.fetchPublicIncidentDetail).mockResolvedValue(detail)
  const router = buildRouter(id)
  await router.push(`/incidents/${id}`)
  await router.isReady()
  const w = mount(StatusIncidentView, { global: { plugins: [router] } })
  await flushPromises()
  return w
}

describe('StatusIncidentView — US7', () => {
  beforeEach(() => vi.clearAllMocks())

  it('renders the incident title and "Incident Report for <brand>" subtitle', async () => {
    const w = await render(mkDetail())
    expect(w.text()).toContain('Elevated errors on Opus 4.7')
    expect(w.text()).toContain('Incident Report for Acme')
  })

  it('paints the title in major-severity color', async () => {
    const w = await render(mkDetail({ severity: 'major' }))
    expect(w.find('h1').classes()).toContain('text-orange-500')
  })

  it('renders one timeline entry per update in newest-first order', async () => {
    const w = await render(mkDetail())
    const entries = w.findAll('[data-update-status]')
    expect(entries).toHaveLength(3)
    expect(entries[0]?.attributes('data-update-status')).toBe('resolved')
    expect(entries[1]?.attributes('data-update-status')).toBe('monitoring')
    expect(entries[2]?.attributes('data-update-status')).toBe('investigating')
  })

  it('shows the affected resource block', async () => {
    const w = await render(mkDetail({ resource_id: 'res-api' }))
    expect(w.find('[data-section="affected"]').exists()).toBe(true)
    expect(w.text()).toContain('res-api')
  })

  it('reloads when the route id changes', async () => {
    const router = buildRouter('inc-1')
    vi.mocked(svc.fetchPublicStatusSummary).mockResolvedValue(summary)
    vi.mocked(svc.fetchPublicIncidentDetail).mockResolvedValue(mkDetail({ id: 'inc-1' }))
    await router.push('/incidents/inc-1')
    await router.isReady()
    const w = mount(StatusIncidentView, { global: { plugins: [router] } })
    await flushPromises()

    vi.mocked(svc.fetchPublicIncidentDetail).mockResolvedValue(mkDetail({ id: 'inc-2', title: 'Second issue' }))
    await router.push('/incidents/inc-2')
    await flushPromises()
    expect(w.text()).toContain('Second issue')
  })
})
