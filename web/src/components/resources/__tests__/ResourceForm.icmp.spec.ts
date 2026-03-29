import { nextTick, ref } from 'vue'
import type { ComponentPublicInstance } from 'vue'
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import ResourceForm from '@/components/resources/ResourceForm.vue'
import type { CreateResource, SystemCapabilities } from '@/types'

const { addResourceMock, updateResourceDataMock, loadCapabilitiesMock, capabilitiesProxy } =
  vi.hoisted(() => {
    const proxy = { value: null as SystemCapabilities | null }
    return {
      addResourceMock: vi.fn(),
      updateResourceDataMock: vi.fn(),
      loadCapabilitiesMock: vi.fn(),
      capabilitiesProxy: proxy,
    }
  })

vi.mock('@/composables/useResources.ts', () => ({
  useResources: () => ({
    addResource: addResourceMock,
    updateResourceData: updateResourceDataMock,
    capabilities: capabilitiesProxy,
    loadCapabilities: loadCapabilitiesMock,
  }),
}))

vi.mock('@/composables/useTags.ts', () => ({
  useTags: () => ({
    tags: ref([]),
    loadTags: vi.fn(),
  }),
}))

vi.mock('@/composables/useComponents.ts', () => ({
  useComponents: () => ({
    components: ref([]),
    loadComponents: vi.fn(),
  }),
}))

describe('ResourceForm ICMP option and availability warning', () => {
  type ResourceFormVm = ComponentPublicInstance & {
    form: CreateResource & { component_id?: string }
  }

  const setFormValues = (
    wrapper: ReturnType<typeof mount>,
    values: Partial<CreateResource & { component_id?: string }>,
  ) => {
    Object.assign((wrapper.vm as unknown as ResourceFormVm).form, values)
  }

  beforeEach(() => {
    addResourceMock.mockReset()
    updateResourceDataMock.mockReset()
    loadCapabilitiesMock.mockReset()
    addResourceMock.mockResolvedValue(undefined)
    capabilitiesProxy.value = null
  })

  it('renders ICMP as a type option in the select', () => {
    const wrapper = mount(ResourceForm)
    const html = wrapper.html()
    expect(html).toContain('ICMP')
  })

  it('shows ICMP-specific target placeholder when type is icmp', async () => {
    const wrapper = mount(ResourceForm)
    setFormValues(wrapper, { type: 'icmp' })
    await nextTick()
    const targetInput = wrapper.find('input[placeholder*="hostname"]')
    expect(targetInput.exists()).toBe(true)
  })

  it('shows ICMP unavailability warning when capabilities are loaded and ICMP is not available', async () => {
    capabilitiesProxy.value = {
      icmp: {
        enabled: true,
        capability_available: false,
        reason: 'requires root or CAP_NET_RAW',
      },
    }
    const wrapper = mount(ResourceForm)
    setFormValues(wrapper, { type: 'icmp' })
    await nextTick()
    const html = wrapper.html()
    expect(html.toLowerCase()).toMatch(/icmp.*unavailable|unavailable.*icmp|cap_net_raw|root/i)
  })

  it('shows ICMP disabled warning when ICMP feature flag is off', async () => {
    capabilitiesProxy.value = {
      icmp: {
        enabled: false,
        capability_available: false,
        reason: '',
      },
    }
    const wrapper = mount(ResourceForm)
    setFormValues(wrapper, { type: 'icmp' })
    await nextTick()
    const html = wrapper.html()
    expect(html.toLowerCase()).toMatch(/icmp.*disabled|enable_icmp/i)
  })

  it('calls loadCapabilities on mount', async () => {
    mount(ResourceForm)
    await nextTick()
    expect(loadCapabilitiesMock).toHaveBeenCalledTimes(1)
  })

  it('does not show ICMP warning when capabilities show ICMP is available', async () => {
    capabilitiesProxy.value = {
      icmp: {
        enabled: true,
        capability_available: true,
        reason: '',
      },
    }
    const wrapper = mount(ResourceForm)
    setFormValues(wrapper, { type: 'icmp' })
    await nextTick()
    // Should NOT show unavailability or disabled warnings
    const html = wrapper.html()
    expect(html.toLowerCase()).not.toMatch(/icmp.*unavailable|cap_net_raw/i)
    expect(html.toLowerCase()).not.toMatch(/icmp.*disabled/i)
  })
})
