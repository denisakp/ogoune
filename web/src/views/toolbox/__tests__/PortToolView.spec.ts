import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'

vi.mock('@nuxt/ui/composables/useToast', () => ({ useToast: () => ({ add: vi.fn() }) }))
const portScan = vi.fn()
vi.mock('@/services/toolboxService', () => ({ portScan: (...a: unknown[]) => portScan(...a) }))

import PortToolView from '../PortToolView.vue'
import { portPresetValues } from '@/schemas/toolbox-port.schema'

const stubs = {
  UFormField: { template: '<div><slot /></div>' },
  UInput: { template: '<input />' },
  USelect: { template: '<select />' },
  UTextarea: { template: '<textarea />' },
  UButton: { template: '<button><slot /></button>' },
  UAlert: { template: '<div />' },
  UBadge: { template: '<span><slot /></span>' },
}

beforeEach(() => portScan.mockReset())

describe('PortToolView', () => {
  it('shows empty state before a run', () => {
    expect(mount(PortToolView, { global: { stubs } }).text()).toContain('Enter a registered host')
  })

  it('default ports text matches the common preset', () => {
    const w = mount(PortToolView, { global: { stubs } })
    // The component seeds portsText from the common preset on creation.
    expect((w.vm as unknown as { state: { portsText: string } }).state.portsText).toContain(
      String(portPresetValues.common[0]),
    )
  })
})
