import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import UKbd from './UKbd.vue'

describe('UKbd', () => {
  it('renders a single key', () => {
    const wrapper = mount(UKbd, { props: { keys: ['⌘'] } })
    expect(wrapper.findAll('kbd')).toHaveLength(1)
    expect(wrapper.text()).toContain('⌘')
  })

  it('renders multiple keys with a "then" separator between them', () => {
    const wrapper = mount(UKbd, { props: { keys: ['⌘', 'K'] } })
    expect(wrapper.findAll('kbd')).toHaveLength(2)
    expect(wrapper.text()).toContain('then')
  })
})
