import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

const bulkAssignMock = vi.fn().mockResolvedValue(undefined)
const createComponentMock = vi.fn().mockResolvedValue({ id: 'new-c', name: 'New' })

vi.mock('@/services/componentService', () => ({
  bulkAssignToComponent: (...a: unknown[]) => bulkAssignMock(...a),
  createComponent: (...a: unknown[]) => createComponentMock(...a),
  bulkRemoveFromComponent: vi.fn(),
}))

const loadComponentsMock = vi.fn().mockResolvedValue(undefined)
vi.mock('@/stores/componentStore', () => ({
  useComponentStore: () => ({
    components: [
      { id: 'c1', name: 'API Cluster' },
      { id: 'c2', name: 'Databases' },
    ],
    loadComponents: loadComponentsMock,
  }),
}))

import GroupResourcesModal from './GroupResourcesModal.vue'

const stubs = {
  UModal: { template: '<div><slot name="body" /></div>', props: ['open', 'title', 'ui'] },
  UInput: {
    template:
      '<input :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
    props: ['modelValue'],
  },
  UButton: {
    template: '<button @click="$emit(\'click\')" :disabled="disabled"><slot /></button>',
    props: ['disabled', 'loading'],
  },
  UIcon: { template: '<span />' },
}

function build(props: Record<string, unknown> = {}) {
  setActivePinia(createPinia())
  return mount(GroupResourcesModal, {
    global: { stubs },
    props: { open: true, selectedIds: ['r1', 'r2', 'r3'], ...props },
  })
}

beforeEach(() => {
  bulkAssignMock.mockClear()
  createComponentMock.mockClear()
  bulkAssignMock.mockResolvedValue(undefined)
  createComponentMock.mockResolvedValue({ id: 'new-c', name: 'New' })
})

describe('GroupResourcesModal', () => {
  it('mounts with default mode pick + canSubmit false', async () => {
    const w = build()
    await flushPromises()
    const vm = w.vm as unknown as { mode: 'pick' | 'create'; canSubmit: boolean }
    expect(vm.mode).toBe('pick')
    expect(vm.canSubmit).toBe(false)
  })

  it('picking an existing component then submit calls bulkAssignToComponent with selectedIds', async () => {
    const w = build()
    await flushPromises()
    const vm = w.vm as unknown as {
      pickedComponentId: string | null
      onSubmit: () => Promise<void>
    }
    vm.pickedComponentId = 'c1'
    await vm.onSubmit()
    expect(createComponentMock).not.toHaveBeenCalled()
    expect(bulkAssignMock).toHaveBeenCalledWith('c1', { resource_ids: ['r1', 'r2', 'r3'] })
  })

  it('+ New component flow: createComponent first, then bulkAssign with new ID', async () => {
    const w = build()
    await flushPromises()
    const vm = w.vm as unknown as {
      mode: 'pick' | 'create'
      newComponentName: string
      onSubmit: () => Promise<void>
    }
    vm.mode = 'create'
    vm.newComponentName = 'Payment Systems'
    await vm.onSubmit()
    expect(createComponentMock).toHaveBeenCalledWith({
      name: 'Payment Systems',
      description: undefined,
      resource_ids: ['r1', 'r2', 'r3'],
    })
    expect(bulkAssignMock).not.toHaveBeenCalled()
  })

  it('canSubmit is false until a target is selected', async () => {
    const w = build()
    await flushPromises()
    const vm = w.vm as unknown as { canSubmit: boolean; pickedComponentId: string | null }
    expect(vm.canSubmit).toBe(false)
    vm.pickedComponentId = 'c1'
    await w.vm.$nextTick()
    expect(vm.canSubmit).toBe(true)
  })
})
