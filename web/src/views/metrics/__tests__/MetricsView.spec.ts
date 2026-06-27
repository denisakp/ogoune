import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

vi.mock('@nuxt/ui/composables/useToast', () => ({ useToast: () => ({ add: vi.fn() }) }))
import MetricsView from '../MetricsView.vue'

const stubs = {
  UButton: { template: '<button><slot /></button>' },
  UBadge: { template: '<span><slot /></span>' },
}

describe('MetricsView', () => {
  it('renders the endpoint, catalog, and integrations', () => {
    const w = mount(MetricsView, { global: { stubs } })
    expect(w.text()).toContain('/metrics')
    expect(w.text()).toContain('ogoune_resource_up')
    expect(w.text()).toContain('Grafana')
  })
})
