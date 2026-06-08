import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import StatusHistoryView from '../StatusHistoryView.vue'
import type {
  PublicStatusIncidentsArchive,
  PublicStatusSummary,
  PublicIncidentSummary,
} from '@/types'

vi.mock('@/services/statusPublicService', () => ({
  fetchPublicStatusSummary: vi.fn(),
  fetchPublicStatusIncidents: vi.fn(),
  fetchPublicStatusUptime: vi.fn(),
  fetchPublicStatusResourceWindows: vi.fn(),
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

function mkIncident(
  id: string,
  started: string,
  resolved: string | null = null,
  severity: 'minor' | 'major' | 'critical' = 'minor',
): PublicIncidentSummary {
  return { id, title: `inc ${id}`, started_at: started, resolved_at: resolved, severity }
}

const summary: PublicStatusSummary = {
  generated_at: '2026-06-04T12:00:00Z',
  verdict: { status: 'operational', label: 'OK', color: 'green' },
  components: [
    { id: 'c-api', name: 'API', aggregated_state: 'up', resources: [] },
    { id: 'c-web', name: 'Web', aggregated_state: 'up', resources: [] },
  ],
  standalone_resources: [],
  current_month_incidents: [],
}

const router = createRouter({
  history: createMemoryHistory(),
  routes: [{ path: '/history', name: 'PublicStatusHistory', component: { template: '<div/>' } }],
})

async function render(incidents: PublicStatusIncidentsArchive) {
  vi.mocked(svc.fetchPublicStatusSummary).mockResolvedValue(summary)
  vi.mocked(svc.fetchPublicStatusIncidents).mockResolvedValue(incidents)
  const w = mount(StatusHistoryView, { global: { plugins: [router] } })
  await flushPromises()
  return w
}

describe('StatusHistoryView — US2', () => {
  beforeEach(() => vi.clearAllMocks())

  it('renders incidents grouped by month', async () => {
    const incidents: PublicStatusIncidentsArchive = {
      generated_at: '2026-06-04T12:00:00Z',
      total: 3,
      months: [
        {
          year_month: '2026-06',
          count: 1,
          incidents: [mkIncident('a', '2026-06-01T09:00:00Z', '2026-06-01T10:00:00Z')],
        },
        {
          year_month: '2026-05',
          count: 2,
          incidents: [
            mkIncident('b', '2026-05-12T09:00:00Z', '2026-05-12T10:00:00Z'),
            mkIncident('c', '2026-05-03T09:00:00Z', '2026-05-03T10:00:00Z'),
          ],
        },
      ],
    }
    const w = await render(incidents)
    const sections = w.findAll('[data-year-month]')
    expect(sections).toHaveLength(2)
    expect(sections[0]?.attributes('data-year-month')).toBe('2026-06')
    expect(w.findAll('[data-incident-id]')).toHaveLength(3)
  })

  it('component filter narrows the request', async () => {
    const empty: PublicStatusIncidentsArchive = { generated_at: 'x', total: 0, months: [] }
    const w = await render(empty)
    const select = w.get('[data-testid="filter-component"]')
    await select.setValue('c-api')
    await flushPromises()
    const lastCall = vi.mocked(svc.fetchPublicStatusIncidents).mock.calls.at(-1)
    expect(lastCall?.[0].component_id).toBe('c-api')
  })

  it('date range filter forwards from/to', async () => {
    const empty: PublicStatusIncidentsArchive = { generated_at: 'x', total: 0, months: [] }
    const w = await render(empty)
    await w.get('[data-testid="filter-from"]').setValue('2026-04-01')
    await flushPromises()
    const lastCall = vi.mocked(svc.fetchPublicStatusIncidents).mock.calls.at(-1)
    expect(lastCall?.[0].from).toBe('2026-04-01')
  })

  it('counter shows "X incidents · all resolved" when all resolved', async () => {
    const archive: PublicStatusIncidentsArchive = {
      generated_at: 'x',
      total: 2,
      months: [
        {
          year_month: '2026-06',
          count: 2,
          incidents: [
            mkIncident('a', '2026-06-01T09:00:00Z', '2026-06-01T10:00:00Z'),
            mkIncident('b', '2026-06-02T09:00:00Z', '2026-06-02T10:00:00Z'),
          ],
        },
      ],
    }
    const w = await render(archive)
    expect(w.get('[data-testid="counter"]').text()).toBe('2 incidents · all resolved')
  })

  it('empty state renders when no months', async () => {
    const empty: PublicStatusIncidentsArchive = { generated_at: 'x', total: 0, months: [] }
    const w = await render(empty)
    expect(w.find('[data-testid="empty-state"]').exists()).toBe(true)
  })
})
