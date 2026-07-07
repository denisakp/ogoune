import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

vi.mock('@nuxt/ui/composables/useToast', () => ({ useToast: () => ({ add: vi.fn() }) }))
import MetricsView from '../MetricsView.vue'
import {
  __setIntegrationsFeedForTests,
  __resetIntegrationsForTests,
} from '@/services/integrationsService'

const stubs = {
  // Forward attrs so @click (onClick) reaches the stub's native button.
  UButton: { inheritAttrs: false, template: '<button v-bind="$attrs"><slot /></button>' },
  UBadge: { template: '<span><slot /></span>' },
}

function lastBlob(): Blob {
  const calls = (URL.createObjectURL as unknown as ReturnType<typeof vi.fn>).mock.calls
  return calls[calls.length - 1]![0] as Blob
}

describe('MetricsView', () => {
  beforeEach(() => {
    vi.stubGlobal('URL', { createObjectURL: vi.fn(() => 'blob:mock'), revokeObjectURL: vi.fn() })
    __resetIntegrationsForTests()
  })

  it('renders the endpoint, catalog, integrations, and scrape hint', () => {
    const w = mount(MetricsView, { global: { stubs } })
    expect(w.text()).toContain('/metrics')
    expect(w.text()).toContain('ogoune_resource_up')
    expect(w.text()).toContain('Grafana')
    expect(w.find('[data-testid="scrape-hint"]').exists()).toBe(true)
  })

  it('Download dashboard fetches the config-derived model and downloads JSON', async () => {
    const feed = {
      fetchAlertRules: vi.fn(async () => 'groups: []'),
      fetchGrafanaDashboard: vi.fn(async () => ({ title: 'seeded' })),
    }
    __setIntegrationsFeedForTests(feed)
    const w = mount(MetricsView, { global: { stubs } })
    await w.find('[data-testid="grafana-download"]').trigger('click')
    await flushPromises()
    expect(feed.fetchGrafanaDashboard).toHaveBeenCalled()
    expect(lastBlob().type).toBe('application/json')
  })

  it('Download rules fetches the config-derived YAML', async () => {
    const feed = {
      fetchAlertRules: vi.fn(async () => 'groups:\n- name: ogoune\n'),
      fetchGrafanaDashboard: vi.fn(async () => ({})),
    }
    __setIntegrationsFeedForTests(feed)
    const w = mount(MetricsView, { global: { stubs } })
    await w.find('[data-testid="alertmanager-examples"]').trigger('click')
    await flushPromises()
    expect(feed.fetchAlertRules).toHaveBeenCalled()
    expect(lastBlob().type).toBe('text/yaml')
  })

  it('falls back to the bundled static asset when the endpoint errors', async () => {
    __setIntegrationsFeedForTests({
      fetchAlertRules: vi.fn(async () => {
        throw new Error('boom')
      }),
      fetchGrafanaDashboard: vi.fn(async () => {
        throw new Error('boom')
      }),
    })
    const w = mount(MetricsView, { global: { stubs } })
    await w.find('[data-testid="alertmanager-examples"]').trigger('click')
    await flushPromises()
    // still downloaded (the static fallback) — no dead button
    expect(lastBlob().type).toBe('text/yaml')
  })
})
