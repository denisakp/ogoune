import { describe, expect, it } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import CronGenerator from './CronGenerator.vue'

type Vm = {
  expr: string
  humanReadable: string
  isValid: boolean
  nextOccurrences: string[]
  onInput: (v: string) => void
}

describe('CronGenerator', () => {
  it('renders a human-readable preview for a valid cron', async () => {
    const w = mount(CronGenerator, { props: { modelValue: '0 2 * * *' } })
    await flushPromises()
    const vm = w.vm as unknown as Vm
    expect(vm.isValid).toBe(true)
    expect(vm.humanReadable.toLowerCase()).toContain('at 02:00')
  })

  it('marks invalid cron and exposes "Invalid cron expression" preview', async () => {
    const w = mount(CronGenerator, { props: { modelValue: 'not a cron' } })
    await flushPromises()
    const vm = w.vm as unknown as Vm
    expect(vm.isValid).toBe(false)
    expect(vm.humanReadable).toBe('Invalid cron expression')
  })

  it('next-5 occurrences list updates reactively when expression changes', async () => {
    const w = mount(CronGenerator, { props: { modelValue: '0 2 * * *' } })
    await flushPromises()
    const vm = w.vm as unknown as Vm
    const before = [...vm.nextOccurrences]
    vm.onInput('0 5 * * *')
    await flushPromises()
    const after = [...vm.nextOccurrences]
    expect(after.length).toBe(5)
    expect(after[0]).not.toBe(before[0])
  })
})
