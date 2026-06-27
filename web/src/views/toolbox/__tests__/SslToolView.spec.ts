import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

vi.mock('@nuxt/ui/composables/useToast', () => ({ useToast: () => ({ add: vi.fn() }) }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: vi.fn() }), useRoute: () => ({ query: {}, params: {}, path: '/toolbox/ssl' }) }))
vi.mock('@/services/toolboxService', () => ({ sslCheck: vi.fn() }))

import SslToolView from '../SslToolView.vue'

const stubs = {
  UForm: { template: '<form><slot /></form>' },
  UFormField: { template: '<div><slot /></div>' },
  UInput: { template: '<input />' },
  UButton: { template: '<button><slot /></button>' },
  UAlert: { template: '<div />' },
  UBadge: { template: '<span><slot /></span>' },
}

describe('SslToolView', () => {
  it('mounts and shows the empty state', () => {
    expect(mount(SslToolView, { global: { stubs } }).text()).toContain('Enter a domain')
  })
})
