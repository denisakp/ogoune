import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

vi.mock('@nuxt/ui/composables/useToast', () => ({ useToast: () => ({ add: vi.fn() }) }))
vi.mock('@/services/toolboxService', () => ({ dnsLookup: vi.fn() }))

import DnsToolView from '../DnsToolView.vue'

const stubs = {
  UForm: { template: '<form><slot /></form>' },
  UFormField: { template: '<div><slot /></div>' },
  UInput: { template: '<input />' },
  USelect: { template: '<select />' },
  UButton: { template: '<button><slot /></button>' },
  UAlert: { template: '<div />' },
  UBadge: { template: '<span><slot /></span>' },
  UIcon: { template: '<span />' },
}

describe('DnsToolView', () => {
  it('mounts and shows the empty state', () => {
    const w = mount(DnsToolView, { global: { stubs } })
    expect(w.text()).toContain('Enter a domain')
    // form + run button render
    expect(w.find('form').exists()).toBe(true)
  })
})
