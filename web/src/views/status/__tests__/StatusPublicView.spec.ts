import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import StatusPublicView from '../StatusPublicView.vue'
import type {
  PublicStatusSummary,
  PublicResource,
  PublicComponent,
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
import * as runtime from '@/composables/useRuntimeConfig'

function mkResource(id: string, state: PublicResource['current_state'] = 'up'): PublicResource {
  return {
    id,
    name: id,
    host: `${id}.acme.com`,
    current_state: state,
    uptime_90d_ratio: 1,
    uptime_ribbon: [{ day: '2026-06-01', ratio: 1 }],
  }
}

function mkComponent(id: string, resources: PublicResource[]): PublicComponent {
  return { id, name: id, aggregated_state: 'up', resources }
}

function mkSummary(overrides: Partial<PublicStatusSummary> = {}): PublicStatusSummary {
  return {
    generated_at: '2026-06-04T12:00:00Z',
    branding: { name: 'Acme Corp' },
    verdict: { status: 'operational', label: 'All Systems Operational', color: 'green' },
    components: [],
    standalone_resources: [],
    current_month_incidents: [],
    ...overrides,
  }
}

const router = createRouter({
  history: createMemoryHistory(),
  routes: [{ path: '/', name: 'PublicStatusCurrent', component: { template: '<div/>' } }],
})

async function renderView(summary: PublicStatusSummary) {
  vi.mocked(svc.fetchPublicStatusSummary).mockResolvedValue(summary)
  const wrapper = mount(StatusPublicView, { global: { plugins: [router] } })
  await flushPromises()
  return wrapper
}

describe('StatusPublicView — US1', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(runtime.loadRuntimeConfig).mockResolvedValue({
      ssl_provider: 'external',
      edition: 'community',
      version: 'test',
      powered_by_required: true,
    })
  })

  it('all up → banner shows operational', async () => {
    const w = await renderView(
      mkSummary({ components: [mkComponent('API', [mkResource('api1')])] }),
    )
    expect(w.find('[data-status="operational"]').exists()).toBe(true)
  })

  it('major outage → banner shows major_outage status', async () => {
    const w = await renderView(
      mkSummary({
        verdict: { status: 'major_outage', label: 'Major Outage', color: 'red' },
        components: [mkComponent('API', [mkResource('api1', 'down')])],
      }),
    )
    expect(w.find('[data-status="major_outage"]').exists()).toBe(true)
  })

  it('renders component groups with nested resources', async () => {
    const w = await renderView(
      mkSummary({
        components: [
          mkComponent('API', [mkResource('api1'), mkResource('api2')]),
          mkComponent('Web', [mkResource('web1')]),
        ],
      }),
    )
    expect(w.findAll('[data-component-id]')).toHaveLength(2)
    expect(w.findAll('[data-resource-id]')).toHaveLength(3)
  })

  it('standalone section appears only when standalone_resources is non-empty', async () => {
    const empty = await renderView(mkSummary())
    expect(empty.find('[data-section="standalone-resources"]').exists()).toBe(false)

    const populated = await renderView(
      mkSummary({ standalone_resources: [mkResource('lonely')] }),
    )
    expect(populated.find('[data-section="standalone-resources"]').exists()).toBe(true)
  })

  it('Powered by Ogoune visible in community / hidden when EE suppresses', async () => {
    const ce = await renderView(mkSummary())
    expect(ce.find('[data-testid="powered-by"]').exists()).toBe(true)

    vi.mocked(runtime.loadRuntimeConfig).mockResolvedValue({
      ssl_provider: 'external',
      edition: 'enterprise',
      version: 'test',
      powered_by_required: false,
    })
    const ee = await renderView(mkSummary())
    expect(ee.find('[data-testid="powered-by"]').exists()).toBe(false)
  })

  it('renders brand name in header + copyright', async () => {
    const w = await renderView(mkSummary())
    expect(w.html()).toContain('Acme Corp')
    expect(w.get('[data-testid="copyright"]').text()).toContain('Acme Corp')
  })
})
