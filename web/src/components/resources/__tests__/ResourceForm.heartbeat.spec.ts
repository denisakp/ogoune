import { nextTick, ref } from 'vue'
import type { ComponentPublicInstance } from 'vue'
import { mount } from '@vue/test-utils'
import { describe, it, expect, beforeEach, vi } from 'vitest'

vi.mock('pinia', () => ({
  storeToRefs: (store: Record<string, unknown>) => {
    const refs: Record<string, unknown> = {}
    for (const key of Object.keys(store)) {
      if (key.startsWith('$') || typeof store[key] === 'function') continue
      refs[key] = store[key]
    }
    return refs
  },
  defineStore: vi.fn(),
}))

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

vi.mock('@/stores/resourceStore', () => ({
  useResourceStore: () => ({
    addResource: addResourceMock,
    updateResourceData: updateResourceDataMock,
    capabilities: capabilitiesProxy,
    loadCapabilities: loadCapabilitiesMock,
    $id: 'resource',
  }),
}))

vi.mock('@/stores/tagStore', () => ({
  useTagStore: () => ({
    tags: ref([]),
    fetchTags: vi.fn(),
    $id: 'tag',
  }),
}))

vi.mock('@/stores/componentStore', () => ({
  useComponentStore: () => ({
    components: ref([]),
    loadComponents: vi.fn(),
    $id: 'component',
  }),
}))

type ResourceFormVm = ComponentPublicInstance & {
  form: CreateResource & { component_id?: string }
}

describe('ResourceForm — Heartbeat type', () => {
  beforeEach(() => {
    addResourceMock.mockReset()
    updateResourceDataMock.mockReset()
  })

  it('renders Heartbeat / Push option in the type selector', () => {
    const wrapper = mount(ResourceForm, { global: { stubs: { 'a-icon': true } } })
    expect(wrapper.html()).toContain('Heartbeat')
  })

  it('shows heartbeat-interval and heartbeat-grace fields when type is heartbeat', async () => {
    const wrapper = mount(ResourceForm, { global: { stubs: { 'a-icon': true } } })
    const vm = wrapper.vm as ResourceFormVm
    vm.form.type = 'heartbeat'
    await nextTick()

    expect(wrapper.find('[data-testid="heartbeat-interval"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="heartbeat-grace"]').exists()).toBe(true)
  })

  it('hides target field when type is heartbeat', async () => {
    const wrapper = mount(ResourceForm, { global: { stubs: { 'a-icon': true } } })
    const vm = wrapper.vm as ResourceFormVm
    vm.form.type = 'heartbeat'
    await nextTick()

    // target input should not be rendered
    const inputs = wrapper.findAll('input')
    const targetInput = inputs.find((i) => i.attributes('placeholder')?.includes('https://'))
    expect(targetInput).toBeUndefined()
  })

  it('calls addResource with heartbeat_interval and heartbeat_grace on submit', async () => {
    addResourceMock.mockResolvedValue({})
    const wrapper = mount(ResourceForm, { global: { stubs: { 'a-icon': true } } })
    const vm = wrapper.vm as ResourceFormVm

    vm.form.name = 'Backup Job'
    vm.form.type = 'heartbeat'
    vm.form.heartbeat_interval = 300
    vm.form.heartbeat_grace = 60
    await nextTick()

    // Trigger submit
    await (wrapper.vm as unknown as { handleSubmit: () => Promise<void> }).handleSubmit()

    expect(addResourceMock).toHaveBeenCalledOnce()
    const payload = addResourceMock.mock.calls[0]?.[0]
    expect(payload.type).toBe('heartbeat')
    expect(payload.heartbeat_interval).toBe(300)
    expect(payload.heartbeat_grace).toBe(60)
  })

  it('rejects invalid heartbeat_interval (below 60)', async () => {
    const wrapper = mount(ResourceForm, { global: { stubs: { 'a-icon': true } } })
    const vm = wrapper.vm as ResourceFormVm

    vm.form.name = 'Bad Interval'
    vm.form.type = 'heartbeat'
    vm.form.heartbeat_interval = 10 // invalid
    vm.form.heartbeat_grace = 60
    await nextTick()

    await (wrapper.vm as unknown as { handleSubmit: () => Promise<void> }).handleSubmit()
    expect(addResourceMock).not.toHaveBeenCalled()
  })

  it('rejects invalid heartbeat_grace (above 3600)', async () => {
    const wrapper = mount(ResourceForm, { global: { stubs: { 'a-icon': true } } })
    const vm = wrapper.vm as ResourceFormVm

    vm.form.name = 'Bad Grace'
    vm.form.type = 'heartbeat'
    vm.form.heartbeat_interval = 300
    vm.form.heartbeat_grace = 9999 // invalid
    await nextTick()

    await (wrapper.vm as unknown as { handleSubmit: () => Promise<void> }).handleSubmit()
    expect(addResourceMock).not.toHaveBeenCalled()
  })
})
