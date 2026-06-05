import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import HeadersEditor from './HeadersEditor.vue'

const stubs = {
  UInput: {
    template:
      '<input :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
    props: ['modelValue', 'placeholder', 'size'],
  },
  UButton: { template: '<button @click="$emit(\'click\')"><slot /></button>' },
}

describe('HeadersEditor', () => {
  it('renders empty by default + Add header button', () => {
    const w = mount(HeadersEditor, { global: { stubs } })
    expect(w.findAll('input').length).toBe(0)
    expect(w.text()).toContain('Add header')
  })

  it('pre-populates rows from modelValue', () => {
    const w = mount(HeadersEditor, {
      global: { stubs },
      props: { modelValue: { 'X-Token': 'abc', 'X-Other': '42' } },
    })
    expect(w.findAll('input').length).toBe(4) // 2 rows × (name + value)
  })

  it('addRow appends an empty row', async () => {
    const w = mount(HeadersEditor, { global: { stubs } })
    ;(w.vm as unknown as { addRow: () => void }).addRow()
    await w.vm.$nextTick()
    expect(w.findAll('input').length).toBe(2)
  })

  it('removeRow emits update:modelValue without the removed entry', async () => {
    const w = mount(HeadersEditor, {
      global: { stubs },
      props: { modelValue: { 'X-Token': 'abc', 'X-Other': '42' } },
    })
    const vm = w.vm as unknown as { rows: Array<{ id: string }>; removeRow: (id: string) => void }
    const firstId = vm.rows[0]!.id
    vm.removeRow(firstId)
    await w.vm.$nextTick()
    const last = w.emitted('update:modelValue')?.at(-1)
    expect(last).toBeDefined()
    expect(Object.keys((last as unknown as [Record<string, string>])[0]).length).toBe(1)
  })

  it('emitChange drops empty-name rows', async () => {
    const w = mount(HeadersEditor, { global: { stubs } })
    const vm = w.vm as unknown as {
      rows: Array<{ id: string; name: string; value: string }>
      addRow: () => void
      emitChange: () => void
    }
    vm.addRow()
    vm.addRow()
    vm.rows[0]!.name = 'X-Valid'
    vm.rows[0]!.value = 'v1'
    // vm.rows[1] left empty
    vm.emitChange()
    const last = w.emitted('update:modelValue')?.at(-1) as unknown as [Record<string, string>]
    expect(last[0]).toEqual({ 'X-Valid': 'v1' })
  })
})
