import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import UEmptyState from './UEmptyState.vue'

describe('UEmptyState', () => {
  it('renders title + description + icon', () => {
    const wrapper = mount(UEmptyState, {
      props: { icon: 'i-lucide-radar', title: 'No monitors', description: 'Add one' },
      global: { stubs: { UIcon: true } },
    })
    expect(wrapper.text()).toContain('No monitors')
    expect(wrapper.text()).toContain('Add one')
  })

  it('renders the actions slot', () => {
    const wrapper = mount(UEmptyState, {
      props: { icon: 'i-lucide-radar', title: 'Empty' },
      slots: { actions: '<button data-test="cta">CTA</button>' },
      global: { stubs: { UIcon: true } },
    })
    expect(wrapper.find('[data-test="cta"]').exists()).toBe(true)
  })
})
