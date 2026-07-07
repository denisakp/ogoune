import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

vi.mock('@nuxt/ui/composables/useToast', () => ({ useToast: () => ({ add: vi.fn() }) }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: vi.fn() }), useRoute: () => ({ query: {}, params: {}, path: '/toolbox/whois' }) }))
vi.mock('@/services/toolboxService', () => ({ whoisLookup: vi.fn() }))

import WhoisToolView from '../WhoisToolView.vue'

const stubs = {
  UForm: { template: '<form><slot /></form>' },
  UFormField: { template: '<div><slot /></div>' },
  UInput: { template: '<input />' },
  UButton: { template: '<button><slot /></button>' },
  UAlert: { template: '<div />' },
  UBadge: { template: '<span><slot /></span>' },
}

describe('WhoisToolView', () => {
  it('mounts and shows the empty state', () => {
    expect(mount(WhoisToolView, { global: { stubs } }).text()).toContain('Enter a domain')
  })
})
