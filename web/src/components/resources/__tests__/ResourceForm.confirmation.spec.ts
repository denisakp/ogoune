import { nextTick, ref } from 'vue'
import type { ComponentPublicInstance } from 'vue'
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import ResourceForm from '@/components/resources/ResourceForm.vue'
import type { CreateResource } from '@/types'

const { addResourceMock, updateResourceDataMock } = vi.hoisted(() => ({
  addResourceMock: vi.fn(),
  updateResourceDataMock: vi.fn(),
}))

vi.mock('@/composables/useResources.ts', () => ({
  useResources: () => ({
    addResource: addResourceMock,
    updateResourceData: updateResourceDataMock,
    loadCapabilities: vi.fn(),
    capabilities: { value: null },
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

describe('ResourceForm confirmation validation', () => {
  type ResourceFormVm = ComponentPublicInstance & {
    form: CreateResource & { component_id?: string }
  }

  const setFormValues = (
    wrapper: ReturnType<typeof mount>,
    values: Partial<CreateResource & { component_id?: string }>,
  ) => {
    Object.assign((wrapper.vm as unknown as ResourceFormVm).form, values)
  }

  const clickSubmit = async (wrapper: ReturnType<typeof mount>) => {
    const submitButton = wrapper
      .findAll('button')
      .find(
        (node) => node.text().includes('Create Monitor') || node.text().includes('Update Monitor'),
      )

    expect(submitButton).toBeDefined()
    await submitButton!.trigger('click')
  }

  beforeEach(() => {
    addResourceMock.mockReset()
    updateResourceDataMock.mockReset()
    addResourceMock.mockResolvedValue(undefined)
  })

  it('blocks submit when confirmation_interval is equal to interval', async () => {
    const wrapper = mount(ResourceForm)

    setFormValues(wrapper, {
      name: 'API health',
      target: 'https://example.com/health',
      interval: 30,
      timeout: 10,
      confirmation_checks: 2,
      confirmation_interval: 30,
    })
    await nextTick()

    await clickSubmit(wrapper)

    expect(addResourceMock).not.toHaveBeenCalled()
  })

  it('blocks submit when confirmation_checks is not positive', async () => {
    const wrapper = mount(ResourceForm)

    setFormValues(wrapper, {
      name: 'API health',
      target: 'https://example.com/health',
      interval: 60,
      timeout: 10,
      confirmation_checks: 0,
      confirmation_interval: 30,
    })
    await nextTick()

    await clickSubmit(wrapper)

    expect(addResourceMock).not.toHaveBeenCalled()
  })

  it('submits when confirmation fields are valid', async () => {
    const wrapper = mount(ResourceForm)

    setFormValues(wrapper, {
      name: 'API health',
      target: 'https://example.com/health',
      interval: 60,
      timeout: 10,
      confirmation_checks: 3,
      confirmation_interval: 20,
    })
    await nextTick()

    await clickSubmit(wrapper)
    await nextTick()

    expect(addResourceMock).toHaveBeenCalledTimes(1)
  })
})
