import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'

vi.mock('@nuxt/ui/composables/useToast', () => ({ useToast: () => ({ add: vi.fn() }) }))
import MetricsView from '../MetricsView.vue'

const stubs = {
  // Forward attrs so @click (onClick) reaches the stub's native button.
  UButton: { inheritAttrs: false, template: '<button v-bind="$attrs"><slot /></button>' },
  UBadge: { template: '<span><slot /></span>' },
}

describe('MetricsView', () => {
  beforeEach(() => {
    vi.stubGlobal('URL', {
      createObjectURL: vi.fn(() => 'blob:mock'),
      revokeObjectURL: vi.fn(),
    })
  })

  it('renders the endpoint, catalog, and integrations', () => {
    const w = mount(MetricsView, { global: { stubs } })
    expect(w.text()).toContain('/metrics')
    expect(w.text()).toContain('ogoune_resource_up')
    expect(w.text()).toContain('Grafana')
  })

  it('Import dashboard triggers a JSON download', async () => {
    const w = mount(MetricsView, { global: { stubs } })
    await w.find('[data-testid="grafana-import"]').trigger('click')
    expect(URL.createObjectURL).toHaveBeenCalledOnce()
    const blob = (URL.createObjectURL as unknown as ReturnType<typeof vi.fn>).mock.calls[0]![0] as Blob
    expect(blob.type).toBe('application/json')
  })

  it('Alertmanager Examples triggers a YAML download', async () => {
    const w = mount(MetricsView, { global: { stubs } })
    await w.find('[data-testid="alertmanager-examples"]').trigger('click')
    expect(URL.createObjectURL).toHaveBeenCalledOnce()
    const blob = (URL.createObjectURL as unknown as ReturnType<typeof vi.fn>).mock.calls[0]![0] as Blob
    expect(blob.type).toBe('text/yaml')
  })
})
