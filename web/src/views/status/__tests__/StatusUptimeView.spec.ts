import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import StatusUptimeView from '../StatusUptimeView.vue'
import type { PublicStatusUptimeRange, PublicStatusSummary } from '@/types'

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

const summary: PublicStatusSummary = {
  generated_at: '2026-06-04T12:00:00Z',
  branding: { name: 'Acme' },
  // Earliest day far in the past so prev navigation is allowed under the
  // new bounds clamp.
  uptime_window: { earliest_day: '2020-01-01', latest_day: '2026-06-04' },
  verdict: { status: 'operational', label: 'OK', color: 'green' },
  components: [{ id: 'c-api', name: 'API', aggregated_state: 'up', resources: [] }],
  standalone_resources: [],
  current_month_incidents: [],
}

const emptyUptime: PublicStatusUptimeRange = { generated_at: '2026-06-04T12:00:00Z', days: [] }

const router = createRouter({
  history: createMemoryHistory(),
  routes: [{ path: '/uptime', name: 'PublicStatusUptime', component: { template: '<div/>' } }],
})

async function render(initial: PublicStatusUptimeRange = emptyUptime) {
  vi.mocked(svc.fetchPublicStatusSummary).mockResolvedValue(summary)
  vi.mocked(svc.fetchPublicStatusUptime).mockResolvedValue(initial)
  const w = mount(StatusUptimeView, { global: { plugins: [router] } })
  await flushPromises()
  return w
}

describe('StatusUptimeView — US3', () => {
  beforeEach(() => vi.clearAllMocks())

  it('renders 3 calendar panels by default', async () => {
    const w = await render()
    expect(w.findAll('[data-month]')).toHaveLength(3)
  })

  it('range navigator arrow shifts the window by 3 months', async () => {
    const w = await render()
    const initialMonths = w.findAll('[data-month]').map((n) => n.attributes('data-month'))
    await w.get('[data-testid="nav-prev"]').trigger('click')
    await flushPromises()
    const after = w.findAll('[data-month]').map((n) => n.attributes('data-month'))
    expect(after).not.toEqual(initialMonths)
  })

  it('component filter forwards component_id in the request', async () => {
    const w = await render()
    await w.get('[data-testid="filter-component"]').setValue('c-api')
    await flushPromises()
    const lastCall = vi.mocked(svc.fetchPublicStatusUptime).mock.calls.at(-1)
    expect(lastCall?.[0].component_id).toBe('c-api')
  })

  it('renders the legend in the footer slot', async () => {
    const w = await render()
    const footer = w.get('[data-testid="public-footer"]')
    expect(footer.text()).toContain('Operational')
    expect(footer.text()).toContain('Outage')
  })

  it('future-month cells render as unknown when no data', async () => {
    const w = await render()
    // No days were provided → all day cells in the visible window should be unknown.
    const banded = w.findAll('[data-band="unknown"]')
    expect(banded.length).toBeGreaterThan(0)
  })
})
