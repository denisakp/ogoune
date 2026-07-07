import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import UFilterChip from './UFilterChip.vue'

describe('UFilterChip', () => {
  it('renders kind + value', () => {
    const wrapper = mount(UFilterChip, {
      props: { kind: 'tag', value: 'production' },
      global: { stubs: { UIcon: true } },
    })
    expect(wrapper.text()).toContain('tag:')
    expect(wrapper.text()).toContain('production')
    expect(wrapper.attributes('data-kind')).toBe('tag')
  })

  it('emits remove on × click', async () => {
    const wrapper = mount(UFilterChip, {
      props: { kind: 'status', value: 'down' },
      global: { stubs: { UIcon: true } },
    })
    await wrapper.find('button').trigger('click')
    expect(wrapper.emitted('remove')).toHaveLength(1)
  })
})
